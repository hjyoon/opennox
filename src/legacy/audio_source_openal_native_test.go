//go:build !server

package legacy

import (
	"math"
	"os"
	"runtime"
	"testing"
	"time"
	"unsafe"

	"github.com/opennox/libs/env"
	"github.com/opennox/opennox/v1/legacy/client/audio/ail"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
)

// Real OpenAL (including its explicit null output driver), with all buffer,
// producer and EOS work entering through the exported C function names.
func TestAudioSourceNativeOpenAL(t *testing.T) {
	if os.Getenv("NOX_TEST_OPENAL") != "true" {
		t.Skip("set NOX_TEST_OPENAL=true for native source playback")
	}
	if env.IsE2E() {
		t.Fatal("unset E2E mode for this real OpenAL check")
	}
	handles.InitHigh()
	defer handles.Release()
	for _, stereo := range []bool{false, true} {
		t.Run(map[bool]string{false: "mono", true: "stereo"}[stereo], func(t *testing.T) {
			runtime.LockOSThread()
			defer runtime.UnlockOSThread()
			driver := ail.WaveOutOpen()
			if driver == 0 {
				t.Fatal("cannot open native OpenAL driver")
			}
			defer func() {
				if err := driver.Close(); err != nil {
					t.Error(err)
				}
			}()
			a, _ := asnOwnedFixture(t)
			audioSourceTestServices = nil
			h := driver.AllocateSample()
			if h == 0 {
				t.Fatal("native sample handle")
			}
			defer h.Release()
			if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(driver) <= math.MaxUint32 || uintptr(h) <= math.MaxUint32) {
				t.Fatal("driver/sample handles must exceed 4 GiB")
			}
			h.Init()
			format := int32(5)
			if stereo {
				format = 7
			}
			h.SetType(format, 0)
			h.SetADPCMBlockSize(256)
			h.SetPlaybackRate(22050)
			h.SetVolume(0)
			v := (*AudioVoiceNative)(a.take(unsafe.Sizeof(AudioVoiceNative{})))
			s := (*AudioSample)(a.take(unsafe.Sizeof(AudioSample{})))
			b := (*AudioBufferNative)(a.take(unsafe.Sizeof(AudioBufferNative{})))
			asnPointer(19, unsafe.Pointer(b), nil, 0, 0, 0)
			for _, size := range []int32{3 * 256, 61 * 256, 4 * 256} {
				c := a.take(unsafe.Sizeof(AudioChunkNative{}))
				data := a.take(uintptr(size)) // zero predictors/indices/nibbles: silent ADPCM
				asnPointer(20, c, data, size, 0, 0)
				asnScalar(9, unsafe.Pointer(b), c, 0)
			}
			asnPointer(4, unsafe.Pointer(v), nil, 0, 0, 0)
			v.OnEnd = asnCallback(5, false)
			endCalls := 0
			audioNativeFixtureCallbackService = func(kind int, p unsafe.Pointer) int32 {
				if kind != 5 || p != unsafe.Pointer(v) {
					t.Fatal("native EOS callback owner")
				}
				endCalls++
				return 0
			}
			asnPointer(3, unsafe.Pointer(v), unsafe.Pointer(b), 0, 0, 0)
			s.Dev, s.Smp, s.Field1 = driver, h, unsafe.Pointer(v)
			s.Data1, s.Data2 = (*byte)(a.take(0x4000)), (*byte)(a.take(0x4000))
			h.SetUserData(s)
			h.RegisterEOBCallback(func() { asnScalar(0, unsafe.Pointer(s), nil, 0) })
			h.RegisterEOSCallback(func() { asnScalar(1, unsafe.Pointer(uintptr(h)), nil, 0) })
			before := driver.PlaybackStats()
			if asnScalar(0, unsafe.Pointer(s), nil, 0) != -1 {
				t.Fatal("initial pump")
			}
			deadline := time.Now().Add(6 * time.Second)
			for (s.Flag7 == 0 || h.Status() == 4) && time.Now().Before(deadline) {
				time.Sleep(10 * time.Millisecond)
				ail.Serve()
			}
			after := driver.PlaybackStats()
			if !after.Native || s.Field3 != 1 || s.Flag7 != 1 || h.Status() != 2 || endCalls != 1 || v.Buffer != nil || after.SampleBuffersQueued-before.SampleBuffersQueued != 68 || after.SampleBuffersProcessed-before.SampleBuffersProcessed != 68 {
				t.Fatalf("native source playback/EOS failed: eof=%d finished=%d status=%d end=%d before=%+v after=%+v", s.Field3, s.Flag7, h.Status(), endCalls, before, after)
			}
			asnScalar(1, unsafe.Pointer(uintptr(h)), nil, 0)
			if endCalls != 1 {
				t.Fatal("duplicate EOS callback")
			}
			t.Logf("driver=%#x sample=%#x source=%p sample-owner=%p OpenAL queued/processed=68/68", driver, h, v, s)
		})
	}
}
