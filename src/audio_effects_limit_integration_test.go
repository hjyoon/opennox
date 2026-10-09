//go:build !server

package opennox

import (
	"context"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/opennox/libs/env"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/client/audio/ail"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/legacy/timer"
)

const nativeFXLimitChild = "NOX_TEST_NATIVE_FX_LIMIT_CHILD"

func TestNativeAudioEffectsEventLimit(t *testing.T) {
	if os.Getenv("NOX_TEST_NATIVE_FX") != "true" {
		t.Skip("set NOX_TEST_NATIVE_FX=true to test real OpenAL event limits")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestNativeAudioEffectsEventLimitChild$", "-test.count=1", "-test.v")
	cmd.Dir = t.TempDir()
	for _, kv := range os.Environ() {
		key, _, _ := strings.Cut(kv, "=")
		if strings.HasPrefix(key, "NOX_TEST_NATIVE_FX_") || strings.HasPrefix(key, "NOX_E2E") || key == "NOX_DATA" {
			continue
		}
		cmd.Env = append(cmd.Env, kv)
	}
	cmd.Env = append(cmd.Env, nativeFXLimitChild+"=true")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("native FX limit subprocess failed: %v\n%s", err, out)
	}
	t.Logf("%s", out)
}

// Generated silent ADPCM and real OpenAL sources, never mock audio. The stock
// male heavy-exertion AUD record has behavior=2, volume=30 and maximum=1.
// GAME.EXE 00451BE0 keeps an equally loud active event unless behavior bit
// 0x10 requests replacement. Server movement must keep sending its events.
func TestNativeAudioEffectsEventLimitChild(t *testing.T) {
	if os.Getenv(nativeFXLimitChild) != "true" {
		t.Skip("only run in the isolated native FX limit subprocess")
	}
	if env.IsE2E() {
		t.Fatal("native event limit test must not use mock audio")
	}
	runtime.LockOSThread()
	t.Cleanup(runtime.UnlockOSThread)
	legacy.InitBlobData()
	previousTicks := timer.PlatformTicks
	timer.PlatformTicks = func() uint64 { return 0 }
	t.Cleanup(func() { timer.PlatformTicks = previousTicks })
	handles.Init()
	t.Cleanup(handles.Release)
	driver := ail.WaveOutOpen()
	if driver == 0 {
		t.Fatal("cannot open native OpenAL playback device")
	}
	t.Cleanup(func() { _ = driver.Close() })
	fx := (*timer.TimerGroup)(legacy.Get_dword_587000_127004())
	*fx = timer.TimerGroup{}
	fx.Timers[0].SetRaw(VolumeMax)
	if legacy.Sub_453070() == 0 {
		t.Fatal("fresh FX gate is disabled")
	}
	fixture := func(t *testing.T, limit uint8, behavior uint8, multi bool) *nativeAudioEffectsState {
		t.Helper()
		entry := &nativeAudioBankEntry{name: "long", rate: 1000, flags: 8, blockSize: 256, data: make([]byte, 256)}
		s := &nativeAudioEffectsState{bank: &nativeAudioBank{entries: map[string]*nativeAudioBankEntry{"long": entry}}}
		for i := 0; i < 4; i++ {
			voice := driver.AllocateSample()
			if voice == 0 || voice.GetSource() == nil {
				t.Fatal("fixture has no real OpenAL voice")
			}
			s.voices = append(s.voices, voice)
		}
		t.Cleanup(s.close)
		def := defaultNativeSoundDef()
		def.enabled, def.behavior, def.field14, def.volume = true, behavior, limit, 163*30
		def.sampleIDs = []string{"long"}
		if multi {
			def.sampleIDs = []string{"long", "long"}
		}
		s.defs[sound.SoundHumanMaleExertionHeavy] = def
		s.defs[sound.SoundShellClick] = def
		return s
	}
	queued := func() uint64 {
		// LoadBuffer marks a sample ready; the unchanged 20 Hz AIL service
		// must submit it before native hardware counters can be asserted.
		time.Sleep(60 * time.Millisecond)
		ail.Serve()
		return driver.PlaybackStats().SampleBuffersQueued
	}
	const heavy = sound.SoundHumanMaleExertionHeavy
	t.Run("overweight-repeated-movement", func(t *testing.T) {
		s := fixture(t, 1, 2, false)
		before := queued()
		if !s.playPanned(heavy, 100, -40) || s.voices[0].Status() != 4 {
			t.Fatal("first exertion did not reach a real active source")
		}
		for i := 0; i < 120; i++ {
			if s.playPanned(heavy, 100, 40) {
				t.Fatalf("movement event %d restarted/overlapped the active exertion", i+1)
			}
		}
		if got := queued() - before; got != 1 {
			t.Fatalf("repeated movement submitted %d buffers, want one", got)
		}
		if !s.play(sound.SoundShellClick, 100) {
			t.Fatal("exertion limit suppressed an unrelated sound ID")
		}
		s.voices[0].Init() // End the sample, then a later movement may play again.
		if !s.play(heavy, 100) || queued()-before != 3 {
			t.Fatal("completed exertion retained its event slot")
		}
	})
	t.Run("louder-event-and-replacement-flag", func(t *testing.T) {
		s := fixture(t, 1, 2, false)
		before := queued()
		if !s.play(heavy, 25) || queued()-before != 1 || !s.play(heavy, 100) ||
			queued()-before != 2 || s.play(heavy, 25) || s.play(heavy, 95) {
			t.Fatal("original louder-first/equal-active ordering was not retained")
		}
		s.defs[heavy].behavior |= 0x10
		if !s.play(heavy, 100) || queued()-before != 3 {
			t.Fatal("explicit newest-event behavior did not replace the old event")
		}
	})
	t.Run("event-limit-not-sample-limit", func(t *testing.T) {
		s := fixture(t, 1, 4, true)
		before := queued()
		if !s.play(heavy, 100) || queued()-before != 2 || s.play(heavy, 100) {
			t.Fatal("multi-sample sound must count as one event")
		}
		s.voices[0].Init()
		if s.play(heavy, 100) {
			t.Fatal("partially completed event was counted as finished")
		}
		s.voices[1].Init()
		if !s.play(heavy, 100) || queued()-before != 4 {
			t.Fatal("fully completed multi-sample event did not release its slot")
		}
	})
	t.Run("unlimited-and-stolen-voice", func(t *testing.T) {
		s := fixture(t, 1, 2, false)
		s.defs[sound.SoundShellClick].field14 = 0
		if !s.play(heavy, 100) {
			t.Fatal("initial limited event rejected")
		}
		for i := 0; i < 4; i++ {
			if !s.play(sound.SoundShellClick, 100) {
				t.Fatal("zero limit should preserve voice-pool stealing")
			}
		}
		if !s.play(heavy, 100) {
			t.Fatal("stolen voice retained the previous sound ID")
		}
	})
}
