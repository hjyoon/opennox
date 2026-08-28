package server

const (
	pickupPotionClassRestrictionFlag4F37D0 = uint32(0x2000)
	pickupPotionQuestFlag4F37D0            = uint32(0x1000)
	pickupPotionPlayerClassLow4F37D0       = uint8(0x04)
	pickupPotionHealthSubClassLow4F37D0    = uint8(0x10)
	pickupPotionManaSubClassLow4F37D0      = uint8(0x20)
	pickupPotionCureSubClassLow4F37D0      = uint8(0x40)

	pickupPotionWarriorClass4F37D0  = uint8(0)
	pickupPotionWizardClass4F37D0   = uint8(1)
	pickupPotionConjurerClass4F37D0 = uint8(2)

	pickupPotionNoCanDoSound4F37D0       = uint32(925)
	pickupPotionRestoreHealthSound4F37D0 = uint32(754)
	pickupPotionRestoreManaSound4F37D0   = uint32(755)
	pickupPotionInventorySound4F37D0     = uint32(832)
	pickupPotionCurePoisonSpell4F37D0    = int32(14)
	pickupPotionOnSoundField4F37D0       = int32(1)
	pickupPotionClassFailMessage4F37D0   = "pickup.c:ObjectEquipClassFail"
)

// pickupPotionHooks4F37D0 exposes every observable field load and call in
// GAME.EXE 004F37D0. Update, Player, HealthData, and PotionUseData are distinct
// comparable tokens so tests can prove which pointers are cached and which
// are reloaded after callbacks mutate the world.
type pickupPotionHooks4F37D0[O, D, H, U, P comparable] struct {
	loadPotionUseData func(O) D
	loadPotionValue   func(D) int32

	gameFlag              func(uint32) int32
	loadOwnerClassLow     func(O) uint8
	loadOwnerUpdate       func(O) U
	loadUpdatePlayer      func(U) P
	loadPlayerClass       func(P) uint8
	playerClassCanUse     func(O, uint8) int32
	classFailureMessage   func(O, string, uint8)
	loadOwnerNetCode      func(O) uint32
	audio                 func(uint32, O, int32, uint32)
	loadPlayerState       func(O) int32
	loadPotionSubClassLow func(O) uint8

	loadOwnerHealth func(O) H
	loadHealthCur   func(H) uint16
	loadHealthMax   func(H) uint16
	scaleHealth     func(int32, uint8) int32
	adjustHealth    func(O, int32)
	scaleMana       func(int32, uint8) int32
	loadManaCur     func(U) uint16
	loadManaMax     func(U) uint16
	addMana         func(O, int32)
	loadOwnerPoison func(O) uint8
	removePoison    func(O)
	spellAudio      func(int32, int32) uint32
	delayedDelete   func(O)
	decay           func(O)
	loadArg4        func() int32
	loadArg3        func() int32
	defaultPickup   func(O, O, int32, int32) int32
}

func pickupPotionKnownClass4F37D0(class uint8) bool {
	return class == pickupPotionWarriorClass4F37D0 ||
		class == pickupPotionWizardClass4F37D0 ||
		class == pickupPotionConjurerClass4F37D0
}

// pickupPotion4F37D0 preserves GAME.EXE 004F37D0.
//
// PotionUseData and its signed Value are eagerly dereferenced before any game
// flag or owner access; there are intentionally no entry nil guards. The
// original Value remains cached for both class-scaled branches. Health reloads
// HealthData for its Cur/Max admission, while mana caches one UpdateData
// pointer for PlayerClass, ManaCur, and ManaMax. Every admission uses a signed
// wrapping int32 sum and consumes only when that sum is strictly below Max.
//
// Cure poison returns canonical one immediately. Any health or mana use also
// returns canonical one after deletion. Otherwise Decay precedes DefaultPickup,
// whose fourth argument is loaded before its third; its full signed int32 is
// returned, and pickup audio is emitted only when that exact result equals one.
func pickupPotion4F37D0[O, D, H, U, P comparable](
	owner, potion O,
	hooks pickupPotionHooks4F37D0[O, D, H, U, P],
) int32 {
	useData := hooks.loadPotionUseData(potion)
	baseAmount := hooks.loadPotionValue(useData)
	amount := baseAmount
	consumed := false

	if hooks.gameFlag(pickupPotionClassRestrictionFlag4F37D0) != 0 &&
		hooks.gameFlag(pickupPotionQuestFlag4F37D0) == 0 &&
		hooks.loadOwnerClassLow(owner)&pickupPotionPlayerClassLow4F37D0 != 0 {
		update := hooks.loadOwnerUpdate(owner)
		player := hooks.loadUpdatePlayer(update)
		class := hooks.loadPlayerClass(player)
		if hooks.playerClassCanUse(potion, class) == 0 {
			hooks.classFailureMessage(owner, pickupPotionClassFailMessage4F37D0, 0)
			netCode := hooks.loadOwnerNetCode(owner)
			hooks.audio(pickupPotionNoCanDoSound4F37D0, owner, 2, netCode)
			return 0
		}
	}

	if hooks.loadPlayerState(owner) == 0 {
		if hooks.loadPotionSubClassLow(potion)&pickupPotionHealthSubClassLow4F37D0 != 0 {
			health := hooks.loadOwnerHealth(owner)
			var zeroHealth H
			if health != zeroHealth {
				if hooks.loadOwnerClassLow(owner)&pickupPotionPlayerClassLow4F37D0 != 0 {
					update := hooks.loadOwnerUpdate(owner)
					player := hooks.loadUpdatePlayer(update)
					class := hooks.loadPlayerClass(player)
					if pickupPotionKnownClass4F37D0(class) {
						amount = hooks.scaleHealth(baseAmount, class)
					}
				}

				health = hooks.loadOwnerHealth(owner)
				cur := int32(hooks.loadHealthCur(health))
				max := int32(hooks.loadHealthMax(health))
				if cur+amount < max {
					hooks.adjustHealth(owner, amount)
					hooks.audio(pickupPotionRestoreHealthSound4F37D0, owner, 0, 0)
					consumed = true
				}
			}
		}

		if hooks.loadPotionSubClassLow(potion)&pickupPotionManaSubClassLow4F37D0 != 0 &&
			hooks.loadOwnerClassLow(owner)&pickupPotionPlayerClassLow4F37D0 != 0 {
			update := hooks.loadOwnerUpdate(owner)
			player := hooks.loadUpdatePlayer(update)
			class := hooks.loadPlayerClass(player)
			if pickupPotionKnownClass4F37D0(class) {
				amount = hooks.scaleMana(baseAmount, class)
			}
			cur := int32(hooks.loadManaCur(update))
			max := int32(hooks.loadManaMax(update))
			if cur+amount < max {
				hooks.addMana(owner, amount)
				hooks.audio(pickupPotionRestoreManaSound4F37D0, owner, 0, 0)
				consumed = true
			}
		}

		if hooks.loadPotionSubClassLow(potion)&pickupPotionCureSubClassLow4F37D0 != 0 &&
			hooks.loadOwnerClassLow(owner)&pickupPotionPlayerClassLow4F37D0 != 0 &&
			hooks.loadOwnerPoison(owner) != 0 {
			hooks.removePoison(owner)
			sound := hooks.spellAudio(
				pickupPotionCurePoisonSpell4F37D0,
				pickupPotionOnSoundField4F37D0,
			)
			hooks.audio(sound, owner, 0, 0)
			hooks.delayedDelete(potion)
			return 1
		}

		if consumed {
			hooks.delayedDelete(potion)
			return 1
		}
	}

	hooks.decay(potion)
	arg4 := hooks.loadArg4()
	arg3 := hooks.loadArg3()
	result := hooks.defaultPickup(owner, potion, arg3, arg4)
	if result == 1 {
		hooks.audio(pickupPotionInventorySound4F37D0, owner, 0, 0)
	}
	return result
}
