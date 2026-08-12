//go:build amd64 || arm64

package server

import "unsafe"

// These are the current native Go offsets, not the original Win32 layout.
// Player and PlayerUpdateData still require a full legacy/runtime split.
var (
	_ = [1]struct{}{}[320-unsafe.Offsetof(PlayerUpdateData{}.Player)]
	_ = [1]struct{}{}[4912-unsafe.Offsetof(Player{}.CameraFollowObj)]
)
