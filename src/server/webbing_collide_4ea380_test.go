package server

import (
	"reflect"
	"testing"
)

type webbingCollideTestObject4EA380 struct {
	name  string
	class uint8
}

func defaultWebbingCollideHooks4EA380() webbingCollideHooks4EA380[*webbingCollideTestObject4EA380] {
	return webbingCollideHooks4EA380[*webbingCollideTestObject4EA380]{
		audio:         func(uint32, *webbingCollideTestObject4EA380) {},
		delayedDelete: func(*webbingCollideTestObject4EA380) {},
		findParent:    func(*webbingCollideTestObject4EA380) *webbingCollideTestObject4EA380 { return nil },
		targetDamage: func(*webbingCollideTestObject4EA380, *webbingCollideTestObject4EA380, *webbingCollideTestObject4EA380, int32, uint32) int32 {
			return 0
		},
		loadClassLow:    func(obj *webbingCollideTestObject4EA380) uint8 { return obj.class },
		loadFPS:         func() uint32 { return 30 },
		applyEnchant:    func(*webbingCollideTestObject4EA380, uint32, uint32, uint32) {},
		priorityMessage: func(*webbingCollideTestObject4EA380, string) {},
	}
}

func TestWebbingCollide4EA380NilTargetReturnsBeforeSource(t *testing.T) {
	hooks := defaultWebbingCollideHooks4EA380()
	hooks.audio = func(uint32, *webbingCollideTestObject4EA380) {
		t.Fatal("nil target emitted audio")
	}
	hooks.delayedDelete = func(*webbingCollideTestObject4EA380) {
		t.Fatal("nil target deleted source")
	}
	hooks.findParent = func(*webbingCollideTestObject4EA380) *webbingCollideTestObject4EA380 {
		t.Fatal("nil target inspected source")
		return nil
	}
	webbingCollide4EA380(nil, nil, (*uint32)(nil), hooks)
}

func TestWebbingCollide4EA380DamageFailureStopsBeforeClass(t *testing.T) {
	source := &webbingCollideTestObject4EA380{name: "source"}
	target := &webbingCollideTestObject4EA380{name: "target", class: 4}
	parent := &webbingCollideTestObject4EA380{name: "parent"}
	events := make([]string, 0, 4)
	hooks := defaultWebbingCollideHooks4EA380()
	hooks.audio = func(id uint32, got *webbingCollideTestObject4EA380) {
		events = append(events, "audio")
		if id != webbingCollideAudio4EA380 || got != source {
			t.Fatalf("audio = %d/%p", id, got)
		}
	}
	hooks.delayedDelete = func(got *webbingCollideTestObject4EA380) {
		events = append(events, "delete")
		if got != source {
			t.Fatalf("delete = %p", got)
		}
	}
	hooks.findParent = func(got *webbingCollideTestObject4EA380) *webbingCollideTestObject4EA380 {
		events = append(events, "parent")
		if got != source {
			t.Fatalf("parent source = %p", got)
		}
		return parent
	}
	hooks.targetDamage = func(gotTarget, gotParent, gotSource *webbingCollideTestObject4EA380, damage int32, damageType uint32) int32 {
		events = append(events, "damage")
		if gotTarget != target || gotParent != parent || gotSource != source || damage != 0 || damageType != 2 {
			t.Fatalf("damage = %p/%p/%p/%d/%d", gotTarget, gotParent, gotSource, damage, damageType)
		}
		return 0
	}
	hooks.loadClassLow = func(*webbingCollideTestObject4EA380) uint8 {
		t.Fatal("failed damage loaded target class")
		return 0
	}
	webbingCollide4EA380(source, target, (*uint32)(nil), hooks)
	want := []string{"audio", "delete", "parent", "damage"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestWebbingCollide4EA380ReloadsClassAfterEnchant(t *testing.T) {
	source := &webbingCollideTestObject4EA380{name: "source"}
	target := &webbingCollideTestObject4EA380{name: "target", class: 2}
	events := make([]string, 0, 9)
	hooks := defaultWebbingCollideHooks4EA380()
	hooks.audio = func(uint32, *webbingCollideTestObject4EA380) {
		events = append(events, "audio")
		target.class = 0
	}
	hooks.delayedDelete = func(*webbingCollideTestObject4EA380) {
		events = append(events, "delete")
		target.class = 0
	}
	hooks.findParent = func(*webbingCollideTestObject4EA380) *webbingCollideTestObject4EA380 {
		events = append(events, "parent")
		return nil
	}
	hooks.targetDamage = func(gotTarget, gotParent, gotSource *webbingCollideTestObject4EA380, damage int32, damageType uint32) int32 {
		events = append(events, "damage")
		target.class = 2
		return -1
	}
	hooks.loadClassLow = func(got *webbingCollideTestObject4EA380) uint8 {
		events = append(events, "class")
		if got != target {
			t.Fatalf("class target = %p", got)
		}
		return got.class
	}
	hooks.loadFPS = func() uint32 {
		events = append(events, "fps")
		return 0x40004001
	}
	hooks.applyEnchant = func(got *webbingCollideTestObject4EA380, enchant, duration, power uint32) {
		events = append(events, "enchant")
		if got != target || enchant != 4 || duration != 0x10004 || power != 3 {
			t.Fatalf("enchant = %p/%d/%#x/%d", got, enchant, duration, power)
		}
		target.class = 4
	}
	hooks.priorityMessage = func(got *webbingCollideTestObject4EA380, message string) {
		events = append(events, "message")
		if got != target || message != webbingCollideMessage4EA380 {
			t.Fatalf("message = %p/%q", got, message)
		}
	}
	webbingCollide4EA380(source, target, (*uint32)(nil), hooks)
	want := []string{"audio", "delete", "parent", "damage", "class", "fps", "enchant", "class", "message"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestWebbingCollide4EA380SecondClassLoadStillRunsWithoutEnchant(t *testing.T) {
	source := &webbingCollideTestObject4EA380{name: "source"}
	target := &webbingCollideTestObject4EA380{name: "target"}
	hooks := defaultWebbingCollideHooks4EA380()
	hooks.targetDamage = func(*webbingCollideTestObject4EA380, *webbingCollideTestObject4EA380, *webbingCollideTestObject4EA380, int32, uint32) int32 {
		return 1
	}
	loads := 0
	hooks.loadClassLow = func(*webbingCollideTestObject4EA380) uint8 {
		loads++
		if loads == 1 {
			return 0
		}
		return 4
	}
	hooks.loadFPS = func() uint32 {
		t.Fatal("class zero loaded FPS")
		return 0
	}
	hooks.applyEnchant = func(*webbingCollideTestObject4EA380, uint32, uint32, uint32) {
		t.Fatal("class zero applied enchant")
	}
	messaged := false
	hooks.priorityMessage = func(*webbingCollideTestObject4EA380, string) {
		messaged = true
	}
	webbingCollide4EA380(source, target, (*uint32)(nil), hooks)
	if loads != 2 || !messaged {
		t.Fatalf("loads/message = %d/%t", loads, messaged)
	}
}
