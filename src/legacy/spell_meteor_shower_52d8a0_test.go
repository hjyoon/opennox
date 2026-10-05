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

type meteorShowerCastLegacyServer52D8A0 struct {
	Server
	srv *server.Server
}

func (s *meteorShowerCastLegacyServer52D8A0) S() *server.Server { return s.srv }

func TestMeteorShowerCastNativeServiceReadsPackedDWORD52D8A0(t *testing.T) {
	owner, freeOwner := alloc.New(server.Object{})
	defer freeOwner()
	caster, freeCaster := alloc.New(server.Object{})
	defer freeCaster()
	arg, freeArg := alloc.New(server.SpellAcceptArg{})
	defer freeArg()
	*caster = server.Object{PosVec: types.Ptf(-100, -200)} // Ordinary MapTraceRay rejects an out-of-grid origin.
	*arg = server.SpellAcceptArg{Obj: owner, Pos: types.Ptf(30, 40)}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, p := range []unsafe.Pointer{unsafe.Pointer(owner), unsafe.Pointer(caster), unsafe.Pointer(arg)} {
			if uintptr(p) <= math.MaxUint32 {
				t.Fatalf("native MeteorShower pointer=%p, want >4 GiB", p)
			}
		}
		if unsafe.Offsetof(server.SpellAcceptArg{}.Pos) != 8 {
			t.Fatal("test requires the native argument layout")
		}
	}
	cache := memmap.PtrUint32(0x5D4594, 2487800)
	previousCache, previousServer := *cache, GetServer
	t.Cleanup(func() { *cache, GetServer = previousCache, previousServer })
	*cache = 0xfedc0123
	GetServer = func() Server { return &meteorShowerCastLegacyServer52D8A0{srv: new(server.Server)} }
	beforeOwner, beforeCaster, beforeArg := *owner, *caster, *arg
	if got := castMeteorShowerNative52D8A0(spell.SPELL_METEOR_SHOWER, owner, owner, caster, arg, 3); got != 0 ||
		*cache != 0xfedc0123 || *owner != beforeOwner || *caster != beforeCaster || *arg != beforeArg {
		t.Fatalf("native service changed blocked cast: result=%d cache=%#x", got, *cache)
	}
}
