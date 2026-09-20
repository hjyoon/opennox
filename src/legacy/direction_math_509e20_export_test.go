package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestDirectionMath509ExCABIRecordsAndResults(t *testing.T) {
	data, freeData := alloc.New(server.DirectionInitData{X: -1, Y: 1})
	defer freeData()
	vector, freeVector := alloc.New(server.IndexedDirectionVector509E20{})
	defer freeVector()
	point, freePoint := alloc.New(types.Ptf(3, 4))
	defer freePoint()
	*data = server.DirectionInitData{X: -1, Y: 1}
	*point = types.Ptf(3, 4)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, pointer := range map[string]unsafe.Pointer{
			"direction data": unsafe.Pointer(data),
			"indexed vector": unsafe.Pointer(vector),
			"float vector":   unsafe.Pointer(point),
		} {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("%s pointer = %p, want native address above 4 GiB", name, pointer)
			}
		}
	}

	if got := directionToAngleExportCall509E00(data); got != 96 {
		t.Fatalf("direction-to-angle result = %d, want 96", got)
	}
	if got := indexedDirectionExportCall509E20(64, vector); got != 6 ||
		*vector != (server.IndexedDirectionVector509E20{Y: 1}) {
		t.Fatalf("indexed direction result/vector = %d/%+v, want 6/{0 1}", got, *vector)
	}
	if got := directionIndexToAngleExportCall509E90(8); got != 32 {
		t.Fatalf("indexed angle result = %d, want 32", got)
	}
	if got := directionOctantExportCall509EA0(192); got != 1 {
		t.Fatalf("octant result = %d, want 1", got)
	}
	normalizeVectorExportCall509F20(point)
	want := types.Ptf(3, 4)
	server.NormalizeVector509F20(&want)
	if math.Float32bits(point.X) != math.Float32bits(want.X) ||
		math.Float32bits(point.Y) != math.Float32bits(want.Y) {
		t.Fatalf("normalized vector bits = %#x/%#x, want %#x/%#x",
			math.Float32bits(point.X), math.Float32bits(point.Y),
			math.Float32bits(want.X), math.Float32bits(want.Y))
	}
	runtime.KeepAlive(data)
	runtime.KeepAlive(vector)
	runtime.KeepAlive(point)
}

func TestDirectionMath509ExCABIAllValidDirections(t *testing.T) {
	for direction := int32(0); direction < 256; direction++ {
		var direct server.IndexedDirectionVector509E20
		wantResult := server.IndexedDirection509E20(direction, &direct)
		var throughC server.IndexedDirectionVector509E20
		if got := indexedDirectionExportCall509E20(direction, &throughC); got != wantResult || throughC != direct {
			t.Fatalf("direction %d C result/vector = %d/%+v, want %d/%+v",
				direction, got, throughC, wantResult, direct)
		}
		if got, want := directionOctantExportCall509EA0(direction), server.DirectionOctant509EA0(direction); got != want {
			t.Fatalf("direction %d C octant = %d, want %d", direction, got, want)
		}
	}
}
