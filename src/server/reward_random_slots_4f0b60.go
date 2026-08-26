package server

// rewardSlotWeightsHalf4F0B60 returns GAME.EXE 004F0B60's five binary32
// weights multiplied by two. Every original value is an exact multiple of
// 0.5, so half units preserve the x87 comparison boundaries without making
// the result depend on a target's floating-point evaluation rules.
func rewardSlotWeightsHalf4F0B60(stage uint32) ([5]uint16, bool) {
	var weights [5]uint16
	switch stage {
	case 0, 10:
		return weights, false
	case 1:
		weights[0], weights[1] = 175, 25
		return weights, true
	case 9:
		weights[3], weights[4] = 25, 175
		return weights, true
	}
	if stage > 10 {
		return weights, false
	}
	index := stage >> 1
	if stage&1 != 0 {
		weights[index-1] = 25
		weights[index] = 150
		weights[index+1] = 25
	} else {
		weights[index-1] = 100
		weights[index] = 100
	}
	return weights, true
}

// rewardRandomSlots4F0B60 reconstructs GAME.EXE 004F0B60. Stage zero and
// ten return their fixed endpoint slots and every unsigned stage above ten
// returns the last slot. Stages one through nine draw the inclusive logic RNG
// range 0..200 exactly once. The original multiplies that integer by binary32
// 0.5, spills it, and compares it inclusively with cumulative x87 weights.
// Comparing the unscaled integer with cumulative half units is exact for all
// int32 callback results, including the observable out-of-contract fallback.
func rewardRandomSlots4F0B60(
	stage uint32,
	randomInt func(int32, int32) int32,
) uint32 {
	if stage == 0 {
		return 1
	}
	if stage >= 10 {
		return 16
	}
	weights, _ := rewardSlotWeightsHalf4F0B60(stage)
	draw := randomInt(0, 200)
	var cumulative int32
	for index, weight := range weights {
		cumulative += int32(weight)
		if draw <= cumulative {
			return uint32(1) << index
		}
	}
	return 1
}
