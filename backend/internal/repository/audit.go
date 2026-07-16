package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
)

// InsertAuditLog inserts an operation_log row using the given executor (either
// *sql.Tx within a service transaction or *sql.DB for non-transactional reads).
// detail must not contain password, hash, token or Authorization values; it is
// marshaled to JSON when a non-string value is supplied.
//
// The audit row is written in the same transaction as the business write so
// that any failure rolls back both the business data and the audit fact.
func InsertAuditLog(exec Executor, ctx context.Context, actorID int64, actorName, action, targetType string, targetID int64, detail any, requestID string) error {
	detailJSON, err := marshalDetail(detail)
	if err != nil {
		return err
	}
	if requestID == "" {
		requestID = "unknown"
	}
	_, err = exec.ExecContext(ctx,
		`INSERT INTO operation_log (operator_id, operator_name, action, target_type, target_id, detail_json, request_id)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		actorID, actorName, action, targetType, targetID, detailJSON, requestID,
	)
	return err
}

func marshalDetail(detail any) (string, error) {
	switch d := detail.(type) {
	case nil:
		return "{}", nil
	case string:
		if d == "" {
			return "{}", nil
		}
		return d, nil
	default:
		b, err := json.Marshal(d)
		if err != nil {
			return "", fmt.Errorf("marshal audit detail: %w", err)
		}
		return string(b), nil
	}
}

// ActorName loads the username for the given actor id from the executor. Used
// so the audit row records a stable operator_name even if the account is later
// renamed or disabled.
func ActorName(exec Executor, ctx context.Context, actorID int64) (string, error) {
	var name sql.NullString
	err := exec.QueryRowContext(ctx, `SELECT username FROM user_account WHERE id = ?`, actorID).Scan(&name)
	if err != nil {
		return "", err
	}
	if !name.Valid || name.String == "" {
		return "unknown", nil
	}
	return name.String, nil
}
