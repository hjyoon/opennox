package legacy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

type maidenPaletteServer4BBF90 struct {
	Server
	srv *server.Server
}

func (s *maidenPaletteServer4BBF90) S() *server.Server { return s.srv }

type maidenPaletteRender4BBF90 struct {
	Render2
	data *noxrender.RenderData
}

func (r *maidenPaletteRender4BBF90) Data() *noxrender.RenderData { return r.data }

type maidenPaletteClient4BBF90 struct {
	Client
	render *maidenPaletteRender4BBF90
}

func (c *maidenPaletteClient4BBF90) R2() Render2 { return c.render }

// The transcript runs the unchanged PE32 MaidenDraw control flow with
// controlled flag, registry, material and MonsterDraw services. This test
// calls its registered C callback with real native records and material
// setters. It certifies palette/order/return behavior, not stock sprite pixels.
func TestMaidenDraw4BBF90MatchesGAMEEXE(t *testing.T) {
	var oracle struct {
		OriginalSHA string `json:"original_sha256"`
		FunctionSHA string `json:"function_sha256"`
		ReceiptSHA  string `json:"receipt_sha256"`
		Start       uint32 `json:"function_start"`
		Bytes       int    `json:"function_bytes"`
		Cases       int    `json:"cases"`
		Unchanged   bool   `json:"executable_bytes_unchanged"`
		Controlled  bool   `json:"all_called_services_controlled"`
		Rows        []struct {
			Name        string            `json:"name"`
			Preview     bool              `json:"preview"`
			ID          uint32            `json:"id"`
			Extents     []uint32          `json:"extents"`
			NPCPresent  bool              `json:"npc_present"`
			Colors      [6]uint32         `json:"colors"`
			RGB         [6][3]byte        `json:"rgb"`
			Mutate      bool              `json:"mutate"`
			DrawReturn  int32             `json:"draw_return"`
			NilViewport bool              `json:"nil_viewport"`
			Returned    int32             `json:"returned"`
			Faults      []json.RawMessage `json:"faults"`
			Materials   []struct {
				Slot   int      `json:"slot"`
				Packed *uint32  `json:"packed"`
				RGB    *[3]byte `json:"rgb"`
			} `json:"materials"`
		} `json:"rows"`
	}
	data, err := os.ReadFile("testdata/maiden_draw_4bbf90_original.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &oracle); err != nil {
		t.Fatal(err)
	}
	if oracle.OriginalSHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" ||
		oracle.FunctionSHA != "6eaa529742c365dec8394e31fe163eb12423670a3949ad764e1512b3224008ff" ||
		oracle.ReceiptSHA != "ccffdef4a7bf0839a9c34861962f4aff17182bde5dd90474a464dd2c4f79e0a2" ||
		oracle.Start != 0x4bbf90 || oracle.Bytes != 226 || oracle.Cases != 144 || len(oracle.Rows) != 144 || !oracle.Unchanged || !oracle.Controlled {
		t.Fatal("not the frozen unchanged GAME.EXE MaidenDraw transcript")
	}
	for _, tc := range oracle.Rows {
		t.Run(tc.Name, func(t *testing.T) {
			if len(tc.Faults) != 0 {
				t.Fatal("unexpected fault in original bounded case")
			}
			srv := new(server.Server)
			srv.NPCs.Init()
			npc := srv.NPCs.New(int(int32(tc.ID)))
			if npc == nil {
				t.Fatal("native NPC allocation failed")
			}
			t.Cleanup(func() { alloc.FreePtr(npc.C()) })
			npc.Color8 = tc.Colors
			if !tc.NPCPresent {
				npc.LiveVal = 0
			}
			vp, freeVP := alloc.New(noxrender.Viewport{})
			dr, freeDR := alloc.New(client.Drawable{})
			ud, freeUD := alloc.New(server.MonsterUpdateData{})
			other, freeOther := alloc.New(server.MonsterUpdateData{})
			renderData, freeRender := noxrender.NewRenderData()
			for _, free := range []func(){freeVP, freeDR, freeUD, freeOther, freeRender} {
				t.Cleanup(free)
			}
			dr.NetCode32 = tc.ID
			for i, rgb := range tc.RGB {
				ud.Color[i] = server.Color3{R: rgb[0], G: rgb[1], B: rgb[2]}
				other.Color[i] = server.Color3{R: 0xc3, G: 0xc3, B: 0xc3}
			}
			objects := make([]*server.Object, len(tc.Extents))
			for i, extent := range tc.Extents {
				obj, free := alloc.New(server.Object{})
				t.Cleanup(free)
				// Original MaidenDraw does not add a MONSTER class gate.
				obj.Extent, obj.UpdateData = extent, unsafe.Pointer(ud)
				objects[i] = obj
				if i == 0 {
					srv.Objs.List = obj
				} else {
					objects[i-1].ObjNext = obj
				}
			}
			pointers := []unsafe.Pointer{vp.C(), dr.C(), npc.C(), renderData.C(), unsafe.Pointer(ud), unsafe.Pointer(other), unsafe.Pointer(&ud.Color[0])}
			for _, obj := range objects {
				pointers = append(pointers, unsafe.Pointer(obj))
			}
			for _, ptr := range pointers {
				if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(ptr) <= math.MaxUint32 {
					t.Fatalf("native record=%p, want above 4 GiB", ptr)
				}
			}
			if unsafe.Sizeof(uintptr(0)) > 4 && unsafe.Offsetof(ud.Color) == 2076 {
				t.Fatal("fixture unexpectedly uses the packed PE32 update layout")
			}
			oldServer, oldClient, oldDraw, oldFlags := GetServer, GetClient, Nox_thing_monster_draw, noxflags.GetGame()
			t.Cleanup(func() {
				GetServer, GetClient, Nox_thing_monster_draw = oldServer, oldClient, oldDraw
				noxflags.ResetGame()
				noxflags.SetGame(oldFlags)
			})
			noxflags.ResetGame()
			if tc.Preview {
				noxflags.SetGame(noxflags.GameFlag22)
			}
			GetServer = func() Server { return &maidenPaletteServer4BBF90{srv: srv} }
			viewClient := &maidenPaletteClient4BBF90{render: &maidenPaletteRender4BBF90{data: renderData}}
			var initial [16]noxrender.Color16
			for slot := range initial {
				initial[slot] = noxrender.Color16{R: uint16(17 + slot), G: uint16(43 + slot), B: uint16(77 + slot)}
				renderData.SetMaterialRGB(slot, int(initial[slot].R), int(initial[slot].G), int(initial[slot].B))
			}
			want := initial
			apply := func(index int) {
				material := tc.Materials[index]
				if material.Slot != index+1 || (material.Packed == nil) == (material.RGB == nil) {
					t.Fatal("invalid original material order/shape")
				}
				if material.Packed != nil {
					word := uint16(*material.Packed)
					want[material.Slot] = noxrender.Color16{R: (word >> 10 & 31) << 3, G: (word >> 5 & 31) << 3, B: (word & 31) << 3}
				} else {
					rgb := *material.RGB
					want[material.Slot] = noxrender.Color16{R: uint16(rgb[0]), G: uint16(rgb[1]), B: uint16(rgb[2])}
				}
			}
			check := func(label string) {
				t.Helper()
				for slot, color := range want {
					if got := renderData.ColorMultOp(slot); got != color {
						t.Fatalf("%s material %d=%+v, want %+v", label, slot, got, color)
					}
				}
			}
			setters, draws := 0, 0
			GetClient = func() Client {
				check(fmt.Sprintf("before setter %d", setters+1))
				if setters >= len(tc.Materials) {
					t.Fatal("unexpected extra material setter")
				}
				apply(setters)
				setters++
				if tc.Mutate {
					// Only future colors are changed after their predecessor was
					// loaded. Cache the update pointer, not an array value or the
					// object's subsequently replaced data_update field.
					if setters < 6 {
						if tc.Preview {
							ud.Color[setters] = server.Color3{R: byte(setters * 31), G: byte(setters*31 + 79), B: byte(setters*31 + 158)}
						} else {
							npc.Color8[setters] = 0xaabb0000 + uint32(setters)*0x101
						}
					}
					if tc.Preview {
						for _, obj := range objects {
							obj.UpdateData = unsafe.Pointer(other)
						}
					}
				}
				return viewClient
			}
			argVP := vp
			if tc.NilViewport {
				argVP = nil
			}
			Nox_thing_monster_draw = func(gotVP *noxrender.Viewport, gotDR *client.Drawable) int {
				draws++
				if gotVP != argVP || gotDR != dr {
					t.Fatalf("MonsterDraw arguments=%p/%p, want %p/%p", gotVP, gotDR, argVP, dr)
				}
				check("MonsterDraw")
				return int(tc.DrawReturn)
			}
			beforeDR := bytes.Clone(unsafe.Slice((*byte)(dr.C()), int(unsafe.Sizeof(*dr))))
			beforeUD := bytes.Clone(unsafe.Slice((*byte)(unsafe.Pointer(ud)), int(unsafe.Sizeof(*ud))))
			got := int32(ccall.CallIntPtr2(Get_nox_thing_maiden_draw(), argVP.C(), dr.C()))
			if got != tc.Returned || setters != len(tc.Materials) {
				t.Fatalf("returned=%d/%d setters=%d/%d", got, tc.Returned, setters, len(tc.Materials))
			}
			wantDraws := 1
			if !tc.Preview && !tc.NPCPresent {
				wantDraws = 0
			}
			if draws != wantDraws {
				t.Fatalf("MonsterDraw calls=%d, want %d", draws, wantDraws)
			}
			check("return, no palette reset")
			if !bytes.Equal(beforeDR, unsafe.Slice((*byte)(dr.C()), int(unsafe.Sizeof(*dr)))) ||
				!tc.Mutate && !bytes.Equal(beforeUD, unsafe.Slice((*byte)(unsafe.Pointer(ud)), int(unsafe.Sizeof(*ud)))) {
				t.Fatal("MaidenDraw changed a read-only drawable or update record")
			}
		})
	}
}
