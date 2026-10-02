//go:build !server

package opennox

import (
	"encoding/binary"
	"fmt"
	"image"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/noximage"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/client/render"
	"github.com/opennox/opennox/v1/client/seat/headless"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestEquippedWeaponColorsMatchGAMEEXE(t *testing.T) {
	checkEquippedColorsMatchGAMEEXE(t, true)
}

// This is a headless material/render integration, not GUI input or a stock
// campaign playback. GAME.EXE 004B8E10/004B8CA0 populate materials 1..6 from
// definition colors 1..6, then apply four modifier colors in slot order.
// Equipment, definitions, effects and render data are all native C-owned.
func checkEquippedColorsMatchGAMEEXE(t *testing.T, weapon bool) {
	t.Helper()
	previousFlags := noxflags.GetGame()
	noxflags.UnsetGame(noxflags.GameHost | noxflags.GameFlag22)
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(previousFlags)
	})
	s := server.New(nil, nil, strman.New())
	t.Cleanup(s.Close)
	typeIDs := make(map[string]int)
	s.Types.ClientTypeByID = func(name string) int {
		if typeIDs[name] == 0 {
			typeIDs[name] = len(typeIDs) + 1
		}
		return typeIDs[name]
	}
	s.Nox_xxx_equipWeapon_4157C0()
	s.Nox_xxx_equipArmor_415AB0()
	r := noxrender.NewRender(nil, s)
	t.Cleanup(r.Part.Free)
	d, freeData := noxrender.NewRenderData()
	t.Cleanup(freeData)
	r.SetData(d)
	c := &Client{Client: &client.Client{Server: s}, r: NewNoxRender(r)}
	previousClient := noxClient
	noxClient = c
	t.Cleanup(func() { noxClient = previousClient })
	def, freeDef := alloc.New(server.Modifier{})
	t.Cleanup(freeDef)
	bit := uint32(1 << 8)
	if weapon {
		def.TypeInd = uint32(s.Weapons.Sub_415840(bit))
		s.Modif.Dword_5d4594_251600 = def
	} else {
		def.TypeInd = uint32(s.Armor.Sub_415CD0(bit))
		s.Modif.Dword_5d4594_251608 = def
	}
	if def.TypeInd == 0 {
		t.Fatal("fixture equipment type is missing")
	}
	var effects [4]*server.ModifierEff
	for i := range effects {
		p, free := alloc.New(server.ModifierEff{})
		t.Cleanup(free)
		p.Color24 = types.RGB{R: byte(224 - 17*i), G: byte(31 + 37*i), B: byte(113 + 29*i)}
		p.Price20 = int32(0x10203040 + i)
		effects[i] = p
		assertEquippedColorNativeAddress(t, unsafe.Pointer(p))
	}
	assertEquippedColorNativeAddress(t, unsafe.Pointer(def))
	assertEquippedColorNativeAddress(t, d.C())
	pix := noximage.NewImage16(image.Rect(0, 0, 16, 4))
	r.SetPixBuffer(pix)
	sc := headless.New(pix.Size())
	t.Cleanup(func() { _ = sc.Close() })
	win, err := render.New(sc)
	if err != nil {
		t.Fatal(err)
	}
	// Every material, including neighboring slots, is sampled at four lumas.
	data := make([]byte, 17)
	binary.LittleEndian.PutUint32(data, 16)
	binary.LittleEndian.PutUint32(data[4:], 4)
	for _, luma := range []byte{0, 64, 128, 255} {
		for slot := 0; slot < 16; slot++ {
			data = append(data, byte(slot<<4)|4, 1, luma)
		}
	}
	img := noxrender.NewRawImage16(4, data)
	type paletteCase struct {
		name                             string
		mask                             byte
		slots                            [4]int32
		missingRecord, missingDefinition bool
	}
	var tests []paletteCase
	for mask := byte(0); mask < 16; mask++ {
		tests = append(tests, paletteCase{name: fmt.Sprintf("mask-%02x", mask), mask: mask, slots: [4]int32{1, 2, 3, 6}})
	}
	tests = append(tests,
		paletteCase{name: "duplicate-last-wins", mask: 15, slots: [4]int32{3, 3, 3, 3}},
		paletteCase{name: "disabled-and-neighbor-slots", mask: 15, slots: [4]int32{-1, 0, 16, 15}},
		paletteCase{name: "missing-record", mask: 15, slots: [4]int32{1, 2, 3, 4}, missingRecord: true},
		paletteCase{name: "missing-definition", mask: 15, slots: [4]int32{1, 2, 3, 4}, missingDefinition: true},
	)
	frames := 0
	for _, holderName := range []string{"player", "NPC"} {
		for _, last := range []bool{false, true} {
			for _, tc := range tests {
				t.Run(fmt.Sprintf("%s/last-%t/%s", holderName, last, tc.name), func(t *testing.T) {
					var want [16]types.RGB
					for i := range want {
						want[i] = types.RGB{R: byte(11 + 7*i), G: byte(213 - 5*i), B: byte(29 + 9*i)}
						d.SetMaterialRGB(i, int(want[i].R), int(want[i].G), int(want[i].B))
					}
					for i := range def.Colors12 {
						def.Colors12[i] = types.RGB{R: byte(35 + 23*i), G: byte(37 + 19*i), B: byte(237 - 21*i)}
					}
					*def.ColorIndexes() = tc.slots
					var mods [4]unsafe.Pointer
					for i, p := range effects {
						if tc.mask&(1<<i) != 0 {
							mods[i] = unsafe.Pointer(p)
						}
					}
					var h server.ArmorAndWeaponHolder
					if holderName == "player" {
						p, free := alloc.New(server.Player{})
						defer free()
						h = p
					} else {
						p, free := alloc.New(server.NPC{})
						defer free()
						h = p
					}
					var arr []server.EquipmentData
					if weapon {
						_, p := h.WeaponData()
						arr = p[:]
					} else {
						_, p := h.ArmorData()
						arr = p[:]
					}
					assertEquippedColorNativeAddress(t, unsafe.Pointer(&arr[0]))
					position := 0
					if last {
						position = len(arr) - 1
					}
					for i := 0; i < position; i++ {
						arr[i].Field0 = 0x80000000
					}
					if p, ok := h.(*server.Player); ok {
						server.ClientEquipPlayerNative417AA0(p, weapon, bit, mods)
					} else {
						server.ClientEquipNPCNative49A3D0(h.(*server.NPC), weapon, bit, mods)
					}
					arr[position].Field20 = 0x13579bdf
					if tc.missingRecord {
						arr[position].Field0 = 0
					}
					before := slices.Clone(arr)
					originalType := def.TypeInd
					if tc.missingDefinition {
						def.TypeInd = 0xffffffff
					}
					beforeDef := *def
					var beforeEffects [4]server.ModifierEff
					for i, p := range effects {
						beforeEffects[i] = *p
					}
					if weapon {
						_, p := h.WeaponData()
						c.sub_4B8E10(p, bit)
					} else {
						_, p := h.ArmorData()
						sub_4B8CA0(p, bit)
					}
					if !tc.missingRecord && !tc.missingDefinition {
						for i := 1; i <= 6; i++ {
							want[i] = def.Colors12[i]
						}
						for i, slot := range tc.slots {
							if mods[i] != nil && slot >= 0 && slot < 16 {
								want[slot] = effects[i].Color24
							}
						}
					}
					if !slices.Equal(arr, before) || *def != beforeDef {
						t.Fatal("material selection changed equipment or definition data")
					}
					for i, p := range effects {
						if *p != beforeEffects[i] {
							t.Fatal("material selection changed modifier effect data")
						}
					}
					def.TypeInd = originalType
					for i, cl := range want {
						got := d.ColorMultOp(i)
						if got != (noxrender.Color16{R: uint16(cl.R), G: uint16(cl.G), B: uint16(cl.B)}) {
							t.Errorf("material %d: got %+v, want %+v", i, got, cl)
						}
					}
					clear(pix.Pix)
					r.DrawImage16(img, image.Point{})
					for y, luma := range []uint16{0, 64, 128, 255} {
						for slot, cl := range want {
							rr := (uint16(cl.R) * luma) >> 8
							gg := (uint16(cl.G) * luma) >> 8
							bb := (uint16(cl.B) * luma) >> 8
							wantPixel := (rr>>3)<<10 | (gg>>3)<<5 | bb>>3
							if got := pix.Pix[16*y+slot]; got != wantPixel {
								t.Errorf("indexed pixel slot=%d luma=%d: got %#04x, want %#04x", slot, luma, got, wantPixel)
							}
						}
					}
					win.CopyBuffer(pix)
					shown, view := sc.Snapshot()
					if shown == nil || view != pix.Rect || !slices.Equal(shown.Pix, pix.Pix) {
						t.Fatal("headless presentation changed equipment colors")
					}
					frames++
				})
			}
		}
	}
	if sc.PresentCount() != uint64(frames) {
		t.Fatal("unexpected headless frame count")
	}
	t.Logf("EQUIPMENT COLORS: weapon=%t %d native player/NPC palettes, %d indexed pixels and headless frames=%d; effect color offset=%d", weapon, frames, frames*64, frames, unsafe.Offsetof(server.ModifierEff{}.Color24))
}

func assertEquippedColorNativeAddress(t *testing.T, p unsafe.Pointer) {
	t.Helper()
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(p) <= 0xffffffff {
		t.Fatalf("fixture must exercise a C-owned address above 4 GiB: %p", p)
	}
}
