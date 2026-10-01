package opennox

import (
	"fmt"
	"slices"
	"testing"

	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

func TestSaveCharacterSelectionRejectsUnavailableSlot(t *testing.T) {
	oldList, oldSlots, oldName := winCharListNames, nox_xxx_saves_arr, nox_savegame_name_1307752
	deleteIndex := memmap.PtrInt32(0x5D4594, 1307772)
	oldDeleteIndex := *deleteIndex
	t.Cleanup(func() {
		winCharListNames, nox_xxx_saves_arr, nox_savegame_name_1307752 = oldList, oldSlots, oldName
		*deleteIndex = oldDeleteIndex
	})
	g := gui.New(nil)
	t.Cleanup(g.DestroyAll)
	list := g.NewWindowRaw(nil, gui.StatusEnabled, 0, 0, 10, 10, nil)
	button := g.NewWindowRaw(nil, gui.StatusEnabled, 0, 0, 10, 10, nil)
	for _, id := range []uint{502, 503} {
		button.SetID(id)
		for _, tc := range []struct {
			name     string
			index    int
			missing  bool
			nilReply bool
			empty    bool
		}{
			{name: "missing list", missing: true},
			{name: "nil response", nilReply: true},
			{name: "unselected", index: -1},
			{name: "negative", index: -2},
			{name: "past slots", index: NOX_SAVEGAME_XXX_MAX},
			{name: "large index", index: 1 << 30},
			{name: "empty slot array", empty: true},
			{name: "empty manual load slot", index: 1},
		} {
			if id == 503 && tc.index == 1 {
				continue // The original empty-delete path records the choice before rejecting it.
			}
			t.Run(fmt.Sprintf("%s button %d", tc.name, id), func(t *testing.T) {
				nox_xxx_saves_arr = make([]server.SaveGameInfo, NOX_SAVEGAME_XXX_MAX)
				copy(nox_xxx_saves_arr[0].PathBuf[:], "Save/AUTOSAVE/Player.plr")
				if tc.empty {
					nox_xxx_saves_arr = nil
				}
				before := slices.Clone(nox_xxx_saves_arr)
				nox_savegame_name_1307752 = "unchanged delete choice"
				*deleteIndex = 7
				calls := 0
				list.SetFunc94(func(_ *gui.Window, ev gui.WindowEvent) gui.WindowEventResp {
					calls++
					a, b := ev.EventArgsC()
					if ev.EventCode() != 0x4014 || a != 0 || b != 0 {
						t.Fatal("character menu wrote to the save list")
					}
					if tc.nilReply {
						return nil
					}
					return gui.RawEventResp(uintptr(uint32(int32(tc.index))))
				})
				winCharListNames = list
				if tc.missing {
					winCharListNames = nil
				}
				resp := nox_xxx_windowSelCharProc_4A5710(nil, &WindowEvent0x4007{Win: button})
				wantCalls := 1
				if tc.missing || tc.empty {
					wantCalls = 0
				}
				if gui.EventRespInt(resp) != 1 || calls != wantCalls || !slices.Equal(nox_xxx_saves_arr, before) ||
					nox_savegame_name_1307752 != "unchanged delete choice" || *deleteIndex != 7 {
					t.Fatal("rejected selection loaded/deleted a save, changed delete choice, or used an invalid query")
				}
			})
		}
	}
}
