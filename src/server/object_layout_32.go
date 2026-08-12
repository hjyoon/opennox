//go:build 386 || arm

package server

import "unsafe"

// Object is still passed directly to legacy C. Keep the 32-bit shared prefix
// byte-for-byte compatible with the original nox_object_t and place Go-only
// handles after it.
var (
	_ = [1]struct{}{}[780-unsafe.Sizeof(Object{})]
	_ = [1]struct{}{}[0-unsafe.Offsetof(Object{}.IDPtr)]
	_ = [1]struct{}{}[4-unsafe.Offsetof(Object{}.TypeInd)]
	_ = [1]struct{}{}[8-unsafe.Offsetof(Object{}.ObjClass)]
	_ = [1]struct{}{}[12-unsafe.Offsetof(Object{}.ObjSubClass)]
	_ = [1]struct{}{}[16-unsafe.Offsetof(Object{}.ObjFlags)]
	_ = [1]struct{}{}[36-unsafe.Offsetof(Object{}.NetCode)]
	_ = [1]struct{}{}[44-unsafe.Offsetof(Object{}.ScriptIDVal)]
	_ = [1]struct{}{}[56-unsafe.Offsetof(Object{}.PosVec)]
	_ = [1]struct{}{}[104-unsafe.Offsetof(Object{}.ZVal)]
	_ = [1]struct{}{}[120-unsafe.Offsetof(Object{}.Mass)]
	_ = [1]struct{}{}[132-unsafe.Offsetof(Object{}.Field33)]
	_ = [1]struct{}{}[148-unsafe.Offsetof(Object{}.Field37)]
	_ = [1]struct{}{}[152-unsafe.Offsetof(Object{}.Field38)]
	_ = [1]struct{}{}[248-unsafe.Offsetof(Object{}.Field62)]
	_ = [1]struct{}{}[256-unsafe.Offsetof(Object{}.legacyMapIndex)]
	_ = [1]struct{}{}[340-unsafe.Offsetof(Object{}.Buffs)]
	_ = [1]struct{}{}[444-unsafe.Offsetof(Object{}.ObjNext)]
	_ = [1]struct{}{}[492-unsafe.Offsetof(Object{}.InvHolder)]
	_ = [1]struct{}{}[504-unsafe.Offsetof(Object{}.InvFirstItem)]
	_ = [1]struct{}{}[508-unsafe.Offsetof(Object{}.ObjOwner)]
	_ = [1]struct{}{}[512-unsafe.Offsetof(Object{}.Field128)]
	_ = [1]struct{}{}[516-unsafe.Offsetof(Object{}.Field129)]
	_ = [1]struct{}{}[556-unsafe.Offsetof(Object{}.HealthData)]
	_ = [1]struct{}{}[560-unsafe.Offsetof(Object{}.Field140)]
	_ = [1]struct{}{}[688-unsafe.Offsetof(Object{}.Init)]
	_ = [1]struct{}{}[692-unsafe.Offsetof(Object{}.InitData)]
	_ = [1]struct{}{}[744-unsafe.Offsetof(Object{}.Update)]
	_ = [1]struct{}{}[760-unsafe.Offsetof(Object{}.ScriptVars)]
	_ = [1]struct{}{}[764-unsafe.Offsetof(Object{}.ScriptPickup)]
	_ = [1]struct{}{}[772-unsafe.Offsetof(Object{}.serverHandle)]
	_ = [1]struct{}{}[20-unsafe.Sizeof(ModifierInitData{})]
	_ = [1]struct{}{}[16-unsafe.Offsetof(ModifierInitData{}.Field16)]
)
