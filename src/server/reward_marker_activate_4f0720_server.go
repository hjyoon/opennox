package server

import "unsafe"

// RewardMarkerActivateRuntime4F0720 supplies the eight reward creators that
// follow 004F0720. Each remains an independently tracked restoration unit.
// Stage is forwarded to every callback because the original call site pushes
// it even for AbilityBook, whose current callee ignores that extra argument.
// Object pointers remain native-width across this boundary.
type RewardMarkerActivateRuntime4F0720 struct {
	SpellBook   func(*Object, uint32) *Object
	AbilityBook func(*Object, uint32) *Object
	FieldGuide  func(*Object, uint32) *Object
	Weapon      func(*Object, uint32) *Object
	Armor       func(*Object, uint32) *Object
	Gem         func(*Object, uint32) *Object
	Potion      func(*Object, uint32) *Object
	Gem2        func(*Object, uint32) *Object
}

type rewardMarkerActivateNativeDeps4F0720 struct {
	loadCachedPlusType  func() uint32
	lookupType          func(string) uint32
	storeCachedPlusType func(uint32)
	randomInt           func(int32, int32) int32
	runtime             RewardMarkerActivateRuntime4F0720
}

func rewardMarkerActivateNative4F0720(
	marker *Object,
	stage uint32,
	deps rewardMarkerActivateNativeDeps4F0720,
) *Object {
	return rewardMarkerActivate4F0720(marker, stage, rewardMarkerActivateHooks4F0720[
		*Object,
		*RewardMarkerInitData,
		*Object,
	]{
		loadCachedPlusType: deps.loadCachedPlusType,
		loadInitData: func(marker *Object) *RewardMarkerInitData {
			return (*RewardMarkerInitData)(marker.InitData)
		},
		lookupType:          deps.lookupType,
		storeCachedPlusType: deps.storeCachedPlusType,
		loadTypeInd: func(marker *Object) uint16 {
			return marker.TypeInd
		},
		loadChanceMode: func(data *RewardMarkerInitData) uint32 {
			return data.ChanceMode
		},
		randomInt: deps.randomInt,
		loadCategoryMask: func(data *RewardMarkerInitData) uint32 {
			return data.CategoryMask
		},
		dispatch: func(kind rewardMarkerDispatch4F0720, marker *Object, stage uint32) *Object {
			switch kind {
			case rewardMarkerDispatchSpellBook4F0720:
				return deps.runtime.SpellBook(marker, stage)
			case rewardMarkerDispatchAbilityBook4F0720:
				return deps.runtime.AbilityBook(marker, stage)
			case rewardMarkerDispatchFieldGuide4F0720:
				return deps.runtime.FieldGuide(marker, stage)
			case rewardMarkerDispatchWeapon4F0720:
				return deps.runtime.Weapon(marker, stage)
			case rewardMarkerDispatchArmor4F0720:
				return deps.runtime.Armor(marker, stage)
			case rewardMarkerDispatchGem4F0720, rewardMarkerDispatchDefaultGem4F0720:
				return deps.runtime.Gem(marker, stage)
			case rewardMarkerDispatchPotion4F0720:
				return deps.runtime.Potion(marker, stage)
			case rewardMarkerDispatchGem2_4F0720:
				return deps.runtime.Gem2(marker, stage)
			default:
				panic("unreachable reward-marker dispatch index")
			}
		},
	})
}

// RewardMarkerActivate4F0720 binds GAME.EXE 004F0720 to native Object and
// fixed-width RewardMarkerInitData fields. The dedicated type cache stays a
// uint32 game value; marker and created-object references stay native pointers.
// There are deliberately no nil or callback guards.
//
//go:noinline
func (s *Server) RewardMarkerActivate4F0720(
	marker *Object,
	stage uint32,
	runtime RewardMarkerActivateRuntime4F0720,
) *Object {
	return rewardMarkerActivateNative4F0720(marker, stage, rewardMarkerActivateNativeDeps4F0720{
		loadCachedPlusType: func() uint32 {
			return s.Types.fast.rewardMarkerPlus
		},
		lookupType: func(name string) uint32 {
			return uint32(s.Types.IndByID(name))
		},
		storeCachedPlusType: func(value uint32) {
			s.Types.fast.rewardMarkerPlus = value
		},
		randomInt: func(minimum, maximum int32) int32 {
			return int32(s.Rand.Logic.IntClamp(int(minimum), int(maximum)))
		},
		runtime: runtime,
	})
}

var (
	_ = [1]struct{}{}[220-unsafe.Sizeof(RewardMarkerInitData{})]
	_ = [1]struct{}{}[0-unsafe.Offsetof(RewardMarkerInitData{}.CategoryMask)]
	_ = [1]struct{}{}[212-unsafe.Offsetof(RewardMarkerInitData{}.ChanceMode)]
	_ = [1]struct{}{}[216-unsafe.Offsetof(RewardMarkerInitData{}.Field216)]
)
