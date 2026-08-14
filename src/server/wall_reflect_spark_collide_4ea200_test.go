package server

import (
	"math"
	"reflect"
	"testing"
)

type wallReflectSparkTestData4EA200 struct {
	damage int32
}

type wallReflectSparkTestObject4EA200 struct {
	name      string
	data      *wallReflectSparkTestData4EA200
	velocityX float32
	velocityY float32
	newX      float32
	newY      float32
	version   int
}

type wallReflectSparkTestCollision4EA200 struct {
	x float32
	y float32
}

func defaultWallReflectSparkHooks4EA200() wallReflectSparkCollideHooks4EA200[
	*wallReflectSparkTestObject4EA200,
	*wallReflectSparkTestCollision4EA200,
	*wallReflectSparkTestData4EA200,
] {
	return wallReflectSparkCollideHooks4EA200[
		*wallReflectSparkTestObject4EA200,
		*wallReflectSparkTestCollision4EA200,
		*wallReflectSparkTestData4EA200,
	]{
		loadCollideData: func(obj *wallReflectSparkTestObject4EA200) *wallReflectSparkTestData4EA200 {
			return obj.data
		},
		loadDamage: func(data *wallReflectSparkTestData4EA200) int32 {
			return data.damage
		},
		findParent: func(*wallReflectSparkTestObject4EA200) *wallReflectSparkTestObject4EA200 {
			return nil
		},
		targetDamage: func(
			*wallReflectSparkTestObject4EA200,
			*wallReflectSparkTestObject4EA200,
			*wallReflectSparkTestObject4EA200,
			int32,
			uint32,
		) int32 {
			return 0
		},
		delayedDelete: func(*wallReflectSparkTestObject4EA200) {},
		loadCollisionY: func(collision *wallReflectSparkTestCollision4EA200) float32 {
			return collision.y
		},
		loadCollisionX: func(collision *wallReflectSparkTestCollision4EA200) float32 {
			return collision.x
		},
		loadVelocityX: func(obj *wallReflectSparkTestObject4EA200) float32 {
			return obj.velocityX
		},
		loadVelocityY: func(obj *wallReflectSparkTestObject4EA200) float32 {
			return obj.velocityY
		},
		storeVelocityX: func(obj *wallReflectSparkTestObject4EA200, value float32) {
			obj.velocityX = value
		},
		storeVelocityY: func(obj *wallReflectSparkTestObject4EA200, value float32) {
			obj.velocityY = value
		},
		loadNewPosY: func(obj *wallReflectSparkTestObject4EA200) float32 {
			return obj.newY
		},
		loadNewPosX: func(obj *wallReflectSparkTestObject4EA200) float32 {
			return obj.newX
		},
		floatToInt: func(value float32) int32 { return int32(value) },
		damageMap:  func(int32, int32, int32, uint32, *wallReflectSparkTestObject4EA200) {},
	}
}

func TestWallReflectSparkCollide4EA200TargetCachesDataAndUsesWholeDamageResult(t *testing.T) {
	oldData := &wallReflectSparkTestData4EA200{damage: -31}
	newData := &wallReflectSparkTestData4EA200{damage: 99}
	parent := &wallReflectSparkTestObject4EA200{name: "parent"}
	source := &wallReflectSparkTestObject4EA200{name: "source", data: oldData}
	target := &wallReflectSparkTestObject4EA200{name: "target", version: 1}
	collision := &wallReflectSparkTestCollision4EA200{x: 1, y: 1}
	events := make([]string, 0, 5)
	hooks := defaultWallReflectSparkHooks4EA200()
	hooks.loadCollideData = func(got *wallReflectSparkTestObject4EA200) *wallReflectSparkTestData4EA200 {
		events = append(events, "data")
		if got != source {
			t.Fatalf("data source = %p", got)
		}
		source.data = newData
		return oldData
	}
	hooks.loadDamage = func(data *wallReflectSparkTestData4EA200) int32 {
		events = append(events, "damage")
		if data != oldData {
			t.Fatalf("data = %p, want cached %p", data, oldData)
		}
		return data.damage
	}
	hooks.findParent = func(got *wallReflectSparkTestObject4EA200) *wallReflectSparkTestObject4EA200 {
		events = append(events, "parent")
		if got != source {
			t.Fatalf("parent source = %p", got)
		}
		target.version = 2
		return parent
	}
	hooks.targetDamage = func(
		gotTarget, gotParent, gotSource *wallReflectSparkTestObject4EA200,
		damage int32,
		damageType uint32,
	) int32 {
		events = append(events, "target-damage")
		if gotTarget != target || gotParent != parent || gotSource != source || damage != -31 || damageType != 11 {
			t.Fatalf("target damage = %p/%p/%p/%d/%d", gotTarget, gotParent, gotSource, damage, damageType)
		}
		if target.version != 2 {
			t.Fatal("Damage callback was not observed after parent lookup")
		}
		return 0x100 // AL is zero, but GAME.EXE checks the whole EAX result.
	}
	hooks.delayedDelete = func(got *wallReflectSparkTestObject4EA200) {
		events = append(events, "delete")
		if got != source {
			t.Fatalf("deleted source = %p", got)
		}
	}
	hooks.loadCollisionY = func(*wallReflectSparkTestCollision4EA200) float32 {
		t.Fatal("target path read collision")
		return 0
	}

	wallReflectSparkCollide4EA200(source, target, collision, hooks)
	want := []string{"data", "damage", "parent", "target-damage", "delete"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestWallReflectSparkCollide4EA200ZeroDamageResultKeepsSource(t *testing.T) {
	source := &wallReflectSparkTestObject4EA200{data: &wallReflectSparkTestData4EA200{damage: 7}}
	target := &wallReflectSparkTestObject4EA200{}
	hooks := defaultWallReflectSparkHooks4EA200()
	hooks.targetDamage = func(
		*wallReflectSparkTestObject4EA200,
		*wallReflectSparkTestObject4EA200,
		*wallReflectSparkTestObject4EA200,
		int32,
		uint32,
	) int32 {
		return 0
	}
	hooks.delayedDelete = func(*wallReflectSparkTestObject4EA200) {
		t.Fatal("zero Damage result deleted source")
	}
	wallReflectSparkCollide4EA200(source, target, nil, hooks)
}

func TestWallReflectSparkCollide4EA200WallOrderCachedDataAndLiveX(t *testing.T) {
	oldData := &wallReflectSparkTestData4EA200{damage: -47}
	newData := &wallReflectSparkTestData4EA200{damage: 88}
	source := &wallReflectSparkTestObject4EA200{
		name:      "source",
		data:      oldData,
		velocityX: 7,
		velocityY: -11,
		newX:      69,
		newY:      46,
	}
	tiny := math.Float32frombits(1)
	collision := &wallReflectSparkTestCollision4EA200{x: tiny, y: tiny}
	events := make([]string, 0, 14)
	inputs := make([]uint32, 0, 2)
	hooks := defaultWallReflectSparkHooks4EA200()
	hooks.loadCollideData = func(got *wallReflectSparkTestObject4EA200) *wallReflectSparkTestData4EA200 {
		events = append(events, "data")
		if got != source {
			t.Fatalf("data source = %p", got)
		}
		return got.data
	}
	hooks.loadCollisionY = func(got *wallReflectSparkTestCollision4EA200) float32 {
		events = append(events, "collision-y")
		if got != collision {
			t.Fatalf("collision Y source = %p", got)
		}
		source.data = newData
		return got.y
	}
	hooks.loadCollisionX = func(got *wallReflectSparkTestCollision4EA200) float32 {
		events = append(events, "collision-x")
		return got.x
	}
	hooks.loadVelocityX = func(obj *wallReflectSparkTestObject4EA200) float32 {
		events = append(events, "velocity-x")
		return obj.velocityX
	}
	hooks.loadVelocityY = func(obj *wallReflectSparkTestObject4EA200) float32 {
		events = append(events, "velocity-y")
		return obj.velocityY
	}
	hooks.storeVelocityX = func(obj *wallReflectSparkTestObject4EA200, value float32) {
		events = append(events, "store-velocity-x")
		obj.velocityX = value
	}
	hooks.storeVelocityY = func(obj *wallReflectSparkTestObject4EA200, value float32) {
		events = append(events, "store-velocity-y")
		obj.velocityY = value
	}
	hooks.loadNewPosY = func(obj *wallReflectSparkTestObject4EA200) float32 {
		events = append(events, "new-y")
		return obj.newY
	}
	hooks.loadDamage = func(data *wallReflectSparkTestData4EA200) int32 {
		events = append(events, "damage")
		if data != oldData {
			t.Fatalf("damage data = %p, want cached %p", data, oldData)
		}
		return data.damage
	}
	hooks.floatToInt = func(value float32) int32 {
		inputs = append(inputs, math.Float32bits(value))
		if len(inputs) == 1 {
			events = append(events, "round-y")
			source.newX = 92
			return -17
		}
		events = append(events, "round-x")
		return 29
	}
	hooks.loadNewPosX = func(obj *wallReflectSparkTestObject4EA200) float32 {
		events = append(events, "new-x")
		return obj.newX
	}
	hooks.damageMap = func(x, y, damage int32, damageType uint32, got *wallReflectSparkTestObject4EA200) {
		events = append(events, "map")
		if x != 29 || y != -17 || damage != -47 || damageType != 11 || got != source {
			t.Fatalf("map = %d/%d/%d/%d/%p", x, y, damage, damageType, got)
		}
	}
	hooks.delayedDelete = func(*wallReflectSparkTestObject4EA200) {
		t.Fatal("wall path deleted source")
	}

	wallReflectSparkCollide4EA200(source, nil, collision, hooks)
	wantEvents := []string{
		"data", "collision-y", "collision-x", "velocity-x", "velocity-y",
		"store-velocity-x", "store-velocity-y", "new-y", "damage", "round-y",
		"new-x", "round-x", "map",
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
	if source.velocityX != 11 || source.velocityY != -7 {
		t.Fatalf("velocity = {%v %v}, want {11 -7}", source.velocityX, source.velocityY)
	}
	gridInverse := math.Float32frombits(wallReflectSparkGridInverseBits4EA200)
	wantInputs := []uint32{
		math.Float32bits(float32(46) * gridInverse),
		math.Float32bits(float32(92) * gridInverse),
	}
	if !reflect.DeepEqual(inputs, wantInputs) {
		t.Fatalf("round inputs = %#v, want %#v", inputs, wantInputs)
	}
}

func TestWallReflectSparkCollide4EA200WallProductBranches(t *testing.T) {
	tiny := math.Float32frombits(1)
	nan := math.Float32frombits(0x7fc12345)
	tests := []struct {
		name      string
		collision wallReflectSparkTestCollision4EA200
		velocityX uint32
		velocityY uint32
		wantX     uint32
		wantY     uint32
	}{
		{
			name:      "positive tiny product stays extended",
			collision: wallReflectSparkTestCollision4EA200{x: tiny, y: tiny},
			velocityX: math.Float32bits(7), velocityY: math.Float32bits(-11),
			wantX: math.Float32bits(11), wantY: math.Float32bits(-7),
		},
		{
			name:      "negative product plain swap",
			collision: wallReflectSparkTestCollision4EA200{x: 1, y: -1},
			velocityX: 0x7fc12345, velocityY: 0x80000000,
			wantX: 0x80000000, wantY: 0x7fc12345,
		},
		{
			name:      "zero product plain swap",
			collision: wallReflectSparkTestCollision4EA200{x: 0, y: -1},
			velocityX: math.Float32bits(3), velocityY: math.Float32bits(5),
			wantX: math.Float32bits(5), wantY: math.Float32bits(3),
		},
		{
			name:      "unordered product plain swap",
			collision: wallReflectSparkTestCollision4EA200{x: nan, y: 1},
			velocityX: math.Float32bits(13), velocityY: math.Float32bits(-17),
			wantX: math.Float32bits(-17), wantY: math.Float32bits(13),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			source := &wallReflectSparkTestObject4EA200{
				data:      &wallReflectSparkTestData4EA200{},
				velocityX: math.Float32frombits(tc.velocityX),
				velocityY: math.Float32frombits(tc.velocityY),
			}
			hooks := defaultWallReflectSparkHooks4EA200()
			wallReflectSparkCollide4EA200(source, nil, &tc.collision, hooks)
			if got := math.Float32bits(source.velocityX); got != tc.wantX {
				t.Errorf("velocity X bits = %#08x, want %#08x", got, tc.wantX)
			}
			if got := math.Float32bits(source.velocityY); got != tc.wantY {
				t.Errorf("velocity Y bits = %#08x, want %#08x", got, tc.wantY)
			}
		})
	}
}

func TestWallReflectSparkCollide4EA200PlainSwapStoresXBeforeY(t *testing.T) {
	source := &wallReflectSparkTestObject4EA200{
		data:      &wallReflectSparkTestData4EA200{},
		velocityX: 3,
		velocityY: 5,
	}
	collision := &wallReflectSparkTestCollision4EA200{x: 1, y: -1}
	events := make([]string, 0, 2)
	hooks := defaultWallReflectSparkHooks4EA200()
	hooks.storeVelocityX = func(obj *wallReflectSparkTestObject4EA200, value float32) {
		events = append(events, "store-x")
		obj.velocityX = value
		obj.velocityY = 99
	}
	hooks.storeVelocityY = func(obj *wallReflectSparkTestObject4EA200, value float32) {
		events = append(events, "store-y")
		obj.velocityY = value
	}
	wallReflectSparkCollide4EA200(source, nil, collision, hooks)
	if !reflect.DeepEqual(events, []string{"store-x", "store-y"}) {
		t.Fatalf("stores = %v", events)
	}
	if source.velocityX != 5 || source.velocityY != 3 {
		t.Fatalf("velocity = {%v %v}, want cached {5 3}", source.velocityX, source.velocityY)
	}
}

func TestWallReflectSparkCollide4EA200NoTargetOrWallReturnsAfterDataCache(t *testing.T) {
	source := &wallReflectSparkTestObject4EA200{}
	events := make([]string, 0, 1)
	hooks := defaultWallReflectSparkHooks4EA200()
	hooks.loadCollideData = func(got *wallReflectSparkTestObject4EA200) *wallReflectSparkTestData4EA200 {
		events = append(events, "data")
		return got.data
	}
	hooks.loadDamage = func(*wallReflectSparkTestData4EA200) int32 {
		t.Fatal("empty collision loaded damage")
		return 0
	}
	hooks.delayedDelete = func(*wallReflectSparkTestObject4EA200) {
		t.Fatal("empty collision deleted source")
	}
	wallReflectSparkCollide4EA200(source, nil, nil, hooks)
	if !reflect.DeepEqual(events, []string{"data"}) {
		t.Fatalf("events = %v", events)
	}
}
