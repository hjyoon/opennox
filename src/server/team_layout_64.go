//go:build amd64 || arm64

package server

import "unsafe"

// These are native Go offsets. The original packed Win32 prefix is 80 bytes;
// pointer-sized int and unsafe.Pointer fields expand it on 64-bit targets.
var (
	_ = [1]struct{}{}[96-unsafe.Sizeof(Team{})]
	_ = [1]struct{}{}[56-unsafe.Offsetof(Team{}.Lessons)]
	_ = [1]struct{}{}[65-unsafe.Offsetof(Team{}.IDVal)]
	_ = [1]struct{}{}[80-unsafe.Offsetof(Team{}.Field_72)]
	_ = [1]struct{}{}[88-unsafe.Offsetof(Team{}.field_76)]
)
