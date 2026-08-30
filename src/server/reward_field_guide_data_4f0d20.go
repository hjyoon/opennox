package server

// rewardFieldGuideDefinition4F0D20 is the native fixed-width meaning of one
// original twelve-byte row at GAME.EXE 005B9BB0. The creator reads only the
// low byte of the original weight dword; guide IDs and slot masks remain exact
// uint32 game values.
type rewardFieldGuideDefinition4F0D20 struct {
	Weight  uint8
	GuideID uint32
	Slots   uint32
}

// rewardFieldGuideDefinitions4F0D20 contains the 31 weighted rows and the
// first zero-ID sentinel consumed by reward field-guide creation.
var rewardFieldGuideDefinitions4F0D20 = [...]rewardFieldGuideDefinition4F0D20{
	{Weight: 4, GuideID: 32, Slots: 0x1f},
	{Weight: 4, GuideID: 34, Slots: 0x1f},
	{Weight: 4, GuideID: 35, Slots: 0x1f},
	{Weight: 4, GuideID: 36, Slots: 0x1f},
	{Weight: 4, GuideID: 17, Slots: 0x1f},
	{Weight: 4, GuideID: 22, Slots: 0x1f},
	{Weight: 4, GuideID: 26, Slots: 0x1f},
	{Weight: 4, GuideID: 24, Slots: 0x1f},
	{Weight: 4, GuideID: 2, Slots: 0x1e},
	{Weight: 4, GuideID: 3, Slots: 0x1e},
	{Weight: 4, GuideID: 33, Slots: 0x1e},
	{Weight: 4, GuideID: 20, Slots: 0x1e},
	{Weight: 4, GuideID: 23, Slots: 0x1e},
	{Weight: 4, GuideID: 11, Slots: 0x1e},
	{Weight: 4, GuideID: 18, Slots: 0x1e},
	{Weight: 4, GuideID: 19, Slots: 0x1e},
	{Weight: 4, GuideID: 29, Slots: 0x1c},
	{Weight: 4, GuideID: 14, Slots: 0x1c},
	{Weight: 4, GuideID: 13, Slots: 0x1c},
	{Weight: 4, GuideID: 21, Slots: 0x1c},
	{Weight: 4, GuideID: 4, Slots: 0x1c},
	{Weight: 4, GuideID: 9, Slots: 0x1c},
	{Weight: 4, GuideID: 16, Slots: 0x18},
	{Weight: 4, GuideID: 10, Slots: 0x18},
	{Weight: 4, GuideID: 27, Slots: 0x18},
	{Weight: 4, GuideID: 39, Slots: 0x18},
	{Weight: 4, GuideID: 40, Slots: 0x18},
	{Weight: 2, GuideID: 15, Slots: 0x10},
	{Weight: 2, GuideID: 31, Slots: 0x10},
	{Weight: 1, GuideID: 37, Slots: 0x10},
	{Weight: 1, GuideID: 38, Slots: 0x10},
	{Slots: 0x1f},
}

// rewardFieldGuideNames4F0D20 is the native form of the unchecked 41-entry
// pointer table at GAME.EXE 00598364 and its referenced creature strings.
var rewardFieldGuideNames4F0D20 = [...]string{
	"GUIDE_INVALID",
	"Bat",
	"BlackBear",
	"Bear",
	"Beholder",
	"Bomber",
	"CarnivorousPlant",
	"AlbinoSpider",
	"SmallAlbinoSpider",
	"EvilCherub",
	"EmberDemon",
	"Ghost",
	"GiantLeech",
	"Imp",
	"FlyingGolem",
	"MechanicalGolem",
	"Mimic",
	"GruntAxe",
	"OgreBrute",
	"OgreWarlord",
	"Scorpion",
	"Shade",
	"Skeleton",
	"SkeletonLord",
	"Spider",
	"SmallSpider",
	"SpittingSpider",
	"StoneGolem",
	"Troll",
	"Urchin",
	"Wasp",
	"WillOWisp",
	"Wolf",
	"BlackWolf",
	"WhiteWolf",
	"Zombie",
	"VileZombie",
	"Demon",
	"Lich",
	"WizardGreen",
	"UrchinShaman",
}

// RewardFieldGuideName4F0D20 returns the exact registered guide-table name.
// Out-of-range IDs retain the native table wrapper's nil/empty result.
func RewardFieldGuideName4F0D20(id int) string {
	if id < 0 || id >= len(rewardFieldGuideNames4F0D20) {
		return ""
	}
	return rewardFieldGuideNames4F0D20[id]
}

// RewardFieldGuideID4F0D20 resolves an exact registered guide-table name.
// Unknown names and GUIDE_INVALID both resolve to the original invalid ID 0.
func RewardFieldGuideID4F0D20(name string) int {
	for id, candidate := range rewardFieldGuideNames4F0D20 {
		if candidate == name {
			return id
		}
	}
	return 0
}
