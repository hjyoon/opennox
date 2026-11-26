//go:build !server

package ail

import (
	"encoding/binary"
	"testing"
)

func TestMP3FrameSyncVariableBitrate(t *testing.T) {
	first := mp3FrameHeader{0xff, 0xfb, 0x90, 0x00}
	second := mp3FrameHeader{0xff, 0xfa, 0xa2, 0xc0}
	const prefix, firstSize, secondSize = 7, 417, 523
	data := make([]byte, prefix+firstSize+secondSize)
	copy(data[prefix:], first[:])
	copy(data[prefix+firstSize:], second[:])

	var sync mp3FrameSync
	skip, size, reset := sync.next(data)
	if !reset || skip != prefix || size != firstSize || sync.header != first {
		t.Fatalf("first frame = (%d, %d, %t, %x)", skip, size, reset, sync.header)
	}
	skip, size, reset = sync.next(data[skip+size:])
	if reset || skip != 0 || size != secondSize || sync.header != second {
		t.Fatalf("second frame = (%d, %d, %t, %x)", skip, size, reset, sync.header)
	}
}

func TestMP3FrameSyncFreeFormat(t *testing.T) {
	first := mp3FrameHeader{0xff, 0xf3, 0x02, 0xc0}
	unpadded := mp3FrameHeader{0xff, 0xf3, 0x00, 0xc0}
	const size = 32
	data := make([]byte, size*3+2)
	copy(data, first[:])
	copy(data[size+1:], unpadded[:])
	copy(data[2*size+1:], first[:])

	var sync mp3FrameSync
	for i, want := range []int{size + 1, size, size + 1} {
		skip, got, reset := sync.next(data)
		if skip != 0 || got != want || reset != (i == 0) || sync.freeFormatBytes != size {
			t.Fatalf("frame %d = (%d, %d, %t), free frame bytes = %d", i, skip, got, reset, sync.freeFormatBytes)
		}
		data = data[got:]
	}

	skip, got, reset := sync.next([]byte{0xff, 0xf3, 0x02})
	if skip != 0 || got != 0 || !reset || sync != (mp3FrameSync{}) {
		t.Fatalf("truncated header = (%d, %d, %t), state = %+v", skip, got, reset, sync)
	}
}

func TestWAVMP3TruncatedHeaderEOF(t *testing.T) {
	wav := mp3FrameWAV(t, 0)
	dataOffset := 20 + int(binary.LittleEndian.Uint32(wav[16:])) + 8
	wav = append(wav[:dataOffset], 0xff, 0xf3, 0x02, 0)
	binary.LittleEndian.PutUint32(wav[4:], uint32(len(wav)-8))
	binary.LittleEndian.PutUint32(wav[dataOffset-4:], 3)

	r := openMP3FrameFixture(t, wav)
	if n, ok := r.Decode(make([]int16, 1152*2)); n != 0 || ok || r.Position() != 0 {
		t.Fatalf("truncated header decoded %d samples, ok=%t, position=%d", n, ok, r.Position())
	}
}
