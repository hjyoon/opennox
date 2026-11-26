package legacy

/*
#include "quest_win_screen_450770.h"
#include "GAME2.h"
*/
import "C"

import (
	"cmp"
	"encoding/binary"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type questWinRow450770 struct {
	Player      *server.Player
	Kills       uint16
	Generators  uint16
	Secrets     uint16
	CoopSecrets uint16
	Score       uint32
}

type questWinRowBytes450770 [unsafe.Sizeof(questWinRow450770{})]byte

var _ [C.sizeof_nox_quest_stats_row_450770]byte = questWinRowBytes450770{}

// The score comparator takes native row pointers, and sorting swaps whole rows.
var questWinStats450770, _ = alloc.New([6]questWinRow450770{})

func questWinRows450770() *[6]questWinRow450770 {
	return questWinStats450770
}

func questWinSortRows450770(count int) {
	// The original swaps whole records, including any native trailing padding.
	rows := (*[6]questWinRowBytes450770)(unsafe.Pointer(questWinRows450770()))
	sortLegacyQsort(rows[:count], questWinCompareRows450960)
}

func questWinCompareRows450960(a, b *questWinRowBytes450770) int {
	const score = unsafe.Offsetof(questWinRow450770{}.Score)
	return cmp.Compare(binary.LittleEndian.Uint32(b[score:]), binary.LittleEndian.Uint32(a[score:]))
}

func questWinCompareC450960(a, b *questWinRow450770) int32 {
	return int32(C.sub_450960(unsafe.Pointer(a), unsafe.Pointer(b)))
}

func questWinRowCSize450770() uintptr { return C.sizeof_nox_quest_stats_row_450770 }
