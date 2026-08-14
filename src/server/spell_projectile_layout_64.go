//go:build amd64 || arm64

package server

import "unsafe"

// These are native runtime offsets. They intentionally differ from the
// original Win32 record after each widened object pointer; legacy C must not
// access this record with fixed 32-bit word indexes on 64-bit targets.
var (
	_ = [1]struct{}{}[40-unsafe.Sizeof(SpellProjectileUpdateData{})]
	_ = [1]struct{}{}[0-unsafe.Offsetof(SpellProjectileUpdateData{}.Field0)]
	_ = [1]struct{}{}[8-unsafe.Offsetof(SpellProjectileUpdateData{}.Target)]
	_ = [1]struct{}{}[16-unsafe.Offsetof(SpellProjectileUpdateData{}.Field8)]
	_ = [1]struct{}{}[24-unsafe.Offsetof(SpellProjectileUpdateData{}.Spell12)]
	_ = [1]struct{}{}[28-unsafe.Offsetof(SpellProjectileUpdateData{}.Level16)]
	_ = [1]struct{}{}[32-unsafe.Offsetof(SpellProjectileUpdateData{}.Field20)]
	_ = [1]struct{}{}[36-unsafe.Offsetof(SpellProjectileUpdateData{}.Field24)]

	_ = [1]struct{}{}[16-unsafe.Sizeof(SpellAcceptArg{})]
	_ = [1]struct{}{}[0-unsafe.Offsetof(SpellAcceptArg{}.Obj)]
	_ = [1]struct{}{}[8-unsafe.Offsetof(SpellAcceptArg{}.Pos)]
)
