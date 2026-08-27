package legacy

import (
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/client/audio/ail"
)

func TestAudioDriverSlotsPreserveNativeWidth(t *testing.T) {
	oldMusic := musicAudioDriver
	oldDialog := dialogAudioDriver
	t.Cleanup(func() {
		setMusicAudioDriver(oldMusic)
		setDialogAudioDriver(oldDialog)
	})

	// Keep the source value nonconstant so this test also compiles on 32-bit
	// targets. The conversion intentionally narrows there and must not narrow on
	// 64-bit targets.
	raw := uint64(0x6c94902d800006)
	drv := ail.Driver(uintptr(raw))
	setMusicAudioDriver(drv)
	setDialogAudioDriver(drv)

	if got := Get_dword_5d4594_816376(); got != drv {
		t.Fatalf("music driver = %#x, want %#x", got, drv)
	}
	if got := dialogAudioDriver; got != drv {
		t.Fatalf("dialog driver = %#x, want %#x", got, drv)
	}

	wantMirror := uint32(drv)
	if unsafe.Sizeof(uintptr(0)) > unsafe.Sizeof(uint32(0)) {
		wantMirror = 1
	}
	if got := musicAudioDriverLegacyMirror(); got != wantMirror {
		t.Fatalf("music C mirror = %#x, want %#x", got, wantMirror)
	}
	if got := dialogAudioDriverLegacyMirror(); got != wantMirror {
		t.Fatalf("dialog C mirror = %#x, want %#x", got, wantMirror)
	}

	setMusicAudioDriver(0)
	setDialogAudioDriver(0)
	if got := musicAudioDriverLegacyMirror(); got != 0 {
		t.Fatalf("disabled music C mirror = %#x, want 0", got)
	}
	if got := dialogAudioDriverLegacyMirror(); got != 0 {
		t.Fatalf("disabled dialog C mirror = %#x, want 0", got)
	}
}
