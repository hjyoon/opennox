package server

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
)

func TestChakramCollide4EAF00NativeLayouts(t *testing.T) {
	ptrSize := unsafe.Sizeof(uintptr(0))
	type layoutCase struct {
		name string
		got  uintptr
		w32  uintptr
		w64  uintptr
	}
	cases := []layoutCase{
		{"update size", unsafe.Sizeof(ChakramUpdateData{}), 28, 40},
		{"update reflections", unsafe.Offsetof(ChakramUpdateData{}.Reflections), 4, 4},
		{"update return target", unsafe.Offsetof(ChakramUpdateData{}.ReturnTarget), 8, 8},
		{"update last hit", unsafe.Offsetof(ChakramUpdateData{}.LastHit), 12, 16},
		{"update owner position", unsafe.Offsetof(ChakramUpdateData{}.OwnerPos), 16, 24},
		{"update return state", unsafe.Offsetof(ChakramUpdateData{}.ReturnState), 24, 32},
		{"attack size", unsafe.Sizeof(ChakramAttackData{}), 32, 48},
		{"attack damage", unsafe.Offsetof(ChakramAttackData{}.Damage), 0, 0},
		{"attack damage type", unsafe.Offsetof(ChakramAttackData{}.DamageType), 4, 4},
		{"attack radius", unsafe.Offsetof(ChakramAttackData{}.Radius), 8, 8},
		{"attack owner", unsafe.Offsetof(ChakramAttackData{}.Owner), 12, 16},
		{"attack X", unsafe.Offsetof(ChakramAttackData{}.PosX), 16, 24},
		{"attack Y", unsafe.Offsetof(ChakramAttackData{}.PosY), 20, 28},
		{"attack field24", unsafe.Offsetof(ChakramAttackData{}.Field24), 24, 32},
		{"attack source", unsafe.Offsetof(ChakramAttackData{}.Source), 28, 40},
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

func TestChakramCollideNative4EAF00WallUsesNativePointPointers(t *testing.T) {
	owner := &Object{}
	item := &Object{}
	update := &ChakramUpdateData{Reflections: 1}
	source := &Object{
		InvFirstItem: item,
		ObjOwner:     owner,
		PosVec:       types.Pointf{X: 10, Y: 20},
		VelVec:       types.Pointf{X: 3, Y: 4},
		UpdateData:   unsafe.Pointer(update),
	}
	collision := &types.Pointf{X: -1, Y: 2}
	var fxPos, reflectedVelocity *types.Pointf
	chakramCollideNative4EAF00(source, nil, collision, chakramCollideNativeDeps4EAF00{
		pointFX: func(id uint32, pos *types.Pointf) {
			if id != chakramImpactFX4EAF00 {
				t.Fatalf("FX = %d, want %d", id, chakramImpactFX4EAF00)
			}
			fxPos = pos
		},
		wallReflect: func(gotCollision, velocity *types.Pointf) {
			if gotCollision != collision {
				t.Fatalf("collision = %p, want %p", gotCollision, collision)
			}
			reflectedVelocity = velocity
		},
		tracePoint: func() (int32, int32, bool) { return 0, 0, false },
	})
	if fxPos != &source.PosVec || reflectedVelocity != &source.VelVec {
		t.Fatalf("native points = (%p, %p), want (%p, %p)", fxPos, reflectedVelocity, &source.PosVec, &source.VelVec)
	}
	if update.Reflections != 0 || update.ReturnTarget != owner || update.ReturnState != 0 {
		t.Fatalf("update = %+v, want exhausted reflections returning to owner", update)
	}
}

func TestChakramCollideNative4EAF00AttackKeepsOneNativeRecord(t *testing.T) {
	owner := &Object{}
	item := &Object{}
	modifier := &Modifier{}
	update := &ChakramUpdateData{Reflections: 1}
	source := &Object{
		TypeInd:      73,
		InvFirstItem: item,
		ObjOwner:     owner,
		PosVec:       types.Pointf{X: 12.5, Y: -4.25},
		UpdateData:   unsafe.Pointer(update),
	}
	source.Shape.Circle.R = 6
	target := &Object{}

	var applyRecord, preRecord *ChakramAttackData
	var gotDamage int32
	chakramCollideNative4EAF00(source, target, nil, chakramCollideNativeDeps4EAF00{
		sameTeam: func(first, second *Object) bool {
			if first != source || second != target {
				t.Fatalf("team args = (%p, %p)", first, second)
			}
			return false
		},
		lookupProjectileClass: func(index uint16) *Modifier {
			if index != source.TypeInd {
				t.Fatalf("type index = %d, want %d", index, source.TypeInd)
			}
			return modifier
		},
		strength: func(got *Object) int32 {
			if got != owner {
				t.Fatalf("strength owner = %p, want %p", got, owner)
			}
			return 37
		},
		calcBoltDamage: func(strength int32, got *Modifier) float32 {
			if strength != 37 || got != modifier {
				t.Fatalf("damage inputs = (%d, %p), want (37, %p)", strength, got, modifier)
			}
			return 7.25
		},
		applyAttackEffect: func(gotSource, gotOwner *Object, attack *ChakramAttackData) {
			if gotSource != source || gotOwner != owner {
				t.Fatalf("apply objects = (%p, %p)", gotSource, gotOwner)
			}
			applyRecord = attack
			attack.Damage = 8.75
		},
		preAttackEffects: func(gotTarget, gotOwner, gotSource *Object, attack *ChakramAttackData) {
			if gotTarget != target || gotOwner != owner || gotSource != source {
				t.Fatalf("pre objects = (%p, %p, %p)", gotTarget, gotOwner, gotSource)
			}
			preRecord = attack
			if attack.Damage != 8.75 {
				t.Fatalf("pre damage = %v, want apply-mutated 8.75", attack.Damage)
			}
			attack.Damage = 9.25
		},
		floatToInt: func(value float64) int32 {
			if value != 9.75 {
				t.Fatalf("round input = %v, want 9.75", value)
			}
			return 10
		},
		targetDamage: func(gotTarget, gotOwner, gotSource *Object, damage int32, typ uint32) {
			if gotTarget != target || gotOwner != owner || gotSource != source || typ != 0 {
				t.Fatalf("damage args = (%p, %p, %p, %d)", gotTarget, gotOwner, gotSource, typ)
			}
			gotDamage = damage
		},
		projectileReflect: func(gotSource, gotTarget *Object) {
			if gotSource != source || gotTarget != target {
				t.Fatalf("reflect args = (%p, %p)", gotSource, gotTarget)
			}
		},
	})
	if applyRecord == nil || applyRecord != preRecord {
		t.Fatalf("effect records = (%p, %p), want one stable pointer", applyRecord, preRecord)
	}
	wantRecord := ChakramAttackData{
		Damage: 9.25, Radius: 36, Owner: owner, PosX: 12.5, PosY: -4.25, Source: source,
	}
	if *applyRecord != wantRecord {
		t.Fatalf("attack = %+v, want %+v", *applyRecord, wantRecord)
	}
	if gotDamage != 10 || update.LastHit != target || update.ReturnTarget != owner || update.Reflections != 0 {
		t.Fatalf("result = (damage %d, update %+v)", gotDamage, update)
	}
}
