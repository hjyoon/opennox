package legacy

import (
	"runtime"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

type suddenDeathNativeArg40A970 struct {
	Kind    string // word, player, or immutable gamedata string
	Word    uint32
	Pointer unsafe.Pointer
	String  string
}

type suddenDeathNativeReply40A970 struct {
	Word         uint32
	Pointer      unsafe.Pointer
	FloatFormat  string
	FloatPayload [10]byte
}

// Fields0..4 are frame/start/deaths/threshold/lifetime DWORDs. Field5 is
// player status, using the named native Player layout and an actual DWORD.
type suddenDeathNativeOperand40A970 struct {
	Site                                      uint32
	Field                                     int
	Base, Address                             unsafe.Pointer
	Offset                                    uintptr
	Width                                     uint32
	Value                                     uint32
	Write, Complete, ValueValid, AddressValid bool
}

// These are numerical store/converter boundaries, not native x87 opcodes or
// private stack snapshots. Bits are the actual numeric helper result used by
// this caller. In particular raw private NaN FSTP payload equivalence is not
// claimed by the public integer-state contract.
type suddenDeathNativeNumeric40A970 struct {
	Site                             uint32
	Kind                             string // store32 or convert32
	Bits, Result                     uint32
	Complete, BitsValid, ResultValid bool
}

type suddenDeathNativeObserver40A970 struct {
	ControlWord uint16
	// Each nonnil cell is actual C-owned diagnostic storage used by the
	// engine itself. It is not a copy of a private Go local and does not
	// claim an alias to the production Go frame or C blob globals.
	Cells   [5]unsafe.Pointer
	Operand func(suddenDeathNativeOperand40A970)
	Service func(site uint32, arguments []suddenDeathNativeArg40A970) (suddenDeathNativeReply40A970, bool)
	Numeric func(suddenDeathNativeNumeric40A970)
}

type suddenDeathNativeObserverFrame40A970 struct {
	observer *suddenDeathNativeObserver40A970
	previous *suddenDeathNativeObserverFrame40A970
}

// Serial diagnostic scopes only. Production has no observer. Distinct frame
// identities also preserve same-observer-pointer nested LIFO restoration.
var suddenDeathNativeObserverStack40A970 *suddenDeathNativeObserverFrame40A970

func pushSuddenDeathNativeObserver40A970(observer *suddenDeathNativeObserver40A970) func() {
	frame := &suddenDeathNativeObserverFrame40A970{observer: observer, previous: suddenDeathNativeObserverStack40A970}
	suddenDeathNativeObserverStack40A970 = frame
	return func() {
		if suddenDeathNativeObserverStack40A970 != frame {
			panic("sudden death observer scopes must restore in LIFO order")
		}
		suddenDeathNativeObserverStack40A970 = frame.previous
	}
}

func suddenDeathNativeObserverCurrent40A970() *suddenDeathNativeObserver40A970 {
	if frame := suddenDeathNativeObserverStack40A970; frame != nil {
		return frame.observer
	}
	return nil
}

// This accessor always exposes the actual production C blob cells. Diagnostic
// redirection is kept separate so a default-entry control can inspect them.
func suddenDeathNativeCell40A970(field int) unsafe.Pointer {
	switch field {
	case 1:
		return memmap.PtrOff(0x5D4594, 3520)
	case 2:
		return memmap.PtrOff(0x5D4594, 3536)
	case 3:
		return memmap.PtrOff(0x5D4594, 3476)
	case 4:
		return memmap.PtrOff(0x5D4594, 1392)
	default:
		panic("unknown sudden death production cell")
	}
}

func suddenDeathNativeEffectiveCell40A970(observer *suddenDeathNativeObserver40A970, field int) unsafe.Pointer {
	if observer != nil && observer.Cells[field] != nil {
		return observer.Cells[field]
	}
	if field == 0 {
		return nil // The real production frame owner is Go state.
	}
	return suddenDeathNativeCell40A970(field)
}

func suddenDeathNativeObserveOperand40A970(observer *suddenDeathNativeObserver40A970, event suddenDeathNativeOperand40A970) {
	if observer != nil && observer.Operand != nil {
		observer.Operand(event)
	}
}

func suddenDeathNativeReadFrame40A970() uint32 {
	observer := suddenDeathNativeObserverCurrent40A970()
	address := suddenDeathNativeEffectiveCell40A970(observer, 0)
	event := suddenDeathNativeOperand40A970{Site: 0x40A970, Field: 0, Base: address, Address: address, Width: 4, AddressValid: address != nil}
	suddenDeathNativeObserveOperand40A970(observer, event)
	var value uint32
	if address != nil {
		value = *(*uint32)(address)
	} else {
		value = gameFrame()
	}
	event.Value, event.Complete, event.ValueValid = value, true, true
	suddenDeathNativeObserveOperand40A970(observer, event)
	runtime.KeepAlive(address)
	return value
}

func suddenDeathNativeWrite40A970(site uint32, field int, value uint32) {
	observer := suddenDeathNativeObserverCurrent40A970()
	address := suddenDeathNativeEffectiveCell40A970(observer, field)
	event := suddenDeathNativeOperand40A970{Site: site, Field: field, Base: address, Address: address, Width: 4, Value: value, Write: true, ValueValid: true, AddressValid: true}
	suddenDeathNativeObserveOperand40A970(observer, event)
	*(*uint32)(address) = value
	event.Complete = true
	suddenDeathNativeObserveOperand40A970(observer, event)
	runtime.KeepAlive(address)
}

//go:noinline
func suddenDeathNativeReadWord40A970(address unsafe.Pointer) uint32 {
	return *(*uint32)(address)
}

func suddenDeathNativeReadStatus40A970(player unsafe.Pointer) uint32 {
	offset := unsafe.Offsetof(server.Player{}.Field3680)
	address := player
	if player != nil {
		address = unsafe.Add(player, offset)
	}
	observer := suddenDeathNativeObserverCurrent40A970()
	event := suddenDeathNativeOperand40A970{Site: 0x40A9C5, Field: 5, Base: player, Address: address, Offset: offset, Width: 4, AddressValid: true}
	suddenDeathNativeObserveOperand40A970(observer, event)
	// No pointer-to-Player struct cast or new admission guard. Foreign partial
	// and unaligned inputs perform their actual scalar operand access.
	value := suddenDeathNativeReadWord40A970(address)
	event.Value, event.Complete, event.ValueValid = value, true, true
	suddenDeathNativeObserveOperand40A970(observer, event)
	runtime.KeepAlive(player)
	return value
}

func suddenDeathNativeService40A970(site uint32, arguments ...suddenDeathNativeArg40A970) suddenDeathNativeReply40A970 {
	if observer := suddenDeathNativeObserverCurrent40A970(); observer != nil && observer.Service != nil {
		if reply, handled := observer.Service(site, append([]suddenDeathNativeArg40A970(nil), arguments...)); handled {
			return reply
		}
	}
	return suddenDeathNativeServiceDefault40A970(site, arguments)
}

func suddenDeathNativeObserveNumeric40A970(event suddenDeathNativeNumeric40A970) {
	if observer := suddenDeathNativeObserverCurrent40A970(); observer != nil && observer.Numeric != nil {
		observer.Numeric(event)
	}
}

func suddenDeathNativeStore32Bits40A970(site uint32, reply suddenDeathNativeReply40A970, cw uint16) uint32 {
	suddenDeathNativeObserveNumeric40A970(suddenDeathNativeNumeric40A970{Site: site, Kind: "store32"})
	value := rawDefaultDamageFloatFromReply4E0B30(rawDefaultDamageNativeReply4E0B30{FloatFormat: reply.FloatFormat, FloatPayload: reply.FloatPayload})
	bits := rawDefaultDamageFloatStore32Bits4E0B30(value, cw)
	suddenDeathNativeObserveNumeric40A970(suddenDeathNativeNumeric40A970{Site: site, Kind: "store32", Bits: bits, Complete: true, BitsValid: true})
	return bits
}

func suddenDeathNativeConvert32Bits40A970(site uint32, bits uint32, cw uint16) uint32 {
	suddenDeathNativeObserveNumeric40A970(suddenDeathNativeNumeric40A970{Site: site, Kind: "convert32", Bits: bits, BitsValid: true})
	value := uint32(rawDefaultDamageFloatConvert32Bits4E0B30(bits, cw))
	suddenDeathNativeObserveNumeric40A970(suddenDeathNativeNumeric40A970{Site: site, Kind: "convert32", Bits: bits, Result: value, Complete: true, BitsValid: true, ResultValid: true})
	return value
}

func sub40A970Native() {
	cw := uint16(0x0e3f)
	if observer := suddenDeathNativeObserverCurrent40A970(); observer != nil && observer.ControlWord != 0 {
		cw = observer.ControlWord
	}
	if cw != 0x0e3f && cw != 0x023f {
		panic("sudden death supports only its two original numerical control words")
	}
	frame := suddenDeathNativeReadFrame40A970()
	suddenDeathNativeWrite40A970(0x40A97B, 1, frame)
	suddenDeathNativeWrite40A970(0x40A980, 2, 0)
	threshold := suddenDeathNativeService40A970(0x40A98A, suddenDeathNativeArg40A970{Kind: "string", String: "SuddenDeathPlayerThreshold"})
	thresholdBits := suddenDeathNativeStore32Bits40A970(0x40A98F, threshold, cw)
	thresholdWord := suddenDeathNativeConvert32Bits40A970(0x40A992, thresholdBits, cw)
	suddenDeathNativeWrite40A970(0x40A99C, 3, thresholdWord)
	lifetime := suddenDeathNativeService40A970(0x40A9A1, suddenDeathNativeArg40A970{Kind: "string", String: "SuddenDeathLifeTime"})
	lifetimeBits := suddenDeathNativeStore32Bits40A970(0x40A9A6, lifetime, cw)
	lifetimeWord := suddenDeathNativeConvert32Bits40A970(0x40A9AD, lifetimeBits, cw)
	suddenDeathNativeWrite40A970(0x40A9B5, 4, lifetimeWord)
	player := suddenDeathNativeService40A970(0x40A9BA).Pointer
	for player != nil {
		status := suddenDeathNativeReadStatus40A970(player)
		if status&0x100 != 0 {
			suddenDeathNativeService40A970(0x40A9D6, suddenDeathNativeArg40A970{Kind: "player", Pointer: player}, suddenDeathNativeArg40A970{Kind: "word", Word: 256})
		}
		player = suddenDeathNativeService40A970(0x40A9DF, suddenDeathNativeArg40A970{Kind: "player", Pointer: player}).Pointer
	}
	suddenDeathNativeService40A970(0x40A9F2, suddenDeathNativeArg40A970{Kind: "word", Word: 0x4000000})
}
