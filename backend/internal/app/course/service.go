package course

import (
	"context"
	"errors"
	"strings"

	"github.com/prelove/zedu/backend/internal/platform/httpserver"
	"github.com/prelove/zedu/backend/internal/repository"
)

// Service is the course-dictionary application service (stage B). It owns
// authorization, validation, hierarchy checks, transaction orchestration and
// audit. It does not depend on net/http.
type Service struct {
	db   repository.DB
	repo *Repository
}

// NewService creates a course-dictionary service backed by the given DB.
func NewService(db repository.DB) *Service {
	return &Service{db: db, repo: NewRepository()}
}

type actor struct {
	id   int64
	role string
}

func fromUser(u httpserver.AuthUser) actor {
	return actor{id: u.UserID, role: u.Role}
}

func (a actor) authorize() error {
	if a.role != "OWNER" && a.role != "OPERATOR" {
		return ErrForbidden
	}
	return nil
}

// ---------- Domain ----------

func (s *Service) CreateDomain(ctx context.Context, u httpserver.AuthUser, w DomainWrite, requestID string) (CourseDomain, error) {
	a := fromUser(u)
	if err := a.authorize(); err != nil {
		return CourseDomain{}, err
	}
	if err := validateDomainWrite(w, true); err != nil {
		return CourseDomain{}, err
	}
	var created CourseDomain
	err := s.inTx(ctx, func(tx repository.Tx) error {
		id, err := s.repo.InsertDomain(ctx, tx, w)
		if err != nil {
			return err
		}
		created, err = s.repo.GetDomain(ctx, tx, id)
		if err != nil {
			return err
		}
		return s.audit(ctx, tx, a.id, "DOMAIN_CREATE", "course_domain", id, map[string]any{"code": created.Code}, requestID)
	})
	if err != nil {
		return CourseDomain{}, err
	}
	return created, nil
}

func (s *Service) UpdateDomain(ctx context.Context, u httpserver.AuthUser, id int64, w DomainWrite, requestID string) (CourseDomain, error) {
	a := fromUser(u)
	if err := a.authorize(); err != nil {
		return CourseDomain{}, err
	}
	if err := validateDomainWrite(w, false); err != nil {
		return CourseDomain{}, err
	}
	var updated CourseDomain
	err := s.inTx(ctx, func(tx repository.Tx) error {
		if _, err := s.repo.GetDomain(ctx, tx, id); err != nil {
			return err
		}
		if err := s.repo.UpdateDomain(ctx, tx, id, w); err != nil {
			return err
		}
		var err error
		updated, err = s.repo.GetDomain(ctx, tx, id)
		if err != nil {
			return err
		}
		return s.audit(ctx, tx, a.id, "DOMAIN_UPDATE", "course_domain", id, map[string]any{"code": updated.Code}, requestID)
	})
	if err != nil {
		return CourseDomain{}, err
	}
	return updated, nil
}

func (s *Service) ListDomains(ctx context.Context, u httpserver.AuthUser, search string, page, pageSize int) (httpserver.ListData, error) {
	a := fromUser(u)
	if err := a.authorize(); err != nil {
		return httpserver.ListData{}, err
	}
	items, total, err := s.repo.ListDomains(ctx, s.db, search, pageSize, (page-1)*pageSize)
	if err != nil {
		return httpserver.ListData{}, err
	}
	return httpserver.NewListData(items, total, httpserver.PageQuery{Page: page, PageSize: pageSize}), nil
}

func validateDomainWrite(w DomainWrite, isCreate bool) error {
	if isCreate {
		if w.Name == nil || strings.TrimSpace(*w.Name) == "" {
			return ErrInvalidState
		}
		if w.Code == nil || strings.TrimSpace(*w.Code) == "" {
			return ErrInvalidState
		}
	}
	if w.Name != nil && strings.TrimSpace(*w.Name) == "" {
		return ErrInvalidState
	}
	if w.Code != nil && strings.TrimSpace(*w.Code) == "" {
		return ErrInvalidState
	}
	if w.Type != nil && *w.Type != "" && !validDomainTypes[*w.Type] {
		return ErrInvalidState
	}
	return nil
}

// ---------- Track ----------

func (s *Service) CreateTrack(ctx context.Context, u httpserver.AuthUser, w TrackWrite, requestID string) (Track, error) {
	a := fromUser(u)
	if err := a.authorize(); err != nil {
		return Track{}, err
	}
	if err := validateTrackWrite(w, true); err != nil {
		return Track{}, err
	}
	var created Track
	err := s.inTx(ctx, func(tx repository.Tx) error {
		if _, err := s.repo.GetDomain(ctx, tx, *w.DomainID); err != nil {
			return err
		}
		id, err := s.repo.InsertTrack(ctx, tx, w)
		if err != nil {
			return err
		}
		created, err = s.repo.GetTrack(ctx, tx, id)
		if err != nil {
			return err
		}
		return s.audit(ctx, tx, a.id, "TRACK_CREATE", "course_track", id, map[string]any{"code": created.Code, "domainId": created.DomainID}, requestID)
	})
	if err != nil {
		return Track{}, err
	}
	return created, nil
}

func (s *Service) UpdateTrack(ctx context.Context, u httpserver.AuthUser, id int64, w TrackWrite, requestID string) (Track, error) {
	a := fromUser(u)
	if err := a.authorize(); err != nil {
		return Track{}, err
	}
	if err := validateTrackWrite(w, false); err != nil {
		return Track{}, err
	}
	var updated Track
	err := s.inTx(ctx, func(tx repository.Tx) error {
		existing, err := s.repo.GetTrack(ctx, tx, id)
		if err != nil {
			return err
		}
		domainID := existing.DomainID
		if w.DomainID != nil {
			domainID = *w.DomainID
		}
		if _, err := s.repo.GetDomain(ctx, tx, domainID); err != nil {
			return err
		}
		if err := s.repo.UpdateTrack(ctx, tx, id, w); err != nil {
			return err
		}
		updated, err = s.repo.GetTrack(ctx, tx, id)
		if err != nil {
			return err
		}
		return s.audit(ctx, tx, a.id, "TRACK_UPDATE", "course_track", id, map[string]any{"code": updated.Code}, requestID)
	})
	if err != nil {
		return Track{}, err
	}
	return updated, nil
}

func (s *Service) ListTracks(ctx context.Context, u httpserver.AuthUser, domainID int64, page, pageSize int) (httpserver.ListData, error) {
	a := fromUser(u)
	if err := a.authorize(); err != nil {
		return httpserver.ListData{}, err
	}
	items, total, err := s.repo.ListTracks(ctx, s.db, domainID, pageSize, (page-1)*pageSize)
	if err != nil {
		return httpserver.ListData{}, err
	}
	return httpserver.NewListData(items, total, httpserver.PageQuery{Page: page, PageSize: pageSize}), nil
}

func validateTrackWrite(w TrackWrite, isCreate bool) error {
	if isCreate {
		if w.DomainID == nil || *w.DomainID <= 0 {
			return ErrInvalidState
		}
		if w.Name == nil || strings.TrimSpace(*w.Name) == "" {
			return ErrInvalidState
		}
		if w.Code == nil || strings.TrimSpace(*w.Code) == "" {
			return ErrInvalidState
		}
	}
	if w.Name != nil && strings.TrimSpace(*w.Name) == "" {
		return ErrInvalidState
	}
	if w.Code != nil && strings.TrimSpace(*w.Code) == "" {
		return ErrInvalidState
	}
	return nil
}

// ---------- Level ----------

func (s *Service) CreateLevel(ctx context.Context, u httpserver.AuthUser, w LevelWrite, requestID string) (Level, error) {
	a := fromUser(u)
	if err := a.authorize(); err != nil {
		return Level{}, err
	}
	if err := validateLevelWrite(w, true); err != nil {
		return Level{}, err
	}
	var created Level
	err := s.inTx(ctx, func(tx repository.Tx) error {
		// Verify full hierarchy: track exists and belongs to a domain.
		track, err := s.repo.GetTrack(ctx, tx, *w.TrackID)
		if err != nil {
			return err
		}
		if _, err := s.repo.GetDomain(ctx, tx, track.DomainID); err != nil {
			return err
		}
		id, err := s.repo.InsertLevel(ctx, tx, w)
		if err != nil {
			return err
		}
		created, err = s.repo.GetLevel(ctx, tx, id)
		if err != nil {
			return err
		}
		return s.audit(ctx, tx, a.id, "LEVEL_CREATE", "course_level", id, map[string]any{"code": created.Code, "trackId": created.TrackID}, requestID)
	})
	if err != nil {
		return Level{}, err
	}
	return created, nil
}

func (s *Service) UpdateLevel(ctx context.Context, u httpserver.AuthUser, id int64, w LevelWrite, requestID string) (Level, error) {
	a := fromUser(u)
	if err := a.authorize(); err != nil {
		return Level{}, err
	}
	if err := validateLevelWrite(w, false); err != nil {
		return Level{}, err
	}
	var updated Level
	err := s.inTx(ctx, func(tx repository.Tx) error {
		existing, err := s.repo.GetLevel(ctx, tx, id)
		if err != nil {
			return err
		}
		trackID := existing.TrackID
		if w.TrackID != nil {
			trackID = *w.TrackID
		}
		// Verify hierarchy: track exists and its domain exists.
		track, err := s.repo.GetTrack(ctx, tx, trackID)
		if err != nil {
			return err
		}
		if _, err := s.repo.GetDomain(ctx, tx, track.DomainID); err != nil {
			return err
		}
		if err := s.repo.UpdateLevel(ctx, tx, id, w); err != nil {
			return err
		}
		updated, err = s.repo.GetLevel(ctx, tx, id)
		if err != nil {
			return err
		}
		return s.audit(ctx, tx, a.id, "LEVEL_UPDATE", "course_level", id, map[string]any{"code": updated.Code}, requestID)
	})
	if err != nil {
		return Level{}, err
	}
	return updated, nil
}

func (s *Service) ListLevels(ctx context.Context, u httpserver.AuthUser, trackID int64, page, pageSize int) (httpserver.ListData, error) {
	a := fromUser(u)
	if err := a.authorize(); err != nil {
		return httpserver.ListData{}, err
	}
	items, total, err := s.repo.ListLevels(ctx, s.db, trackID, pageSize, (page-1)*pageSize)
	if err != nil {
		return httpserver.ListData{}, err
	}
	return httpserver.NewListData(items, total, httpserver.PageQuery{Page: page, PageSize: pageSize}), nil
}

func validateLevelWrite(w LevelWrite, isCreate bool) error {
	if isCreate {
		if w.TrackID == nil || *w.TrackID <= 0 {
			return ErrInvalidState
		}
		if w.Name == nil || strings.TrimSpace(*w.Name) == "" {
			return ErrInvalidState
		}
		if w.Code == nil || strings.TrimSpace(*w.Code) == "" {
			return ErrInvalidState
		}
	}
	if w.Name != nil && strings.TrimSpace(*w.Name) == "" {
		return ErrInvalidState
	}
	if w.Code != nil && strings.TrimSpace(*w.Code) == "" {
		return ErrInvalidState
	}
	if w.MinAge != nil && w.MaxAge != nil && *w.MinAge > *w.MaxAge {
		return ErrInvalidState
	}
	return nil
}

// ---------- Capability Tag ----------

func (s *Service) CreateTag(ctx context.Context, u httpserver.AuthUser, w TagWrite, requestID string) (CapabilityTag, error) {
	a := fromUser(u)
	if err := a.authorize(); err != nil {
		return CapabilityTag{}, err
	}
	if err := validateTagWrite(w, true); err != nil {
		return CapabilityTag{}, err
	}
	var created CapabilityTag
	err := s.inTx(ctx, func(tx repository.Tx) error {
		if _, err := s.repo.GetDomain(ctx, tx, *w.DomainID); err != nil {
			return err
		}
		id, err := s.repo.InsertTag(ctx, tx, w)
		if err != nil {
			return err
		}
		created, err = s.repo.GetTag(ctx, tx, id)
		if err != nil {
			return err
		}
		return s.audit(ctx, tx, a.id, "TAG_CREATE", "skill_tag", id, map[string]any{"code": created.Code, "domainId": created.DomainID}, requestID)
	})
	if err != nil {
		return CapabilityTag{}, err
	}
	return created, nil
}

func (s *Service) UpdateTag(ctx context.Context, u httpserver.AuthUser, id int64, w TagWrite, requestID string) (CapabilityTag, error) {
	a := fromUser(u)
	if err := a.authorize(); err != nil {
		return CapabilityTag{}, err
	}
	if err := validateTagWrite(w, false); err != nil {
		return CapabilityTag{}, err
	}
	var updated CapabilityTag
	err := s.inTx(ctx, func(tx repository.Tx) error {
		existing, err := s.repo.GetTag(ctx, tx, id)
		if err != nil {
			return err
		}
		domainID := existing.DomainID
		if w.DomainID != nil {
			domainID = *w.DomainID
		}
		if _, err := s.repo.GetDomain(ctx, tx, domainID); err != nil {
			return err
		}
		if err := s.repo.UpdateTag(ctx, tx, id, w); err != nil {
			return err
		}
		updated, err = s.repo.GetTag(ctx, tx, id)
		if err != nil {
			return err
		}
		return s.audit(ctx, tx, a.id, "TAG_UPDATE", "skill_tag", id, map[string]any{"code": updated.Code}, requestID)
	})
	if err != nil {
		return CapabilityTag{}, err
	}
	return updated, nil
}

func (s *Service) ListTags(ctx context.Context, u httpserver.AuthUser, domainID int64, page, pageSize int) (httpserver.ListData, error) {
	a := fromUser(u)
	if err := a.authorize(); err != nil {
		return httpserver.ListData{}, err
	}
	items, total, err := s.repo.ListTags(ctx, s.db, domainID, pageSize, (page-1)*pageSize)
	if err != nil {
		return httpserver.ListData{}, err
	}
	return httpserver.NewListData(items, total, httpserver.PageQuery{Page: page, PageSize: pageSize}), nil
}

func validateTagWrite(w TagWrite, isCreate bool) error {
	if isCreate {
		if w.DomainID == nil || *w.DomainID <= 0 {
			return ErrInvalidState
		}
		if w.Name == nil || strings.TrimSpace(*w.Name) == "" {
			return ErrInvalidState
		}
		if w.Code == nil || strings.TrimSpace(*w.Code) == "" {
			return ErrInvalidState
		}
	}
	if w.Name != nil && strings.TrimSpace(*w.Name) == "" {
		return ErrInvalidState
	}
	if w.Code != nil && strings.TrimSpace(*w.Code) == "" {
		return ErrInvalidState
	}
	return nil
}

// ---------- shared helpers ----------

func (s *Service) inTx(ctx context.Context, fn func(repository.Tx) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

// audit writes the operation_log row inside the current transaction.
func (s *Service) audit(ctx context.Context, tx repository.Tx, actorID int64, action, targetType string, targetID int64, detail map[string]any, requestID string) error {
	name, err := repository.ActorName(tx, ctx, actorID)
	if err != nil {
		return err
	}
	return repository.InsertAuditLog(tx, ctx, actorID, name, action, targetType, targetID, detail, requestID)
}

// auditRead is retained for stage C enrollment/assignment paths that may audit
// after a read; stage B update paths audit inside inTx so business + audit share
// one transaction.
var _ = errors.Is
