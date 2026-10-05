package server

import (
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/prand"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

func TestMonsterFleeServer544760SchedulesRealSelfCast(t *testing.T) {
	for _, related := range []bool{false, true} {
		for _, duration := range []bool{false, true} {
			name := "self"
			if related {
				name = "related"
			}
			if duration {
				name += "/duration"
			} else {
				name += "/instant"
			}
			t.Run(name, func(t *testing.T) {
				s, unit, update := monsterScriptHitFixture515A30(t)
				s.frame = 100
				s.SetTickRate(30)
				s.Rand.Logic = prand.New(544760)
				flags := things.SpellMobsCanCast
				if duration {
					flags |= things.SpellDuration
				}
				s.Spells.byID = map[spell.ID]*SpellDef{
					spell.SPELL_HASTE: {ID: spell.SPELL_HASTE, Def: things.Spell{Flags: flags}},
				}
				unit.ObjFlags = object.FlagEnabled
				unit.SpeedBase = 2
				update.StatusFlags = object.MonStatusCanCastSpells
				update.Aggression = 0.8
				update.CurrentEnemy = &Object{PosVec: types.Ptf(20, 30)}
				update.Field70 = s.Frame()
				update.Field364_0, update.Field364_2 = 7, 7
				update.Field370_0, update.Field370_2 = 9, 9
				mask := monsterFightSelfSpellMask540B90
				if related {
					mask = monsterFleeRelatedSpellMask541050
					update.AIStack[0] = AIStackItem{Action: uint32(ai.DEPENDENCY_UNINTERRUPTABLE)}
					update.AIStackInd = 1
				}
				unsafe.Slice(&update.Field373, SpellsMax-1)[int(spell.SPELL_HASTE)-1] = mask
				flee := &update.AIStack[update.AIStackInd]
				*flee = AIStackItem{Action: uint32(ai.ACTION_FLEE)}
				flee.SetArgs(types.Ptf(-1, -2), uint32(0x12345678))
				pathCalls := 0
				s.MonsterActionFlee544760(unit, func([]types.Pointf, *Object, *types.Pointf) int {
					pathCalls++
					return 0
				})
				wantAction := ai.ACTION_CAST_SPELL_ON_OBJECT
				wantIndex := int8(2)
				if related {
					wantIndex++
				}
				if duration {
					wantAction = ai.ACTION_CAST_DURATION_SPELL
					wantIndex++
				}
				head := update.AIStackHead()
				if update.AIStackInd != wantIndex || head.Type() != wantAction ||
					head.ArgU32(0) != uint32(spell.SPELL_HASTE) || head.ArgU32(1) != 0 ||
					head.Args[2] != uintptr(unsafe.Pointer(unit)) || !s.AI.StackChanged {
					t.Fatalf("real cast stack = %#v index %d, changed %t", head, update.AIStackInd, s.AI.StackChanged)
				}
				if flee.ArgPos(0) != update.CurrentEnemy.PosVec || flee.ArgU32(2) != 0x12345678 || pathCalls != 0 {
					t.Fatalf("cached FLEE = %#v, path calls %d", flee, pathCalls)
				}
				if related {
					if update.Field371 != 109 || update.Field365 != 0 {
						t.Fatalf("related cooldown = %d, self cooldown %d", update.Field371, update.Field365)
					}
				} else if update.Field365 != 107 || update.Field371 != 0 {
					t.Fatalf("self cooldown = %d, related cooldown %d", update.Field365, update.Field371)
				}
				runtime.KeepAlive(unit)
			})
		}
	}
}
