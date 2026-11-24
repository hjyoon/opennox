package legacy

import (
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestFloorBrushNativeCABI(t *testing.T) {
	const x, y, tile, edge = 5, 7, 7, 9
	row, freeRow := alloc.Make([]worldTileCell411160(nil), y+2)
	t.Cleanup(freeRow)
	rows, freeRows := alloc.Make([]unsafe.Pointer(nil), x+1)
	t.Cleanup(freeRows)
	rows[x] = unsafe.Pointer(&row[0])
	owner := (*unsafe.Pointer)(worldTileHeadCell411160())
	pool, subtileMode := game4TilesTestBrushState()
	mode := memmap.PtrUint32(0x973F18, 35916)
	oldOwner, oldPool, oldSubtileMode, oldMode := *owner, *pool, *subtileMode, *mode
	t.Cleanup(func() { *owner, *pool, *subtileMode, *mode = oldOwner, oldPool, oldSubtileMode, oldMode })
	*owner, *pool, *subtileMode, *mode = unsafe.Pointer(&rows[0]), nil, 0, 0

	def := &worldTileDefinitions411160()[tile]
	oldWidth, oldHeight := def.Field52, def.Field53
	t.Cleanup(func() { def.Field52, def.Field53 = oldWidth, oldHeight })
	def.Field52, def.Field53 = 2, 3
	dims := thingEdgeBytes411850(edge, 28696, 2)
	oldDims := [2]byte(dims)
	t.Cleanup(func() { copy(dims, oldDims[:]) })
	dims[0], dims[1] = 3, 3

	for _, index := range []uintptr{3 + 12*2, 3 + 12*4, 6 + 12*6} {
		word := memmap.PtrUint32(0x587000, 282736+4*index)
		old := *word
		t.Cleanup(func() { *word = old })
		*word = 6
	}

	cell := &row[y]
	cell.Flag = 0xa4
	brush := mapgenFloorBrush{tileID: tile, edgeID: edge, edgeVariation: 2}
	if got := game4TilesTestApplyBrush(x, y, 1, 0, &brush); got != 1 || cell.Flag != 0xa5 || cell.Value1 != tile || cell.Field2 != 4 {
		t.Fatalf("base left brush returned %d and wrote %+v", got, cell)
	}
	if got := game4TilesTestApplyBrushToken(x, y, 2, &brush); got != 1 || cell.Flag != 0xa7 || cell.Value6 != tile || cell.Field7 != 0 {
		t.Fatalf("token brush returned %d and wrote %+v", got, cell)
	}
	if row[y+1] != (worldTileCell411160{}) {
		t.Fatal("a base brush overwrote the next row cell")
	}

	*subtileMode = 1
	if got := game4TilesTestApplyBrush(x, y, 1, 0, &brush); got != 1 || cell.List5 == nil || cell.List10 != nil {
		t.Fatalf("subtile brush returned %d and wrote %+v", got, cell)
	}
	first := cell.List5
	t.Cleanup(func() {
		*pool = oldPool
		game4TilesTestFreeBrushNodes(unsafe.Pointer(first))
	})
	if got := game4TilesTestApplyBrush(x, y, 1, 0, &brush); got != 0 || first.Next != nil {
		t.Fatalf("duplicate brush returned %d and wrote %+v", got, first)
	}
	brush.edgeVariation = 4
	if got := game4TilesTestApplyBrush(x, y, 1, 0, &brush); got != 1 || first.Next == nil {
		t.Fatalf("second subtile returned %d and wrote %+v", got, first)
	}
	second := first.Next
	brush.edgeVariation = 3
	if got := game4TilesTestApplyBrush(x, y, 1, 1, &brush); got != 1 || cell.List5 != first || first.Next != nil || first.Variation != 6 || *pool != unsafe.Pointer(second) {
		t.Fatalf("merge returned %d, wrote %+v and recycled %p", got, first, *pool)
	}

	*subtileMode = 0
	if got := game4TilesTestApplyBrush(x, y, 1, 0, nil); got != 1 || cell.List5 != nil || cell.Value1 != 255 || cell.Flag != 0xa6 ||
		*pool != unsafe.Pointer(first) || first.Next != second {
		t.Fatalf("clearing the left layer returned %d, wrote %+v and recycled %p", got, cell, *pool)
	}
}
