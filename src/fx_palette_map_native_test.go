//go:build !server

package opennox

import (
	"bytes"
	"encoding/json"
	"image"
	"os"
	"os/exec"
	"slices"
	"strings"
	"testing"
	"unsafe"

	noxcolor "github.com/opennox/libs/color"
	"github.com/opennox/libs/noximage"
	"github.com/opennox/libs/strman"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/client/render"
	"github.com/opennox/opennox/v1/client/seat/headless"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func fxPaletteOriginal199(t *testing.T) []uint32 {
	t.Helper()
	var raw struct {
		Calls [][4]uint32 `json:"calls"`
		Words []uint32    `json:"table_words"`
	}
	data, err := os.ReadFile("legacy/testdata/fx_palette_4b5cd0_original.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	if len(raw.Calls) != 256 || len(raw.Words) != 256 {
		t.Fatal("missing original palette")
	}
	words := make([]uint32, 256)
	for i, word := range raw.Words {
		index := int(word & 255)
		call := raw.Calls[index]
		if call[3] != word {
			t.Fatal("original token/store mismatch")
		}
		words[i] = noxcolor.RGB5551Color(byte(call[0]), byte(call[1]), byte(call[2])).Color32()
	}
	return words
}

func fxPaletteChild199(t *testing.T, mode string) bool {
	t.Helper()
	const variable = "OPENNOX_TEST_FX_PALETTE_R199"
	if os.Getenv(variable) == mode {
		legacy.InitBlobData()
		// This is the actual parent used by map startup, with the genuine
		// C-to-Go RGB5551 export. No palette writer/converter is substituted.
		legacy.Sub_4B64C0()
		return true
	}
	cmd := exec.Command(os.Args[0], "-test.run=^"+t.Name()+"$", "-test.count=1", "-test.v")
	cmd.Env = append(os.Environ(), variable+"="+mode)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("actual map palette %s: %v\n%s", mode, err, output)
	}
	if !bytes.Contains(output, []byte("FX PALETTE ACTUAL PASS: "+mode)) {
		t.Fatalf("missing actual map/raster receipt: %s", output)
	}
	t.Log(strings.TrimSpace(string(output)))
	return false
}

func TestFXPaletteActualMapInitializer4B64C0(t *testing.T) {
	if !fxPaletteChild199(t, "palette") {
		return
	}
	base := uintptr(1312500)
	words := unsafe.Slice(memmap.PtrUint32(0x5D4594, base), 256)
	if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(unsafe.Pointer(&words[0])) <= 0xffffffff {
		t.Fatal("actual high-address palette required")
	}
	want := fxPaletteOriginal199(t)
	for i, word := range words {
		if word != want[i] {
			t.Fatalf("actual parent/CGo RGB palette[%d]=%#x, original %#x", i, word, want[i])
		}
	}
	t.Log("FX PALETTE ACTUAL PASS: palette; real map parent -> complete initializer -> genuine RGB5551 export -> all 256 stores")
}

func TestFXPaletteActualMagicTailFrames4B5E10(t *testing.T) {
	if !fxPaletteChild199(t, "tail") {
		return
	}
	original := fxPaletteOriginal199(t)
	s := server.New(nil, nil, strman.New())
	t.Cleanup(s.Close)
	s.SetTickRate(60)
	s.SetFrame(100)
	s.Rand.Logic.Reset(93)
	s.Rand.Other.Reset(17)
	cl := client.NewClient(nil, nil, s)
	t.Cleanup(cl.Close)
	t.Cleanup(func() { alloc.FreePtr(cl.Viewport().C()) })
	cl.Objs.Init(4)
	t.Cleanup(cl.Objs.Free)
	r := cl.Render()
	t.Cleanup(r.Part.Free)
	d, freeData := noxrender.NewRenderData()
	t.Cleanup(freeData)
	r.SetData(d)
	pix := noximage.NewImage16(image.Rect(0, 0, 144, 144))
	r.SetPixBuffer(pix)
	c := &Client{Client: cl, srv: &Server{Server: s}, r: NewNoxRender(r)}
	oldClient, oldServer := noxClient, noxServer
	noxClient, noxServer = c, c.srv
	t.Cleanup(func() { noxClient, noxServer = oldClient, oldServer })
	vp, freeVP := alloc.New(noxrender.Viewport{})
	t.Cleanup(freeVP)
	vp.World, vp.Screen = image.Rect(100, 200, 228, 328), image.Rect(8, 9, 136, 137)
	vp.Size = pix.Size()
	beforeVP := registeredEffectBytes4B7540(vp)
	reset := func() {
		for i := range pix.Pix {
			pix.Pix[i] = 0x2104
		}
		*d = noxrender.RenderData{}
		d.SetClip(true)
		d.SetClipRect(vp.Screen)
		d.SetClipRect2(vp.Screen)
	}
	seat := headless.New(pix.Size())
	t.Cleanup(func() { _ = seat.Close() })
	output, err := render.New(seat)
	if err != nil {
		t.Fatal(err)
	}
	frames := 0
	for _, remaining := range []uint32{1, 7, 15, 30, 60} {
		for _, direct := range []bool{false, true} {
			dr := cl.Objs.New()
			if dr == nil {
				t.Fatal("native pool exhausted")
			}
			cl.Objs.Count, cl.Objs.List1 = 1, dr
			dr.ObjFlags = 0x1000000
			dr.PosVec = image.Pt(164, 280)
			dr.ZVal, dr.ZVal2 = 8, 1
			dr.Deadline = 100 + remaining
			dr.UnionEffect().Field_108, dr.UnionEffect().Field_109 = 140, 250
			dr.DrawFuncPtr = legacy.Get_nox_thing_magic_tail_link_draw()
			if !client.IsRegisteredDrawFunc(dr.DrawFuncPtr) || (unsafe.Sizeof(uintptr(0)) > 4 && (uintptr(dr.C()) <= 0xffffffff || uintptr(vp.C()) <= 0xffffffff)) {
				t.Fatal("actual native registered records required")
			}
			state := magicTailDrawStateFor4B5E10(dr, vp, 100, 60)
			color := noxcolor.RGBA5551(original[state.colorIndex])
			reset()
			d.SetColor2(color)
			d.SetAlphaEnabled(true)
			d.SetAlpha(128)
			c.r.DrawLine(state.point, state.tail, color)
			d.SetAlphaEnabled(false)
			want := slices.Clone(pix.Pix)
			reset()
			result := 0
			if direct {
				result = dr.CallDraw(vp)
			} else {
				var handled bool
				result, handled = c.callMagicDrawableDraw4B98A0(dr, vp)
				if !handled {
					t.Fatal("world tail dispatch not handled")
				}
			}
			if result != 1 || !slices.Equal(pix.Pix, want) {
				t.Fatalf("actual initialized palette tail differs from original RGB primitive raster: remaining=%d index=%d direct=%t", remaining, state.colorIndex, direct)
			}
			if !bytes.Equal(beforeVP, registeredEffectBytes4B7540(vp)) || dr.PosVec != image.Pt(164, 280) || dr.ZVal != 8 || dr.ZVal2 != 1 || dr.Deadline != 100+remaining || dr.UnionEffect().Field_108 != 140 || dr.UnionEffect().Field_109 != 250 || s.Frame() != 100 || s.TickRate() != 60 || s.Rand.Logic.Index() != 93 || s.Rand.Other.Index() != 17 || cl.Objs.Count != 1 || cl.Objs.List1 != dr {
				t.Fatal("tail changed non-light inputs/frame/RNG/native pool")
			}
			presentParticleRegistration4B01B0(t, pix, output, seat, want, 0)
			c.Nox_xxx_spriteDeleteStatic_45A4E0_drawable(dr)
			if cl.Objs.Count != 0 || cl.Objs.List1 != nil {
				t.Fatal("actual pool cleanup incomplete")
			}
			frames++
		}
	}
	t.Logf("FX PALETTE ACTUAL PASS: tail; %d real world/C callback full-raster/headless-upload frames using actual map-initialized palette and sealed original RGB", frames)
}
