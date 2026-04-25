package git

import (
	"runtime"
	"testing"
)

func TestHintsForCurrentOS_NonEmpty(t *testing.T) {
	hints := hintsForCurrentOS()
	if len(hints) == 0 {
		t.Fatalf("expected at least one hint for OS %q", runtime.GOOS)
	}
	for i, h := range hints {
		if h.Label == "" {
			t.Errorf("hint[%d] missing Label: %+v", i, h)
		}
		if h.Command == "" && h.URL == "" {
			t.Errorf("hint[%d] must have Command or URL: %+v", i, h)
		}
	}
}

func TestHintsForCurrentOS_PlatformShape(t *testing.T) {
	hints := hintsForCurrentOS()
	switch runtime.GOOS {
	case "windows":
		if hints[0].Method != "winget" {
			t.Errorf("windows top hint should be winget, got %q", hints[0].Method)
		}
	case "darwin":
		if hints[0].Method != "brew" {
			t.Errorf("darwin top hint should be brew, got %q", hints[0].Method)
		}
	case "linux":
		if hints[0].Method != "apt" {
			t.Errorf("linux top hint should be apt, got %q", hints[0].Method)
		}
	}
}
