package legacy

/*
#include "GAME4.h"
*/
import "C"

import (
	"bytes"
	"encoding/binary"
	"io"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/internal/binfile"
	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

const mapgenLoadMinimumRecord503830 = int32(56)

type mapgenLoadReader503830 struct {
	raw *binfile.File
	cf  *cryptfile.CryptFile
	end int64
}

func (r mapgenLoadReader503830) position() (int64, bool) {
	pos, err := r.raw.Seek(0, io.SeekCurrent)
	return pos, err == nil
}

func (r mapgenLoadReader503830) readRaw(dst []byte) bool {
	pos, ok := r.position()
	if !ok || pos > r.end || int64(len(dst)) > r.end-pos {
		return false
	}
	_, err := io.ReadFull(r.raw, dst)
	return err == nil
}

func (r mapgenLoadReader503830) readCrypt(dst []byte) bool {
	pos, ok := r.position()
	if !ok || pos > r.end || int64(len(dst)) > r.end-pos {
		return false
	}
	n, err := r.cf.ReadWrite(dst)
	return err == nil && n == len(dst)
}

func (r mapgenLoadReader503830) readRawU32() (uint32, bool) {
	var buf [4]byte
	if !r.readRaw(buf[:]) {
		return 0, false
	}
	return binary.LittleEndian.Uint32(buf[:]), true
}

func (r mapgenLoadReader503830) readCryptU32() (uint32, bool) {
	var buf [4]byte
	if !r.readCrypt(buf[:]) {
		return 0, false
	}
	return binary.LittleEndian.Uint32(buf[:]), true
}

func mapgenCString503830(buf []byte) string {
	if i := bytes.IndexByte(buf, 0); i >= 0 {
		buf = buf[:i]
	}
	return string(buf)
}

func mapgenOrderCorners503830(corners *[8]int32) {
	topX, topY := corners[0], corners[1]
	for i := 1; i < 4; i++ {
		x, y := corners[2*i], corners[2*i+1]
		if y < topY {
			topX, topY = x, y
		}
	}

	leftX, leftY := corners[2], corners[3]
	for _, i := range [...]int{0, 2, 3} {
		x, y := corners[2*i], corners[2*i+1]
		if x < leftX {
			leftX, leftY = x, y
		}
	}

	rightX, rightY := corners[4], corners[5]
	for _, i := range [...]int{0, 1, 3} {
		x, y := corners[2*i], corners[2*i+1]
		if x > rightX {
			rightX, rightY = x, y
		}
	}

	bottomX, bottomY := corners[6], corners[7]
	for i := 0; i < 3; i++ {
		x, y := corners[2*i], corners[2*i+1]
		if y > bottomY {
			bottomX, bottomY = x, y
		}
	}

	*corners = [8]int32{topX, topY, leftX, leftY, rightX, rightY, bottomX, bottomY}
}

func mapgenBounds503830(corners *[8]int32) [4]int32 {
	minY, maxY := corners[7], corners[1]
	if corners[1] < corners[7] {
		minY, maxY = corners[1], corners[7]
	}
	minX, maxX := corners[4], corners[2]
	if corners[2] < corners[4] {
		minX, maxX = corners[2], corners[4]
	}
	if minX < 0 {
		minX = 0
	}
	if minY < 0 {
		minY = 0
	}
	if maxX >= 5888 {
		maxX = 5887
	}
	if maxY >= 5888 {
		maxY = 5887
	}
	return [4]int32{minX, minY, maxX, maxY}
}

// mapgenLoadCurrent503830 restores the body of GAME.EXE 00503830 after the
// compatibility wrapper has selected and opened an AreaMap record. Parsing
// and all transient pointers stay native-width in Go; the legacy C symbol is
// retained only for callers that identify this routine by its original ABI.
func mapgenLoadCurrent503830() bool {
	cf := cryptfile.Global()
	if cf == nil || cf.File == nil || cf.File.File == nil {
		return false
	}
	raw := cf.File.File

	var lengthBuf [4]byte
	if _, err := io.ReadFull(raw, lengthBuf[:]); err != nil {
		return false
	}
	recordLength := int32(binary.LittleEndian.Uint32(lengthBuf[:]))
	if recordLength < mapgenLoadMinimumRecord503830 {
		return false
	}
	afterLength, err := raw.Seek(0, io.SeekCurrent)
	if err != nil {
		return false
	}
	fileSize, err := raw.Size()
	if err != nil || afterLength > fileSize || int64(recordLength) > fileSize-afterLength {
		return false
	}
	reader := mapgenLoadReader503830{raw: raw, cf: cf, end: afterLength + int64(recordLength)}

	var one [1]byte
	if !reader.readRaw(one[:]) || one[0] >= 64 {
		return false
	}
	name := make([]byte, int(one[0]))
	if !reader.readRaw(name) || !reader.readRaw(one[:]) {
		return false
	}
	if !reader.readRaw(one[:]) {
		return false
	}
	version := one[0]
	var ignored [8]byte
	if !reader.readRaw(ignored[:]) {
		return false
	}
	if version > 1 {
		attachmentLength, ok := reader.readRawU32()
		if !ok || int32(attachmentLength) < 0 {
			return false
		}
		pos, ok := reader.position()
		if !ok || pos > reader.end || int64(attachmentLength) > reader.end-pos {
			return false
		}
		if _, err := raw.Seek(int64(attachmentLength), io.SeekCurrent); err != nil {
			return false
		}
	}

	magic, ok := reader.readRawU32()
	if !ok || magic != mapgenMagic502ED0 {
		return false
	}
	wallX, ok := reader.readRawU32()
	if !ok {
		return false
	}
	wallY, ok := reader.readRawU32()
	if !ok {
		return false
	}
	*memmap.PtrUint32(0x5D4594, 739980) = wallX
	*memmap.PtrUint32(0x5D4594, 739984) = wallY

	var corners [8]int32
	for _, index := range [...]int{0, 1, 6, 7, 2, 3, 4, 5} {
		value, ok := reader.readRawU32()
		if !ok {
			return false
		}
		corners[index] = int32(value)
	}
	mapgenOrderCorners503830(&corners)
	for i, value := range corners {
		*memmap.PtrUint32(0x5D4594, uintptr(1599500+4*i)) = uint32(value)
	}
	bounds := mapgenBounds503830(&corners)

	cf.SetXOR(true)
	for {
		if !reader.readCrypt(one[:]) {
			return false
		}
		if one[0] == 0 {
			cf.SetXOR(false)
			return true
		}
		sectionName := make([]byte, int(one[0]))
		if !reader.readCrypt(sectionName) {
			return false
		}
		sectionLength, ok := reader.readCryptU32()
		if !ok {
			return false
		}
		sectionStart, ok := reader.position()
		if !ok || sectionStart > reader.end || int64(sectionLength) > reader.end-sectionStart {
			return false
		}

		if Nox_xxx_mapReadSection_426EA0 == nil {
			return false
		}
		name := mapgenCString503830(sectionName)
		handled, err := Nox_xxx_mapReadSection_426EA0(unsafe.Pointer(&corners[0]), name)
		if err != nil {
			mapLog.Println(err)
			return false
		}
		if !handled && !mapgenLoadPlaceObject503830(name, unsafe.Pointer(&bounds[0])) {
			return false
		}

		pos, ok := reader.position()
		if !ok || pos > sectionStart+int64(sectionLength) {
			return false
		}
	}
}

//export nox_mapgenLoadNative_503830
func nox_mapgenLoadNative_503830(_ C.int) C.int {
	return C.int(bool2int(mapgenLoadCurrent503830()))
}

// The original loader calls an object's transfer callback and then places the
// object. Keep both object pointers native-width across the C/Go boundary.
type mapgenLoadObjectDeps503830 struct {
	newObject   func(string) *server.Object
	xfer        func(*server.Object, unsafe.Pointer) error
	freeObject  func(*server.Object)
	placeObject func(*server.Object, *ntype.Point32) int32
}

func mapgenLoadObjectWithDeps503830(name string, bounds unsafe.Pointer, deps mapgenLoadObjectDeps503830) bool {
	obj := deps.newObject(name)
	if obj == nil {
		return false
	}
	if obj.Xfer == nil || deps.xfer(obj, bounds) != nil {
		deps.freeObject(obj)
		return false
	}
	// GAME.EXE ignores the placement result. Placement owns the object,
	// including its cleanup on rejection.
	deps.placeObject(obj, (*ntype.Point32)(bounds))
	return true
}

var mapgenLoadPlaceObject503830 = func(name string, bounds unsafe.Pointer) bool {
	s := GetServer().S()
	return mapgenLoadObjectWithDeps503830(name, bounds, mapgenLoadObjectDeps503830{
		newObject: s.NewObjectByTypeID,
		xfer: func(obj *server.Object, bounds unsafe.Pointer) error {
			return obj.CallXfer(bounds)
		},
		freeObject: func(obj *server.Object) {
			s.Objs.FreeObject(obj)
		},
		placeObject: func(obj *server.Object, bounds *ntype.Point32) int32 {
			return Nox_xxx_servMapLoadPlaceObj_4F3F50(obj, nil, bounds)
		},
	})
}

//export nox_mapgenLoadPlaceObject_503830
func nox_mapgenLoadPlaceObject_503830(name *C.char, bounds unsafe.Pointer) C.int {
	return C.int(bool2int(mapgenLoadPlaceObject503830(GoString(name), bounds)))
}
