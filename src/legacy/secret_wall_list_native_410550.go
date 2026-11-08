package legacy

import (
	"runtime"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

type secretWallNativeArg410550 struct {
	Word      uint32
	Pointer   unsafe.Pointer
	IsPointer bool
}

type secretWallNativeReply410550 struct {
	Word     uint32
	Pointer  unsafe.Pointer
	ECX, EDX uint32
}

// Width and Offset describe actual native memory. Pointer values never travel
// through Word. Fields4/5 are reached-site local argument reads, with no claim
// that the native private stack equals the original caller stack.
type secretWallNativeOperand410550 struct {
	Site                                                 uint32
	Field                                                int
	Base, Address                                        unsafe.Pointer
	Offset                                               uintptr
	Width                                                uint32
	ValueWord                                            uint32
	ValuePointer                                         unsafe.Pointer
	IsPointer, Write, Complete, ValueValid, AddressValid bool
}

type secretWallNativeObserver410550 struct {
	Operand func(secretWallNativeOperand410550)
	Service func(uint32, []secretWallNativeArg410550) (secretWallNativeReply410550, bool)
	// Optional actual C-owned/shared native pointer cell. Nil selects the
	// production C symbol; this diagnostic cell is not a PE/memmap alias.
	HeadCell unsafe.Pointer
}

type secretWallNativeObserverFrame410550 struct {
	Observer *secretWallNativeObserver410550
	Previous *secretWallNativeObserverFrame410550
}

var secretWallNativeObserverHead410550 *secretWallNativeObserverFrame410550

func secretWallNativeObserverCurrent410550() *secretWallNativeObserver410550 {
	if secretWallNativeObserverHead410550 == nil {
		return nil
	}
	return secretWallNativeObserverHead410550.Observer
}

func pushSecretWallNativeObserver410550(observer *secretWallNativeObserver410550) func() {
	frame := &secretWallNativeObserverFrame410550{observer, secretWallNativeObserverHead410550}
	secretWallNativeObserverHead410550 = frame
	return func() {
		if secretWallNativeObserverHead410550 != frame {
			panic("secret-wall410550 observer pop order")
		}
		secretWallNativeObserverHead410550 = frame.Previous
	}
}

// Next, Wall pointer, and unsigned wall ID are distinct native fields.
func secretWallNativeOffsets410550() [3]uintptr {
	var node server.SecretWall
	var wall server.Wall
	return [3]uintptr{unsafe.Offsetof(node.Next), unsafe.Offsetof(node.Wall), unsafe.Offsetof(wall.Field10)}
}

func secretWallNativeSelectedHeadCell410550() unsafe.Pointer {
	if observer := secretWallNativeObserverCurrent410550(); observer != nil && observer.HeadCell != nil {
		return observer.HeadCell
	}
	return secretWallNativeHeadCell410550()
}

func secretWallNativeAddress410550(base unsafe.Pointer, offset uintptr) unsafe.Pointer {
	if base == nil {
		// Do not expose a manufactured low pointer to Go/GC. Offset still
		// identifies the semantic field; physical nil-EA equivalence is false.
		return nil
	}
	return unsafe.Add(base, offset)
}

func secretWallNativeObserve410550(event secretWallNativeOperand410550) {
	if observer := secretWallNativeObserverCurrent410550(); observer != nil && observer.Operand != nil {
		observer.Operand(event)
	}
}

//go:noinline
func secretWallNativeLoadPointer410550(address unsafe.Pointer) unsafe.Pointer {
	if uintptr(address)%unsafe.Alignof(unsafe.Pointer(nil)) == 0 {
		return *(*unsafe.Pointer)(address)
	}
	// Unaligned storage is foreign memory only, never a Go pointer field.
	return secretWallNativeForeignLoadPointer410550(address)
}

//go:noinline
func secretWallNativeLoadWord410550(address unsafe.Pointer) uint16 {
	if uintptr(address)%unsafe.Alignof(uint16(0)) == 0 {
		return *(*uint16)(address)
	}
	return secretWallNativeForeignLoadWord410550(address)
}

//go:noinline
func secretWallNativeStorePointer410550(address, value unsafe.Pointer) {
	if uintptr(address)%unsafe.Alignof(unsafe.Pointer(nil)) == 0 {
		// Retain Go's pointer write barrier for actual Go-owned slots.
		*(*unsafe.Pointer)(address) = value
	} else {
		secretWallNativeForeignStorePointer410550(address, value)
	}
	runtime.KeepAlive(value)
}

func secretWallNativeReadPointer410550(site uint32, field int, base unsafe.Pointer, offset uintptr) unsafe.Pointer {
	address := secretWallNativeAddress410550(base, offset)
	event := secretWallNativeOperand410550{Site: site, Field: field, Base: base, Address: address, Offset: offset, Width: uint32(unsafe.Sizeof(unsafe.Pointer(nil))), IsPointer: true, AddressValid: true}
	secretWallNativeObserve410550(event)
	value := secretWallNativeLoadPointer410550(address)
	event.ValuePointer, event.Complete, event.ValueValid = value, true, true
	secretWallNativeObserve410550(event)
	runtime.KeepAlive(base)
	runtime.KeepAlive(value)
	return value
}

func secretWallNativeReadWord410550(site uint32, field int, base unsafe.Pointer, offset uintptr) uint16 {
	address := secretWallNativeAddress410550(base, offset)
	event := secretWallNativeOperand410550{Site: site, Field: field, Base: base, Address: address, Offset: offset, Width: 2, AddressValid: true}
	secretWallNativeObserve410550(event)
	value := secretWallNativeLoadWord410550(address)
	event.ValueWord, event.Complete, event.ValueValid = uint32(value), true, true
	secretWallNativeObserve410550(event)
	runtime.KeepAlive(base)
	return value
}

func secretWallNativeWritePointer410550(site uint32, field int, base unsafe.Pointer, offset uintptr, value unsafe.Pointer) {
	address := secretWallNativeAddress410550(base, offset)
	event := secretWallNativeOperand410550{Site: site, Field: field, Base: base, Address: address, Offset: offset, Width: uint32(unsafe.Sizeof(unsafe.Pointer(nil))), ValuePointer: value, IsPointer: true, Write: true, ValueValid: true, AddressValid: true}
	secretWallNativeObserve410550(event)
	secretWallNativeStorePointer410550(address, value)
	event.Complete = true
	secretWallNativeObserve410550(event)
	runtime.KeepAlive(base)
	runtime.KeepAlive(value)
}

func secretWallNativeService410550(site uint32, args ...secretWallNativeArg410550) secretWallNativeReply410550 {
	defer func() { runtime.KeepAlive(args) }()
	if observer := secretWallNativeObserverCurrent410550(); observer != nil && observer.Service != nil {
		if reply, handled := observer.Service(site, args); handled {
			return reply
		}
	}
	return secretWallNativeServiceDefault410550(site, args)
}

func secretWallLookupNative410550(rawArgument uint32) unsafe.Pointer {
	node := secretWallNativeService410550(0x410551).Pointer
	if node == nil {
		return nil
	}
	// Original41055A reads only the WORD, and only after First is nonnil.
	argument := uint16(rawArgument)
	id := secretWallNativeReadWord410550(0x41055a, 4, unsafe.Pointer(&argument), 0)
	offsets := secretWallNativeOffsets410550()
	for {
		wall := secretWallNativeReadPointer410550(0x41055f, 2, node, offsets[1])
		currentID := secretWallNativeReadWord410550(0x410562, 3, wall, offsets[2])
		if currentID == id {
			// This actual second load must survive after-read mutations.
			return secretWallNativeReadPointer410550(0x410579, 2, node, offsets[1])
		}
		node = secretWallNativeService410550(0x410569, secretWallNativeArg410550{Pointer: node, IsPointer: true}).Pointer
		if node == nil {
			return nil
		}
	}
}

func secretWallClearNative410730() unsafe.Pointer {
	headCell := secretWallNativeSelectedHeadCell410550()
	defer func() { runtime.KeepAlive(headCell) }()
	node := secretWallNativeReadPointer410550(0x410730, 0, headCell, 0)
	if node == nil {
		secretWallNativeWritePointer410550(0x410753, 0, headCell, 0, nil)
		return nil
	}
	offsets := secretWallNativeOffsets410550()
	for {
		next := secretWallNativeReadPointer410550(0x41073a, 1, node, offsets[0])
		secretWallNativeService410550(0x41073d, secretWallNativeArg410550{Pointer: node, IsPointer: true})
		node = next
		if node == nil {
			secretWallNativeWritePointer410550(0x41074b, 0, headCell, 0, nil)
			return nil
		}
	}
}

func secretWallRemoveNative4107A0(argument unsafe.Pointer) unsafe.Pointer {
	headCell := secretWallNativeSelectedHeadCell410550()
	var previous, node, target unsafe.Pointer
	defer func() {
		runtime.KeepAlive(headCell)
		runtime.KeepAlive(argument)
		runtime.KeepAlive(previous)
		runtime.KeepAlive(node)
		runtime.KeepAlive(target)
	}()
	node = secretWallNativeReadPointer410550(0x4107a0, 0, headCell, 0)
	if node == nil {
		return nil
	}
	// Original4107AD reads the argument only after a nonnil initial head.
	target = secretWallNativeReadPointer410550(0x4107ad, 5, unsafe.Pointer(&argument), 0)
	for node != target {
		previous = node
		node = secretWallNativeService410550(0x4107b8, secretWallNativeArg410550{Pointer: node, IsPointer: true}).Pointer
		if node == nil {
			return nil
		}
	}
	offsets := secretWallNativeOffsets410550()
	// One live read is cached before Next; its after-read effect must not
	// silently change the comparison into another head load.
	liveHead := secretWallNativeReadPointer410550(0x4107c7, 0, headCell, 0)
	if node == liveHead {
		next := secretWallNativeService410550(0x4107d2, secretWallNativeArg410550{Pointer: node, IsPointer: true}).Pointer
		secretWallNativeWritePointer410550(0x4107da, 0, headCell, 0, next)
		secretWallNativeService410550(0x4107e0, secretWallNativeArg410550{Pointer: target, IsPointer: true})
	} else {
		next := secretWallNativeService410550(0x4107eb, secretWallNativeArg410550{Pointer: node, IsPointer: true}).Pointer
		secretWallNativeWritePointer410550(0x4107f3, 1, previous, offsets[0], next)
		secretWallNativeService410550(0x4107f6, secretWallNativeArg410550{Pointer: target, IsPointer: true})
	}
	// Current C returns this typed cached matched pointer. Original post-Free
	// EAX remains diagnostic/unqualified; it never manufactures this pointer.
	return node
}
