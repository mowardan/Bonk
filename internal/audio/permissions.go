package audio

/*
#cgo darwin LDFLAGS: -framework AVFoundation -framework CoreFoundation
#include <CoreFoundation/CoreFoundation.h>
#include <objc/runtime.h>
#include <objc/message.h>

static int getMicrophoneAuthStatus(void) {
    Class cls = (Class)objc_getClass("AVCaptureDevice");
    if (!cls) {
        return 3; // Fallback to authorized if class not found
    }
    SEL sel = sel_registerName("authorizationStatusForMediaType:");
    CFStringRef mediaType = CFSTR("soun"); // AVMediaTypeAudio
    int (*msgSend)(id, SEL, CFStringRef) = (int (*)(id, SEL, CFStringRef))objc_msgSend;
    return msgSend((id)cls, sel, mediaType);
}
*/
import "C"
import "fmt"

// PermissionStatus represents the current state of macOS microphone access.
type PermissionStatus int

const (
	StatusNotDetermined PermissionStatus = 0
	StatusRestricted    PermissionStatus = 1
	StatusDenied        PermissionStatus = 2
	StatusAuthorized    PermissionStatus = 3
)

// CheckMicrophonePermission returns the current authorization status via AVFoundation.
func CheckMicrophonePermission() PermissionStatus {
	return PermissionStatus(C.getMicrophoneAuthStatus())
}

// FormatPermissionHelp returns friendly instructions for granting microphone permission in macOS.
func FormatPermissionHelp() string {
	return `
================================================================================
⚠️  Microphone Permission Required
================================================================================
bonk uses your MacBook's built-in microphone to detect physical impact shockwaves.
Audio is processed entirely in volatile memory and NEVER recorded, stored, or sent
over the network.

To grant permission:
1. Open System Settings
2. Navigate to "Privacy & Security" -> "Microphone"
3. Toggle ON the switch next to your terminal application (e.g., Terminal, iTerm2, Ghostty)
4. Re-launch bonk
================================================================================`
}

func (s PermissionStatus) String() string {
	switch s {
	case StatusAuthorized:
		return "Authorized"
	case StatusDenied:
		return "Denied"
	case StatusRestricted:
		return "Restricted"
	case StatusNotDetermined:
		return "Not Determined"
	default:
		return fmt.Sprintf("Unknown (%d)", int(s))
	}
}
