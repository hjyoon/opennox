package server

import (
	"encoding/binary"
	"math"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

// MonsterGeneratorDamageRuntime4E27D0 supplies services outside the object
// layout for the native-width restoration of GAME.EXE 004E27D0.
type MonsterGeneratorDamageRuntime4E27D0 struct {
	Frame   func() uint32
	PointFX func(byte, byte, types.Pointf)
	Audio   func(int, *Object)
	Script  func(*ScriptCallback, *Object, *Object, ScriptEventType)
	Default DamageFunc
}

func monsterGeneratorDamagePointFXPacket523150(op, subtype byte, pos types.Pointf) [6]byte {
	var packet [6]byte
	packet[0] = op
	packet[1] = subtype
	binary.LittleEndian.PutUint16(packet[2:4], uint16(int32(math.RoundToEven(float64(pos.X)))))
	binary.LittleEndian.PutUint16(packet[4:6], uint16(int32(math.RoundToEven(float64(pos.Y)))))
	return packet
}

// Nox_xxx_netSendPointFx2_523150 restores the fixed six-byte point-effect
// packet built by GAME.EXE 00523150.
func (s *Server) Nox_xxx_netSendPointFx2_523150(op, subtype byte, pos types.Pointf) {
	packet := monsterGeneratorDamagePointFXPacket523150(op, subtype, pos)
	s.Nox_xxx_netSendFxAllCli_523030(pos, packet[:])
}

func monsterGeneratorDamageNative4E27D0(
	target, source, weapon *Object,
	damage int32,
	typ object.DamageType,
	runtime MonsterGeneratorDamageRuntime4E27D0,
) bool {
	return monsterGeneratorDamage4E27D0(target, source, weapon, damage, typ,
		monsterGeneratorDamageHooks4E27D0[*Object, *MonsterGenUpdateData, *HealthData]{
			loadFlags: func(object *Object) uint32 {
				return uint32(object.ObjFlags)
			},
			loadUpdate: func(object *Object) *MonsterGenUpdateData {
				// GAME.EXE has no class gate before loading offset 748.
				return (*MonsterGenUpdateData)(object.UpdateData)
			},
			frame: runtime.Frame,
			loadEffectFrame: func(object *Object) uint32 {
				return object.Frame134
			},
			loadPosX: func(object *Object) float32 {
				return object.PosVec.X
			},
			loadPosY: func(object *Object) float32 {
				return object.PosVec.Y
			},
			normalize: chestOpenNormalizeVector509F20,
			pointFX:   runtime.PointFX,
			audio:     runtime.Audio,
			getHP:     UnitGetHP4EE780,
			defaultDamage: func(target, source, weapon *Object, damage int32, typ object.DamageType) bool {
				return runtime.Default(target, source, weapon, damage, typ)
			},
			scriptDamage: func(update *MonsterGenUpdateData, source, target *Object, event ScriptEventType) {
				block := (*ScriptCallback)(unsafe.Pointer(&update.Field48))
				runtime.Script(block, source, target, event)
			},
			loadHealth: func(object *Object) *HealthData {
				return object.HealthData
			},
			loadHealthMax: func(health *HealthData) uint16 {
				return health.Max
			},
			loadHealthCur: func(health *HealthData) uint16 {
				return health.Cur
			},
			loadXStatus: func(object *Object) uint32 {
				return object.Field5
			},
			setXStatus: func(object *Object, status uint32) {
				object.SetXStatus(status)
			},
			unsetXStatus: func(object *Object, status uint32) {
				object.UnsetXStatus(status)
			},
		},
	)
}

// MonsterGeneratorDamage4E27D0 executes the MonsterGeneratorDamage callback
// without narrowing any Object or update-data pointer to the PE32 ABI.
func MonsterGeneratorDamage4E27D0(
	target, source, weapon *Object,
	damage int32,
	typ object.DamageType,
	runtime MonsterGeneratorDamageRuntime4E27D0,
) bool {
	return monsterGeneratorDamageNative4E27D0(target, source, weapon, damage, typ, runtime)
}
