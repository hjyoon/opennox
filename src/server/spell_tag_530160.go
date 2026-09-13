package server

import (
	"encoding/binary"

	"github.com/opennox/libs/noxnet/netmsg"

	"github.com/opennox/opennox/v1/common/ntype"
)

const (
	spellTagDurationBalance530160 = "TagDurationPerLevel"
	spellTagPlayerClass530160     = byte(4)
	spellTagInvalidFlags530160    = uint32(0x8020)
)

type spellTagHooks530160[Record, Object comparable] struct {
	loadCaster      func(Record) Object
	loadTarget      func(Record) Object
	loadClass       func(Object) byte
	loadFlags       func(Object) uint32
	loadPlayerIndex func(Object) int32
	loadLevel       func(Record) uint32
	loadFrame       func() uint32
	storeFrame      func(Record, uint32)
	loadBalance     func(string) float64
	floatToInt      func(float32) int32
	mark            func(int32, Object, uint32)
	unmark          func(int32, Object, uint32)
	unitCode        func(Object) uint16
	objectType      func(Object) uint16
	send            func(int32, [7]byte)
}

func spellTagPacket530160(code, objectType uint16, action byte) [7]byte {
	var packet [7]byte
	packet[0] = byte(netmsg.MSG_INTERESTING_ID)
	binary.LittleEndian.PutUint16(packet[1:3], code)
	binary.LittleEndian.PutUint16(packet[3:5], objectType)
	packet[5] = action
	packet[6] = 1 // Tag marker, distinct from the ordinary interesting-ID marker.
	return packet
}

// spellTagCreate530160 follows GAME.EXE 00530160. A missing caster is
// rejected instead of reproducing the original's premature UpdateData read.
func spellTagCreate530160[Record, Object comparable](record Record, h spellTagHooks530160[Record, Object]) int32 {
	var nilObject Object
	caster := h.loadCaster(record)
	if caster == nilObject || h.loadFlags(caster)&spellTagInvalidFlags530160 != 0 || h.loadClass(caster)&spellTagPlayerClass530160 == 0 {
		return 1
	}
	target := h.loadTarget(record)
	if target == nilObject || h.loadFlags(target)&spellTagInvalidFlags530160 != 0 {
		return 1
	}

	duration := h.floatToInt(float32(h.loadBalance(spellTagDurationBalance530160)))
	level := h.loadLevel(record)
	h.storeFrame(record, h.loadFrame()+level*uint32(duration))
	h.mark(h.loadPlayerIndex(caster), h.loadTarget(record), 1)
	code := h.unitCode(h.loadTarget(record))
	objectType := h.objectType(h.loadTarget(record))
	h.send(h.loadPlayerIndex(caster), spellTagPacket530160(code, objectType, 1))
	return 0
}

// spellTagDestroy530270 follows GAME.EXE 00530270. The original callback's
// return register is undefined; duration destruction ignores it.
func spellTagDestroy530270[Record, Object comparable](record Record, h spellTagHooks530160[Record, Object]) {
	var nilObject Object
	caster := h.loadCaster(record)
	if caster == nilObject || h.loadClass(caster)&spellTagPlayerClass530160 == 0 {
		return
	}
	target := h.loadTarget(record)
	if target == nilObject {
		return
	}
	if h.loadClass(target)&spellTagPlayerClass530160 == 0 {
		h.unmark(h.loadPlayerIndex(caster), target, 1)
	}
	code := h.unitCode(h.loadTarget(record))
	objectType := h.objectType(h.loadTarget(record))
	h.send(h.loadPlayerIndex(caster), spellTagPacket530160(code, objectType, 2))
}

func spellTagServerHooks530160(s *Server) spellTagHooks530160[*DurSpell, *Object] {
	return spellTagHooks530160[*DurSpell, *Object]{
		loadCaster: func(record *DurSpell) *Object { return record.Caster16 },
		loadTarget: func(record *DurSpell) *Object { return record.Target48 },
		loadClass:  func(obj *Object) byte { return byte(obj.ObjClass) },
		loadFlags:  func(obj *Object) uint32 { return uint32(obj.ObjFlags) },
		loadPlayerIndex: func(obj *Object) int32 {
			return int32(byte(obj.ControllingPlayer().PlayerInd))
		},
		loadLevel:   func(record *DurSpell) uint32 { return record.Level },
		loadFrame:   s.Frame,
		storeFrame:  func(record *DurSpell, frame uint32) { record.Frame68 = frame },
		loadBalance: s.Balance.Float,
		floatToInt:  playerUnitInitFloatToInt4EFE80,
		mark: func(index int32, obj *Object, flags uint32) {
			s.Players.Nox_xxx_netMarkMinimapObject_417190(ntype.PlayerInd(index), obj, flags)
		},
		unmark: func(index int32, obj *Object, flags uint32) {
			s.Players.Nox_xxx_netUnmarkMinimapObj_417300(ntype.PlayerInd(index), obj, flags)
		},
		unitCode:   func(obj *Object) uint16 { return uint16(s.GetUnitNetCode(obj)) },
		objectType: func(obj *Object) uint16 { return obj.TypeInd },
		send: func(index int32, packet [7]byte) {
			s.NetSendPacketXxx0(int(index), packet[:], nil, 1)
		},
	}
}

// SpellTagCreate530160 binds the Tag creation callback to native-width
// duration and object fields.
//
//go:noinline
func (s *Server) SpellTagCreate530160(record *DurSpell) int32 {
	return spellTagCreate530160(record, spellTagServerHooks530160(s))
}

// SpellTagDestroy530270 binds the Tag removal callback to native-width
// duration and object fields.
//
//go:noinline
func (s *Server) SpellTagDestroy530270(record *DurSpell) {
	spellTagDestroy530270(record, spellTagServerHooks530160(s))
}
