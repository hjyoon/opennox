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

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/client/audio/ail"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/legacy/timer"
)

const nativeFXServiceChild = "NOX_TEST_NATIVE_FX_SERVICE_CHILD"

// API integration fixture: generated silent ADPCM, the actual C Options
// procedure and the production 30 Hz timer via AIL_serve. No direct invocation
// of the timer callback, E2E mock handles, stock assets or personal settings.
func TestNativeAudioEffectsLiveService(t *testing.T) {
	if os.Getenv("NOX_TEST_NATIVE_FX") != "true" {
		t.Skip("set NOX_TEST_NATIVE_FX=true to test already-playing native FX")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestNativeAudioEffectsLiveServiceChild$", "-test.count=1", "-test.v")
	cmd.Dir = t.TempDir()
	for _, kv := range os.Environ() {
		key, _, _ := strings.Cut(kv, "=")
		if strings.HasPrefix(key, "NOX_TEST_NATIVE_FX_") || strings.HasPrefix(key, "NOX_E2E") || key == "NOX_DATA" {
			continue
		}
		cmd.Env = append(cmd.Env, kv)
	}
	cmd.Env = append(cmd.Env, nativeFXServiceChild+"=true")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("native FX service subprocess failed: %v\n%s", err, out)
	}
	t.Logf("%s", out)
}

func TestNativeAudioEffectsLiveServiceChild(t *testing.T) {
	if os.Getenv(nativeFXServiceChild) != "true" {
		t.Skip("only run in the isolated native FX service subprocess")
	}
	if env.IsE2E() {
		t.Fatal("live FX service test must not use E2E mock audio")
	}
	runtime.LockOSThread()
	t.Cleanup(runtime.UnlockOSThread)
	legacy.InitBlobData()
	previousTicks, ticks := timer.PlatformTicks, uint64(0)
	timer.PlatformTicks = func() uint64 { return ticks }
	t.Cleanup(func() { timer.PlatformTicks = previousTicks })
	handles.Init()
	t.Cleanup(handles.Release)
	t.Cleanup(sub_43E9F0)
	driver := ail.WaveOutOpen()
	if driver == 0 {
		t.Fatal("cannot open the configured OpenAL playback device")
	}
	t.Cleanup(func() {
		if err := driver.Close(); err != nil {
			t.Errorf("close native FX service driver: %v", err)
		}
	})
	entry := &nativeAudioBankEntry{name: "silence", rate: 1000, flags: 8, blockSize: 256, data: make([]byte, 256*256)}
	stereo := *entry
	stereo.name, stereo.flags = "stereo", 9
	nativeAudioFX.bank = &nativeAudioBank{entries: map[string]*nativeAudioBankEntry{"silence": entry, "stereo": &stereo}}
	for i := 0; i < 4; i++ {
		voice := driver.AllocateSample()
		if voice == 0 || voice.GetSource() == nil {
			t.Fatal("native FX sample has no real OpenAL source")
		}
		nativeAudioFX.voices = append(nativeAudioFX.voices, voice)
	}
	t.Cleanup(nativeAudioFX.close)
	fx := (*timer.TimerGroup)(legacy.Get_dword_587000_127004())
	*fx = timer.TimerGroup{}
	fx.Timers[0].SetRaw(VolumeMax)
	configSetVolume(1, VolumeFX) // A deliberately different startup value.
	for i, volume := range []uint32{0x4000, 163 * 255, 0} {
		def := defaultNativeSoundDef()
		def.enabled, def.volume, def.sampleIDs = true, volume, []string{"silence"}
		if i == 1 {
			def.sampleIDs = []string{"stereo"}
		}
		nativeAudioFX.defs[i+1] = def
		if !nativeAudioFX.play(sound.ID(i+1), 100) {
			t.Fatal("long native FX sample was not admitted")
		}
	}
	voices := nativeAudioFX.voices[:3]
	sources := make([]openal.Source, len(voices))
	events := make([]uint32, len(voices))
	for i, voice := range voices {
		sources[i] = openal.Source(*voice.GetSource())
		events[i] = voice.UserData().(uint32)
	}
	idle := nativeAudioFX.voices[3]
	idle.SetVolume(19)
	// An active sample outside the native FX pool must not be mixed or stopped,
	// even when its user data happens to be a DWORD too.
	foreign := driver.AllocateSample()
	foreign.Init()
	foreign.SetType(5, 0)
	foreign.SetADPCMBlockSize(entry.blockSize)
	foreign.SetPlaybackRate(int(entry.rate))
	foreign.SetVolume(23)
	foreign.SetUserData(uint32(12345))
	ready := foreign.BufferReady()
	if ready < 0 {
		t.Fatal("foreign sample has no ready buffer")
	}
	foreign.LoadBuffer(uint32(ready), entry.data)
	foreignSource := openal.Source(*foreign.GetSource())
	g := gui.New(nil)
	t.Cleanup(g.DestroyAll)
	previousClient := noxClient
	noxClient = &Client{Client: &client.Client{GUI: g}}
	t.Cleanup(func() { noxClient = previousClient })
	root := g.NewWindowRaw(nil, gui.StatusEnabled, 0, 0, 640, 480, nil)
	slider := g.NewWindowRaw(root, gui.StatusEnabled, 0, 0, 70, 12, nil)
	slider.SetID(351)
	checkbox := g.NewWindowRaw(root, gui.StatusEnabled, 0, 0, 70, 12, nil)
	checkbox.SetID(361)
	root.SetFunc94C(legacy.OptionsInGameProc4ADF30())
	if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(root.C()) <= math.MaxUint32 || uintptr(slider.C()) <= math.MaxUint32 || uintptr(checkbox.C()) <= math.MaxUint32) {
		t.Fatal("native C Options fixture did not exceed 4 GiB")
	}
	if startAudioServices() != 0 || audioTimer93944 == math.MaxUint32 {
		t.Fatal("production audio timer did not start")
	}
	audioTimer := audioTimer93944
	if startAudioServices() != 0 || audioTimer93944 != audioTimer {
		t.Fatal("starting audio services again replaced the live timer")
	}
	drive := func(label string, ready func() bool) {
		t.Helper()
		deadline := time.Now().Add(2 * time.Second)
		for time.Now().Before(deadline) {
			ail.Serve()
			if err := openal.Err(); err != nil {
				t.Fatalf("%s: OpenAL error: %v", label, err)
			}
			if ready() {
				return
			}
			time.Sleep(2 * time.Millisecond)
		}
		for _, source := range append(sources, foreignSource) {
			t.Logf("source=%d state=%v queued=%d processed=%d gain=%g", source, source.State(), source.BuffersQueued(), source.BuffersProcessed(), source.Getf(openal.AlGain))
		}
		t.Fatalf("%s: production timer did not apply live state: current=%d target=%d gains=%g/%g/%g enabled=%d playback=%+v", label, fx.Timers[0].Current>>16, fx.Timers[0].Target>>16, sources[0].Getf(openal.AlGain), sources[1].Getf(openal.AlGain), sources[2].Getf(openal.AlGain), legacy.Sub_453070(), driver.PlaybackStats())
	}
	drive("start real playback", func() bool {
		for _, source := range append(sources, foreignSource) {
			if source.State() != openal.Playing || source.BuffersQueued() == 0 {
				return false
			}
		}
		return true
	})
	checkGains := func(live int, first int) bool {
		for i := first; i < len(sources); i++ {
			mixed := uint64(events[i]) * uint64(live) / VolumeMax
			want := min(127, int(127*mixed>>14))
			if math.Abs(float64(sources[i].Getf(openal.AlGain))-float64(float32(want)/127)) > 1e-6 || sources[i].State() != openal.Playing {
				return false
			}
		}
		return true
	}
	for _, live := range []int{4850, 8126, 0, VolumeMax, 1, VolumeMax} {
		if live == 0 {
			// Test a zero live mix without toggling admission: C's zero slider
			// also toggles the checkbox, which is tested separately below.
			fx.Timers[0].SetRaw(0)
		} else if gui.EventRespInt(root.Func94(&gui.RawEvent{Event: 0x4009, Arg1: uintptr(slider.C()), Arg2: uintptr(live)})) != 0 {
			t.Fatal("actual C FX slider returned the wrong result")
		}
		drive("C slider/live mix", func() bool { return fx.Timers[0].Current == uint32(live)<<16 && checkGains(live, 0) })
		if configGetVolume(VolumeFX) != 1 {
			t.Fatal("live FX service rewrote the startup setting")
		}
	}
	fx.Timers[0].SetParams(1000, VolumeMax)
	fx.Timers[0].SetInterp(0)
	ticks = 250
	wantCurrent := uint32(VolumeMax)<<16 - uint32(ticks)*fx.Timers[0].DeltaPerTick
	drive("interpolated current, not target", func() bool { return fx.Timers[0].Current == wantCurrent && checkGains(int(wantCurrent>>16), 0) })
	ticks = 500
	wantCurrent = uint32(VolumeMax)<<16 - uint32(ticks)*fx.Timers[0].DeltaPerTick
	drive("second interpolation tick", func() bool { return fx.Timers[0].Current == wantCurrent && checkGains(int(wantCurrent>>16), 0) })
	voices[0].Init()
	voices[0].SetVolume(17)
	fx.Timers[0].SetRaw(4850)
	drive("stopped voice is not revived", func() bool { return fx.Timers[0].Current == 4850<<16 && checkGains(4850, 1) })
	if sources[0].State() == openal.Playing || sources[0].Getf(openal.AlGain) != float32(17)/127 {
		t.Fatal("live mix changed or restarted an explicitly stopped native sample")
	}
	toggle := func() {
		t.Helper()
		before := legacy.Sub_453070()
		if gui.EventRespInt(root.Func94(&gui.RawEvent{Event: 0x4007, Arg1: uintptr(checkbox.C()), Arg2: ^uintptr(0)})) != 1 || legacy.Sub_453070() != 1-before {
			t.Fatal("actual C FX checkbox did not toggle its live flag")
		}
	}
	toggle()
	drive("C mute stops and discards native voices", func() bool {
		for _, source := range sources {
			if source.State() == openal.Playing || source.BuffersQueued() != 0 {
				return false
			}
		}
		return true
	})
	toggle()
	// Let several real production ticks pass. Re-enabling admits future
	// events, but must not restart the sounds that mute discarded.
	until := time.Now().Add(150 * time.Millisecond)
	for time.Now().Before(until) {
		ail.Serve()
		time.Sleep(2 * time.Millisecond)
	}
	for i, voice := range voices {
		if voice.Status() == 4 || sources[i].BuffersQueued() != 0 || voice.UserData() != events[i] {
			t.Fatal("re-enabling revived a discarded FX voice or changed its original event metadata")
		}
	}
	drive("hardware-consumed buffer", func() bool { return driver.PlaybackStats().SampleBuffersProcessed != 0 })
	if foreignSource.State() != openal.Playing || foreignSource.Getf(openal.AlGain) != float32(23)/127 || foreign.UserData() != uint32(12345) || idle.Status() == 4 || openal.Source(*idle.GetSource()).Getf(openal.AlGain) != float32(19)/127 {
		t.Fatal("FX service modified a foreign or idle sample")
	}
	if !nativeAudioFX.play(sound.ID(1), 33) {
		t.Fatal("re-enabled FX did not admit a future sound")
	}
	reused := nativeAudioFX.voices[(nativeAudioFX.next+len(nativeAudioFX.voices)-1)%len(nativeAudioFX.voices)]
	if reused.UserData() != uint32(163*33) {
		t.Fatal("future sound did not replace the chosen voice's original event volume")
	}
	stats := driver.PlaybackStats()
	if !stats.Native || stats.SampleBuffersQueued == 0 || stats.SampleBuffersProcessed == 0 {
		t.Fatalf("actual decode/queue/drain was not observed: %+v", stats)
	}
	nativeAudioFX.close()
	until = time.Now().Add(150 * time.Millisecond)
	for time.Now().Before(until) {
		ail.Serve()
		time.Sleep(2 * time.Millisecond)
	}
	if after := driver.PlaybackStats(); after.SampleSources != 1 || foreign.Status() != 4 || foreignSource.Getf(openal.AlGain) != float32(23)/127 {
		t.Fatalf("servicing the closed FX pool touched the remaining foreign sample: %+v", after)
	}
	if err := openal.Err(); err != nil {
		t.Fatalf("OpenAL error after mute/re-enable/close: %v", err)
	}
	t.Logf("production timer=%#x renderer=%q ALSOFT_DRIVERS=%q native playback=%+v; mono/stereo real C slider/mute, zero recovery, interpolation, stopped/foreign/idle/closed scope and future-event admission checked", audioTimer, openal.GetString(0xB003), os.Getenv("ALSOFT_DRIVERS"), stats)
}
