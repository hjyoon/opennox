package server

import (
	"github.com/opennox/libs/object"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/unit/ai"
)

// monsterMainDodge547210 restores GAME.EXE 00547994..005479F9 after health
// retreat and weapon/shield defense. Capability belongs to the entry-cached
// MonsterDef, while passive/confusion and whole-stack DODGE checks read the
// live unit. The original does not recheck admission after the missile query,
// and imposes no movement-speed, recent-move, or cast-head restriction here.
func (s *Server) monsterMainDodge547210(unit *Object, update *MonsterUpdateData, runtime MonsterMainRuntime547210) bool {
	if unit == nil || unit.UpdateData == nil || update == nil ||
		!noxflags.HasGame(noxflags.GameModeCoop) ||
		!(unit.UpdateDataMonster().Aggression >= monsterMainPassiveAggressionLimit547210) ||
		unit.HasEnchant(ENCHANT_CONFUSED) || update.MonsterDef == nil ||
		!update.MonsterDef.StatusFlags92.Has(object.MonStatusCanDodge) ||
		unit.UpdateDataMonster().HasAction(ai.ACTION_DODGE) || runtime.TestShield == nil ||
		runtime.TestShield(unit) == 0 {
		return false
	}
	if runtime.TileAt == nil ||
		(runtime.RandomInt == nil || runtime.RandomFloat == nil) && s.Rand.Logic == nil {
		return false // missing runtime services are not evidence of a clear dodge
	}
	return s.monsterMainCheckDodgeables547C50(unit, runtime)
}
