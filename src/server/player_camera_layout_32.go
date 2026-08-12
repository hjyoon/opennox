//go:build 386 || arm

package server

import "unsafe"

var (
	_ = [1]struct{}{}[276-unsafe.Offsetof(PlayerUpdateData{}.Player)]
	_ = [1]struct{}{}[2056-unsafe.Offsetof(Player{}.PlayerUnit)]
	_ = [1]struct{}{}[3628-unsafe.Offsetof(Player{}.CameraFollowObj)]
	_ = [1]struct{}{}[3680-unsafe.Offsetof(Player{}.Field3680)]
)
