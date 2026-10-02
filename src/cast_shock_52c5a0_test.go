package opennox

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// Exercise the production admission/selector, unchanged six-argument C
// wrapper, Shock export, and real BuffApply export without replacing any of
// those services. The empty balance/definition fixture deliberately supplies
// zero duration/sound; nonzero stock balance values are covered in server.
func TestCastShockRoot52C5A0AppliesNativeBuffThroughActualSelector(t *testing.T) {
	base := server.New(nil, nil, strman.New())
	t.Cleanup(base.Close)
	s := &Server{Server: base}
	caster, freeCaster := alloc.New(server.Object{})
	target, freeTarget := alloc.New(server.Object{})
	arg, freeArg := alloc.New(server.SpellAcceptArg{})
	for _, free := range []func(){freeCaster, freeTarget, freeArg} {
		t.Cleanup(free)
	}
	// A client-persistent object uses its native sync slots directly, without
	// requiring a separately loaded ObjectType definition in this fixture.
	*target = server.Object{ObjClass: object.ClassClientPersist, TypeInd: 0x8123, Buffs: uint32(1) << 4}
	*arg = server.SpellAcceptArg{Obj: target, Pos: types.Ptf(-12.5, 33.25)}
	before := *arg
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, ptr := range []unsafe.Pointer{unsafe.Pointer(caster), unsafe.Pointer(target), unsafe.Pointer(arg)} {
			if uintptr(ptr) <= math.MaxUint32 {
				t.Fatalf("root Shock pointer = %p, want C-owned address above 4 GiB", ptr)
			}
		}
	}
	cache := legacy.Get_dword_5d4594_2487712_ptr()
	oldCache, oldServer := *cache, legacy.GetServer
	t.Cleanup(func() { *cache, legacy.GetServer = oldCache, oldServer })
	*cache = 123
	legacy.GetServer = func() legacy.Server { return s }
	if got := s.SpellAccept4FD400(spell.SPELL_SHOCK, caster, caster, nil, arg, 0x180); got != 1 {
		t.Fatalf("actual Shock selector result = %d, want canonical success", got)
	}
	if target.Buffs != uint32(1)<<4|uint32(1)<<server.ENCHANT_SHOCK ||
		target.BuffsDur[server.ENCHANT_SHOCK] != 0 || target.BuffsPower[server.ENCHANT_SHOCK] != 0x80 ||
		target.Field38 != math.MaxUint32 || *arg != before {
		t.Fatalf("actual Shock buff/argument = %#x/%d/%d/%#x/%+v", target.Buffs,
			target.BuffsDur[server.ENCHANT_SHOCK], target.BuffsPower[server.ENCHANT_SHOCK], target.Field38, *arg)
	}
	// The real buff service rejects an already-active zero timer. The cast
	// still succeeds, without changing the old power or acceptance argument.
	if got := s.SpellAccept4FD400(spell.SPELL_SHOCK, caster, caster, nil, arg, 3); got != 1 ||
		target.BuffsPower[server.ENCHANT_SHOCK] != 0x80 || *arg != before {
		t.Fatalf("active-buff rejection changed cast result/power/arg = %d/%d/%+v", got,
			target.BuffsPower[server.ENCHANT_SHOCK], *arg)
	}
}
