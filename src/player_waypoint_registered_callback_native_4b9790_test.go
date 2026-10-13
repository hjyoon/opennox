//go:build !server

package opennox

import (
	"bytes"
	"fmt"
	"image"
	"os"
	"os/exec"
	"reflect"
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
	"github.com/opennox/opennox/v1/internal/binfile"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/server"
)

const registeredWaypointChild4B9790 = "OPENNOX_TEST_REGISTERED_PLAYER_WAYPOINT_R169"

type registeredWaypointCase4B9790 struct {
	alpha byte
	color uint16
	style string
}

func (c registeredWaypointCase4B9790) name() string {
	return fmt.Sprintf("alpha-%d/color-%04x/%s", c.alpha, c.color, c.style)
}

var registeredWaypointCases4B9790 = func() []registeredWaypointCase4B9790 {
	var out []registeredWaypointCase4B9790
	for _, alpha := range []byte{0, 64, 128, 255} {
		for _, color := range []uint16{0x03e0, 0x7fff} {
			for _, style := range []string{"center", "clip-corner", "edge-right", "outside", "screen-offset"} {
				out = append(out, registeredWaypointCase4B9790{alpha, color, style})
			}
		}
	}
	return out
}()

// Each child executes all 132 unchanged original frames through the real DRAW
// registry and native-width C entry. Expected endpoints come from unhooked
// PE32 waypoint/circle execution, while pixels use the declared native line
// renderer: this is not an independent original Windows framebuffer oracle.
func TestRegisteredPlayerWaypointCallbackNative4B9790(t *testing.T) {
	if name := os.Getenv(registeredWaypointChild4B9790); name != "" {
		testRegisteredWaypointChild4B9790(t, name)
		return
	}
	for _, def := range registeredWaypointCases4B9790 {
		name := def.name()
		t.Run(strings.ReplaceAll(name, "/", "-"), func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run=^TestRegisteredPlayerWaypointCallbackNative4B9790$", "-test.v")
			cmd.Env = append(os.Environ(), registeredWaypointChild4B9790+"="+name)
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("registered native waypoint callback failed: %v\n%s", err, out)
			}
			if bytes.Count(out, []byte("REGISTERED WAYPOINT FRAME PASS:")) != 132 {
				t.Fatalf("missing original/native/C-entry/raster/headless frame receipts: %s", out)
			}
			t.Logf("REGISTERED WAYPOINT CALLBACK PASS: %s frames=132", name)
		})
	}
}

func registeredWaypointType4B9790(t *testing.T, c *Client) *client.ObjectType {
	t.Helper()
	var raw []byte
	for _, value := range []string{"PlayerWaypoint", "DRAW", "PlayerWaypointDraw"} {
		raw = append(raw, byte(len(value)))
		raw = append(raw, value...)
	}
	for len(raw)%8 != 0 {
		raw = append(raw, 0)
	}
	raw = append(raw, make([]byte, 8)...)
	raw = append(raw, 0)
	buf, _ := alloc.CloneSlice(raw)
	f := binfile.NewMemFile(unsafe.Pointer(&buf[0]), len(buf))
	scratch, freeScratch := alloc.Make([]byte{}, 256)
	err := c.Things.ReadType(f, scratch)
	remaining := len(f.Data())
	freeScratch()
	f.Free()
	if err != nil || remaining != 0 {
		t.Fatalf("actual DRAW parser failed: %v unread=%d", err, remaining)
	}
	typ := c.Things.TypeByID("PlayerWaypoint")
	if typ == nil || typ.DrawFunc != legacy.Get_nox_thing_player_waypoint_draw() || !client.IsRegisteredDrawFunc(typ.DrawFunc) || typ.DrawData != nil || client.DrawableDataKind(typ.DrawFunc) != 0 {
		t.Fatal("actual kind0 registered waypoint parser failed")
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(typ.C()) <= 0xffffffff {
		t.Fatal("actual native type must be above 4 GiB")
	}
	return typ
}

func testRegisteredWaypointChild4B9790(t *testing.T, name string) {
	var def registeredWaypointCase4B9790
	found := false
	for _, candidate := range registeredWaypointCases4B9790 {
		if candidate.name() == name {
			def, found = candidate, true
			break
		}
	}
	if !found {
		t.Fatal("unknown frozen waypoint case")
	}
	oracle := loadWaypointCircleOriginal4B9790(t)
	t.Chdir(t.TempDir())
	handles.Init()
	t.Cleanup(handles.Release)
	s := server.New(nil, nil, strman.New())
	t.Cleanup(s.Close)
	s.SetTickRate(60)
	s.Rand.Logic.Reset(93)
	s.Rand.Other.Reset(17)
	cl := client.NewClient(nil, nil, s)
	t.Cleanup(cl.Close)
	t.Cleanup(func() { alloc.FreePtr(cl.Viewport().C()) })
	cl.Objs.Init(4)
	t.Cleanup(cl.Objs.Free)
	cl.Things.Nox_things_free_44C580_B()
	t.Cleanup(cl.Things.Nox_things_free_44C580_B)
	r := cl.Render()
	t.Cleanup(r.Part.Free)
	d, freeData := noxrender.NewRenderData()
	t.Cleanup(freeData)
	r.SetData(d)
	pix := noximage.NewImage16(image.Rect(0, 0, 64, 64))
	r.SetPixBuffer(pix)
	c := &Client{Client: cl, srv: &Server{Server: s}, r: NewNoxRender(r)}
	oldClient, oldServer := noxClient, noxServer
	noxClient, noxServer = c, c.srv
	t.Cleanup(func() { noxClient, noxServer = oldClient, oldServer })
	typ := registeredWaypointType4B9790(t, c)
	beforeType := registeredEffectBytes4B7540(typ)
	vp, freeVP := alloc.New(noxrender.Viewport{})
	t.Cleanup(freeVP)
	vp.World, vp.Screen = image.Rect(100, 200, 164, 264), image.Rect(7, 11, 71, 75)
	vp.Size = pix.Size()
	center, clip := image.Pt(32, 32), image.Rect(0, 0, 63, 63)
	switch def.style {
	case "clip-corner":
		center, clip = image.Pt(4, 4), image.Rect(3, 3, 32, 32)
	case "edge-right":
		center = image.Pt(62, 32)
	case "outside":
		center = image.Pt(96, 96)
	case "screen-offset":
		vp.Screen = image.Rect(17, 29, 81, 93)
	}
	beforeVP := registeredEffectBytes4B7540(vp)
	oldColor := memmap.Uint32(0x85B3FC, 940)
	*memmap.PtrUint32(0x85B3FC, 940) = uint32(def.color)
	t.Cleanup(func() { *memmap.PtrUint32(0x85B3FC, 940) = oldColor })
	color := noxcolor.RGBA5551(def.color)
	sc := headless.New(pix.Size())
	t.Cleanup(func() { _ = sc.Close() })
	output, err := render.New(sc)
	if err != nil {
		t.Fatal(err)
	}
	resetRender := func() {
		for i := range pix.Pix {
			pix.Pix[i] = 0x2104
		}
		*d = noxrender.RenderData{}
		d.SetAlpha(def.alpha)
		d.SetAlphaEnabled(false)
		d.SetClip(true)
		d.SetClipRect(pix.Rect)
		d.SetClipRect2(clip)
		d.SetColor2(noxcolor.RGBA5551(0x001f))
	}
	for _, row := range oracle.Rows {
		s.SetFrame(row.Frame)
		dr := c.Nox_xxx_spriteLoadAdd_45A360_drawable(c.Things.IndByID("PlayerWaypoint"), vp.World.Min.Add(center))
		if dr == nil || dr.DrawFuncPtr != typ.DrawFunc || !client.IsRegisteredDrawFunc(dr.DrawFuncPtr) {
			t.Fatal("real type/pool/spawn did not retain registered callback")
		}
		if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(dr.C()) <= 0xffffffff || uintptr(vp.C()) <= 0xffffffff) {
			t.Fatal("actual C-owned drawable/viewport must be above 4 GiB")
		}
		dr.ZVal, dr.VelZ = 17, -3
		dr.UnionEffect().Field_108, dr.UnionEffect().Field_109 = 0x12345678, 0x90abcdef
		c.Objs.TransparentDecay(dr, 30)
		beforeDR := registeredEffectBytes4B7540(dr)
		beforeGraph := registeredRainGraph4B7310(t, c)
		logic, other := s.Rand.Logic.Index(), s.Rand.Other.Index()
		assertInputs := func(route string) {
			if !bytes.Equal(beforeDR, registeredEffectBytes4B7540(dr)) || !bytes.Equal(beforeType, registeredEffectBytes4B7540(typ)) || !bytes.Equal(beforeVP, registeredEffectBytes4B7540(vp)) || !reflect.DeepEqual(beforeGraph, registeredRainGraph4B7310(t, c)) {
				t.Fatalf("%s mutated native drawable/type/viewport/whole graph frame=%08x", route, row.Frame)
			}
			if c.Objs.Count != 1 || c.Objs.List1 != dr || c.Objs.DeadlineList != dr || s.Frame() != row.Frame || s.TickRate() != 60 || s.Rand.Logic.Index() != logic || s.Rand.Other.Index() != other || memmap.Uint32(0x85B3FC, 940) != uint32(def.color) {
				t.Fatalf("%s changed lifetime/frame/RNG/palette frame=%08x", route, row.Frame)
			}
			if d.IsAlphaEnabled() || d.Alpha() != def.alpha || d.Color2() != color {
				t.Fatalf("%s final alpha/color differs from original frame=%08x", route, row.Frame)
			}
		}
		resetRender()
		if got := c.callDrawableDraw4B6B80(dr, vp); got != 1 {
			t.Fatalf("native world return=%d", got)
		}
		assertInputs("native-world")
		worldData, worldPixels := registeredEffectBytes4B7540(d), slices.Clone(pix.Pix)
		t.Logf("WAYPOINT NATIVE CONTROL PASS: %s frame=%08x", name, row.Frame)
		resetRender()
		if got := dr.CallDraw(vp); got != 1 {
			t.Fatalf("registered C entry return=%d frame=%08x", got, row.Frame)
		}
		assertInputs("registered-C-entry")
		if !bytes.Equal(worldData, registeredEffectBytes4B7540(d)) || !slices.Equal(worldPixels, pix.Pix) {
			t.Fatalf("registered C entry differs from whole native render state/raster frame=%08x", row.Frame)
		}
		resetRender()
		d.SetColor2(color)
		d.SetAlphaEnabled(true)
		drawOriginal := func(line [2][2]int) {
			r.DrawLine(waypointOriginalPoint4B9790(line[0], center), waypointOriginalPoint4B9790(line[1], center), color)
		}
		for _, line := range oracle.Circle {
			drawOriginal(line)
		}
		for _, line := range row.Star {
			drawOriginal(line)
		}
		d.SetAlphaEnabled(false)
		assertInputs("original-endpoint-replay")
		if !bytes.Equal(worldData, registeredEffectBytes4B7540(d)) || !slices.Equal(worldPixels, pix.Pix) {
			t.Fatalf("whole render state/raster differs from original ordered endpoints frame=%08x", row.Frame)
		}
		presentParticleRegistration4B01B0(t, pix, output, sc, worldPixels, int(row.Frame))
		c.Nox_xxx_spriteDeleteStatic_45A4E0_drawable(dr)
		if c.Objs.Count != 0 || c.Objs.List1 != nil || c.Objs.DeadlineList != nil {
			t.Fatal("real drawable/deadline cleanup failed")
		}
		t.Logf("REGISTERED WAYPOINT FRAME PASS: %s frame=%08x", name, row.Frame)
	}
	if sc.PresentCount() != 132 {
		t.Fatal("missing real headless frame uploads")
	}
}
