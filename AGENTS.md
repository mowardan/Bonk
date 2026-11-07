# AGENTS.md

> Guidelines for AI agents working in this repository.

## Project Overview

**bonk** is a native macOS application that detects physical impacts (slaps and taps) on MacBook laptops using real-time CoreAudio acoustic transient analysis and plays funny audio reactions.

- **Platform**: macOS 11.0+ on both Apple Silicon (`arm64`) and Intel (`x86_64`)
- **Runtime requirement**: Standard unprivileged user. **NEVER requires sudo or root.**
- **Architecture**: Modular Go package using Apple CoreAudio (`AudioQueue`) and in-memory DSP algorithms.

## Commands

### Build & Test

```bash
# Run all unit tests
go test -v ./...

# Build native binary (bin/bonk)
make build

# Build universal binary (arm64 + amd64)
make build-universal
```

### Run (No sudo needed)

```bash
./bin/bonk
./bin/bonk --sexy
./bin/bonk --halo
./bin/bonk --custom /path/to/mp3s
./bin/bonk --sensitivity 0.35
```

## Code Organization

```
bonk/
├── cmd/
│   └── bonk/
│       └── main.go              # Application CLI entrypoint and orchestrator
├── internal/
│   ├── audio/
│   │   ├── capture.go          # CoreAudio AudioQueue real-time microphone capture
│   │   ├── playback.go         # Beep-based audio playback with feedback ducking
│   │   └── permissions.go      # AVFoundation microphone permission checker
│   ├── detection/
│   │   ├── detector.go         # Core impact detection state machine
│   │   ├── energy.go           # RMS, peak amplitude, and crest factor DSP
│   │   ├── filter.go           # Adaptive noise floor and rejection filters
│   │   └── transient.go        # Transient attack and rise-time extraction
│   ├── modes/
│   │   ├── pack.go             # Embedded sound pack manager
│   │   └── tracker.go          # Escalation curve and score tracking
│   └── config/
│       └── config.go           # CLI flags and runtime configuration
├── assets/
│   ├── assets.go               # Embedded audio asset declarations
│   └── sounds/                 # Embedded MP3 assets (pain, sexy, halo, lizard)
├── tests/                      # Synthetic audio DSP and integration tests
├── Makefile                    # Multi-architecture and universal build rules
├── go.mod
├── go.sum
├── PRIVACY.md
├── SECURITY.md
└── README.md
```

## Key Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/gopxl/beep/v2` | CoreAudio MP3 audio decoding and speaker playback |
| `github.com/spf13/cobra` | CLI command and flag parsing |
| `Apple CoreAudio / AudioToolbox` | System microphone audio capture via `AudioQueue` |
| `Apple AVFoundation` | Microphone permission status checks |

## Architectural Constraints

1. **Strictly Non-Root**: Never introduce root checks, `os.Geteuid() == 0`, or `sudo` requirements.
2. **Ephemeral Memory Only**: Never write microphone audio to disk or transmit it over the network.
3. **Universal Compatibility**: Must compile and run on both Apple Silicon (`arm64`) and Intel (`x86_64`) Macs.
4. **Self-Ducking**: Always ensure `detector.SetPlaying(true)` is called while reaction sounds are playing to prevent speaker feedback loops.
