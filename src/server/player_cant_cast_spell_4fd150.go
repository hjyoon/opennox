package server

const (
	playerCantCastModeKOTR4FD150     = uint32(0x10)
	playerCantCastModeCTF4FD150      = uint32(0x20)
	playerCantCastModeFlagBall4FD150 = uint32(0x40)

	playerCantCastSpellCantHoldCrown4FD150 = uint32(0x80000)
	playerCantCastInventoryFlag4FD150      = uint32(0x10000000)
	playerCantCastAntiMagic4FD150          = uint8(29)
)

type playerCantCastSummonType4FD150 uint8

const (
	playerCantCastPixie4FD150 playerCantCastSummonType4FD150 = iota
	playerCantCastMagicMissile4FD150
	playerCantCastSmallFist4FD150
	playerCantCastMediumFist4FD150
	playerCantCastLargeFist4FD150
	playerCantCastDeathBall4FD150
	playerCantCastMeteor4FD150
)

// playerCantCastSpellHooks4FD150 exposes every observable scalar load,
// pointer traversal, and helper call in GAME.EXE 004FD150. Object and list
// nodes are generic so the semantic core cannot accidentally assume PE32
// pointer width or a native Go structure layout.
type playerCantCastSpellHooks4FD150[O, I comparable] struct {
	findParent  func(O) O
	hasGameFlag func(uint32) int32

	loadCrownTypeCache     func() uint32
	storeCrownTypeCache    func(uint32)
	loadGameBallTypeCache  func() uint32
	storeGameBallTypeCache func(uint32)
	lookupObjectType       func(string) uint32
	spellHasFlags          func(int32, uint32) int32

	loadFirstOwned func(O) I
	loadOwnedType  func(I) uint16
	loadNextOwned  func(I) I
	hasTeam        func(O) int32

	loadFirstInventory func(O) I
	loadInventoryFlags func(I) uint32
	loadNextInventory  func(I) I

	hasEnchant     func(O, uint8) int32
	spellAllowed   func(int32) int32
	loadSummonType func(playerCantCastSummonType4FD150) uint32
	countOwnedType func(O, uint32) int32
	spellPower     func(int32, O) int32
	balanceFloat   func(string, int32) float64
}

// playerCantCastSpell4FD150 preserves GAME.EXE 004FD150, including game-mode
// precedence (KOTR, FlagBall, then CTF), lazy cache timing, low-16-bit object
// type comparisons, and the original helper-call order for summon limits.
// The initial owner-chain lookup is intentionally retained even though its
// result is discarded by the original function.
func playerCantCastSpell4FD150[O, I comparable](
	unit O,
	spellID int32,
	bypassModeRules int32,
	hooks playerCantCastSpellHooks4FD150[O, I],
) int32 {
	hooks.findParent(unit)

	if bypassModeRules == 0 {
		switch {
		case hooks.hasGameFlag(playerCantCastModeKOTR4FD150) != 0:
			if hooks.loadCrownTypeCache() == 0 {
				hooks.storeCrownTypeCache(hooks.lookupObjectType("Crown"))
			}
			if hooks.spellHasFlags(spellID, playerCantCastSpellCantHoldCrown4FD150) != 0 {
				var zero I
				if first := hooks.loadFirstOwned(unit); first != zero {
					wanted := hooks.loadCrownTypeCache()
					for current := first; current != zero; current = hooks.loadNextOwned(current) {
						if uint32(hooks.loadOwnedType(current)) == wanted {
							if hooks.hasTeam(unit) != 0 {
								return 17
							}
							break
						}
					}
				}
			}

		case hooks.hasGameFlag(playerCantCastModeFlagBall4FD150) != 0:
			if hooks.loadGameBallTypeCache() == 0 {
				hooks.storeGameBallTypeCache(hooks.lookupObjectType("GameBall"))
			}
			if hooks.spellHasFlags(spellID, playerCantCastSpellCantHoldCrown4FD150) != 0 {
				var zero I
				if first := hooks.loadFirstOwned(unit); first != zero {
					wanted := hooks.loadGameBallTypeCache()
					for current := first; current != zero; current = hooks.loadNextOwned(current) {
						if uint32(hooks.loadOwnedType(current)) == wanted {
							return 16
						}
					}
				}
			}

		case hooks.hasGameFlag(playerCantCastModeCTF4FD150) != 0:
			if hooks.spellHasFlags(spellID, playerCantCastSpellCantHoldCrown4FD150) != 0 {
				var zero I
				for item := hooks.loadFirstInventory(unit); item != zero; item = hooks.loadNextInventory(item) {
					if hooks.loadInventoryFlags(item)&playerCantCastInventoryFlag4FD150 != 0 {
						return 13
					}
				}
			}
		}
	}

	if hooks.hasEnchant(unit, playerCantCastAntiMagic4FD150) != 0 {
		return 14
	}
	if hooks.spellAllowed(spellID) == 0 {
		return 10
	}

	count := func(kind playerCantCastSummonType4FD150) int32 {
		return hooks.countOwnedType(unit, hooks.loadSummonType(kind))
	}
	switch spellID {
	case 29:
		if count(playerCantCastLargeFist4FD150) > 0 {
			return 3
		}
		if count(playerCantCastMediumFist4FD150) > 0 {
			return 3
		}
		if count(playerCantCastSmallFist4FD150) > 0 {
			return 3
		}
		return 0
	case 31:
		if count(playerCantCastDeathBall4FD150) > 0 {
			return 3
		}
		return 0
	case 50:
		current := count(playerCantCastMagicMissile4FD150)
		power := hooks.spellPower(spellID, unit)
		limit := x87TruncSignedQwordLow566DCC(hooks.balanceFloat("MagicMissileCount", power-1))
		if current < limit {
			return 0
		}
		return 3
	case 52:
		if count(playerCantCastMeteor4FD150) > 0 {
			return 3
		}
		return 0
	case 58:
		current := count(playerCantCastPixie4FD150)
		power := hooks.spellPower(spellID, unit)
		limit := x87TruncSignedQwordLow566DCC(hooks.balanceFloat("PixieCount", power-1))
		if current < limit {
			return 0
		}
		return 3
	default:
		return 0
	}
}
