package noxrender

import (
	"bytes"
	"reflect"
	"testing"
)

// This reference widens before arithmetic and does not use ColorIntensity,
// SplitColor16, the table-index expression or any renderer blend routine.
// It is an independent table/formula contract, not an original-EXE oracle.
func colorIntensityReferenceR274(r, g, b byte) byte {
	return byte((76*(int(r)|7) + 150*(int(g)|7) + 28*(int(b)|7)) >> 8)
}

func TestColorIntensityAllRGB555CellsR274(t *testing.T) {
	r := new(NoxRender)
	r.initColorTablesRev()
	if len(r.colors.revTable) != 0x8002 {
		t.Fatalf("reverse table length = %d, want %d", len(r.colors.revTable), 0x8002)
	}
	before := append([]byte(nil), r.colors.revTable...)
	checked := 0
	for red := 0; red < 32; red++ {
		for green := 0; green < 32; green++ {
			for blue := 0; blue < 32; blue++ {
				r8, g8, b8 := byte(red<<3), byte(green<<3), byte(blue<<3)
				want := colorIntensityReferenceR274(r8, g8, b8)
				// Enumerating cells in RGB order also audits initialization of
				// every table entry independently of the production lookup.
				index := red*1024 + green*32 + blue
				if got := r.colors.revTable[index]; got != want {
					t.Fatalf("table[%d], RGB555=(%d,%d,%d): got %d, want %d", index, red, green, blue, got, want)
				}
				if got := r.ColorIntensity(r8, g8, b8); got != want {
					t.Fatalf("RGB555 cell (%d,%d,%d): got %d, want %d", red, green, blue, got, want)
				}
				checked++
			}
		}
	}
	if checked != 32768 {
		t.Fatalf("checked %d cells, want 32768", checked)
	}
	if !bytes.Equal(r.colors.revTable, before) {
		t.Fatal("ColorIntensity changed the reverse table")
	}
}

func TestColorIntensityAllLowBitAliasesAndImmutableStateR274(t *testing.T) {
	r := new(NoxRender)
	r.initColorTablesRev()
	r.p = &RenderData{
		useClip: 1, useAlpha: 3, multiply14: 5, flag16: 7, colorize17: 11,
		colorMult: RGB{R: 13, G: 17, B: 19}, lightColor: RGB{R: 23, G: 29, B: 31},
		alpha: 37, field_260: 41, field_261: 43, field_262: 47,
	}
	r.p.materials[3] = RenderMat{Color: RGB{R: 53, G: 59, B: 61}, Color32: 67}
	r.interlacing, r.interlacingY = true, 1
	r.dword_5d4594_3799476, r.dword_5d4594_3799484 = 71, 73
	beforeData, beforeRenderer := *r.p, *r
	beforeTable := append([]byte(nil), r.colors.revTable...)
	checked := 0
	for red := 0; red < 32; red++ {
		for green := 0; green < 32; green++ {
			for blue := 0; blue < 32; blue++ {
				r0, g0, b0 := byte(red<<3), byte(green<<3), byte(blue<<3)
				want := colorIntensityReferenceR274(r0, g0, b0)
				for rlow := byte(0); rlow < 8; rlow++ {
					for glow := byte(0); glow < 8; glow++ {
						for blow := byte(0); blow < 8; blow++ {
							r8, g8, b8 := r0|rlow, g0|glow, b0|blow
							if got := r.ColorIntensity(r8, g8, b8); got != want {
								t.Fatalf("RGB=(%d,%d,%d), cell=(%d,%d,%d): got %d, want %d", r8, g8, b8, red, green, blue, got, want)
							}
							checked++
						}
					}
				}
			}
		}
	}
	if checked != 256*256*256 {
		t.Fatalf("checked %d inputs, want %d", checked, 256*256*256)
	}
	if !bytes.Equal(r.colors.revTable, beforeTable) {
		t.Error("ColorIntensity changed the reverse table")
	}
	if !reflect.DeepEqual(*r.p, beforeData) {
		t.Error("ColorIntensity changed RenderData or its materials")
	}
	if !reflect.DeepEqual(*r, beforeRenderer) {
		t.Error("ColorIntensity changed renderer state")
	}
}

func TestColorIntensityLiteralTableAddressesR274(t *testing.T) {
	// Distinct sentinel bytes verify the actual table address independently of
	// both the weighted-color formula and initColorTablesRev.
	r := new(NoxRender)
	r.colors.revTable = make([]byte, 32768)
	cases := []struct {
		r, g, b byte
		index   int
		value   byte
	}{
		{0, 0, 0, 0x0000, 11},
		{255, 0, 0, 0x7c00, 23},
		{0, 255, 0, 0x03e0, 37},
		{0, 0, 255, 0x001f, 53},
		{255, 255, 255, 0x7fff, 71},
		{0xa8, 0x50, 0xd0, 0x555a, 97},
		{0xaf, 0x57, 0xd7, 0x555a, 97},
	}
	for _, tc := range cases {
		r.colors.revTable[tc.index] = tc.value
	}
	before := append([]byte(nil), r.colors.revTable...)
	for _, tc := range cases {
		if got := r.ColorIntensity(tc.r, tc.g, tc.b); got != tc.value {
			t.Errorf("RGB=(%d,%d,%d): table address want %#04x (value %d), got value %d", tc.r, tc.g, tc.b, tc.index, tc.value, got)
		}
	}
	if !bytes.Equal(r.colors.revTable, before) {
		t.Fatal("ColorIntensity changed the sentinel table")
	}
}
