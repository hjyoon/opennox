package server

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

type continualOriginal218 struct {
	OriginalSHA string `json:"original_sha256"`
	Function    struct {
		Address string
		Bytes   int
		SHA256  string
	}
	Executions int
	Guards     bool `json:"code_input_guards_and_callee_saved_registers_verified"`
	Cases      []struct {
		Input struct {
			Name            string
			Frame, Rate     uint32
			Charge, Maximum uint8
			Cls, Owner      uint32
			Index           uint8
			NilItem         bool `json:"nil_item"`
			NilEffect       bool `json:"nil_effect"`
			NilUse          bool `json:"nil_use"`
			NilUpdate       bool `json:"nil_update"`
			NilPlayer       bool `json:"nil_player"`
		}
		Expected struct {
			Charge  uint8
			Reports [][3]uint32
			Fault   *string
		}
	}
}

func TestContinualReplenishment4E02C0MatchesWholeOriginalAndFaultOrdering(t *testing.T) {
	data, err := os.ReadFile("../legacy/testdata/continual_replenishment_original_4e02c0_v2.json")
	if err != nil {
		t.Fatal(err)
	}
	var original continualOriginal218
	if err = json.Unmarshal(data, &original); err != nil {
		t.Fatal(err)
	}
	if original.OriginalSHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || original.Function.Address != "0x4e02c0" || original.Function.Bytes != 137 || original.Function.SHA256 != "56ba4d3256e141aacd6eeae507571aebead435fefb43e2205049fcf34c93edc0" || len(original.Cases) != 1484 || original.Executions != 4452 || !original.Guards {
		t.Fatal("sealed original coverage changed")
	}
	effect, freeEffect := alloc.New(ModifierEff{})
	defer freeEffect()
	item, freeItem := alloc.New(Object{})
	defer freeItem()
	use, freeUse := alloc.New(WandUseData{})
	defer freeUse()
	owner, freeOwner := alloc.New(Object{})
	defer freeOwner()
	update, freeUpdate := alloc.New(PlayerUpdateData{})
	defer freeUpdate()
	player, freePlayer := alloc.New(Player{})
	defer freePlayer()
	regions := []struct {
		p unsafe.Pointer
		n uintptr
	}{
		{effect.C(), unsafe.Sizeof(*effect)}, {item.CObj(), unsafe.Sizeof(*item)},
		{unsafe.Pointer(use), unsafe.Sizeof(*use)}, {owner.CObj(), unsafe.Sizeof(*owner)},
		{unsafe.Pointer(update), unsafe.Sizeof(*update)}, {unsafe.Pointer(player), unsafe.Sizeof(*player)},
	}
	for _, r := range regions {
		if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(r.p) <= 0xffffffff {
			t.Fatalf("native address %p below4GiB", r.p)
		}
	}
	executions := 0
	for index, row := range original.Cases {
		for _, pattern := range []byte{0, 0x5a, 0xff} {
			for _, r := range regions {
				for i := range unsafe.Slice((*byte)(r.p), r.n) {
					unsafe.Slice((*byte)(r.p), r.n)[i] = pattern
				}
			}
			in := row.Input
			effect.Update100.Val = int32(in.Rate)
			item.ObjClass = object.Class(in.Cls)
			item.UseData.Ptr = unsafe.Pointer(use)
			if in.NilUse {
				item.UseData.Ptr = nil
			}
			item.InvHolder = owner
			if in.Owner == 0 {
				item.InvHolder = nil
			}
			owner.ObjClass = object.Class(in.Owner)
			owner.UpdateData = unsafe.Pointer(update)
			if in.NilUpdate {
				owner.UpdateData = nil
			}
			update.Player = player
			if in.NilPlayer {
				update.Player = nil
			}
			player.PlayerInd = in.Index
			use.Charge, use.MaxCharge = in.Charge, in.Maximum
			before := make([][]byte, len(regions))
			for i, r := range regions {
				before[i] = bytes.Clone(unsafe.Slice((*byte)(r.p), r.n))
			}
			before[2][108] = row.Expected.Charge
			mod, it := effect, item
			if in.NilEffect {
				mod = nil
			}
			if in.NilItem {
				it = nil
			}
			frames := 0
			reports := make([][3]uint32, 0)
			var caught any
			func() {
				defer func() { caught = recover() }()
				EffectContinualReplenishment4E02C0(mod, it, func() uint32 { frames++; return in.Frame }, func(ind uint8, got *Object, charge, max uint8) {
					if got != item {
						t.Fatalf("case%d report changed item", index)
					}
					reports = append(reports, [3]uint32{uint32(ind), uint32(charge), uint32(max)})
				})
			}()
			if (caught != nil) != (row.Expected.Fault != nil) {
				t.Fatalf("case%d %s pattern%x fault=%v expected=%v", index, in.Name, pattern, caught, row.Expected.Fault)
			}
			if row.Expected.Fault != nil && *row.Expected.Fault == "divide" && !strings.Contains(caught.(error).Error(), "divide by zero") {
				t.Fatalf("case%d wrong divide fault %v", index, caught)
			}
			wantFrames := 1
			if in.NilItem {
				wantFrames = 0
			}
			if frames != wantFrames || !reflect.DeepEqual(reports, row.Expected.Reports) {
				t.Fatalf("case%d frames/reports=%d/%v want%d/%v", index, frames, reports, wantFrames, row.Expected.Reports)
			}
			for i, r := range regions {
				if !bytes.Equal(before[i], unsafe.Slice((*byte)(r.p), r.n)) {
					t.Fatalf("case%d %s guard%x region%d changed beyond original byte write", index, in.Name, pattern, i)
				}
			}
			executions++
		}
	}
	if executions != 4452 {
		t.Fatal("coverage changed")
	}
	t.Log("whole original callback comparison: 1484 cases x3 native guard patterns, including5 fault cases x3")
}

func TestContinualReplenishment4E02C0ReportSeesUpdatedByteAndDoesNotWriteProgress(t *testing.T) {
	effect := ModifierEff{Update100: ModifierEffFnc{Val: 240}}
	data := WandUseData{Charge: 255, MaxCharge: 255, Progress: 0x12345678}
	player := Player{PlayerInd: 0xfe}
	update := PlayerUpdateData{Player: &player}
	owner := Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(&update)}
	item := Object{ObjClass: object.ClassWand, UseData: UseDataPtr{Ptr: unsafe.Pointer(&data)}, InvHolder: &owner}
	calls := 0
	EffectContinualReplenishment4E02C0(&effect, &item, func() uint32 { return 240 }, func(index uint8, got *Object, charge, max uint8) {
		calls++
		if index != 0xfe || got != &item || charge != 0 || max != 255 || data.Charge != 0 || data.Progress != 0x12345678 {
			t.Fatal("byte wrap/report ordering/progress changed")
		}
		// No follow-up traversal or write is permitted after the original report.
		item.InvHolder = nil
		data.Charge = 17
		data.MaxCharge = 19
		data.Progress = 23
	})
	if calls != 1 || data.Charge != 17 || data.MaxCharge != 19 || data.Progress != 23 {
		t.Fatal("callback mutation was overwritten")
	}
}
