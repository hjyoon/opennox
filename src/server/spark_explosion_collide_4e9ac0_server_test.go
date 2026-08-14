package server

import (
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestSparkExplosionCollide4E9AC0NativeLayout(t *testing.T) {
	wantPos := uintptr(56)
	wantDirection := uintptr(124)
	wantCollideData := uintptr(700)
	wantDamage := uintptr(716)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantPos = 60
		wantDirection = 128
		wantCollideData = 776
		wantDamage = 808
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"SparkExplosionCollideData size", unsafe.Sizeof(SparkExplosionCollideData{}), 1},
		{"SparkExplosionCollideData.Power", unsafe.Offsetof(SparkExplosionCollideData{}.Power), 0},
		{"Object.PosVec", unsafe.Offsetof(Object{}.PosVec), wantPos},
		{"Object.Direction1", unsafe.Offsetof(Object{}.Direction1), wantDirection},
		{"Object.CollideData", unsafe.Offsetof(Object{}.CollideData), wantCollideData},
		{"Object.Damage", unsafe.Offsetof(Object{}.Damage), wantDamage},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestSparkExplosionCollideNative4E9AC0UsesCachedDataAndPointers(t *testing.T) {
	oldData := &SparkExplosionCollideData{Power: 9}
	newData := &SparkExplosionCollideData{Power: 250}
	source := &Object{PosVec: types.Ptf(10.25, -3.5), CollideData: unsafe.Pointer(oldData)}
	target := &Object{PosVec: types.Ptf(-7.75, 4.5)}
	parent := &Object{}
	collision := &types.Pointf{X: 123.5, Y: -77.25}
	var events []string

	sparkExplosionCollideNative4E9AC0(source, target, collision, sparkExplosionCollideNativeDeps4E9AC0{
		gameFlagsCheck: func(flag uint32) int32 {
			if flag == sparkExplosionQuestFlag4E9AC0 {
				events = append(events, "quest")
				return 0
			}
			events = append(events, "coop")
			return -1
		},
		findParent: func(got *Object) *Object {
			events = append(events, "parent")
			if got != source {
				t.Fatalf("parent source = %p, want %p", got, source)
			}
			return parent
		},
		mapPushUnits: func(pos types.Pointf, first, second, force float32, owner *Object, arg6, arg7 int32) {
			events = append(events, "push")
			if pos != source.PosVec || first != 6 || second != 0 ||
				force != math.Float32frombits(sparkExplosionPushForceBits4E9AC0) ||
				owner != source || arg6 != 0 || arg7 != 0 {
				t.Fatalf("unexpected push arguments: %#v %g %g %g %p %d %d", pos, first, second, force, owner, arg6, arg7)
			}
			source.CollideData = unsafe.Pointer(newData)
			oldData.Power = 10
		},
		targetDamage: func(gotTarget, gotParent, gotSource *Object, damage int32, damageType object.DamageType) int32 {
			events = append(events, "target-damage")
			if gotTarget != target || gotParent != parent || gotSource != source || damage != 5 || damageType != object.DamageType(sparkExplosionDamageType4E9AC0) {
				t.Fatalf("unexpected target damage: %p %p %p %d %d", gotTarget, gotParent, gotSource, damage, damageType)
			}
			oldData.Power = 12
			return 73
		},
		mapDamageUnits: func(pos types.Pointf, radius, inner float32, damage int32, damageType object.DamageType, gotSource, excluded *Object) {
			events = append(events, "area-damage")
			if pos != source.PosVec || radius != 4 || inner != 0 || damage != 6 ||
				damageType != object.DamageType(sparkExplosionDamageType4E9AC0) || gotSource != source || excluded != target {
				t.Fatalf("unexpected area damage: %#v %g %g %d %d %p %p", pos, radius, inner, damage, damageType, gotSource, excluded)
			}
			oldData.Power = 14
		},
		sparkFX: func(pos types.Pointf, power uint8) {
			events = append(events, "fx")
			if pos != source.PosVec || power != 14 {
				t.Fatalf("spark FX = %#v/%d, want %#v/14", pos, power, source.PosVec)
			}
		},
		audio: func(id uint32, obj *Object, kind int32, code uint32) {
			events = append(events, "audio")
			if id != sparkExplosionDetonateAudio4E9AC0 || obj != source || kind != 0 || code != 0 {
				t.Fatalf("audio = %d/%p/%d/%d", id, obj, kind, code)
			}
		},
		scorch: func(pos types.Pointf, kind int32) {
			events = append(events, "scorch")
			if pos != source.PosVec || kind != sparkExplosionScorchType4E9AC0 {
				t.Fatalf("scorch = %#v/%d", pos, kind)
			}
		},
		delayedDelete: func(got *Object) {
			events = append(events, "delete")
			if got != source {
				t.Fatalf("delete = %p, want %p", got, source)
			}
		},
	})

	want := []string{
		"quest", "push", "parent", "target-damage", "coop",
		"area-damage", "fx", "audio", "scorch", "delete",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
	if source.CollideData != unsafe.Pointer(newData) {
		t.Fatalf("source collide data = %p, want replacement %p", source.CollideData, newData)
	}
	if collision.X != 123.5 || collision.Y != -77.25 {
		t.Fatalf("collision was modified: %#v", collision)
	}
}

func TestSparkExplosionCollideNative4E9AC0ReflectPointerOrder(t *testing.T) {
	data := &SparkExplosionCollideData{Power: 7}
	source := &Object{PosVec: types.Ptf(1.25, 2.5), CollideData: unsafe.Pointer(data)}
	target := &Object{
		PosVec:     types.Ptf(-4.5, 8.75),
		Direction1: Dir16(uint16(0xfffa)),
		Buffs:      uint32(1) << sparkExplosionReflectEnchant4E9AC0,
	}
	var events []string
	sparkExplosionCollideNative4E9AC0(source, target, nil, sparkExplosionCollideNativeDeps4E9AC0{
		checkDirection: func(first types.Pointf, direction int16, second types.Pointf) int32 {
			events = append(events, "direction")
			if first != target.PosVec || direction != -6 || second != source.PosVec {
				t.Fatalf("direction args = %#v/%d/%#v", first, direction, second)
			}
			return -1
		},
		reflect: func(gotSource, gotTarget *Object) {
			events = append(events, "reflect")
			if gotSource != source || gotTarget != target {
				t.Fatalf("reflect = %p/%p", gotSource, gotTarget)
			}
		},
		clearOwner: func(got *Object) {
			events = append(events, "clear")
			if got != source {
				t.Fatalf("clear owner = %p", got)
			}
		},
		setOwner: func(owner, obj *Object) {
			events = append(events, "set")
			if owner != target || obj != source {
				t.Fatalf("set owner = %p/%p", owner, obj)
			}
		},
		audio: func(id uint32, obj *Object, kind int32, code uint32) {
			events = append(events, "audio")
			if id != sparkExplosionReflectAudio4E9AC0 || obj != target || kind != 0 || code != 0 {
				t.Fatalf("reflect audio = %d/%p/%d/%d", id, obj, kind, code)
			}
		},
		gameFlagsCheck: func(flag uint32) int32 {
			events = append(events, "quest")
			if flag != sparkExplosionQuestFlag4E9AC0 {
				t.Fatalf("flag = %#x, want %#x", flag, sparkExplosionQuestFlag4E9AC0)
			}
			return 0
		},
	})

	want := []string{"direction", "reflect", "clear", "set", "audio", "quest"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
}
