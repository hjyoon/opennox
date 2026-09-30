package legacy

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestQuestWin450770NativeRowsAndUnsignedPrefixSort(t *testing.T) {
	rows := questWinRows450770()
	old := *rows
	t.Cleanup(func() { *rows = old })
	*rows = [6]questWinRow450770{}
	if got := questWinRowCSize450770(); got != unsafe.Sizeof(rows[0]) {
		t.Fatalf("C row=%d Go row=%d", got, unsafe.Sizeof(rows[0]))
	}
	for i, score := range []uint32{1, 0x80000000, math.MaxUint32, 7, 9, 11} {
		player, free := alloc.New(server.Player{})
		t.Cleanup(free)
		if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(unsafe.Pointer(player)) <= math.MaxUint32 {
			t.Fatalf("player=%p, want above 4 GiB", player)
		}
		rows[i] = questWinRow450770{player, uint16(i + 1), uint16(i + 2), uint16(i + 3), uint16(i + 4), score}
	}
	before := *rows
	if questWinCompareC450960(&rows[0], &rows[1]) != 1 || questWinCompareC450960(&rows[2], &rows[1]) != -1 || questWinCompareC450960(&rows[2], &rows[2]) != 0 {
		t.Fatal("C comparator did not preserve unsigned descending score comparisons")
	}
	questWinSortRows450770(3)
	for dst, src := range []int{2, 1, 0, 3, 4, 5} {
		if rows[dst] != before[src] {
			t.Fatalf("row %d=%+v want source %d=%+v", dst, rows[dst], src, before[src])
		}
	}
	questWinSortRows450770(0)
	questWinSortRows450770(1)
	for dst, src := range []int{2, 1, 0, 3, 4, 5} {
		if rows[dst] != before[src] {
			t.Fatalf("zero/singleton prefix changed row %d", dst)
		}
	}
}
