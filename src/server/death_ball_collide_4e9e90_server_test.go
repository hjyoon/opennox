package server

import (
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/ntype"
)

func TestDeathBallCollide4E9E90NativeLayout(t *testing.T) {
	wantClass := uintptr(8)
	wantPos := uintptr(56)
	wantNewPos := uintptr(64)
	wantPrevPos := uintptr(72)
	wantVelocity := uintptr(80)
	wantDamage := uintptr(716)
	wantUpdate := uintptr(748)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantClass = 12
		wantPos = 60
		wantNewPos = 68
		wantPrevPos = 76
		wantVelocity = 84
		wantDamage = 808
		wantUpdate = 872
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantClass},
		{"Object.PosVec", unsafe.Offsetof(Object{}.PosVec), wantPos},
		{"Object.NewPos", unsafe.Offsetof(Object{}.NewPos), wantNewPos},
		{"Object.PrevPos", unsafe.Offsetof(Object{}.PrevPos), wantPrevPos},
		{"Object.VelVec", unsafe.Offsetof(Object{}.VelVec), wantVelocity},
		{"Object.Damage", unsafe.Offsetof(Object{}.Damage), wantDamage},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantUpdate},
		{"DoorUpdateData size", unsafe.Sizeof(DoorUpdateData{}), 52},
		{"DoorUpdateData.CurrentDirection", unsafe.Offsetof(DoorUpdateData{}.CurrentDirection), 12},
		{"Point32 size", unsafe.Sizeof(ntype.Point32{}), 8},
		{"Point32.X", unsafe.Offsetof(ntype.Point32{}.X), 0},
		{"Point32.Y", unsafe.Offsetof(ntype.Point32{}.Y), 4},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestDeathBallCollideNative4E9E90DoorFieldsAndDirection(t *testing.T) {
	update := &DoorUpdateData{CurrentDirection: 7}
	source := &Object{
		PrevPos: types.Ptf(2, 3),
		VelVec:  types.Ptf(6, -8),
	}
	target := &Object{
		ObjClass:   object.ClassDoor,
		PosVec:     types.Ptf(100, 50),
		UpdateData: unsafe.Pointer(update),
	}
	var events []string
	var normal types.Pointf

	deathBallCollideNative4E9E90(source, target, nil, deathBallCollideNativeDeps4E9E90{
		doorReflect: func(got *Object, x, y float32) {
			events = append(events, "reflect")
			if got != source {
				t.Fatalf("reflect source = %p", got)
			}
			normal = types.Ptf(x, y)
			got.VelVec.X, got.VelVec.Y = deathBallDoorReflect57B770(got.VelVec.X, got.VelVec.Y, x, y)
		},
		audio: func(id uint32, got *Object) {
			events = append(events, "audio")
			if id != deathBallDoorReflectAudio4E9E90 || got != source {
				t.Fatalf("audio = %d/%p", id, got)
			}
		},
	})

	if source.NewPos != source.PrevPos {
		t.Fatalf("NewPos = %v, want PrevPos %v", source.NewPos, source.PrevPos)
	}
	if normal != (types.Pointf{X: -27, Y: -18}) {
		t.Fatalf("normal = %v, want {-27 -18}", normal)
	}
	if reflect.DeepEqual(source.VelVec, types.Ptf(6, -8)) {
		t.Fatalf("velocity was not reflected: %v", source.VelVec)
	}
	if !reflect.DeepEqual(events, []string{"reflect", "audio"}) {
		t.Fatalf("events = %#v", events)
	}
}

func TestDeathBallCollideNative4E9E90NonDoorDamageArguments(t *testing.T) {
	source := &Object{}
	target := &Object{ObjClass: object.ClassPlayer}
	parent := &Object{}
	var gotTarget, gotParent, gotSource *Object
	var gotDamage int32
	var gotType object.DamageType

	deathBallCollideNative4E9E90(source, target, nil, deathBallCollideNativeDeps4E9E90{
		balanceFloat: func(key string) float64 {
			if key != deathBallCollideDamageKey4E9E90 {
				t.Fatalf("balance key = %q", key)
			}
			return 10.5
		},
		floatToInt: playerCollideRound4E8460,
		findParent: func(got *Object) *Object {
			if got != source {
				t.Fatalf("owner source = %p, want %p", got, source)
			}
			return parent
		},
		targetDamage: func(gotTarg, gotPar, gotSrc *Object, damage int32, damageType object.DamageType) int32 {
			gotTarget, gotParent, gotSource = gotTarg, gotPar, gotSrc
			gotDamage, gotType = damage, damageType
			return -1
		},
	})

	if gotTarget != target || gotParent != parent || gotSource != source ||
		gotDamage != 10 || gotType != object.DamageCrush {
		t.Fatalf("damage = %p/%p/%p/%d/%d", gotTarget, gotParent, gotSource, gotDamage, gotType)
	}
}

func TestDeathBallCollideNative4E9E90WallUsesFixedWidthTrace(t *testing.T) {
	source := &Object{VelVec: types.Ptf(2, 3)}
	collision := &types.Pointf{X: 1, Y: 1}
	point := &ntype.Point32{X: math.MinInt32, Y: math.MaxInt32}
	var got [4]int32
	var gotSource *Object

	deathBallCollideNative4E9E90(source, nil, collision, deathBallCollideNativeDeps4E9E90{
		wallReflect: spellProjectileWallReflect57B810,
		audio:       func(uint32, *Object) {},
		traceHitPoint: func() *ntype.Point32 {
			return point
		},
		balanceFloat: func(string) float64 { return 20.5 },
		floatToInt:   playerCollideRound4E8460,
		damageMap: func(x, y, damage int32, damageType object.DamageType, source *Object) {
			got = [4]int32{x, y, damage, int32(damageType)}
			gotSource = source
		},
	})

	if source.VelVec != (types.Pointf{X: -3, Y: -2}) {
		t.Fatalf("velocity = %v, want {-3 -2}", source.VelVec)
	}
	if got != [4]int32{math.MinInt32, math.MaxInt32, 20, int32(object.DamageCrush)} || gotSource != source {
		t.Fatalf("map = %#v/%p", got, gotSource)
	}
}
