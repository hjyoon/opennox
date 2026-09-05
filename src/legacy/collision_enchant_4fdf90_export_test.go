package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func TestCollisionEnchantExport4FDF90PreservesNativePointers(t *testing.T) {
	oldCollide := Nox_xxx_collide_4FDF90
	t.Cleanup(func() { Nox_xxx_collide_4FDF90 = oldCollide })

	source := new(server.Object)
	target := new(server.Object)

	var pin runtime.Pinner
	pin.Pin(source)
	pin.Pin(target)
	defer pin.Unpin()

	var gotSource, gotTarget *server.Object
	Nox_xxx_collide_4FDF90 = func(source, target *server.Object) {
		gotSource, gotTarget = source, target
	}
	collisionEnchantExportCall4FDF90(source, target)
	if gotSource != source || gotTarget != target {
		t.Fatalf("export args = %p/%p, want %p/%p", gotSource, gotTarget, source, target)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, pointer := range map[string]uintptr{
			"source": uintptr(unsafe.Pointer(source)),
			"target": uintptr(unsafe.Pointer(target)),
		} {
			if pointer <= math.MaxUint32 {
				t.Fatalf("%s pointer = %#x, want native address above 4 GiB", name, pointer)
			}
		}
	}

	gotSource, gotTarget = source, target
	collisionEnchantExportCall4FDF90(nil, nil)
	if gotSource != nil || gotTarget != nil {
		t.Fatalf("nil export args = %p/%p", gotSource, gotTarget)
	}

	runtime.KeepAlive(source)
	runtime.KeepAlive(target)
}
