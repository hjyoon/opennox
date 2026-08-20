package server

import (
	"unsafe"

	"github.com/opennox/libs/object"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
)

// UnitDamageClearRuntime4EE5E0 supplies effects that live above package server
// or retain their own restoration scope. Every Object-bearing argument remains
// native-width; the fixed damage, HP, and enchant values keep their original
// widths at the boundary.
type UnitDamageClearRuntime4EE5E0 struct {
	BreakHarpoon  func(*Object)
	SetHP         func(*Object, uint16)
	BuffOff       func(*Object, EnchantID)
	SoloReward    func(*Object)
	MonsterDie    func(*Object)
	DelayedDelete func(*Object)
}

type unitDamageClearNativeDeps4EE5E0 struct {
	engineFlag    func(uint32) int32
	breakHarpoon  func(*Object)
	setHP         func(*Object, uint16)
	buffOff       func(*Object, EnchantID)
	isZombie      func(*Object) bool
	soloReward    func(*Object)
	monsterDie    func(*Object)
	callDeath     func(unsafe.Pointer, *Object)
	delayedDelete func(*Object)
	informOwnerHP func(*Object)
}

func unitDamageClearNative4EE5E0(
	unit *Object,
	damage int32,
	deps unitDamageClearNativeDeps4EE5E0,
) {
	unitDamageClear4EE5E0(unitDamageClearHooks4EE5E0[
		*Object,
		*HealthData,
		*PlayerUpdateData,
		*Player,
		unsafe.Pointer,
	]{
		loadUnitArg: func() *Object {
			return unit
		},
		loadHealth: func(unit *Object) *HealthData {
			return unit.HealthData
		},
		loadMaximum: func(health *HealthData) uint16 {
			return health.Max
		},
		engineFlag: deps.engineFlag,
		loadClassLow: func(unit *Object) uint8 {
			return uint8(unit.ObjClass)
		},
		loadUpdateData: func(unit *Object) *PlayerUpdateData {
			return (*PlayerUpdateData)(unit.UpdateData)
		},
		loadPlayer: func(update *PlayerUpdateData) *Player {
			return update.Player
		},
		loadPlayerClass: func(player *Player) uint8 {
			if player == nil {
				panic("GAME.EXE 004EE5E0 Player class load through nil Player")
			}
			return uint8(player.Info().PlayerClass())
		},
		loadHarpoonTarget: func(update *PlayerUpdateData) *Object {
			return update.HarpoonTarg
		},
		breakHarpoon: deps.breakHarpoon,
		loadDamageArg: func() int32 {
			return damage
		},
		loadCurrent: func(health *HealthData) uint16 {
			return health.Cur
		},
		setHP: deps.setHP,
		loadFlags: func(unit *Object) uint32 {
			return uint32(unit.ObjFlags)
		},
		storeFlags: func(unit *Object, flags uint32) {
			unit.ObjFlags = object.Flags(flags)
		},
		buffOff: func(unit *Object, enchant int32) {
			deps.buffOff(unit, EnchantID(enchant))
		},
		isZombie: func(unit *Object) int32 {
			if deps.isZombie(unit) {
				return 1
			}
			return 0
		},
		soloReward: deps.soloReward,
		monsterDie: deps.monsterDie,
		loadDeath: func(unit *Object) unsafe.Pointer {
			return unit.Death
		},
		callDeath:     deps.callDeath,
		delayedDelete: deps.delayedDelete,
		informOwnerHP: deps.informOwnerHP,
	})
}

// UnitDamageClear4EE5E0 applies GAME.EXE 004EE5E0 through native-width object
// layouts. The registered Death callback is invoked directly with its native
// function pointer; the runtime fields identify the still-separate ports.
func (s *Server) UnitDamageClear4EE5E0(
	unit *Object,
	damage int32,
	runtime UnitDamageClearRuntime4EE5E0,
) {
	unitDamageClearNative4EE5E0(unit, damage, unitDamageClearNativeDeps4EE5E0{
		engineFlag: func(flag uint32) int32 {
			if noxflags.HasEngine(noxflags.EngineFlag(flag)) {
				return 1
			}
			return 0
		},
		breakHarpoon: runtime.BreakHarpoon,
		setHP:        runtime.SetHP,
		buffOff:      runtime.BuffOff,
		isZombie:     s.IsZombie,
		soloReward:   runtime.SoloReward,
		monsterDie:   runtime.MonsterDie,
		callDeath: func(death unsafe.Pointer, unit *Object) {
			ccall.CallVoidPtr(death, unit.CObj())
		},
		delayedDelete: runtime.DelayedDelete,
		informOwnerHP: s.MobInformOwnerHP4EE4C0,
	})
}
