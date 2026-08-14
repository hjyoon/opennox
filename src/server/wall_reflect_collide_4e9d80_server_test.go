package server

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestWallReflectCollide4E9D80NativeLayout(t *testing.T) {
	wantPos := uintptr(56)
	wantNewPos := uintptr(64)
	wantVelocity := uintptr(80)
	wantCollide := uintptr(696)
	wantCollideData := uintptr(700)
	wantDamage := uintptr(716)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantPos = 60
		wantNewPos = 68
		wantVelocity = 84
		wantCollide = 768
		wantCollideData = 776
		wantDamage = 808
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"ProjectileCollideData size", unsafe.Sizeof(ProjectileCollideData{}), 8},
		{"ProjectileCollideData.Damage", unsafe.Offsetof(ProjectileCollideData{}.Damage), 0},
		{"ProjectileCollideData.Field4", unsafe.Offsetof(ProjectileCollideData{}.Field4), 4},
		{"Object.PosVec", unsafe.Offsetof(Object{}.PosVec), wantPos},
		{"Object.NewPos", unsafe.Offsetof(Object{}.NewPos), wantNewPos},
		{"Object.VelVec", unsafe.Offsetof(Object{}.VelVec), wantVelocity},
		{"Object.Collide", unsafe.Offsetof(Object{}.Collide), wantCollide},
		{"Object.CollideData", unsafe.Offsetof(Object{}.CollideData), wantCollideData},
		{"Object.Damage", unsafe.Offsetof(Object{}.Damage), wantDamage},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestWallReflectCollideNative4E9D80CachedDataAndCallbackIdentity(t *testing.T) {
	oldData := &ProjectileCollideData{Damage: -11, Field4: 0x12345678}
	newData := &ProjectileCollideData{Damage: 99}
	var yellowToken byte
	yellow := unsafe.Pointer(&yellowToken)
	source := &Object{Collide: yellow, CollideData: unsafe.Pointer(oldData)}
	target := &Object{}
	parent := &Object{}
	var events []string

	wallReflectCollideNative4E9D80(source, target, nil, wallReflectCollideNativeDeps4E9D80{
		sameTeam: func(first, second *Object) int32 {
			events = append(events, "team")
			if first != source || second != target {
				t.Fatalf("team args = %p/%p", first, second)
			}
			source.CollideData = unsafe.Pointer(newData)
			return 0
		},
		gameFlagsCheck: func(flag uint32) int32 {
			events = append(events, "quest")
			if flag != wallReflectQuestFlag4E9D80 {
				t.Fatalf("flag = %#x", flag)
			}
			return 1
		},
		yellowStarCollide: yellow,
		findParent: func(got *Object) *Object {
			events = append(events, "parent")
			if got != source {
				t.Fatalf("parent source = %p", got)
			}
			return parent
		},
		targetDamage: func(gotTarget, gotParent, gotSource *Object, damage int32, damageType object.DamageType) int32 {
			events = append(events, "damage")
			if gotTarget != target || gotParent != parent || gotSource != source || damage != -33 || damageType != object.DamageImpact {
				t.Fatalf("damage args = %p/%p/%p/%d/%d", gotTarget, gotParent, gotSource, damage, damageType)
			}
			return -1
		},
		delayedDelete: func(got *Object) {
			events = append(events, "delete")
			if got != source {
				t.Fatalf("delete = %p", got)
			}
		},
	})

	if !reflect.DeepEqual(events, []string{"team", "quest", "parent", "damage", "delete"}) {
		t.Fatalf("events = %#v", events)
	}
	if oldData.Field4 != 0x12345678 || source.CollideData != unsafe.Pointer(newData) {
		t.Fatalf("collide data changed unexpectedly: %#x/%p", oldData.Field4, source.CollideData)
	}
}

func TestWallReflectCollideNative4E9D80WallFields(t *testing.T) {
	data := &ProjectileCollideData{Damage: 17, Field4: -1}
	source := &Object{
		NewPos:      types.Ptf(-69, 46),
		VelVec:      types.Ptf(2, 3),
		CollideData: unsafe.Pointer(data),
	}
	collision := &types.Pointf{X: 1, Y: 1}
	var gotMap [4]int32
	var gotSource *Object

	wallReflectCollideNative4E9D80(source, nil, collision, wallReflectCollideNativeDeps4E9D80{
		wallReflect: spellProjectileWallReflect57B810,
		floatToInt:  playerCollideRound4E8460,
		damageMap: func(x, y, damage int32, damageType object.DamageType, got *Object) {
			gotMap = [4]int32{x, y, damage, int32(damageType)}
			gotSource = got
		},
		delayedDelete: func(*Object) { t.Fatal("wall branch deleted source") },
	})

	if source.VelVec != (types.Pointf{X: -3, Y: -2}) {
		t.Fatalf("velocity = %#v, want {-3 -2}", source.VelVec)
	}
	if gotMap != [4]int32{-3, 2, 17, int32(object.DamageImpact)} || gotSource != source {
		t.Fatalf("map = %v/%p", gotMap, gotSource)
	}
}

func TestYellowStarShotCollideNative4E9E50PositionAndPointers(t *testing.T) {
	source := &Object{PosVec: types.Ptf(12.5, -7.25)}
	target := &Object{}
	collision := &types.Pointf{X: 3, Y: 4}
	var events []string

	yellowStarShotCollideNative4E9E50(source, target, collision, yellowStarShotCollideNativeDeps4E9E50{
		gameFlagsCheck: func(flag uint32) int32 {
			events = append(events, "flag")
			if flag != yellowStarSuppressFXFlag4E9E50 {
				t.Fatalf("flag = %#x", flag)
			}
			return 0
		},
		pointFX: func(id uint32, pos types.Pointf) {
			events = append(events, "fx")
			if id != yellowStarPointFX4E9E50 || pos != source.PosVec {
				t.Fatalf("FX = %d/%#v", id, pos)
			}
		},
		wallCollide: func(gotSource, gotTarget *Object, gotCollision *types.Pointf) {
			events = append(events, "wall")
			if gotSource != source || gotTarget != target || gotCollision != collision {
				t.Fatalf("wall args = %p/%p/%p", gotSource, gotTarget, gotCollision)
			}
		},
	})

	if !reflect.DeepEqual(events, []string{"flag", "fx", "wall"}) {
		t.Fatalf("events = %#v", events)
	}
}
