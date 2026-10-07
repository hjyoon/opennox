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

func TestCurePoisonNativeDispatch52CDB0(t *testing.T) {
	second, freeSecond := alloc.New(server.Object{})
	defer freeSecond()
	owner, freeOwner := alloc.New(server.Object{})
	defer freeOwner()
	caster, freeCaster := alloc.New(server.Object{})
	defer freeCaster()
	arg, freeArg := alloc.New(server.SpellAcceptArg{})
	defer freeArg()
	*arg = server.SpellAcceptArg{Obj: second, Pos: types.Ptf(-1, 2)}
	wantArg := *arg
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, ptr := range []unsafe.Pointer{unsafe.Pointer(second), unsafe.Pointer(owner), unsafe.Pointer(caster), unsafe.Pointer(arg)} {
			if uintptr(ptr) <= math.MaxUint32 {
				t.Fatalf("native dispatch allocation=%p, want >4 GiB", ptr)
			}
		}
	}
	previous := curePoisonCastCall52CDB0
	t.Cleanup(func() { curePoisonCastCall52CDB0 = previous })
	calls := 0
	curePoisonCastCall52CDB0 = func(id spell.ID, gotSecond, gotOwner, gotCaster *server.Object, gotArg *server.SpellAcceptArg, power int) int {
		calls++
		if id != spell.SPELL_CURE_POISON || gotSecond != second || gotOwner != owner || gotCaster != caster || gotArg != arg || power != math.MinInt32 {
			t.Fatal("six native selector arguments changed")
		}
		return -17
	}
	if got := Nox_xxx_castCurePoison_52CDB0(spell.SPELL_CURE_POISON, second, owner, caster, arg, math.MinInt32); got != -17 || calls != 1 || *arg != wantArg {
		t.Fatalf("result/calls/arg=%d/%d/%+v", got, calls, *arg)
	}
}
