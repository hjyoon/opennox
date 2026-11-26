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

func TestMP3HybridScalarReference(t *testing.T) {
	// Cycle through four block types and three mixed-block boundaries. The
	// spectrum and the retained overlap both contribute to the reference hash.
	const granules = 12
	data := make([]byte, 0, granules*(mp3SpectralSamples+mp3OverlapSamples)*4)
	var spectrum [mp3SpectralSamples]float32
	var overlap [mp3OverlapSamples]float32
	var seed uint32 = 68934692
	sample := func() float32 {
		seed = seed*1664525 + 1013904223
		return float32(int32(seed>>14)-131072) / 1024
	}
	for i := range granules {
		for j := range spectrum {
			spectrum[j] = sample()
		}
		if i%11 == 0 {
			for j := range overlap {
				overlap[j] = sample()
			}
		}
		mp3Hybrid(&spectrum, &overlap, uint8(i%4), 2*(i%3))
		for _, v := range spectrum {
			data = binary.LittleEndian.AppendUint32(data, math.Float32bits(v))
		}
		for _, v := range overlap {
			data = binary.LittleEndian.AppendUint32(data, math.Float32bits(v))
		}
	}

	var want string
	switch {
	case mp3LinuxARM64Contract:
		want = "0dc6108d2b5bd28a47fbbee116d57231f33094626c8ca0cc72d6a865e6f0c9c2"
	case runtime.GOARCH == "arm64":
		want = "0c096a2ab0964a7c3c4c4218417e061c46f1e5f3fc9c228d63f4884002dc6b01"
	default:
		want = "6607e702ed4f95f5c8d0f1fe82835e2a258cdb9bd74559d77667a6d9b1dcb0c6"
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(data)); got != want {
		t.Fatalf("MP3 spectrum and overlap hash = %s, want %s", got, want)
	}
}
