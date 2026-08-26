package server

import "unsafe"

type rewardAbilityBookNativeDeps4F0C70 struct {
	randomInt          func(int32, int32) int32
	createObjectByType func(string) *Object
}

func rewardAbilityBookNative4F0C70(
	marker *Object,
	deps rewardAbilityBookNativeDeps4F0C70,
) *Object {
	return rewardAbilityBook4F0C70(marker, rewardAbilityBookHooks4F0C70[
		*Object,
		*RewardMarkerInitData,
		*Object,
	]{
		loadInitData: func(marker *Object) *RewardMarkerInitData {
			return (*RewardMarkerInitData)(marker.InitData)
		},
		loadFlags: func(data *RewardMarkerInitData) uint8 {
			return data.RewardFlags
		},
		loadExplicitAbility: func(data *RewardMarkerInitData, index int) uint8 {
			return data.Abilities[index]
		},
		randomInt: deps.randomInt,
		createObjectByType: func(typeName string) *Object {
			return deps.createObjectByType(typeName)
		},
		isNilObject: func(object *Object) bool {
			return object == nil
		},
		storeAbility: func(object *Object, ability uint8) {
			object.UseDataAbilityReward().Ability = ability
		},
	})
}

// RewardAbilityBook4F0C70 binds GAME.EXE 004F0C70 to native Object pointers,
// fixed-width RewardMarkerInitData and AbilityRewardUseData, the server logic
// RNG, and the native object factory. There are deliberately no nil guards.
//
//go:noinline
func (s *Server) RewardAbilityBook4F0C70(marker *Object) *Object {
	return rewardAbilityBookNative4F0C70(marker, rewardAbilityBookNativeDeps4F0C70{
		randomInt: func(minimum, maximum int32) int32 {
			return int32(s.Rand.Logic.IntClamp(int(minimum), int(maximum)))
		},
		createObjectByType: s.NewObjectByTypeID,
	})
}

var (
	_ = [1]struct{}{}[220-unsafe.Sizeof(RewardMarkerInitData{})]
	_ = [1]struct{}{}[4-unsafe.Offsetof(RewardMarkerInitData{}.RewardFlags)]
	_ = [1]struct{}{}[145-unsafe.Offsetof(RewardMarkerInitData{}.Abilities)]
	_ = [1]struct{}{}[1-unsafe.Sizeof(AbilityRewardUseData{})]
	_ = [1]struct{}{}[0-unsafe.Offsetof(AbilityRewardUseData{}.Ability)]
)
