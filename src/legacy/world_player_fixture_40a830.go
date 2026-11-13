package legacy

/*
#include <stdint.h>
#include <stdlib.h>
#include "world_player_reference_40a830.h"
*/
import "C"
import (
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
	"runtime"
	"unsafe"
)

func worldPlayerBool40A830(v bool) uint32 {
	if v {
		return 1
	}
	return 0
}

// These two private symbols are only service seams for the exact renamed
// C reference copies. They never replace an entry algorithm with its result.
//
//export nox_world_player_reference_pointer
func nox_world_player_reference_pointer(op C.int32_t, p unsafe.Pointer, a, b C.uint32_t, q unsafe.Pointer) unsafe.Pointer {
	r := worldPlayerRuntimeCurrent40A830()
	switch int(op) {
	case 1:
		return unsafe.Pointer(r.PlayerFirst())
	case 2:
		return unsafe.Pointer(r.PlayerNext((*server.Player)(p)))
	case 3:
		return unsafe.Pointer(r.UnitFirst())
	case 4:
		return unsafe.Pointer(r.UnitNext((*server.Object)(p)))
	case 5:
		return unsafe.Pointer(r.TeamFirst())
	case 6:
		return unsafe.Pointer(r.TeamNext((*server.Team)(p)))
	case 7:
		return unsafe.Pointer(r.WallCreate(int32(a), int32(b)))
	case 8:
		return r.WallAttach((*server.Wall)(p), (*server.Object)(q))
	case 9:
		return unsafe.Pointer(r.ArmorDefinition(uint16(a)))
	default:
		panic("world/player reference pointer service")
	}
}

//export nox_world_player_reference_word
func nox_world_player_reference_word(op C.int32_t, p unsafe.Pointer, a, b C.uint32_t, q unsafe.Pointer, n C.uint32_t) C.uint32_t {
	r := worldPlayerRuntimeCurrent40A830()
	var word uint32
	switch int(op) {
	case 1:
		word = worldPlayerBool40A830(r.TeamContains((*server.ObjectTeam)(p), server.TeamID(a)))
	case 2:
		word = worldPlayerBool40A830(r.EngineNoRendering())
	case 3:
		word = worldPlayerBool40A830(r.GameplayTeams())
	case 4:
		word = r.GameFlag(uint32(a))
	case 5:
		word = r.GameFlags()
	case 6:
		word = uint32(r.TimerStatus())
	case 7:
		word = uint32(r.TimerMode(int16(uint16(a))))
	case 8:
		r.TimerStart()
	case 9:
		word = uint32(r.TimerSet(int32(a)))
	case 10:
		if n != 7 || b != 0 || q != nil {
			panic("world/player reference fixed seven-byte packet ABI")
		}
		packet := append([]byte(nil), unsafe.Slice((*byte)(p), int(n))...)
		word = uint32(r.Send(int32(a), packet))
	case 11:
		r.Lesson((*server.Object)(p))
	case 12:
		r.Minimap(true, byte(a), (*server.Object)(p), uint32(b))
	case 13:
		r.Minimap(false, byte(a), (*server.Object)(p), uint32(b))
	default:
		panic("world/player reference word service")
	}
	runtime.KeepAlive(p)
	runtime.KeepAlive(q)
	return C.uint32_t(word)
}
func worldPlayerFixtureWord40A830(backend string, entry int, p unsafe.Pointer, a uint32) uint32 {
	if backend != "go" {
		ref := 0
		if backend == "old" {
			ref = 1
		}
		return uint32(C.world_player_fixture_word(C.int(ref), C.int(entry), p, C.uint32_t(a)))
	}
	switch entry {
	case 0:
		return uint32(countNonEliminatedNative40A830((*server.Team)(p)))
	case 1:
		return uint32(gamePlayAnyNative40A8A0())
	case 2:
		return uint32(cliCanTalkNative4100F0((*uint16)(p)))
	case 3:
		return uint32(tileAllocNative410F60())
	case 4:
		tileFreeNative410FC0()
		return 0
	case 5:
		return uint32(tileFromPointNative411160((*types.Pointf)(p)))
	case 6:
		return uint32(tileAllowsTeleportNative411A90((*types.Pointf)(p)))
	case 7:
		return uint32(checkPauseNative413A50())
	case 8:
		return uint32(settingsAcquiredNative4169D0(int32(a)))
	case 9:
		return uint32(playerUnsetNative417530((*server.Player)(p), a))
	case 10:
		return uint32(netReportStatusNative417630((*server.Player)(p)))
	default:
		panic("world/player fixture word entry")
	}
}
func worldPlayerFixturePointer40A830(backend string, entry int, p, q unsafe.Pointer, a, b int32) unsafe.Pointer {
	if backend != "go" {
		ref := 0
		if backend == "old" {
			ref = 1
		}
		return C.world_player_fixture_pointer(C.int(ref), C.int(entry), p, q, C.int32_t(a), C.int32_t(b))
	}
	switch entry {
	case 0:
		return doorAttachWallNative410360((*server.Object)(p), a, b)
	case 1:
		return unsafe.Pointer(secretWallBlockNative410760((*server.SecretWall)(p)))
	case 2:
		return unsafe.Pointer(secretWallFirstNative410780())
	case 3:
		return unsafe.Pointer(secretWallNextNative410790((*server.SecretWall)(p)))
	case 4:
		return unsafe.Pointer(modifierNextNative4133C0((*server.ModifierEff)(p)))
	case 5:
		return cliGamedataNative416590(a)
	case 6:
		return playerForceLessonsNative416E50(a)
	case 7:
		return unsafe.Pointer(playerByNameNative4170D0((*uint16)(p)))
	case 8:
		return minimapAllNative417470((*server.Object)(p), uint32(a), false)
	case 9:
		return minimapAllNative417470((*server.Object)(p), uint32(a), true)
	case 10:
		return sendAllStatusNative4175C0((*server.Player)(p))
	default:
		panic("world/player fixture pointer entry")
	}
}
func worldPlayerFixtureArmor40A830(backend string, p *server.Object) float64 {
	if backend == "go" {
		return itemDefendNative415C00(p)
	}
	ref := 0
	if backend == "old" {
		ref = 1
	}
	return float64(C.world_player_fixture_armor(C.int(ref), unsafe.Pointer(p)))
}
func worldPlayerFixtureLayout40A830(i int) uintptr {
	return uintptr(C.world_player_fixture_layout(C.int(i)))
}
func worldPlayerFixtureAlloc40A830(n uintptr) unsafe.Pointer { return game4SupportAllocate(1, n) }
func worldPlayerFixtureFree40A830(p unsafe.Pointer)          { game4SupportReleaseMemory(p) }
func worldPlayerFixtureTileOwner40A830(p unsafe.Pointer, n int32) {
	C.world_player_fixture_tile_owner(p, C.int(n))
}
func worldPlayerFixtureModifierNext40A830(p, next unsafe.Pointer) {
	C.world_player_fixture_modifier_next(p, next)
}
func worldPlayerFixtureDefendCallback40A830() unsafe.Pointer {
	return C.world_player_fixture_defend_callback()
}
func worldPlayerFixtureDefendSeen40A830(e, item unsafe.Pointer) bool {
	return C.world_player_fixture_defend_seen(e, item) != 0
}
