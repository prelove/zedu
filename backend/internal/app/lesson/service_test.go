package lesson

import "testing"

func TestValidateWriteRejectsInvalidDuration(t *testing.T) {
	if err := validateWrite(Write{DurationMin: 9, Timezone: "Asia/Tokyo", MeetingType: "OFFLINE"}); err == nil {
		t.Fatal("duration below M4a lower bound must be rejected")
	}
}

func TestValidateWriteAcceptsOfflineLessonWithoutLink(t *testing.T) {
	if err := validateWrite(Write{StartAt: "2026-08-01T19:00:00", DurationMin: 60, Timezone: "Asia/Tokyo", MeetingType: "OFFLINE"}); err != nil {
		t.Fatalf("valid offline lesson: %v", err)
	}
}

func TestParseBusinessTimeNormalizesTokyoLocalTimeToUTC(t *testing.T) {
	got, err := parseBusinessTime("2026-08-01T19:00:00", "Asia/Tokyo")
	if err != nil {
		t.Fatalf("parse local lesson time: %v", err)
	}
	if want := "2026-08-01T10:00:00Z"; got.Format("2006-01-02T15:04:05Z") != want {
		t.Fatalf("UTC time = %s, want %s", got.Format("2006-01-02T15:04:05Z"), want)
	}
}

func TestValidateWriteRejectsInvalidWeChatURL(t *testing.T) {
	err := validateWrite(Write{StartAt: "2026-08-01T19:00:00", DurationMin: 60, Timezone: "Asia/Tokyo", MeetingType: "WECHAT", MeetingLink: "not-a-url"})
	if err == nil {
		t.Fatal("invalid WeChat meeting link must be rejected")
	}
}
