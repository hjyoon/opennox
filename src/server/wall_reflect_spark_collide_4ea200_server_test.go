package server

import (
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func defaultWallReflectSparkNativeDeps4EA200() wallReflectSparkCollideNativeDeps4EA200 {
	return wallReflectSparkCollideNativeDeps4EA200{
		findParent:    func(obj *Object) *Object { return obj.ObjOwner },
		targetDamage:  func(*Object, *Object, *Object, int32, object.DamageType) int32 { return 0 },
		delayedDelete: func(*Object) {},
		floatToInt:    playerCollideRound4E8460,
		damageMap:     func(int32, int32, int32, object.DamageType, *Object) {},
	}
}

func TestWallReflectSparkCollide4EA200NativeLayout(t *testing.T) {
	wantNewPos := uintptr(64)
	wantVelocity := uintptr(80)
	wantCollideData := uintptr(700)
	wantDamage := uintptr(716)
	wantObjectSize := uintptr(780)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantNewPos = 68
		wantVelocity = 84
		wantCollideData = 776
		wantDamage = 808
		wantObjectSize = 928
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.NewPos", unsafe.Offsetof(Object{}.NewPos), wantNewPos},
		{"Object.VelVec", unsafe.Offsetof(Object{}.VelVec), wantVelocity},
		{"Object.CollideData", unsafe.Offsetof(Object{}.CollideData), wantCollideData},
		{"Object.Damage", unsafe.Offsetof(Object{}.Damage), wantDamage},
		{"ProjectileCollideData size", unsafe.Sizeof(ProjectileCollideData{}), 8},
		{"ProjectileCollideData.Damage", unsafe.Offsetof(ProjectileCollideData{}.Damage), 0},
		{"ProjectileCollideData.Field4", unsafe.Offsetof(ProjectileCollideData{}.Field4), 4},
		{"Pointf size", unsafe.Sizeof(types.Pointf{}), 8},
		{"Pointf.X", unsafe.Offsetof(types.Pointf{}.X), 0},
		{"Pointf.Y", unsafe.Offsetof(types.Pointf{}.Y), 4},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestWallReflectSparkCollideNative4EA200TargetUsesCachedDataAndLiveDamage(t *testing.T) {
	oldData := &ProjectileCollideData{Damage: -31, Field4: 0x12345678}
	parent := &Object{}
	source := &Object{CollideData: unsafe.Pointer(oldData), ObjOwner: parent}
	target := &Object{ObjFlags: 1}
	events := make([]string, 0, 4)
	deps := defaultWallReflectSparkNativeDeps4EA200()
	deps.findParent = func(got *Object) *Object {
		events = append(events, "parent")
		if got != source {
			t.Fatalf("parent source = %p", got)
		}
		target.ObjFlags = 2
		return parent
	}
	deps.targetDamage = func(
		gotTarget, gotParent, gotSource *Object,
		damage int32,
		damageType object.DamageType,
	) int32 {
		events = append(events, "damage")
		if gotTarget != target || gotParent != parent || gotSource != source || damage != -31 || damageType != 11 {
			t.Fatalf("damage = %p/%p/%p/%d/%d", gotTarget, gotParent, gotSource, damage, damageType)
		}
		if target.ObjFlags != 2 {
			t.Fatal("Damage callback was not observed live after parent lookup")
		}
		return 0x100
	}
	deps.delayedDelete = func(got *Object) {
		events = append(events, "delete")
		if got != source {
			t.Fatalf("deleted source = %p", got)
		}
	}
	wallReflectSparkCollideNative4EA200(source, target, nil, deps)
	if !reflect.DeepEqual(events, []string{"parent", "damage", "delete"}) {
		t.Fatalf("events = %v", events)
	}
}

func TestWallReflectSparkCollideNative4EA200WallReflectsAndUsesLiveX(t *testing.T) {
	data := &ProjectileCollideData{Damage: -47, Field4: 91}
	tiny := math.Float32frombits(1)
	source := &Object{
		CollideData: unsafe.Pointer(data),
		VelVec:      types.Ptf(7, -11),
		NewPos:      types.Ptf(69, 46),
	}
	collision := &types.Pointf{X: tiny, Y: tiny}
	inputs := make([]uint32, 0, 2)
	deps := defaultWallReflectSparkNativeDeps4EA200()
	deps.floatToInt = func(value float32) int32 {
		inputs = append(inputs, math.Float32bits(value))
		if len(inputs) == 1 {
			source.NewPos.X = 92
			return -17
		}
		return 29
	}
	deps.damageMap = func(x, y, damage int32, damageType object.DamageType, got *Object) {
		if x != 29 || y != -17 || damage != -47 || damageType != 11 || got != source {
			t.Fatalf("map = %d/%d/%d/%d/%p", x, y, damage, damageType, got)
		}
	}
	deps.delayedDelete = func(*Object) { t.Fatal("wall path deleted source") }
	wallReflectSparkCollideNative4EA200(source, nil, collision, deps)
	if source.VelVec != (types.Pointf{X: 11, Y: -7}) {
		t.Fatalf("velocity = %v, want {11 -7}", source.VelVec)
	}
	gridInverse := math.Float32frombits(wallReflectSparkGridInverseBits4EA200)
	wantInputs := []uint32{
		math.Float32bits(float32(46) * gridInverse),
		math.Float32bits(float32(92) * gridInverse),
	}
	if !reflect.DeepEqual(inputs, wantInputs) {
		t.Fatalf("round inputs = %#v, want %#v", inputs, wantInputs)
	}
}
