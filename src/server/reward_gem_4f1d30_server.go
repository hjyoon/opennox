package server

import "unsafe"

type rewardGemNativeDeps4F1D30 struct {
	pickSlots          func(uint32) uint32
	randomInt          func(int32, int32, string, int32) int32
	createObjectByType func(string) *Object
}

func rewardGemNative4F1D30(stage uint32, deps rewardGemNativeDeps4F1D30) *Object {
	return rewardGem4F1D30(stage, rewardGemHooks4F1D30[*Object, *GoldInitData]{
		pickSlots:          deps.pickSlots,
		randomInt:          deps.randomInt,
		createObjectByType: deps.createObjectByType,
		isNilObject: func(object *Object) bool {
			return object == nil
		},
		loadInitData: func(object *Object) *GoldInitData {
			return (*GoldInitData)(object.InitData)
		},
		loadGoldMaximum: func(index rewardGemGoldRange4F1D30) int32 {
			return rewardGemGoldAmountBounds4F1D30[index].Maximum
		},
		loadGoldMinimum: func(index rewardGemGoldRange4F1D30) int32 {
			return rewardGemGoldAmountBounds4F1D30[index].Minimum
		},
		storeGoldAmount: func(data *GoldInitData, amount int32) {
			data.Amount = uint32(amount)
		},
	})
}

// RewardGem4F1D30 binds GAME.EXE 004F1D30 to native Object and GoldInitData
// pointers, the exact immutable gold-amount bounds, the server logic RNG, and
// the native name-based object factory. The marker is intentionally ignored
// because the original function never reads its first argument.
//
//go:noinline
func (s *Server) RewardGem4F1D30(_ *Object, stage uint32) *Object {
	randomInt := func(minimum, maximum int32, _ string, _ int32) int32 {
		return int32(s.Rand.Logic.IntClamp(int(minimum), int(maximum)))
	}
	return rewardGemNative4F1D30(stage, rewardGemNativeDeps4F1D30{
		pickSlots: func(stage uint32) uint32 {
			return rewardRandomSlots4F0B60(stage, func(minimum, maximum int32) int32 {
				return randomInt(minimum, maximum, "", 0)
			})
		},
		randomInt:          randomInt,
		createObjectByType: s.NewObjectByTypeID,
	})
}

// RewardGem2_4F1F00 preserves 004F1F00's exact two-argument forwarding
// wrapper while keeping both marker dispatch destinations visible.
//
//go:noinline
func (s *Server) RewardGem2_4F1F00(marker *Object, stage uint32) *Object {
	return s.RewardGem4F1D30(marker, stage)
}

var (
	_ = [1]struct{}{}[4-unsafe.Sizeof(GoldInitData{})]
	_ = [1]struct{}{}[0-unsafe.Offsetof(GoldInitData{}.Amount)]
)
