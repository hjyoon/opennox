//go:build !server

package opennox

import (
	"context"
	"fmt"
	"math"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/opennox/libs/env"
	"github.com/timshannon/go-openal/openal"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/client/audio/ail"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/legacy/timer"
)

const nativeFXVolumeChild = "NOX_TEST_NATIVE_FX_VOLUME_CHILD"

// This exercises the actual native sample submission and reads its OpenAL
// source gain, not a predicted setting or a mock handle. The generated ADPCM
// is silent; ALSOFT_DRIVERS=null can select OpenAL Soft's real headless backend
// when a hardware device is unavailable. No stock assets or personal configuration
// are used. Isolate
// blob/timer initialization from the rest of the root package's fixtures.
func TestNativeAudioEffectsLiveVolume(t *testing.T) {
	if os.Getenv("NOX_TEST_NATIVE_FX") != "true" {
		t.Skip("set NOX_TEST_NATIVE_FX=true to test real native FX sample gain")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestNativeAudioEffectsLiveVolumeChild$", "-test.count=1", "-test.v")
	cmd.Dir = t.TempDir()
	for _, kv := range os.Environ() {
		key, _, _ := strings.Cut(kv, "=")
		if strings.HasPrefix(key, "NOX_TEST_NATIVE_FX_") || strings.HasPrefix(key, "NOX_E2E") || key == "NOX_DATA" {
			continue
		}
		cmd.Env = append(cmd.Env, kv)
	}
	cmd.Env = append(cmd.Env, nativeFXVolumeChild+"=true")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("native FX volume subprocess failed: %v\n%s", err, out)
	}
	t.Logf("%s", out)
}

func TestNativeAudioEffectsLiveVolumeChild(t *testing.T) {
	if os.Getenv(nativeFXVolumeChild) != "true" {
		t.Skip("only run in the isolated native FX subprocess")
	}
	if env.IsE2E() {
		t.Fatal("native FX gain test must not use E2E mock audio")
	}
	runtime.LockOSThread()
	t.Cleanup(runtime.UnlockOSThread)
	legacy.InitBlobData()
	prevTicks := timer.PlatformTicks
	timer.PlatformTicks = func() uint64 { return 0 }
	t.Cleanup(func() { timer.PlatformTicks = prevTicks })
	handles.Init()
	t.Cleanup(handles.Release)
	driver := ail.WaveOutOpen()
	if driver == 0 {
		t.Fatal("cannot open the configured OpenAL playback device")
	}
	t.Cleanup(func() {
		if err := driver.Close(); err != nil {
			t.Errorf("close native FX driver: %v", err)
		}
	})
	voice := driver.AllocateSample()
	if voice == 0 || voice.GetSource() == nil {
		t.Fatal("native FX sample has no real OpenAL source")
	}
	entry := &nativeAudioBankEntry{name: "silence", rate: 22050, flags: 8, blockSize: 256, data: make([]byte, 256)}
	state := nativeAudioEffectsState{
		bank:   &nativeAudioBank{entries: map[string]*nativeAudioBankEntry{"silence": entry}},
		voices: []ail.Sample{voice},
	}
	t.Cleanup(state.close)
	source := openal.Source(*voice.GetSource())
	t.Logf("native FX driver=%#x sample=%#x source=%d renderer=%q ALSOFT_DRIVERS=%q", driver, voice, source, openal.GetString(0xB003 /* AL_RENDERER */), os.Getenv("ALSOFT_DRIVERS"))
	fx := (*timer.TimerGroup)(legacy.Get_dword_587000_127004())
	if fx == nil {
		t.Fatal("FX timer group was not bound by blob initialization")
	}
	def := defaultNativeSoundDef()
	check := func(name string, live, requested int) {
		t.Helper()
		event := uint32(uint64(163*requested) * uint64(def.volume) >> 14)
		mixed := uint64(event) * uint64(live) / VolumeMax
		want := int(127 * mixed >> 14)
		if want > 127 {
			want = 127
		}
		if !state.playSampleLocked(sound.SoundShellClick, &def, "SILENCE", requested) {
			t.Fatalf("%s: valid silent ADPCM sample was not submitted", name)
		}
		got := source.Getf(openal.AlGain)
		if math.Abs(float64(got)-float64(float32(want)/127)) > 1e-6 {
			t.Fatalf("%s: submitted OpenAL gain=%g (%g/127), want %d/127", name, got, got*127, want)
		}
		if err := openal.Err(); err != nil {
			t.Fatalf("%s: OpenAL error: %v", name, err)
		}
	}
	count := 0
	for _, flags := range []uint32{8, 9} {
		entry.flags = flags
		for _, startup := range []int{VolumeMax, VolumeMax / 3, 1, 0} {
			configSetVolume(startup, VolumeFX)
			for _, live := range []int{0, 1, 100, 4850, 8126, 8192, 12000, VolumeMax} {
				for _, requested := range []int{100, 33, 1} {
					for _, eventDef := range []uint32{0x4000, 163 * 80, 163 * 255, 1} {
						def.volume = eventDef
						*fx = timer.TimerGroup{}
						fx.Timers[0].Current = uint32((live+1)%(VolumeMax+1)) << 16
						// A real GUI slider sends SetRaw. Do not advance its timer
						// in the fixture: production sample submission must do so.
						fx.Timers[0].SetRaw(uint32(live))
						name := fmt.Sprintf("flags=%d startup=%d live=%d requested=%d def=%d", flags, startup, live, requested, eventDef)
						check(name, live, requested)
						if fx.Timers[0].Current != uint32(live)<<16 || fx.Timers[0].Flags&2 == 0 {
							t.Fatalf("%s: pending raw FX timer was not updated: %+v", name, fx.Timers[0])
						}
						if fx.Timers[0].Target != uint32(live)<<16 || configGetVolume(VolumeFX) != startup {
							t.Fatalf("%s: sample submission changed the target or startup scalar", name)
						}
						count++
					}
				}
			}
		}
	}
	// An interpolated timer must use Current, not jump directly to Target.
	// The fractional part is discarded before mixing, as in TimerGroup.Mix.
	*fx = timer.TimerGroup{}
	fx.Timers[0].Current = 8126<<16 | 0xffff
	fx.Timers[0].Target = VolumeMax << 16
	def.volume = 0x4000
	check("interpolated current, not target", 8126, 100)
	if fx.Timers[0].Current != 8126<<16|0xffff || fx.Timers[0].Target != VolumeMax<<16 {
		t.Fatalf("interpolated FX timer was snapped to its target: %+v", fx.Timers[0])
	}
	t.Logf("checked %d raw mono/stereo submissions plus interpolated current and original mix-before-127 quantization", count)
}
