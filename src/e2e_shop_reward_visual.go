package opennox

import (
	"fmt"
	"image"

	"github.com/opennox/libs/client/keybind"
	"github.com/opennox/libs/client/seat"
	"github.com/opennox/libs/noximage"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/things"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy"
)

// Compare an opaque stock item silhouette against the already rendered live
// quantity dialog. The reference is separate; no expected pixels are written
// into the actual framebuffer, and the real item's position is not changed.
func (sc *e2eScenario) AssertItemAmountIcon(item, name string) {
	sc.add(0, name, func() {
		active, _, _ := legacy.Nox_gui_itemAmountState()
		win := legacy.Get_nox_gui_itemAmount_dialog_1319228()
		drawable := legacy.ClientItemAmountDrawable()
		if !active || win == nil || win.GetFlags().IsHidden() || drawable == nil || drawable.TypeIDVal != uint32(noxClient.Things.IndByID(item)) {
			e2eError(fmt.Errorf("item amount icon unavailable: active=%t window=%p drawable=%p item=%q", active, win, drawable, item))
			return
		}
		live := noxClient.r.PixBuffer()
		reference := noximage.NewImage16(live.Rect)
		data, savedDrawable := *noxClient.r.Data(), *drawable
		noxClient.r.SetPixBuffer(reference)
		defer func() {
			noxClient.r.SetPixBuffer(live)
			*noxClient.r.Data() = data
			*drawable = savedDrawable
		}()
		viewport := noxrender.Viewport{Screen: live.Rect, Size: live.Rect.Size()}
		drawable.CallDraw(&viewport)
		pixels, matching := e2eItemIconPixelCount(live, reference)
		if pixels == 0 || matching != pixels {
			e2eError(fmt.Errorf("item amount icon missing: item=%s pos=%v reference_pixels=%d matching_pixels=%d", item, savedDrawable.Pos(), pixels, matching))
			return
		}
		e2eLog.Printf("ITEM AMOUNT ICON VERIFIED: item=%s drawable=%p pos=%v reference_pixels=%d matching_pixels=%d live_frame=true", item, drawable, savedDrawable.Pos(), pixels, matching)
	})
}

func e2eItemIconPixelCount(actual, reference *noximage.Image16) (pixels, matching int) {
	for y := reference.Rect.Min.Y; y < reference.Rect.Max.Y; y++ {
		for x := reference.Rect.Min.X; x < reference.Rect.Max.X; x++ {
			value := reference.Pix[reference.PixOffset(x, y)]
			if value == 0 {
				continue
			}
			pixels++
			if image.Pt(x, y).In(actual.Rect) && actual.Pix[actual.PixOffset(x, y)] == value {
				matching++
			}
		}
	}
	return
}

// Award through the real server path and observe the client packet, natural
// book animation, insertion, and stock 0x600 default. Other entries (including
// manually changed targets) must remain byte-for-byte unchanged.
func (sc *e2eScenario) CheckBookRewardDefault(id spell.ID, toggle bool, name string) {
	var before [25][2]uint32
	var index, row, maxTrail int
	var flags things.SpellFlags
	sc.addWhen(0, name+" prepare actual award", 1200, func() bool {
		return noxServer.Players.HostUnit() != nil && legacy.Get_dword_8531A0_2576() != nil && memmap.Uint32(0x5D4594, 1047520) == 0
	}, func() {
		var ok bool
		before, row, ok = legacy.ClientQuickbarSnapshot()
		def := noxServer.Spells.DefByInd(id)
		if !ok || def == nil || id <= 0 || id >= 137 {
			e2eError(fmt.Errorf("book award setup unavailable: spell=%d row=%d ok=%t", id, row, ok))
			return
		}
		flags = noxServer.Spells.Flags(id)
		if flags&0x15000 != 0 {
			e2eError(fmt.Errorf("book award spell is a family source: %d flags=%x", id, uint32(flags)))
			return
		}
		index = -1
		for _, entry := range before {
			if entry[0] == uint32(id) {
				e2eError(fmt.Errorf("spell %d already occurs in quickbar", id))
				return
			}
		}
		for n := 0; n < 5 && index < 0; n++ {
			for slot := 0; slot < 5; slot++ {
				i := ((row+n)%5)*5 + slot
				if before[i][0] == 0 {
					index = i
					break
				}
			}
		}
		if index < 0 {
			e2eError(fmt.Errorf("no free slot for book award %d", id))
			return
		}
		unit := noxServer.Players.HostUnit()
		if got := legacy.Nox_xxx_spellGrantToPlayer_4FB550(unit, id, 1, 1, 1); got != 1 {
			e2eError(fmt.Errorf("server book award failed: spell=%d result=%d", id, got))
			return
		}
		e2eLog.Printf("BOOK AWARD PREPARED: spell=%d flags=%x row=%d slot=%d real_server_packet=true", id, uint32(flags), index/5, index%5)
	})
	sc.addWhen(0, name+" observe normal reward completion", 1200, func() bool {
		if trail := int(memmap.Uint32(0x5D4594, 1046680)); trail > maxTrail {
			maxTrail = trail
		}
		entries, _, ok := legacy.ClientQuickbarSnapshot()
		return ok && index >= 0 && entries[index][0] == uint32(id) && memmap.Uint32(0x5D4594, 1047520) == 0
	}, func() {
		entries, gotRow, ok := legacy.ClientQuickbarSnapshot()
		want := before
		want[index][0] = uint32(id)
		want[index][1] &^= 0xff
		if flags&0x600 != 0 {
			want[index][1] |= 1
		}
		if !ok || entries != want || gotRow != index/5 || legacy.Get_dword_8531A0_2576().SpellLvl[id] != 1 {
			e2eError(fmt.Errorf("book default/preservation mismatch: spell=%d flags=%x row=%d want_row=%d entries=%v want=%v", id, uint32(flags), gotRow, index/5, entries, want))
			return
		}
		self := entries[index][1]&1 != 0
		e2eLog.Printf("BOOK DEFAULT VERIFIED: spell=%d stock_flags=%x cast_on_self=%t max_natural_trail=%d preserved_other_entries=true", id, uint32(flags), self, maxTrail)
	})
	// Close the book through its normal input before testing the real toggle.
	sc.Key(keybind.KeyB, name+" close book with B")
	sc.Wait(30, name+" settle book close")
	if !toggle {
		return
	}
	sc.add(0, name+" press actual target toggle", func() {
		win := legacy.ClientQuickbarNugget(index % 5)
		if flags&0x200400 != 0 || win == nil || win.GetFlags().IsHidden() {
			e2eError(fmt.Errorf("award target toggle unavailable: spell=%d flags=%x win=%p", id, uint32(flags), win))
			return
		}
		pos := win.GlobalPos().Add(win.Size().Div(2))
		e2eQueueInput(&seat.MouseMoveEvent{Pos: pos}, &seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: true})
	})
	sc.Input(1, name+" release target toggle", &seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: false})
	sc.add(5, name+" verify manual target choice", func() {
		entries, _, ok := legacy.ClientQuickbarSnapshot()
		want := before
		want[index][0] = uint32(id)
		want[index][1] &^= 0xff
		if flags&0x600 == 0 {
			want[index][1] |= 1
		}
		if !ok || entries != want {
			e2eError(fmt.Errorf("manual target toggle failed: spell=%d entries=%v want=%v", id, entries, want))
			return
		}
		e2eLog.Printf("BOOK MANUAL TARGET VERIFIED: spell=%d cast_on_self=%t actual_mouse_click=true", id, entries[index][1]&1 != 0)
	})
	var manual [25][2]uint32
	sc.add(0, name+" repeat real award after manual choice", func() {
		manual, _, _ = legacy.ClientQuickbarSnapshot()
		if got := legacy.Nox_xxx_spellGrantToPlayer_4FB550(noxServer.Players.HostUnit(), id, 1, 1, 0); got != 1 {
			e2eError(fmt.Errorf("repeat server book award failed: spell=%d result=%d", id, got))
		}
	})
	sc.addWhen(0, name+" verify existing manual choice survives repeated award", 1200, func() bool {
		return legacy.Get_dword_8531A0_2576().SpellLvl[id] == 2 && memmap.Uint32(0x5D4594, 1047520) == 0
	}, func() {
		entries, _, ok := legacy.ClientQuickbarSnapshot()
		if !ok || entries != manual {
			e2eError(fmt.Errorf("repeated award reset existing target choice: spell=%d entries=%v want=%v", id, entries, manual))
			return
		}
		e2eLog.Printf("BOOK REPEATED AWARD VERIFIED: spell=%d client_level=2 manual_target_preserved=true", id)
	})
	sc.Key(keybind.KeyB, name+" close repeated award book")
	sc.Wait(30, name+" settle repeated award book close")
}
