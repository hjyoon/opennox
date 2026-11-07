package legacy

import (
	"runtime"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

type rawSkeletonArgument4E23C0 struct {
	Pointer   unsafe.Pointer
	Word      uint32
	IsPointer bool
}

type rawSkeletonService4E23C0 struct {
	Site      uint32
	Arguments []rawSkeletonArgument4E23C0
}

type rawSkeletonOperand4E23C0 struct {
	Site                   uint32
	Base, Address, Pointer unsafe.Pointer
	Offset                 uintptr
	EffectiveAddress       uintptr // Named-field EA; a Go nil trap may report address zero.
	Width                  int
	Word                   uint32
	IsPointer, Complete    bool
}

// A diagnostic observer supplies independently chosen external service inputs.
// Nil dispatch uses the actual product endpoints. It never supplies a caller
// result, branch schedule or replacement memory load.
type rawSkeletonObserver4E23C0 struct {
	Operand func(rawSkeletonOperand4E23C0)
	Service func(rawSkeletonService4E23C0) (uint32, bool)
	Frame   func() uint8
}

type rawSkeletonObserverFrame4E23C0 struct {
	Value  *rawSkeletonObserver4E23C0
	Parent *rawSkeletonObserverFrame4E23C0
}

var rawSkeletonObserverStack4E23C0 *rawSkeletonObserverFrame4E23C0

func pushRawSkeletonObserver4E23C0(value *rawSkeletonObserver4E23C0) func() {
	frame := &rawSkeletonObserverFrame4E23C0{Value: value, Parent: rawSkeletonObserverStack4E23C0}
	rawSkeletonObserverStack4E23C0 = frame
	return func() {
		if rawSkeletonObserverStack4E23C0 != frame {
			panic("raw Skeleton observer LIFO")
		}
		rawSkeletonObserverStack4E23C0 = frame.Parent
	}
}

func rawSkeletonObserveOperand4E23C0(event rawSkeletonOperand4E23C0) {
	if frame := rawSkeletonObserverStack4E23C0; frame != nil && frame.Value != nil && frame.Value.Operand != nil {
		frame.Value.Operand(event)
	}
}

func rawSkeletonOffsets4E23C0() [7]uintptr {
	return [7]uintptr{
		unsafe.Offsetof(server.Object{}.PosVec),
		unsafe.Offsetof(server.Object{}.PrevPos),
		unsafe.Offsetof(server.Object{}.PrevPos) + 4,
		unsafe.Offsetof(server.Object{}.Direction1),
		unsafe.Offsetof(server.Object{}.UpdateData),
		unsafe.Offsetof(server.MonsterUpdateData{}.Field120_0),
		unsafe.Offsetof(server.MonsterUpdateData{}.Field120_1),
	}
}

func rawSkeletonRead4E23C0(base unsafe.Pointer, offset uintptr, width int, site uint32, pointer bool) rawSkeletonOperand4E23C0 {
	address := base
	if base != nil {
		address = unsafe.Add(base, offset)
	}
	// Keep invalid small addresses out of Go pointer slots. Nil retains its
	// actual Go fault probe; the numeric named-field EA is separate metadata.
	event := rawSkeletonOperand4E23C0{Site: site, Base: base, Address: address, Offset: offset, EffectiveAddress: uintptr(base) + offset, Width: width, IsPointer: pointer}
	rawSkeletonObserveOperand4E23C0(event)
	if pointer {
		if width != int(unsafe.Sizeof(unsafe.Pointer(nil))) {
			panic("raw Skeleton full native pointer width")
		}
		if uintptr(address)%unsafe.Alignof(unsafe.Pointer(nil)) == 0 {
			event.Pointer = *(*unsafe.Pointer)(address)
		} else {
			// Unaligned foreign input still transports every native pointer
			// bit; it is never reconstructed from an original 32-bit word.
			event.Pointer = rawSkeletonForeignPointer4E23C0(*(*uintptr)(address))
		}
	} else {
		switch width {
		case 1:
			event.Word = uint32(*(*uint8)(address))
		case 2:
			event.Word = uint32(*(*uint16)(address))
		case 4:
			event.Word = *(*uint32)(address)
		default:
			panic("raw Skeleton scalar read width")
		}
	}
	event.Complete = true
	rawSkeletonObserveOperand4E23C0(event)
	runtime.KeepAlive(base)
	return event
}

func rawSkeletonCall4E23C0(site uint32, args ...rawSkeletonArgument4E23C0) uint32 {
	call := rawSkeletonService4E23C0{Site: site, Arguments: args}
	if frame := rawSkeletonObserverStack4E23C0; frame != nil && frame.Value != nil && frame.Value.Service != nil {
		if result, handled := frame.Value.Service(call); handled {
			return result
		}
	}
	return rawSkeletonFallbackService4E23C0(call)
}

func rawSkeletonFrame4E23C0() uint8 {
	if frame := rawSkeletonObserverStack4E23C0; frame != nil && frame.Value != nil && frame.Value.Frame != nil {
		return frame.Value.Frame()
	}
	return uint8(gameFrameHook())
}

// Complete raw 004E23C0 caller. Keep the original lazy MOV order, signed WORD
// direction, AL bit test, pre-Action update pointer cache and signed EAX return.
// The ordinary registered bool SkeletonDamage route has separate admission.
func rawSkeletonDamageNative4E23C0(target, source, weapon unsafe.Pointer, damage, typ int32) int32 {
	defer runtime.KeepAlive(target)
	defer runtime.KeepAlive(source)
	defer runtime.KeepAlive(weapon)
	ptr := func(p unsafe.Pointer) rawSkeletonArgument4E23C0 {
		return rawSkeletonArgument4E23C0{Pointer: p, IsPointer: true}
	}
	word := func(v uint32) rawSkeletonArgument4E23C0 { return rawSkeletonArgument4E23C0{Word: v} }
	if rawSkeletonCall4E23C0(0x4e23ce, ptr(target), word(23)) != 0 {
		if rawSkeletonFrame4E23C0()&3 == 0 {
			rawSkeletonCall4E23C0(0x4e23ea, word(71), ptr(target), word(0), word(0))
		}
		return 1
	}
	if source != nil {
		offsets := rawSkeletonOffsets4E23C0()
		selected := weapon
		xSite, ySite := uint32(0x4e2413), uint32(0x4e2416)
		if selected == nil {
			selected = source
			xSite, ySite = 0x4e2423, 0x4e2426
		}
		var point [2]uint32
		point[0] = rawSkeletonRead4E23C0(selected, offsets[1], 4, xSite, false).Word
		point[1] = rawSkeletonRead4E23C0(selected, offsets[2], 4, ySite, false).Word
		direction := int32(int16(rawSkeletonRead4E23C0(target, offsets[3], 2, 0x4e2431, false).Word))
		position := unsafe.Add(target, offsets[0]) // LEA only; no target coordinate read.
		if rawSkeletonCall4E23C0(0x4e243f, ptr(position), word(uint32(direction)), ptr(unsafe.Pointer(&point[0])))&1 != 0 {
			update := rawSkeletonRead4E23C0(target, offsets[4], int(unsafe.Sizeof(unsafe.Pointer(nil))), 0x4e244b, true).Pointer
			if rawSkeletonCall4E23C0(0x4e2452, ptr(target)) == 21 {
				first := rawSkeletonRead4E23C0(update, offsets[5], 1, 0x4e245f, false).Word
				second := rawSkeletonRead4E23C0(update, offsets[6], 1, 0x4e2465, false).Word
				if second > first>>1 {
					rawSkeletonCall4E23C0(0x4e247b, word(878), ptr(target), word(0), word(0))
					return 1
				}
			}
		}
		runtime.KeepAlive(point)
	}
	return int32(rawSkeletonCall4E23C0(0x4e249d, ptr(target), ptr(source), ptr(weapon), word(uint32(damage)), word(uint32(typ))))
}
