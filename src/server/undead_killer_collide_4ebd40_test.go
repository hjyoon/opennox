package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type undeadKillerTestState4EBD40 struct {
	events       []string
	classLow     uint8
	subclassLow  uint8
	hp           uint16
	remaining    int32
	damageResult int32
	onHP         func(*undeadKillerTestState4EBD40)
	onParent     func(*undeadKillerTestState4EBD40)
	onDamage     func(*undeadKillerTestState4EBD40)
	onDelete     func(*undeadKillerTestState4EBD40)
	panicOnData  bool
	panicOnSpell bool
	parent       int
	damageFn     int
	data         int
	spell        int
	stored       []int32
}

func (s *undeadKillerTestState4EBD40) event(format string, args ...any) {
	s.events = append(s.events, fmt.Sprintf(format, args...))
}

func (s *undeadKillerTestState4EBD40) hooks() undeadKillerCollideHooks4EBD40[int, int, int, int] {
	return undeadKillerCollideHooks4EBD40[int, int, int, int]{
		loadClassLow: func(target int) uint8 {
			s.event("class:%d", target)
			return s.classLow
		},
		loadSubclassLow: func(target int) uint8 {
			s.event("subclass:%d", target)
			return s.subclassLow
		},
		loadCollideData: func(source int) int {
			s.event("data:%d", source)
			if s.panicOnData {
				panic("nil source collide-data load")
			}
			return s.data
		},
		loadHP: func(target int) uint16 {
			s.event("hp:%d", target)
			if s.onHP != nil {
				s.onHP(s)
			}
			return s.hp
		},
		loadSpell: func(data int) int {
			s.event("spell:%d", data)
			if s.panicOnSpell {
				panic("nil collide-data dereference")
			}
			return s.spell
		},
		loadRemaining: func(spell int) int32 {
			s.event("remaining:%d=%d", spell, s.remaining)
			return s.remaining
		},
		findParentPlayer: func(source int) int {
			s.event("parent:%d=%d", source, s.parent)
			if s.onParent != nil {
				s.onParent(s)
			}
			return s.parent
		},
		loadTargetDamage: func(target int) int {
			s.event("damage-fn:%d=%d", target, s.damageFn)
			return s.damageFn
		},
		callTargetDamage: func(fn, target, parent, source int, damage int32, damageType uint32) int32 {
			s.event("damage:%d:%d:%d:%d:%d:%d", fn, target, parent, source, damage, damageType)
			if s.onDamage != nil {
				s.onDamage(s)
			}
			return s.damageResult
		},
		delayedDelete: func(source int) {
			s.event("delete:%d", source)
			if s.onDelete != nil {
				s.onDelete(s)
			}
		},
		storeRemaining: func(spell int, value int32) {
			s.event("store:%d=%d", spell, value)
			s.remaining = value
			s.stored = append(s.stored, value)
		},
	}
}

func newEligibleUndeadKillerState4EBD40() *undeadKillerTestState4EBD40 {
	return &undeadKillerTestState4EBD40{
		classLow:     undeadKillerMonsterClassLow4EBD40,
		subclassLow:  undeadKillerUndeadSubclassLow4EBD40,
		data:         31,
		spell:        41,
		parent:       51,
		damageFn:     61,
		damageResult: math.MinInt32,
	}
}

func assertUndeadKillerEvents4EBD40(t *testing.T, got, want []string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("events:\n got %#v\nwant %#v", got, want)
	}
}

func TestUndeadKillerCollide4EBD40NilTargetCollisionContract(t *testing.T) {
	t.Run("nil target and nil collision deletes without reading source", func(t *testing.T) {
		state := newEligibleUndeadKillerState4EBD40()
		state.panicOnData = true
		undeadKillerCollide4EBD40(0, 0, 0, state.hooks())
		assertUndeadKillerEvents4EBD40(t, state.events, []string{"delete:0"})
	})

	t.Run("non-nil collision suppresses the nil-target delete", func(t *testing.T) {
		state := newEligibleUndeadKillerState4EBD40()
		undeadKillerCollide4EBD40(11, 0, 0x7fffffff, state.hooks())
		assertUndeadKillerEvents4EBD40(t, state.events, nil)
	})
}

func TestUndeadKillerCollide4EBD40TargetGuardsAreLowByteAndOrdered(t *testing.T) {
	t.Run("non-monster stops before subclass", func(t *testing.T) {
		state := newEligibleUndeadKillerState4EBD40()
		state.classLow = 0xfd
		undeadKillerCollide4EBD40(11, 21, 0, state.hooks())
		assertUndeadKillerEvents4EBD40(t, state.events, []string{"class:21"})
	})

	t.Run("non-undead stops before source data", func(t *testing.T) {
		state := newEligibleUndeadKillerState4EBD40()
		state.subclassLow = 0xbf
		state.panicOnData = true
		undeadKillerCollide4EBD40(11, 21, 0, state.hooks())
		assertUndeadKillerEvents4EBD40(t, state.events, []string{"class:21", "subclass:21"})
	})
}

func TestUndeadKillerCollide4EBD40PartialBudgetUsesEntryRemaining(t *testing.T) {
	state := newEligibleUndeadKillerState4EBD40()
	state.hp = 4
	state.remaining = 10
	state.onDamage = func(s *undeadKillerTestState4EBD40) {
		s.remaining = 1234
		s.damageFn = 999
	}
	undeadKillerCollide4EBD40(11, 21, 0x1234, state.hooks())
	if state.remaining != 6 || !reflect.DeepEqual(state.stored, []int32{6}) {
		t.Fatalf("remaining/stores = %d/%v, want 6/[6]", state.remaining, state.stored)
	}
	assertUndeadKillerEvents4EBD40(t, state.events, []string{
		"class:21", "subclass:21", "data:11", "hp:21", "spell:31", "remaining:41=10",
		"parent:11=51", "damage-fn:21=61", "damage:61:21:51:11:4:6", "store:41=6",
	})
}

func TestUndeadKillerCollide4EBD40CachesDataBeforeHPAndLoadsDamageAfterParent(t *testing.T) {
	state := newEligibleUndeadKillerState4EBD40()
	state.hp = 4
	state.remaining = 10
	state.onHP = func(s *undeadKillerTestState4EBD40) {
		s.data = 99
	}
	state.onParent = func(s *undeadKillerTestState4EBD40) {
		s.damageFn = 62
	}
	undeadKillerCollide4EBD40(11, 21, 0, state.hooks())
	assertUndeadKillerEvents4EBD40(t, state.events, []string{
		"class:21", "subclass:21", "data:11", "hp:21", "spell:31", "remaining:41=10",
		"parent:11=51", "damage-fn:21=62", "damage:62:21:51:11:4:6", "store:41=6",
	})
}

func TestUndeadKillerCollide4EBD40ConsumedBudgetUsesPostDeleteLiveRemaining(t *testing.T) {
	state := newEligibleUndeadKillerState4EBD40()
	state.hp = 5
	state.remaining = 4
	state.onDamage = func(s *undeadKillerTestState4EBD40) { s.remaining = 20 }
	state.onDelete = func(s *undeadKillerTestState4EBD40) { s.remaining = 31 }
	undeadKillerCollide4EBD40(11, 21, 0, state.hooks())
	if state.remaining != 27 || !reflect.DeepEqual(state.stored, []int32{27}) {
		t.Fatalf("remaining/stores = %d/%v, want 27/[27]", state.remaining, state.stored)
	}
	assertUndeadKillerEvents4EBD40(t, state.events, []string{
		"class:21", "subclass:21", "data:11", "hp:21", "spell:31", "remaining:41=4",
		"parent:11=51", "damage-fn:21=61", "damage:61:21:51:11:4:6", "delete:11",
		"remaining:41=31", "store:41=27",
	})
}

func TestUndeadKillerCollide4EBD40EqualBudgetUsesConsumedBranch(t *testing.T) {
	state := newEligibleUndeadKillerState4EBD40()
	state.hp = 7
	state.remaining = 7
	undeadKillerCollide4EBD40(11, 21, 0, state.hooks())
	if state.remaining != 0 {
		t.Fatalf("remaining = %d, want 0", state.remaining)
	}
	assertUndeadKillerEvents4EBD40(t, state.events, []string{
		"class:21", "subclass:21", "data:11", "hp:21", "spell:31", "remaining:41=7",
		"parent:11=51", "damage-fn:21=61", "damage:61:21:51:11:7:6", "delete:11",
		"remaining:41=7", "store:41=0",
	})
}

func TestUndeadKillerCollide4EBD40ZeroBudgetDeletesWithoutDamageOrStore(t *testing.T) {
	state := newEligibleUndeadKillerState4EBD40()
	state.hp = math.MaxUint16
	state.remaining = 0
	undeadKillerCollide4EBD40(11, 21, 0, state.hooks())
	assertUndeadKillerEvents4EBD40(t, state.events, []string{
		"class:21", "subclass:21", "data:11", "hp:21", "spell:31", "remaining:41=0", "delete:11",
	})
}

func TestUndeadKillerCollide4EBD40NegativeBudgetIsNonzeroConsumedDamage(t *testing.T) {
	state := newEligibleUndeadKillerState4EBD40()
	state.hp = 1
	state.remaining = -3
	state.onDelete = func(s *undeadKillerTestState4EBD40) { s.remaining = 5 }
	undeadKillerCollide4EBD40(11, 21, 0, state.hooks())
	if state.remaining != 8 {
		t.Fatalf("remaining = %d, want 8", state.remaining)
	}
	assertUndeadKillerEvents4EBD40(t, state.events, []string{
		"class:21", "subclass:21", "data:11", "hp:21", "spell:31", "remaining:41=-3",
		"parent:11=51", "damage-fn:21=61", "damage:61:21:51:11:-3:6", "delete:11",
		"remaining:41=5", "store:41=8",
	})
}

func TestUndeadKillerCollide4EBD40PositiveBudgetAndZeroHPPreservesBudget(t *testing.T) {
	state := newEligibleUndeadKillerState4EBD40()
	state.hp = 0
	state.remaining = 9
	undeadKillerCollide4EBD40(11, 21, 0, state.hooks())
	if state.remaining != 9 {
		t.Fatalf("remaining = %d, want 9", state.remaining)
	}
	assertUndeadKillerEvents4EBD40(t, state.events, []string{
		"class:21", "subclass:21", "data:11", "hp:21", "spell:31", "remaining:41=9",
		"parent:11=51", "damage-fn:21=61", "damage:61:21:51:11:0:6", "store:41=9",
	})
}

func TestUndeadKillerCollide4EBD40ConsumedSubtractionWrapsInt32(t *testing.T) {
	state := newEligibleUndeadKillerState4EBD40()
	state.hp = 1
	state.remaining = math.MinInt32
	state.onDelete = func(s *undeadKillerTestState4EBD40) { s.remaining = math.MaxInt32 }
	undeadKillerCollide4EBD40(11, 21, 0, state.hooks())
	if state.remaining != -1 {
		t.Fatalf("remaining = %d, want -1", state.remaining)
	}
}

func TestUndeadKillerCollide4EBD40LoadsDataBeforeHPAndSpellAfterHP(t *testing.T) {
	t.Run("source data fault precedes HP", func(t *testing.T) {
		state := newEligibleUndeadKillerState4EBD40()
		state.panicOnData = true
		defer func() {
			if recover() == nil {
				t.Fatal("expected collide-data panic")
			}
			assertUndeadKillerEvents4EBD40(t, state.events, []string{"class:21", "subclass:21", "data:0"})
		}()
		undeadKillerCollide4EBD40(0, 21, 0, state.hooks())
	})

	t.Run("collide-data dereference follows HP", func(t *testing.T) {
		state := newEligibleUndeadKillerState4EBD40()
		state.panicOnSpell = true
		defer func() {
			if recover() == nil {
				t.Fatal("expected spell panic")
			}
			assertUndeadKillerEvents4EBD40(t, state.events, []string{
				"class:21", "subclass:21", "data:11", "hp:21", "spell:31",
			})
		}()
		undeadKillerCollide4EBD40(11, 21, 0, state.hooks())
	})
}
