//go:build 386 || arm

package server

import "unsafe"

var (
	_ = [1]struct{}{}[276-unsafe.Offsetof(PlayerUpdateData{}.Player)]
	_ = [1]struct{}{}[3628-unsafe.Offsetof(Player{}.CameraFollowObj)]
)
