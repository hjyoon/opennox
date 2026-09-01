package server

const (
	spellRewardUsePlayerClass53F9E0  = uint8(0x04)
	spellRewardUseQuestMask53F9E0    = uint32(0x1800)
	spellRewardUseWarrior53F9E0      = uint8(1)
	spellRewardUseConjurer53F9E0     = uint8(2)
	spellRewardUseFailureSound53F9E0 = int32(925)
	spellRewardUseFailureKind53F9E0  = int32(2)
	spellRewardUseFailureMsg53F9E0   = "use.c:SpellRewardClassFail"
)

type spellRewardUseHooks53F9E0[O any, U any, P any, D any] struct {
	loadItemArg       func() O
	loadOwnerArg      func() O
	loadUseData       func(O) D
	loadClassLow      func(O) uint8
	loadUpdateData    func(O) U
	loadPlayer        func(U) P
	loadPlayerClass   func(P) uint8
	loadSpell         func(D) uint8
	checkSpellClass   func(uint8, uint8) int32
	primaryMessage    func(O, string, uint8)
	loadNetCode       func(O) uint32
	audit             func(int32, O, int32, uint32)
	gameFlagsCheck    func(uint32) int32
	loadSpellLevel    func(P, int32) uint32
	grantSpell        func(O, int32, int32, int32, int32) int32
	delayedDeleteItem func(O)
}

// useSpellReward53F9E0 preserves GAME.EXE 0053F9E0. The item UseData pointer
// and owner UpdateData pointer are cached before the Player-class branch. The
// Player pointer and one-byte spell ID remain live: class validation reads the
// spell once, the Quest level check reads it again, and the grant service reads
// it a third time.
//
// Only exact player classes one and two may continue. Both an invalid class
// and a class-incompatible spell emit the original priority message, then load
// the owner's live NetCode for sound 925, and return canonical zero. A valid
// class path always returns one. Grant success schedules the cached item for
// deletion; grant failure emits the same live-NetCode sound. There are
// deliberately no nil, spell-index, or callback guards beyond the original
// Player-class bit test.
func useSpellReward53F9E0[O any, U any, P any, D any](
	hooks spellRewardUseHooks53F9E0[O, U, P, D],
) int32 {
	item := hooks.loadItemArg()
	owner := hooks.loadOwnerArg()
	data := hooks.loadUseData(item)
	classLow := hooks.loadClassLow(owner)
	update := hooks.loadUpdateData(owner)
	if classLow&spellRewardUsePlayerClass53F9E0 == 0 {
		return 0
	}

	player := hooks.loadPlayer(update)
	class := hooks.loadPlayerClass(player)
	if class != spellRewardUseWarrior53F9E0 && class != spellRewardUseConjurer53F9E0 {
		hooks.primaryMessage(owner, spellRewardUseFailureMsg53F9E0, 0)
		netCode := hooks.loadNetCode(owner)
		hooks.audit(spellRewardUseFailureSound53F9E0, owner, spellRewardUseFailureKind53F9E0, netCode)
		return 0
	}

	spell := hooks.loadSpell(data)
	if hooks.checkSpellClass(class, spell) != 0 {
		hooks.primaryMessage(owner, spellRewardUseFailureMsg53F9E0, 0)
		netCode := hooks.loadNetCode(owner)
		hooks.audit(spellRewardUseFailureSound53F9E0, owner, spellRewardUseFailureKind53F9E0, netCode)
		return 0
	}

	questArg := int32(0)
	if hooks.gameFlagsCheck(spellRewardUseQuestMask53F9E0) != 0 {
		player = hooks.loadPlayer(update)
		spell = hooks.loadSpell(data)
		if hooks.loadSpellLevel(player, int32(spell)) == 0 {
			questArg = 1
		}
	}

	spell = hooks.loadSpell(data)
	if hooks.grantSpell(owner, int32(spell), 1, questArg, 0) != 0 {
		hooks.delayedDeleteItem(item)
	} else {
		netCode := hooks.loadNetCode(owner)
		hooks.audit(spellRewardUseFailureSound53F9E0, owner, spellRewardUseFailureKind53F9E0, netCode)
	}
	return 1
}
