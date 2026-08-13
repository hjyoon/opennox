package server

import (
	"math"
	"reflect"
	"testing"
)

func TestObjectUpdateCollider4E7350CenterUsesNewPositionBitsAndOrder(t *testing.T) {
	p := &colliderProbe4E7290{
		kind:    1,
		posX:    float32Bits4E7290(100),
		posY:    float32Bits4E7290(200),
		newPosX: 0x7fa54321,
		newPosY: 0x80000000,
	}
	if got := objectUpdateCollider4E7350(p); got != p {
		t.Fatalf("return = %p, want input %p", got, p)
	}
	if got, want := [4]uint32{p.minX, p.minY, p.maxX, p.maxY}, [4]uint32{p.newPosX, p.newPosY, p.newPosX, p.newPosY}; got != want {
		t.Fatalf("bounds = %#v, want raw NewPos copies %#v", got, want)
	}
	wantEvents := []string{"kind", "new-pos-x", "min-x", "new-pos-y", "max-x", "min-y", "max-y"}
	if !reflect.DeepEqual(p.events, wantEvents) {
		t.Fatalf("events = %v, want %v", p.events, wantEvents)
	}
}

func TestObjectUpdateCollider4E7350CircleUsesNewPositionAndOrder(t *testing.T) {
	p := &colliderProbe4E7290{
		kind:    2,
		posX:    float32Bits4E7290(100),
		posY:    float32Bits4E7290(200),
		newPosX: float32Bits4E7290(12.5),
		newPosY: float32Bits4E7290(-3.25),
		radius:  float32Bits4E7290(2.5),
	}
	objectUpdateCollider4E7350(p)
	want := [4]uint32{
		float32Bits4E7290(10), float32Bits4E7290(-5.75),
		float32Bits4E7290(15), float32Bits4E7290(-0.75),
	}
	if got := [4]uint32{p.minX, p.minY, p.maxX, p.maxY}; got != want {
		t.Fatalf("bounds = %#v, want %#v", got, want)
	}
	wantEvents := []string{
		"kind",
		"new-pos-x", "radius", "min-x",
		"new-pos-y", "radius", "min-y",
		"radius", "new-pos-x", "max-x",
		"radius", "new-pos-y", "max-y",
	}
	if !reflect.DeepEqual(p.events, wantEvents) {
		t.Fatalf("events = %v, want %v", p.events, wantEvents)
	}
}

func TestObjectUpdateCollider4E7350BoxUsesNewPositionAndOrder(t *testing.T) {
	p := &colliderProbe4E7290{
		kind:    3,
		posX:    float32Bits4E7290(100),
		posY:    float32Bits4E7290(200),
		newPosX: float32Bits4E7290(10),
		newPosY: float32Bits4E7290(20),
		boxMinX: float32Bits4E7290(-4),
		boxMinY: float32Bits4E7290(-7),
		boxMaxX: float32Bits4E7290(6),
		boxMaxY: float32Bits4E7290(9),
	}
	objectUpdateCollider4E7350(p)
	want := [4]uint32{
		float32Bits4E7290(6), float32Bits4E7290(13),
		float32Bits4E7290(16), float32Bits4E7290(29),
	}
	if got := [4]uint32{p.minX, p.minY, p.maxX, p.maxY}; got != want {
		t.Fatalf("bounds = %#v, want %#v", got, want)
	}
	wantEvents := []string{
		"kind",
		"box-min-x", "new-pos-x", "min-x",
		"box-min-y", "new-pos-y", "min-y",
		"box-max-x", "new-pos-x", "max-x",
		"box-max-y", "new-pos-y", "max-y",
	}
	if !reflect.DeepEqual(p.events, wantEvents) {
		t.Fatalf("events = %v, want %v", p.events, wantEvents)
	}
}

func TestObjectUpdateCollider4E7350UnknownShapeDoesNotReadPositionOrWrite(t *testing.T) {
	for _, kind := range []uint32{0, 4, math.MaxUint32} {
		p := &colliderProbe4E7290{
			kind: kind,
			minX: 1, minY: 2, maxX: 3, maxY: 4,
		}
		if got := objectUpdateCollider4E7350(p); got != p {
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

func TestObjectUpdateCollider4E7350NilFaultsOnKind(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("nil object did not panic")
		}
	}()
	var p *colliderProbe4E7290
	objectUpdateCollider4E7350(p)
}
