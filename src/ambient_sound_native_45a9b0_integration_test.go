//go:build !server

package opennox

import (
	"context"
	"image"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/client/audio/ail"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/legacy/timer"
	"github.com/timshannon/go-openal/openal"
)

// The same C entry, native C-owned drawables and real OpenAL null device are
// exercised before and after the fix. Generated ADPCM keeps assets immutable.
func TestNativeAmbientPublicCPlayback45A9B0(t *testing.T) {
	const child = "NOX_TEST_AMBIENT_PUBLIC_C_CHILD"
	if os.Getenv(child) != "true" {
		if os.Getenv("NOX_TEST_NATIVE_FX") != "true" {
			t.Skip("set NOX_TEST_NATIVE_FX=true for actual ambient OpenAL playback")
		}
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestNativeAmbientPublicCPlayback45A9B0$", "-test.v")
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
			t.Fatalf("actual ambient C entry: %v", err)
		}
		return
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	legacy.InitBlobData()
	handles.Init()
	defer handles.Release()
	srv := NewServer(nil, nil, strman.New())
	defer srv.Close()
	c := &Client{Client: client.NewClient(nil, nil, srv.Server), srv: srv}
	defer c.Client.Close()
	oldClient, oldGet, oldWidth := noxClient, legacy.GetClient, nox_win_width
	noxClient, legacy.GetClient, nox_win_width = c, func() legacy.Client { return c }, 640
	defer func() { noxClient, legacy.GetClient, nox_win_width = oldClient, oldGet, oldWidth }()
	*c.Viewport() = noxrender.Viewport{Screen: image.Rect(0, 0, 640, 480), World: image.Rect(0, 0, 640, 480), Size: image.Pt(640, 480)}
	driver := ail.WaveOutOpen()
	if driver == 0 {
		t.Fatal("real OpenAL device unavailable")
	}
	defer driver.Close()
	defer nativeAudioFX.close()
	nativeAudioFX.resetDefs()
	entry := &nativeAudioBankEntry{name: "ambient", rate: 22050, flags: 8, blockSize: 256, data: make([]byte, 256*8)}
	// A nonzero predictor proves an audible PCM signal was submitted.
	for i := 0; i < len(entry.data); i += 256 {
		entry.data[i], entry.data[i+1] = 0xe8, 3
	}
	nativeAudioFX.bank = &nativeAudioBank{entries: map[string]*nativeAudioBankEntry{"ambient": entry}}
	for i := 0; i < 2; i++ {
		v := driver.AllocateSample()
		if v == 0 {
			t.Fatal("native voice")
		}
		nativeAudioFX.voices = append(nativeAudioFX.voices, v)
	}
	def := defaultNativeSoundDef()
	def.enabled, def.behavior, def.sampleIDs = true, 7, []string{"ambient"}
	nativeAudioFX.defs[sound.SoundLoopAmbBrook2] = def
	fx := (*timer.TimerGroup)(legacy.Get_dword_587000_127004())
	*fx = timer.TimerGroup{}
	fx.Timers[0].SetRaw(VolumeMax)
	source, listener := plasmaNativeAllocation531580[client.Drawable](t), plasmaNativeAllocation531580[client.Drawable](t)
	source.AudioLoop = uint32(sound.SoundLoopAmbBrook2)
	source.ObjFlags = object.Flags(0x1000000)
	source.PosVec, listener.PosVec = image.Pt(320, 200), image.Pt(320, 200)
	c.Objs.List8Add(source)
	legacy.Sub_45A9B0(source, listener)
	until := time.Now().Add(250 * time.Millisecond)
	for time.Now().Before(until) {
		ail.Serve()
		if driver.PlaybackStats().SampleBuffersQueued != 0 {
			break
		}
		time.Sleep(2 * time.Millisecond)
	}
	if stats := driver.PlaybackStats(); !stats.Native || stats.SampleBuffersQueued == 0 {
		t.Fatalf("audible nearby Amb emitter submitted no native buffers: %+v", stats)
	}
	var voice ail.Sample
	for _, v := range nativeAudioFX.voices {
		if v.Status() == 4 {
			voice = v
			break
		}
	}
	if voice == 0 {
		t.Fatal("no owned playing ambient voice")
	}
	al := openal.Source(*voice.GetSource())
	initial := driver.PlaybackStats().SampleBuffersQueued
	initialGain := al.Getf(openal.AlGain)
	source.PosVec = image.Pt(160, 200)
	legacy.Sub_45A9B0(source, listener)
	if voice.Status() != 4 || driver.PlaybackStats().SampleBuffersQueued != initial {
		t.Fatal("moving an emitter restarted its buffer")
	}
	if gain := al.Getf(openal.AlGain); gain <= 0 || gain >= initialGain {
		t.Fatal("moving emitter did not update distance gain")
	}
	until = time.Now().Add(550 * time.Millisecond)
	for time.Now().Before(until) {
		legacy.Sub_45A9B0(source, listener)
		ail.Serve()
		time.Sleep(2 * time.Millisecond)
	}
	if driver.PlaybackStats().SampleBuffersQueued <= initial+8 {
		t.Fatal("ambient sample did not repeat after natural completion")
	}
	source.PosVec = image.Pt(5000, 5000)
	legacy.Sub_45A9B0(source, listener)
	if voice.Status() == 4 || al.BuffersQueued() != 0 {
		t.Fatal("out-of-range emitter retained playback or queued hardware buffers")
	}
	source.PosVec = image.Pt(320, 200)
	legacy.Sub_45A9B0(source, listener)
	ail.Serve()
	c.sub_459F70(source)
	for _, v := range nativeAudioFX.voices {
		if v.Status() == 4 {
			t.Fatal("drawable unlink retained an ambient voice")
		}
	}
	if source.Field_124 != nil {
		t.Fatal("native loop stored a channel pointer in the old PE32 handle")
	}
	t.Logf("native ambient C entry=%p listener=%p actual OpenAL playback=%+v; moving, natural repeat, distance stop and drawable unlink verified", source, listener, driver.PlaybackStats())
}
