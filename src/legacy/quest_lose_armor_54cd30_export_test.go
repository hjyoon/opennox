package legacy

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/prand"
	"github.com/opennox/libs/strman"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type questLoseArmorLegacyServer54CD30 struct {
	Server
	native  *server.Server
	deleted *server.Object
	deletes int
}

func (s *questLoseArmorLegacyServer54CD30) S() *server.Server { return s.native }
func (s *questLoseArmorLegacyServer54CD30) DelayedDelete(item *server.Object) {
	s.deleted, s.deletes = item, s.deletes+1
}

func TestQuestLoseArmor54CD30NativeCGoRoundTrip(t *testing.T) {
	s := server.New(nil, nil, strman.New())
	t.Cleanup(s.Close)
	outer := &questLoseArmorLegacyServer54CD30{native: s}
	oldServer := GetServer
	GetServer = func() Server { return outer }
	t.Cleanup(func() { GetServer = oldServer })
	u, freeUnit := alloc.New(server.Object{})
	t.Cleanup(freeUnit)
	a, freeA := alloc.New(server.Object{})
	t.Cleanup(freeA)
	b, freeB := alloc.New(server.Object{})
	t.Cleanup(freeB)
	*u = server.Object{InvFirstItem: a, Worth: 0xaabbccdd}
	*a = server.Object{TypeInd: 65535, ObjClass: 0x82000000, ObjFlags: 0x80000100, InvNextItem: b}
	*b = server.Object{TypeInd: 32768, ObjClass: 0x02000000, ObjFlags: 0x100}
	beforeUnit, beforeA, beforeB := *u, *a, *b
	s.Rand.Logic = prand.New(81)
	wantRNG := prand.New(81)
	want := []*server.Object{a, b}[wantRNG.IntClamp(0, 1)]
	questLoseArmorCEntry54CD30(u)
	if outer.deletes != 1 || outer.deleted != want || s.Rand.Logic.Index() != wantRNG.Index() {
		t.Fatalf("delete = %d/%p, want %p", outer.deletes, outer.deleted, want)
	}
	if *u != beforeUnit || *a != beforeA || *b != beforeB {
		t.Fatal("native callback changed unrelated object fields")
	}
	u.InvFirstItem = nil
	s.Rand.Logic = nil
	questLoseArmorCEntry54CD30(u)
	if outer.deletes != 1 {
		t.Fatal("empty inventory called delayed deletion")
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, pointer := range []unsafe.Pointer{unsafe.Pointer(u), unsafe.Pointer(a), unsafe.Pointer(b)} {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("C-owned pointer %p must exceed 4 GiB", pointer)
			}
		}
	}
}
