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
