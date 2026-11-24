package legacy

/*
#include "GAME1_1.h"
#include "GAME4_2.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/noxcrypt"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

const (
	mapWallOffsetX = 739980
	mapWallOffsetY = 739984
	mapWallFlags   = 739992
	mapgenMinWallX = 2487252
	mapgenMinWallY = 2487256
)

func SetMapWallOffset(x, y uint32) {
	*memmap.PtrUint32(0x5D4594, mapWallOffsetX) = x
	*memmap.PtrUint32(0x5D4594, mapWallOffsetY) = y
}

//export nox_xxx_wallGet_426A30
func nox_xxx_wallGet_426A30() C.int {
	return C.int(memmap.Uint32(0x5D4594, mapWallFlags))
}

//export nox_xxx_mapGetWallSize_426A70
func nox_xxx_mapGetWallSize_426A70() *C.char {
	return (*C.char)(memmap.PtrOff(0x5D4594, mapWallOffsetX))
}

//export nox_xxx_mapWall_426A80
func nox_xxx_mapWall_426A80(offset *C.int) {
	words := (*[2]C.int)(unsafe.Pointer(offset))
	*memmap.PtrUint32(0x5D4594, mapWallOffsetX) = uint32(words[0])
	*memmap.PtrUint32(0x5D4594, mapWallOffsetY) = uint32(words[1])
}

//export nox_xxx_mapSaveMap_51E010
func nox_xxx_mapSaveMap_51E010(path *C.char, compress C.int) C.int {
	name := GoString(path)
	if len(name) < 4 || nox_xxx_cryptOpen_426910(path, int32(cryptfile.WriteOnly), crypt.MapKey) == 0 {
		return 0
	}
	_ = cryptfile.Global().WriteU32(0xFADEFACE)
	checksum := nox_xxx_cryptFlush_4268E0()
	if nox_xxx_map_51E140() == 0 {
		nox_xxx_cryptClose_4269F0()
		return 0
	}
	sub_4268F0(checksum)
	nox_xxx_cryptClose_4269F0()
	if compress == 0 {
		return 1
	}
	suffix := alloc.GoStringS(memmap.Slice(0x587000, 253112)[:5])
	return C.int(bool2int(compressGeneratedMap(name, name[:len(name)-4]+suffix)))
}

//export nox_xxx_map_51E140
func nox_xxx_map_51E140() C.int {
	minX := memmap.PtrUint32(0x5D4594, mapgenMinWallX)
	minY := memmap.PtrUint32(0x5D4594, mapgenMinWallY)
	*minX, *minY = 256, 256
	GetServer().S().Walls.EachWallXxx(func(wall *server.Wall) bool {
		mapCountWalls51DEA0(wall, minX, minY)
		return true
	})
	_, _ = cryptfile.Global().ReadWrite(memmap.Slice(0x5D4594, mapgenMinWallX)[:4])
	_, _ = cryptfile.Global().ReadWrite(memmap.Slice(0x5D4594, mapgenMinWallY)[:4])
	SetMapWallOffset(*minX, *minY)
	nox_xxx_mapSetWallInGlobalDir0pr1_5004D0()
	if nox_xxx_mapWriteSectionsMB_426E20(nil) == 0 {
		nox_xxx_cryptClose_4269F0()
		return 0
	}
	nox_xxx_map_5004F0()
	var end [1]byte
	_, _ = cryptfile.Global().ReadWrite(end[:])
	return 1
}
