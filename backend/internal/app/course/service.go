package course

import (
	"context"
	"errors"
	"strings"
	"time"

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
	} else {
		// P2-2: an empty PATCH body is not a valid write.
		if w == (DomainWrite{}) {
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
		// P1-3: reparenting a track that is referenced by capabilities or
		// enrollments would orphan those references; reject 42201.
		if w.DomainID != nil && *w.DomainID != existing.DomainID {
			if n, err := s.repo.CountCapabilitiesByTrack(ctx, tx, id); err != nil {
				return err
			} else if n > 0 {
				return ErrInvalidState
			}
			if n, err := s.repo.CountEnrollmentsByTrack(ctx, tx, id); err != nil {
				return err
			} else if n > 0 {
				return ErrInvalidState
			}
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
	} else {
		// P2-2: an empty PATCH body is not a valid write.
		if w == (TrackWrite{}) {
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
		// P1-3: reparenting a level that is referenced by capabilities or
		// enrollments would orphan those references; reject 42201.
		if w.TrackID != nil && *w.TrackID != existing.TrackID {
			if n, err := s.repo.CountCapabilitiesByLevel(ctx, tx, id); err != nil {
				return err
			} else if n > 0 {
				return ErrInvalidState
			}
			if n, err := s.repo.CountEnrollmentsByLevel(ctx, tx, id); err != nil {
				return err
			} else if n > 0 {
				return ErrInvalidState
			}
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
	} else {
		// P2-2: an empty PATCH body is not a valid write.
		if w == (LevelWrite{}) {
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

// inTx runs fn inside a single database transaction. If the business function
// returns a validation/conflict error but Rollback itself fails, the final
// error is ErrDatabase (50002) — not the original error — because the
// transaction state is indeterminate. BeginTx and Commit failures also return
// ErrDatabase.
func (s *Service) inTx(ctx context.Context, fn func(repository.Tx) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return repository.ErrDatabase
	}
	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return repository.ErrDatabase
		}
		return err
	}
	if err := tx.Commit(); err != nil {
		return repository.ErrDatabase
	}
	return nil
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

// ========== Enrollment & Assignment (stage C) ==========

// validEnrollmentTransitions defines the allowed status transitions for an
// enrollment. Terminal states (COMPLETED, CANCELLED) cannot be restored.
var validEnrollmentTransitions = map[string]map[string]bool{
	"ACTIVE":    {"PAUSED": true, "COMPLETED": true, "CANCELLED": true},
	"PAUSED":    {"ACTIVE": true, "CANCELLED": true},
	"COMPLETED": {},
	"CANCELLED": {},
}

var validEnrollmentTypes = map[string]bool{"ONE_TO_ONE": true, "GROUP": true, "TRIAL": true}

// ---------- Enrollment ----------

// CreateEnrollment creates an enrollment for an active student after verifying
// the course hierarchy. M2 creates no lesson, financial, notification or email
// records.
func (s *Service) CreateEnrollment(ctx context.Context, u httpserver.AuthUser, studentID int64, w EnrollmentWrite, requestID string) (Enrollment, error) {
	a := fromUser(u)
	if err := a.authorize(); err != nil {
		return Enrollment{}, err
	}
	if err := validateEnrollmentWrite(w, true); err != nil {
		return Enrollment{}, err
	}
	var created Enrollment
	err := s.inTx(ctx, func(tx repository.Tx) error {
		// Student must exist and be ACTIVE.
		student, err := s.repo.GetStudentActive(ctx, tx, studentID)
		if err != nil {
			return err
		}
		_ = student
		// Verify course hierarchy.
		if err := s.verifyEnrollmentHierarchy(ctx, tx, w); err != nil {
			return err
		}
		id, err := s.repo.InsertEnrollment(ctx, tx, studentID, w)
		if err != nil {
			return err
		}
		created, err = s.repo.GetEnrollment(ctx, tx, id)
		if err != nil {
			return err
		}
		return s.audit(ctx, tx, a.id, "ENROLLMENT_CREATE", "enrollment", id, map[string]any{"studentId": studentID, "trackId": created.TrackID}, requestID)
	})
	if err != nil {
		return Enrollment{}, err
	}
	return created, nil
}

// UpdateEnrollment patches an enrollment. The student must still be ACTIVE
// (P1-2). A currentLevelId change does NOT overwrite the enrollment's
// current_level_id; instead a student_level_event row is written in the same
// transaction (P1-4). Course-selection changes (domain/track/target level) do
// update the enrollment and are audited with before/after snapshots.
func (s *Service) UpdateEnrollment(ctx context.Context, u httpserver.AuthUser, id int64, w EnrollmentWrite, requestID string) (Enrollment, error) {
	a := fromUser(u)
	if err := a.authorize(); err != nil {
		return Enrollment{}, err
	}
	if err := validateEnrollmentWrite(w, false); err != nil {
		return Enrollment{}, err
	}
	// P2-2: an empty PATCH body is not a valid write.
	if w == (EnrollmentWrite{}) {
		return Enrollment{}, ErrInvalidState
	}
	var updated Enrollment
	err := s.inTx(ctx, func(tx repository.Tx) error {
		existing, err := s.repo.GetEnrollment(ctx, tx, id)
		if err != nil {
			return err
		}
		// P1-2: the student must still be ACTIVE to mutate the enrollment.
		if _, err := s.repo.GetStudentActive(ctx, tx, existing.StudentID); err != nil {
			return err
		}
		if w.Status != nil && *w.Status != existing.Status {
			allowed, ok := validEnrollmentTransitions[existing.Status]
			if !ok || !allowed[*w.Status] {
				return ErrInvalidState
			}
		}
		// The level history is authoritative after the first transition. A level
		// event does not overwrite the enrollment snapshot by design.
		effectiveLevel, err := s.repo.EffectiveEnrollmentLevel(ctx, tx, id, existing.CurrentLevelID)
		if err != nil {
			return err
		}
		// P1-4: a currentLevelId change writes a level event but does NOT
		// overwrite the enrollment's current_level_id.
		var levelEventFrom, levelEventTo *int64
		if w.CurrentLevelID != nil {
			if (w.DomainID != nil && *w.DomainID != existing.DomainID) || (w.TrackID != nil && *w.TrackID != existing.TrackID) {
				// Mixing a transition event with a course-selection move would retain
				// the old snapshot level under a new track. Keep each write atomic and
				// unambiguous instead of creating an internally inconsistent enrollment.
				return ErrInvalidState
			}
			if effectiveLevel != nil && *effectiveLevel == *w.CurrentLevelID {
				return ErrInvalidState
			}
			levelEventTo = w.CurrentLevelID
			levelEventFrom = effectiveLevel
			// Verify the new level belongs to the (possibly new) track.
			hw := EnrollmentWrite{
				DomainID:       w.DomainID,
				TrackID:        w.TrackID,
				CurrentLevelID: w.CurrentLevelID,
				TargetLevelID:  w.TargetLevelID,
			}
			if hw.DomainID == nil {
				hw.DomainID = &existing.DomainID
			}
			if hw.TrackID == nil {
				hw.TrackID = &existing.TrackID
			}
			if hw.TargetLevelID == nil {
				hw.TargetLevelID = existing.TargetLevelID
			}
			if err := s.verifyEnrollmentHierarchy(ctx, tx, hw); err != nil {
				return err
			}
		} else if w.DomainID != nil || w.TrackID != nil || w.TargetLevelID != nil {
			// If hierarchy fields change, re-verify.
			hw := EnrollmentWrite{
				DomainID:       w.DomainID,
				TrackID:        w.TrackID,
				CurrentLevelID: w.CurrentLevelID,
				TargetLevelID:  w.TargetLevelID,
			}
			if hw.DomainID == nil {
				hw.DomainID = &existing.DomainID
			}
			if hw.TrackID == nil {
				hw.TrackID = &existing.TrackID
			}
			if hw.CurrentLevelID == nil {
				hw.CurrentLevelID = existing.CurrentLevelID
			}
			if hw.TargetLevelID == nil {
				hw.TargetLevelID = existing.TargetLevelID
			}
			if err := s.verifyEnrollmentHierarchy(ctx, tx, hw); err != nil {
				return err
			}
		}
		// Build the persisted patch: strip currentLevelId so the enrollment's
		// current_level_id is NOT overwritten by a level-change request.
		persist := w
		persist.CurrentLevelID = nil
		if err := s.repo.UpdateEnrollment(ctx, tx, id, persist); err != nil {
			return err
		}
		// P1-4: write the level event inside the same transaction.
		if levelEventTo != nil {
			if err := s.repo.InsertLevelEvent(ctx, tx, existing.StudentID, id, levelEventFrom, levelEventTo, "MANUAL", time.Now().UTC().Format("2006-01-02"), a.id); err != nil {
				return err
			}
		}
		updated, err = s.repo.GetEnrollment(ctx, tx, id)
		if err != nil {
			return err
		}
		// P1-4: audit with before/after snapshots for course-selection changes.
		detail := map[string]any{}
		if levelEventTo != nil {
			detail["levelEvent"] = map[string]any{"from": ptrInt64(levelEventFrom), "to": *levelEventTo}
		}
		if w.DomainID != nil || w.TrackID != nil || w.TargetLevelID != nil || w.Status != nil || w.Note != nil {
			detail["before"] = map[string]any{
				"domainId":      existing.DomainID,
				"trackId":       existing.TrackID,
				"targetLevelId": existing.TargetLevelID,
				"status":        existing.Status,
			}
			detail["after"] = map[string]any{
				"domainId":      updated.DomainID,
				"trackId":       updated.TrackID,
				"targetLevelId": updated.TargetLevelID,
				"status":        updated.Status,
			}
		}
		return s.audit(ctx, tx, a.id, "ENROLLMENT_UPDATE", "enrollment", id, detail, requestID)
	})
	if err != nil {
		return Enrollment{}, err
	}
	return updated, nil
}

// GetEnrollment returns an enrollment by id.
func (s *Service) GetEnrollment(ctx context.Context, u httpserver.AuthUser, id int64) (Enrollment, error) {
	a := fromUser(u)
	if err := a.authorize(); err != nil {
		return Enrollment{}, err
	}
	return s.repo.GetEnrollment(ctx, s.db, id)
}

// ListEnrollments returns a page of enrollments for a student.
func (s *Service) ListEnrollments(ctx context.Context, u httpserver.AuthUser, studentID int64, page, pageSize int) (httpserver.ListData, error) {
	a := fromUser(u)
	if err := a.authorize(); err != nil {
		return httpserver.ListData{}, err
	}
	items, total, err := s.repo.ListEnrollments(ctx, s.db, studentID, pageSize, (page-1)*pageSize)
	if err != nil {
		return httpserver.ListData{}, err
	}
	return httpserver.NewListData(items, total, httpserver.PageQuery{Page: page, PageSize: pageSize}), nil
}

func validateEnrollmentWrite(w EnrollmentWrite, isCreate bool) error {
	if isCreate {
		if w.DomainID == nil || w.TrackID == nil {
			return ErrInvalidState
		}
	}
	// P1-5: reject non-positive level IDs at the service layer so they never
	// reach the database (which would surface as 50002).
	if w.CurrentLevelID != nil && *w.CurrentLevelID <= 0 {
		return ErrInvalidState
	}
	if w.TargetLevelID != nil && *w.TargetLevelID <= 0 {
		return ErrInvalidState
	}
	if w.DomainID != nil && *w.DomainID <= 0 {
		return ErrInvalidState
	}
	if w.TrackID != nil && *w.TrackID <= 0 {
		return ErrInvalidState
	}
	if w.EnrollmentType != nil && *w.EnrollmentType != "" && !validEnrollmentTypes[*w.EnrollmentType] {
		return ErrInvalidState
	}
	if w.Status != nil && *w.Status != "" {
		if !enrollmentStatusExists(*w.Status) {
			return ErrInvalidState
		}
	}
	return nil
}

func enrollmentStatusExists(s string) bool {
	_, ok := validEnrollmentTransitions[s]
	return ok
}

// verifyEnrollmentHierarchy confirms domain exists, track belongs to domain,
// and current/target levels (if provided) belong to the track.
func (s *Service) verifyEnrollmentHierarchy(ctx context.Context, tx repository.Tx, w EnrollmentWrite) error {
	domainID := ptrInt64(w.DomainID)
	trackID := ptrInt64(w.TrackID)
	if domainID <= 0 || trackID <= 0 {
		return ErrInvalidState
	}
	if _, err := s.repo.GetDomain(ctx, tx, domainID); err != nil {
		return err
	}
	track, err := s.repo.GetTrack(ctx, tx, trackID)
	if err != nil {
		return err
	}
	if track.DomainID != domainID {
		return ErrInvalidState
	}
	for _, lvl := range []*int64{w.CurrentLevelID, w.TargetLevelID} {
		if lvl != nil && *lvl > 0 {
			level, err := s.repo.GetLevel(ctx, tx, *lvl)
			if err != nil {
				return err
			}
			if level.TrackID != trackID {
				return ErrInvalidState
			}
		}
	}
	return nil
}

// ---------- Assignment ----------

// CreateAssignment creates an ACTIVE assignment for an enrollment. If an ACTIVE
// assignment already exists, it is ended and the new one created in a single
// transaction (replacement). The partial unique index guarantees at most one
// ACTIVE assignment; the end-before-insert within one tx means the insert never
// races a leftover ACTIVE. M2 creates no lesson, attendance, payment,
// notification, payout or email records.
// validAssignmentRoleTypes mirrors migration 003 CHECK constraint.
var validAssignmentRoleTypes = map[string]bool{"MAIN": true, "SUBSTITUTE": true, "ASSISTANT": true}

func (s *Service) CreateAssignment(ctx context.Context, u httpserver.AuthUser, enrollmentID int64, w AssignmentWrite, requestID string) (Assignment, error) {
	a := fromUser(u)
	if err := a.authorize(); err != nil {
		return Assignment{}, err
	}
	if w.TeacherID == nil || *w.TeacherID <= 0 {
		return Assignment{}, ErrInvalidState
	}
	// P1-5: reject non-enum roleType at the service layer so it never reaches
	// the database (which would surface as 50002).
	if w.RoleType != nil && *w.RoleType != "" && !validAssignmentRoleTypes[*w.RoleType] {
		return Assignment{}, ErrInvalidState
	}
	var created Assignment
	err := s.inTx(ctx, func(tx repository.Tx) error {
		enrollment, err := s.repo.GetEnrollment(ctx, tx, enrollmentID)
		if err != nil {
			return err
		}
		if enrollment.Status != "ACTIVE" {
			return ErrInvalidState
		}
		teacher, err := s.repo.GetTeacherActive(ctx, tx, *w.TeacherID)
		if err != nil {
			return err
		}
		_ = teacher
		// End current ACTIVE assignment if any (replacement), with reason.
		reason := ptrString(w.Reason)
		if reason == "" {
			reason = "REPLACE"
		}
		endedID, ended, err := s.repo.EndActiveAssignment(ctx, tx, enrollmentID, reason)
		if err != nil {
			return err
		}
		if ended {
			if err := s.audit(ctx, tx, a.id, "ASSIGNMENT_END", "assignment", endedID, map[string]any{"enrollmentId": enrollmentID, "reason": reason}, requestID); err != nil {
				return err
			}
		}
		id, err := s.repo.InsertAssignment(ctx, tx, enrollmentID, enrollment.StudentID, *w.TeacherID, w)
		if err != nil {
			return err
		}
		created, err = s.repo.GetAssignment(ctx, tx, id)
		if err != nil {
			return err
		}
		return s.audit(ctx, tx, a.id, "ASSIGNMENT_CREATE", "assignment", id, map[string]any{"enrollmentId": enrollmentID, "teacherId": created.TeacherID}, requestID)
	})
	if err != nil {
		return Assignment{}, err
	}
	return created, nil
}

// EndAssignment ends the current ACTIVE assignment by id. ENDED assignments
// cannot be restored; ending a non-ACTIVE assignment returns 42201. A concurrent
// end surfaces as ended=false (no half-written record).
func (s *Service) EndAssignment(ctx context.Context, u httpserver.AuthUser, id int64, w EndAssignmentWrite, requestID string) (Assignment, error) {
	a := fromUser(u)
	if err := a.authorize(); err != nil {
		return Assignment{}, err
	}
	var ended Assignment
	err := s.inTx(ctx, func(tx repository.Tx) error {
		existing, err := s.repo.GetAssignment(ctx, tx, id)
		if err != nil {
			return err
		}
		if existing.Status != "ACTIVE" {
			return ErrInvalidState
		}
		ended2, err := s.repo.EndAssignmentByID(ctx, tx, id, ptrString(w.Reason))
		if err != nil {
			return err
		}
		if !ended2 {
			// Concurrent end: no row was ACTIVE anymore.
			return ErrConflict
		}
		ended, err = s.repo.GetAssignment(ctx, tx, id)
		if err != nil {
			return err
		}
		return s.audit(ctx, tx, a.id, "ASSIGNMENT_END", "assignment", id, map[string]any{"enrollmentId": existing.EnrollmentID, "reason": ptrString(w.Reason)}, requestID)
	})
	if err != nil {
		return Assignment{}, err
	}
	return ended, nil
}

// ListAssignments returns a page of assignments for an enrollment.
func (s *Service) ListAssignments(ctx context.Context, u httpserver.AuthUser, enrollmentID int64, page, pageSize int) (httpserver.ListData, error) {
	a := fromUser(u)
	if err := a.authorize(); err != nil {
		return httpserver.ListData{}, err
	}
	if _, err := s.repo.GetEnrollment(ctx, s.db, enrollmentID); err != nil {
		return httpserver.ListData{}, err
	}
	items, total, err := s.repo.ListAssignments(ctx, s.db, enrollmentID, pageSize, (page-1)*pageSize)
	if err != nil {
		return httpserver.ListData{}, err
	}
	return httpserver.NewListData(items, total, httpserver.PageQuery{Page: page, PageSize: pageSize}), nil
}

// ---------- helpers ----------
