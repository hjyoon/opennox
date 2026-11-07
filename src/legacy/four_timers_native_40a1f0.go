package legacy

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

// Field 0 is the live active DWORD, fields 1/2 the deadline low/high DWORDs,
// and field 3 the selected live limit BYTE. Addresses are the actual C blobs.
// A before-read observation has no value and must not add a read.
type timerNativeOperand40A1F0 struct {
	Site       uint32
	Field      int
	Address    unsafe.Pointer
	Width      uint32
	Value      uint32
	Write      bool
	Complete   bool
	ValueValid bool
}

// These are declared service reply words, not emulated native CPU registers.
// Clock consumes EDX:EAX; other services consume the entire EAX DWORD.
// Physical native ECX/EDX equivalence is not claimed for the C int entries.
type timerNativeReply40A1F0 struct{ EAX, ECX, EDX uint32 }

type timerNativeObserver40A1F0 struct {
	Operand func(timerNativeOperand40A1F0)
	Service func(site uint32, arguments []uint32) (timerNativeReply40A1F0, bool)
}

type timerNativeObserverFrame40A1F0 struct {
	observer *timerNativeObserver40A1F0
	previous *timerNativeObserverFrame40A1F0
}

// Serial diagnostic scopes only; production has no observer. Nested scopes
// restore in LIFO order. This is not a concurrency facility.
var timerNativeObserverStack40A1F0 *timerNativeObserverFrame40A1F0

func pushTimerNativeObserver40A1F0(observer *timerNativeObserver40A1F0) func() {
	frame := &timerNativeObserverFrame40A1F0{observer: observer, previous: timerNativeObserverStack40A1F0}
	timerNativeObserverStack40A1F0 = frame
	return func() {
		if timerNativeObserverStack40A1F0 != frame {
			panic("timer observer scopes must be restored in LIFO order")
		}
		timerNativeObserverStack40A1F0 = frame.previous
	}
}

func timerNativeObserverCurrent40A1F0() *timerNativeObserver40A1F0 {
	if frame := timerNativeObserverStack40A1F0; frame != nil {
		return frame.observer
	}
	return nil
}

func timerNativeCell40A1F0(field int, index uint32) unsafe.Pointer {
	switch field {
	case 0:
		return memmap.PtrOff(0x587000, 4660)
	case 1:
		return memmap.PtrOff(0x5D4594, 3468)
	case 2:
		return memmap.PtrOff(0x5D4594, 3472)
	case 3:
		// Preserve DWORD offset arithmetic. The frozen independent ModeLookup
		// reply domain is indices 0..5; no new bounds guard or clamp is added.
		return memmap.PtrOff(0x5D4594, uintptr(uint32(3500)+index))
	default:
		panic("unknown timer live field")
	}
}

func timerNativeRead40A1F0(site uint32, field int, index uint32) uint32 {
	address := timerNativeCell40A1F0(field, index)
	width := uint32(4)
	if field == 3 {
		width = 1
	}
	observer := timerNativeObserverCurrent40A1F0()
	if observer != nil && observer.Operand != nil {
		observer.Operand(timerNativeOperand40A1F0{Site: site, Field: field, Address: address, Width: width})
	}
	var value uint32
	if width == 1 {
		value = uint32(*(*byte)(address))
	} else {
		value = *(*uint32)(address)
	}
	if observer != nil && observer.Operand != nil {
		observer.Operand(timerNativeOperand40A1F0{Site: site, Field: field, Address: address, Width: width, Value: value, Complete: true, ValueValid: true})
	}
	return value
}

func timerNativeWrite40A1F0(site uint32, field int, value uint32) {
	address := timerNativeCell40A1F0(field, 0)
	observer := timerNativeObserverCurrent40A1F0()
	if observer != nil && observer.Operand != nil {
		observer.Operand(timerNativeOperand40A1F0{Site: site, Field: field, Address: address, Width: 4, Value: value, Write: true, ValueValid: true})
	}
	*(*uint32)(address) = value
	if observer != nil && observer.Operand != nil {
		observer.Operand(timerNativeOperand40A1F0{Site: site, Field: field, Address: address, Width: 4, Value: value, Write: true, Complete: true, ValueValid: true})
	}
}

func timerNativeService40A1F0(site uint32, arguments ...uint32) timerNativeReply40A1F0 {
	if observer := timerNativeObserverCurrent40A1F0(); observer != nil && observer.Service != nil {
		// Declared memory effects may change live cells, but the original
		// cached/pushed argument words do not become mutable callback inputs.
		if reply, handled := observer.Service(site, append([]uint32(nil), arguments...)); handled {
			return reply
		}
	}
	switch site {
	case 0x40A1FD:
		// The C helper canonicalizes its own bool. Handled raw replies above
		// remain unmodified, including nonzero values with a zero low byte.
		return timerNativeReply40A1F0{EAX: rawDefaultDamageGameFlagsNative4E0B30(arguments[0])}
	case 0x40A20F:
		return timerNativeReply40A1F0{EAX: uint32(timerNativeNotify40A1F0(int32(arguments[0]), int32(arguments[1])))}
	case 0x40A252:
		return timerNativeReply40A1F0{EAX: uint32(nox_common_gameFlags_getVal_40A5B0())}
	case 0x40A258:
		// The caller's pushed DWORD is observed above before the existing
		// helper's public signed-short input conversion.
		return timerNativeReply40A1F0{EAX: uint32(sub409A70Native(int16(arguments[0])))}
	case 0x40A233, 0x40A262, 0x40A311:
		// Consume the existing full-width Go clock. The old unrelated public
		// C.uint nox_platform_get_ticks export is deliberately unchanged.
		clock := PlatformTicks()
		return timerNativeReply40A1F0{EAX: uint32(clock), EDX: uint32(clock >> 32)}
	default:
		panic("unknown timer service site")
	}
}

// These bodies follow the complete sealed original caller instructions. The
// seven external services retain their explicit, separately qualified scope.
func sub40A1F0Native(argument int32) int32 {
	word := uint32(argument)
	timerNativeWrite40A1F0(0x40A1F7, 0, word)
	result := timerNativeService40A1F0(0x40A1FD, 1).EAX
	if result != 0 {
		result = timerNativeService40A1F0(0x40A20F, 159, word).EAX
	}
	return int32(result)
}

func sub40A230Native() int32 {
	clock := timerNativeService40A1F0(0x40A233)
	deadlineLow := timerNativeRead40A1F0(0x40A23A, 1, 0)
	return int32(deadlineLow - clock.EAX)
}

func sub40A250Native() int64 {
	flags := timerNativeService40A1F0(0x40A252).EAX
	index := timerNativeService40A1F0(0x40A258, flags).EAX
	clock := timerNativeService40A1F0(0x40A262)
	limit := timerNativeRead40A1F0(0x40A26B, 3, index)
	// Original LEA/SHL operations are DWORD arithmetic followed by CDQ.
	duration := limit * 60000
	signedDuration := int64(int32(duration))
	deadline := (uint64(clock.EDX)<<32 | uint64(clock.EAX)) + uint64(signedDuration)
	timerNativeWrite40A1F0(0x40A28A, 1, uint32(deadline))
	timerNativeWrite40A1F0(0x40A290, 2, uint32(deadline>>32))
	return signedDuration
}

func sub40A310Native(argument int32) int64 {
	clock := timerNativeService40A1F0(0x40A311)
	// Original CDQ sign-extends the raw input DWORD before ADD/ADC.
	result := int64(argument)
	deadline := (uint64(clock.EDX)<<32 | uint64(clock.EAX)) + uint64(result)
	timerNativeWrite40A1F0(0x40A323, 1, uint32(deadline))
	timerNativeWrite40A1F0(0x40A329, 2, uint32(deadline>>32))
	return result
}
