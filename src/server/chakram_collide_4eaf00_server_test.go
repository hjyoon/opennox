package server

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/ntype"
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

func TestChakramCollideServerDeps4EAF00NativeRuntimeBoundary(t *testing.T) {
	weapon := &Object{}
	playerUpdate := &PlayerUpdateData{EquippedWeapon: weapon}
	owner := &Object{UpdateData: unsafe.Pointer(playerUpdate)}
	trace := &ntype.Point32{X: -123, Y: 456}

	var (
		mapX, mapY, mapDamage int32
		mapType               object.DamageType
		mapSource             *Object
		createdItem           *Object
		createdOwner          *Object
		createdPos            types.Pointf
		droppedOwner          *Object
		droppedItem           *Object
		droppedPos            *types.Pointf
	)
	runtime := ChakramCollideRuntime4EAF00{
		TraceHitPoint: func() *ntype.Point32 { return trace },
		DamageMap: func(x, y, damage int32, typ object.DamageType, source *Object) {
			mapX, mapY, mapDamage, mapType, mapSource = x, y, damage, typ, source
		},
		Drop: func(owner, item *Object, pos *types.Pointf) {
			droppedOwner, droppedItem, droppedPos = owner, item, pos
		},
		CreateAt: func(item, owner *Object, pos types.Pointf) {
			createdItem, createdOwner, createdPos = item, owner, pos
		},
	}
	deps := chakramCollideServerDeps4EAF00(&Server{}, runtime)

	if !deps.ownerHasWeapon(owner) {
		t.Fatal("native equipped-weapon pointer was not observed")
	}
	playerUpdate.EquippedWeapon = nil
	if deps.ownerHasWeapon(owner) {
		t.Fatal("cleared equipped-weapon pointer still reported present")
	}

	collision := &types.Pointf{X: 2, Y: 3}
	velocity := &types.Pointf{X: 4, Y: -5}
	deps.wallReflect(collision, velocity)
	if *velocity != (types.Pointf{X: 5, Y: -4}) {
		t.Fatalf("positive-product reflection = %+v, want {5 -4}", *velocity)
	}
	collision.X = 0
	deps.wallReflect(collision, velocity)
	if *velocity != (types.Pointf{X: -4, Y: 5}) {
		t.Fatalf("zero-product reflection = %+v, want {-4 5}", *velocity)
	}

	x, y, ok := deps.tracePoint()
	if !ok || x != trace.X || y != trace.Y {
		t.Fatalf("trace = (%d, %d, %t), want (%d, %d, true)", x, y, ok, trace.X, trace.Y)
	}
	source := &Object{}
	deps.damageMap(x, y, 7, 9, source)
	if mapX != trace.X || mapY != trace.Y || mapDamage != 7 || mapType != 9 || mapSource != source {
		t.Fatalf("map damage = (%d, %d, %d, %d, %p)", mapX, mapY, mapDamage, mapType, mapSource)
	}

	if got := deps.floatToInt(9.75); got != 9 {
		t.Fatalf("positive truncation = %d, want 9", got)
	}
	if got := deps.floatToInt(-9.75); got != -9 {
		t.Fatalf("negative truncation = %d, want -9", got)
	}
	item := &Object{}
	pos := &types.Pointf{X: 12.5, Y: -3.25}
	deps.drop(owner, item, pos)
	if droppedOwner != owner || droppedItem != item || droppedPos != pos {
		t.Fatalf("drop = (%p, %p, %p), want (%p, %p, %p)", droppedOwner, droppedItem, droppedPos, owner, item, pos)
	}
	deps.createAt(item, nil, pos)
	if createdItem != item || createdOwner != nil || createdPos != *pos {
		t.Fatalf("create = (%p, %p, %+v), want (%p, nil, %+v)",
			createdItem, createdOwner, createdPos, item, *pos)
	}
}
