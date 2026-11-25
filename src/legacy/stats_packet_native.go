package legacy

/*
#include <stdlib.h>
#include "GAME1_2.h"

extern uint32_t dword_5d4594_741648;
extern uint32_t dword_5d4594_741652;
*/
import "C"

import (
	"encoding/binary"
	"time"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func statsCompressRLE(input []byte) []byte {
	var frequency [256]int
	for _, value := range input {
		frequency[value]++
	}
	var marker byte
	for i, count := range frequency {
		if count == 0 {
			marker = byte(i)
			break
		}
		if count < frequency[marker] {
			marker = byte(i)
		}
	}

	output := make([]byte, 1, len(input)+1)
	output[0] = marker
	for start := 0; start < len(input); {
		value := input[start]
		end := start + 1
		for end < len(input) && input[end] == value {
			end++
		}
		// The original counts each complete run in a single wrapping byte.
		run := byte(end - start)
		if run > 3 {
			output = append(output, marker, run, value)
		} else {
			for range run {
				if value == marker {
					output = append(output, marker)
				}
				output = append(output, value)
			}
		}
		start = end
	}
	return output
}

func statsRandomIndexCells() (first, second *uint32) {
	return (*uint32)(unsafe.Pointer(&C.dword_5d4594_741648)), (*uint32)(unsafe.Pointer(&C.dword_5d4594_741652))
}

func statsNextRandom(seed *int32) float64 {
	const modulus int32 = 1000000000
	const initConstant int32 = 161803398
	table := memmap.PtrT[[56]int32](0x5D4594, 741384)
	ready := memmap.PtrUint32(0x5D4594, 741656)
	first, second := statsRandomIndexCells()

	if *seed < 0 || *ready == 0 {
		*ready = 1
		start := *seed
		if start < 0 {
			start = -start
		}
		last := (initConstant - start) % modulus
		table[55] = last
		next := int32(1)
		for i := 1; i < 55; i++ {
			index := (21 * i) % 55
			table[index] = next
			next, last = last-next, next
			if next < 0 {
				next += modulus
			}
		}
		for range 4 {
			for i := 1; i <= 55; i++ {
				diff := table[i] - table[1+(i+30)%55]
				if diff < 0 {
					diff += modulus
				}
				table[i] = diff
			}
		}
		*first, *second, *seed = 0, 31, 1
	}
	*first++
	*second++
	if *first == 56 {
		*first = 1
	}
	if *second == 56 {
		*second = 1
	}
	next := table[*first] - table[*second]
	if next < 0 {
		next += modulus
	}
	table[*first] = next
	return mulFloat64TowardZero(float64(next), memmap.Float64(0x581450, 8368))
}

func statsPacketSeed(unix int64) int32 {
	seed := int32(unix)
	if seed > 0 {
		seed = -seed
	}
	return seed
}

func statsPacketScramble(input []byte, currentTime func() int64) []byte {
	if len(input) < 15 {
		return nil
	}
	seed := statsPacketSeed(currentTime())
	key := uint32(seed)
	mask := make([]byte, len(input))
	random := statsNextRandom(&seed)
	for i := range mask {
		value := x87TruncQwordLow(random * 255)
		if value < 0 {
			value = -value
		}
		mask[i] = byte(value)
		random = statsNextRandom(&seed)
	}

	selectorSeed := statsPacketSeed(currentTime())
	factor := 241
	if len(input) < 241 {
		factor = len(input) - 14
	}
	selector := byte(x87TruncQword(statsNextRandom(&selectorSeed)*float64(factor)) + 10)
	output := make([]byte, len(input)+5)
	position := 0
	for i, value := range input {
		if position == 5 {
			output[position] = selector
			position++
		}
		if position == int(selector) {
			binary.BigEndian.PutUint32(output[position:], key)
			position += 4
		}
		output[position] = value ^ mask[i]
		position++
	}
	return output
}

func statsPacketUnixTime() int64 { return time.Now().Unix() }

//export sub_42AAA0
func sub_42AAA0(seed *C.int) C.double {
	return C.double(statsNextRandom((*int32)(unsafe.Pointer(seed))))
}

//export sub_42A970
func sub_42A970(input, output *C.uint8_t, length *C.int) C.int {
	data := unsafe.Slice((*byte)(unsafe.Pointer(input)), max(0, int(*length)))
	encoded := statsCompressRLE(data)
	copy(unsafe.Slice((*byte)(unsafe.Pointer(output)), len(encoded)), encoded)
	*length = C.int(len(encoded))
	return *length
}

//export sub_42AC50
func sub_42AC50(input *C.uint8_t, length *C.size_t) *C.uint8_t {
	// This ABI passes a 32-bit length slot on each architecture.
	word := (*uint32)(unsafe.Pointer(length))
	if int32(*word) < 15 {
		*word = ^uint32(1)
		return nil
	}
	data := unsafe.Slice((*byte)(unsafe.Pointer(input)), int(*word))
	encoded := statsPacketScramble(data, statsPacketUnixTime)
	out := C.calloc(1, C.size_t(len(encoded)))
	if out == nil {
		return nil
	}
	copy(unsafe.Slice((*byte)(out), len(encoded)), encoded)
	*word = uint32(len(encoded))
	return (*C.uint8_t)(out)
}

//export sub_42A8B0
func sub_42A8B0(input *C.uint8_t, length *C.int) *C.uint16_t {
	if length == nil || input == nil && *length > 0 {
		return nil
	}
	data := unsafe.Slice((*byte)(unsafe.Pointer(input)), max(0, int(*length)))
	encoded := statsPacketScramble(statsCompressRLE(data), statsPacketUnixTime)
	if encoded == nil {
		*length = -2
		return nil
	}

	fieldLength := int(uint16(len(encoded)))
	paddedLength := (fieldLength + 3) &^ 3
	packetLength := 12 + paddedLength
	ptr := C.calloc(1, C.size_t(packetLength))
	if ptr == nil {
		return nil
	}
	packet := unsafe.Slice((*byte)(ptr), packetLength)
	binary.BigEndian.PutUint16(packet, uint16(packetLength))
	tag := memmap.PtrT[[4]byte](0x587000, 71480)
	copy(packet[4:8], tag[:alloc.StrLenS(tag[:])])
	binary.BigEndian.PutUint16(packet[8:], 20)
	binary.BigEndian.PutUint16(packet[10:], uint16(fieldLength))
	copy(packet[12:], encoded[:fieldLength])
	*length = C.int(packetLength)
	return (*C.uint16_t)(ptr)
}
