package lifecycle

import (
	"sync"
	"time"
)

type ClockMode int

const (
	Realtime ClockMode = iota
	Offset
	Frozen
)

var (
	mu         sync.RWMutex
	mode       ClockMode = Realtime
	offset     time.Duration
	frozenTime time.Time
)

// Now returns the current server time, respecting offset/freeze settings.
func Now() time.Time {
	mu.RLock()
	defer mu.RUnlock()

	switch mode {
	case Offset:
		return time.Now().UTC().Add(offset)
	case Frozen:
		return frozenTime
	default:
		return time.Now().UTC()
	}
}

// SetOffset sets a time offset from real time.
func SetOffset(seconds int) {
	mu.Lock()
	defer mu.Unlock()
	mode = Offset
	offset = time.Duration(seconds) * time.Second
}

// Freeze freezes the clock at the given instant.
func Freeze(t time.Time) {
	mu.Lock()
	defer mu.Unlock()
	mode = Frozen
	frozenTime = t
}

// Reset returns to real-time mode.
func Reset() {
	mu.Lock()
	defer mu.Unlock()
	mode = Realtime
	offset = 0
}

// Mode returns the current clock mode as a string.
func Mode() string {
	mu.RLock()
	defer mu.RUnlock()
	switch mode {
	case Offset:
		return "offset"
	case Frozen:
		return "frozen"
	default:
		return "realtime"
	}
}
