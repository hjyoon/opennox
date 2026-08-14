package server

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestDamageCollideNative4E9430(t *testing.T) {
	data := &DamageCollideData{Damage: 1, Reserved: [3]uint8{2, 3, 4}, DamageType: -11}
	parent := &Object{ObjClass: object.ClassPlayer}
	source := &Object{CollideData: unsafe.Pointer(data), ObjOwner: parent}
	health := &HealthData{}
	target := &Object{HealthData: health}
	var markerBefore, markerAfter byte
	target.Damage = unsafe.Pointer(&markerBefore)

	var gotTarget, gotSource, gotAttacker *Object
	var gotDamage, gotDamageType int32
	damageCollideNative4E9430(source, target, unsafe.Pointer(&markerBefore), damageCollideNativeDeps4E9430{
		loadFrameLow: func() uint8 { return 1 },
		findParent: func(obj *Object) *Object {
			target.Damage = unsafe.Pointer(&markerAfter)
			return obj.FindOwnerChainPlayer()
		},
		damage: func(targetArg, sourceArg, attackerArg *Object, damage, damageType int32) int32 {
			if targetArg.Damage != unsafe.Pointer(&markerAfter) {
				t.Fatalf("damage callback pointer was not observed after parent callback")
			}
			gotTarget, gotSource, gotAttacker = targetArg, sourceArg, attackerArg
			gotDamage, gotDamageType = damage, damageType
			return -1
		},
	})
	if gotTarget != target || gotSource != parent || gotAttacker != source || gotDamage != 1 || gotDamageType != -11 {
		t.Fatalf("native callback = (%p,%p,%p,%d,%d)", gotTarget, gotSource, gotAttacker, gotDamage, gotDamageType)
	}
	if data.Reserved != [3]uint8{2, 3, 4} || target.HealthData != health {
		t.Fatalf("native records changed: data=%+v health=%p", data, target.HealthData)
	}
}

func TestDamageCollideNative4E9430NilTargetDoesNotDereferenceData(t *testing.T) {
	source := &Object{}
	damageCollideNative4E9430(source, nil, nil, damageCollideNativeDeps4E9430{
		loadFrameLow: func() uint8 { t.Fatal("frame read"); return 0 },
		findParent:   func(*Object) *Object { t.Fatal("parent lookup"); return nil },
		damage: func(*Object, *Object, *Object, int32, int32) int32 {
			t.Fatal("damage callback")
			return 0
		},
	})
}

func TestDamageCollide4E9430Layouts(t *testing.T) {
	wantHealth, wantCollideData, wantDamage := uintptr(556), uintptr(700), uintptr(716)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantHealth, wantCollideData, wantDamage = 616, 776, 808
	}
	tests := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"DamageCollideData size", unsafe.Sizeof(DamageCollideData{}), 8},
		{"DamageCollideData.Damage", unsafe.Offsetof(DamageCollideData{}.Damage), 0},
		{"DamageCollideData.DamageType", unsafe.Offsetof(DamageCollideData{}.DamageType), 4},
		{"Object.HealthData", unsafe.Offsetof(Object{}.HealthData), wantHealth},
		{"Object.CollideData", unsafe.Offsetof(Object{}.CollideData), wantCollideData},
		{"Object.Damage", unsafe.Offsetof(Object{}.Damage), wantDamage},
	}
	for _, tc := range tests {
		if tc.got != tc.want {
			t.Errorf("%s = %d, want %d", tc.name, tc.got, tc.want)
		}
	}
}
