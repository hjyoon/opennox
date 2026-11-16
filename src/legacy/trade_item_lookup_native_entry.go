package legacy

/*
#include "GAME4_1.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

//export sub_50FFE0
func sub_50FFE0(head unsafe.Pointer, code C.int) unsafe.Pointer {
	return unsafe.Pointer(game4TradeFindItem50FFE0((*server.TradeItem)(head), uint32(code)))
}
