package opennox

import (
	"reflect"
	"runtime"
	"testing"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestSpellRuntimeCleanupNative4FCA80PreservesPointersAndOrder(t *testing.T) {
	allocatorClass := new(alloc.Class)
	allocator := alloc.ClassT[server.MagicEntityClass]{Class: allocatorClass}
	caster := new(server.Object)
	var events []string

	got := spellRuntimeCleanupNative4FCA80(spellRuntimeCleanupNativeDeps4FCA80{
		freeDurations: func() {
			events = append(events, "free-durations")
		},
		loadMagicClass: func() alloc.ClassT[server.MagicEntityClass] {
			events = append(events, "load-magic")
			return allocator
		},
		freeMagicClass: func(value alloc.ClassT[server.MagicEntityClass]) {
			events = append(events, "free-magic")
			if value.Class != allocatorClass {
				t.Fatalf("allocator pointer = %p, want exact %p", value.Class, allocatorClass)
			}
		},
		loadImaginaryCaster: func() *server.Object {
			events = append(events, "load-caster")
			return caster
		},
		clearMagicEntityHead: func() {
			events = append(events, "clear-head")
		},
		delayedDelete: func(value *server.Object) {
			events = append(events, "delayed-delete")
			if value != caster {
				t.Fatalf("caster pointer = %p, want exact %p", value, caster)
			}
		},
		clearImaginaryCaster: func() {
			events = append(events, "clear-caster")
		},
	})

	if got != 1 {
		t.Fatalf("result = %d, want canonical 1", got)
	}
	want := []string{
		"free-durations",
		"load-magic",
		"free-magic",
		"load-caster",
		"clear-head",
		"delayed-delete",
		"clear-caster",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
	runtime.KeepAlive(allocatorClass)
	runtime.KeepAlive(caster)
}

func TestSpellRuntimeCleanupNative4FCA80ForwardsNilPointers(t *testing.T) {
	allocator := alloc.ClassT[server.MagicEntityClass]{}
	magicFreed := false
	deleteCalled := false

	got := spellRuntimeCleanupNative4FCA80(spellRuntimeCleanupNativeDeps4FCA80{
		freeDurations: func() {},
		loadMagicClass: func() alloc.ClassT[server.MagicEntityClass] {
			return allocator
		},
		freeMagicClass: func(value alloc.ClassT[server.MagicEntityClass]) {
			magicFreed = true
			if value.Class != nil {
				t.Fatalf("allocator pointer = %p, want nil", value.Class)
			}
		},
		loadImaginaryCaster:  func() *server.Object { return nil },
		clearMagicEntityHead: func() {},
		delayedDelete: func(value *server.Object) {
			deleteCalled = true
			if value != nil {
				t.Fatalf("caster pointer = %p, want nil", value)
			}
		},
		clearImaginaryCaster: func() {},
	})

	if got != 1 || !magicFreed || !deleteCalled {
		t.Fatalf("result/free/delete = (%d, %t, %t), want (1, true, true)", got, magicFreed, deleteCalled)
	}
}
