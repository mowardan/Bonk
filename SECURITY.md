# Security Policy for bonk

## Security Principles

**bonk** was engineered from the ground up to operate under strict least-privilege principles:

1. **No Root Privileges / No `sudo`:** `bonk` runs entirely as an unprivileged user process in user space. It never requests, requires, or checks for root permissions.
2. **No Kernel or Private Extensions:** `bonk` does not install kernel extensions (kexts), system extensions, privileged helper tools, or LaunchDaemons.
3. **Public Apple APIs Only:** Audio input and output use Apple's public, documented frameworks (`AudioToolbox.framework`, `CoreFoundation.framework`, and `AVFoundation.framework`).
4. **Offline Operation:** The binary has no network capabilities and cannot communicate over the internet.
5. **No Shell Execution:** `bonk` does not invoke external subshells (`/bin/sh`, `bash`, `zsh`), sub-processes, or remote scripts.

---

## Supported Versions

Security updates and patches are provided for the latest release:

| Version | Supported          |
| ------- | ------------------ |
| 2.x     | :white_check_mark: |
| < 2.0   | :x:                |

---

## Reporting a Vulnerability

If you discover any security concern or potential vulnerability in `bonk`:

1. Do not open a public GitHub issue.
2. Report the vulnerability privately via GitHub Security Advisories or by emailing the project maintainers.
3. Please include:
   - macOS version and hardware architecture (Intel vs Apple Silicon).
   - Detailed steps to reproduce the issue.
   - Proof-of-concept code if applicable.

We will review and respond to reports within 48 hours.
