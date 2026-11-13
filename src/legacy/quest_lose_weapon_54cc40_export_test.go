package legacy

import (
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/player"
	"github.com/opennox/libs/strman"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type questLoseWeaponLegacyServer54CC40 struct {
	Server
	native *server.Server
	delete func(*server.Object)
}

func (s *questLoseWeaponLegacyServer54CC40) S() *server.Server { return s.native }
func (s *questLoseWeaponLegacyServer54CC40) DelayedDelete(item *server.Object) {
	s.delete(item)
}

func TestQuestLoseWeapon54CC40NativeCGoRoundTrip(t *testing.T) {
	s := server.NewHigh(nil, nil, strman.New())
	t.Cleanup(s.Close)
	outer := &questLoseWeaponLegacyServer54CC40{native: s}
	oldServer, oldCanUse := GetServer, Nox_xxx_playerClassCanUseItem_57B3D0
	GetServer = func() Server { return outer }
	t.Cleanup(func() {
		GetServer, Nox_xxx_playerClassCanUseItem_57B3D0 = oldServer, oldCanUse
	})
	u, freeU := alloc.NewHigh(server.Object{})
	w, freeW := alloc.NewHigh(server.Object{})
	a, freeA := alloc.NewHigh(server.Object{})
	b, freeB := alloc.NewHigh(server.Object{})
	tail, freeTail := alloc.NewHigh(server.Object{})
	bypassed, freeBypassed := alloc.NewHigh(server.Object{})
	update, freeUpdate := alloc.NewHigh(server.PlayerUpdateData{})
	replacement, freeReplacement := alloc.NewHigh(server.PlayerUpdateData{})
	p, freeP := alloc.NewHigh(server.Player{})
	changed, freeChanged := alloc.NewHigh(server.Player{})
	replaced, freeReplaced := alloc.NewHigh(server.Player{})
	data, freeData := alloc.NewHigh(server.ModifierInitData{})
	mod, freeMod := alloc.NewHigh(server.ModifierEff{})
	for _, free := range []func(){freeU, freeW, freeA, freeB, freeTail, freeBypassed, freeUpdate, freeReplacement, freeP, freeChanged, freeReplaced, freeData, freeMod} {
		t.Cleanup(free)
	}
	p.Info().SetPlayerClass(player.Class(3))
	changed.Info().SetPlayerClass(player.Class(255))
	replaced.Info().SetPlayerClass(player.Class(17))
	update.Player, replacement.Player = p, replaced
	data.Modifiers[0], data.Field16 = mod, 0xfedcba98
	*u = server.Object{UpdateData: unsafe.Pointer(update), InvFirstItem: w, Worth: 0xaabbccdd}
	*w = server.Object{ObjFlags: 0x80000100, ObjClass: 0x1000, ObjSubClass: 0x104, InitData: unsafe.Pointer(data), InvNextItem: a}
	*a = server.Object{ObjClass: 0x01000000, InvNextItem: b}
	*b = server.Object{ObjClass: 0x1000, ObjSubClass: 0x10002, InvNextItem: bypassed}
	*tail, *bypassed = server.Object{ObjClass: 0x1000}, server.Object{ObjClass: 0x1000}
	beforeU, beforeW, beforeA, beforeB, beforeTail, beforeData := *u, *w, *a, *b, *tail, *data
	var items []*server.Object
	var classes []player.Class
	Nox_xxx_playerClassCanUseItem_57B3D0 = func(item *server.Object, class player.Class) bool {
		items, classes = append(items, item), append(classes, class)
		switch item {
		case a:
			u.UpdateData = unsafe.Pointer(replacement)
			update.Player = changed
			u.InvFirstItem = bypassed
			return false
		case b:
			b.InvNextItem = tail
			return true
		case tail:
			return false
		default:
			t.Errorf("unexpected spare %p", item)
			return false
		}
	}
	deletes := 0
	outer.delete = func(item *server.Object) {
		deletes++
		if item != w || len(items) != 3 {
			t.Errorf("delete=%p after %d class checks, want %p/3", item, len(items), w)
		}
	}
	s.Rand.Logic = nil
	questLoseWeaponCEntry54CC40(u)
	if !reflect.DeepEqual(items, []*server.Object{a, b, tail}) || !reflect.DeepEqual(classes, []player.Class{3, 255, 255}) || deletes != 1 {
		t.Fatalf("items=%v classes=%v deletes=%d", items, classes, deletes)
	}
	beforeU.UpdateData, beforeU.InvFirstItem = unsafe.Pointer(replacement), bypassed
	beforeB.InvNextItem = tail
	if *u != beforeU || *w != beforeW || *a != beforeA || *b != beforeB || *tail != beforeTail || *data != beforeData {
		t.Fatal("native entry changed fields outside explicit callback mutations")
	}
	// These returns do not need player/class/deletion bindings.
	Nox_xxx_playerClassCanUseItem_57B3D0, outer.delete = nil, nil
	u.UpdateData = nil
	u.InvFirstItem = w
	w.ObjSubClass = 0x10104
	w.InitData = nil
	questLoseWeaponCEntry54CC40(u)
	u.InvFirstItem = nil
	questLoseWeaponCEntry54CC40(u)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, pointer := range []unsafe.Pointer{
			unsafe.Pointer(u), unsafe.Pointer(w), unsafe.Pointer(a), unsafe.Pointer(b), unsafe.Pointer(tail), unsafe.Pointer(bypassed),
			unsafe.Pointer(update), unsafe.Pointer(replacement), unsafe.Pointer(p), unsafe.Pointer(changed), unsafe.Pointer(replaced), unsafe.Pointer(data), unsafe.Pointer(mod),
		} {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("C-owned pointer %p must exceed 4 GiB", pointer)
			}
		}
	}
}
