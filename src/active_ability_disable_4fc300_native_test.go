package opennox

import (
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func TestActiveAbilityDisableNative4FC300Layout(t *testing.T) {
	type layoutCheck struct {
		name string
		got  uintptr
		v32  uintptr
		v64  uintptr
	}
	checks := []layoutCheck{
		{"Object.UpdateData", unsafe.Offsetof(server.Object{}.UpdateData), 748, 872},
		{"PlayerUpdateData.HarpoonBolt", unsafe.Offsetof(server.PlayerUpdateData{}.HarpoonBolt), 136, 160},
		{"ExecAbilityClass.Abil", unsafe.Offsetof(server.ExecAbilityClass{}.Abil), 0, 0},
		{"ExecAbilityClass.Unit", unsafe.Offsetof(server.ExecAbilityClass{}.Unit), 4, 8},
		{"ExecAbilityClass.Frame", unsafe.Offsetof(server.ExecAbilityClass{}.Frame), 8, 16},
		{"ExecAbilityClass.Active", unsafe.Offsetof(server.ExecAbilityClass{}.Active), 12, 20},
		{"ExecAbilityClass.Next", unsafe.Offsetof(server.ExecAbilityClass{}.Next), 16, 24},
		{"ExecAbilityClass.Prev", unsafe.Offsetof(server.ExecAbilityClass{}.Prev), 20, 32},
		{"ExecAbilityClass size", unsafe.Sizeof(server.ExecAbilityClass{}), 24, 40},
		{"Ability width", unsafe.Sizeof(server.Ability(0)), 4, 4},
	}
	for _, check := range checks {
		want := check.v64
		if unsafe.Sizeof(uintptr(0)) == 4 {
			want = check.v32
		}
		if check.got != want {
			t.Errorf("%s = %d, want %d", check.name, check.got, want)
		}
	}
}

func TestActiveAbilityDisableNative4FC300BindsFieldsCallbacksAndPointerWidth(t *testing.T) {
	bolt := new(server.Object)
	replacementBolt := new(server.Object)
	update := &server.PlayerUpdateData{HarpoonBolt: bolt}
	unit := &server.Object{UpdateData: unsafe.Pointer(update)}
	stale := &server.ExecAbilityClass{Unit: new(server.Object), Abil: server.AbilityWarcry}
	match := &server.ExecAbilityClass{
		Unit: unit, Abil: server.AbilityHarpoon, Frame: math.MaxUint32, Active: math.MaxUint32,
	}
	head := stale
	allocator := uintptr(0x4fc300)
	var events []string
	deps := activeAbilityDisableNativeDeps4FC300{
		breakHarpoon: func(gotUnit, gotBolt *server.Object) {
			if gotUnit != unit || gotBolt != bolt {
				t.Fatalf("Harpoon break = (%p,%p), want (%p,%p)", gotUnit, gotBolt, unit, bolt)
			}
			events = append(events, "break")
			update.HarpoonBolt = replacementBolt
		},
		disableEnchant: func(*server.Object, server.EnchantID) {
			t.Fatal("Harpoon path disabled an enchant")
		},
		reportActive: func(gotUnit *server.Object, ability server.Ability, active bool) {
			if gotUnit != unit || ability != server.AbilityHarpoon || active {
				t.Fatalf("active report = (%p,%d,%v), want (%p,%d,false)", gotUnit, ability, active, unit, server.AbilityHarpoon)
			}
			events = append(events, "report")
			head = match
		},
		loadExecHead: func() *server.ExecAbilityClass {
			events = append(events, "head")
			return head
		},
		storeExecHead: func(record *server.ExecAbilityClass) {
			events = append(events, "head-store")
			head = record
		},
		loadExecAllocator: func() uintptr {
			events = append(events, "allocator")
			return allocator
		},
		freeExec: func(gotAllocator uintptr, record *server.ExecAbilityClass) {
			if gotAllocator != allocator || record != match {
				t.Fatalf("free = (%#x,%p), want (%#x,%p)", gotAllocator, record, allocator, match)
			}
			if record.Frame != math.MaxUint32 || record.Active != math.MaxUint32 {
				t.Fatalf("record was inspected or cleared before free: %+v", *record)
			}
			events = append(events, "free")
			*record = server.ExecAbilityClass{}
		},
	}

	var pin runtime.Pinner
	pin.Pin(unit)
	pin.Pin(bolt)
	defer pin.Unpin()
	if unsafe.Sizeof(uintptr(0)) == 8 {
		if uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
			t.Fatalf("unit pointer = %p, want native address above 4 GiB", unit)
		}
		if uintptr(unsafe.Pointer(bolt)) <= math.MaxUint32 {
			t.Fatalf("bolt pointer = %p, want native address above 4 GiB", bolt)
		}
	}

	activeAbilityDisableNative4FC300(unit, server.AbilityHarpoon, deps)
	if want := []string{"break", "report", "head", "head-store", "allocator", "free"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %q, want %q", events, want)
	}
	if update.HarpoonBolt != replacementBolt {
		t.Fatal("Harpoon break did not run after the bolt load")
	}
	if head != nil {
		t.Fatalf("execution head = %p, want nil", head)
	}
	if *match != (server.ExecAbilityClass{}) {
		t.Fatalf("released record = %+v, want zero", *match)
	}
	if *stale == (server.ExecAbilityClass{}) {
		t.Fatal("execution head was loaded before the inactive report")
	}
	runtime.KeepAlive(update)
	runtime.KeepAlive(bolt)
	runtime.KeepAlive(unit)
}

func TestActiveAbilityDisableNative4FC300GatesAndSpecialPaths(t *testing.T) {
	t.Run("nil unit", func(t *testing.T) {
		activeAbilityDisableNative4FC300(nil, server.AbilityHarpoon, activeAbilityDisableNativeDeps4FC300{})
	})

	for _, ability := range []server.Ability{
		server.Ability(math.MinInt32),
		server.AbilityInvalid,
		server.AbilityMax,
		server.Ability(math.MaxInt32),
	} {
		t.Run(ability.String(), func(t *testing.T) {
			activeAbilityDisableNative4FC300(new(server.Object), ability, activeAbilityDisableNativeDeps4FC300{})
		})
	}

	t.Run("Infravis", func(t *testing.T) {
		activeAbilityDisableNative4FC300(new(server.Object), server.AbilityInfravis, activeAbilityDisableNativeDeps4FC300{})
	})

	t.Run("Tread Lightly", func(t *testing.T) {
		unit := new(server.Object)
		var events []string
		activeAbilityDisableNative4FC300(unit, server.AbilityTreadLightly, activeAbilityDisableNativeDeps4FC300{
			disableEnchant: func(gotUnit *server.Object, enchant server.EnchantID) {
				if gotUnit != unit || enchant != server.ENCHANT_SNEAK {
					t.Fatalf("disable enchant = (%p,%d), want (%p,%d)", gotUnit, enchant, unit, server.ENCHANT_SNEAK)
				}
				events = append(events, "enchant")
			},
			reportActive: func(gotUnit *server.Object, ability server.Ability, active bool) {
				if gotUnit != unit || ability != server.AbilityTreadLightly || active {
					t.Fatalf("active report = (%p,%d,%v), want (%p,%d,false)", gotUnit, ability, active, unit, server.AbilityTreadLightly)
				}
				events = append(events, "report")
			},
			loadExecHead: func() *server.ExecAbilityClass {
				events = append(events, "head")
				return nil
			},
		})
		if want := []string{"enchant", "report", "head"}; !reflect.DeepEqual(events, want) {
			t.Fatalf("events = %q, want %q", events, want)
		}
	})

	t.Run("nil Harpoon UpdateData", func(t *testing.T) {
		called := false
		defer func() {
			if recover() == nil {
				t.Fatal("nil UpdateData did not fault at the HarpoonBolt load")
			}
			if called {
				t.Fatal("Harpoon callback ran after a faulting bolt load")
			}
		}()
		activeAbilityDisableNative4FC300(new(server.Object), server.AbilityHarpoon, activeAbilityDisableNativeDeps4FC300{
			breakHarpoon: func(*server.Object, *server.Object) { called = true },
		})
	})

	t.Run("nil Harpoon bolt", func(t *testing.T) {
		update := new(server.PlayerUpdateData)
		unit := &server.Object{UpdateData: unsafe.Pointer(update)}
		var events []string
		activeAbilityDisableNative4FC300(unit, server.AbilityHarpoon, activeAbilityDisableNativeDeps4FC300{
			breakHarpoon: func(gotUnit, gotBolt *server.Object) {
				if gotUnit != unit || gotBolt != nil {
					t.Fatalf("Harpoon break = (%p,%p), want (%p,nil)", gotUnit, gotBolt, unit)
				}
				events = append(events, "break")
			},
			reportActive: func(*server.Object, server.Ability, bool) {
				events = append(events, "report")
			},
			loadExecHead: func() *server.ExecAbilityClass {
				events = append(events, "head")
				return nil
			},
		})
		if want := []string{"break", "report", "head"}; !reflect.DeepEqual(events, want) {
			t.Fatalf("events = %q, want %q", events, want)
		}
		runtime.KeepAlive(update)
	})
}
