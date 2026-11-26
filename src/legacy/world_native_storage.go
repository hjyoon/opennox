package legacy

/*
#include "world_native_storage.h"
*/
import "C"

import "github.com/opennox/opennox/v1/legacy/common/alloc"

var archiveAndMapgenStorage, _ = alloc.New(C.nox_archive_mapgen_state_t{})

//export nox_archive_mapgen_storage
func nox_archive_mapgen_storage() *C.nox_archive_mapgen_state_t { return archiveAndMapgenStorage }

var worldOwnerStorage, _ = alloc.New(C.nox_world_owner_state_t{})

//export nox_world_owner_storage
func nox_world_owner_storage() *C.nox_world_owner_state_t { return worldOwnerStorage }

var worldWordStorage = func() *C.nox_world_word_state_t {
	// The settings comparisons read up to 16 bytes on either side of a word.
	allocation, _ := alloc.New(struct {
		before [16]byte
		state  C.nox_world_word_state_t
		after  [16]byte
	}{})
	state := &allocation.state
	state.global_record_word_60044 = 0xffffffff
	state.buff_iteration_count = 0x1d
	state.next_polygon_index = 0x1
	return state
}()

//export nox_world_word_storage
func nox_world_word_storage() *C.nox_world_word_state_t { return worldWordStorage }
