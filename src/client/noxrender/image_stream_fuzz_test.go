package noxrender

import (
	"encoding/binary"
	"image"
	"log/slog"
	"testing"

	"github.com/opennox/libs/noximage"
)

func FuzzDrawImagePixdata(f *testing.F) {
	for _, payload := range [][]byte{
		nil, {2}, {2, 0}, {1, 3}, {2, 3, 1, 0, 2, 0, 3, 0},
		{2, 2, 1, 0, 2, 0, 2, 2, 3, 0, 4, 0},
		{4, 2, 1, 2, 5, 2, 0xff, 0xff, 0, 0},
	} {
		f.Add(uint8(1), uint8(1), int8(0), int8(-1), uint8(0), false, payload)
		f.Add(uint8(1), uint8(1), int8(-1), int8(0), uint8(0), true, payload)
	}
	f.Fuzz(func(t *testing.T, w, h uint8, x, y int8, typ uint8, interlaced bool, payload []byte) {
		data := make([]byte, 17, 17+len(payload))
		binary.LittleEndian.PutUint32(data, uint32(w%8)+1)
		binary.LittleEndian.PutUint32(data[4:], uint32(h%8)+1)
		data = append(data, payload...)
		pix := noximage.NewImage16(image.Rect(0, 0, 8, 8))
		d := newRenderData(8, 8)
		d.SetClip(true)
		r := NewRender(slog.Default(), nil)
		r.SetPixBuffer(pix)
		r.SetData(d)
		r.SetInterlacing(interlaced, int(typ>>7))
		imgTypes := [...]int{3, 4, 5, 6, 8}
		r.DrawImage16(NewRawImage16(imgTypes[int(typ)%len(imgTypes)], data), image.Pt(int(x%9), int(y%9)))
	})
}
