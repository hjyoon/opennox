//go:build !server

package opennox

import (
	"context"
	"encoding/binary"
	"math"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"
	"time"
	"unsafe"

	"github.com/opennox/libs/env"
	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/timshannon/go-openal/openal"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/client/audio/ail"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/legacy/timer"
)

const nativeNetworkAudioChild = "NOX_TEST_NATIVE_FX_NETWORK_CHILD"

// Exercise the public packet switch and a real OpenAL source in a separate
// process. The generated silent ADPCM and null device do not require stock
// assets, a GUI, or the user's audio configuration.
func TestNativeNetworkAudioPlayback(t *testing.T) {
	if os.Getenv("NOX_TEST_NATIVE_FX") != "true" {
		t.Skip("set NOX_TEST_NATIVE_FX=true to test network FX on real OpenAL")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestNativeNetworkAudioPlaybackChild$", "-test.count=1", "-test.v")
	cmd.Dir = t.TempDir()
	for _, kv := range os.Environ() {
		key, _, _ := strings.Cut(kv, "=")
		if strings.HasPrefix(key, "NOX_TEST_NATIVE_FX_") || strings.HasPrefix(key, "NOX_E2E") || key == "NOX_DATA" {
			continue
		}
		cmd.Env = append(cmd.Env, kv)
	}
	cmd.Env = append(cmd.Env, nativeNetworkAudioChild+"=true")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("native network audio subprocess failed: %v\n%s", err, out)
	}
	t.Logf("%s", out)
}

func TestNativeNetworkAudioPlaybackChild(t *testing.T) {
	if os.Getenv(nativeNetworkAudioChild) != "true" {
		t.Skip("only run in the isolated native network audio subprocess")
	}
	if env.IsE2E() {
		t.Fatal("network audio test must not use E2E mock audio")
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
		t.Fatal("cannot open the configured OpenAL device")
	}
	t.Cleanup(func() {
		if err := driver.Close(); err != nil {
			t.Errorf("close network audio driver: %v", err)
		}
	})
	voice := driver.AllocateSample()
	if voice == 0 || voice.GetSource() == nil {
		t.Fatal("network audio sample has no real OpenAL source")
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(driver) <= math.MaxUint32 || uintptr(voice) <= math.MaxUint32) {
		t.Fatal("native audio handles did not exceed 4 GiB")
	}
	entry := &nativeAudioBankEntry{name: "silence", rate: 22050, flags: 8, blockSize: 256, data: make([]byte, 256)}
	nativeAudioFX.bank = &nativeAudioBank{entries: map[string]*nativeAudioBankEntry{"silence": entry}}
	nativeAudioFX.voices = []ail.Sample{voice}
	def := defaultNativeSoundDef()
	def.enabled, def.sampleIDs = true, []string{"silence"}
	nativeAudioFX.defs[sound.SoundRunOnStone] = def
	t.Cleanup(nativeAudioFX.close)
	source := openal.Source(*voice.GetSource())
	t.Logf("network audio driver=%#x sample=%#x source=%d renderer=%q", driver, voice, source, openal.GetString(0xB003))
	fx := (*timer.TimerGroup)(legacy.Get_dword_587000_127004())
	if fx == nil || legacy.Sub_453070() != 1 {
		t.Fatal("FX group and enabled flag were not initialized")
	}
	*memmap.PtrUint32(0x5D4594, 815764) = 1
	c := &Client{Client: &client.Client{}}
	var frame uint32
	var frame16 uint16
	for _, op := range []netmsg.Op{netmsg.MSG_AUDIO_EVENT, netmsg.MSG_AUDIO_PLAYER_EVENT} {
		*fx = timer.TimerGroup{}
		fx.Timers[0].SetRaw(VolumeMax)
		voice.SetVolume(17)
		voice.SetUserData(nil)
		packet := []byte{byte(op), 0, 0, 0}
		binary.LittleEndian.PutUint16(packet[2:], uint16(sound.SoundRunOnStone)|uint16(50)<<10)
		before := driver.PlaybackStats().SampleBuffersQueued
		processed := driver.PlaybackStats().SampleBuffersProcessed
		if got := c.nox_xxx_netOnPacketRecvCli48EA70_switch(0, op, packet, &frame, &frame16); got != 4 {
			t.Fatalf("%s: consumed %d bytes, want 4", op, got)
		}
		wantEvent := uint32(16300)
		wantGain := float32(126) / 127
		if event, ok := voice.UserData().(uint32); !ok || event != wantEvent {
			t.Fatalf("%s: valid network FX did not reach native sample submission: saved volume=%v (%T), want %d; queued=%d -> %d",
				op, voice.UserData(), voice.UserData(), wantEvent, before, driver.PlaybackStats().SampleBuffersQueued)
		}
		if got := source.Getf(openal.AlGain); math.Abs(float64(got-wantGain)) > 1e-6 {
			t.Fatalf("%s: actual network source gain=%g, want %g", op, got, wantGain)
		}
		waitNativeNetworkAudioBuffer48EA70(t, driver, before+1, processed+1)
		if got := driver.PlaybackStats().SampleBuffersQueued; got != before+1 {
			t.Fatalf("%s: actual OpenAL queued buffers=%d, want %d", op, got, before+1)
		}
		if err := openal.Err(); err != nil {
			t.Fatalf("%s: OpenAL error: %v", op, err)
		}
	}
	checkNativeNetworkAudioCases48EA70(t, c, driver, voice, entry, fx)
}

func checkNativeNetworkAudioCases48EA70(t *testing.T, c *Client, driver ail.Driver, voice ail.Sample, entry *nativeAudioBankEntry, fx *timer.TimerGroup) {
	t.Helper()
	source := openal.Source(*voice.GetSource())
	g := gui.New(nil)
	t.Cleanup(g.DestroyAll)
	root := g.NewWindowRaw(nil, gui.StatusEnabled, 0, 0, 640, 480, nil)
	checkbox := g.NewWindowRaw(root, gui.StatusEnabled, 0, 0, 70, 12, nil)
	checkbox.SetID(361)
	root.SetFunc94C(legacy.OptionsInGameProc4ADF30())
	setEnabled := func(enabled bool) {
		t.Helper()
		if (legacy.Sub_453070() != 0) == enabled {
			return
		}
		response := root.Func94(&gui.RawEvent{Event: 0x4007, Arg1: uintptr(checkbox.C()), Arg2: ^uintptr(0)})
		if gui.EventRespInt(response) != 1 || (legacy.Sub_453070() != 0) != enabled {
			t.Fatal("actual Options checkbox did not set the live FX enabled flag")
		}
	}
	configSetVolume(0, VolumeFX)
	var frame uint32 = 0x12345678
	var frame16 uint16 = 0x9876
	count := 0
	for _, op := range []netmsg.Op{netmsg.MSG_AUDIO_EVENT, netmsg.MSG_AUDIO_PLAYER_EVENT} {
		for _, flags := range []uint32{8, 9} {
			entry.flags = flags
			for pan := -128; pan <= 127; pan++ {
				requested := []int{2, 32, 80, 100, 126}[(pan+128)%5]
				live := []int{0, 1, 4850, 8126, VolumeMax}[(pan+129)%5]
				defVolume := []uint32{0x4000, 163 * 80, 163 * 255, 1}[(pan+128)%4]
				nativeAudioFX.defs[sound.SoundRunOnStone].volume = defVolume
				*fx = timer.TimerGroup{}
				fx.Timers[0].SetRaw(uint32(live))
				voice.SetVolume(17)
				voice.SetPan(17)
				voice.SetUserData(nil)
				packet := []byte{byte(op), byte(pan), 0, 0, 0xa5, 0x5a}
				binary.LittleEndian.PutUint16(packet[2:4], uint16(sound.SoundRunOnStone)|uint16(requested/2)<<10)
				if got := c.nox_xxx_netOnPacketRecvCli48EA70_switch(0, op, packet, &frame, &frame16); got != 4 {
					t.Fatalf("%s pan=%d: consumed %d bytes, want 4", op, pan, got)
				}
				clampedVolume := requested
				if clampedVolume > 100 {
					clampedVolume = 100
				}
				event := uint32(uint64(163*clampedVolume) * uint64(defVolume) >> 14)
				mixed := uint64(event) * uint64(live) / VolumeMax
				gain := int(127 * mixed >> 14)
				if gain > 127 {
					gain = 127
				}
				wantGain := float32(gain) / 127
				if got := source.Getf(openal.AlGain); math.Abs(float64(got-wantGain)) > 1e-6 {
					t.Fatalf("%s flags=%d pan=%d live=%d requested=%d def=%d: actual gain=%g, want %g", op, flags, pan, live, requested, defVolume, got, wantGain)
				}
				clampedPan := math.Max(-50, math.Min(50, float64(pan)))
				samplePan := (127 * (int(math.Trunc(clampedPan*8192/50)) + 8192)) >> 14
				wantX := float32(samplePan-63) / 64
				wantZ := float32(math.Sqrt(float64(1 - wantX*wantX)))
				x, y, z := source.Get3f(openal.AlPosition)
				if x != wantX || y != 0 || z != wantZ {
					t.Fatalf("%s flags=%d signed pan=%d: actual OpenAL position=(%g,%g,%g), want=(%g,0,%g)", op, flags, pan, x, y, z, wantX, wantZ)
				}
				if got, ok := voice.UserData().(uint32); !ok || got != event || voice.BufferReady() != 1 || voice.Status() != 4 {
					t.Fatalf("%s flags=%d pan=%d: wrong event metadata or no submitted buffer: saved=%v want=%d stats=%+v", op, flags, pan, voice.UserData(), event, driver.PlaybackStats())
				}
				if frame != 0x12345678 || frame16 != 0x9876 || configGetVolume(VolumeFX) != 0 {
					t.Fatal("audio packet changed timestamps or startup FX configuration")
				}
				if err := openal.Err(); err != nil {
					t.Fatalf("%s flags=%d pan=%d: OpenAL error: %v", op, flags, pan, err)
				}
				count++
			}
		}
	}
	// The 20 Hz driver serves pending sample buffers, not LoadBuffer itself.
	// Above, inspect each sample before the single-voice pool can be reused;
	// here drain the last matrix event through the actual driver. Both opcode
	// paths were independently queued and drained before this gain/pan matrix.
	before := driver.PlaybackStats()
	waitNativeNetworkAudioBuffer48EA70(t, driver, before.SampleBuffersQueued+1, before.SampleBuffersProcessed+1)
	// Guarded packets still consume exactly four bytes, with no sample/timer
	// submission. Empty and truncated public-switch inputs must also be safe.
	for _, op := range []netmsg.Op{netmsg.MSG_AUDIO_EVENT, netmsg.MSG_AUDIO_PLAYER_EVENT} {
		for _, mode := range []string{"disconnected", "disabled", "zero-volume", "zero-id", "undefined-id"} {
			*memmap.PtrUint32(0x5D4594, 815764) = 1
			setEnabled(true)
			*fx = timer.TimerGroup{}
			fx.Timers[0].SetRaw(VolumeMax)
			id, volume := sound.SoundRunOnStone, 100
			switch mode {
			case "disconnected":
				*memmap.PtrUint32(0x5D4594, 815764) = 0
			case "disabled":
				setEnabled(false)
			case "zero-volume":
				volume = 0
			case "zero-id":
				id = 0
			case "undefined-id":
				id = 1023
			}
			packet := []byte{byte(op), 50, 0, 0}
			binary.LittleEndian.PutUint16(packet[2:], uint16(id)|uint16(volume/2)<<10)
			voice.SetVolume(17)
			voice.SetPan(17)
			voice.SetUserData("sentinel")
			before := driver.PlaybackStats().SampleBuffersQueued
			if got := c.nox_xxx_netOnPacketRecvCli48EA70_switch(0, op, packet, &frame, &frame16); got != 4 {
				t.Fatalf("%s %s: consumed=%d, want 4", op, mode, got)
			}
			x, _, _ := source.Get3f(openal.AlPosition)
			if source.Getf(openal.AlGain) != float32(17)/127 || x != float32(17-63)/64 || voice.UserData() != "sentinel" || driver.PlaybackStats().SampleBuffersQueued != before {
				t.Fatalf("%s %s changed the real native source or submitted a buffer", op, mode)
			}
		}
		*memmap.PtrUint32(0x5D4594, 815764) = 1
		setEnabled(true)
		for size := 0; size < 4; size++ {
			want := -1
			if size == 0 {
				want = 0 // Public packet switch's existing empty-input contract.
			}
			if got := c.nox_xxx_netOnPacketRecvCli48EA70_switch(0, op, make([]byte, size), &frame, &frame16); got != want {
				t.Fatalf("%s short size=%d: consumed=%d, want %d", op, size, got, want)
			}
		}
	}
	// A spatial event must not leave a subsequent UI sound panned to a side.
	nativeAudioFX.defs[sound.SoundShellClick] = nativeAudioFX.defs[sound.SoundRunOnStone]
	*fx = timer.TimerGroup{}
	fx.Timers[0].SetRaw(VolumeMax)
	if !nativeAudioFX.play(sound.SoundShellClick, 100) {
		t.Fatal("centered UI effect was rejected after spatial playback")
	}
	x, y, z := source.Get3f(openal.AlPosition)
	if x != 0 || y != 0 || z != 1 {
		t.Fatalf("UI effect retained network pan: (%g,%g,%g)", x, y, z)
	}
	before = driver.PlaybackStats()
	waitNativeNetworkAudioBuffer48EA70(t, driver, before.SampleBuffersQueued+1, before.SampleBuffersProcessed+1)
	t.Logf("verified %d public A6/A7 mono/stereo submissions over all signed pan bytes, live/definition/wire gains, admission guards, timestamps, and centered UI reuse", count)
}

func waitNativeNetworkAudioBuffer48EA70(t *testing.T, driver ail.Driver, queued, processed uint64) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		ail.Serve()
		stats := driver.PlaybackStats()
		if stats.SampleBuffersQueued >= queued && stats.SampleBuffersProcessed >= processed {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("real OpenAL did not queue and process network FX buffers: want queued>=%d processed>=%d actual=%+v", queued, processed, driver.PlaybackStats())
}
