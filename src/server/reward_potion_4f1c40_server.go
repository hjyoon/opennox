package server

type rewardPotionNativeDeps4F1C40 struct {
	objects           []rewardObjectDefinition4F0640
	pickSlots         func(uint32) uint32
	randomInt         func(int32, int32) int32
	objectTypeAllowed func(uint32) bool
	createObject      func(uint32) *Object
}

func rewardPotionNative4F1C40(stage uint32, deps rewardPotionNativeDeps4F1C40) *Object {
	return rewardPotion4F1C40(stage, rewardPotionHooks4F1C40[*Object]{
		pickSlots: deps.pickSlots,
		loadObjectName: func(index int) string {
			return deps.objects[index].Name
		},
		loadObjectWeight: func(index int) uint8 {
			return uint8(deps.objects[index].Weight)
		},
		loadObjectType: func(index int) uint32 {
			return deps.objects[index].TypeInd
		},
		loadObjectKind: func(index int) uint8 {
			return uint8(deps.objects[index].Kind)
		},
		loadObjectSlots: func(index int) uint32 {
			return deps.objects[index].Slots
		},
		objectTypeAllowed: deps.objectTypeAllowed,
		randomInt:         deps.randomInt,
		createObject:      deps.createObject,
	})
}

// RewardPotion4F1C40 binds GAME.EXE 004F1C40 to the per-server reward
// definitions, logic RNG, native object registry, and native object factory.
// The marker is intentionally ignored because the original function never
// reads its first argument. Object references are never converted to PE32
// integer cells.
//
//go:noinline
func (s *Server) RewardPotion4F1C40(_ *Object, stage uint32) *Object {
	randomInt := func(minimum, maximum int32) int32 {
		return int32(s.Rand.Logic.IntClamp(int(minimum), int(maximum)))
	}
	return rewardPotionNative4F1C40(stage, rewardPotionNativeDeps4F1C40{
		objects: s.rewardDefinitions.Objects[:],
		pickSlots: func(stage uint32) uint32 {
			return rewardRandomSlots4F0B60(stage, randomInt)
		},
		randomInt: randomInt,
		objectTypeAllowed: func(typeInd uint32) bool {
			return s.Types.ByInd(int(typeInd)).Allowed()
		},
		createObject: func(typeInd uint32) *Object {
			return s.NewObjectByTypeInd(int(typeInd))
		},
	})
}
