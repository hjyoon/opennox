package server

import noxflags "github.com/opennox/opennox/v1/common/flags"

type unitIsHostileMimicNativeDeps4E7F90 struct {
	loadMimicCache  func() uint32
	lookupType      func(string) uint32
	storeMimicCache func(uint32)
	isEnemy         func(*Object, *Object) int32
	isQuest         func() int32
}

func unitIsHostileMimicNative4E7F90(
	obj, obj2 *Object,
	deps unitIsHostileMimicNativeDeps4E7F90,
) int32 {
	return unitIsHostileMimic4E7F90(obj, obj2, unitIsHostileMimicHooks4E7F90[*Object]{
		loadMimicCache:  deps.loadMimicCache,
		lookupType:      deps.lookupType,
		storeMimicCache: deps.storeMimicCache,
		isEnemy:         deps.isEnemy,
		isQuest:         deps.isQuest,
		loadType: func(obj *Object) uint16 {
			return obj.TypeInd
		},
		loadClassLow: func(obj *Object) uint8 {
			return uint8(obj.ObjClass)
		},
		loadOwner: func(obj *Object) *Object {
			return obj.ObjOwner
		},
	})
}

func (s *Server) isHostileMimicResult4E7F90(obj, obj2 *Object) int32 {
	return unitIsHostileMimicNative4E7F90(obj, obj2, unitIsHostileMimicNativeDeps4E7F90{
		loadMimicCache: func() uint32 {
			return uint32(s.Types.fast.mimic)
		},
		lookupType: func(name string) uint32 {
			return uint32(s.Types.IndByID(name))
		},
		storeMimicCache: func(value uint32) {
			s.Types.fast.mimic = int(value)
		},
		isEnemy: func(obj, obj2 *Object) int32 {
			if s.IsEnemyTo(obj, obj2) {
				return 1
			}
			return 0
		},
		isQuest: func() int32 {
			if noxflags.HasGame(noxflags.GameModeQuest) {
				return 1
			}
			return 0
		},
	})
}

func (s *Server) IsHostileMimicXxx(obj, obj2 *Object) bool { // nox_xxx_unitIsHostileMimic_4E7F90
	return s.isHostileMimicResult4E7F90(obj, obj2) == 1
}
