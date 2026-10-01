package legacy

import (
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/strman"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type questLoseGemsLegacyServer54D080 struct {
	Server
	native *server.Server
	delete func(*server.Object)
}

func (s *questLoseGemsLegacyServer54D080) S() *server.Server                 { return s.native }
func (s *questLoseGemsLegacyServer54D080) DelayedDelete(item *server.Object) { s.delete(item) }

func TestQuestLoseGems54D080NativeCGoRoundTripAndLiveGoldBinding(t *testing.T) {
	s := server.New(nil, nil, strman.New())
	t.Cleanup(s.Close)
	outer := &questLoseGemsLegacyServer54D080{native: s}
	oldServer := GetServer
	GetServer = func() Server { return outer }
	t.Cleanup(func() { GetServer = oldServer })
	types := questGemTypeCache54D080()
	oldTypes := [3]uint32{*types[0], *types[1], *types[2]}
	t.Cleanup(func() {
		for slot := range types {
			*types[slot] = oldTypes[slot]
		}
	})
	for slot := range types {
		*types[slot] = uint32(11 + slot)
	}
	u, freeU := alloc.New(server.Object{})
	a, freeA := alloc.New(server.Object{})
	b, freeB := alloc.New(server.Object{})
	c, freeC := alloc.New(server.Object{})
	d, freeD := alloc.New(server.Object{})
	e, freeE := alloc.New(server.Object{})
	f, freeF := alloc.New(server.Object{})
	bypass, freeBypass := alloc.New(server.Object{})
	update, freeUpdate := alloc.New(server.PlayerUpdateData{})
	replacement, freeReplacement := alloc.New(server.PlayerUpdateData{})
	p, freeP := alloc.New(server.Player{})
	changed, freeChanged := alloc.New(server.Player{})
	for _, free := range []func(){freeU, freeA, freeB, freeC, freeD, freeE, freeF, freeBypass, freeUpdate, freeReplacement, freeP, freeChanged} {
		t.Cleanup(free)
	}
	p.GoldVal, changed.GoldVal = 37, math.MaxUint32-199
	update.Player, replacement.Player = p, changed
	*u = server.Object{UpdateData: unsafe.Pointer(update), InvFirstItem: a, Worth: 0xaabbccdd}
	*a = server.Object{TypeInd: 11, Worth: 1001, InvNextItem: b}
	*b = server.Object{TypeInd: 11, Worth: 999, InvNextItem: c}
	*c = server.Object{TypeInd: 11, Worth: 777, InvNextItem: d}
	*d = server.Object{TypeInd: 12, Worth: 123, InvNextItem: e}
	*e = server.Object{TypeInd: 12, Worth: 456, InvNextItem: f}
	*f = server.Object{TypeInd: 13, Worth: math.MaxUint32}
	*bypass = server.Object{TypeInd: 11, Worth: 10000}
	beforeU, beforeA, beforeB, beforeC, beforeD, beforeE, beforeF := *u, *a, *b, *c, *d, *e, *f
	var deleted []*server.Object
	outer.delete = func(item *server.Object) {
		deleted = append(deleted, item)
		if len(deleted) == 1 {
			if item != a || p.GoldVal != 37 || changed.GoldVal != math.MaxUint32-199 {
				t.Fatal("first odd credit ran before its deletion")
			}
			u.UpdateData = unsafe.Pointer(replacement)
		} else if changed.GoldVal != 300 || p.GoldVal != 37 {
			t.Fatalf("positive half-price did not use fresh unit update: old=%d live=%d", p.GoldVal, changed.GoldVal)
		}
		item.InvNextItem = nil
		u.InvFirstItem = bypass // Must not restart or follow the changed link.
	}
	s.Rand.Logic = nil
	questLoseGemsCEntry54D080(u)
	if !reflect.DeepEqual(deleted, []*server.Object{a, b, d, f}) || p.GoldVal != 37 || changed.GoldVal != 3221225772 {
		t.Fatalf("deleted=%v old gold=%d live gold=%d", deleted, p.GoldVal, changed.GoldVal)
	}
	beforeU.UpdateData, beforeU.InvFirstItem = unsafe.Pointer(replacement), bypass
	beforeA.InvNextItem, beforeB.InvNextItem, beforeD.InvNextItem = nil, nil, nil
	if *u != beforeU || *a != beforeA || *b != beforeB || *c != beforeC || *d != beforeD || *e != beforeE || *f != beforeF ||
		*types[0] != 11 || *types[1] != 12 || *types[2] != 13 {
		t.Fatal("native entry changed fields outside explicit callback/gold mutations")
	}
	// Initialized-cache, empty inventory needs neither player nor deletion.
	u.UpdateData, u.InvFirstItem, outer.delete = nil, nil, nil
	questLoseGemsCEntry54D080(u)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, pointer := range []unsafe.Pointer{
			unsafe.Pointer(u), unsafe.Pointer(a), unsafe.Pointer(b), unsafe.Pointer(c), unsafe.Pointer(d), unsafe.Pointer(e), unsafe.Pointer(f), unsafe.Pointer(bypass),
			unsafe.Pointer(update), unsafe.Pointer(replacement), unsafe.Pointer(p), unsafe.Pointer(changed),
		} {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("C-owned pointer %p must exceed 4 GiB", pointer)
			}
		}
	}
}
