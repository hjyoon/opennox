//go:build !server

package ail

// Layer III stereo and subband processing follow minimp3 by lieff, CC0-1.0.

// Store both coefficient channels followed by the scalefactors. The scalar
// 8 kHz mixed-block reorder spans 24 values into the following region.
type mp3SpectralWorkspace [2*mp3SpectralSamples + 40]float32

func (w *mp3SpectralWorkspace) channel(ch int) *[mp3SpectralSamples]float32 {
	return (*[mp3SpectralSamples]float32)(w[ch*mp3SpectralSamples:])
}

func (g *mp3Granule) longSubbands(header mp3FrameHeader) int {
	if g.mixedBlock == 0 {
		return 0
	}
	if header.bandRateIndex() == 2 {
		return 4
	}
	return 2
}

func (w *mp3SpectralWorkspace) unpack(header mp3FrameHeader, granules []mp3Granule, bits *mp3BitReader, intensity *[2][39]uint8) {
	channels := header.channelCount()
	scales := (*[40]float32)(w[2*mp3SpectralSamples:])
	first, second := w.channel(0), w.channel(1)
	clear(w[:2*mp3SpectralSamples])

	for ch := range channels {
		g := &granules[ch]
		end := bits.pos + int(g.dataBits)
		readMP3Scalefactors(header, g, ch, bits, &intensity[ch], scales)
		decodeMP3Huffman(w.channel(ch), g, scales, bits, end)
	}

	if header[3]&0x10 != 0 {
		mp3IntensityStereo(first, second, header, &granules[0], &intensity[1], granules[1].scalefacCompress&1)
	} else if header.midSideStereo() {
		mp3MidSide(first[:], second[:])
	}

	for ch := range channels {
		g := &granules[ch]
		boundaries := mp3Subbands - 1
		if g.shortBands != 0 {
			long := g.longSubbands(header)
			boundaries = long - 1
			mp3ReorderShortBands(w[ch*mp3SpectralSamples+long*mp3SamplesPerSubband:], g.bandWidths[g.longBands:])
		}
		mp3Antialias(w.channel(ch), boundaries)
	}
}

func mp3MidSide(left, right []float32) {
	for i, a := range left {
		b := right[i]
		left[i], right[i] = a+b, a-b
	}
}

var mp3IntensityPan = [7][2]float32{
	{0, 1}, {0.21132487, 0.78867513}, {0.36602540, 0.63397460}, {0.5, 0.5},
	{0.63397460, 0.36602540}, {0.78867513, 0.21132487}, {1, 0},
}

func mp3IntensityStereo(left, right *[mp3SpectralSamples]float32, header mp3FrameHeader, g *mp3Granule, intensity *[39]uint8, mpeg2Shift uint16) {
	top := [3]int{-1, -1, -1}
	bands := int(g.longBands) + int(g.shortBands)
	offset := 0
	for i, width := range g.bandWidths[:bands] {
		for _, sample := range right[offset : offset+int(width)] {
			if sample != 0 {
				top[i%3] = i
				break
			}
		}
		offset += int(width)
	}
	if g.longBands != 0 {
		maximum := max(top[0], top[1], top[2])
		top = [3]int{maximum, maximum, maximum}
	}
	mpeg1 := header.granuleCount() == 2
	windows := 1
	if g.shortBands != 0 {
		windows = 3
	}
	var defaultPosition uint8
	maxPosition := uint8(64)
	if mpeg1 {
		defaultPosition, maxPosition = 3, 7
	}
	for i := range windows {
		last := bands - windows + i
		previous := last - windows
		position := defaultPosition
		if top[i] < previous {
			position = intensity[previous]
		}
		intensity[last] = position
	}

	offset = 0
	midSide := header[3]&0x20 != 0
	for i, bandWidth := range g.bandWidths {
		width := int(bandWidth)
		if width == 0 {
			break
		}
		l, r := left[offset:offset+width], right[offset:offset+width]
		position := intensity[i]
		if i > top[i%3] && position < maxPosition {
			var leftGain, rightGain float32
			if mpeg1 {
				leftGain, rightGain = mp3IntensityPan[position][0], mp3IntensityPan[position][1]
			} else {
				leftGain, rightGain = 1, mp3ScaleByQuarterSteps(1, (int(position)+1)>>1<<mpeg2Shift)
				if position&1 != 0 {
					leftGain, rightGain = rightGain, leftGain
				}
			}
			if midSide {
				leftGain *= 1.41421356
				rightGain *= 1.41421356
			}
			for j, v := range l {
				l[j], r[j] = v*leftGain, v*rightGain
			}
		} else if midSide {
			mp3MidSide(l, r)
		}
		offset += width
	}
}

func mp3ReorderShortBands(samples []float32, bands []uint8) {
	var reordered [mp3SpectralSamples]float32
	offset := 0
	for band := 0; bands[band] != 0; band += 3 {
		width := int(bands[band])
		for i := range width {
			reordered[offset+3*i] = samples[offset+i]
			reordered[offset+3*i+1] = samples[offset+width+i]
			reordered[offset+3*i+2] = samples[offset+2*width+i]
		}
		offset += 3 * width
	}
	copy(samples, reordered[:offset])
}

var mp3AntialiasCos = [8]float32{
	0.85749293, 0.88174200, 0.94962865, 0.98331459, 0.99551782, 0.99916056, 0.99989920, 0.99999316,
}
var mp3AntialiasSin = [8]float32{
	0.51449576, 0.47173197, 0.31337745, 0.18191320, 0.09457419, 0.04096558, 0.01419856, 0.00369997,
}

func mp3Antialias(spectrum *[mp3SpectralSamples]float32, boundaries int) {
	for band := 1; band <= boundaries; band++ {
		position := band * mp3SamplesPerSubband
		for i := range mp3AntialiasCos {
			u, d := spectrum[position+i], spectrum[position-1-i]
			cos, sin := mp3AntialiasCos[i], mp3AntialiasSin[i]
			spectrum[position+i] = mp3ProductDifference(u, cos, d, sin)
			spectrum[position-1-i] = u*sin + float32(d*cos)
		}
	}
}
