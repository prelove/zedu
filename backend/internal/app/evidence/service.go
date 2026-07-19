package evidence

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"

	"github.com/prelove/zedu/backend/internal/platform/httpserver"
	"github.com/prelove/zedu/backend/internal/repository"
)

type Service struct {
	db      repository.DB
	repo    Repository
	storage *Storage
}

type AttachmentContent struct {
	Attachment Attachment
	File       *os.File
}

func NewService(db repository.DB, storage *Storage) *Service {
	return &Service{db: db, repo: Repository{}, storage: storage}
}

func (s *Service) UploadAttachment(ctx context.Context, user httpserver.AuthUser, paymentID int64, originalName string, body io.Reader, requestID string) (Attachment, error) {
	if !isFinanceRole(user.Role) || paymentID <= 0 {
		return Attachment{}, ErrForbidden
	}

	staged, err := s.storage.Stage(body)
	if err != nil {
		if err == ErrInvalidState {
			return Attachment{}, err
		}
		return Attachment{}, repository.ErrDatabase
	}
	defer s.storage.DiscardTemp(staged.path)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Attachment{}, repository.ErrDatabase
	}
	defer tx.Rollback()

	status, err := s.repo.PaymentStatus(ctx, tx, paymentID)
	if err == sql.ErrNoRows {
		return Attachment{}, ErrNotFound
	}
	if err != nil {
		return Attachment{}, repository.ErrDatabase
	}
	if status != "CONFIRMED" {
		return Attachment{}, ErrInvalidState
	}

	count, err := s.repo.CountAttachments(ctx, tx, paymentID)
	if err != nil {
		return Attachment{}, repository.ErrDatabase
	}
	if count >= maxAttachmentsPerPayment {
		return Attachment{}, ErrInvalidState
	}

	attachment := Attachment{
		PaymentID:  paymentID,
		FileName:   sanitizeFileName(originalName, staged.ext),
		FileType:   staged.mime,
		FileSize:   staged.size,
		UploadedBy: user.UserID,
	}

	actorName, err := repository.ActorName(tx, ctx, user.UserID)
	if err != nil {
		return Attachment{}, repository.ErrDatabase
	}

	attachmentID, err := s.repo.NextAttachmentID(ctx, tx)
	if err != nil {
		return Attachment{}, repository.ErrDatabase
	}
	attachment.ID = attachmentID
	attachment.FilePath = pathForAttachment(paymentID, attachmentID, staged.ext)
	if _, err := s.repo.InsertAttachment(ctx, tx, attachment); err != nil {
		return Attachment{}, repository.ErrDatabase
	}
	if err := repository.InsertAuditLog(tx, ctx, user.UserID, actorName, "PAYMENT_ATTACHMENT_UPLOAD", "payment_attachment", attachmentID, map[string]any{
		"paymentId":      paymentID,
		"attachmentId":   attachmentID,
		"fileName":       attachment.FileName,
		"fileType":       attachment.FileType,
		"fileSize":       attachment.FileSize,
		"uploadedBy":     user.UserID,
		"uploadedByRole": user.Role,
	}, requestID); err != nil {
		return Attachment{}, repository.ErrDatabase
	}
	if err := tx.Commit(); err != nil {
		return Attachment{}, repository.ErrDatabase
	}

	if _, err := s.storage.Publish(staged, paymentID, attachmentID); err != nil {
		if err := s.compensateAttachment(ctx, attachmentID); err != nil {
			return Attachment{}, repository.ErrDatabase
		}
		return Attachment{}, repository.ErrDatabase
	}

	item, err := s.repo.AttachmentByID(ctx, s.db, paymentID, attachmentID)
	if err != nil {
		return Attachment{}, repository.ErrDatabase
	}
	return item, nil
}

func (s *Service) ListAttachments(ctx context.Context, user httpserver.AuthUser, paymentID int64, page, pageSize int) (Page[Attachment], error) {
	if !isFinanceRole(user.Role) || paymentID <= 0 {
		return Page[Attachment]{}, ErrForbidden
	}
	exists, err := s.repo.PaymentExists(ctx, s.db, paymentID)
	if err != nil {
		return Page[Attachment]{}, repository.ErrDatabase
	}
	if !exists {
		return Page[Attachment]{}, ErrNotFound
	}

	items, total, err := s.repo.ListAttachments(ctx, s.db, paymentID, pageSize, (page-1)*pageSize)
	if err != nil {
		return Page[Attachment]{}, repository.ErrDatabase
	}
	return Page[Attachment]{Items: items, Page: page, PageSize: pageSize, Total: total}, nil
}

func (s *Service) OpenAttachment(ctx context.Context, user httpserver.AuthUser, paymentID, attachmentID int64, requestID string) (AttachmentContent, error) {
	if !isFinanceRole(user.Role) || paymentID <= 0 || attachmentID <= 0 {
		return AttachmentContent{}, ErrForbidden
	}

	item, err := s.repo.AttachmentByID(ctx, s.db, paymentID, attachmentID)
	if err == sql.ErrNoRows {
		return AttachmentContent{}, ErrNotFound
	}
	if err != nil {
		return AttachmentContent{}, repository.ErrDatabase
	}

	path, err := s.storage.ResolveContentPath(item.FilePath)
	if err != nil {
		return AttachmentContent{}, repository.ErrDatabase
	}
	file, err := os.Open(path)
	if err != nil {
		return AttachmentContent{}, repository.ErrDatabase
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		_ = file.Close()
		return AttachmentContent{}, repository.ErrDatabase
	}
	defer tx.Rollback()

	actorName, err := repository.ActorName(tx, ctx, user.UserID)
	if err != nil {
		_ = file.Close()
		return AttachmentContent{}, repository.ErrDatabase
	}
	if err := repository.InsertAuditLog(tx, ctx, user.UserID, actorName, "PAYMENT_ATTACHMENT_DOWNLOAD", "payment_attachment", attachmentID, map[string]any{
		"paymentId":      paymentID,
		"attachmentId":   attachmentID,
		"fileName":       item.FileName,
		"fileType":       item.FileType,
		"fileSize":       item.FileSize,
		"downloadedBy":   user.UserID,
		"downloadedRole": user.Role,
	}, requestID); err != nil {
		_ = file.Close()
		return AttachmentContent{}, repository.ErrDatabase
	}
	if err := tx.Commit(); err != nil {
		_ = file.Close()
		return AttachmentContent{}, repository.ErrDatabase
	}

	return AttachmentContent{Attachment: item, File: file}, nil
}

func (s *Service) compensateAttachment(ctx context.Context, attachmentID int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := s.repo.DeleteUploadAudit(ctx, tx, attachmentID); err != nil {
		return err
	}
	if err := s.repo.DeleteAttachment(ctx, tx, attachmentID); err != nil {
		return err
	}
	return tx.Commit()
}

func isFinanceRole(role string) bool {
	return role == "OWNER" || role == "OPERATOR"
}

func pathForAttachment(paymentID, attachmentID int64, ext string) string {
	return path.Join("payments", strconv.FormatInt(paymentID, 10), fmt.Sprintf("%d%s", attachmentID, ext))
}
