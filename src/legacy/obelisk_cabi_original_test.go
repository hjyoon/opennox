package legacy

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"math"
	"os"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func obeliskCFrameOriginal[T any](t *testing.T) (*T, []byte) {
	t.Helper()
	var zero T
	p, free := alloc.Make([]byte(nil), int(unsafe.Sizeof(zero))+64)
	t.Cleanup(free)
	for n := range p {
		p[n] = 0xa5
	}
	v := (*T)(unsafe.Pointer(&p[32]))
	*v = zero
	if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(unsafe.Pointer(v)) <= math.MaxUint32 {
		t.Fatal("native frame below 4 GiB")
	}
	return v, p
}

func TestObeliskCEntryOriginal(t *testing.T) {
	blob, err := os.ReadFile("../server/testdata/obelisk_startup_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		SHA        string            `json:"original_sha256"`
		Unchanged  bool              `json:"original_unchanged"`
		CW         uint16            `json:"actual_startup_control_word"`
		Contexts   int               `json:"contexts"`
		Executions int               `json:"root_executions"`
		Faults     []json.RawMessage `json:"memory_faults"`
		Rows       []struct {
			Mode  string `json:"mode"`
			Input struct {
				Kind         string    `json:"kind"`
				Name         string    `json:"name"`
				Count        int       `json:"count"`
				Source       [2]uint32 `json:"source"`
				Mana         int32     `json:"mana"`
				LastFrame    uint32    `json:"last_frame"`
				Frame        uint32    `json:"frame"`
				TickRate     uint32    `json:"tick_rate"`
				ItemClass    uint32    `json:"item_class"`
				ItemSubclass uint32    `json:"item_subclass"`
				Progress     int32     `json:"progress"`
				Charge       uint8     `json:"charge"`
				MaxCharge    uint8     `json:"max_charge"`
				Amount       int32     `json:"amount"`
				Rate         int32     `json:"rate"`
				Slot         int       `json:"slot"`
				NilUse       bool      `json:"nil_use"`
				NilItem      bool      `json:"nil_item"`
				ZeroObjects  bool      `json:"zero_objects"`
			} `json:"input"`
			ReturnRaw   uint32      `json:"return_raw"`
			ReturnOwner bool        `json:"return_owner_address"`
			Preserved   bool        `json:"nonoutput_bytes_preserved"`
			ObjectWords [][3]uint32 `json:"object_words"`
			Mana        []uint32    `json:"mana"`
			Wands       [][3]uint32 `json:"wands"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(blob, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || !fixture.Unchanged || fixture.CW != 0xe3f || fixture.Contexts != 772 || fixture.Executions != 1544 || len(fixture.Rows) != 1544 || len(fixture.Faults) != 0 {
		t.Fatal("incomplete original Obelisk fixture")
	}
	srv := server.New(nil, nil, strman.New())
	t.Cleanup(srv.Close)
	oldServer := GetServer
	GetServer = func() Server { return &obeliskUpdateLegacyServer53C580{srv: srv} }
	t.Cleanup(func() { GetServer = oldServer })
	counts := map[string]int{}
	for _, row := range fixture.Rows {
		i := row.Input
		if row.Mode != "actual-startup" || (i.Kind == "main" && i.Count != 0) || (i.Kind == "rate" && i.ItemClass&0x1000 != 0 && i.ItemSubclass&0x4000000 != 0) {
			continue
		}
		// These C entries execute real charge, modifier and passive-sync paths.
		// Live-player callbacks and balance contexts are covered by the server
		// matrix, without sending artificial protection tokens into the runtime.
		counts[i.Kind]++
		t.Run(i.Name, func(t *testing.T) {
			owner, ownerFrame := obeliskCFrameOriginal[server.Object](t)
			item, itemFrame := obeliskCFrameOriginal[server.Object](t)
			mana, manaFrame := obeliskCFrameOriginal[server.ObeliskUpdateData](t)
			wand, wandFrame := obeliskCFrameOriginal[server.WandUseData](t)
			attrs, attrsFrame := obeliskCFrameOriginal[server.ModifierInitData](t)
			effect, effectFrame := obeliskCFrameOriginal[server.ModifierEff](t)
			owner.PosVec = rigidNativePoint(i.Source)
			owner.UpdateData = unsafe.Pointer(mana)
			owner.Field34 = i.LastFrame
			item.ObjClass, item.ObjSubClass = object.Class(i.ItemClass), object.SubClass(i.ItemSubclass)
			item.InitData, item.UseData.Ptr = unsafe.Pointer(attrs), unsafe.Pointer(wand)
			if i.NilUse {
				item.UseData.Ptr = nil
			}
			if !i.ZeroObjects {
				owner.Field38, item.Field38 = 0x11223344, 0x11223344
			}
			mana.Mana = i.Mana
			wand.Progress, wand.Charge, wand.MaxCharge = uint32(i.Progress), i.Charge, i.MaxCharge
			effect.Attack40.Fnc, effect.Attack40.Val = obeliskReplenishmentEffectC53C940(), i.Rate
			if i.Slot == 2 || i.Slot == 3 {
				attrs.Modifiers[i.Slot] = effect
			}
			frames := [][]byte{ownerFrame, itemFrame, manaFrame, wandFrame, attrsFrame, effectFrame}
			want := make([][]byte, len(frames))
			for n, p := range frames {
				want[n] = bytes.Clone(p)
			}
			binary.LittleEndian.PutUint32(want[0][32+unsafe.Offsetof(owner.ObjFlags):], row.ObjectWords[0][0])
			binary.LittleEndian.PutUint32(want[0][32+unsafe.Offsetof(owner.Field34):], row.ObjectWords[0][1])
			binary.LittleEndian.PutUint32(want[0][32+unsafe.Offsetof(owner.Field38):], row.ObjectWords[0][2])
			binary.LittleEndian.PutUint32(want[2][32:], row.Mana[0])
			want[3][32+unsafe.Offsetof(wand.Charge)] = byte(row.Wands[0][0])
			want[3][32+unsafe.Offsetof(wand.MaxCharge)] = byte(row.Wands[0][1])
			binary.LittleEndian.PutUint32(want[3][32+unsafe.Offsetof(wand.Progress):], row.Wands[0][2])
			srv.SetFrame(i.Frame)
			srv.SetTickRate(i.TickRate)
			var got int
			switch i.Kind {
			case "main":
				got = obeliskUpdateC53C580(owner)
			case "charge":
				got = int(obeliskRechargeItemC53C520(item, i.Amount))
			case "rate":
				p := item
				if i.NilItem {
					p = nil
				}
				got = int(obeliskRechargeRateC53C940(p))
			default:
				t.Fatal("unexpected original entry")
			}
			result := int(int32(row.ReturnRaw))
			if row.ReturnOwner {
				result = int(uintptr(unsafe.Pointer(owner)))
			}
			if !row.Preserved || got != result {
				t.Fatalf("return=%d original=%d preserved=%t", got, result, row.Preserved)
			}
			for n, p := range frames {
				if !bytes.Equal(p, want[n]) {
					t.Fatalf("native frame %d differs from original output", n)
				}
			}
		})
	}
	if counts["main"] != 43 || counts["charge"] != 169 || counts["rate"] != 22 || len(counts) != 3 {
		t.Fatalf("C entry coverage=%v", counts)
	}
}
