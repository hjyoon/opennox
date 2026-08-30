package server

const (
	spellGrantPlayerClass4FB550        = uint8(0x04)
	spellGrantFirstSpell4FB550         = int32(1)
	spellGrantSpellCount4FB550         = int32(137)
	spellGrantCoopQuestFlag4FB550      = uint32(0x1800)
	spellGrantQuestFlag4FB550          = uint32(0x1000)
	spellGrantSoloFlag4FB550           = uint32(0x0800)
	spellGrantMaxCoopLevel4FB550       = uint32(3)
	spellGrantMaxLevel4FB550           = uint32(5)
	spellGrantFamilySourceA4FB550      = int32(0x1000)
	spellGrantFamilyMemberA4FB550      = int32(0x2000)
	spellGrantFamilySourceB4FB550      = int32(0x4000)
	spellGrantFamilyMemberB4FB550      = int32(0x8000)
	spellGrantFamilySourceC4FB550      = int32(0x10000)
	spellGrantFamilyMemberC4FB550      = int32(0x20000)
	spellGrantSoloSuppressMask4FB550   = int32(0x15000)
	spellGrantGlyphSpell4FB550         = int32(34)
	spellGrantSound4FB550              = uint32(226)
	spellGrantSoundKind4FB550          = int32(0)
	spellGrantRewardKind4FB550         = int32(0)
	spellGrantInvalidMessage4FB550     = "AwardSpellError"
	spellGrantMaxMessage4FB550         = "MaxSpellLevel"
	spellGrantMessagePath4FB550        = `C:\NoxPost\src\Server\Magic\plyrspel.c`
	spellGrantInvalidMessageLine4FB550 = 339
	spellGrantMaxMessageLine4FB550     = 351
	spellGrantSingleMessageLine4FB550  = 386
)

type spellGrantHooks4FB550[
	O comparable,
	U any,
	P comparable,
	T comparable,
	M any,
] struct {
	loadUnitArg      func() O
	loadClassLow     func(O) uint8
	loadUpdateData   func(O) U
	loadPlayer       func(U) P
	loadSpellLevel   func(P, int32) uint32
	storeSpellLevel  func(P, int32, uint32)
	loadProtection   func(P) uint32
	gameFlagsCheck   func(uint32) int32
	loadString       func(string, string, int) M
	sendLineMessage  func(O, M)
	awardProtection  func(uint32, int32, int32)
	spellHasFlags    func(int32, int32) int32
	spellIsValid     func(int32) int32
	audio            func(uint32, O, int32, uint32)
	loadNotifyField  func(P) uint32
	rewardNotify     func(O, int32, O, int32)
	checkPlayerState func(O) int32
	firstPlayer      func() P
	nextPlayer       func(P) P
	loadPlayerUnit   func(P) O
	loadTrade        func(U) T
	shopExit         func(T)
	reportSpellAward func(O, int32, int32, int32)
}

func spellGrantQuestSingleLevel4FB550(spellID int32) bool {
	switch spellID {
	case 34, 134, 45, 46, 47, 48, 49,
		117, 118, 119, 120, 121, 122, 123, 124, 125, 19:
		return true
	default:
		return false
	}
}

// spellGrantToPlayer4FB550 preserves GAME.EXE 004FB550. The original unit is
// loaded before its Player class gate; invalid spell IDs therefore do not
// touch UpdateData. UpdateData is cached once after validation, while its
// Player pointer, spell levels, protection token, notify field, and trade
// session remain live observations at their original access points.
//
// Level increments wrap at 32 bits. The x86 JLE clamps are signed comparisons,
// and a nonzero override is stored with its exact 32-bit bit pattern after the
// selected-spell clamps. Family propagation deliberately keeps the original
// executable's asymmetric behavior: it stores and max-clamps each matching
// member, but the Quest clamp still targets the initially selected spell and
// protection pairs that selected ID with the member's live level. A selected
// spell that is itself a family member is consequently updated and protected
// twice.
//
// Notification arguments retain their distinct predicates. Any nonzero value
// emits audio, but Solo shop close requires both notify and shop to equal one;
// the final report always receives the original 32-bit arguments. There are no
// defensive nil guards beyond the explicit class and iterator checks present
// in the executable.
func spellGrantToPlayer4FB550[
	O comparable,
	U any,
	P comparable,
	T comparable,
	M any,
](
	spellID, notify, shop, override int32,
	hooks spellGrantHooks4FB550[O, U, P, T, M],
) int32 {
	unit := hooks.loadUnitArg()
	if hooks.loadClassLow(unit)&spellGrantPlayerClass4FB550 == 0 {
		return 0
	}
	if spellID < spellGrantFirstSpell4FB550 || spellID >= spellGrantSpellCount4FB550 {
		message := hooks.loadString(
			spellGrantInvalidMessage4FB550,
			spellGrantMessagePath4FB550,
			spellGrantInvalidMessageLine4FB550,
		)
		hooks.sendLineMessage(unit, message)
		return 0
	}

	update := hooks.loadUpdateData(unit)
	if hooks.gameFlagsCheck(spellGrantCoopQuestFlag4FB550) != 0 {
		player := hooks.loadPlayer(update)
		if hooks.loadSpellLevel(player, spellID) == spellGrantMaxCoopLevel4FB550 {
			message := hooks.loadString(
				spellGrantMaxMessage4FB550,
				spellGrantMessagePath4FB550,
				spellGrantMaxMessageLine4FB550,
			)
			hooks.sendLineMessage(unit, message)
			return 0
		}
	}
	player := hooks.loadPlayer(update)
	if hooks.loadSpellLevel(player, spellID) == spellGrantMaxLevel4FB550 {
		message := hooks.loadString(
			spellGrantMaxMessage4FB550,
			spellGrantMessagePath4FB550,
			spellGrantMaxMessageLine4FB550,
		)
		hooks.sendLineMessage(unit, message)
		return 0
	}

	if hooks.gameFlagsCheck(spellGrantQuestFlag4FB550) != 0 && spellGrantQuestSingleLevel4FB550(spellID) {
		player = hooks.loadPlayer(update)
		if hooks.loadSpellLevel(player, spellID) != 0 {
			message := hooks.loadString(
				spellGrantMaxMessage4FB550,
				spellGrantMessagePath4FB550,
				spellGrantSingleMessageLine4FB550,
			)
			hooks.sendLineMessage(unit, message)
			return 0
		}
	}

	player = hooks.loadPlayer(update)
	level := hooks.loadSpellLevel(player, spellID)
	level++
	hooks.storeSpellLevel(player, spellID, level)

	player = hooks.loadPlayer(update)
	level = hooks.loadSpellLevel(player, spellID)
	if int32(level) > int32(spellGrantMaxLevel4FB550) {
		hooks.storeSpellLevel(player, spellID, spellGrantMaxLevel4FB550)
	}

	if hooks.gameFlagsCheck(spellGrantQuestFlag4FB550) != 0 {
		player = hooks.loadPlayer(update)
		level = hooks.loadSpellLevel(player, spellID)
		if int32(level) > int32(spellGrantMaxCoopLevel4FB550) {
			hooks.storeSpellLevel(player, spellID, spellGrantMaxCoopLevel4FB550)
		}
	}
	if override != 0 {
		player = hooks.loadPlayer(update)
		hooks.storeSpellLevel(player, spellID, uint32(override))
	}

	player = hooks.loadPlayer(update)
	level = hooks.loadSpellLevel(player, spellID)
	protection := hooks.loadProtection(player)
	hooks.awardProtection(protection, spellID, int32(level))

	family := int32(0)
	switch {
	case hooks.spellHasFlags(spellID, spellGrantFamilySourceA4FB550) != 0:
		family = spellGrantFamilyMemberA4FB550
	case hooks.spellHasFlags(spellID, spellGrantFamilySourceB4FB550) != 0:
		family = spellGrantFamilyMemberB4FB550
	case hooks.spellHasFlags(spellID, spellGrantFamilySourceC4FB550) != 0:
		family = spellGrantFamilyMemberC4FB550
	}
	if family != 0 {
		for related := spellGrantFirstSpell4FB550; related < spellGrantSpellCount4FB550; related++ {
			if hooks.spellHasFlags(related, family) == 0 || hooks.spellIsValid(related) == 0 {
				continue
			}

			player = hooks.loadPlayer(update)
			if override != 0 {
				hooks.storeSpellLevel(player, related, uint32(override))
			} else {
				level = hooks.loadSpellLevel(player, related)
				level++
				hooks.storeSpellLevel(player, related, level)
			}

			player = hooks.loadPlayer(update)
			level = hooks.loadSpellLevel(player, related)
			if int32(level) > int32(spellGrantMaxLevel4FB550) {
				hooks.storeSpellLevel(player, related, spellGrantMaxLevel4FB550)
			}

			if hooks.gameFlagsCheck(spellGrantQuestFlag4FB550) != 0 {
				player = hooks.loadPlayer(update)
				level = hooks.loadSpellLevel(player, spellID)
				if int32(level) > int32(spellGrantMaxCoopLevel4FB550) {
					hooks.storeSpellLevel(player, spellID, spellGrantMaxCoopLevel4FB550)
				}
			}

			player = hooks.loadPlayer(update)
			level = hooks.loadSpellLevel(player, related)
			protection = hooks.loadProtection(player)
			hooks.awardProtection(protection, spellID, int32(level))
		}
		unit = hooks.loadUnitArg()
	}

	if notify != 0 {
		shouldNotify := int32(1)
		hooks.audio(spellGrantSound4FB550, unit, spellGrantSoundKind4FB550, 0)
		if hooks.gameFlagsCheck(spellGrantSoloFlag4FB550) != 0 {
			if spellID == spellGrantGlyphSpell4FB550 ||
				hooks.spellHasFlags(spellID, spellGrantSoloSuppressMask4FB550) == 1 {
				shouldNotify = 0
			}
		}
		if hooks.gameFlagsCheck(spellGrantQuestFlag4FB550) != 0 {
			player = hooks.loadPlayer(update)
			if hooks.loadNotifyField(player) == 0 {
				shouldNotify = 0
			}
		}
		if shouldNotify != 0 {
			hooks.rewardNotify(unit, spellGrantRewardKind4FB550, unit, spellID)
			if hooks.checkPlayerState(unit) == 0 {
				var zeroPlayer P
				var zeroUnit O
				for recipient := hooks.firstPlayer(); recipient != zeroPlayer; recipient = hooks.nextPlayer(recipient) {
					recipientUnit := hooks.loadPlayerUnit(recipient)
					if recipientUnit != unit && recipientUnit != zeroUnit {
						hooks.rewardNotify(recipientUnit, spellGrantRewardKind4FB550, unit, spellID)
					}
				}
			}
		}
	}

	if hooks.gameFlagsCheck(spellGrantSoloFlag4FB550) == 1 && notify == 1 && shop == 1 {
		trade := hooks.loadTrade(update)
		var zeroTrade T
		if trade != zeroTrade {
			hooks.shopExit(trade)
		}
	}
	hooks.reportSpellAward(unit, spellID, notify, shop)
	return 1
}
