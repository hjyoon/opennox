//go:build !server

package opennox

import (
	"bytes"
	"crypto/sha256"
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
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type waypointCircleOriginalRow4B9790 struct {
	Frame uint32       `json:"frame"`
	Star  [5][2][2]int `json:"star"`
}

type waypointCircleOriginalOracle4B9790 struct {
	PE     string                            `json:"original_sha256"`
	Table  string                            `json:"direction_table_sha256"`
	Circle [16][2][2]int                     `json:"circle"`
	Rows   []waypointCircleOriginalRow4B9790 `json:"rows"`
}

func loadWaypointCircleOriginal4B9790(t *testing.T) waypointCircleOriginalOracle4B9790 {
	t.Helper()
	b, err := os.ReadFile("testdata/player_waypoint_circle_4b9790.json")
	if err != nil {
		t.Fatal(err)
	}
	var oracle waypointCircleOriginalOracle4B9790
	if err := json.Unmarshal(b, &oracle); err != nil {
		t.Fatal(err)
	}
	if oracle.PE != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || len(oracle.Rows) != 132 {
		t.Fatal("wrong original waypoint transcript")
	}
	// The original immutable 256-entry table is a separate input invariant.
	// Expected points come from executing PE32, not from this Go table/math.
	table := make([]byte, 8*len(sincosTable16))
	for i, p := range sincosTable16 {
		binary.LittleEndian.PutUint32(table[8*i:], uint32(p.X))
		binary.LittleEndian.PutUint32(table[8*i+4:], uint32(p.Y))
	}
	if fmt.Sprintf("%x", sha256.Sum256(table)) != oracle.Table {
		t.Fatal("native direction table differs from original")
	}
	return oracle
}

func waypointOriginalPoint4B9790(p [2]int, center image.Point) image.Point {
	return center.Add(image.Pt(p[0], p[1]))
}

func waypointOriginalBytes4B9790[T any](p *T) []byte {
	return bytes.Clone(unsafe.Slice((*byte)(unsafe.Pointer(p)), int(unsafe.Sizeof(*p))))
}

func TestPlayerWaypointCircleOriginal4B9790Geometry(t *testing.T) {
	oracle := loadWaypointCircleOriginal4B9790(t)
	for _, row := range oracle.Rows {
		for _, center := range []image.Point{{0, 0}, {50, 60}, {-30, -20}} {
			t.Run(fmt.Sprintf("frame-%08x/center-%d-%d", row.Frame, center.X, center.Y), func(t *testing.T) {
				dr, freeDR := alloc.New(client.Drawable{})
				t.Cleanup(freeDR)
				vp, freeVP := alloc.New(noxrender.Viewport{})
				t.Cleanup(freeVP)
				dr.PosVec = image.Pt(100, 200).Add(center)
				vp.World = image.Rect(100, 200, 740, 680)
				vp.Screen = image.Rect(17, 29, 657, 509)
				if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(dr.C()) <= 0xffffffff || uintptr(vp.C()) <= 0xffffffff) {
					t.Fatal("fixtures are not high-address native C-owned records")
				}
				beforeDR, beforeVP := waypointOriginalBytes4B9790(dr), waypointOriginalBytes4B9790(vp)
				got := playerWaypointDrawStateFor4B9790(dr, vp, row.Frame)
				if got.center != center {
					t.Fatalf("world origin, without Screen.Min: got=%v want=%v", got.center, center)
				}
				for i, expected := range oracle.Circle {
					// State stores cyclic geometry. PE32's closing draw reverses
					// the last edge; actual ordered alpha raster is checked below.
					if i == len(oracle.Circle)-1 {
						expected[0], expected[1] = expected[1], expected[0]
					}
					want := indicatorLine4B9790{
						from: waypointOriginalPoint4B9790(expected[0], center),
						to:   waypointOriginalPoint4B9790(expected[1], center),
					}
					if got.circle[i] != want {
						t.Errorf("original IMUL/SAR circle segment %d: got=%v want=%v", i, got.circle[i], want)
					}
				}
				for i, expected := range row.Star {
					want := indicatorLine4B9790{
						from: waypointOriginalPoint4B9790(expected[0], center),
						to:   waypointOriginalPoint4B9790(expected[1], center),
					}
					if got.lines[i] != want {
						t.Errorf("original star truncation segment %d: got=%v want=%v", i, got.lines[i], want)
					}
				}
				if !bytes.Equal(beforeDR, waypointOriginalBytes4B9790(dr)) || !bytes.Equal(beforeVP, waypointOriginalBytes4B9790(vp)) {
					t.Fatal("waypoint state mutated native input records")
				}
			})
		}
	}
}

func TestPlayerWaypointCircleOriginal4B9790NativeHeadless(t *testing.T) {
	oracle := loadWaypointCircleOriginal4B9790(t)
	for _, frame := range []uint32{0, 32, 77, 127} {
		row := oracle.Rows[frame]
		for _, alpha := range []byte{0, 64, 128, 255} {
			for _, clipped := range []bool{false, true} {
				for _, color := range []uint16{0x03e0, 0x7fff} {
					t.Run(fmt.Sprintf("frame-%d/alpha-%d/clip-%t/color-%04x", frame, alpha, clipped, color), func(t *testing.T) {
						s := server.New(nil, nil, strman.New())
						t.Cleanup(s.Close)
						s.SetFrame(frame)
						r := noxrender.NewRender(nil, s)
						t.Cleanup(r.Part.Free)
						d, freeData := noxrender.NewRenderData()
						t.Cleanup(freeData)
						r.SetData(d)
						pix := noximage.NewImage16(image.Rect(0, 0, 32, 32))
						for i := range pix.Pix {
							pix.Pix[i] = 0x2104
						}
						r.SetPixBuffer(pix)
						d.SetAlpha(alpha)
						d.SetAlphaEnabled(false)
						d.SetClip(true)
						d.SetClipRect(pix.Rect)
						d.SetClipRect2(image.Rect(0, 0, 31, 31))
						if clipped {
							d.SetClipRect2(image.Rect(14, 8, 23, 21))
						}
						d.SetColor2(noxcolor.RGBA5551(0x001f))
						c := &Client{Client: &client.Client{Server: s}, srv: &Server{Server: s}, r: NewNoxRender(r)}
						dr, freeDR := alloc.New(client.Drawable{})
						t.Cleanup(freeDR)
						dr.PosVec = image.Pt(116, 216)
						dr.DrawFuncPtr = legacy.Get_nox_thing_player_waypoint_draw()
						vp, freeVP := alloc.New(noxrender.Viewport{})
						t.Cleanup(freeVP)
						vp.World = image.Rect(100, 200, 132, 232)
						vp.Screen = image.Rect(7, 11, 39, 43)
						if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(dr.C()) <= 0xffffffff || uintptr(vp.C()) <= 0xffffffff) {
							t.Fatal("actual draw dispatch requires native high-address records")
						}
						beforeDR, beforeVP := waypointOriginalBytes4B9790(dr), waypointOriginalBytes4B9790(vp)
						cl := noxcolor.RGBA5551(color)
						oldColor := memmap.Uint32(0x85B3FC, 940)
						*memmap.PtrUint32(0x85B3FC, 940) = uint32(color)
						t.Cleanup(func() { *memmap.PtrUint32(0x85B3FC, 940) = oldColor })
						logic, other := s.Rand.Logic.Index(), s.Rand.Other.Index()
						if got, handled := c.callIndicatorDraw4B9790(dr, vp); !handled || got != 1 {
							t.Fatalf("production native draw dispatch: return=%d handled=%t", got, handled)
						}
						got := slices.Clone(pix.Pix)
						if d.IsAlphaEnabled() || d.Alpha() != alpha || d.Color2() != cl || s.Frame() != frame || s.Rand.Logic.Index() != logic || s.Rand.Other.Index() != other {
							t.Fatal("original final alpha/color/frame/RNG state differs")
						}
						// Replay the original ordered endpoints through the same
						// declared native line renderer, not a Windows raster oracle.
						for i := range pix.Pix {
							pix.Pix[i] = 0x2104
						}
						d.SetAlphaEnabled(true)
						drawOriginal := func(line [2][2]int) {
							r.DrawLine(waypointOriginalPoint4B9790(line[0], image.Pt(16, 16)), waypointOriginalPoint4B9790(line[1], image.Pt(16, 16)), cl)
						}
						for _, line := range oracle.Circle {
							drawOriginal(line)
						}
						for _, line := range row.Star {
							drawOriginal(line)
						}
						d.SetAlphaEnabled(false)
						if !slices.Equal(got, pix.Pix) {
							diff := 0
							for i, word := range pix.Pix {
								if word != got[i] {
									diff++
								}
							}
							t.Fatalf("complete raster differs from original ordered endpoints: %d pixels", diff)
						}
						if !bytes.Equal(beforeDR, waypointOriginalBytes4B9790(dr)) || !bytes.Equal(beforeVP, waypointOriginalBytes4B9790(vp)) {
							t.Fatal("actual draw mutated native input records")
						}
						sc := headless.New(pix.Size())
						t.Cleanup(func() { _ = sc.Close() })
						output, err := render.New(sc)
						if err != nil {
							t.Fatal(err)
						}
						presentParticleRegistration4B01B0(t, pix, output, sc, got, 0)
						t.Log("WAYPOINT ORIGINAL FRAME: real native dispatch/raster/upload with PE32 circle and star endpoints")
					})
				}
			}
		}
	}
}
