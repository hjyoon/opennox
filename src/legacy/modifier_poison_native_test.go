package legacy

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

func TestModifierPoisonOriginalNativeC(t *testing.T) {
	blob, err := os.ReadFile("testdata/modifier_poison_original.json")
	if err != nil {
		t.Fatal(err)
	}
	type event struct {
		Service        string   `json:"service"`
		Direction      uint32   `json:"direction_bits32"`
		Maximum        uint32   `json:"maximum_bits32"`
		Increment      int32    `json:"increment"`
		TargetPosition []uint32 `json:"target_position_bits32"`
		SourcePosition []uint32 `json:"source_position_bits32"`
		Index          int      `json:"player_index"`
		Value          int      `json:"value"`
	}
	var fixture struct {
		SHA       string            `json:"original_sha256"`
		Unchanged bool              `json:"original_unchanged"`
		CW        uint16            `json:"actual_startup_control_word"`
		Faults    []json.RawMessage `json:"memory_faults"`
		Rows      []struct {
			Input struct {
				Name             string `json:"name"`
				Class            int    `json:"target_class"`
				State            int    `json:"state"`
				Direction        int    `json:"direction"`
				Power            int32  `json:"power"`
				Facing           int    `json:"facing_return"`
				Activate         int    `json:"activate_return"`
				FacingMutation   int    `json:"facing_mutation"`
				ActivateMutation int    `json:"activate_mutation"`
			} `json:"input"`
			Events    []event `json:"events"`
			Preserved bool    `json:"nonoutput_bytes_preserved"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(blob, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || !fixture.Unchanged || fixture.CW != 0xe3f || len(fixture.Rows) != 240 || len(fixture.Faults) != 0 {
		t.Fatal("incomplete original poison effect fixture")
	}
	for _, row := range fixture.Rows {
		t.Run(row.Input.Name, func(t *testing.T) {
			in := row.Input
			update, liveUpdate := modifierStatusPlayerUpdates(t)
			update.State = server.PlayerState(in.State)
			target := selfWeaponElectricNativeAlloc(t, server.Object{
				ObjClass: object.Class(in.Class), UpdateData: unsafe.Pointer(update),
				PosVec:     types.Pointf{X: -12.75, Y: math.Float32frombits(0x80000000)},
				Direction1: server.Dir16(in.Direction), NetCode: 301,
			})
			owner := selfWeaponElectricNativeAlloc(t, server.Object{
				PosVec: types.Pointf{X: 29.5, Y: -4.25}, NetCode: 269,
			})
			effect := selfWeaponElectricNativeAlloc(t, server.ModifierEff{AttackPreDmg64: server.ModifierEffFnc{Val: in.Power}})
			mutate := func(activation bool, mutation int) {
				switch mutation {
				case 1:
					target.ObjClass = 0
				case 2:
					target.ObjClass = object.ClassMonster
					if activation {
						target.ObjClass = object.ClassPlayer
					}
				case 3:
					target.UpdateData = unsafe.Pointer(liveUpdate)
				case 4:
					effect.AttackPreDmg64.Val = -129
					target.PosVec.X, owner.PosVec.Y = 77.5, 88.25
				}
			}
			got := make([]event, 0, len(row.Events))
			facing := func(pos *types.Pointf, dir int16, from *types.Pointf) int {
				if pos != &target.PosVec || from != &owner.PosVec {
					t.Fatal("poison facing lost live position pointers")
				}
				got = append(got, event{
					Service: "facing", Direction: uint32(int32(dir)),
					TargetPosition: []uint32{math.Float32bits(pos.X), math.Float32bits(pos.Y)},
					SourcePosition: []uint32{math.Float32bits(from.X), math.Float32bits(from.Y)},
				})
				mutate(false, in.FacingMutation)
				return in.Facing
			}
			activate := func(unit *server.Object, increment, maximum int32) int32 {
				if unit != target {
					t.Fatal("poison activation lost target")
				}
				got = append(got, event{Service: "activate", Increment: increment, Maximum: uint32(maximum)})
				mutate(true, in.ActivateMutation)
				return int32(in.Activate)
			}
			inform := func(index byte, value int32) {
				got = append(got, event{Service: "inform", Index: int(index), Value: int(value)})
			}
			applyModifierPoison4E0850(effect, owner, target, facing, activate, inform)
			if !reflect.DeepEqual(got, row.Events) || !row.Preserved || target.NetCode != 301 || owner.NetCode != 269 {
				t.Fatalf("native poison events %+v, original %+v", got, row.Events)
			}
		})
	}
}

// Keep the actual four-slot pre-damage entry, typed PoisonEffect CGo callback,
// poison cap, logic RNG and poison setter. No activation service is replaced.
func TestModifierPoisonActualPreDamage(t *testing.T) {
	for _, slot := range []int{0, 1, 2, 3} {
		for _, row := range []struct {
			name  string
			start uint8
			power int32
			want  uint8
		}{
			{"first", 0, 4, 1},
			{"increment", 3, 4, 4},
			{"at-cap", 4, 4, 4},
			{"above-cap", 5, 4, 5},
			{"zero-cap", 0, 0, 0},
			{"negative-cap", 0, -1, 0},
			{"full-positive-power", 0, math.MaxInt32, 1},
			{"full-negative-power", 0, math.MinInt32, 0},
		} {
			t.Run(fmt.Sprintf("slot-%d/%s", slot, row.name), func(t *testing.T) {
				srv := npcReflectLegacyServer4E17B0(t)
				health := selfWeaponElectricNativeAlloc(t, server.HealthData{Cur: 77, Max: 80, Field16: 99})
				target := selfWeaponElectricNativeAlloc(t, server.Object{ObjClass: object.ClassMonster, HealthData: health, Poison540: row.start, Field542: 1234})
				source := selfWeaponElectricNativeAlloc(t, server.Object{})
				effect := selfWeaponElectricNativeAlloc(t, server.ModifierEff{AttackPreDmg64: server.ModifierEffFnc{Fnc: itemPreDamageFunctionsNative4E13B0().poison, Val: row.power}})
				data := selfWeaponElectricNativeAlloc(t, server.ModifierInitData{})
				data.Modifiers[slot] = effect
				weapon := selfWeaponElectricNativeAlloc(t, server.Object{ObjClass: object.ClassWeapon, InitData: unsafe.Pointer(data)})
				index := srv.Rand.Logic.Index()
				if result := Nox_xxx_itemApplyPreDamageEffect_4E13B0(target, source, weapon, nil); result != 0 || target.Poison540 != row.want || srv.Rand.Logic.Index() != index+1 {
					t.Fatalf("actual pre-damage result/poison/RNG=%d/%d/%d", result, target.Poison540, srv.Rand.Logic.Index())
				}
				wantFrame, wantTimer := uint32(99), uint16(1234)
				if row.want != row.start {
					wantTimer = 1000
				}
				if row.start == 0 && row.want > 0 {
					wantFrame = 1400
				}
				if health.Cur != 77 || health.Max != 80 || health.Field16 != wantFrame || target.Field542 != wantTimer || effect.AttackPreDmg64.Val != row.power || data.Modifiers[slot] != effect {
					t.Fatal("actual poison setter, cap, frame or nonoutput state changed")
				}
			})
		}
	}
}
