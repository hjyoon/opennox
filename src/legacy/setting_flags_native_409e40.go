package legacy

import "unsafe"

// Field 0 is the actual C subflags DWORD; field 1 is the actual C int dirty cell.
// A before-read event has no loaded value. Observing it must not add a read.
type settingFlagsNativeOperand409E40 struct {
	Site       uint32
	Field      int
	Address    unsafe.Pointer
	Width      uint32
	Value      uint32
	Write      bool
	Complete   bool
	ValueValid bool
}

type settingFlagsNativeObserver409E40 struct {
	Operand func(settingFlagsNativeOperand409E40)
	Service func(site, mask uint32) (uint32, bool)
}

type settingFlagsNativeObserverFrame409E40 struct {
	observer *settingFlagsNativeObserver409E40
	previous *settingFlagsNativeObserverFrame409E40
}

// Serial diagnostic scopes only. Production has no observer. This is not a
// concurrency facility; reentrant diagnostics must restore scopes in LIFO order.
var settingFlagsNativeObserverStack409E40 *settingFlagsNativeObserverFrame409E40

func pushSettingFlagsNativeObserver409E40(observer *settingFlagsNativeObserver409E40) func() {
	frame := &settingFlagsNativeObserverFrame409E40{
		observer: observer,
		previous: settingFlagsNativeObserverStack409E40,
	}
	settingFlagsNativeObserverStack409E40 = frame
	return func() {
		if settingFlagsNativeObserverStack409E40 != frame {
			panic("setting flags observer scopes must be restored in LIFO order")
		}
		settingFlagsNativeObserverStack409E40 = frame.previous
	}
}

func settingFlagsNativeObserverCurrent409E40() *settingFlagsNativeObserver409E40 {
	if frame := settingFlagsNativeObserverStack409E40; frame != nil {
		return frame.observer
	}
	return nil
}

func settingFlagsNativeRead409E40(site uint32, field int) uint32 {
	address := settingFlagsNativeCell409E40(field)
	observer := settingFlagsNativeObserverCurrent409E40()
	if observer != nil && observer.Operand != nil {
		observer.Operand(settingFlagsNativeOperand409E40{
			Site: site, Field: field, Address: address, Width: 4,
		})
	}
	value := *(*uint32)(address)
	if observer != nil && observer.Operand != nil {
		observer.Operand(settingFlagsNativeOperand409E40{
			Site: site, Field: field, Address: address, Width: 4,
			Value: value, Complete: true, ValueValid: true,
		})
	}
	return value
}

func settingFlagsNativeWrite409E40(site uint32, field int, value uint32) {
	address := settingFlagsNativeCell409E40(field)
	observer := settingFlagsNativeObserverCurrent409E40()
	if observer != nil && observer.Operand != nil {
		observer.Operand(settingFlagsNativeOperand409E40{
			Site: site, Field: field, Address: address, Width: 4,
			Value: value, Write: true, ValueValid: true,
		})
	}
	*(*uint32)(address) = value
	if observer != nil && observer.Operand != nil {
		observer.Operand(settingFlagsNativeOperand409E40{
			Site: site, Field: field, Address: address, Width: 4,
			Value: value, Write: true, Complete: true, ValueValid: true,
		})
	}
}

func settingFlagsNativeService409E40(site, mask uint32) uint32 {
	if observer := settingFlagsNativeObserverCurrent409E40(); observer != nil && observer.Service != nil {
		if word, handled := observer.Service(site, mask); handled {
			return word
		}
	}
	switch site {
	case 0x409E90, 0x409F0A, 0x409F52:
		// C canonicalizes its own bool; handled observer DWORDs stay unmodified.
		return rawDefaultDamageGameFlagsNative4E0B30(mask)
	case 0x409EA4, 0x409F28:
		return uint32(settingFlagsNativeRefresh409E40())
	default:
		panic("unknown setting flags service site")
	}
}

// These bodies follow the complete sealed original instructions, including
// live rereads and raw full-EAX returns. They do not model external services.
func sub409E40Native(argument int32) int32 {
	word := uint32(argument)
	if settingFlagsNativeRead409E40(0x409E44, 0) != word {
		settingFlagsNativeWrite409E40(0x409E4E, 0, word)
		settingFlagsNativeWrite409E40(0x409E53, 1, 1)
	}
	return argument
}

func sub409E70Native(argument int32) int32 {
	word := uint32(argument)
	result := settingFlagsNativeRead409E40(0x409E70, 0) & word
	if result != word {
		// The second original MOV reads the live cell again.
		flags := settingFlagsNativeRead409E40(0x409E80, 0)
		settingFlagsNativeWrite409E40(0x409E8A, 0, flags|word)
		result = settingFlagsNativeService409E40(0x409E90, 1)
		if result != 0 && word&0x2000 != 0 {
			result = settingFlagsNativeService409E40(0x409EA4, 0)
		}
		settingFlagsNativeWrite409E40(0x409EA9, 1, 1)
	}
	return int32(result)
}

func sub409EC0Native(argument int32) int32 {
	word := uint32(argument)
	flags := settingFlagsNativeRead409E40(0x409EC4, 0)
	if word&flags == 0 {
		return argument
	}
	result := ^word
	flags &= result
	// ECX is already computed before the original dirty store. Do not reread
	// subflags if an observer changes it after that completed store.
	settingFlagsNativeWrite409E40(0x409ED2, 1, 1)
	settingFlagsNativeWrite409E40(0x409EDC, 0, flags)
	return int32(result)
}

func sub409EF0Native(argument int32) int32 {
	word := uint32(argument)
	flags := settingFlagsNativeRead409E40(0x409EF0, 0)
	settingFlagsNativeWrite409E40(0x409F04, 0, flags^word)
	result := settingFlagsNativeService409E40(0x409F0A, 1)
	if result != 0 && word&0x2000 != 0 {
		result = settingFlagsNativeRead409E40(0x409F1E, 0)
		if result&0x2000 != 0 {
			result = settingFlagsNativeService409E40(0x409F28, 0)
		}
	}
	settingFlagsNativeWrite409E40(0x409F2D, 1, 1)
	return int32(result)
}

func sub409F40Native(argument int32) int32 {
	word := uint32(argument)
	if word == 0x2000 && settingFlagsNativeService409E40(0x409F52, 1056) != 0 {
		return 1
	}
	if settingFlagsNativeRead409E40(0x409F65, 0)&word != 0 {
		return 1
	}
	return 0
}
