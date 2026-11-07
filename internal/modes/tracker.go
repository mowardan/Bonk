package modes

import (
	"math"
	"math/rand"
	"sync"
	"time"
)

const (
	// DecayHalfLife controls how quickly escalation intensity fades (in seconds).
	DecayHalfLife = 30.0
)

// EscalationTracker manages hit counts, intensity scoring, and audio file selection.
type EscalationTracker struct {
	mu       sync.Mutex
	score    float64
	lastTime time.Time
	total    int
	halfLife float64
	scale    float64
	pack     *SoundPack
}

// NewEscalationTracker initializes a new tracker for the given pack.
func NewEscalationTracker(pack *SoundPack, cooldown time.Duration) *EscalationTracker {
	cooldownSec := cooldown.Seconds()
	if cooldownSec <= 0 {
		cooldownSec = 0.75
	}

	// scale maps the exponential curve so that sustained max-rate
	// slapping (one per cooldown) converges smoothly to the final file index.
	ssMax := 1.0 / (1.0 - math.Pow(0.5, cooldownSec/DecayHalfLife))
	fileCount := len(pack.Files)
	if fileCount == 0 {
		fileCount = 1
	}
	scale := (ssMax - 1) / math.Log(float64(fileCount+1))
	if scale <= 0 {
		scale = 1.0
	}

	return &EscalationTracker{
		halfLife: DecayHalfLife,
		scale:    scale,
		pack:     pack,
	}
}

// Record registers a new impact and returns the cumulative hit number and current escalation score.
func (t *EscalationTracker) Record(now time.Time) (int, float64) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.lastTime.IsZero() {
		elapsed := now.Sub(t.lastTime).Seconds()
		t.score *= math.Pow(0.5, elapsed/t.halfLife)
	}
	t.score += 1.0
	t.lastTime = now
	t.total++
	return t.total, t.score
}

// GetFile selects the appropriate sound file based on the play mode and current score.
func (t *EscalationTracker) GetFile(score float64) string {
	t.mu.Lock()
	defer t.mu.Unlock()

	if len(t.pack.Files) == 0 {
		return ""
	}

	if t.pack.Mode == ModeRandom {
		return t.pack.Files[rand.Intn(len(t.pack.Files))]
	}

	// Escalation mode: 1-exp(-x) curve maps score to file index
	maxIdx := len(t.pack.Files) - 1
	idx := int(float64(len(t.pack.Files)) * (1.0 - math.Exp(-(score-1)/t.scale)))
	if idx < 0 {
		idx = 0
	}
	if idx > maxIdx {
		idx = maxIdx
	}
	return t.pack.Files[idx]
}

// Reset clears the tracker state.
func (t *EscalationTracker) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.score = 0
	t.lastTime = time.Time{}
	t.total = 0
}
