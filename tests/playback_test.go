package tests

import (
	"math"
	"testing"

	"bonk/internal/audio"

	"github.com/gopxl/beep/v2"
)

type dummyStreamer struct{}

func (dummyStreamer) Stream(samples [][2]float64) (int, bool) { return 0, false }
func (dummyStreamer) Err() error                              { return nil }

func TestAmplitudeToVolume(t *testing.T) {
	t.Run("clamping at boundaries", func(t *testing.T) {
		if v := audio.AmplitudeToVolume(0.01); v != -3.0 {
			t.Errorf("below min amplitude should return -3.0, got %f", v)
		}
		if v := audio.AmplitudeToVolume(0.05); v != -3.0 {
			t.Errorf("at min amplitude should return -3.0, got %f", v)
		}
		if v := audio.AmplitudeToVolume(0.80); v != 0.0 {
			t.Errorf("at max amplitude should return 0.0, got %f", v)
		}
		if v := audio.AmplitudeToVolume(1.50); v != 0.0 {
			t.Errorf("above max amplitude should return 0.0, got %f", v)
		}
	})

	t.Run("monotonic volume scaling", func(t *testing.T) {
		prev := audio.AmplitudeToVolume(0.05)
		for amp := 0.10; amp <= 0.80; amp += 0.05 {
			cur := audio.AmplitudeToVolume(amp)
			if cur < prev-1e-9 {
				t.Errorf("volume non-monotonic: at amp %f got %f < prev %f", amp, cur, prev)
			}
			prev = cur
		}
	})

	t.Run("valid finite numbers", func(t *testing.T) {
		for _, amp := range []float64{0, 0.05, 0.1, 0.35, 0.5, 0.8, 2.0} {
			v := audio.AmplitudeToVolume(amp)
			if math.IsNaN(v) || math.IsInf(v, 0) {
				t.Errorf("got NaN or Inf at amp %f", amp)
			}
		}
	})
}

func TestResampleForPlayback(t *testing.T) {
	src := dummyStreamer{}

	t.Run("matching sample rate and normal speed returns original source", func(t *testing.T) {
		res := audio.ResampleForPlayback(src, 44100, 44100, 1.0)
		if res != beep.Streamer(src) {
			t.Errorf("expected source to be returned unchanged")
		}
	})

	t.Run("mismatched sample rate creates resampler", func(t *testing.T) {
		res := audio.ResampleForPlayback(src, 48000, 44100, 1.0)
		r, ok := res.(*beep.Resampler)
		if !ok {
			t.Fatalf("expected *beep.Resampler, got %T", res)
		}
		expectedRatio := 48000.0 / 44100.0
		if math.Abs(r.Ratio()-expectedRatio) > 1e-6 {
			t.Errorf("expected ratio %f, got %f", expectedRatio, r.Ratio())
		}
	})

	t.Run("speed ratio folds into resampler", func(t *testing.T) {
		res := audio.ResampleForPlayback(src, 44100, 44100, 1.5)
		r, ok := res.(*beep.Resampler)
		if !ok {
			t.Fatalf("expected *beep.Resampler, got %T", res)
		}
		expectedRatio := 1.5
		if math.Abs(r.Ratio()-expectedRatio) > 1e-6 {
			t.Errorf("expected ratio %f, got %f", expectedRatio, r.Ratio())
		}
	})
}
