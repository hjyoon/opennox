package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type deathBallTestObject4E9E90 struct {
	class      uint8
	update     int
	previousX  float32
	previousY  float32
	newX       float32
	newY       float32
	positionX  float32
	positionY  float32
	direction  int32
	directionX int32
	directionY int32
}

type deathBallTraceTestPoint4E9E90 struct {
	x int32
	y int32
}

func TestDeathBallCollide4E9E90DoorOrderCachesCopyButReloadsDelta(t *testing.T) {
	source := &deathBallTestObject4E9E90{previousX: 2, previousY: 3}
	target := &deathBallTestObject4E9E90{
		class:      deathBallDoorClassMask4E9E90,
		update:     91,
		positionX:  100,
		positionY:  50,
		direction:  7,
		directionX: 18,
		directionY: -27,
	}
	var events []string
	prevXReads := 0
	prevYReads := 0
	directionReads := 0

	deathBallCollide4E9E90(source, target, 0, deathBallCollideHooks4E9E90[
		*deathBallTestObject4E9E90,
		int,
		int,
		*deathBallTraceTestPoint4E9E90,
	]{
		loadClassByte: func(obj *deathBallTestObject4E9E90) uint8 {
			events = append(events, "class")
			return obj.class
		},
		loadDoorUpdate: func(obj *deathBallTestObject4E9E90) int {
			events = append(events, "update")
			return obj.update
		},
		loadPrevX: func(obj *deathBallTestObject4E9E90) float32 {
			prevXReads++
			events = append(events, fmt.Sprintf("prev-x:%d", prevXReads))
			return obj.previousX
		},
		loadPrevY: func(obj *deathBallTestObject4E9E90) float32 {
			prevYReads++
			events = append(events, fmt.Sprintf("prev-y:%d", prevYReads))
			return obj.previousY
		},
		storeNewX: func(obj *deathBallTestObject4E9E90, value float32) {
			events = append(events, "store-new-x")
			obj.newX = value
			// The original has already cached both words for the copy.
			obj.previousY = 30
		},
		storeNewY: func(obj *deathBallTestObject4E9E90, value float32) {
			events = append(events, "store-new-y")
			obj.newY = value
			// Delta X is a later live reload.
			obj.previousX = 20
		},
		loadPosX: func(obj *deathBallTestObject4E9E90) float32 {
			events = append(events, "pos-x")
			return obj.positionX
		},
		loadPosY: func(obj *deathBallTestObject4E9E90) float32 {
			events = append(events, "pos-y")
			return obj.positionY
		},
		loadDoorDirection: func(update int) int32 {
			events = append(events, "direction")
			directionReads++
			if update != target.update {
				t.Fatalf("update = %d, want %d", update, target.update)
			}
			if directionReads == 2 {
				return 8
			}
			return target.direction
		},
		loadDirectionY: func(direction int32) int32 {
			events = append(events, "direction-y")
			if direction != target.direction {
				t.Fatalf("Y direction = %d", direction)
			}
			return target.directionY
		},
		loadDirectionX: func(direction int32) int32 {
			events = append(events, "direction-x")
			if direction != 8 {
				t.Fatalf("X direction = %d", direction)
			}
			return target.directionX
		},
		doorReflect: func(got *deathBallTestObject4E9E90, normalX, normalY float32) {
			events = append(events, "reflect")
			if got != source || normalX != -27 || normalY != -18 {
				t.Fatalf("reflect = %p/(%g,%g), want %p/(-27,-18)", got, normalX, normalY, source)
			}
		},
		audio: func(id uint32, got *deathBallTestObject4E9E90) {
			events = append(events, "audio")
			if id != deathBallDoorReflectAudio4E9E90 || got != source {
				t.Fatalf("audio = %d/%p", id, got)
			}
		},
		balanceFloat: func(string) float64 {
			t.Fatal("Door branch loaded balance")
			return 0
		},
		wallReflect: func(int, *deathBallTestObject4E9E90) {
			t.Fatal("Door branch used collision pointer")
		},
	})

	wantEvents := []string{
		"class", "update", "prev-x:1", "prev-y:1", "store-new-x", "store-new-y",
		"pos-x", "prev-x:2", "pos-y", "direction", "prev-y:2",
		"direction-y", "direction", "direction-x", "reflect", "audio",
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %#v, want %#v", events, wantEvents)
	}
	if source.newX != 2 || source.newY != 3 {
		t.Fatalf("NewPos = (%g,%g), want cached (2,3)", source.newX, source.newY)
	}
}

func TestDeathBallCollide4E9E90NonDoorBalanceParentAndLiveDamage(t *testing.T) {
	source := &deathBallTestObject4E9E90{}
	target := &deathBallTestObject4E9E90{class: 0x7f}
	parent := &deathBallTestObject4E9E90{}
	var events []string

	deathBallCollide4E9E90(source, target, 99, deathBallCollideHooks4E9E90[
		*deathBallTestObject4E9E90,
		int,
		int,
		*deathBallTraceTestPoint4E9E90,
	]{
		loadClassByte: func(obj *deathBallTestObject4E9E90) uint8 {
			events = append(events, "class")
			return obj.class
		},
		balanceFloat: func(key string) float64 {
			events = append(events, "balance")
			if key != deathBallCollideDamageKey4E9E90 {
				t.Fatalf("balance key = %q", key)
			}
			return 10.5
		},
		floatToInt: func(value float32) int32 {
			events = append(events, "round")
			if math.Float32bits(value) != math.Float32bits(10.5) {
				t.Fatalf("round input = %g", value)
			}
			return 10
		},
		findParent: func(got *deathBallTestObject4E9E90) *deathBallTestObject4E9E90 {
			events = append(events, "parent")
			if got != source {
				t.Fatalf("parent source = %p", got)
			}
			return parent
		},
		targetDamage: func(gotTarget, gotParent, gotSource *deathBallTestObject4E9E90, damage int32, damageType uint32) int32 {
			events = append(events, "damage")
			if gotTarget != target || gotParent != parent || gotSource != source ||
				damage != 10 || damageType != deathBallDamageType4E9E90 {
				t.Fatalf("damage args = %p/%p/%p/%d/%d", gotTarget, gotParent, gotSource, damage, damageType)
			}
			// The callback return is ignored, including a nonzero full EAX.
			return 0x100
		},
		wallReflect: func(int, *deathBallTestObject4E9E90) {
			t.Fatal("target branch used collision pointer")
		},
	})

	if want := []string{"class", "balance", "round", "parent", "damage"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
}

func TestDeathBallCollide4E9E90WallTraceBalanceThenYBeforeX(t *testing.T) {
	source := &deathBallTestObject4E9E90{}
	point := &deathBallTraceTestPoint4E9E90{x: 5, y: 4}
	var events []string

	deathBallCollide4E9E90(source, nil, 73, deathBallCollideHooks4E9E90[
		*deathBallTestObject4E9E90,
		int,
		int,
		*deathBallTraceTestPoint4E9E90,
	]{
		loadClassByte: func(*deathBallTestObject4E9E90) uint8 {
			t.Fatal("nil target read class")
			return 0
		},
		wallReflect: func(collision int, got *deathBallTestObject4E9E90) {
			events = append(events, "reflect")
			if collision != 73 || got != source {
				t.Fatalf("reflect = %d/%p", collision, got)
			}
		},
		audio: func(id uint32, got *deathBallTestObject4E9E90) {
			events = append(events, "audio")
			if id != deathBallWallReflectAudio4E9E90 || got != source {
				t.Fatalf("audio = %d/%p", id, got)
			}
		},
		traceHitPoint: func() *deathBallTraceTestPoint4E9E90 {
			events = append(events, "trace")
			return point
		},
		balanceFloat: func(key string) float64 {
			events = append(events, "balance")
			point.y = 40
			return 7.75
		},
		floatToInt: func(value float32) int32 {
			events = append(events, "round")
			if value != 7.75 {
				t.Fatalf("round input = %g", value)
			}
			point.x = 50
			return 8
		},
		loadTraceY: func(got *deathBallTraceTestPoint4E9E90) int32 {
			events = append(events, "trace-y")
			got.x = 60
			return got.y
		},
		loadTraceX: func(got *deathBallTraceTestPoint4E9E90) int32 {
			events = append(events, "trace-x")
			return got.x
		},
		damageMap: func(x, y, damage int32, damageType uint32, got *deathBallTestObject4E9E90) {
			events = append(events, "map")
			if x != 60 || y != 40 || damage != 8 || damageType != deathBallDamageType4E9E90 || got != source {
				t.Fatalf("map = %d/%d/%d/%d/%p", x, y, damage, damageType, got)
			}
		},
	})

	want := []string{"reflect", "audio", "trace", "balance", "round", "trace-y", "trace-x", "map"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
}

func TestDeathBallCollide4E9E90NilWallAndMissingTraceShortCircuit(t *testing.T) {
	var events []string
	hooks := deathBallCollideHooks4E9E90[int, int, int, int]{
		wallReflect: func(int, int) { events = append(events, "reflect") },
		audio:       func(uint32, int) { events = append(events, "audio") },
		traceHitPoint: func() int {
			events = append(events, "trace")
			return 0
		},
		balanceFloat: func(string) float64 {
			t.Fatal("missing trace loaded balance")
			return 0
		},
	}
	deathBallCollide4E9E90(0, 0, 0, hooks)
	if len(events) != 0 {
		t.Fatalf("nil collision events = %#v", events)
	}
	deathBallCollide4E9E90(1, 0, 2, hooks)
	if want := []string{"reflect", "audio", "trace"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("missing trace events = %#v, want %#v", events, want)
	}
}

func TestDeathBallTraceHitResult537760ReadyGate(t *testing.T) {
	pointCalls := 0
	point := func() int {
		pointCalls++
		return 91
	}
	if got := deathBallTraceHitResult537760(func() uint32 { return 0 }, point); got != 0 || pointCalls != 0 {
		t.Fatalf("zero ready = %d, point calls %d", got, pointCalls)
	}
	if got := deathBallTraceHitResult537760(func() uint32 { return 0x80000000 }, point); got != 91 || pointCalls != 1 {
		t.Fatalf("nonzero ready = %d, point calls %d", got, pointCalls)
	}
}

func TestDeathBallDoorDirection4E9E90ExactTableAndBounds(t *testing.T) {
	want := [...]deathBallDirection4E9E90{
		{x: -23, y: -23}, {x: -18, y: -27}, {x: -12, y: -30}, {x: -6, y: -31},
		{x: 0, y: -32}, {x: 6, y: -31}, {x: 12, y: -30}, {x: 18, y: -27},
		{x: 23, y: -23}, {x: 27, y: -18}, {x: 30, y: -12}, {x: 31, y: -6},
		{x: 32, y: 0}, {x: 31, y: 6}, {x: 30, y: 12}, {x: 27, y: 18},
		{x: 23, y: 23}, {x: 18, y: 27}, {x: 12, y: 30}, {x: 6, y: 31},
		{x: 0, y: 32}, {x: -6, y: 31}, {x: -12, y: 30}, {x: -18, y: 27},
		{x: -23, y: 23}, {x: -27, y: 18}, {x: -30, y: 12}, {x: -31, y: 6},
		{x: -32, y: 0}, {x: -31, y: -6}, {x: -30, y: -12}, {x: -27, y: -18},
	}
	for i, expected := range want {
		if got := deathBallDoorDirection4E9E90(int32(i)); got != expected {
			t.Fatalf("direction %d = %#v, want %#v", i, got, expected)
		}
	}
	for _, invalid := range []int32{-1, 32, math.MaxInt32} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("direction %d did not panic", invalid)
				}
			}()
			_ = deathBallDoorDirection4E9E90(invalid)
		}()
	}
}

func TestDeathBallDoorReflect57B770RepresentativeBits(t *testing.T) {
	tests := []struct {
		velocityX float32
		velocityY float32
		normalX   float32
		normalY   float32
		wantX     uint32
		wantY     uint32
	}{
		{velocityX: 3, velocityY: 4, normalX: 0, normalY: 32, wantX: 0x403ece3c, wantY: 0xc07e6850},
		{velocityX: -7.25, velocityY: 1.5, normalX: 23, normalY: -23, wantX: 0x3fbed32a, wantY: 0xc0e6947e},
		{velocityX: math.Float32frombits(0x80000000), velocityY: 0, normalX: 0, normalY: 0, wantX: 0x80000000, wantY: 0},
	}
	for _, tc := range tests {
		x, y := deathBallDoorReflect57B770(tc.velocityX, tc.velocityY, tc.normalX, tc.normalY)
		if gotX, gotY := math.Float32bits(x), math.Float32bits(y); gotX != tc.wantX || gotY != tc.wantY {
			t.Fatalf("velocity=(%08x,%08x) normal=(%08x,%08x) result=(%08x,%08x), want (%08x,%08x)",
				math.Float32bits(tc.velocityX), math.Float32bits(tc.velocityY),
				math.Float32bits(tc.normalX), math.Float32bits(tc.normalY),
				gotX, gotY, tc.wantX, tc.wantY)
		}
	}
}

func TestDeathBallDoorReflect57B770ReloadAndStoreOrder(t *testing.T) {
	type velocity struct {
		x float32
		y float32
	}
	value := &velocity{x: 3, y: 4}
	var events []string
	xReads := 0
	yReads := 0
	deathBallDoorReflectCore57B770(value, 0, 32, deathBallDoorReflectHooks57B770[*velocity]{
		loadVelocityX: func(got *velocity) float32 {
			events = append(events, "load-x")
			xReads++
			if got != value {
				t.Fatalf("X receiver = %p, want %p", got, value)
			}
			if xReads == 2 {
				return 30
			}
			return got.x
		},
		loadVelocityY: func(got *velocity) float32 {
			events = append(events, "load-y")
			yReads++
			if got != value {
				t.Fatalf("Y receiver = %p, want %p", got, value)
			}
			if yReads == 2 {
				return 40
			}
			return got.y
		},
		storeVelocityX: func(got *velocity, result float32) {
			events = append(events, "store-x")
			got.x = result
		},
		storeVelocityY: func(got *velocity, result float32) {
			events = append(events, "store-y")
			got.y = result
		},
	})
	if want := []string{"load-x", "load-y", "load-y", "load-x", "store-x", "store-y"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
	if xReads != 2 || yReads != 2 {
		t.Fatalf("reads = X:%d Y:%d, want 2/2", xReads, yReads)
	}
}
