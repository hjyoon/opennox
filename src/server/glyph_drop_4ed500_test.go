package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type glyphDropTestPoint4ED500 struct {
	x float32
	y float32
}

type glyphDropTestData4ED500 struct {
	point glyphDropTestPoint4ED500
}

type glyphDropTestObject4ED500 struct {
	name       string
	position   glyphDropTestPoint4ED500
	data       *glyphDropTestData4ED500
	direction1 uint16
	direction2 uint16
}

type glyphDropTestWorld4ED500 struct {
	owner           *glyphDropTestObject4ED500
	glyph           *glyphDropTestObject4ED500
	point           *glyphDropTestPoint4ED500
	trapResult      int32
	directionResult int32
	events          []string
	faultAt         int
	pointXLoads     int
	pointYLoads     int
	afterStoreY     func(*glyphDropTestWorld4ED500)
}

func (w *glyphDropTestWorld4ED500) event(name string) {
	w.events = append(w.events, name)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(name)
	}
}

func newGlyphDropTestWorld4ED500() *glyphDropTestWorld4ED500 {
	data := &glyphDropTestData4ED500{}
	return &glyphDropTestWorld4ED500{
		owner: &glyphDropTestObject4ED500{
			name:     "owner",
			position: glyphDropTestPoint4ED500{x: 21.5, y: 5.75},
		},
		glyph:           &glyphDropTestObject4ED500{name: "glyph", data: data},
		point:           &glyphDropTestPoint4ED500{x: 13.5, y: -2.25},
		trapResult:      1,
		directionResult: 0x1234abcd,
	}
}

func (w *glyphDropTestWorld4ED500) hooks() glyphDropHooks4ED500[
	*glyphDropTestObject4ED500,
	*glyphDropTestData4ED500,
	*glyphDropTestPoint4ED500,
] {
	return glyphDropHooks4ED500[
		*glyphDropTestObject4ED500,
		*glyphDropTestData4ED500,
		*glyphDropTestPoint4ED500,
	]{
		dropTrap: func(owner, glyph *glyphDropTestObject4ED500, point *glyphDropTestPoint4ED500) int32 {
			w.event("trap")
			if owner != w.owner || glyph != w.glyph || point != w.point {
				panic("trap arguments")
			}
			return w.trapResult
		},
		loadInitData: func(glyph *glyphDropTestObject4ED500) *glyphDropTestData4ED500 {
			w.event("init-data")
			return glyph.data
		},
		loadPointX: func(point *glyphDropTestPoint4ED500) float32 {
			w.pointXLoads++
			w.event(fmt.Sprintf("point-x-%d", w.pointXLoads))
			return point.x
		},
		storeGlyphX: func(data *glyphDropTestData4ED500, value float32) {
			w.event("store-glyph-x")
			data.point.x = value
		},
		loadPointY: func(point *glyphDropTestPoint4ED500) float32 {
			w.pointYLoads++
			w.event(fmt.Sprintf("point-y-%d", w.pointYLoads))
			return point.y
		},
		storeGlyphY: func(data *glyphDropTestData4ED500, value float32) {
			w.event("store-glyph-y")
			data.point.y = value
			if w.afterStoreY != nil {
				w.afterStoreY(w)
			}
		},
		loadObjectX: func(owner *glyphDropTestObject4ED500) float32 {
			w.event("owner-x")
			return owner.position.x
		},
		loadObjectY: func(owner *glyphDropTestObject4ED500) float32 {
			w.event("owner-y")
			return owner.position.y
		},
		vectorDirection: func(x, y float32) int32 {
			w.event("direction")
			if x != 8 || y != 8 {
				panic(fmt.Sprintf("direction vector %v/%v", x, y))
			}
			return w.directionResult
		},
		storeDirection2: func(glyph *glyphDropTestObject4ED500, value uint16) {
			w.event("store-direction-2")
			glyph.direction2 = value
		},
		storeDirection1: func(glyph *glyphDropTestObject4ED500, value uint16) {
			w.event("store-direction-1")
			glyph.direction1 = value
		},
		audio: func(id uint32, glyph *glyphDropTestObject4ED500, kind int32, code uint32) {
			w.event("audio")
			if id != 825 || glyph != w.glyph || kind != 0 || code != 0 {
				panic("audio arguments")
			}
		},
	}
}

func glyphDropSuccessEvents4ED500() []string {
	return []string{
		"trap", "init-data", "point-x-1", "store-glyph-x", "point-y-1", "store-glyph-y",
		"owner-x", "point-x-2", "owner-y", "point-y-2", "direction",
		"store-direction-2", "store-direction-1", "audio",
	}
}

func verifyGlyphDropFaultPrefixes4ED500(t *testing.T, want []string) {
	t.Helper()
	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			w := newGlyphDropTestWorld4ED500()
			w.faultAt = faultAt
			defer func() {
				if got := recover(); got != want[faultAt-1] {
					t.Fatalf("panic = %v, want %q", got, want[faultAt-1])
				}
				if !reflect.DeepEqual(w.events, want[:faultAt]) {
					t.Fatalf("events = %v, want %v", w.events, want[:faultAt])
				}
			}()
			glyphDrop4ED500(w.owner, w.glyph, w.point, w.hooks())
		})
	}
}

func TestGlyphDrop4ED500TrapGateUsesFullEAX(t *testing.T) {
	for _, result := range []int32{0, 1, -1, math.MinInt32} {
		t.Run(fmt.Sprintf("%08x", uint32(result)), func(t *testing.T) {
			w := newGlyphDropTestWorld4ED500()
			w.trapResult = result
			got := glyphDrop4ED500(w.owner, w.glyph, w.point, w.hooks())
			if result == 0 {
				if got != 0 || !reflect.DeepEqual(w.events, []string{"trap"}) {
					t.Fatalf("result/events = %d/%v", got, w.events)
				}
				return
			}
			if got != 1 || !reflect.DeepEqual(w.events, glyphDropSuccessEvents4ED500()) {
				t.Fatalf("result/events = %d/%v", got, w.events)
			}
		})
	}
}

func TestGlyphDrop4ED500ExactSuccessTraceAndLowDirectionWord(t *testing.T) {
	w := newGlyphDropTestWorld4ED500()
	if got := glyphDrop4ED500(w.owner, w.glyph, w.point, w.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if !reflect.DeepEqual(w.events, glyphDropSuccessEvents4ED500()) {
		t.Fatalf("events = %v, want %v", w.events, glyphDropSuccessEvents4ED500())
	}
	if w.glyph.data.point != *w.point {
		t.Fatalf("glyph point = %#v, want %#v", w.glyph.data.point, *w.point)
	}
	if w.glyph.direction2 != 0xabcd || w.glyph.direction1 != 0xabcd {
		t.Fatalf("directions = %#04x/%#04x, want abcd/abcd", w.glyph.direction2, w.glyph.direction1)
	}
	verifyGlyphDropFaultPrefixes4ED500(t, glyphDropSuccessEvents4ED500())
}

func TestGlyphDrop4ED500ReloadsAliasedPointAfterStores(t *testing.T) {
	w := newGlyphDropTestWorld4ED500()
	w.afterStoreY = func(w *glyphDropTestWorld4ED500) {
		// Models a3 overlapping a different part of the glyph-data destination.
		w.point.x = 101.5
		w.owner.position.y = 13.75
	}
	hooks := w.hooks()
	hooks.vectorDirection = func(x, y float32) int32 {
		w.event("direction")
		if x != -80 || y != 16 {
			t.Fatalf("live vector = %v/%v, want -80/16", x, y)
		}
		return 0x10002
	}
	if got := glyphDrop4ED500(w.owner, w.glyph, w.point, hooks); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if w.glyph.data.point.x != 13.5 || w.glyph.data.point.y != -2.25 {
		t.Fatalf("stored pre-alias point = %#v", w.glyph.data.point)
	}
	if w.glyph.direction1 != 2 || w.glyph.direction2 != 2 {
		t.Fatalf("directions = %d/%d, want 2/2", w.glyph.direction1, w.glyph.direction2)
	}
}

func TestGlyphDropSubtract4ED500RoundsAtBinary32Spill(t *testing.T) {
	left := math.Float32frombits(0x3f800001)
	right := math.Float32frombits(0x33800000)
	want := float32(float64(left) - float64(right))
	if got := glyphDropSubtract4ED500(left, right); math.Float32bits(got) != math.Float32bits(want) {
		t.Fatalf("subtract bits = %08x, want %08x", math.Float32bits(got), math.Float32bits(want))
	}
}
