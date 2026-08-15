package server

import (
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestUndeadKillerCollide4EBD40NativeLayout(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantClass := uintptr(8)
	wantSubclass := uintptr(12)
	wantHealth := uintptr(556)
	wantCollideData := uintptr(700)
	wantDamage := uintptr(716)
	wantDataSize := uintptr(4)
	wantDurSpellSize := uintptr(120)
	wantSpellField72 := uintptr(72)
	wantSpellFlags88 := uintptr(88)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantClass = 12
		wantSubclass = 16
		wantHealth = 616
		wantCollideData = 776
		wantDamage = 808
		wantDataSize = 8
		wantDurSpellSize = 184
		wantSpellField72 = 100
		wantSpellFlags88 = 120
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantClass},
		{"Object.ObjSubClass", unsafe.Offsetof(Object{}.ObjSubClass), wantSubclass},
		{"Object.HealthData", unsafe.Offsetof(Object{}.HealthData), wantHealth},
		{"Object.CollideData", unsafe.Offsetof(Object{}.CollideData), wantCollideData},
		{"Object.Damage", unsafe.Offsetof(Object{}.Damage), wantDamage},
		{"HealthData size", unsafe.Sizeof(HealthData{}), 20},
		{"HealthData.Cur", unsafe.Offsetof(HealthData{}.Cur), 0},
		{"UndeadKillerCollideData size", unsafe.Sizeof(UndeadKillerCollideData{}), wantDataSize},
		{"UndeadKillerCollideData.Spell", unsafe.Offsetof(UndeadKillerCollideData{}.Spell), 0},
		{"DurSpell size", unsafe.Sizeof(DurSpell{}), wantDurSpellSize},
		{"DurSpell.Field72", unsafe.Offsetof(DurSpell{}.Field72), wantSpellField72},
		{"DurSpell.Flags88", unsafe.Offsetof(DurSpell{}.Flags88), wantSpellFlags88},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func defaultUndeadKillerNativeDeps4EBD40() undeadKillerCollideNativeDeps4EBD40 {
	return undeadKillerCollideNativeDeps4EBD40{
		findParentPlayer: func(*Object) *Object { return nil },
		callTargetDamage: func(
			unsafe.Pointer,
			*Object, *Object, *Object,
			int32,
			object.DamageType,
		) int32 {
			return 0
		},
		delayedDelete: func(*Object) {},
	}
}

func TestUndeadKillerCollideNative4EBD40BindsPartialBudgetFields(t *testing.T) {
	spell := &DurSpell{Field72: 10, Flags88: 0x12345678}
	data := &UndeadKillerCollideData{Spell: spell}
	source := &Object{CollideData: unsafe.Pointer(data)}
	parent := &Object{}
	oldDamage := 41
	newDamage := 42
	target := &Object{
		ObjClass:    object.ClassMonster,
		ObjSubClass: object.SubClass(object.MonsterUndead),
		HealthData:  &HealthData{Cur: 4, Max: 99},
		Damage:      unsafe.Pointer(&oldDamage),
	}
	collision := &types.Pointf{X: math.Float32frombits(0x7fc12345), Y: math.Float32frombits(0x80000000)}
	events := make([]string, 0, 4)
	deps := defaultUndeadKillerNativeDeps4EBD40()
	deps.findParentPlayer = func(got *Object) *Object {
		events = append(events, "parent")
		if got != source {
			t.Fatalf("parent source = %p, want %p", got, source)
		}
		target.Damage = unsafe.Pointer(&newDamage)
		return parent
	}
	deps.callTargetDamage = func(
		fn unsafe.Pointer,
		gotTarget, gotParent, gotSource *Object,
		damage int32,
		damageType object.DamageType,
	) int32 {
		events = append(events, "damage")
		if fn != unsafe.Pointer(&newDamage) || gotTarget != target || gotParent != parent || gotSource != source ||
			damage != 4 || damageType != object.DamageType(undeadKillerDamageType4EBD40) {
			t.Fatalf("damage args = %p/%p/%p/%p/%d/%d", fn, gotTarget, gotParent, gotSource, damage, damageType)
		}
		spell.Field72 = 777
		return math.MinInt32
	}
	deps.delayedDelete = func(*Object) {
		t.Fatal("partial budget must not delete source")
	}

	undeadKillerCollideNative4EBD40(source, target, collision, deps)
	if spell.Field72 != 6 || spell.Flags88 != 0x12345678 {
		t.Fatalf("spell fields = %d/%#x, want 6/%#x", spell.Field72, spell.Flags88, uint32(0x12345678))
	}
	if target.HealthData.Cur != 4 || target.HealthData.Max != 99 || target.Damage != unsafe.Pointer(&newDamage) {
		t.Fatalf("target fields changed unexpectedly: %#v", target)
	}
	if math.Float32bits(collision.X) != 0x7fc12345 || math.Float32bits(collision.Y) != 0x80000000 {
		t.Fatalf("collision changed: %#v", *collision)
	}
	if !reflect.DeepEqual(events, []string{"parent", "damage"}) {
		t.Fatalf("events = %v", events)
	}
}

func TestUndeadKillerCollideNative4EBD40ConsumedBudgetReloadsAfterDelete(t *testing.T) {
	spell := &DurSpell{Field72: -3, Flags88: 0xabcdef01}
	data := &UndeadKillerCollideData{Spell: spell}
	source := &Object{CollideData: unsafe.Pointer(data)}
	target := &Object{
		ObjClass:    object.ClassMonster,
		ObjSubClass: object.SubClass(object.MonsterUndead),
		HealthData:  &HealthData{Cur: 1},
	}
	events := make([]string, 0, 3)
	deps := defaultUndeadKillerNativeDeps4EBD40()
	deps.callTargetDamage = func(
		_ unsafe.Pointer,
		gotTarget, _, gotSource *Object,
		damage int32,
		damageType object.DamageType,
	) int32 {
		events = append(events, "damage")
		if gotTarget != target || gotSource != source || damage != -3 || damageType != object.DamageType(6) {
			t.Fatalf("damage args = %p/%p/%d/%d", gotTarget, gotSource, damage, damageType)
		}
		spell.Field72 = 20
		return 0
	}
	deps.delayedDelete = func(got *Object) {
		events = append(events, "delete")
		if got != source {
			t.Fatalf("deleted = %p, want %p", got, source)
		}
		spell.Field72 = math.MaxInt32
	}

	undeadKillerCollideNative4EBD40(source, target, nil, deps)
	if spell.Field72 != math.MinInt32+2 {
		t.Fatalf("remaining = %d, want %d", spell.Field72, int32(math.MinInt32+2))
	}
	if spell.Flags88 != 0xabcdef01 {
		t.Fatalf("flags = %#x, want %#x", spell.Flags88, uint32(0xabcdef01))
	}
	if !reflect.DeepEqual(events, []string{"damage", "delete"}) {
		t.Fatalf("events = %v", events)
	}
}

func TestUndeadKillerCollideNative4EBD40NilHealthUsesZeroHP(t *testing.T) {
	spell := &DurSpell{Field72: 9}
	source := &Object{CollideData: unsafe.Pointer(&UndeadKillerCollideData{Spell: spell})}
	target := &Object{
		ObjClass:    object.ClassMonster,
		ObjSubClass: object.SubClass(object.MonsterUndead),
	}
	deps := defaultUndeadKillerNativeDeps4EBD40()
	called := false
	deps.callTargetDamage = func(
		_ unsafe.Pointer,
		gotTarget, _, gotSource *Object,
		damage int32,
		damageType object.DamageType,
	) int32 {
		called = true
		if gotTarget != target || gotSource != source || damage != 0 || damageType != object.DamageType(6) {
			t.Fatalf("damage args = %p/%p/%d/%d", gotTarget, gotSource, damage, damageType)
		}
		return 0
	}
	undeadKillerCollideNative4EBD40(source, target, nil, deps)
	if !called || spell.Field72 != 9 {
		t.Fatalf("called/remaining = %v/%d, want true/9", called, spell.Field72)
	}
}

func TestUndeadKillerCollideNative4EBD40NilTargetUsesCollisionPointerOnly(t *testing.T) {
	deps := defaultUndeadKillerNativeDeps4EBD40()
	deleted := 0
	deps.delayedDelete = func(got *Object) {
		deleted++
		if got != nil {
			t.Fatalf("deleted = %p, want nil", got)
		}
	}
	undeadKillerCollideNative4EBD40(nil, nil, nil, deps)
	if deleted != 1 {
		t.Fatalf("delete count = %d, want 1", deleted)
	}
	undeadKillerCollideNative4EBD40(nil, nil, &types.Pointf{}, deps)
	if deleted != 1 {
		t.Fatalf("delete count after non-nil collision = %d, want 1", deleted)
	}
}
