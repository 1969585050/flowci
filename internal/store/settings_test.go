package store

import "testing"

func TestSettings_SaveAndGet(t *testing.T) {
	cleanup := setupTestStore(t)
	defer cleanup()

	if err := SaveSettings("theme", "dark"); err != nil {
		t.Fatalf("SaveSettings: %v", err)
	}
	if err := SaveSettings("locale", "zh-CN"); err != nil {
		t.Fatalf("SaveSettings: %v", err)
	}

	m, err := GetSettings()
	if err != nil {
		t.Fatalf("GetSettings: %v", err)
	}
	if m["theme"] != "dark" {
		t.Errorf("theme = %q, want dark", m["theme"])
	}
	if m["locale"] != "zh-CN" {
		t.Errorf("locale = %q, want zh-CN", m["locale"])
	}
}

// TestSettings_UpsertOverwrite 同一 key 二次写入应覆盖。
func TestSettings_UpsertOverwrite(t *testing.T) {
	cleanup := setupTestStore(t)
	defer cleanup()

	_ = SaveSettings("theme", "dark")
	_ = SaveSettings("theme", "light")

	m, _ := GetSettings()
	if m["theme"] != "light" {
		t.Errorf("theme should be overwritten to 'light', got %q", m["theme"])
	}
}

func TestGetSettings_Empty(t *testing.T) {
	cleanup := setupTestStore(t)
	defer cleanup()

	m, err := GetSettings()
	if err != nil {
		t.Fatalf("GetSettings: %v", err)
	}
	if len(m) != 0 {
		t.Errorf("expected empty map, got %d keys", len(m))
	}
}
