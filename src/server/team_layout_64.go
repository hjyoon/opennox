//go:build amd64 || arm64

package server

import "unsafe"

// The fixed-width prefix keeps the original field offsets on every target;
// only the native pointer at +72 expands the record on 64-bit targets.
var (
	_ = [1]struct{}{}[88-unsafe.Sizeof(Team{})]
	_ = [1]struct{}{}[4-unsafe.Sizeof(Team{}.Lessons)]
	_ = [1]struct{}{}[52-unsafe.Offsetof(Team{}.Lessons)]
	_ = [1]struct{}{}[57-unsafe.Offsetof(Team{}.IDVal)]
	_ = [1]struct{}{}[72-unsafe.Offsetof(Team{}.Field_72)]
	_ = [1]struct{}{}[80-unsafe.Offsetof(Team{}.field_76)]
)
