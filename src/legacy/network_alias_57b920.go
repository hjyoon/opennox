package legacy

/*
#include "GAME5_2.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/internal/netalias"
)

func aliasTable57B920(data unsafe.Pointer) netalias.Table {
	return unsafe.Slice((*byte)(data), netalias.TableSize)
}

//export sub_57B920
func sub_57B920(data unsafe.Pointer) C.int {
	clear(aliasTable57B920(data))
	return 0
}

//export nox_xxx_cliGenerateAlias_57B9A0
func nox_xxx_cliGenerateAlias_57B9A0(data unsafe.Pointer, code, typeID C.int, now C.uint) C.char {
	table := aliasTable57B920(data)
	alias := netalias.FindAvailable(int32(code), int32(typeID), uint32(now), table.Lookup)
	return C.char(int8(alias))
}

//export sub_57BA10
func sub_57BA10(data unsafe.Pointer, code, typeID C.short, deadline C.int) unsafe.Pointer {
	entry := netalias.Entry{Code: uint16(code), TypeID: uint16(typeID), Deadline: uint32(deadline)}
	entry.PutRecord(unsafe.Slice((*byte)(data), netalias.RecordSize))
	return data
}
