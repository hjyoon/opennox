package server

const (
	crownPickupPlayerClass4F3400  = uint8(0x04)
	crownPickupEnchant4F3400      = uint32(30)
	crownPickupEnchantDur4F3400   = uint32(0)
	crownPickupEnchantPower4F3400 = uint32(5)
	crownPickupSound4F3400        = uint32(313)
	crownPickupInformCode4F3400   = uint8(10)
	crownPickupMinimapFlag4F3400  = uint32(1)
)

type crownPickupHooks4F3400[O, D, U any] struct {
	loadCrownUpdate  func(O) D
	loadClassLow     func(O) uint8
	defaultPickup    func(O, O, int32, int32) uint32
	loadPlayerUpdate func(O) U
	loadFrame        func() uint32
	storePickupFrame func(U, uint32)
	setOwner         func(O, O)
	applyEnchant     func(O, uint32, uint32, uint32)
	playAudio        func(uint32, O, int32, uint32)
	loadNetCode      func(O) uint32
	hasTeam          func(O) bool
	loadTeamID       func(O) uint8
	informPickup     func(uint8, uint32, uint32)
	unmarkMinimap    func(O, uint32)
	clearPending     func(D)
}

// crownPickup4F3400 preserves GAME.EXE 004F3400. The crown update pointer is
// cached before the Player gate. On the Player path its pending target is
// cleared last, even when DefaultPickup fails; the non-Player path leaves it
// untouched. Both pickup flags are forwarded exactly as supplied.
func crownPickup4F3400[O, D, U any](
	who, crown O,
	flag1, flag2 int32,
	hooks crownPickupHooks4F3400[O, D, U],
) uint32 {
	update := hooks.loadCrownUpdate(crown)
	if hooks.loadClassLow(who)&crownPickupPlayerClass4F3400 == 0 {
		return 0
	}

	result := hooks.defaultPickup(who, crown, flag1, flag2)
	if result != 0 {
		playerUpdate := hooks.loadPlayerUpdate(who)
		frame := hooks.loadFrame()
		hooks.storePickupFrame(playerUpdate, frame)
		hooks.setOwner(who, crown)
		hooks.applyEnchant(
			who,
			crownPickupEnchant4F3400,
			crownPickupEnchantDur4F3400,
			crownPickupEnchantPower4F3400,
		)
		hooks.playAudio(crownPickupSound4F3400, who, 0, 0)
		netCode := hooks.loadNetCode(who)
		teamID := uint32(0)
		if hooks.hasTeam(who) {
			teamID = uint32(hooks.loadTeamID(who))
		}
		hooks.informPickup(crownPickupInformCode4F3400, netCode, teamID)
		hooks.unmarkMinimap(crown, crownPickupMinimapFlag4F3400)
	}
	hooks.clearPending(update)
	return result
}
