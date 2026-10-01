package server

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/player"
)

func TestQuestLoseWeapon54CC40NativeCachedUpdateLivePlayerAndNext(t *testing.T) {
	var s Server // This helper neither needs nor advances an RNG.
	p, changed, replaced := &Player{}, &Player{}, &Player{}
	p.Info().SetPlayerClass(player.Class(3))
	changed.Info().SetPlayerClass(player.Class(255))
	replaced.Info().SetPlayerClass(player.Class(17))
	update, replacement := &PlayerUpdateData{Player: p}, &PlayerUpdateData{Player: replaced}
	mod := &ModifierEff{}
	data := &ModifierInitData{Modifiers: [4]*ModifierEff{nil, nil, nil, mod}, Field16: 0xaabbccdd}
	tail, bypassed := &Object{ObjClass: 0x1000}, &Object{ObjClass: 0x1000}
	b := &Object{ObjClass: 0x01000000, ObjSubClass: 0x10002, InvNextItem: bypassed}
	a := &Object{ObjClass: 0x1000, InvNextItem: b}
	w := &Object{ObjClass: 0x81001000, ObjFlags: 0x80000100, ObjSubClass: 0x104, InitData: unsafe.Pointer(data), InvNextItem: a}
	u := &Object{UpdateData: unsafe.Pointer(update), InvFirstItem: w, Worth: 0x12345678}
	beforeW, beforeA, beforeB, beforeTail, beforeData := *w, *a, *b, *tail, *data
	var items []*Object
	var classes []uint8
	deletes := 0
	s.QuestLoseWeapon54CC40(u, func(item *Object, class uint8) int32 {
		items, classes = append(items, item), append(classes, class)
		switch item {
		case a:
			u.UpdateData = unsafe.Pointer(replacement)
			update.Player = changed // The cached update's Player link is live.
			u.InvFirstItem = bypassed
			return 1
		case b:
			b.InvNextItem = tail // Even after a match, the next link is read live.
			return 2             // The original accepts exactly 1, not any nonzero value.
		case tail:
			return -1
		default:
			t.Fatalf("unexpected spare %p", item)
			return 0
		}
	}, func(item *Object) {
		deletes++
		if item != w || len(items) != 3 {
			t.Fatalf("delete=%p after %d spare checks, want %p/3", item, len(items), w)
		}
	})
	if !reflect.DeepEqual(items, []*Object{a, b, tail}) || !reflect.DeepEqual(classes, []uint8{3, 255, 255}) || deletes != 1 {
		t.Fatalf("items=%v classes=%v deletes=%d", items, classes, deletes)
	}
	beforeB.InvNextItem = tail
	if *w != beforeW || *a != beforeA || *b != beforeB || *tail != beforeTail || *data != beforeData || u.Worth != 0x12345678 {
		t.Fatal("helper changed fields outside explicit callback mutations")
	}
}

func TestQuestLoseWeapon54CC40NativeMissingBindings(t *testing.T) {
	var s Server
	for _, tc := range []struct {
		name   string
		setup  func() *Object
		canUse func(*Object, uint8) int32
		delete func(*Object)
		panics bool
	}{
		{"nil-unit", func() *Object { return nil }, nil, nil, true},
		{"empty-no-update", func() *Object { return &Object{} }, nil, nil, false},
		{"protected-no-modifier-data", func() *Object {
			return &Object{InvFirstItem: &Object{ObjClass: 0x1000, ObjFlags: 0x100, ObjSubClass: 0x10104}}
		}, nil, nil, false},
		{"plain-all-four-modifiers-nil", func() *Object {
			data := &ModifierInitData{}
			return &Object{InvFirstItem: &Object{ObjClass: 0x1000, ObjFlags: 0x100, ObjSubClass: 4, InitData: unsafe.Pointer(data)}}
		}, nil, nil, false},
		{"no-spare-no-update", func() *Object {
			return &Object{InvFirstItem: &Object{ObjClass: 0x1000, ObjFlags: 0x100}}
		}, nil, nil, false},
		{"missing-modifier-data", func() *Object {
			return &Object{InvFirstItem: &Object{ObjClass: 0x1000, ObjFlags: 0x100, ObjSubClass: 4}}
		}, nil, nil, true},
		{"missing-update", func() *Object {
			spare := &Object{ObjClass: 0x1000}
			return &Object{InvFirstItem: &Object{ObjClass: 0x1000, ObjFlags: 0x100, InvNextItem: spare}}
		}, nil, nil, true},
		{"missing-player", func() *Object {
			update := &PlayerUpdateData{}
			spare := &Object{ObjClass: 0x1000}
			return &Object{UpdateData: unsafe.Pointer(update), InvFirstItem: &Object{ObjClass: 0x1000, ObjFlags: 0x100, InvNextItem: spare}}
		}, nil, nil, true},
		{"missing-can-use", questWeaponNativeFixture54CC40, nil, nil, true},
		{"no-match-no-delete", questWeaponNativeFixture54CC40, func(*Object, uint8) int32 { return 2 }, nil, false},
		{"missing-delete", questWeaponNativeFixture54CC40, func(*Object, uint8) int32 { return 1 }, nil, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if got := recover() != nil; got != tc.panics {
					t.Fatalf("panic=%v, want %v", got, tc.panics)
				}
			}()
			s.QuestLoseWeapon54CC40(tc.setup(), tc.canUse, tc.delete)
		})
	}
}

func questWeaponNativeFixture54CC40() *Object {
	update := &PlayerUpdateData{Player: &Player{}}
	spare := &Object{ObjClass: 0x1000}
	return &Object{UpdateData: unsafe.Pointer(update), InvFirstItem: &Object{ObjClass: 0x1000, ObjFlags: 0x100, InvNextItem: spare}}
}

func TestQuestLoseWeapon54CC40NativeSearchDoesNotNormalizePlayerClass(t *testing.T) {
	var s Server
	for _, class := range []uint8{0, 1, 2, 3, 127, 128, 255} {
		u := questWeaponNativeFixture54CC40()
		update := (*PlayerUpdateData)(u.UpdateData)
		update.Player.Info().SetPlayerClass(player.Class(class))
		calls := 0
		s.QuestLoseWeapon54CC40(u, func(item *Object, got uint8) int32 {
			calls++
			if item != u.InvFirstItem.InvNextItem || got != class {
				t.Fatalf("spare=%p class=%d, want class=%d", item, got, class)
			}
			return 0
		}, nil)
		if calls != 1 {
			t.Fatalf("class=%d, calls=%d", class, calls)
		}
	}
}
