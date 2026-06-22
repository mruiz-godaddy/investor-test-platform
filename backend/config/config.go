package config

import "sync"

type Config struct {
	mu                      sync.RWMutex
	AutoFinalize            bool // default: true — background goroutine auto-transitions expired listings
	StatusTransitionDelayMs int  // default: 0 — artificial delay between endTime passing and status change
	FinalizerIntervalMs     int  // default: 1000 — how often the background goroutine checks (ms)
	AutoExtWindowSec        int  // default: 60 — last N seconds before end time that triggers extension
	AutoExtSeconds          int  // default: 300 — how many seconds to extend by
	IncludeBin              bool // default: false — include BIN/closeout/OCO listings in app-facing feeds
}

func New() *Config {
	return &Config{
		AutoFinalize:            true,
		StatusTransitionDelayMs: 0,
		FinalizerIntervalMs:     1000,
		AutoExtWindowSec:        60,
		AutoExtSeconds:          300,
		IncludeBin:              false,
	}
}

func (c *Config) GetAutoFinalize() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.AutoFinalize
}

func (c *Config) SetAutoFinalize(v bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.AutoFinalize = v
}

func (c *Config) GetStatusTransitionDelayMs() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.StatusTransitionDelayMs
}

func (c *Config) SetStatusTransitionDelayMs(v int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.StatusTransitionDelayMs = v
}

func (c *Config) GetFinalizerIntervalMs() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.FinalizerIntervalMs
}

func (c *Config) SetFinalizerIntervalMs(v int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.FinalizerIntervalMs = v
}

func (c *Config) GetAutoExtWindowSec() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.AutoExtWindowSec
}

func (c *Config) SetAutoExtWindowSec(v int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.AutoExtWindowSec = v
}

func (c *Config) GetAutoExtSeconds() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.AutoExtSeconds
}

func (c *Config) SetAutoExtSeconds(v int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.AutoExtSeconds = v
}

func (c *Config) GetIncludeBin() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.IncludeBin
}

func (c *Config) SetIncludeBin(v bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.IncludeBin = v
}

// Update applies non-nil fields from an update request.
// Returns the current state after applying.
func (c *Config) Update(autoFinalize *bool, delayMs *int, intervalMs *int, autoExtWindowSec *int, autoExtSeconds *int, includeBin *bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if autoFinalize != nil {
		c.AutoFinalize = *autoFinalize
	}
	if delayMs != nil {
		c.StatusTransitionDelayMs = *delayMs
	}
	if intervalMs != nil {
		c.FinalizerIntervalMs = *intervalMs
	}
	if autoExtWindowSec != nil {
		c.AutoExtWindowSec = *autoExtWindowSec
	}
	if autoExtSeconds != nil {
		c.AutoExtSeconds = *autoExtSeconds
	}
	if includeBin != nil {
		c.IncludeBin = *includeBin
	}
}

// Snapshot returns a copy of the current config for JSON serialization.
func (c *Config) Snapshot() ConfigSnapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return ConfigSnapshot{
		AutoFinalize:            c.AutoFinalize,
		StatusTransitionDelayMs: c.StatusTransitionDelayMs,
		FinalizerIntervalMs:     c.FinalizerIntervalMs,
		AutoExtWindowSec:        c.AutoExtWindowSec,
		AutoExtSeconds:          c.AutoExtSeconds,
		IncludeBin:              c.IncludeBin,
	}
}

type ConfigSnapshot struct {
	AutoFinalize            bool `json:"autoFinalize"`
	StatusTransitionDelayMs int  `json:"statusTransitionDelayMs"`
	FinalizerIntervalMs     int  `json:"finalizerIntervalMs"`
	AutoExtWindowSec        int  `json:"autoExtWindowSec"`
	AutoExtSeconds          int  `json:"autoExtSeconds"`
	IncludeBin              bool `json:"includeBin"`
}
