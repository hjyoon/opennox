//go:build !server

package opennox

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"reflect"
	"slices"
	"testing"
	"unsafe"

	noxcolor "github.com/opennox/libs/color"
	"github.com/opennox/libs/noximage"

	core "github.com/opennox/opennox/v1/client/noxrender"
)

// These are setter-state contracts from retained 433E40/433CD0 instruction
// reads and the existing mode-1 scalar split receipt. They do not execute an
// original Windows framebuffer, C transition, GPU, GUI or personal saved game.
// The reference deliberately uses neither SplitColor nor ColorNRGBA.
func colorMultPackedReferenceR278(word uint16) [3]byte {
	return [3]byte{byte((word & 0x7c00) >> 7), byte((word & 0x03e0) >> 2), byte((word & 0x001f) * 8)}
}

func colorMultIntensityReferenceR278(rgb [3]byte) byte {
	return byte((76*(int(rgb[0])|7) + 150*(int(rgb[1])|7) + 28*(int(rgb[2])|7)) >> 8)
}

type colorMultFixtureR278 struct {
	r       *NoxRender
	data    *core.RenderData
	seed    []byte
	dword   int
	pix     *noximage.Image16
	pixSeed []uint16
}

func colorMultDataBytesR278(d *core.RenderData) []byte {
	return unsafe.Slice((*byte)(d.C()), int(unsafe.Sizeof(*d)))
}

func newColorMultFixtureR278(t *testing.T) *colorMultFixtureR278 {
	t.Helper()
	render := core.NewRender(nil, nil)
	t.Cleanup(render.Part.Free)
	data, freeData := core.NewRenderData()
	t.Cleanup(freeData)
	// NewRenderData resets selected fields; initialize every byte explicitly,
	// including unused fields and padding, to make neighbor checks meaningful.
	for i := range colorMultDataBytesR278(data) {
		colorMultDataBytesR278(data)[i] = byte(0x59 + i*37)
	}
	data.SetClip(true)
	data.SetClipRect(image.Rect(2, 3, 13, 17))
	data.SetClipRect2(image.Rect(2, 3, 12, 16))
	data.SetRect3(image.Rect(-3, -5, 23, 29))
	data.SetAlphaEnabled(true)
	data.SetAlpha(113)
	data.SetMultiply14(1)
	data.SetColorize17(1)
	data.SetFlag16(true)
	data.SetColorMultA(core.Color16{R: 19, G: 71, B: 137})
	for i := 0; i < 16; i++ {
		data.SetMaterialRGB(i, 17+i*11, 251-i*13, 29+i*7)
	}
	alpha, ok := reflect.TypeOf(*data).FieldByName("alpha")
	if !ok || alpha.Type.Kind() != reflect.Uint32 || alpha.Offset < 4 || alpha.Offset > unsafe.Sizeof(*data)-4 {
		t.Fatal("cannot locate stable typed alpha neighbor of original intensity DWORD")
	}
	// Original +1032 is the whole DWORD immediately before alpha. This works
	// both before and after unifying the two uint16 fields and never derives
	// expected ownership from Intensity258's current offset or current type.
	dword := int(alpha.Offset) - 4
	binary.LittleEndian.PutUint32(colorMultDataBytesR278(data)[dword:dword+4], 0xa17ec35b)
	render.SetData(data)
	pix := noximage.NewImage16(image.Rect(0, 0, 7, 3))
	for i := range pix.Pix {
		pix.Pix[i] = uint16(0x1357 + i*0x0311)
	}
	render.SetPixBuffer(pix)
	return &colorMultFixtureR278{
		r: NewNoxRender(render), data: data, seed: append([]byte(nil), colorMultDataBytesR278(data)...),
		dword: dword, pix: pix, pixSeed: append([]uint16(nil), pix.Pix...),
	}
}

func (f *colorMultFixtureR278) reset() {
	copy(colorMultDataBytesR278(f.data), f.seed)
}

func (f *colorMultFixtureR278) stateDifference(rgb [3]byte) string {
	var want core.RenderData
	copy(colorMultDataBytesR278(&want), f.seed)
	want.SetColorMultA(core.Color16{R: uint16(rgb[0]), G: uint16(rgb[1]), B: uint16(rgb[2])})
	want.SetFlag16(rgb == [3]byte{255, 255, 255})
	binary.LittleEndian.PutUint32(colorMultDataBytesR278(&want)[f.dword:f.dword+4], uint32(colorMultIntensityReferenceR278(rgb)))
	gotBytes, wantBytes := colorMultDataBytesR278(f.data), colorMultDataBytesR278(&want)
	if bytes.Equal(gotBytes, wantBytes) {
		return ""
	}
	for i := range gotBytes {
		if gotBytes[i] != wantBytes[i] {
			return fmt.Sprintf("byte %d got %#02x, want %#02x; logical DWORD got %#08x, want %#08x", i, gotBytes[i], wantBytes[i], binary.LittleEndian.Uint32(gotBytes[f.dword:f.dword+4]), binary.LittleEndian.Uint32(wantBytes[f.dword:f.dword+4]))
		}
	}
	panic("unequal renderer byte slices without a differing byte")
}

func (f *colorMultFixtureR278) assertSurfaceAndOwner(t *testing.T) {
	t.Helper()
	if f.r.Data() != f.data || f.r.PixBuffer() != f.pix || !slices.Equal(f.pix.Pix, f.pixSeed) {
		t.Fatal("color setter changed the renderer data owner, pixel owner, or pixels")
	}
}

func checkPackedColorMultAllWordsR278(t *testing.T, f *colorMultFixtureR278, call func(uint16) (int, bool)) {
	t.Helper()
	var badRGB, badFlag, badReturn, badState int
	var firstRGB, firstFlag, firstReturn, firstState string
	for word := 0; word < 65536; word++ {
		f.reset()
		rgb := colorMultPackedReferenceR278(uint16(word))
		gotReturn, hasReturn := call(uint16(word)) // exactly one actual adapter call
		gotRGB := f.data.ColorMultA()
		wantRGB := core.Color16{R: uint16(rgb[0]), G: uint16(rgb[1]), B: uint16(rgb[2])}
		if gotRGB != wantRGB {
			badRGB++
			if firstRGB == "" {
				firstRGB = fmt.Sprintf("word=%04x got=%+v want=%+v", word, gotRGB, wantRGB)
			}
		}
		if f.data.Flag16() { // packed five-bit components can never equal255
			badFlag++
			if firstFlag == "" {
				firstFlag = fmt.Sprintf("word=%04x unexpectedly set the full-byte white flag", word)
			}
		}
		if want := int(colorMultIntensityReferenceR278(rgb)); hasReturn && gotReturn != want {
			badReturn++
			if firstReturn == "" {
				firstReturn = fmt.Sprintf("word=%04x got=%d want=%d", word, gotReturn, want)
			}
		}
		if diff := f.stateDifference(rgb); diff != "" {
			badState++
			if firstState == "" {
				firstState = fmt.Sprintf("word=%04x %s", word, diff)
			}
		}
		f.assertSurfaceAndOwner(t)
	}
	// Aggregate rather than failing on word0's DWORD mismatch, so baseline
	// diagnostics independently expose the packed-max conversion divergence.
	for _, failure := range []struct {
		kind, first string
		count       int
	}{{"RGB", firstRGB, badRGB}, {"Flag16", firstFlag, badFlag}, {"return", firstReturn, badReturn}, {"whole RenderData", firstState, badState}} {
		if failure.count != 0 {
			t.Errorf("all65536 %s: %d mismatches; first %s", failure.kind, failure.count, failure.first)
		}
	}
}

func TestPackedColorMultSub433E40AllWordsR278(t *testing.T) {
	f := newColorMultFixtureR278(t)
	oldClient := noxClient
	noxClient = &Client{r: f.r}
	t.Cleanup(func() { noxClient = oldClient })
	checkPackedColorMultAllWordsR278(t, f, func(word uint16) (int, bool) {
		return sub433E40(noxcolor.RGBA5551(word)), true
	})
}

func TestPackedColorMultMethodAllWordsR278(t *testing.T) {
	f := newColorMultFixtureR278(t)
	checkPackedColorMultAllWordsR278(t, f, func(word uint16) (int, bool) {
		f.r.SetColorMultAndIntensity(noxcolor.RGBA5551(word))
		return 0, false // existing public method has no result
	})
}

func TestColorMultRGBWholeDWORDR278(t *testing.T) {
	f := newColorMultFixtureR278(t)
	for _, rgb := range [][3]byte{
		{255, 255, 255}, {248, 248, 248}, {0, 0, 0},
		{255, 0, 0}, {0, 255, 0}, {0, 0, 255},
		{0xa8, 0x50, 0xd0}, {0xaf, 0x57, 0xd7}, {1, 127, 254},
	} {
		f.reset()
		got := f.r.SetColorMultAndIntensityRGB(rgb[0], rgb[1], rgb[2])
		if want := colorMultIntensityReferenceR278(rgb); got != want {
			t.Errorf("direct RGB=%v return=%d want=%d", rgb, got, want)
		}
		if diff := f.stateDifference(rgb); diff != "" {
			t.Errorf("direct RGB=%v: %s", rgb, diff)
		}
		if rgb == [3]byte{255, 255, 255} && (!f.data.Flag16() || f.data.ColorMultA() != (core.Color16{R: 255, G: 255, B: 255})) {
			t.Error("direct full-byte white must retain255 multipliers and white flag")
		}
		f.assertSurfaceAndOwner(t)
	}
}
