package legacy

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/spell"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestLockPublicSelector52CE90KeepsSixArgumentsAndSignedResult(t *testing.T) {
	second, freeSecond := alloc.New(server.Object{})
	caster, freeCaster := alloc.New(server.Object{})
	aim, freeAim := alloc.New(server.Object{})
	arg, freeArg := alloc.New(server.SpellAcceptArg{})
	for _, free := range []func(){freeSecond, freeCaster, freeAim, freeArg} {
		t.Cleanup(free)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, ptr := range []unsafe.Pointer{unsafe.Pointer(second), unsafe.Pointer(caster), unsafe.Pointer(aim), unsafe.Pointer(arg)} {
			if uintptr(ptr) <= math.MaxUint32 {
				t.Fatalf("native pointer=%p want >4 GiB", ptr)
			}
		}
	}
	previous := lockCastCall52CE90
	t.Cleanup(func() { lockCastCall52CE90 = previous })
	calls := 0
	lockCastCall52CE90 = func(id spell.ID, a2, a3, a4 *server.Object, sa *server.SpellAcceptArg, level int) int {
		calls++
		if id != spell.ID(math.MinInt32) || a2 != second || a3 != caster || a4 != aim || sa != arg || level != math.MinInt32 {
			t.Fatal("Lock selector narrowed or reordered six arguments")
		}
		return math.MinInt32
	}
	if got := Nox_xxx_castLock_52CE90(spell.ID(math.MinInt32), second, caster, aim, arg, math.MinInt32); got != math.MinInt32 || calls != 1 {
		t.Fatalf("result/calls=%d/%d", got, calls)
	}
}
