//go:build !server

package legacy

/*
#include "client__gui__guigen.h"
*/
import "C"
import "unsafe"

// Declaration-only access to the actual local-map-generation C entry.
// Network tests use the existing full C dispatcher, not this bridge.
func MapGenClientTextC4A9D00(packet unsafe.Pointer) int16 {
	return int16(C.nox_xxx_mapGenClientText_4A9D00((*C.uchar)(packet)))
}
