package detection

import (
	"math"
)

// CalculateRMS calculates the Root Mean Square (RMS) energy of a 16-bit PCM buffer.
// Returns a normalized value in the range [0.0, 1.0].
func CalculateRMS(samples []int16) float64 {
	if len(samples) == 0 {
		return 0.0
	}
	var sumSquares float64
	for _, s := range samples {
		// Normalize sample to [-1.0, 1.0]
		norm := float64(s) / 32768.0
		sumSquares += norm * norm
	}
	return math.Sqrt(sumSquares / float64(len(samples)))
}

// CalculatePeak finds the maximum absolute amplitude in a 16-bit PCM buffer.
// Returns a normalized value in the range [0.0, 1.0].
func CalculatePeak(samples []int16) float64 {
	if len(samples) == 0 {
		return 0.0
	}
	var maxAbs int16
	for _, s := range samples {
		abs := s
		if abs < 0 {
			if abs == math.MinInt16 {
				abs = math.MaxInt16
			} else {
				abs = -abs
			}
		}
		if abs > maxAbs {
			maxAbs = abs
		}
	}
	return float64(maxAbs) / 32768.0
}

// CalculateCrestFactor computes the ratio of peak amplitude to RMS energy.
// Impulsive transients (like chassis impacts) have high crest factors (typically > 3.0),
// whereas sustained sounds like speech or music have lower crest factors (typically 1.4 - 2.5).
func CalculateCrestFactor(peak, rms float64) float64 {
	if rms <= 1e-6 {
		return 1.0
	}
	return peak / rms
}
