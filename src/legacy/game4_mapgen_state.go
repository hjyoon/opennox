package legacy

// #include "game4_mapgen_state.h"
import "C"

import "github.com/opennox/opennox/v1/legacy/common/alloc"

// The archive and list addresses are also used by native callers.
var game4Mapgen, _ = alloc.New(C.nox_game4_mapgen_state{})

//export nox_game4_mapgen_storage
func nox_game4_mapgen_storage() *C.nox_game4_mapgen_state {
	return game4Mapgen
}
