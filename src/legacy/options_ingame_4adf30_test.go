package legacy

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/common/sound"
)

// These are unit-test callbacks, not E2E input. The headless audit must still
// reach this procedure through the stock Options.wnd and actual mouse events.
func TestOptionsInGame4ADF30NativeExtensionDispatch(t *testing.T) {
	g := gui.New(nil)
	t.Cleanup(g.DestroyAll)
	root := g.NewWindowRaw(nil, gui.StatusEnabled, 0, 0, 640, 480, nil)
	control := g.NewWindowRaw(root, gui.StatusEnabled, 0, 0, 70, 12, nil)
	root.SetFunc94C(OptionsInGameProc4ADF30())
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, win := range []*gui.Window{root, control} {
			if uintptr(win.C()) <= math.MaxUint32 {
				t.Fatalf("native window %p did not exceed 4 GiB", win)
			}
		}
	}

	oldExtension, oldSound := Nox_gui_menu_proc_ext, ClientPlaySoundSpecial
	t.Cleanup(func() {
		Nox_gui_menu_proc_ext, ClientPlaySoundSpecial = oldExtension, oldSound
	})
	var ids []int
	var sounds []sound.ID
	var result int32
	Nox_gui_menu_proc_ext = func(id int) int {
		ids = append(ids, id)
		return int(result)
	}
	ClientPlaySoundSpecial = func(id sound.ID, volume int) {
		if volume != 100 {
			t.Fatalf("sound volume=%d, want 100", volume)
		}
		sounds = append(sounds, id)
	}
	click := func() int {
		return gui.EventRespInt(root.Func94(&gui.RawEvent{
			Event: 0x4007, Arg1: uintptr(control.C()), Arg2: ^uintptr(0),
		}))
	}
	for id := 380; id <= 389; id++ {
		control.SetID(uint(id))
		for _, value := range []int32{math.MinInt32, -1, 0, 1, math.MaxInt32} {
			result = value
			ids, sounds = nil, nil
			if got := click(); got != int(value) {
				t.Fatalf("id=%d: C return=%d, want signed DWORD %d", id, got, value)
			}
			if len(ids) != 1 || ids[0] != id || len(sounds) != 0 {
				t.Fatalf("id=%d: extension=%v sounds=%v, want one delegate and no duplicate sound", id, ids, sounds)
			}
		}
	}
	// These old controls are outside the PE32 in-game click switch's actions.
	// Keep its default sound/return instead of sending them to the extension.
	for _, id := range []uint{0, 310, 320, 321, 322, 323, 330, 331, 332, 333, 334, 379, math.MaxUint32} {
		control.SetID(id)
		ids, sounds = nil, nil
		if got := click(); got != 1 || len(ids) != 0 || len(sounds) != 1 || sounds[0] != sound.SoundShellClick {
			t.Fatalf("legacy id=%d: return=%d extension=%v sounds=%v", id, got, ids, sounds)
		}
	}
	for _, event := range []int{1, 2, 21, 23, 0x4000, 0x4006, 0x4008, 0x400a} {
		ids, sounds = nil, nil
		if got := gui.EventRespInt(root.Func94(&gui.RawEvent{Event: event})); got != 0 || len(ids) != 0 || len(sounds) != 0 {
			t.Fatalf("event=%#x: return=%d extension=%v sounds=%v", event, got, ids, sounds)
		}
	}
	t.Logf("C procedure=%p native root=%p control=%p", OptionsInGameProc4ADF30(), root, control)
}
