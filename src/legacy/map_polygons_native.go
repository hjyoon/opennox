package legacy

/*
#include <stdlib.h>
#include "GAME1_1.h"
#include "GAME1_2.h"
*/
import "C"

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"unsafe"

	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

const (
	mapPolygonVersion       = 4
	mapPolygonAngleCapacity = (552228 - 535844) / 16
	mapPolygonCapacity      = 255
)

// The map polygon table keeps its fixed-width records. Data and vertex
// allocations live in the existing native sidecars keyed by record address.
type mapPolygonRecord struct {
	DataMarker   uint32
	Name         [76]byte
	Index        uint32
	Active       uint32
	Bounds       [4]int32
	Color        [3]byte
	_            byte
	VertexMarker uint32
	Handlers     [2]server.ScriptCallback
	VertexCount  uint16
	AudioZone    byte
	_            byte
	Flags        uint32
	Visited      uint32
}

var _ = [1]struct{}{}[unsafe.Sizeof(Nox_player_polygon_check_data{})-unsafe.Sizeof(mapPolygonRecord{})]

func readMapPolygonAngles(cf *cryptfile.CryptFile, version int16, debug func(string, ...any)) error {
	count, err := cf.ReadU32()
	if err != nil {
		return err
	}
	debug("[polygon] angles=%d version=%d\n", count, version)
	if count >= mapPolygonAngleCapacity {
		return fmt.Errorf("map has %d polygon vertices, maximum %d", count, mapPolygonAngleCapacity-1)
	}
	for index := uint32(1); index <= count; index++ {
		remap, err := cf.ReadU32()
		if err != nil {
			return err
		}
		var coords [8]byte
		if err := mapSectionTransfer(cf, coords[:]); err != nil {
			return err
		}
		x := math.Float32frombits(binary.LittleEndian.Uint32(coords[:4]))
		y := math.Float32frombits(binary.LittleEndian.Uint32(coords[4:]))
		if Nox_xxx_polygonSetAngle_420D40(x, y, index, remap) == nil {
			return fmt.Errorf("initialize map polygon vertex %d", index)
		}
		if version < 3 {
			if _, err := cf.ReadU32(); err != nil {
				return err
			}
		}
	}
	return nil
}

func mapPolygonBounds(vertices []uint32) ([4]int32, error) {
	var bounds [4]int32
	for i, index := range vertices {
		if index >= mapPolygonAngleCapacity {
			return bounds, fmt.Errorf("invalid map polygon vertex index %d", index)
		}
		angle := Nox_xxx_polygonGetAngle_421030(index)
		for axis, value := range [...]float32{angle.X, angle.Y} {
			if i == 0 {
				bounds[axis] = server.X87TruncateSignedDWORD419A70(value)
				bounds[axis+2] = bounds[axis]
				continue
			}
			v := float64(value)
			if v >= float64(bounds[axis]) {
				if v > float64(bounds[axis+2]) {
					bounds[axis+2] = server.X87TruncateSignedDWORD419A70(value)
				}
			} else {
				bounds[axis] = server.X87TruncateSignedDWORD419A70(value)
			}
		}
	}
	return bounds, nil
}

func mapPolygonHandlers(cf *cryptfile.CryptFile, poly *mapPolygonRecord, context unsafe.Pointer) error {
	deps := scriptHandlerXferRuntimeDeps4F5580(cf)
	var contextErr error
	deps.storeContext = func(dst unsafe.Pointer, name []byte) {
		if dst == nil || len(name) >= 128 {
			contextErr = fmt.Errorf("invalid map polygon script context for %d-byte name", len(name))
			return
		}
		buf := unsafe.Slice((*byte)(dst), len(name)+1)
		copy(buf, name)
		buf[len(name)] = 0
	}
	for i := range poly.Handlers {
		if scriptHandlerXferNative4F5580(&poly.Handlers[i], context, deps) == 0 {
			return fmt.Errorf("invalid map polygon script handler %d", i+1)
		}
		if contextErr != nil {
			return contextErr
		}
		if context != nil {
			context = unsafe.Add(context, 128)
		}
	}
	return nil
}

func readMapPolygon(cf *cryptfile.CryptFile, version int16, index uint32, debug func(string, ...any)) (bool, error) {
	debug("[polygon] record=%d begin\n", index)
	raw := sub_421230()
	if raw == nil {
		return false, fmt.Errorf("initialize map polygon")
	}
	poly := (*mapPolygonRecord)(unsafe.Pointer(raw))
	size, err := cf.ReadU8()
	if err != nil {
		return false, err
	}
	if int(size) >= len(poly.Name) {
		return false, fmt.Errorf("map polygon name is %d bytes, maximum %d", size, len(poly.Name)-1)
	}
	if err := mapSectionTransfer(cf, poly.Name[:size]); err != nil {
		return false, err
	}
	poly.Name[size] = 0
	if version < 3 {
		if _, err := cf.ReadU32(); err != nil {
			return false, err
		}
		var legacy [12]byte
		if err := mapSectionTransfer(cf, legacy[:]); err != nil {
			return false, err
		}
		poly.Color = [3]byte{legacy[0], legacy[4], legacy[8]}
	} else {
		for i := range poly.Color {
			if poly.Color[i], err = cf.ReadU8(); err != nil {
				return false, err
			}
		}
	}
	if poly.AudioZone, err = cf.ReadU8(); err != nil {
		return false, err
	}
	if poly.VertexCount, err = cf.ReadU16(); err != nil {
		return false, err
	}
	if poly.VertexCount == 0 {
		return false, fmt.Errorf("map polygon has no vertices")
	}
	sizeBytes := 4 * int(poly.VertexCount)
	vertexData := C.calloc(C.size_t(poly.VertexCount), 4)
	if vertexData == nil {
		return false, fmt.Errorf("allocate map polygon vertices")
	}
	if err := mapSectionTransfer(cf, unsafe.Slice((*byte)(vertexData), sizeBytes)); err != nil {
		C.free(vertexData)
		return false, err
	}
	nox_xxx_polygonSetVertexIndicesNative(unsafe.Pointer(poly), (*C.uint32_t)(vertexData))
	sub_421040(unsafe.Pointer(poly))
	debug("[polygon] record=%d vertices=%d ptr=%p\n", index, poly.VertexCount, vertexData)
	poly.Bounds, err = mapPolygonBounds(unsafe.Slice((*uint32)(vertexData), poly.VertexCount))
	if err != nil {
		return false, err
	}
	if version >= 2 {
		context := nox_xxx_polygonGetDataNative(unsafe.Pointer(poly))
		debug("[polygon] record=%d handler1 data=%p\n", index, context)
		if err := mapPolygonHandlers(cf, poly, context); err != nil {
			return false, err
		}
		debug("[polygon] record=%d handlers done\n", index)
	}
	if version >= 4 {
		if poly.Flags, err = cf.ReadU32(); err != nil {
			return false, err
		}
		debug("[polygon] record=%d done\n", index)
	}
	return poly.Flags&1 != 0, nil
}

func writeMapPolygonAngles(cf *cryptfile.CryptFile) error {
	var count uint32
	for angle := nox_xxx_polygon_420CA0(); angle != nil; angle = nox_xxx_polygon_420CD0((*C.uint32_t)(unsafe.Pointer(angle))) {
		count++
	}
	if err := cf.WriteU32(count); err != nil {
		return err
	}
	for angle := nox_xxx_polygon_420CA0(); angle != nil; angle = nox_xxx_polygon_420CD0((*C.uint32_t)(unsafe.Pointer(angle))) {
		raw := unsafe.Slice((*byte)(unsafe.Pointer(angle)), 12)
		if err := mapSectionTransfer(cf, raw[:4]); err != nil {
			return err
		}
		if err := mapSectionTransfer(cf, raw[4:]); err != nil {
			return err
		}
	}
	return nil
}

func writeMapPolygon(cf *cryptfile.CryptFile, poly *mapPolygonRecord) error {
	name := poly.Name[:alloc.StrLenS(poly.Name[:])]
	if err := cf.WriteU8(byte(len(name))); err != nil {
		return err
	}
	if err := mapSectionTransfer(cf, name); err != nil {
		return err
	}
	for _, color := range poly.Color {
		if err := cf.WriteU8(color); err != nil {
			return err
		}
	}
	if err := cf.WriteU8(poly.AudioZone); err != nil {
		return err
	}
	if err := cf.WriteU16(poly.VertexCount); err != nil {
		return err
	}
	vertexData := nox_xxx_polygonGetVertexIndicesNative(unsafe.Pointer(poly))
	if vertexData == nil && poly.VertexCount != 0 {
		return fmt.Errorf("map polygon has %d missing vertices", poly.VertexCount)
	}
	if err := mapSectionTransfer(cf, unsafe.Slice((*byte)(unsafe.Pointer(vertexData)), int(poly.VertexCount)*4)); err != nil {
		return err
	}
	context := nox_xxx_polygonGetDataNative(unsafe.Pointer(poly))
	if err := mapPolygonHandlers(cf, poly, context); err != nil {
		return err
	}
	return cf.WriteU32(poly.Flags)
}

func mapRWPolygons(cf *cryptfile.CryptFile, regionPresent bool) error {
	if regionPresent {
		return nil
	}
	if cf == nil {
		cf = cryptfile.Global()
		if cf == nil {
			return fmt.Errorf("map polygons require an open crypt file")
		}
	}
	version, err := cf.ReadWriteU16(mapPolygonVersion)
	if err != nil {
		return err
	}
	if int16(version) > mapPolygonVersion {
		return fmt.Errorf("unsupported map polygon version %d", version)
	}
	if cf.ReadOnly() {
		debug := func(string, ...any) {}
		if _, enabled := os.LookupEnv("NOX_DEBUG_POLYGONS"); enabled {
			debug = func(format string, args ...any) { fmt.Fprintf(os.Stderr, format, args...) }
		}
		if err := readMapPolygonAngles(cf, int16(version), debug); err != nil {
			return err
		}
		count, err := cf.ReadU32()
		if err != nil {
			return err
		}
		debug("[polygon] records=%d\n", count)
		if count >= mapPolygonCapacity {
			return fmt.Errorf("map has %d polygons, maximum %d", count, mapPolygonCapacity-1)
		}
		var questRegions C.int
		for i := range count {
			quest, err := readMapPolygon(cf, int16(version), i+1, debug)
			if err != nil {
				return fmt.Errorf("map polygon %d: %w", i+1, err)
			}
			if quest {
				questRegions++
			}
		}
		sub_4D72D0(questRegions)
		return nil
	}

	if err := writeMapPolygonAngles(cf); err != nil {
		return err
	}
	var count uint32
	for p := Nox_xxx_polygonGetNext_4210A0(); p != nil; p = Sub_4210E0(p) {
		count++
	}
	if err := cf.WriteU32(count); err != nil {
		return err
	}
	for p := Nox_xxx_polygonGetNext_4210A0(); p != nil; p = Sub_4210E0(p) {
		if err := writeMapPolygon(cf, (*mapPolygonRecord)(unsafe.Pointer(p))); err != nil {
			return err
		}
	}
	return nil
}

//export nox_server_mapRWPolygons_428CD0
func nox_server_mapRWPolygons_428CD0(region C.int) C.int {
	return C.int(bool2int(mapRWPolygons(nil, region != 0) == nil))
}
