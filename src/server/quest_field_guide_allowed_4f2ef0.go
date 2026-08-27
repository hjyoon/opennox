package server

type questFieldGuideFamily4F2EF0 [6]uint32

// questFieldGuideFamily24Data4F2EF0 is the native fixed-width meaning of the
// zero-terminated family list at GAME.EXE 005B98A0. The first value is a
// header; only the following values are tested by 004F2EF0.
var questFieldGuideFamily24Data4F2EF0 = questFieldGuideFamily4F2EF0{24, 7, 8, 25, 26, 0}

// questFieldGuideFamilies4F2EF0 is the native-pointer form of the PE32 table
// at GAME.EXE 005B98B8: one family pointer followed by a null sentinel.
var questFieldGuideFamilies4F2EF0 = [...]*questFieldGuideFamily4F2EF0{
	&questFieldGuideFamily24Data4F2EF0,
	nil,
}

type questFieldGuideAllowedHooks4F2EF0 struct {
	loadRewardGuideID func(index int) uint32
	loadRewardSlots   func(index int) uint32
	loadFamily        func(index int) *questFieldGuideFamily4F2EF0
	loadFamilyValue   func(family *questFieldGuideFamily4F2EF0, index int) uint32
}

// questFieldGuideAllowed4F2EF0 reconstructs GAME.EXE 004F2EF0. It first
// scans the live reward-guide rows and reads Slots only for a matching ID.
// It then scans every non-null family even after an earlier match. A family
// header is only a nonzero gate; matching begins at member one.
//
// The member comparison intentionally occurs before the zero-sentinel check.
// Consequently guide zero matches the terminator of any nonempty family.
func questFieldGuideAllowed4F2EF0(
	guideID int32,
	hooks questFieldGuideAllowedHooks4F2EF0,
) int32 {
	target := uint32(guideID)
	index := 0
	currentGuideID := hooks.loadRewardGuideID(index)
	allowed := false
	if currentGuideID != 0 {
		for {
			if currentGuideID == target && hooks.loadRewardSlots(index) != 0 {
				allowed = true
				break
			}
			index++
			currentGuideID = hooks.loadRewardGuideID(index)
			if currentGuideID == 0 {
				break
			}
		}
	}

	familyIndex := 0
	family := hooks.loadFamily(familyIndex)
	if family != nil {
		for {
			header := hooks.loadFamilyValue(family, 0)
			if header != 0 {
				memberIndex := 1
				for {
					member := hooks.loadFamilyValue(family, memberIndex)
					if member == target {
						allowed = true
						break
					}
					memberIndex++
					if member == 0 {
						break
					}
				}
			}
			familyIndex++
			family = hooks.loadFamily(familyIndex)
			if family == nil {
				break
			}
		}
	}
	if allowed {
		return 1
	}
	return 0
}

// QuestFieldGuideAllowed4F2EF0 binds the original predicate to the exact
// fixed-width reward rows and the native-width guide-family pointer table.
//
//go:noinline
func QuestFieldGuideAllowed4F2EF0(guideID int32) int32 {
	return questFieldGuideAllowed4F2EF0(guideID, questFieldGuideAllowedHooks4F2EF0{
		loadRewardGuideID: func(index int) uint32 {
			return rewardFieldGuideDefinitions4F0D20[index].GuideID
		},
		loadRewardSlots: func(index int) uint32 {
			return rewardFieldGuideDefinitions4F0D20[index].Slots
		},
		loadFamily: func(index int) *questFieldGuideFamily4F2EF0 {
			return questFieldGuideFamilies4F2EF0[index]
		},
		loadFamilyValue: func(family *questFieldGuideFamily4F2EF0, index int) uint32 {
			return family[index]
		},
	})
}
