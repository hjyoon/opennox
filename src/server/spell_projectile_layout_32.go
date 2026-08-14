//go:build 386 || arm

package server

import "unsafe"

// SpellProjectileUpdateData matches the original Win32 seven-word record on
// 32-bit targets. SpellAcceptArg is the original three-word dispatch argument.
var (
	_ = [1]struct{}{}[28-unsafe.Sizeof(SpellProjectileUpdateData{})]
	_ = [1]struct{}{}[0-unsafe.Offsetof(SpellProjectileUpdateData{}.Field0)]
	_ = [1]struct{}{}[4-unsafe.Offsetof(SpellProjectileUpdateData{}.Target)]
	_ = [1]struct{}{}[8-unsafe.Offsetof(SpellProjectileUpdateData{}.Field8)]
	_ = [1]struct{}{}[12-unsafe.Offsetof(SpellProjectileUpdateData{}.Spell12)]
	_ = [1]struct{}{}[16-unsafe.Offsetof(SpellProjectileUpdateData{}.Level16)]
	_ = [1]struct{}{}[20-unsafe.Offsetof(SpellProjectileUpdateData{}.Field20)]
	_ = [1]struct{}{}[24-unsafe.Offsetof(SpellProjectileUpdateData{}.Field24)]

	_ = [1]struct{}{}[12-unsafe.Sizeof(SpellAcceptArg{})]
	_ = [1]struct{}{}[0-unsafe.Offsetof(SpellAcceptArg{}.Obj)]
	_ = [1]struct{}{}[4-unsafe.Offsetof(SpellAcceptArg{}.Pos)]
)
