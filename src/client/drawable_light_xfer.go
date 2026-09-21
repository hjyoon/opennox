package client

import (
	"encoding/binary"
	"math"
)

// DrawableLightXferSize is the size of GAME.EXE's PE32 drawable fields
// 34..68. The map stream stays 140 bytes even when Go int and pointer fields
// in Drawable widen on a 64-bit host.
const DrawableLightXferSize = 140

// LightXferData returns the fixed-width drawable light block consumed by
// InvisibleLightXfer. RGB values are Go ints at runtime, but GAME.EXE stored
// them as signed 32-bit integers in this block.
func (s *Drawable) LightXferData() [DrawableLightXferSize]byte {
	var out [DrawableLightXferSize]byte
	if s == nil {
		return out
	}
	putU32 := func(off int, v uint32) {
		binary.LittleEndian.PutUint32(out[off:], v)
	}
	putU16 := func(off int, v uint16) {
		binary.LittleEndian.PutUint16(out[off:], v)
	}

	putU32(0, s.LightFlags)
	putU32(4, math.Float32bits(s.LightIntensity))
	putU32(8, s.LightIntensityRad)
	putU32(12, s.LightIntensityU16)
	putU32(16, uint32(int32(s.LightColor.R)))
	putU32(20, uint32(int32(s.LightColor.G)))
	putU32(24, uint32(int32(s.LightColor.B)))
	putU16(28, s.LightDir)
	putU16(30, s.LightPenumbra)
	putU32(32, s.Field_42)
	putU32(36, s.Field_43)
	putU32(40, s.Field_44)
	for i, v := range s.data_45 {
		putU32(44+4*i, v)
	}
	for i, v := range s.data_50 {
		putU32(64+4*i, v)
	}
	for i, v := range s.data_60 {
		putU32(104+4*i, v)
	}
	putU32(124, s.Field_65)
	putU32(128, s.Field_66)
	putU32(132, s.Field_67)
	putU32(136, s.Field_68)
	return out
}

// ApplyLightXferData restores GAME.EXE's fixed-width PE32 light block into a
// native-width Drawable. In particular, RGB components must be sign-extended
// into Go ints instead of copying the packed bytes over the widened fields.
func (s *Drawable) ApplyLightXferData(in *[DrawableLightXferSize]byte) {
	if s == nil || in == nil {
		return
	}
	u32 := func(off int) uint32 {
		return binary.LittleEndian.Uint32(in[off:])
	}
	u16 := func(off int) uint16 {
		return binary.LittleEndian.Uint16(in[off:])
	}

	s.LightFlags = u32(0)
	s.LightIntensity = math.Float32frombits(u32(4))
	s.LightIntensityRad = u32(8)
	s.LightIntensityU16 = u32(12)
	s.LightColor.R = int(int32(u32(16)))
	s.LightColor.G = int(int32(u32(20)))
	s.LightColor.B = int(int32(u32(24)))
	s.LightDir = u16(28)
	s.LightPenumbra = u16(30)
	s.Field_42 = u32(32)
	s.Field_43 = u32(36)
	s.Field_44 = u32(40)
	for i := range s.data_45 {
		s.data_45[i] = u32(44 + 4*i)
	}
	for i := range s.data_50 {
		s.data_50[i] = u32(64 + 4*i)
	}
	for i := range s.data_60 {
		s.data_60[i] = u32(104 + 4*i)
	}
	s.Field_65 = u32(124)
	s.Field_66 = u32(128)
	s.Field_67 = u32(132)
	s.Field_68 = u32(136)
}
