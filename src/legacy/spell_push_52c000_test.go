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

type pushCastLegacyServer52C000 struct {
	Server
	srv *server.Server
}

func (s *pushCastLegacyServer52C000) S() *server.Server { return s.srv }

// Exercise the real public spell selector entry, not a test-only replacement
// of that entry. The former six-int C callee truncated caster +56 before the
// radial-push export could read its origin; PosVec is also at +60 on wide hosts.
func TestPushCastGoEntryPreservesNativePosition52C000(t *testing.T) {
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
				t.Fatalf("Push pointer=%p, want >4 GiB", pointer)
			}
		}
	}
	oldServer, oldPush := GetServer, mapPushUnitsAroundCall52E040
	t.Cleanup(func() { GetServer, mapPushUnitsAroundCall52E040 = oldServer, oldPush })
	GetServer = func() Server { return &pushCastLegacyServer52C000{srv: new(server.Server)} }
	calls := 0
	mapPushUnitsAroundCall52E040 = func(origin types.Pointf, outer, inner, force float32, source *server.Object, callback, callbackArg int) {
		calls++
		if origin != caster.PosVec || outer != 600 || inner != 10 || force != 0 || source != nil || callback != 0 || callbackArg != 0 {
			t.Fatalf("Push radial arguments=%v/%g/%g/%g/%p/%d/%d", origin, outer, inner, force, source, callback, callbackArg)
		}
	}
	if got := Nox_xxx_castPush_52C000(spell.SPELL_PUSH, second, owner, caster, arg, 3); got != 1 || calls != 1 || *arg != before {
		t.Fatalf("Push result/calls/arg=%d/%d/%+v", got, calls, *arg)
	}
}

func TestPushCastGoDispatchPreservesNativeArguments52C000(t *testing.T) {
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
	previous := pushCastCall52C000
	t.Cleanup(func() { pushCastCall52C000 = previous })
	calls := 0
	pushCastCall52C000 = func(id spell.ID, gotSecond, gotOwner, gotCaster *server.Object, gotArg *server.SpellAcceptArg, level int) int {
		calls++
		if id != spell.SPELL_PUSH || gotSecond != second || gotOwner != owner || gotCaster != caster || gotArg != arg || level != -3 {
			t.Fatalf("Push dispatch=%s/%p/%p/%p/%p/%d", id, gotSecond, gotOwner, gotCaster, gotArg, level)
		}
		return -17
	}
	if got := Nox_xxx_castPush_52C000(spell.SPELL_PUSH, second, owner, caster, arg, -3); got != -17 || calls != 1 || *arg != before {
		t.Fatalf("Push result/calls/arg=%d/%d/%+v", got, calls, *arg)
	}
}
