package cfixture

/*
#include "world_player_reference_40a830.h"
*/
import "C"

import "unsafe"

var WorldReferencePointer func(op int32, p unsafe.Pointer, a, b uint32, q unsafe.Pointer) unsafe.Pointer
var WorldReferenceWord func(op int32, p unsafe.Pointer, a, b uint32, q unsafe.Pointer, n uint32) uint32

//export nox_world_player_reference_pointer
func nox_world_player_reference_pointer(op C.int32_t, p unsafe.Pointer, a, b C.uint32_t, q unsafe.Pointer) unsafe.Pointer {
	return WorldReferencePointer(int32(op), p, uint32(a), uint32(b), q)
}

//export nox_world_player_reference_word
func nox_world_player_reference_word(op C.int32_t, p unsafe.Pointer, a, b C.uint32_t, q unsafe.Pointer, n C.uint32_t) C.uint32_t {
	return C.uint32_t(WorldReferenceWord(int32(op), p, uint32(a), uint32(b), q, uint32(n)))
}

func referenceFlag(reference bool) C.int {
	if reference {
		return 1
	}
	return 0
}

func WorldWord(reference bool, entry int, p unsafe.Pointer, arg uint32) uint32 {
	return uint32(C.world_player_fixture_word(referenceFlag(reference), C.int(entry), p, C.uint32_t(arg)))
}

func WorldPointer(reference bool, entry int, p, q unsafe.Pointer, a, b int32) unsafe.Pointer {
	return C.world_player_fixture_pointer(referenceFlag(reference), C.int(entry), p, q, C.int32_t(a), C.int32_t(b))
}

func WorldArmor(reference bool, p unsafe.Pointer) float64 {
	return float64(C.world_player_fixture_armor(referenceFlag(reference), p))
}

func WorldLayout(field int) uintptr {
	return uintptr(C.world_player_fixture_layout(C.int(field)))
}

func WorldTileOwner(head unsafe.Pointer, capacity int32) {
	C.world_player_fixture_tile_owner(head, C.int(capacity))
}

func WorldModifierNext(p, next unsafe.Pointer) {
	C.world_player_fixture_modifier_next(p, next)
}

func WorldDefendCallback() unsafe.Pointer {
	return C.world_player_fixture_defend_callback()
}

func WorldDefendSeen(effect, item unsafe.Pointer) bool {
	return C.world_player_fixture_defend_seen(effect, item) != 0
}
