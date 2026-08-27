package server

type randomFieldGuideLossEligibilityHooks4F2530 struct {
	loadGuideID func(index int) uint32
	loadSlots   func(index int) uint32
}

// randomFieldGuideLossEligible4F2530 reconstructs GAME.EXE 004F2530. The
// original starts at the GuideID member of the first twelve-byte reward-guide
// row. It scans live GuideID values until the zero-ID sentinel, reads Slots
// only for a matching ID, and accepts the first matching row whose Slots value
// is nonzero. A zero guide ID remains ineligible even after a match, although
// the sentinel makes that final original check redundant for a well-formed
// table.
func randomFieldGuideLossEligible4F2530(
	guideID int32,
	hooks randomFieldGuideLossEligibilityHooks4F2530,
) int32 {
	index := 0
	currentGuideID := hooks.loadGuideID(index)
	if currentGuideID == 0 {
		return 0
	}
	target := uint32(guideID)
	for {
		if currentGuideID == target && hooks.loadSlots(index) != 0 {
			break
		}
		currentGuideID = hooks.loadGuideID(index + 1)
		index++
		if currentGuideID == 0 {
			return 0
		}
	}
	if guideID != 0 {
		return 1
	}
	return 0
}

// RandomFieldGuideLossEligible4F2530 binds the original predicate to the
// exact fixed-width reward-guide rows reconstructed from GAME.EXE 005B9BB0.
//
//go:noinline
func RandomFieldGuideLossEligible4F2530(guideID int32) int32 {
	return randomFieldGuideLossEligible4F2530(guideID, randomFieldGuideLossEligibilityHooks4F2530{
		loadGuideID: func(index int) uint32 {
			return rewardFieldGuideDefinitions4F0D20[index].GuideID
		},
		loadSlots: func(index int) uint32 {
			return rewardFieldGuideDefinitions4F0D20[index].Slots
		},
	})
}
