package legacy

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// Exercise the real selector entry with C-owned objects, not the injectable
// dispatch. No balance file makes the radius zero; the native map still has
// to enumerate the missile at the caster's exact live position.
func TestInversionCastGoEntryNativePositionAndOwner52BEB0(t *testing.T) {
	srv := server.New(nil, nil, strman.New())
	t.Cleanup(srv.Close)
	srv.Map.Init()
	second, freeSecond := alloc.New(server.Object{})
	t.Cleanup(freeSecond)
	owner, freeOwner := alloc.New(server.Object{})
	t.Cleanup(freeOwner)
	caster, freeCaster := alloc.New(server.Object{})
	t.Cleanup(freeCaster)
	*caster = server.Object{PosVec: types.Ptf(123.5, 456.25)}
	arg, freeArg := alloc.New(server.SpellAcceptArg{})
	t.Cleanup(freeArg)
	*arg = server.SpellAcceptArg{Obj: second, Pos: types.Ptf(1000, 2000)}
	missile, freeMissile := alloc.New(server.Object{})
	*missile = server.Object{
		ObjClass: object.ClassMissile, ObjFlags: object.FlagActive,
		PosVec: caster.PosVec, NewPos: caster.PosVec,
	}
	t.Cleanup(freeMissile)
	srv.Map.AddObjectToIndex(missile)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, p := range []unsafe.Pointer{unsafe.Pointer(owner), unsafe.Pointer(caster), unsafe.Pointer(missile), unsafe.Pointer(arg)} {
			if uintptr(p) <= math.MaxUint32 {
				t.Fatalf("inversion pointer=%p, want >4 GiB", p)
			}
		}
	}
	oldServer, oldChangeOwner := GetServer, Nox_xxx_changeOwner_52BE40
	t.Cleanup(func() { GetServer, Nox_xxx_changeOwner_52BE40 = oldServer, oldChangeOwner })
	GetServer = func() Server { return &monsterMainLegacyServer547210{srv: srv} }
	changes := 0
	Nox_xxx_changeOwner_52BE40 = func(gotMissile, gotOwner *server.Object) {
		changes++
		if gotMissile != missile || gotOwner != owner {
			t.Fatalf("inversion callback=%p/%p want %p/%p", gotMissile, gotOwner, missile, owner)
		}
	}
	before := *arg
	t.Logf("real inversion selector: caster=%p owner=%p missile=%p", caster, owner, missile)
	if got := Sub_52BEB0(spell.SPELL_INVERSION, second, owner, caster, arg, -3); got != 1 || changes != 1 || *arg != before {
		t.Fatalf("inversion result/changes/arg=%d/%d/%+v", got, changes, *arg)
	}
}

func TestInversionCastGoDispatch52BEB0PreservesAllArguments(t *testing.T) {
	second, freeSecond := alloc.New(server.Object{})
	t.Cleanup(freeSecond)
	owner, freeOwner := alloc.New(server.Object{})
	t.Cleanup(freeOwner)
	caster, freeCaster := alloc.New(server.Object{})
	t.Cleanup(freeCaster)
	arg, freeArg := alloc.New(server.SpellAcceptArg{})
	t.Cleanup(freeArg)
	*arg = server.SpellAcceptArg{Obj: second}
	previous := inversionCastCall52BEB0
	t.Cleanup(func() { inversionCastCall52BEB0 = previous })
	calls := 0
	inversionCastCall52BEB0 = func(id spell.ID, gotSecond, gotOwner, gotCaster *server.Object, gotArg *server.SpellAcceptArg, level int) int {
		calls++
		if id != spell.SPELL_INVERSION || gotSecond != second || gotOwner != owner || gotCaster != caster || gotArg != arg || level != -3 {
			t.Fatal("inversion dispatch narrowed or replaced an argument")
		}
		return -17
	}
	if Sub_52BEB0(spell.SPELL_INVERSION, second, owner, caster, arg, -3) != -17 || calls != 1 {
		t.Fatal("inversion dispatch changed native return")
	}
}
