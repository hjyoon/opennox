package opennox

import "math"

// The three original color-light interpolators run in the established
// gameplay precision53/chop context. Inputs are a uint32 frame, positive
// signed-byte count, uint16 delay and byte keys, so quotients and their FMA
// residuals remain finite and normal. Correct division locally, retaining
// the existing exact add/multiply/spill primitives without changing FENV.
//
//go:noinline
func colorLightDivideChop53_4CE440(a, b float64) float64 {
	quotient := float64(a / b)
	remainder := math.FMA(-quotient, b, a)
	if quotient > 0 && remainder < 0 || quotient < 0 && remainder > 0 {
		return math.Nextafter(quotient, 0)
	}
	return quotient
}

func colorLightFrameSelection4CE440(cycle float64, count byte, delay uint16, loop bool) (from, to int, subframe byte) {
	// The original CRT truncates to a qword, then FISUB reloads its signed
	// low DWORD. With an admitted count>=2, uint32 frames keep it in range.
	whole := int32(int64(cycle))
	phase := fireballAddChop53_52C790(cycle, -float64(whole))
	scaled := fireballMulChop53_52C790(phase, float64(int8(count)))
	index := int(int64(scaled))
	if loop {
		from, to = index, index+1
		if to >= int(count) {
			to = 0
		}
	} else if whole&1 != 0 {
		from = int(count) - index - 1
		to = from - 1
		if to < 0 {
			to = 0
		}
	} else {
		from, to = index, index+1
		if to >= int(count) {
			to = int(count) - 1
		}
	}
	remaining := fireballAddChop53_52C790(scaled, -float64(index))
	// Original AND EAX,255 deliberately wraps subframes for delays>255.
	subframe = byte(uint64(fireballMulChop53_52C790(remaining, float64(delay))))
	return from, to, subframe
}

func colorLightLinear4CE440(start, end int, subframe byte, delay uint16) float64 {
	// FDIV precedes FMUL: delta/delay*byteSubframe+start. Computing a
	// shared subframe/delay fraction changes both signed rounding and RGB.
	step := colorLightDivideChop53_4CE440(float64(end-start), float64(delay))
	return fireballAddChop53_52C790(fireballMulChop53_52C790(step, float64(subframe)), float64(start))
}
