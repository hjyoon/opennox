package legacy

import (
	"runtime"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

// Cache is a real native pointer cell, not the packed four-byte PE slot.
// Only serial diagnostics replace its owner or provide external lookup inputs.
type material7Observer4133D0 struct {
	Cache       **server.ModifierEff
	Operand     func(material7Operand4133D0)
	ResolveID   func(name string) int32
	ResolveDesc func(id int32) *server.ModifierEff
}
type material7Operand4133D0 struct {
	Site                                   uint32
	Field                                  int // cache=0, Object.InitData=1, Object.ObjClass=2, material slot=3
	Base, Address, Pointer                 unsafe.Pointer
	Offset, EffectiveAddress               uintptr
	Width                                  int
	Word                                   uint32
	IsPointer, Write, Complete, ValueValid bool
}
type material7ObserverFrame4133D0 struct {
	Value    *material7Observer4133D0
	Previous *material7ObserverFrame4133D0
}

var material7ObserverStack4133D0 *material7ObserverFrame4133D0

func pushMaterial7Observer4133D0(value *material7Observer4133D0) func() {
	frame := &material7ObserverFrame4133D0{Value: value, Previous: material7ObserverStack4133D0}
	material7ObserverStack4133D0 = frame
	return func() {
		if material7ObserverStack4133D0 != frame {
			panic("Material7 observer LIFO")
		}
		material7ObserverStack4133D0 = frame.Previous
	}
}
func material7Current4133D0() *material7Observer4133D0 {
	if frame := material7ObserverStack4133D0; frame != nil {
		return frame.Value
	}
	return nil
}
func material7Observe4133D0(event material7Operand4133D0) {
	if value := material7Current4133D0(); value != nil && value.Operand != nil {
		value.Operand(event)
	}
}
func material7Offsets4133D0() [3]uintptr {
	return [3]uintptr{unsafe.Offsetof(server.Object{}.InitData), unsafe.Offsetof(server.Object{}.ObjClass), unsafe.Offsetof(server.ModifierInitData{}.Modifiers) + unsafe.Sizeof((*server.ModifierEff)(nil))}
}
func material7ReadPointer4133D0(base unsafe.Pointer, offset uintptr, site uint32, field int) unsafe.Pointer {
	address := base
	if base != nil {
		address = unsafe.Add(base, offset)
	}
	event := material7Operand4133D0{Site: site, Field: field, Base: base, Address: address, Offset: offset, EffectiveAddress: uintptr(base) + offset, Width: int(unsafe.Sizeof(unsafe.Pointer(nil))), IsPointer: true}
	material7Observe4133D0(event)
	if uintptr(address)%unsafe.Alignof(unsafe.Pointer(nil)) == 0 {
		event.Pointer = *(*unsafe.Pointer)(address)
	} else {
		event.Pointer = material7ForeignPointer4133D0(*(*uintptr)(address))
	}
	event.Complete = true
	event.ValueValid = true
	material7Observe4133D0(event)
	runtime.KeepAlive(base)
	return event.Pointer
}
func material7ReadClass4133D0(item unsafe.Pointer, offset uintptr) uint32 {
	address := item
	if item != nil {
		address = unsafe.Add(item, offset)
	}
	event := material7Operand4133D0{Site: 0x4133fd, Field: 2, Base: item, Address: address, Offset: offset, EffectiveAddress: uintptr(item) + offset, Width: 4}
	material7Observe4133D0(event)
	event.Word = *(*uint32)(address)
	event.Complete = true
	event.ValueValid = true
	material7Observe4133D0(event)
	runtime.KeepAlive(item)
	return event.Word
}
func material7CacheCell4133D0() **server.ModifierEff {
	if value := material7Current4133D0(); value != nil && value.Cache != nil {
		return value.Cache
	}
	return GetServer().S().Modif.Material7CacheCell4133D0()
}
func material7ResolveID4133D0() int32 {
	if value := material7Current4133D0(); value != nil && value.ResolveID != nil {
		return value.ResolveID("Material7")
	}
	return int32(GetServer().S().Modif.Nox_xxx_modifGetIdByName413290("Material7"))
}
func material7ResolveDesc4133D0(id int32) *server.ModifierEff {
	if value := material7Current4133D0(); value != nil && value.ResolveDesc != nil {
		return value.ResolveDesc(id)
	}
	return GetServer().S().Modif.Nox_xxx_modifGetDescById413330(int(id))
}

// Complete original 72-byte caller order. In particular, the first cache value
// controls cold admission even if a diagnostic mutates the cell afterwards.
// InitData is loaded eagerly, then kept through both external resolver calls.
// A nil resolver result remains cold next time; eligible nil==nil is true.
func rawMaterial7Native4133D0(item unsafe.Pointer) int32 {
	defer runtime.KeepAlive(item)
	cell := material7CacheCell4133D0()
	cached := material7ReadPointer4133D0(unsafe.Pointer(cell), 0, 0x4133d0, 0)
	cold := cached == nil
	offsets := material7Offsets4133D0()
	init := material7ReadPointer4133D0(item, offsets[0], 0x4133dd, 1)
	if cold {
		id := material7ResolveID4133D0()
		resolved := material7ResolveDesc4133D0(id)
		cached = unsafe.Pointer(resolved)
		event := material7Operand4133D0{Site: 0x4133f8, Field: 0, Base: unsafe.Pointer(cell), Address: unsafe.Pointer(cell), EffectiveAddress: uintptr(unsafe.Pointer(cell)), Width: int(unsafe.Sizeof(unsafe.Pointer(nil))), Pointer: cached, IsPointer: true, Write: true, ValueValid: true}
		material7Observe4133D0(event)
		*cell = resolved
		event.Complete = true
		material7Observe4133D0(event)
	}
	if material7ReadClass4133D0(item, offsets[1])&0x13001000 == 0 {
		return 0
	}
	material := material7ReadPointer4133D0(init, offsets[2], 0x413406, 3)
	if material == cached {
		return 1
	}
	return 0
}
