package server

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

func meleeMonsterTestObject532130(t *testing.T) *Object {
	t.Helper()
	unit := monsterActionTestObject50A910(t)
	update := unit.UpdateDataMonster()
	update.AIStackInd = 0
	update.AIStack[0].Action = uint32(ai.ACTION_MELEE_ATTACK)
	update.MonsterDef = &MonsterDef{
		MeleeAttackFrame108:          2,
		MeleeAttackRange112:          15,
		MeleeAttackDamage116:         6,
		MeleeAttackImpact120:         4,
		MeleeAttackDamageType124:     7,
		MeleeAttackDelayMin128:       20,
		MeleeAttackDelayMax132:       30,
		MeleeAttackPoisonChange136:   25,
		MeleeAttackPoisonStrength140: 2,
		MeleeAttackPoisonMax144:      5,
	}
	return unit
}

func TestMonsterActionMeleeStart532130AttackStateAndOrder(t *testing.T) {
	unit := meleeMonsterTestObject532130(t)
	update := unit.UpdateDataMonster()
	var sounds [17]uint32
	sounds[6] = 606
	update.SoundSet122 = unsafe.Pointer(&sounds[0])
	var events []string
	handled := monsterActionMeleeStart532130(unit, monsterActionMeleeStartHooks532130{
		frame: func() uint32 { return 100 },
		random: func(minimum, maximum int) int {
			if minimum != 20 || maximum != 30 {
				t.Fatalf("random bounds = %d..%d", minimum, maximum)
			}
			events = append(events, "random")
			return 24
		},
		buffOff: func(got *Object, enchant EnchantID) {
			if got != unit {
				t.Fatalf("buff unit = %p", got)
			}
			events = append(events, "buff:"+string(rune(enchant)))
		},
		audio: func(id uint32, got *Object) {
			if id != 606 || got != unit || unit.Field34 != 100 || update.Field128 != 124 {
				t.Fatalf("audio state = %d/%p/%d/%d", id, got, unit.Field34, update.Field128)
			}
			events = append(events, "audio")
		},
		push: func(ai.ActionType, ...any) *AIStackItem { t.Fatal("ready attack pushed an action"); return nil },
	})
	if !handled {
		t.Fatal("non-NPC melee start was not handled")
	}
	want := []string{"buff:\x00", "buff:\x17", "random", "audio"}
	if len(events) != len(want) {
		t.Fatalf("events = %q, want %q", events, want)
	}
	for i := range want {
		if events[i] != want[i] {
			t.Fatalf("events = %q, want %q", events, want)
		}
	}
}

func TestMonsterActionMeleeStart532130CooldownStack(t *testing.T) {
	unit := meleeMonsterTestObject532130(t)
	target := &Object{}
	update := unit.UpdateDataMonster()
	update.CurrentEnemy = target
	update.Field128 = 150
	var actions []ai.ActionType
	var pushed []*AIStackItem
	handled := monsterActionMeleeStart532130(unit, monsterActionMeleeStartHooks532130{
		frame:  func() uint32 { return 100 },
		random: func(int, int) int { t.Fatal("cooldown branch used RNG"); return 0 },
		push: func(action ai.ActionType, args ...any) *AIStackItem {
			actions = append(actions, action)
			item := &AIStackItem{Action: uint32(action)}
			item.SetArgs(args...)
			pushed = append(pushed, item)
			return item
		},
	})
	if !handled || len(actions) != 2 || actions[0] != ai.DEPENDENCY_OBJECT_CLOSER_THAN || actions[1] != ai.ACTION_WAIT {
		t.Fatalf("cooldown actions = %v", actions)
	}
	if pushed[0].ArgF32(0) != 18 || pushed[0].Args[2] != uintptr(unsafe.Pointer(target)) || pushed[1].ArgU32(0) != 150 {
		t.Fatalf("cooldown args = %#v", pushed)
	}
}

func TestMonsterActionMeleeUpdate532440StrikeAndCompletion(t *testing.T) {
	unit := meleeMonsterTestObject532130(t)
	update := unit.UpdateDataMonster()
	marker := unsafe.Pointer(unit)
	update.MonsterDef.MeleeStrikeFunc236 = marker
	update.Field120_1 = 2
	update.Field120_2 = 0
	update.Field120_3 = 1
	var sounds [17]uint32
	sounds[8] = 808
	update.SoundSet122 = unsafe.Pointer(&sounds[0])
	var events []string
	handled := monsterActionMeleeUpdate532440(unit, monsterActionMeleeUpdateHooks532440{
		canStrike: func(got unsafe.Pointer) bool {
			events = append(events, "can")
			return got == marker
		},
		strike: func(got *Object, fn unsafe.Pointer) int {
			if got != unit || fn != marker {
				t.Fatalf("strike args = %p/%p", got, fn)
			}
			events = append(events, "strike")
			return 1
		},
		audio: func(id uint32, got *Object) {
			if id != 808 || got != unit {
				t.Fatalf("audio = %d/%p", id, got)
			}
			events = append(events, "audio")
		},
		pop: func() int { events = append(events, "pop"); return 0 },
	})
	want := []string{"can", "strike", "audio", "pop"}
	if !handled || len(events) != len(want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	for i := range want {
		if events[i] != want[i] {
			t.Fatalf("events = %v, want %v", events, want)
		}
	}
}

func TestMonsterActionMeleeUpdate532440RejectsUnknownStrike(t *testing.T) {
	unit := meleeMonsterTestObject532130(t)
	marker := unsafe.Pointer(unit)
	unit.UpdateDataMonster().MonsterDef.MeleeStrikeFunc236 = marker
	popped := false
	if monsterActionMeleeUpdate532440(unit, monsterActionMeleeUpdateHooks532440{
		canStrike: func(unsafe.Pointer) bool { return false },
		strike:    func(*Object, unsafe.Pointer) int { t.Fatal("unknown strike was called"); return 0 },
		pop:       func() int { popped = true; return 0 },
	}) {
		t.Fatal("unknown strike was handled")
	}
	if popped {
		t.Fatal("unknown strike mutated the stack")
	}
}

func TestMonsterPickMeleeTarget549440FacingRangeAndNearestEdge(t *testing.T) {
	unit := meleeMonsterTestObject532130(t)
	unit.PosVec = types.Ptf(100, 100)
	unit.Direction1 = DirFromVec(types.Ptf(1, 0))
	unit.Shape.Circle.R = 4
	behind := &Object{ObjClass: object.ClassPlayer, PosVec: types.Ptf(90, 100)}
	friend := &Object{ObjClass: object.ClassPlayer, PosVec: types.Ptf(105, 100)}
	farCenterLarge := &Object{ObjClass: object.ClassPlayer, PosVec: types.Ptf(112, 100)}
	farCenterLarge.Shape.Circle.R = 8
	nearCenterSmall := &Object{ObjClass: object.ClassPlayer, PosVec: types.Ptf(108, 100)}
	nearCenterSmall.Shape.Circle.R = 1
	candidates := []*Object{behind, friend, nearCenterSmall, farCenterLarge}
	var gotRect types.Rectf
	target := monsterPickMeleeTarget549440(unit, false, monsterPickMeleeTargetHooks549440{
		eachInRect: func(rect types.Rectf, fn func(*Object) bool) {
			gotRect = rect
			for _, candidate := range candidates {
				fn(candidate)
			}
		},
		isEnemy: func(_ *Object, candidate *Object) bool { return candidate != friend },
	})
	if target != farCenterLarge {
		t.Fatalf("target = %p, want large-radius target %p", target, farCenterLarge)
	}
	if gotRect.Min != (types.Ptf(51, 51)) || gotRect.Max != (types.Ptf(149, 149)) {
		t.Fatalf("search rect = %v", gotRect)
	}
}

func TestMonsterStrikeSpider549BC0DamageForcePoisonOrder(t *testing.T) {
	unit := meleeMonsterTestObject532130(t)
	unit.PosVec = types.Ptf(10, 20)
	target := &Object{PosVec: types.Ptf(30, 20)}
	var events []string
	got := monsterStrikeSpider549BC0(unit, monsterStrikeSpiderHooks549BC0{
		pickTarget: func(got *Object, allowFriendly bool) *Object {
			if got != unit || allowFriendly {
				t.Fatalf("pick args = %p/%v", got, allowFriendly)
			}
			events = append(events, "pick")
			return target
		},
		trace: func(from, to types.Pointf, flags MapTraceFlags) bool {
			if from != unit.PosVec || to != target.PosVec || flags != MapTraceFlags(5) {
				t.Fatalf("trace = %v/%v/%d", from, to, flags)
			}
			events = append(events, "trace")
			return true
		},
		damage: func(gotTarget, source, attacker *Object, damage int, damageType object.DamageType) bool {
			if gotTarget != target || source != unit || attacker != unit || damage != 6 || damageType != 7 {
				t.Fatalf("damage args = %p/%p/%p/%d/%d", gotTarget, source, attacker, damage, damageType)
			}
			events = append(events, "damage")
			return false
		},
		applyForce: func(got *Object, origin types.Pointf, force float64) {
			if got != target || origin != unit.PosVec || force != 4 {
				t.Fatalf("force args = %p/%v/%g", got, origin, force)
			}
			events = append(events, "force")
		},
		random: func(minimum, maximum int) int {
			if minimum != 1 || maximum != 100 {
				t.Fatalf("random = %d..%d", minimum, maximum)
			}
			events = append(events, "random")
			return 25
		},
		activatePoison: func(got *Object, increment, maximum int32) int32 {
			if got != target || increment != 2 || maximum != 5 {
				t.Fatalf("poison args = %p/%d/%d", got, increment, maximum)
			}
			events = append(events, "poison")
			return 1
		},
		priorityMessage: func(got *Object, id strman.ID, value byte) {
			if got != target || id != "aifunc.c:Poisoned" || value != 0 {
				t.Fatalf("message args = %p/%q/%d", got, id, value)
			}
			events = append(events, "message")
		},
	})
	want := []string{"pick", "trace", "damage", "force", "random", "poison", "message"}
	if got != 1 || len(events) != len(want) {
		t.Fatalf("result/events = %d/%v, want 1/%v", got, events, want)
	}
	for i := range want {
		if events[i] != want[i] {
			t.Fatalf("events = %v, want %v", events, want)
		}
	}
}
