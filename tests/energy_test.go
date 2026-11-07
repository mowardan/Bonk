package tests

import (
	"math"
	"testing"

	"bonk/internal/detection"
)

func TestCalculateRMS(t *testing.T) {
	t.Run("silence returns zero", func(t *testing.T) {
		samples := make([]int16, 512)
		rms := detection.CalculateRMS(samples)
		if rms != 0.0 {
			t.Errorf("expected 0.0, got %f", rms)
		}
	})

	t.Run("full scale square wave returns 1.0", func(t *testing.T) {
		samples := make([]int16, 512)
		for i := range samples {
			if i%2 == 0 {
				samples[i] = 32767
			} else {
				samples[i] = -32767
			}
		}
		rms := detection.CalculateRMS(samples)
		// 32767/32768 is ~0.999969
		if math.Abs(rms-1.0) > 0.001 {
			t.Errorf("expected ~1.0, got %f", rms)
		}
	})

	t.Run("empty slice returns zero", func(t *testing.T) {
		if rms := detection.CalculateRMS(nil); rms != 0.0 {
			t.Errorf("expected 0.0, got %f", rms)
		}
	})
}

func TestCalculatePeak(t *testing.T) {
	t.Run("silence returns zero", func(t *testing.T) {
		samples := make([]int16, 256)
		peak := detection.CalculatePeak(samples)
		if peak != 0.0 {
			t.Errorf("expected 0.0, got %f", peak)
		}
	})

	t.Run("finds positive peak", func(t *testing.T) {
		samples := []int16{0, 100, 16384, 50, 0}
		peak := detection.CalculatePeak(samples)
		expected := 16384.0 / 32768.0 // 0.5
		if math.Abs(peak-expected) > 1e-6 {
			t.Errorf("expected %f, got %f", expected, peak)
		}
	})

	t.Run("finds negative peak and handles MinInt16", func(t *testing.T) {
		samples := []int16{0, -100, math.MinInt16, 50}
		peak := detection.CalculatePeak(samples)
		if peak < 0.999 {
			t.Errorf("expected peak close to 1.0, got %f", peak)
		}
	})
}

func TestCalculateCrestFactor(t *testing.T) {
	t.Run("silence returns 1.0", func(t *testing.T) {
		crest := detection.CalculateCrestFactor(0, 0)
		if crest != 1.0 {
			t.Errorf("expected 1.0, got %f", crest)
		}
	})

	t.Run("sine wave crest factor is approximately sqrt(2)", func(t *testing.T) {
		// A pure sine wave has crest factor = Peak / (Peak / sqrt(2)) = sqrt(2) ~ 1.414
		peak := 0.5
		rms := peak / math.Sqrt(2)
		crest := detection.CalculateCrestFactor(peak, rms)
		if math.Abs(crest-math.Sqrt(2)) > 0.01 {
			t.Errorf("expected ~1.414, got %f", crest)
		}
	})

	t.Run("sharp impulse has high crest factor", func(t *testing.T) {
		peak := 0.9
		rms := 0.15
		crest := detection.CalculateCrestFactor(peak, rms)
		if crest < 5.0 {
			t.Errorf("expected crest factor >= 5.0, got %f", crest)
		}
	})
}
