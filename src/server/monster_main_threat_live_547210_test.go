package server

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/common/unit/ai"
)

func TestMonsterMainThreatFlee547210LivePredicatesAndCachedEnemySpellRecord(t *testing.T) {
	for _, gate := range []string{"live-allows", "live-passive", "live-cast-head", "live-recent-move"} {
		t.Run(gate, func(t *testing.T) {
			s, unit, cached, enemy := newMonsterMainFleeTest547210(t)
			cached.Aggression = 0 // not the record used by 00534440
			cached.StatusFlags, cached.Field376 = object.MonStatusCanCastSpells, 1
			live := new(MonsterUpdateData)
			live.Aggression = 1
			live.AIStack[0].Action = uint32(ai.ACTION_WAIT)
			switch gate {
			case "live-passive":
				live.Aggression = 0
			case "live-cast-head":
				live.AIStack[0].Action = uint32(ai.ACTION_CAST_DURATION_SPELL)
			case "live-recent-move":
				live.Field127 = s.Frame() - 1
			}
			unit.UpdateData = unsafe.Pointer(live)
			before := *live
			distances, casts := 0, 0
			handled := s.monsterMainThreatFlee547210(unit, cached, nil, MonsterMainRuntime547210{
				Distance: func(got, target *Object) float64 {
					distances++
					if got != unit || target != enemy {
						t.Fatal("distance did not retain MainAI's cached enemy")
					}
					return 20
				},
				CastSpell: func(int32, *Object, *SpellAcceptArg) { casts++ },
				RandomInt: func(int, int) int { return 3 },
			})
			want := gate == "live-allows"
			if handled != want || (casts == 1) != want || (distances == 1) != want || *live != before ||
				(want && cached.Field371 != 103) {
				t.Fatal("live predicates/cached spell record distinction was lost")
			}
		})
	}
}
