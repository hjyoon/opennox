package server

const (
	fieldGuideUsePlayerClass53F930      = uint8(0x04)
	fieldGuideUseQuestFlag53F930        = uint32(0x1000)
	fieldGuideUseConjurerClass53F930    = uint8(2)
	fieldGuideUseClassMessage53F930     = "pickup.c:ObjectEquipClassFail"
	fieldGuideUseDuplicateMessage53F930 = "objcoll.c:AlreadyHaveGuide"
)

type fieldGuideUseHooks53F930[O any, U any, P any, D any] struct {
	loadOwnerArg      func() O
	loadClassLow      func(O) uint8
	loadItemArg       func() O
	loadUpdateData    func(O) U
	loadUseData       func(O) D
	loadCreature      func(D) string
	guideByName       func(string) int32
	gameFlagsCheck    func(uint32) int32
	loadPlayer        func(U) P
	loadPlayerClass   func(P) uint8
	loadGuideLevel    func(P, int32) uint32
	primaryMessage    func(O, string, uint8)
	awardGuide        func(O, int32, int32) int32
	delayedDeleteItem func(O)
}

// useFieldGuide53F930 preserves GAME.EXE 0053F930. The owner class is read
// before the item argument, so a non-Player never observes item state. On the
// Player path the item, owner UpdateData, and item UseData are cached before
// either lookup callback. The Player pointer itself remains live: Quest class
// validation and ownership testing reload it separately from cached
// UpdateData.
//
// Quest mode admits only exact class two. An already-owned guide emits the
// original priority message and leaves the item intact. Otherwise the guide
// award return value is deliberately ignored, the cached item is scheduled
// for deletion, and the function returns canonical one. There are no nil or
// guide-index guards beyond the original Player-class gate.
func useFieldGuide53F930[O any, U any, P any, D any](
	hooks fieldGuideUseHooks53F930[O, U, P, D],
) int32 {
	owner := hooks.loadOwnerArg()
	if hooks.loadClassLow(owner)&fieldGuideUsePlayerClass53F930 == 0 {
		return 0
	}

	item := hooks.loadItemArg()
	update := hooks.loadUpdateData(owner)
	data := hooks.loadUseData(item)
	creature := hooks.loadCreature(data)
	guide := hooks.guideByName(creature)
	if hooks.gameFlagsCheck(fieldGuideUseQuestFlag53F930) != 0 {
		player := hooks.loadPlayer(update)
		if hooks.loadPlayerClass(player) != fieldGuideUseConjurerClass53F930 {
			hooks.primaryMessage(owner, fieldGuideUseClassMessage53F930, 0)
			return 0
		}
	}

	player := hooks.loadPlayer(update)
	if hooks.loadGuideLevel(player, guide) != 0 {
		hooks.primaryMessage(owner, fieldGuideUseDuplicateMessage53F930, 0)
		return 0
	}

	hooks.awardGuide(owner, guide, 1)
	hooks.delayedDeleteItem(item)
	return 1
}
