package config

import "testing"

func TestDefaults(t *testing.T) {
	c := New()
	if !c.GetAutoFinalize() {
		t.Error("expected autoFinalize=true")
	}
	if c.GetStatusTransitionDelayMs() != 0 {
		t.Error("expected delay=0")
	}
	if c.GetFinalizerIntervalMs() != 1000 {
		t.Error("expected interval=1000")
	}
}

func TestUpdate(t *testing.T) {
	c := New()
	f := false
	d := 5000
	c.Update(&f, &d, nil, nil, nil)

	if c.GetAutoFinalize() {
		t.Error("expected autoFinalize=false after update")
	}
	if c.GetStatusTransitionDelayMs() != 5000 {
		t.Error("expected delay=5000 after update")
	}
	// interval unchanged
	if c.GetFinalizerIntervalMs() != 1000 {
		t.Error("expected interval unchanged at 1000")
	}
}

func TestSnapshot(t *testing.T) {
	c := New()
	snap := c.Snapshot()
	if snap.AutoFinalize != true {
		t.Error("snapshot autoFinalize mismatch")
	}
	if snap.FinalizerIntervalMs != 1000 {
		t.Error("snapshot interval mismatch")
	}
}
