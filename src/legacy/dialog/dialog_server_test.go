//go:build server

package dialog

import (
	"testing"

	"github.com/opennox/libs/strman"

	"github.com/opennox/opennox/v1/legacy/client/audio/ail"
	"github.com/opennox/opennox/v1/legacy/timer"
)

func TestDialogServerKeepsAudioDisabled(t *testing.T) {
	previousTicks := timer.PlatformTicks
	timer.PlatformTicks = func() uint64 { return 0 }
	t.Cleanup(func() { timer.PlatformTicks = previousTicks })
	if result := ail.Startup(); result != -1 {
		t.Fatalf("server audio startup = %d, want -1", result)
	}
	defer ail.Shutdown()
	if driver := ail.WaveOutOpen(); driver != 0 {
		t.Fatalf("server audio driver = %v, want 0", driver)
	}
	sm := strman.New()
	if err := sm.ReadJSON("testdata/dialogs.json"); err != nil {
		t.Fatal(err)
	}
	variant, ok := sm.GetVariantInFile("Con03B.scr:WorkerHurt", "C:\\NoxPost\\src\\client\\Audio\\AudDiag.c")
	if !ok || variant.Str2 != "empty" {
		t.Fatalf("test dialog variant = %v, found = %t", variant, ok)
	}
	var enabled, state, flag, fallback, initialized, volume uint32
	var driver ail.Driver
	var driverCalls int
	var resetWrites []uint32
	unexpected := func() { t.Fatal("disabled server audio invoked a client callback") }
	d := NewDialog(
		"testdata", &enabled, &state, &flag, &fallback, &initialized, &driver, &volume,
		func() *strman.StringManager { return sm },
		new(timer.TimerGroup), new(timer.TimerGroup), new(timer.TimerGroup), new(timer.TimerGroup),
		func() string { unexpected(); return "" },
		func() ail.Driver { driverCalls++; return ail.WaveOutOpen() },
		unexpected, unexpected,
		func() bool { unexpected(); return false },
		func(value uint32) { resetWrites = append(resetWrites, value) },
		func() uint32 { unexpected(); return 0 },
	)
	checkInactive := func(t *testing.T, wantInitialized bool) {
		t.Helper()
		if d.IsInitialized() != wantInitialized || enabled != 0 || driver != 0 || d.State() != 0 ||
			d.FileToRead() != "" || d.CurrentPlayingFile() != "" || d.GetStream() != 0 || d.Sub_44D930() {
			t.Fatalf("server dialog queued audio (initialized=%t, want %t): %#v", d.IsInitialized(), wantInitialized, d)
		}
	}
	t.Run("before-initialization", func(t *testing.T) { checkInactive(t, false) })
	t.Run("initialize", func(t *testing.T) {
		if result := d.Nox_xxx_WorkerHurt_44D810(); result != 1 {
			t.Fatalf("initialization = %d, want 1", result)
		}
		checkInactive(t, true)
	})
	t.Run("initialize-again", func(t *testing.T) {
		if result := d.Nox_xxx_WorkerHurt_44D810(); result != 1 {
			t.Fatalf("reinitialization = %d, want 1", result)
		}
		checkInactive(t, true)
	})
	t.Run("play-request", func(t *testing.T) {
		if result := d.PlayFile("empty", 100); result != 1 {
			t.Fatalf("play request = %d, want 1", result)
		}
		checkInactive(t, true)
	})
	t.Run("updates", func(t *testing.T) {
		for i := 0; i < 3; i++ {
			d.Sub_44D3A0()
			checkInactive(t, true)
		}
	})
	t.Run("shutdown", func(t *testing.T) {
		d.Sub_44D8C0()
		checkInactive(t, false)
	})
	if driverCalls != 1 || len(resetWrites) != 1 || resetWrites[0] != 0 {
		t.Fatalf("initialization was not idempotent: driver calls = %d, reset writes = %v", driverCalls, resetWrites)
	}
}
