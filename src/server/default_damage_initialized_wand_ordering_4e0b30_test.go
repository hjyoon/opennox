package server

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

// The fixture executes unchanged DefaultDamage, its melee and Hammer predicates,
// and reached original item-pre-damage. Shock and HP calls are explicit services.
func TestDefaultDamageInitializedWandOriginalOrdering4E0B30(t *testing.T) {
	blob, err := os.ReadFile("testdata/default_damage_initialized_wand_ordering_4e0b30.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		SHA  string `json:"original_sha256"`
		Rows []struct {
			Source         string               `json:"source_kind"`
			WeaponSubclass object.SubClass      `json:"weapon_subclass"`
			WandFlags      uint32               `json:"wand_flags"`
			Player         bool                 `json:"player"`
			Type           object.DamageType    `json:"damage_type"`
			Subclass       object.SubClass      `json:"subclass"`
			Enemy          bool                 `json:"enemy"`
			NoUpdate       bool                 `json:"no_update"`
			Shock          bool                 `json:"shock"`
			Retaliation    []int32              `json:"retaliation"`
			Events         []string             `json:"events"`
			Forward        []int32              `json:"forward_damage"`
			Marker         uint32               `json:"marker"`
			Kind           uint32               `json:"marker_type"`
			Status         object.MonsterStatus `json:"status"`
			SourceState    PlayerState          `json:"source_state"`
			Frame          uint32               `json:"frame"`
			Position       []float32            `json:"position"`
			Return         int                  `json:"return"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(blob, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || len(fixture.Rows) != 768 {
		t.Fatal("original ordering fixture")
	}
	for _, row := range fixture.Rows {
		t.Run(fmt.Sprintf("source-%s/subclass-%x/flags-%x/player-%t/type-%d/enemy-%t/no-update-%t/shock-%t", row.Source, uint32(row.WeaponSubclass), row.WandFlags, row.Player, row.Type, row.Enemy, row.NoUpdate, row.Shock), func(t *testing.T) {
			v := damageMeleeUnitFixture4E17B0(t, row.Player)
			v.ObjSubClass = row.Subclass
			if row.NoUpdate {
				v.ObjFlags |= object.FlagNoUpdate
			}
			if row.Shock {
				v.Buffs |= 1 << defaultDamageShockEnchant4E0B30
			}
			if row.Player {
				u := v.UpdateDataPlayer()
				u.Field76, u.Field75 = 99, 77
			} else {
				u := v.UpdateDataMonster()
				u.Field547, u.Field546 = 99, 77
			}
			a := damageMeleeUnitFixture4E17B0(t, row.Source == "player")
			if row.Source == "player" {
				a.UpdateDataPlayer().State = 0
				a.UpdateDataPlayer().Player.PlayerUnit = a
			}
			if row.Source == "world" || row.Source == "zero" {
				a.ObjClass, a.UpdateData = object.ClassDoor, nil
				if row.Source == "zero" {
					a.ObjClass = 0
				}
			}
			a.PrevPos = types.Ptf(-33, -44)
			mods, free := alloc.New(ModifierInitData{})
			*mods = ModifierInitData{}
			t.Cleanup(free)
			w := &Object{ObjClass: object.ClassWand, ObjSubClass: row.WeaponSubclass, TypeInd: 73, InitData: unsafe.Pointer(mods), ObjOwner: a, PrevPos: types.Ptf(32.5, -18.25)}
			wandUse, freeWand := alloc.New(WandUseData{})
			*wandUse = WandUseData{Flags: row.WandFlags}
			t.Cleanup(freeWand)
			w.UseData = UseDataPtr{Ptr: unsafe.Pointer(wandUse)}
			if row.Source == "terminal" {
				a = w
				w.ObjOwner = nil
			}
			if row.Source == "none" {
				a = nil
				w.ObjOwner = nil
			}
			var events, wantEvents []string
			for _, e := range row.Events {
				if !strings.HasPrefix(e, "mode-") && e != "late-defend" && e != "ball-boundary" {
					wantEvents = append(wantEvents, e)
				}
			}
			r := damageMeleeWorldRuntime4E0B30(t)
			r.IsEnemy = func(victim, source *Object) bool {
				if victim != v || source != a {
					t.Fatal("enemy identity")
				}
				events = append(events, "enemy")
				return row.Enemy
			}
			r.Audio = func(id int, unit *Object) {
				if id != 135 || unit != a {
					t.Fatal("Shock audio identity")
				}
				events = append(events, "shock-audio")
			}
			r.BuffOff = func(unit *Object, id EnchantID) {
				if unit != v || (id != 0 && id != 22) {
					t.Fatal("buff identity")
				}
				unit.Buffs &^= 1 << id
				events = append(events, fmt.Sprintf("buff-off-%d", id))
			}
			r.BalanceFloatInd = func(key string, index int) float64 {
				if key != "ShockDamage" || index != 4 {
					t.Fatal("Shock balance")
				}
				events = append(events, "shock-balance9")
				return 9
			}
			var retaliation, forward []int32
			r.CallDamage = func(victim, source, weapon *Object, damage int32, typ object.DamageType) bool {
				if victim != a || source != v || weapon != nil || typ != object.DamageElectric {
					t.Fatal("Shock registered callback")
				}
				retaliation = append(retaliation, damage)
				events = append(events, "retaliation")
				return true
			}
			r.PlayerSetState = func(unit *Object, state PlayerState) bool {
				if unit != a || state != PlayerState23 {
					t.Fatal("Shock source state")
				}
				unit.UpdateDataPlayer().State = state
				events = append(events, "source-state23")
				return true
			}
			r.ElectricProtection = func(unit *Object) float64 {
				if unit != v {
					t.Fatal("electric protection identity")
				}
				events = append(events, "zero-electric-protection")
				return 0
			}
			r.DefaultDamageSound = func(unit, weapon *Object) {
				if unit != v || weapon != w {
					t.Fatal("default sound identity")
				}
				events = append(events, "default-sound")
			}
			r.PlayerDamageSound = func(unit, weapon *Object) {
				if unit != v || weapon != w {
					t.Fatal("player sound identity")
				}
				events = append(events, "player-sound")
			}
			r.DamageClear = func(unit *Object, damage int32) {
				if unit != v {
					t.Fatal("HP identity")
				}
				forward = append(forward, damage)
				events = append(events, "HP-boundary")
				unit.HealthData.Cur -= uint16(damage)
			}
			result := DefaultDamageWorld4E0B30(v, a, w, 9, row.Type, r)
			marker, kind := uint32(0), uint32(0)
			status := object.MonsterStatus(0)
			if row.Player {
				u := v.UpdateDataPlayer()
				marker, kind = u.Field76, u.Field75
			} else {
				u := v.UpdateDataMonster()
				marker, kind, status = u.Field547, u.Field546, u.StatusFlags
			}
			equalNumbers := func(a, b []int32) bool {
				if len(a) != len(b) {
					return false
				}
				for i := range a {
					if a[i] != b[i] {
						return false
					}
				}
				return true
			}
			if result != (row.Return != 0) || !reflect.DeepEqual(events, wantEvents) || !equalNumbers(forward, row.Forward) || !equalNumbers(retaliation, row.Retaliation) || marker != row.Marker || kind != row.Kind || status != row.Status || v.Frame134 != row.Frame || len(row.Position) != 2 || v.Pos132 != types.Ptf(row.Position[0], row.Position[1]) {
				t.Fatalf("result=%t/%d events=%v/%v HP=%v/%v retaliation=%v/%v marker=%d/%d kind=%d/%d status=%x/%x frame=%d/%d pos=%v/%v", result, row.Return, events, wantEvents, forward, row.Forward, retaliation, row.Retaliation, marker, row.Marker, kind, row.Kind, status, row.Status, v.Frame134, row.Frame, v.Pos132, row.Position)
			}
			if row.Source == "player" && a.UpdateDataPlayer().State != row.SourceState {
				t.Fatal("source state")
			}
			if len(forward) > 0 {
				if v.Obj130 != w || v.Field131 != uint32(row.Type) || v.HealthData.Cur != 191 {
					t.Fatal("actual weapon/HP attribution")
				}
			} else if v.Obj130 != nil || v.HealthData.Cur != 200 {
				t.Fatal("early-exit attribution/HP")
			}
		})
	}
}
