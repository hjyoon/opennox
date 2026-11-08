package legacy

import (
	"runtime"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/server"
)

type wallAttachNativeArg410390 struct {
	Word      uint32
	Pointer   unsafe.Pointer
	IsPointer bool
}

type wallAttachNativeReply410390 struct {
	Word     uint32
	Pointer  unsafe.Pointer
	ECX, EDX uint32
}

// ValueBits contains the actual scalar load/store bits at Width. A coordinate
// read is a native int; the engine narrows it only when forming the int32 local.
// Pointer stores instead carry ValuePointer and leave ValueBits zero.
type wallAttachNativeOperand410390 struct {
	Site                                   uint32
	Field                                  int
	Base, Address                          unsafe.Pointer
	Offset                                 uintptr
	Width                                  uint32
	ValueBits                              uint64
	ValuePointer                           unsafe.Pointer
	IsPointer, Write, Complete, ValueValid bool
	AddressValid                           bool
}

type wallAttachNativeObserver410390 struct {
	Operand func(wallAttachNativeOperand410390)
	Service func(uint32, []wallAttachNativeArg410390) (wallAttachNativeReply410390, bool)
}

// The diagnostic seam is serial and reentrant. Each push owns a distinct frame,
// including when the same observer pointer is pushed twice.
type wallAttachNativeObserverFrame410390 struct {
	Observer *wallAttachNativeObserver410390
	Previous *wallAttachNativeObserverFrame410390
}

var wallAttachNativeObserverHead410390 *wallAttachNativeObserverFrame410390

func wallAttachNativeObserverCurrent410390() *wallAttachNativeObserver410390 {
	if wallAttachNativeObserverHead410390 == nil {
		return nil
	}
	return wallAttachNativeObserverHead410390.Observer
}

func pushWallAttachNativeObserver410390(observer *wallAttachNativeObserver410390) func() {
	frame := &wallAttachNativeObserverFrame410390{observer, wallAttachNativeObserverHead410390}
	wallAttachNativeObserverHead410390 = frame
	return func() {
		if wallAttachNativeObserverHead410390 != frame {
			panic("wall-attach410390 observer pop order")
		}
		wallAttachNativeObserverHead410390 = frame.Previous
	}
}

// Draw X, Draw Y, wall flags, wall client data, and wall tile low BYTE.
func wallAttachNativeOffsets410390() [5]uintptr {
	var drawable client.Drawable
	var wall server.Wall
	return [5]uintptr{
		unsafe.Offsetof(drawable.PosVec) + unsafe.Offsetof(drawable.PosVec.X),
		unsafe.Offsetof(drawable.PosVec) + unsafe.Offsetof(drawable.PosVec.Y),
		unsafe.Offsetof(wall.Flags4),
		unsafe.Offsetof(wall.ClientData),
		unsafe.Offsetof(wall.Field8),
	}
}

func wallAttachNativeAddress410390(base unsafe.Pointer, offset uintptr) unsafe.Pointer {
	if base == nil {
		// Preserve an actual Go nil load, without exposing a low unsafe.Pointer
		// to the observer/GC. Offset remains logical field metadata.
		return nil
	}
	return unsafe.Add(base, offset)
}

func wallAttachNativeObserve410390(event wallAttachNativeOperand410390) {
	if observer := wallAttachNativeObserverCurrent410390(); observer != nil && observer.Operand != nil {
		observer.Operand(event)
	}
}

//go:noinline
func wallAttachNativeLoadInt410390(address unsafe.Pointer) int {
	return *(*int)(address)
}

//go:noinline
func wallAttachNativeLoadByte410390(address unsafe.Pointer) byte {
	return *(*byte)(address)
}

//go:noinline
func wallAttachNativeStoreWord410390(address unsafe.Pointer, value uint32) {
	*(*uint32)(address) = value
}

//go:noinline
func wallAttachNativeStoreByte410390(address unsafe.Pointer, value byte) {
	*(*byte)(address) = value
}

//go:noinline
func wallAttachNativeStorePointer410390(address, value unsafe.Pointer) {
	if uintptr(address)%unsafe.Alignof(unsafe.Pointer(nil)) == 0 {
		// Actual typed Go pointer slots retain the compiler's write barrier.
		*(*unsafe.Pointer)(address) = value
	} else {
		// Unaligned diagnostic slots belong to foreign memory. C memcpy keeps
		// the full pointer and cgo checks; no uintptr pointer reconstruction.
		wallAttachNativeForeignStorePointer410390(address, value)
	}
	runtime.KeepAlive(value)
}

func wallAttachNativeReadInt410390(site uint32, field int, base unsafe.Pointer, offset uintptr) int {
	address := wallAttachNativeAddress410390(base, offset)
	event := wallAttachNativeOperand410390{Site: site, Field: field, Base: base, Address: address, Offset: offset, Width: uint32(unsafe.Sizeof(int(0))), AddressValid: true}
	wallAttachNativeObserve410390(event)
	value := wallAttachNativeLoadInt410390(address)
	event.ValueBits, event.Complete, event.ValueValid = uint64(uintptr(value)), true, true
	wallAttachNativeObserve410390(event)
	runtime.KeepAlive(base)
	return value
}

func wallAttachNativeReadByte410390(site uint32, field int, base unsafe.Pointer, offset uintptr) byte {
	address := wallAttachNativeAddress410390(base, offset)
	event := wallAttachNativeOperand410390{Site: site, Field: field, Base: base, Address: address, Offset: offset, Width: 1, AddressValid: true}
	wallAttachNativeObserve410390(event)
	value := wallAttachNativeLoadByte410390(address)
	event.ValueBits, event.Complete, event.ValueValid = uint64(value), true, true
	wallAttachNativeObserve410390(event)
	runtime.KeepAlive(base)
	return value
}

func wallAttachNativeWriteScalar410390(site uint32, field int, base unsafe.Pointer, offset uintptr, width uint32, value uint32) {
	address := wallAttachNativeAddress410390(base, offset)
	if width == 1 {
		value = uint32(byte(value))
	}
	event := wallAttachNativeOperand410390{Site: site, Field: field, Base: base, Address: address, Offset: offset, Width: width, ValueBits: uint64(value), Write: true, ValueValid: true, AddressValid: true}
	wallAttachNativeObserve410390(event)
	if width == 1 {
		wallAttachNativeStoreByte410390(address, byte(value))
	} else {
		wallAttachNativeStoreWord410390(address, value)
	}
	event.Complete = true
	wallAttachNativeObserve410390(event)
	runtime.KeepAlive(base)
}

func wallAttachNativeWritePointer410390(site uint32, field int, base unsafe.Pointer, offset uintptr, value unsafe.Pointer) {
	address := wallAttachNativeAddress410390(base, offset)
	event := wallAttachNativeOperand410390{Site: site, Field: field, Base: base, Address: address, Offset: offset, Width: uint32(unsafe.Sizeof(unsafe.Pointer(nil))), ValuePointer: value, IsPointer: true, Write: true, ValueValid: true, AddressValid: true}
	wallAttachNativeObserve410390(event)
	wallAttachNativeStorePointer410390(address, value)
	event.Complete = true
	wallAttachNativeObserve410390(event)
	runtime.KeepAlive(base)
	runtime.KeepAlive(value)
}

func wallAttachNativeService410390(site uint32, args ...wallAttachNativeArg410390) wallAttachNativeReply410390 {
	defer func() { runtime.KeepAlive(args) }()
	if observer := wallAttachNativeObserverCurrent410390(); observer != nil && observer.Service != nil {
		if reply, handled := observer.Service(site, args); handled {
			return reply
		}
	}
	return wallAttachNativeServiceDefault410390(site, args)
}

func sub410390Native(drawable unsafe.Pointer, x, y int32) unsafe.Pointer {
	var wall, polygon unsafe.Pointer
	var local [2]int32
	defer func() {
		runtime.KeepAlive(drawable)
		runtime.KeepAlive(wall)
		runtime.KeepAlive(polygon)
		runtime.KeepAlive(&local)
	}()
	word := func(value int32) wallAttachNativeArg410390 { return wallAttachNativeArg410390{Word: uint32(value)} }
	wall = wallAttachNativeService410390(0x4103a0, word(x), word(y)).Pointer
	if wall == nil {
		wall = wallAttachNativeService410390(0x4103b0, word(x), word(y)).Pointer
	}
	if wall == nil {
		return nil
	}
	offsets := wallAttachNativeOffsets410390()
	localPointer := unsafe.Pointer(&local[0])
	px := wallAttachNativeReadInt410390(0x4103c4, 0, drawable, offsets[0])
	py := wallAttachNativeReadInt410390(0x4103c7, 1, drawable, offsets[1])
	wallAttachNativeWriteScalar410390(0x4103ca, 2, localPointer, 0, 4, uint32(int32(px)))
	flags := wallAttachNativeReadByte410390(0x4103ce, 3, wall, offsets[2])
	wallAttachNativeWritePointer410390(0x4103d1, 4, wall, offsets[3], drawable)
	wallAttachNativeWriteScalar410390(0x4103dc, 5, localPointer, 4, 4, uint32(int32(py)))
	wallAttachNativeWriteScalar410390(0x4103e0, 3, wall, offsets[2], 1, uint32(flags|0x10))
	point := wallAttachNativeArg410390{Pointer: localPointer, IsPointer: true}
	polygon = wallAttachNativeService410390(0x4103e3, point, word(0)).Pointer
	if polygon == nil {
		polygon = wallAttachNativeService410390(0x4103fa, point, wallAttachNativeArg410390{Word: 0x41200000}, word(0)).Pointer
	}
	if polygon != nil {
		tile := wallAttachNativeReadByte410390(0x410406, 6, polygon, 130)
		wallAttachNativeWriteScalar410390(0x41040d, 7, wall, offsets[4], 1, uint32(tile))
	} else {
		wallAttachNativeWriteScalar410390(0x410416, 7, wall, offsets[4], 1, 1)
	}
	return polygon
}
