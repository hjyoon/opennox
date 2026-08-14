package server

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
)

func TestArrowCollide4EB490NativeLayouts(t *testing.T) {
	ptrSize := unsafe.Sizeof(uintptr(0))
	type layoutCase struct {
		name string
		got  uintptr
		w32  uintptr
		w64  uintptr
	}
	cases := []layoutCase{
		{"collide size", unsafe.Sizeof(ArrowCollideData{}), 8, 16},
		{"collide field zero", unsafe.Offsetof(ArrowCollideData{}.Field0), 0, 0},
		{"collide owner", unsafe.Offsetof(ArrowCollideData{}.Owner), 4, 8},
		{"attack size", unsafe.Sizeof(ArrowAttackData{}), 32, 48},
		{"attack damage", unsafe.Offsetof(ArrowAttackData{}.Damage), 0, 0},
		{"attack damage type", unsafe.Offsetof(ArrowAttackData{}.DamageType), 4, 4},
		{"attack radius", unsafe.Offsetof(ArrowAttackData{}.Radius), 8, 8},
		{"attack owner", unsafe.Offsetof(ArrowAttackData{}.Owner), 12, 16},
		{"attack X", unsafe.Offsetof(ArrowAttackData{}.PosX), 16, 24},
		{"attack Y", unsafe.Offsetof(ArrowAttackData{}.PosY), 20, 28},
		{"attack field 24", unsafe.Offsetof(ArrowAttackData{}.Field24), 24, 32},
		{"attack source", unsafe.Offsetof(ArrowAttackData{}.Source), 28, 40},
	}
	for _, tc := range cases {
		want := tc.w64
		if ptrSize == 4 {
			want = tc.w32
		}
		if tc.got != want {
			t.Errorf("%s = %d, want %d on pointer size %d", tc.name, tc.got, want, ptrSize)
		}
	}
}

func TestArrowCollideNative4EB490UsesNativeDataAndOneAttackRecord(t *testing.T) {
	strengthOwner := &Object{}
	attackOwner := &Object{}
	liveAttackOwner := &Object{}
	data := &ArrowCollideData{Field0: 0x89abcdef, Owner: attackOwner}
	source := &Object{
		TypeInd: 7, ObjOwner: strengthOwner,
		CollideData: unsafe.Pointer(data),
		PosVec:      types.Pointf{X: 12.5, Y: -4.25},
	}
	source.Shape.Circle.R = 6.75
	target := &Object{}
	modifier := &Modifier{}

	var applyRecord, preRecord *ArrowAttackData
	var gotDamage int32
	arrowCollideNative4EB490(source, target, nil, arrowCollideNativeDeps4EB490{
		lookupProjectileClass: func(index uint16) *Modifier {
			if index != 7 {
				t.Fatalf("type index = %d, want 7", index)
			}
			return modifier
		},
		strength: func(owner *Object) int32 {
			if owner != strengthOwner {
				t.Fatalf("strength owner = %p, want %p", owner, strengthOwner)
			}
			return 37
		},
		gameFlag:         func(uint32) bool { return false },
		findParentPlayer: func(*Object) *Object { return strengthOwner },
		isEnemy:          func(*Object, *Object) bool { return false },
		calcBoltDamage: func(strength int32, got *Modifier) float64 {
			if strength != 37 || got != modifier {
				t.Fatalf("damage inputs = (%d, %p), want (37, %p)", strength, got, modifier)
			}
			return 7.25
		},
		floatToInt: func(value float64) int32 {
			if value != 9.75 {
				t.Fatalf("conversion input = %v, want 9.75", value)
			}
			return 9
		},
		loadArcherBoltType: func() uint32 { return 99 },
		lookupType: func(string) uint32 {
			t.Fatal("lookup called with populated ArcherBolt cache")
			return 0
		},
		storeArcherBoltType: func(uint32) { t.Fatal("ArcherBolt cache stored") },
		applyAttackEffect: func(gotSource, gotOwner *Object, attack *ArrowAttackData) {
			if gotSource != source || gotOwner != attackOwner {
				t.Fatalf("apply objects = (%p, %p)", gotSource, gotOwner)
			}
			applyRecord = attack
			data.Owner = liveAttackOwner
			attack.Damage = 8.75
		},
		preAttackEffects: func(gotTarget, gotOwner, gotSource *Object, attack *ArrowAttackData) {
			if gotTarget != target || gotOwner != liveAttackOwner || gotSource != source {
				t.Fatalf("pre objects = (%p, %p, %p)", gotTarget, gotOwner, gotSource)
			}
			preRecord = attack
			attack.Damage = 9.25
		},
		targetDamage: func(gotTarget, parent, gotSource *Object, damage int32, damageType uint32) int32 {
			if gotTarget != target || parent != strengthOwner || gotSource != source || damageType != 11 {
				t.Fatalf("damage objects/type = (%p, %p, %p, %d)", gotTarget, parent, gotSource, damageType)
			}
			gotDamage = damage
			return 0x100
		},
		delayedDelete: func(*Object) { t.Fatal("AL-zero ordinary Arrow was deleted") },
	})

	if applyRecord == nil || applyRecord != preRecord {
		t.Fatalf("effect records = (%p, %p), want one stable native pointer", applyRecord, preRecord)
	}
	want := ArrowAttackData{
		Damage: 9.25, DamageType: 11, Radius: 6.75, Owner: attackOwner,
		PosX: 12.5, PosY: -4.25, Source: source,
	}
	if *applyRecord != want {
		t.Fatalf("attack = %+v, want %+v", *applyRecord, want)
	}
	if gotDamage != 9 || data.Field0 != 0x89abcdef {
		t.Fatalf("damage/field = (%d, %#x), want (9, %#x)", gotDamage, data.Field0, uint32(0x89abcdef))
	}
}

func TestArrowTruncFloat64ToInt32_4EB490(t *testing.T) {
	tests := []struct {
		value float64
		want  int32
	}{
		{12.999, 12},
		{-12.999, -12},
		{0x1p32 + 7, 7},
		{-0x1p32 - 7, -7},
		{0x1p63 - 0x1p10, -1024},
		{-0x1p63, 0},
		{0x1p63, 0},
		{math.Inf(1), 0},
		{math.Inf(-1), 0},
		{math.NaN(), 0},
	}
	for _, tc := range tests {
		if got := arrowTruncFloat64ToInt32_4EB490(tc.value); got != tc.want {
			t.Errorf("truncate(%v) = %d (%#x), want %d (%#x)", tc.value, got, uint32(got), tc.want, uint32(tc.want))
		}
	}
}
