package noxrender

import (
	"encoding/binary"
	"fmt"
	"image"
	"slices"
	"testing"

	noxcolor "github.com/opennox/libs/color"
	"github.com/opennox/libs/noximage"
)

// GAME.EXE 004C79AB..004C79D9 uses row[y+i] and advances the source
// by 2*width bytes. It does not repeatedly overwrite row[y+1].
func TestUncompressedImageOriginal4C7860Rows(t *testing.T) {
	for _, typ := range []int{2, 7} {
		for _, size := range []image.Point{{1, 1}, {2, 2}, {3, 4}, {7, 3}} {
			for _, pos := range []image.Point{{0, 0}, {2, 3}, {4, 5}, {10, 8}} {
				for _, crop := range []uint32{0, 1} {
					t.Run(fmt.Sprintf("type-%d/size-%dx%d/pos-%dx%d/crop-%d", typ, size.X, size.Y, pos.X, pos.Y, crop), func(t *testing.T) {
						pix := noximage.NewImage16(image.Rect(0, 0, 24, 20))
						for i := range pix.Pix {
							pix.Pix[i] = 0x1234
						}
						d, free := NewRenderData()
						t.Cleanup(free)
						d.SetClip(true)
						d.SetClipRect(pix.Rect)
						d.SetColor2(noxcolor.RGBA5551(0x001f))
						r := NewRender(nil, nil)
						t.Cleanup(r.Part.Free)
						r.SetData(d)
						r.SetPixBuffer(pix)
						r.Set_dword_5d4594_3799484(int(crop))
						payload := make([]byte, 17+2*size.X*size.Y)
						binary.LittleEndian.PutUint32(payload, uint32(size.X))
						binary.LittleEndian.PutUint32(payload[4:], uint32(size.Y))
						binary.LittleEndian.PutUint32(payload[8:], 1)
						binary.LittleEndian.PutUint32(payload[12:], 2)
						palette := [...]uint16{0x7fff, 0x7c00, 0x03e0, 0x001f, 0x4210, 0x0000, 0x7fe0}
						for y := 0; y < size.Y; y++ {
							for x := 0; x < size.X; x++ {
								binary.LittleEndian.PutUint16(payload[17+2*(y*size.X+x):], palette[(y*3+x)%len(palette)])
							}
						}
						before := append([]byte(nil), payload...)
						want := append([]uint16(nil), pix.Pix...)
						for y := 0; y < size.Y-int(crop); y++ {
							for x := 0; x < size.X; x++ {
								want[(pos.Y+2+y)*24+pos.X+1+x] = palette[(y*3+x)%len(palette)]
							}
						}
						r.DrawImage16(NewRawImage16(typ, payload), pos)
						if !slices.Equal(payload, before) || uint16(noxcolor.ToRGBA5551Color(d.Color2())) != 0x001f {
							t.Fatal("uncompressed drawing changed source bytes or persistent color state")
						}
						if !slices.Equal(pix.Pix, want) {
							for i, v := range want {
								if pix.Pix[i] != v {
									t.Fatalf("004C79B7 row advance differs at (%d,%d): got=%04x original=%04x", i%24, i/24, pix.Pix[i], v)
								}
							}
						}
						t.Log("UNCOMPRESSED ROWS: complete distinct-color raster, original offsets and crop, source bytes unchanged")
					})
				}
			}
		}
	}
}
