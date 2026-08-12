//go:build amd64 || arm64

package server

import "unsafe"

// These offsets mirror the native-pointer nox_object_t layout asserted in
// legacy/defs.h. The fixed-width legacy map slots prevent pointer expansion in
// fields 64 through 84; Go-only handles start after the 912-byte C prefix.
var (
	_ = [1]struct{}{}[928-unsafe.Sizeof(Object{})]
	_ = [1]struct{}{}[0-unsafe.Offsetof(Object{}.IDPtr)]
	_ = [1]struct{}{}[8-unsafe.Offsetof(Object{}.TypeInd)]
	_ = [1]struct{}{}[12-unsafe.Offsetof(Object{}.ObjClass)]
	_ = [1]struct{}{}[16-unsafe.Offsetof(Object{}.ObjSubClass)]
	_ = [1]struct{}{}[20-unsafe.Offsetof(Object{}.ObjFlags)]
	_ = [1]struct{}{}[40-unsafe.Offsetof(Object{}.NetCode)]
	_ = [1]struct{}{}[48-unsafe.Offsetof(Object{}.ScriptIDVal)]
	_ = [1]struct{}{}[52-unsafe.Offsetof(Object{}.TeamVal)]
	_ = [1]struct{}{}[60-unsafe.Offsetof(Object{}.PosVec)]
	_ = [1]struct{}{}[108-unsafe.Offsetof(Object{}.ZVal)]
	_ = [1]struct{}{}[124-unsafe.Offsetof(Object{}.Mass)]
	_ = [1]struct{}{}[136-unsafe.Offsetof(Object{}.Field33)]
	_ = [1]struct{}{}[152-unsafe.Offsetof(Object{}.Field37)]
	_ = [1]struct{}{}[156-unsafe.Offsetof(Object{}.Field38)]
	_ = [1]struct{}{}[252-unsafe.Offsetof(Object{}.Field62)]
	_ = [1]struct{}{}[260-unsafe.Offsetof(Object{}.legacyMapIndex)]
	_ = [1]struct{}{}[344-unsafe.Offsetof(Object{}.Buffs)]
	_ = [1]struct{}{}[448-unsafe.Offsetof(Object{}.ObjNext)]
	_ = [1]struct{}{}[464-unsafe.Offsetof(Object{}.DeletedNext)]
	_ = [1]struct{}{}[472-unsafe.Offsetof(Object{}.DeletedAt)]
	_ = [1]struct{}{}[520-unsafe.Offsetof(Object{}.InvHolder)]
	_ = [1]struct{}{}[528-unsafe.Offsetof(Object{}.InvNextItem)]
	_ = [1]struct{}{}[544-unsafe.Offsetof(Object{}.InvFirstItem)]
	_ = [1]struct{}{}[552-unsafe.Offsetof(Object{}.ObjOwner)]
	_ = [1]struct{}{}[560-unsafe.Offsetof(Object{}.Field128)]
	_ = [1]struct{}{}[568-unsafe.Offsetof(Object{}.Field129)]
	_ = [1]struct{}{}[616-unsafe.Offsetof(Object{}.HealthData)]
	_ = [1]struct{}{}[624-unsafe.Offsetof(Object{}.Field140)]
	_ = [1]struct{}{}[752-unsafe.Offsetof(Object{}.Init)]
	_ = [1]struct{}{}[760-unsafe.Offsetof(Object{}.InitData)]
	_ = [1]struct{}{}[864-unsafe.Offsetof(Object{}.Update)]
	_ = [1]struct{}{}[872-unsafe.Offsetof(Object{}.UpdateData)]
	_ = [1]struct{}{}[896-unsafe.Offsetof(Object{}.ScriptVars)]
	_ = [1]struct{}{}[904-unsafe.Offsetof(Object{}.ScriptPickup)]
	_ = [1]struct{}{}[912-unsafe.Offsetof(Object{}.serverHandle)]
	_ = [1]struct{}{}[40-unsafe.Sizeof(ModifierInitData{})]
	_ = [1]struct{}{}[32-unsafe.Offsetof(ModifierInitData{}.Field16)]
)
