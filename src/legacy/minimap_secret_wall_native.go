package legacy

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

// The original minimap hides closing and fully open secret walls.
func secretWallHiddenMinimap472600(ptr unsafe.Pointer) bool {
	state := (*server.SecretWall)(ptr).State
	return state == 2 || state == 3
}
