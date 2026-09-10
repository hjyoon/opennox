package server

import (
	"encoding/binary"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/common/sound"
)

// CharmRuntime5011F0 supplies the services still owned by package legacy.
// Duration records, objects, update data, definitions, and players stay at
// native pointer width throughout all three charm phases.
type CharmRuntime5011F0 struct {
	GameFlag          func(uint32) bool
	GuideSize         func(int32) int32
	Charmable         func(uint16) int32
	FloatToInt        func(float32) int32
	Distance          func(*Object, *Object) float64
	CheckLimit        func(*Object, int32) bool
	BuffApplyRuntime  BuffApplyRuntime4FF380
	BuffOffRuntime    SpellBuffOffRuntime4FF5B0
	Attribution       func(*Object, *Object)
	UnitOrderRuntime  UnitOrderRuntime533900
	ChangeTeam        func(*ObjectTeam, uint32)
	SetHP             func(*Object, uint16)
	QuestSpawnCleanup func(*Object)
}

func charmRecordDword5011F0(record *DurSpell) uint32 {
	return uint32(uintptr(unsafe.Pointer(record)))
}

func charmObjectDword501690(obj *Object) int32 {
	return int32(uint32(uintptr(unsafe.Pointer(obj))))
}

func (s *Server) charmPrivateMessage5011F0(obj *Object, message string, value byte) {
	s.NetPriMsgToPlayer(obj, strman.ID(message), value)
}

func (s *Server) charmAudio5011F0(id int32, obj *Object) {
	s.Audio.EventObj(sound.ID(id), obj, 0, 0)
}

func (s *Server) charmReportTotalHealth5013E0(playerInd byte, obj *Object) int {
	if obj.HealthData == nil {
		return 0
	}
	var packet [7]byte
	packet[0] = byte(netmsg.MSG_REPORT_TOTAL_HEALTH)
	binary.LittleEndian.PutUint16(packet[1:3], uint16(s.GetUnitNetCode(obj)))
	health := obj.HealthData
	binary.LittleEndian.PutUint16(packet[3:5], health.Cur)
	binary.LittleEndian.PutUint16(packet[5:7], health.Max)
	return s.NetSendPacketXxx1(int(playerInd), packet[:], nil, 1)
}

func (s *Server) charmReportAcquire5013E0(
	playerInd byte,
	obj *Object,
	gameFlag func(uint32) bool,
) int {
	var packet [5]byte
	packet[0] = byte(netmsg.MSG_REPORT_ACQUIRE_CREATURE)
	binary.LittleEndian.PutUint16(packet[1:3], uint16(s.GetUnitNetCode(obj)))
	binary.LittleEndian.PutUint16(packet[3:5], obj.TypeInd)
	if gameFlag(charmAcquireTypeFlag5013E0) {
		packet[4] |= 0x80
	}
	s.NetSendPacketXxx1(int(playerInd), packet[:], nil, 1)
	return s.charmReportTotalHealth5013E0(playerInd, obj)
}

func (s *Server) charmSendSimpleObject5013E0(
	playerInd byte,
	obj *Object,
	floatToInt func(float32) int32,
) int {
	var packet [9]byte
	packet[0] = byte(netmsg.MSG_SIMPLE_OBJ)
	binary.LittleEndian.PutUint16(packet[3:5], obj.TypeInd)
	binary.LittleEndian.PutUint16(packet[1:3], uint16(s.GetUnitNetCode(obj)))
	x := obj.PosVec.X
	binary.LittleEndian.PutUint16(packet[5:7], uint16(floatToInt(x)))
	y := obj.PosVec.Y
	binary.LittleEndian.PutUint16(packet[7:9], uint16(floatToInt(y)))
	return s.NetSendPacketXxx1(int(playerInd), packet[:], nil, 1)
}

// CharmStart5011F0 binds GAME.EXE 005011F0 to native-width server state.
//
//go:noinline
func (s *Server) CharmStart5011F0(record *DurSpell, runtime CharmRuntime5011F0) int32 {
	return charmStart5011F0(record, charmStartHooks5011F0[
		*DurSpell,
		*Object,
		*PlayerUpdateData,
		*Player,
	]{
		loadRecordFlag: func(record *DurSpell) uint32 {
			return record.Flag20
		},
		balanceFloat: func(key string) float64 {
			return s.Balance.Float(key)
		},
		balanceFloatInd: func(key string, selector int32) float64 {
			return s.Balance.FloatInd(key, int(selector))
		},
		floatToInt: runtime.FloatToInt,
		loadLevel: func(record *DurSpell) uint32 {
			return record.Level
		},
		loadTarget: func(record *DurSpell) *Object {
			return record.Target48
		},
		storeTarget: func(record *DurSpell, target *Object) {
			record.Target48 = target
		},
		loadCaster: func(record *DurSpell) *Object {
			return record.Caster16
		},
		loadSpell: func(record *DurSpell) uint32 {
			return record.Spell
		},
		spellFlags: func(id uint32) uint32 {
			return uint32(s.Spells.Flags(spell.ID(id)))
		},
		loadPosition: func(record *DurSpell) *types.Pointf {
			return &record.Pos2
		},
		searchTarget: func(
			position *types.Pointf,
			caster *Object,
			flags uint32,
			distance float32,
			mode int32,
			self *Object,
		) *Object {
			return s.Nox_xxx_spellFlySearchTarget(
				position, caster, things.SpellFlags(flags), distance, int(mode), self,
			)
		},
		loadClassLow: func(obj *Object) byte {
			return byte(obj.ObjClass)
		},
		monitored: Nox_xxx_creatureIsMonitored_500CC0,
		loadTypeIndex: func(obj *Object) uint16 {
			return obj.TypeInd
		},
		charmable: runtime.Charmable,
		loadUpdate: func(obj *Object) *PlayerUpdateData {
			return (*PlayerUpdateData)(obj.UpdateData)
		},
		loadPlayer: func(update *PlayerUpdateData) *Player {
			return update.Player
		},
		loadGuideLevel: func(player *Player, guide int32) uint32 {
			return player.BeastScrollLvl[guide]
		},
		privateMessage: s.charmPrivateMessage5011F0,
		audio:          s.charmAudio5011F0,
		guideSize:      runtime.GuideSize,
		recordDword:    charmRecordDword5011F0,
		frame:          s.Frame,
		storeDeadline: func(record *DurSpell, deadline uint32) {
			record.Frame68 = deadline
		},
		buffApply: func(target *Object, buff int32, duration int16, power int8) {
			s.BuffApply4FF380(target, buff, duration, power, runtime.BuffApplyRuntime)
		},
		attribution:      runtime.Attribution,
		durationRayStart: s.DurationRayStart4FF130,
	})
}

// CharmFinish5013E0 binds GAME.EXE 005013E0 to native-width server state.
//
//go:noinline
func (s *Server) CharmFinish5013E0(record *DurSpell, runtime CharmRuntime5011F0) int32 {
	return charmFinish5013E0(record, charmFinishHooks5013E0[
		*DurSpell,
		*Object,
		*MonsterUpdateData,
		*MonsterDef,
		*PlayerUpdateData,
		*Player,
	]{
		loadTarget: func(record *DurSpell) *Object {
			return record.Target48
		},
		loadCaster: func(record *DurSpell) *Object {
			return record.Caster16
		},
		loadObjectFlags: func(obj *Object) uint32 {
			return uint32(obj.ObjFlags)
		},
		distance: runtime.Distance,
		loadClassLow: func(obj *Object) byte {
			return byte(obj.ObjClass)
		},
		privateMessage: s.charmPrivateMessage5011F0,
		audio:          s.charmAudio5011F0,
		loadDeadline: func(record *DurSpell) uint32 {
			return record.Frame68
		},
		frame: s.Frame,
		loadSubclass: func(obj *Object) uint32 {
			return uint32(obj.ObjSubClass)
		},
		storeSubclass: func(obj *Object, subclass uint32) {
			obj.ObjSubClass = object.SubClass(subclass)
		},
		loadTypeIndex: func(obj *Object) uint16 {
			return obj.TypeInd
		},
		charmable:  runtime.Charmable,
		checkLimit: runtime.CheckLimit,
		buffOff: func(target *Object, buff int32) int32 {
			return s.SpellBuffOff4FF5B0(target, buff, runtime.BuffOffRuntime)
		},
		findParent: func(obj *Object) *Object {
			return obj.FindOwnerChainPlayer()
		},
		clearOwner: s.ObjClearOwner,
		setOwner:   s.ObjSetOwner,
		hasTeam: func(obj *Object) bool {
			return obj.HasTeam()
		},
		loadNetCode: func(obj *Object) uint32 {
			return obj.NetCode
		},
		changeTeam: func(obj *Object, netCode uint32) {
			runtime.ChangeTeam(obj.TeamPtr(), netCode)
		},
		loadMonsterUpdate: func(obj *Object) *MonsterUpdateData {
			return (*MonsterUpdateData)(obj.UpdateData)
		},
		loadMonsterStatus: func(update *MonsterUpdateData) uint32 {
			return uint32(update.StatusFlags)
		},
		storeMonsterStatus: func(update *MonsterUpdateData, status uint32) {
			update.StatusFlags = object.MonsterStatus(status)
		},
		gameFlag: runtime.GameFlag,
		typeHealthMax: func(typeIndex uint16) uint16 {
			return s.Types.ByInd(int(typeIndex)).Health().Max
		},
		unitHP: UnitGetHP4EE780,
		loadMonsterDef: func(update *MonsterUpdateData) *MonsterDef {
			return update.MonsterDef
		},
		monsterDefIsNil: func(definition *MonsterDef) bool {
			return definition == nil
		},
		monsterQuestMax: func(definition *MonsterDef) uint16 {
			return uint16(definition.HealthQuest72)
		},
		storeHealthMax: func(obj *Object, maximum uint16) {
			obj.HealthData.Max = maximum
		},
		setHP: runtime.SetHP,
		loadPlayerUpdate: func(obj *Object) *PlayerUpdateData {
			return (*PlayerUpdateData)(obj.UpdateData)
		},
		loadPlayer: func(update *PlayerUpdateData) *Player {
			return update.Player
		},
		loadSummonOrder: func(player *Player) uint32 {
			return player.SummonOrderAll
		},
		loadPlayerIndex: func(player *Player) byte {
			return player.PlayerInd
		},
		orderUnit: func(owner, creature *Object, order uint32) {
			s.OrderUnit533900(owner, creature, order, runtime.UnitOrderRuntime)
		},
		reportAcquire: func(playerInd byte, obj *Object) {
			s.charmReportAcquire5013E0(playerInd, obj, runtime.GameFlag)
		},
		markMinimap: func(playerInd byte, obj *Object, flags uint32) {
			s.Players.Nox_xxx_netMarkMinimapObject_417190(ntype.PlayerInd(playerInd), obj, flags)
		},
		sendSimpleObject: func(playerInd byte, obj *Object) {
			s.charmSendSimpleObject5013E0(playerInd, obj, runtime.FloatToInt)
		},
		questSpawnCleanup: runtime.QuestSpawnCleanup,
		loadSpell: func(record *DurSpell) uint32 {
			return record.Spell
		},
		spellAudio: func(id uint32, selector int32) int32 {
			return int32(s.Spells.DefByInd(spell.ID(id)).GetAudio(int(selector)))
		},
	})
}

// CharmCancel501690 binds GAME.EXE 00501690 to native-width server state.
//
//go:noinline
func (s *Server) CharmCancel501690(record *DurSpell, runtime CharmRuntime5011F0) int32 {
	return charmCancel501690(record, charmCancelHooks501690[*DurSpell, *Object]{
		loadTarget: func(record *DurSpell) *Object {
			return record.Target48
		},
		loadObjectFlags: func(obj *Object) uint32 {
			return uint32(obj.ObjFlags)
		},
		objectDword: charmObjectDword501690,
		buffOff: func(target *Object, buff int32) int32 {
			return s.SpellBuffOff4FF5B0(target, buff, runtime.BuffOffRuntime)
		},
	})
}
