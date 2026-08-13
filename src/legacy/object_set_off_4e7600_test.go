package legacy

import (
	"reflect"
	"testing"
)

type objectSetOffFixture4E7600 struct {
	flags uint32
	class uint32
}

func objectSetOffTestHooks4E7600(events *[]string) objectSetOffHooks4E7600[*objectSetOffFixture4E7600] {
	return objectSetOffHooks4E7600[*objectSetOffFixture4E7600]{
		flags: func(obj *objectSetOffFixture4E7600) uint32 {
			*events = append(*events, "flags")
			return obj.flags
		},
		class: func(obj *objectSetOffFixture4E7600) uint32 {
			*events = append(*events, "class")
			return obj.class
		},
		audio: func(*objectSetOffFixture4E7600) {
			*events = append(*events, "audio:236")
		},
		setOnOff: func(obj *objectSetOffFixture4E7600, enabled bool) {
			*events = append(*events, "set-off")
			if enabled {
				obj.flags |= objectEnabledFlag4E7600
			} else {
				obj.flags &^= objectEnabledFlag4E7600
			}
		},
		setFlags: func(obj *objectSetOffFixture4E7600, flags uint32) {
			*events = append(*events, "set-flags")
			obj.flags = flags
		},
	}
}

func TestObjectSetOff4E7600InitialEnabledAndElevatorOrder(t *testing.T) {
	for _, tc := range []struct {
		name       string
		flags      uint32
		class      uint32
		wantEvents []string
	}{
		{
			name:       "enabled elevator plays sound",
			flags:      objectEnabledFlag4E7600,
			class:      objectElevatorClass4E7600,
			wantEvents: []string{"flags", "class", "audio:236", "set-off", "class"},
		},
		{
			name:       "enabled ordinary object skips sound",
			flags:      objectEnabledFlag4E7600,
			class:      0x2,
			wantEvents: []string{"flags", "class", "set-off", "class"},
		},
		{
			name:       "disabled elevator skips initial class read",
			class:      objectElevatorClass4E7600,
			wantEvents: []string{"flags", "set-off", "class"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			obj := &objectSetOffFixture4E7600{flags: tc.flags, class: tc.class}
			var events []string
			got := objectSetOff4E7600(obj, objectSetOffTestHooks4E7600(&events))
			if got != tc.class {
				t.Fatalf("result = %#x, want class %#x", got, tc.class)
			}
			if !reflect.DeepEqual(events, tc.wantEvents) {
				t.Fatalf("events = %v, want %v", events, tc.wantEvents)
			}
		})
	}
}

func TestObjectSetOff4E7600ReloadsClassThenReturnsFullFlags(t *testing.T) {
	obj := &objectSetOffFixture4E7600{
		flags: objectEnabledFlag4E7600 | 0x80000120,
		class: objectElevatorClass4E7600,
	}
	var events []string
	hooks := objectSetOffTestHooks4E7600(&events)
	hooks.audio = func(obj *objectSetOffFixture4E7600) {
		events = append(events, "audio:236")
		obj.class = 0x2
	}
	hooks.setOnOff = func(obj *objectSetOffFixture4E7600, enabled bool) {
		events = append(events, "set-off")
		obj.flags = 0x80000120
		obj.class = objectNoCollideClass4E7600 | 0xa5
	}

	got := objectSetOff4E7600(obj, hooks)
	if got != 0x80000160 {
		t.Fatalf("result = %#x, want full updated flags 0x80000160", got)
	}
	if obj.flags != got {
		t.Fatalf("stored flags = %#x, want returned value %#x", obj.flags, got)
	}
	wantEvents := []string{"flags", "class", "audio:236", "set-off", "class", "flags", "set-flags"}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
}

func TestObjectSetOff4E7600NoCollideClassMask(t *testing.T) {
	for _, class := range []uint32{0x2000, 0x40000, 0x10000000, 0x10042000} {
		obj := &objectSetOffFixture4E7600{
			flags: objectEnabledFlag4E7600 | 0x80,
			class: class,
		}
		var events []string
		got := objectSetOff4E7600(obj, objectSetOffTestHooks4E7600(&events))
		if got != 0xc0 || obj.flags != 0xc0 {
			t.Fatalf("class %#x: result/flags = (%#x, %#x), want (0xc0, 0xc0)", class, got, obj.flags)
		}
		wantEvents := []string{"flags", "class", "set-off", "class", "flags", "set-flags"}
		if !reflect.DeepEqual(events, wantEvents) {
			t.Fatalf("class %#x: events = %v, want %v", class, events, wantEvents)
		}
	}
}

func TestObjectSetOff4E7600ReturnsReloadedClassWithoutFlagsRead(t *testing.T) {
	obj := &objectSetOffFixture4E7600{class: objectElevatorClass4E7600}
	var events []string
	hooks := objectSetOffTestHooks4E7600(&events)
	hooks.setOnOff = func(obj *objectSetOffFixture4E7600, enabled bool) {
		events = append(events, "set-off")
		obj.flags = 0xffffffff
		obj.class = 0x80000002
	}

	got := objectSetOff4E7600(obj, hooks)
	if got != 0x80000002 {
		t.Fatalf("result = %#x, want reloaded class 0x80000002", got)
	}
	wantEvents := []string{"flags", "set-off", "class"}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
}

func TestObjectSetOff4E7600NilFaultsOnFirstFlagsRead(t *testing.T) {
	var events []string
	defer func() {
		if recover() == nil {
			t.Fatal("nil object returned without a panic")
		}
		if !reflect.DeepEqual(events, []string{"flags"}) {
			t.Fatalf("events = %v, want first flags read only", events)
		}
	}()
	objectSetOff4E7600[*objectSetOffFixture4E7600](nil, objectSetOffTestHooks4E7600(&events))
}
