package onboarding

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/prelove/zedu/backend/internal/platform/httpserver"
)

const (
	settingTemplate      = "onboarding.template"
	settingInitializedAt = "onboarding.initialized_at"
	// onboardingAuditTargetID identifies the singleton system onboarding target.
	// It is stable because onboarding config is global to this single-instance MVP.
	onboardingAuditTargetID = int64(1)
)

// Handler provides Owner-only, explicit business-template initialization.
type Handler struct {
	db     *sql.DB
	logger *slog.Logger
}

// NewHandler creates an onboarding handler backed by the application database.
func NewHandler(db *sql.DB, logger *slog.Logger) *Handler {
	return &Handler{db: db, logger: logger}
}

// MountRoutes mounts Owner-only onboarding endpoints on an existing mux.
func MountRoutes(mux *http.ServeMux, h *Handler, db *sql.DB, jwtSecret string) {
	authMW := httpserver.AuthMiddleware(jwtSecret, db)
	mux.Handle("POST /onboarding/initialize", authMW(httpserver.RequireRole("OWNER", http.HandlerFunc(h.Initialize))))
	mux.Handle("POST /onboarding/reset", authMW(httpserver.RequireRole("OWNER", http.HandlerFunc(h.Reset))))
}

type templateRequest struct {
	Template string `json:"template"`
}

type templateResponse struct {
	Template string `json:"template"`
	Reused   bool   `json:"reused"`
}

// Initialize applies a selected business template exactly once. A repeated
// request returns the existing result without adding records or a new audit fact.
func (h *Handler) Initialize(w http.ResponseWriter, r *http.Request) {
	template, ok := h.decodeTemplate(w, r)
	if !ok {
		return
	}
	user, _ := httpserver.UserFromContext(r.Context())
	response, err := h.initialize(r.Context(), user, template, false)
	if err != nil {
		h.writeDatabaseError(w, r, "initialize onboarding", err)
		return
	}
	httpserver.WriteSuccess(w, http.StatusOK, response)
}

// Reset replaces template data only when no protected business record exists.
func (h *Handler) Reset(w http.ResponseWriter, r *http.Request) {
	template, ok := h.decodeTemplate(w, r)
	if !ok {
		return
	}
	user, _ := httpserver.UserFromContext(r.Context())
	response, err := h.initialize(r.Context(), user, template, true)
	if err != nil {
		if err == errProtectedDataExists {
			httpserver.WriteErrorFromContext(w, r, http.StatusUnprocessableEntity, httpserver.CodeInvalidState, "RESET_NOT_ALLOWED")
			return
		}
		h.writeDatabaseError(w, r, "reset onboarding", err)
		return
	}
	httpserver.WriteSuccess(w, http.StatusOK, response)
}

func (h *Handler) decodeTemplate(w http.ResponseWriter, r *http.Request) (string, bool) {
	var req templateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || !validTemplate(req.Template) {
		httpserver.WriteErrorFromContext(w, r, http.StatusUnprocessableEntity, httpserver.CodeInvalidState, "INVALID_TEMPLATE")
		return "", false
	}
	return req.Template, true
}

func validTemplate(template string) bool {
	return template == "japanese" || template == "k12" || template == "blank"
}

var errProtectedDataExists = fmt.Errorf("protected business data exists")

func (h *Handler) initialize(ctx context.Context, user httpserver.AuthUser, template string, reset bool) (templateResponse, error) {
	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		return templateResponse{}, err
	}
	defer tx.Rollback()

	var existing string
	err = tx.QueryRowContext(ctx, `SELECT config_value FROM system_settings WHERE config_key = ?`, settingTemplate).Scan(&existing)
	if err == nil && !reset {
		if err := tx.Commit(); err != nil {
			return templateResponse{}, err
		}
		return templateResponse{Template: existing, Reused: true}, nil
	}
	if err != nil && err != sql.ErrNoRows {
		return templateResponse{}, err
	}

	if reset {
		protected, err := hasProtectedBusinessData(ctx, tx)
		if err != nil {
			return templateResponse{}, err
		}
		if protected {
			return templateResponse{}, errProtectedDataExists
		}
		if err := clearTemplateData(ctx, tx); err != nil {
			return templateResponse{}, err
		}
	}

	if err := insertTemplate(ctx, tx, template); err != nil {
		return templateResponse{}, err
	}
	if err := writeSettings(ctx, tx, template, user.UserID); err != nil {
		return templateResponse{}, err
	}
	if err := writeAudit(ctx, tx, user.UserID, template, reset); err != nil {
		return templateResponse{}, err
	}
	if err := tx.Commit(); err != nil {
		return templateResponse{}, err
	}
	return templateResponse{Template: template}, nil
}

func hasProtectedBusinessData(ctx context.Context, tx *sql.Tx) (bool, error) {
	for _, table := range []string{"student", "teacher", "student_course_enrollment", "student_teacher_assignment"} {
		var count int
		if err := tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+table).Scan(&count); err != nil {
			return false, err
		}
		if count > 0 {
			return true, nil
		}
	}
	return false, nil
}

func clearTemplateData(ctx context.Context, tx *sql.Tx) error {
	for _, table := range []string{"course_level", "skill_tag", "course_track", "course_domain"} {
		if _, err := tx.ExecContext(ctx, "DELETE FROM "+table); err != nil {
			return err
		}
	}
	_, err := tx.ExecContext(ctx, `DELETE FROM system_settings WHERE config_key IN (?, ?)`, settingTemplate, settingInitializedAt)
	return err
}

func writeSettings(ctx context.Context, tx *sql.Tx, template string, userID int64) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	for _, setting := range []struct {
		key, value, description string
	}{
		{settingTemplate, template, "selected business template"},
		{settingInitializedAt, now, "business template initialization timestamp"},
	} {
		if _, err := tx.ExecContext(ctx, `INSERT INTO system_settings (config_key, config_value, description, updated_by)
			VALUES (?, ?, ?, ?)
			ON CONFLICT(config_key) DO UPDATE SET config_value = excluded.config_value, description = excluded.description, updated_by = excluded.updated_by, updated_at = CURRENT_TIMESTAMP`,
			setting.key, setting.value, setting.description, userID); err != nil {
			return err
		}
	}
	return nil
}

func writeAudit(ctx context.Context, tx *sql.Tx, actorID int64, template string, reset bool) error {
	var actorName string
	if err := tx.QueryRowContext(ctx, `SELECT username FROM user_account WHERE id = ?`, actorID).Scan(&actorName); err != nil {
		return err
	}
	detail, err := json.Marshal(map[string]string{"template": template})
	if err != nil {
		return err
	}
	action := "ONBOARDING_INITIALIZE"
	if reset {
		action = "ONBOARDING_RESET"
	}
	requestID := httpserver.RequestIDFromContext(ctx)
	if requestID == "" {
		requestID = "unknown"
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO operation_log (operator_id, operator_name, action, target_type, target_id, detail_json, request_id)
		VALUES (?, ?, ?, ?, ?, ?, ?)`, actorID, actorName, action, "system", onboardingAuditTargetID, string(detail), requestID)
	return err
}

func insertTemplate(ctx context.Context, tx *sql.Tx, template string) error {
	switch template {
	case "blank":
		return nil
	case "japanese":
		return insertJapaneseTemplate(ctx, tx)
	case "k12":
		return insertK12Template(ctx, tx)
	default:
		return fmt.Errorf("invalid template %q", template)
	}
}

func insertJapaneseTemplate(ctx context.Context, tx *sql.Tx) error {
	domainID, err := insertDomain(ctx, tx, "日语", "JAPANESE", "LANGUAGE", 1)
	if err != nil {
		return err
	}
	jlptID, err := insertTrack(ctx, tx, domainID, "JLPT备考", "JLPT", 1)
	if err != nil {
		return err
	}
	if err := insertLevels(ctx, tx, jlptID, []templateItem{{"入门", "BEGINNER"}, {"N5", "N5"}, {"N4", "N4"}, {"N3", "N3"}, {"N2", "N2"}, {"N1", "N1"}}); err != nil {
		return err
	}
	conversationID, err := insertTrack(ctx, tx, domainID, "日常会话", "CONVERSATION", 2)
	if err != nil {
		return err
	}
	if err := insertLevels(ctx, tx, conversationID, []templateItem{{"初级", "BEGINNER"}, {"中级", "INTERMEDIATE"}, {"高级", "ADVANCED"}}); err != nil {
		return err
	}
	if _, err := insertTrack(ctx, tx, domainID, "商务日语", "BUSINESS", 3); err != nil {
		return err
	}
	if _, err := insertTrack(ctx, tx, domainID, "少儿日语", "CHILDREN", 4); err != nil {
		return err
	}
	return insertTags(ctx, tx, domainID, []templateItem{{"词汇", "VOCABULARY"}, {"语法", "GRAMMAR"}, {"阅读", "READING"}, {"听力", "LISTENING"}, {"口语", "SPEAKING"}, {"写作", "WRITING"}, {"综合", "COMPREHENSIVE"}, {"面试技巧", "INTERVIEW"}, {"商务敬语", "BUSINESS_KEIGO"}})
}

func insertK12Template(ctx context.Context, tx *sql.Tx) error {
	domains := []struct {
		name, code string
		levels     []templateItem
	}{
		{"小学数学", "PRIMARY_MATH", []templateItem{{"小一", "G1"}, {"小二", "G2"}, {"小三", "G3"}, {"小四", "G4"}, {"小五", "G5"}, {"小六", "G6"}}},
		{"初中数学", "JUNIOR_MATH", []templateItem{{"初一", "G7"}, {"初二", "G8"}, {"初三", "G9"}}},
		{"初中物理", "JUNIOR_PHYSICS", []templateItem{{"初一", "G7"}, {"初二", "G8"}, {"初三", "G9"}}},
		{"初中化学", "JUNIOR_CHEMISTRY", []templateItem{{"初一", "G7"}, {"初二", "G8"}, {"初三", "G9"}}},
	}
	for position, domain := range domains {
		domainID, err := insertDomain(ctx, tx, domain.name, domain.code, "K12", position+1)
		if err != nil {
			return err
		}
		trackID, err := insertTrack(ctx, tx, domainID, "同步辅导", "SUPPORT", 1)
		if err != nil {
			return err
		}
		if err := insertLevels(ctx, tx, trackID, domain.levels); err != nil {
			return err
		}
		for index, item := range []templateItem{{"期末冲刺", "FINAL_REVIEW"}, {"中考冲刺", "EXAM_REVIEW"}, {"专题强化", "TOPIC"}} {
			if _, err := insertTrack(ctx, tx, domainID, item.name, item.code, index+2); err != nil {
				return err
			}
		}
		if err := insertTags(ctx, tx, domainID, []templateItem{{"基础概念", "BASICS"}, {"计算", "CALCULATION"}, {"应用题", "APPLICATION"}, {"函数", "FUNCTION"}, {"几何", "GEOMETRY"}, {"力学", "MECHANICS"}, {"电学", "ELECTRICITY"}, {"实验", "EXPERIMENT"}, {"错题整理", "ERROR_REVIEW"}}); err != nil {
			return err
		}
	}
	return nil
}

type templateItem struct{ name, code string }

func insertDomain(ctx context.Context, tx *sql.Tx, name, code, kind string, order int) (int64, error) {
	result, err := tx.ExecContext(ctx, `INSERT INTO course_domain (name, code, type, sort_order) VALUES (?, ?, ?, ?)`, name, code, kind, order)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func insertTrack(ctx context.Context, tx *sql.Tx, domainID int64, name, code string, order int) (int64, error) {
	result, err := tx.ExecContext(ctx, `INSERT INTO course_track (domain_id, name, code, sort_order) VALUES (?, ?, ?, ?)`, domainID, name, code, order)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func insertLevels(ctx context.Context, tx *sql.Tx, trackID int64, levels []templateItem) error {
	for index, level := range levels {
		if _, err := tx.ExecContext(ctx, `INSERT INTO course_level (track_id, name, code, sort_order) VALUES (?, ?, ?, ?)`, trackID, level.name, level.code, index+1); err != nil {
			return err
		}
	}
	return nil
}

func insertTags(ctx context.Context, tx *sql.Tx, domainID int64, tags []templateItem) error {
	for index, tag := range tags {
		if _, err := tx.ExecContext(ctx, `INSERT INTO skill_tag (domain_id, name, code, sort_order) VALUES (?, ?, ?, ?)`, domainID, tag.name, tag.code, index+1); err != nil {
			return err
		}
	}
	return nil
}

func (h *Handler) writeDatabaseError(w http.ResponseWriter, r *http.Request, message string, err error) {
	h.logger.Error(message, slog.String("request_id", httpserver.RequestIDFromContext(r.Context())), slog.Any("error", err))
	httpserver.WriteErrorFromContext(w, r, http.StatusInternalServerError, httpserver.CodeDatabase, "DATABASE_ERROR")
}
