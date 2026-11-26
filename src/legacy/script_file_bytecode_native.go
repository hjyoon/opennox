package legacy

/*
#include <stdio.h>

*/
import "C"

import (
	"encoding/binary"
	"math"
	"unsafe"

	"github.com/opennox/noxscript/ns/asm"
)

func readScriptWord(file *FILE) (uint32, bool) {
	var bytes [4]byte
	n := nox_fs_fread(file, unsafe.Pointer(&bytes[0]), 4)
	return binary.LittleEndian.Uint32(bytes[:]), n == 4
}

func writeScriptWord(file *FILE, value uint32) C.int {
	var bytes [4]byte
	binary.LittleEndian.PutUint32(bytes[:], value)
	return C.int(nox_fs_fwrite(file, unsafe.Pointer(&bytes[0]), 4) / 4)
}

func readScriptInt(file *FILE) (C.int, bool) {
	value, ok := readScriptWord(file)
	return C.int(value), ok
}

//export nox_script_readInt_542B70
func nox_script_readInt_542B70(file *C.FILE) C.int {
	value, _ := readScriptInt(file)
	return value
}

//export nox_script_readFloat_542B90
func nox_script_readFloat_542B90(file *C.FILE) C.double {
	bits, _ := readScriptWord(file)
	return C.double(math.Float32frombits(bits))
}

//export nox_script_writeInt_542BB0
func nox_script_writeInt_542BB0(value C.int, file *C.FILE) C.int {
	return writeScriptWord(file, uint32(value))
}

//export nox_script_writeFloat_542BD0
func nox_script_writeFloat_542BD0(value C.float, file *C.FILE) C.int {
	return writeScriptWord(file, math.Float32bits(float32(value)))
}

func relocateScriptCallArgument(file *FILE, offset C.long, adjustment C.int) {
	nox_fs_fseek(file, offset, C.SEEK_CUR)
	value := nox_script_readInt_542B70(file)
	nox_fs_fseek(file, -4, C.SEEK_CUR)
	nox_script_writeInt_542BB0(value+adjustment, file)
}

//export nox_script_readWriteYyy_542380
func nox_script_readWriteYyy_542380(source, destination *C.FILE, relocate C.int) C.int {
	for {
		opcode, ok := readScriptInt(source)
		if !ok {
			return 0
		}
		switch asm.Op(opcode) {
		case asm.OpLoadVarInt, asm.OpLoadVarFloat, asm.OpLoadVarPtr:
			global := nox_script_readInt_542B70(source)
			index := nox_script_readInt_542B70(source)
			if relocate != 0 && global != 0 && index >= 4 {
				index += C.int(legacyWords.dword_5d4594_2489420) - 4
			}
			nox_script_writeInt_542BB0(opcode, destination)
			nox_script_writeInt_542BB0(global, destination)
			nox_script_writeInt_542BB0(index, destination)
		case asm.OpLoadVarString:
			global := nox_script_readInt_542B70(source)
			index := int8(nox_script_readInt_542B70(source))
			if relocate != 0 && global != 0 {
				index += int8(legacyWords.dword_5d4594_2489424)
			}
			nox_script_writeInt_542BB0(opcode, destination)
			nox_script_writeInt_542BB0(global, destination)
			nox_script_writeInt_542BB0(C.int(index), destination)
		case asm.OpPushInt, asm.OpPushString:
			value := int8(nox_script_readInt_542B70(source))
			if asm.Op(opcode) == asm.OpPushString && relocate != 0 {
				value += int8(legacyWords.dword_5d4594_2489424)
			}
			nox_script_writeInt_542BB0(opcode, destination)
			nox_script_writeInt_542BB0(C.int(value), destination)
		case asm.OpPushFloat:
			value := C.float(nox_script_readFloat_542B90(source))
			nox_script_writeInt_542BB0(opcode, destination)
			nox_script_writeFloat_542BD0(value, destination)
		case asm.OpJump, asm.OpJumpIf, asm.OpJumpIfNot:
			nox_script_writeInt_542BB0(opcode, destination)
			value := int8(nox_script_readInt_542B70(source))
			nox_script_writeInt_542BB0(C.int(value), destination)
		case asm.OpCallBuiltin:
			builtin := nox_script_readInt_542B70(source)
			if relocate != 0 && bool(nox_script_shouldReadMoreXxx(int32(builtin))) {
				adjustment := C.int(legacyWords.dword_5d4594_2489428) - 2
				relocateScriptCallArgument(destination, -4, adjustment)
				if bool(nox_script_shouldReadEvenMoreXxx(int32(builtin))) {
					relocateScriptCallArgument(destination, -12, adjustment)
					nox_fs_fseek(destination, 8, C.SEEK_CUR)
				}
			}
			nox_script_writeInt_542BB0(opcode, destination)
			nox_script_writeInt_542BB0(builtin, destination)
		case asm.OpCallScript:
			index := nox_script_readInt_542B70(source)
			if relocate != 0 {
				index += C.int(legacyWords.dword_5d4594_2489428) - 2
			}
			nox_script_writeInt_542BB0(opcode, destination)
			nox_script_writeInt_542BB0(index, destination)
		case asm.OpReturn:
			return nox_script_writeInt_542BB0(opcode, destination)
		default:
			if opcode < C.int(asm.OpIntAdd) || opcode > C.int(asm.OpStringAdd) {
				return opcode
			}
			nox_script_writeInt_542BB0(opcode, destination)
		}
	}
}
