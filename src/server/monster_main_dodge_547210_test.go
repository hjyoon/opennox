package server

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/unit/ai"
)

func monsterMainDodgeFixture547210(t *testing.T) (*Server, *Object, *MonsterUpdateData, MonsterMainRuntime547210) {
	t.Helper()
	s, unit, update, runtime := monsterMainDodgeServiceFixture547C50(t)
	s.Rand.init(nil) // this unit fixture starts with a zero-value server
	noxflags.SetGame(noxflags.GameModeCoop)
	update.Aggression = 0.5
	update.MonsterDef.StatusFlags92 = object.MonStatusCanDodge
	runtime.TestShield = func(*Object) int { return 1 }
	return s, unit, update, runtime
}

func TestMonsterMainDodge547210EntryOriginalAdmissions(t *testing.T) {
	for _, mode := range []string{
		"ordinary", "NPC", "passive-boundary", "active-boundary", "attack-at-will-boundary",
		"melee", "missile", "object-cast", "location-cast", "duration-cast", "face-object", "waiting",
		"freeze-and-stun", "antimagic", "disabled", "destroyed", "bot", "recent-move",
		"zero-base-speed", "zero-current-speed", "full-stack", "dead-head",
	} {
		t.Run(mode, func(t *testing.T) {
			s, unit, update, runtime := monsterMainDodgeFixture547210(t)
			destination := types.Ptf(100, 215)
			switch mode {
			case "NPC":
				unit.ObjSubClass = object.SubClass(object.MonsterNPC)
			case "passive-boundary":
				update.Aggression = monsterMainPassiveAggressionLimit547210
			case "active-boundary":
				update.Aggression = monsterMainActiveAggressionLimit547210
			case "attack-at-will-boundary":
				update.Aggression = monsterMainAttackAtWillAggression547210
			case "melee", "missile", "object-cast", "location-cast", "duration-cast", "face-object", "waiting":
				update.AIStackInd = 1
				update.AIStack[1].Action = uint32(map[string]ai.ActionType{
					"melee": ai.ACTION_MELEE_ATTACK, "missile": ai.ACTION_MISSILE_ATTACK,
					"object-cast": ai.ACTION_CAST_SPELL_ON_OBJECT, "location-cast": ai.ACTION_CAST_SPELL_ON_LOCATION,
					"duration-cast": ai.ACTION_CAST_DURATION_SPELL, "face-object": ai.ACTION_FACE_OBJECT,
					"waiting": ai.ACTION_WAIT,
				}[mode])
			case "freeze-and-stun":
				unit.Buffs = 1<<ENCHANT_FREEZE | 1<<ENCHANT_HELD
			case "antimagic":
				unit.Buffs = 1 << ENCHANT_ANTI_MAGIC
			case "disabled":
				unit.ObjFlags &^= object.FlagEnabled
			case "destroyed":
				unit.ObjFlags |= object.FlagDestroyed
			case "bot":
				update.StatusFlags = object.MonStatusBot
			case "recent-move":
				update.Field127 = s.Frame()
			case "zero-base-speed":
				unit.SpeedBase = 0
			case "zero-current-speed":
				unit.SpeedCur = 0
				destination = unit.PosVec
			case "full-stack":
				update.AIStackInd = int8(len(update.AIStack) - 1)
				update.AIStackHead().Action = uint32(ai.ACTION_WAIT)
			case "dead-head":
				update.AIStackHead().Action = uint32(ai.ACTION_DEAD)
			}
			before, queries, tiles := *update, 0, 0
			runtime.TestShield = func(got *Object) int {
				queries++
				if got != unit {
					t.Fatal("missile query lost native unit")
				}
				return -1 // original tests nonzero, not positive
			}
			runtime.TileAt = func(types.Pointf) int { tiles++; return 0 }
			if !s.MonsterMainNativeRuntime547210(unit, runtime) || queries != 1 || tiles != 1 {
				t.Fatalf("original dodge admission missing: queries=%d tiles=%d stack=%+v", queries, tiles, update.GetAIStack())
			}
			if mode == "full-stack" || mode == "dead-head" {
				if *update != before {
					t.Fatal("rejected pushes changed stack, but branch must still return success")
				}
				return
			}
			index := int(update.AIStackInd)
			if index < 2 || update.AIStack[0].Type() != ai.ACTION_FIGHT ||
				update.AIStack[index-1].Type() != ai.DEPENDENCY_TIME || update.AIStack[index-1].ArgU32(0) != s.Frame()+s.TickRate() ||
				update.AIStackHead().Type() != ai.ACTION_DODGE || update.AIStackHead().ArgPos(0) != destination ||
				update.AIStackHead().ArgU32(2) != 0 {
				t.Fatalf("dodge native stack=%+v, want destination %v", update.GetAIStack(), destination)
			}
		})
	}
}

func TestMonsterMainDodge547210EntryCachedDefinitionAndLiveGates(t *testing.T) {
	for _, mode := range []string{"cached-definition", "live-definition-only", "live-passive", "live-unordered", "live-dodge"} {
		t.Run(mode, func(t *testing.T) {
			s, unit, cached, runtime := monsterMainDodgeFixture547210(t)
			live := &MonsterUpdateData{Aggression: 0.5, AIStackInd: 0, MonsterDef: &MonsterDef{}}
			live.AIStack[0].Action = uint32(ai.ACTION_FIGHT)
			want := mode == "cached-definition"
			switch mode {
			case "live-definition-only":
				cached.MonsterDef.StatusFlags92 = 0
				live.MonsterDef.StatusFlags92 = object.MonStatusCanDodge
			case "live-passive":
				live.Aggression = math.Nextafter32(monsterMainPassiveAggressionLimit547210, 0)
			case "live-unordered":
				live.Aggression = float32(math.NaN())
			case "live-dodge":
				live.AIStack[0].Action = uint32(ai.ACTION_DODGE)
				live.AIStackInd = 1
				live.AIStack[1].Action = uint32(ai.ACTION_WAIT)
			}
			beforeLive, beforeCached, queries := *live, *cached, 0
			runtime.GUICursorActive = func() bool { unit.UpdateData = unsafe.Pointer(live); return true }
			runtime.TestShield = func(*Object) int { queries++; return 1 }
			s.MonsterMainNativeRuntime547210(unit, runtime)
			if want {
				if queries != 1 || live.AIStackInd != 2 || live.AIStackHead().Type() != ai.ACTION_DODGE || *cached != beforeCached {
					t.Fatalf("cached definition did not govern live stack: queries=%d live=%+v", queries, live.GetAIStack())
				}
			} else {
				// NaN is unequal to itself; check its original bits separately
				// before comparing the other fields of the two snapshots.
				after := *live
				aggressionUnchanged := math.Float32bits(after.Aggression) == math.Float32bits(beforeLive.Aggression)
				after.Aggression, beforeLive.Aggression = 0, 0
				if queries != 0 || !aggressionUnchanged || after != beforeLive || *cached != beforeCached {
					t.Fatalf("definition/status/whole-stack sources mixed: queries=%d live=%+v", queries, live.GetAIStack())
				}
			}
		})
	}
}

func TestMonsterMainDodge547210OriginalRejections(t *testing.T) {
	for _, mode := range []string{
		"noncoop", "below-passive", "unordered", "negative-infinity", "confused", "nil-definition",
		"other-definition-flag", "update-status-only", "dodge-head", "dodge-below-wait", "nil-probe", "probe-miss",
	} {
		t.Run(mode, func(t *testing.T) {
			s, unit, update, runtime := monsterMainDodgeFixture547210(t)
			queries, wantQueries := 0, 0
			runtime.TestShield = func(*Object) int { queries++; return 1 }
			runtime.RandomFloat = func(float32, float32) float64 { t.Fatal("rejected gate consumed RNG"); return 2 }
			switch mode {
			case "noncoop":
				noxflags.UnsetGame(noxflags.GameModeCoop)
			case "below-passive":
				update.Aggression = math.Nextafter32(monsterMainPassiveAggressionLimit547210, 0)
			case "unordered":
				update.Aggression = float32(math.NaN())
			case "negative-infinity":
				update.Aggression = float32(math.Inf(-1))
			case "confused":
				unit.Buffs = 1 << ENCHANT_CONFUSED
			case "nil-definition":
				update.MonsterDef = nil
			case "other-definition-flag":
				update.MonsterDef.StatusFlags92 = object.MonStatusCanBlock
			case "update-status-only":
				update.MonsterDef.StatusFlags92 = 0
				update.StatusFlags = object.MonStatusCanDodge
			case "dodge-head", "dodge-below-wait":
				update.AIStack[0].Action = uint32(ai.ACTION_DODGE)
				if mode == "dodge-below-wait" {
					update.AIStackInd = 1
					update.AIStack[1].Action = uint32(ai.ACTION_WAIT)
				}
			case "nil-probe":
				runtime.TestShield = nil
			case "probe-miss":
				wantQueries = 1
				runtime.TestShield = func(*Object) int { queries++; return 0 }
			}
			before := update.AIStack
			if s.monsterMainDodge547210(unit, update, runtime) || queries != wantQueries || update.AIStack != before {
				t.Fatalf("rejected %s gate admitted dodge: queries=%d stack=%+v", mode, queries, update.GetAIStack())
			}
		})
	}
}

func TestMonsterMainDodge547210RuntimeContainmentAndDefaultRNG(t *testing.T) {
	for _, mode := range []string{"missing-tile", "missing-float-and-logic", "missing-integer-and-logic", "default-float", "default-integer", "default-both"} {
		t.Run(mode, func(t *testing.T) {
			s, unit, update, runtime := monsterMainDodgeFixture547210(t)
			want, queries := true, 0
			runtime.TestShield = func(*Object) int { queries++; return 1 }
			switch mode {
			case "missing-tile":
				runtime.TileAt = nil
				want = false
			case "missing-float-and-logic":
				runtime.RandomFloat, s.Rand.Logic = nil, nil
				want = false
			case "missing-integer-and-logic":
				runtime.RandomInt, s.Rand.Logic = nil, nil
				want = false
			case "default-float":
				runtime.RandomFloat = nil
			case "default-integer":
				runtime.RandomInt = nil
			case "default-both":
				runtime.RandomInt, runtime.RandomFloat = nil, nil
			}
			before := *update
			if got := s.monsterMainDodge547210(unit, update, runtime); got != want || queries != 1 {
				t.Fatalf("runtime %s = %t queries=%d, want %t", mode, got, queries, want)
			}
			if want && update.AIStackHead().Type() != ai.ACTION_DODGE || !want && *update != before {
				t.Fatal("runtime containment or actual server RNG changed scheduling")
			}
		})
	}
}

func TestMonsterMainDodge547210NoClearDestinationFallsThrough(t *testing.T) {
	for _, mode := range []string{"ray", "obstacles", "lava"} {
		t.Run(mode, func(t *testing.T) {
			s, unit, update, runtime := monsterMainDodgeFixture547210(t)
			queries, attempts := 0, 0
			runtime.TestShield = func(*Object) int { queries++; return 1 }
			runtime.RandomFloat = func(float32, float32) float64 { attempts++; return 2 }
			switch mode {
			case "ray":
				runtime.TraceRay = func(types.Pointf, types.Pointf, MapTraceFlags) bool { return false }
			case "obstacles":
				runtime.TraceObstacles = func(*Object, types.Pointf, types.Pointf) bool { return false }
			case "lava":
				runtime.TileAt = func(types.Pointf) int { return 6 }
			}
			before := *update
			if s.monsterMainDodge547210(unit, update, runtime) || queries != 1 || attempts != 5 || *update != before {
				t.Fatalf("five unsuccessful attempts did not fall through: queries=%d attempts=%d stack=%+v", queries, attempts, update.GetAIStack())
			}
		})
	}
}

func TestMonsterMainDodge547210ProbeMutationDoesNotRecheckAdmission(t *testing.T) {
	for _, mode := range []string{"coop", "aggression", "confused", "definition", "existing-dodge", "replacement-update"} {
		t.Run(mode, func(t *testing.T) {
			s, unit, cached, runtime := monsterMainDodgeFixture547210(t)
			live, queries := cached, 0
			runtime.TestShield = func(*Object) int {
				queries++
				switch mode {
				case "coop":
					noxflags.UnsetGame(noxflags.GameModeCoop)
				case "aggression":
					cached.Aggression = 0
				case "confused":
					unit.Buffs = 1 << ENCHANT_CONFUSED
				case "definition":
					cached.MonsterDef.StatusFlags92 = 0
				case "existing-dodge":
					cached.AIStack[0].Action = uint32(ai.ACTION_DODGE)
				case "replacement-update":
					live = &MonsterUpdateData{AIStackInd: 0}
					live.AIStack[0].Action = uint32(ai.ACTION_WAIT)
					unit.UpdateData = unsafe.Pointer(live)
				}
				return 1
			}
			if !s.MonsterMainNativeRuntime547210(unit, runtime) || queries != 1 ||
				live.AIStackInd != 2 || live.AIStack[1].Type() != ai.DEPENDENCY_TIME || live.AIStackHead().Type() != ai.ACTION_DODGE {
				t.Fatalf("post-query gate was rechecked: queries=%d stack=%+v", queries, live.GetAIStack())
			}
		})
	}
}

func TestMonsterMainDodge547210EntryEarlierBranchesPreempt(t *testing.T) {
	for _, mode := range []string{"dead", "uninterruptible", "throttle", "fear", "health-retreat", "weapon-block", "shield-block"} {
		t.Run(mode, func(t *testing.T) {
			s, unit, update, runtime := monsterMainDodgeFixture547210(t)
			queries, wantQueries := 0, 0
			switch mode {
			case "dead":
				unit.ObjFlags |= object.FlagDead
			case "uninterruptible":
				update.AIStack[0].Action = uint32(ai.DEPENDENCY_UNINTERRUPTABLE)
				update.AIStackInd = 1
				update.AIStack[1].Action = uint32(ai.ACTION_WAIT)
			case "throttle":
				update.AIStack[0].Action = uint32(ai.ACTION_IDLE)
				update.Field137 = s.Frame() - 1 + uint32(byte(unit.NetCode))
			case "fear":
				unit.Buffs = 1 << ENCHANT_AFRAID
			case "health-retreat":
				unit.HealthData.Cur = 0
			case "weapon-block":
				unit.ObjSubClass = object.SubClass(object.MonsterNPC)
				update.WeaponEquipFlags = 0x400
				wantQueries = 1
			case "shield-block":
				update.StatusFlags = object.MonStatusCanBlock
				wantQueries = 1
			}
			runtime.TestShield = func(*Object) int { queries++; return 1 }
			runtime.RandomFloat = func(float32, float32) float64 { t.Fatal("dodge ran before earlier branch"); return 2 }
			if !s.MonsterMainNativeRuntime547210(unit, runtime) || queries != wantQueries || update.HasAction(ai.ACTION_DODGE) {
				t.Fatalf("dodge preempted %s: queries=%d stack=%+v", mode, queries, update.GetAIStack())
			}
		})
	}
}
