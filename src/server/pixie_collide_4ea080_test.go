package server

import (
	"math"
	"reflect"
	"testing"
)

type pixieCollideTestData4EA080 struct {
	damage int32
}

type pixieCollideTestObject4EA080 struct {
	name          string
	class         uint32
	flags         uint32
	owner         *pixieCollideTestObject4EA080
	data          *pixieCollideTestData4EA080
	posX          float32
	posY          float32
	newX          float32
	newY          float32
	velocityX     float32
	velocityY     float32
	direction1    int16
	direction2    uint16
	damageVersion int
}

func defaultPixieCollideHooks4EA080() pixieCollideHooks4EA080[
	*pixieCollideTestObject4EA080,
	int,
	*pixieCollideTestData4EA080,
] {
	return pixieCollideHooks4EA080[*pixieCollideTestObject4EA080, int, *pixieCollideTestData4EA080]{
		loadCollideData: func(obj *pixieCollideTestObject4EA080) *pixieCollideTestData4EA080 { return obj.data },
		isEnemy:         func(*pixieCollideTestObject4EA080, *pixieCollideTestObject4EA080) int32 { return 1 },
		loadClass:       func(obj *pixieCollideTestObject4EA080) uint32 { return obj.class },
		loadFlags:       func(obj *pixieCollideTestObject4EA080) uint32 { return obj.flags },
		loadOwner:       func(obj *pixieCollideTestObject4EA080) *pixieCollideTestObject4EA080 { return obj.owner },
		checkInversion:  func(*pixieCollideTestObject4EA080, *pixieCollideTestObject4EA080) int32 { return 0 },
		changeOwner:     func(*pixieCollideTestObject4EA080, *pixieCollideTestObject4EA080) {},
		hasEnchant:      func(*pixieCollideTestObject4EA080, uint32) int32 { return 0 },
		loadDirection:   func(obj *pixieCollideTestObject4EA080) int16 { return obj.direction1 },
		checkDirection:  func(*pixieCollideTestObject4EA080, int16, *pixieCollideTestObject4EA080) int32 { return 0 },
		loadDamage:      func(data *pixieCollideTestData4EA080) int32 { return data.damage },
		findParent:      func(obj *pixieCollideTestObject4EA080) *pixieCollideTestObject4EA080 { return obj.owner },
		targetDamage: func(*pixieCollideTestObject4EA080, *pixieCollideTestObject4EA080, *pixieCollideTestObject4EA080, int32, uint32) int32 {
			return 0
		},
		audio:           func(uint32, *pixieCollideTestObject4EA080) {},
		delayedDelete:   func(*pixieCollideTestObject4EA080) {},
		wallReflect:     func(int, *pixieCollideTestObject4EA080) {},
		vectorDirection: func(*pixieCollideTestObject4EA080) int32 { return 0 },
		loadVelocityX:   func(obj *pixieCollideTestObject4EA080) float32 { return obj.velocityX },
		loadVelocityY:   func(obj *pixieCollideTestObject4EA080) float32 { return obj.velocityY },
		loadNewPosX:     func(obj *pixieCollideTestObject4EA080) float32 { return obj.newX },
		loadNewPosY:     func(obj *pixieCollideTestObject4EA080) float32 { return obj.newY },
		storeDirection2: func(obj *pixieCollideTestObject4EA080, value uint16) { obj.direction2 = value },
		storeNewPosX:    func(obj *pixieCollideTestObject4EA080, value float32) { obj.newX = value },
		storeNewPosY:    func(obj *pixieCollideTestObject4EA080, value float32) { obj.newY = value },
		floatToInt:      func(value float32) int32 { return int32(value) },
		damageMap:       func(int32, int32, int32, uint32, *pixieCollideTestObject4EA080) {},
	}
}

func TestPixieCollide4EA080EarlyTargetGuardsAreSilent(t *testing.T) {
	tests := []struct {
		name        string
		enemy       int32
		targetClass uint32
		targetFlags uint32
	}{
		{name: "not enemy", enemy: 0, targetClass: pixiePlayerClass4EA080},
		{name: "unsupported class", enemy: 1, targetClass: 1},
		{name: "flag 0x20", enemy: 1, targetClass: 2, targetFlags: 0x20},
		{name: "flag 0x8000", enemy: 1, targetClass: 0x20000, targetFlags: 0x8000},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			source := &pixieCollideTestObject4EA080{data: &pixieCollideTestData4EA080{damage: 7}}
			target := &pixieCollideTestObject4EA080{class: tc.targetClass, flags: tc.targetFlags}
			hooks := defaultPixieCollideHooks4EA080()
			hooks.isEnemy = func(gotSource, gotTarget *pixieCollideTestObject4EA080) int32 {
				if gotSource != source || gotTarget != target {
					t.Fatalf("enemy args = %p/%p", gotSource, gotTarget)
				}
				return tc.enemy
			}
			hooks.checkInversion = func(*pixieCollideTestObject4EA080, *pixieCollideTestObject4EA080) int32 {
				t.Fatal("guarded collision checked inversion")
				return 0
			}
			hooks.loadDamage = func(*pixieCollideTestData4EA080) int32 {
				t.Fatal("guarded collision loaded damage")
				return 0
			}
			hooks.wallReflect = func(int, *pixieCollideTestObject4EA080) {
				t.Fatal("target collision reflected wall")
			}
			hooks.delayedDelete = func(*pixieCollideTestObject4EA080) {
				t.Fatal("guarded collision deleted source")
			}
			pixieCollide4EA080(source, target, 91, hooks)
		})
	}
}

func TestPixieCollide4EA080TargetDamageCachesDataAndClassAndOrdersEffects(t *testing.T) {
	oldData := &pixieCollideTestData4EA080{damage: -31}
	newData := &pixieCollideTestData4EA080{damage: 99}
	parent := &pixieCollideTestObject4EA080{name: "parent", class: pixiePlayerClass4EA080}
	source := &pixieCollideTestObject4EA080{name: "source", data: oldData, owner: parent}
	target := &pixieCollideTestObject4EA080{name: "target", class: pixiePlayerClass4EA080, damageVersion: 1}
	events := make([]string, 0, 13)
	hooks := defaultPixieCollideHooks4EA080()
	hooks.loadCollideData = func(got *pixieCollideTestObject4EA080) *pixieCollideTestData4EA080 {
		events = append(events, "data")
		if got != source {
			t.Fatalf("data source = %p", got)
		}
		return got.data
	}
	hooks.isEnemy = func(first, second *pixieCollideTestObject4EA080) int32 {
		events = append(events, "enemy")
		if first != source || second != target {
			t.Fatalf("enemy args = %p/%p", first, second)
		}
		source.data = newData
		return 1
	}
	hooks.loadClass = func(obj *pixieCollideTestObject4EA080) uint32 {
		events = append(events, "class:"+obj.name)
		return obj.class
	}
	hooks.loadFlags = func(obj *pixieCollideTestObject4EA080) uint32 {
		events = append(events, "flags:"+obj.name)
		return obj.flags
	}
	hooks.loadOwner = func(obj *pixieCollideTestObject4EA080) *pixieCollideTestObject4EA080 {
		events = append(events, "owner")
		return obj.owner
	}
	hooks.checkInversion = func(gotTarget, gotSource *pixieCollideTestObject4EA080) int32 {
		events = append(events, "inversion")
		if gotTarget != target || gotSource != source {
			t.Fatalf("inversion args = %p/%p", gotTarget, gotSource)
		}
		target.class = 2
		return 0
	}
	hooks.hasEnchant = func(got *pixieCollideTestObject4EA080, enchant uint32) int32 {
		events = append(events, "enchant")
		if got != target || enchant != pixieReflectEnchant4EA080 {
			t.Fatalf("enchant args = %p/%d", got, enchant)
		}
		return 0
	}
	hooks.loadDamage = func(data *pixieCollideTestData4EA080) int32 {
		events = append(events, "damage")
		if data != oldData {
			t.Fatalf("data = %p, want cached %p", data, oldData)
		}
		return data.damage
	}
	hooks.findParent = func(got *pixieCollideTestObject4EA080) *pixieCollideTestObject4EA080 {
		events = append(events, "parent")
		if got != source {
			t.Fatalf("parent source = %p", got)
		}
		target.damageVersion = 2
		return parent
	}
	hooks.targetDamage = func(gotTarget, gotParent, gotSource *pixieCollideTestObject4EA080, damage int32, damageType uint32) int32 {
		events = append(events, "target-damage")
		if gotTarget != target || gotParent != parent || gotSource != source || damage != -31 || damageType != pixieDamageType4EA080 {
			t.Fatalf("damage args = %p/%p/%p/%d/%d", gotTarget, gotParent, gotSource, damage, damageType)
		}
		if target.damageVersion != 2 {
			t.Fatal("Damage callback was not observed after parent lookup")
		}
		return math.MinInt32
	}
	hooks.audio = func(id uint32, obj *pixieCollideTestObject4EA080) {
		events = append(events, "audio")
		if id != pixieDamageAudio4EA080 || obj != source {
			t.Fatalf("audio = %d/%p", id, obj)
		}
	}
	hooks.delayedDelete = func(obj *pixieCollideTestObject4EA080) {
		events = append(events, "delete")
		if obj != source {
			t.Fatalf("delete = %p", obj)
		}
	}

	pixieCollide4EA080(source, target, 77, hooks)
	want := []string{
		"data", "enemy", "class:target", "flags:target", "owner", "class:parent", "flags:parent",
		"inversion", "enchant", "damage", "parent", "target-damage", "audio", "delete",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestPixieCollide4EA080OwnerSuppressionIsSilent(t *testing.T) {
	owner := &pixieCollideTestObject4EA080{name: "owner", class: pixiePlayerClass4EA080, flags: pixieOwnerSuppressFlag4EA080}
	source := &pixieCollideTestObject4EA080{name: "source", data: &pixieCollideTestData4EA080{}, owner: owner}
	target := &pixieCollideTestObject4EA080{name: "target", class: pixiePlayerClass4EA080}
	hooks := defaultPixieCollideHooks4EA080()
	hooks.checkInversion = func(*pixieCollideTestObject4EA080, *pixieCollideTestObject4EA080) int32 {
		t.Fatal("suppressed collision checked inversion")
		return 0
	}
	hooks.loadDamage = func(*pixieCollideTestData4EA080) int32 {
		t.Fatal("suppressed collision loaded damage")
		return 0
	}
	hooks.delayedDelete = func(*pixieCollideTestObject4EA080) {
		t.Fatal("suppressed collision deleted source")
	}
	pixieCollide4EA080(source, target, 0, hooks)
}

func TestPixieCollide4EA080PlayerInversionChangesOwnerOnly(t *testing.T) {
	source := &pixieCollideTestObject4EA080{data: &pixieCollideTestData4EA080{}}
	target := &pixieCollideTestObject4EA080{class: pixiePlayerClass4EA080}
	hooks := defaultPixieCollideHooks4EA080()
	events := make([]string, 0, 2)
	hooks.checkInversion = func(gotTarget, gotSource *pixieCollideTestObject4EA080) int32 {
		events = append(events, "inversion")
		if gotTarget != target || gotSource != source {
			t.Fatalf("inversion = %p/%p", gotTarget, gotSource)
		}
		return 1
	}
	hooks.changeOwner = func(gotSource, gotTarget *pixieCollideTestObject4EA080) {
		events = append(events, "owner")
		if gotSource != source || gotTarget != target {
			t.Fatalf("owner = %p/%p", gotSource, gotTarget)
		}
	}
	hooks.hasEnchant = func(*pixieCollideTestObject4EA080, uint32) int32 {
		t.Fatal("inversion path checked enchant")
		return 0
	}
	hooks.delayedDelete = func(*pixieCollideTestObject4EA080) { t.Fatal("inversion path deleted source") }
	pixieCollide4EA080(source, target, 0, hooks)
	if !reflect.DeepEqual(events, []string{"inversion", "owner"}) {
		t.Fatalf("events = %v", events)
	}
}

func TestPixieCollide4EA080ReflectiveShieldOrderAndArguments(t *testing.T) {
	source := &pixieCollideTestObject4EA080{data: &pixieCollideTestData4EA080{}}
	target := &pixieCollideTestObject4EA080{class: pixiePlayerClass4EA080, direction1: -123}
	hooks := defaultPixieCollideHooks4EA080()
	events := make([]string, 0, 6)
	hooks.checkInversion = func(*pixieCollideTestObject4EA080, *pixieCollideTestObject4EA080) int32 {
		events = append(events, "inversion")
		return 0
	}
	hooks.hasEnchant = func(got *pixieCollideTestObject4EA080, enchant uint32) int32 {
		events = append(events, "enchant")
		if got != target || enchant != 27 {
			t.Fatalf("enchant = %p/%d", got, enchant)
		}
		return 1
	}
	hooks.loadDirection = func(got *pixieCollideTestObject4EA080) int16 {
		events = append(events, "direction")
		return got.direction1
	}
	hooks.checkDirection = func(gotTarget *pixieCollideTestObject4EA080, direction int16, gotSource *pixieCollideTestObject4EA080) int32 {
		events = append(events, "check")
		if gotTarget != target || direction != -123 || gotSource != source {
			t.Fatalf("direction check = %p/%d/%p", gotTarget, direction, gotSource)
		}
		return 3
	}
	hooks.changeOwner = func(gotSource, gotTarget *pixieCollideTestObject4EA080) {
		events = append(events, "owner")
		if gotSource != source || gotTarget != target {
			t.Fatalf("owner = %p/%p", gotSource, gotTarget)
		}
	}
	hooks.audio = func(id uint32, got *pixieCollideTestObject4EA080) {
		events = append(events, "audio")
		if id != pixieReflectAudio4EA080 || got != target {
			t.Fatalf("audio = %d/%p", id, got)
		}
	}
	hooks.loadDamage = func(*pixieCollideTestData4EA080) int32 {
		t.Fatal("reflected collision loaded damage")
		return 0
	}
	hooks.delayedDelete = func(*pixieCollideTestObject4EA080) { t.Fatal("reflected collision deleted source") }
	pixieCollide4EA080(source, target, 0, hooks)
	want := []string{"inversion", "enchant", "direction", "check", "owner", "audio"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestPixieCollide4EA080WallOrderExtendedYAndLiveX(t *testing.T) {
	oldData := &pixieCollideTestData4EA080{damage: -47}
	newData := &pixieCollideTestData4EA080{damage: 88}
	source := &pixieCollideTestObject4EA080{name: "source", data: oldData}
	const collision = 73
	events := make([]string, 0, 16)
	inputs := make([]uint32, 0, 2)
	results := []int32{-17, 29}
	hooks := defaultPixieCollideHooks4EA080()
	hooks.loadCollideData = func(obj *pixieCollideTestObject4EA080) *pixieCollideTestData4EA080 {
		events = append(events, "data")
		return obj.data
	}
	hooks.wallReflect = func(gotCollision int, gotSource *pixieCollideTestObject4EA080) {
		events = append(events, "reflect")
		if gotCollision != collision || gotSource != source {
			t.Fatalf("reflect = %d/%p", gotCollision, gotSource)
		}
		source.velocityX = 1
		source.newX = 2
		source.velocityY = math.Float32frombits(0xc09f9db2)
		source.newY = math.Float32frombits(0x3ce56042)
		source.data = newData
	}
	hooks.vectorDirection = func(got *pixieCollideTestObject4EA080) int32 {
		events = append(events, "vector")
		if got != source || got.velocityX != 1 || got.velocityY != math.Float32frombits(0xc09f9db2) {
			t.Fatalf("vector source = %#v", got)
		}
		return 0x12345
	}
	hooks.loadVelocityX = func(obj *pixieCollideTestObject4EA080) float32 {
		events = append(events, "vel-x")
		return obj.velocityX
	}
	hooks.loadNewPosX = func(obj *pixieCollideTestObject4EA080) float32 { events = append(events, "new-x"); return obj.newX }
	hooks.storeDirection2 = func(obj *pixieCollideTestObject4EA080, value uint16) {
		events = append(events, "direction2")
		obj.direction2 = value
	}
	hooks.storeNewPosX = func(obj *pixieCollideTestObject4EA080, value float32) {
		events = append(events, "store-x")
		obj.newX = value
	}
	hooks.loadVelocityY = func(obj *pixieCollideTestObject4EA080) float32 {
		events = append(events, "vel-y")
		return obj.velocityY
	}
	hooks.loadNewPosY = func(obj *pixieCollideTestObject4EA080) float32 { events = append(events, "new-y"); return obj.newY }
	hooks.storeNewPosY = func(obj *pixieCollideTestObject4EA080, value float32) {
		events = append(events, "store-y")
		obj.newY = value
	}
	hooks.loadDamage = func(data *pixieCollideTestData4EA080) int32 {
		events = append(events, "damage")
		if data != oldData {
			t.Fatalf("damage data = %p, want %p", data, oldData)
		}
		source.newY = 1000
		return data.damage
	}
	hooks.floatToInt = func(value float32) int32 {
		events = append(events, "float")
		inputs = append(inputs, math.Float32bits(value))
		if len(inputs) == 1 {
			source.newX = 92
		}
		result := results[0]
		results = results[1:]
		return result
	}
	hooks.damageMap = func(x, y, damage int32, damageType uint32, got *pixieCollideTestObject4EA080) {
		events = append(events, "map")
		if x != 29 || y != -17 || damage != -47 || damageType != pixieDamageType4EA080 || got != source {
			t.Fatalf("map = %d/%d/%d/%d/%p", x, y, damage, damageType, got)
		}
	}
	hooks.delayedDelete = func(*pixieCollideTestObject4EA080) { t.Fatal("wall path deleted source") }

	pixieCollide4EA080(source, nil, collision, hooks)
	wantEvents := []string{
		"data", "reflect", "vector", "vel-x", "new-x", "direction2", "store-x",
		"vel-y", "new-y", "store-y", "damage", "float", "new-x", "float", "map",
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
	if source.direction2 != 0x2345 || source.newX != 92 {
		t.Fatalf("direction/new X = %#x/%g", source.direction2, source.newX)
	}
	gridInverse := float64(math.Float32frombits(pixieGridInverseBits4EA080))
	wantInputs := []uint32{
		0xbe5cd3ec,
		math.Float32bits(float32(float64(float32(92)) * gridInverse)),
	}
	if !reflect.DeepEqual(inputs, wantInputs) {
		t.Fatalf("conversion inputs = %#v, want %#v", inputs, wantInputs)
	}
	spilledY := float32(source.velocityY + math.Float32frombits(0x3ce56042))
	if math.Float32bits(float32(float64(spilledY)*gridInverse)) != 0xbe5cd3ed {
		t.Fatal("test vector no longer distinguishes the original unspilled Y sum")
	}
}

func TestPixieCollide4EA080NoTargetOrWallDeletesAfterDataCache(t *testing.T) {
	source := &pixieCollideTestObject4EA080{data: &pixieCollideTestData4EA080{}}
	events := make([]string, 0, 2)
	hooks := defaultPixieCollideHooks4EA080()
	hooks.loadCollideData = func(obj *pixieCollideTestObject4EA080) *pixieCollideTestData4EA080 {
		events = append(events, "data")
		return obj.data
	}
	hooks.delayedDelete = func(got *pixieCollideTestObject4EA080) {
		events = append(events, "delete")
		if got != source {
			t.Fatalf("delete = %p", got)
		}
	}
	pixieCollide4EA080(source, nil, 0, hooks)
	if !reflect.DeepEqual(events, []string{"data", "delete"}) {
		t.Fatalf("events = %v", events)
	}
}
