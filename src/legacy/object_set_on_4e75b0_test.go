package legacy

import (
	"reflect"
	"testing"
)

type objectSetOnFixture4E75B0 struct {
	flags uint32
	class uint32
}

func objectSetOnTestHooks4E75B0(events *[]string, helperResult byte) objectSetOnHooks4E75B0[*objectSetOnFixture4E75B0] {
	return objectSetOnHooks4E75B0[*objectSetOnFixture4E75B0]{
		flags: func(obj *objectSetOnFixture4E75B0) uint32 {
			*events = append(*events, "flags")
			return obj.flags
		},
		class: func(obj *objectSetOnFixture4E75B0) uint32 {
			*events = append(*events, "class")
			return obj.class
		},
		audio: func(*objectSetOnFixture4E75B0) {
			*events = append(*events, "audio:235")
		},
		setOnOff: func(obj *objectSetOnFixture4E75B0, enabled bool) {
			*events = append(*events, "set-on")
			if enabled {
				obj.flags |= objectEnabledFlag4E75B0
			}
		},
		clearFlags: func(obj *objectSetOnFixture4E75B0, flags uint32) {
			*events = append(*events, "clear:0x40")
			obj.flags &^= flags
		},
		hasCollideOrUpdate: func(*objectSetOnFixture4E75B0) byte {
			*events = append(*events, "collide-or-update")
			return helperResult
		},
	}
}

func TestObjectSetOn4E75B0InitialEnabledAndElevatorOrder(t *testing.T) {
	for _, tc := range []struct {
		name       string
		flags      uint32
		class      uint32
		wantEvents []string
	}{
		{
			name:       "disabled elevator plays sound",
			class:      objectElevatorClass4E75B0,
			wantEvents: []string{"flags", "class", "audio:235", "set-on", "class", "collide-or-update"},
		},
		{
			name:       "disabled ordinary object skips sound",
			class:      0x2,
			wantEvents: []string{"flags", "class", "set-on", "class", "collide-or-update"},
		},
		{
			name:       "enabled elevator skips initial class read",
			flags:      objectEnabledFlag4E75B0,
			class:      objectElevatorClass4E75B0,
			wantEvents: []string{"flags", "set-on", "class", "collide-or-update"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			obj := &objectSetOnFixture4E75B0{flags: tc.flags, class: tc.class}
			var events []string
			got := objectSetOn4E75B0(obj, objectSetOnTestHooks4E75B0(&events, 0xfe))
			if got != 0xfe {
				t.Fatalf("result = %#x, want 0xfe", got)
			}
			if !reflect.DeepEqual(events, tc.wantEvents) {
				t.Fatalf("events = %v, want %v", events, tc.wantEvents)
			}
		})
	}
}

func TestObjectSetOn4E75B0ReloadsThenCachesClass(t *testing.T) {
	obj := &objectSetOnFixture4E75B0{
		flags: objectNoCollideFlag4E75B0 | 0x80000000,
		class: objectElevatorClass4E75B0,
	}
	var events []string
	hooks := objectSetOnTestHooks4E75B0(&events, 0xee)
	hooks.audio = func(obj *objectSetOnFixture4E75B0) {
		events = append(events, "audio:235")
		obj.class = 0x20000002
	}
	hooks.setOnOff = func(obj *objectSetOnFixture4E75B0, enabled bool) {
		events = append(events, "set-on")
		obj.flags |= objectEnabledFlag4E75B0
		obj.class = objectClearClass4E75B0 | 0xa5
	}
	originalClear := hooks.clearFlags
	hooks.clearFlags = func(obj *objectSetOnFixture4E75B0, flags uint32) {
		originalClear(obj, flags)
		obj.class = 0x2
	}
	hooks.hasCollideOrUpdate = func(*objectSetOnFixture4E75B0) byte {
		t.Fatal("cached Missile class called collide/update helper")
		return 0
	}

	got := objectSetOn4E75B0(obj, hooks)
	if got != 0xa5 {
		t.Fatalf("result = %#x, want cached class byte 0xa5", got)
	}
	wantEvents := []string{"flags", "class", "audio:235", "set-on", "class", "clear:0x40"}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
	if obj.flags != objectEnabledFlag4E75B0|0x80000000 {
		t.Fatalf("flags = %#x, want enabled and neighboring high bit only", obj.flags)
	}
}

func TestObjectSetOn4E75B0ClearClassMask(t *testing.T) {
	for _, class := range []uint32{0x2000, 0x40000, 0x10000000, 0x10042000} {
		t.Run("class", func(t *testing.T) {
			obj := &objectSetOnFixture4E75B0{
				flags: objectEnabledFlag4E75B0 | objectNoCollideFlag4E75B0 | 0x80,
				class: class,
			}
			var events []string
			got := objectSetOn4E75B0(obj, objectSetOnTestHooks4E75B0(&events, 0x7f))
			if got != 0x7f {
				t.Fatalf("result = %#x, want helper result 0x7f", got)
			}
			if obj.flags != objectEnabledFlag4E75B0|0x80 {
				t.Fatalf("flags = %#x, NoCollide was not cleared independently", obj.flags)
			}
			wantEvents := []string{"flags", "set-on", "class", "clear:0x40", "collide-or-update"}
			if !reflect.DeepEqual(events, wantEvents) {
				t.Fatalf("events = %v, want %v", events, wantEvents)
			}
		})
	}
}

func TestObjectSetOn4E75B0NilFaultsOnFirstFlagsRead(t *testing.T) {
	var events []string
	defer func() {
		if recover() == nil {
			t.Fatal("nil object returned without a panic")
		}
		if !reflect.DeepEqual(events, []string{"flags"}) {
			t.Fatalf("events = %v, want first flags read only", events)
		}
	}()
	objectSetOn4E75B0[*objectSetOnFixture4E75B0](nil, objectSetOnTestHooks4E75B0(&events, 0))
}
