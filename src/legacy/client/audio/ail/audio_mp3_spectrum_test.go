//go:build !server

package ail

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"runtime"
	"testing"
)

func TestMP3EightKHzMixedSpectrum(t *testing.T) {
	header := mp3FrameHeader{0xff, 0xe3, 0x98, 0x70}
	mixed := mp3Granule{
		bandWidths:      mp3MixedBandWidths[1][:],
		blockType:       mp3ShortBlock,
		mixedBlock:      1,
		longBands:       6,
		shortBands:      30,
		regionBandCount: [3]uint8{7, 255, 0},
		subblockGain:    [3]uint8{1, 2, 5},
		preflag:         1,
		scalefacScale:   1,
		countOneTable:   1,
	}
	left, right := mixed, mixed
	left.dataBits, left.bigValues, left.scalefacCompress, left.globalGain = 1713, 61, 421, 163
	left.huffmanTable = [3]uint8{31, 24, 7}
	right.dataBits, right.bigValues, right.scalefacCompress, right.globalGain = 1828, 37, 287, 142
	right.huffmanTable = [3]uint8{15, 28, 3}

	var payload [mp3ReservoirSize + mp3MaxFreeFormatBytes]byte
	var seed uint32 = 28732489
	for i := range payload {
		seed = seed*1664525 + 1013904223
		payload[i] = byte(seed >> 9)
	}
	bits := mp3BitReader{data: payload[:]}
	var coefficients mp3SpectralWorkspace
	var intensity [2][39]uint8
	coefficients.unpack(header, []mp3Granule{left, right}, &bits, &intensity)
	if bits.pos != int(left.dataBits)+int(right.dataBits) {
		t.Fatalf("consumed %d bits", bits.pos)
	}

	data := make([]byte, 0, 8*mp3SpectralSamples)
	for _, v := range coefficients[:2*mp3SpectralSamples] {
		data = binary.LittleEndian.AppendUint32(data, math.Float32bits(v))
	}

	// The C reference initializes its scratch storage to make this case reproducible.
	var want string
	switch {
	case mp3LinuxARM64Contract:
		want = "05a30ad17ee93cc3f29bdabe919119cc0b2b0e4eca326d021a32083cd2690b63"
	case runtime.GOARCH == "arm64":
		want = "8050544b5ac9c2e69ee16ecc738c34a7ef6c5b03e5397beaa63b006c7664e937"
	default:
		want = "6158e80c80bb0ab97416b6cac45f79345cba9dae0bd544c8b20f3df9c1e812e6"
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(data)); got != want {
		t.Fatalf("8 kHz MP3 spectrum hash = %s, want %s", got, want)
	}
}
