package opennox

import (
	"fmt"
	"image"
	"unsafe"

	"github.com/opennox/libs/client/seat"
	"github.com/opennox/libs/noximage"
	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

// mode 0 checks the top row; mode 1 checks a scrolled viewport.
func (sc *e2eScenario) CheckInventoryClipping(mode int, name string) {
	sc.addWhen(0, name, 1200, func() bool {
		return legacy.Nox_client_inventoryAnimationState() == 2 && legacy.InventoryWindow() != nil
	}, func() {
		columns, rows, filled, lastRow := legacy.InventoryGridMetrics()
		offset := legacy.InventoryScrollOffset()
		if columns != 4 || rows != 21 || filled < 13 || lastRow < 3 || (mode == 0 && offset != 0) || (mode == 1 && offset <= 0) {
			e2eError(fmt.Errorf("inventory grid/scroll contract failed: columns=%d capacityRows=%d filled=%d lastRow=%d offset=%d mode=%d", columns, rows, filled, lastRow, offset, mode))
			return
		}
		data := noxClient.r.Data()
		before := *data
		live := noxClient.r.PixBuffer()
		pix := noximage.NewImage16(live.Rect)
		noxClient.r.SetPixBuffer(pix)
		defer noxClient.r.SetPixBuffer(live)
		pos := legacy.InventoryWindow().GlobalPos().Add(image.Pt(254, 13))
		clip := image.Rectangle{Min: pos, Max: pos.Add(image.Pt(260, 150))}
		if !legacy.InventoryDrawTrayClipped(pos) {
			e2eError(fmt.Errorf("inventory stock tray clip setup failed"))
			return
		}
		bands := [3]int{}
		outside := 0
		for y := pix.Rect.Min.Y; y < pix.Rect.Max.Y; y++ {
			for x := pix.Rect.Min.X; x < pix.Rect.Max.X; x++ {
				if pix.Pix[pix.PixOffset(x, y)] == 0 {
					continue
				}
				if !image.Pt(x, y).In(clip) {
					outside++
				} else if x >= pos.X+60 {
					bands[(y-pos.Y)/50]++
				}
			}
		}
		if outside != 0 || bands[0] == 0 || bands[1] == 0 || bands[2] == 0 || data.Clip() != before.Clip() || data.ClipRect() != before.ClipRect() || data.ClipRect2() != before.ClipRect2() {
			e2eError(fmt.Errorf("inventory tray pixel/restore contract failed: rect=%v rowPixels=%v outside=%d", clip, bands, outside))
			return
		}
		if err := e2eInventoryIconPixels(pix, clip); err != nil {
			e2eError(err)
			return
		}
		if err := e2eInventoryWeaponIconPixels(false); err != nil {
			e2eError(err)
			return
		}
		e2eLog.Printf("INVENTORY CLIPPING VERIFIED: visible=4x3 capacity=4x21 cells=%d last_row=%d offset=%d row_pixels=%v outside_pixels=0 restore=true", filled, lastRow, offset, bands)
	})
}

// Compare the stock Sword/GreatSword silhouettes with the real tray pass.
// Background pixels alone cannot demonstrate that inventory icons are visible.
// The reference uses an explicitly native-width viewport; client cells, their
// positions (already set by the tray pass), and draw callbacks are unchanged.
func e2eInventoryIconPixels(tray *noximage.Image16, clip image.Rectangle) error {
	sword := uint32(noxClient.Things.IndByID("Sword"))
	greatSword := uint32(noxClient.Things.IndByID("GreatSword"))
	viewport := noxrender.Viewport{Screen: tray.Rect, Size: tray.Rect.Size()}
	offset := legacy.InventoryScrollOffset()
	icons := 0
	defer noxClient.r.SetPixBuffer(tray)
	for row := 0; row < 20; row++ {
		for column := 0; column < 4; column++ {
			position := clip.Min.Add(image.Pt(60+50*column+25, 50*row-offset+25))
			if !position.In(clip) {
				continue
			}
			drawable := legacy.InventoryCellDrawable(column, row)
			if drawable == nil || (drawable.TypeIDVal != sword && drawable.TypeIDVal != greatSword) {
				continue
			}
			if drawable.Pos() != position {
				return fmt.Errorf("inventory icon position: cell=%d/%d pos=%v want=%v", column, row, drawable.Pos(), position)
			}
			reference := noximage.NewImage16(tray.Rect)
			noxClient.r.SetPixBuffer(reference)
			drawable.CallDraw(&viewport)
			pixels, matching := 0, 0
			for y := clip.Min.Y; y < clip.Max.Y; y++ {
				for x := clip.Min.X; x < clip.Max.X; x++ {
					index := reference.PixOffset(x, y)
					if reference.Pix[index] == 0 {
						continue
					}
					pixels++
					if reference.Pix[index] == tray.Pix[index] {
						matching++
					}
				}
			}
			// These two stock items use opaque static silhouettes. Every pixel
			// must survive the real tray callback, both before and after scrolling.
			if pixels == 0 || matching != pixels {
				return fmt.Errorf("inventory icon missing: cell=%d/%d type=%d pos=%v reference_pixels=%d matching_pixels=%d", column, row, drawable.TypeIDVal, drawable.Pos(), pixels, matching)
			}
			icons++
			e2eLog.Printf("INVENTORY ICON VERIFIED: cell=%d/%d type=%d reference_pixels=%d matching_pixels=%d", column, row, drawable.TypeIDVal, pixels, matching)
		}
	}
	if icons == 0 {
		return fmt.Errorf("no visible stock inventory icons were checked")
	}
	return nil
}

func e2eInventoryWeaponIconPixels(alternate bool) error {
	drawable := legacy.InventoryWeaponDrawable(alternate)
	win := legacy.InventoryWeaponWindow(alternate)
	if drawable == nil || win == nil {
		return fmt.Errorf("inventory weapon slot is empty: alternate=%t", alternate)
	}
	live := noxClient.r.PixBuffer()
	actual := noximage.NewImage16(live.Rect)
	noxClient.r.SetPixBuffer(actual)
	defer noxClient.r.SetPixBuffer(live)
	if !legacy.InventoryDrawWeapon(alternate) {
		return fmt.Errorf("inventory weapon callback did not draw: alternate=%t", alternate)
	}
	reference := noximage.NewImage16(live.Rect)
	noxClient.r.SetPixBuffer(reference)
	viewport := noxrender.Viewport{Screen: live.Rect, Size: live.Rect.Size()}
	drawable.CallDraw(&viewport)
	clip := live.Rect
	if alternate {
		// The alternate slot prints its hotkey after the icon at y+41.
		clip = image.Rectangle{Min: win.GlobalPos(), Max: win.GlobalPos().Add(image.Pt(win.Size().X, 41))}.Intersect(clip)
	}
	pixels, matching := 0, 0
	for y := clip.Min.Y; y < clip.Max.Y; y++ {
		for x := clip.Min.X; x < clip.Max.X; x++ {
			index := reference.PixOffset(x, y)
			if reference.Pix[index] == 0 {
				continue
			}
			pixels++
			if reference.Pix[index] == actual.Pix[index] {
				matching++
			}
		}
	}
	if pixels == 0 || pixels != matching {
		return fmt.Errorf("inventory weapon icon missing: alternate=%t type=%d pos=%v reference_pixels=%d matching_pixels=%d", alternate, drawable.TypeIDVal, drawable.Pos(), pixels, matching)
	}
	e2eLog.Printf("INVENTORY WEAPON ICON VERIFIED: alternate=%t type=%d reference_pixels=%d matching_pixels=%d", alternate, drawable.TypeIDVal, pixels, matching)
	return nil
}

func (sc *e2eScenario) DragInventoryItemToAlternate(typeID, name string) {
	sc.add(0, name, func() {
		typeIndex := uint32(noxClient.Things.IndByID(typeID))
		win := legacy.InventoryWindow()
		if win == nil || legacy.Sub_4675B0() == 5 {
			e2eError(fmt.Errorf("inventory item cannot be dragged to alternate slot: item=%s", typeID))
			return
		}
		// The initial Sword is already equipped. Choose a visible, unequipped
		// cell so the normal drop path selects a secondary weapon, not dequip.
		column, row := -1, -1
		offset := legacy.InventoryScrollOffset()
		for r := 0; r < 20 && row == -1; r++ {
			y := 50*r - offset + 25
			if y < 0 || y >= 150 {
				continue
			}
			for c := 0; c < 4; c++ {
				drawable := legacy.InventoryCellDrawable(c, r)
				if drawable != nil && drawable.TypeIDVal == typeIndex && !legacy.InventoryCellEquipped(c, r) {
					column, row = c, r
					break
				}
			}
		}
		if row == -1 {
			e2eError(fmt.Errorf("no visible unequipped inventory item: item=%s", typeID))
			return
		}
		pos := win.GlobalPos().Add(image.Pt(314+50*column+25, 13+50*row-offset+25))
		e2eLog.Printf("INVENTORY ALTERNATE DRAG: item=%s cell=%d/%d start=%v mode=%d captured=%p", typeID, column, row, pos, legacy.Sub_4675B0(), noxClient.GUI.Captured())
		e2eQueueInput(&seat.MouseMoveEvent{Pos: pos})
	})
	sc.Input(2, name+" press item", &seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: true})
	sc.addWhen(2, name+" wait for dragged item", 120, legacy.Nox_client_inventoryHasDragged, func() {
		win := legacy.InventoryWeaponWindow(true)
		if win == nil {
			e2eError(fmt.Errorf("alternate weapon window unavailable"))
			return
		}
		pos := win.GlobalPos().Add(win.Size().Div(2))
		e2eLog.Printf("INVENTORY ALTERNATE DRAGGING: item=%s end=%v hidden=%t captured=%p target=%p", typeID, pos, win.GetFlags().IsHidden(), noxClient.GUI.Captured(), noxClient.GUI.Captured().ChildByPos(pos))
		e2eQueueInput(&seat.MouseMoveEvent{Pos: pos})
	})
	sc.Input(3, name+" release in alternate slot", &seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: false})
	sc.addWhen(0, name+" wait for secondary weapon report", 1200, func() bool {
		drawable := legacy.InventoryWeaponDrawable(true)
		item := noxServer.SecondaryWeapon53AB90(noxServer.Players.HostUnit())
		return drawable != nil && drawable.TypeIDVal == uint32(noxClient.Things.IndByID(typeID)) && item != nil && int(item.TypeInd) == noxServer.Types.ByID(typeID).Ind() && item.NetCode&0x7fff == drawable.NetCode32&0x7fff
	}, nil)
	sc.add(10, name+" verify alternate icon after packet synchronization", func() {
		drawable := legacy.InventoryWeaponDrawable(true)
		item := noxServer.SecondaryWeapon53AB90(noxServer.Players.HostUnit())
		if drawable == nil || item == nil || drawable.TypeIDVal != uint32(noxClient.Things.IndByID(typeID)) || item.NetCode&0x7fff != drawable.NetCode32&0x7fff || legacy.Nox_client_inventoryHasDragged() {
			e2eError(fmt.Errorf("alternate weapon selection was not retained after synchronization: item=%s", typeID))
			return
		}
		if err := e2eInventoryWeaponIconPixels(true); err != nil {
			e2eError(err)
		}
		e2eLog.Printf("INVENTORY ALTERNATE SYNCHRONIZED: item=%s server_netcode=%d client_netcode=%d dragged=false", typeID, item.NetCode, drawable.NetCode32)
	})
}

func (sc *e2eScenario) ClickInventoryScroll(down bool, name string) {
	sc.add(0, name, func() {
		button := legacy.InventoryScrollButton(down)
		if button == nil {
			e2eError(fmt.Errorf("inventory scroll button unavailable"))
			return
		}
		pos := button.GlobalPos().Add(button.Size().Div(2))
		e2eLog.Printf("INVENTORY SCROLL CLICK: down=%t pos=%v offset=%d", down, pos, legacy.InventoryScrollOffset())
		e2eQueueInput(&seat.MouseMoveEvent{Pos: pos}, &seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: true})
	})
	sc.Input(1, "", &seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: false})
}

func (sc *e2eScenario) ClickInventoryIdentify(name string) {
	sc.add(0, name, func() {
		win := legacy.InventoryWindow()
		if win == nil || legacy.Nox_client_inventoryAnimationState() != 2 {
			e2eError(fmt.Errorf("inventory is not open"))
			return
		}
		pos := win.GlobalPos().Add(image.Pt(284, 88))
		e2eQueueInput(&seat.MouseMoveEvent{Pos: pos}, &seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: true})
	})
	sc.Input(1, "", &seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: false})
}

func e2eInventoryDescription() (*gui.Window, *gui.ScrollListBoxData) {
	parent := legacy.InventoryIdentifyWindow()
	if parent == nil || parent.GetFlags().IsHidden() {
		return nil, nil
	}
	win := parent.ChildByID(9156)
	if win == nil || win.WidgetData == nil {
		return nil, nil
	}
	return win, (*gui.ScrollListBoxData)(win.WidgetData)
}

// mode 0 checks the initial description; mode 1 requires actual scrolling.
func (sc *e2eScenario) CheckItemDescriptionClipping(mode int, name string) {
	sc.addWhen(0, name, 1200, func() bool {
		win, d := e2eInventoryDescription()
		return win != nil && d.Items != nil && d.Field_11_0 != 0
	}, func() {
		win, d := e2eInventoryDescription()
		if (mode == 0 && d.Field_13_1 != 0) || (mode == 1 && d.Field_13_1 == 0) {
			e2eError(fmt.Errorf("description did not scroll: total=%d viewport=%d", d.Field_10, d.Field_13_0))
			return
		}
		data := noxClient.r.Data()
		before := *data
		live := noxClient.r.PixBuffer()
		pix := noximage.NewImage16(live.Rect)
		noxClient.r.SetPixBuffer(pix)
		defer noxClient.r.SetPixBuffer(live)
		// Render the real 9156 callback with unchanged stock description rows.
		win.Draw()
		clip := image.Rectangle{Min: win.GlobalPos(), Max: win.GlobalPos().Add(win.Size())}
		if win.DrawData().Text() != "" {
			clip.Min.Y += noxClient.r.FontHeight(win.DrawData().Font()) + 1
		}
		if d.Field_3 != 0 {
			clip.Max.X -= 10
		}
		inside, outside := 0, 0
		for y := pix.Rect.Min.Y; y < pix.Rect.Max.Y; y++ {
			for x := pix.Rect.Min.X; x < pix.Rect.Max.X; x++ {
				if pix.Pix[pix.PixOffset(x, y)] == 0 {
					continue
				}
				if image.Pt(x, y).In(clip) {
					inside++
				} else {
					outside++
				}
			}
		}
		if outside != 0 || inside == 0 || data.Clip() != before.Clip() || data.ClipRect() != before.ClipRect() || data.ClipRect2() != before.ClipRect2() {
			e2eError(fmt.Errorf("description pixel/restore contract failed: rect=%v inside=%d outside=%d", clip, inside, outside))
			return
		}
		items := unsafe.Slice(d.Items, int(d.Field_11_0))
		for i := range items {
			e2eLog.Printf("DESCRIPTION ROW: index=%d bottom=%d height=%d text=%q", i, items[i].Field_0, items[i].Field_130, alloc.GoString16(&items[i].Text[0]))
		}
		e2eLog.Printf("DESCRIPTION CLIPPING VERIFIED: rect=%v rows=%d total=%d viewport=%d offset=%d inside_pixels=%d outside_pixels=0 restore=true", clip, d.Field_11_0, d.Field_10, d.Field_13_0, d.Field_13_1, inside)
	})
}

func (sc *e2eScenario) ClickItemDescriptionScroll(name string) {
	sc.add(0, name, func() {
		_, d := e2eInventoryDescription()
		if d == nil || d.Field_8 == nil {
			e2eError(fmt.Errorf("description scroll button unavailable"))
			return
		}
		button := legacy.AsWindowP(d.Field_8)
		pos := button.GlobalPos().Add(button.Size().Div(2))
		e2eLog.Printf("DESCRIPTION SCROLL CLICK: pos=%v offset=%d total=%d viewport=%d", pos, d.Field_13_1, d.Field_10, d.Field_13_0)
		e2eQueueInput(&seat.MouseMoveEvent{Pos: pos}, &seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: true})
	})
	sc.Input(1, "", &seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: false})
}
