package server

import (
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestCollisionEnchantNative4FDF90PreservesPointersFieldsAndLiveDamageSlot(t *testing.T) {
	source := &Object{
		Buffs: uint32(1) << ENCHANT_SHOCK,
	}
	source.BuffsPower[ENCHANT_SHOCK] = 7
	target := &Object{
		ObjClass: object.ClassMonster,
	}
	initialDamage := new(byte)
	liveDamage := new(byte)
	target.Damage = unsafe.Pointer(initialDamage)

	var events []string
	collisionEnchantNative4FDF90(source, target, collisionEnchantNativeDeps4FDF90{
		isEnemy: func(gotSource, gotTarget *Object) bool {
			if gotSource != source || gotTarget != target {
				t.Fatalf("enemy args = %p/%p, want %p/%p", gotSource, gotTarget, source, target)
			}
			events = append(events, "enemy")
			source.BuffsPower[ENCHANT_SHOCK] = 0
			return true
		},
		audio: func(id int32, gotSource *Object, kind int32, code uint32) {
			if id != collisionEnchantShockAudio4FDF90 || gotSource != source || kind != 0 || code != 0 {
				t.Fatalf("audio args = %d/%p/%d/%d", id, gotSource, kind, code)
			}
			events = append(events, "audio")
		},
		disableEnchant: func(gotSource *Object, enchant EnchantID) {
			if gotSource != source {
				t.Fatalf("disable source = %p, want %p", gotSource, source)
			}
			events = append(events, "disable:"+enchant.String())
			switch enchant {
			case ENCHANT_SHOCK:
				target.Damage = unsafe.Pointer(liveDamage)
			case ENCHANT_INVISIBLE:
				if source.ObjClass == 0 {
					source.ObjClass = object.ClassPlayer
					target.ObjClass = object.Class(collisionEnchantWallClass4FDF90)
				}
			default:
				t.Fatalf("unexpected enchant %v", enchant)
			}
		},
		balanceFloatTable: func(key string, index int32) float64 {
			if key != collisionEnchantShockDamageBalance4FDF90 || index != -1 {
				t.Fatalf("balance args = %q/%d, want %q/-1", key, index, collisionEnchantShockDamageBalance4FDF90)
			}
			events = append(events, "balance")
			return 16_777_217
		},
		floatToInt: func(value float32) int32 {
			if got := math.Float32bits(value); got != math.Float32bits(16_777_216) {
				t.Fatalf("damage float bits = %08x, want binary32 spill", got)
			}
			events = append(events, "float-to-int")
			return math.MinInt32
		},
		callDamage: func(
			fnc unsafe.Pointer,
			gotTarget, gotSource, gotWeapon *Object,
			damage int32,
			damageType uint32,
		) int32 {
			if fnc != unsafe.Pointer(liveDamage) {
				t.Fatalf("damage slot = %p, want live %p (initial %p)", fnc, liveDamage, initialDamage)
			}
			if gotTarget != target || gotSource != source || gotWeapon != source ||
				damage != math.MinInt32 || damageType != collisionEnchantShockDamageType4FDF90 {
				t.Fatalf("damage args = %p/%p/%p/%d/%d", gotTarget, gotSource, gotWeapon, damage, damageType)
			}
			events = append(events, "damage")
			target.ObjClass = object.Class(collisionEnchantUnitOrWallClass4FDF90)
			return math.MinInt32
		},
		unitsOnSameTeam: func(gotTarget, gotSource *Object) bool {
			if gotTarget != target || gotSource != source {
				t.Fatalf("team args = %p/%p, want %p/%p", gotTarget, gotSource, target, source)
			}
			events = append(events, "same-team")
			return false
		},
	})

	want := []string{
		"enemy", "audio", "disable:ENCHANT_SHOCK", "balance", "float-to-int", "damage",
		"same-team", "disable:ENCHANT_INVISIBLE", "disable:ENCHANT_INVISIBLE",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, pointer := range map[string]uintptr{
			"source": uintptr(unsafe.Pointer(source)),
			"target": uintptr(unsafe.Pointer(target)),
		} {
			if pointer <= math.MaxUint32 {
				t.Fatalf("%s pointer = %#x, want native address above 4 GiB", name, pointer)
			}
		}
	}
	runtime.KeepAlive(source)
	runtime.KeepAlive(target)
	runtime.KeepAlive(initialDamage)
	runtime.KeepAlive(liveDamage)
}

func TestCollisionEnchantNative4FDF90DamageDispatchHasNoNilGuard(t *testing.T) {
	source := &Object{Buffs: uint32(1) << ENCHANT_SHOCK}
	target := &Object{ObjClass: object.ClassMonster}

	defer func() {
		if recover() == nil {
			t.Fatal("nil target damage callback was guarded instead of faulting")
		}
	}()
	collisionEnchantNative4FDF90(source, target, collisionEnchantNativeDeps4FDF90{
		isEnemy:           func(*Object, *Object) bool { return true },
		audio:             func(int32, *Object, int32, uint32) {},
		disableEnchant:    func(*Object, EnchantID) {},
		balanceFloatTable: func(string, int32) float64 { return 1 },
		floatToInt:        playerCollideRound4E8460,
		callDamage:        collisionEnchantCallDamage4FDF90,
		unitsOnSameTeam:   func(*Object, *Object) bool { t.Fatal("continued after nil damage fault"); return false },
	})
}

func TestCollisionEnchantNative4FDF90LayoutFields(t *testing.T) {
	ptrSize := unsafe.Sizeof(uintptr(0))
	for _, field := range []struct {
		name       string
		got        uintptr
		want32     uintptr
		wantNative uintptr
	}{
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), 8, 12},
		{"Object.ObjFlags", unsafe.Offsetof(Object{}.ObjFlags), 16, 20},
		{"Object.Buffs", unsafe.Offsetof(Object{}.Buffs), 340, 344},
		{"Object.BuffsPower", unsafe.Offsetof(Object{}.BuffsPower), 408, 412},
		{"Object.Damage", unsafe.Offsetof(Object{}.Damage), 716, 808},
	} {
		want := field.wantNative
		if ptrSize == 4 {
			want = field.want32
		}
		if field.got != want {
			t.Errorf("%s offset = %d, want %d for %d-bit", field.name, field.got, want, ptrSize*8)
		}
	}
	if got := unsafe.Sizeof(Object{}.BuffsPower[0]); got != 1 {
		t.Fatalf("Shock power element size = %d, want 1", got)
	}
}
