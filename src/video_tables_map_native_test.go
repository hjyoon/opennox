//go:build !server

package opennox

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/server"
)

func videoTableOriginal200(t *testing.T) ([]int32, []string) {
	t.Helper()
	var row struct {
		SHA   string               `json:"original_sha256"`
		Trig  []int32              `json:"trig_words"`
		Calls [][2]json.RawMessage `json:"calls"`
	}
	data, err := os.ReadFile("legacy/testdata/video_tables_4a9c80_4aee30_original.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &row); err != nil {
		t.Fatal(err)
	}
	if row.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || len(row.Trig) != 320 || len(row.Calls) != 36 {
		t.Fatal("wrong original video initialization")
	}
	names := make([]string, 36)
	for i, call := range row.Calls {
		if err := json.Unmarshal(call[0], &names[i]); err != nil {
			t.Fatal(err)
		}
	}
	return row.Trig, names
}

func videoTableChild200(t *testing.T, mode string) bool {
	t.Helper()
	const key = "OPENNOX_TEST_VIDEO_TABLES_R200"
	if os.Getenv(key) == mode {
		legacy.InitBlobData()
		return true
	}
	cmd := exec.Command(os.Args[0], "-test.run=^"+t.Name()+"$", "-test.count=1", "-test.v")
	cmd.Env = append(os.Environ(), key+"="+mode)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("actual video initializer %s: %v\n%s", mode, err, output)
	}
	if !bytes.Contains(output, []byte("VIDEO TABLES ACTUAL PASS: "+mode)) {
		t.Fatalf("missing completed actual receipt: %s", output)
	}
	t.Log(strings.TrimSpace(string(output)))
	return false
}

func actualVideoRender200(t *testing.T) (*Client, *noximage.Image16, *noxrender.RenderData) {
	t.Helper()
	s := server.New(nil, nil, strman.New())
	t.Cleanup(s.Close)
	s.SetFrame(100)
	s.SetTickRate(60)
	s.Rand.Logic.Reset(93)
	s.Rand.Other.Reset(17)
	cl := client.NewClient(nil, nil, s)
	t.Cleanup(cl.Close)
	t.Cleanup(func() { alloc.FreePtr(cl.Viewport().C()) })
	r := cl.Render()
	t.Cleanup(r.Part.Free)
	d, free := noxrender.NewRenderData()
	t.Cleanup(free)
	r.SetData(d)
	pix := noximage.NewImage16(image.Rect(0, 0, 64, 64))
	r.SetPixBuffer(pix)
	c := &Client{Client: cl, srv: &Server{Server: s}, r: NewNoxRender(r)}
	oldClient, oldServer := noxClient, noxServer
	noxClient, noxServer = c, c.srv
	t.Cleanup(func() { noxClient, noxServer = oldClient, oldServer })
	return c, pix, d
}

func assertTrigTable200(t *testing.T, original []int32) []int32 {
	t.Helper()
	actual := unsafe.Slice(memmap.PtrInt32(0x5D4594, 1309840), 320)
	if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(unsafe.Pointer(&actual[0])) <= 0xffffffff {
		t.Fatal("native high-address table required")
	}
	for i, want := range original {
		if actual[i] != want {
			t.Fatalf("original FSIN/truncation word[%d]: actual=%d original=%d", i, actual[i], want)
		}
	}
	return actual
}

func TestVideoTrigActualInitializerAndParticlePoints4AEE30(t *testing.T) {
	if !videoTableChild200(t, "trig") {
		return
	}
	original, _ := videoTableOriginal200(t)
	c, _, _ := actualVideoRender200(t)
	before := slices.Clone(unsafe.Slice(memmap.PtrInt32(0x5D4594, 1309840-32), 320+16))
	sub_4AE520()
	t.Cleanup(c.r.circleSeg.Free)
	actual := assertTrigTable200(t, original)
	all := unsafe.Slice(memmap.PtrInt32(0x5D4594, 1309840-32), 320+16)
	if !slices.Equal(all[:8], before[:8]) || !slices.Equal(all[328:], before[328:]) {
		t.Fatal("actual table guard changed")
	}
	points := 0
	for _, index := range []int{-9, 0, 1, 63, 64, 127, 128, 191, 192, 255, 256, 300} {
		clamped := max(0, min(256, index))
		for _, radius := range []int{0, 1, 9, 127, -1, -31} {
			xword := before[328] // Existing clamp=256 reads the unchanged word after the 320 initialized entries.
			if clamped+64 < len(original) {
				xword = original[clamped+64]
			}
			want := image.Pt((radius*int(xword))>>16, (radius*int(original[clamped]))>>16)
			if got := sub_4AEDA0(index, radius); got != want {
				t.Fatalf("actual particle trig point index=%d radius=%d: got=%v original=%v", index, radius, got, want)
			}
			points++
		}
	}
	if !slices.Equal(actual, original) || c.Server.Frame() != 100 || c.Server.Rand.Logic.Index() != 93 || c.Server.Rand.Other.Index() != 17 {
		t.Fatal("consumer changed initialized table/frame/RNG")
	}
	t.Logf("VIDEO TABLES ACTUAL PASS: trig; real circle initializer/CGo FSIN table 320 words/guards and %d actual particle points match original", points)
}

func TestVideoCompassActualCGoHandlesHeadless4A9C80(t *testing.T) {
	if !videoTableChild200(t, "compass") {
		return
	}
	_, names := videoTableOriginal200(t)
	t.Chdir(t.TempDir())
	generatorBag4BC720(t)
	handles.Init()
	t.Cleanup(handles.Release)
	c, pix, d := actualVideoRender200(t)
	if err := c.r.GetBag().ReadVideoBag(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(c.r.GetBag().Free)
	seat := headless.New(pix.Size())
	t.Cleanup(func() { _ = seat.Close() })
	output, err := render.New(seat)
	if err != nil {
		t.Fatal(err)
	}
	table := unsafe.Slice(memmap.PtrUint32(0x5D4594, 1309516), 38)
	if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(unsafe.Pointer(&table[0])) <= 0xffffffff {
		t.Fatal("native high-address compass table required")
	}
	oldText := table[36]
	table[37] = 0xffffffff
	var images [36]*noxrender.Image
	calls := 0
	oldLoader := legacy.Nox_xxx_gLoadImg
	legacy.Nox_xxx_gLoadImg = func(name string) *noxrender.Image {
		if calls >= 36 || name != names[calls] {
			t.Fatalf("actual C formatter/load sequence call=%d name=%q", calls, name)
		}
		img := c.r.GetBag().ImageByIndex(calls % len(generatorColors4BC720))
		images[calls] = img
		calls++
		return img
	}
	t.Cleanup(func() { legacy.Nox_xxx_gLoadImg = oldLoader })
	if legacy.Nox_xxx_compassGenStrings_4A9C80() != 1 || calls != 36 || table[36] != oldText || table[37] != 0 {
		t.Fatal("actual compass calls/return/unrelated text/state differ")
	}
	for i, img := range images {
		slot := i - 4
		if i < 4 {
			slot = i + 32
		}
		if table[slot] != uint32(uintptr(img.C())) {
			t.Fatal("actual compass handle store/registry contract")
		}
	}
	aliasWord, freeAlias := alloc.New(uintptr(0))
	t.Cleanup(freeAlias)
	reset := func() {
		for i := range pix.Pix {
			pix.Pix[i] = 0x2104
		}
		*d = noxrender.RenderData{}
		d.SetClip(true)
		d.SetClipRect(pix.Rect)
		d.SetClipRect2(pix.Rect)
	}
	for frame := 0; frame < 32; frame++ {
		reset()
		c.r.DrawImage16(images[frame%4], image.Pt(20, 20))
		c.r.DrawImage16(images[frame+4], image.Pt(24, 20))
		want := slices.Clone(pix.Pix)
		reset()
		for _, slot := range []int{32 + frame%4, frame} {
			// A C-owned word models the original opaque DWORD-to-image cast.
			// Only the existing registry resolves it; no alias is dereferenced.
			*aliasWord = uintptr(table[slot])
			key := *(*noxrender.ImageHandle)(unsafe.Pointer(aliasWord))
			img := c.r.GetBag().AsImage(key)
			pos := image.Pt(20, 20)
			if slot < 32 {
				pos = image.Pt(24, 20)
			}
			if img == nil {
				t.Fatal("actual DWORD compass handle unresolved")
			}
			c.r.DrawImage16(img, pos)
		}
		if !slices.Equal(pix.Pix, want) {
			t.Fatalf("actual compass handles/frame %d raster", frame)
		}
		presentParticleRegistration4B01B0(t, pix, output, seat, want, 0)
	}
	if c.Server.Frame() != 100 || c.Server.Rand.Logic.Index() != 93 || c.Server.Rand.Other.Index() != 17 {
		t.Fatal("compass consumed frame/RNG")
	}
	t.Log("VIDEO TABLES ACTUAL PASS: compass; actual C formatter/CGo 36 image loads/DWORD handles, 32 paired renderer/headless frames; declared colored bag, not Windows compass raster")
}

func TestVideoTrigActualCircleSegmentHeadless4AEE30(t *testing.T) {
	if !videoTableChild200(t, "circle") {
		return
	}
	original, _ := videoTableOriginal200(t)
	c, pix, d := actualVideoRender200(t)
	sub_4AE520()
	t.Cleanup(c.r.circleSeg.Free)
	actual := assertTrigTable200(t, original)
	snapshot := slices.Clone(actual)
	seat := headless.New(pix.Size())
	t.Cleanup(func() { _ = seat.Close() })
	output, err := render.New(seat)
	if err != nil {
		t.Fatal(err)
	}
	reset := func(clipped bool) {
		for i := range pix.Pix {
			pix.Pix[i] = 0x2104
		}
		*d = noxrender.RenderData{}
		d.SetClip(true)
		d.SetClipRect(pix.Rect)
		d.SetClipRect2(pix.Rect)
		d.SetRect3(pix.Rect)
		if clipped {
			d.SetClipRect(image.Rect(24, 21, 47, 46))
			d.SetClipRect2(image.Rect(24, 21, 47, 46))
		}
	}
	frames := 0
	for angle := 0; angle < 256; angle++ {
		for _, radius := range []int{6, 19} {
			for _, clipped := range []bool{false, true} {
				reset(clipped)
				copy(actual, original)
				c.r.DrawCircleSegment(32, 32, radius, angle, noxcolor.RGBA5551(0x03e0))
				want := slices.Clone(pix.Pix)
				copy(actual, snapshot)
				reset(clipped)
				c.r.DrawCircleSegment(32, 32, radius, angle, noxcolor.RGBA5551(0x03e0))
				if !slices.Equal(pix.Pix, want) {
					t.Fatalf("actual initialized circle raster angle=%d radius=%d clip=%t", angle, radius, clipped)
				}
				presentParticleRegistration4B01B0(t, pix, output, seat, want, 0)
				frames++
			}
		}
	}
	if !slices.Equal(actual, snapshot) || c.Server.Frame() != 100 || c.Server.Rand.Logic.Index() != 93 || c.Server.Rand.Other.Index() != 17 {
		t.Fatal("circle changed actual table/frame/RNG")
	}
	t.Log(fmt.Sprintf("VIDEO TABLES ACTUAL PASS: circle; %d actual native circle/headless frames compared with original table using existing raster service", frames))
}
