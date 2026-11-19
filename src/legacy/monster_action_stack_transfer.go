package legacy

/*
#include "GAME4_2.h"
*/
import "C"

import (
	"bytes"
	"encoding/binary"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

func shiftedMonsterActionFrame(value uint32, delta int32) int32 {
	return int32(value + uint32(delta))
}

//export nox_xxx_AssignIfGreater_52A420
func nox_xxx_AssignIfGreater_52A420(value *C.int, delta C.int) C.int {
	shifted := shiftedMonsterActionFrame(uint32(*value), int32(delta))
	*value = C.int(max(shifted, 1))
	return C.int(shifted)
}

func legacyActionTransferBytes(data []byte) C.int {
	ptr := (*C.uchar)(unsafe.Pointer(unsafe.SliceData(data)))
	return C.int(nox_xxx_fileReadWrite_426AC0_file3_fread_impl(ptr, C.size_t(len(data)), nil, 0))
}

//go:nocheckptr
func legacyActionWaypointIndex(address uintptr) uint32 {
	return (*server.Waypoint)(unsafe.Pointer(address)).Index
}

//export sub_52A440
func sub_52A440(_ C.int, record unsafe.Pointer, delta C.int) C.int {
	if record == nil {
		return 0
	}
	item := (*server.AIStackItem)(record)
	name, ok := monsterActionNameNative534650(item.Action)
	if !ok {
		return 0
	}
	reading := cryptfile.Global().ReadOnly()
	var nameBytes [256]byte
	length := [1]byte{byte(copy(nameBytes[:], name))}
	legacyActionTransferBytes(length[:])
	storedName := nameBytes[:length[0]]
	legacyActionTransferBytes(storedName)
	if end := bytes.IndexByte(storedName, 0); end >= 0 {
		storedName = storedName[:end]
	}
	item.Action = uint32(monsterActionByName528DB0(string(storedName), 72, ai.ACTION_IDLE))
	base := 255604 + 16*uintptr(item.Action)
	count := [1]byte{memmap.Uint8(0x587000, base)}
	legacyActionTransferBytes(count[:])
	if int(count[0]) > len(item.Args)/2 {
		return 0
	}
	for i := 0; i < int(count[0]); i++ {
		slot := i * 2
		kind := memmap.Uint32(0x587000, base+4+4*uintptr(i))
		value := uint32(item.Args[slot])
		width := 4
		switch kind {
		case 0:
			width = 8
		case 1:
			if !reading && item.Args[slot] != 0 {
				value = uint32(item.ArgObj(slot).ScriptIDVal)
			}
		case 2:
			if !reading && item.Args[slot] != 0 {
				value = legacyActionWaypointIndex(item.Args[slot])
			}
		case 3, 4, 5, 6:
		case 7:
			width = 1
		default:
			return C.int(kind)
		}
		var data [8]byte
		binary.LittleEndian.PutUint32(data[:4], value)
		binary.LittleEndian.PutUint32(data[4:], uint32(item.Args[slot+1]))
		legacyActionTransferBytes(data[:width])
		if !reading {
			continue
		}
		value = binary.LittleEndian.Uint32(data[:4])
		if kind == 5 {
			value = uint32(max(shiftedMonsterActionFrame(value, int32(delta)), 1))
		}
		item.Args[slot] = uintptr(value)
		if width == 8 {
			item.Args[slot+1] = uintptr(binary.LittleEndian.Uint32(data[4:]))
		}
	}
	var field [4]byte
	binary.LittleEndian.PutUint32(field[:], item.Field5)
	result := legacyActionTransferBytes(field[:])
	item.Field5 = binary.LittleEndian.Uint32(field[:])
	return result
}
