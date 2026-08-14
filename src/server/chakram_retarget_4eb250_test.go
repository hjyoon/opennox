package server

import (
	"math"
	"testing"
)

type chakramRetargetTestObject4EB250 struct {
	name      string
	class     uint32
	flags     uint32
	invisible bool
	mapOK     bool
	x         float32
	y         float32
	velX      float32
	velY      float32
	speed     float32
	owner     *chakramRetargetTestObject4EB250
}

type chakramRetargetTestUpdate4EB250 struct {
	last  *chakramRetargetTestObject4EB250
	state uint8
}

type chakramRetargetTestFixture4EB250 struct {
	events     []string
	update     *chakramRetargetTestUpdate4EB250
	candidates []*chakramRetargetTestObject4EB250
	rect       chakramRetargetRect4EB250
}

func (f *chakramRetargetTestFixture4EB250) hooks() chakramRetargetHooks4EB250[
	*chakramRetargetTestObject4EB250,
	*chakramRetargetTestUpdate4EB250,
] {
	return chakramRetargetHooks4EB250[
		*chakramRetargetTestObject4EB250,
		*chakramRetargetTestUpdate4EB250,
	]{
		loadUpdateData: func(obj *chakramRetargetTestObject4EB250) *chakramRetargetTestUpdate4EB250 {
			f.events = append(f.events, "update:"+obj.name)
			return f.update
		},
		loadLastHit: func(update *chakramRetargetTestUpdate4EB250) *chakramRetargetTestObject4EB250 {
			f.events = append(f.events, "last")
			return update.last
		},
		loadOwner: func(obj *chakramRetargetTestObject4EB250) *chakramRetargetTestObject4EB250 {
			f.events = append(f.events, "owner:"+obj.name)
			return obj.owner
		},
		loadClass: func(obj *chakramRetargetTestObject4EB250) uint32 {
			f.events = append(f.events, "class:"+obj.name)
			return obj.class
		},
		loadFlags: func(obj *chakramRetargetTestObject4EB250) uint32 {
			f.events = append(f.events, "flags:"+obj.name)
			return obj.flags
		},
		hasEnchant: func(obj *chakramRetargetTestObject4EB250, enchant uint32) bool {
			f.events = append(f.events, "enchant:"+obj.name)
			if enchant != chakramRetargetInvisible4EB250 {
				panic("unexpected enchant")
			}
			return obj.invisible
		},
		mapCheck: func(candidate, source *chakramRetargetTestObject4EB250) bool {
			f.events = append(f.events, "map:"+candidate.name+":"+source.name)
			return candidate.mapOK
		},
		loadPosX: func(obj *chakramRetargetTestObject4EB250) float32 {
			f.events = append(f.events, "x:"+obj.name)
			return obj.x
		},
		loadPosY: func(obj *chakramRetargetTestObject4EB250) float32 {
			f.events = append(f.events, "y:"+obj.name)
			return obj.y
		},
		loadSpeed: func(obj *chakramRetargetTestObject4EB250) float32 {
			f.events = append(f.events, "speed:"+obj.name)
			return obj.speed
		},
		eachInRect: func(rect chakramRetargetRect4EB250, callback func(*chakramRetargetTestObject4EB250)) {
			f.events = append(f.events, "each")
			f.rect = rect
			for _, candidate := range f.candidates {
				callback(candidate)
			}
		},
		storeState: func(update *chakramRetargetTestUpdate4EB250, state uint8) {
			f.events = append(f.events, "state")
			update.state = state
		},
		storeVelocityX: func(obj *chakramRetargetTestObject4EB250, value float32) {
			f.events = append(f.events, "vel-x")
			obj.velX = value
		},
		storeVelocityY: func(obj *chakramRetargetTestObject4EB250, value float32) {
			f.events = append(f.events, "vel-y")
			obj.velY = value
		},
	}
}

func TestChakramRetarget4EB250FiltersAndKeepsFirstEqualDistance(t *testing.T) {
	owner := &chakramRetargetTestObject4EB250{name: "owner"}
	last := &chakramRetargetTestObject4EB250{name: "last"}
	source := &chakramRetargetTestObject4EB250{name: "source", x: 10, y: 20, speed: 11, owner: owner}
	wrongClass := &chakramRetargetTestObject4EB250{name: "wrong-class", mapOK: true}
	flagged := &chakramRetargetTestObject4EB250{name: "flagged", class: 2, flags: 0x20, mapOK: true}
	invisible := &chakramRetargetTestObject4EB250{name: "invisible", class: 2, invisible: true, mapOK: true}
	owner.class, owner.mapOK = 2, true
	last.class, last.mapOK = 2, true
	blocked := &chakramRetargetTestObject4EB250{name: "blocked", class: 2}
	edge := &chakramRetargetTestObject4EB250{name: "edge", class: 2, mapOK: true, x: 410, y: 20}
	first := &chakramRetargetTestObject4EB250{name: "first", class: 2, mapOK: true, x: 13, y: 24}
	equal := &chakramRetargetTestObject4EB250{name: "equal", class: 2, mapOK: true, x: 7, y: 16}
	f := &chakramRetargetTestFixture4EB250{
		update: &chakramRetargetTestUpdate4EB250{last: last},
		candidates: []*chakramRetargetTestObject4EB250{
			wrongClass, flagged, invisible, owner, last, blocked, edge, first, equal,
		},
	}
	got := chakramRetarget4EB250(source, f.hooks())
	if got != first {
		t.Fatalf("target = %v, want first equal-distance candidate", got.name)
	}
	if f.rect != (chakramRetargetRect4EB250{MinX: -390, MinY: -380, MaxX: 410, MaxY: 420}) {
		t.Fatalf("rect = %+v", f.rect)
	}
	if f.update.state != chakramReturnStateSeek4EAF00 {
		t.Fatalf("state = %d, want %d", f.update.state, chakramReturnStateSeek4EAF00)
	}
	denominator := float32(math.Sqrt(25) + float64(float32(0.1)))
	wantX := float32(float64(3) * 11 / float64(denominator))
	wantY := float32(float64(4) * 11 / float64(denominator))
	if source.velX != wantX || source.velY != wantY {
		t.Fatalf("velocity = (%v, %v), want (%v, %v)", source.velX, source.velY, wantX, wantY)
	}
	for _, forbidden := range []string{"flags:wrong-class", "enchant:flagged", "map:invisible:source", "map:owner:source", "map:last:source"} {
		for _, event := range f.events {
			if event == forbidden {
				t.Fatalf("filter order reached forbidden event %q in %q", forbidden, f.events)
			}
		}
	}
}

func TestChakramRetarget4EB250UnorderedDistanceMatchesX87Replacement(t *testing.T) {
	source := &chakramRetargetTestObject4EB250{name: "source", speed: 10}
	nanCandidate := &chakramRetargetTestObject4EB250{
		name: "nan", class: 2, mapOK: true, x: math.Float32frombits(0x7fc01234),
	}
	normal := &chakramRetargetTestObject4EB250{name: "normal", class: 2, mapOK: true, x: 3, y: 4}
	f := &chakramRetargetTestFixture4EB250{
		update:     &chakramRetargetTestUpdate4EB250{},
		candidates: []*chakramRetargetTestObject4EB250{nanCandidate, normal},
	}
	if got := chakramRetarget4EB250(source, f.hooks()); got != normal {
		t.Fatalf("target = %v, want later normal candidate after NaN best", got.name)
	}
	if math.IsNaN(float64(source.velX)) || math.IsNaN(float64(source.velY)) {
		t.Fatalf("normal replacement produced NaN velocity (%v, %v)", source.velX, source.velY)
	}
}

func TestChakramRetarget4EB250AcceptsExact400UnitBoundary(t *testing.T) {
	source := &chakramRetargetTestObject4EB250{name: "source", speed: 10}
	edge := &chakramRetargetTestObject4EB250{name: "edge", class: 2, mapOK: true, x: 400}
	f := &chakramRetargetTestFixture4EB250{
		update:     &chakramRetargetTestUpdate4EB250{},
		candidates: []*chakramRetargetTestObject4EB250{edge},
	}
	if got := chakramRetarget4EB250(source, f.hooks()); got != edge {
		t.Fatalf("target = %v, want exact-boundary candidate", got)
	}
}

func TestChakramRetarget4EB250NoTargetLeavesStateAndVelocityUntouched(t *testing.T) {
	source := &chakramRetargetTestObject4EB250{name: "source", x: 9, y: 8, velX: 6, velY: 5, speed: 7}
	f := &chakramRetargetTestFixture4EB250{
		update: &chakramRetargetTestUpdate4EB250{state: 1},
		candidates: []*chakramRetargetTestObject4EB250{
			{name: "wrong", class: 0, mapOK: true},
		},
	}
	if got := chakramRetarget4EB250(source, f.hooks()); got != nil {
		t.Fatalf("target = %v, want nil", got.name)
	}
	if source.velX != 6 || source.velY != 5 || f.update.state != 1 {
		t.Fatalf("mutated no-target state: velocity (%v, %v), state %d", source.velX, source.velY, f.update.state)
	}
	for _, forbidden := range []string{"state", "vel-x", "vel-y", "speed:source"} {
		for _, event := range f.events {
			if event == forbidden {
				t.Fatalf("unexpected event %q in %q", forbidden, f.events)
			}
		}
	}
}
