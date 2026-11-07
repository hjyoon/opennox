package legacy

import (
	"runtime"
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/server"
)

type twoSettingsPointerNativeArg40A740 struct {
	Word      uint32
	Pointer   unsafe.Pointer
	IsPointer bool
}

type twoSettingsPointerNativeReply40A740 struct {
	Word    uint32
	Pointer unsafe.Pointer
	ECX     uint32
	EDX     uint32
}

// Fields: info flag BYTE, object update pointer, update player pointer,
// player status BYTE, player index BYTE, engine flags DWORD. Offset is a
// native named-layout offset. It is never an original PE address.
type twoSettingsPointerNativeOperand40A740 struct {
	Site         uint32
	Field        int
	Base         unsafe.Pointer
	Address      unsafe.Pointer
	Offset       uintptr
	Width        uint32
	IsPointer    bool
	ValueWord    uint32
	ValuePointer unsafe.Pointer
	Complete     bool
	ValueValid   bool
	AddressValid bool
}

type twoSettingsPointerNativeObserver40A740 struct {
	Operand func(twoSettingsPointerNativeOperand40A740)
	Service func(site uint32, arguments []twoSettingsPointerNativeArg40A740) (twoSettingsPointerNativeReply40A740, bool)
	// Diagnostic storage only: a real owned native DWORD or protected page.
	// Production nil observer reads the actual Go-owned full engine flags.
	// This cell is not claimed to alias a packed PE blob or production state.
	EngineCell unsafe.Pointer
}

type twoSettingsPointerNativeObserverFrame40A740 struct {
	observer *twoSettingsPointerNativeObserver40A740
	previous *twoSettingsPointerNativeObserverFrame40A740
}

// Serial diagnostic scopes only. Reentrant callbacks restore in LIFO order.
var twoSettingsPointerNativeObserverStack40A740 *twoSettingsPointerNativeObserverFrame40A740

func pushTwoSettingsPointerNativeObserver40A740(observer *twoSettingsPointerNativeObserver40A740) func() {
	frame := &twoSettingsPointerNativeObserverFrame40A740{observer: observer, previous: twoSettingsPointerNativeObserverStack40A740}
	twoSettingsPointerNativeObserverStack40A740 = frame
	return func() {
		if twoSettingsPointerNativeObserverStack40A740 != frame {
			panic("two settings pointer observer scopes must be restored in LIFO order")
		}
		twoSettingsPointerNativeObserverStack40A740 = frame.previous
	}
}

func twoSettingsPointerNativeObserverCurrent40A740() *twoSettingsPointerNativeObserver40A740 {
	if frame := twoSettingsPointerNativeObserverStack40A740; frame != nil {
		return frame.observer
	}
	return nil
}

func twoSettingsPointerNativeAddress40A740(base unsafe.Pointer, offset uintptr) unsafe.Pointer {
	if base == nil {
		// Still perform the actual nil load below. Do not manufacture a low
		// nonnil Go pointer by adding an offset to nil. Original and native
		// physical fault addresses are explicitly different observations.
		return nil
	}
	return unsafe.Add(base, offset)
}

//go:noinline
func twoSettingsPointerNativeLoadPointer40A740(address unsafe.Pointer) unsafe.Pointer {
	if uintptr(address)%unsafe.Alignof(unsafe.Pointer(nil)) == 0 {
		// Legal Go pointer fields retain their pointer-typed load and GC
		// semantics. Nil also takes this path and performs the actual read.
		return *(*unsafe.Pointer)(address)
	}
	// A misaligned pointer slot is foreign storage, not a legal Go pointer
	// field. Read its full native word before converting those foreign bits
	// in C, following the existing raw damage operand reader. This avoids a
	// checkptr alignment rejection before the actual guarded memory load.
	bits := *(*uintptr)(address)
	value := twoSettingsPointerNativeForeignPointer40A740(bits)
	runtime.KeepAlive(address)
	return value
}

//go:noinline
func twoSettingsPointerNativeLoadByte40A740(address unsafe.Pointer) byte {
	return *(*byte)(address)
}

//go:noinline
func twoSettingsPointerNativeLoadWord40A740(address unsafe.Pointer) uint32 {
	return *(*uint32)(address)
}

func twoSettingsPointerNativeReadPointer40A740(site uint32, field int, base unsafe.Pointer, offset uintptr) unsafe.Pointer {
	address := twoSettingsPointerNativeAddress40A740(base, offset)
	observer := twoSettingsPointerNativeObserverCurrent40A740()
	event := twoSettingsPointerNativeOperand40A740{Site: site, Field: field, Base: base, Address: address, Offset: offset, Width: uint32(unsafe.Sizeof(unsafe.Pointer(nil))), IsPointer: true, AddressValid: true}
	if observer != nil && observer.Operand != nil {
		observer.Operand(event)
	}
	value := twoSettingsPointerNativeLoadPointer40A740(address)
	runtime.KeepAlive(base)
	event.ValuePointer, event.Complete, event.ValueValid = value, true, true
	if observer != nil && observer.Operand != nil {
		observer.Operand(event)
	}
	return value
}

func twoSettingsPointerNativeReadByte40A740(site uint32, field int, base unsafe.Pointer, offset uintptr) uint32 {
	address := twoSettingsPointerNativeAddress40A740(base, offset)
	observer := twoSettingsPointerNativeObserverCurrent40A740()
	event := twoSettingsPointerNativeOperand40A740{Site: site, Field: field, Base: base, Address: address, Offset: offset, Width: 1, AddressValid: true}
	if observer != nil && observer.Operand != nil {
		observer.Operand(event)
	}
	value := uint32(twoSettingsPointerNativeLoadByte40A740(address))
	runtime.KeepAlive(base)
	event.ValueWord, event.Complete, event.ValueValid = value, true, true
	if observer != nil && observer.Operand != nil {
		observer.Operand(event)
	}
	return value
}

func twoSettingsPointerNativeReadEngine40A740(site uint32) uint32 {
	observer := twoSettingsPointerNativeObserverCurrent40A740()
	event := twoSettingsPointerNativeOperand40A740{Site: site, Field: 5, Width: 4}
	if observer != nil && observer.EngineCell != nil {
		event.Base, event.Address, event.AddressValid = observer.EngineCell, observer.EngineCell, true
	}
	if observer != nil && observer.Operand != nil {
		observer.Operand(event)
	}
	var value uint32
	if event.AddressValid {
		value = twoSettingsPointerNativeLoadWord40A740(event.Address)
	} else {
		// The actual production owner is Go state, not a fake C blob DWORD.
		value = uint32(noxflags.GetEngine())
	}
	event.ValueWord, event.Complete, event.ValueValid = value, true, true
	if observer != nil && observer.Operand != nil {
		observer.Operand(event)
	}
	runtime.KeepAlive(event.Base)
	return value
}

func twoSettingsPointerNativeService40A740(site uint32, arguments ...twoSettingsPointerNativeArg40A740) twoSettingsPointerNativeReply40A740 {
	if observer := twoSettingsPointerNativeObserverCurrent40A740(); observer != nil && observer.Service != nil {
		// Keep original cached/pushed arguments immutable to slice mutation.
		if reply, handled := observer.Service(site, append([]twoSettingsPointerNativeArg40A740(nil), arguments...)); handled {
			return reply
		}
	}
	return twoSettingsPointerNativeServiceDefault40A740(site, arguments)
}

func sub40A740Native() int32 {
	info := twoSettingsPointerNativeService40A740(0x40A741).Pointer
	check := twoSettingsPointerNativeService40A740(0x40A74D, twoSettingsPointerNativeArg40A740{Word: 128}).Word
	if check != 0 && twoSettingsPointerNativeReadByte40A740(0x40A759, 0, info, 53)&0x80 != 0 {
		return 1
	}
	return 0
}

func sub40A770Native() int32 {
	var count uint32
	teamMode := twoSettingsPointerNativeService40A740(0x40A778, twoSettingsPointerNativeArg40A740{Word: 4}).Word
	if teamMode != 0 {
		team := twoSettingsPointerNativeService40A740(0x40A784).Pointer
		for team != nil {
			// This is deliberately a fresh global scan for every team. The
			// original does not inspect team membership or call TeamCompare.
			unit := twoSettingsPointerNativeService40A740(0x40A79A).Pointer
			for unit != nil {
				update := twoSettingsPointerNativeReadPointer40A740(0x40A7A3, 1, unit, unsafe.Offsetof(server.Object{}.UpdateData))
				player := twoSettingsPointerNativeReadPointer40A740(0x40A7A9, 2, update, unsafe.Offsetof(server.PlayerUpdateData{}.Player))
				status := twoSettingsPointerNativeReadByte40A740(0x40A7AF, 3, player, unsafe.Offsetof(server.Player{}.Field3680))
				if status&1 == 0 {
					index := twoSettingsPointerNativeReadByte40A740(0x40A7B7, 4, player, unsafe.Offsetof(server.Player{}.PlayerInd))
					if index != 31 || twoSettingsPointerNativeReadEngine40A740(0x40A7C0)&0x40000 == 0 {
						count++
						break
					}
				}
				unit = twoSettingsPointerNativeService40A740(0x40A7C9, twoSettingsPointerNativeArg40A740{Pointer: unit, IsPointer: true}).Pointer
			}
			team = twoSettingsPointerNativeService40A740(0x40A7D9, twoSettingsPointerNativeArg40A740{Pointer: team, IsPointer: true}).Pointer
		}
		return int32(count)
	}
	player := twoSettingsPointerNativeService40A740(0x40A7EE).Pointer
	for player != nil {
		status := twoSettingsPointerNativeReadByte40A740(0x40A7FE, 3, player, unsafe.Offsetof(server.Player{}.Field3680))
		if status&1 == 0 {
			index := twoSettingsPointerNativeReadByte40A740(0x40A806, 4, player, unsafe.Offsetof(server.Player{}.PlayerInd))
			if index != 31 || twoSettingsPointerNativeReadEngine40A740(0x40A80F)&0x40000 == 0 {
				count++
			}
		}
		player = twoSettingsPointerNativeService40A740(0x40A819, twoSettingsPointerNativeArg40A740{Pointer: player, IsPointer: true}).Pointer
	}
	return int32(count)
}
