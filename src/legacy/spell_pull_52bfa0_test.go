package legacy

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type pullCastLegacyServer52BFA0 struct {
	Server
	srv *server.Server
}

func (s *pullCastLegacyServer52BFA0) S() *server.Server { return s.srv }

// Exercise the actual public selector entry with C-owned native pointers.
// The former six-int C callee narrowed caster +56 before the radial-push
// export could read its origin; the wide-host position is at +60 as well.
func TestPullCastGoEntryPreservesNativePosition52BFA0(t *testing.T) {
	second, freeSecond := alloc.New(server.Object{})
	defer freeSecond()
	owner, freeOwner := alloc.New(server.Object{})
	defer freeOwner()
	caster, freeCaster := alloc.New(server.Object{})
	defer freeCaster()
	arg, freeArg := alloc.New(server.SpellAcceptArg{})
	defer freeArg()
	*caster = server.Object{PosVec: types.Ptf(123.5, -456.25)}
	*owner = server.Object{PosVec: types.Ptf(7, 8)}
	*arg = server.SpellAcceptArg{Obj: second, Pos: types.Ptf(-1000, 2000)}
	before := *arg
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, pointer := range []unsafe.Pointer{unsafe.Pointer(second), unsafe.Pointer(owner), unsafe.Pointer(caster), unsafe.Pointer(arg)} {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("Pull pointer=%p, want >4 GiB", pointer)
			}
		}
	}
	oldServer, oldPush := GetServer, mapPushUnitsAroundCall52E040
	t.Cleanup(func() { GetServer, mapPushUnitsAroundCall52E040 = oldServer, oldPush })
	GetServer = func() Server { return &pullCastLegacyServer52BFA0{srv: new(server.Server)} }
	calls := 0
	mapPushUnitsAroundCall52E040 = func(origin types.Pointf, outer, inner, force float32, source *server.Object, callback, callbackArg int) {
		calls++
		if origin != caster.PosVec || outer != 600 || inner != 10 || math.Float32bits(force) != 0x80000000 || source != nil || callback != 0 || callbackArg != 0 {
			t.Fatalf("Pull radial arguments=%v/%g/%g/%#x/%p/%d/%d", origin, outer, inner, math.Float32bits(force), source, callback, callbackArg)
		}
	}
	if got := Nox_xxx_castPull_52BFA0(spell.SPELL_PULL, second, owner, caster, arg, 3); got != 1 || calls != 1 || *arg != before {
		t.Fatalf("Pull result/calls/arg=%d/%d/%+v", got, calls, *arg)
	}
}

func TestPullCastGoDispatchPreservesNativeArguments52BFA0(t *testing.T) {
	second, freeSecond := alloc.New(server.Object{})
	defer freeSecond()
	owner, freeOwner := alloc.New(server.Object{})
	defer freeOwner()
	caster, freeCaster := alloc.New(server.Object{})
	defer freeCaster()
	arg, freeArg := alloc.New(server.SpellAcceptArg{})
	defer freeArg()
	*arg = server.SpellAcceptArg{Obj: second, Pos: types.Ptf(-1, 2)}
	before := *arg
	previous := pullCastCall52BFA0
	t.Cleanup(func() { pullCastCall52BFA0 = previous })
	calls := 0
	pullCastCall52BFA0 = func(id spell.ID, gotSecond, gotOwner, gotCaster *server.Object, gotArg *server.SpellAcceptArg, level int) int {
		calls++
		if id != spell.SPELL_PULL || gotSecond != second || gotOwner != owner || gotCaster != caster || gotArg != arg || level != -3 {
			t.Fatalf("Pull dispatch=%s/%p/%p/%p/%p/%d", id, gotSecond, gotOwner, gotCaster, gotArg, level)
		}
		return -17
	}
	if got := Nox_xxx_castPull_52BFA0(spell.SPELL_PULL, second, owner, caster, arg, -3); got != -17 || calls != 1 || *arg != before {
		t.Fatalf("Pull result/calls/arg=%d/%d/%+v", got, calls, *arg)
	}
}
