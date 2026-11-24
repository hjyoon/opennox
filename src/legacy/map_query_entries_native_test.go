package legacy

import (
	"math"
	"slices"
	"testing"
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestMapTypeQueriesNativeCEntries(t *testing.T) {
	origin, freeOrigin := alloc.NewHigh([349]uint32{})
	t.Cleanup(freeOrigin)
	entry, freeEntry := alloc.NewHigh(Nox_map_list_item{})
	t.Cleanup(freeEntry)
	old := Nox_mapToGameFlags
	t.Cleanup(func() { Nox_mapToGameFlags = old })
	var calls []int32
	Nox_mapToGameFlags = func(value int) noxflags.GameFlag {
		calls = append(calls, int32(value))
		return noxflags.GameFlag(uint32(value) ^ 0x1210)
	}
	origin[348], entry.Field_7 = 0x87654321, 0xfedcba98
	first, second := mapTypeQueriesC(unsafe.Pointer(origin), entry)
	if first != int32(origin[348]^0x1210) || second != int32(entry.Field_7^0x1210) ||
		!slices.Equal(calls, []int32{int32(origin[348]), int32(entry.Field_7)}) {
		t.Fatalf("map flags returned %08x and %08x with inputs %v", uint32(first), uint32(second), calls)
	}
}

func TestMapBookAndQuestQueryNativeCEntries(t *testing.T) {
	engine, game := noxflags.GetEngine(), noxflags.GetGame()
	mode := memmap.PtrUint32(0x5D4594, 1548484)
	warp := memmap.PtrUint32(questStateBase4D71F0, questWarpOnOffset4D7520)
	oldMode, oldWarp := *mode, *warp
	t.Cleanup(func() {
		noxflags.ResetEngine()
		noxflags.SetEngine(engine)
		noxflags.ResetGame()
		noxflags.SetGame(game)
		*mode, *warp = oldMode, oldWarp
	})
	noxflags.ResetEngine()
	*mode = 0
	if got := legacyNoRenderingC(); got != 0 {
		t.Fatalf("rendering mode = %d, want 0", got)
	}
	noxflags.SetEngine(noxflags.EngineNoRendering)
	if got := legacyNoRenderingC(); got != 1 {
		t.Fatalf("headless engine mode = %d, want 1", got)
	}
	noxflags.ResetEngine()
	*mode = 0x80000000
	if got := legacyNoRenderingC(); got != 1 {
		t.Fatalf("headless state word = %d, want 1", got)
	}

	for _, quest := range []bool{false, true} {
		noxflags.ResetGame()
		if quest {
			noxflags.SetGame(noxflags.GameModeQuest)
		}
		for _, index := range []int32{math.MinInt32, 36, 37, 38, 39, 40, 41, 110, 111, 112, 113, 114, 115, math.MaxInt32} {
			creature, spell := legacyBookQueriesC(index)
			if creature != int32(bool2int(quest || index < 37 || index > 40)) ||
				spell != int32(bool2int(quest || index < 111 || index > 114)) {
				t.Fatalf("quest=%v index=%d returned creature=%d spell=%d", quest, index, creature, spell)
			}
		}
	}

	for _, value := range []uint32{0, 1, 0x80000000, 0xffffffff} {
		*warp = value
		if got := legacyQuestWarpFlagC(); got != int32(value) {
			t.Fatalf("quest warp status = %d, want %d", got, int32(value))
		}
	}
}
