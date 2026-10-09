package opennox

import (
	"fmt"
	"image"
	"unsafe"

	"github.com/opennox/libs/client/keybind"
	"github.com/opennox/libs/client/seat"
	"github.com/opennox/libs/noximage"
	"github.com/opennox/libs/spell"
	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/legacy"
)

// Use real mouse input for insertion, row selection, and reopening. The only
// starting fixture is the main quickbar spell supplied by the YAML. Never write
// Trap Set entries, window flags, callbacks, or the live framebuffer here.
func (sc *e2eScenario) CheckTrapQuickbarIcons(name string) {
	const id = spell.SPELL_FIREBALL
	var want, mainBefore [25][2]uint32
	var mainRow int
	sc.addWhen(0, name+" observe starting fixture", 1200, func() bool {
		return nox_client_isConnected() && noxServer.Players.HostUnit() != nil &&
			legacy.ClientTrapQuickbarRoot() != nil && legacy.ClientQuickbarButton(0) != nil
	}, func() {
		var row int
		var open, ok bool
		want, row, open = legacy.ClientTrapQuickbarSnapshot()
		mainBefore, mainRow, ok = legacy.ClientQuickbarSnapshot()
		def := noxServer.Spells.DefByInd(id)
		if !ok || row != 0 || open || mainBefore[5*mainRow][0] != uint32(id) ||
			def == nil || !def.IsEnabled() || !noxServer.Spells.CanUseInTraps(id) || def.Icon == nil {
			e2eError(fmt.Errorf("Trap Set starting fixture unavailable: row=%d open=%t main_row=%d spell=%v", row, open, mainRow, def))
			return
		}
		for row := 0; row < 3; row++ {
			for slot := 0; slot < 3; slot++ {
				if want[5*row+slot][0] != 0 {
					e2eError(fmt.Errorf("Trap Set starting slot is not empty: row=%d slot=%d", row, slot))
					return
				}
			}
		}
		for slot := 0; slot < 3; slot++ {
			win := legacy.ClientTrapQuickbarButton(slot)
			if win == nil || win.Parent() != legacy.ClientTrapQuickbarRoot() ||
				(unsafe.Sizeof(uintptr(0)) == 8 && uintptr(win.C()) <= 0xffffffff) {
				e2eError(fmt.Errorf("Trap Set native slot pointer unavailable: slot=%d win=%p", slot, win))
				return
			}
			e2eLog.Printf("TRAP SET NATIVE SLOT: slot=%d win=%p pos=%v flags=%#x", slot, win, win.GlobalPos(), win.GetFlags())
		}
		e2eLog.Printf("TRAP SET FIXTURE: spell=%d main_row=%d numeric_entries=%v", id, mainRow, want)
	})

	click := func(control int, label string) {
		sc.add(0, name+" "+label, func() {
			win := legacy.ClientTrapQuickbarControl(control)
			if !e2eTrapWindowVisible(win) {
				e2eError(fmt.Errorf("Trap Set control is not visible: control=%d win=%p", control, win))
				return
			}
			pos := win.GlobalPos().Add(win.Size().Div(2))
			e2eQueueInput(&seat.MouseMoveEvent{Pos: pos}, &seat.MouseButtonEvent{Pressed: true})
		})
		sc.Input(1, name+" release control", &seat.MouseButtonEvent{Pressed: false})
		sc.Input(2, name+" move cursor clear of icons", &seat.MouseMoveEvent{Pos: image.Pt(700, 380)})
		sc.Wait(30, name+" settle control")
	}
	assert := func(row int, open, rendered bool, label string) {
		sc.add(0, name+" "+label, func() {
			entries, gotRow, gotOpen := legacy.ClientTrapQuickbarSnapshot()
			main, gotMainRow, ok := legacy.ClientQuickbarSnapshot()
			root := legacy.ClientTrapQuickbarRoot()
			visible := open && rendered // F11 hides the root but keeps the selected tray open.
			if entries != want || gotRow != row || gotOpen != open || !ok ||
				main != mainBefore || gotMainRow != mainRow || root == nil ||
				e2eTrapWindowVisible(root) != visible || nox_client_renderGUI_80828 != rendered {
				e2eError(fmt.Errorf("Trap Set state mismatch: phase=%s row=%d/%d open=%t/%t rendered=%t/%t root_visible=%t/%t main_row=%d/%d main_unchanged=%t entries=%v want=%v", label, gotRow, row, gotOpen, open, nox_client_renderGUI_80828, rendered, e2eTrapWindowVisible(root), visible, gotMainRow, mainRow, main == mainBefore, entries, want))
				return
			}
			for slot := 0; slot < 3; slot++ {
				win := legacy.ClientTrapQuickbarButton(slot)
				if win == nil || win.GetFlags().IsHidden() || e2eTrapWindowVisible(win) != visible {
					e2eError(fmt.Errorf("Trap Set slot remains hidden: phase=%s slot=%d win=%p", label, slot, win))
					return
				}
				if entries[5*row+slot][0] == 0 {
					continue
				}
				pixels, matching := e2eTrapQuickbarIconPixels(win, id)
				if pixels < 64 || (open && rendered && matching != pixels) ||
					((!open || !rendered) && matching == pixels) {
					e2eError(fmt.Errorf("Trap Set icon mismatch: phase=%s row=%d slot=%d reference=%d matching=%d open=%t rendered=%t", label, row, slot, pixels, matching, open, rendered))
					return
				}
				e2eLog.Printf("TRAP SET ICON VERIFIED: phase=%s row=%d slot=%d spell=%d reference_pixels=%d matching_pixels=%d open=%t rendered=%t live_frame=true", label, row, slot, id, pixels, matching, open, rendered)
			}
			e2eLog.Printf("TRAP SET STATE VERIFIED: phase=%s row=%d open=%t rendered=%t main_unchanged=true reserved_entries_unchanged=true", label, row, open, rendered)
		})
	}

	click(0, "open actual Trap Set")
	assert(0, true, true, "opened empty tray")
	for row := 0; row < 3; row++ {
		if row != 0 {
			click(4, "select next trap row")
			assert(row, true, true, "next empty row")
		}
		sc.add(0, name+" press source spell", func() {
			win := legacy.ClientQuickbarButton(0)
			e2eQueueInput(&seat.MouseMoveEvent{Pos: win.GlobalPos().Add(win.Size().Div(2))}, &seat.MouseButtonEvent{Pressed: true})
		})
		sc.add(4, name+" drag onto actual trap slot", func() {
			win := legacy.ClientTrapQuickbarButton(row)
			e2eQueueInput(&seat.MouseMoveEvent{Pos: win.GlobalPos().Add(win.Size().Div(2))})
		})
		sc.Input(4, name+" release actual spell drop", &seat.MouseButtonEvent{Pressed: false})
		sc.Input(2, name+" move cursor clear of icons", &seat.MouseMoveEvent{Pos: image.Pt(700, 380)})
		sc.add(30, name+" expect only the actual drop", func() { want[5*row+row][0] = uint32(id) })
		assert(row, true, true, "real mouse drop")
		sc.Screen(fmt.Sprintf("Trap Set row %d spell icon", row+1))
	}
	// Cycle through the existing rows without writing their selected indices.
	click(4, "wrap to first row")
	assert(0, true, true, "wrapped first row")
	click(3, "previous row wraps to third")
	assert(2, true, true, "previous third row")
	click(0, "close actual Trap Set")
	assert(2, false, true, "closed tray")
	click(0, "reopen actual Trap Set")
	assert(2, true, true, "reopened tray")
	sc.Key(keybind.KeyF11, name+" hide HUD through F11")
	sc.Wait(30, name+" settle hidden HUD")
	assert(2, true, false, "F11 hidden HUD")
	sc.Key(keybind.KeyF11, name+" show HUD through F11")
	sc.Wait(30, name+" settle restored HUD")
	assert(2, true, true, "F11 restored HUD")
	click(4, "return to first row")
	assert(0, true, true, "final first row")
}

func e2eTrapWindowVisible(win *gui.Window) bool {
	if win == nil {
		return false
	}
	for ; win != nil; win = win.Parent() {
		if win.GetFlags().IsHidden() {
			return false
		}
	}
	return true
}

// Draw the stock opaque spell silhouette in a separate buffer, then compare it
// to the frame already produced by normal GUI traversal. Do not invoke the
// slot's production draw function or copy reference pixels into the live frame.
func e2eTrapQuickbarIconPixels(win *gui.Window, id spell.ID) (pixels, matching int) {
	r := noxClient.r
	live := r.PixBuffer()
	data := *r.Data()
	defer func() { r.SetPixBuffer(live); *r.Data() = data }()
	reference := noximage.NewImage16(live.Rect)
	r.SetPixBuffer(reference)
	r.Data().Reset()
	r.Data().SetClip(false)
	icon := (*noxrender.Image)(noxServer.Spells.DefByInd(id).Icon)
	r.DrawImage16(icon, win.GlobalPos())
	return e2eItemIconPixelCount(live, reference)
}
