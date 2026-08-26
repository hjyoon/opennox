package client

import (
	"encoding/binary"
	"math"
	"testing"

	"github.com/opennox/opennox/v1/client/noxrender"
)

func TestDrawableLightXferDataPE32Layout(t *testing.T) {
	dr := &Drawable{
		LightFlags:        0x01020304,
		LightIntensity:    1.25,
		LightIntensityRad: 0x11121314,
		LightIntensityU16: 0x21222324,
		LightColor: noxrender.RGB{
			R: 0x31323334,
			G: -2,
			B: 0x41424344,
		},
		LightDir:      0x5152,
		LightPenumbra: 0x5354,
		Field_42:      0x61626364,
		Field_43:      0x71727374,
		Field_44:      0x81828384,
		Field_65:      0x91929394,
		Field_66:      0xa1a2a3a4,
		Field_67:      0xb1b2b3b4,
		Field_68:      0xc1c2c3c4,
	}
	for i := range dr.data_45 {
		dr.data_45[i] = 0xd0000000 + uint32(i)
	}
	for i := range dr.data_50 {
		dr.data_50[i] = 0xe0000000 + uint32(i)
	}
	for i := range dr.data_60 {
		dr.data_60[i] = 0xf0000000 + uint32(i)
	}

	got := dr.LightXferData()
	u32 := func(off int) uint32 { return binary.LittleEndian.Uint32(got[off:]) }
	u16 := func(off int) uint16 { return binary.LittleEndian.Uint16(got[off:]) }
	checks32 := map[int]uint32{
		0:   dr.LightFlags,
		4:   math.Float32bits(dr.LightIntensity),
		8:   dr.LightIntensityRad,
		12:  dr.LightIntensityU16,
		16:  uint32(int32(dr.LightColor.R)),
		20:  uint32(int32(dr.LightColor.G)),
		24:  uint32(int32(dr.LightColor.B)),
		32:  dr.Field_42,
		36:  dr.Field_43,
		40:  dr.Field_44,
		124: dr.Field_65,
		128: dr.Field_66,
		132: dr.Field_67,
		136: dr.Field_68,
	}
	for off, want := range checks32 {
		if v := u32(off); v != want {
			t.Errorf("u32[%d] = %#x, want %#x", off, v, want)
		}
	}
	if v := u16(28); v != dr.LightDir {
		t.Errorf("direction = %#x, want %#x", v, dr.LightDir)
	}
	if v := u16(30); v != dr.LightPenumbra {
		t.Errorf("penumbra = %#x, want %#x", v, dr.LightPenumbra)
	}
	for i, want := range dr.data_45 {
		if v := u32(44 + 4*i); v != want {
			t.Errorf("data45[%d] = %#x, want %#x", i, v, want)
		}
	}
	for i, want := range dr.data_50 {
		if v := u32(64 + 4*i); v != want {
			t.Errorf("data50[%d] = %#x, want %#x", i, v, want)
		}
	}
	for i, want := range dr.data_60 {
		if v := u32(104 + 4*i); v != want {
			t.Errorf("data60[%d] = %#x, want %#x", i, v, want)
		}
	}
	if len(got) != 140 {
		t.Fatalf("light block size = %d, want 140", len(got))
	}
}
