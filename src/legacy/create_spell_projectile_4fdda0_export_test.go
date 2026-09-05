package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/spell"

	"github.com/opennox/opennox/v1/server"
)

func TestCreateSpellProjectileExport4FDDA0PreservesNativePointersAndSignedDword(t *testing.T) {
	oldCreate := Nox_xxx_createSpellFly_4FDDA0
	t.Cleanup(func() { Nox_xxx_createSpellFly_4FDDA0 = oldCreate })

	source := new(server.Object)
	target := new(server.Object)
	projectile := new(server.Object)

	var pin runtime.Pinner
	pin.Pin(source)
	pin.Pin(target)
	pin.Pin(projectile)
	defer pin.Unpin()

	var gotSource, gotTarget *server.Object
	var gotSpell spell.ID
	Nox_xxx_createSpellFly_4FDDA0 = func(
		source, target *server.Object,
		spellID spell.ID,
	) *server.Object {
		gotSource, gotTarget, gotSpell = source, target, spellID
		return projectile
	}

	if got := createSpellProjectileExportCall4FDDA0(source, target, math.MinInt32); got != projectile {
		t.Fatalf("export result = %p, want %p", got, projectile)
	}
	if gotSource != source || gotTarget != target || int32(gotSpell) != math.MinInt32 {
		t.Fatalf("export args = %p/%p/%d, want %p/%p/%d", gotSource, gotTarget, gotSpell, source, target, int32(math.MinInt32))
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, pointer := range map[string]uintptr{
			"source":     uintptr(unsafe.Pointer(source)),
			"target":     uintptr(unsafe.Pointer(target)),
			"projectile": uintptr(unsafe.Pointer(projectile)),
		} {
			if pointer <= math.MaxUint32 {
				t.Fatalf("%s pointer = %#x, want native address above 4 GiB", name, pointer)
			}
		}
	}

	Nox_xxx_createSpellFly_4FDDA0 = func(
		source, target *server.Object,
		spellID spell.ID,
	) *server.Object {
		gotSource, gotTarget, gotSpell = source, target, spellID
		return nil
	}
	if got := createSpellProjectileExportCall4FDDA0(nil, nil, math.MaxInt32); got != nil {
		t.Fatalf("nil export result = %p", got)
	}
	if gotSource != nil || gotTarget != nil || int32(gotSpell) != math.MaxInt32 {
		t.Fatalf("nil export args = %p/%p/%d", gotSource, gotTarget, gotSpell)
	}

	runtime.KeepAlive(source)
	runtime.KeepAlive(target)
	runtime.KeepAlive(projectile)
}
