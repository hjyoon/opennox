//go:build !server

package ail

import "runtime"

// The scalar Layer III inverse transform follows minimp3 by lieff, CC0-1.0.
const (
	mp3Subbands          = 32
	mp3SamplesPerSubband = 18
	mp3SpectralSamples   = mp3Subbands * mp3SamplesPerSubband
	mp3OverlapPerSubband = mp3SamplesPerSubband / 2
	mp3OverlapSamples    = mp3Subbands * mp3OverlapPerSubband
	mp3ShortBlock        = 2
	mp3StopBlock         = 3
)

// The Linux ARM64 GCC reference contracts multiply-adds across statements.
const mp3LinuxARM64Contract = runtime.GOOS == "linux" && runtime.GOARCH == "arm64"

var mp3LongWindow = [18]float32{
	0.99904822, 0.99144486, 0.97629601, 0.95371695, 0.92387953, 0.88701083, 0.84339145, 0.79335334, 0.73727734,
	0.04361938, 0.13052619, 0.21643961, 0.30070580, 0.38268343, 0.46174861, 0.53729961, 0.60876143, 0.67559021,
}

var mp3StopWindow = [18]float32{
	1, 1, 1, 1, 1, 1, 0.99144486, 0.92387953, 0.79335334,
	0, 0, 0, 0, 0, 0, 0.13052619, 0.38268343, 0.60876143,
}

var mp3LongTwiddle = [18]float32{
	0.73727734, 0.79335334, 0.84339145, 0.88701083, 0.92387953, 0.95371695, 0.97629601, 0.99144486, 0.99904822,
	0.67559021, 0.60876143, 0.53729961, 0.46174861, 0.38268343, 0.30070580, 0.21643961, 0.13052619, 0.04361938,
}

var mp3ShortTwiddle = [6]float32{0.79335334, 0.92387953, 0.99144486, 0.60876143, 0.38268343, 0.13052619}

func mp3Hybrid(samples *[mp3SpectralSamples]float32, overlap *[mp3OverlapSamples]float32, blockType uint8, longSubbands int) {
	for band := range mp3Subbands {
		block := (*[18]float32)(samples[band*18:])
		history := (*[9]float32)(overlap[band*9:])
		switch {
		case band < longSubbands:
			mp3InverseLong(block, history, &mp3LongWindow)
		case blockType == mp3ShortBlock:
			mp3InverseShort(block, history)
		case blockType == mp3StopBlock:
			mp3InverseLong(block, history, &mp3StopWindow)
		default:
			mp3InverseLong(block, history, &mp3LongWindow)
		}
		// Frequency inversion compensates for the alternating odd subbands.
		if band&1 != 0 {
			for i := 1; i < len(block); i += 2 {
				block[i] = -block[i]
			}
		}
	}
}

func mp3InverseLong(block *[18]float32, overlap *[9]float32, window *[18]float32) {
	var co, si [9]float32
	co[0], si[0] = -block[0], block[17]
	for i := range 4 {
		si[8-2*i] = block[4*i+1] - block[4*i+2]
		co[1+2*i] = block[4*i+1] + block[4*i+2]
		si[7-2*i] = block[4*i+4] - block[4*i+3]
		co[2+2*i] = -(block[4*i+3] + block[4*i+4])
	}
	mp3InverseDCT9(&co)
	mp3InverseDCT9(&si)
	for i := 1; i < len(si); i += 2 {
		si[i] = -si[i]
	}
	for i := range co {
		saved := overlap[i]
		twiddleCos, twiddleSin := mp3LongTwiddle[i], mp3LongTwiddle[9+i]
		// The scalar C expression rounds the right-hand product before the add.
		sum := co[i]*twiddleSin + float32(si[i]*twiddleCos)
		overlap[i] = co[i]*twiddleCos - float32(si[i]*twiddleSin)
		block[i] = saved*window[i] - float32(sum*window[9+i])
		block[17-i] = saved*window[9+i] + float32(sum*window[i])
	}
}

func mp3InverseDCT9(y *[9]float32) {
	s0, s2, s4, s6, s8 := y[0], y[2], y[4], y[6], y[8]
	t0 := s0 + s6*0.5
	s0 -= s6
	sum24, sum28, diff48 := s4+s2, s8+s2, s4-s8
	s4 += s8 - s2

	s2 = s0 - s4*0.5
	y[4] = s4 + s0
	if mp3LinuxARM64Contract {
		s8 = (t0 - sum28*0.76604444) + diff48*0.17364818
		s0 = (t0 - sum24*0.93969262) + sum28*0.76604444
		s4 = (t0 + sum24*0.93969262) - diff48*0.17364818
	} else {
		t4 := float32(sum24 * 0.93969262)
		t2 := float32(sum28 * 0.76604444)
		t6 := float32(diff48 * 0.17364818)
		s8 = t0 - t2 + t6
		s0 = t0 - t4 + t2
		s4 = t0 + t4 - t6
	}

	s1, s3, s5, s7 := y[1], y[3], y[5], y[7]
	sum17 := s1 + s7
	odd := s1 - s5 - s7
	t0 = float32((s5 + s1) * 0.98480775)
	t4 := float32((s5 - s7) * 0.34202014)
	if mp3LinuxARM64Contract {
		s5 = (t0 - s3*0.86602540) - sum17*0.64278761
		s7 = (t4 - s3*0.86602540) - t0
		s3 = (t4 + s3*0.86602540) - sum17*0.64278761
		y[1], y[7] = s2+odd*0.86602540, s2-odd*0.86602540
	} else {
		s3 = float32(s3 * 0.86602540)
		t2 := float32(sum17 * 0.64278761)
		s1 = float32(odd * 0.86602540)
		s5 = t0 - s3 - t2
		s7 = t4 - s3 - t0
		s3 = t4 + s3 - t2
		y[1], y[7] = s2+s1, s2-s1
	}

	y[0], y[2], y[3] = s4-s7, s0-s3, s8+s5
	y[5], y[6], y[8] = s8-s5, s0+s3, s4+s7
}

func mp3InverseShort(block *[18]float32, overlap *[9]float32) {
	coefficients := *block
	copy(block[:6], overlap[:6])
	mp3InverseWindow12(coefficients[:], block[6:12], overlap[6:])
	mp3InverseWindow12(coefficients[1:], block[12:], overlap[6:])
	mp3InverseWindow12(coefficients[2:], overlap[:6], overlap[6:])
}

func mp3InverseWindow12(coefficients, dst, overlap []float32) {
	co := mp3InverseDCT3(-coefficients[0], coefficients[6]+coefficients[3], coefficients[12]+coefficients[9])
	si := mp3InverseDCT3(coefficients[15], coefficients[12]-coefficients[9], coefficients[6]-coefficients[3])
	si[1] = -si[1]
	for i := range co {
		saved := overlap[i]
		sum := co[i]*mp3ShortTwiddle[3+i] + float32(si[i]*mp3ShortTwiddle[i])
		overlap[i] = co[i]*mp3ShortTwiddle[i] - float32(si[i]*mp3ShortTwiddle[3+i])
		dst[i] = saved*mp3ShortTwiddle[2-i] - float32(sum*mp3ShortTwiddle[5-i])
		dst[5-i] = saved*mp3ShortTwiddle[5-i] + float32(sum*mp3ShortTwiddle[2-i])
	}
}

func mp3InverseDCT3(x0, x1, x2 float32) [3]float32 {
	a1 := x0 - x2*0.5
	if mp3LinuxARM64Contract {
		return [3]float32{a1 + x1*0.86602540, x0 + x2, a1 - x1*0.86602540}
	}
	m1 := float32(x1 * 0.86602540)
	return [3]float32{a1 + m1, x0 + x2, a1 - m1}
}
