package legacy

/*
#include "GAME1_1.h"
*/
import "C"

import (
	"encoding/binary"
	"fmt"
	"io"
	"unsafe"

	"github.com/opennox/opennox/v1/internal/cryptfile"
)

func floorTileFieldsIO(cf *cryptfile.CryptFile, data [4]uint32) ([4]uint32, error) {
	tile, err := cf.ReadWriteU8(byte(data[0]))
	if err != nil {
		return data, err
	}
	variation, err := cf.ReadWriteU16(uint16(data[1]))
	if err != nil {
		return data, err
	}
	edge, err := cf.ReadWriteU8(byte(data[2]))
	if err != nil {
		return data, err
	}
	style, err := cf.ReadWriteU8(byte(data[3]))
	if err != nil {
		return data, err
	}
	return [4]uint32{uint32(tile), uint32(int32(int16(variation))), uint32(edge), uint32(style)}, nil
}

func floorTileRawFieldsIO(cf *cryptfile.CryptFile, data [4]uint32) ([4]uint32, error) {
	var raw [16]byte
	for i, value := range data {
		binary.LittleEndian.PutUint32(raw[i*4:], value)
	}
	n, err := cf.ReadWrite(raw[:])
	if err != nil {
		return data, err
	}
	if n != len(raw) {
		if cf.ReadOnly() {
			return data, io.ErrUnexpectedEOF
		}
		return data, io.ErrShortWrite
	}
	for i := range data {
		data[i] = binary.LittleEndian.Uint32(raw[i*4:])
	}
	return data, nil
}

func floorTileLayerIO(cf *cryptfile.CryptFile, fields *[4]uint32, head **C.nox_tile_list_node_t, dense bool) error {
	transfer := floorTileFieldsIO
	if dense {
		transfer = floorTileRawFieldsIO
	}
	data, err := transfer(cf, *fields)
	if err != nil {
		return err
	}
	*fields = data

	var count byte
	if !cf.ReadOnly() {
		for node := *head; node != nil; node = node.next {
			count++
		}
	}
	count, err = cf.ReadWriteU8(count)
	if err != nil {
		return err
	}
	if cf.ReadOnly() {
		*head = nil
		tail := head
		for range count {
			node := nox_xxx_tileListAddNewSubtile_422160(0, 0, 0, 0)
			if node == nil {
				return fmt.Errorf("allocate floor subtile")
			}
			row, err := transfer(cf, [4]uint32{})
			if err != nil {
				nox_xxx_tileFreeTileOne_4221E0(unsafe.Pointer(node))
				return err
			}
			node.field_0, node.field_1, node.field_2, node.field_3 = C.int(row[0]), C.int(row[1]), C.int(row[2]), C.int(row[3])
			*tail = node
			tail = &node.next
		}
		return nil
	}
	for node := *head; node != nil; node = node.next {
		row := [4]uint32{uint32(node.field_0), uint32(node.field_1), uint32(node.field_2), uint32(node.field_3)}
		if _, err := transfer(cf, row); err != nil {
			return err
		}
	}
	return nil
}

//export nox_xxx_tileReadOne_422A40
func nox_xxx_tileReadOne_422A40(_ C.int, fields *C.uint32_t, head **C.nox_tile_list_node_t) C.uchar {
	cf := cryptfile.Global()
	if cf == nil {
		return 0
	}
	err := floorTileLayerIO(cf, (*[4]uint32)(unsafe.Pointer(fields)), head, false)
	return C.uchar(bool2int(err == nil))
}
