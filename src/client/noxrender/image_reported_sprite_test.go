package noxrender

import (
	"encoding/binary"
	"image"
	"log/slog"
	"path/filepath"
	"testing"

	"github.com/opennox/libs/bag"
	"github.com/opennox/libs/noxtest"
	"github.com/stretchr/testify/require"
)

func TestDrawImageReportedSprite(t *testing.T) {
	f, err := bag.Open(filepath.Join(noxtest.DataPath(t), "video.bag"))
	require.NoError(t, err)
	defer f.Close()
	imgs, err := f.Images()
	require.NoError(t, err)
	require.Greater(t, len(imgs), 15086)
	img := imgs[15086]
	data, err := img.Raw()
	require.NoError(t, err)
	hdr, sz, err := img.DecodeHeader()
	require.NoError(t, err)
	t.Logf("reported sprite: index=%d name=%q type=%d bytes=%d offset=%v size=%v", img.Index, img.Name, img.Type, len(data), hdr.Point, sz)
	require.GreaterOrEqual(t, len(data), 17)
	_, ok := skipPixdata(data[17:], sz.X, sz.Y)
	require.True(t, ok, "stock sprite stream must contain all declared rows")

	const pad = 20
	w, h := sz.X+2*pad, sz.Y+2*pad
	require.Less(t, w, 4096)
	require.Less(t, h, 4096)
	pos := image.Pt(pad, pad).Sub(hdr.Point)
	ref := newBackgroundPattern16(w, h)
	refBefore := append([]uint16(nil), ref.Pix...)
	rd := newRenderData(w, h)
	r := NewRender(slog.Default(), nil)
	r.SetPixBuffer(ref)
	r.SetData(rd)
	require.NotPanics(t, func() { r.DrawImage16(NewRawImage16(int(img.Type), data), pos) })
	require.NotEqual(t, refBefore, ref.Pix, "valid stock sprite must draw visible pixels")

	for top := 0; top < sz.Y; top++ {
		pix := newBackgroundPattern16(w, h)
		before := append([]uint16(nil), pix.Pix...)
		clip := image.Rect(pad, pad+top, pad+sz.X, pad+sz.Y)
		d := newRenderData(w, h)
		d.SetClip(true)
		d.SetClipRect(clip)
		r := NewRender(slog.Default(), nil)
		r.SetPixBuffer(pix)
		r.SetData(d)
		require.NotPanics(t, func() { r.DrawImage16(NewRawImage16(int(img.Type), data), pos) })
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				i := y*pix.Stride + x
				want := before[i]
				if image.Pt(x, y).In(clip) {
					want = ref.Pix[i]
				}
				require.Equal(t, want, pix.Pix[i], "top=%d pixel=(%d,%d)", top, x, y)
			}
		}
	}

	for phase := 0; phase < 2; phase++ {
		for cut := 0; cut <= len(data); cut++ {
			pix := newBackgroundPattern16(w, h)
			d := newRenderData(w, h)
			d.SetClip(true)
			d.SetClipRect(image.Rect(pad+1, pad+1, pad+sz.X, pad+sz.Y))
			r := NewRender(slog.Default(), nil)
			r.SetPixBuffer(pix)
			r.SetData(d)
			r.SetInterlacing(true, phase)
			require.NotPanics(t, func() {
				r.DrawImage16(NewRawImage16(int(img.Type), data[:cut]), pos)
			}, "phase=%d bytes=%d", phase, cut)
		}
	}
}

func TestStockSpritePixelRows(t *testing.T) {
	f, err := bag.Open(filepath.Join(noxtest.DataPath(t), "video.bag"))
	require.NoError(t, err)
	defer f.Close()
	imgs, err := f.Images()
	require.NoError(t, err)
	checked := 0
	for _, img := range imgs {
		switch img.Type & 0x3f {
		case 3, 4, 5, 6:
		default:
			continue
		}
		data, err := img.Raw()
		require.NoError(t, err, "sprite=%d name=%q", img.Index, img.Name)
		require.GreaterOrEqual(t, len(data), 17, "sprite=%d name=%q", img.Index, img.Name)
		w := int(int32(binary.LittleEndian.Uint32(data)))
		h := int(int32(binary.LittleEndian.Uint32(data[4:])))
		pix := data[17:]
		for y := 0; y < h; y++ {
			for covered := 0; covered < w; {
				run, next, ok := nextPixdataRun(pix)
				require.True(t, ok, "sprite=%d name=%q row=%d covered=%d size=(%d,%d)", img.Index, img.Name, y, covered, w, h)
				require.LessOrEqual(t, run.n, w-covered, "sprite=%d name=%q row=%d", img.Index, img.Name, y)
				covered += run.n
				pix = next
			}
		}
		checked++
	}
	t.Logf("validated %d stock sprite pixel streams", checked)
}
