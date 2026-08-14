package server

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestDeathBallFragmentCollide4E9FE0NativeLayout(t *testing.T) {
	wantNewPos := uintptr(64)
	wantVel := uintptr(80)
	wantDamage := uintptr(716)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantNewPos = 68
		wantVel = 84
		wantDamage = 808
	}
	if got := unsafe.Offsetof(Object{}.NewPos); got != wantNewPos {
		t.Fatalf("Object.NewPos offset = %d, want %d", got, wantNewPos)
	}
	if got := unsafe.Offsetof(Object{}.VelVec); got != wantVel {
		t.Fatalf("Object.VelVec offset = %d, want %d", got, wantVel)
	}
	if got := unsafe.Offsetof(Object{}.Damage); got != wantDamage {
		t.Fatalf("Object.Damage offset = %d, want %d", got, wantDamage)
	}
	if got := unsafe.Sizeof(types.Pointf{}); got != 8 {
		t.Fatalf("Pointf size = %d, want 8", got)
	}
	if got := unsafe.Offsetof(types.Pointf{}.X); got != 0 {
		t.Fatalf("Pointf.X offset = %d, want 0", got)
	}
	if got := unsafe.Offsetof(types.Pointf{}.Y); got != 4 {
		t.Fatalf("Pointf.Y offset = %d, want 4", got)
	}
}

func TestDeathBallFragmentCollideNative4E9FE0TargetUsesLiveDamageThenDeletes(t *testing.T) {
	source := &Object{}
	target := &Object{}
	parent := &Object{}
	collision := &types.Pointf{X: 1, Y: 2}
	events := make([]string, 0, 3)
	target.ObjFlags = 1

	deathBallFragmentCollideNative4E9FE0(
		source,
		target,
		collision,
		deathBallFragmentCollideNativeDeps4E9FE0{
			findParent: func(got *Object) *Object {
				events = append(events, "parent")
				if got != source {
					t.Fatalf("parent source = %p", got)
				}
				target.ObjFlags = 2
				return parent
			},
			targetDamage: func(gotTarget, gotParent, gotSource *Object, damage int32, damageType object.DamageType) int32 {
				events = append(events, "damage")
				if gotTarget != target || gotParent != parent || gotSource != source ||
					damage != 20 || damageType != object.DamageCrush {
					t.Fatalf("damage = %p/%p/%p/%d/%d", gotTarget, gotParent, gotSource, damage, damageType)
				}
				if target.ObjFlags != 2 {
					t.Fatal("target state was cached before parent lookup")
				}
				return 0
			},
			wallReflect: func(*types.Pointf, *Object) {
				t.Fatal("target branch read collision")
			},
			delayedDelete: func(got *Object) {
				events = append(events, "delete")
				if got != source {
					t.Fatalf("deleted = %p", got)
				}
			},
		},
	)

	if want := []string{"parent", "damage", "delete"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestDeathBallFragmentCollideNative4E9FE0WallUsesLiveYThenX(t *testing.T) {
	source := &Object{NewPos: types.Ptf(1, 2)}
	collision := &types.Pointf{X: 3, Y: 4}
	events := make([]string, 0, 5)
	results := []int32{-9, 12}
	var got [4]int32
	var gotSource *Object

	deathBallFragmentCollideNative4E9FE0(
		source,
		nil,
		collision,
		deathBallFragmentCollideNativeDeps4E9FE0{
			wallReflect: func(gotCollision *types.Pointf, gotSource *Object) {
				events = append(events, "reflect")
				if gotCollision != collision || gotSource != source {
					t.Fatalf("reflect = %p/%p", gotCollision, gotSource)
				}
				source.NewPos.Y = 69
			},
			audio: func(id uint32, gotSource *Object) {
				events = append(events, "audio")
				if id != 37 || gotSource != source {
					t.Fatalf("audio = %d/%p", id, gotSource)
				}
			},
			floatToInt: func(float32) int32 {
				events = append(events, "float")
				if len(results) == 2 {
					source.NewPos.X = 92
				}
				result := results[0]
				results = results[1:]
				return result
			},
			damageMap: func(x, y, damage int32, damageType object.DamageType, source *Object) {
				events = append(events, "map")
				got = [4]int32{x, y, damage, int32(damageType)}
				gotSource = source
			},
			delayedDelete: func(*Object) {
				t.Fatal("wall path deleted source")
			},
		},
	)

	if want := []string{"reflect", "audio", "float", "float", "map"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	if got != [4]int32{12, -9, 20, int32(object.DamageCrush)} || gotSource != source {
		t.Fatalf("map = %#v/%p", got, gotSource)
	}
}
