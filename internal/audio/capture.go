package audio

/*
#cgo darwin LDFLAGS: -framework AudioToolbox -framework CoreFoundation
#include <AudioToolbox/AudioToolbox.h>

extern void bonkAudioInputCallback(void *userData, short *samples, int count);

static void cAudioQueueInputCallback(void *userData, AudioQueueRef inAQ, AudioQueueBufferRef inBuffer, const AudioTimeStamp *inStartTime, UInt32 inNumberPacketDescriptions, const AudioStreamPacketDescription *inPacketDescs) {
    if (inBuffer->mAudioDataByteSize > 0) {
        int sampleCount = inBuffer->mAudioDataByteSize / sizeof(short);
        bonkAudioInputCallback(userData, (short *)inBuffer->mAudioData, sampleCount);
    }
    AudioQueueEnqueueBuffer(inAQ, inBuffer, 0, NULL);
}

static OSStatus bonkStartCapture(AudioQueueRef *outAQ, void *userData, double sampleRate, int bufferSamples) {
    AudioStreamBasicDescription desc = {0};
    desc.mSampleRate = sampleRate;
    desc.mFormatID = kAudioFormatLinearPCM;
    desc.mFormatFlags = kLinearPCMFormatFlagIsSignedInteger | kLinearPCMFormatFlagIsPacked;
    desc.mBytesPerPacket = 2;
    desc.mFramesPerPacket = 1;
    desc.mBytesPerFrame = 2;
    desc.mChannelsPerFrame = 1;
    desc.mBitsPerChannel = 16;

    OSStatus status = AudioQueueNewInput(&desc, cAudioQueueInputCallback, userData, NULL, NULL, 0, outAQ);
    if (status != 0) return status;

    UInt32 bufferBytes = bufferSamples * sizeof(short);
    for (int i = 0; i < 3; i++) {
        AudioQueueBufferRef buf;
        status = AudioQueueAllocateBuffer(*outAQ, bufferBytes, &buf);
        if (status != 0) return status;
        status = AudioQueueEnqueueBuffer(*outAQ, buf, 0, NULL);
        if (status != 0) return status;
    }

    return AudioQueueStart(*outAQ, NULL);
}

static OSStatus bonkStopCapture(AudioQueueRef aq) {
    AudioQueueStop(aq, true);
    return AudioQueueDispose(aq, true);
}
*/
import "C"
import (
	"fmt"
	"sync"
	"unsafe"
)

// Global registry for dispatching C callbacks to Go capture instances
var (
	captureRegistryMu sync.Mutex
	captureRegistry   = make(map[uintptr]*Capture)
	nextCaptureID     uintptr
)

// FrameHandler receives a frame of 16-bit PCM audio samples.
type FrameHandler func(samples []int16)

// Capture manages real-time microphone audio input via Apple CoreAudio AudioQueue.
type Capture struct {
	id            uintptr
	sampleRate    float64
	frameSize     int
	handler       FrameHandler
	mu            sync.Mutex
	aq            C.AudioQueueRef
	running       bool
	sampleScratch []int16
}

// NewCapture creates an AudioQueue capture device.
// Defaults: 16,000 Hz, 512 samples per frame (~32ms per buffer).
func NewCapture(handler FrameHandler) *Capture {
	captureRegistryMu.Lock()
	nextCaptureID++
	id := nextCaptureID
	c := &Capture{
		id:            id,
		sampleRate:    16000.0,
		frameSize:     512,
		handler:       handler,
		sampleScratch: make([]int16, 512),
	}
	captureRegistry[id] = c
	captureRegistryMu.Unlock()
	return c
}

//export bonkAudioInputCallback
func bonkAudioInputCallback(userData unsafe.Pointer, samples *C.short, count C.int) {
	id := uintptr(userData)
	captureRegistryMu.Lock()
	c, ok := captureRegistry[id]
	captureRegistryMu.Unlock()

	if !ok || c == nil || count <= 0 {
		return
	}

	n := int(count)
	slice := unsafe.Slice((*int16)(unsafe.Pointer(samples)), n)

	c.mu.Lock()
	if c.handler != nil && c.running {
		c.handler(slice)
	}
	c.mu.Unlock()
}

// Start begins recording audio from the default microphone.
func (c *Capture) Start() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.running {
		return nil
	}

	var aq C.AudioQueueRef
	status := C.bonkStartCapture(&aq, unsafe.Pointer(c.id), C.double(c.sampleRate), C.int(c.frameSize))
	if status != 0 {
		// Common CoreAudio errors
		if status == 0x7065726d || status == -66748 { // 'perm'
			return fmt.Errorf("microphone access denied by macOS: %s", FormatPermissionHelp())
		}
		return fmt.Errorf("AudioQueueStart failed with OSStatus %d", int(status))
	}

	c.aq = aq
	c.running = true
	return nil
}

// Stop terminates microphone audio capture and releases CoreAudio resources.
func (c *Capture) Stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.running {
		return nil
	}

	C.bonkStopCapture(c.aq)
	c.running = false
	c.aq = nil
	return nil
}

// IsRunning returns whether capture is currently active.
func (c *Capture) IsRunning() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.running
}
