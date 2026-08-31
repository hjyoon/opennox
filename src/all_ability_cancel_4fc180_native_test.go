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

func TestAllAbilityCancelNative4FC180Layout(t *testing.T) {
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
		{"Player class byte", 0, 0, 0},
		{"ExecAbilityClass.Abil", unsafe.Offsetof(server.ExecAbilityClass{}.Abil), 0, 0},
		{"ExecAbilityClass.Unit", unsafe.Offsetof(server.ExecAbilityClass{}.Unit), 4, 8},
		{"ExecAbilityClass.Next", unsafe.Offsetof(server.ExecAbilityClass{}.Next), 16, 24},
		{"ExecAbilityClass.Prev", unsafe.Offsetof(server.ExecAbilityClass{}.Prev), 20, 32},
		{"ExecAbilityClass size", unsafe.Sizeof(server.ExecAbilityClass{}), 24, 40},
	}
	// Player.Info starts at 2185/2189 and its class byte is 66 bytes in.
	player := new(server.Player)
	checks[4].got = uintptr(player.Info().C()) - uintptr(unsafe.Pointer(player)) + 66
	checks[4].v32 = 2251
	checks[4].v64 = 2255
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

func TestAllAbilityCancelNative4FC180BindsFieldsCallbacksAndPointerWidth(t *testing.T) {
	players := make([]*server.Player, int(server.AbilityMax)-1)
	for i := range players {
		players[i] = &server.Player{PlayerInd: uint8(i + 1)}
	}
	players[0].Info().SetPlayerClass(playerlib.Warrior)
	update := &server.PlayerUpdateData{Player: players[0]}
	unit := &server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
	otherUnit := new(server.Object)

	tail := &server.ExecAbilityClass{Unit: otherUnit, Abil: server.AbilityBerserk, Active: 0x44}
	match2 := &server.ExecAbilityClass{Unit: unit, Abil: server.Ability(-99), Frame: math.MaxUint32, Active: 0x33, Next: tail}
	match1 := &server.ExecAbilityClass{Unit: unit, Abil: server.AbilityMax, Frame: math.MaxUint32, Active: 0x22, Next: match2}
	prefix := &server.ExecAbilityClass{Unit: otherUnit, Abil: server.AbilityWarcry, Next: match1}
	match1.Prev = prefix
	match2.Prev = match1
	tail.Prev = match2
	stale := &server.ExecAbilityClass{Unit: otherUnit}
	head := stale
	allocator := uintptr(0x4fc180)
	var events []string

	deps := allAbilityCancelNativeDeps4FC180{
		storeCooldown: func(index uint8, ability server.Ability, value int32) {
			wantIndex := uint8(ability)
			if index != wantIndex || ability <= server.AbilityInvalid || ability >= server.AbilityMax || value != 0 {
				t.Fatalf("cooldown store = [%d][%d]=%d, want [%d][%d]=0", index, ability, value, wantIndex, ability)
			}
			events = append(events, "cooldown")
			if ability+1 < server.AbilityMax {
				update.Player = players[int(ability)]
			}
		},
		resetAbilities: func(gotUnit *server.Object, ability server.Ability) {
			if gotUnit != unit || ability != server.AbilityMax {
				t.Fatalf("reset = (%p,%d), want (%p,%d)", gotUnit, ability, unit, server.AbilityMax)
			}
			events = append(events, "reset")
			head = prefix
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
			if gotUnit != unit || active {
				t.Fatalf("active report = (%p,%d,%v), want unit %p inactive", gotUnit, ability, active, unit)
			}
			events = append(events, "report")
		},
		loadExecAllocator: func() uintptr {
			events = append(events, "allocator")
			return allocator
		},
		freeExec: func(gotAllocator uintptr, record *server.ExecAbilityClass) {
			if gotAllocator != allocator || (record != match1 && record != match2) {
				t.Fatalf("free = (%#x,%p), want allocator %#x and a matching record", gotAllocator, record, allocator)
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

	allAbilityCancelNative4FC180(unit, deps)
	if want := []string{
		"cooldown", "cooldown", "cooldown", "cooldown", "cooldown",
		"reset", "head", "report", "allocator", "free", "report", "allocator", "free",
	}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %q, want %q", events, want)
	}
	if head != prefix || prefix.Next != tail || tail.Prev != prefix {
		t.Fatal("all matching records were not unlinked from the native list")
	}
	if *match1 != (server.ExecAbilityClass{}) || *match2 != (server.ExecAbilityClass{}) {
		t.Fatalf("released records = %+v / %+v, want zero", *match1, *match2)
	}
	if *stale == (server.ExecAbilityClass{}) {
		t.Fatal("execution head was loaded before the aggregate reset callback")
	}
	runtime.KeepAlive(players)
	runtime.KeepAlive(update)
	runtime.KeepAlive(unit)
}

func TestAllAbilityCancelNative4FC180GatesAndPointerFaults(t *testing.T) {
	t.Run("nil unit", func(t *testing.T) {
		allAbilityCancelNative4FC180(nil, allAbilityCancelNativeDeps4FC180{})
	})
	t.Run("non-player", func(t *testing.T) {
		unit := &server.Object{ObjClass: object.ClassMonster}
		allAbilityCancelNative4FC180(unit, allAbilityCancelNativeDeps4FC180{})
	})
	t.Run("non-Warrior", func(t *testing.T) {
		player := new(server.Player)
		player.Info().SetPlayerClass(playerlib.Wizard)
		update := &server.PlayerUpdateData{Player: player}
		unit := &server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
		allAbilityCancelNative4FC180(unit, allAbilityCancelNativeDeps4FC180{})
		runtime.KeepAlive(update)
	})
	t.Run("nil UpdateData", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("nil UpdateData did not fault at the Player load")
			}
		}()
		unit := &server.Object{ObjClass: object.ClassPlayer}
		allAbilityCancelNative4FC180(unit, allAbilityCancelNativeDeps4FC180{})
	})
	t.Run("nil Player", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("nil Player did not fault at the class-byte load")
			}
		}()
		update := new(server.PlayerUpdateData)
		unit := &server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
		allAbilityCancelNative4FC180(unit, allAbilityCancelNativeDeps4FC180{})
	})
	t.Run("Player reload before second cooldown", func(t *testing.T) {
		player := &server.Player{PlayerInd: 1}
		player.Info().SetPlayerClass(playerlib.Warrior)
		update := &server.PlayerUpdateData{Player: player}
		unit := &server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
		stores := 0
		defer func() {
			if recover() == nil {
				t.Fatal("nil reloaded Player did not fault before the second cooldown store")
			}
			if stores != 1 {
				t.Fatalf("cooldown stores = %d, want one completed prefix store", stores)
			}
		}()
		allAbilityCancelNative4FC180(unit, allAbilityCancelNativeDeps4FC180{
			storeCooldown: func(uint8, server.Ability, int32) {
				stores++
				update.Player = nil
			},
		})
	})
}
