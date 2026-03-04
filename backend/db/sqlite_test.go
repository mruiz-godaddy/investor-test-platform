package db

import "testing"

func TestNewAndSeed(t *testing.T) {
	d := New(":memory:")
	defer d.Close()
	d.SeedDefaults()

	var count int
	d.Conn.QueryRow("SELECT COUNT(*) FROM shoppers").Scan(&count)
	if count != 2 {
		t.Errorf("expected 2 seeded shoppers, got %d", count)
	}
}

func TestDropAll(t *testing.T) {
	d := New(":memory:")
	defer d.Close()
	d.SeedDefaults()
	d.DropAll()

	var count int
	d.Conn.QueryRow("SELECT COUNT(*) FROM shoppers").Scan(&count)
	if count != 0 {
		t.Errorf("expected 0 shoppers after drop, got %d", count)
	}
}
