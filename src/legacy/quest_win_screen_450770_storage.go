package legacy

/*
#include "quest_win_screen_450770.h"
#include "GAME2.h"
static void quest_stats_sort_450770(size_t count) {
    qsort(nox_quest_stats_450770, count, sizeof(nox_quest_stats_450770[0]), sub_450960);
}
*/
import "C"

import (
	"unsafe"

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

func questWinRows450770() *[6]questWinRow450770 {
	return (*[6]questWinRow450770)(unsafe.Pointer(&C.nox_quest_stats_450770[0]))
}

func questWinSortRows450770(count int) { C.quest_stats_sort_450770(C.size_t(count)) }

func questWinCompareC450960(a, b *questWinRow450770) int32 {
	return int32(C.sub_450960(unsafe.Pointer(a), unsafe.Pointer(b)))
}

func questWinRowCSize450770() uintptr { return C.sizeof_nox_quest_stats_row_450770 }
