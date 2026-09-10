package server

import (
	"encoding/binary"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/internal/netlist"
)

// UnitOrderRuntime533900 owns the three legacy services that cannot live in
// package server without creating an import cycle. All object and definition
// values cross this boundary at native pointer width.
type UnitOrderRuntime533900 struct {
	MonsterDefByType func(int) *MonsterDef
	Banish           func(*Object)
	Observe          func(*Object, *Object)
}

func (s *Server) monsterCommandSend528BD0(unit, source *Object, command string, arg int16) {
	packet := make([]byte, len(command)+12)
	packet[0] = byte(netmsg.MSG_TEXT_MESSAGE)
	binary.LittleEndian.PutUint16(packet[1:3], uint16(s.GetUnitNetCode(unit)))
	packet[3] = 8
	binary.LittleEndian.PutUint16(packet[4:6], uint16(int64(unit.PosVec.X)))
	binary.LittleEndian.PutUint16(packet[6:8], uint16(int64(unit.PosVec.Y)))
	packet[8] = byte(len(command) + 1)
	binary.LittleEndian.PutUint16(packet[9:11], uint16(arg))
	copy(packet[11:], command)
	packet = packet[:int(packet[8])+11]
	if source != nil {
		if uint8(source.ObjClass)&unitOrderPlayerClass533900 != 0 {
			s.NetList.AddToMsgListCli(source.UpdateDataPlayer().Player.PlayerIndex(), netlist.Kind1, packet)
		}
		return
	}
	for playerUnit := s.Players.FirstUnit(); playerUnit != nil; playerUnit = s.Players.NextUnit(playerUnit) {
		s.NetList.AddToMsgListCli(playerUnit.UpdateDataPlayer().Player.PlayerIndex(), netlist.Kind1, packet)
	}
}

func (s *Server) playUnitOrderSound5339A0(unit *Object) {
	if uint8(unit.ObjClass)&unitOrderMonsterClass533900 == 0 {
		return
	}
	update := (*MonsterUpdateData)(unit.UpdateData)
	if update.SoundSet122 == nil {
		return
	}
	id := *(*uint32)(unsafe.Add(update.SoundSet122, 68))
	s.Audio.EventObj(sound.ID(id), unit, 0, 0)
}

func unitOrderCanShoot534280(unit *Object) bool {
	update := (*MonsterUpdateData)(unit.UpdateData)
	if object.SubClass(unit.ObjSubClass).AsMonster().Has(object.MonsterNPC) {
		return update.WeaponEquipFlags&0x047f00fe != 0
	}
	return update.MonsterDef.MissileName148[0] != 0
}

func (s *Server) enactUnitOrder5339A0(source, unit *Object, order uint32, runtime UnitOrderRuntime533900) {
	enactUnitOrder5339A0(source, unit, order, unitOrderEnactHooks5339A0[*Object, *MonsterUpdateData, *MonsterDef, *AIStackItem]{
		loadUpdate: func(obj *Object) *MonsterUpdateData {
			return (*MonsterUpdateData)(obj.UpdateData)
		},
		loadClassLow: func(obj *Object) uint8 {
			return uint8(obj.ObjClass)
		},
		isZombie: s.IsZombie,
		loadObjectFlags: func(obj *Object) uint32 {
			return uint32(obj.ObjFlags)
		},
		loadMonsterDef: func(update *MonsterUpdateData) *MonsterDef {
			return update.MonsterDef
		},
		storeMonsterDef: func(update *MonsterUpdateData, def *MonsterDef) {
			update.MonsterDef = def
		},
		loadTypeIndex: func(obj *Object) uint16 {
			return obj.TypeInd
		},
		monsterDefByType: func(typ uint16) *MonsterDef {
			return runtime.MonsterDefByType(int(typ))
		},
		loadOwner: func(obj *Object) *Object {
			return obj.ObjOwner
		},
		banish:         runtime.Banish,
		observe:        runtime.Observe,
		monsterCommand: s.monsterCommandSend528BD0,
		playOrderSound: s.playUnitOrderSound5339A0,
		loadStatus: func(update *MonsterUpdateData) uint32 {
			return uint32(update.StatusFlags)
		},
		storeStatus: func(update *MonsterUpdateData, status uint32) {
			update.StatusFlags = object.MonsterStatus(status)
		},
		storeAggression: func(update *MonsterUpdateData, value float32) {
			update.Aggression = value
		},
		storeSightRange: func(update *MonsterUpdateData, value float32) {
			update.SightRange = value
		},
		isMoving: func(obj *Object) bool {
			return obj.SpeedBase >= unitOrderMovingSpeed5339A0
		},
		canShoot: unitOrderCanShoot534280,
		clearActionStack: func(obj *Object) {
			obj.ClearActionStack()
		},
		pushAction: func(obj *Object, action ai.ActionType) *AIStackItem {
			return obj.MonsterPushActionImpl(action, "go", 0)
		},
		setGuardArgs: func(action *AIStackItem, obj *Object) {
			action.SetArgs(obj.PosVec, uint32(obj.Direction1))
		},
		setEscortArgs: func(action *AIStackItem, obj *Object) {
			action.SetArgs(obj.PosVec, obj)
		},
	})
}

// OrderUnit533900 restores GAME.EXE 00533900 and 005339A0 without passing
// Object, UpdateData, MonsterDef, or AI stack pointers through PE32 integers.
func (s *Server) OrderUnit533900(owner, creature *Object, order uint32, runtime UnitOrderRuntime533900) {
	unitOrder533900(owner, creature, order, unitOrderDispatchHooks533900[*Object, *MonsterUpdateData]{
		loadClassLow: func(obj *Object) uint8 {
			return uint8(obj.ObjClass)
		},
		firstOwned: func(obj *Object) *Object {
			return obj.Field129
		},
		nextOwned: func(obj *Object) *Object {
			return obj.Field128
		},
		loadUpdate: func(obj *Object) *MonsterUpdateData {
			return (*MonsterUpdateData)(obj.UpdateData)
		},
		loadStatus: func(update *MonsterUpdateData) uint32 {
			return uint32(update.StatusFlags)
		},
		localOrder: func(obj *Object, order uint32) {
			s.Nox_xxx_orderUnitLocal_500C70(obj.UpdateDataPlayer().Player.PlayerIndex(), order)
		},
		enactOrder: func(source, unit *Object, order uint32) {
			s.enactUnitOrder5339A0(source, unit, order, runtime)
		},
	})
}
