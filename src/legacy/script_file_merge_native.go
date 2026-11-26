package legacy

/*
#include <stdio.h>
#include "server__script__file.h"

*/
import "C"

import "io"

func transferScriptBytes(source, destination *FILE, size int64) bool {
	if size < 0 {
		return false
	}
	var writer io.Writer = io.Discard
	if destination != nil {
		writer = fileByHandle(destination)
	}
	_, err := io.CopyN(writer, fileByHandle(source), size)
	return err == nil
}

func transferScriptPairHeader(first, second, destination *FILE, size int64) bool {
	return transferScriptBytes(first, nil, size) && transferScriptBytes(second, destination, size)
}

func transferLengthPrefixedScriptBytes(source, destination *FILE) bool {
	size, ok := readScriptInt(source)
	if !ok || size < 0 {
		return false
	}
	if destination != nil && nox_script_writeInt_542BB0(size, destination) != 1 {
		return false
	}
	return transferScriptBytes(source, destination, int64(size))
}

//export nox_script_readWriteJjj_5418C0
func nox_script_readWriteJjj_5418C0(first, second, destination *C.FILE) C.int {
	if !transferScriptPairHeader(first, second, destination, 4) {
		return 0
	}
	countFirst, firstOK := readScriptInt(first)
	countSecond, secondOK := readScriptInt(second)
	if !firstOK || !secondOK || countFirst < 0 || countSecond < 0 {
		return 0
	}
	legacyWords.dword_5d4594_2489424 = C.uint(countFirst)
	nox_script_writeInt_542BB0(countFirst+countSecond, destination)
	for _, input := range [...]struct {
		file  *FILE
		count C.int
	}{{first, countFirst}, {second, countSecond}} {
		for range input.count {
			if !transferLengthPrefixedScriptBytes(input.file, destination) {
				return 0
			}
		}
	}
	return countSecond
}

func transferScriptGlobalsPrefix(first, second, destination *FILE) bool {
	return transferScriptPairHeader(first, second, destination, 4) &&
		transferLengthPrefixedScriptBytes(first, nil) &&
		transferLengthPrefixedScriptBytes(second, destination) &&
		transferScriptPairHeader(first, second, destination, 8) &&
		transferScriptPairHeader(first, second, destination, 4)
}

//export nox_script_readWriteVvv_541E40
func nox_script_readWriteVvv_541E40(first, second, destination *C.FILE) C.int {
	if !transferScriptGlobalsPrefix(first, second, destination) {
		return 0
	}
	countFirst, firstOK := readScriptInt(first)
	countSecond, secondOK := readScriptInt(second)
	if !firstOK || !secondOK || countFirst < 0 || countSecond < 0 {
		return 0
	}
	nox_script_writeInt_542BB0(countFirst+countSecond, destination)
	if !transferScriptPairHeader(first, second, destination, 4) ||
		!transferScriptPairHeader(first, second, destination, 4) {
		return 0
	}
	codeFirst, firstOK := readScriptInt(first)
	_, secondOK = readScriptInt(second)
	if !firstOK || !secondOK {
		return 0
	}
	nox_script_writeInt_542BB0(codeFirst, destination)
	if !transferScriptPairHeader(first, second, destination, 4) ||
		!transferScriptGlobalsPrefix(first, second, destination) {
		return 0
	}
	countFirst, firstOK = readScriptInt(first)
	countSecond, secondOK = readScriptInt(second)
	if !firstOK || !secondOK || countFirst < 0 || countSecond < 4 {
		return 0
	}
	legacyWords.dword_5d4594_2489420 = C.uint(countFirst)
	nox_script_writeInt_542BB0(countFirst+countSecond-4, destination)
	if !transferScriptPairHeader(first, second, destination, 4) ||
		!transferScriptBytes(first, destination, 4*int64(countFirst)) ||
		!transferScriptBytes(second, nil, 16) ||
		!transferScriptBytes(second, destination, 4*int64(countSecond-4)) ||
		!transferScriptPairHeader(first, second, destination, 4) {
		return 0
	}
	codeFirst, firstOK = readScriptInt(first)
	codeSecond, secondOK := readScriptInt(second)
	if !firstOK || !secondOK {
		return 0
	}
	nox_script_writeInt_542BB0(codeFirst+codeSecond-4, destination)
	nox_script_readWriteYyy_542380(first, destination, 0)
	nox_fs_fseek(destination, -4, C.SEEK_CUR)
	return nox_script_readWriteYyy_542380(second, destination, 1)
}

//export nox_script_readWriteIii_541D80
func nox_script_readWriteIii_541D80(source, destination *C.FILE) C.int {
	if !transferScriptBytes(source, destination, 4) {
		return 0
	}
	count, ok := readScriptInt(source)
	if !ok {
		return 0
	}
	nox_script_writeInt_542BB0(count, destination)
	if !transferScriptBytes(source, destination, 4+4*int64(max(0, count))) {
		return 0
	}
	return count
}

func transferScriptFunction(source, destination *FILE, relocate C.int) {
	if !transferScriptBytes(source, destination, 4) ||
		!transferLengthPrefixedScriptBytes(source, destination) ||
		!transferScriptBytes(source, destination, 8) {
		return
	}
	nox_script_readWriteIii_541D80(source, destination)
	if transferScriptBytes(source, destination, 8) {
		nox_script_readWriteYyy_542380(source, destination, relocate)
	}
}

//export nox_script_readWriteXxx_541A50
func nox_script_readWriteXxx_541A50(first, second, destination *C.FILE) C.int {
	if !transferScriptPairHeader(first, second, destination, 4) {
		return 0
	}
	countFirst, firstOK := readScriptInt(first)
	countSecond, secondOK := readScriptInt(second)
	if !firstOK || !secondOK || countFirst < 2 || countSecond < 2 {
		return 0
	}
	legacyWords.dword_5d4594_2489428 = C.uint(countFirst)
	nox_script_writeInt_542BB0(countFirst+countSecond-2, destination)
	result := nox_script_readWriteVvv_541E40(first, second, destination)
	for mode, input := range [...]struct {
		file  *FILE
		count C.int
	}{{first, countFirst - 2}, {second, countSecond - 2}} {
		for range input.count {
			transferScriptFunction(input.file, destination, C.int(mode))
			result = input.count
		}
	}
	return result
}

//export nox_script_readWriteWww_5417C0
func nox_script_readWriteWww_5417C0(first, second, destination *C.FILE) C.size_t {
	legacyWords.dword_5d4594_2489420, legacyWords.dword_5d4594_2489424, legacyWords.dword_5d4594_2489428 = 0, 0, 0
	if !transferScriptPairHeader(first, second, destination, 8) {
		return 0
	}
	nox_script_readWriteJjj_5418C0(first, second, destination)
	nox_script_readWriteXxx_541A50(first, second, destination)
	if !transferScriptPairHeader(first, second, destination, 4) {
		return 0
	}
	return 4
}
