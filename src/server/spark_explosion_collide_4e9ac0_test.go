package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type sparkExplosionTestData4E9AC0 struct {
	power uint8
}

func TestSparkExplosionCollide4E9AC0ReflectsThenRunsQuestGate(t *testing.T) {
	var events []string
	sparkExplosionCollide4E9AC0(1, 2, struct{ unread int }{unread: 99}, sparkExplosionCollideHooks4E9AC0[int, int]{
		loadCollideData: func(source int) int {
			events = append(events, fmt.Sprintf("data:%d", source))
			return 77
		},
		hasEnchant: func(target int, enchant uint32) int32 {
			events = append(events, fmt.Sprintf("enchant:%d:%d", target, enchant))
			return -1
		},
		loadDirection: func(target int) int16 {
			events = append(events, fmt.Sprintf("direction:%d", target))
			return -7
		},
		checkDirection: func(target int, direction int16, source int) int32 {
			events = append(events, fmt.Sprintf("check:%d:%d:%d", target, direction, source))
			return 3
		},
		reflect: func(source, target int) {
			events = append(events, fmt.Sprintf("reflect:%d:%d", source, target))
		},
		clearOwner: func(source int) {
			events = append(events, fmt.Sprintf("clear:%d", source))
		},
		setOwner: func(owner, source int) {
			events = append(events, fmt.Sprintf("owner:%d:%d", owner, source))
		},
		audio: func(id uint32, obj int, kind int32, code uint32) {
			events = append(events, fmt.Sprintf("audio:%d:%d:%d:%d", id, obj, kind, code))
		},
		gameFlagsCheck: func(flag uint32) int32 {
			events = append(events, fmt.Sprintf("flag:%#x", flag))
			return 1
		},
		findParent: func(source int) int {
			events = append(events, fmt.Sprintf("parent:%d", source))
			return 3
		},
		classLow: func(obj int) uint8 {
			events = append(events, fmt.Sprintf("class:%d", obj))
			return sparkExplosionPlayerClass4E9AC0
		},
		isEnemy: func(parent, target int) int32 {
			events = append(events, fmt.Sprintf("enemy:%d:%d", parent, target))
			return 0
		},
	})

	want := []string{
		"data:1",
		"enchant:2:27",
		"direction:2",
		"check:2:-7:1",
		"reflect:1:2",
		"clear:1",
		"owner:2:1",
		"audio:122:2:0:0",
		"flag:0x1000",
		"parent:1",
		"class:3",
		"class:2",
		"enemy:3:2",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
}

func TestSparkExplosionCollide4E9AC0ReloadsCachedPowerInOriginalOrder(t *testing.T) {
	data := &sparkExplosionTestData4E9AC0{power: 9}
	var events []string
	sparkExplosionCollide4E9AC0(1, 2, nil, sparkExplosionCollideHooks4E9AC0[int, *sparkExplosionTestData4E9AC0]{
		loadCollideData: func(source int) *sparkExplosionTestData4E9AC0 {
			events = append(events, "data")
			return data
		},
		loadPower: func(got *sparkExplosionTestData4E9AC0) uint8 {
			if got != data {
				t.Fatalf("data = %p, want cached %p", got, data)
			}
			events = append(events, fmt.Sprintf("power:%d", got.power))
			return got.power
		},
		hasEnchant: func(target int, enchant uint32) int32 {
			events = append(events, "enchant")
			return 0
		},
		gameFlagsCheck: func(flag uint32) int32 {
			events = append(events, fmt.Sprintf("flag:%#x", flag))
			if flag == sparkExplosionCoopFlag4E9AC0 {
				return 2
			}
			return 0
		},
		mapPushUnits: func(pos int, first, second, force float32, source int, arg6, arg7 int32) {
			events = append(events, "push")
			if pos != 1 || source != 1 || first != 6 || second != 0 ||
				force != math.Float32frombits(sparkExplosionPushForceBits4E9AC0) || arg6 != 0 || arg7 != 0 {
				t.Fatalf("push = (%d,%g,%g,%g,%d,%d,%d)", pos, first, second, force, source, arg6, arg7)
			}
			data.power = 10
		},
		findParent: func(source int) int {
			events = append(events, "parent")
			return 3
		},
		targetDamage: func(target, parent, source int, damage int32, damageType uint32) int32 {
			events = append(events, "damage")
			if target != 2 || parent != 3 || source != 1 || damage != 5 || damageType != sparkExplosionDamageType4E9AC0 {
				t.Fatalf("damage = (%d,%d,%d,%d,%d)", target, parent, source, damage, damageType)
			}
			data.power = 12
			return -9
		},
		mapDamageUnits: func(pos int, radius, inner float32, damage int32, damageType uint32, source, excluded int) {
			events = append(events, "area")
			if pos != 1 || radius != 4 || inner != 0 || damage != 6 ||
				damageType != sparkExplosionDamageType4E9AC0 || source != 1 || excluded != 2 {
				t.Fatalf("area = (%d,%g,%g,%d,%d,%d,%d)", pos, radius, inner, damage, damageType, source, excluded)
			}
			data.power = 14
		},
		sparkFX: func(pos int, power uint8) {
			events = append(events, fmt.Sprintf("fx:%d:%d", pos, power))
		},
		audio: func(id uint32, obj int, kind int32, code uint32) {
			events = append(events, fmt.Sprintf("audio:%d:%d", id, obj))
		},
		scorch: func(pos int, kind int32) {
			events = append(events, fmt.Sprintf("scorch:%d:%d", pos, kind))
		},
		delayedDelete: func(source int) {
			events = append(events, fmt.Sprintf("delete:%d", source))
		},
	})

	want := []string{
		"data", "enchant", "flag:0x1000", "power:9", "push",
		"power:10", "parent", "damage", "flag:0x800", "power:12",
		"area", "power:14", "fx:1:14", "audio:42:1", "scorch:1:2", "delete:1",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
}

func TestSparkExplosionCollide4E9AC0NilTargetUsesNonCoopInnerRadius(t *testing.T) {
	data := &sparkExplosionTestData4E9AC0{power: 255}
	var gotInner float32
	var gotDamage int32
	var gotExcluded int
	sparkExplosionCollide4E9AC0(1, 0, 123, sparkExplosionCollideHooks4E9AC0[int, *sparkExplosionTestData4E9AC0]{
		loadCollideData: func(int) *sparkExplosionTestData4E9AC0 { return data },
		loadPower:       func(data *sparkExplosionTestData4E9AC0) uint8 { return data.power },
		gameFlagsCheck:  func(uint32) int32 { return 0 },
		mapPushUnits:    func(int, float32, float32, float32, int, int32, int32) {},
		mapDamageUnits: func(_ int, _ float32, inner float32, damage int32, _ uint32, _, excluded int) {
			gotInner = inner
			gotDamage = damage
			gotExcluded = excluded
		},
		sparkFX:       func(int, uint8) {},
		audio:         func(uint32, int, int32, uint32) {},
		scorch:        func(int, int32) {},
		delayedDelete: func(int) {},
	})

	if gotInner != math.Float32frombits(sparkExplosionInnerRadiusBits4E9AC0) || gotDamage != 127 || gotExcluded != 0 {
		t.Fatalf("area inner/damage/excluded = %g/%d/%d, want 15/127/0", gotInner, gotDamage, gotExcluded)
	}
}

func TestSparkExplosionCollide4E9AC0DirectionRequiresLowBit(t *testing.T) {
	data := &sparkExplosionTestData4E9AC0{power: 3}
	reflected := false
	sparkExplosionCollide4E9AC0(1, 2, nil, sparkExplosionCollideHooks4E9AC0[int, *sparkExplosionTestData4E9AC0]{
		loadCollideData: func(int) *sparkExplosionTestData4E9AC0 { return data },
		loadPower:       func(data *sparkExplosionTestData4E9AC0) uint8 { return data.power },
		hasEnchant:      func(int, uint32) int32 { return 1 },
		loadDirection:   func(int) int16 { return 17 },
		checkDirection:  func(int, int16, int) int32 { return 2 },
		reflect:         func(int, int) { reflected = true },
		gameFlagsCheck:  func(uint32) int32 { return 0 },
		mapPushUnits:    func(int, float32, float32, float32, int, int32, int32) {},
		findParent:      func(int) int { return 3 },
		targetDamage:    func(int, int, int, int32, uint32) int32 { return 0 },
		mapDamageUnits:  func(int, float32, float32, int32, uint32, int, int) {},
		sparkFX:         func(int, uint8) {},
		audio:           func(uint32, int, int32, uint32) {},
		scorch:          func(int, int32) {},
		delayedDelete:   func(int) {},
	})
	if reflected {
		t.Fatal("direction result 2 reflected; GAME.EXE tests only bit zero")
	}
}
