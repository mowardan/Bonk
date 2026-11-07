<p align="center">
  <h1 align="center">💥 bonk</h1>
  <p align="center"><strong>Acoustic physical impact reaction detector for macOS.</strong></p>
  <p align="center"><em>Slap your MacBook, it yells back. Zero sudo. Zero root. Pure local DSP.</em></p>
</p>

---

## Highlights

* 🔒 **Zero Sudo / Zero Root:** Runs 100% in unprivileged user space. No administrator rights or `sudo` needed.
* 🎙️ **Microphone-Powered Transient DSP:** Uses native Apple CoreAudio to detect physical chassis impacts (slaps and taps) by analyzing acoustic impulse shockwaves.
* 💻 **Universal Mac Support:** Works on **all Apple Silicon Macs** (M1, M2, M3, M4, M-series) and **all Intel Macs**.
* 🛑 **Speech & Typing Rejection:** Intelligent acoustic filters distinguish physical unibody slaps from human speech, singing, ambient music, and keyboard typing clicks.
* 🛡️ **Privacy First:** 100% offline. Audio is processed entirely in volatile memory and never recorded to disk or sent over any network.
* 🔇 **Self-Ducking Feedback Prevention:** Automatically suppresses microphone detection while reaction sounds are playing through the speakers to avoid acoustic feedback loops.

---

## How It Works

When you physically hit, tap, or slap your MacBook's unibody chassis, a mechanical stress wave propagates through the aluminum enclosure directly into the internal microphone diaphragm. 

`bonk` captures audio via Apple's documented CoreAudio `AudioQueue` framework and runs real-time digital signal processing:

1. **RMS Energy & Peak Extraction:** Computes instantaneous energy and peak amplitude on short 32ms audio frames.
2. **Crest Factor Analysis:** Impulsive structural shockwaves exhibit an extremely high peak-to-RMS ratio (crest factor > 2.6) compared to continuous speech or music (~1.4 - 2.2).
3. **Attack Rate Jump:** Measures rapid onset acceleration (instantaneous rise within < 3ms).
4. **Adaptive Noise Floor Tracking:** Dynamically adjusts the detection threshold to ambient room noise (air conditioning, office chatter) using an asymmetric exponential moving average.
5. **Rejection Filters:** Rejects sustained vocal formants and high-frequency scissor-switch keyboard clicks.

---

## Requirements

* **Operating System:** macOS 11.0 (Big Sur) or newer (tested on macOS 12, 13, 14, 15, and 26+).
* **Architecture:** Universal support for both **Apple Silicon** (`arm64`) and **Intel** (`x86_64`).
* **Permissions:** Standard macOS microphone permission (requested automatically upon first run).

---

## Installation

### Option A: Install via Homebrew / Pre-built Universal Binary

Download the latest universal release:

```bash
curl -fsSL -o bonk https://github.com/bonk-project/bonk/releases/latest/download/bonk_universal
chmod +x bonk
sudo mv bonk /usr/local/bin/bonk   # Optional: move to PATH (standard install, bonk itself runs without sudo)
```

### Option B: Build From Source

Requires Go 1.22+ and Apple Clang (Command Line Tools):

```bash
git clone https://github.com/bonk-project/bonk.git
cd bonk

# Build native binary for your Mac
make build

# Or build a Universal Binary (arm64 + x86_64)
make build-universal
```

The compiled binary will be placed in `./bin/bonk`.

---

## Usage

Run `bonk` directly in your terminal (**no `sudo` required!**):

```bash
# 1. Normal mode — yells "ow!" when slapped
bonk

# 2. Sexy mode — reaction sounds intensify as you slap more frequently
bonk --sexy

# 3. Halo mode — plays Halo death audio clips on impact
bonk --halo

# 4. Lizard mode — escalating lizard reactions
bonk --lizard

# 5. Custom audio directory — play your own MP3 files
bonk --custom /path/to/my/sounds
```

### Options & Tuning

```text
Flags:
  -s, --sexy                Escalating intensity mode (reactions intensify with slap rate)
  -H, --halo                Play random Halo death sounds on impact
  -l, --lizard              Escalating lizard sounds mode
  -c, --custom string       Path to custom directory containing MP3 audio files
      --sensitivity float   Acoustic sensitivity (0.01 to 1.0, higher = more sensitive) (default 0.25)
      --cooldown int        Minimum cooldown between responses in milliseconds (default 750)
      --volume-scaling      Scale playback volume by impact intensity (harder hit = louder yell)
      --speed float         Playback speed multiplier (e.g. 0.5 = half speed, 2.0 = double) (default 1)
      --stdio               JSON stream mode for scripting and GUI integration
  -v, --version             Display version
  -h, --help                Help for bonk
```

### Sensitivity Tuning

* **Increase sensitivity (for gentle taps):**
  ```bash
  bonk --sensitivity 0.50
  ```
* **Decrease sensitivity (requires a firm slap):**
  ```bash
  bonk --sensitivity 0.10
  ```
* **Enable volume scaling (harder hits yell louder):**
  ```bash
  bonk --volume-scaling
  ```

---

## Microphone Permission

When you launch `bonk` for the first time, macOS will display a prompt asking:
> *"Terminal" would like to access the microphone.*

Click **Allow**. 

### Troubleshooting Permissions:
If you accidentally clicked "Don't Allow" or microphone access was previously disabled:
1. Open **System Settings** on your Mac.
2. Navigate to **Privacy & Security** $\rightarrow$ **Microphone**.
3. Locate your terminal application (e.g. *Terminal*, *iTerm2*, or *Ghostty*) and toggle the switch **ON**.
4. Re-run `bonk`.

---

## Manual Test Procedure

To verify impact detection in real time:

1. Launch `bonk` in verbose mode:
   ```bash
   ./bin/bonk --sensitivity 0.25
   ```
2. Speak normally or type on the keyboard — notice that `bonk` rejects speech and typing.
3. Firmly tap or slap the aluminum palm rest to the left or right of your trackpad.
4. You will see an immediate detection log:
   ```text
   impact #1 [amp=0.642 conf=0.88 score=1.00] -> sounds/pain/00_Ow.mp3
   ```
   and hear the reaction audio played through your Mac's speakers.

---

## Privacy & Security Guarantee

* **No sudo:** Never run `bonk` with `sudo`. It does not need administrator rights.
* **No root services:** Installs no background daemons, LaunchAgents, or LaunchDaemons.
* **No disk recordings:** Audio buffers are ephemeral and processed in volatile RAM only.
* **No network activity:** Zero telemetry, zero analytics, and zero outbound network calls.

Read full details in [PRIVACY.md](PRIVACY.md) and [SECURITY.md](SECURITY.md).

---

## Uninstallation

Because `bonk` is a self-contained single binary with no system files or background daemons, uninstalling it is as simple as deleting the binary:

```bash
rm -f /usr/local/bin/bonk
```

---

## License

This project is licensed under the [MIT License](LICENSE).
Third-party sound assets remain the property of their respective creators and are included for personal, non-commercial entertainment purposes.
