package legacy

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type fumbleCastLegacyServer52C060 struct {
	Server
	srv *server.Server
}

func (s *fumbleCastLegacyServer52C060) S() *server.Server { return s.srv }

// The public entry formerly loaded the low DWORD of the native argument's
// target pointer into a C int before reading its class/subclass. A protected
// monster needs no inventory or physics side effects, but still faults on
// that pointer load in the old implementation.
func TestFumbleCastGoEntryPreservesTarget52C060(t *testing.T) {
	target, freeTarget := alloc.New(server.Object{})
	defer freeTarget()
	caster, freeCaster := alloc.New(server.Object{})
	defer freeCaster()
	arg, freeArg := alloc.New(server.SpellAcceptArg{})
	defer freeArg()
	*target = server.Object{ObjClass: object.ClassMonster, ObjSubClass: object.SubClass(0x2000)}
	*caster = server.Object{PosVec: types.Ptf(123.5, -456.25)}
	*arg = server.SpellAcceptArg{Obj: target, Pos: types.Ptf(-1000, 2000)}
	before := *arg
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, pointer := range []unsafe.Pointer{unsafe.Pointer(target), unsafe.Pointer(caster), unsafe.Pointer(arg)} {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("Fumble pointer=%p, want >4 GiB", pointer)
			}
		}
	}
	previous := GetServer
	t.Cleanup(func() { GetServer = previous })
	GetServer = func() Server { return &fumbleCastLegacyServer52C060{srv: new(server.Server)} }
	if got := Nox_xxx_castFumble_52C060(spell.SPELL_FUMBLE, nil, nil, caster, arg, 3); got != 1 || *arg != before {
		t.Fatalf("Fumble result/arg=%d/%+v", got, *arg)
	}
}

func TestFumbleCastGoDispatchPreservesNativeArguments52C060(t *testing.T) {
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
	previous := fumbleCastCall52C060
	t.Cleanup(func() { fumbleCastCall52C060 = previous })
	calls := 0
	fumbleCastCall52C060 = func(id spell.ID, gotSecond, gotOwner, gotCaster *server.Object, gotArg *server.SpellAcceptArg, level int) int {
		calls++
		if id != spell.SPELL_FUMBLE || gotSecond != second || gotOwner != owner || gotCaster != caster || gotArg != arg || level != -3 {
			t.Fatalf("Fumble dispatch=%s/%p/%p/%p/%p/%d", id, gotSecond, gotOwner, gotCaster, gotArg, level)
		}
		return -17
	}
	if got := Nox_xxx_castFumble_52C060(spell.SPELL_FUMBLE, second, owner, caster, arg, -3); got != -17 || calls != 1 || *arg != before {
		t.Fatalf("Fumble result/calls/arg=%d/%d/%+v", got, calls, *arg)
	}
}
