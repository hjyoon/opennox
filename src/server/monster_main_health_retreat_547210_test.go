package server

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/unit/ai"
)

func monsterMainHealthRetreatFixture547210(t *testing.T) (*Server, *Object, *MonsterUpdateData) {
	t.Helper()
	s, unit, update, _ := newMonsterMainFleeTest547210(t)
	oldFlags := noxflags.GetGame()
	noxflags.ResetGame()
	t.Cleanup(func() { noxflags.ResetGame(); noxflags.SetGame(oldFlags) })
	unit.ObjSubClass = 0
	unit.HealthData = &HealthData{Cur: 1, Field2: 1, Max: 25}
	update.CurrentEnemy, update.PreferredEnemy = nil, nil
	update.Aggression, update.RetreatLevel, update.FleeRange = 0.83, 0.04, 0
	update.Field363 = s.Frame() + 1
	return s, unit, update
}

func TestMonsterMainHealthRetreat547210EntryOriginalAdmissions(t *testing.T) {
	for _, mode := range []string{
		"active", "passive-rounded-ratio", "melee", "object-cast", "location-cast", "duration-cast",
		"caster-antimagic", "healthy-caster-antimagic", "noncaster-antimagic", "confused",
		"freeze-and-stun", "disabled", "destroyed", "NaN-aggression", "NaN-threshold", "bot",
		"summoned-coop", "monitor-coop", "summoned-noncoop", "monitor-noncoop",
	} {
		t.Run(mode, func(t *testing.T) {
			s, unit, update := monsterMainHealthRetreatFixture547210(t)
			want := ai.ACTION_RETREAT
			switch mode {
			case "passive-rounded-ratio":
				update.Aggression = 0
			case "melee", "object-cast", "location-cast", "duration-cast":
				action := map[string]ai.ActionType{"melee": ai.ACTION_MELEE_ATTACK, "object-cast": ai.ACTION_CAST_SPELL_ON_OBJECT,
					"location-cast": ai.ACTION_CAST_SPELL_ON_LOCATION, "duration-cast": ai.ACTION_CAST_DURATION_SPELL}[mode]
				update.AIStackInd = 1
				update.AIStack[1].Action = uint32(action)
			case "caster-antimagic", "healthy-caster-antimagic":
				update.StatusFlags = object.MonStatusCanCastSpells
				unit.Buffs = 1 << ENCHANT_ANTI_MAGIC
				if mode == "healthy-caster-antimagic" {
					unit.HealthData.Cur = unit.HealthData.Max
				}
			case "noncaster-antimagic":
				unit.Buffs = 1 << ENCHANT_ANTI_MAGIC
			case "confused":
				unit.Buffs = 1 << ENCHANT_CONFUSED
			case "freeze-and-stun":
				unit.Buffs = 1<<ENCHANT_FREEZE | 1<<ENCHANT_HELD
			case "disabled":
				unit.ObjFlags &^= object.FlagEnabled
			case "destroyed":
				unit.ObjFlags |= object.FlagDestroyed
			case "NaN-aggression":
				update.Aggression = float32(math.NaN())
			case "NaN-threshold":
				update.RetreatLevel = float32(math.NaN())
			case "bot":
				update.StatusFlags = object.MonStatusBot
			case "summoned-coop", "summoned-noncoop":
				update.StatusFlags = object.MonStatusSummoned
			case "monitor-coop", "monitor-noncoop":
				unit.ObjSubClass = object.SubClass(object.MonsterMonitor)
			}
			if mode == "summoned-coop" || mode == "monitor-coop" {
				noxflags.SetGame(noxflags.GameModeCoop)
				want = ai.ACTION_RETREAT_TO_MASTER
			}
			sounds := [14]uint32{13: 0x12345678}
			update.SoundSet122 = unsafe.Pointer(&sounds[0])
			order := []string{}
			handled := s.MonsterMainNativeRuntime547210(unit, MonsterMainRuntime547210{
				AudioEvent: func(id uint32, got *Object) {
					if id != sounds[13] || got != unit || update.AIStackHead().Type() != want {
						t.Fatal("retreat sound did not follow native scheduling")
					}
					order = append(order, "audio")
				},
				ScriptCallback: func(block *ScriptCallback, caller, trigger *Object, event ScriptEventType) {
					if block != &update.ScriptRetreat || caller != nil || trigger != unit || event != NoxEventMonsterMoveXXX {
						t.Fatal("retreat callback lost its cached block or native trigger")
					}
					order = append(order, "script")
				},
				RandomInt:    func(int, int) int { t.Fatal("retreat consumed RNG"); return 0 },
				SearchEdible: func(*Object, float32) *Object { t.Fatal("retreat reached food search"); return nil },
			})
			if !handled || update.AIStackHead().Type() != want || len(order) != 2 || order[0] != "audio" || order[1] != "script" {
				t.Fatalf("retreat = handled %t, head %v, callbacks %v", handled, update.AIStackHead().Type(), order)
			}
			index := int(update.AIStackInd)
			if index < 1 || update.AIStack[index-1].Type() != ai.DEPENDENCY_NOT_CORNERED ||
				update.AIStack[index].Args != [4]uintptr{} || update.AIStack[0].Type() != ai.ACTION_FIGHT {
				t.Fatalf("retreat native stack = %+v", update.GetAIStack())
			}
		})
	}
}

func TestMonsterMainHealthRetreat547210CachedRecordLiveGatesAndSound(t *testing.T) {
	s, unit, cached := monsterMainHealthRetreatFixture547210(t)
	noxflags.SetGame(noxflags.GameModeCoop)
	unit.Field5 = 0x10
	live := &MonsterUpdateData{AIStackInd: 0, Aggression: 0}
	live.AIStack[0].Action = uint32(ai.ACTION_FIGHT)
	oldSound := [14]uint32{13: 101}
	newSound := [14]uint32{13: 202}
	cached.SoundSet122 = unsafe.Pointer(&oldSound[0])
	calls := 0
	if !s.MonsterMainNativeRuntime547210(unit, MonsterMainRuntime547210{
		GUICursorActive: func() bool {
			unit.UpdateData = unsafe.Pointer(live)
			cached.SoundSet122 = unsafe.Pointer(&newSound[0])
			oldSound[13] = 303
			return true
		},
		AudioEvent: func(id uint32, got *Object) {
			calls++
			if id != 303 || got != unit || live.AIStackHead().Type() != ai.ACTION_RETREAT {
				t.Fatal("retreat did not retain entry sound pointer and live action stack")
			}
		},
		ScriptCallback: func(block *ScriptCallback, _, trigger *Object, _ ScriptEventType) {
			calls++
			if block != &cached.ScriptRetreat || trigger != unit {
				t.Fatal("retreat did not retain entry update record")
			}
		},
	}) || calls != 2 || cached.AIStackInd != 0 || live.AIStackInd != 2 {
		t.Fatalf("cached/live retreat calls %d, cached index %d, live index %d", calls, cached.AIStackInd, live.AIStackInd)
	}
}

func TestMonsterMainHealthRetreat547210OriginalRejections(t *testing.T) {
	for _, mode := range []string{"no-max", "stationary", "NaN-speed", "recent-move", "wrapped-recent-move",
		"flee-below-head", "retreat-below-head", "master-below-head", "above-threshold", "healthy-noncaster-antimagic"} {
		t.Run(mode, func(t *testing.T) {
			s, unit, update := monsterMainHealthRetreatFixture547210(t)
			switch mode {
			case "no-max":
				unit.HealthData.Max = 0
			case "stationary":
				unit.SpeedBase = math.Nextafter32(float32(0.0099999998), 0)
			case "NaN-speed":
				unit.SpeedBase = float32(math.NaN())
			case "recent-move":
				update.Field127 = s.Frame() - 89
			case "wrapped-recent-move":
				s.SetFrame(2)
				update.Field127 = ^uint32(0) - 2
			case "flee-below-head", "retreat-below-head", "master-below-head":
				update.AIStack[0].Action = uint32(map[string]ai.ActionType{
					"flee-below-head": ai.ACTION_FLEE, "retreat-below-head": ai.ACTION_RETREAT, "master-below-head": ai.ACTION_RETREAT_TO_MASTER}[mode])
				update.AIStackInd = 1
				update.AIStack[1].Action = uint32(ai.ACTION_WAIT)
			case "above-threshold":
				update.RetreatLevel = math.Nextafter32(0.04, 0)
			case "healthy-noncaster-antimagic":
				unit.HealthData.Cur = unit.HealthData.Max
				unit.Buffs = 1 << ENCHANT_ANTI_MAGIC
			}
			before := *update
			if s.monsterMainHealthRetreat547210(unit, update, nil, MonsterMainRuntime547210{
				AudioEvent:     func(uint32, *Object) { t.Fatal("rejected retreat played sound") },
				ScriptCallback: func(*ScriptCallback, *Object, *Object, ScriptEventType) { t.Fatal("rejected retreat called script") },
			}) || update.AIStack != before.AIStack || update.AIStackInd != before.AIStackInd || s.AI.StackChanged {
				t.Fatal("rejected retreat changed stack or executed callbacks")
			}
		})
	}
}

func TestMonsterMainHealthRetreat547210PostPushPredicateAndRejectedPush(t *testing.T) {
	for _, mode := range []string{"summoned-after-cancel", "monitor-after-cancel", "full-stack", "dead-head"} {
		t.Run(mode, func(t *testing.T) {
			s, unit, update := monsterMainHealthRetreatFixture547210(t)
			noxflags.SetGame(noxflags.GameModeCoop)
			update.AIStack[0].Action = uint32(ai.ACTION_WAIT)
			update.AIStack[0].Field5 = 1
			old := aiActions[ai.ACTION_WAIT]
			aiActions[ai.ACTION_WAIT] = monsterMainFearCancel547210{cancel: func(*Object) {
				if mode == "summoned-after-cancel" {
					update.StatusFlags |= object.MonStatusSummoned
				} else {
					unit.ObjSubClass = object.SubClass(object.MonsterMonitor)
				}
				s.SetFrame(201)
			}}
			t.Cleanup(func() {
				if old == nil {
					delete(aiActions, ai.ACTION_WAIT)
				} else {
					aiActions[ai.ACTION_WAIT] = old
				}
			})
			if mode == "full-stack" {
				update.AIStackInd = int8(len(update.AIStack) - 1)
				update.AIStack[update.AIStackInd].Action = uint32(ai.ACTION_WAIT)
			} else if mode == "dead-head" {
				update.AIStack[0].Action = uint32(ai.ACTION_DEAD)
			}
			calls := 0
			if !s.monsterMainHealthRetreat547210(unit, update, nil, MonsterMainRuntime547210{
				ScriptCallback: func(*ScriptCallback, *Object, *Object, ScriptEventType) { calls++ },
			}) || calls != 1 {
				t.Fatal("eligible retreat must still return/call script when normal pushes reject")
			}
			if mode == "summoned-after-cancel" || mode == "monitor-after-cancel" {
				if update.AIStackHead().Type() != ai.ACTION_RETREAT_TO_MASTER || s.Frame() != 201 {
					t.Fatal("retreat read master predicate before the ordinary push callback")
				}
			}
		})
	}
}
