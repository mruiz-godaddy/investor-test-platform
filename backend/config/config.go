package config

import "sync"

type Config struct {
	mu                      sync.RWMutex
	AutoFinalize            bool // default: true — background goroutine auto-transitions expired listings
	StatusTransitionDelayMs int  // default: 0 — artificial delay between endTime passing and status change
	FinalizerIntervalMs     int  // default: 1000 — how often the background goroutine checks (ms)
}

func New() *Config {
	return &Config{
		AutoFinalize:            true,
		StatusTransitionDelayMs: 0,
		FinalizerIntervalMs:     1000,
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

// Update applies non-nil fields from an update request.
// Returns the current state after applying.
func (c *Config) Update(autoFinalize *bool, delayMs *int, intervalMs *int) {
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
}

// Snapshot returns a copy of the current config for JSON serialization.
func (c *Config) Snapshot() ConfigSnapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return ConfigSnapshot{
		AutoFinalize:            c.AutoFinalize,
		StatusTransitionDelayMs: c.StatusTransitionDelayMs,
		FinalizerIntervalMs:     c.FinalizerIntervalMs,
	}
}

type ConfigSnapshot struct {
	AutoFinalize            bool `json:"autoFinalize"`
	StatusTransitionDelayMs int  `json:"statusTransitionDelayMs"`
	FinalizerIntervalMs     int  `json:"finalizerIntervalMs"`
}
