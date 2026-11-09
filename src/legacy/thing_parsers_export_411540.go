package legacy

/*
#include "memfile.h"
#include "GAME1.h"
*/
import "C"
import "unsafe"

//export nox_thing_read_FLOR_411540
func nox_thing_read_FLOR_411540(f *nox_memfile, p *C.uchar) C.int {
	return C.int(thingReadFloorNative411540(f, unsafe.Pointer(p)))
}

//export nox_thing_read_EDGE_411850
func nox_thing_read_EDGE_411850(f *nox_memfile, p *C.uchar) C.int {
	return C.int(thingReadEdgeNative411850(f, unsafe.Pointer(p)))
}

//export nox_thing_skip_AUD_414D40
func nox_thing_skip_AUD_414D40(f *nox_memfile) C.int { return C.int(thingSkipAudioNative414D40(f)) }

//export nox_thing_read_FLOR_414DB0
func nox_thing_read_FLOR_414DB0(f *nox_memfile) C.int { return C.int(thingReadFloorNative414DB0(f)) }

//export nox_thing_read_EDGE_414E70
func nox_thing_read_EDGE_414E70(f *nox_memfile, p unsafe.Pointer) C.int {
	return C.int(thingReadEdgeNative414E70(f, p))
}

//export nox_thing_read_WALL_414F60
func nox_thing_read_WALL_414F60(f *nox_memfile, p unsafe.Pointer) C.int {
	return C.int(thingReadWallNative414F60(f, p))
}

//export nox_thing_skip_spells_415100
func nox_thing_skip_spells_415100(f *nox_memfile) C.int { return C.int(thingSkipSpellsNative415100(f)) }

//export nox_thing_read_image_415240
func nox_thing_read_image_415240(f *nox_memfile) C.int { return C.int(thingReadImageNative415240(f)) }

//export nox_thing_read_ability_415320
func nox_thing_read_ability_415320(f *nox_memfile) C.int {
	return C.int(thingReadAbilityNative415320(f))
}

//export nox_thing_read_audio_415660
func nox_thing_read_audio_415660(f *nox_memfile, p *C.char) C.int {
	return C.int(thingReadAudioNative415660(f, unsafe.Pointer(p)))
}
