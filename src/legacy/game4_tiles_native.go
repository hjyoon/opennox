package legacy

/*
#include <stdint.h>
#include <stdlib.h>
#include <string.h>
#include "GAME4.h"
#include "GAME4_1.h"
#include "GAME4_3.h"
#include "GAME3_2.h"

extern nox_tile_coord_entry_t* dword_5d4594_1599556;
extern uint32_t nox_tile_def_cnt;
extern nox_tileDef_t nox_tile_defs_arr[176];

// Give cgo a unique fully prototyped entry instead of calling the shared
// historical K&R declaration directly. This is only a raw 72-byte copy.
static void game4_tiles_snapshot_504150(void* dst) {
	memcpy(dst, sub_4D3C70(), 72);
}
*/
import "C"

import (
	"math"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

//export nox_xxx_tileAllocTileInCoordList_5040A0
func nox_xxx_tileAllocTileInCoordList_5040A0(x, y C.int, rawKind C.float) *C.nox_tile_coord_entry_t {
	entry := (*C.nox_tile_coord_entry_t)(C.calloc(1, C.sizeof_nox_tile_coord_entry_t))
	if entry == nil {
		return nil
	}
	entry.layer = (*C.nox_tile_layer_t)(C.calloc(1, C.sizeof_nox_tile_layer_t))
	if entry.layer == nil {
		C.free(unsafe.Pointer(entry))
		return nil
	}
	entry.prev = C.dword_5d4594_1599556
	if entry.prev != nil {
		entry.prev.next = entry
	}
	C.dword_5d4594_1599556 = entry
	*memmap.PtrUint32(0x5D4594, 1599560)++
	// a3 is a raw DWORD carrier despite its historical float ABI spelling.
	entry.kind = C.uint8_t(math.Float32bits(float32(rawKind)))
	entry.x = C.float(float32(x) * 46)
	entry.y = C.float(float32(y) * 46)
	if entry.kind == 1 {
		entry.x = C.float(float32(entry.x) + 23)
	} else {
		entry.y = C.float(float32(entry.y) + 23)
	}
	return entry
}

type game4TilesServices struct {
	snapshot  func() [72]byte
	restore   func(*[72]byte)
	image     func(int32)
	variation func(int32)
	mode      func(int32)
	base      func(*C.float2)
	byte3     func(int32)
	byte4     func(int32)
	subtile   func(*C.float2)
}

func game4TilesDefaultServices() game4TilesServices {
	return game4TilesServices{
		snapshot: func() [72]byte {
			var state [72]byte
			C.game4_tiles_snapshot_504150(unsafe.Pointer(&state[0]))
			return state
		},
		restore:   func(state *[72]byte) { C.nox_xxx_tileInitdataClear_4D3C50(unsafe.Pointer(&state[0])) },
		image:     func(value int32) { C.nox_xxx_tileCheckImage_51D540(C.int(value)) },
		variation: func(value int32) { C.nox_xxx_tileCheckImageVari_51D570(C.int(value)) },
		mode:      func(value int32) { C.nox_xxx_tile_51D5C0(C.int(value)) },
		base:      func(at *C.float2) { C.sub_51D8F0(at) },
		byte3:     func(value int32) { C.nox_xxx_tileCheckByte3_544070(C.int(value)) },
		byte4:     func(value int32) { C.nox_xxx_tileCheckByte4_5440A0(C.int(value)) },
		subtile:   func(at *C.float2) { C.nox_xxx_tileSubtile_544310(at) },
	}
}

// Test-only overrides replace actual external service boundaries, without
// replacing list traversal, cached/live reads, transparent lookup or arithmetic.
var game4TilesServiceFactory = game4TilesDefaultServices

//export nox_xxx_tileInit_504150
func nox_xxx_tileInit_504150(x, y C.int) C.int {
	services := game4TilesServiceFactory()
	transparent := memmap.PtrInt32(0x587000, 229704)
	if *transparent == -1 && C.nox_tile_def_cnt > 0 {
		// The current C initializes v2 to zero, not -1. An absent
		// TransparentFloor deliberately leaves the cache -1 and continues.
		for i := uint32(0); i < uint32(C.nox_tile_def_cnt); i++ {
			if C.strcmp((*C.char)(unsafe.Pointer(&C.nox_tile_defs_arr[i].name[0])), game4TilesTransparentName()) == 0 {
				*transparent = int32(i)
				break
			}
		}
	}
	state := services.snapshot()
	entry := C.dword_5d4594_1599556
	if entry != nil {
		baseX, baseY := float32(x), float32(y)
		for entry != nil {
			at := C.float2{field_0: C.float(baseX + float32(entry.x)), field_4: C.float(baseY + float32(entry.y))}
			services.image(int32(entry.layer.field_0))
			services.variation(int32(entry.layer.field_1))
			services.mode(1)
			if uint32(entry.layer.field_0) != uint32(*transparent) {
				services.base(&at)
			}
			for node := entry.layer.subtiles; node != nil; node = node.next {
				services.byte3(int32(node.field_2))
				services.byte4(int32(node.field_3))
				services.image(int32(node.field_0))
				services.variation(int32(node.field_1))
				services.subtile(&at)
			}
			entry = entry.prev
		}
	}
	services.restore(&state)
	return 1
}

// C-owned interned storage is stable across collections and is not a Go
// string pointer retained by a C caller. strcmp remains the original service.
func game4TilesTransparentName() *C.char { return internCStr("TransparentFloor") }
