package opennox

import (
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/things"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/server"
)

// gameCaptureMagicAllowed4FDC10 preserves GAME.EXE 004FDC10 without treating
// native Object, PlayerUpdateData, Player, or owned-object links as PE32
// integers. CTF has priority over FlagBall, which has priority over KOTR when
// malformed input contains more than one game-mode bit.
func gameCaptureMagicAllowed4FDC10(
	spellCannotHoldCrown bool,
	unit *server.Object,
	gameFlags noxflags.GameFlag,
	gameBallID, crownID uint16,
) bool {
	if unit == nil {
		return false
	}
	if !spellCannotHoldCrown || !unit.Class().Has(object.ClassPlayer) {
		return true
	}

	switch {
	case gameFlags.Has(noxflags.GameModeCTF):
		update := unit.UpdateDataPlayer()
		return update.Player.WeaponEquip&1 == 0
	case gameFlags.Has(noxflags.GameModeFlagBall):
		for owned := unit.Field129; owned != nil; owned = owned.Field128 {
			if owned.TypeInd == gameBallID {
				return false
			}
		}
	case gameFlags.Has(noxflags.GameModeKOTR):
		for owned := unit.Field129; owned != nil; owned = owned.Field128 {
			if owned.TypeInd == crownID && unit.HasTeam() {
				return false
			}
		}
	}
	return true
}

func (s *Server) gameCaptureMagic4FDC10(spellID spell.ID, unit *server.Object) bool {
	// GAME.EXE initializes both object-type caches together before checking the
	// unit argument. The server caches retain the same lazy lookup behavior.
	gameBallID := uint16(s.Types.GameBallID())
	crownID := uint16(s.Types.CrownID())
	return gameCaptureMagicAllowed4FDC10(
		s.Spells.HasFlags(spellID, things.SpellCantHoldCrown),
		unit,
		noxflags.GetGame(),
		gameBallID,
		crownID,
	)
}
