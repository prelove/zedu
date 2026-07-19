package lesson

import (
	"database/sql"
	"strings"
	"time"
)

const lessonSelect = `SELECT id,lesson_no,enrollment_id,assignment_id,teacher_id,student_id,scheduled_start_at,scheduled_end_at,duration_min,timezone,meeting_type,meeting_link,lesson_topic,note,status,cancel_reason FROM lesson`

type lessonScanner interface{ Scan(...any) error }

func scanLesson(row lessonScanner) (Lesson, error) {
	var lesson Lesson
	var start, end time.Time
	var link, topic, note, reason sql.NullString
	err := row.Scan(&lesson.ID, &lesson.LessonNo, &lesson.EnrollmentID, &lesson.AssignmentID, &lesson.TeacherID, &lesson.StudentID, &start, &end, &lesson.DurationMin, &lesson.Timezone, &lesson.MeetingType, &link, &topic, &note, &lesson.Status, &reason)
	if err != nil {
		return Lesson{}, err
	}
	lesson.ScheduledStartAt = start.UTC().Format(time.RFC3339)
	lesson.ScheduledEndAt = end.UTC().Format(time.RFC3339)
	lesson.MeetingLink = nullString(link)
	lesson.LessonTopic = nullString(topic)
	lesson.Note = nullString(note)
	lesson.CancelReason = nullString(reason)
	return lesson, nil
}

func nullString(value sql.NullString) string {
	if value.Valid {
		return value.String
	}
	return ""
}

func normalizePage(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

func lessonWhere(filter ListFilter) (string, []any, error) {
	clauses := make([]string, 0, 5)
	args := make([]any, 0, 5)
	if filter.StudentID > 0 {
		clauses, args = append(clauses, "student_id=?"), append(args, filter.StudentID)
	}
	if filter.TeacherID > 0 {
		clauses, args = append(clauses, "teacher_id=?"), append(args, filter.TeacherID)
	}
	if filter.Status != "" {
		if filter.Status != "SCHEDULED" && filter.Status != "COMPLETED" && filter.Status != "CANCELLED" {
			return "", nil, ErrInvalidState
		}
		clauses, args = append(clauses, "status=?"), append(args, filter.Status)
	}
	for _, bound := range []struct{ value, op string }{{filter.From, ">="}, {filter.To, "<="}} {
		if bound.value == "" {
			continue
		}
		parsed, err := time.Parse(time.RFC3339, bound.value)
		if err != nil {
			return "", nil, err
		}
		clauses, args = append(clauses, "scheduled_start_at"+bound.op+"?"), append(args, parsed.UTC())
	}
	if len(clauses) == 0 {
		return "", args, nil
	}
	return " WHERE " + strings.Join(clauses, " AND "), args, nil
}
