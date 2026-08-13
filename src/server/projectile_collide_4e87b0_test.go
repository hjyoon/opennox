package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type projectileCollideTestObject4E87B0 struct {
	name        string
	typeIndex   uint16
	collideData *projectileCollideTestData4E87B0
}

type projectileCollideTestData4E87B0 struct {
	damage int32
}

type projectileCollideTestPoint4E87B0 struct {
	x int32
	y int32
}

type projectileCollideTestState4E87B0 struct {
	t                 *testing.T
	events            []string
	throwingStoneType uint32
	impShotType       uint32
	lookup            map[string]uint32
	gameData          map[string]float64
	floatResult       int32
	floatInput        float32
	parent            *projectileCollideTestObject4E87B0
	damageResult      uint8
	point             *projectileCollideTestPoint4E87B0
	onLoadCollideData func()
	onLookup          func(string)
	onLoadPointY      func(*projectileCollideTestPoint4E87B0)
	gotDamageTarget   *projectileCollideTestObject4E87B0
	gotDamageSource   *projectileCollideTestObject4E87B0
	gotDamageAttacker *projectileCollideTestObject4E87B0
	gotDamage         int32
	gotDamageType     uint32
	gotMapX           int32
	gotMapY           int32
	gotMapDamage      int32
	gotMapDamageType  uint32
	gotMapProjectile  *projectileCollideTestObject4E87B0
	deleted           []*projectileCollideTestObject4E87B0
}

func (s *projectileCollideTestState4E87B0) hooks() projectileCollideHooks4E87B0[
	*projectileCollideTestObject4E87B0,
	*uint32,
	*projectileCollideTestPoint4E87B0,
	*projectileCollideTestData4E87B0,
] {
	return projectileCollideHooks4E87B0[
		*projectileCollideTestObject4E87B0,
		*uint32,
		*projectileCollideTestPoint4E87B0,
		*projectileCollideTestData4E87B0,
	]{
		loadCollideData: func(obj *projectileCollideTestObject4E87B0) *projectileCollideTestData4E87B0 {
			s.events = append(s.events, "collide-data")
			if s.onLoadCollideData != nil {
				s.onLoadCollideData()
			}
			return obj.collideData
		},
		loadThrowingStoneType: func() uint32 {
			s.events = append(s.events, "throw-cache")
			return s.throwingStoneType
		},
		lookupType: func(name string) uint32 {
			s.events = append(s.events, "lookup:"+name)
			if s.onLookup != nil {
				s.onLookup(name)
			}
			return s.lookup[name]
		},
		storeThrowingStone: func(value uint32) {
			s.events = append(s.events, fmt.Sprintf("store-throw:%d", value))
			s.throwingStoneType = value
		},
		storeImpShot: func(value uint32) {
			s.events = append(s.events, fmt.Sprintf("store-imp:%d", value))
			s.impShotType = value
		},
		loadType: func(obj *projectileCollideTestObject4E87B0) uint16 {
			s.events = append(s.events, "type")
			return obj.typeIndex
		},
		loadImpShotType: func() uint32 {
			s.events = append(s.events, "imp-cache")
			return s.impShotType
		},
		gameDataFloat: func(name string) float64 {
			s.events = append(s.events, "balance:"+name)
			return s.gameData[name]
		},
		floatToInt: func(value float32) int32 {
			s.events = append(s.events, "float-to-int")
			s.floatInput = value
			return s.floatResult
		},
		loadDamage: func(data *projectileCollideTestData4E87B0) int32 {
			s.events = append(s.events, "damage-data")
			return data.damage
		},
		findParentPlayer: func(obj *projectileCollideTestObject4E87B0) *projectileCollideTestObject4E87B0 {
			s.events = append(s.events, "parent")
			return s.parent
		},
		damage: func(target, source, attacker *projectileCollideTestObject4E87B0, damage int32, damageType uint32) uint8 {
			s.events = append(s.events, "damage-callback")
			s.gotDamageTarget = target
			s.gotDamageSource = source
			s.gotDamageAttacker = attacker
			s.gotDamage = damage
			s.gotDamageType = damageType
			return s.damageResult
		},
		traceHitPoint: func() *projectileCollideTestPoint4E87B0 {
			s.events = append(s.events, "trace-point")
			return s.point
		},
		loadPointY: func(point *projectileCollideTestPoint4E87B0) int32 {
			s.events = append(s.events, "point-y")
			if s.onLoadPointY != nil {
				s.onLoadPointY(point)
			}
			return point.y
		},
		loadPointX: func(point *projectileCollideTestPoint4E87B0) int32 {
			s.events = append(s.events, "point-x")
			return point.x
		},
		damageMap: func(x, y, damage int32, damageType uint32, projectile *projectileCollideTestObject4E87B0) {
			s.events = append(s.events, "damage-map")
			s.gotMapX = x
			s.gotMapY = y
			s.gotMapDamage = damage
			s.gotMapDamageType = damageType
			s.gotMapProjectile = projectile
		},
		delayedDelete: func(obj *projectileCollideTestObject4E87B0) {
			s.events = append(s.events, "delete")
			s.deleted = append(s.deleted, obj)
		},
	}
}

func TestProjectileCollide4E87B0DefaultTargetOrderArgumentsAndDelete(t *testing.T) {
	data := &projectileCollideTestData4E87B0{damage: -0x1020304}
	projectile := &projectileCollideTestObject4E87B0{name: "projectile", typeIndex: 9, collideData: data}
	other := &projectileCollideTestObject4E87B0{name: "other"}
	parent := &projectileCollideTestObject4E87B0{name: "parent"}
	collision := uint32(0x7fa12345)
	state := &projectileCollideTestState4E87B0{
		t:                 t,
		throwingStoneType: 7,
		impShotType:       8,
		parent:            parent,
		damageResult:      0x80,
	}

	projectileCollide4E87B0(projectile, other, &collision, state.hooks())

	wantEvents := []string{
		"throw-cache", "collide-data", "throw-cache", "type", "imp-cache",
		"damage-data", "parent", "damage-callback", "delete",
	}
	if !reflect.DeepEqual(state.events, wantEvents) {
		t.Fatalf("events = %v, want %v", state.events, wantEvents)
	}
	if state.gotDamageTarget != other || state.gotDamageSource != parent || state.gotDamageAttacker != projectile {
		t.Fatalf("damage objects = (%p, %p, %p), want (%p, %p, %p)", state.gotDamageTarget, state.gotDamageSource, state.gotDamageAttacker, other, parent, projectile)
	}
	if state.gotDamage != data.damage || state.gotDamageType != projectileCollideDamageType4E87B0 {
		t.Fatalf("damage = (%d, %d), want (%d, %d)", state.gotDamage, state.gotDamageType, data.damage, projectileCollideDamageType4E87B0)
	}
	if !reflect.DeepEqual(state.deleted, []*projectileCollideTestObject4E87B0{projectile}) {
		t.Fatalf("deleted = %v, want projectile once", state.deleted)
	}
	if collision != 0x7fa12345 {
		t.Fatalf("collision word changed to %#x", collision)
	}
}

func TestProjectileCollide4E87B0ZeroDamageLowBytePreservesProjectile(t *testing.T) {
	projectile := &projectileCollideTestObject4E87B0{typeIndex: 3, collideData: &projectileCollideTestData4E87B0{damage: 99}}
	other := &projectileCollideTestObject4E87B0{}
	state := &projectileCollideTestState4E87B0{
		t:                 t,
		throwingStoneType: 1,
		impShotType:       2,
		damageResult:      0,
	}

	projectileCollide4E87B0(projectile, other, (*uint32)(nil), state.hooks())

	wantEvents := []string{
		"throw-cache", "collide-data", "throw-cache", "type", "imp-cache",
		"damage-data", "parent", "damage-callback",
	}
	if !reflect.DeepEqual(state.events, wantEvents) {
		t.Fatalf("events = %v, want %v", state.events, wantEvents)
	}
	if len(state.deleted) != 0 {
		t.Fatalf("zero AL deleted %d objects", len(state.deleted))
	}
}

func TestProjectileCollide4E87B0InitializesBothCachesBeforeLiveTypeRead(t *testing.T) {
	oldData := &projectileCollideTestData4E87B0{damage: 17}
	newData := &projectileCollideTestData4E87B0{damage: 88}
	projectile := &projectileCollideTestObject4E87B0{typeIndex: 3, collideData: oldData}
	other := &projectileCollideTestObject4E87B0{}
	state := &projectileCollideTestState4E87B0{
		t:            t,
		lookup:       map[string]uint32{projectileCollideThrowingStoneType4E87B0: 23, projectileCollideImpShotType4E87B0: 29},
		damageResult: 1,
	}
	state.onLookup = func(name string) {
		if name == projectileCollideImpShotType4E87B0 {
			projectile.typeIndex = 31
			projectile.collideData = newData
		}
	}
	state.onLoadCollideData = func() {
		// The entry cache value was already sampled. Mutating the live cache
		// here must not suppress the two initialization lookups.
		state.throwingStoneType = 99
	}

	projectileCollide4E87B0(projectile, other, (*uint32)(nil), state.hooks())

	wantEvents := []string{
		"throw-cache", "collide-data",
		"lookup:ThrowingStone", "store-throw:23", "lookup:ImpShot", "store-imp:29",
		"throw-cache", "type", "imp-cache", "damage-data", "parent", "damage-callback", "delete",
	}
	if !reflect.DeepEqual(state.events, wantEvents) {
		t.Fatalf("events = %v, want %v", state.events, wantEvents)
	}
	if state.gotDamage != oldData.damage {
		t.Fatalf("damage = %d, want cached collide-data damage %d", state.gotDamage, oldData.damage)
	}
	if state.throwingStoneType != 23 || state.impShotType != 29 {
		t.Fatalf("caches = (%d, %d), want (23, 29)", state.throwingStoneType, state.impShotType)
	}
}

func TestProjectileCollide4E87B0PopulatedThrowCacheDoesNotRepairImpCache(t *testing.T) {
	projectile := &projectileCollideTestObject4E87B0{
		typeIndex:   9,
		collideData: &projectileCollideTestData4E87B0{damage: 27},
	}
	state := &projectileCollideTestState4E87B0{
		t:                 t,
		throwingStoneType: 7,
		impShotType:       0,
	}

	projectileCollide4E87B0(projectile, nil, (*uint32)(nil), state.hooks())

	wantEvents := []string{
		"throw-cache", "collide-data", "throw-cache", "type", "imp-cache",
		"damage-data", "trace-point", "delete",
	}
	if !reflect.DeepEqual(state.events, wantEvents) {
		t.Fatalf("events = %v, want %v", state.events, wantEvents)
	}
	if state.impShotType != 0 {
		t.Fatalf("ImpShot cache = %d, want unchanged zero", state.impShotType)
	}
}

func TestProjectileCollide4E87B0BalanceOverridesSpillToBinary32(t *testing.T) {
	tests := []struct {
		name          string
		typeIndex     uint16
		throwingStone uint32
		impShot       uint32
		balanceName   string
		balance       float64
	}{
		{name: "ThrowingStone", typeIndex: 17, throwingStone: 17, impShot: 19, balanceName: projectileCollideUrchinDamage4E87B0, balance: 16777217.25},
		{name: "ImpShot", typeIndex: 19, throwingStone: 17, impShot: 19, balanceName: projectileCollideImpShotDamage4E87B0, balance: math.SmallestNonzeroFloat64},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			projectile := &projectileCollideTestObject4E87B0{typeIndex: tc.typeIndex}
			other := &projectileCollideTestObject4E87B0{}
			state := &projectileCollideTestState4E87B0{
				t:                 t,
				throwingStoneType: tc.throwingStone,
				impShotType:       tc.impShot,
				gameData:          map[string]float64{tc.balanceName: tc.balance},
				floatResult:       -77,
				damageResult:      1,
			}

			projectileCollide4E87B0(projectile, other, (*uint32)(nil), state.hooks())

			if state.floatInput != float32(tc.balance) {
				t.Fatalf("float input bits = %#08x, want %#08x", math.Float32bits(state.floatInput), math.Float32bits(float32(tc.balance)))
			}
			if state.gotDamage != -77 {
				t.Fatalf("damage = %d, want rounded override -77", state.gotDamage)
			}
			if state.gotDamageType != projectileCollideDamageType4E87B0 {
				t.Fatalf("damage type = %d", state.gotDamageType)
			}
			if !containsProjectileCollideEvent4E87B0(state.events, "balance:"+tc.balanceName) || containsProjectileCollideEvent4E87B0(state.events, "damage-data") {
				t.Fatalf("override events = %v", state.events)
			}
		})
	}
}

func TestProjectileCollide4E87B0WallPointReadsYThenLiveXAndDeletes(t *testing.T) {
	projectile := &projectileCollideTestObject4E87B0{typeIndex: 3, collideData: &projectileCollideTestData4E87B0{damage: 41}}
	point := &projectileCollideTestPoint4E87B0{x: 10, y: -20}
	state := &projectileCollideTestState4E87B0{
		t:                 t,
		throwingStoneType: 1,
		impShotType:       2,
		point:             point,
	}
	state.onLoadPointY = func(got *projectileCollideTestPoint4E87B0) {
		if got != point {
			t.Fatalf("Y point = %p, want %p", got, point)
		}
		got.x = 77
	}

	projectileCollide4E87B0(projectile, nil, (*uint32)(nil), state.hooks())

	wantEvents := []string{
		"throw-cache", "collide-data", "throw-cache", "type", "imp-cache", "damage-data",
		"trace-point", "point-y", "point-x", "damage-map", "delete",
	}
	if !reflect.DeepEqual(state.events, wantEvents) {
		t.Fatalf("events = %v, want %v", state.events, wantEvents)
	}
	if state.gotMapX != 77 || state.gotMapY != -20 || state.gotMapDamage != 41 || state.gotMapDamageType != projectileCollideDamageType4E87B0 || state.gotMapProjectile != projectile {
		t.Fatalf("map args = (%d,%d,%d,%d,%p)", state.gotMapX, state.gotMapY, state.gotMapDamage, state.gotMapDamageType, state.gotMapProjectile)
	}
	if !reflect.DeepEqual(state.deleted, []*projectileCollideTestObject4E87B0{projectile}) {
		t.Fatalf("deleted = %v", state.deleted)
	}
}

func TestProjectileCollide4E87B0NilWallPointStillDeletes(t *testing.T) {
	projectile := &projectileCollideTestObject4E87B0{typeIndex: 3, collideData: &projectileCollideTestData4E87B0{damage: 41}}
	state := &projectileCollideTestState4E87B0{t: t, throwingStoneType: 1, impShotType: 2}
	projectileCollide4E87B0(projectile, nil, (*uint32)(nil), state.hooks())
	wantEvents := []string{
		"throw-cache", "collide-data", "throw-cache", "type", "imp-cache", "damage-data", "trace-point", "delete",
	}
	if !reflect.DeepEqual(state.events, wantEvents) {
		t.Fatalf("events = %v, want %v", state.events, wantEvents)
	}
}

func TestProjectileCollide4E87B0SpecialTypeDoesNotDereferenceNilCollideData(t *testing.T) {
	projectile := &projectileCollideTestObject4E87B0{typeIndex: 7}
	state := &projectileCollideTestState4E87B0{
		t:                 t,
		throwingStoneType: 7,
		impShotType:       8,
		gameData:          map[string]float64{projectileCollideUrchinDamage4E87B0: 5},
		floatResult:       5,
	}
	projectileCollide4E87B0(projectile, nil, (*uint32)(nil), state.hooks())
	if containsProjectileCollideEvent4E87B0(state.events, "damage-data") {
		t.Fatalf("special type dereferenced nil collide data: %v", state.events)
	}
}

func TestProjectileCollide4E87B0OrdinaryTypeFaultsOnCachedNilCollideData(t *testing.T) {
	projectile := &projectileCollideTestObject4E87B0{typeIndex: 9}
	state := &projectileCollideTestState4E87B0{t: t, throwingStoneType: 7, impShotType: 8}
	defer func() {
		if recover() == nil {
			t.Fatal("ordinary projectile with nil collide data did not fault")
		}
		wantEvents := []string{"throw-cache", "collide-data", "throw-cache", "type", "imp-cache", "damage-data"}
		if !reflect.DeepEqual(state.events, wantEvents) {
			t.Fatalf("events before fault = %v, want %v", state.events, wantEvents)
		}
	}()
	projectileCollide4E87B0(projectile, nil, (*uint32)(nil), state.hooks())
}

func TestProjectileCollide4E87B0NilProjectileFaultsBeforeCache(t *testing.T) {
	state := &projectileCollideTestState4E87B0{t: t, throwingStoneType: 1, impShotType: 2}
	defer func() {
		if recover() == nil {
			t.Fatal("nil projectile did not fault")
		}
		if !reflect.DeepEqual(state.events, []string{"throw-cache", "collide-data"}) {
			t.Fatalf("events before fault = %v, want cache then collide-data", state.events)
		}
	}()
	projectileCollide4E87B0(
		(*projectileCollideTestObject4E87B0)(nil),
		(*projectileCollideTestObject4E87B0)(nil),
		(*uint32)(nil),
		state.hooks(),
	)
}

func containsProjectileCollideEvent4E87B0(events []string, want string) bool {
	for _, event := range events {
		if event == want {
			return true
		}
	}
	return false
}
