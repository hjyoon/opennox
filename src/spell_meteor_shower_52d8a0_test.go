package opennox

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type meteorShowerCastRootServer52D8A0 struct {
	legacy.Server
	srv *server.Server
}

func (s *meteorShowerCastRootServer52D8A0) S() *server.Server { return s.srv }

func TestMeteorShowerCastRootInstantSelectorUsesNativeEntry52D8A0(t *testing.T) {
	caster, freeCaster := alloc.New(server.Object{})
	defer freeCaster()
	arg, freeArg := alloc.New(server.SpellAcceptArg{})
	defer freeArg()
	*caster = server.Object{PosVec: types.Ptf(-100, -200)}
	*arg = server.SpellAcceptArg{Pos: types.Ptf(30, 40)}
	if unsafe.Sizeof(uintptr(0)) == 8 &&
		(uintptr(unsafe.Pointer(caster)) <= math.MaxUint32 || uintptr(unsafe.Pointer(arg)) <= math.MaxUint32) {
		t.Fatalf("native selector pointers=%p/%p, want >4 GiB", caster, arg)
	}
	s := &Server{Server: new(server.Server)}
	cache := memmap.PtrUint32(0x5D4594, 2487800)
	previousCache, previousServer := *cache, legacy.GetServer
	t.Cleanup(func() { *cache, legacy.GetServer = previousCache, previousServer })
	*cache = 321
	legacy.GetServer = func() legacy.Server { return &meteorShowerCastRootServer52D8A0{srv: s.Server} }
	beforeCaster, beforeArg := *caster, *arg
	if got := s.spellAcceptInstant4FD400(spell.SPELL_METEOR_SHOWER, nil, nil, caster, arg, 3); got != 0 ||
		*cache != 321 || *caster != beforeCaster || *arg != beforeArg {
		t.Fatalf("selector result/cache=%d/%d", got, *cache)
	}
}
