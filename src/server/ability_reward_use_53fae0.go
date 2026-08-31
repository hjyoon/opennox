package server

const (
	abilityRewardUsePlayerClass53FAE0  = uint8(0x04)
	abilityRewardUseQuestMask53FAE0    = uint32(0x1800)
	abilityRewardUseAuditSound53FAE0   = int32(925)
	abilityRewardUseAuditKind53FAE0    = int32(2)
	abilityRewardUseClassMessage53FAE0 = "pickup.c:ObjectEquipClassFail"
)

type abilityRewardUseHooks53FAE0[O any, U any, P any, D any] struct {
	loadItemArg      func() O
	loadOwnerArg     func() O
	loadUseData      func(O) D
	loadClassLow     func(O) uint8
	loadUpdateData   func(O) U
	loadPlayer       func(U) P
	loadPlayerClass  func(P) uint8
	primaryMessage   func(O, string, uint8)
	loadNetCode      func(O) uint32
	audit            func(int32, O, int32, uint32)
	gameFlagsCheck   func(uint32) int32
	loadAbility      func(D) uint8
	loadAbilityLevel func(P, int32) uint32
	rewardAbility    func(O, int32, int32) int32
	delayedDelete    func(O)
}

// useAbilityReward53FAE0 preserves GAME.EXE 0053FAE0. The item UseData pointer
// is cached before any owner dereference. The owner class byte is read before
// UpdateData, but UpdateData is still loaded before the class branch. Only a
// Player whose class byte is exactly zero may use the reward.
//
// The Quest/Solo mask conditionally observes the cached UpdateData's live
// Player and the cached UseData's live ability byte. The ability byte is loaded
// again for the service call, so callback mutation between those reads is
// visible. A valid-class path always returns one: service success schedules the
// item for deletion, while any zero service result emits the exact audit event.
// There are deliberately no defensive nil guards.
func useAbilityReward53FAE0[O any, U any, P any, D any](
	hooks abilityRewardUseHooks53FAE0[O, U, P, D],
) int32 {
	item := hooks.loadItemArg()
	owner := hooks.loadOwnerArg()
	data := hooks.loadUseData(item)
	classLow := hooks.loadClassLow(owner)
	update := hooks.loadUpdateData(owner)
	if classLow&abilityRewardUsePlayerClass53FAE0 == 0 {
		return 0
	}

	player := hooks.loadPlayer(update)
	if hooks.loadPlayerClass(player) != 0 {
		hooks.primaryMessage(owner, abilityRewardUseClassMessage53FAE0, 0)
		netCode := hooks.loadNetCode(owner)
		hooks.audit(abilityRewardUseAuditSound53FAE0, owner, abilityRewardUseAuditKind53FAE0, netCode)
		return 0
	}

	rewardArg := int32(0)
	if hooks.gameFlagsCheck(abilityRewardUseQuestMask53FAE0) != 0 {
		player = hooks.loadPlayer(update)
		ability := int32(hooks.loadAbility(data))
		if hooks.loadAbilityLevel(player, ability) == 0 {
			rewardArg = 1
		}
	}
	ability := int32(hooks.loadAbility(data))
	if hooks.rewardAbility(owner, ability, rewardArg) != 0 {
		hooks.delayedDelete(item)
	} else {
		netCode := hooks.loadNetCode(owner)
		hooks.audit(abilityRewardUseAuditSound53FAE0, owner, abilityRewardUseAuditKind53FAE0, netCode)
	}
	return 1
}
