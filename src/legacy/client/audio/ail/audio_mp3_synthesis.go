//go:build !server

package ail

// The scalar polyphase synthesis filter follows minimp3 by lieff, CC0-1.0.
const (
	mp3QMFTaps     = 15
	mp3QMFStride   = 2 * mp3Subbands
	mp3QMFSaved    = mp3QMFTaps * mp3QMFStride
	mp3QMFWorkSize = (mp3QMFTaps + mp3SamplesPerSubband) * mp3QMFStride
)

type mp3SynthesisState [mp3QMFSaved]float32

func (s *mp3SynthesisState) synthesize(spectrum *[2][mp3SpectralSamples]float32, channels int, pcm []int16) {
	for ch := range channels {
		mp3SynthesisDCT(&spectrum[ch])
	}

	var lines [mp3QMFWorkSize]float32
	copy(lines[:], s[:])
	for slot := 0; slot < mp3SamplesPerSubband; slot += 2 {
		mp3SynthesizeSlots(spectrum, slot, lines[slot*mp3QMFStride:], channels, pcm[slot*channels*mp3Subbands:])
	}

	history := lines[mp3SamplesPerSubband*mp3QMFStride:]
	if channels == 1 {
		for i := 0; i < len(s); i += 2 {
			s[i] = history[i]
		}
	} else {
		copy(s[:], history)
	}
}

var mp3SynthesisSecant = [8][3]float32{
	{10.19000816, 0.50060302, 0.50241929},
	{3.40760851, 0.50547093, 0.52249861},
	{2.05778098, 0.51544732, 0.56694406},
	{1.48416460, 0.53104258, 0.64682180},
	{1.16943991, 0.55310392, 0.78815460},
	{0.97256821, 0.58293498, 1.06067765},
	{0.83934963, 0.62250412, 1.72244716},
	{0.74453628, 0.67480832, 5.10114861},
}

func mp3SynthesisDCT(spectrum *[mp3SpectralSamples]float32) {
	for slot := range mp3SamplesPerSubband {
		var work [4][8]float32
		for i, sec := range &mp3SynthesisSecant {
			x0 := spectrum[slot+i*mp3SamplesPerSubband]
			x1 := spectrum[slot+(15-i)*mp3SamplesPerSubband]
			x2 := spectrum[slot+(16+i)*mp3SamplesPerSubband]
			x3 := spectrum[slot+(31-i)*mp3SamplesPerSubband]
			sum03, sum12 := x0+x3, x1+x2
			diff03 := float32((x0 - x3) * sec[1])
			work[0][i] = sum03 + sum12
			work[1][i] = (sum03 - sum12) * sec[2]
			if mp3LinuxARM64Contract {
				work[2][i] = diff03 + (x1-x2)*sec[0]
				work[3][i] = (diff03 - (x1-x2)*sec[0]) * sec[2]
			} else {
				diff12 := float32((x1 - x2) * sec[0])
				work[2][i] = diff03 + diff12
				work[3][i] = (diff03 - diff12) * sec[2]
			}
		}
		for i := range work {
			mp3SynthesisDCT8(&work[i])
		}
		for i := range 7 {
			j := slot + 4*i*mp3SamplesPerSubband
			spectrum[j] = work[0][i]
			spectrum[j+mp3SamplesPerSubband] = work[2][i] + work[3][i] + work[3][i+1]
			spectrum[j+2*mp3SamplesPerSubband] = work[1][i] + work[1][i+1]
			spectrum[j+3*mp3SamplesPerSubband] = work[2][i+1] + work[3][i] + work[3][i+1]
		}
		j := slot + 28*mp3SamplesPerSubband
		spectrum[j] = work[0][7]
		spectrum[j+mp3SamplesPerSubband] = work[2][7] + work[3][7]
		spectrum[j+2*mp3SamplesPerSubband] = work[1][7]
		spectrum[j+3*mp3SamplesPerSubband] = work[3][7]
	}
}

func mp3SynthesisDCT8(x *[8]float32) {
	x0, x1, x2, x3 := x[0], x[1], x[2], x[3]
	x4, x5, x6, x7 := x[4], x[5], x[6], x[7]
	xt := x0 - x7
	x0 += x7
	x7 = x1 - x6
	x1 += x6
	x6 = x2 - x5
	x2 += x5
	x5 = x3 - x4
	x3 += x4
	x4 = x0 - x3
	x0 += x3
	x3 = x1 - x2
	x1 += x2
	x[0] = x0 + x1
	x[4] = (x0 - x1) * 0.70710677
	x5 += x6
	if mp3LinuxARM64Contract {
		x6 = (x6 + x7) * 0.70710677
		x3 = (x3 + x4) * 0.70710677
	} else {
		x6 = float32((x6 + x7) * 0.70710677)
		x3 = float32((x3 + x4) * 0.70710677)
	}
	x7 += xt
	x5 -= x7 * 0.198912367
	x7 += x5 * 0.382683432
	x5 -= x7 * 0.198912367
	x0 = xt - x6
	xt += x6
	x[1] = (xt + x7) * 0.50979561
	x[2] = (x4 + x3) * 0.54119611
	x[3] = (x0 - x5) * 0.60134488
	x[5] = (x0 + x5) * 0.89997619
	x[6] = (x4 - x3) * 1.30656302
	x[7] = (xt - x7) * 2.56291556
}

func mp3SynthesizeSlots(spectrum *[2][mp3SpectralSamples]float32, slot int, lines []float32, channels int, pcm []int16) {
	left, right := &spectrum[0], &spectrum[channels-1]
	for j := range 2 {
		z := mp3QMFSaved + 4*(15+16*j)
		lines[z] = left[slot+j+mp3SamplesPerSubband*16]
		lines[z+1] = right[slot+j+mp3SamplesPerSubband*16]
		lines[z+2] = left[slot+j]
		lines[z+3] = right[slot+j]
		for ch := range channels {
			z := 4*15 + mp3QMFStride*j + ch
			mp3SynthesisPair(pcm[j*mp3Subbands*channels+ch:], channels, lines[z:])
		}
	}

	for row := range mp3SynthesisWindow {
		weights := &mp3SynthesisWindow[row]
		i := 14 - row
		z := mp3QMFSaved + 4*i
		mirror := slot + mp3SamplesPerSubband*(31-i)
		next := slot + mp3SamplesPerSubband*(1+i)
		lines[z], lines[z+1] = left[mirror], right[mirror]
		lines[z+2], lines[z+3] = left[mirror+1], right[mirror+1]
		lines[z+mp3QMFStride], lines[z+mp3QMFStride+1] = left[next+1], right[next+1]
		lines[z-mp3QMFStride+2], lines[z-mp3QMFStride+3] = left[next], right[next]
		var a, b [4]float32
		for j := range 4 {
			vz, vy := lines[z+j], lines[z+j-mp3QMFStride*15]
			a[j] = mp3SynthesisDifference(vz, weights[0], vy, weights[1])
			b[j] = vz*weights[1] + float32(vy*weights[0])
		}
		for tap := 1; tap < 8; tap++ {
			w0, w1 := weights[2*tap], weights[2*tap+1]
			front := z - mp3QMFStride*tap
			back := z - mp3QMFStride*(15-tap)
			for j := range 4 {
				vz, vy := lines[front+j], lines[back+j]
				b[j] += vz*w1 + float32(vy*w0)
				if tap&1 != 0 {
					a[j] += mp3SynthesisDifference(vy, w1, vz, w0)
				} else {
					a[j] += mp3SynthesisDifference(vz, w0, vy, w1)
				}
			}
		}
		for ch := range channels {
			pcm[(15-i)*channels+ch] = mp3PCM16(a[ch])
			pcm[(17+i)*channels+ch] = mp3PCM16(b[ch])
			pcm[(47-i)*channels+ch] = mp3PCM16(a[2+ch])
			pcm[(49+i)*channels+ch] = mp3PCM16(b[2+ch])
		}
	}
}

func mp3SynthesisDifference(a, wa, b, wb float32) float32 {
	if mp3LinuxARM64Contract {
		return float32(a*wa) - b*wb
	}
	return a*wa - float32(b*wb)
}

func mp3SynthesisPair(pcm []int16, channels int, z []float32) {
	a := (z[14*64] - z[0]) * 29
	a += (z[1*64] + z[13*64]) * 213
	a += (z[12*64] - z[2*64]) * 459
	a += (z[3*64] + z[11*64]) * 2037
	a += (z[10*64] - z[4*64]) * 5153
	a += (z[5*64] + z[9*64]) * 6574
	a += (z[8*64] - z[6*64]) * 37489
	a += z[7*64] * 75038
	pcm[0] = mp3PCM16(a)

	z = z[2:]
	a = z[14*64] * 104
	a += z[12*64] * 1567
	a += z[10*64] * 9727
	a += z[8*64] * 64019
	a += z[6*64] * -9975
	a += z[4*64] * -45
	a += z[2*64] * 146
	a += z[0] * -5
	pcm[16*channels] = mp3PCM16(a)
}

func mp3PCM16(sample float32) int16 {
	if sample >= 32766.5 {
		return 32767
	}
	if sample <= -32767.5 {
		return -32768
	}
	n := int16(sample + 0.5)
	if n < 0 {
		n--
	}
	return n
}

var mp3SynthesisWindow = [15][16]float32{
	{-1, 26, -31, 208, 218, 401, -519, 2063, 2000, 4788, -5517, 7134, 5959, 35640, -39336, 74992},
	{-1, 24, -35, 202, 222, 347, -581, 2080, 1952, 4425, -5879, 7640, 5288, 33791, -41176, 74856},
	{-1, 21, -38, 196, 225, 294, -645, 2087, 1893, 4063, -6237, 8092, 4561, 31947, -43006, 74630},
	{-1, 19, -41, 190, 227, 244, -711, 2085, 1822, 3705, -6589, 8492, 3776, 30112, -44821, 74313},
	{-1, 17, -45, 183, 228, 197, -779, 2075, 1739, 3351, -6935, 8840, 2935, 28289, -46617, 73908},
	{-1, 16, -49, 176, 228, 153, -848, 2057, 1644, 3004, -7271, 9139, 2037, 26482, -48390, 73415},
	{-2, 14, -53, 169, 227, 111, -919, 2032, 1535, 2663, -7597, 9389, 1082, 24694, -50137, 72835},
	{-2, 13, -58, 161, 224, 72, -991, 2001, 1414, 2330, -7910, 9592, 70, 22929, -51853, 72169},
	{-2, 11, -63, 154, 221, 36, -1064, 1962, 1280, 2006, -8209, 9750, -998, 21189, -53534, 71420},
	{-2, 10, -68, 147, 215, 2, -1137, 1919, 1131, 1692, -8491, 9863, -2122, 19478, -55178, 70590},
	{-3, 9, -73, 139, 208, -29, -1210, 1870, 970, 1388, -8755, 9935, -3300, 17799, -56778, 69679},
	{-3, 8, -79, 132, 200, -57, -1283, 1817, 794, 1095, -8998, 9966, -4533, 16155, -58333, 68692},
	{-4, 7, -85, 125, 189, -83, -1356, 1759, 605, 814, -9219, 9959, -5818, 14548, -59838, 67629},
	{-4, 7, -91, 117, 177, -106, -1428, 1698, 402, 545, -9416, 9916, -7154, 12980, -61289, 66494},
	{-5, 6, -97, 111, 163, -127, -1498, 1634, 185, 288, -9585, 9838, -8540, 11455, -62684, 65290},
}
