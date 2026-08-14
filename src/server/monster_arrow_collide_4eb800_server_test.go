package server

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/ntype"
)

func TestMonsterArrowCollide4EB800NativeLayout(t *testing.T) {
	if got := unsafe.Sizeof(MonsterArrowCollideData{}); got != 8 {
		t.Fatalf("collide-data size = %d, want 8", got)
	}
	if got := unsafe.Offsetof(MonsterArrowCollideData{}.CoopDamage); got != 0 {
		t.Fatalf("CoopDamage offset = %d, want 0", got)
	}
	if got := unsafe.Offsetof(MonsterArrowCollideData{}.OtherDamage); got != 4 {
		t.Fatalf("OtherDamage offset = %d, want 4", got)
	}
}

func TestMonsterArrowCollideNative4EB800UsesFixedWidthData(t *testing.T) {
	data := &MonsterArrowCollideData{CoopDamage: -31, OtherDamage: 47}
	parent := &Object{}
	source := &Object{CollideData: unsafe.Pointer(data)}
	target := &Object{}

	deleted := false
	monsterArrowCollideNative4EB800(source, target, nil, monsterArrowCollideNativeDeps4EB800{
		gameFlag: func(flag uint32) bool {
			if flag != monsterArrowCoopFlag4EB800 {
				t.Fatalf("game flag = %#x, want %#x", flag, monsterArrowCoopFlag4EB800)
			}
			return true
		},
		findParent: func(got *Object) *Object {
			if got != source {
				t.Fatalf("parent source = %p, want %p", got, source)
			}
			return parent
		},
		targetDamage: func(gotTarget, gotParent, gotSource *Object, damage int32, damageType object.DamageType) int32 {
			if gotTarget != target || gotParent != parent || gotSource != source {
				t.Fatalf("damage objects = (%p,%p,%p), want (%p,%p,%p)", gotTarget, gotParent, gotSource, target, parent, source)
			}
			if damage != -31 || damageType != object.DamageImpale {
				t.Fatalf("damage/type = (%d,%d), want (-31,%d)", damage, damageType, object.DamageImpale)
			}
			data.CoopDamage = 999
			return 0
		},
		tracePoint: func() *ntype.Point32 {
			t.Fatal("trace requested on target path")
			return nil
		},
		damageMap: func(int32, int32, int32, object.DamageType, *Object) {
			t.Fatal("map damaged on target path")
		},
		delayedDelete: func(got *Object) {
			if got != source {
				t.Fatalf("deleted object = %p, want %p", got, source)
			}
			deleted = true
		},
	})

	if !deleted || data.OtherDamage != 47 {
		t.Fatalf("deleted/data = %t/%+v", deleted, data)
	}
}

func TestMonsterArrowCollideNative4EB800WallUsesNativeTrace(t *testing.T) {
	data := &MonsterArrowCollideData{CoopDamage: -31, OtherDamage: 47}
	source := &Object{CollideData: unsafe.Pointer(data)}
	trace := &ntype.Point32{X: 12, Y: -7}

	mapCalls := 0
	deleteCalls := 0
	monsterArrowCollideNative4EB800(source, nil, nil, monsterArrowCollideNativeDeps4EB800{
		gameFlag: func(uint32) bool { return false },
		findParent: func(*Object) *Object {
			t.Fatal("parent requested on wall path")
			return nil
		},
		targetDamage: func(*Object, *Object, *Object, int32, object.DamageType) int32 {
			t.Fatal("target damaged on wall path")
			return 0
		},
		tracePoint: func() *ntype.Point32 { return trace },
		damageMap: func(x, y, damage int32, damageType object.DamageType, gotSource *Object) {
			mapCalls++
			if x != 12 || y != -7 || damage != 47 || damageType != object.DamageImpact || gotSource != source {
				t.Fatalf("map args = (%d,%d,%d,%d,%p)", x, y, damage, damageType, gotSource)
			}
		},
		delayedDelete: func(got *Object) {
			deleteCalls++
			if got != source {
				t.Fatalf("deleted object = %p, want %p", got, source)
			}
		},
	})

	if mapCalls != 1 || deleteCalls != 1 {
		t.Fatalf("map/delete calls = %d/%d, want 1/1", mapCalls, deleteCalls)
	}
}
