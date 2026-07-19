package evidence

import (
	"context"
	"database/sql"

	"github.com/prelove/zedu/backend/internal/repository"
)

type Repository struct{}

func (Repository) NextAttachmentID(ctx context.Context, exec repository.Executor) (int64, error) {
	var nextID int64
	err := exec.QueryRowContext(ctx, `SELECT COALESCE(MAX(id), 0) + 1 FROM payment_attachment`).Scan(&nextID)
	return nextID, err
}

func (Repository) PaymentStatus(ctx context.Context, exec repository.Executor, paymentID int64) (string, error) {
	var status string
	err := exec.QueryRowContext(ctx, `SELECT status FROM student_payment WHERE id = ?`, paymentID).Scan(&status)
	return status, err
}

func (Repository) CountAttachments(ctx context.Context, exec repository.Executor, paymentID int64) (int, error) {
	var count int
	err := exec.QueryRowContext(ctx, `SELECT COUNT(*) FROM payment_attachment WHERE payment_id = ?`, paymentID).Scan(&count)
	return count, err
}

func (Repository) InsertAttachment(ctx context.Context, exec repository.Executor, attachment Attachment) (int64, error) {
	result, err := exec.ExecContext(ctx, `INSERT INTO payment_attachment (id, payment_id, file_name, file_path, file_type, file_size, uploaded_by) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		attachment.ID, attachment.PaymentID, attachment.FileName, attachment.FilePath, attachment.FileType, attachment.FileSize, attachment.UploadedBy,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (Repository) DeleteAttachment(ctx context.Context, exec repository.Executor, attachmentID int64) error {
	_, err := exec.ExecContext(ctx, `DELETE FROM payment_attachment WHERE id = ?`, attachmentID)
	return err
}

func (Repository) DeleteUploadAudit(ctx context.Context, exec repository.Executor, attachmentID int64) error {
	_, err := exec.ExecContext(ctx, `DELETE FROM operation_log WHERE action='PAYMENT_ATTACHMENT_UPLOAD' AND target_type='payment_attachment' AND target_id=?`, attachmentID)
	return err
}

func (Repository) ListAttachments(ctx context.Context, exec repository.Executor, paymentID int64, limit, offset int) ([]Attachment, int, error) {
	var total int
	if err := exec.QueryRowContext(ctx, `SELECT COUNT(*) FROM payment_attachment WHERE payment_id = ?`, paymentID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := exec.QueryContext(ctx, `SELECT id, payment_id, file_name, file_path, file_type, file_size, uploaded_by, uploaded_at FROM payment_attachment WHERE payment_id = ? ORDER BY id ASC LIMIT ? OFFSET ?`,
		paymentID, limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]Attachment, 0)
	for rows.Next() {
		var item Attachment
		if err := rows.Scan(&item.ID, &item.PaymentID, &item.FileName, &item.FilePath, &item.FileType, &item.FileSize, &item.UploadedBy, &item.UploadedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (Repository) AttachmentByID(ctx context.Context, exec repository.Executor, paymentID, attachmentID int64) (Attachment, error) {
	var item Attachment
	err := exec.QueryRowContext(ctx, `SELECT id, payment_id, file_name, file_path, file_type, file_size, uploaded_by, uploaded_at FROM payment_attachment WHERE payment_id = ? AND id = ?`,
		paymentID, attachmentID,
	).Scan(&item.ID, &item.PaymentID, &item.FileName, &item.FilePath, &item.FileType, &item.FileSize, &item.UploadedBy, &item.UploadedAt)
	return item, err
}

func (Repository) PaymentExists(ctx context.Context, exec repository.Executor, paymentID int64) (bool, error) {
	var exists int
	err := exec.QueryRowContext(ctx, `SELECT 1 FROM student_payment WHERE id = ?`, paymentID).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}
