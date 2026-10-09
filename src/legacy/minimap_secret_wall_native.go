package legacy

/*
#include "secret_wall.h"
*/
import "C"
import "unsafe"

// Exercise the same typed state predicate used by the native minimap pass.
func secretWallHiddenMinimap472600(ptr unsafe.Pointer) bool {
	return C.nox_secret_wall_hidden_minimap_472600((*C.nox_secret_wall_t)(ptr)) != 0
}
