package legacy

import (
	"runtime"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

type settingEligibilityNativeArg40AA70 struct {
	Word      uint32
	Pointer   unsafe.Pointer
	IsPointer bool
}

type settingEligibilityNativeReply40AA70 struct {
	Word    uint32
	Pointer unsafe.Pointer
	ECX     uint32
	EDX     uint32
}

// Field0 is cached Info52 BYTE; Field1 argument status DWORD; Field2 argument
// key DWORD; Field3 iterator2140 DWORD; Field4 the independent global DWORD.
// Offset names the native layout. Address is the actual load address, and a
// nil base performs an actual nil load without manufacturing a low Go pointer.
type settingEligibilityNativeOperand40AA70 struct {
	Site         uint32
	Field        int
	Base         unsafe.Pointer
	Address      unsafe.Pointer
	Offset       uintptr
	Width        uint32
	ValueWord    uint32
	Complete     bool
	ValueValid   bool
	AddressValid bool
}

type settingEligibilityNativeObserver40AA70 struct {
	Operand func(settingEligibilityNativeOperand40AA70)
	Service func(site uint32, arguments []settingEligibilityNativeArg40AA70) (settingEligibilityNativeReply40AA70, bool)
	// Optional diagnostic C-owned/shared DWORD. It is loaded directly, never
	// copied to a private shadow. Nil uses the actual production memmap cell.
	// Diagnostic storage is not claimed to alias the production owner or PE.
	GlobalCell unsafe.Pointer
}

type settingEligibilityNativeObserverFrame40AA70 struct {
	observer *settingEligibilityNativeObserver40AA70
	previous *settingEligibilityNativeObserverFrame40AA70
}

// Serial diagnostic scopes only. Reentrant helpers restore the exact frame.
var settingEligibilityNativeObserverStack40AA70 *settingEligibilityNativeObserverFrame40AA70

func pushSettingEligibilityNativeObserver40AA70(observer *settingEligibilityNativeObserver40AA70) func() {
	frame := &settingEligibilityNativeObserverFrame40AA70{observer: observer, previous: settingEligibilityNativeObserverStack40AA70}
	settingEligibilityNativeObserverStack40AA70 = frame
	return func() {
		if settingEligibilityNativeObserverStack40AA70 != frame {
			panic("setting eligibility observer scopes must be restored in LIFO order")
		}
		settingEligibilityNativeObserverStack40AA70 = frame.previous
	}
}

func settingEligibilityNativeObserverCurrent40AA70() *settingEligibilityNativeObserver40AA70 {
	if frame := settingEligibilityNativeObserverStack40AA70; frame != nil {
		return frame.observer
	}
	return nil
}

// This always identifies the production owner, even inside a diagnostic scope.
func settingEligibilityNativeCell40AA70() unsafe.Pointer {
	return memmap.PtrOff(0x5D4594, 3508)
}

func settingEligibilityNativeAddress40AA70(base unsafe.Pointer, offset uintptr) unsafe.Pointer {
	if base == nil {
		// Physical original nil+offset and native nil effective addresses are
		// separate observations. Do not reconstruct a Go pointer from a DWORD.
		return nil
	}
	return unsafe.Add(base, offset)
}

//go:noinline
func settingEligibilityNativeLoadByte40AA70(address unsafe.Pointer) byte {
	return *(*byte)(address)
}

//go:noinline
func settingEligibilityNativeLoadWord40AA70(address unsafe.Pointer) uint32 {
	return *(*uint32)(address)
}

func settingEligibilityNativeRead40AA70(site uint32, field int, base unsafe.Pointer, offset uintptr, width uint32) uint32 {
	address := settingEligibilityNativeAddress40AA70(base, offset)
	observer := settingEligibilityNativeObserverCurrent40AA70()
	event := settingEligibilityNativeOperand40AA70{Site: site, Field: field, Base: base, Address: address, Offset: offset, Width: width, AddressValid: true}
	if observer != nil && observer.Operand != nil {
		observer.Operand(event)
	}
	var value uint32
	if width == 1 {
		value = uint32(settingEligibilityNativeLoadByte40AA70(address))
	} else {
		value = settingEligibilityNativeLoadWord40AA70(address)
	}
	event.ValueWord, event.Complete, event.ValueValid = value, true, true
	if observer != nil && observer.Operand != nil {
		observer.Operand(event)
	}
	runtime.KeepAlive(base)
	return value
}

func settingEligibilityNativeReadGlobal40AA70() uint32 {
	var cell unsafe.Pointer
	if observer := settingEligibilityNativeObserverCurrent40AA70(); observer != nil {
		cell = observer.GlobalCell
	}
	if cell == nil {
		cell = settingEligibilityNativeCell40AA70()
	}
	return settingEligibilityNativeRead40AA70(0x40AAB3, 4, cell, 0, 4)
}

func settingEligibilityNativeService40AA70(site uint32, arguments ...settingEligibilityNativeArg40AA70) settingEligibilityNativeReply40AA70 {
	defer func() { runtime.KeepAlive(arguments) }()
	if observer := settingEligibilityNativeObserverCurrent40AA70(); observer != nil && observer.Service != nil {
		// Intercept before typed/default helpers. A callback receives a copy
		// of the actual cached argument slice and cannot replace those words.
		if reply, handled := observer.Service(site, append([]settingEligibilityNativeArg40AA70(nil), arguments...)); handled {
			return reply
		}
	}
	return settingEligibilityNativeServiceDefault40AA70(site, arguments)
}

func sub40AA70Native(player unsafe.Pointer) int32 {
	defer func() { runtime.KeepAlive(player) }()
	// The actual getter is called first and its full pointer remains cached.
	info := settingEligibilityNativeService40AA70(0x40AA72).Pointer
	defer func() { runtime.KeepAlive(info) }()
	if player != nil {
		if settingEligibilityNativeService40AA70(0x40AA8A, settingEligibilityNativeArg40AA70{Word: 0x1000}).Word != 0 {
			count := settingEligibilityNativeService40AA70(0x40AA96).Word
			if int32(count) < 6 {
				return 1
			}
			return 0
		}
		status := settingEligibilityNativeRead40AA70(0x40AAA8, 1, player, unsafe.Offsetof(server.Player{}.Field3680), 4)
		if status&0x100 != 0 && settingEligibilityNativeReadGlobal40AA70() == 0 {
			return 0
		}
		if settingEligibilityNativeService40AA70(0x40AAC1).Word != 0 ||
			settingEligibilityNativeService40AA70(0x40AACF, settingEligibilityNativeArg40AA70{Word: 0x8000}).Word != 0 {
			key := settingEligibilityNativeRead40AA70(0x40AADF, 2, player, unsafe.Offsetof(server.Player{}.Field2068), 4)
			if key == 0 {
				return 0
			}
			team := settingEligibilityNativeService40AA70(0x40AAED, settingEligibilityNativeArg40AA70{Word: key}).Pointer
			if team == nil {
				limit := settingEligibilityNativeRead40AA70(0x40AAFD, 0, info, 52, 1)
				clamp := settingEligibilityNativeService40AA70(0x40AB02, settingEligibilityNativeArg40AA70{Word: 96}).Word != 0
				if !clamp && settingEligibilityNativeService40AA70(0x40AB10, settingEligibilityNativeArg40AA70{Word: 16}).Word != 0 {
					clamp = settingEligibilityNativeService40AA70(0x40AB1E, settingEligibilityNativeArg40AA70{Word: 4}).Word != 0
				}
				if clamp && int32(limit) > 2 {
					limit = 2
				}
				count := settingEligibilityNativeService40AA70(0x40AB34).Word & 0xff
				if int32(count) >= int32(limit) {
					return 0
				}
				// The second Check96 and char call are independent sites.
				if settingEligibilityNativeService40AA70(0x40AB49, settingEligibilityNativeArg40AA70{Word: 96}).Word != 0 {
					second := settingEligibilityNativeService40AA70(0x40AB56).Word & 0xff
					maximum := settingEligibilityNativeService40AA70(0x40AB63).Word
					if int32(second) >= int32(maximum) {
						return 0
					}
				}
			}
		}
	}
	if settingEligibilityNativeService40AA70(0x40AB77, settingEligibilityNativeArg40AA70{Word: 128}).Word != 0 {
		return 1
	}
	if settingEligibilityNativeService40AA70(0x40AB88, settingEligibilityNativeArg40AA70{Word: 1024}).Word == 0 {
		return 1
	}
	iterator := settingEligibilityNativeService40AA70(0x40AB94).Pointer
	for iterator != nil {
		word := settingEligibilityNativeRead40AA70(0x40AB9D, 3, iterator, unsafe.Offsetof(server.Player{}.Field2140), 4)
		// Original TEST ECX/JG is signed. The retained C uint32 comparison
		// cannot stand in for this predicate on 0x80000000..0xffffffff.
		if int32(word) > 0 {
			if settingEligibilityNativeService40AA70(0x40ABBC).Word == 0 {
				return 1
			}
			return 0
		}
		iterator = settingEligibilityNativeService40AA70(0x40ABA8, settingEligibilityNativeArg40AA70{Pointer: iterator, IsPointer: true}).Pointer
	}
	return 1
}
