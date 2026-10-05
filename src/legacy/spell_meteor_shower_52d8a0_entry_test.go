package legacy

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestMeteorShowerCastGoDispatchPreservesNativeArguments52D8A0(t *testing.T) {
	second, freeSecond := alloc.New(server.Object{})
	defer freeSecond()
	owner, freeOwner := alloc.New(server.Object{})
	defer freeOwner()
	caster, freeCaster := alloc.New(server.Object{})
	defer freeCaster()
	arg, freeArg := alloc.New(server.SpellAcceptArg{})
	defer freeArg()
	*arg = server.SpellAcceptArg{Obj: second, Pos: types.Ptf(101.5, -202.25)}
	before := *arg
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, p := range []unsafe.Pointer{unsafe.Pointer(second), unsafe.Pointer(owner), unsafe.Pointer(caster), unsafe.Pointer(arg)} {
			if uintptr(p) <= math.MaxUint32 {
				t.Fatalf("native dispatch pointer=%p, want >4 GiB", p)
			}
		}
	}
	previous := meteorShowerCastCall52D8A0
	t.Cleanup(func() { meteorShowerCastCall52D8A0 = previous })
	for _, level := range []int{3, 0, -3, math.MinInt32, math.MaxInt32} {
		calls := 0
		meteorShowerCastCall52D8A0 = func(id spell.ID, gotSecond, gotOwner, gotCaster *server.Object, gotArg *server.SpellAcceptArg, gotLevel int) int {
			calls++
			if id != spell.SPELL_METEOR_SHOWER || gotSecond != second || gotOwner != owner || gotCaster != caster ||
				gotArg != arg || gotLevel != level {
				t.Fatalf("native dispatch=%s/%p/%p/%p/%p/%d", id, gotSecond, gotOwner, gotCaster, gotArg, gotLevel)
			}
			return -17
		}
		if got := Nox_xxx_castMeteorShower_52D8A0(spell.SPELL_METEOR_SHOWER, second, owner, caster, arg, level); got != -17 || calls != 1 || *arg != before {
			t.Fatalf("dispatch result/calls/arg=%d/%d/%+v", got, calls, *arg)
		}
	}
}

func TestMeteorShowerCastGoEntryReadsNativeArgument52D8A0(t *testing.T) {
	caster, freeCaster := alloc.New(server.Object{})
	defer freeCaster()
	arg, freeArg := alloc.New(server.SpellAcceptArg{})
	defer freeArg()
	*caster = server.Object{PosVec: types.Ptf(-100, -200)}
	*arg = server.SpellAcceptArg{Pos: types.Ptf(30, 40)}
	if unsafe.Sizeof(uintptr(0)) == 8 &&
		(uintptr(unsafe.Pointer(caster)) <= math.MaxUint32 || uintptr(unsafe.Pointer(arg)) <= math.MaxUint32) {
		t.Fatalf("native entry pointers=%p/%p, want >4 GiB", caster, arg)
	}
	cache := memmap.PtrUint32(0x5D4594, 2487800)
	previousCache, previousServer := *cache, GetServer
	t.Cleanup(func() { *cache, GetServer = previousCache, previousServer })
	*cache = 321
	GetServer = func() Server { return &meteorShowerCastLegacyServer52D8A0{srv: new(server.Server)} }
	beforeCaster, beforeArg := *caster, *arg
	// No dispatch replacement: the public entry must read the actual C-owned
	// native argument and reach the ordinary out-of-grid trace rejection.
	if got := Nox_xxx_castMeteorShower_52D8A0(spell.SPELL_METEOR_SHOWER, nil, nil, caster, arg, 3); got != 0 ||
		*cache != 321 || *caster != beforeCaster || *arg != beforeArg {
		t.Fatalf("entry result/cache=%d/%d", got, *cache)
	}
}
