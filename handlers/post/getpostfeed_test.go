package post

import (
	"testing"
	"time"
)

func TestDefaultPublishedAt(t *testing.T) {
	now := time.Date(2026, 10, 19, 9, 0, 0, 0, time.UTC)
	taken := time.Date(2026, 10, 12, 4, 45, 15, 0, time.UTC)
	if got := defaultPublishedAt(taken, now); !got.Equal(taken) {
		t.Errorf("taken_at set: got %v, want %v", got, taken)
	}
	if got := defaultPublishedAt(time.Time{}, now); !got.Equal(now) {
		t.Errorf("taken_at missing: got %v, want %v", got, now)
	}
}
