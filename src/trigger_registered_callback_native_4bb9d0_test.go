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

type triggerOriginalRow4BB9D0 struct {
	Name      string         `json:"name"`
	ShapeBits [8]uint32      `json:"shape_bits"`
	Base      [2]int         `json:"base"`
	Lines     [4][2][2]int32 `json:"lines"`
	Returned  int            `json:"returned"`
	Unchanged bool           `json:"guarded_bytes_unchanged"`
}

type triggerOriginalOracle4BB9D0 struct {
	PE      string                     `json:"original_sha256"`
	Control uint16                     `json:"x87_control_word"`
	Entries map[string]int             `json:"unhooked_entries"`
	Faults  []json.RawMessage          `json:"memory_faults"`
	Rows    []triggerOriginalRow4BB9D0 `json:"rows"`
}

func loadTriggerOriginal4BB9D0(t *testing.T) triggerOriginalOracle4BB9D0 {
	t.Helper()
	b, err := os.ReadFile("testdata/trigger_draw_4bb9d0.json")
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(b)) != "1c0814c0dbf8b9515f5d19ab14639a319c07ba22398d011edbcb8f588570e3c4" {
		t.Fatal("frozen original trigger transcript changed")
	}
	var out triggerOriginalOracle4BB9D0
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	if out.PE != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || out.Control != 0x037f || len(out.Rows) != 27 || out.Entries["0x4bb9d0"] != 27 || out.Entries["0x419a70"] != 216 || len(out.Faults) != 0 {
		t.Fatal("wrong original trigger execution")
	}
	for _, row := range out.Rows {
		if row.Returned != 1 || !row.Unchanged {
			t.Fatal("incomplete original guard/return receipt")
		}
	}
	return out
}

func applyTriggerOriginalRow4BB9D0(dr *client.Drawable, row triggerOriginalRow4BB9D0) {
	box := &dr.Shape.Box
	fields := []*float32{&box.LeftTop, &box.LeftBottom, &box.LeftBottom2, &box.LeftTop2, &box.RightTop, &box.RightBottom, &box.RightBottom2, &box.RightTop2}
	for i, ptr := range fields {
		*ptr = math.Float32frombits(row.ShapeBits[i])
	}
	dr.UnionEffect().Field_108 = 0x89abcdef
	dr.UnionEffect().Field_109 = 0xbeef0123
	dr.UnionEffect().Field_110 = 0x12345678
}

// Original ADD/LEA instructions produce signed dword endpoints. Translate the
// frozen original base-zero transcript with explicit dword wrap, independently
// of the production shape conversion and native image.Point arithmetic.
func triggerOriginalPoint4BB9D0(p [2]int32, base image.Point) image.Point {
	return image.Pt(int(int32(uint32(p[0])+uint32(base.X))), int(int32(uint32(p[1])+uint32(base.Y))))
}

const registeredTriggerChild4BB9D0 = "OPENNOX_TEST_REGISTERED_TRIGGER_DRAW_R171"

type registeredTriggerCase4BB9D0 struct {
	alpha   byte
	enabled bool
	style   string
}

func (c registeredTriggerCase4BB9D0) name() string {
	return fmt.Sprintf("alpha-%d/enabled-%t/%s", c.alpha, c.enabled, c.style)
}

var registeredTriggerCases4BB9D0 = func() []registeredTriggerCase4BB9D0 {
	var out []registeredTriggerCase4BB9D0
	for _, alpha := range []byte{0, 64, 128, 255} {
		for _, enabled := range []bool{false, true} {
			for _, style := range []string{"center", "clip-corner", "edge-right", "outside", "screen-offset", "negative-origin", "high-word-origin", "word-coincidence-control"} {
				out = append(out, registeredTriggerCase4BB9D0{alpha, enabled, style})
			}
		}
	}
	return out
}()

// All 27 original shapes/origins run through the real C DRAW registration in
// each child. The complete native renderer/raster/upload comparison is against
// declared line services, not an independent original Windows framebuffer.
func TestRegisteredTriggerCallbackNative4BB9D0(t *testing.T) {
	if name := os.Getenv(registeredTriggerChild4BB9D0); name != "" {
		testRegisteredTriggerChild4BB9D0(t, name)
		return
	}
	for _, def := range registeredTriggerCases4BB9D0 {
		name := def.name()
		t.Run(strings.ReplaceAll(name, "/", "-"), func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run=^TestRegisteredTriggerCallbackNative4BB9D0$", "-test.v")
			cmd.Env = append(os.Environ(), registeredTriggerChild4BB9D0+"="+name)
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("registered native trigger failed: %v\n%s", err, out)
			}
			if bytes.Count(out, []byte("REGISTERED TRIGGER FRAME PASS:")) != 27 {
				t.Fatalf("missing original/native/C-entry/raster/headless receipts: %s", out)
			}
			t.Logf("REGISTERED TRIGGER CALLBACK PASS: %s frames=27", name)
		})
	}
}

func registeredTriggerType4BB9D0(t *testing.T, c *Client) *client.ObjectType {
	t.Helper()
	var raw []byte
	for _, value := range []string{"Trigger", "DRAW", "TriggerDraw"} {
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
	typ := c.Things.TypeByID("Trigger")
	if typ == nil || typ.DrawFunc != legacy.Get_nox_thing_trigger_draw() || !client.IsRegisteredDrawFunc(typ.DrawFunc) || typ.DrawData != nil || client.DrawableDataKind(typ.DrawFunc) != 0 {
		t.Fatal("actual registered kind0 trigger type failed")
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(typ.C()) <= 0xffffffff {
		t.Fatal("actual C-owned type must be above 4 GiB")
	}
	return typ
}

func testRegisteredTriggerChild4BB9D0(t *testing.T, name string) {
	var def registeredTriggerCase4BB9D0
	found := false
	for _, candidate := range registeredTriggerCases4BB9D0 {
		if candidate.name() == name {
			def, found = candidate, true
			break
		}
	}
	if !found {
		t.Fatal("unknown frozen trigger case")
	}
	oracle := loadTriggerOriginal4BB9D0(t)
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
	typ := registeredTriggerType4BB9D0(t, c)
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
	case "negative-origin":
		center = image.Pt(-263, -267)
	case "high-word-origin":
		if unsafe.Sizeof(uintptr(0)) == 8 {
			wide := int(uint64(1) << 32)
			vp.Screen = image.Rect(wide+7, wide+11, wide+71, wide+75)
			vp.World = image.Rect(wide+100, wide+200, wide+164, wide+264)
		}
	case "word-coincidence-control":
		center = image.Pt(32, 32)
		vp.Screen, vp.World = image.Rect(0, 0, 64, 64), image.Rect(64, 0, 128, 64)
	}
	base := center.Add(vp.Screen.Min)
	oldBlack := legacy.Get_nox_color_black_2650656()
	legacy.Set_nox_color_black_2650656(0)
	t.Cleanup(func() { legacy.Set_nox_color_black_2650656(oldBlack) })
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
		rowBase := image.Pt(row.Base[0], row.Base[1])
		dr := c.Nox_xxx_spriteLoadAdd_45A360_drawable(c.Things.IndByID("Trigger"), vp.World.Min.Add(center).Add(rowBase))
		if dr == nil || dr.DrawFuncPtr != typ.DrawFunc {
			t.Fatal("real type/pool/spawn lost trigger callback")
		}
		if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(dr.C()) <= 0xffffffff || uintptr(vp.C()) <= 0xffffffff) {
			t.Fatal("actual native drawable/viewport must be above 4 GiB")
		}
		applyTriggerOriginalRow4BB9D0(dr, row)
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
			if d.IsAlphaEnabled() || d.Alpha() != def.alpha || d.Color2() != noxcolor.RGBA5551(0) || legacy.Get_nox_color_black_2650656() != 0 || reflect.ValueOf(r).Elem().FieldByName("points").Len() != 0 {
				t.Fatalf("%s original final alpha/color/point-queue differs case=%s", route, row.Name)
			}
		}
		// Replay unchanged original endpoints through the real native point/
		// line frontend. No Windows framebuffer or whole-world equivalence
		// is claimed for these declared services.
		resetRender()
		r.ClearPoints()
		d.SetColor2(noxcolor.RGBA5551(0))
		d.SetAlphaEnabled(true)
		for _, line := range row.Lines {
			r.AddPoint(triggerOriginalPoint4BB9D0(line[0], base))
			r.AddPoint(triggerOriginalPoint4BB9D0(line[1], base))
			if !r.DrawLineFromPoints(d.Color2()) {
				t.Fatal("real original endpoint frontend did not consume two points")
			}
		}
		d.SetAlphaEnabled(false)
		assertInputs("original-endpoint-replay")
		wantData, wantPixels := registeredEffectBytes4B7540(d), slices.Clone(pix.Pix)
		t.Logf("TRIGGER ORIGINAL CONTROL PASS: %s case=%s", name, row.Name)
		for _, direct := range []bool{true, false} {
			resetRender()
			r.ClearPoints()
			got := 0
			if direct {
				got = dr.CallDraw(vp)
			} else {
				got = c.callDrawableDraw4B6B80(dr, vp)
			}
			if got != 1 {
				t.Fatalf("real TriggerDraw return=%d direct=%t", got, direct)
			}
			assertInputs(fmt.Sprintf("direct=%t", direct))
			if !bytes.Equal(wantData, registeredEffectBytes4B7540(d)) || !slices.Equal(wantPixels, pix.Pix) {
				t.Errorf("registered TriggerDraw differs from original complete state/raster: direct=%t style=%s case=%s", direct, def.style, row.Name)
			}
		}
		if !t.Failed() {
			presentParticleRegistration4B01B0(t, pix, output, sc, wantPixels, tick)
			t.Logf("REGISTERED TRIGGER FRAME PASS: %s case=%s", name, row.Name)
		}
		c.Nox_xxx_spriteDeleteStatic_45A4E0_drawable(dr)
		if c.Objs.Count != 0 || c.Objs.List1 != nil || c.Objs.DeadlineList != nil {
			t.Fatal("actual drawable/deadline cleanup failed")
		}
	}
	if !t.Failed() && sc.PresentCount() != 27 {
		t.Fatal("missing real headless uploads")
	}
}
