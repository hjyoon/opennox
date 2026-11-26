package legacy

/*
#include "GAME5_2.h"

typedef const nox_object_t protection_const_object_t;

*/
import "C"

import (
	"encoding/binary"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// PlayerProtectionChecksum XORs the complete little-endian words in data.
func PlayerProtectionChecksum(data []byte) uint32 {
	return playerProtectionChecksumWords(unsafe.Pointer(unsafe.SliceData(data)), uintptr(len(data)/4))
}

func playerProtectionChecksumWords(data unsafe.Pointer, count uintptr) uint32 {
	var checksum uint32
	for i := uintptr(0); i < count; i++ {
		word := (*[4]byte)(unsafe.Add(data, 4*i))
		checksum ^= binary.LittleEndian.Uint32(word[:])
	}
	return checksum
}

//export nox_xxx_protectionStringCRC_56FAC0
func nox_xxx_protectionStringCRC_56FAC0(data *C.int, length C.uint) C.int {
	return C.int(playerProtectionChecksumWords(unsafe.Pointer(data), uintptr(length>>2)))
}

//export nox_xxx_protectionStringCRCLen_56FAE0
func nox_xxx_protectionStringCRCLen_56FAE0(data *C.int, length C.uint) C.int {
	if data == nil {
		return 0
	}
	return nox_xxx_protectionStringCRC_56FAC0(data, length)
}

func playerProtectionChecksumRecord(token int32) *playerProtectionRecord {
	if token < minimumPlayerProtectionToken {
		return nil
	}
	record := findPlayerProtectionRecord(uint32(token))
	if record == nil {
		nullsub_31(1)
	}
	return record
}

//export sub_56FB00
func sub_56FB00(data *C.int, length C.uint, token C.int) C.int {
	record := playerProtectionChecksumRecord(int32(token))
	if record == nil {
		return 0
	}
	if C.uint32_t(nox_xxx_protectionStringCRCLen_56FAE0(data, length))^legacyWords.dword_5d4594_2516348 == C.uint32_t(record.value) {
		return 1
	}
	nullsub_31(1)
	return 0
}

func playerProtectionItemChecksum(item *server.Object) uint32 {
	if item == nil {
		return 0
	}
	checksum := objectTraceNetCode4E4C00(item)
	checksum ^= uint32(Nox_xxx_unitGetHP_4EE780(item))
	checksum ^= uint32(objectTraceType4E4C10(item))
	data := objectTraceInit4E4C30(item)
	size := nox_xxx_objectGetInitDataSize_4E4C50(asObjectC(item))
	if data != nil {
		checksum ^= playerProtectionChecksumWords(data, uintptr(size>>2))
	}
	id := objectTraceID4E4C80(item)
	return checksum ^ playerProtectionChecksumWords(id, uintptr(alloc.Strlen(id)/4))
}

//export sub_56FB60
func sub_56FB60(item *C.nox_object_t) C.int {
	return C.int(playerProtectionItemChecksum(asObjectS((*nox_object_t)(item))))
}

func updatePlayerProtectionItem(token int32, item *server.Object) int32 {
	if token < minimumPlayerProtectionToken {
		return token
	}
	record := playerProtectionChecksumRecord(token)
	if record == nil {
		return 0
	}
	legacyWords.dword_5d4594_2516328 ^= C.uint32_t(record.value)
	checksum := playerProtectionItemChecksum(item)
	record.value ^= checksum
	legacyWords.dword_5d4594_2516328 ^= C.uint32_t(record.value)
	return int32(legacyWords.dword_5d4594_2516328)
}

//export nox_xxx_protect_56FBF0
func nox_xxx_protect_56FBF0(token C.int, item *C.nox_object_t) C.int {
	return C.int(updatePlayerProtectionItem(int32(token), asObjectS((*nox_object_t)(item))))
}

//export nox_xxx_protect_56FC50
func nox_xxx_protect_56FC50(token C.int, item *C.protection_const_object_t) C.int {
	return C.int(updatePlayerProtectionItem(int32(token), asObjectS((*nox_object_t)(unsafe.Pointer(item)))))
}

func playerProtectionIndexBit(index, level int32) uint32 {
	if level == 0 {
		return 0
	}
	// The original x86 shift retains the low five bits of the index.
	return uint32(1) << (uint32(index) & 31)
}

//export sub_56FCB0
func sub_56FCB0(index, level C.int) C.int {
	return C.int(playerProtectionIndexBit(int32(index), int32(level)))
}

//export nox_xxx_playerAwardSpellProtectionCRC_56FCE0
func nox_xxx_playerAwardSpellProtectionCRC_56FCE0(token, index, level C.int) C.int {
	if token < minimumPlayerProtectionToken {
		return token
	}
	record := playerProtectionChecksumRecord(int32(token))
	if record == nil {
		return 0
	}
	legacyWords.dword_5d4594_2516328 ^= C.uint32_t(record.value)
	key := uint32(legacyWords.dword_5d4594_2516348)
	record.value = ((record.value ^ key) | playerProtectionIndexBit(int32(index), int32(level))) ^ key
	legacyWords.dword_5d4594_2516328 ^= C.uint32_t(record.value)
	return C.int(record.value)
}

//export nox_xxx_playerApplyProtectionCRC_56FD50
func nox_xxx_playerApplyProtectionCRC_56FD50(token C.int, levels unsafe.Pointer, count C.int) C.int {
	record := playerProtectionChecksumRecord(int32(token))
	if record == nil {
		return 0
	}
	var bits uint32
	for i := int32(1); i < int32(count); i++ {
		word := (*[4]byte)(unsafe.Add(levels, uintptr(i)*4))
		bits |= playerProtectionIndexBit(i, int32(binary.LittleEndian.Uint32(word[:])))
	}
	if bits^uint32(legacyWords.dword_5d4594_2516348) == record.value {
		return 1
	}
	nullsub_31(1)
	return 0
}
