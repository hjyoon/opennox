//go:build !server

package opennox

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"image"
	"os"
	"slices"
	"testing"
	"unsafe"

	noxcolor "github.com/opennox/libs/color"
	"github.com/opennox/libs/noximage"
	"github.com/opennox/libs/strman"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/client/render"
	"github.com/opennox/opennox/v1/client/seat/headless"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type vortexOriginalInput4B9F50 struct {
	Angle, Spin, Radius, Rise, Z, X, Y int
	Glow, Center                       uint32
}

type vortexOriginalRow4B9F50 struct {
	Input       vortexOriginalInput4B9F50
	Events      [][]json.RawMessage
	Returned    int
	FinalColor  uint32 `json:"final_color"`
	FinalZ      uint16 `json:"final_z"`
	FinalMotion uint32 `json:"final_motion"`
}

// Expected commands/state come from the unchanged PE32 Vortex, projection
// and FISTP bodies. Replay declared rendering services through this native
// renderer; this is not an original Windows framebuffer/full-world oracle.
func TestVortexOriginalColors4B9F50NativeHeadless(t *testing.T) {
	b, err := os.ReadFile("testdata/vortex_colors_4b9f50.json")
	if err != nil {
		t.Fatal(err)
	}
	var oracle struct {
		Original string `json:"original_sha256"`
		Rows     []vortexOriginalRow4B9F50
	}
	if err := json.Unmarshal(b, &oracle); err != nil {
		t.Fatal(err)
	}
	if oracle.Original != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || len(oracle.Rows) != 1030 {
		t.Fatal("wrong frozen original transcript")
	}
	for index, row := range oracle.Rows {
		for _, direct := range []bool{false, true} {
			t.Run(fmt.Sprintf("case-%04d/direct-%t", index, direct), func(t *testing.T) {
				s := server.New(nil, nil, strman.New())
				t.Cleanup(s.Close)
				s.SetFrame(7)
				cl := client.NewClient(nil, nil, s)
				t.Cleanup(cl.Close)
				t.Cleanup(func() { alloc.FreePtr(cl.Viewport().C()) })
				cl.Objs.Init(2)
				t.Cleanup(cl.Objs.Free)
				r := cl.Render()
				t.Cleanup(r.Part.Free)
				r.Part.RenderGlow = true
				d, freeData := noxrender.NewRenderData()
				t.Cleanup(freeData)
				r.SetData(d)
				pix := noximage.NewImage16(image.Rect(0, 0, 160, 160))
				r.SetPixBuffer(pix)
				c := &Client{Client: cl, srv: &Server{Server: s}, r: NewNoxRender(r)}
				oldClient := noxClient
				noxClient = c
				t.Cleanup(func() { noxClient = oldClient })
				vp, freeVP := alloc.New(noxrender.Viewport{})
				t.Cleanup(freeVP)
				vp.Screen = image.Rect(7, 11, 135, 139)
				vp.World = image.Rect(100, 200, 228, 328)
				vp.Size = vp.Screen.Size()
				dr := cl.Objs.New()
				if dr == nil {
					t.Fatal("native drawable allocation")
				}
				cl.Objs.Count, cl.Objs.List1 = 1, dr
				dr.ObjFlags, dr.DrawFuncPtr = 0x1000000, legacy.Get_nox_thing_vortex_draw()
				inp := row.Input
				dr.PosVec, dr.ZVal = image.Pt(inp.X, inp.Y), uint16(inp.Z)
				effect := dr.UnionEffect()
				effect.Field_108, effect.Field_109 = inp.Glow, inp.Center
				effect.Field_110, effect.Field_111 = uint32(inp.X), uint32(inp.Y)
				effect.Field_112 = uint32(byte(inp.Angle)) | uint32(byte(inp.Spin))<<8 | uint32(byte(inp.Radius))<<16 | uint32(byte(inp.Rise))<<24
				if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(dr.C()) <= 0xffffffff || uintptr(vp.C()) <= 0xffffffff) {
					t.Fatal("requires real C-owned high native records")
				}
				beforeDR, beforeVP := registeredEffectBytes4B7540(dr), registeredEffectBytes4B7540(vp)
				logic, other := s.Rand.Logic.Index(), s.Rand.Other.Index()
				initialData := *d
				seed := image.Pt(3, 9)
				prepare := func() {
					for i := range pix.Pix {
						pix.Pix[i] = 0x2104
					}
					*d = initialData
					d.SetColor2(noxcolor.RGBA5551(0x7fff))
					d.SetAlpha(128)
					d.SetAlphaEnabled(index&1 != 0)
					d.SetClip(true)
					d.SetClipRect(pix.Rect)
					d.SetClipRect2(pix.Rect)
					if index&2 != 0 {
						d.SetClipRect2(image.Rect(63, 51, 79, 83))
					}
					r.ClearPoints()
					r.AddPoint(seed)
				}
				prepare()
				var gotReturn int
				if direct {
					gotReturn = dr.CallDraw(vp)
				} else {
					var handled bool
					gotReturn, handled = c.callBubbleVortexDraw4B7540(dr, vp)
					if !handled {
						t.Fatal("world dispatch rejected Vortex")
					}
				}
				gotPixels, gotData := slices.Clone(pix.Pix), registeredEffectBytes4B7540(d)
				if gotReturn != row.Returned || cl.Objs.Count != row.Returned {
					t.Errorf("original return/lifetime: got=%d/%d want=%d", gotReturn, cl.Objs.Count, row.Returned)
				}
				if d.Color2() != noxcolor.RGBA5551(row.FinalColor) {
					t.Errorf("original final color: got=%04x want=%04x", d.Color2(), noxcolor.RGBA5551(row.FinalColor))
				}
				if p, ok := r.LastPoint(true); !ok || p != seed {
					t.Error("original point queue prefix was not preserved")
				}
				if cl.Objs.Count == 1 {
					wantDR := slices.Clone(beforeDR)
					binary.LittleEndian.PutUint16(wantDR[unsafe.Offsetof(dr.ZVal):], row.FinalZ)
					unionOffset := unsafe.Offsetof(dr.Union) + unsafe.Offsetof(client.DrawableUnionEffect{}.Field_112)
					binary.LittleEndian.PutUint32(wantDR[unionOffset:], row.FinalMotion)
					if !bytes.Equal(registeredEffectBytes4B7540(dr), wantDR) {
						t.Error("original complete live drawable state differs")
					}
					c.Nox_xxx_spriteDeleteStatic_45A4E0_drawable(dr)
				}
				if cl.Objs.Count != 0 || cl.Objs.List1 != nil {
					t.Error("real native drawable cleanup incomplete")
				}
				if !bytes.Equal(beforeVP, registeredEffectBytes4B7540(vp)) || s.Frame() != 7 || s.Rand.Logic.Index() != logic || s.Rand.Other.Index() != other {
					t.Error("viewport/frame/RNG changed")
				}
				// Replay exactly the original ordered primitive service calls.
				prepare()
				for _, raw := range row.Events {
					var command string
					if err := json.Unmarshal(raw[0], &command); err != nil {
						t.Fatal(err)
					}
					values := make([]int64, len(raw)-1)
					for i := range values {
						if err := json.Unmarshal(raw[i+1], &values[i]); err != nil {
							t.Fatal(err)
						}
					}
					switch command {
					case "rgb", "delete":
					case "glow":
						r.DrawGlow(image.Pt(int(values[0]), int(values[1])), noxcolor.RGBA5551(values[2]), int(values[3]), int(values[4]))
					case "color":
						d.SetColor2(noxcolor.RGBA5551(values[0]))
					case "dot":
						r.DrawPoint(image.Pt(int(values[0]), int(values[1])), int(values[2]), d.Color2())
					case "add":
						r.AddPoint(image.Pt(int(values[0]), int(values[1])))
					case "line":
						if !r.DrawLineFromPoints(d.Color2()) {
							t.Fatal("original line queue service failed")
						}
					default:
						t.Fatalf("undeclared original command: %s", command)
					}
				}
				if !slices.Equal(gotPixels, pix.Pix) || !bytes.Equal(gotData, registeredEffectBytes4B7540(d)) {
					t.Error("complete native raster/render state differs from original ordered services")
				}
				sc := headless.New(pix.Size())
				t.Cleanup(func() { _ = sc.Close() })
				output, err := render.New(sc)
				if err != nil {
					t.Fatal(err)
				}
				presentParticleRegistration4B01B0(t, pix, output, sc, gotPixels, 0)
				t.Log("VORTEX ORIGINAL FRAME: original colors/state/ordered services, real native callback/world route, complete raster/headless upload")
			})
		}
	}
}
