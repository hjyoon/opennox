//go:build !server

package ail

// Layer III scalefactor decoding follows minimp3 by lieff, CC0-1.0.
var mp3ScaleFactorPartitions = [3][28]uint8{
	{6, 5, 5, 5, 6, 5, 5, 5, 6, 5, 7, 3, 11, 10, 0, 0, 7, 7, 7, 0, 6, 6, 6, 3, 8, 8, 5, 0},
	{8, 9, 6, 12, 6, 9, 9, 9, 6, 9, 12, 6, 15, 18, 0, 0, 6, 15, 12, 0, 6, 12, 9, 6, 6, 18, 9, 0},
	{9, 9, 6, 12, 9, 9, 9, 9, 9, 9, 12, 6, 18, 18, 0, 0, 12, 12, 12, 0, 12, 9, 9, 6, 15, 12, 9, 0},
}

var mp3MPEG1ScalefactorWidths = [16]uint8{0, 1, 2, 3, 12, 5, 6, 7, 9, 10, 11, 13, 14, 15, 18, 19}
var mp3MPEG2ScalefactorBases = [6][4]int{
	{5, 5, 4, 4}, {5, 5, 4, 1}, {4, 3, 1, 1},
	{5, 6, 6, 1}, {4, 4, 4, 1}, {4, 3, 1, 1},
}
var mp3ScalefactorPreemphasis = [10]uint8{1, 1, 1, 1, 2, 2, 3, 3, 3, 2}
var mp3QuarterGains = [4]float32{9.31322575e-10, 7.83145814e-10, 6.58544508e-10, 5.53767716e-10}

func (h mp3FrameHeader) midSideStereo() bool {
	return h[3]&0xe0 == 0x60
}

func mp3ScaleByQuarterSteps(value float32, steps int) float32 {
	for {
		chunk := min(120, steps)
		value *= mp3QuarterGains[chunk&3] * float32(int32(1<<30)>>(chunk>>2))
		steps -= chunk
		if steps == 0 {
			return value
		}
	}
}

func readMP3Scalefactors(header mp3FrameHeader, g *mp3Granule, ch int, bits *mp3BitReader, intensity *[39]uint8, result *[40]float32) {
	partition := 0
	if g.shortBands != 0 {
		partition = 1
		if g.longBands == 0 {
			partition = 2
		}
	}
	counts := mp3ScaleFactorPartitions[partition][:]
	var widths [4]uint8
	reuse := int(g.scalefacReuseMask)
	mpeg1 := header.granuleCount() == 2
	if mpeg1 {
		encoded := mp3MPEG1ScalefactorWidths[g.scalefacCompress]
		widths = [4]uint8{encoded >> 2, encoded >> 2, encoded & 3, encoded & 3}
	} else {
		group := 0
		compress := int(g.scalefacCompress)
		if ch == 1 && header[3]&0x10 != 0 {
			group = 3
			compress >>= 1
		}
		for {
			modulus := 1
			for i := len(widths) - 1; i >= 0; i-- {
				base := mp3MPEG2ScalefactorBases[group][i]
				widths[i] = uint8(compress / modulus % base)
				modulus *= base
			}
			group++
			compress -= modulus
			if compress < 0 {
				break
			}
		}
		counts = counts[4*group:]
		reuse = 0
	}

	var raw [40]uint8
	offset := 0
	for p, width := range widths {
		length := int(counts[p])
		if length == 0 {
			break
		}
		factors := raw[offset : offset+length]
		positions := intensity[offset : offset+length]
		switch {
		case reuse&8 != 0:
			copy(factors, positions)
		case width == 0:
			clear(positions)
		default:
			maxFactor := uint8((1 << width) - 1)
			for i := range factors {
				factor := uint8(bits.read(int(width)))
				factors[i], positions[i] = factor, factor
				if !mpeg1 && factor == maxFactor {
					positions[i] = 255
				}
			}
		}
		offset += length
		reuse <<= 1
	}

	shift := g.scalefacScale + 1
	longBands, shortBands := int(g.longBands), int(g.shortBands)
	if shortBands != 0 {
		for i := longBands; i < longBands+shortBands; i += 3 {
			for w, gain := range g.subblockGain {
				raw[i+w] += gain << (3 - shift)
			}
		}
	} else if g.preflag != 0 {
		for i, factor := range mp3ScalefactorPreemphasis {
			raw[11+i] += factor
		}
	}

	const maximumGain = 44
	gainSteps := int(g.globalGain) - 214
	if header.midSideStereo() {
		gainSteps -= 2
	}
	gain := mp3ScaleByQuarterSteps(1<<(maximumGain/4), maximumGain-gainSteps)
	for i, factor := range raw[:longBands+shortBands] {
		result[i] = mp3ScaleByQuarterSteps(gain, int(factor)<<shift)
	}
}
