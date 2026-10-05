package server

import (
	"fmt"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/unit/ai"
)

func monsterMainBlockFixture547210(t *testing.T, weapon bool) (*Server, *Object, *MonsterUpdateData) {
	t.Helper()
	s, unit, update := monsterMainHealthRetreatFixture547210(t)
	unit.HealthData.Cur = unit.HealthData.Max
	if weapon {
		unit.ObjSubClass = object.SubClass(object.MonsterNPC)
		update.WeaponEquipFlags = 0x400
	} else {
		update.StatusFlags = object.MonStatusCanBlock
	}
	return s, unit, update
}

func TestMonsterMainBlock547210EntryTransitions(t *testing.T) {
	for _, weapon := range []bool{false, true} {
		for _, mode := range []string{"fight", "face-object", "cast-location", "existing-block", "waiting", "weapon-block",
			"confused-melee-head", "confused-wait-head", "disabled", "destroyed", "antimagic", "full-stack", "dead-head"} {
			t.Run(fmt.Sprintf("weapon=%t/%s", weapon, mode), func(t *testing.T) {
				s, unit, update := monsterMainBlockFixture547210(t, weapon)
				s.SetTickRate(31)
				want := ai.ACTION_BLOCK_ATTACK
				deadline := uint32(115)
				if weapon {
					want, deadline = ai.ACTION_WAIT, 131
				}
				switch mode {
				case "face-object", "cast-location":
					update.AIStackInd = 1
					update.AIStack[1].Action = uint32(map[string]ai.ActionType{"face-object": ai.ACTION_FACE_OBJECT, "cast-location": ai.ACTION_CAST_SPELL_ON_LOCATION}[mode])
				case "existing-block":
					update.AIStackInd = 2
					update.AIStack[1] = AIStackItem{Action: uint32(ai.ACTION_BLOCK_ATTACK), Args: [4]uintptr{900}}
					update.AIStack[2].Action = uint32(ai.ACTION_FACE_OBJECT)
					if !weapon {
						want, deadline = ai.ACTION_FACE_OBJECT, 0
					}
				case "waiting":
					update.AIStack[0] = AIStackItem{Action: uint32(ai.ACTION_WAIT), Args: [4]uintptr{900}}
					if weapon {
						deadline = 900
					}
				case "weapon-block":
					update.AIStack[0] = AIStackItem{Action: uint32(ai.ACTION_WEAPON_BLOCK), Args: [4]uintptr{900}}
					if weapon {
						want, deadline = ai.ACTION_WEAPON_BLOCK, 900
					}
				case "confused-melee-head", "confused-wait-head":
					unit.Buffs = 1 << ENCHANT_CONFUSED
					update.AIStack[0].Action = uint32(map[string]ai.ActionType{"confused-melee-head": ai.ACTION_MELEE_ATTACK, "confused-wait-head": ai.ACTION_WAIT}[mode])
				case "disabled":
					unit.ObjFlags &^= object.FlagEnabled
				case "destroyed":
					unit.ObjFlags |= object.FlagDestroyed
				case "antimagic":
					unit.Buffs = 1 << ENCHANT_ANTI_MAGIC
				case "full-stack":
					update.AIStackInd = int8(len(update.AIStack) - 1)
					update.AIStack[update.AIStackInd].Action = uint32(ai.ACTION_FIGHT)
					want, deadline = ai.ACTION_FIGHT, 0
				case "dead-head":
					update.AIStack[0].Action = uint32(ai.ACTION_DEAD)
					want, deadline = ai.ACTION_DEAD, 0
				}
				calls := 0
				if !s.MonsterMainNativeRuntime547210(unit, MonsterMainRuntime547210{TestShield: func(got *Object) int {
					calls++
					if got != unit {
						t.Fatal("shield probe lost native unit")
					}
					return -1
				}}) || calls != 1 || update.AIStackHead().Type() != want || update.AIStackHead().ArgU32(0) != deadline {
					t.Fatalf("block = calls %d, head %+v, want %v/%d", calls, update.AIStackHead(), want, deadline)
				}
			})
		}
	}
}

func TestMonsterMainBlock547210RetreatPreemptsProbe(t *testing.T) {
	s, unit, update := monsterMainBlockFixture547210(t, false)
	unit.HealthData.Cur = 1
	if !s.MonsterMainNativeRuntime547210(unit, MonsterMainRuntime547210{
		TestShield: func(*Object) int { t.Fatal("block probe ran before eligible retreat"); return 1 },
	}) || update.AIStackHead().Type() != ai.ACTION_RETREAT {
		t.Fatal("original health retreat did not preempt missile blocking")
	}
}

func TestMonsterMainHasShield5342C0LiveCapabilities(t *testing.T) {
	for _, npc := range []bool{false, true} {
		for _, armor := range []uint32{0, 0x1000000, 0x2000000, 0x3000000, 0x80000000} {
			for _, block := range []bool{false, true} {
				t.Run(fmt.Sprintf("NPC=%t/armor=%x/status=%t", npc, armor, block), func(t *testing.T) {
					_, unit, update := monsterMainBlockFixture547210(t, false)
					update.StatusFlags = 0
					if npc {
						unit.ObjSubClass = object.SubClass(object.MonsterNPC)
					}
					if block {
						update.StatusFlags = object.MonStatusCanBlock
					}
					update.ArmorEquipFlags = armor
					want := block
					if npc {
						want = armor&0x3000000 != 0
					}
					if monsterMainHasShield5342C0(unit) != want {
						t.Fatal("NPC equipment and ordinary status gates were mixed")
					}
				})
			}
		}
	}
}

func TestMonsterMainBlock547210ProbeMayReplaceLiveStackButNotCachedHead(t *testing.T) {
	for _, weapon := range []bool{false, true} {
		t.Run(fmt.Sprint(weapon), func(t *testing.T) {
			s, unit, update := monsterMainBlockFixture547210(t, weapon)
			live := &MonsterUpdateData{AIStackInd: 1, StatusFlags: object.MonStatusCanBlock}
			live.AIStack[0].Action = uint32(ai.ACTION_BLOCK_ATTACK)
			live.AIStack[1].Action = uint32(ai.ACTION_WAIT)
			calls := 0
			if !s.monsterMainBlock547210(unit, update, &update.AIStack[0], MonsterMainRuntime547210{TestShield: func(*Object) int {
				calls++
				unit.UpdateData = unsafe.Pointer(live)
				return 1
			}}) || calls != 1 {
				t.Fatal("eligible query did not complete native block branch")
			}
			if weapon {
				if live.AIStackInd != 2 || live.AIStackHead().Type() != ai.ACTION_WAIT || live.AIStackHead().ArgU32(0) != 130 {
					t.Fatal("weapon block reread live WAIT instead of its cached FIGHT head")
				}
			} else if live.AIStackInd != 1 {
				t.Fatal("shield block ignored existing BLOCK in replacement live stack")
			}
		})
	}
}

func TestMonsterMainBlock547210PostPushClockAndUnsignedWrap(t *testing.T) {
	for _, weapon := range []bool{false, true} {
		t.Run(fmt.Sprint(weapon), func(t *testing.T) {
			s, unit, update := monsterMainBlockFixture547210(t, weapon)
			update.AIStack[0].Field5 = 1
			// FIGHT's normal push cancellation also uses this test action type;
			// only the registration key controls which stack entry it handles.
			oldFight := aiActions[ai.ACTION_FIGHT]
			aiActions[ai.ACTION_FIGHT] = monsterMainFearCancel547210{cancel: func(*Object) { s.SetFrame(^uint32(0) - 2); s.SetTickRate(9) }}
			t.Cleanup(func() {
				if oldFight == nil {
					delete(aiActions, ai.ACTION_FIGHT)
				} else {
					aiActions[ai.ACTION_FIGHT] = oldFight
				}
			})
			if !s.monsterMainBlock547210(unit, update, &update.AIStack[0], MonsterMainRuntime547210{TestShield: func(*Object) int { return 1 }}) {
				t.Fatal("native block was not scheduled")
			}
			want := uint32(1)
			if weapon {
				want = 6
			}
			if update.AIStackHead().ArgU32(0) != want {
				t.Fatal("block deadline was computed before push or without uint32 wrap")
			}
		})
	}
}

func TestMonsterMainBlock547210NoExtraCoopGate(t *testing.T) {
	s, unit, update := monsterMainBlockFixture547210(t, false)
	noxflags.SetGame(noxflags.GameModeCoop)
	if !s.monsterMainBlock547210(unit, update, &update.AIStack[0], MonsterMainRuntime547210{TestShield: func(*Object) int { return 1 }}) {
		t.Fatal("co-op wrongly suppressed native shielding")
	}
}

func TestMonsterMainBlock547210ExcludedHeadsAndMissedThreat(t *testing.T) {
	for _, weapon := range []bool{false, true} {
		for _, head := range []ai.ActionType{ai.ACTION_MELEE_ATTACK, ai.ACTION_MISSILE_ATTACK, ai.ACTION_FIGHT} {
			t.Run(fmt.Sprintf("weapon=%t/head=%v", weapon, head), func(t *testing.T) {
				s, unit, update := monsterMainBlockFixture547210(t, weapon)
				update.AIStack[0].Action = uint32(head)
				before := *update
				calls := 0
				if s.monsterMainBlock547210(unit, update, &update.AIStack[0], MonsterMainRuntime547210{TestShield: func(*Object) int { calls++; return 0 }}) {
					t.Fatal("excluded or missed missile probe completed block branch")
				}
				wantCalls := 0
				if head == ai.ACTION_FIGHT {
					wantCalls = 1
				}
				if calls != wantCalls || update.AIStack != before.AIStack || update.AIStackInd != before.AIStackInd || s.AI.StackChanged {
					t.Fatal("excluded or missed missile query changed stack/probe count")
				}
			})
		}
	}
}

func TestMonsterMainBlock547210WeaponMissRechecksLiveShieldAndHead(t *testing.T) {
	for _, mode := range []string{"shield-after-weapon-miss", "cached-head-becomes-wait", "cached-head-becomes-weapon-block", "weapon-before-shield"} {
		t.Run(mode, func(t *testing.T) {
			s, unit, update := monsterMainBlockFixture547210(t, true)
			update.ArmorEquipFlags = 0
			calls := 0
			if !s.monsterMainBlock547210(unit, update, &update.AIStack[0], MonsterMainRuntime547210{TestShield: func(*Object) int {
				calls++
				if mode == "shield-after-weapon-miss" && calls == 1 {
					update.ArmorEquipFlags = 0x1000000
					return 0
				}
				if mode == "cached-head-becomes-wait" {
					update.AIStack[0].Action = uint32(ai.ACTION_WAIT)
				}
				if mode == "cached-head-becomes-weapon-block" {
					update.AIStack[0].Action = uint32(ai.ACTION_WEAPON_BLOCK)
				}
				return 1
			}}) {
				t.Fatal("eligible native block did not complete")
			}
			wantCalls := 1
			if mode == "shield-after-weapon-miss" {
				wantCalls = 2
				if update.AIStackHead().Type() != ai.ACTION_BLOCK_ATTACK {
					t.Fatal("weapon miss did not use live shield capability")
				}
			} else if mode == "weapon-before-shield" {
				if update.AIStackHead().Type() != ai.ACTION_WAIT {
					t.Fatal("shield incorrectly preempted weapon block")
				}
			} else if update.AIStackInd != 0 || s.AI.StackChanged {
				t.Fatal("post-query cached WAIT/WEAPON_BLOCK was ignored")
			}
			if calls != wantCalls {
				t.Fatal("MainAI repeated or skipped shield query")
			}
		})
	}
}

func TestMonsterMainBlock547210LiveSubclassAndCachedWeaponRecord(t *testing.T) {
	s, unit, cached := monsterMainBlockFixture547210(t, true)
	noxflags.SetGame(noxflags.GameModeCoop)
	unit.Field5 = 0x10
	live := &MonsterUpdateData{AIStackInd: 0}
	live.AIStack[0].Action = uint32(ai.ACTION_FIGHT)
	calls := 0
	if !s.MonsterMainNativeRuntime547210(unit, MonsterMainRuntime547210{
		GUICursorActive: func() bool { unit.UpdateData = unsafe.Pointer(live); return true },
		TestShield:      func(*Object) int { calls++; return 1 },
	}) || calls != 1 || live.AIStackHead().Type() != ai.ACTION_WAIT || live.WeaponEquipFlags != 0 || cached.AIStackInd != 0 {
		t.Fatal("weapon branch lost cached equipment/head after live record replacement")
	}
}
