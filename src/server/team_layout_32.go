//go:build 386 || arm

package server

import "unsafe"

var (
	_ = [1]struct{}{}[80-unsafe.Sizeof(Team{})]
	_ = [1]struct{}{}[4-unsafe.Sizeof(Team{}.Lessons)]
	_ = [1]struct{}{}[52-unsafe.Offsetof(Team{}.Lessons)]
	_ = [1]struct{}{}[57-unsafe.Offsetof(Team{}.IDVal)]
	_ = [1]struct{}{}[72-unsafe.Offsetof(Team{}.Field_72)]
	_ = [1]struct{}{}[76-unsafe.Offsetof(Team{}.field_76)]
)
