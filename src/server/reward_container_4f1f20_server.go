package server

import (
	"unsafe"

	"github.com/opennox/libs/types"
)

// RewardContainerRuntime4F1F20 supplies services that still cross the outer
// server or later sequential restoration units. Object and point references
// remain native-width throughout the boundary.
type RewardContainerRuntime4F1F20 struct {
	QuestStage        func() uint32
	PreprocessMarkers func()
	PreprocessRewards func()
	ActivateMarker    func(*Object, uint32) *Object
	CreateAt          func(*Object, *Object, types.Pointf)
	DelayedDelete     func(*Object)
	DetachInventory   func(*Object, *Object)
	InventoryPut      func(*Object, *Object, uint32)
}

type rewardContainerNativeDeps4F1F20 struct {
	loadMarkerCache  func() uint32
	storeMarkerCache func(uint32)
	loadPlusCache    func() uint32
	storePlusCache   func(uint32)
	lookupType       func(string) uint32
	firstObject      func() *Object
	chestInit        unsafe.Pointer
	newObject        func(string) *Object
	randomReachable  func(float32, *types.Pointf, *types.Pointf) *types.Pointf
	runtime          RewardContainerRuntime4F1F20
}

func rewardContainerNative4F1F20(deps rewardContainerNativeDeps4F1F20) {
	rewardContainerProcess4F1F20(rewardContainerHooks4F1F20[*Object, *RewardMarkerInitData]{
		loadQuestStage:    deps.runtime.QuestStage,
		loadMarkerCache:   deps.loadMarkerCache,
		storeMarkerCache:  deps.storeMarkerCache,
		loadPlusCache:     deps.loadPlusCache,
		storePlusCache:    deps.storePlusCache,
		lookupType:        deps.lookupType,
		preprocessMarkers: deps.runtime.PreprocessMarkers,
		preprocessRewards: deps.runtime.PreprocessRewards,
		firstObject:       deps.firstObject,
		nextObject: func(object *Object) *Object {
			return object.ObjNext
		},
		loadTypeInd: func(object *Object) uint16 {
			return object.TypeInd
		},
		loadInit: func(object *Object) unsafe.Pointer {
			return object.Init
		},
		isChestInit: func(init unsafe.Pointer) bool {
			return init == deps.chestInit
		},
		loadInitData: func(object *Object) *RewardMarkerInitData {
			return (*RewardMarkerInitData)(object.InitData)
		},
		loadField216Low: func(data *RewardMarkerInitData) uint8 {
			return uint8(data.Field216)
		},
		firstInventory: func(object *Object) *Object {
			return object.InvFirstItem
		},
		nextInventoryItem: func(object *Object) *Object {
			return object.InvNextItem
		},
		activateMarker:  deps.runtime.ActivateMarker,
		detachInventory: deps.runtime.DetachInventory,
		delayedDelete:   deps.runtime.DelayedDelete,
		inventoryPut:    deps.runtime.InventoryPut,
		loadClass: func(object *Object) uint32 {
			return uint32(object.ObjClass)
		},
		loadSubclassLow: func(object *Object) uint8 {
			return uint8(object.ObjSubClass)
		},
		newObject: deps.newObject,
		loadPosX: func(object *Object) float32 {
			return object.PosVec.X
		},
		loadPosY: func(object *Object) float32 {
			return object.PosVec.Y
		},
		createAt: deps.runtime.CreateAt,
		randomReachable: func(radius float32, center *Object, output *types.Pointf) {
			deps.randomReachable(radius, &center.PosVec, output)
		},
	})
}

// RewardContainerProcess4F1F20 binds GAME.EXE 004F1F20 to native-width
// Object links, the server type registry, dedicated fixed-width caches, the
// native object factory, and the restored random reachable-point helper.
// Chest identity remains the exact registered ChestInit callback pointer.
//
//go:noinline
func (s *Server) RewardContainerProcess4F1F20(runtime RewardContainerRuntime4F1F20) {
	rewardContainerNative4F1F20(rewardContainerNativeDeps4F1F20{
		loadMarkerCache: func() uint32 {
			return s.Types.fast.rewardContainerMarker
		},
		storeMarkerCache: func(value uint32) {
			s.Types.fast.rewardContainerMarker = value
		},
		loadPlusCache: func() uint32 {
			return s.Types.fast.rewardContainerMarkerPlus
		},
		storePlusCache: func(value uint32) {
			s.Types.fast.rewardContainerMarkerPlus = value
		},
		lookupType: func(name string) uint32 {
			return uint32(s.Types.IndByID(name))
		},
		firstObject:     s.Objs.First,
		chestInit:       initFuncs["ChestInit"].Func,
		newObject:       s.NewObjectByTypeID,
		randomReachable: s.RandomReachablePointAroundInto4ED970,
		runtime:         runtime,
	})
}

var (
	_ = [1]struct{}{}[220-unsafe.Sizeof(RewardMarkerInitData{})]
	_ = [1]struct{}{}[216-unsafe.Offsetof(RewardMarkerInitData{}.Field216)]
)
