package legacy

import (
	"image"
	"image/color"
	"os"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/common/memmap"
)

type lineTrailTestClient struct {
	Client
	render *lineTrailTestRender
}

func (c *lineTrailTestClient) R2() Render2 { return c.render }

type lineTrailTestRender struct {
	Render2
	state  *noxrender.RenderData
	points []image.Point
	alpha  []byte
}

func (r *lineTrailTestRender) Data() *noxrender.RenderData { return r.state }
func (r *lineTrailTestRender) AddPoint(point image.Point)  { r.points = append(r.points, point) }

func (r *lineTrailTestRender) DrawLineFromPoints(color.Color, ...image.Point) bool {
	if r.state.IsAlphaEnabled() {
		r.alpha = append(r.alpha, r.state.Alpha())
	}
	return true
}

func TestDrawableLineTrailPreservesAlphaStepsAndCoordinates(t *testing.T) {
	loadOriginalRenderTable(t, 194136, 8)
	state, free := noxrender.NewRenderData()
	defer free()
	render := &lineTrailTestRender{state: state}
	oldClient := GetClient
	GetClient = func() Client { return &lineTrailTestClient{render: render} }
	defer func() { GetClient = oldClient }()

	dr := &client.Drawable{PosVec: image.Pt(140, 160), ZVal: 3, ZVal2: 65534}
	fx := &client.DrawableFX{
		Count: 5,
		Trail: [6][2]int32{{135, 158}, {132, 150}, {128, 145}, {122, 141}, {120, 139}, {115, 132}},
	}
	view := &noxrender.Viewport{
		Screen: image.Rect(10, 20, 650, 500),
		World:  image.Rect(100, 100, 740, 580),
	}
	if got := drawLineTrail(dr, fx, view); got != 1 {
		t.Fatalf("line trail result = %d, want 1", got)
	}
	if want := []byte{192, 129, 66, 3, 196}; !slices.Equal(render.alpha, want) || state.IsAlphaEnabled() {
		t.Fatalf("line alpha = %v, left enabled = %t, want %v and disabled", render.alpha, state.IsAlphaEnabled(), want)
	}
	want := []image.Point{
		{X: 38, Y: 69}, {X: 33, Y: 67},
		{X: 33, Y: 67}, {X: 30, Y: 59},
		{X: 30, Y: 59}, {X: 26, Y: 54},
		{X: 26, Y: 54}, {X: 20, Y: 50},
		{X: 20, Y: 50}, {X: 18, Y: 48},
	}
	if !slices.Equal(render.points, want) {
		t.Fatalf("line endpoints = %v, want %v", render.points, want)
	}
	if trail := ([6][2]int32{{140, 160}, {135, 158}, {132, 150}, {128, 145}, {122, 141}, {120, 139}}); fx.Trail != trail || fx.Count != 5 {
		t.Fatalf("line history = %v count %d, want %v count 5", fx.Trail, fx.Count, trail)
	}
}

func loadOriginalRenderTable(t *testing.T, offset uintptr, size int) {
	t.Helper()
	blob, err := os.ReadFile("../common/memmap/nox/blobdata/blob_587000.dat")
	if err != nil {
		t.Fatal(err)
	}
	end := int(offset) + size
	if len(blob) < end {
		t.Fatalf("shipped render table at %#x is truncated", offset)
	}
	data := unsafe.Slice((*byte)(memmap.PtrOff(0x587000, offset)), size)
	saved := slices.Clone(data)
	t.Cleanup(func() { copy(data, saved) })
	copy(data, blob[int(offset):end])
}
