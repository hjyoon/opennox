package server

const (
	crownDropGameFlag4ED5E0     = uint32(16)
	crownDropGameplayFlag4ED5E0 = uint32(4)
	crownDropEnchant4ED5E0      = uint32(30)
	crownDropInformCode4ED5E0   = uint8(11)
	crownDropMinimapFlag4ED5E0  = uint32(1)
)

// crownDropHooks4ED5E0 exposes the delayed argument loads and callback
// boundaries in GAME.EXE 004ED5E0. The Crown update pointer and each Player
// update pointer are cached at their original loads; Crown team IDs and the
// final owner fields remain live reads.
type crownDropHooks4ED5E0[O comparable, D, U, P any] struct {
	gameFlag     func(uint32) int32
	gameplayFlag func(uint32) int32

	loadCrownArg    func() O
	loadFrame       func() uint32
	loadTeamID      func(O) uint8
	loadCrownUpdate func(O) D
	firstPlayer     func() O
	loadPlayerData  func(O) U
	teamContains    func(O, uint8) int32
	loadPickupFrame func(U) uint32
	nextPlayer      func(O) O

	loadOwnerArg      func() O
	storePickupTarget func(D, O)
	loadPointArg      func() P
	defaultDrop       func(O, O, P) int32
	clearOwner        func(O)
	buffOff           func(O, uint32)
	loadNetCode       func(O) uint32
	hasTeam           func(O) int32
	informDrop        func(uint8, uint32, uint32)
	markMinimap       func(O, uint32)
}

// crownDrop4ED5E0 preserves GAME.EXE 004ED5E0. KOTR target selection compares
// pickup frames as signed IA-32 integers and keeps the first minimum. With no
// Player unit, the Crown itself remains the candidate. DefaultDrop gates every
// post-drop side effect with its full EAX result.
func crownDrop4ED5E0[O comparable, D, U, P any](hooks crownDropHooks4ED5E0[O, D, U, P]) int32 {
	var zero O
	var crown O
	var owner O

	if hooks.gameFlag(crownDropGameFlag4ED5E0) != 0 &&
		hooks.gameplayFlag(crownDropGameplayFlag4ED5E0) != 0 {
		crown = hooks.loadCrownArg()
		bestFrame := int32(hooks.loadFrame())
		if hooks.loadTeamID(crown) != 0 {
			update := hooks.loadCrownUpdate(crown)
			player := hooks.firstPlayer()
			candidate := hooks.loadCrownArg()
			if player != zero {
				for player != zero {
					liveCrown := hooks.loadCrownArg()
					playerData := hooks.loadPlayerData(player)
					teamID := hooks.loadTeamID(liveCrown)
					if hooks.teamContains(player, teamID) != 0 {
						frame := int32(hooks.loadPickupFrame(playerData))
						if frame < bestFrame {
							bestFrame = frame
							candidate = player
						}
					}
					player = hooks.nextPlayer(player)
				}
				crown = hooks.loadCrownArg()
			}
			owner = hooks.loadOwnerArg()
			if candidate != zero && owner != candidate {
				hooks.storePickupTarget(update, candidate)
			}
		} else {
			owner = hooks.loadOwnerArg()
		}
	} else {
		crown = hooks.loadCrownArg()
		owner = hooks.loadOwnerArg()
	}

	point := hooks.loadPointArg()
	if hooks.defaultDrop(owner, crown, point) == 0 {
		return 0
	}
	hooks.clearOwner(crown)
	hooks.buffOff(owner, crownDropEnchant4ED5E0)
	netCode := hooks.loadNetCode(owner)
	teamID := uint32(0)
	if hooks.hasTeam(owner) != 0 {
		teamID = uint32(hooks.loadTeamID(owner))
	}
	hooks.informDrop(crownDropInformCode4ED5E0, netCode, teamID)
	hooks.markMinimap(crown, crownDropMinimapFlag4ED5E0)
	return 1
}
