package opennox

import (
	"encoding/json"
	"fmt"
	"image"
	"maps"
	"math"
	"os"
	"slices"
	"strings"
	"unsafe"

	"github.com/opennox/libs/client/keybind"
	"github.com/opennox/libs/client/seat"
	"github.com/opennox/libs/common"
	"github.com/opennox/libs/datapath"
	"github.com/opennox/libs/ifs"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// The expectation is test-only scalar data, outside all game save slots. No
// game loader reads it, and no native pointer survives the process boundary.
const e2eManualSaveSnapshotFile = ".e2e-manual-save.json"

type e2eSavedInventoryItem struct {
	Type              string
	Equipped          bool
	Health, MaxHealth uint16
	Modifiers         [4]string
}

type e2eManualSaveSnapshot struct {
	Version, Slot, Stage, Class, Strength int
	Map, Name                             string
	Position                              types.Pointf
	Health, MaxHealth, Mana, MaxMana      uint16
	Gold, Experience                      uint32
	Level                                 uint8
	Inventory                             []e2eSavedInventoryItem
}

func e2eManualSaveSnapshotNow(slot int) (e2eManualSaveSnapshot, error) {
	unit, ud := e2eHostPlayerUnit()
	if slot < 1 || slot >= NOX_SAVEGAME_XXX_MAX || unit == nil || ud == nil || ud.Player == nil ||
		unit.HealthData == nil || unit.HealthData.Cur == 0 || unit.Flags().HasAny(object.FlagDead|object.FlagDestroyed) {
		return e2eManualSaveSnapshot{}, fmt.Errorf("manual save requires a live campaign player and a non-auto slot: %d", slot)
	}
	pl := ud.Player
	snap := e2eManualSaveSnapshot{
		Version: 1, Slot: slot, Stage: int(sub_450750()), Class: int(pl.PlayerClass()), Strength: unit.Strength(),
		Map: e2eMapBaseName(legacy.Nox_xxx_mapGetMapName_409B40()), Name: pl.Name(), Position: unit.PosVec,
		Health: unit.HealthData.Cur, MaxHealth: unit.HealthData.Max, Mana: ud.ManaCur, MaxMana: ud.ManaMax,
		Gold: pl.GoldVal, Experience: math.Float32bits(unit.Experience), Level: pl.Level,
	}
	seen := make(map[*server.Object]bool)
	for item := unit.InvFirstItem; item != nil; item = item.InvNextItem {
		if seen[item] || item.InvHolder != unit || item.Flags().Has(object.FlagDestroyed) {
			return snap, fmt.Errorf("manual save inventory has a cycle, foreign holder or destroyed item")
		}
		seen[item] = true
		entry := e2eSavedInventoryItem{Type: item.ObjectTypeC().ID(), Equipped: item.Flags().Has(object.FlagEquipped)}
		if item.HealthData != nil {
			entry.Health, entry.MaxHealth = item.HealthData.Cur, item.HealthData.Max
		}
		if item.InitData != nil && item.Class().HasAny(object.ClassWeapon|object.ClassArmor|object.ClassWand) {
			for i, mod := range item.InitDataModifier().Modifiers {
				if mod != nil {
					entry.Modifiers[i] = mod.Name()
				}
			}
		}
		snap.Inventory = append(snap.Inventory, entry)
	}
	// Inventory insertion order is not a serialization contract; retain every
	// duplicate item and its equipment/durability/modifiers in canonical order.
	slices.SortFunc(snap.Inventory, func(a, b e2eSavedInventoryItem) int {
		aa, _ := json.Marshal(a)
		bb, _ := json.Marshal(b)
		return strings.Compare(string(aa), string(bb))
	})
	return snap, nil
}

func (want e2eManualSaveSnapshot) compare(got e2eManualSaveSnapshot) error {
	if want.Version != 1 || want.Slot < 1 || want.Slot >= NOX_SAVEGAME_XXX_MAX || want.Map == "" ||
		want.Name == "" || want.Class < 0 || want.Class > 2 || want.Stage < 1 || want.Health == 0 ||
		want.Health > want.MaxHealth || want.Mana > want.MaxMana || len(want.Inventory) == 0 ||
		math.IsNaN(float64(want.Position.X)) || math.IsNaN(float64(want.Position.Y)) ||
		math.IsInf(float64(want.Position.X), 0) || math.IsInf(float64(want.Position.Y), 0) {
		return fmt.Errorf("invalid manual save expectation: %+v", want)
	}
	pos := got.Position
	wantPos := want.Position
	want.Position, got.Position = types.Pointf{}, types.Pointf{}
	wantItems, gotItems := want.Inventory, got.Inventory
	want.Inventory, got.Inventory = nil, nil
	wa, _ := json.Marshal(want)
	ga, _ := json.Marshal(got)
	if string(wa) != string(ga) || !slices.Equal(wantItems, gotItems) {
		return fmt.Errorf("manual save player/inventory changed: got=%s items=%+v want=%s items=%+v", ga, gotItems, wa, wantItems)
	}
	if d := math.Hypot(float64(pos.X-wantPos.X), float64(pos.Y-wantPos.Y)); math.IsNaN(d) || math.IsInf(d, 0) || d > 2 {
		return fmt.Errorf("manual save location=%v, want %v (distance %.3f)", pos, wantPos, d)
	}
	return nil
}

func e2eSaveControlVisible(win *gui.Window) bool {
	if win == nil || !win.GetFlags().IsEnabled() {
		return false
	}
	for w := win; w != nil; w = w.Parent() {
		if w.GetFlags().IsHidden() {
			return false
		}
	}
	return true
}

// These helpers only queue genuine keyboard/mouse events. Widget positions
// use canvas pixels; even headless menus have a 640x480 canvas in a 1024x768
// input viewport. Convert to logical window coordinates before raw playback.
// The 0x4014 event is a read-only listbox query, not a write to selection.
func (sc *e2eScenario) e2eClickSaveControl(name string, control func() *gui.Window) {
	sc.addWhen(0, name, 1200, func() bool { return e2eSaveControlVisible(control()) }, func() {
		win := control()
		pos := win.GlobalPos().Add(win.Size().Div(2))
		e2eLog.Printf("MANUAL SAVE INPUT: %s id=%d point=%v", name, win.ID(), pos)
		e2eQueueRawInput(&seat.MouseMoveEvent{Pos: noxClient.Inp.DrawPosToWindow(pos), Relative: false})
	})
	sc.Input(1, "", &seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: true})
	sc.Input(1, "", &seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: false})
	sc.Wait(2, "")
}

func (sc *e2eScenario) e2eOpenSaveMenu(name string) {
	sc.Key(keybind.KeyEsc, name+" Escape")
	sc.e2eClickSaveControl(name+" open Save/Load", func() *gui.Window { return noxClient.GUI.ChildByID(9003) })
	sc.addWhen(0, name+" wait for save list", 1200, func() bool { return e2eSaveControlVisible(dword_5d4594_1082856) }, nil)
}

func (sc *e2eScenario) e2eSelectSaveRow(slot int, menu bool, name string) {
	list := func() *gui.Window {
		if menu {
			return winCharListNames
		}
		return dword_5d4594_1082864
	}
	sc.addWhen(0, name+" locate row", 1200, func() bool {
		return e2eSaveControlVisible(list()) && (!menu || nox_wnd_xxx_1307748.State() == gui.AnimInDone)
	}, func() {
		win := list()
		d := (*gui.ScrollListBoxData)(win.WidgetData)
		if d == nil || d.Items == nil || slot < 0 || slot >= int(d.Field_11_0) || d.Field_11_0 > d.Count {
			e2eError(fmt.Errorf("manual save list has no populated row %d", slot))
			return
		}
		items := unsafe.Slice(d.Items, int(d.Count))
		top := 0
		if slot != 0 {
			top = int(items[slot-1].Field_0) + 1
		}
		bottom := int(items[slot].Field_0)
		titleHeight := 0
		if win.DrawData().Text() != "" && win.GUI().Render() != nil {
			titleHeight = win.GUI().Render().FontHeight(win.DrawData().Font()) + 1
		}
		y := titleHeight + (top+bottom)/2 - int(d.Field_13_1)
		if y < 0 || y >= win.Size().Y {
			e2eError(fmt.Errorf("manual save row %d is outside visible viewport", slot))
			return
		}
		pos := win.GlobalPos().Add(image.Pt(win.Size().X/2, y))
		e2eLog.Printf("MANUAL SAVE ROW INPUT: slot=%d menu=%t point=%v list=%v size=%v rows=%d edges=%d..%d scroll=%d title=%q capture=%p hit=%p",
			slot, menu, pos, win.GlobalPos(), win.Size(), d.Field_11_0, top, bottom, d.Field_13_1, win.DrawData().Text(),
			noxClient.GUI.Captured(), noxClient.GUI.Captured().ChildByPos(pos))
		e2eQueueRawInput(&seat.MouseMoveEvent{Pos: noxClient.Inp.DrawPosToWindow(pos), Relative: false})
	})
	sc.Input(1, "", &seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: true})
	sc.Input(1, "", &seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: false})
	var waited int
	sc.addWhen(2, name+" verify native selection", 120, func() bool {
		waited++
		// The outer list has the save-menu procedure installed in Func94,
		// rather than a listbox query procedure. Query the two real native
		// columns; the timestamp column is the original slot source.
		lists := []*gui.Window{dword_5d4594_1082864, dword_5d4594_1082868}
		if menu {
			lists = []*gui.Window{winCharListNames, winCharListStyle}
		}
		selected := make([]int, len(lists))
		ready := true
		for i, win := range lists {
			selected[i] = gui.EventRespInt(win.Func94(gui.AsWindowEvent(0x4014, 0, 0)))
			ready = ready && win != nil && selected[i] == slot
		}
		if !ready && (waited == 1 || waited == 120) {
			e2eLog.Printf("MANUAL SAVE SELECTION WAIT: slot=%d menu=%t selections=%v", slot, menu, selected)
		}
		return ready
	}, func() { e2eLog.Printf("MANUAL SAVE SLOT SELECTED: slot=%d menu=%t real-mouse=true", slot, menu) })
}

func (sc *e2eScenario) e2eSaveButton(id uint, name string) {
	sc.e2eClickSaveControl(name, func() *gui.Window { return dword_5d4594_1082856.ChildByID(id) })
}

func (sc *e2eScenario) e2eSaveConfirmation(yes bool, name string) {
	id := uint(guiDialogNoID)
	if yes {
		id = guiDialogYesID
	}
	sc.e2eClickSaveControl(name, func() *gui.Window { return nox_gui_curDialog_830224.ChildByID(id) })
}

type e2eManualSaveFixture struct {
	want                   e2eManualSaveSnapshot
	slot                   int
	autoHashes, slotHashes map[string]string
}

func (f *e2eManualSaveFixture) capture() {
	var err error
	f.want, err = e2eManualSaveSnapshotNow(f.slot)
	if err != nil {
		e2eError(err)
		return
	}
	e2eLog.Printf("MANUAL SAVE CAPTURE: %+v", f.want)
}

func (sc *e2eScenario) e2eWaitManualSave(f *e2eManualSaveFixture, overwrite bool, name string) {
	saveName := fmt.Sprintf(common.SaveFormat, f.slot)
	sc.addWhen(0, name, 2400, func() bool {
		if !overwrite && e2eSaveControlVisible(nox_gui_curDialog_830224.ChildByID(guiDialogYesID)) {
			e2eError(fmt.Errorf("empty manual slot %d unexpectedly asks to overwrite a populated slot (native selection=%d)",
				f.slot, gui.EventRespInt(dword_5d4594_1082864.Func94(gui.AsWindowEvent(0x4014, 0, 0)))))
			return false
		}
		if nox_xxx_gameGet_4DB1B0() {
			return false
		}
		path := datapath.Save(saveName, common.PlayerFile)
		if fi, err := ifs.Stat(path); err != nil || fi.Size() == 0 {
			return false
		}
		var meta server.SaveGameInfo
		if legacy.Sub_41A000(path, &meta) == 0 {
			return false
		}
		if datapath.SaveNameFromPath(meta.Path()) != saveName || meta.Player.Name() != f.want.Name ||
			int(meta.Player.PlayerClass()) != f.want.Class || e2eMapBaseName(alloc.GoStringS(meta.MapNameBuf[:])) != f.want.Map ||
			int(meta.Stage) != f.want.Stage || meta.Timestamp.Time().Year() < 1997 {
			e2eError(fmt.Errorf("manual slot %s metadata does not match captured player", saveName))
			return false
		}
		mapPath := datapath.Save(saveName, f.want.Map, f.want.Map+".map")
		if fi, err := ifs.Stat(mapPath); err != nil || fi.Size() == 0 {
			return false
		}
		hashes := e2eHashDir(datapath.Save(saveName))
		if overwrite && maps.Equal(hashes, f.slotHashes) {
			return false
		}
		if !maps.Equal(e2eHashDir(datapath.Save(common.SaveAuto)), f.autoHashes) {
			e2eError(fmt.Errorf("manual slot %d changed AUTOSAVE instead of preserving it", f.slot))
			return false
		}
		f.slotHashes = hashes
		return true
	}, func() {
		e2eLog.Printf("MANUAL SAVE FILE VERIFIED: slot=%d overwrite=%t files=%d autosave-unchanged=true", f.slot, overwrite, len(f.slotHashes))
	})
}

func (sc *e2eScenario) e2eWaitManualRestored(want func() e2eManualSaveSnapshot, name string) {
	var lastErr error
	var waited int
	sc.addWhen(0, name, 3600, func() bool {
		waited++
		expected := want()
		got, err := e2eManualSaveSnapshotNow(expected.Slot)
		if err == nil {
			err = expected.compare(got)
		}
		lastErr = err
		if err != nil && (waited == 1 || waited%600 == 0) {
			e2eLog.Printf("MANUAL SAVE RESTORE WAIT: %s: %v", name, err)
		}
		if waited >= 3600 && err != nil {
			e2eError(fmt.Errorf("manual save restore did not match: %w", err))
		}
		if err != nil || !nox_client_isConnected() || nox_xxx_gameGet_4DB1B0() {
			return false
		}
		unit, _ := e2eHostPlayerUnit()
		drawable := noxClient.ClientPlayerUnit()
		return drawable != nil && drawable.NetCode32 == uint32(noxServer.GetUnitNetCode(unit)) &&
			!e2eSaveControlVisible(dword_5d4594_1082856)
	}, func() {
		if lastErr != nil {
			e2eError(lastErr)
			return
		}
		e2eLog.Printf("MANUAL SAVE RESTORED: %+v exact-stats-and-inventory=true position-tolerance=2", want())
	})
}

func (sc *e2eScenario) CheckManualSaveLoad(slot int, name string) {
	f := &e2eManualSaveFixture{slot: slot}
	sc.add(0, name+" capture initial player and autosave", func() {
		f.capture()
		f.autoHashes = e2eHashDir(datapath.Save(common.SaveAuto))
	})
	sc.e2eOpenSaveMenu(name)
	sc.e2eSelectSaveRow(slot, false, name)
	sc.add(0, name+" verify initial manual slot is empty", func() {
		if slot < 1 || slot >= NOX_SAVEGAME_XXX_MAX || nox_savegame_arr_1064948[slot].Path() != "" {
			e2eError(fmt.Errorf("manual save fixture requires an empty non-auto slot %d", slot))
		}
	})
	sc.e2eSaveButton(501, name+" save empty manual slot")
	sc.e2eWaitManualSave(f, false, name+" verify first manual file")
	sc.RunFor(0.25, 70, name+" real movement before overwrite")
	sc.Wait(30, "")
	sc.add(0, name+" capture overwrite position", func() {
		before := f.want.Position
		f.capture()
		if d := math.Hypot(float64(f.want.Position.X-before.X), float64(f.want.Position.Y-before.Y)); d < 8 {
			e2eError(fmt.Errorf("manual save overwrite fixture moved %.3f, want at least 8", d))
		}
	})
	sc.e2eOpenSaveMenu(name + " cancel overwrite")
	sc.e2eSelectSaveRow(slot, false, name)
	sc.e2eSaveButton(501, name+" request overwrite")
	sc.e2eSaveConfirmation(false, name+" refuse overwrite")
	sc.add(0, name+" verify refusal leaves files unchanged", func() {
		if !maps.Equal(f.slotHashes, e2eHashDir(datapath.Save(fmt.Sprintf(common.SaveFormat, slot)))) ||
			!maps.Equal(f.autoHashes, e2eHashDir(datapath.Save(common.SaveAuto))) {
			e2eError(fmt.Errorf("declining overwrite changed a save"))
			return
		}
		e2eLog.Printf("MANUAL SAVE OVERWRITE REFUSED: slot=%d manual-and-auto-files-unchanged=true", slot)
	})
	sc.e2eSaveButton(501, name+" request overwrite again")
	sc.e2eSaveConfirmation(true, name+" confirm overwrite")
	sc.e2eWaitManualSave(f, true, name+" verify overwritten manual file")
	sc.RunFor(0.75, 70, name+" real movement before reload")
	sc.Wait(30, "")
	sc.add(0, name+" prove reload must restore an older location", func() {
		got, err := e2eManualSaveSnapshotNow(slot)
		if err != nil {
			e2eError(err)
			return
		}
		if d := math.Hypot(float64(got.Position.X-f.want.Position.X), float64(got.Position.Y-f.want.Position.Y)); d < 8 {
			e2eError(fmt.Errorf("manual reload fixture moved %.3f away from saved location, want at least 8", d))
		}
	})
	sc.e2eOpenSaveMenu(name + " reload")
	sc.e2eSelectSaveRow(slot, false, name)
	sc.e2eSaveButton(502, name+" load manual slot")
	sc.e2eSaveConfirmation(true, name+" confirm live reload")
	sc.e2eWaitManualRestored(func() e2eManualSaveSnapshot { return f.want }, name+" verify restored player")
	sc.add(0, name+" record scalar expectation for fresh process", func() {
		data, err := json.Marshal(f.want)
		if err == nil {
			err = os.WriteFile(datapath.Save(e2eManualSaveSnapshotFile), data, 0600)
		}
		if err != nil {
			e2eError(err)
		}
	})
	sc.Wait(100, name+" wait for original load fade")
	sc.Screen(name + " restored player")
	sc.WaitPlayerReloaded(name + " bind restored player")
	sc.RunFor(0.75, 70, name+" real control after live reload")
	sc.AssertPlayerMovedAfterReload(name + " verify control after live reload")
}

func (sc *e2eScenario) SelectManualCampaignSave(slot int, name string) {
	sc.e2eSelectSaveRow(slot, true, name)
}

func (sc *e2eScenario) AssertManualSaveRestored(name string) {
	var want e2eManualSaveSnapshot
	sc.add(0, name+" read test-only scalar expectation", func() {
		data, err := os.ReadFile(datapath.Save(e2eManualSaveSnapshotFile))
		if err == nil {
			err = json.Unmarshal(data, &want)
		}
		if err == nil {
			err = want.compare(want)
		}
		if err != nil {
			e2eError(err)
		}
	})
	sc.e2eWaitManualRestored(func() e2eManualSaveSnapshot { return want }, name)
}
