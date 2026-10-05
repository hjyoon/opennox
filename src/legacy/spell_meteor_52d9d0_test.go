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
