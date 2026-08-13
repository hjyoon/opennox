package server

import (
	"math"
	"reflect"
	"testing"
)

type projectileSparkTestObject4E8880 struct {
	name        string
	newX        float32
	newY        float32
	collideData *projectileSparkTestData4E8880
}

type projectileSparkTestData4E8880 struct {
	damage int32
}

type projectileSparkTestState4E8880 struct {
	events       []string
	parent       *projectileSparkTestObject4E8880
	damageResult uint8
	floatResults []int32
	floatInputs  []uint32
	damageArgs   struct {
		target, source, attacker *projectileSparkTestObject4E8880
		damage                   int32
		damageType               uint32
	}
	mapArgs struct {
		x, y, damage int32
		damageType   uint32
		projectile   *projectileSparkTestObject4E8880
	}
	deleted []*projectileSparkTestObject4E8880
	onLoadY func(*projectileSparkTestObject4E8880)
	onFloat func(int, *projectileSparkTestObject4E8880)
}

func (s *projectileSparkTestState4E8880) hooks(projectile *projectileSparkTestObject4E8880) projectileSparkCollideHooks4E8880[
	*projectileSparkTestObject4E8880,
	*projectileSparkTestData4E8880,
] {
	return projectileSparkCollideHooks4E8880[
		*projectileSparkTestObject4E8880,
		*projectileSparkTestData4E8880,
	]{
		loadCollideData: func(obj *projectileSparkTestObject4E8880) *projectileSparkTestData4E8880 {
			s.events = append(s.events, "collide-data")
			return obj.collideData
		},
		loadDamage: func(data *projectileSparkTestData4E8880) int32 {
			s.events = append(s.events, "damage-data")
			return data.damage
		},
		findParent: func(*projectileSparkTestObject4E8880) *projectileSparkTestObject4E8880 {
			s.events = append(s.events, "parent")
			return s.parent
		},
		damage: func(target, source, attacker *projectileSparkTestObject4E8880, damage int32, damageType uint32) uint8 {
			s.events = append(s.events, "damage-callback")
			s.damageArgs.target = target
			s.damageArgs.source = source
			s.damageArgs.attacker = attacker
			s.damageArgs.damage = damage
			s.damageArgs.damageType = damageType
			return s.damageResult
		},
		loadNewPosY: func(obj *projectileSparkTestObject4E8880) float32 {
			s.events = append(s.events, "new-y")
			value := obj.newY
			if s.onLoadY != nil {
				s.onLoadY(obj)
			}
			return value
		},
		loadNewPosX: func(obj *projectileSparkTestObject4E8880) float32 {
			s.events = append(s.events, "new-x")
			return obj.newX
		},
		floatToInt: func(value float32) int32 {
			s.events = append(s.events, "float-to-int")
			i := len(s.floatInputs)
			s.floatInputs = append(s.floatInputs, math.Float32bits(value))
			if s.onFloat != nil {
				s.onFloat(i, projectile)
			}
			result := s.floatResults[0]
			s.floatResults = s.floatResults[1:]
			return result
		},
		damageMap: func(x, y, damage int32, damageType uint32, gotProjectile *projectileSparkTestObject4E8880) {
			s.events = append(s.events, "damage-map")
			s.mapArgs.x = x
			s.mapArgs.y = y
			s.mapArgs.damage = damage
			s.mapArgs.damageType = damageType
			s.mapArgs.projectile = gotProjectile
		},
		delayedDelete: func(obj *projectileSparkTestObject4E8880) {
			s.events = append(s.events, "delete")
			s.deleted = append(s.deleted, obj)
		},
	}
}

func TestProjectileSparkCollide4E8880TargetOrderArgumentsALAndIgnoredCollision(t *testing.T) {
	data := &projectileSparkTestData4E8880{damage: -0x1020304}
	projectile := &projectileSparkTestObject4E8880{name: "projectile", collideData: data}
	other := &projectileSparkTestObject4E8880{name: "other"}
	parent := &projectileSparkTestObject4E8880{name: "parent"}
	collision := uint32(0xa5c3f17e)
	state := &projectileSparkTestState4E8880{parent: parent, damageResult: 0x80}

	projectileSparkCollide4E8880(projectile, other, &collision, state.hooks(projectile))

	wantEvents := []string{"collide-data", "damage-data", "parent", "damage-callback", "delete"}
	if !reflect.DeepEqual(state.events, wantEvents) {
		t.Fatalf("events = %v, want %v", state.events, wantEvents)
	}
	if state.damageArgs.target != other || state.damageArgs.source != parent || state.damageArgs.attacker != projectile ||
		state.damageArgs.damage != data.damage || state.damageArgs.damageType != projectileSparkDamageType4E8880 {
		t.Fatalf("damage args = %+v", state.damageArgs)
	}
	if !reflect.DeepEqual(state.deleted, []*projectileSparkTestObject4E8880{projectile}) {
		t.Fatalf("deleted = %v, want projectile once", state.deleted)
	}
	if collision != 0xa5c3f17e {
		t.Fatalf("ignored collision changed to %#x", collision)
	}
}

func TestProjectileSparkCollide4E8880TargetZeroALKeepsProjectile(t *testing.T) {
	projectile := &projectileSparkTestObject4E8880{collideData: &projectileSparkTestData4E8880{damage: 17}}
	other := &projectileSparkTestObject4E8880{}
	state := &projectileSparkTestState4E8880{}

	projectileSparkCollide4E8880(projectile, other, (*uint32)(nil), state.hooks(projectile))

	wantEvents := []string{"collide-data", "damage-data", "parent", "damage-callback"}
	if !reflect.DeepEqual(state.events, wantEvents) {
		t.Fatalf("events = %v, want %v", state.events, wantEvents)
	}
	if len(state.deleted) != 0 {
		t.Fatalf("zero AL deleted %d objects", len(state.deleted))
	}
}

func TestProjectileSparkCollide4E8880MapOrderCachedDataLiveXAndDelete(t *testing.T) {
	oldData := &projectileSparkTestData4E8880{damage: -73}
	newData := &projectileSparkTestData4E8880{damage: 99}
	projectile := &projectileSparkTestObject4E8880{newX: -46, newY: 23, collideData: oldData}
	state := &projectileSparkTestState4E8880{floatResults: []int32{7, -8}}
	state.onLoadY = func(obj *projectileSparkTestObject4E8880) {
		obj.collideData = newData
	}
	state.onFloat = func(index int, obj *projectileSparkTestObject4E8880) {
		if index == 0 {
			obj.newX = 69
		}
	}

	projectileSparkCollide4E8880(projectile, nil, (*uint32)(nil), state.hooks(projectile))

	wantEvents := []string{
		"collide-data", "new-y", "damage-data", "float-to-int",
		"new-x", "float-to-int", "damage-map", "delete",
	}
	if !reflect.DeepEqual(state.events, wantEvents) {
		t.Fatalf("events = %v, want %v", state.events, wantEvents)
	}
	if wantBits := []uint32{0x3f800000, 0x40400000}; !reflect.DeepEqual(state.floatInputs, wantBits) {
		t.Fatalf("scaled coordinate bits = %08x, want %08x", state.floatInputs, wantBits)
	}
	if state.mapArgs.x != -8 || state.mapArgs.y != 7 || state.mapArgs.damage != oldData.damage ||
		state.mapArgs.damageType != projectileSparkDamageType4E8880 || state.mapArgs.projectile != projectile {
		t.Fatalf("map args = %+v", state.mapArgs)
	}
	if !reflect.DeepEqual(state.deleted, []*projectileSparkTestObject4E8880{projectile}) {
		t.Fatalf("deleted = %v, want projectile once", state.deleted)
	}
}

func TestProjectileSparkCollide4E8880NilDataFaultsAfterYBeforeConversion(t *testing.T) {
	projectile := &projectileSparkTestObject4E8880{newX: 46, newY: 23}
	state := &projectileSparkTestState4E8880{floatResults: []int32{1, 2}}
	defer func() {
		if recover() == nil {
			t.Fatal("nil collide-data did not fault")
		}
		wantEvents := []string{"collide-data", "new-y", "damage-data"}
		if !reflect.DeepEqual(state.events, wantEvents) {
			t.Fatalf("events = %v, want %v", state.events, wantEvents)
		}
	}()

	projectileSparkCollide4E8880(projectile, nil, (*uint32)(nil), state.hooks(projectile))
}

func TestProjectileSparkCollide4E8880NilProjectileFaultsBeforeBranchReads(t *testing.T) {
	state := &projectileSparkTestState4E8880{}
	hooks := state.hooks(nil)
	defer func() {
		if recover() == nil {
			t.Fatal("nil projectile did not fault")
		}
		if want := []string{"collide-data"}; !reflect.DeepEqual(state.events, want) {
			t.Fatalf("events = %v, want %v", state.events, want)
		}
	}()

	projectileSparkCollide4E8880[*projectileSparkTestObject4E8880](nil, nil, (*uint32)(nil), hooks)
}
