package tests

import (
	"math"
	"math/rand"
	"testing"
	"time"

	"bonk/internal/detection"
)

// Helper generators for synthetic audio frames (512 samples at 16 kHz = 32ms)

func makeSilence(n int) []int16 {
	return make([]int16, n)
}

func makeNoise(n int, amplitude float64) []int16 {
	samples := make([]int16, n)
	ampInt := int(amplitude * 32767.0)
	if ampInt <= 0 {
		return samples
	}
	r := rand.New(rand.NewSource(42))
	for i := range samples {
		samples[i] = int16(r.Intn(2*ampInt) - ampInt)
	}
	return samples
}

// makeImpactImpulse simulates a physical chassis slap:
// Extremely sharp rise within 1-2 samples, high amplitude, low-frequency structural ring, rapid decay.
func makeImpactImpulse(n int, peakAmp float64) []int16 {
	samples := make([]int16, n)
	for i := 0; i < n; i++ {
		t := float64(i)
		// Exponential decay with ~150Hz structural resonance, peak at t=0
		decay := math.Exp(-t / 30.0)
		osc := math.Cos(2.0 * math.Pi * 150.0 * (t / 16000.0))
		val := peakAmp * decay * osc
		if val > 1.0 {
			val = 1.0
		}
		if val < -1.0 {
			val = -1.0
		}
		samples[i] = int16(val * 32767.0)
	}
	return samples
}

// makeSpeechFrame simulates continuous human speech:
// Harmonic vocal formants (120Hz pitch + 800Hz / 1800Hz formants), sustained amplitude, low crest factor (~2.0).
func makeSpeechFrame(n int, amp float64) []int16 {
	samples := make([]int16, n)
	for i := 0; i < n; i++ {
		t := float64(i) / 16000.0
		s1 := math.Sin(2.0 * math.Pi * 130.0 * t)
		s2 := 0.7 * math.Sin(2.0 * math.Pi * 780.0 * t)
		s3 := 0.5 * math.Sin(2.0 * math.Pi * 1850.0 * t)
		comb := (s1 + s2 + s3) / 2.2
		samples[i] = int16(comb * amp * 32767.0)
	}
	return samples
}

// makeTypingClick simulates a keypress:
// Low overall energy, high-frequency localized click, very little sub-500Hz energy.
func makeTypingClick(n int, peakAmp float64) []int16 {
	samples := make([]int16, n)
	// Click happens around sample 50, rapid oscillation at 3500 Hz, decays in 15 samples
	for i := 50; i < 70 && i < n; i++ {
		t := float64(i - 50)
		samples[i] = int16(peakAmp * math.Exp(-t/4.0) * math.Sin(2.0*math.Pi*3500.0*(t/16000.0)) * 32767.0)
	}
	return samples
}

func TestSyntheticImpactDetection(t *testing.T) {
	cfg := detection.DefaultConfig()
	cfg.Sensitivity = 0.30
	det := detection.NewDetector(cfg)

	// Feed ambient noise first to establish noise floor
	for i := 0; i < 10; i++ {
		det.ProcessFrame(makeNoise(512, 0.01))
	}

	// Feed a chassis slap impact (peak 0.65)
	slap := makeImpactImpulse(512, 0.65)
	event := det.ProcessFrame(slap)

	if event == nil {
		t.Fatalf("expected physical impact to be detected, got nil")
	}

	if event.Amplitude < 0.5 {
		t.Errorf("expected detected amplitude > 0.5, got %f", event.Amplitude)
	}
	if event.CrestFactor < 2.6 {
		t.Errorf("expected crest factor >= 2.6, got %f", event.CrestFactor)
	}
}

func TestSpeechRejection(t *testing.T) {
	cfg := detection.DefaultConfig()
	det := detection.NewDetector(cfg)

	// Establish noise floor
	for i := 0; i < 5; i++ {
		det.ProcessFrame(makeNoise(512, 0.01))
	}

	// Simulate ongoing speech across multiple frames
	detected := false
	for i := 0; i < 8; i++ {
		speech := makeSpeechFrame(512, 0.40)
		if ev := det.ProcessFrame(speech); ev != nil {
			detected = true
			break
		}
	}

	if detected {
		t.Errorf("expected speech frame to be rejected, but impact was triggered")
	}
}

func TestTypingRejection(t *testing.T) {
	cfg := detection.DefaultConfig()
	det := detection.NewDetector(cfg)

	// Establish noise floor
	for i := 0; i < 5; i++ {
		det.ProcessFrame(makeNoise(512, 0.01))
	}

	// Feed typing keystroke
	keystroke := makeTypingClick(512, 0.18)
	if ev := det.ProcessFrame(keystroke); ev != nil {
		t.Errorf("expected keyboard typing click to be rejected, but impact was triggered")
	}
}

func TestAmbientNoiseAdaptation(t *testing.T) {
	cfg := detection.DefaultConfig()
	det := detection.NewDetector(cfg)

	// Feed gradually increasing background noise (e.g. music or room AC ramp up)
	for amp := 0.01; amp <= 0.08; amp += 0.005 {
		for i := 0; i < 5; i++ {
			ev := det.ProcessFrame(makeNoise(512, amp))
			if ev != nil {
				t.Fatalf("background noise ramp triggered false positive at amp %f", amp)
			}
		}
	}
}

func TestCooldownAndDuplicateSuppression(t *testing.T) {
	cfg := detection.DefaultConfig()
	cfg.Cooldown = 500 * time.Millisecond
	det := detection.NewDetector(cfg)

	// Establish noise floor
	for i := 0; i < 5; i++ {
		det.ProcessFrame(makeNoise(512, 0.01))
	}

	// First impact: should trigger
	ev1 := det.ProcessFrame(makeImpactImpulse(512, 0.70))
	if ev1 == nil {
		t.Fatalf("first impact should trigger")
	}

	// Second immediate impact (echo / bounce within cooldown): should be suppressed
	ev2 := det.ProcessFrame(makeImpactImpulse(512, 0.70))
	if ev2 != nil {
		t.Errorf("immediate second impact should be suppressed by cooldown")
	}

	// Wait past cooldown
	time.Sleep(550 * time.Millisecond)

	// Third impact: should trigger again
	ev3 := det.ProcessFrame(makeImpactImpulse(512, 0.70))
	if ev3 == nil {
		t.Errorf("third impact after cooldown should trigger")
	}
}

func TestDuckingPlaybackSuppression(t *testing.T) {
	cfg := detection.DefaultConfig()
	det := detection.NewDetector(cfg)

	// Simulate speaker output actively playing
	det.SetPlaying(true)

	// Strong sound occurs while playing (e.g. speakers playing loud sound)
	ev := det.ProcessFrame(makeImpactImpulse(512, 0.85))
	if ev != nil {
		t.Errorf("impact should be suppressed while speaker is playing to prevent feedback loop")
	}

	det.SetPlaying(false)
}

func TestAdaptiveSensitivity(t *testing.T) {
	// Low sensitivity detector
	cfgLow := detection.DefaultConfig()
	cfgLow.Sensitivity = 0.05
	detLow := detection.NewDetector(cfgLow)

	// High sensitivity detector
	cfgHigh := detection.DefaultConfig()
	cfgHigh.Sensitivity = 0.85
	detHigh := detection.NewDetector(cfgHigh)

	// Initialize both
	for i := 0; i < 5; i++ {
		noise := makeNoise(512, 0.01)
		detLow.ProcessFrame(noise)
		detHigh.ProcessFrame(noise)
	}

	// Light tap (peak 0.22)
	lightTap := makeImpactImpulse(512, 0.22)

	evLow := detLow.ProcessFrame(lightTap)
	evHigh := detHigh.ProcessFrame(lightTap)

	if evLow != nil {
		t.Errorf("low sensitivity should NOT trigger on light tap")
	}
	if evHigh == nil {
		t.Errorf("high sensitivity SHOULD trigger on light tap")
	}
}
