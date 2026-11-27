package legacy

import (
	"image"
	"unsafe"

	"github.com/opennox/opennox/v1/client/gui"
)

func windowFieldOffset46AE40(win *gui.Window, x, y int32) *gui.Window {
	if win != nil {
		win.DrawData().ImgPtVal = image.Pt(int(x), int(y))
	}
	return win
}

func windowFieldImage(win *gui.Window, value unsafe.Pointer, index int) int32 {
	if win == nil {
		return -2
	}
	draw := win.DrawData()
	var slot unsafe.Pointer
	switch index {
	case 1:
		slot = unsafe.Pointer(&draw.BgImageHnd)
	case 2:
		slot = unsafe.Pointer(&draw.HlImageHnd)
	case 3:
		slot = unsafe.Pointer(&draw.SelImageHnd)
	default:
		panic("invalid window image entry")
	}
	storeForeignPointer(slot, value)
	return 0
}

func windowFieldFont46AF40(win *gui.Window) unsafe.Pointer {
	if win == nil {
		return nil
	}
	return win.DrawData().FontPtr
}
