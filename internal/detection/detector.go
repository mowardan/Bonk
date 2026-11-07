package detection

import (
	"sync"
	"time"
)

// ImpactEvent represents a detected physical impact.
type ImpactEvent struct {
	Time        time.Time
	Amplitude   float64 // Normalized amplitude [0.0, 1.0]
	CrestFactor float64
	AttackRatio float64
	Confidence  float64
}

// Config controls detector sensitivity and timing parameters.
type Config struct {
	// Sensitivity: 0.01 (least sensitive, requires hard hit) to 1.0 (most sensitive, triggers on light tap). Default: 0.25.
	Sensitivity float64

	// Cooldown: Minimum duration between detected impact events. Default: 750ms.
	Cooldown time.Duration

	// MinPeak: Absolute minimum peak amplitude required regardless of noise floor.
	MinPeak float64

	// MinCrestFactor: Minimum crest factor (peak/RMS) required to identify an impulse.
	MinCrestFactor float64

	// MinAttackRatio: Minimum sudden increase in amplitude relative to baseline.
	MinAttackRatio float64
}

// DefaultConfig returns recommended detection defaults tuned for MacBook unibody impacts.
func DefaultConfig() Config {
	return Config{
		Sensitivity:    0.25,
		Cooldown:       750 * time.Millisecond,
		MinPeak:        0.12,
		MinCrestFactor: 2.6,
		MinAttackRatio: 3.0,
	}
}

// Detector processes real-time audio frames and flags physical chassis impacts.
type Detector struct {
	cfg Config

	mu                 sync.RWMutex
	noiseFloor         *AdaptiveNoiseFloor
	prevRMS            float64
	lastFrameTime      time.Time
	sustainedFrames    int
	lastTriggerTime    time.Time
	isPlaying          bool
	suppressUntil      time.Time
}

// NewDetector creates a new Detector with the specified configuration.
func NewDetector(cfg Config) *Detector {
	if cfg.Sensitivity <= 0 {
		cfg.Sensitivity = 0.25
	}
	if cfg.Cooldown <= 0 {
		cfg.Cooldown = 750 * time.Millisecond
	}
	if cfg.MinPeak <= 0 {
		cfg.MinPeak = 0.12
	}
	if cfg.MinCrestFactor <= 0 {
		cfg.MinCrestFactor = 2.6
	}
	if cfg.MinAttackRatio <= 0 {
		cfg.MinAttackRatio = 3.0
	}

	return &Detector{
		cfg:        cfg,
		noiseFloor: NewAdaptiveNoiseFloor(),
	}
}

// SetPlaying signals that an audio response is currently playing through the speakers.
// This activates self-ducking to prevent the laptop's own audio from triggering feedback loops.
func (d *Detector) SetPlaying(playing bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.isPlaying = playing
	if playing {
		// Extend suppression window 300ms past playback stop
		d.suppressUntil = time.Now().Add(300 * time.Millisecond)
	}
}

// UpdateSensitivity updates the sensitivity setting at runtime.
func (d *Detector) UpdateSensitivity(s float64) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if s > 0 && s <= 1.0 {
		d.cfg.Sensitivity = s
	}
}

// UpdateCooldown updates the cooldown setting at runtime.
func (d *Detector) UpdateCooldown(c time.Duration) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if c > 0 {
		d.cfg.Cooldown = c
	}
}

// ProcessFrame inspects an incoming slice of 16-bit PCM samples.
// Returns an *ImpactEvent if a physical impact is detected, or nil otherwise.
func (d *Detector) ProcessFrame(samples []int16) *ImpactEvent {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now()

	// If there was a gap between frames (e.g. pause or test delay), reset transient baseline
	if !d.lastFrameTime.IsZero() && now.Sub(d.lastFrameTime) > 100*time.Millisecond {
		d.prevRMS = d.noiseFloor.Current()
		d.sustainedFrames = 0
	}
	d.lastFrameTime = now

	// 1. Check self-suppression during playback / ducking
	if d.isPlaying || now.Before(d.suppressUntil) {
		return nil
	}

	// 2. Extract features
	currentFloor := d.noiseFloor.Current()
	features := AnalyzeTransient(samples, d.prevRMS, currentFloor)

	// Update sustained sound tracker (speech/noise persistence)
	if features.RMS > currentFloor*2.0 {
		d.sustainedFrames++
	} else {
		if d.sustainedFrames > 0 {
			d.sustainedFrames--
		}
	}

	// 3. Update noise floor only when not in an active high transient
	if features.AttackRatio < 2.5 {
		d.noiseFloor.Update(features.RMS)
	}

	d.prevRMS = features.RMS

	// 4. Check Cooldown
	if now.Sub(d.lastTriggerTime) < d.cfg.Cooldown {
		return nil
	}

	// 5. Dynamic Threshold calculation:
	// Sensitivity scales from 0.01 (harder threshold) to 1.0 (lower threshold)
	// Base threshold adapts with ambient noise floor.
	thresholdScale := (1.0 - d.cfg.Sensitivity) // 0.0 (max sens) to 0.99 (min sens)
	dynamicThreshold := currentFloor*3.0 + d.cfg.MinPeak + thresholdScale*0.45

	// Clamp minimum threshold
	if dynamicThreshold < d.cfg.MinPeak*0.5 {
		dynamicThreshold = d.cfg.MinPeak * 0.5
	}

	// 6. Impact criteria:
	// - Peak must clear the dynamic threshold
	// - Must exhibit high crest factor (impulse spike)
	// - Must exhibit sudden attack jump
	if features.Peak < dynamicThreshold {
		return nil
	}

	if features.CrestFactor < d.cfg.MinCrestFactor {
		return nil
	}

	if features.AttackRatio < d.cfg.MinAttackRatio {
		return nil
	}

	// 7. Rejection filters
	if IsSpeechLike(features, d.sustainedFrames) {
		return nil
	}

	if IsTypingLike(features) {
		return nil
	}

	// Impact confirmed!
	d.lastTriggerTime = now

	// Calculate confidence metric based on feature margin
	conf := (features.CrestFactor / 3.5) * (features.AttackRatio / 4.0)
	if conf > 1.0 {
		conf = 1.0
	}

	return &ImpactEvent{
		Time:        now,
		Amplitude:   features.Peak,
		CrestFactor: features.CrestFactor,
		AttackRatio: features.AttackRatio,
		Confidence:  conf,
	}
}
