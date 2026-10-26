package opennox

import (
	"fmt"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/datapath"
	"github.com/opennox/libs/ifs"
	"github.com/opennox/opennox/v1/client/gui"
	noxflags "github.com/opennox/opennox/v1/common/flags"
)

// Exercise the complete scanner with an empty private Save directory and real
// native listboxes already containing rows. This isolates the original final
// 4A5649 insertion-position guard; headless coverage separately parses a real
// Quest save and loads its default row without any selection input.
func TestSaveCharacterDefaultSelectionNative(t *testing.T) {
	oldList, oldNames, oldStyle, oldWin := winCharList, winCharListNames, winCharListStyle, winSelSave
	oldSaves, oldData, oldFlags := nox_xxx_saves_arr, datapath.Data(), noxflags.GetGame()
	oldCWD, err := ifs.Workdir()
	if err != nil {
		t.Fatal(err)
	}
	dataDir := t.TempDir()
	t.Cleanup(func() {
		winCharList, winCharListNames, winCharListStyle, winSelSave = oldList, oldNames, oldStyle, oldWin
		nox_xxx_saves_arr = oldSaves
		datapath.SetData(oldData)
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
		if err := ifs.Chdir(oldCWD); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})
	datapath.SetData(dataDir)
	noxflags.ResetGame()
	for _, tc := range []struct {
		name                     string
		nameRows, styleRows      int
		selectedLast, zeroInsert bool
		wantNames, wantStyle     int
	}{
		{name: "empty", wantNames: -1, wantStyle: -1},
		{name: "one", nameRows: 1, styleRows: 1},
		{name: "multiple", nameRows: 3, styleRows: 3},
		{name: "full", nameRows: NOX_SAVEGAME_XXX_MAX, styleRows: NOX_SAVEGAME_XXX_MAX},
		{name: "reset existing selection", nameRows: 3, styleRows: 3, selectedLast: true},
		{name: "zero insertion preserves selection", nameRows: 3, styleRows: 3, selectedLast: true, zeroInsert: true, wantNames: 2, wantStyle: 2},
		{name: "zero insertion keeps unselected", nameRows: 3, styleRows: 3, zeroInsert: true, wantNames: -1, wantStyle: -1},
		{name: "different column lengths", nameRows: 3, styleRows: 2},
		{name: "empty style column", nameRows: 3, wantStyle: -1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := gui.New(nil)
			t.Cleanup(g.DestroyAll)
			root := g.NewWindowRaw(nil, gui.StatusEnabled, 0, 0, 640, 480, nil)
			buttons := make(map[uint]*gui.Window)
			for _, id := range []uint{501, 502, 503} {
				button := g.NewWindowRaw(root, gui.StatusEnabled, 0, 0, 10, 10, nil)
				button.SetID(id)
				buttons[id] = button
			}
			lists := make([]*gui.Window, 3)
			for i, rows := range []int{tc.nameRows, tc.nameRows, tc.styleRows} {
				draw := gui.WindowData{Style: gui.StyleScrollListBox | gui.StyleMouseTrack}
				list := gui.NewScrollListBoxRaw(g, root, gui.StatusEnabled, 0, 0, 200, 300, &draw,
					&gui.ScrollListBoxData{Count: NOX_SAVEGAME_XXX_MAX, Line_height: 17})
				if list == nil {
					t.Fatal("native list construction failed")
				}
				lists[i] = list
				for row := range rows {
					list.Func94(&WindowEvent0x400d{Str: fmt.Sprintf("row %d", row), Val: 14})
				}
				if tc.selectedLast && rows > 0 {
					list.Func94(gui.AsWindowEvent(0x4013, uintptr(rows-1), 0))
				}
			}
			data := make([]gui.ScrollListBoxData, len(lists))
			items := make([][]gui.ScrollListBoxItem, len(lists))
			for i, list := range lists {
				d := (*gui.ScrollListBoxData)(list.WidgetData)
				// A real, unrelated native button occupies the former PE32
				// count offset. Reading its bytes cannot replace the guard.
				d.Field_8 = unsafe.Pointer(buttons[502])
				d.Field_9 = unsafe.Pointer(buttons[503])
				if unsafe.Sizeof(uintptr(0)) == 8 {
					if unsafe.Offsetof(d.Field_11_1) != 62 || unsafe.Offsetof(d.Field_8) != 40 ||
						uintptr(list.WidgetData) <= uintptr(^uint32(0)) || uintptr(d.Field_8) <= uintptr(^uint32(0)) {
						t.Fatal("fixture does not exercise the native LP64 count and high pointers")
					}
				}
				if tc.zeroInsert && i == 1 {
					list.Func94(gui.AsWindowEvent(0x4012, 0, 0))
				}
				data[i] = *d
				items[i] = slices.Clone(unsafe.Slice(d.Items, int(d.Count)))
			}
			winSelSave, winCharList, winCharListNames, winCharListStyle = root, lists[0], lists[1], lists[2]
			nox_xxx_findAutosaves_4A5150()
			for i, want := range []int{tc.wantNames, tc.wantStyle} {
				if got := gui.EventRespInt(lists[i+1].Func94(gui.AsWindowEvent(0x4014, 0, 0))); got != want {
					t.Errorf("column %d default selection=%d, want %d", i+1, got, want)
				}
			}
			for i, list := range lists {
				d := (*gui.ScrollListBoxData)(list.WidgetData)
				if *d != data[i] || !slices.Equal(unsafe.Slice(d.Items, int(d.Count)), items[i]) {
					t.Errorf("scanner changed list %d data/items instead of only native selection", i)
				}
			}
			if !buttons[501].GetFlags().IsEnabled() || buttons[502].GetFlags().IsEnabled() || buttons[503].GetFlags().IsEnabled() ||
				nox_xxx_saves_arr != nil || noxflags.GetGame() != 0 {
				t.Fatal("empty scan changed original New/Load/Delete availability, saves, or game flags")
			}
			if cwd, err := ifs.Workdir(); err != nil || ifs.Normalize(cwd) != ifs.Normalize(datapath.Data()) {
				t.Fatalf("scanner working directory=%q error=%v, want data directory", cwd, err)
			}
		})
	}
}
