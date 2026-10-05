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

type meteorCastLegacyServer52D9D0 struct {
	Server
	srv *server.Server
}

func (s *meteorCastLegacyServer52D9D0) S() *server.Server { return s.srv }

func TestMeteorCastNativeServiceUsesCSharedCache52D9D0(t *testing.T) {
	last, freeLast := alloc.New(server.Object{TypeInd: 321})
	defer freeLast()
	first, freeFirst := alloc.New(server.Object{TypeInd: 99, Field128: last})
	defer freeFirst()
	owner, freeOwner := alloc.New(server.Object{Field129: first})
	defer freeOwner()
	caster, freeCaster := alloc.New(server.Object{PosVec: types.Ptf(10, 20)})
	defer freeCaster()
	arg, freeArg := alloc.New(server.SpellAcceptArg{Obj: first, Pos: types.Ptf(30, 40)})
	defer freeArg()
	// alloc.New reserves zeroed C storage; its argument specifies the type,
	// not an initial record value.
	*last = server.Object{TypeInd: 321}
	*first = server.Object{TypeInd: 99, Field128: last}
	*owner = server.Object{Field129: first}
	*caster = server.Object{PosVec: types.Ptf(10, 20)}
	*arg = server.SpellAcceptArg{Obj: first, Pos: types.Ptf(30, 40)}
	cache := Get_dword_5d4594_2487804_ptr()
	retired := memmap.PtrUint32(0x5D4594, 2487804)
	if cache == retired {
		t.Fatal("C-owned cache incorrectly aliases the retired PE32 blob")
	}
	previousCache, previousRetired, previousServer := *cache, *retired, GetServer
	t.Cleanup(func() { *cache, *retired, GetServer = previousCache, previousRetired, previousServer })
	*cache, *retired = 321, 123
	GetServer = func() Server { return &meteorCastLegacyServer52D9D0{srv: new(server.Server)} }
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, ptr := range []unsafe.Pointer{unsafe.Pointer(first), unsafe.Pointer(last), unsafe.Pointer(owner), unsafe.Pointer(caster), unsafe.Pointer(arg)} {
			if uintptr(ptr) <= math.MaxUint32 {
				t.Fatalf("native Meteor pointer %p is not above 4 GiB", ptr)
			}
		}
	}
	beforeOwner, beforeFirst, beforeLast, beforeCaster, beforeArg := *owner, *first, *last, *caster, *arg
	// The real service must walk both native-width owned links and reject the
	// duplicate before touching trace/placement services. No outcome is supplied.
	got := castMeteorNative52D9D0(spell.SPELL_METEOR, first, owner, caster, arg, 3)
	if got != 0 || *cache != 321 || *retired != 123 || *owner != beforeOwner || *first != beforeFirst ||
		*last != beforeLast || *caster != beforeCaster || *arg != beforeArg {
		t.Fatalf("Meteor native duplicate changed state: result=%d cache=%d blob=%d", got, *cache, *retired)
	}
}

func TestMeteorCastGoDispatchPreservesNativeArguments52D9D0(t *testing.T) {
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
		for _, ptr := range []unsafe.Pointer{unsafe.Pointer(second), unsafe.Pointer(owner), unsafe.Pointer(caster), unsafe.Pointer(arg)} {
			if uintptr(ptr) <= math.MaxUint32 {
				t.Fatalf("Meteor dispatch pointer=%p, want >4 GiB", ptr)
			}
		}
	}
	previous := meteorCastCall52D9D0
	t.Cleanup(func() { meteorCastCall52D9D0 = previous })
	for _, level := range []int{3, 0, -3, math.MinInt32, math.MaxInt32} {
		calls := 0
		meteorCastCall52D9D0 = func(id spell.ID, gotSecond, gotOwner, gotCaster *server.Object, gotArg *server.SpellAcceptArg, gotLevel int) int {
			calls++
			if id != spell.SPELL_METEOR || gotSecond != second || gotOwner != owner || gotCaster != caster || gotArg != arg || gotLevel != level {
				t.Fatalf("Meteor native dispatch=%s/%p/%p/%p/%p/%d", id, gotSecond, gotOwner, gotCaster, gotArg, gotLevel)
			}
			return -17
		}
		if got := Nox_xxx_castMeteor_52D9D0(spell.SPELL_METEOR, second, owner, caster, arg, level); got != -17 || calls != 1 || *arg != before {
			t.Fatalf("Meteor dispatch result/calls/arg=%d/%d/%+v", got, calls, *arg)
		}
	}
}

// Exercise the public entry without replacing its dispatch hook. The reported
// six-int C callee faults at owner+0x204 before finding this owned Meteor.
func TestMeteorCastGoEntryWalksNativeOwnedPointers52D9D0(t *testing.T) {
	last, freeLast := alloc.New(server.Object{})
	defer freeLast()
	first, freeFirst := alloc.New(server.Object{})
	defer freeFirst()
	owner, freeOwner := alloc.New(server.Object{})
	defer freeOwner()
	*last = server.Object{TypeInd: 321}
	*first = server.Object{TypeInd: 99, Field128: last}
	*owner = server.Object{Field129: first}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, ptr := range []unsafe.Pointer{unsafe.Pointer(first), unsafe.Pointer(last), unsafe.Pointer(owner)} {
			if uintptr(ptr) <= math.MaxUint32 {
				t.Fatalf("Meteor entry pointer=%p, want >4 GiB", ptr)
			}
		}
	}
	cache := Get_dword_5d4594_2487804_ptr()
	previousCache, previousServer := *cache, GetServer
	t.Cleanup(func() { *cache, GetServer = previousCache, previousServer })
	*cache = 321
	GetServer = func() Server { return &meteorCastLegacyServer52D9D0{srv: new(server.Server)} }
	beforeOwner, beforeFirst, beforeLast := *owner, *first, *last
	// The duplicate check precedes accesses to caster and arg in GAME.EXE.
	if got := Nox_xxx_castMeteor_52D9D0(spell.SPELL_METEOR, nil, owner, nil, nil, 3); got != 0 ||
		*cache != 321 || *owner != beforeOwner || *first != beforeFirst || *last != beforeLast {
		t.Fatalf("Meteor native entry result/cache=%d/%d", got, *cache)
	}
}
