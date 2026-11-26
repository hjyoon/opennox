//go:build !server

package ail

// Layer III side information follows minimp3 by lieff, CC0-1.0.
type mp3BitReader struct {
	data []byte
	pos  int
}

func (b *mp3BitReader) read(bits int) uint32 {
	start := b.pos
	b.pos += bits
	if bits == 0 || b.pos > len(b.data)*8 {
		return 0
	}
	index, skip := start/8, start%8
	value := uint32(b.data[index]) & (0xff >> skip)
	left := bits + skip - 8
	for ; left > 0; left -= 8 {
		index++
		value = value<<8 | uint32(b.data[index])
	}
	return value >> -left
}

type mp3Granule struct {
	bandWidths        []uint8
	dataBits          uint16
	bigValues         uint16
	scalefacCompress  uint16
	globalGain        uint8
	blockType         uint8
	mixedBlock        uint8
	longBands         uint8
	shortBands        uint8
	huffmanTable      [3]uint8
	regionBandCount   [3]uint8
	subblockGain      [3]uint8
	preflag           uint8
	scalefacScale     uint8
	countOneTable     uint8
	scalefacReuseMask uint8
}

type mp3SideInfo struct {
	mainDataBegin  int
	mainDataOffset int
	granules       [4]mp3Granule
}

func (h mp3FrameHeader) bandRateIndex() int {
	version := int(h[1]>>3&1) + int(h[1]>>4&1)
	return 3*version + int(h[2]>>2&3)
}

func (h mp3FrameHeader) channelCount() int {
	if h[3]&0xc0 == 0xc0 {
		return 1
	}
	return 2
}

func (h mp3FrameHeader) granuleCount() int {
	if h[1]&8 != 0 {
		return 2
	}
	return 1
}

func readMP3SideInfo(frame []byte, header mp3FrameHeader) (mp3SideInfo, bool) {
	var info mp3SideInfo
	if len(frame) < mp3HeaderSize {
		return info, false
	}
	bits := mp3BitReader{data: frame[mp3HeaderSize:]}
	if header[1]&1 == 0 {
		bits.read(16) // CRC precedes side information when protection is enabled.
	}
	bandRate := max(0, header.bandRateIndex()-1)
	channels, granules := header.channelCount(), header.granuleCount()
	count := channels * granules
	mpeg1 := granules == 2
	scalefacBits := 9
	var reuse uint32
	if mpeg1 {
		scalefacBits = 4
		info.mainDataBegin = int(bits.read(9))
		reuse = bits.read(7 + count)
	} else {
		info.mainDataBegin = int(bits.read(8+channels)) >> channels
	}

	var dataBits int
	for i := 0; i < count; i++ {
		if channels == 1 {
			reuse <<= 4
		}
		g := &info.granules[i]
		g.dataBits = uint16(bits.read(12))
		dataBits += int(g.dataBits)
		g.bigValues = uint16(bits.read(9))
		if g.bigValues > 288 {
			return mp3SideInfo{}, false
		}
		g.globalGain = uint8(bits.read(8))
		g.scalefacCompress = uint16(bits.read(scalefacBits))
		g.bandWidths, g.longBands = mp3LongBandWidths[bandRate][:], 22

		var tables uint32
		if bits.read(1) != 0 {
			g.blockType = uint8(bits.read(2))
			if g.blockType == 0 {
				return mp3SideInfo{}, false
			}
			g.mixedBlock = uint8(bits.read(1))
			g.regionBandCount[0], g.regionBandCount[1] = 7, 255
			if g.blockType == 2 {
				reuse &= 0x0f0f
				if g.mixedBlock == 0 {
					g.regionBandCount[0] = 8
					g.bandWidths, g.longBands, g.shortBands = mp3ShortBandWidths[bandRate][:], 0, 39
				} else {
					g.bandWidths, g.longBands, g.shortBands = mp3MixedBandWidths[bandRate][:], 6, 30
					if mpeg1 {
						g.longBands = 8
					}
				}
			}
			tables = bits.read(10) << 5
			for s := range g.subblockGain {
				g.subblockGain[s] = uint8(bits.read(3))
			}
		} else {
			tables = bits.read(15)
			g.regionBandCount = [3]uint8{uint8(bits.read(4)), uint8(bits.read(3)), 255}
		}
		g.huffmanTable = [3]uint8{uint8(tables >> 10), uint8(tables >> 5 & 31), uint8(tables & 31)}
		if mpeg1 {
			g.preflag = uint8(bits.read(1))
		} else if g.scalefacCompress >= 500 {
			g.preflag = 1
		}
		g.scalefacScale = uint8(bits.read(1))
		g.countOneTable = uint8(bits.read(1))
		g.scalefacReuseMask = uint8(reuse >> 12 & 15)
		reuse <<= 4
	}

	limit := len(bits.data) * 8
	if bits.pos > limit || dataBits+bits.pos > limit+info.mainDataBegin*8 {
		return mp3SideInfo{}, false
	}
	info.mainDataOffset = mp3HeaderSize + bits.pos/8
	return info, true
}

var mp3LongBandWidths = [8][23]uint8{
	{6, 6, 6, 6, 6, 6, 8, 10, 12, 14, 16, 20, 24, 28, 32, 38, 46, 52, 60, 68, 58, 54, 0},
	{12, 12, 12, 12, 12, 12, 16, 20, 24, 28, 32, 40, 48, 56, 64, 76, 90, 2, 2, 2, 2, 2, 0},
	{6, 6, 6, 6, 6, 6, 8, 10, 12, 14, 16, 20, 24, 28, 32, 38, 46, 52, 60, 68, 58, 54, 0},
	{6, 6, 6, 6, 6, 6, 8, 10, 12, 14, 16, 18, 22, 26, 32, 38, 46, 54, 62, 70, 76, 36, 0},
	{6, 6, 6, 6, 6, 6, 8, 10, 12, 14, 16, 20, 24, 28, 32, 38, 46, 52, 60, 68, 58, 54, 0},
	{4, 4, 4, 4, 4, 4, 6, 6, 8, 8, 10, 12, 16, 20, 24, 28, 34, 42, 50, 54, 76, 158, 0},
	{4, 4, 4, 4, 4, 4, 6, 6, 6, 8, 10, 12, 16, 18, 22, 28, 34, 40, 46, 54, 54, 192, 0},
	{4, 4, 4, 4, 4, 4, 6, 6, 8, 10, 12, 16, 20, 24, 30, 38, 46, 56, 68, 84, 102, 26, 0},
}

var mp3ShortBandWidths = [8][40]uint8{
	{4, 4, 4, 4, 4, 4, 4, 4, 4, 6, 6, 6, 8, 8, 8, 10, 10, 10, 12, 12, 12, 14, 14, 14, 18, 18, 18, 24, 24, 24, 30, 30, 30, 40, 40, 40, 18, 18, 18, 0},
	{8, 8, 8, 8, 8, 8, 8, 8, 8, 12, 12, 12, 16, 16, 16, 20, 20, 20, 24, 24, 24, 28, 28, 28, 36, 36, 36, 2, 2, 2, 2, 2, 2, 2, 2, 2, 26, 26, 26, 0},
	{4, 4, 4, 4, 4, 4, 4, 4, 4, 6, 6, 6, 6, 6, 6, 8, 8, 8, 10, 10, 10, 14, 14, 14, 18, 18, 18, 26, 26, 26, 32, 32, 32, 42, 42, 42, 18, 18, 18, 0},
	{4, 4, 4, 4, 4, 4, 4, 4, 4, 6, 6, 6, 8, 8, 8, 10, 10, 10, 12, 12, 12, 14, 14, 14, 18, 18, 18, 24, 24, 24, 32, 32, 32, 44, 44, 44, 12, 12, 12, 0},
	{4, 4, 4, 4, 4, 4, 4, 4, 4, 6, 6, 6, 8, 8, 8, 10, 10, 10, 12, 12, 12, 14, 14, 14, 18, 18, 18, 24, 24, 24, 30, 30, 30, 40, 40, 40, 18, 18, 18, 0},
	{4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 6, 6, 6, 8, 8, 8, 10, 10, 10, 12, 12, 12, 14, 14, 14, 18, 18, 18, 22, 22, 22, 30, 30, 30, 56, 56, 56, 0},
	{4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 6, 6, 6, 6, 6, 6, 10, 10, 10, 12, 12, 12, 14, 14, 14, 16, 16, 16, 20, 20, 20, 26, 26, 26, 66, 66, 66, 0},
	{4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 6, 6, 6, 8, 8, 8, 12, 12, 12, 16, 16, 16, 20, 20, 20, 26, 26, 26, 34, 34, 34, 42, 42, 42, 12, 12, 12, 0},
}

var mp3MixedBandWidths = [8][40]uint8{
	{6, 6, 6, 6, 6, 6, 6, 6, 6, 8, 8, 8, 10, 10, 10, 12, 12, 12, 14, 14, 14, 18, 18, 18, 24, 24, 24, 30, 30, 30, 40, 40, 40, 18, 18, 18, 0},
	{12, 12, 12, 4, 4, 4, 8, 8, 8, 12, 12, 12, 16, 16, 16, 20, 20, 20, 24, 24, 24, 28, 28, 28, 36, 36, 36, 2, 2, 2, 2, 2, 2, 2, 2, 2, 26, 26, 26, 0},
	{6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 8, 8, 8, 10, 10, 10, 14, 14, 14, 18, 18, 18, 26, 26, 26, 32, 32, 32, 42, 42, 42, 18, 18, 18, 0},
	{6, 6, 6, 6, 6, 6, 6, 6, 6, 8, 8, 8, 10, 10, 10, 12, 12, 12, 14, 14, 14, 18, 18, 18, 24, 24, 24, 32, 32, 32, 44, 44, 44, 12, 12, 12, 0},
	{6, 6, 6, 6, 6, 6, 6, 6, 6, 8, 8, 8, 10, 10, 10, 12, 12, 12, 14, 14, 14, 18, 18, 18, 24, 24, 24, 30, 30, 30, 40, 40, 40, 18, 18, 18, 0},
	{4, 4, 4, 4, 4, 4, 6, 6, 4, 4, 4, 6, 6, 6, 8, 8, 8, 10, 10, 10, 12, 12, 12, 14, 14, 14, 18, 18, 18, 22, 22, 22, 30, 30, 30, 56, 56, 56, 0},
	{4, 4, 4, 4, 4, 4, 6, 6, 4, 4, 4, 6, 6, 6, 6, 6, 6, 10, 10, 10, 12, 12, 12, 14, 14, 14, 16, 16, 16, 20, 20, 20, 26, 26, 26, 66, 66, 66, 0},
	{4, 4, 4, 4, 4, 4, 6, 6, 4, 4, 4, 6, 6, 6, 8, 8, 8, 12, 12, 12, 16, 16, 16, 20, 20, 20, 26, 26, 26, 34, 34, 34, 42, 42, 42, 12, 12, 12, 0},
}
