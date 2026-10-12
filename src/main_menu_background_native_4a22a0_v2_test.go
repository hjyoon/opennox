//go:build !server

package opennox

import (
	"encoding/binary"
	"fmt"
	"image"
	"slices"
	"testing"
	"unsafe"

	noxcolor "github.com/opennox/libs/color"
	"github.com/opennox/libs/noximage"
	"github.com/opennox/libs/strman"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/client/render"
	"github.com/opennox/opennox/v1/client/seat/headless"
	"github.com/opennox/opennox/v1/server"
)

func mainMenuBackgroundFixture4A22A0(t *testing.T) (*Client, *gui.Window, *noximage.Image16, *render.Renderer, *headless.Seat) {
	t.Helper()
	s := server.New(nil, nil, strman.New())
	t.Cleanup(s.Close)
	r := noxrender.NewRender(nil, s)
	t.Cleanup(r.Part.Free)
	d, free := noxrender.NewRenderData()
	t.Cleanup(free)
	r.SetData(d)
	pix := noximage.NewImage16(image.Rect(0, 0, 24, 20))
	r.SetPixBuffer(pix)
	d.SetClip(true)
	d.SetClipRect(pix.Rect)
	d.SetClipRect2(image.Rect(0, 0, 23, 19))
	d.SetColor2(noxcolor.RGBA5551(0x001f))
	g := gui.New(r)
	t.Cleanup(g.DestroyAll)
	parent := g.NewWindowRaw(nil, gui.StatusEnabled, 5, 7, 18, 12, nil)
	win := g.NewWindowRaw(parent, gui.StatusEnabled, 3, 4, 4, 3, nil)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(win)) <= 0xffffffff {
		t.Fatal("actual window is not a native-width C-owned record")
	}
	c := &Client{Client: &client.Client{Server: s, GUI: g}, srv: &Server{Server: s}, r: NewNoxRender(r)}
	previousClient, previousEyes := noxClient, mainMenuEyes
	noxClient, mainMenuEyes = c, nil
	t.Cleanup(func() { noxClient, mainMenuEyes = previousClient, previousEyes })
	sc := headless.New(pix.Size())
	t.Cleanup(func() { _ = sc.Close() })
	window, err := render.New(sc)
	if err != nil {
		t.Fatal(err)
	}
	return c, win, pix, window, sc
}

func TestMainMenuBackgroundOriginal4A22A0OpaqueColorState(t *testing.T) {
	for _, cl := range []uint16{0, 0x7c00, 0x03e0, 0x7fff} {
		for _, clipped := range []bool{false, true} {
			for _, empty := range []bool{false, true} {
				t.Run(fmt.Sprintf("color-%04x/clip-%t/empty-%t", cl, clipped, empty), func(t *testing.T) {
					c, win, pix, output, sc := mainMenuBackgroundFixture4A22A0(t)
					draw, free := gui.NewWindowData()
					t.Cleanup(free)
					draw.SetBackgroundColor(noxcolor.RGBA5551(cl))
					if empty {
						win.SizeVal.X = 0
					}
					if clipped {
						c.r.Data().SetClipRect(image.Rect(10, 12, 12, 14))
					}
					beforeWindow := *win
					beforeDraw := *draw
					want := make([]uint16, len(pix.Pix))
					if !empty {
						for y := 11; y < 14; y++ {
							for x := 8; x < 12; x++ {
								if !clipped || x >= 10 && y >= 12 {
									want[y*24+x] = cl
								}
							}
						}
					}
					if got := drawMainMenuBackground(win, draw); got != 1 {
						t.Fatalf("original callback return=%d want=1", got)
					}
					if got := uint16(noxcolor.ToRGBA5551Color(c.r.Data().Color2())); got != cl {
						t.Fatalf("004A2321 must store background color before fill, even for an empty/clipped rect: got=%04x want=%04x", got, cl)
					}
					if *win != beforeWindow || *draw != beforeDraw || c.srv.Rand.Other.Index() != 0 || !slices.Equal(pix.Pix, want) {
						t.Fatal("original native window/data/RNG or complete background raster differs")
					}
					presentParticleRegistration4B01B0(t, pix, output, sc, want, 0)
					t.Log("MENU BACKGROUND COLOR: original persistent color state and one real complete raster/upload frame")
				})
			}
		}
	}
}

func TestMainMenuBackgroundOriginal4A22A0TransparentPreservesColor(t *testing.T) {
	c, win, pix, output, sc := mainMenuBackgroundFixture4A22A0(t)
	draw, free := gui.NewWindowData()
	t.Cleanup(free)
	draw.BgColorVal = 0x80000000 // original sentinel, not an arbitrary zero color.
	if drawMainMenuBackground(win, draw) != 1 || uint16(noxcolor.ToRGBA5551Color(c.r.Data().Color2())) != 0x001f {
		t.Fatal("original transparent branch must not change color2")
	}
	presentParticleRegistration4B01B0(t, pix, output, sc, make([]uint16, len(pix.Pix)), 0)
}

func TestMainMenuBackgroundOriginal4A22A0PartialEyeLoad(t *testing.T) {
	for _, hidden := range []bool{false, true} {
		for _, blink := range []uint32{0, 1, 2} {
			t.Run(fmt.Sprintf("hidden-%t/blink-%d", hidden, blink), func(t *testing.T) {
				c, win, pix, output, sc := mainMenuBackgroundFixture4A22A0(t)
				draw, free := gui.NewWindowData()
				t.Cleanup(free)
				draw.BgColorVal = 0x80000000
				payload := make([]byte, 25)
				binary.LittleEndian.PutUint32(payload, 2)
				binary.LittleEndian.PutUint32(payload[4:], 2)
				for i := 17; i < len(payload); i += 2 {
					binary.LittleEndian.PutUint16(payload[i:], 0x7fff)
				}
				img := noxrender.NewRawImage(2, payload)
				t.Cleanup(img.Free)
				mainMenuEyes = newMainMenuEyes()[:2]
				mainMenuEyes[0].image = nil
				mainMenuEyes[0].phaseTicks = 17
				mainMenuEyes[0].blinkCooldown = 9
				mainMenuEyes[1].image = img
				mainMenuEyes[1].spec.X, mainMenuEyes[1].spec.Y = 15, 3
				mainMenuEyes[1].hidden = hidden
				mainMenuEyes[1].blinkTicks = blink
				mainMenuEyes[1].phaseTicks = 5
				mainMenuEyes[1].blinkCooldown = 7
				before := append([]mainMenuEye(nil), mainMenuEyes...)
				want := make([]uint16, len(pix.Pix))
				if !hidden && blink == 0 {
					for y := 3; y < 5; y++ {
						for x := 15; x < 17; x++ {
							want[y*24+x] = 0x7fff
						}
					}
				}
				if drawMainMenuBackground(win, draw) != 1 {
					t.Fatal("original partial load callback return differs")
				}
				if !slices.Equal(mainMenuEyes, before) || c.srv.Rand.Other.Index() != 0 {
					t.Fatal("004A2347 skips all eye update/RNG when first image is nil")
				}
				presentParticleRegistration4B01B0(t, pix, output, sc, want, 0)
				t.Log("MENU PARTIAL IMAGE: original separate draw pass, no eye-state/RNG update, one real raster/upload frame")
			})
		}
	}
}
