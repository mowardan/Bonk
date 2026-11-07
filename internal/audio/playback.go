package audio

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"os"
	"sync"
	"time"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/effects"
	"github.com/gopxl/beep/v2/mp3"
	"github.com/gopxl/beep/v2/speaker"
)

// PlaybackEngine plays reaction sound files through the system output.
type PlaybackEngine struct {
	mu             sync.Mutex
	speakerInit    bool
	targetRate     beep.SampleRate
	volumeScaling  bool
	speedRatio     float64
	onPlayingState func(playing bool)
}

// NewPlaybackEngine initializes a playback engine.
func NewPlaybackEngine(volumeScaling bool, speedRatio float64, onPlayingState func(playing bool)) *PlaybackEngine {
	if speedRatio <= 0 {
		speedRatio = 1.0
	}
	return &PlaybackEngine{
		volumeScaling:  volumeScaling,
		speedRatio:     speedRatio,
		onPlayingState: onPlayingState,
	}
}

// SetVolumeScaling toggles amplitude-based volume scaling.
func (p *PlaybackEngine) SetVolumeScaling(enabled bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.volumeScaling = enabled
}

// SetSpeed updates playback speed multiplier.
func (p *PlaybackEngine) SetSpeed(speed float64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if speed > 0 {
		p.speedRatio = speed
	}
}

// AmplitudeToVolume maps impact amplitude [0.05, 0.80] to decibel volume gain [-3.0, 0.0] (base 2).
func AmplitudeToVolume(amplitude float64) float64 {
	const (
		minAmp = 0.05
		maxAmp = 0.80
		minVol = -3.0 // ~1/8th volume
		maxVol = 0.0  // full volume
	)

	if amplitude <= minAmp {
		return minVol
	}
	if amplitude >= maxAmp {
		return maxVol
	}

	t := (amplitude - minAmp) / (maxAmp - minAmp)
	t = math.Log(1+t*99) / math.Log(100) // Logarithmic curve
	return minVol + t*(maxVol-minVol)
}

// ResampleForPlayback wraps source so it matches the speaker sample rate with speed scaling.
func ResampleForPlayback(source beep.Streamer, fileRate, speakerRate beep.SampleRate, speed float64) beep.Streamer {
	srcRate := fileRate
	if speed != 1.0 && speed > 0 {
		srcRate = beep.SampleRate(int(float64(fileRate) * speed))
	}
	if srcRate == speakerRate {
		return source
	}
	return beep.Resample(4, srcRate, speakerRate, source)
}

// PlayData plays an in-memory MP3 byte slice.
func (p *PlaybackEngine) PlayData(data []byte, amplitude float64) error {
	streamer, format, err := mp3.Decode(io.NopCloser(bytes.NewReader(data)))
	if err != nil {
		return fmt.Errorf("mp3 decode: %w", err)
	}
	return p.playStreamer(streamer, format, amplitude)
}

// PlayFile plays an MP3 file from disk.
func (p *PlaybackEngine) PlayFile(filePath string, amplitude float64) error {
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open file %s: %w", filePath, err)
	}
	defer f.Close()

	streamer, format, err := mp3.Decode(f)
	if err != nil {
		return fmt.Errorf("mp3 decode %s: %w", filePath, err)
	}
	return p.playStreamer(streamer, format, amplitude)
}

func (p *PlaybackEngine) playStreamer(streamer beep.StreamSeekCloser, format beep.Format, amplitude float64) error {
	defer streamer.Close()

	p.mu.Lock()
	if !p.speakerInit {
		if err := speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10)); err != nil {
			p.mu.Unlock()
			return fmt.Errorf("speaker.Init: %w", err)
		}
		p.targetRate = format.SampleRate
		p.speakerInit = true
	}
	targetRate := p.targetRate
	volScaling := p.volumeScaling
	speed := p.speedRatio
	notify := p.onPlayingState
	p.mu.Unlock()

	var source beep.Streamer = streamer
	if volScaling {
		source = &effects.Volume{
			Streamer: streamer,
			Base:     2,
			Volume:   AmplitudeToVolume(amplitude),
			Silent:   false,
		}
	}

	source = ResampleForPlayback(source, format.SampleRate, targetRate, speed)

	// Notify detector to duck input
	if notify != nil {
		notify(true)
	}

	done := make(chan struct{})
	speaker.Play(beep.Seq(source, beep.Callback(func() {
		close(done)
	})))

	<-done

	if notify != nil {
		notify(false)
	}
	return nil
}
