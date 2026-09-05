package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

func TestCastSpellByUserExport4FDD20PreservesNativePointersAndSignedDwords(t *testing.T) {
	oldCast := Nox_xxx_castSpellByUser_4FDD20
	t.Cleanup(func() { Nox_xxx_castSpellByUser_4FDD20 = oldCast })

	caster := new(server.Object)
	target := new(server.Object)
	arg := &server.SpellAcceptArg{
		Obj: target,
		Pos: types.Pointf{X: -123.5, Y: 456.25},
	}

	var pin runtime.Pinner
	pin.Pin(caster)
	pin.Pin(target)
	pin.Pin(arg)
	defer pin.Unpin()

	var gotID int32
	var gotCaster *server.Object
	var gotArg *server.SpellAcceptArg
	Nox_xxx_castSpellByUser_4FDD20 = func(id int32, caster *server.Object, arg *server.SpellAcceptArg) int32 {
		gotID, gotCaster, gotArg = id, caster, arg
		return math.MinInt32
	}

	if got := castSpellByUserExportCall4FDD20(math.MaxInt32, caster, arg); got != math.MinInt32 {
		t.Fatalf("export result = %d, want %d", got, int32(math.MinInt32))
	}
	if gotID != math.MaxInt32 || gotCaster != caster || gotArg != arg {
		t.Fatalf("export call = %d/%p/%p, want %d/%p/%p", gotID, gotCaster, gotArg, int32(math.MaxInt32), caster, arg)
	}
	if gotArg.Obj != target || gotArg.Pos != (types.Pointf{X: -123.5, Y: 456.25}) {
		t.Fatalf("export arg = %p/%v, want %p/(-123.5,456.25)", gotArg.Obj, gotArg.Pos, target)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, ptr := range map[string]uintptr{
			"caster": uintptr(unsafe.Pointer(caster)),
			"target": uintptr(unsafe.Pointer(target)),
			"arg":    uintptr(unsafe.Pointer(arg)),
		} {
			if ptr <= math.MaxUint32 {
				t.Fatalf("%s pointer = %#x, want native address above 4 GiB", name, ptr)
			}
		}
	}

	Nox_xxx_castSpellByUser_4FDD20 = func(id int32, caster *server.Object, arg *server.SpellAcceptArg) int32 {
		gotID, gotCaster, gotArg = id, caster, arg
		return math.MaxInt32
	}
	if got := castSpellByUserExportCall4FDD20(math.MinInt32, nil, nil); got != math.MaxInt32 {
		t.Fatalf("nil-pointer export result = %d, want %d", got, int32(math.MaxInt32))
	}
	if gotID != math.MinInt32 || gotCaster != nil || gotArg != nil {
		t.Fatalf("nil-pointer export call = %d/%p/%p", gotID, gotCaster, gotArg)
	}

	runtime.KeepAlive(caster)
	runtime.KeepAlive(target)
	runtime.KeepAlive(arg)
}
