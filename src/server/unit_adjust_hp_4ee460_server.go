package server

import noxflags "github.com/opennox/opennox/v1/common/flags"

// UnitAdjustHPRuntime4EE460 supplies the still-legacy 004E4560 HP setter.
// All pointer-bearing reads in 004EE460/004EE4C0 and current-HP reporting are
// native Go; the setter remains an explicit dependency for its own audit.
type UnitAdjustHPRuntime4EE460 struct {
	SetHP func(*Object, uint16)
}

type unitAdjustHPNativeDeps4EE460 struct {
	gameFlag      func(uint32) int32
	setHP         func(*Object, uint16)
	informOwnerHP func(*Object)
}

func unitAdjustHPNative4EE460(
	unit *Object,
	delta int32,
	deps unitAdjustHPNativeDeps4EE460,
) {
	unitAdjustHP4EE460(unitAdjustHPHooks4EE460[*Object, *HealthData]{
		gameFlag: deps.gameFlag,
		loadUnitArg: func() *Object {
			return unit
		},
		loadHealth: func(unit *Object) *HealthData {
			return unit.HealthData
		},
		loadCurrent: func(health *HealthData) uint16 {
			return health.Cur
		},
		loadMaximum: func(health *HealthData) uint16 {
			return health.Max
		},
		loadDeltaArg: func() int32 {
			return delta
		},
		setHP: deps.setHP,
		loadClassLow: func(unit *Object) uint8 {
			return uint8(unit.ObjClass)
		},
		informOwnerHP: deps.informOwnerHP,
	})
}

func mobInformOwnerHPNative4EE4C0(obj *Object, report func(uint8, *Object)) {
	mobInformOwnerHP4EE4C0(mobInformOwnerHPHooks4EE4C0[*Object, *PlayerUpdateData, *Player]{
		loadObjectArg: func() *Object {
			return obj
		},
		loadOwner: func(obj *Object) *Object {
			return obj.ObjOwner
		},
		loadClassLow: func(owner *Object) uint8 {
			return uint8(owner.ObjClass)
		},
		loadUpdateData: func(owner *Object) *PlayerUpdateData {
			return (*PlayerUpdateData)(owner.UpdateData)
		},
		loadPlayer: func(update *PlayerUpdateData) *Player {
			return update.Player
		},
		loadPlayerInd: func(player *Player) uint8 {
			return player.PlayerInd
		},
		reportHP: report,
	})
}

// MobInformOwnerHP4EE4C0 resolves the native owner/Player chain and invokes
// the restored current-HP reporter with the exact PlayerInd byte.
func (s *Server) MobInformOwnerHP4EE4C0(obj *Object) {
	mobInformOwnerHPNative4EE4C0(obj, func(playerInd uint8, obj *Object) {
		_ = s.CurrentHPReport4D8620(int32(playerInd), obj)
	})
}

// UnitAdjustHP4EE460 applies the original wrapped adjustment and live class
// notification using native-width objects. The runtime setter is the sole
// remaining dependency on the separately scoped 004E4560 cluster.
func (s *Server) UnitAdjustHP4EE460(unit *Object, delta int32, runtime UnitAdjustHPRuntime4EE460) {
	unitAdjustHPNative4EE460(unit, delta, unitAdjustHPNativeDeps4EE460{
		gameFlag: func(flag uint32) int32 {
			if noxflags.HasGame(noxflags.GameFlag(flag)) {
				return 1
			}
			return 0
		},
		setHP: runtime.SetHP,
		informOwnerHP: func(unit *Object) {
			s.MobInformOwnerHP4EE4C0(unit)
		},
	})
}
