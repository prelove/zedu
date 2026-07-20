package dashboard

import (
	"testing"
	"time"
)

func TestTokyoDayRangeUsesTokyoCalendarDay(t *testing.T) {
	// 2026-01-01 18:00 UTC is already 2026-01-02 in Tokyo.
	now := time.Date(2026, 1, 1, 18, 0, 0, 0, time.UTC)
	start, end := tokyoDayRange(now)
	if got, want := start.UTC(), time.Date(2026, 1, 1, 15, 0, 0, 0, time.UTC); !got.Equal(want) {
		t.Fatalf("start = %s, want %s", got, want)
	}
	if got, want := end.UTC(), time.Date(2026, 1, 2, 15, 0, 0, 0, time.UTC); !got.Equal(want) {
		t.Fatalf("end = %s, want %s", got, want)
	}
}
