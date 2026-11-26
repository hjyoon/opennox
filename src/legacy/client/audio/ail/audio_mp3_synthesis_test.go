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

func TestMP3SynthesisScalarReference(t *testing.T) {
	const granules = 12
	const recordSize = 2*2*mp3SpectralSamples + 4*(2*mp3SpectralSamples+mp3QMFSaved)
	data := make([]byte, 0, granules*recordSize)
	var spectrum [2][mp3SpectralSamples]float32
	var state mp3SynthesisState
	var seed uint32 = 1237788
	sample := func(gain float32) float32 {
		seed = seed*1664525 + 1013904223
		return float32(int32(seed>>14)-131072) * gain
	}
	for i := range granules {
		gain := [4]float32{0x1p-20, 0x1p-22, 0x1p-24, 0x1p-17}[i%4]
		channels := 1
		if i%3 != 0 {
			channels = 2
		}
		var pcm [2 * mp3SpectralSamples]int16
		for ch := range spectrum {
			for j := range spectrum[ch] {
				spectrum[ch][j] = sample(gain)
			}
		}
		if i == 0 {
			for j := range state {
				state[j] = sample(gain)
			}
		}
		state.synthesize(&spectrum[0], &spectrum[1], channels, pcm[:])
		for _, v := range pcm {
			data = binary.LittleEndian.AppendUint16(data, uint16(v))
		}
		for ch := range spectrum {
			for _, v := range spectrum[ch] {
				data = binary.LittleEndian.AppendUint32(data, math.Float32bits(v))
			}
		}
		for _, v := range state {
			data = binary.LittleEndian.AppendUint32(data, math.Float32bits(v))
		}
	}

	var want string
	switch {
	case mp3LinuxARM64Contract:
		want = "04246f2b3f006c2f7a86a72533f267715990e1e6fc9bb6b9a8e2bf790739048e"
	case runtime.GOARCH == "arm64":
		want = "4bf734999bdf00ca895d27769bfe2f43443108bd9058e24e57a012bc43bc4832"
	default:
		want = "aea8b80a5203e5ebc8e9fc0a771ff8b3eeb6fe25e89712d52aa9ae56db358b83"
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(data)); got != want {
		t.Fatalf("MP3 PCM, synthesis coefficients and history hash = %s, want %s", got, want)
	}
}

func TestMP3SynthesisPairRounding(t *testing.T) {
	var line [mp3QMFSaved]float32
	// Coefficients from the stock dialog at the rounding boundary.
	coefficients := [...]uint32{0x3d3ff30a, 0x3da9dfe0, 0xbd81d0ac, 0xbbd1f9f8, 0xbe137842, 0xbd16b6d9, 0xbdd9e36e, 0x3d31feec}
	for i, coefficient := range coefficients {
		line[(14-2*i)*mp3QMFStride+2] = math.Float32frombits(coefficient)
	}

	var pcm [17]int16
	mp3SynthesisPair(pcm[:], 1, line[:])
	want := int16(531)
	if runtime.GOARCH == "arm64" && !mp3LinuxARM64Contract {
		want = 530
	}
	if pcm[16] != want {
		t.Fatalf("MP3 PCM midpoint = %d, want %d", pcm[16], want)
	}
}
