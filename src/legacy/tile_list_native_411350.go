package legacy

import (
	"runtime"
	"unsafe"
)

// Matches defs.h: four DWORDs, native Next at16; size24 on ARM64.
type tileListNativeNode411350 struct {
	Value, Preserved, Edge, Variation int32
	Next                              *tileListNativeNode411350
}

type tileListNativeOperand411350 struct {
	Site                                          uint32
	Field                                         int // value0, edge1, variation2, next3, X4, Y5, edge-byte0=6, edge-byte1=7
	Base, Address                                 unsafe.Pointer
	Offset                                        uintptr // actual native offset, not a PE-pointer encoding
	Width                                         uint32
	ValueWord                                     uint32
	ValuePointer                                  unsafe.Pointer
	LogicalPEAddress                              uint32 // metadata only; never converted into a Go pointer
	IsPointer, Complete, ValueValid, AddressValid bool
}

type tileListNativeCall411350 struct {
	Site, Target uint32
	// ABI-indexed slots: 490 words0/1; 3A0 arg0=Pointer, arg1=Words[1].
	Words    [2]uint32
	Pointer  unsafe.Pointer
	Result   int32
	Complete bool
}

type tileListNativeObserver411350 struct {
	Operand func(tileListNativeOperand411350)
	Call    func(tileListNativeCall411350) // notification; cannot replace either actual helper body
	// Optional diagnostic resolver names an actual C-owned/guarded BYTE address.
	// Default nil selects the actual85B3FC owner with the raw32 offset below.
	EdgeByteAddress func(site, rawIndex, rawOffset, logicalPEAddress uint32) unsafe.Pointer
}

// Proposed bridge functions, to be implemented/reviewed after root chooses scope:
// tileListNativeReadWord411350(site,field,base,offset) reads exactlyDWORD4.
// tileListNativeReadPointer411350(site,field,base,offset) reads nativePTR8.
// tileListNativeReadEdgeByte411350(site,index,offset) resolves then reads exactlyBYTE1.
// Each emits BEFORE with ValueValid=false, then actual load, then COMPLETE.
// No extra preload; a fault has no fabricated COMPLETE/value/AFTER event.
// tileListNativeObserveCall411350(event) is optional, serial and LIFO-scoped.

func sub411350Native(head, xy unsafe.Pointer, initial int32) int32 {
	defer func() { runtime.KeepAlive(head); runtime.KeepAlive(xy) }()
	if head == nil {
		return initial
	} // no eager coordinate/node/EDGE access
	value := initial
	var layout tileListNativeNode411350
	for node := head; node != nil; {
		variation := int32(tileListNativeReadWord411350(0x411363, 2, node, unsafe.Offsetof(layout.Variation)))
		edge := int32(tileListNativeReadWord411350(0x411366, 1, node, unsafe.Offsetof(layout.Edge)))
		call := tileListNativeCall411350{Site: 0x41136b, Target: 0x411490, Words: [2]uint32{uint32(edge), uint32(variation)}}
		tileListNativeObserveCall411350(call)
		selector := sub411490Native(edge, variation)
		call.Result, call.Complete = selector, true
		tileListNativeObserveCall411350(call)
		call = tileListNativeCall411350{Site: 0x411372, Target: 0x4113a0, Words: [2]uint32{0, uint32(selector)}, Pointer: xy}
		tileListNativeObserveCall411350(call)
		accepted := sub4113A0Native(xy, selector)
		call.Result, call.Complete = accepted, true
		tileListNativeObserveCall411350(call)
		if accepted != 0 {
			value = int32(tileListNativeReadWord411350(0x41137e, 0, node, unsafe.Offsetof(layout.Value)))
		}
		// Load after both actual calls and the optional lazy value read.
		node = tileListNativeReadPointer411350(0x411380, 3, node, unsafe.Offsetof(layout.Next))
	}
	return value
}

func sub4113A0Native(xy unsafe.Pointer, selector int32) int32 {
	defer runtime.KeepAlive(xy)
	x := int32(tileListNativeReadWord411350(0x4113a7, 4, xy, 0))
	y := int32(tileListNativeReadWord411350(0x4113a9, 5, xy, 4))
	ge, le := x >= y, x <= y
	anti := int32(uint32(46) - uint32(y)) // SUB DWORD, then signed comparisons
	geAnti, leAnti := x >= anti, x <= anti
	if uint32(selector) > 11 {
		return 0
	} // after both coordinate loads
	var accepted bool
	switch selector {
	case 0:
		accepted = le && geAnti
	case 1:
		accepted = le
	case 2:
		accepted = le && leAnti
	case 3:
		accepted = geAnti
	case 4:
		accepted = leAnti
	case 5:
		accepted = geAnti && ge
	case 6:
		accepted = ge
	case 7:
		accepted = ge && leAnti
	case 8:
		accepted = le || geAnti
	case 9:
		accepted = le || leAnti
	case 10:
		accepted = ge || leAnti
	case 11:
		accepted = ge || geAnti
	}
	if accepted {
		return 1
	}
	return 0
}

func sub411490Native(edge, variation int32) int32 {
	offset := uint32(28696) + uint32(edge)*uint32(60)
	a := int32(tileListNativeReadEdgeByte411350(0x4114a2, uint32(edge), offset))
	b := int32(tileListNativeReadEdgeByte411350(0x4114a8, uint32(edge), offset+uint32(1)))
	if a == 3 && b == 3 {
		return variation
	}
	if variation == 0 {
		return 0
	}
	if variation <= a-2 {
		return 1
	}
	if variation == a-1 {
		return 2
	}
	bound := a + 2*b - 4
	if variation < bound {
		return int32((uint32(a)^uint32(variation))&1) + 3
	}
	if variation == bound {
		return 5
	}
	if variation <= 2*(b+a)-6 {
		return 6
	}
	if variation == 2*(b+a)-5 {
		return 7
	}
	delta := 2 * (int32(6) - b - a)
	return int32(uint32(variation) + uint32(delta)) // full DWORD wrap
}
