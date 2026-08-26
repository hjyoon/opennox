package server

import "unsafe"

type rewardFieldGuideNativeDeps4F0D20 struct {
	rows               []rewardFieldGuideDefinition4F0D20
	pickSlots          func(uint32) uint32
	randomInt          func(int32, int32) int32
	createObjectByType func(string) *Object
}

func rewardFieldGuideNative4F0D20(
	marker *Object,
	stage uint32,
	deps rewardFieldGuideNativeDeps4F0D20,
) *Object {
	return rewardFieldGuide4F0D20(marker, stage, rewardFieldGuideHooks4F0D20[
		*Object,
		*RewardMarkerInitData,
		[]rewardFieldGuideDefinition4F0D20,
		*Object,
		*FieldGuideUseData,
	]{
		loadInitData: func(marker *Object) *RewardMarkerInitData {
			return (*RewardMarkerInitData)(marker.InitData)
		},
		loadFlags: func(data *RewardMarkerInitData) uint8 {
			return data.RewardFlags
		},
		loadExplicitGuide: func(data *RewardMarkerInitData, index int) uint8 {
			return data.Guides[index]
		},
		pickSlots: deps.pickSlots,
		rows:      deps.rows,
		loadRowWeight: func(rows []rewardFieldGuideDefinition4F0D20, index int) uint8 {
			return rows[index].Weight
		},
		loadRowGuideID: func(rows []rewardFieldGuideDefinition4F0D20, index int) uint32 {
			return rows[index].GuideID
		},
		loadRowSlots: func(rows []rewardFieldGuideDefinition4F0D20, index int) uint32 {
			return rows[index].Slots
		},
		randomInt: deps.randomInt,
		createObjectByType: func(typeName string) *Object {
			return deps.createObjectByType(typeName)
		},
		isNilObject: func(object *Object) bool {
			return object == nil
		},
		loadUseData: func(object *Object) *FieldGuideUseData {
			return object.UseDataFieldGuide()
		},
		guideName: func(guideID uint32) string {
			return rewardFieldGuideNames4F0D20[guideID]
		},
		storeGuide: func(useData *FieldGuideUseData, name string) {
			useData.SetCreature(name)
		},
	})
}

// RewardFieldGuide4F0D20 binds GAME.EXE 004F0D20 to native Object pointers,
// fixed-width RewardMarkerInitData and FieldGuideUseData, the exact native
// guide table, the server logic RNG, and the native object factory. There are
// deliberately no nil, row, or guide-ID guards beyond original branches.
//
//go:noinline
func (s *Server) RewardFieldGuide4F0D20(marker *Object, stage uint32) *Object {
	randomInt := func(minimum, maximum int32) int32 {
		return int32(s.Rand.Logic.IntClamp(int(minimum), int(maximum)))
	}
	return rewardFieldGuideNative4F0D20(marker, stage, rewardFieldGuideNativeDeps4F0D20{
		rows: rewardFieldGuideDefinitions4F0D20[:],
		pickSlots: func(stage uint32) uint32 {
			return rewardRandomSlots4F0B60(stage, randomInt)
		},
		randomInt:          randomInt,
		createObjectByType: s.NewObjectByTypeID,
	})
}

var (
	_ = [1]struct{}{}[220-unsafe.Sizeof(RewardMarkerInitData{})]
	_ = [1]struct{}{}[4-unsafe.Offsetof(RewardMarkerInitData{}.RewardFlags)]
	_ = [1]struct{}{}[151-unsafe.Offsetof(RewardMarkerInitData{}.Guides)]
	_ = [1]struct{}{}[212-unsafe.Offsetof(RewardMarkerInitData{}.ChanceMode)]
	_ = [1]struct{}{}[64-unsafe.Sizeof(FieldGuideUseData{})]
	_ = [1]struct{}{}[0-unsafe.Offsetof(FieldGuideUseData{}.CreatureBuf)]
)
