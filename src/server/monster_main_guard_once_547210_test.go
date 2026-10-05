package server

import (
	"testing"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

func TestMonsterMainGuard547210AlreadyProbedPrefixDoesNotRepeat(t *testing.T) {
	s, unit, update := monsterScriptHitFixture515A30(t)
	s.SetFrame(208)
	update.Aggression = 1
	update.AIStackInd = 1
	update.AIStack[0].Action = uint32(ai.ACTION_GUARD)
	update.AIStack[1].Action = uint32(ai.ACTION_WAIT)
	s.monsterMainGuardEnemyStimulus547210(unit, update, MonsterMainRuntime547210{
		guardStimulusDone: true,
		EnemyAggro: func(*Object, float32) *Object {
			t.Fatal("stable classifier repeated the entry-prefix GUARD query")
			return nil
		},
	}, ai.ACTION_WAIT)
}
