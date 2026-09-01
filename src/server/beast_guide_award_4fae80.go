package server

const (
	beastGuideAwardPlayerClass4FAE80        = uint8(0x04)
	beastGuideAwardFirstGuide4FAE80         = int32(1)
	beastGuideAwardGuideCount4FAE80         = int32(41)
	beastGuideAwardLevel4FAE80              = uint32(1)
	beastGuideAwardSound4FAE80              = uint32(227)
	beastGuideAwardSoundKind4FAE80          = int32(0)
	beastGuideAwardRewardKind4FAE80         = int32(1)
	beastGuideAwardInvalidMessage4FAE80     = "AwardGuideError"
	beastGuideAwardMessagePath4FAE80        = `C:\NoxPost\src\Server\Magic\PlyrGide.c`
	beastGuideAwardInvalidMessageLine4FAE80 = 39
)

type beastGuideAwardHooks4FAE80[
	O comparable,
	U any,
	P any,
	R comparable,
	M any,
] struct {
	loadUnitArg      func() O
	loadClassLow     func(O) uint8
	loadUpdateData   func(O) U
	loadPlayer       func(U) P
	loadGuideLevel   func(P, int32) uint32
	storeGuideLevel  func(P, int32, uint32)
	loadProtection   func(P) uint32
	loadString       func(string, string, int) M
	sendLineMessage  func(O, M)
	awardProtection  func(uint32, int32, int32)
	audio            func(uint32, O, int32, uint32)
	rewardNotify     func(O, int32, O, int32)
	relatedGuides    func(int32) []int32
	firstPlayer      func() R
	nextPlayer       func(R) R
	loadPlayerUnit   func(R) O
	reportGuideAward func(O, int32, int32, int32)
}

// beastGuideAward4FAE80 preserves GAME.EXE 004FAE80. The Player class gate
// precedes signed guide validation and UpdateData access. Invalid guide IDs
// take the original localized line-message path; an already-owned guide
// returns silently. UpdateData is cached once, while its Player pointer is
// reloaded for the initial protection update and twice for each related guide.
//
// Any nonzero notify value emits source audio and notification before related
// guide propagation, then walks every active Player record. PlayerUnit is
// loaded from each record and only non-nil units distinct from the source are
// notified. The final report receives the original notify value and zero shop
// flag. There are deliberately no defensive nil or related-index guards.
func beastGuideAward4FAE80[
	O comparable,
	U any,
	P any,
	R comparable,
	M any,
](
	guide, notify int32,
	hooks beastGuideAwardHooks4FAE80[O, U, P, R, M],
) int32 {
	unit := hooks.loadUnitArg()
	if hooks.loadClassLow(unit)&beastGuideAwardPlayerClass4FAE80 == 0 {
		return 0
	}
	if guide < beastGuideAwardFirstGuide4FAE80 || guide >= beastGuideAwardGuideCount4FAE80 {
		message := hooks.loadString(
			beastGuideAwardInvalidMessage4FAE80,
			beastGuideAwardMessagePath4FAE80,
			beastGuideAwardInvalidMessageLine4FAE80,
		)
		hooks.sendLineMessage(unit, message)
		return 0
	}

	update := hooks.loadUpdateData(unit)
	player := hooks.loadPlayer(update)
	if hooks.loadGuideLevel(player, guide) != 0 {
		return 0
	}
	hooks.storeGuideLevel(player, guide, beastGuideAwardLevel4FAE80)

	player = hooks.loadPlayer(update)
	level := hooks.loadGuideLevel(player, guide)
	protection := hooks.loadProtection(player)
	hooks.awardProtection(protection, guide, int32(level))

	if notify != 0 {
		hooks.audio(beastGuideAwardSound4FAE80, unit, beastGuideAwardSoundKind4FAE80, 0)
		hooks.rewardNotify(unit, beastGuideAwardRewardKind4FAE80, unit, guide)
	}

	for _, related := range hooks.relatedGuides(guide) {
		player = hooks.loadPlayer(update)
		hooks.storeGuideLevel(player, related, beastGuideAwardLevel4FAE80)
		player = hooks.loadPlayer(update)
		level = hooks.loadGuideLevel(player, related)
		protection = hooks.loadProtection(player)
		hooks.awardProtection(protection, related, int32(level))
	}

	if notify != 0 {
		var zeroPlayer R
		var zeroUnit O
		for recipient := hooks.firstPlayer(); recipient != zeroPlayer; recipient = hooks.nextPlayer(recipient) {
			recipientUnit := hooks.loadPlayerUnit(recipient)
			if recipientUnit != unit && recipientUnit != zeroUnit {
				hooks.rewardNotify(recipientUnit, beastGuideAwardRewardKind4FAE80, unit, guide)
			}
		}
	}

	hooks.reportGuideAward(unit, guide, notify, 0)
	return 1
}
