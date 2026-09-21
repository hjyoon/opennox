package server

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
)

func TestAIPathMapIndexFlags50AB50(t *testing.T) {
	var paths serverAIPaths

	for _, tc := range []struct {
		x int
		y int
	}{
		{x: -1, y: 0},
		{x: aiMapIndexSize, y: 0},
		{x: 0, y: -1},
		{x: 0, y: aiMapIndexSize},
	} {
		if got := paths.MapIndexFlags(tc.x, tc.y); got != 0 {
			t.Fatalf("flags at (%d, %d) = %#x, want 0", tc.x, tc.y, got)
		}
	}

	paths.MapIndex(0, 0).Flags8 = 0xa55a
	paths.MapIndex(aiMapIndexSize-1, aiMapIndexSize-1).Flags8 = 0x5aa5
	paths.MapIndex(3, 7).Flags8 = 0x1234
	paths.MapIndex(7, 3).Flags8 = 0x4321
	for _, tc := range []struct {
		x    int
		y    int
		want AIMapIndexFlags
	}{
		{x: 0, y: 0, want: 0xa55a},
		{x: aiMapIndexSize - 1, y: aiMapIndexSize - 1, want: 0x5aa5},
		{x: 3, y: 7, want: 0x1234},
		{x: 7, y: 3, want: 0x4321},
	} {
		if got := paths.MapIndexFlags(tc.x, tc.y); got != tc.want {
			t.Fatalf("flags at (%d, %d) = %#x, want %#x", tc.x, tc.y, got, tc.want)
		}
	}
}

func TestAIPathStorageNativeLayouts50AB90(t *testing.T) {
	ptrSize := unsafe.Sizeof(uintptr(0))
	wantVisitSize := uintptr(16)
	wantField4 := uintptr(4)
	wantField8 := uintptr(8)
	wantFlags12 := uintptr(12)
	wantField13 := uintptr(13)
	wantField14 := uintptr(14)
	if ptrSize == 8 {
		wantVisitSize = 32
		wantField4 = 8
		wantField8 = 16
		wantFlags12 = 24
		wantField13 = 25
		wantField14 = 26
	}

	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{name: "AIMapIndexNode size", got: unsafe.Sizeof(AIMapIndexNode{}), want: 12},
		{name: "AIMapIndexNode.Index0", got: unsafe.Offsetof(AIMapIndexNode{}.Index0), want: 0},
		{name: "AIMapIndexNode.IndexGen4", got: unsafe.Offsetof(AIMapIndexNode{}.IndexGen4), want: 4},
		{name: "AIMapIndexNode.Flags8", got: unsafe.Offsetof(AIMapIndexNode{}.Flags8), want: 8},
		{name: "AIMapIndexNode.Field10", got: unsafe.Offsetof(AIMapIndexNode{}.Field10), want: 10},
		{name: "AIVisitNode size", got: unsafe.Sizeof(AIVisitNode{}), want: wantVisitSize},
		{name: "AIVisitNode.X0", got: unsafe.Offsetof(AIVisitNode{}.X0), want: 0},
		{name: "AIVisitNode.Y2", got: unsafe.Offsetof(AIVisitNode{}.Y2), want: 2},
		{name: "AIVisitNode.Field4", got: unsafe.Offsetof(AIVisitNode{}.Field4), want: wantField4},
		{name: "AIVisitNode.Field8", got: unsafe.Offsetof(AIVisitNode{}.Field8), want: wantField8},
		{name: "AIVisitNode.Flags12", got: unsafe.Offsetof(AIVisitNode{}.Flags12), want: wantFlags12},
		{name: "AIVisitNode.Field13", got: unsafe.Offsetof(AIVisitNode{}.Field13), want: wantField13},
		{name: "AIVisitNode.Field14", got: unsafe.Offsetof(AIVisitNode{}.Field14), want: wantField14},
		{name: "Pointf size", got: unsafe.Sizeof(types.Pointf{}), want: 8},
		{name: "Pointf.Y", got: unsafe.Offsetof(types.Pointf{}.Y), want: 4},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}
}

func TestAIPathStorageLifecycle50AB90And50ABF0(t *testing.T) {
	srv := new(Server)
	paths := serverAIPaths{lastFrame: math.MaxUint32}
	paths.Init(srv)
	t.Cleanup(func() {
		if paths.Valid() {
			paths.Free()
		}
	})

	if !paths.Valid() {
		t.Fatal("visit-node allocator is invalid after initialization")
	}
	if paths.s != srv {
		t.Fatalf("server = %p, want %p", paths.s, srv)
	}
	if got := len(paths.points); got != 1024 {
		t.Fatalf("point count = %d, want 1024", got)
	}
	if got := cap(paths.points); got != 1024 {
		t.Fatalf("point capacity = %d, want 1024", got)
	}
	if paths.lastFrame != 0 {
		t.Fatalf("last frame = %d, want 0", paths.lastFrame)
	}

	first := paths.NewVisitNode()
	second := paths.NewVisitNode()
	if first == nil || second == nil || first == second {
		t.Fatalf("visit nodes = %p/%p, want two distinct records", first, second)
	}
	first.Field4 = second
	second.Field8 = first
	if first.Field4 != second || second.Field8 != first {
		t.Fatalf("native visit links did not round trip: %p/%p", first.Field4, second.Field8)
	}
	paths.points[1023] = types.Pointf{X: 12.5, Y: -7.25}
	if got := paths.points[1023]; got != (types.Pointf{X: 12.5, Y: -7.25}) {
		t.Fatalf("last point = %+v", got)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		if got := uintptr(unsafe.Pointer(first)); got <= math.MaxUint32 {
			t.Fatalf("visit-node address = %#x, want a native pointer above 4 GiB", got)
		}
		if got := uintptr(unsafe.Pointer(unsafe.SliceData(paths.points))); got <= math.MaxUint32 {
			t.Fatalf("point-buffer address = %#x, want a native pointer above 4 GiB", got)
		}
	}

	paths.Free()
	if paths.Valid() {
		t.Fatal("visit-node allocator remained valid after cleanup")
	}
	if paths.allocVisit.Class != nil {
		t.Fatalf("allocator handle after cleanup = %p, want nil", paths.allocVisit.Class)
	}
	if paths.points != nil {
		t.Fatalf("point buffer after cleanup = %p, want nil", unsafe.SliceData(paths.points))
	}
}
