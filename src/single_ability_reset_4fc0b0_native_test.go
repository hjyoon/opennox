package opennox

import (
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	playerlib "github.com/opennox/libs/player"

	"github.com/opennox/opennox/v1/server"
)

func TestSingleAbilityResetNative4FC0B0Layout(t *testing.T) {
	type layoutCheck struct {
		name string
		got  uintptr
		v32  uintptr
		v64  uintptr
	}
	checks := []layoutCheck{
		{"Object.ObjClass", unsafe.Offsetof(server.Object{}.ObjClass), 8, 12},
		{"Object.UpdateData", unsafe.Offsetof(server.Object{}.UpdateData), 748, 872},
		{"PlayerUpdateData.Player", unsafe.Offsetof(server.PlayerUpdateData{}.Player), 276, 336},
		{"Player.PlayerInd", unsafe.Offsetof(server.Player{}.PlayerInd), 2064, 2068},
		{"ExecAbilityClass.Abil", unsafe.Offsetof(server.ExecAbilityClass{}.Abil), 0, 0},
		{"ExecAbilityClass.Unit", unsafe.Offsetof(server.ExecAbilityClass{}.Unit), 4, 8},
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

func TestSingleAbilityResetNative4FC0B0BindsNativeFieldsAndCallbacks(t *testing.T) {
	player := &server.Player{PlayerInd: 0xfe}
	player.Info().SetPlayerClass(playerlib.Warrior)
	update := &server.PlayerUpdateData{Player: player}
	unit := &server.Object{
		ObjClass:   object.ClassPlayer,
		UpdateData: unsafe.Pointer(update),
	}
	stale := &server.ExecAbilityClass{Unit: new(server.Object), Abil: server.AbilityWarcry}
	match := &server.ExecAbilityClass{
		Unit: unit, Abil: server.AbilityBerserk, Frame: math.MaxUint32, Active: math.MaxUint32,
	}
	head := stale
	allocator := uintptr(0x4fc0b0)
	var events []string
	deps := singleAbilityResetNativeDeps4FC0B0{
		storeCooldown: func(index uint8, ability server.Ability, value int32) {
			if index != player.PlayerInd || ability != server.AbilityBerserk || value != 0 {
				t.Fatalf("cooldown store = [%d][%d]=%d", index, ability, value)
			}
			events = append(events, "cooldown")
		},
		resetAbility: func(gotUnit *server.Object, ability server.Ability) {
			if gotUnit != unit || ability != server.AbilityBerserk {
				t.Fatalf("reset = (%p,%d), want (%p,%d)", gotUnit, ability, unit, server.AbilityBerserk)
			}
			events = append(events, "reset")
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
		reportActive: func(gotUnit *server.Object, ability server.Ability, active bool) {
			if gotUnit != unit || ability != server.AbilityBerserk || active {
				t.Fatalf("active report = (%p,%d,%v), want (%p,%d,false)", gotUnit, ability, active, unit, server.AbilityBerserk)
			}
			events = append(events, "report")
		},
		loadExecAllocator: func() uintptr {
			events = append(events, "allocator")
			return allocator
		},
		freeExec: func(gotAllocator uintptr, record *server.ExecAbilityClass) {
			if gotAllocator != allocator || record != match {
				t.Fatalf("free = (%#x,%p), want (%#x,%p)", gotAllocator, record, allocator, match)
			}
			events = append(events, "free")
			*record = server.ExecAbilityClass{}
		},
	}

	var pin runtime.Pinner
	pin.Pin(unit)
	defer pin.Unpin()
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
		t.Fatalf("unit pointer = %p, want native address above 4 GiB", unit)
	}

	singleAbilityResetNative4FC0B0(unit, server.AbilityBerserk, deps)
	if want := []string{"cooldown", "reset", "head", "report", "head-store", "allocator", "free"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %q, want %q", events, want)
	}
	if head != nil {
		t.Fatalf("execution head = %p, want nil", head)
	}
	if *match != (server.ExecAbilityClass{}) {
		t.Fatalf("released record = %+v, want zero", *match)
	}
	if *stale == (server.ExecAbilityClass{}) {
		t.Fatal("head was loaded before the reset callback")
	}
	runtime.KeepAlive(update)
	runtime.KeepAlive(player)
	runtime.KeepAlive(unit)
}

func TestSingleAbilityResetNative4FC0B0GatesAndPointerFaults(t *testing.T) {
	t.Run("nil unit", func(t *testing.T) {
		singleAbilityResetNative4FC0B0(nil, server.AbilityBerserk, singleAbilityResetNativeDeps4FC0B0{})
	})
	t.Run("non-player", func(t *testing.T) {
		unit := &server.Object{ObjClass: object.ClassMonster}
		singleAbilityResetNative4FC0B0(unit, server.AbilityBerserk, singleAbilityResetNativeDeps4FC0B0{})
	})
	t.Run("non-Warrior", func(t *testing.T) {
		player := new(server.Player)
		player.Info().SetPlayerClass(playerlib.Wizard)
		update := &server.PlayerUpdateData{Player: player}
		unit := &server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
		singleAbilityResetNative4FC0B0(unit, server.AbilityBerserk, singleAbilityResetNativeDeps4FC0B0{})
		runtime.KeepAlive(update)
	})
	t.Run("nil UpdateData", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("nil UpdateData did not fault at the Player load")
			}
		}()
		unit := &server.Object{ObjClass: object.ClassPlayer}
		singleAbilityResetNative4FC0B0(unit, server.AbilityBerserk, singleAbilityResetNativeDeps4FC0B0{})
	})
	t.Run("nil Player", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("nil Player did not fault at the class-byte load")
			}
		}()
		update := new(server.PlayerUpdateData)
		unit := &server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
		singleAbilityResetNative4FC0B0(unit, server.AbilityBerserk, singleAbilityResetNativeDeps4FC0B0{})
	})
}
