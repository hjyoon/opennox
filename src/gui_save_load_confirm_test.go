package opennox

import (
	"testing"

	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/server"
)

func TestSaveLoadConfirmRejectsUnavailableSlot(t *testing.T) {
	oldList, oldSlots := dword_5d4594_1082864, nox_savegame_arr_1064948
	t.Cleanup(func() {
		dword_5d4594_1082864, nox_savegame_arr_1064948 = oldList, oldSlots
	})
	g := gui.New(nil)
	t.Cleanup(g.DestroyAll)
	win := g.NewWindowRaw(nil, gui.StatusEnabled, 0, 0, 10, 10, nil)
	for _, tc := range []struct {
		name     string
		index    int
		missing  bool
		nilReply bool
	}{
		{name: "missing list", missing: true},
		{name: "nil response", nilReply: true},
		{name: "unselected", index: -1},
		{name: "negative", index: -2},
		{name: "past slots", index: NOX_SAVEGAME_XXX_MAX},
		{name: "large index", index: 1 << 30},
		{name: "empty first manual slot", index: 1},
		{name: "empty last manual slot", index: NOX_SAVEGAME_XXX_MAX - 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			nox_savegame_arr_1064948 = [NOX_SAVEGAME_XXX_MAX]server.SaveGameInfo{}
			// Selecting an empty manual slot must not fall back to a populated
			// AUTOSAVE. No client or loader is installed in this unit test.
			copy(nox_savegame_arr_1064948[0].PathBuf[:], "Save/AUTOSAVE/Player.plr")
			before := nox_savegame_arr_1064948
			calls := 0
			win.SetFunc94(func(_ *gui.Window, ev gui.WindowEvent) gui.WindowEventResp {
				calls++
				a, b := ev.EventArgsC()
				if ev.EventCode() != 0x4014 || a != 0 || b != 0 {
					t.Fatal("load confirmation wrote to the save list")
				}
				if tc.nilReply {
					return nil
				}
				return gui.RawEventResp(uintptr(uint32(int32(tc.index))))
			})
			dword_5d4594_1082864 = win
			if tc.missing {
				dword_5d4594_1082864 = nil
			}
			nox_savegame_sub_46CBD0()
			wantCalls := 1
			if tc.missing {
				wantCalls = 0
			}
			if calls != wantCalls || nox_savegame_arr_1064948 != before {
				t.Fatalf("rejected selection changed saves or query count: calls=%d, want=%d", calls, wantCalls)
			}
		})
	}
}
