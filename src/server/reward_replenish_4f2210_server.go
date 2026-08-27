package server

import "unsafe"

// RewardReplenishRuntime4F2210 supplies the two legacy game-state reads and
// delayed deletion. Object traversal, type lookup, temporary arrays, InitData,
// and random selection stay on the native Server side.
type RewardReplenishRuntime4F2210 struct {
	QuestStage    func() uint32
	PlayerCount   func() int32
	DelayedDelete func(*Object)
}

type rewardReplenishObjectArray4F2210 struct {
	objects []*Object
}

type rewardReplenishNativeDeps4F2210 struct {
	loadMarkerCache  func() uint32
	storeMarkerCache func(uint32)
	loadPlusCache    func() uint32
	storePlusCache   func(uint32)
	loadPotionCache  func() uint32
	storePotionCache func(uint32)
	lookupType       func(string) uint32
	firstObject      func() *Object
	randomInt        func(int32, int32, string, int32) int32
	runtime          RewardReplenishRuntime4F2210
}

func rewardReplenishNative4F2210(deps rewardReplenishNativeDeps4F2210) {
	rewardReplenish4F2210(rewardReplenishHooks4F2210[
		*Object,
		*RewardMarkerInitData,
		*rewardReplenishObjectArray4F2210,
	]{
		loadQuestStage:   deps.runtime.QuestStage,
		loadPlayerCount:  deps.runtime.PlayerCount,
		loadMarkerCache:  deps.loadMarkerCache,
		storeMarkerCache: deps.storeMarkerCache,
		loadPlusCache:    deps.loadPlusCache,
		storePlusCache:   deps.storePlusCache,
		loadPotionCache:  deps.loadPotionCache,
		storePotionCache: deps.storePotionCache,
		lookupType:       deps.lookupType,
		firstObject:      deps.firstObject,
		nextObject: func(object *Object) *Object {
			return object.ObjNext
		},
		loadTypeInd: func(object *Object) uint16 {
			return object.TypeInd
		},
		loadInitData: func(object *Object) *RewardMarkerInitData {
			return (*RewardMarkerInitData)(object.InitData)
		},
		loadFieldLow: func(data *RewardMarkerInitData) uint8 {
			return uint8(data.Field216)
		},
		storeFieldLow: func(data *RewardMarkerInitData, value uint8) {
			data.Field216 = data.Field216&^uint32(0xff) | uint32(value)
		},
		allocObjects: func(count int32) *rewardReplenishObjectArray4F2210 {
			return &rewardReplenishObjectArray4F2210{
				objects: make([]*Object, int(count)),
			}
		},
		storeObject: func(array *rewardReplenishObjectArray4F2210, index int32, object *Object) {
			array.objects[index] = object
		},
		loadObject: func(array *rewardReplenishObjectArray4F2210, index int32) *Object {
			return array.objects[index]
		},
		freeObjects: func(array *rewardReplenishObjectArray4F2210) {
			array.objects = nil
		},
		randomInt:     deps.randomInt,
		delayedDelete: deps.runtime.DelayedDelete,
	})
}

// RewardReplenish4F2210 binds GAME.EXE 004F2210 to native Object links and
// InitData, dedicated uint32 type caches, and the server logic RNG. No object
// pointer is encoded into the original PE32 integer representation.
//
//go:noinline
func (s *Server) RewardReplenish4F2210(runtime RewardReplenishRuntime4F2210) {
	rewardReplenishNative4F2210(rewardReplenishNativeDeps4F2210{
		loadMarkerCache: func() uint32 {
			return s.Types.fast.rewardReplenishMarker
		},
		storeMarkerCache: func(value uint32) {
			s.Types.fast.rewardReplenishMarker = value
		},
		loadPlusCache: func() uint32 {
			return s.Types.fast.rewardReplenishMarkerPlus
		},
		storePlusCache: func(value uint32) {
			s.Types.fast.rewardReplenishMarkerPlus = value
		},
		loadPotionCache: func() uint32 {
			return s.Types.fast.rewardReplenishRedPotion
		},
		storePotionCache: func(value uint32) {
			s.Types.fast.rewardReplenishRedPotion = value
		},
		lookupType: func(name string) uint32 {
			return uint32(s.Types.IndByID(name))
		},
		firstObject: s.Objs.First,
		randomInt: func(minimum, maximum int32, _ string, _ int32) int32 {
			return int32(s.Rand.Logic.IntClamp(int(minimum), int(maximum)))
		},
		runtime: runtime,
	})
}

var (
	_ = [1]struct{}{}[220-unsafe.Sizeof(RewardMarkerInitData{})]
	_ = [1]struct{}{}[216-unsafe.Offsetof(RewardMarkerInitData{}.Field216)]
)
