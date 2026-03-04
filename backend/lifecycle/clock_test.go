package lifecycle

import (
	"testing"
	"time"
)

func TestNowRealtime(t *testing.T) {
	Reset()
	now := Now()
	realNow := time.Now().UTC()
	diff := realNow.Sub(now)
	if diff < 0 {
		diff = -diff
	}
	if diff > time.Second {
		t.Errorf("realtime Now() off by %v", diff)
	}
}

func TestSetOffset(t *testing.T) {
	defer Reset()
	SetOffset(3600) // +1 hour
	now := Now()
	realNow := time.Now().UTC()
	diff := now.Sub(realNow)
	if diff < 3599*time.Second || diff > 3601*time.Second {
		t.Errorf("offset Now() expected ~3600s ahead, got diff=%v", diff)
	}
	if Mode() != "offset" {
		t.Errorf("expected mode 'offset', got %q", Mode())
	}
}

func TestFreeze(t *testing.T) {
	defer Reset()
	frozen := time.Date(2025, 6, 8, 12, 0, 0, 0, time.UTC)
	Freeze(frozen)
	now := Now()
	if !now.Equal(frozen) {
		t.Errorf("frozen Now() = %v, want %v", now, frozen)
	}
	if Mode() != "frozen" {
		t.Errorf("expected mode 'frozen', got %q", Mode())
	}
}

func TestReset(t *testing.T) {
	SetOffset(9999)
	Reset()
	if Mode() != "realtime" {
		t.Errorf("expected mode 'realtime' after reset, got %q", Mode())
	}
}
