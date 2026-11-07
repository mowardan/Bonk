package detection

// AdaptiveNoiseFloor estimates the background ambient noise floor dynamically.
// It uses an asymmetric exponential moving average:
// - Slowly rises during persistent background noise (fans, air conditioning, room noise).
// - Quickly falls when ambient noise quiets down.
// - Rejects sudden impulsive transients so impacts do not artificially raise the noise floor.
type AdaptiveNoiseFloor struct {
	floor     float64
	minFloor  float64
	maxFloor  float64
	alphaUp   float64
	alphaDown float64
}

// NewAdaptiveNoiseFloor creates a noise floor tracker with defaults.
func NewAdaptiveNoiseFloor() *AdaptiveNoiseFloor {
	return &AdaptiveNoiseFloor{
		floor:     0.010,
		minFloor:  0.002,
		maxFloor:  0.200,
		alphaUp:   0.005, // Slowly adapt up (avoids rising on short bursts)
		alphaDown: 0.050, // Rapidly adapt down when quiet
	}
}

// Update updates the noise floor estimate with the current frame RMS.
func (nf *AdaptiveNoiseFloor) Update(frameRMS float64) float64 {
	// Clamp frameRMS to avoid runaway
	if frameRMS > nf.maxFloor {
		frameRMS = nf.maxFloor
	}

	if frameRMS > nf.floor {
		// Only adapt upward slowly
		nf.floor += nf.alphaUp * (frameRMS - nf.floor)
	} else {
		// Adapt downward faster
		nf.floor += nf.alphaDown * (frameRMS - nf.floor)
	}

	if nf.floor < nf.minFloor {
		nf.floor = nf.minFloor
	}
	return nf.floor
}

// Current returns the current noise floor.
func (nf *AdaptiveNoiseFloor) Current() float64 {
	return nf.floor
}

// Reset resets the noise floor to baseline.
func (nf *AdaptiveNoiseFloor) Reset() {
	nf.floor = 0.010
}

// IsSpeechLike checks if the features correspond to human speech rather than a mechanical impact.
// Speech is characterized by:
// - Low crest factor (< 2.6) due to sustained vowel energy
// - Slower attack ratio (< 3.0)
// - Sustained RMS across consecutive frames
func IsSpeechLike(features TransientFeatures, consecutiveSustainedFrames int) bool {
	// If the sound is sustained over multiple frames (e.g. speaking a syllable), it is speech
	if consecutiveSustainedFrames >= 3 {
		return true
	}
	// Speech vowel formants have a low peak-to-RMS ratio compared to mechanical shockwaves
	if features.CrestFactor < 2.5 && features.AttackRatio < 3.2 {
		return true
	}
	return false
}

// IsTypingLike checks if the features correspond to typing on a laptop keyboard.
// Keystrokes are localized, produce little chassis displacement, and have low low-frequency energy.
func IsTypingLike(features TransientFeatures) bool {
	// Typing has very high high-frequency difference energy (low LowFreqRatio)
	// and lower absolute peak amplitude.
	if features.LowFreqRatio < 0.20 && features.Peak < 0.25 {
		return true
	}
	return false
}
