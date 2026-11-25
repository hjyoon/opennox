package legacy

import (
	"encoding/binary"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/internal/cfixture"
)

func TestClientInventoryCompactionPreservesStackOrderAndAlternatePointer(t *testing.T) {
	grid, layout := cfixture.ClientSecondaryGridLayout198()
	original := slices.Clone(grid)
	selected, previous := cfixture.ClientSecondaryGridGlobals198()

	destination, freeDestination := alloc.NewHigh(client.Drawable{TypeIDVal: 7})
	defer freeDestination()
	source, freeSource := alloc.NewHigh(client.Drawable{TypeIDVal: 7})
	defer freeSource()

	defer func() {
		copy(grid, original)
		cfixture.ClientSecondaryGridRestore198(selected, previous)
	}()
	clear(grid)

	cell := func(i int) []byte { return grid[i*layout[0] : (i+1)*layout[0]] }
	writeCode := func(dst []byte, i int, code uint32) {
		binary.LittleEndian.PutUint32(dst[layout[2]+4*i:], code)
	}
	readCode := func(src []byte, i int) uint32 {
		return binary.LittleEndian.Uint32(src[layout[2]+4*i:])
	}

	first, gap, occupied, reserved := cell(0), cell(21), cell(42), cell(20)
	*(*unsafe.Pointer)(unsafe.Pointer(&first[0])) = destination.C()
	*(*unsafe.Pointer)(unsafe.Pointer(&occupied[0])) = source.C()
	first[layout[5]], occupied[layout[5]], reserved[layout[5]] = 30, 5, 1
	binary.LittleEndian.PutUint32(occupied[layout[4]:], 1)
	for i := range 5 {
		writeCode(occupied, i, 200+uint32(i))
	}
	writeCode(reserved, 0, 999)
	checkReserved := slices.Clone(reserved)
	cfixture.ClientSecondaryGridPrepare198(42)

	sub_461B50()

	if first[layout[5]] != 32 || readCode(first, 30) != 204 || readCode(first, 31) != 203 {
		t.Fatalf("merged stack has size %d and new codes %d %d", first[layout[5]], readCode(first, 30), readCode(first, 31))
	}
	if gap[layout[5]] != 3 || occupied[layout[5]] != 0 || *(*unsafe.Pointer)(unsafe.Pointer(&gap[0])) != source.C() {
		t.Fatalf("retained stack has size %d and source size %d", gap[layout[5]], occupied[layout[5]])
	}
	for i := range 3 {
		if got := readCode(gap, i); got != uint32(200+i) {
			t.Fatalf("remaining code %d = %d", i, got)
		}
	}
	if got, _ := cfixture.ClientSecondaryGridGlobals198(); got != uintptr(unsafe.Pointer(&gap[0])) {
		t.Fatalf("alternate weapon points to %#x, want %p", got, &gap[0])
	}
	if !slices.Equal(reserved, checkReserved) {
		t.Fatal("compaction modified the reserved inventory row")
	}
}
