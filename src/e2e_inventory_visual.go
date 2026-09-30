package opennox

import (
	"fmt"
	"image"
	"unsafe"

	"github.com/opennox/libs/client/seat"
	"github.com/opennox/libs/noximage"
	"github.com/opennox/opennox/v1/client/gui"
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
		e2eLog.Printf("INVENTORY CLIPPING VERIFIED: visible=4x3 capacity=4x21 cells=%d last_row=%d offset=%d row_pixels=%v outside_pixels=0 restore=true", filled, lastRow, offset, bands)
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
