package server

import (
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/common/unit/ai"
)

func TestMonsterMainBlink547210EntryCallsImmediateServiceBeforeExistingFlee(t *testing.T) {
	for _, existingFlee := range []bool{false, true} {
		t.Run(map[bool]string{false: "fight", true: "already-fleeing"}[existingFlee], func(t *testing.T) {
			s, unit, update, _ := newMonsterMainFleeTest547210(t)
			update.StatusFlags = object.MonStatusCanCastSpells
			update.Field376 = 0x80000000 // any nonzero spell slot qualifies
			update.Field371 = s.Frame()
			update.Field370_0, update.Field370_2 = 7, 7
			if existingFlee {
				update.AIStack[0].Action = uint32(ai.ACTION_FLEE)
			}
			beforeStack := update.AIStack
			calls, randomCalls := 0, 0
			handled := s.MonsterMainNativeRuntime547210(unit, MonsterMainRuntime547210{
				CastSpell: func(id int32, caster *Object, arg *SpellAcceptArg) {
					calls++
					if id != 4 || caster != unit || arg.Obj != unit || arg.Pos != unit.PosVec {
						t.Fatal("MainAI did not pass the live native unit and position to Blink")
					}
				},
				RandomInt: func(min, max int) int {
					randomCalls++
					if calls != 1 || min != 7 || max != 7 {
						t.Fatal("Blink cooldown randomization did not follow the immediate cast")
					}
					return 7
				},
			})
			if !handled || calls != 1 || randomCalls != 1 || update.Field371 != 107 ||
				update.AIStack != beforeStack || update.AIStackInd != 0 {
				t.Fatalf("MainAI Blink = handled %t, cast %d, RNG %d, deadline %d, stack index %d",
					handled, calls, randomCalls, update.Field371, update.AIStackInd)
			}
		})
	}
}
