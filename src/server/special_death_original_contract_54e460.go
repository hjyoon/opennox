package server

import "unsafe"

// SpecialDeathCallback54E460 retains the entry-cached update allocation. The
// original script boundary observes its callback address without reading the
// callback record. A malformed nil base is represented without a small Go
// pointer, including across reentrant callbacks and garbage collections.
type SpecialDeathCallback54E460 struct{ Data unsafe.Pointer }

func (p SpecialDeathCallback54E460) NativeCallback() *ScriptCallback {
	if p.Data == nil {
		return nil
	}
	return (*ScriptCallback)(unsafe.Add(p.Data, unsafe.Offsetof(MonsterGenUpdateData{}.Field56)))
}

func (p SpecialDeathCallback54E460) Address() uintptr {
	return uintptr(p.Data) + unsafe.Offsetof(MonsterGenUpdateData{}.Field56)
}

// SpecialDeathRuntime54E460 declares only the original external services for
// Marker, Boulder and MonsterGenerator. The embedded debris services retain
// source-relative origins and DWORD table indices. Quest keeps the integer
// world result while its consuming branch tests only nonzero.
type SpecialDeathRuntime54E460 struct {
	GolemDeathRuntime549E90
	DestroyedName unsafe.Pointer
	Parent        func(*Object) *Object
	Delete        func(*Object) uint32
	Frame         func() uint32
	Timer         func(uint32) uint32
	Mode          func(int32) uint32
	Script        func(SpecialDeathCallback54E460, *Object, *Object, uint32) uint32
	GeneratorFX   func(SkeletonOrigin54A310, uint32) uint32
	Quest         func() int32
	Credit        func(*Object) uint32
}
