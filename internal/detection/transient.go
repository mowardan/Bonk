package detection

// TransientFeatures holds temporal and spectral characteristics of an audio frame.
type TransientFeatures struct {
	Peak             float64
	RMS              float64
	CrestFactor      float64
	AttackRatio      float64 // Current peak relative to previous frame RMS
	ZeroCrossingRate float64
	LowFreqRatio     float64 // Approximate low-frequency concentration (chassis thud vs keystroke click)
}

// AnalyzeTransient extracts transient features from the current audio frame
// given the RMS energy of the preceding frame and baseline noise floor.
func AnalyzeTransient(samples []int16, prevRMS, noiseFloor float64) TransientFeatures {
	peak := CalculatePeak(samples)
	rms := CalculateRMS(samples)
	crest := CalculateCrestFactor(peak, rms)

	refRMS := prevRMS
	if noiseFloor > refRMS {
		refRMS = noiseFloor
	}
	if refRMS < 0.001 {
		refRMS = 0.001
	}

	attackRatio := peak / refRMS

	// Zero Crossing Rate (ZCR)
	var zcr float64
	var zeroCrossings int
	if len(samples) > 1 {
		for i := 1; i < len(samples); i++ {
			if (samples[i-1] >= 0 && samples[i] < 0) || (samples[i-1] < 0 && samples[i] >= 0) {
				zeroCrossings++
			}
		}
		zcr = float64(zeroCrossings) / float64(len(samples)-1)
	}

	// Approximate low-frequency content via simple 1st-order difference filter (high-pass proxy).
	// A chassis slap has high structural bass (high energy in low-pass / low difference energy).
	// Keystroke clicks have very high sample-to-sample difference energy (predominantly high-frequency).
	var diffEnergy float64
	var rawEnergy float64
	for i := 1; i < len(samples); i++ {
		diff := float64(samples[i] - samples[i-1])
		raw := float64(samples[i])
		diffEnergy += diff * diff
		rawEnergy += raw * raw
	}

	lowFreqRatio := 0.5
	if diffEnergy > 0 && rawEnergy > 0 {
		// If diffEnergy is much larger than rawEnergy, it's dominated by high-frequency clicks (e.g. typing)
		ratio := rawEnergy / (rawEnergy + diffEnergy*0.25)
		if ratio > 1.0 {
			ratio = 1.0
		}
		lowFreqRatio = ratio
	}

	return TransientFeatures{
		Peak:             peak,
		RMS:              rms,
		CrestFactor:      crest,
		AttackRatio:      attackRatio,
		ZeroCrossingRate: zcr,
		LowFreqRatio:     lowFreqRatio,
	}
}
