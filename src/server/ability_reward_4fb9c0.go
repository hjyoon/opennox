package server

const (
	abilityRewardPlayerClass4FB9C0        = uint8(0x04)
	abilityRewardFirstAbility4FB9C0       = int32(1)
	abilityRewardAbilityLimit4FB9C0       = int32(6)
	abilityRewardLevel4FB9C0              = uint32(5)
	abilityRewardQuestFlag4FB9C0          = uint32(0x1000)
	abilityRewardNotifyKind4FB9C0         = int32(2)
	abilityRewardInvalidMessage4FB9C0     = "AwardAbilityError"
	abilityRewardOwnedMessage4FB9C0       = "use.c:HadAbility"
	abilityRewardMessagePath4FB9C0        = `C:\NoxPost\src\Server\Ability\Ability.c`
	abilityRewardInvalidMessageLine4FB9C0 = 108
)

type abilityRewardHooks4FB9C0[
	O comparable,
	U any,
	P comparable,
	M any,
] struct {
	loadUnitArg       func() O
	loadClassLow      func(O) uint8
	loadUpdateData    func(O) U
	loadPlayer        func(U) P
	loadAbilityLevel  func(P, int32) uint32
	storeAbilityLevel func(P, int32, uint32)
	loadProtection    func(P) uint32
	loadString        func(string, string, int) M
	sendLineMessage   func(O, M)
	primaryMessage    func(O, string, uint8)
	awardProtection   func(uint32, int32, int32)
	reportAbility     func(O, int32, int32)
	gameFlagsCheck    func(uint32) int32
	rewardNotify      func(O, int32, O, int32)
	checkPlayerState  func(O) int32
	firstPlayerUnit   func() O
	nextPlayerUnit    func(O) O
}

// abilityRewardServ4FB9C0 preserves GAME.EXE 004FB9C0. It intentionally has
// no nil guard before the Player-class read. Invalid signed ability IDs are
// rejected before UpdateData is touched. For a valid ID, UpdateData is cached
// once while its Player pointer is reloaded at the three original observation
// points: initial ownership, signed level clamp, and protection update.
//
// The redundant level-five clamp is retained because callback mutation can
// replace the live Player. Its JLE is signed, so values with bit 31 set are not
// clamped. The report precedes the Quest flag check; Quest notifies the source
// first and broadcasts only when the source is not in the suppressed player
// state. Iterator results are native object identities and the source is
// skipped by exact identity comparison.
func abilityRewardServ4FB9C0[
	O comparable,
	U any,
	P comparable,
	M any,
](
	ability, rewardArg int32,
	hooks abilityRewardHooks4FB9C0[O, U, P, M],
) int32 {
	unit := hooks.loadUnitArg()
	if hooks.loadClassLow(unit)&abilityRewardPlayerClass4FB9C0 == 0 {
		return 0
	}
	if ability < abilityRewardFirstAbility4FB9C0 || ability >= abilityRewardAbilityLimit4FB9C0 {
		message := hooks.loadString(
			abilityRewardInvalidMessage4FB9C0,
			abilityRewardMessagePath4FB9C0,
			abilityRewardInvalidMessageLine4FB9C0,
		)
		hooks.sendLineMessage(unit, message)
		return 0
	}

	update := hooks.loadUpdateData(unit)
	player := hooks.loadPlayer(update)
	if hooks.loadAbilityLevel(player, ability) != 0 {
		hooks.primaryMessage(unit, abilityRewardOwnedMessage4FB9C0, 0)
		return 0
	}

	hooks.storeAbilityLevel(player, ability, abilityRewardLevel4FB9C0)
	player = hooks.loadPlayer(update)
	level := hooks.loadAbilityLevel(player, ability)
	if int32(level) > int32(abilityRewardLevel4FB9C0) {
		hooks.storeAbilityLevel(player, ability, abilityRewardLevel4FB9C0)
	}

	player = hooks.loadPlayer(update)
	level = hooks.loadAbilityLevel(player, ability)
	protection := hooks.loadProtection(player)
	hooks.awardProtection(protection, ability, int32(level))
	hooks.reportAbility(unit, ability, rewardArg)

	if hooks.gameFlagsCheck(abilityRewardQuestFlag4FB9C0) != 0 {
		hooks.rewardNotify(unit, abilityRewardNotifyKind4FB9C0, unit, ability)
		if hooks.checkPlayerState(unit) == 0 {
			var zero O
			for recipient := hooks.firstPlayerUnit(); recipient != zero; recipient = hooks.nextPlayerUnit(recipient) {
				if recipient != unit {
					hooks.rewardNotify(recipient, abilityRewardNotifyKind4FB9C0, unit, ability)
				}
			}
		}
	}
	return 1
}
