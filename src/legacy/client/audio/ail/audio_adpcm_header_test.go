//go:build !server

package ail

import (
	"slices"
	"testing"
)

func TestADPCMInitialStepIndexBounds(t *testing.T) {
	for _, tc := range []struct {
		name    string
		decode  func([]int16, []byte) []int16
		offsets []int
	}{
		{"mono", decodeADPCMMono, []int{2}},
		{"stereo-left", decodeADPCMStereo, []int{2}},
		{"stereo-right", decodeADPCMStereo, []int{6}},
		{"stereo-both", decodeADPCMStereo, []int{2, 6}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			block := make([]byte, 1024)
			block[8] = 0x71
			block[12] = 0x42
			for i, offset := range tc.offsets {
				block[offset] = []byte{147, 255}[i]
			}
			original := slices.Clone(block)
			valid := slices.Clone(block)
			for _, offset := range tc.offsets {
				valid[offset] = 88
			}
			got, want := tc.decode(nil, block), tc.decode(nil, valid)
			if !slices.Equal(got, want) {
				t.Fatal("the initial ADPCM index must be clamped before decoding its first nibble")
			}
			if !slices.Equal(block, original) {
				t.Fatal("the decoder changed its input")
			}
		})
	}
}
