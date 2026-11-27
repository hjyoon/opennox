package legacy

import (
	"encoding/binary"
	"runtime"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

type tileListNativeObserverFrame411350 struct {
	Observer *tileListNativeObserver411350
	Previous *tileListNativeObserverFrame411350
}

var tileListNativeObserverHead411350 *tileListNativeObserverFrame411350

func tileListNativeObserverCurrent411350() *tileListNativeObserver411350 {
	if tileListNativeObserverHead411350 == nil {
		return nil
	}
	return tileListNativeObserverHead411350.Observer
}

func pushTileListNativeObserver411350(observer *tileListNativeObserver411350) func() {
	frame := &tileListNativeObserverFrame411350{observer, tileListNativeObserverHead411350}
	tileListNativeObserverHead411350 = frame
	return func() {
		if tileListNativeObserverHead411350 != frame {
			panic("tile-list411350 observer pop order")
		}
		tileListNativeObserverHead411350 = frame.Previous
	}
}

func tileListNativeObserveOperand411350(event tileListNativeOperand411350) {
	if observer := tileListNativeObserverCurrent411350(); observer != nil && observer.Operand != nil {
		observer.Operand(event)
	}
}

func tileListNativeObserveCall411350(event tileListNativeCall411350) {
	if observer := tileListNativeObserverCurrent411350(); observer != nil && observer.Call != nil {
		observer.Call(event)
	}
	runtime.KeepAlive(event.Pointer)
}

func tileListNativeAddress411350(base unsafe.Pointer, offset uintptr) unsafe.Pointer {
	if base == nil {
		// Retain the logical offset without manufacturing a low Go pointer.
		// Physical nil effective-address equivalence is not claimed.
		return nil
	}
	return unsafe.Add(base, offset)
}

//go:noinline
func tileListNativeLoadWord411350(address unsafe.Pointer) uint32 {
	if uintptr(address)%unsafe.Alignof(uint32(0)) == 0 {
		return *(*uint32)(address)
	}
	return binary.NativeEndian.Uint32(unsafe.Slice((*byte)(address), 4))
}

//go:noinline
func tileListNativeLoadPointer411350(address unsafe.Pointer) unsafe.Pointer {
	if uintptr(address)%unsafe.Alignof(unsafe.Pointer(nil)) == 0 {
		return *(*unsafe.Pointer)(address)
	}
	return loadUnalignedNativePointer(address)
}

func loadUnalignedNativePointer(address unsafe.Pointer) unsafe.Pointer {
	var value unsafe.Pointer
	width := unsafe.Sizeof(value)
	copy(unsafe.Slice((*byte)(unsafe.Pointer(&value)), width), unsafe.Slice((*byte)(address), width))
	return value
}

func tileListNativeReadWord411350(site uint32, field int, base unsafe.Pointer, offset uintptr) uint32 {
	address := tileListNativeAddress411350(base, offset)
	event := tileListNativeOperand411350{Site: site, Field: field, Base: base, Address: address, Offset: offset, Width: 4, AddressValid: true}
	tileListNativeObserveOperand411350(event)
	value := tileListNativeLoadWord411350(address)
	event.ValueWord, event.Complete, event.ValueValid = value, true, true
	tileListNativeObserveOperand411350(event)
	runtime.KeepAlive(base)
	return value
}

func tileListNativeReadPointer411350(site uint32, field int, base unsafe.Pointer, offset uintptr) unsafe.Pointer {
	address := tileListNativeAddress411350(base, offset)
	event := tileListNativeOperand411350{Site: site, Field: field, Base: base, Address: address, Offset: offset, Width: uint32(unsafe.Sizeof(unsafe.Pointer(nil))), IsPointer: true, AddressValid: true}
	tileListNativeObserveOperand411350(event)
	value := tileListNativeLoadPointer411350(address)
	event.ValuePointer, event.Complete, event.ValueValid = value, true, true
	tileListNativeObserveOperand411350(event)
	runtime.KeepAlive(base)
	runtime.KeepAlive(value)
	return value
}

func tileListNativeReadEdgeByte411350(site, index, rawOffset uint32) uint8 {
	logical := uint32(0x85b3fc) + rawOffset
	var base, address unsafe.Pointer
	var offset uintptr
	if observer := tileListNativeObserverCurrent411350(); observer != nil && observer.EdgeByteAddress != nil {
		address = observer.EdgeByteAddress(site, index, rawOffset, logical)
		base = address
	} else {
		// The existing EDGE byte owner has stride60. It is distinct from
		// the native TileDef array and contains no relocated pointer here.
		offset = uintptr(rawOffset)
		address = memmap.PtrOff(0x85b3fc, offset)
		// The base describes the operand. Only the addressed edge byte is read.
		base = unsafe.Add(address, -int(offset))
	}
	field := 6
	if site == 0x4114a8 {
		field = 7
	}
	event := tileListNativeOperand411350{Site: site, Field: field, Base: base, Address: address, Offset: offset, Width: 1, LogicalPEAddress: logical, AddressValid: true}
	tileListNativeObserveOperand411350(event)
	value := *(*uint8)(address)
	event.ValueWord, event.Complete, event.ValueValid = uint32(value), true, true
	tileListNativeObserveOperand411350(event)
	runtime.KeepAlive(base)
	return value
}
