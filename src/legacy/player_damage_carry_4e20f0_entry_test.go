package legacy

import (
	"encoding/json"
	"math"
	"os"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestPlayerDamageCarryPublicCEntry4E20F0(t *testing.T) {
	data, err := os.ReadFile("../server/testdata/player_damage_carry_precision_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var original struct {
		Cases []struct {
			Name        string       `json:"name"`
			Class       object.Class `json:"class_bits"`
			Amount      uint32       `json:"amount_bits32"`
			Carry       uint32       `json:"carry_bits32"`
			TargetNil   bool         `json:"target_nil"`
			OutputNil   bool         `json:"output_nil"`
			WantOutput  uint32       `json:"expected_output_bits32"`
			WantMonster uint32       `json:"expected_monster_carry_bits32"`
			WantPlayer  uint32       `json:"expected_player_carry_bits32"`
		} `json:"carry_cases"`
	}
	if err := json.Unmarshal(data, &original); err != nil {
		t.Fatal(err)
	}
	if len(original.Cases) != 704 {
		t.Fatalf("original carry cases = %d, want 704", len(original.Cases))
	}
	target, freeTarget := alloc.NewHigh(server.Object{})
	t.Cleanup(freeTarget)
	size := max(unsafe.Sizeof(server.PlayerUpdateData{}), unsafe.Sizeof(server.MonsterUpdateData{}))
	update, freeUpdate := alloc.MakeHigh([]byte(nil), int(size))
	t.Cleanup(freeUpdate)
	output, freeOutput := alloc.NewHigh([3]int32{})
	t.Cleanup(freeOutput)
	target.UpdateData = unsafe.Pointer(&update[0])
	player := (*server.PlayerUpdateData)(target.UpdateData)
	monster := (*server.MonsterUpdateData)(target.UpdateData)
	for _, ptr := range []unsafe.Pointer{unsafe.Pointer(target), target.UpdateData, unsafe.Pointer(&output[1])} {
		if cgoABIPointerSize == 8 && uintptr(ptr) <= math.MaxUint32 {
			t.Fatalf("fixture pointer %p must be above 4 GiB", ptr)
		}
	}
	const sentinel = 0x6b6b6b6b
	for _, tc := range original.Cases {
		t.Run(tc.Name, func(t *testing.T) {
			target.ObjClass = tc.Class
			player.Field21, monster.Field1 = tc.Carry, tc.Carry
			*output = [3]int32{sentinel, sentinel, sentinel}
			obj, out := target, &output[1]
			if tc.TargetNil {
				obj = nil
			}
			if tc.OutputNil {
				out = nil
			}
			playerDamageCarryPublicCall4E20F0(obj, out, math.Float32frombits(tc.Amount))
			if uint32(output[1]) != tc.WantOutput || player.Field21 != tc.WantPlayer || monster.Field1 != tc.WantMonster {
				t.Fatalf("C result/output bits = %08x/%08x/%08x, want %08x/%08x/%08x",
					output[1], player.Field21, monster.Field1, tc.WantOutput, tc.WantPlayer, tc.WantMonster)
			}
			if output[0] != sentinel || output[2] != sentinel || target.ObjClass != tc.Class {
				t.Fatal("public C entry changed adjacent data")
			}
		})
	}
	t.Run("reload-class-after-output-alias", func(t *testing.T) {
		target.ObjClass = object.ClassMonster
		player.Field21, monster.Field1 = 0, math.Float32bits(0.5)
		playerDamageCarryPublicCall4E20F0(target, (*int32)(unsafe.Pointer(&target.ObjClass)), 3.75)
		if target.ObjClass != object.ClassPlayer || player.Field21 != math.Float32bits(0.25) || monster.Field1 != math.Float32bits(0.5) {
			t.Fatal("public C entry must store output before selecting the carry write destination")
		}
	})
}
