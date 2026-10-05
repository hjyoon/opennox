package opennox

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/spell"

	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type meteorCastRootServer52D9D0 struct {
	legacy.Server
	srv *server.Server
}

func (s *meteorCastRootServer52D9D0) S() *server.Server { return s.srv }

func TestMeteorCastRootInstantSelectorUsesNativeEntry52D9D0(t *testing.T) {
	owned, freeOwned := alloc.New(server.Object{})
	defer freeOwned()
	owner, freeOwner := alloc.New(server.Object{})
	defer freeOwner()
	*owned = server.Object{TypeInd: 321}
	*owner = server.Object{Field129: owned}
	if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(owner)) <= math.MaxUint32 || uintptr(unsafe.Pointer(owned)) <= math.MaxUint32) {
		t.Fatalf("Meteor selector pointers owner=%p owned=%p, want >4 GiB", owner, owned)
	}
	s := &Server{Server: new(server.Server)}
	cache := legacy.Get_dword_5d4594_2487804_ptr()
	previousCache, previousServer := *cache, legacy.GetServer
	t.Cleanup(func() { *cache, legacy.GetServer = previousCache, previousServer })
	*cache = 321
	legacy.GetServer = func() legacy.Server { return &meteorCastRootServer52D9D0{srv: s.Server} }
	beforeOwner, beforeOwned := *owner, *owned
	if got := s.spellAcceptInstant4FD400(spell.SPELL_METEOR, nil, owner, nil, nil, 3); got != 0 ||
		*cache != 321 || *owner != beforeOwner || *owned != beforeOwned {
		t.Fatalf("Meteor selector result/cache=%d/%d", got, *cache)
	}
}
