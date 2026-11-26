package legacy

/*
#include <stdlib.h>
#include "GAME5_2.h"

*/
import "C"

import (
	"math"
	"time"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

// The legacy ABI exposes the first two words and links native records through
// the shared head and tail. Records live in the C heap for that ABI.
type playerProtectionRecord struct {
	token uint32
	value uint32
	next  *playerProtectionRecord
	prev  *playerProtectionRecord
}

func playerProtectionHead() *playerProtectionRecord {
	return (*playerProtectionRecord)(legacyPointers.playerProtectionHead)
}

func playerProtectionCount() *uint16 {
	return memmap.PtrUint16(0x587000, 311204)
}

//export sub_56F1C0
func sub_56F1C0() C.int {
	sub_56FF00(int32(time.Now().Unix()))
	legacyPointers.playerProtectionTail = nil
	legacyWords.dword_5d4594_2516348 = C.uint32_t(gameFrame())
	legacyPointers.playerProtectionHead = nil
	*playerProtectionCount() = 0
	legacyWords.dword_5d4594_2516356 = minimumPlayerProtectionToken
	legacyWords.dword_5d4594_2516348 ^= C.uint32_t(nox_xxx_protect_56F240())
	legacyWords.dword_5d4594_2516328 = ^legacyWords.dword_5d4594_2516348
	*memmap.PtrUint32(0x5D4594, 2516340) = uint32(nox_xxx_protectionCreateInt_56F400(0))
	sub_56F250()
	last := nox_xxx_protectionCreateInt_56F400(1)
	*memmap.PtrUint32(0x5D4594, 2516332) = uint32(last)
	return last
}

//export nox_xxx_protect_56F240
func nox_xxx_protect_56F240() C.int {
	return C.int(sub_56FF80(1, -1))
}

//export sub_56F250
func sub_56F250() C.int {
	var result C.int
	for range 7 {
		result = createPlayerProtectionRecord(uint32(legacyWords.dword_5d4594_2516356), 0)
		legacyWords.dword_5d4594_2516356++
	}
	return result
}

//export nox_xxx_protectionCreateStructForInt_56F280
func nox_xxx_protectionCreateStructForInt_56F280(token, value C.int) C.int {
	return createPlayerProtectionRecord(uint32(token), uint32(value))
}

//export nox_xxx_protectionCreateStructForFloat_56F480
func nox_xxx_protectionCreateStructForFloat_56F480(token C.int, value C.float) C.int {
	return createPlayerProtectionRecord(uint32(token), math.Float32bits(float32(value)))
}

func createPlayerProtectionRecord(token, value uint32) C.int {
	node := (*playerProtectionRecord)(C.calloc(1, C.size_t(unsafe.Sizeof(playerProtectionRecord{}))))
	if node == nil {
		return 0
	}
	mask := uint32(legacyWords.dword_5d4594_2516348)
	node.token = token ^ mask
	node.value = value ^ mask
	legacyWords.dword_5d4594_2516328 ^= C.uint32_t(node.token ^ node.value)
	return insertPlayerProtectionRecord(node)
}

//export sub_56F2F0
func sub_56F2F0(words *C.uint32_t) C.int {
	return insertPlayerProtectionRecord((*playerProtectionRecord)(unsafe.Pointer(words)))
}

func insertPlayerProtectionRecord(node *playerProtectionRecord) C.int {
	count := playerProtectionCount()
	if *count == 0 {
		*count++
		legacyPointers.playerProtectionTail = unsafe.Pointer(node)
		legacyPointers.playerProtectionHead = unsafe.Pointer(node)
		return 1
	}

	before := playerProtectionHead()
	index := uint16(nox_common_randomInt_415FA0(0, int32(*count)-1))
	for i := uint16(0); before != nil && i != index; i++ {
		before = before.next
	}
	node.prev = before.prev
	if legacyWords.dword_5d4594_2516348 == 0 {
		nullsub_31(1)
	}
	node.next = before
	before.prev = node
	if node.prev != nil {
		node.prev.next = node
		*count++
	} else {
		*count++
		legacyPointers.playerProtectionHead = unsafe.Pointer(node)
	}
	return 1
}

//export sub_56F3B0
func sub_56F3B0() *C.uint32_t {
	for node := playerProtectionHead(); node != nil; {
		next := node.next
		C.free(unsafe.Pointer(node))
		node = next
	}
	legacyWords.dword_5d4594_2516328 = 0
	*playerProtectionCount() = 0
	legacyWords.dword_5d4594_2516348 = 0
	legacyPointers.playerProtectionTail = nil
	legacyPointers.playerProtectionHead = nil
	return nil
}

//export nox_xxx_protectionCreateInt_56F400
func nox_xxx_protectionCreateInt_56F400(value C.int) C.int {
	if createPlayerProtectionRecord(uint32(legacyWords.dword_5d4594_2516356), uint32(value)) == 0 {
		nullsub_31(1)
		return 0
	}
	token := legacyWords.dword_5d4594_2516356
	legacyWords.dword_5d4594_2516356++
	return C.int(token)
}

//export sub_56F4F0
func sub_56F4F0(token *C.int) C.int {
	result := sub_56F510(*token)
	if result != 0 {
		*token = 0
	}
	return result
}

//export sub_56F510
func sub_56F510(token C.int) C.int {
	node := findPlayerProtectionRecord(uint32(token))
	if node == nil {
		return 0
	}
	if node.prev != nil {
		node.prev.next = node.next
	} else {
		legacyPointers.playerProtectionHead = unsafe.Pointer(node.next)
	}
	if node.next != nil {
		node.next.prev = node.prev
	} else {
		legacyPointers.playerProtectionTail = unsafe.Pointer(node.prev)
	}
	legacyWords.dword_5d4594_2516328 ^= C.uint32_t(node.token ^ node.value)
	*playerProtectionCount()--
	C.free(unsafe.Pointer(node))
	return 1
}

//export sub_56F590
func sub_56F590(token C.int) *C.uint32_t {
	return (*C.uint32_t)(unsafe.Pointer(findPlayerProtectionRecord(uint32(token))))
}

func findPlayerProtectionRecord(token uint32) *playerProtectionRecord {
	encoded := token ^ uint32(legacyWords.dword_5d4594_2516348)
	for node := playerProtectionHead(); node != nil; node = node.next {
		if node.token == encoded {
			return node
		}
	}
	nullsub_31(1)
	return nil
}

//export nox_xxx_protectData_56F5C0
func nox_xxx_protectData_56F5C0() C.int {
	if legacyWords.dword_5d4594_2516348 == 0 {
		nullsub_31(1)
	}
	frame := gameFrame()
	oldMask := uint32(legacyWords.dword_5d4594_2516348)
	newMask := frame ^ uint32(nox_xxx_protect_56F240())
	delta := oldMask ^ newMask
	legacyWords.dword_5d4594_2516328 = C.uint32_t(^newMask)
	count := playerProtectionCount()
	for i := int32(0); i < int32(*count)>>2; i++ {
		low := nox_common_randomInt_415FA0(0, int32(*count)>>1)
		size := int32(*count)
		high := nox_common_randomInt_415FA0((size>>1)+1, size-1)
		if low != high {
			second := sub_56F6F0(C.int(high))
			first := sub_56F6F0(C.int(low))
			sub_56F720((*C.int)(unsafe.Pointer(first)), (*C.int)(unsafe.Pointer(second)))
		}
	}

	legacyWords.dword_5d4594_2516348 = 0
	for node := playerProtectionHead(); node != nil; node = node.next {
		node.token ^= delta
		node.value ^= delta
		legacyWords.dword_5d4594_2516328 ^= C.uint32_t(node.token ^ node.value)
	}
	*memmap.PtrUint32(0x5D4594, 2516364)++
	legacyWords.dword_5d4594_2516348 = C.uint32_t(newMask)
	if newMask == 0 {
		nullsub_31(1)
	}
	return C.int(newMask)
}

//export sub_56F6F0
func sub_56F6F0(index C.int) *C.uint32_t {
	for node := playerProtectionHead(); node != nil; node = node.next {
		if index == 0 {
			return (*C.uint32_t)(unsafe.Pointer(node))
		}
		index--
	}
	nullsub_31(1)
	return nil
}

//export sub_56F720
func sub_56F720(first, second *C.int) {
	if first == nil || second == nil {
		nullsub_31(1)
		return
	}
	a, b := (*[2]C.int)(unsafe.Pointer(first)), (*[2]C.int)(unsafe.Pointer(second))
	old := *a
	a[0] = b[0]
	a[1] = b[1]
	*b = old
	*memmap.PtrUint32(0x5D4594, 2516360)++
	if legacyWords.dword_5d4594_2516348 == 0 {
		nullsub_31(1)
	}
}
