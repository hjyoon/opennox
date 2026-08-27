package server

import (
	"unsafe"

	"github.com/opennox/libs/types"
)

// RewardAnkhReplaceRuntime4F2110 supplies the outer object creation and
// delayed-delete services. Every object reference remains a native pointer.
type RewardAnkhReplaceRuntime4F2110 struct {
	CreateAt      func(*Object, *Object, types.Pointf)
	DelayedDelete func(*Object)
}

type rewardAnkhReplaceNativeDeps4F2110 struct {
	loadMarkerCache  func() uint32
	storeMarkerCache func(uint32)
	loadPlusCache    func() uint32
	storePlusCache   func(uint32)
	lookupType       func(string) uint32
	firstObject      func() *Object
	randomInt        func(int32, int32, string, int32) int32
	newObject        func(string) *Object
	runtime          RewardAnkhReplaceRuntime4F2110
}

func rewardAnkhReplaceNative4F2110(deps rewardAnkhReplaceNativeDeps4F2110) {
	rewardAnkhReplace4F2110(rewardAnkhReplaceHooks4F2110[*Object, *RewardMarkerInitData]{
		loadMarkerCache:  deps.loadMarkerCache,
		storeMarkerCache: deps.storeMarkerCache,
		loadPlusCache:    deps.loadPlusCache,
		storePlusCache:   deps.storePlusCache,
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
		loadCategoryLow: func(data *RewardMarkerInitData) uint8 {
			return uint8(data.CategoryMask)
		},
		randomInt: deps.randomInt,
		newObject: deps.newObject,
		loadPosX: func(object *Object) float32 {
			return object.PosVec.X
		},
		loadPosY: func(object *Object) float32 {
			return object.PosVec.Y
		},
		createAt:      deps.runtime.CreateAt,
		delayedDelete: deps.runtime.DelayedDelete,
	})
}

// RewardAnkhReplace4F2110 binds GAME.EXE 004F2110 to native Object links and
// InitData, dedicated uint32 caches, the server logic RNG, type registry, and
// object factory. There are deliberately no nil-data or callback guards.
//
//go:noinline
func (s *Server) RewardAnkhReplace4F2110(runtime RewardAnkhReplaceRuntime4F2110) {
	rewardAnkhReplaceNative4F2110(rewardAnkhReplaceNativeDeps4F2110{
		loadMarkerCache: func() uint32 {
			return s.Types.fast.rewardAnkhMarker
		},
		storeMarkerCache: func(value uint32) {
			s.Types.fast.rewardAnkhMarker = value
		},
		loadPlusCache: func() uint32 {
			return s.Types.fast.rewardAnkhMarkerPlus
		},
		storePlusCache: func(value uint32) {
			s.Types.fast.rewardAnkhMarkerPlus = value
		},
		lookupType: func(name string) uint32 {
			return uint32(s.Types.IndByID(name))
		},
		firstObject: s.Objs.First,
		randomInt: func(minimum, maximum int32, _ string, _ int32) int32 {
			return int32(s.Rand.Logic.IntClamp(int(minimum), int(maximum)))
		},
		newObject: s.NewObjectByTypeID,
		runtime:   runtime,
	})
}

var (
	_ = [1]struct{}{}[220-unsafe.Sizeof(RewardMarkerInitData{})]
	_ = [1]struct{}{}[0-unsafe.Offsetof(RewardMarkerInitData{}.CategoryMask)]
)
