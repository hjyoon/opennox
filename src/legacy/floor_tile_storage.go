package legacy

/*
#include "floor_tile_storage.h"
*/
import "C"

import "github.com/opennox/opennox/v1/legacy/common/alloc"

// Go owns the fixed-size native table shared by the terrain loaders and renderer.
var floorTileStorage, _ = alloc.NewHigh(C.nox_floor_tile_storage_t{})

var floorEditState, _ = alloc.New(C.nox_floor_edit_state_t{})

//export nox_floor_edit_storage
func nox_floor_edit_storage() *C.nox_floor_edit_state_t { return floorEditState }

//export nox_floor_tile_storage
func nox_floor_tile_storage() *C.nox_floor_tile_storage_t {
	return floorTileStorage
}
