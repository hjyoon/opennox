package opennox

import (
	"fmt"
	"image"
	"testing"

	"github.com/opennox/libs/noximage"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/server"
)

func TestE2EEquipmentColorSchedule(t *testing.T) {
	for _, mode := range []int{0, 1} {
		var sc e2eScenario
		sc.CheckEquipmentColors(mode, "read-only equipment color check")
		if len(sc.steps) != 1 || sc.steps[0].ready == nil || sc.steps[0].fnc == nil || sc.steps[0].waitTimeout != 1200 {
			t.Fatal("equipment color check has no bounded live observation")
		}
	}
	for _, mode := range []int{-1, 2} {
		t.Run(fmt.Sprint(mode), func(t *testing.T) {
			var sc e2eScenario
			defer func() {
				if recover() == nil || len(sc.steps) != 0 {
					t.Fatal("unsupported equipment color mode scheduled")
				}
			}()
			sc.CheckEquipmentColors(mode, "invalid")
		})
	}
}

func TestE2EEquipmentPaletteReference(t *testing.T) {
	var initial [16]noxrender.Color16
	for slot := range initial {
		initial[slot] = noxrender.Color16{R: uint16(slot + 1), G: uint16(slot + 31), B: uint16(slot + 61)}
	}
	def := server.Modifier{
		Colors12: [8]types.RGB{
			{R: 99, G: 100, B: 101}, // slot zero must not replace material zero
			{R: 10, G: 20, B: 30}, {R: 11, G: 21, B: 31}, {R: 12, G: 22, B: 32},
			{R: 13, G: 23, B: 33}, {R: 14, G: 24, B: 34}, {R: 15, G: 25, B: 35},
			{R: 200, G: 201, B: 202}, // definition slot seven is not a body material
		},
		Effectiveness36: 1, Material40: 2, PriEnchant44: 3, SecEnchant48: 6,
	}
	base := initial
	base[1] = noxrender.Color16{R: 10, G: 20, B: 30}
	base[2] = noxrender.Color16{R: 11, G: 21, B: 31}
	base[3] = noxrender.Color16{R: 12, G: 22, B: 32}
	base[4] = noxrender.Color16{R: 13, G: 23, B: 33}
	base[5] = noxrender.Color16{R: 14, G: 24, B: 34}
	base[6] = noxrender.Color16{R: 15, G: 25, B: 35}
	mods := [4]*server.ModifierEff{
		{Color24: types.RGB{R: 210, G: 61, B: 97}}, {Color24: types.RGB{R: 189, G: 82, B: 116}},
		{Color24: types.RGB{R: 167, G: 103, B: 137}}, {Color24: types.RGB{R: 146, G: 124, B: 158}},
	}
	before := def
	for mask := uint(0); mask < 16; mask++ {
		var present [4]*server.ModifierEff
		want := base
		for i, slot := range []int{1, 2, 3, 6} {
			if mask&(1<<i) != 0 {
				present[i] = mods[i]
				color := mods[i].Color24
				want[slot] = noxrender.Color16{R: uint16(color.R), G: uint16(color.G), B: uint16(color.B)}
			}
		}
		if got := e2eEquipmentPalette(&def, present, initial); got != want || def != before {
			t.Fatalf("mask=%x got=%v want=%v or definition changed", mask, got, want)
		}
	}
	def.Effectiveness36, def.Material40, def.PriEnchant44, def.SecEnchant48 = 4, 4, 4, 4
	want := base
	want[4] = noxrender.Color16{R: 146, G: 124, B: 158}
	if got := e2eEquipmentPalette(&def, mods, initial); got != want {
		t.Fatalf("last modifier must win: got=%v want=%v", got, want)
	}
	def.Effectiveness36, def.Material40, def.PriEnchant44, def.SecEnchant48 = -1, 16, 0, 15
	want = base
	want[0] = noxrender.Color16{R: 167, G: 103, B: 137}
	want[15] = noxrender.Color16{R: 146, G: 124, B: 158}
	if got := e2eEquipmentPalette(&def, mods, initial); got != want {
		t.Fatalf("bounded modifier slots: got=%v want=%v", got, want)
	}
}

func TestE2EEquipmentLayerPixelsRejectsBlankAndCorrupt(t *testing.T) {
	rect := image.Rect(3, 4, 7, 6)
	actual, reference := noximage.NewImage16(rect), noximage.NewImage16(rect)
	if _, err := e2eEquipmentLayerPixels(actual, reference); err == nil {
		t.Fatal("blank images falsely prove a stock layer was rendered")
	}
	actual.Pix[0], reference.Pix[0] = 0x1234, 0x1234
	actual.Pix[5], reference.Pix[5] = 0x5678, 0x5678
	if count, err := e2eEquipmentLayerPixels(actual, reference); err != nil || count != 2 {
		t.Fatalf("matching stock pixels=%d err=%v", count, err)
	}
	actual.Pix[5] ^= 1
	if _, err := e2eEquipmentLayerPixels(actual, reference); err == nil {
		t.Fatal("one wrong RGB555 bit was not detected")
	}
	for _, candidate := range []*noximage.Image16{nil, noximage.NewImage16(image.Rectangle{}), noximage.NewImage16(image.Rect(0, 0, 4, 2))} {
		if _, err := e2eEquipmentLayerPixels(candidate, reference); err == nil {
			t.Fatal("incompatible bounds accepted")
		}
	}
	for _, length := range []int{1, 9} {
		short := *reference
		short.Pix = make([]uint16, length)
		short.Pix[0] = 0x1234
		if _, err := e2eEquipmentLayerPixels(&short, &short); err == nil {
			t.Fatal("nonempty but incomplete or oversized pixel storage accepted")
		}
	}
}
