# Privacy Policy for bonk

**Last updated:** November 2026

Privacy and local execution are fundamental engineering principles of **bonk**. This document explicitly outlines how audio data is handled and your privacy rights when using this application.

---

## 1. Zero Data Collection or Exfiltration

- **No Audio Storage:** `bonk` **never** records, saves, or writes microphone audio to disk or permanent storage.
- **No Network Transmission:** `bonk` contains **zero network libraries**, opens zero network sockets, and makes zero inbound or outbound connections. Your audio never leaves your device.
- **No Telemetry or Analytics:** `bonk` does not track usage statistics, crash telemetry, device identifiers, or user behavior.
- **No File or Credential Access:** `bonk` does not access your Keychain, browser data, passwords, home directory documents, or environment variables.

---

## 2. In-Memory Ephemeral DSP

- Audio is streamed directly from Apple's public CoreAudio `AudioQueue` into small, fixed-size ring buffers in volatile RAM (typically ~32ms per frame, 512 samples at 16 kHz).
- The DSP engine immediately calculates mathematical metrics (RMS energy, peak amplitude, and crest factor) to identify impact impulses.
- Audio samples are discarded immediately after frame analysis is complete. At no point is an audio file formed or retained in memory.

---

## 3. Why Microphone Permission is Required

macOS requires explicit user consent before any application can read input from the built-in microphone.

- `bonk` uses the microphone strictly to measure the acoustic shockwave transmitted through the laptop chassis when you physically tap or slap it.
- Standard macOS permission controls (TCC) are fully respected. If access is not granted or revoked, `bonk` does not attempt to bypass permissions; it cleanly informs you and exits.

---

## 4. How to Revoke Access

You maintain complete control over microphone access. To revoke permission at any time:

1. Open **System Settings** on your Mac.
2. Select **Privacy & Security** from the sidebar.
3. Click **Microphone**.
4. Toggle OFF the permission for your terminal application (e.g. _Terminal_, _iTerm2_, or _Ghostty_).

---

## 5. Security & Verification

`bonk` is 100% open-source. You can inspect the entire audio capture and DSP pipeline in [`internal/audio/capture.go`](internal/audio/capture.go) and [`internal/detection/detector.go`](internal/detection/detector.go) to verify that no network or persistent storage operations exist.
