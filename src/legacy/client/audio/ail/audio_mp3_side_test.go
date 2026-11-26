//go:build !server

package ail

import "testing"

func TestMP3SideInfoCRCAndMixedBlock(t *testing.T) {
	header := mp3FrameHeader{0xff, 0xf2, 0x84, 0xc0}
	frame := append(header[:], 0x55, 0xaa, 0x05, 0x81, 0x00, 0x33, 0x4b, 0xf1, 0xa5, 0x89, 0x9f)
	info, ok := readMP3SideInfo(frame, header)
	if !ok || info.mainDataBegin != 5 || info.mainDataOffset != 15 {
		t.Fatalf("mixed frame side information = %+v, valid = %t", info, ok)
	}
	g := info.granules[0]
	if g.dataBits != 32 || g.bigValues != 12 || g.globalGain != 210 || g.scalefacCompress != 504 ||
		g.blockType != 2 || g.mixedBlock != 1 || g.longBands != 6 || g.shortBands != 30 ||
		g.preflag != 1 || g.scalefacScale != 1 || g.countOneTable != 1 ||
		g.huffmanTable != [3]uint8{5, 17, 0} || g.regionBandCount != [3]uint8{7, 255, 0} ||
		g.subblockGain != [3]uint8{1, 4, 7} {
		t.Fatalf("mixed frame granule = %+v", g)
	}
	if _, ok := readMP3SideInfo(frame[:len(frame)-1], header); ok {
		t.Fatal("accepted truncated side information")
	}
	frame[6] = 3 // The reservoir can supply 24 bits, but the granule requests 32.
	if _, ok := readMP3SideInfo(frame, header); ok {
		t.Fatal("accepted granule data beyond the available reservoir")
	}
}
