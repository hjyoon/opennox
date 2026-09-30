package legacy

import (
	"image"
	"log/slog"
	"strconv"
	"testing"
	"unsafe"

	noxcolor "github.com/opennox/libs/color"
	"github.com/opennox/libs/noximage"

	"github.com/opennox/opennox/v1/client/noxrender"
)

func TestDrawClipRect49F6F0CEntry(t *testing.T) {
	for _, tc := range []struct {
		name   string
		bounds image.Rectangle
		x, y   int32
		w, h   int32
		want   image.Rectangle
	}{
		{"minimap", image.Rect(0, 0, 1024, 768), 0, 299, 170, 170, image.Rect(0, 299, 170, 469)},
		{"inventory viewport", image.Rect(0, 0, 1024, 768), 254, 13, 260, 150, image.Rect(254, 13, 514, 163)},
		{"clipped", image.Rect(10, 20, 100, 80), -5, 15, 120, 50, image.Rect(10, 20, 100, 65)},
		{"one pixel", image.Rect(0, 0, 10, 10), 9, 9, 1, 1, image.Rect(9, 9, 10, 10)},
		{"touch right", image.Rect(0, 0, 10, 10), 10, 0, 5, 5, image.Rectangle{}},
		{"touch bottom", image.Rect(0, 0, 10, 10), 0, 10, 5, 5, image.Rectangle{}},
		{"outside", image.Rect(0, 0, 10, 10), -20, -20, 5, 5, image.Rectangle{}},
		{"zero width", image.Rect(0, 0, 10, 10), 1, 1, 0, 5, image.Rectangle{}},
		{"zero height", image.Rect(0, 0, 10, 10), 1, 1, 5, 0, image.Rectangle{}},
		{"negative width", image.Rect(0, 0, 10, 10), 5, 1, -4, 5, image.Rectangle{}},
		{"negative height", image.Rect(0, 0, 10, 10), 1, 5, 5, -4, image.Rectangle{}},
		{"signed overflow", image.Rect(0, 0, 10, 10), 2147483647, 1, 2, 5, image.Rectangle{}},
	} {
		for _, enabled := range []bool{false, true} {
			t.Run(tc.name+map[bool]string{false: "/disabled", true: "/enabled"}[enabled], func(t *testing.T) {
				data, free := noxrender.NewRenderData()
				defer free()
				if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(data.C()) <= 0xFFFFFFFF {
					t.Fatal("fixture must exercise a C-owned pointer above 4 GiB")
				}
				data.SetRect3(tc.bounds)
				data.SetClip(enabled)
				oldClip := image.Rect(2, 3, 8, 9)
				oldClip2 := image.Rect(2, 3, 7, 8)
				data.SetClipRect(oldClip)
				data.SetClipRect2(oldClip2)
				before := *data
				got := drawClipRectCEntry49F6F0(data, tc.x, tc.y, tc.w, tc.h)
				if tc.want.Empty() {
					if got != 0 || *data != before {
						t.Fatalf("empty intersection result=%d or render data changed", got)
					}
					return
				}
				if got != 1 || data.ClipRect() != tc.want || data.ClipRect2() != (image.Rectangle{Min: tc.want.Min, Max: tc.want.Max.Sub(image.Pt(1, 1))}) {
					t.Fatalf("result=%d clip=%v inclusive=%v; want success=1 clip=%v", got, data.ClipRect(), data.ClipRect2(), tc.want)
				}
				data.SetClipRect(oldClip)
				data.SetClipRect2(oldClip2)
				if *data != before {
					t.Fatal("clip setup changed the enable flag, render bounds, or other render state")
				}
			})
		}
	}
}

func TestDrawClipRect49F6F0MinimapPixels(t *testing.T) {
	data, free := noxrender.NewRenderData()
	defer free()
	bounds := image.Rect(0, 0, 320, 240)
	data.SetRect3(bounds)
	data.SetClip(true)
	data.SetClipRect(bounds)
	data.SetClipRect2(image.Rect(0, 0, 319, 239))
	clip := image.Rect(0, 93, 53, 146)
	if got := drawClipRectCEntry49F6F0(data, 0, 93, 53, 53); got != 1 {
		t.Fatalf("minimap clip setup returned %d, want 1", got)
	}
	for _, radius := range []int{2, 3, 4, 5, 6} {
		t.Run(strconv.Itoa(radius), func(t *testing.T) {
			pix := noximage.NewImage16(bounds)
			r := noxrender.NewRender(slog.Default(), nil)
			r.SetData(data)
			r.SetPixBuffer(pix)
			cl := noxcolor.RGB5551Color(255, 255, 0)
			for _, p := range []image.Point{
				clip.Min.Add(image.Pt(20, 20)),
				image.Pt(0, 110), image.Pt(52, 110),
				image.Pt(20, 93), image.Pt(20, 145),
				image.Pt(70, 110), image.Pt(20, 70), image.Pt(20, 170),
			} {
				r.DrawPoint(p, radius, cl)
				r.DrawCircle(p.X, p.Y, 3, cl)
			}
			r.DrawLine(image.Pt(0, 70), image.Pt(100, 170), cl)
			changed := 0
			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				for x := bounds.Min.X; x < bounds.Max.X; x++ {
					if pix.Pix[pix.PixOffset(x, y)] == 0 {
						continue
					}
					changed++
					if !image.Pt(x, y).In(clip) {
						t.Fatalf("minimap marker leaked to (%d,%d) outside %v", x, y, clip)
					}
				}
			}
			if changed == 0 {
				t.Fatal("in-bounds minimap markers were not drawn")
			}
		})
	}
}
