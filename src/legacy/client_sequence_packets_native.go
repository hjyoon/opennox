package legacy

/*
#include "GAME2_3.h"
typedef const uint8_t nox_seq_const_byte;
*/
import "C"

import (
	"bytes"
	"container/list"
	"encoding/binary"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

type clientSequencePacket struct {
	sequence uint16
	arrived  uint64
	data     []byte
}

type clientSequenceQueue struct {
	packets          list.List
	current, pending *list.Element
	generation       uint64
}

var clientSequences clientSequenceQueue

func clientSequenceNumber() *uint16 {
	return memmap.PtrUint16(0x5D4594, 1197360)
}

func clientOldSequence(expected, sequence uint16) bool {
	const window = 10000
	if expected < window && sequence >= 0xffff-(window-expected) {
		return true
	}
	return sequence < expected && sequence >= expected-min(expected, window)
}

func clientSequenceTicks() uint64 {
	return uint64(uint32(PlatformTicks()))
}

func (q *clientSequenceQueue) find(sequence uint16) (before *list.Element, index int) {
	for before = q.packets.Front(); before != nil; before = before.Next() {
		if before.Value.(clientSequencePacket).sequence >= sequence {
			break
		}
		index++
	}
	return before, index
}

func (q *clientSequenceQueue) add(expected, sequence uint16, data []byte) int {
	if clientOldSequence(expected, sequence) {
		return 1
	}
	before, index := q.find(sequence)
	if before != nil && before.Value.(clientSequencePacket).sequence == sequence {
		return 1
	}
	packet := clientSequencePacket{sequence: sequence, arrived: clientSequenceTicks(), data: bytes.Clone(data)}
	var inserted *list.Element
	if before != nil {
		inserted = q.packets.InsertBefore(packet, before)
	} else {
		inserted = q.packets.PushBack(packet)
	}
	if expected == sequence {
		q.current = inserted
	}
	return index
}

func (q *clientSequenceQueue) drain(expected *uint16, deliver func([]byte)) int32 {
	var result uint64
	if q.current == nil && q.pending != nil {
		packet := q.pending.Value.(clientSequencePacket)
		result = clientSequenceTicks() - packet.arrived
		if result > 30000 {
			*expected = packet.sequence
			q.current = q.pending
			q.pending = q.pending.Next()
		}
	}

	entry := q.current
	generation := q.generation
	for entry != nil {
		packet := entry.Value.(clientSequencePacket)
		result = uint64(packet.sequence)
		if packet.sequence != *expected {
			break
		}
		next := entry.Next()
		if next == nil {
			next = q.packets.Front()
			if next == entry {
				next = nil
			}
		}
		deliver(packet.data)
		if q.generation != generation {
			return int32(result)
		}
		*expected++
		q.packets.Remove(entry)
		entry = next
	}
	// The legacy drain remembers only the stopping point of this pass.
	q.current, q.pending = nil, entry
	return int32(result)
}

func resetClientSequencePackets() {
	clientSequences.packets.Init()
	clientSequences.current, clientSequences.pending = nil, nil
	clientSequences.generation++
	*clientSequenceNumber() = 0
}

func queueClientSequencePacket(data []byte) int {
	return clientSequences.add(*clientSequenceNumber(), binary.LittleEndian.Uint16(data[1:3]), data[4:])
}

// ClientSequencePacket48D5A0 handles a complete MSG_SEQ_IMPORTANT envelope.
func ClientSequencePacket48D5A0(data []byte, connected bool) int {
	if len(data) < 4 {
		return -1
	}
	size := 4 + int(data[3])
	if len(data) < size {
		return -1
	}
	if connected {
		queueClientSequencePacket(data[:size])
	}
	return size
}

//export sub_48D4F0
func sub_48D4F0(expected, sequence C.ushort) C.int {
	return C.int(bool2int(clientOldSequence(uint16(expected), uint16(sequence))))
}

//export sub_48D560
func sub_48D560(sequence C.ushort) C.int {
	match, _ := clientSequences.find(uint16(sequence))
	return C.int(bool2int(match != nil && match.Value.(clientSequencePacket).sequence == uint16(sequence)))
}

//export sub_48D5A0
func sub_48D5A0(data *C.nox_seq_const_byte) C.int {
	if data == nil {
		return 0
	}
	p := (*byte)(unsafe.Pointer(data))
	size := 4 + int(*(*byte)(unsafe.Add(unsafe.Pointer(data), 3)))
	return C.int(queueClientSequencePacket(unsafe.Slice(p, size)))
}

//export sub_48D660
func sub_48D660() C.int {
	return C.int(clientSequences.drain(clientSequenceNumber(), func(data []byte) {
		Nox_xxx_netOnPacketRecvCli_48EA70(server.HostPlayerIndex, unsafe.SliceData(data), len(data))
	}))
}

//export sub_48D740
func sub_48D740() C.int {
	resetClientSequencePackets()
	return 0
}

//export sub_48D760
func sub_48D760() {
	resetClientSequencePackets()
}

//export sub_48D7B0
func sub_48D7B0() *C.int { return nil }
