package server

import (
	"unsafe"

	"github.com/opennox/libs/strman"

	"github.com/opennox/opennox/v1/common/sound"
)

// ActivatePoisonRuntime4EE7E0 identifies the poison-protection callback by
// address. Poison state, player status, and reporting now remain inside the
// restored native-width server path.
type ActivatePoisonRuntime4EE7E0 struct {
	PoisonProtectEngage unsafe.Pointer
}

type activatePoisonNativeDeps4EE7E0 struct {
	poisonProtection func(*Object) float64
	randomInt        func(int32, int32, string, int32) int32
	priorityMessage  func(*Object, string, uint8)
	setPoison        func(*Object, int32)
	audio            func(uint32, *Object, int32, uint32)
	frame            func() uint32
}

func activatePoisonNative4EE7E0(
	unit *Object,
	increment, maximum int32,
	deps activatePoisonNativeDeps4EE7E0,
) int32 {
	return activatePoison4EE7E0(activatePoisonHooks4EE7E0[
		*Object,
		*PlayerUpdateData,
		*Player,
		*HealthData,
	]{
		loadUnitArg: func() *Object {
			return unit
		},
		loadCurrent: func(unit *Object) uint8 {
			return unit.Poison540
		},
		loadFlagsLow: func(unit *Object) uint8 {
			return uint8(unit.ObjFlags)
		},
		testBuff: func(unit *Object, enchant uint32) int32 {
			if unit.HasEnchant(EnchantID(enchant)) {
				return 1
			}
			return 0
		},
		loadClass: func(unit *Object) uint32 {
			return uint32(unit.ObjClass)
		},
		loadUpdateData: func(unit *Object) *PlayerUpdateData {
			return (*PlayerUpdateData)(unit.UpdateData)
		},
		loadPlayer: func(update *PlayerUpdateData) *Player {
			return update.Player
		},
		loadPlayerFlags: func(player *Player) uint32 {
			return player.Field3680
		},
		loadSubClass: func(unit *Object) uint32 {
			return uint32(unit.ObjSubClass)
		},
		poisonProtection: deps.poisonProtection,
		floatToInt:       activatePoisonRound4EE7E0,
		randomInt:        deps.randomInt,
		priorityMessage:  deps.priorityMessage,
		loadIncrementArg: func() int32 {
			return increment
		},
		loadMaximumArg: func() int32 {
			return maximum
		},
		setPoison: deps.setPoison,
		audio:     deps.audio,
		loadHealth: func(unit *Object) *HealthData {
			return unit.HealthData
		},
		frame: deps.frame,
		storePoisonFrame: func(health *HealthData, frame uint32) {
			health.Field16 = frame
		},
	})
}

// ActivatePoison4EE7E0 binds GAME.EXE 004EE7E0 to native-width server layouts.
// Poison protection 004E0040 and poison state assignment 004EEA90 both stay
// inside their restored native-width server implementations.
func (s *Server) ActivatePoison4EE7E0(
	unit *Object,
	increment, maximum int32,
	runtime ActivatePoisonRuntime4EE7E0,
) int32 {
	return activatePoisonNative4EE7E0(unit, increment, maximum, activatePoisonNativeDeps4EE7E0{
		poisonProtection: func(unit *Object) float64 {
			return s.PoisonProtection4E0040(unit, PoisonProtectionRuntime4E0040{
				PoisonProtectEngage: runtime.PoisonProtectEngage,
			})
		},
		randomInt: func(minimum, maximum int32, _ string, _ int32) int32 {
			return int32(s.Rand.Logic.IntClamp(int(minimum), int(maximum)))
		},
		priorityMessage: func(unit *Object, message string, value uint8) {
			s.NetPriMsgToPlayer(unit, strman.ID(message), value)
		},
		setPoison: func(unit *Object, value int32) {
			s.SetPoison4EEA90(unit, value)
		},
		audio: func(id uint32, unit *Object, kind int32, code uint32) {
			s.Audio.EventObj(sound.ID(id), unit, int(kind), code)
		},
		frame: s.Frame,
	})
}
