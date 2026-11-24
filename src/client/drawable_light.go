package client

import (
	"math"

	"github.com/opennox/opennox/v1/client/noxrender"
)

func (s *Drawable) SetLightColor(r, g, b byte) {
	s.SetLightColorRGB(noxrender.RGB{R: int(r), G: int(g), B: int(b)})
}

func (s *Drawable) SetLightColorRGB(color noxrender.RGB) {
	if s == nil {
		return
	}
	s.LightFlags = 2
	s.LightColor = color
}

func (s *Drawable) SetLightDirection(degrees int32) int64 {
	value := drawableLightAngle(degrees)
	if s != nil {
		s.LightDir = uint16(value)
		s.Field_42 = 0
	}
	return value
}

func (s *Drawable) SetLightPenumbra(degrees int32) int64 {
	value := drawableLightAngle(degrees)
	if s != nil {
		s.LightPenumbra = uint16(value)
	}
	return value
}

func drawableLightAngle(degrees int32) int64 {
	// The PE32 binary32 multiplier times 65536 is exactly 0xb60b61 / 65536.
	// The 0.5 bias and the integer boundaries are exact at x87 precision53.
	return (int64(degrees)*0xb60b61 + (1 << 15)) / (1 << 16)
}

func (s *Drawable) SetLightIntensity(v float32) {
	if s == nil {
		return
	}
	if v > 63 {
		v = 63
	}
	s.LightIntensity = v
	s.LightIntensityU16 = drawableFixedLightIntensity(v)
	s.LightIntensityRad = uint32(LightRadius(v))
}

func drawableFixedLightIntensity(v float32) uint32 {
	fixed := float64(v) * 65536
	biased := float64(fixed + 0.5)
	// The multiplication is exact. For large negative values, chop the
	// addition toward zero before converting to a signed QWORD.
	if biased < 0 && biased-fixed < 0.5 {
		biased = math.Nextafter(biased, 0)
	}
	if !(biased >= -0x1p63 && biased < 0x1p63) {
		return 0 // Low DWORD of the x87 integer-indefinite value.
	}
	return uint32(int64(biased))
}

func LightRadius(intens float32) int {
	const (
		minIntensity = 1.0
		bias         = 4.0
		divisor      = 961.0
		scale        = float64(float32(0.001))
	)
	if !(intens > minIntensity) {
		return 0
	}
	value := float64(intens)
	falloff := (bias+min(value, 31))/(bias+minIntensity) + 1
	radius := math.Sqrt(falloff * (value * value / (divisor * scale)))
	if !(radius < 0x1p63) {
		return 0
	}
	return int(int32(int64(radius)))
}
