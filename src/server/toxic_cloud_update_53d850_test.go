package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestToxicCloudUpdate53D850PreservesCachedDataAndReloadsOwner(t *testing.T) {
	oldOwner := new(Object)
	newOwner := new(Object)
	oldData := &ToxicCloudUpdateData{Duration: 2}
	newData := &ToxicCloudUpdateData{Duration: 99}
	cloud := &Object{
		Field34:    99,
		PosVec:     types.Ptf(10, 20),
		ObjOwner:   oldOwner,
		UpdateData: unsafe.Pointer(oldData),
	}
	blocked := &Object{ObjClass: object.ClassPlayer, PosVec: types.Ptf(30, 40)}
	target := &Object{ObjClass: object.ClassMonster, PosVec: types.Ptf(50, 60)}
	nonUnit := &Object{ObjClass: object.ClassImmobile, PosVec: types.Ptf(70, 80)}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(cloud)) <= math.MaxUint32 {
		t.Fatalf("cloud pointer = %p, want address above the ABI32 range", cloud)
	}

	var events []string
	frameCalls := 0
	deps := toxicCloudUpdateDeps53D850{
		frame: func() uint32 {
			frameCalls++
			frame := uint32(99 + frameCalls)
			events = append(events, fmt.Sprintf("frame:%d", frame))
			return frame
		},
		eachInCircle: func(pos types.Pointf, radius float32, visit func(*Object) bool) {
			events = append(events, fmt.Sprintf("circle:%v:%.0f", pos, radius))
			if !visit(nonUnit) || !visit(blocked) || !visit(target) {
				t.Fatal("circle visitor stopped")
			}
		},
		randomInt: func(minimum, maximum int) int {
			events = append(events, fmt.Sprintf("random:%d:%d", minimum, maximum))
			switch {
			case minimum == 3 && maximum == 10:
				return 7
			case minimum == 5 && maximum == 10:
				return 6
			default:
				t.Fatalf("unexpected random range %d..%d", minimum, maximum)
				return 0
			}
		},
		traceRay: func(from, to types.Pointf, flags MapTraceFlags) bool {
			events = append(events, fmt.Sprintf("trace:%v:%v:%d", from, to, flags))
			if from != cloud.PosVec || flags != MapTraceFlags(5) {
				t.Fatalf("trace = %v -> %v flags %d", from, to, flags)
			}
			return to == target.PosVec
		},
		findOwnerPlayer: func(got *Object) *Object {
			if got != cloud {
				t.Fatalf("owner source = %p, want %p", got, cloud)
			}
			if got.ObjOwner == oldOwner {
				events = append(events, "owner:old")
			} else if got.ObjOwner == newOwner {
				events = append(events, "owner:new")
			} else {
				t.Fatalf("unexpected owner %p", got.ObjOwner)
			}
			return got.ObjOwner
		},
		damage: func(gotTarget, source, weapon *Object, damage int32, damageType object.DamageType) {
			events = append(events, fmt.Sprintf("damage:%d:%d", damage, damageType))
			if gotTarget != target || source != oldOwner || weapon != cloud || damage != 7 || damageType != object.DamagePoison {
				t.Fatalf("damage = %p/%p/%p/%d/%d", gotTarget, source, weapon, damage, damageType)
			}
			cloud.ObjOwner = newOwner
			cloud.UpdateData = unsafe.Pointer(newData)
		},
		isEnemy: func(source, gotTarget *Object) bool {
			events = append(events, "enemy")
			if source != newOwner || gotTarget != target {
				t.Fatalf("enemy = %p/%p, want %p/%p", source, gotTarget, newOwner, target)
			}
			return true
		},
		activatePoison: func(got *Object, increment, maximum int32) int32 {
			events = append(events, "poison")
			if got != target || increment != 1 || maximum != 1 {
				t.Fatalf("poison = %p/%d/%d", got, increment, maximum)
			}
			return 1
		},
		delayedDelete: func(*Object) { t.Fatal("live cloud was deleted") },
	}

	toxicCloudUpdateNative53D850(cloud, toxicCloudRadius53D850, true, deps)
	wantEvents := []string{
		"frame:100",
		"circle:{10 20}:75",
		"trace:{10 20}:{30 40}:5",
		"trace:{10 20}:{50 60}:5",
		"random:3:10",
		"owner:old",
		fmt.Sprintf("damage:7:%d", object.DamagePoison),
		"owner:new",
		"enemy",
		"poison",
		"random:5:10",
		"frame:101",
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events =\n%q\nwant\n%q", events, wantEvents)
	}
	if cloud.Field34 != 107 {
		t.Fatalf("next frame = %d, want 107", cloud.Field34)
	}
	if oldData.Duration != 1 || newData.Duration != 99 {
		t.Fatalf("durations = old %d new %d, want 1/99", oldData.Duration, newData.Duration)
	}
}

func TestSmallToxicCloudUpdate53D960UsesZeroDamageAndDeletesAtZero(t *testing.T) {
	data := &ToxicCloudUpdateData{Duration: 1}
	cloud := &Object{Field34: 49, PosVec: types.Ptf(1, 2), UpdateData: unsafe.Pointer(data)}
	target := &Object{ObjClass: object.ClassPlayer, PosVec: types.Ptf(3, 4)}
	frame := uint32(50)
	deleted := 0
	poisoned := 0

	toxicCloudUpdateNative53D850(cloud, smallToxicCloudRadius53D960, false, toxicCloudUpdateDeps53D850{
		frame: func() uint32 {
			value := frame
			frame++
			return value
		},
		eachInCircle: func(pos types.Pointf, radius float32, visit func(*Object) bool) {
			if pos != cloud.PosVec || radius != 35 || !visit(target) {
				t.Fatalf("circle = %v/%.1f", pos, radius)
			}
		},
		randomInt: func(minimum, maximum int) int {
			if minimum != 5 || maximum != 10 {
				t.Fatalf("small cloud requested damage random %d..%d", minimum, maximum)
			}
			return 7
		},
		traceRay: func(from, to types.Pointf, flags MapTraceFlags) bool {
			return from == cloud.PosVec && to == target.PosVec && flags == MapTraceFlags(5)
		},
		findOwnerPlayer: func(*Object) *Object { return nil },
		damage: func(gotTarget, source, weapon *Object, damage int32, damageType object.DamageType) {
			if gotTarget != target || source != nil || weapon != cloud || damage != 0 || damageType != object.DamagePoison {
				t.Fatalf("damage = %p/%p/%p/%d/%d", gotTarget, source, weapon, damage, damageType)
			}
		},
		isEnemy: func(*Object, *Object) bool { return false },
		activatePoison: func(*Object, int32, int32) int32 {
			poisoned++
			return 1
		},
		delayedDelete: func(got *Object) {
			if got != cloud {
				t.Fatalf("deleted = %p, want %p", got, cloud)
			}
			deleted++
		},
	})
	if cloud.Field34 != 58 || data.Duration != 0 || deleted != 1 || poisoned != 0 {
		t.Fatalf("state = next %d duration %d deleted %d poisoned %d, want 58/0/1/0",
			cloud.Field34, data.Duration, deleted, poisoned)
	}
}

func TestToxicCloudUpdate53D850StrictFrameBoundaryAndSignedWrap(t *testing.T) {
	data := &ToxicCloudUpdateData{Duration: math.MinInt32}
	cloud := &Object{Field34: 100, UpdateData: unsafe.Pointer(data)}
	deleted := 0
	toxicCloudUpdateNative53D850(cloud, toxicCloudRadius53D850, true, toxicCloudUpdateDeps53D850{
		frame: func() uint32 { return 100 },
		eachInCircle: func(types.Pointf, float32, func(*Object) bool) {
			t.Fatal("Field34 == frame must not scan")
		},
		randomInt:     func(int, int) int { t.Fatal("Field34 == frame must not draw random"); return 0 },
		delayedDelete: func(*Object) { deleted++ },
	})
	if data.Duration != math.MaxInt32 || deleted != 0 || cloud.Field34 != 100 {
		t.Fatalf("state = duration %d deleted %d frame %d, want %d/0/100",
			data.Duration, deleted, cloud.Field34, math.MaxInt32)
	}
}
