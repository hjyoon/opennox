package server

import (
	"math"
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

func assertMeleeEvents549380(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("events = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("events = %v, want %v", got, want)
		}
	}
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

func TestMonsterStrikeDefault549380NoTargetReturnsOne(t *testing.T) {
	unit := &Object{}
	update := &MonsterUpdateData{}
	var events []string
	got := monsterStrikeDefault549380(unit, monsterStrikeDefaultHooks549380{
		loadUpdate: func(got *Object) *MonsterUpdateData {
			if got != unit {
				t.Fatalf("update unit = %p, want %p", got, unit)
			}
			events = append(events, "load-update")
			return update
		},
		pickTarget: func(got *Object, allowFriendly bool) *Object {
			if got != unit || allowFriendly {
				t.Fatalf("pick args = %p/%v", got, allowFriendly)
			}
			events = append(events, "pick")
			return nil
		},
	})
	if got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	assertMeleeEvents549380(t, events, []string{"load-update", "pick"})
}

func TestMonsterStrikeDefault549380TraceFailureOrder(t *testing.T) {
	unit := &Object{}
	target := &Object{}
	update := &MonsterUpdateData{}
	var events []string
	got := monsterStrikeDefault549380(unit, monsterStrikeDefaultHooks549380{
		loadUpdate: func(*Object) *MonsterUpdateData {
			events = append(events, "load-update")
			return update
		},
		pickTarget: func(*Object, bool) *Object {
			events = append(events, "pick")
			return target
		},
		loadUnitY: func(*Object) float32 {
			events = append(events, "unit-y")
			return 22
		},
		loadUnitX: func(*Object) float32 {
			events = append(events, "unit-x")
			return 11
		},
		loadTargetX: func(*Object) float32 {
			events = append(events, "target-x")
			return 33
		},
		loadTargetY: func(*Object) float32 {
			events = append(events, "target-y")
			return 44
		},
		trace: func(from, to types.Pointf, flags MapTraceFlags) int32 {
			events = append(events, "trace")
			if from != (types.Ptf(11, 22)) || to != (types.Ptf(33, 44)) || flags != MapTraceFlags(5) {
				t.Fatalf("trace = %v/%v/%d", from, to, flags)
			}
			return 0
		},
	})
	if got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	assertMeleeEvents549380(t, events, []string{
		"load-update", "pick", "unit-y", "unit-x", "target-x", "target-y", "trace",
	})
}

func TestMonsterStrikeDefault549380LiveReloadAndFieldOrder(t *testing.T) {
	unit := &Object{ObjClass: object.ClassMonster, PosVec: types.Ptf(1, 2)}
	target := &Object{PosVec: types.Ptf(3, 4)}
	firstDef := &MonsterDef{
		MeleeAttackDamage116:     0xfffffffe,
		MeleeAttackImpact120:     1,
		MeleeAttackDamageType124: 0x80000007,
	}
	secondDef := &MonsterDef{MeleeAttackImpact120: 3.5}
	cachedUpdate := &MonsterUpdateData{MonsterDef: firstDef}
	replacementUpdate := &MonsterUpdateData{}
	unit.UpdateData = unsafe.Pointer(cachedUpdate)
	var events []string
	defLoads := 0
	got := monsterStrikeDefault549380(unit, monsterStrikeDefaultHooks549380{
		loadUpdate: func(got *Object) *MonsterUpdateData {
			events = append(events, "load-update")
			if got != unit {
				t.Fatalf("update unit = %p, want %p", got, unit)
			}
			return got.UpdateDataMonster()
		},
		pickTarget: func(got *Object, allowFriendly bool) *Object {
			events = append(events, "pick")
			if got != unit || allowFriendly {
				t.Fatalf("pick args = %p/%v", got, allowFriendly)
			}
			unit.UpdateData = unsafe.Pointer(replacementUpdate)
			unit.PosVec = types.Ptf(11, 22)
			target.PosVec = types.Ptf(33, 44)
			return target
		},
		loadUnitY: func(got *Object) float32 {
			events = append(events, "unit-y")
			return got.PosVec.Y
		},
		loadUnitX: func(got *Object) float32 {
			events = append(events, "unit-x")
			return got.PosVec.X
		},
		loadTargetX: func(got *Object) float32 {
			events = append(events, "target-x")
			return got.PosVec.X
		},
		loadTargetY: func(got *Object) float32 {
			events = append(events, "target-y")
			return got.PosVec.Y
		},
		trace: func(from, to types.Pointf, flags MapTraceFlags) int32 {
			events = append(events, "trace")
			if from != (types.Ptf(11, 22)) || to != (types.Ptf(33, 44)) || flags != MapTraceFlags(5) {
				t.Fatalf("trace = %v/%v/%d", from, to, flags)
			}
			return 1
		},
		loadMonsterDef: func(got *MonsterUpdateData) *MonsterDef {
			defLoads++
			events = append(events, []string{"load-def-1", "load-def-2"}[defLoads-1])
			if got != cachedUpdate {
				t.Fatalf("update reload = %p, want cached %p", got, cachedUpdate)
			}
			return got.MonsterDef
		},
		loadDamageType: func(got *MonsterDef) uint32 {
			events = append(events, "damage-type")
			if got != firstDef {
				t.Fatalf("damage type def = %p, want %p", got, firstDef)
			}
			return got.MeleeAttackDamageType124
		},
		loadDamage: func(got *MonsterDef) uint32 {
			events = append(events, "damage-value")
			if got != firstDef {
				t.Fatalf("damage def = %p, want %p", got, firstDef)
			}
			return got.MeleeAttackDamage116
		},
		damage: func(gotTarget, source, attacker *Object, damage, damageType uint32) {
			events = append(events, "damage")
			if gotTarget != target || source != unit || attacker != unit || damage != 0xfffffffe || damageType != 0x80000007 {
				t.Fatalf("damage args = %p/%p/%p/%#x/%#x", gotTarget, source, attacker, damage, damageType)
			}
			cachedUpdate.MonsterDef = secondDef
			unit.PosVec = types.Ptf(77, 88)
		},
		loadImpact: func(got *MonsterDef) float32 {
			events = append(events, "impact")
			if got != secondDef {
				t.Fatalf("impact def = %p, want %p", got, secondDef)
			}
			return got.MeleeAttackImpact120
		},
		applyForce: func(gotUnit, gotTarget *Object, impact float32) {
			events = append(events, "force")
			if gotUnit != unit || gotTarget != target || gotUnit.PosVec != (types.Ptf(77, 88)) || impact != 3.5 {
				t.Fatalf("force args = %p/%p/%v/%g", gotUnit, gotTarget, gotUnit.PosVec, impact)
			}
		},
	})
	if got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	assertMeleeEvents549380(t, events, []string{
		"load-update", "pick", "unit-y", "unit-x", "target-x", "target-y", "trace",
		"load-def-1", "damage-type", "damage-value", "damage", "load-def-2", "impact", "force",
	})
}

func TestMonsterStrikeDefault549380ImpactIsOrderedStrictPositive(t *testing.T) {
	tests := []struct {
		name      string
		impact    float32
		wantForce bool
	}{
		{name: "negative", impact: -1},
		{name: "negative-zero", impact: math.Float32frombits(0x80000000)},
		{name: "positive-zero", impact: 0},
		{name: "nan", impact: math.Float32frombits(0x7fc00000)},
		{name: "positive-subnormal", impact: math.Float32frombits(1), wantForce: true},
		{name: "positive-infinity", impact: float32(math.Inf(1)), wantForce: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			unit := &Object{}
			target := &Object{}
			def := &MonsterDef{}
			update := &MonsterUpdateData{MonsterDef: def}
			damageCalls := 0
			forceCalls := 0
			got := monsterStrikeDefault549380(unit, monsterStrikeDefaultHooks549380{
				loadUpdate:     func(*Object) *MonsterUpdateData { return update },
				pickTarget:     func(*Object, bool) *Object { return target },
				loadUnitY:      func(*Object) float32 { return 0 },
				loadUnitX:      func(*Object) float32 { return 0 },
				loadTargetX:    func(*Object) float32 { return 0 },
				loadTargetY:    func(*Object) float32 { return 0 },
				trace:          func(types.Pointf, types.Pointf, MapTraceFlags) int32 { return 1 },
				loadMonsterDef: func(got *MonsterUpdateData) *MonsterDef { return got.MonsterDef },
				loadDamageType: func(*MonsterDef) uint32 { return 0 },
				loadDamage:     func(*MonsterDef) uint32 { return 0 },
				damage:         func(*Object, *Object, *Object, uint32, uint32) { damageCalls++ },
				loadImpact:     func(*MonsterDef) float32 { return tc.impact },
				applyForce:     func(*Object, *Object, float32) { forceCalls++ },
			})
			wantForceCalls := 0
			if tc.wantForce {
				wantForceCalls = 1
			}
			if got != 1 || damageCalls != 1 || forceCalls != wantForceCalls {
				t.Fatalf("result/damage/force = %d/%d/%d, want 1/1/%d", got, damageCalls, forceCalls, wantForceCalls)
			}
		})
	}
}

func TestMonsterStrikeDefault549380FaultPrefixes(t *testing.T) {
	sequence := []string{
		"load-update", "pick", "unit-y", "unit-x", "target-x", "target-y", "trace",
		"load-def-1", "damage-type", "damage-value", "damage", "load-def-2", "impact", "force",
	}
	for failIndex, failAt := range sequence {
		t.Run(failAt, func(t *testing.T) {
			unit := &Object{}
			target := &Object{}
			def := &MonsterDef{}
			update := &MonsterUpdateData{MonsterDef: def}
			var events []string
			sentinel := &struct{}{}
			mark := func(event string) {
				events = append(events, event)
				if event == failAt {
					panic(sentinel)
				}
			}
			defLoads := 0
			hooks := monsterStrikeDefaultHooks549380{
				loadUpdate: func(*Object) *MonsterUpdateData {
					mark("load-update")
					return update
				},
				pickTarget: func(*Object, bool) *Object {
					mark("pick")
					return target
				},
				loadUnitY:   func(*Object) float32 { mark("unit-y"); return 0 },
				loadUnitX:   func(*Object) float32 { mark("unit-x"); return 0 },
				loadTargetX: func(*Object) float32 { mark("target-x"); return 0 },
				loadTargetY: func(*Object) float32 { mark("target-y"); return 0 },
				trace: func(types.Pointf, types.Pointf, MapTraceFlags) int32 {
					mark("trace")
					return 1
				},
				loadMonsterDef: func(*MonsterUpdateData) *MonsterDef {
					defLoads++
					if defLoads == 1 {
						mark("load-def-1")
					} else {
						mark("load-def-2")
					}
					return def
				},
				loadDamageType: func(*MonsterDef) uint32 { mark("damage-type"); return 0 },
				loadDamage:     func(*MonsterDef) uint32 { mark("damage-value"); return 0 },
				damage:         func(*Object, *Object, *Object, uint32, uint32) { mark("damage") },
				loadImpact:     func(*MonsterDef) float32 { mark("impact"); return 1 },
				applyForce:     func(*Object, *Object, float32) { mark("force") },
			}
			func() {
				defer func() {
					if recovered := recover(); recovered != sentinel {
						t.Fatalf("panic = %#v, want sentinel", recovered)
					}
				}()
				monsterStrikeDefault549380(unit, hooks)
				t.Fatal("strike returned instead of faulting")
			}()
			assertMeleeEvents549380(t, events, sequence[:failIndex+1])
		})
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
