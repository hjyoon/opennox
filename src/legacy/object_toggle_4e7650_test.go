package legacy

import (
	"reflect"
	"testing"
)

func TestObjectToggle4E7650EnabledCallsSetOff(t *testing.T) {
	var events []string
	result, wasEnabled := objectToggle4E7650(7, objectToggleHooks4E7650[int]{
		flags: func(obj int) uint32 {
			events = append(events, "flags")
			if obj != 7 {
				t.Fatalf("flags object = %d, want 7", obj)
			}
			return 0xa5010000
		},
		setOff: func(obj int) uint32 {
			events = append(events, "off")
			return 0x89abcdef
		},
		setOn: func(obj int) byte {
			t.Fatal("setOn called for enabled object")
			return 0
		},
	})
	if result != 0xef {
		t.Fatalf("result = %#x, want low byte %#x", result, byte(0xef))
	}
	if !wasEnabled {
		t.Fatal("wasEnabled = false, want true")
	}
	if want := []string{"flags", "off"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestObjectToggle4E7650DisabledCallsSetOn(t *testing.T) {
	var events []string
	result, wasEnabled := objectToggle4E7650(9, objectToggleHooks4E7650[int]{
		flags: func(obj int) uint32 {
			events = append(events, "flags")
			return 0xfeffffff &^ objectEnabledFlag4E7650
		},
		setOff: func(obj int) uint32 {
			t.Fatal("setOff called for disabled object")
			return 0
		},
		setOn: func(obj int) byte {
			events = append(events, "on")
			return 0xa5
		},
	})
	if result != 0xa5 {
		t.Fatalf("result = %#x, want %#x", result, byte(0xa5))
	}
	if wasEnabled {
		t.Fatal("wasEnabled = true, want false")
	}
	if want := []string{"flags", "on"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestObjectToggle4E7650ReadsFlagsOnce(t *testing.T) {
	loads := 0
	result, wasEnabled := objectToggle4E7650(1, objectToggleHooks4E7650[int]{
		flags: func(int) uint32 {
			loads++
			if loads == 1 {
				return objectEnabledFlag4E7650
			}
			return 0
		},
		setOff: func(int) uint32 { return 0x10203040 },
		setOn:  func(int) byte { return 0xff },
	})
	if loads != 1 {
		t.Fatalf("flags loads = %d, want 1", loads)
	}
	if result != 0x40 || !wasEnabled {
		t.Fatalf("result, wasEnabled = %#x, %v; want %#x, true", result, wasEnabled, byte(0x40))
	}
}

func TestObjectToggle4E7650NilFaultsBeforeDispatch(t *testing.T) {
	type toggleObject struct {
		flags uint32
	}
	var obj *toggleObject
	dispatched := false
	defer func() {
		if recover() == nil {
			t.Fatal("nil object did not fault")
		}
		if dispatched {
			t.Fatal("state transition ran before nil flags fault")
		}
	}()
	objectToggle4E7650(obj, objectToggleHooks4E7650[*toggleObject]{
		flags: func(obj *toggleObject) uint32 { return obj.flags },
		setOff: func(*toggleObject) uint32 {
			dispatched = true
			return 0
		},
		setOn: func(*toggleObject) byte {
			dispatched = true
			return 0
		},
	})
}
