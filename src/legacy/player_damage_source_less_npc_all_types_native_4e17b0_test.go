package legacy

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"runtime"
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	noxflags "github.com/opennox/opennox/v1/common/flags"
)

// Original PlayerDamage/carry/DefaultDamage execute together in the fixture
// oracle, with armor wear, scale1.0, protection0 and final HP as service
// boundaries. Here the production C dispatcher and real UnitSetHP execute.
// Raw -1 is the callback ABI; public DoDamage(DamageTrue) is a separate caller.
func TestSourceLessNPCAllTypesNativeCallback4E17B0(t *testing.T) {
	_ = npcReflectLegacyServer4E17B0(t)
	blob, err := os.ReadFile("../server/testdata/player_damage_source_less_npc_integrated_4e17b0.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		OriginalSHA string `json:"original_sha256"`
		Rows        []struct {
			Type     object.DamageType    `json:"damage_type"`
			Subclass object.SubClass      `json:"subclass"`
			Forward  []uint32             `json:"forward_damage"`
			Carry    uint32               `json:"carry_bits"`
			Marker   uint32               `json:"marker"`
			Kind     uint32               `json:"marker_type"`
			Status   object.MonsterStatus `json:"status_flags"`
			Electric []byte               `json:"electric_state_bytes"`
			Return   int                  `json:"return"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(blob, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.OriginalSHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || len(fixture.Rows) != 120 {
		t.Fatal("unexpected original integrated NPC fixture")
	}
	for _, row := range fixture.Rows {
		for _, mode := range []struct{ god, nilHP bool }{{}, {god: true}, {nilHP: true}} {
			t.Run(fmt.Sprintf("type-%d/subclass-%x/god-%t/nil-hp-%t", int32(row.Type), uint32(row.Subclass), mode.god, mode.nilHP), func(t *testing.T) {
				if mode.god {
					noxflags.SetEngine(noxflags.EngineGodMode)
				} else {
					noxflags.UnsetEngine(noxflags.EngineGodMode)
				}
				var pin runtime.Pinner
				defer pin.Unpin()
				target, player := npcReflectLegacyObjects4E17B0(t, &pin)
				if mode.nilHP {
					target.HealthData = nil
				}
				target.ObjSubClass = row.Subclass
				target.Obj130, target.Pos132, target.Frame134 = player, types.Ptf(29, 31), 77
				ud := target.UpdateDataMonster()
				ud.Field518, ud.Field1 = math.Float32bits(.5), math.Float32bits(.25)
				ud.Field523_2, ud.Field523_3 = 0x11, 0xab
				if len(row.Forward) > 1 || len(row.Electric) != 2 {
					t.Fatal("invalid original boundary/state vector")
				}
				wantHP := uint16(200)
				if len(row.Forward) != 0 {
					wantHP -= uint16(row.Forward[0])
				}
				result := objectDamageDispatchCallNative(target, nil, nil, 9, row.Type)
				hpOK := target.HealthData == nil
				if !mode.nilHP {
					hpOK = target.HealthData != nil && target.HealthData.Cur == wantHP
				}
				if result != (row.Return != 0) || !hpOK || ud.Field1 != row.Carry || ud.Field547 != row.Marker || ud.Field546 != row.Kind || ud.StatusFlags != row.Status || ud.Field523_2 != row.Electric[0] || ud.Field523_3 != row.Electric[1] {
					t.Fatalf("C NPC result=%t HP-valid=%t want=%d carry=%x/%x marker=%d/%d kind=%d/%d status=%x/%x electric=%x,%x", result, hpOK, wantHP, ud.Field1, row.Carry, ud.Field547, row.Marker, ud.Field546, row.Kind, ud.StatusFlags, row.Status, ud.Field523_2, ud.Field523_3)
				}
				if len(row.Forward) == 0 {
					if target.Field38 != 0 || target.Obj130 != player || target.Pos132 != types.Ptf(29, 31) || target.Frame134 != 77 {
						t.Fatal("immune NPC crossed attribution/HP boundary")
					}
				} else {
					wantSync := uint32(math.MaxUint32)
					if mode.nilHP {
						wantSync = 0
					}
					if target.Field38 != wantSync || target.Obj130 != nil || target.Pos132 != (types.Pointf{}) || target.Frame134 != 1400 || target.Field131 != uint32(row.Type) {
						t.Fatal("source-less NPC attribution/HP sync")
					}
				}
				runtime.KeepAlive(target)
				runtime.KeepAlive(player)
			})
		}
	}
}
