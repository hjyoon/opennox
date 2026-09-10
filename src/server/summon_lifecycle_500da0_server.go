package server

import (
	"encoding/binary"
	"math"
	"unsafe"

	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/common/sound"
)

// SummonRuntime500DA0 supplies the root/legacy services which cannot live in
// package server. Records and objects themselves remain native-width pointers.
type SummonRuntime500DA0 struct {
	GameFlags            func(uint32) int32
	GuideSize            func(int32) int32
	CheckLimit           func(*Object, int32) bool
	MapTileAllowTeleport func(*types.Pointf) int32
	SummonAt             func(int, types.Pointf, *Object, Dir16) *Object
	LoadEffectID         func() uint16
	StoreEffectID        func(uint16)
	FloatToInt           func(float32) int32
}

func storeSummonPayload500DA0(record *DurSpell, payload summonPayload500DA0) {
	x := math.Float32bits(payload.position.X)
	y := math.Float32bits(payload.position.Y)
	record.Field72 = int32(uint32(payload.typeID) | x<<16)
	record.Field76 = uintptr(x>>16 | y<<16)
	record.Field80 = y>>16 |
		uint32(payload.direction)<<16 |
		uint32(byte(payload.effectID))<<24
	record.Field84 = record.Field84&0xffff0000 |
		uint32(payload.effectID>>8) |
		uint32(payload.complete)<<8
}

func summonPayloadTypeID500DA0(record *DurSpell) uint16 {
	return uint16(record.Field72)
}

func summonPayloadPosition500DA0(record *DurSpell) types.Pointf {
	word72 := uint32(record.Field72)
	word76 := uint32(record.Field76)
	return types.Pointf{
		X: math.Float32frombits(word72>>16 | word76<<16),
		Y: math.Float32frombits(word76>>16 | record.Field80<<16),
	}
}

func summonPayloadDirection500DA0(record *DurSpell) byte {
	return byte(record.Field80 >> 16)
}

func summonPayloadEffectID500DA0(record *DurSpell) uint16 {
	return uint16(record.Field80>>24) | uint16(byte(record.Field84))<<8
}

func summonPayloadComplete500DA0(record *DurSpell) byte {
	return byte(record.Field84 >> 8)
}

func storeSummonPayloadComplete500DA0(record *DurSpell, complete byte) {
	record.Field84 = record.Field84&^0xff00 | uint32(complete)<<8
}

func (s *Server) summonPlacement500F40(
	record *DurSpell,
	destination *types.Pointf,
	runtime SummonRuntime500DA0,
) int32 {
	return summonPlacement500F40(record, destination, summonPlacementHooks500F40[
		*DurSpell,
		*Object,
		*types.Pointf,
		*PlayerUpdateData,
		*Player,
	]{
		loadCaster: func(record *DurSpell) *Object {
			return record.Caster16
		},
		loadClassLowByte: func(obj *Object) byte {
			return byte(obj.ObjClass)
		},
		loadObjectX: func(obj *Object) float32 {
			return obj.PosVec.X
		},
		loadObjectY: func(obj *Object) float32 {
			return obj.PosVec.Y
		},
		loadTargetX: func(record *DurSpell) float32 {
			return record.Pos2.X
		},
		loadTargetY: func(record *DurSpell) float32 {
			return record.Pos2.Y
		},
		trace: func(from, to types.Pointf, flags MapTraceFlags) int32 {
			if s.MapTraceRayAt(from, to, nil, nil, flags) {
				return 1
			}
			return 0
		},
		tileAllow: runtime.MapTileAllowTeleport,
		storeX: func(destination *types.Pointf, value float32) {
			destination.X = value
		},
		storeY: func(destination *types.Pointf, value float32) {
			destination.Y = value
		},
		loadUpdate: func(obj *Object) *PlayerUpdateData {
			return (*PlayerUpdateData)(obj.UpdateData)
		},
		loadPlayer: func(update *PlayerUpdateData) *Player {
			return update.Player
		},
		loadPlayerIndex: func(player *Player) byte {
			return player.PlayerInd
		},
		inform: func(index, code byte, value int32) {
			_ = s.NetInformTextMsg(ntype.PlayerInd(index), code, int(value))
		},
	})
}

func (s *Server) sendSummonStartEffect500DA0(
	effectID uint16,
	position types.Pointf,
	direction byte,
	typeID uint16,
	duration uint16,
) {
	var packet [12]byte
	packet[0] = 126
	binary.LittleEndian.PutUint16(packet[1:3], uint16(int64(position.X)))
	binary.LittleEndian.PutUint16(packet[3:5], uint16(int64(position.Y)))
	binary.LittleEndian.PutUint16(packet[5:7], effectID)
	binary.LittleEndian.PutUint16(packet[7:9], typeID)
	packet[9] = direction
	binary.LittleEndian.PutUint16(packet[10:12], duration)
	_ = s.NetSendPacketXxx0(255, packet[:], nil, 1)
}

func (s *Server) sendSummonCancelEffect5011C0(effectID uint16) {
	var packet [3]byte
	packet[0] = 127
	binary.LittleEndian.PutUint16(packet[1:3], effectID)
	_ = s.NetSendPacketXxx0(255, packet[:], nil, 1)
}

func recordAddressDword500DA0(record *DurSpell) uint32 {
	return uint32(uintptr(unsafe.Pointer(record)))
}

func checkSummonLimitResult500DA0(runtime SummonRuntime500DA0, caster *Object, guide int32) int32 {
	if runtime.CheckLimit(caster, guide) {
		return 1
	}
	return 0
}

func (s *Server) summonPrivateMessage500DA0(caster *Object, message string, value byte) {
	s.NetPriMsgToPlayer(caster, strman.ID(message), value)
}

func (s *Server) summonAudioComplete5010D0(obj *Object) {
	s.Audio.EventObj(sound.SoundSummonComplete, obj, 0, 0)
}

func (s *Server) summonAudioAbort5011C0(position types.Pointf) {
	s.Audio.EventPos(sound.SoundSummonAbort, position, 0, 0)
}

// SummonStart500DA0 binds GAME.EXE 00500DA0 and its placement helper to
// native-width duration/object/update/player pointers.
//
//go:noinline
func (s *Server) SummonStart500DA0(record *DurSpell, runtime SummonRuntime500DA0) int32 {
	return summonStart500DA0(record, summonStartHooks500DA0[
		*DurSpell,
		*Object,
		*PlayerUpdateData,
		*Player,
	]{
		loadSpell: func(record *DurSpell) uint32 {
			return record.Spell
		},
		loadCaster: func(record *DurSpell) *Object {
			return record.Caster16
		},
		loadObjectFlags: func(obj *Object) uint32 {
			return uint32(obj.ObjFlags)
		},
		loadRecordFlag: func(record *DurSpell) uint32 {
			return record.Flag20
		},
		loadClassLowByte: func(obj *Object) byte {
			return byte(obj.ObjClass)
		},
		loadUpdate: func(obj *Object) *PlayerUpdateData {
			return (*PlayerUpdateData)(obj.UpdateData)
		},
		gameFlags: runtime.GameFlags,
		loadPlayer: func(update *PlayerUpdateData) *Player {
			return update.Player
		},
		loadGuideLevel: func(player *Player, guide int32) uint32 {
			return player.BeastScrollLvl[guide]
		},
		privateMessage: s.summonPrivateMessage500DA0,
		checkLimit: func(caster *Object, guide int32) int32 {
			return checkSummonLimitResult500DA0(runtime, caster, guide)
		},
		place: func(record *DurSpell, destination *types.Pointf) int32 {
			return s.summonPlacement500F40(record, destination, runtime)
		},
		loadDirectionLow: func(obj *Object) byte {
			return byte(obj.Direction1)
		},
		guideName: func(guide int32) string {
			return RewardFieldGuideName4F0D20(int(guide))
		},
		typeID: func(name string) uint16 {
			return uint16(s.Types.IndByID(name))
		},
		loadEffectID:  runtime.LoadEffectID,
		storeEffectID: runtime.StoreEffectID,
		storePayload:  storeSummonPayload500DA0,
		guideSize:     runtime.GuideSize,
		balanceFloat: func(key string, selector int32) float64 {
			return s.Balance.FloatInd(key, int(selector))
		},
		floatToInt:  runtime.FloatToInt,
		recordDword: recordAddressDword500DA0,
		frame:       s.Frame,
		storeDeadline: func(record *DurSpell, deadline uint32) {
			record.Frame68 = deadline
		},
		sendStartEffect: s.sendSummonStartEffect500DA0,
	})
}

// SummonFinish5010D0 binds GAME.EXE 005010D0 to native-width Go state.
//
//go:noinline
func (s *Server) SummonFinish5010D0(record *DurSpell, runtime SummonRuntime500DA0) int32 {
	return summonFinish5010D0(record, summonFinishHooks5010D0[
		*DurSpell,
		*Object,
		*PlayerUpdateData,
		*Player,
	]{
		loadCaster: func(record *DurSpell) *Object {
			return record.Caster16
		},
		loadObjectFlags: func(obj *Object) uint32 {
			return uint32(obj.ObjFlags)
		},
		loadDeadline: func(record *DurSpell) uint32 {
			return record.Frame68
		},
		frame: s.Frame,
		loadSpell: func(record *DurSpell) uint32 {
			return record.Spell
		},
		loadClassLowByte: func(obj *Object) byte {
			return byte(obj.ObjClass)
		},
		loadUpdate: func(obj *Object) *PlayerUpdateData {
			return (*PlayerUpdateData)(obj.UpdateData)
		},
		gameFlags: runtime.GameFlags,
		loadPlayer: func(update *PlayerUpdateData) *Player {
			return update.Player
		},
		loadGuideLevel: func(player *Player, guide int32) uint32 {
			return player.BeastScrollLvl[guide]
		},
		privateMessage: s.summonPrivateMessage500DA0,
		checkLimit: func(caster *Object, guide int32) int32 {
			return checkSummonLimitResult500DA0(runtime, caster, guide)
		},
		loadPayloadDirection: summonPayloadDirection500DA0,
		loadPayloadTypeID:    summonPayloadTypeID500DA0,
		loadPayloadPosition:  summonPayloadPosition500DA0,
		summon: func(typeID uint16, position types.Pointf, caster *Object, direction byte) *Object {
			return runtime.SummonAt(int(typeID), position, caster, Dir16(direction))
		},
		audioObject:   s.summonAudioComplete5010D0,
		storeComplete: storeSummonPayloadComplete500DA0,
	})
}

// SummonCancel5011C0 binds GAME.EXE 005011C0 to the semantic payload stored in
// the native-width duration record.
//
//go:noinline
func (s *Server) SummonCancel5011C0(record *DurSpell) {
	summonCancel5011C0(record, summonCancelHooks5011C0[*DurSpell]{
		loadComplete:        summonPayloadComplete500DA0,
		loadEffectID:        summonPayloadEffectID500DA0,
		sendCancelEffect:    s.sendSummonCancelEffect5011C0,
		loadPayloadPosition: summonPayloadPosition500DA0,
		audioPosition:       s.summonAudioAbort5011C0,
	})
}
