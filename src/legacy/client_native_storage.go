package legacy

/*
#include "client_native_storage.h"
*/
import "C"

import "github.com/opennox/opennox/v1/legacy/common/alloc"

var clientDisplayStorage, _ = alloc.New(C.nox_client_display_state_t{})

//export nox_client_display_storage
func nox_client_display_storage() *C.nox_client_display_state_t { return clientDisplayStorage }

var clientAudioStorage = func() *C.nox_client_audio_state_t {
	state, _ := alloc.New(C.nox_client_audio_state_t{})
	state.music_enabled = 0x1
	state.music_idle = 0x1
	return state
}()

//export nox_client_audio_storage
func nox_client_audio_storage() *C.nox_client_audio_state_t { return clientAudioStorage }

var clientItemStorage, _ = alloc.New(C.nox_client_item_state_t{})

//export nox_client_item_storage
func nox_client_item_storage() *C.nox_client_item_state_t { return clientItemStorage }
