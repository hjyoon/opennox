package legacy

import (
	"fmt"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type shockCastLegacyServer52C5A0 struct {
	Server
	srv *server.Server
}

func TestCastShockRealCWrapper52C5A0GlyphAndDamagePointers(t *testing.T) {
	target, freeTarget := alloc.New(server.Object{})
	caster, freeCaster := alloc.New(server.Object{})
	context, freeContext := alloc.New(server.Object{})
	arg, freeArg := alloc.New(server.SpellAcceptArg{})
	for _, free := range []func(){freeTarget, freeCaster, freeContext, freeArg} {
		t.Cleanup(free)
	}
	context.TypeInd = 0x8123
	*arg = server.SpellAcceptArg{Obj: target, Pos: types.Ptf(-1000, 2000)}
	before := *arg
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, ptr := range []unsafe.Pointer{unsafe.Pointer(target), unsafe.Pointer(caster), unsafe.Pointer(context), unsafe.Pointer(arg)} {
			if uintptr(ptr) <= math.MaxUint32 {
				t.Fatalf("C-owned Shock pointer=%p, want above 4 GiB", ptr)
			}
		}
	}
	cache := Get_dword_5d4594_2487712_ptr()
	oldCache, oldServer := *cache, GetServer
	t.Cleanup(func() { *cache, GetServer = oldCache, oldServer })
	*cache = 0x8123
	GetServer = func() Server { return &shockCastLegacyServer52C5A0{srv: new(server.Server)} }
	callback := objectDamageNativeProbePtr()
	calls := 0
	server.RegisterObjectDamageGo(fmt.Sprintf("ShockNativeWidthTest%d", objectDamageNativeTestSequence.Add(1)), callback,
		func(unit, source, weapon *server.Object, damage int32, typ object.DamageType) bool {
			calls++
			if unit != target || source != caster || weapon != caster || damage != 0 || typ != object.DamageElectric {
				t.Fatalf("C Shock damage=%p/%p/%p/%d/%d", unit, source, weapon, damage, typ)
			}
			return false
		})
	target.Damage = callback
	if got := Nox_xxx_useShock_52C5A0(spell.SPELL_SHOCK, nil, caster, context, arg, 3); got != 1 || calls != 1 || *arg != before {
		t.Fatalf("Glyph result/calls/arg=%d/%d/%+v", got, calls, *arg)
	}
}

func TestCastShockCEntry52C5A0SignedDwordsAndNativeLinks(t *testing.T) {
	caster, freeCaster := alloc.New(server.Object{})
	context, freeContext := alloc.New(server.Object{})
	arg, freeArg := alloc.New(server.SpellAcceptArg{})
	for _, free := range []func(){freeCaster, freeContext, freeArg} {
		t.Cleanup(free)
	}
	old := castShockCall52C5A0
	t.Cleanup(func() { castShockCall52C5A0 = old })
	calls := 0
	castShockCall52C5A0 = func(source, glyph *server.Object, argument *server.SpellAcceptArg, power int32) int32 {
		calls++
		if source != caster || glyph != context || argument != arg || power != math.MinInt32 {
			t.Fatalf("C Shock export=%p/%p/%p/%d", source, glyph, argument, power)
		}
		return math.MinInt32
	}
	if got := Nox_xxx_useShock_52C5A0(spell.ID(math.MinInt32), nil, caster, context, arg, math.MinInt32); got != math.MinInt32 || calls != 1 {
		t.Fatalf("signed C result/calls=%d/%d", got, calls)
	}
}

func TestCastShockCEntry52C5A0NilTargetSkipsServerLookup(t *testing.T) {
	arg, freeArg := alloc.New(server.SpellAcceptArg{})
	context, freeContext := alloc.New(server.Object{})
	t.Cleanup(freeArg)
	t.Cleanup(freeContext)
	cache := Get_dword_5d4594_2487712_ptr()
	oldCache, oldServer := *cache, GetServer
	t.Cleanup(func() { *cache, GetServer = oldCache, oldServer })
	*cache = 0
	GetServer = func() Server { panic("Shock queried the server before its original target gate") }
	if got := Nox_xxx_useShock_52C5A0(spell.SPELL_SHOCK, nil, nil, context, arg, 1); got != 0 || *cache != 0 {
		t.Fatalf("nil target result/cache=%d/%d", got, *cache)
	}
}

func (s *shockCastLegacyServer52C5A0) S() *server.Server { return s.srv }

// Use the unchanged public Go wrapper and actual six-argument C entrypoint.
// Both the acceptance argument and the target are native-width C allocations.
func TestCastShockRealCWrapper52C5A0PreservesTarget(t *testing.T) {
	target, freeTarget := alloc.New(server.Object{})
	arg, freeArg := alloc.New(server.SpellAcceptArg{})
	t.Cleanup(freeTarget)
	t.Cleanup(freeArg)
	*arg = server.SpellAcceptArg{Obj: target, Pos: types.Ptf(-12.5, 33.25)}
	before := *arg
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, ptr := range []unsafe.Pointer{unsafe.Pointer(target), unsafe.Pointer(arg)} {
			if uintptr(ptr) <= math.MaxUint32 {
				t.Fatalf("Shock pointer=%p, want above 4 GiB", ptr)
			}
		}
	}
	t.Logf("real C Shock: target=%p arg=%p", target, arg)
	cache := Get_dword_5d4594_2487712_ptr()
	oldCache, oldServer, oldApply := *cache, GetServer, buffApplyExportImpl4FF380
	t.Cleanup(func() { *cache, GetServer, buffApplyExportImpl4FF380 = oldCache, oldServer, oldApply })
	*cache = 123
	GetServer = func() Server { return &shockCastLegacyServer52C5A0{srv: new(server.Server)} }
	calls := 0
	buffApplyExportImpl4FF380 = func(unit *server.Object, buff int32, duration int16, power int8) {
		calls++
		if unit != target || buff != 22 || duration != 0 || power != -128 {
			t.Errorf("Shock buff=%p/%d/%d/%d, want %p/22/0/-128", unit, buff, duration, power, target)
		}
	}
	if got := Nox_xxx_useShock_52C5A0(spell.SPELL_SHOCK, nil, nil, nil, arg, 0x180); got != 1 || calls != 1 || *arg != before {
		t.Fatalf("Shock result/calls/arg=%d/%d/%+v", got, calls, *arg)
	}
	arg.Obj = nil
	if got := Nox_xxx_useShock_52C5A0(spell.SPELL_SHOCK, nil, nil, nil, arg, 1); got != 0 || calls != 1 {
		t.Fatalf("nil target result/calls=%d/%d", got, calls)
	}
}
