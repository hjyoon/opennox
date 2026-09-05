package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

type positionDeltaLegacyServer4FEA70 struct {
	Server
	srv *server.Server
}

func (s *positionDeltaLegacyServer4FEA70) S() *server.Server {
	return s.srv
}

func TestPositionDeltaExport4FEA70PreservesNativePointersAndBinary32(t *testing.T) {
	srv := new(server.Server)
	oldGetServer := GetServer
	GetServer = func() Server { return &positionDeltaLegacyServer4FEA70{srv: srv} }
	t.Cleanup(func() { GetServer = oldGetServer })

	object := &server.Object{PosVec: types.Pointf{
		X: math.Float32frombits(0x41200001),
		Y: math.Float32frombits(0xc1a00001),
	}}
	point := &types.Pointf{
		X: math.Float32frombits(0x41600001),
		Y: math.Float32frombits(0xc1700001),
	}
	var pin runtime.Pinner
	pin.Pin(object)
	pin.Pin(point)
	defer pin.Unpin()

	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, pointer := range map[string]uintptr{
			"object": uintptr(unsafe.Pointer(object)),
			"point":  uintptr(unsafe.Pointer(point)),
		} {
			if pointer <= math.MaxUint32 {
				t.Fatalf("%s pointer = %#x, want native address above 4 GiB", name, pointer)
			}
		}
	}

	if got := positionDeltaExportCall4FEA70(object, point); got != 1 {
		t.Fatalf("first CGo result = %d, want canonical one", got)
	}
	if got, want := math.Float32bits(object.PosVec.X), uint32(0x41200001); got != want {
		t.Fatalf("object X bits = %#08x, want %#08x", got, want)
	}
	if got, want := math.Float32bits(point.Y), uint32(0xc1700001); got != want {
		t.Fatalf("point Y bits = %#08x, want %#08x", got, want)
	}

	point.X = object.PosVec.X
	point.Y = object.PosVec.Y
	if got := positionDeltaExportCall4FEA70(object, point); got != 0 {
		t.Fatalf("live replacement CGo result = %d, want zero", got)
	}
	runtime.KeepAlive(object)
	runtime.KeepAlive(point)
}
