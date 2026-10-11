//go:build !server

package opennox

import (
	"encoding/binary"
	"encoding/json"
	"image"
	"os"
	"slices"
	"testing"
	"unsafe"

	noxcolor "github.com/opennox/libs/color"
	"github.com/opennox/libs/noximage"
	"github.com/opennox/libs/strman"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/client/render"
	"github.com/opennox/opennox/v1/client/seat/headless"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/internal/npcdrawprobe"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func init() {
	client.RegisterDraw("NPCPaletteReadOnlyTestBoundary", npcdrawprobe.Callback(), 0, nil)
}

// The frozen fixture executes the unchanged GAME.EXE NPCDraw, with explicit
// material, empty equipment, frame, cached top-effect and non-ally services.
// This test calls the real native-width DrawNPC and C draw callbacks. It then
// renders every captured palette through the real indexed renderer and a
// headless seat. Null body images and a read-only top-effect callback deliberately
// exclude stock sprite pixels, GUI input and whole-original-world equivalence.
func TestDrawNPCPaletteMatchesUnchangedGAMEEXE(t *testing.T) {
	var oracle struct {
		OriginalSHA string            `json:"original_sha256"`
		FunctionSHA string            `json:"function_sha256"`
		Start       uint32            `json:"function_start"`
		Bytes       int               `json:"function_bytes"`
		Entries     map[string]int    `json:"unhooked_entries"`
		Faults      []json.RawMessage `json:"memory_faults"`
		Unchanged   bool              `json:"executable_bytes_unchanged"`
		Initial     [16]uint16        `json:"initial"`
		Colors      [6]uint32         `json:"npc_colors"`
		Rows        []struct {
			Name           string       `json:"name"`
			Buffs          uint32       `json:"buffs"`
			Poison         uint32       `json:"poison"`
			Frame          uint32       `json:"frame"`
			Direction      byte         `json:"direction"`
			Count          uint16       `json:"count"`
			MissingNPC     bool         `json:"missing_npc"`
			Returned       int          `json:"returned"`
			Effects        []string     `json:"effect_names"`
			EffectPalettes [][16]uint16 `json:"effect_palettes"`
			FinalPalette   [16]uint16   `json:"final_palette"`
			GuardsIntact   bool         `json:"guarded_bytes_unchanged"`
		} `json:"rows"`
	}
	data, err := os.ReadFile("testdata/draw_npc_palette_4b9050.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &oracle); err != nil {
		t.Fatal(err)
	}
	if oracle.OriginalSHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" ||
		oracle.FunctionSHA != "f5d3eb7b66c8b8cec76784cf0b1f1e1e3d34a075491a57511650bbe387ac545c" ||
		oracle.Start != 0x4b9050 || oracle.Bytes != 1047 || !oracle.Unchanged ||
		len(oracle.Faults) != 0 || len(oracle.Rows) != 74 || oracle.Entries["0x4b9050"] != 74 {
		t.Fatal("not the frozen unchanged NPCDraw transcript")
	}
	previousFlags := noxflags.GetGame()
	noxflags.UnsetGame(noxflags.GameHost | noxflags.GameFlag22)
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(previousFlags)
	})
	previousWhite, previousBlue := nox_color_white_2523948, nox_color_blue_2650684
	nox_color_white_2523948, nox_color_blue_2650684 = noxcolor.RGBA5551(0x7fff), noxcolor.RGBA5551(0x001f)
	t.Cleanup(func() {
		nox_color_white_2523948, nox_color_blue_2650684 = previousWhite, previousBlue
	})
	// No HP-bar state is supplied by this bounded oracle. Preserve the entire
	// existing ally table instead of leaking fixture state into other tests.
	allyBytes := unsafe.Slice((*byte)(memmap.PtrOff(0x5D4594, 1200916)), 512)
	previousAllyBytes := slices.Clone(allyBytes)
	clear(allyBytes)
	t.Cleanup(func() { copy(allyBytes, previousAllyBytes) })
	s := server.New(nil, nil, strman.New())
	t.Cleanup(s.Close)
	cl := client.NewClient(nil, nil, s)
	t.Cleanup(cl.Close)
	t.Cleanup(func() { alloc.FreePtr(cl.Viewport().C()) })
	cl.Objs.Init(4)
	t.Cleanup(cl.Objs.Free)
	r := cl.Render()
	t.Cleanup(r.Part.Free)
	d, freeData := noxrender.NewRenderData()
	t.Cleanup(freeData)
	r.SetData(d)
	c := &Client{Client: cl, r: NewNoxRender(r)}
	previousClient := noxClient
	noxClient = c
	t.Cleanup(func() { noxClient = previousClient })
	s.NPCs.Init()
	npc := s.NPCs.New(0x1234)
	if npc == nil {
		t.Fatal("native NPC allocation failed")
	}
	// Init allocates the array once; New above is its first, base entry.
	t.Cleanup(func() { alloc.FreePtr(npc.C()) })
	host, dr, skull, crown := cl.Objs.New(), cl.Objs.New(), cl.Objs.New(), cl.Objs.New()
	if host == nil || dr == nil || skull == nil || crown == nil {
		t.Fatal("native drawable allocation failed")
	}
	dd, freeDD := alloc.New(client.PlayerDrawData{})
	t.Cleanup(freeDD)
	naked, freeNaked := alloc.New(client.PlayerEquipAnimation{})
	t.Cleanup(freeNaked)
	frames, freeFrames := alloc.Make([]noxrender.ImageHandle{}, 1)
	t.Cleanup(freeFrames)
	dd.Size = uint32(unsafe.Sizeof(*dd))
	naked.Size = uint32(unsafe.Sizeof(*naked))
	for i := range naked.Frames {
		naked.Frames[i] = &frames[0]
	}
	dd.Anim[0].Naked = naked
	dd.Anim[0].Base.Kind = client.AnimSlave
	host.DrawData = unsafe.Pointer(dd)
	dr.NetCode32 = 0x1234
	dr.PosVec = image.Pt(63, 123)
	skull.DrawFuncPtr, crown.DrawFuncPtr = npcdrawprobe.Callback(), npcdrawprobe.Callback()
	c.info.skull, c.info.crown = skull, crown
	playerSlot := memmap.PtrPtr(0x852978, 8)
	previousPlayer := *playerSlot
	*playerSlot = host.C()
	t.Cleanup(func() { *playerSlot = previousPlayer })
	vp := cl.Viewport()
	vp.Screen.Min, vp.World.Min = image.Pt(7, 11), image.Pt(13, 23)
	for _, p := range []unsafe.Pointer{d.C(), npc.C(), host.C(), dr.C(), skull.C(), crown.C(),
		vp.C(), unsafe.Pointer(dd), unsafe.Pointer(naked), unsafe.Pointer(&frames[0])} {
		assertNPCPaletteNativeAddress(t, p)
	}
	pix := noximage.NewImage16(image.Rect(0, 0, 16, 4))
	r.SetPixBuffer(pix)
	sc := headless.New(pix.Size())
	t.Cleanup(func() { _ = sc.Close() })
	win, err := render.New(sc)
	if err != nil {
		t.Fatal(err)
	}
	imageData := make([]byte, 17)
	binary.LittleEndian.PutUint32(imageData, 16)
	binary.LittleEndian.PutUint32(imageData[4:], 4)
	for _, luma := range []byte{0, 64, 128, 255} {
		for slot := 0; slot < 16; slot++ {
			imageData = append(imageData, byte(slot<<4)|4, 1, luma)
		}
	}
	img := noxrender.NewRawImage16(4, imageData)
	presentations, effectCalls := 0, 0
	checkPalette := func(t *testing.T, label string, got *noxrender.RenderData, want [16]uint16) {
		t.Helper()
		assertNPCPaletteNativeAddress(t, got.C())
		for slot, word := range want {
			if color := got.ColorMultOp(slot); color != npcPaletteReferenceColor(word) {
				t.Errorf("%s material %d: got %+v, want word=%04x RGB=%+v", label, slot, color, word, npcPaletteReferenceColor(word))
			}
		}
		r.SetData(got)
		clear(pix.Pix)
		r.DrawImage16(img, image.Point{})
		for y, luma := range []uint16{0, 64, 128, 255} {
			for slot, word := range want {
				cl := npcPaletteReferenceColor(word)
				rr, gg, bb := cl.R*luma>>8, cl.G*luma>>8, cl.B*luma>>8
				wantPixel := (rr>>3)<<10 | (gg>>3)<<5 | bb>>3
				if pixel := pix.Pix[16*y+slot]; pixel != wantPixel {
					t.Errorf("%s indexed pixel slot=%d luma=%d: got=%04x want=%04x", label, slot, luma, pixel, wantPixel)
				}
			}
		}
		win.CopyBuffer(pix)
		shown, view := sc.Snapshot()
		if shown == nil || view != pix.Rect || !slices.Equal(shown.Pix, pix.Pix) {
			t.Fatal("headless presentation changed the palette sample")
		}
		presentations++
		r.SetData(d)
	}
	for _, tc := range oracle.Rows {
		t.Run(tc.Name, func(t *testing.T) {
			if !tc.GuardsIntact || len(tc.Effects) != len(tc.EffectPalettes) {
				t.Fatal("invalid frozen boundary evidence")
			}
			d.Reset()
			for slot, word := range oracle.Initial {
				d.SetMaterial(slot, noxcolor.RGBA5551(word))
			}
			npc.Color8, npc.Field1312, npc.LiveVal = oracle.Colors, tc.Poison, 1
			if tc.MissingNPC {
				npc.LiveVal = 0
			}
			s.SetFrame(tc.Frame)
			dr.Buffs, dr.AnimDir = tc.Buffs, tc.Direction
			dd.Anim[0].Base.Cnt40 = tc.Count
			beforeNPC, beforeDD, beforeNaked, beforeDR := *npc, *dd, *naked, *dr
			recorder := npcdrawprobe.New(d.C(), unsafe.Sizeof(*d))
			defer recorder.Close()
			assertNPCPaletteNativeAddress(t, recorder.C())
			if result := c.DrawNPC(vp, dr); result != tc.Returned {
				t.Errorf("DrawNPC returned %d, want %d", result, tc.Returned)
			}
			if *npc != beforeNPC || *dd != beforeDD || *naked != beforeNaked || *dr != beforeDR || frames[0] != nil || host.DrawData != unsafe.Pointer(dd) {
				t.Fatal("DrawNPC modified a read-only native NPC, animation or drawable record")
			}
			if !recorder.GuardsIntact() || recorder.Count() != len(tc.Effects) {
				t.Fatalf("effect boundaries: calls=%d want=%d guards=%t", recorder.Count(), len(tc.Effects), recorder.GuardsIntact())
			}
			for i, name := range tc.Effects {
				effect := skull
				if name == "crown" {
					effect = crown
				} else if name != "skull" {
					t.Fatalf("unexpected original effect %q", name)
				}
				if recorder.Drawable(i) != uintptr(effect.C()) || effect.PosVec != image.Pt(57, 61) {
					t.Fatal("top-effect order, native drawable address or position differs")
				}
				checkPalette(t, name+"-boundary", (*noxrender.RenderData)(recorder.Snapshot(i)), tc.EffectPalettes[i])
				effectCalls++
			}
			checkPalette(t, "after-NPC", d, tc.FinalPalette)
		})
	}
	if presentations != 146 || effectCalls != 72 || sc.PresentCount() != uint64(presentations) {
		t.Fatalf("missing native/headless evidence: frames=%d effects=%d present=%d", presentations, effectCalls, sc.PresentCount())
	}
	t.Logf("NPC PALETTE: 74 unchanged PE32 cases, 72 real C top-effect boundaries, 146 headless palette frames, 9344 indexed pixels; native C-owned addresses >4 GiB; stock sprite/HP/preview/wholeworld NOT claimed")
}

func npcPaletteReferenceColor(word uint16) noxrender.Color16 {
	return noxrender.Color16{R: ((word >> 10) & 31) * 8, G: ((word >> 5) & 31) * 8, B: (word & 31) * 8}
}

func assertNPCPaletteNativeAddress(t *testing.T, p unsafe.Pointer) {
	t.Helper()
	if p == nil || unsafe.Sizeof(uintptr(0)) == 8 && uintptr(p) <= 0xffffffff {
		t.Fatalf("fixture needs a C-owned pointer above 4 GiB: %p", p)
	}
}
