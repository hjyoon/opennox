package legacy

/*
#include <stdint.h>
// Opaque full-width transport only; never dereference here.
static void* rdd_native_foreign_pointer_4e0b30(uintptr_t bits){return(void*)bits;}
*/
import "C"

import (
	"runtime"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

// Native observer plumbing only. There is no parent admission, damage branch,
// arithmetic result or root return implementation in this file. Go heap GC and
// concurrent dispatch remain separate gates from a foreign fixture graph.
type rawDefaultDamageNativeValue4E0B30 struct {
	Word    uint32
	Pointer unsafe.Pointer
}
type rawDefaultDamageNativeOperand4E0B30 struct {
	Site            uint32
	Base, Address   unsafe.Pointer
	Field, Width    int
	Write, Complete bool
	Value           rawDefaultDamageNativeValue4E0B30
}
type rawDefaultDamageNativeArgument4E0B30 struct {
	Kind        string // word, object, damage, string, coordinates
	Word        uint32
	Pointer     unsafe.Pointer
	Damage      *int32
	String      string
	Coordinates [4]uint32
}
type rawDefaultDamageNativeService4E0B30 struct {
	Indirect  bool
	Site      uint32
	Callback  unsafe.Pointer
	Arguments []rawDefaultDamageNativeArgument4E0B30
}
type rawDefaultDamageNativeReply4E0B30 struct {
	Word         uint32
	Pointer      unsafe.Pointer
	FloatFormat  string
	FloatPayload [10]byte
}
type rawDefaultDamageNativeCall4E0B30 struct {
	Site, Entry uint32
	Arguments   []rawDefaultDamageNativeArgument4E0B30
	Complete    bool
	Result      uint32
}
type rawDefaultDamageNativeCaller4E0B30 struct {
	Site            uint32
	Pointer         *int32
	Value           uint32
	Complete, Entry bool
}
type rawDefaultDamageNativeObserver4E0B30 struct {
	ControlWord    uint16
	Caller         func(rawDefaultDamageNativeCaller4E0B30)
	Operand        func(rawDefaultDamageNativeOperand4E0B30)
	Service        func(rawDefaultDamageNativeService4E0B30) (rawDefaultDamageNativeReply4E0B30, bool)
	BeforeCallback func(site uint32, target unsafe.Pointer)
	Call           func(rawDefaultDamageNativeCall4E0B30)
	Frame          func(site uint32, width int) uint32
}
type rawDefaultDamageNativeObserverFrame4E0B30 struct {
	Value  *rawDefaultDamageNativeObserver4E0B30
	Parent *rawDefaultDamageNativeObserverFrame4E0B30
}

var rawDefaultDamageNativeObserverStack4E0B30 *rawDefaultDamageNativeObserverFrame4E0B30

func rawDefaultDamagePushNativeObserver4E0B30(next *rawDefaultDamageNativeObserver4E0B30) func() {
	frame := &rawDefaultDamageNativeObserverFrame4E0B30{Value: next, Parent: rawDefaultDamageNativeObserverStack4E0B30}
	rawDefaultDamageNativeObserverStack4E0B30 = frame
	return func() {
		if rawDefaultDamageNativeObserverStack4E0B30 != frame {
			panic("native damage observer LIFO")
		}
		rawDefaultDamageNativeObserverStack4E0B30 = frame.Parent
	}
}
func rawDefaultDamageObserveNativeOperand4E0B30(event rawDefaultDamageNativeOperand4E0B30) {
	if frame := rawDefaultDamageNativeObserverStack4E0B30; frame != nil && frame.Value != nil && frame.Value.Operand != nil {
		frame.Value.Operand(event)
	}
}
func rawDefaultDamageReadNativeOperand4E0B30(base unsafe.Pointer, field, width int, site uint32, pointer bool) (value rawDefaultDamageNativeValue4E0B30) {
	offsets := rawDefaultDamageNativeOffsets4E0B30()
	if field < 0 || field >= len(offsets) {
		panic("undeclared native damage field")
	}
	offset := offsets[field]
	address := base
	if base != nil {
		address = unsafe.Add(base, offset)
	}
	event := rawDefaultDamageNativeOperand4E0B30{Site: site, Base: base, Address: address, Field: field, Width: width}
	rawDefaultDamageObserveNativeOperand4E0B30(event)
	if pointer {
		if width != int(unsafe.Sizeof(uintptr(0))) {
			panic("full native pointer load required")
		}
		if uintptr(address)%unsafe.Alignof(unsafe.Pointer(nil)) == 0 {
			value.Pointer = *(*unsafe.Pointer)(address)
		} else {
			bits := *(*uintptr)(address)
			value.Pointer = C.rdd_native_foreign_pointer_4e0b30(C.uintptr_t(bits))
		}
	} else {
		switch width {
		case 1:
			value.Word = uint32(*(*uint8)(address))
		case 2:
			value.Word = uint32(*(*uint16)(address))
		case 4:
			value.Word = *(*uint32)(address)
		default:
			panic("native damage scalar load width")
		}
	}
	event.Complete = true
	event.Value = value
	rawDefaultDamageObserveNativeOperand4E0B30(event)
	runtime.KeepAlive(base)
	return value
}
func rawDefaultDamageWriteNativeOperand4E0B30(base unsafe.Pointer, field, width int, site uint32, pointer bool, value rawDefaultDamageNativeValue4E0B30) {
	offsets := rawDefaultDamageNativeOffsets4E0B30()
	if field < 0 || field >= len(offsets) {
		panic("undeclared native damage field")
	}
	offset := offsets[field]
	address := base
	if base != nil {
		address = unsafe.Add(base, offset)
	}
	if !pointer {
		switch width {
		case 1:
			value.Word = uint32(uint8(value.Word))
		case 2:
			value.Word = uint32(uint16(value.Word))
		case 4:
		default:
			panic("native damage scalar store width")
		}
	}
	event := rawDefaultDamageNativeOperand4E0B30{Site: site, Base: base, Address: address, Field: field, Width: width, Write: true, Value: value}
	rawDefaultDamageObserveNativeOperand4E0B30(event)
	if pointer {
		if width != int(unsafe.Sizeof(uintptr(0))) {
			panic("full native pointer store required")
		}
		if uintptr(address)%unsafe.Alignof(unsafe.Pointer(nil)) == 0 {
			*(*unsafe.Pointer)(address) = value.Pointer
		} else {
			*(*uintptr)(address) = uintptr(value.Pointer)
		}
	} else {
		switch width {
		case 1:
			*(*uint8)(address) = uint8(value.Word)
		case 2:
			*(*uint16)(address) = uint16(value.Word)
		case 4:
			*(*uint32)(address) = value.Word
		}
	}
	event.Complete = true
	rawDefaultDamageObserveNativeOperand4E0B30(event)
	runtime.KeepAlive(base)
	runtime.KeepAlive(value.Pointer)
}
func rawDefaultDamageDispatchObservedService4E0B30(call rawDefaultDamageNativeService4E0B30) (rawDefaultDamageNativeReply4E0B30, bool) {
	frame := rawDefaultDamageNativeObserverStack4E0B30
	if frame == nil || frame.Value == nil || frame.Value.Service == nil {
		return rawDefaultDamageNativeReply4E0B30{}, false
	}
	return frame.Value.Service(call)
}
func rawDefaultDamageBeforeNativeCallback4E0B30(site uint32, target unsafe.Pointer) {
	if frame := rawDefaultDamageNativeObserverStack4E0B30; frame != nil && frame.Value != nil && frame.Value.BeforeCallback != nil {
		frame.Value.BeforeCallback(site, target)
	}
}

func rawDefaultDamageObserveNativeCall4E0B30(call rawDefaultDamageNativeCall4E0B30) {
	if frame := rawDefaultDamageNativeObserverStack4E0B30; frame != nil && frame.Value != nil && frame.Value.Call != nil {
		frame.Value.Call(call)
	}
}
func rawDefaultDamageInvokeAttack4E0B30(site uint32, source, weapon unsafe.Pointer) int32 {
	call := rawDefaultDamageNativeCall4E0B30{Site: site, Entry: 0x4e1400, Arguments: []rawDefaultDamageNativeArgument4E0B30{{Kind: "object", Pointer: source}, {Kind: "object", Pointer: weapon}}}
	rawDefaultDamageObserveNativeCall4E0B30(call)
	result := damagePredicateRawAttackNative4E1400(source, weapon)
	call.Complete = true
	call.Result = uint32(result)
	rawDefaultDamageObserveNativeCall4E0B30(call)
	return result
}
func rawDefaultDamageInvokeFriendly4E0B30(site uint32, weapon unsafe.Pointer) int32 {
	call := rawDefaultDamageNativeCall4E0B30{Site: site, Entry: 0x4e1470, Arguments: []rawDefaultDamageNativeArgument4E0B30{{Kind: "object", Pointer: weapon}}}
	rawDefaultDamageObserveNativeCall4E0B30(call)
	result := damagePredicateRawFriendlyNative4E1470(weapon)
	call.Complete = true
	call.Result = uint32(result)
	rawDefaultDamageObserveNativeCall4E0B30(call)
	return result
}
func rawDefaultDamageObserveConversion4E0B30(site uint32, complete bool, float32Bits, result uint32) {
	rawDefaultDamageObserveNativeCall4E0B30(rawDefaultDamageNativeCall4E0B30{Site: site, Entry: 0x419a70, Arguments: []rawDefaultDamageNativeArgument4E0B30{{Kind: "word", Word: float32Bits}}, Complete: complete, Result: result})
}
func rawDefaultDamageObservedFrame4E0B30(site uint32, width int) (uint32, bool) {
	frame := rawDefaultDamageNativeObserverStack4E0B30
	if frame == nil || frame.Value == nil || frame.Value.Frame == nil {
		return 0, false
	}
	return frame.Value.Frame(site, width), true
}

// The CW is an independent original execution input, not an expected result.
// A nil observer leaves the real startup/default arithmetic policy to the engine.
func rawDefaultDamageObservedControlWord4E0B30() (uint16, bool) {
	frame := rawDefaultDamageNativeObserverStack4E0B30
	if frame == nil || frame.Value == nil {
		return 0, false
	}
	cw := frame.Value.ControlWord
	if cw != 0x0e3f && cw != 0x023f {
		panic("undeclared original damage FPU control input")
	}
	return cw, true
}

// Named Go field offsets; fixture C23 and Go-derived monster4 correspondences
// are independently gated by the frozen initial controls. No POSIX dependency.
func rawDefaultDamageNativeOffsets4E0B30() [27]uintptr {
	return [27]uintptr{
		unsafe.Offsetof(server.Object{}.TypeInd),
		unsafe.Offsetof(server.Object{}.ObjClass),
		unsafe.Offsetof(server.Object{}.ObjSubClass),
		unsafe.Offsetof(server.Object{}.ObjFlags),
		unsafe.Offsetof(server.Object{}.PosVec),
		unsafe.Offsetof(server.Object{}.PosVec) + 4,
		unsafe.Offsetof(server.Object{}.PrevPos),
		unsafe.Offsetof(server.Object{}.PrevPos) + 4,
		unsafe.Offsetof(server.Object{}.ObjOwner),
		unsafe.Offsetof(server.Object{}.Obj130),
		unsafe.Offsetof(server.Object{}.Field131),
		unsafe.Offsetof(server.Object{}.Pos132),
		unsafe.Offsetof(server.Object{}.Pos132) + 4,
		unsafe.Offsetof(server.Object{}.Frame134),
		unsafe.Offsetof(server.Object{}.HealthData),
		unsafe.Offsetof(server.Object{}.Damage),
		unsafe.Offsetof(server.Object{}.DamageSound),
		unsafe.Offsetof(server.Object{}.UseData) + unsafe.Offsetof(server.UseDataPtr{}.Ptr),
		unsafe.Offsetof(server.Object{}.UpdateData),
		unsafe.Offsetof(server.PlayerUpdateData{}.State),
		unsafe.Offsetof(server.PlayerUpdateData{}.Field40_0),
		unsafe.Offsetof(server.MonsterUpdateData{}.StatusFlags),
		unsafe.Offsetof(server.MonsterUpdateData{}.Field523_2),
		unsafe.Offsetof(server.MonsterUpdateData{}.Field546),
		unsafe.Offsetof(server.MonsterUpdateData{}.Field547),
		unsafe.Offsetof(server.WandUseData{}.Flags),
		32,
	}
}

func rawDefaultDamageObserveCaller4E0B30(cell *int32, site uint32, complete, entry bool) {
	frame := rawDefaultDamageNativeObserverStack4E0B30
	if frame == nil || frame.Value == nil || frame.Value.Caller == nil {
		return
	}
	value := uint32(*cell)
	frame.Value.Caller(rawDefaultDamageNativeCaller4E0B30{Site: site, Pointer: cell, Value: value, Complete: complete, Entry: entry})
	runtime.KeepAlive(cell)
}

// Native literal bits are independently provided by the implementation and
// returned unchanged for its arithmetic. This is a semantic constant-input
// observation, not a PE-image load or a C-owned native raw-memory probe.
func rawDefaultDamageNativeConstant4E0B30(site, bits uint32) uint32 {
	event := rawDefaultDamageNativeOperand4E0B30{Site: site, Field: -4, Width: 4}
	rawDefaultDamageObserveNativeOperand4E0B30(event)
	event.Complete = true
	event.Value.Word = bits
	rawDefaultDamageObserveNativeOperand4E0B30(event)
	return bits
}
