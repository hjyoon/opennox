package server

import (
	"math"
	"reflect"
	"testing"
)

type colliderProbe4E7290 struct {
	kind                   uint32
	posX, posY, radius     uint32
	boxMinX, boxMinY       uint32
	boxMaxX, boxMaxY       uint32
	minX, minY, maxX, maxY uint32
	events                 []string
}

func (p *colliderProbe4E7290) event(name string) { p.events = append(p.events, name) }
func (p *colliderProbe4E7290) colliderKind4E7290() uint32 {
	p.event("kind")
	return p.kind
}
func (p *colliderProbe4E7290) colliderPosXBits4E7290() uint32 {
	p.event("pos-x")
	return p.posX
}
func (p *colliderProbe4E7290) colliderPosYBits4E7290() uint32 {
	p.event("pos-y")
	return p.posY
}
func (p *colliderProbe4E7290) colliderRadiusBits4E7290() uint32 {
	p.event("radius")
	return p.radius
}
func (p *colliderProbe4E7290) colliderBoxMinXBits4E7290() uint32 {
	p.event("box-min-x")
	return p.boxMinX
}
func (p *colliderProbe4E7290) colliderBoxMinYBits4E7290() uint32 {
	p.event("box-min-y")
	return p.boxMinY
}
func (p *colliderProbe4E7290) colliderBoxMaxXBits4E7290() uint32 {
	p.event("box-max-x")
	return p.boxMaxX
}
func (p *colliderProbe4E7290) colliderBoxMaxYBits4E7290() uint32 {
	p.event("box-max-y")
	return p.boxMaxY
}
func (p *colliderProbe4E7290) colliderStoreMinXBits4E7290(v uint32) {
	p.event("min-x")
	p.minX = v
}
func (p *colliderProbe4E7290) colliderStoreMinYBits4E7290(v uint32) {
	p.event("min-y")
	p.minY = v
}
func (p *colliderProbe4E7290) colliderStoreMaxXBits4E7290(v uint32) {
	p.event("max-x")
	p.maxX = v
}
func (p *colliderProbe4E7290) colliderStoreMaxYBits4E7290(v uint32) {
	p.event("max-y")
	p.maxY = v
}

func float32Bits4E7290(v float32) uint32 { return math.Float32bits(v) }

func TestObjectUpdateCollider4E7290CenterBitCopiesAndOrder(t *testing.T) {
	p := &colliderProbe4E7290{
		kind: 1,
		posX: 0x7fa12345,
		posY: 0x80000000,
	}
	if got := objectUpdateCollider4E7290(p); got != p {
		t.Fatalf("return = %p, want input %p", got, p)
	}
	if got, want := [4]uint32{p.minX, p.minY, p.maxX, p.maxY}, [4]uint32{p.posX, p.posY, p.posX, p.posY}; got != want {
		t.Fatalf("bounds = %#v, want raw copies %#v", got, want)
	}
	wantEvents := []string{"kind", "pos-x", "min-x", "pos-y", "max-x", "min-y", "max-y"}
	if !reflect.DeepEqual(p.events, wantEvents) {
		t.Fatalf("events = %v, want %v", p.events, wantEvents)
	}
}

func TestObjectUpdateCollider4E7290CircleArithmeticAndOrder(t *testing.T) {
	p := &colliderProbe4E7290{
		kind:   2,
		posX:   float32Bits4E7290(12.5),
		posY:   float32Bits4E7290(-3.25),
		radius: float32Bits4E7290(2.5),
	}
	objectUpdateCollider4E7290(p)
	want := [4]uint32{
		float32Bits4E7290(10), float32Bits4E7290(-5.75),
		float32Bits4E7290(15), float32Bits4E7290(-0.75),
	}
	if got := [4]uint32{p.minX, p.minY, p.maxX, p.maxY}; got != want {
		t.Fatalf("bounds = %#v, want %#v", got, want)
	}
	wantEvents := []string{
		"kind",
		"pos-x", "radius", "min-x",
		"pos-y", "radius", "min-y",
		"radius", "pos-x", "max-x",
		"radius", "pos-y", "max-y",
	}
	if !reflect.DeepEqual(p.events, wantEvents) {
		t.Fatalf("events = %v, want %v", p.events, wantEvents)
	}
}

func TestObjectUpdateCollider4E7290BoxFieldMappingAndOrder(t *testing.T) {
	p := &colliderProbe4E7290{
		kind:    3,
		posX:    float32Bits4E7290(10),
		posY:    float32Bits4E7290(20),
		boxMinX: float32Bits4E7290(-4),
		boxMinY: float32Bits4E7290(-7),
		boxMaxX: float32Bits4E7290(6),
		boxMaxY: float32Bits4E7290(9),
	}
	objectUpdateCollider4E7290(p)
	want := [4]uint32{
		float32Bits4E7290(6), float32Bits4E7290(13),
		float32Bits4E7290(16), float32Bits4E7290(29),
	}
	if got := [4]uint32{p.minX, p.minY, p.maxX, p.maxY}; got != want {
		t.Fatalf("bounds = %#v, want %#v", got, want)
	}
	wantEvents := []string{
		"kind",
		"box-min-x", "pos-x", "min-x",
		"box-min-y", "pos-y", "min-y",
		"box-max-x", "pos-x", "max-x",
		"box-max-y", "pos-y", "max-y",
	}
	if !reflect.DeepEqual(p.events, wantEvents) {
		t.Fatalf("events = %v, want %v", p.events, wantEvents)
	}
}

func TestObjectUpdateCollider4E7290UnknownShapeDoesNotWrite(t *testing.T) {
	for _, kind := range []uint32{0, 4, math.MaxUint32} {
		p := &colliderProbe4E7290{
			kind: kind,
			minX: 1, minY: 2, maxX: 3, maxY: 4,
		}
		if got := objectUpdateCollider4E7290(p); got != p {
			t.Fatalf("kind %d return = %p, want %p", kind, got, p)
		}
		if got, want := [4]uint32{p.minX, p.minY, p.maxX, p.maxY}, [4]uint32{1, 2, 3, 4}; got != want {
			t.Fatalf("kind %d bounds = %#v, want unchanged %#v", kind, got, want)
		}
		if !reflect.DeepEqual(p.events, []string{"kind"}) {
			t.Fatalf("kind %d events = %v, want only kind", kind, p.events)
		}
	}
}

func TestObjectUpdateCollider4E7290NilFaultsOnKind(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("nil object did not panic")
		}
	}()
	var p *colliderProbe4E7290
	objectUpdateCollider4E7290(p)
}
