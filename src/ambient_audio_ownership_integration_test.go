//go:build !server

package opennox

import (
	"image"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/client/audio/ail"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/legacy/timer"
)

func TestNativeAmbientVoiceOwnership45A9B0(t *testing.T) {
	const child = "NOX_TEST_AMBIENT_OWNERSHIP_CHILD"
	if os.Getenv(child) != "true" {
		if os.Getenv("NOX_TEST_NATIVE_FX") != "true" {
			t.Skip("set NOX_TEST_NATIVE_FX=true for actual native ambient ownership")
		}
		cmd := exec.Command(os.Args[0], "-test.run=^TestNativeAmbientVoiceOwnership45A9B0$", "-test.v")
		cmd.Dir = t.TempDir()
		for _, kv := range os.Environ() {
			k, _, _ := strings.Cut(kv, "=")
			if strings.HasPrefix(k, "NOX_E2E") || k == child || k == "NOX_DATA" {
				continue
			}
			cmd.Env = append(cmd.Env, kv)
		}
		cmd.Env = append(cmd.Env, child+"=true")
		out, err := cmd.CombinedOutput()
		t.Logf("%s", out)
		if err != nil {
			t.Fatal(err)
		}
		return
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	legacy.InitBlobData()
	handles.Init()
	defer handles.Release()
	driver := ail.WaveOutOpen()
	if driver == 0 {
		t.Fatal("native device")
	}
	defer driver.Close()
	defer nativeAudioFX.close()
	entry := &nativeAudioBankEntry{name: "owned", rate: 1000, flags: 8, blockSize: 256, data: make([]byte, 256*256)}
	nativeAudioFX.bank = &nativeAudioBank{entries: map[string]*nativeAudioBankEntry{"owned": entry}}
	for i := 0; i < 2; i++ {
		voice := driver.AllocateSample()
		if voice == 0 {
			t.Fatal("voice")
		}
		nativeAudioFX.voices = append(nativeAudioFX.voices, voice)
	}
	nativeAudioFX.resetDefs()
	def := defaultNativeSoundDef()
	def.enabled, def.behavior, def.sampleIDs = true, 7, []string{"owned"}
	for _, id := range []sound.ID{sound.SoundLoopAmbBrook2, sound.SoundShellClick, sound.SoundShellSelect} {
		nativeAudioFX.defs[id] = def
	}
	fx := (*timer.TimerGroup)(legacy.Get_dword_587000_127004())
	*fx = timer.TimerGroup{}
	fx.Timers[0].SetRaw(VolumeMax)
	listener := plasmaNativeAllocation531580[client.Drawable](t)
	a, b := plasmaNativeAllocation531580[client.Drawable](t), plasmaNativeAllocation531580[client.Drawable](t)
	for _, dr := range []*client.Drawable{a, b} {
		dr.AudioLoop = uint32(sound.SoundLoopAmbBrook2)
		dr.ObjFlags = object.Flags(0x1000000)
	}
	update := func(dr *client.Drawable) { nativeAudioFX.updateDrawableLoop(dr, listener, 0, 0, 640) }
	update(a)
	owner := nativeAudioFX.ambient[a]
	if owner == nil || owner.event == nil {
		t.Fatal("initial ambient owner")
	}
	stolen := owner.event.voices[0]
	if !nativeAudioFX.play(sound.SoundShellClick, 100) || !nativeAudioFX.play(sound.SoundShellSelect, 100) {
		t.Fatal("gameplay voice stealing")
	}
	if len(owner.event.voices) != 0 || stolen.Status() != 4 {
		t.Fatal("stolen voice retains ambient ownership")
	}
	nativeAudioFX.stopDrawableLoop(a)
	if stolen.Status() != 4 {
		t.Fatal("deleting old owner stopped the replacement gameplay effect")
	}
	for _, v := range nativeAudioFX.voices {
		v.Init()
	}
	nativeAudioFX.events = nil
	nativeAudioFX.next = 0
	nativeAudioFX.defs[sound.SoundLoopAmbBrook2].field14 = 1
	a.PosVec = image.Pt(400, 0)
	update(a)
	old := nativeAudioFX.ambient[a].event
	update(b)
	current := nativeAudioFX.ambient[b].event
	if current == nil || current == old {
		t.Fatal("louder ambient did not displace old same-sound event")
	}
	// Reuse the freed sample for a gameplay event before unlinking the stale
	// old owner. It must never stop or retune that new sample generation.
	if !nativeAudioFX.play(sound.SoundShellClick, 100) {
		t.Fatal("replacement gameplay event")
	}
	nativeAudioFX.stopDrawableLoop(a)
	for _, v := range nativeAudioFX.voices {
		if v.Status() != 4 {
			t.Fatal("stale owner stopped a current voice")
		}
	}
	nativeAudioFX.resetDefs()
	if len(nativeAudioFX.ambient) != 0 || len(current.voices) != 1 || current.voices[0].Status() == 4 {
		t.Fatal("definition reload retained positional playback")
	}
	if nativeAudioFX.voices[0].Status() != 4 {
		t.Fatal("definition reload stopped unrelated gameplay audio")
	}
	nativeAudioFX.defs[sound.SoundLoopAmbBrook2] = def
	update(b)
	if nativeAudioFX.ambient[b] == nil {
		t.Fatal("reloaded bank has no owner")
	}
	legacy.Sub_453050()
	update(b)
	if len(nativeAudioFX.ambient) != 0 {
		t.Fatal("FX mute retained ambient owner")
	}
	nativeAudioFX.close()
	if len(nativeAudioFX.events) != 0 || len(nativeAudioFX.ambient) != 0 || driver.PlaybackStats().SampleSources != 0 {
		t.Fatal("closed driver retained voices/events/owners")
	}
	t.Logf("actual native pool: gameplay stealing, same-sound replacement, stale-owner deletion, definition reload, mute and close verified")
}
