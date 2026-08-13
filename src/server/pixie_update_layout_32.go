//go:build 386 || arm

package server

import "unsafe"

// PixieUpdateData is byte-for-byte compatible with the original seven-word
// record on 32-bit targets.
var (
	_ = [1]struct{}{}[28-unsafe.Sizeof(PixieUpdateData{})]
	_ = [1]struct{}{}[0-unsafe.Offsetof(PixieUpdateData{}.Owner)]
	_ = [1]struct{}{}[4-unsafe.Offsetof(PixieUpdateData{}.Target)]
	_ = [1]struct{}{}[8-unsafe.Offsetof(PixieUpdateData{}.Field8)]
	_ = [1]struct{}{}[12-unsafe.Offsetof(PixieUpdateData{}.SpellID)]
	_ = [1]struct{}{}[16-unsafe.Offsetof(PixieUpdateData{}.Field16)]
	_ = [1]struct{}{}[20-unsafe.Offsetof(PixieUpdateData{}.Deadline)]
	_ = [1]struct{}{}[24-unsafe.Offsetof(PixieUpdateData{}.LastOwnerVisibleFrame)]
)
