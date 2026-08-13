package server

import (
	"math"
	"reflect"
	"testing"
)

type colliderLimitProbe4E7410 struct {
	*colliderProbe4E7290
	flags uint8
}

func (p *colliderLimitProbe4E7410) colliderFlagsByte4E7410() uint8 {
	p.event("flags")
	return p.flags
}

func (p *colliderLimitProbe4E7410) colliderMinXBits4E7410() uint32 {
	p.event("read-min-x")
	return p.minX
}

func (p *colliderLimitProbe4E7410) colliderMinYBits4E7410() uint32 {
	p.event("read-min-y")
	return p.minY
}

func (p *colliderLimitProbe4E7410) colliderMaxXBits4E7410() uint32 {
	p.event("read-max-x")
	return p.maxX
}

func (p *colliderLimitProbe4E7410) colliderMaxYBits4E7410() uint32 {
	p.event("read-max-y")
	return p.maxY
}

func newColliderLimitProbe4E7410() *colliderLimitProbe4E7410 {
	return &colliderLimitProbe4E7410{colliderProbe4E7290: &colliderProbe4E7290{}}
}

func TestObjectColliderExtentBelowLimit4E7410X87C0Semantics(t *testing.T) {
	// This exact binary32 difference is below 85, but rounding the subtraction
	// to binary32 first produces 85. GAME.EXE keeps the difference in x87.
	maxNear := math.Float32frombits(0x42800000) // 64
	minNear := math.Float32frombits(0xc1a7ffff) // nextafter(-21, +Inf)
	if got := maxNear - minNear; got != 85 {
		t.Fatalf("binary32 control subtraction = %v, want 85", got)
	}

	tests := []struct {
		name         string
		maxBits      uint32
		minBits      uint32
		wantC0Branch bool
	}{
		{"extended precision below", math.Float32bits(maxNear), math.Float32bits(minNear), true},
		{"equal", math.Float32bits(85), math.Float32bits(0), false},
		{"greater", math.Float32bits(86), math.Float32bits(0), false},
		{"unordered max", 0x7fa12345, math.Float32bits(0), true},
		{"unordered subtraction", math.Float32bits(float32(math.Inf(1))), math.Float32bits(float32(math.Inf(1))), true},
		{"positive infinity", math.Float32bits(float32(math.Inf(1))), math.Float32bits(0), false},
		{"negative infinity", math.Float32bits(float32(math.Inf(-1))), math.Float32bits(0), true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := objectColliderExtentBelowLimit4E7410(tc.maxBits, tc.minBits); got != tc.wantC0Branch {
				t.Fatalf("C0 branch = %v, want %v", got, tc.wantC0Branch)
			}
		})
	}
}

func TestObjectColliderAllowed4E7410NoCollideShortCircuit(t *testing.T) {
	p := newColliderLimitProbe4E7410()
	p.flags = 0x40
	p.kind = 99
	p.minX, p.minY, p.maxX, p.maxY = 1, 2, 3, 4
	if !objectColliderAllowed4E7410(p) {
		t.Fatal("NoCollide object was rejected")
	}
	if got, want := [4]uint32{p.minX, p.minY, p.maxX, p.maxY}, [4]uint32{1, 2, 3, 4}; got != want {
		t.Fatalf("bounds = %#v, want unchanged %#v", got, want)
	}
	if !reflect.DeepEqual(p.events, []string{"flags"}) {
		t.Fatalf("events = %v, want flag read only", p.events)
	}
}

func TestObjectColliderAllowed4E7410RefreshesBeforeOrderedBoundsReads(t *testing.T) {
	p := newColliderLimitProbe4E7410()
	p.kind = 3
	p.posX = math.Float32bits(10)
	p.posY = math.Float32bits(20)
	p.boxMinX = math.Float32bits(-20)
	p.boxMaxX = math.Float32bits(20)
	p.boxMinY = math.Float32bits(-30)
	p.boxMaxY = math.Float32bits(30)
	if !objectColliderAllowed4E7410(p) {
		t.Fatal("40x60 bounds were rejected")
	}
	wantSuffix := []string{"read-max-x", "read-min-x", "read-max-y", "read-min-y"}
	if got := p.events[len(p.events)-len(wantSuffix):]; !reflect.DeepEqual(got, wantSuffix) {
		t.Fatalf("bounds read suffix = %v, want %v; all events: %v", got, wantSuffix, p.events)
	}
	for i, event := range p.events[:len(p.events)-len(wantSuffix)] {
		if len(event) >= 5 && event[:5] == "read-" {
			t.Fatalf("bounds read %q occurred before refresh completed at event %d: %v", event, i, p.events)
		}
	}
}

func TestObjectColliderAllowed4E7410WidthFailureSkipsHeightRead(t *testing.T) {
	p := newColliderLimitProbe4E7410()
	p.kind = 3
	p.boxMinX = math.Float32bits(-42.5)
	p.boxMaxX = math.Float32bits(42.5)
	p.boxMinY = math.Float32bits(-1)
	p.boxMaxY = math.Float32bits(1)
	if objectColliderAllowed4E7410(p) {
		t.Fatal("width exactly 85 was accepted")
	}
	wantSuffix := []string{"read-max-x", "read-min-x"}
	if got := p.events[len(p.events)-len(wantSuffix):]; !reflect.DeepEqual(got, wantSuffix) {
		t.Fatalf("bounds read suffix = %v, want %v; all events: %v", got, wantSuffix, p.events)
	}
	for _, event := range p.events {
		if event == "read-max-y" || event == "read-min-y" {
			t.Fatalf("height was read after width rejection: %v", p.events)
		}
	}
}

func TestObjectColliderAllowed4E7410UnorderedWidthContinuesToHeight(t *testing.T) {
	p := newColliderLimitProbe4E7410()
	p.kind = 1
	p.posX = 0x7fa12345
	p.posY = math.Float32bits(0)
	if !objectColliderAllowed4E7410(p) {
		t.Fatal("unordered width with zero height was rejected")
	}
	wantSuffix := []string{"read-max-x", "read-min-x", "read-max-y", "read-min-y"}
	if got := p.events[len(p.events)-len(wantSuffix):]; !reflect.DeepEqual(got, wantSuffix) {
		t.Fatalf("bounds read suffix = %v, want %v; all events: %v", got, wantSuffix, p.events)
	}
}

func TestObjectColliderAllowed4E7410NilFaultsOnFlagRead(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("nil object did not panic")
		}
	}()
	var p *colliderLimitProbe4E7410
	objectColliderAllowed4E7410(p)
}
