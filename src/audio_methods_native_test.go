//go:build nox_audio_methods_fixture && !server

package opennox

import (
	"math"
	"os"
	"runtime"
	"testing"
	"time"
	"unsafe"

	"github.com/opennox/libs/env"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/client/audio/ail"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
)

func TestAudioDefaultMethodsNativeCABI(t *testing.T) {
	if os.Getenv("NOX_TEST_OPENAL") != "true" {
		t.Skip("set NOX_TEST_OPENAL=true for default native C method playback")
	}
	if env.IsE2E() {
		t.Fatal("unset E2E mock mode")
	}
	handles.InitHigh()
	defer handles.Release()
	for _, fallback := range []bool{false, true} {
		t.Run(map[bool]string{false: "buffer-format", true: "context-format"}[fallback], func(t *testing.T) {
			runtime.LockOSThread()
			defer runtime.UnlockOSThread()
			oldDriver := audioDev
			audioDev = 0
			defer func() { audioDev = oldDriver }()
			contexts, freeCtx := alloc.MakeHigh([]legacy.AudioContextNative{}, 1)
			defer freeCtx()
			ctx := &contexts[0]
			ctx.Group.Init()
			ctx.Methods = legacy.AudioMethodsFixtureClass().Methods
			ctx.Format60 = [7]uint32{4, 2, 22050, 2, 2, 0, 256}
			if result := legacy.AudioMethodsFixtureContextCreate(ctx); result != 0 || audioDev == 0 || ctx.Limit != 24 {
				t.Fatalf("real C context creation: %d / %x / limit%d", result, audioDev, ctx.Limit)
			}
			defer legacy.AudioMethodsFixtureContextDestroy(ctx)
			driver := audioDev
			if got := legacy.AudioMethodsFixtureDriver(); got != uintptr(driver) || unsafe.Sizeof(uintptr(0)) == 8 && got <= math.MaxUint32 {
				t.Fatalf("real driver getter truncated: %x != %x", got, driver)
			}
			voice := legacy.AudioMethodsFixtureCreate(ctx)
			if voice == nil {
				t.Fatal("real C source constructor")
			}
			defer func() {
				if voice != nil {
					legacy.AudioMethodsFixtureDestroy(voice)
				}
			}()
			s := voice.Sample
			if s == nil || s.Smp == 0 || s.Dev != driver || s.Field1 != unsafe.Pointer(voice) || s.Data1 == nil || s.Data2 == nil {
				t.Fatal("default create callback native sample/owner/buffers")
			}
			buffers, freeBuffer := alloc.MakeHigh([]legacy.AudioBufferNative{}, 1)
			defer freeBuffer()
			b := &buffers[0]
			legacy.AudioMethodsFixtureInitBuffer(b)
			data, freeData := alloc.MakeHigh([]byte{}, 8*256)
			defer freeData()
			format, freeFormat := alloc.MakeHigh([][7]uint32{}, 1)
			defer freeFormat()
			format[0] = [7]uint32{4, 2, 11025, 1, 2, 0, 256}
			b.Data, b.Size = &data[0], uint32(len(data))
			var expectedFormat unsafe.Pointer
			if fallback {
				expectedFormat = unsafe.Pointer(&ctx.Format60[0])
			} else {
				b.Format = &format[0]
				expectedFormat = unsafe.Pointer(b.Format)
			}
			legacy.AudioMethodsFixtureSelect(voice, b)
			voice.Repeat = -1
			if result := legacy.AudioMethodsFixturePlay(voice); result != 0 || voice.Flags&1 == 0 || s.Field4 != expectedFormat || s.Flag7 != 0 {
				t.Fatalf("default play callback native format/state: result%d format%p want%p flags%x finished%d", result, s.Field4, expectedFormat, voice.Flags, s.Flag7)
			}
			deadline := time.Now().Add(4 * time.Second)
			for driver.PlaybackStats().SampleBuffersQueued == 0 && time.Now().Before(deadline) {
				time.Sleep(10 * time.Millisecond)
				ail.Serve()
			}
			stats := driver.PlaybackStats()
			if !stats.Native || stats.SampleSources != 1 || stats.SampleBuffersQueued == 0 || s.Smp.Status() != 4 {
				t.Fatalf("default method did not reach native OpenAL: %+v status%d", stats, s.Smp.Status())
			}
			voice.Mixed.Timers[0].Current = 0
			legacy.AudioMethodsFixtureRefresh(voice)
			if legacy.AudioMethodsFixturePause(voice) != 0 {
				t.Fatal("default C pause callback") // AIL pause stops input consumption; queued OpenAL buffers may still report playing.
			}
			if legacy.AudioMethodsFixtureStop(voice) != 0 || s.Flag7 != 1 || voice.Flags&5 != 0 || voice.Buffer != nil || voice.Repeat != 0 {
				t.Fatal("default stop→EOS native callback state")
			}
			if legacy.AudioMethodsFixtureStop(voice) != 0 || s.Flag7 != 1 {
				t.Fatal("idempotent native stop")
			}
			legacy.AudioMethodsFixtureDestroy(voice)
			voice = nil
			if driver.PlaybackStats().SampleSources != 0 {
				t.Fatal("default destructor retained a native sample")
			}
			t.Logf("C default create/play/refresh/pause/stop/destroy passed with native driver=%#x, format=%p", driver, expectedFormat)
		})
	}
}
