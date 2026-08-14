package server

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func defaultWebbingCollideNativeDeps4EA380() webbingCollideNativeDeps4EA380 {
	return webbingCollideNativeDeps4EA380{
		audio:           func(uint32, *Object) {},
		delayedDelete:   func(*Object) {},
		findParent:      func(*Object) *Object { return nil },
		targetDamage:    func(*Object, *Object, *Object, int32, object.DamageType) int32 { return 0 },
		loadFPS:         func() uint32 { return 30 },
		applyEnchant:    func(*Object, EnchantID, uint32, uint32) {},
		priorityMessage: func(*Object, string) {},
	}
}

func TestWebbingCollide4EA380NativeLayout(t *testing.T) {
	wantClass := uintptr(8)
	wantDamage := uintptr(716)
	wantSize := uintptr(780)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantClass = 12
		wantDamage = 808
		wantSize = 928
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantSize},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantClass},
		{"Object.Damage", unsafe.Offsetof(Object{}.Damage), wantDamage},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestWebbingCollideNative4EA380UsesLiveFieldsAndExactOrder(t *testing.T) {
	source := &Object{}
	target := &Object{}
	parent := &Object{}
	collision := &types.Pointf{X: 3.5, Y: -8.25}
	markerBefore := byte(1)
	markerAfter := byte(2)
	target.Damage = unsafe.Pointer(&markerBefore)
	events := make([]string, 0, 7)
	deps := defaultWebbingCollideNativeDeps4EA380()
	deps.audio = func(id uint32, got *Object) {
		events = append(events, "audio")
		if id != 351 || got != source {
			t.Fatalf("audio = %d/%p", id, got)
		}
	}
	deps.delayedDelete = func(got *Object) {
		events = append(events, "delete")
		if got != source {
			t.Fatalf("delete = %p", got)
		}
	}
	deps.findParent = func(got *Object) *Object {
		events = append(events, "parent")
		if got != source {
			t.Fatalf("parent source = %p", got)
		}
		target.Damage = unsafe.Pointer(&markerAfter)
		return parent
	}
	deps.targetDamage = func(gotTarget, gotParent, gotSource *Object, damage int32, damageType object.DamageType) int32 {
		events = append(events, "damage")
		if gotTarget != target || gotParent != parent || gotSource != source || damage != 0 || damageType != object.DamageType(2) {
			t.Fatalf("damage = %p/%p/%p/%d/%d", gotTarget, gotParent, gotSource, damage, damageType)
		}
		if gotTarget.Damage != unsafe.Pointer(&markerAfter) {
			t.Fatal("Damage field was not observed after parent lookup")
		}
		target.ObjClass = object.ClassMonster
		return -1
	}
	deps.loadFPS = func() uint32 {
		events = append(events, "fps")
		return 0x40004001
	}
	deps.applyEnchant = func(got *Object, enchant EnchantID, duration, power uint32) {
		events = append(events, "enchant")
		if got != target || enchant != ENCHANT_SLOWED || duration != 0x10004 || power != 3 {
			t.Fatalf("enchant = %p/%d/%#x/%d", got, enchant, duration, power)
		}
		target.ObjClass = object.ClassPlayer
	}
	deps.priorityMessage = func(got *Object, message string) {
		events = append(events, "message")
		if got != target || message != "objcoll.c:WebbingSlow" {
			t.Fatalf("message = %p/%q", got, message)
		}
	}
	webbingCollideNative4EA380(source, target, collision, deps)
	want := []string{"audio", "delete", "parent", "damage", "fps", "enchant", "message"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestWebbingCollide4EA380ServerBindingNilTarget(t *testing.T) {
	(&Server{}).WebbingCollide4EA380(nil, nil, nil, WebbingCollideRuntime4EA380{})
}
