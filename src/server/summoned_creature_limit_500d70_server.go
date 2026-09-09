package server

func summonedCreatureLimitNative500D70(
	owner *Object,
	guideIndex int32,
	loadGuideSize func(int32) int32,
	isMonitored func(owner, unit *Object) bool,
) bool {
	return summonedCreatureLimitCheck500D70(owner, guideIndex, summonedCreatureLimitHooks500D70[*Object]{
		loadGuideSize: loadGuideSize,
		countControlled: func(owner *Object) int32 {
			return controlledCreatureCountNative500D10(owner, isMonitored)
		},
	})
}

// CheckSummonedCreaturesLimit500D70 binds the exact GAME.EXE 00500D70
// arithmetic and call order to native-width Object links.
func CheckSummonedCreaturesLimit500D70(
	owner *Object,
	guideIndex int32,
	loadGuideSize func(int32) int32,
) bool {
	return summonedCreatureLimitNative500D70(
		owner,
		guideIndex,
		loadGuideSize,
		Nox_xxx_creatureIsMonitored_500CC0,
	)
}
