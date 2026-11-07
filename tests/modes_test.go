package tests

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"bonk/internal/modes"
)

func TestSoundPackLoading(t *testing.T) {
	t.Run("load normal pack", func(t *testing.T) {
		pack, err := modes.LoadPack("normal", "")
		if err != nil {
			t.Fatalf("failed to load normal pack: %v", err)
		}
		if len(pack.Files) == 0 {
			t.Errorf("normal pack has 0 files")
		}
		if pack.Mode != modes.ModeRandom {
			t.Errorf("expected ModeRandom, got %v", pack.Mode)
		}
	})

	t.Run("load sexy pack", func(t *testing.T) {
		pack, err := modes.LoadPack("sexy", "")
		if err != nil {
			t.Fatalf("failed to load sexy pack: %v", err)
		}
		if len(pack.Files) == 0 {
			t.Errorf("sexy pack has 0 files")
		}
		if pack.Mode != modes.ModeEscalation {
			t.Errorf("expected ModeEscalation, got %v", pack.Mode)
		}
	})

	t.Run("load halo pack", func(t *testing.T) {
		pack, err := modes.LoadPack("halo", "")
		if err != nil {
			t.Fatalf("failed to load halo pack: %v", err)
		}
		if len(pack.Files) == 0 {
			t.Errorf("halo pack has 0 files")
		}
	})

	t.Run("load lizard pack", func(t *testing.T) {
		pack, err := modes.LoadPack("lizard", "")
		if err != nil {
			t.Fatalf("failed to load lizard pack: %v", err)
		}
		if len(pack.Files) == 0 {
			t.Errorf("lizard pack has 0 files")
		}
	})

	t.Run("load custom pack with missing dir errors", func(t *testing.T) {
		_, err := modes.LoadPack("custom", "/nonexistent/directory/path/sounds")
		if err == nil {
			t.Errorf("expected error for nonexistent custom directory, got nil")
		}
	})

	t.Run("load custom pack with valid files", func(t *testing.T) {
		tmpDir := t.TempDir()
		// Create mock mp3s
		os.WriteFile(filepath.Join(tmpDir, "01.mp3"), []byte("mock"), 0644)
		os.WriteFile(filepath.Join(tmpDir, "02.mp3"), []byte("mock"), 0644)
		os.WriteFile(filepath.Join(tmpDir, "ignore.txt"), []byte("txt"), 0644)

		pack, err := modes.LoadPack("custom", tmpDir)
		if err != nil {
			t.Fatalf("failed to load valid custom pack: %v", err)
		}
		if len(pack.Files) != 2 {
			t.Errorf("expected 2 mp3 files, got %d", len(pack.Files))
		}
	})
}

func TestEscalationTracker(t *testing.T) {
	pack, err := modes.LoadPack("sexy", "")
	if err != nil {
		t.Fatalf("failed to load sexy pack: %v", err)
	}

	cooldown := 750 * time.Millisecond
	tracker := modes.NewEscalationTracker(pack, cooldown)

	now := time.Now()

	// Initial hit
	total, score1 := tracker.Record(now)
	if total != 1 {
		t.Errorf("expected total 1, got %d", total)
	}
	if score1 != 1.0 {
		t.Errorf("expected score 1.0, got %f", score1)
	}

	firstFile := tracker.GetFile(score1)
	if firstFile == "" {
		t.Errorf("expected valid file, got empty string")
	}

	// Rapid succession hits should increase score monotonically
	prevScore := score1
	for i := 2; i <= 15; i++ {
		now = now.Add(cooldown + 10*time.Millisecond)
		tot, score := tracker.Record(now)
		if tot != i {
			t.Errorf("expected total %d, got %d", i, tot)
		}
		if score <= prevScore {
			t.Errorf("expected score to escalate: current %f <= prev %f", score, prevScore)
		}
		prevScore = score
	}

	highFile := tracker.GetFile(prevScore)
	if highFile == firstFile {
		t.Errorf("escalated score should map to a later file in the pack")
	}

	// Decay test: after 60s inactivity (2 half-lives), score should significantly decay
	now = now.Add(60 * time.Second)
	_, decayedScore := tracker.Record(now)
	if decayedScore >= prevScore {
		t.Errorf("score should decay after 60s, got %f vs prev %f", decayedScore, prevScore)
	}
}
