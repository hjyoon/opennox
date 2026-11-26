//go:build !server

package ail

import (
	"bytes"
	"testing"
)

func TestMP3BitReservoirRebuild(t *testing.T) {
	var reservoir mp3BitReservoir
	var scratch [mp3ReservoirSize + mp3MaxFreeFormatBytes]byte
	data, ready := reservoir.restore(scratch[:], []byte{10, 11, 12, 13}, 3)
	if ready || !bytes.Equal(data, []byte{10, 11, 12, 13}) {
		t.Fatalf("missing first-frame prefix = %v, ready=%t", data, ready)
	}
	reservoir.save(data, 0)

	data, ready = reservoir.restore(scratch[:], []byte{14, 15, 16}, 3)
	if !ready || !bytes.Equal(data, []byte{11, 12, 13, 14, 15, 16}) {
		t.Fatalf("assembled frame = %v, ready=%t", data, ready)
	}
	reservoir.save(data, 9)
	data, ready = reservoir.restore(scratch[:], []byte{17}, 5)
	if ready || !bytes.Equal(data, []byte{13, 14, 15, 16, 17}) {
		t.Fatalf("short prefix after consuming two bytes = %v, ready=%t", data, ready)
	}

	payload := make([]byte, mp3ReservoirSize+16)
	for i := range payload {
		payload[i] = byte(i)
	}
	reservoir.save(payload, 8)
	data, ready = reservoir.restore(scratch[:], nil, mp3ReservoirSize)
	if !ready || !bytes.Equal(data, payload[16:]) {
		t.Fatal("reservoir did not retain the newest 511 bytes")
	}
}
