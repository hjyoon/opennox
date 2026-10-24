package opennox

import "math"

// The original light setters use a binary32 degrees-to-turns coefficient,
// then separate 53-bit, round-toward-zero operations and signed-qword FISTP.
func colorLightAngle4CE960(degrees int) int64 {
	value := fireballMulChop53_52C790(float64(int32(degrees)), float64(math.Float32frombits(0x3b360b61)))
	value = fireballMulChop53_52C790(value, 65536)
	return colorLightTruncate4CE960(fireballAddChop53_52C790(value, 0.5))
}

func colorLightTruncate4CE960(value float64) int64 {
	// x87 masked invalid FISTP produces the signed-qword indefinite value.
	// Callers which reload EAX subsequently interpret only its signed low DWORD.
	if math.IsNaN(value) || value >= 0x1p63 || value < -0x1p63 {
		return math.MinInt64
	}
	return int64(value)
}

func colorLightDivide4CE960(a, b float64) float64 {
	if b < 0 {
		return -colorLightDivideChop53_4CE440(a, -b)
	}
	return colorLightDivideChop53_4CE440(a, b)
}

func colorLightTrackedAngle4CE8C0(dx, dy int32) int32 {
	// IMUL and ADD precede FILD in the PE32 path, so squared distance wraps
	// as a signed DWORD. Coincident/invalid distances still reach the setter.
	square := int32(uint32(dx)*uint32(dx) + uint32(dy)*uint32(dy))
	distance := aiDependencySqrtChop53_546B63(float64(square))
	ratio := colorLightDivide4CE960(float64(dx), distance)
	// The CRT acos temporarily selects nearest rounding. Its finite path is
	// atan2(sqrt((1+x)*(1-x)),x); 419AF0 then spills the result to binary32
	// after the caller's chopped control word has been restored.
	angle := math.NaN()
	if ratio >= -1 && ratio <= 1 {
		plus, minus := float64(1+ratio), float64(1-ratio)
		product := float64(plus * minus)
		angle = math.Atan2(math.Sqrt(product), ratio)
	}
	radians := fireballSpill32_52C790(angle)
	degrees := fireballMulChop53_52C790(float64(radians), float64(math.Float32frombits(0x42652ee0)))
	if dy < 0 {
		degrees = fireballAddChop53_52C790(360, -degrees)
	}
	return int32(colorLightTruncate4CE960(degrees))
}
