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

func TestFistCastGoDispatchPreservesNativeArguments52D3C0(t *testing.T) {
	second, freeSecond := alloc.New(server.Object{})
	defer freeSecond()
	owner, freeOwner := alloc.New(server.Object{})
	defer freeOwner()
	caster, freeCaster := alloc.New(server.Object{})
	defer freeCaster()
	arg, freeArg := alloc.New(server.SpellAcceptArg{})
	defer freeArg()
	*arg = server.SpellAcceptArg{Obj: second, Pos: types.Ptf(101.5, 202.25)}
	before := *arg
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, p := range []unsafe.Pointer{unsafe.Pointer(second), unsafe.Pointer(owner), unsafe.Pointer(caster), unsafe.Pointer(arg)} {
			if uintptr(p) <= math.MaxUint32 {
				t.Fatalf("pointer=%p, want >4 GiB", p)
			}
		}
	}
	previous := fistCastCall52D3C0
	t.Cleanup(func() { fistCastCall52D3C0 = previous })
	calls := 0
	fistCastCall52D3C0 = func(id spell.ID, gotSecond, gotOwner, gotCaster *server.Object, gotArg *server.SpellAcceptArg, level int) int {
		calls++
		if id != spell.SPELL_FIST || gotSecond != second || gotOwner != owner || gotCaster != caster || gotArg != arg || level != 3 {
			t.Fatalf("native dispatch=%s/%p/%p/%p/%p/%d", id, gotSecond, gotOwner, gotCaster, gotArg, level)
		}
		return -17
	}
	if got := Nox_xxx_castFist_52D3C0(spell.SPELL_FIST, second, owner, caster, arg, 3); got != -17 || calls != 1 || *arg != before {
		t.Fatalf("result/calls/arg=%d/%d/%+v", got, calls, *arg)
	}
}
