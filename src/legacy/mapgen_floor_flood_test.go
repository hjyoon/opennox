package legacy

import (
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestFloorFloodNativeCABI(t *testing.T) {
	const columns, rows, tile, edgeID = 12, 14, 7, 9
	grid, freeGrid := alloc.Make([]worldTileCell411160(nil), columns*rows)
	t.Cleanup(freeGrid)
	owners, freeOwners := alloc.Make([]unsafe.Pointer(nil), columns)
	t.Cleanup(freeOwners)
	nodes, freeNodes := alloc.Make([]tileListNativeNode411350(nil), 10)
	t.Cleanup(freeNodes)
	for i := range nodes[:len(nodes)-1] {
		nodes[i].Next = &nodes[i+1]
	}
	for x := range owners {
		owners[x] = unsafe.Pointer(&grid[x*rows])
	}
	at := func(x, y int) *worldTileCell411160 { return &grid[x*rows+y] }
	for i := range grid {
		grid[i].Value1, grid[i].Value6 = tile, tile
	}

	owner := (*unsafe.Pointer)(worldTileHeadCell411160())
	pool, floorMode := game4TilesTestBrushState()
	count := game4TilesTestWorklistCount()
	scratch := memmap.Slice(0x973F18, 16200)[:6004]
	saved := [6004]byte(scratch)
	mode, selectedTile := memmap.PtrUint32(0x973F18, 35916), memmap.PtrUint32(0x973F18, 35912)
	_, edge, variation := game4TilesTestEdgeWords()
	oldOwner, oldPool := *owner, *pool
	oldCount, oldFloorMode, oldMode, oldTile := *count, *floorMode, *mode, *selectedTile
	oldEdge, oldVariation := *edge, *variation
	t.Cleanup(func() {
		*owner, *pool = oldOwner, oldPool
		*count, *floorMode, *mode, *selectedTile = oldCount, oldFloorMode, oldMode, oldTile
		*edge, *variation = oldEdge, oldVariation
		copy(scratch, saved[:])
	})
	*owner, *pool = unsafe.Pointer(&owners[0]), unsafe.Pointer(&nodes[0])
	*floorMode, *mode, *selectedTile, *edge, *variation = 1, 0, tile, edgeID, 2
	clear(scratch)
	*count = 500
	game4TilesTestEnqueue(9, 11, 1, tile)
	if *count != 500 || !floorFloodTestOverflow() {
		t.Fatal("worklist cap did not report overflow")
	}
	*count = 0
	game4TilesTestEnqueue(9, 11, 1, tile)
	game4TilesTestEnqueue(9, 11, 1, tile)
	game4TilesTestEnqueue(9, 10, 3, tile)
	game4TilesTestEnqueue(9, 1, 1, tile)
	game4TilesTestEnqueue(1, 11, 2, tile)
	if *count != 2 {
		t.Fatalf("deduplication or border filtering left %d entries", *count)
	}
	out := [3]uint32{}
	if game4TilesTestDequeue(&out) != 1 || out != [3]uint32{9, 10, 3} || *count != 1 {
		t.Fatalf("first native pop = %v, count = %d", out, *count)
	}
	if game4TilesTestDequeue(&out) != 1 || out != [3]uint32{9, 11, 1} || *count != 0 {
		t.Fatalf("second native pop = %v, count = %d", out, *count)
	}
	if game4TilesTestDequeue(&out) != 0 || out != [3]uint32{9, 11, 1} {
		t.Fatalf("empty native pop wrote %v", out)
	}

	def := &worldTileDefinitions411160()[tile]
	oldWidth, oldHeight := def.Field52, def.Field53
	t.Cleanup(func() { def.Field52, def.Field53 = oldWidth, oldHeight })
	def.Field52, def.Field53 = 2, 3
	dims := thingEdgeBytes411850(edgeID, 28696, 2)
	oldDims := [2]byte(dims)
	t.Cleanup(func() { copy(dims, oldDims[:]) })
	dims[0], dims[1] = 3, 3

	// This seed belongs to the left quarter at (9, 11). The adjacent tiles
	// have another base tile and receive the boundary subtiles in native order.
	at(9, 11).Value1 = 5
	brush := mapgenFloorBrush{tileID: tile, edgeID: edgeID}
	status := game4TilesTestApplyFlood(&[2]int32{449, 495}, &brush)
	first := at(10, 11).List10
	if status != 1 || *count != 0 || floorFloodTestOverflow() || brush.edgeVariation != 7 || first != &nodes[0] {
		t.Fatalf("flood status = %d, count = %d, edge variation = %d", status, *count, brush.edgeVariation)
	}
	for i, want := range []struct {
		node *tileListNativeNode411350
		kind int32
	}{
		{first, 4}, {at(10, 10).List10, 1}, {at(9, 11).List10, 6}, {at(9, 10).List10, 3},
		{at(9, 10).List5, 0}, {at(10, 11).List5, 2}, {at(8, 11).List5, 5}, {at(9, 12).List5, 7},
	} {
		node := &nodes[i]
		if node != want.node || node.Value != tile || node.Edge != edgeID || node.Variation != want.kind || node.Next != nil {
			t.Fatalf("flood emission %d = %+v at %p, want edge variation %d at %p", i, node, node, want.kind, want.node)
		}
	}
	if game4TilesTestApplySubtile(448.5, 494.5) != 1 || *floorMode != 0 || at(9, 11).List5 == nil {
		t.Fatal("native subtile producer did not paint the selected quarter and reset the mode")
	}
	*selectedTile = 255
	if game4TilesTestApplySubtile(448.5, 494.5) != 1 || *floorMode != 0 || at(9, 11).List5 != nil {
		t.Fatal("native no-tile sentinel did not clear the quarter and reset the mode")
	}
}

func floorFloodTestOverflow() bool { return memmap.Uint32(0x973F18, 22200) != 0 }
