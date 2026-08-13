//go:build amd64 || arm64

package server

import "unsafe"

// Native pointers widen the first two fields. Legacy code must use the named
// record rather than fixed four-byte word indexes on 64-bit targets.
var (
	_ = [1]struct{}{}[40-unsafe.Sizeof(PixieUpdateData{})]
	_ = [1]struct{}{}[0-unsafe.Offsetof(PixieUpdateData{}.Owner)]
	_ = [1]struct{}{}[8-unsafe.Offsetof(PixieUpdateData{}.Target)]
	_ = [1]struct{}{}[16-unsafe.Offsetof(PixieUpdateData{}.Field8)]
	_ = [1]struct{}{}[20-unsafe.Offsetof(PixieUpdateData{}.SpellID)]
	_ = [1]struct{}{}[24-unsafe.Offsetof(PixieUpdateData{}.Field16)]
	_ = [1]struct{}{}[28-unsafe.Offsetof(PixieUpdateData{}.Deadline)]
	_ = [1]struct{}{}[32-unsafe.Offsetof(PixieUpdateData{}.LastOwnerVisibleFrame)]
)
