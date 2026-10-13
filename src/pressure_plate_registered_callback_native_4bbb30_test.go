//go:build !server

package opennox

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image"
	"math"
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
	"github.com/opennox/opennox/v1/internal/binfile"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/server"
)

type pressurePlateOriginalRow4BBB30 struct {
	Name      string        `json:"name"`
	ShapeBits [8]uint32     `json:"shape_bits"`
	RGB       [6]byte       `json:"rgb"`
	Visible   bool          `json:"visible"`
	Colors    [][3]byte     `json:"colors"`
	Lines     [][2][2]int32 `json:"lines"`
	Returned  int           `json:"returned"`
	Unchanged bool          `json:"guarded_bytes_unchanged"`
}

type pressurePlateOriginalOracle4BBB30 struct {
	PE      string                           `json:"original_sha256"`
	Control uint16                           `json:"x87_control_word"`
	Entries map[string]int                   `json:"unhooked_entries"`
	Faults  []json.RawMessage                `json:"memory_faults"`
	Rows    []pressurePlateOriginalRow4BBB30 `json:"rows"`
}

func loadPressurePlateOriginal4BBB30(t *testing.T) pressurePlateOriginalOracle4BBB30 {
	t.Helper()
	b, err := os.ReadFile("testdata/pressure_plate_4bbb30.json")
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(b)) != "6719ed3780dd3916136bc13e0d4f77f19a12f07778b6e29b106902efe0a4b05a" {
		t.Fatal("frozen unhooked original transcript changed")
	}
	var out pressurePlateOriginalOracle4BBB30
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	if out.PE != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || out.Control != 0x037f || len(out.Rows) != 72 || out.Entries["0x4bbb30"] != 72 || out.Entries["0x419a70"] != 504 || len(out.Faults) != 0 {
		t.Fatal("wrong original pressure-plate execution")
	}
	for _, row := range out.Rows {
		want := 0
		if row.Visible {
			want = 8
		}
		if len(row.Lines) != want || row.Returned != 1 || !row.Unchanged {
			t.Fatal("incomplete original geometry/guard receipt")
		}
	}
	return out
}

func applyPressurePlateOriginalRow4BBB30(dr *client.Drawable, row pressurePlateOriginalRow4BBB30) {
	box := &dr.Shape.Box
	fields := []*float32{&box.LeftTop, &box.LeftBottom, &box.LeftBottom2, &box.LeftTop2, &box.RightTop, &box.RightBottom, &box.RightBottom2, &box.RightTop2}
	for i, ptr := range fields {
		*ptr = math.Float32frombits(row.ShapeBits[i])
	}
	rgb := row.RGB
	dr.UnionEffect().Field_108 = uint32(rgb[0]) | uint32(rgb[1])<<8 | uint32(rgb[2])<<16 | uint32(rgb[3])<<24
	// PE32 visibility reads exactly six bytes. These next two bytes must
	// neither activate a zero-RGB plate nor be modified by the draw.
	dr.UnionEffect().Field_109 = uint32(rgb[4]) | uint32(rgb[5])<<8 | 0xbeef0000
	dr.UnionEffect().Field_110 = 0x12345678
}

// Original ADD/LEA instructions produce signed dword endpoints. Translate the
// frozen original base-zero transcript with explicit dword wrap, independently
// of the production shape conversion and native image.Point arithmetic.
func pressurePlateOriginalPoint4BBB30(p [2]int32, base image.Point) image.Point {
	return image.Pt(int(int32(uint32(p[0])+uint32(base.X))), int(int32(uint32(p[1])+uint32(base.Y))))
}

func assertPressurePlateOriginalGeometry4BBB30(t *testing.T, dr *client.Drawable, vp *noxrender.Viewport, row pressurePlateOriginalRow4BBB30, base image.Point) {
	t.Helper()
	got := pressurePlateDrawStateFor4BBB30(dr, vp)
	if got.visible != row.Visible {
		t.Fatalf("original visibility differs case=%s", row.Name)
	}
	wantColors := [2]noxcolor.RGBA5551{
		noxcolor.RGB5551Color(row.RGB[0], row.RGB[1], row.RGB[2]),
		noxcolor.RGB5551Color(row.RGB[3], row.RGB[4], row.RGB[5]),
	}
	if got.colors != wantColors {
		t.Fatalf("original six low-byte RGB colors differ case=%s", row.Name)
	}
	for i, line := range row.Lines {
		want := indicatorLine4B9790{from: pressurePlateOriginalPoint4BBB30(line[0], base), to: pressurePlateOriginalPoint4BBB30(line[1], base)}
		if got.lines[i/4][i%4] != want {
			t.Errorf("original signed-dword x87 geometry case=%s base=%v line=%d: got=%v want=%v", row.Name, base, i, got.lines[i/4][i%4], want)
		}
	}
}

func TestPressurePlateOriginal4BBB30Geometry(t *testing.T) {
	oracle := loadPressurePlateOriginal4BBB30(t)
	for _, row := range oracle.Rows {
		for _, base := range []image.Point{{0, 0}, {256, 256}, {-256, -256}} {
			t.Run(fmt.Sprintf("%s/base-%d-%d", strings.ReplaceAll(row.Name, "/", "-"), base.X, base.Y), func(t *testing.T) {
				dr, freeDR := alloc.New(client.Drawable{})
				t.Cleanup(freeDR)
				vp, freeVP := alloc.New(noxrender.Viewport{})
				t.Cleanup(freeVP)
				vp.World, vp.Screen = image.Rect(100, 200, 164, 264), image.Rect(7, 11, 71, 75)
				dr.PosVec = vp.World.Min.Add(base).Sub(vp.Screen.Min)
				applyPressurePlateOriginalRow4BBB30(dr, row)
				if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(dr.C()) <= 0xffffffff || uintptr(vp.C()) <= 0xffffffff) {
					t.Fatal("geometry requires real high-address C-owned records")
				}
				beforeDR, beforeVP := registeredEffectBytes4B7540(dr), registeredEffectBytes4B7540(vp)
				assertPressurePlateOriginalGeometry4BBB30(t, dr, vp, row, base)
				if !bytes.Equal(beforeDR, registeredEffectBytes4B7540(dr)) || !bytes.Equal(beforeVP, registeredEffectBytes4B7540(vp)) {
					t.Fatal("geometry mutated native input bytes")
				}
			})
		}
	}
}

const registeredPressurePlateChild4BBB30 = "OPENNOX_TEST_REGISTERED_PRESSURE_PLATE_R170"

type registeredPressurePlateCase4BBB30 struct {
	alpha   byte
	enabled bool
	style   string
}

func (c registeredPressurePlateCase4BBB30) name() string {
	return fmt.Sprintf("alpha-%d/enabled-%t/%s", c.alpha, c.enabled, c.style)
}

var registeredPressurePlateCases4BBB30 = func() []registeredPressurePlateCase4BBB30 {
	var out []registeredPressurePlateCase4BBB30
	for _, alpha := range []byte{0, 64, 128, 255} {
		for _, enabled := range []bool{false, true} {
			for _, style := range []string{"center", "clip-corner", "edge-right", "outside", "screen-offset"} {
				out = append(out, registeredPressurePlateCase4BBB30{alpha, enabled, style})
			}
		}
	}
	return out
}()

// All 72 original shapes/colors run through the real C DRAW registration in
// each child. The complete native renderer/raster/upload comparison is against
// declared line services, not an independent original Windows framebuffer.
func TestRegisteredPressurePlateCallbackNative4BBB30(t *testing.T) {
	if name := os.Getenv(registeredPressurePlateChild4BBB30); name != "" {
		testRegisteredPressurePlateChild4BBB30(t, name)
		return
	}
	for _, def := range registeredPressurePlateCases4BBB30 {
		name := def.name()
		t.Run(strings.ReplaceAll(name, "/", "-"), func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run=^TestRegisteredPressurePlateCallbackNative4BBB30$", "-test.v")
			cmd.Env = append(os.Environ(), registeredPressurePlateChild4BBB30+"="+name)
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("registered native pressure plate failed: %v\n%s", err, out)
			}
			if bytes.Count(out, []byte("REGISTERED PRESSURE PLATE FRAME PASS:")) != 72 {
				t.Fatalf("missing original/native/C-entry/raster/headless receipts: %s", out)
			}
			t.Logf("REGISTERED PRESSURE PLATE CALLBACK PASS: %s frames=72", name)
		})
	}
}

func registeredPressurePlateType4BBB30(t *testing.T, c *Client) *client.ObjectType {
	t.Helper()
	var raw []byte
	for _, value := range []string{"PressurePlate", "DRAW", "PressurePlateDraw"} {
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
	typ := c.Things.TypeByID("PressurePlate")
	if typ == nil || typ.DrawFunc != legacy.Get_nox_thing_pressure_plate_draw() || !client.IsRegisteredDrawFunc(typ.DrawFunc) || typ.DrawData != nil || client.DrawableDataKind(typ.DrawFunc) != 0 {
		t.Fatal("actual registered kind0 pressure-plate type failed")
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(typ.C()) <= 0xffffffff {
		t.Fatal("actual C-owned type must be above 4 GiB")
	}
	return typ
}

func testRegisteredPressurePlateChild4BBB30(t *testing.T, name string) {
	var def registeredPressurePlateCase4BBB30
	found := false
	for _, candidate := range registeredPressurePlateCases4BBB30 {
		if candidate.name() == name {
			def, found = candidate, true
			break
		}
	}
	if !found {
		t.Fatal("unknown frozen pressure-plate case")
	}
	oracle := loadPressurePlateOriginal4BBB30(t)
	t.Chdir(t.TempDir())
	handles.Init()
	t.Cleanup(handles.Release)
	s := server.New(nil, nil, strman.New())
	t.Cleanup(s.Close)
	s.SetTickRate(60)
	s.SetFrame(100)
	if def.style == "outside" {
		s.SetFrame(0xffffffff)
	}
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
	typ := registeredPressurePlateType4BBB30(t, c)
	beforeType := registeredEffectBytes4B7540(typ)
	vp, freeVP := alloc.New(noxrender.Viewport{})
	t.Cleanup(freeVP)
	vp.World, vp.Screen = image.Rect(100, 200, 164, 264), image.Rect(7, 11, 71, 75)
	vp.Size = pix.Size()
	center, clip := image.Pt(25, 21), image.Rect(0, 0, 63, 63)
	switch def.style {
	case "clip-corner":
		center, clip = image.Pt(-3, -7), image.Rect(3, 3, 32, 32)
	case "edge-right":
		center = image.Pt(55, 21)
	case "outside":
		center = image.Pt(89, 85)
	case "screen-offset":
		vp.Screen = image.Rect(17, 29, 81, 93)
	}
	base := center.Add(vp.Screen.Min)
	beforeVP := registeredEffectBytes4B7540(vp)
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
		d.SetAlphaEnabled(def.enabled)
		d.SetClip(true)
		d.SetClipRect(pix.Rect)
		d.SetClipRect2(clip)
		d.SetColor2(noxcolor.RGBA5551(0x001f))
	}
	frame, logic, other := s.Frame(), s.Rand.Logic.Index(), s.Rand.Other.Index()
	for tick, row := range oracle.Rows {
		dr := c.Nox_xxx_spriteLoadAdd_45A360_drawable(c.Things.IndByID("PressurePlate"), vp.World.Min.Add(center))
		if dr == nil || dr.DrawFuncPtr != typ.DrawFunc {
			t.Fatal("real type/pool/spawn lost pressure-plate callback")
		}
		if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(dr.C()) <= 0xffffffff || uintptr(vp.C()) <= 0xffffffff) {
			t.Fatal("actual native drawable/viewport must be above 4 GiB")
		}
		applyPressurePlateOriginalRow4BBB30(dr, row)
		dr.ZVal, dr.VelZ = 17, -3
		c.Objs.TransparentDecay(dr, 30)
		beforeDR, beforeGraph := registeredEffectBytes4B7540(dr), registeredRainGraph4B7310(t, c)
		assertInputs := func(route string) {
			if !bytes.Equal(beforeDR, registeredEffectBytes4B7540(dr)) || !bytes.Equal(beforeType, registeredEffectBytes4B7540(typ)) || !bytes.Equal(beforeVP, registeredEffectBytes4B7540(vp)) || !reflect.DeepEqual(beforeGraph, registeredRainGraph4B7310(t, c)) {
				t.Fatalf("%s mutated native drawable/type/viewport/whole graph case=%s", route, row.Name)
			}
			if c.Objs.Count != 1 || c.Objs.List1 != dr || c.Objs.DeadlineList != dr || s.Frame() != frame || s.TickRate() != 60 || s.Rand.Logic.Index() != logic || s.Rand.Other.Index() != other {
				t.Fatalf("%s changed lifetime/frame/RNG case=%s", route, row.Name)
			}
			wantColor := noxcolor.RGBA5551(0x001f)
			if row.Visible {
				rgb := row.Colors[1]
				wantColor = noxcolor.RGB5551Color(rgb[0], rgb[1], rgb[2])
			}
			if d.IsAlphaEnabled() != def.enabled || d.Alpha() != def.alpha || d.Color2() != wantColor {
				t.Fatalf("%s alpha/final RGB differs case=%s", route, row.Name)
			}
		}
		resetRender()
		if got := c.callDrawableDraw4B6B80(dr, vp); got != 1 {
			t.Fatalf("native world return=%d", got)
		}
		assertInputs("native-world")
		worldData, worldPixels := registeredEffectBytes4B7540(d), slices.Clone(pix.Pix)
		t.Logf("PRESSURE PLATE NATIVE CONTROL PASS: %s case=%s", name, row.Name)
		resetRender()
		if got := dr.CallDraw(vp); got != 1 {
			t.Fatalf("registered C entry return=%d", got)
		}
		assertInputs("registered-C-entry")
		if !bytes.Equal(worldData, registeredEffectBytes4B7540(d)) || !slices.Equal(worldPixels, pix.Pix) {
			t.Fatalf("registered C entry differs from whole native render state/raster case=%s", row.Name)
		}
		assertPressurePlateOriginalGeometry4BBB30(t, dr, vp, row, base)
		resetRender()
		for i, line := range row.Lines {
			if i%4 == 0 {
				rgb := row.Colors[i/4]
				d.SetColor2(noxcolor.RGB5551Color(rgb[0], rgb[1], rgb[2]))
			}
			r.DrawLine(pressurePlateOriginalPoint4BBB30(line[0], base), pressurePlateOriginalPoint4BBB30(line[1], base), d.Color2())
		}
		assertInputs("original-endpoint-replay")
		if !bytes.Equal(worldData, registeredEffectBytes4B7540(d)) || !slices.Equal(worldPixels, pix.Pix) {
			t.Fatalf("whole render state/raster differs from original ordered RGB/endpoints case=%s", row.Name)
		}
		presentParticleRegistration4B01B0(t, pix, output, sc, worldPixels, tick)
		c.Nox_xxx_spriteDeleteStatic_45A4E0_drawable(dr)
		if c.Objs.Count != 0 || c.Objs.List1 != nil || c.Objs.DeadlineList != nil {
			t.Fatal("actual drawable/deadline cleanup failed")
		}
		t.Logf("REGISTERED PRESSURE PLATE FRAME PASS: %s case=%s", name, row.Name)
	}
	if sc.PresentCount() != 72 {
		t.Fatal("missing real headless uploads")
	}
}
