//go:build !server

package opennox

import (
	"context"
	"math"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"
	"time"
	"unsafe"

	"github.com/opennox/libs/env"
	"github.com/timshannon/go-openal/openal"

	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/client/audio/ail"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/legacy/timer"
)

const nativeFXEnabledChild = "NOX_TEST_NATIVE_FX_ENABLED_CHILD"

// This is an API integration fixture, not simulated GUI input. It invokes
// the actual C in-game Options procedure and unchanged native sound callback,
// then reads the real OpenAL source. Stock Options.wnd and queued mouse input
// are covered separately by the headless GUI audit.
func TestNativeAudioEffectsEnabled(t *testing.T) {
	if os.Getenv("NOX_TEST_NATIVE_FX") != "true" {
		t.Skip("set NOX_TEST_NATIVE_FX=true to test real native FX enable/re-enable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestNativeAudioEffectsEnabledChild$", "-test.count=1", "-test.v")
	cmd.Dir = t.TempDir()
	for _, kv := range os.Environ() {
		key, _, _ := strings.Cut(kv, "=")
		if strings.HasPrefix(key, "NOX_TEST_NATIVE_FX_") || strings.HasPrefix(key, "NOX_E2E") || key == "NOX_DATA" {
			continue
		}
		cmd.Env = append(cmd.Env, kv)
	}
	cmd.Env = append(cmd.Env, nativeFXEnabledChild+"=true")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("native FX enabled subprocess failed: %v\n%s", err, out)
	}
	t.Logf("%s", out)
}

func TestNativeAudioEffectsEnabledChild(t *testing.T) {
	if os.Getenv(nativeFXEnabledChild) != "true" {
		t.Skip("only run in the isolated native FX enabled subprocess")
	}
	if env.IsE2E() {
		t.Fatal("native FX enable test must not use E2E mock audio")
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
	nativeAudioFX.bank = &nativeAudioBank{entries: map[string]*nativeAudioBankEntry{"silence": entry}}
	nativeAudioFX.voices = []ail.Sample{voice}
	def := defaultNativeSoundDef()
	def.enabled, def.sampleIDs = true, []string{"silence"}
	nativeAudioFX.defs[sound.SoundShellClick] = def
	t.Cleanup(nativeAudioFX.close)
	source := openal.Source(*voice.GetSource())
	t.Logf("native FX driver=%#x sample=%#x source=%d renderer=%q ALSOFT_DRIVERS=%q", driver, voice, source, openal.GetString(0xB003 /* AL_RENDERER */), os.Getenv("ALSOFT_DRIVERS"))
	fx := (*timer.TimerGroup)(legacy.Get_dword_587000_127004())
	if fx == nil || legacy.Sub_453070() != 1 {
		t.Fatal("fresh process did not initialize the FX timer and enabled flag")
	}
	g := gui.New(nil)
	t.Cleanup(g.DestroyAll)
	root := g.NewWindowRaw(nil, gui.StatusEnabled, 0, 0, 640, 480, nil)
	checkbox := g.NewWindowRaw(root, gui.StatusEnabled, 0, 0, 70, 12, nil)
	checkbox.SetID(361)
	root.SetFunc94C(legacy.OptionsInGameProc4ADF30())
	if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(root.C()) <= math.MaxUint32 || uintptr(checkbox.C()) <= math.MaxUint32) {
		t.Fatal("native Options windows did not exceed 4 GiB")
	}
	t.Logf("C Options procedure=%p native root=%p checkbox=%p", legacy.OptionsInGameProc4ADF30(), root, checkbox)
	toggle := func(t *testing.T) {
		t.Helper()
		before, startup := legacy.Sub_453070(), configGetVolume(VolumeFX)
		got := gui.EventRespInt(root.Func94(&gui.RawEvent{Event: 0x4007, Arg1: uintptr(checkbox.C()), Arg2: ^uintptr(0)}))
		if got != 1 || legacy.Sub_453070() != 1-before || configGetVolume(VolumeFX) != startup {
			t.Fatalf("C checkbox toggle: return=%d enabled=%d previous=%d startup=%d", got, legacy.Sub_453070(), before, configGetVolume(VolumeFX))
		}
	}
	setRaw := func(volume int) {
		*fx = timer.TimerGroup{}
		fx.Timers[0].SetRaw(uint32(volume))
	}
	checkGain := func(t *testing.T, volume, requested int) {
		t.Helper()
		if requested > 100 {
			requested = 100
		}
		event := uint64(163 * requested)
		mixed := event * uint64(volume) / VolumeMax
		want := int(127 * mixed >> 14)
		got := source.Getf(openal.AlGain)
		if math.Abs(float64(got)-float64(float32(want)/127)) > 1e-6 {
			t.Fatalf("submitted native gain=%g (%g/127), want %d/127", got, got*127, want)
		}
		if err := openal.Err(); err != nil {
			t.Fatalf("OpenAL error: %v", err)
		}
	}
	t.Run("mute-with-nonzero-startup", func(t *testing.T) {
		configSetVolume(VolumeMax, VolumeFX)
		setRaw(4850)
		voice.SetVolume(17)
		before := source.Getf(openal.AlGain)
		toggle(t)
		t.Cleanup(func() { toggle(t) })
		if got := source.Getf(openal.AlGain); got != before {
			t.Errorf("muted C checkbox still submitted its click: gain %g -> %g", before, got)
		}
		for _, requested := range []int{1, 33, 100, 101} {
			if nativeAudioFX.play(sound.SoundShellClick, requested) {
				t.Errorf("muted FX admitted sound with requested volume %d and startup volume %d", requested, configGetVolume(VolumeFX))
			}
		}
		if got := source.Getf(openal.AlGain); got != before || fx.Timers[0].Current != 0 {
			t.Errorf("muted events modified the real source or pending FX timer: gain=%g timer=%+v", got, fx.Timers[0])
		}
	})
	t.Run("re-enable-with-zero-startup", func(t *testing.T) {
		configSetVolume(0, VolumeFX)
		setRaw(8126)
		voice.SetVolume(17)
		toggle(t)
		if source.Getf(openal.AlGain) != float32(17)/127 {
			t.Error("disabling FX submitted a new sample")
		}
		toggle(t)
		checkGain(t, 8126, 100) // The actual C checkbox click uses the native callback.
		if !nativeAudioFX.play(sound.SoundShellClick, 33) {
			t.Error("enabled FX remained blocked by zero startup volume")
		} else {
			checkGain(t, 8126, 33)
		}
		if configGetVolume(VolumeFX) != 0 {
			t.Fatal("re-enabling rewrote the startup scalar instead of using live state")
		}
	})
	t.Run("request-guards-and-clamping", func(t *testing.T) {
		configSetVolume(0, VolumeFX)
		setRaw(VolumeMax)
		for _, requested := range []int{-1, 0} {
			if nativeAudioFX.play(sound.SoundShellClick, requested) {
				t.Errorf("nonpositive requested volume %d was admitted", requested)
			}
		}
		for _, requested := range []int{1, 33, 100, 101, math.MaxInt} {
			if !nativeAudioFX.play(sound.SoundShellClick, requested) {
				t.Errorf("enabled FX rejected requested volume %d", requested)
				continue
			}
			checkGain(t, VolumeMax, requested)
		}
		for _, id := range []sound.ID{-1, 0, nativeSoundDefCount} {
			if nativeAudioFX.play(id, 100) {
				t.Errorf("invalid sound ID %d was admitted", id)
			}
		}
	})
}
