package legacy

import (
	"fmt"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/player"

	"github.com/opennox/opennox/v1/server"
)

func TestWeaponRoll10001EE0ForwardHighPointersAndCandidateOrder(t *testing.T) {
	const (
		owner       = uint64(0x7f38d5b7b3d0)
		ownerAlias  = uint64(0x00000000d5b7b3d0)
		update      = uint64(0x6f38e5c8c4e0)
		playerOne   = uint64(0x5f38f6d9d5f0)
		playerTwo   = uint64(0x4f3807eae600)
		current     = uint64(0x7f38d5b7a110)
		zeroFlags   = uint64(0x7f38d5b7a220)
		quiver      = uint64(0x7f38d5b7a330)
		classDenied = uint64(0x7f38d5b7a440)
		weak        = uint64(0x7f38d5b7a550)
		target      = uint64(0x7f38d5b7a660)
		unvisited   = uint64(0x7f38d5b7a770)
	)

	var events []string
	record := func(format string, args ...any) {
		events = append(events, fmt.Sprintf(format, args...))
	}
	next := map[uint64]uint64{
		current:     zeroFlags,
		zeroFlags:   quiver,
		quiver:      classDenied,
		classDenied: weak,
		weak:        target,
		target:      unvisited,
	}
	flags := map[uint64]uint32{
		zeroFlags:   0,
		quiver:      weaponRollQuiverFlag10001EE0,
		classDenied: 4,
		weak:        8,
		target:      16,
	}
	activePlayer := playerOne

	hooks := weaponRollHooks10001EE0[uint64, uint64, uint64]{
		loadUpdate: func(got uint64) uint64 {
			record("update:%x", got)
			if got == ownerAlias {
				t.Fatal("owner pointer was narrowed to its low-32-bit alias")
			}
			if got != owner {
				t.Fatalf("update owner = %#x, want %#x", got, owner)
			}
			return update
		},
		loadPlayer: func(got uint64) uint64 {
			record("player:%x=%x", got, activePlayer)
			if got != update {
				t.Fatalf("player update = %#x, want %#x", got, update)
			}
			return activePlayer
		},
		loadPlayerStatus: func(got uint64) uint32 {
			record("status:%x", got)
			return 0
		},
		loadState: func(got uint64) server.PlayerState {
			record("state:%x", got)
			return server.PlayerState2
		},
		loadEquipped: func(got uint64) uint64 {
			record("equipped:%x=%x", got, current)
			return current
		},
		loadFirstItem: func(uint64) uint64 {
			panic("equipped path loaded InvFirstItem")
		},
		loadNextItem: func(got uint64) uint64 {
			value := next[got]
			record("next:%x=%x", got, value)
			return value
		},
		loadPrevItem: func(uint64) uint64 {
			panic("forward path loaded Field125")
		},
		loadWeaponFlags: func(got uint64) uint32 {
			value := flags[got]
			record("flags:%x=%x", got, value)
			return value
		},
		loadPlayerClass: func(got uint64) uint8 {
			value := map[uint64]uint8{playerOne: 1, playerTwo: 2}[got]
			record("class:%x=%d", got, value)
			return value
		},
		classCanUse: func(got uint64, class uint8) bool {
			value := got != classDenied
			record("can-use:%x:%d=%t", got, class, value)
			if got == classDenied {
				activePlayer = playerTwo
			}
			return value
		},
		checkStrength: func(gotOwner, gotItem uint64) bool {
			value := gotItem == target
			record("strength:%x:%x=%t", gotOwner, gotItem, value)
			return value
		},
		tryDequip: func(gotOwner, gotItem uint64) bool {
			record("dequip:%x:%x", gotOwner, gotItem)
			return true
		},
		tryEquip: func(gotOwner, gotItem uint64) bool {
			record("equip:%x:%x", gotOwner, gotItem)
			return true
		},
	}

	if got := weaponRoll10001EE0(owner, 1, hooks); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	want := []string{
		"update:7f38d5b7b3d0",
		"player:6f38e5c8c4e0=5f38f6d9d5f0",
		"status:5f38f6d9d5f0",
		"state:6f38e5c8c4e0",
		"equipped:6f38e5c8c4e0=7f38d5b7a110",
		"next:7f38d5b7a110=7f38d5b7a220", "flags:7f38d5b7a220=0",
		"next:7f38d5b7a220=7f38d5b7a330", "flags:7f38d5b7a330=2",
		"next:7f38d5b7a330=7f38d5b7a440", "flags:7f38d5b7a440=4",
		"update:7f38d5b7b3d0", "player:6f38e5c8c4e0=5f38f6d9d5f0", "class:5f38f6d9d5f0=1",
		"can-use:7f38d5b7a440:1=false",
		"next:7f38d5b7a440=7f38d5b7a550", "flags:7f38d5b7a550=8",
		"update:7f38d5b7b3d0", "player:6f38e5c8c4e0=4f3807eae600", "class:4f3807eae600=2",
		"can-use:7f38d5b7a550:2=true", "strength:7f38d5b7b3d0:7f38d5b7a550=false",
		"next:7f38d5b7a550=7f38d5b7a660", "flags:7f38d5b7a660=10",
		"update:7f38d5b7b3d0", "player:6f38e5c8c4e0=4f3807eae600", "class:4f3807eae600=2",
		"can-use:7f38d5b7a660:2=true", "strength:7f38d5b7b3d0:7f38d5b7a660=true",
		"dequip:7f38d5b7b3d0:7f38d5b7a110", "equip:7f38d5b7b3d0:7f38d5b7a660",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events:\n got: %#v\nwant: %#v", events, want)
	}
}

func TestWeaponRoll10001EE0GateShortCircuit(t *testing.T) {
	tests := []struct {
		name   string
		status uint32
		state  server.PlayerState
		want   []string
	}{
		{name: "player status", status: 2, state: server.PlayerState2, want: []string{"update", "player", "status"}},
		{name: "update state", status: 0, state: server.PlayerState1, want: []string{"update", "player", "status", "state"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var events []string
			hooks := weaponRollHooks10001EE0[uint64, uint64, uint64]{
				loadUpdate: func(uint64) uint64 { events = append(events, "update"); return 2 },
				loadPlayer: func(uint64) uint64 { events = append(events, "player"); return 3 },
				loadPlayerStatus: func(uint64) uint32 {
					events = append(events, "status")
					return tc.status
				},
				loadState: func(uint64) server.PlayerState {
					events = append(events, "state")
					return tc.state
				},
				loadEquipped: func(uint64) uint64 { panic("gate loaded equipped weapon") },
			}
			if got := weaponRoll10001EE0(uint64(1), 1, hooks); got != 0 {
				t.Fatalf("result = %d, want 0", got)
			}
			if !reflect.DeepEqual(events, tc.want) {
				t.Fatalf("events = %v, want %v", events, tc.want)
			}
		})
	}
}

func TestWeaponRoll10001EE0ReverseDequipFailureStopsBeforeEquip(t *testing.T) {
	const owner, update, playerData, current, target = uint64(11), uint64(12), uint64(13), uint64(14), uint64(15)
	var events []string
	record := func(event string) { events = append(events, event) }
	hooks := weaponRollHooks10001EE0[uint64, uint64, uint64]{
		loadUpdate:       func(uint64) uint64 { return update },
		loadPlayer:       func(uint64) uint64 { return playerData },
		loadPlayerStatus: func(uint64) uint32 { return 0 },
		loadState:        func(uint64) server.PlayerState { return server.PlayerState2 },
		loadEquipped:     func(uint64) uint64 { return current },
		loadFirstItem:    func(uint64) uint64 { panic("loaded first item") },
		loadNextItem:     func(uint64) uint64 { panic("reverse path loaded next item") },
		loadPrevItem: func(got uint64) uint64 {
			record(fmt.Sprintf("prev:%d", got))
			return target
		},
		loadWeaponFlags: func(uint64) uint32 { record("flags"); return 4 },
		loadPlayerClass: func(uint64) uint8 { record("class"); return 0 },
		classCanUse:     func(uint64, uint8) bool { record("can-use"); return true },
		checkStrength:   func(uint64, uint64) bool { record("strength"); return true },
		tryDequip: func(gotOwner, gotItem uint64) bool {
			record(fmt.Sprintf("dequip:%d:%d", gotOwner, gotItem))
			return false
		},
		tryEquip: func(uint64, uint64) bool { panic("equip ran after failed dequip") },
	}
	if got := weaponRoll10001EE0(owner, 0, hooks); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	want := []string{"prev:14", "flags", "class", "can-use", "strength", "dequip:11:14"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestWeaponRoll10001EE0UnequippedIgnoresDirectionAndStopsAfterEquip(t *testing.T) {
	const owner, update, playerData, quiver, target, unvisited = uint64(21), uint64(22), uint64(23), uint64(24), uint64(25), uint64(26)
	var events []string
	next := map[uint64]uint64{quiver: target, target: unvisited}
	record := func(event string) { events = append(events, event) }
	hooks := weaponRollHooks10001EE0[uint64, uint64, uint64]{
		loadUpdate:       func(uint64) uint64 { return update },
		loadPlayer:       func(uint64) uint64 { return playerData },
		loadPlayerStatus: func(uint64) uint32 { return 0 },
		loadState:        func(uint64) server.PlayerState { return server.PlayerState2 },
		loadEquipped:     func(uint64) uint64 { return 0 },
		loadFirstItem: func(got uint64) uint64 {
			record(fmt.Sprintf("first:%d", got))
			return quiver
		},
		loadNextItem: func(got uint64) uint64 {
			record(fmt.Sprintf("next:%d", got))
			return next[got]
		},
		loadPrevItem: func(uint64) uint64 { panic("unequipped path used direction link") },
		loadWeaponFlags: func(got uint64) uint32 {
			record(fmt.Sprintf("flags:%d", got))
			if got == quiver {
				return weaponRollQuiverFlag10001EE0
			}
			return 4
		},
		loadPlayerClass: func(uint64) uint8 { record("class"); return 0 },
		classCanUse:     func(uint64, uint8) bool { record("can-use"); return true },
		checkStrength:   func(uint64, uint64) bool { record("strength"); return true },
		tryDequip:       func(uint64, uint64) bool { panic("unequipped path tried dequip") },
		tryEquip: func(gotOwner, gotItem uint64) bool {
			record(fmt.Sprintf("equip:%d:%d", gotOwner, gotItem))
			return false
		},
	}
	if got := weaponRoll10001EE0(owner, -7, hooks); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	want := []string{"first:21", "flags:24", "next:24", "flags:25", "class", "can-use", "strength", "equip:21:25"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestWeaponRollNative10001EE0KeepsNativeObjectLinks(t *testing.T) {
	playerData := &server.Player{Field3680: 0}
	playerData.Info().SetPlayerClass(player.Warrior)
	update := &server.PlayerUpdateData{State: server.PlayerState2, Player: playerData}
	owner := &server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
	current := &server.Object{}
	target := &server.Object{}
	update.EquippedWeapon = current
	current.Field125 = target

	if unsafe.Sizeof(uintptr(0)) > 4 {
		for name, ptr := range map[string]unsafe.Pointer{
			"owner": unsafe.Pointer(owner), "update": unsafe.Pointer(update), "player": unsafe.Pointer(playerData),
			"current": unsafe.Pointer(current), "target": unsafe.Pointer(target),
		} {
			if uintptr(ptr) <= uintptr(^uint32(0)) {
				t.Fatalf("%s pointer %#x does not exercise a high native address", name, uintptr(ptr))
			}
		}
	}

	var events []string
	got := weaponRollNative10001EE0(owner, 0, weaponRollNativeDeps10001EE0{
		loadWeaponFlags: func(got *server.Object) uint32 {
			events = append(events, fmt.Sprintf("flags:%p", got))
			if got != target {
				t.Fatalf("flags object = %p, want %p", got, target)
			}
			return 4
		},
		classCanUse: func(got *server.Object, class player.Class) bool {
			events = append(events, fmt.Sprintf("class:%p:%d", got, class))
			return got == target && class == player.Warrior
		},
		checkStrength: func(gotOwner, gotItem *server.Object) bool {
			events = append(events, fmt.Sprintf("strength:%p:%p", gotOwner, gotItem))
			return gotOwner == owner && gotItem == target
		},
		tryDequip: func(gotOwner, gotItem *server.Object) bool {
			events = append(events, fmt.Sprintf("dequip:%p:%p", gotOwner, gotItem))
			return gotOwner == owner && gotItem == current
		},
		tryEquip: func(gotOwner, gotItem *server.Object) bool {
			events = append(events, fmt.Sprintf("equip:%p:%p", gotOwner, gotItem))
			return gotOwner == owner && gotItem == target
		},
	})
	if got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if len(events) != 5 {
		t.Fatalf("events = %v, want five candidate/dequip/equip calls", events)
	}
}

func TestWeaponRoll10001EE0FaultPrefixes(t *testing.T) {
	const (
		owner      = uint64(0x7f38d5b7b3d0)
		update     = uint64(0x6f38e5c8c4e0)
		playerData = uint64(0x5f38f6d9d5f0)
		current    = uint64(0x7f38d5b7a110)
		target     = uint64(0x7f38d5b7a220)
	)
	const injectedFault = "10001EE0 injected fault"
	golden := []string{
		"load-update",
		"load-player",
		"load-status",
		"load-state",
		"load-equipped",
		"load-next",
		"load-flags",
		"load-update",
		"load-player",
		"load-class",
		"class-can-use",
		"check-strength",
		"try-dequip",
		"try-equip",
	}

	run := func(faultAt int) (result int32, panicValue any, events []string) {
		record := func(event string) {
			events = append(events, event)
			if len(events) == faultAt {
				panic(injectedFault)
			}
		}
		hooks := weaponRollHooks10001EE0[uint64, uint64, uint64]{
			loadUpdate: func(got uint64) uint64 {
				record("load-update")
				if got != owner {
					panic("unexpected owner")
				}
				return update
			},
			loadPlayer: func(got uint64) uint64 {
				record("load-player")
				if got != update {
					panic("unexpected update")
				}
				return playerData
			},
			loadPlayerStatus: func(got uint64) uint32 {
				record("load-status")
				if got != playerData {
					panic("unexpected player")
				}
				return 0
			},
			loadState: func(got uint64) server.PlayerState {
				record("load-state")
				if got != update {
					panic("unexpected state owner")
				}
				return server.PlayerState2
			},
			loadEquipped: func(got uint64) uint64 {
				record("load-equipped")
				if got != update {
					panic("unexpected equipped owner")
				}
				return current
			},
			loadFirstItem: func(uint64) uint64 { panic("equipped path loaded first item") },
			loadNextItem: func(got uint64) uint64 {
				record("load-next")
				if got != current {
					panic("unexpected next-item owner")
				}
				return target
			},
			loadPrevItem: func(uint64) uint64 { panic("forward path loaded previous item") },
			loadWeaponFlags: func(got uint64) uint32 {
				record("load-flags")
				if got != target {
					panic("unexpected flags item")
				}
				return 4
			},
			loadPlayerClass: func(got uint64) uint8 {
				record("load-class")
				if got != playerData {
					panic("unexpected class player")
				}
				return uint8(player.Warrior)
			},
			classCanUse: func(got uint64, class uint8) bool {
				record("class-can-use")
				if got != target || class != uint8(player.Warrior) {
					panic("unexpected class check")
				}
				return true
			},
			checkStrength: func(gotOwner, gotItem uint64) bool {
				record("check-strength")
				if gotOwner != owner || gotItem != target {
					panic("unexpected strength check")
				}
				return true
			},
			tryDequip: func(gotOwner, gotItem uint64) bool {
				record("try-dequip")
				if gotOwner != owner || gotItem != current {
					panic("unexpected dequip")
				}
				return true
			},
			tryEquip: func(gotOwner, gotItem uint64) bool {
				record("try-equip")
				if gotOwner != owner || gotItem != target {
					panic("unexpected equip")
				}
				return true
			},
		}

		func() {
			defer func() { panicValue = recover() }()
			result = weaponRoll10001EE0(owner, 1, hooks)
		}()
		return result, panicValue, events
	}

	result, panicValue, events := run(0)
	if panicValue != nil {
		t.Fatalf("complete path panic = %v", panicValue)
	}
	if result != 1 {
		t.Fatalf("complete path result = %d, want 1", result)
	}
	if !reflect.DeepEqual(events, golden) {
		t.Fatalf("complete path events = %v, want %v", events, golden)
	}

	for faultAt := 1; faultAt <= len(golden); faultAt++ {
		t.Run(fmt.Sprintf("fault_%02d_%s", faultAt, golden[faultAt-1]), func(t *testing.T) {
			_, panicValue, events := run(faultAt)
			if panicValue != injectedFault {
				t.Fatalf("panic = %v, want %q", panicValue, injectedFault)
			}
			want := golden[:faultAt]
			if !reflect.DeepEqual(events, want) {
				t.Fatalf("events = %v, want prefix %v", events, want)
			}
		})
	}
}
