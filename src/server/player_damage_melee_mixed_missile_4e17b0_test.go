package server

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

// Original direct callback probes retain qualifying MISSILE|WEAPON forms and six source identities.
// Signed/default values do not invoke the public root DoDamage wrapper.
func TestPlayerDamageMeleeMixedMissileMatchesGAMEEXE4E17B0(t *testing.T) {
	data, err := os.ReadFile("testdata/player_damage_melee_mixed_missile_4e17b0.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		OriginalSHA string `json:"original_sha256"`
		Rows        []struct {
			Player         bool              `json:"player"`
			SourceKind     string            `json:"source_kind"`
			WeaponKind     string            `json:"weapon_kind"`
			WeaponSubclass object.SubClass   `json:"weapon_subclass"`
			Type           object.DamageType `json:"damage_type"`
			Raw            int32             `json:"raw"`
			Forward        uint32            `json:"forward_damage"`
			Marker         uint32            `json:"marker"`
			Kind           uint32            `json:"marker_type"`
			Carry          uint32            `json:"carry_bits"`
			Wear           []uint32          `json:"wear"`
			Return         int               `json:"return"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.OriginalSHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || len(fixture.Rows) != 2304 {
		t.Fatal("unexpected original initialized-monster source fixture")
	}
	for _, row := range fixture.Rows {
		t.Run(fmt.Sprintf("source-%s/weapon-%s/subclass-%x/player-%t/type-%d/raw-%d", row.SourceKind, row.WeaponKind, uint32(row.WeaponSubclass), row.Player, row.Type, row.Raw), func(t *testing.T) {
			v := damageMeleeUnitFixture4E17B0(t, row.Player)
			armor := damageMeleeArmorFixture4E17B0(v, .5, .25)
			attacker := damageMeleeUnitFixture4E17B0(t, true)
			weapon := damageMeleeUnitFixture4E17B0(t, false)
			attacker.UpdateDataPlayer().Player.PlayerUnit = attacker
			if row.SourceKind == "monster" {
				attacker = damageMeleeUnitFixture4E17B0(t, false)
			}
			if row.SourceKind == "world" || row.SourceKind == "zero" {
				attacker = damageMeleeUnitFixture4E17B0(t, false)
				attacker.ObjClass, attacker.ObjSubClass, attacker.UpdateData = object.ClassDoor, 0, nil
				if row.SourceKind == "zero" {
					attacker.ObjClass = 0
				}
			}
			if row.WeaponKind == "world" {
				weapon.ObjClass, weapon.ObjSubClass, weapon.UpdateData = object.ClassArmor, 0, nil
			}
			weapon.ObjClass, weapon.ObjSubClass, weapon.UpdateData, weapon.HealthData = object.ClassMissile|object.ClassWeapon, row.WeaponSubclass, nil, nil
			modifiers, free := alloc.New(ModifierInitData{})
			*modifiers = ModifierInitData{}
			t.Cleanup(free)
			weapon.InitData = unsafe.Pointer(modifiers)
			weapon.TypeInd, weapon.ObjOwner = 73, attacker
			if row.SourceKind == "terminal" {
				attacker = weapon
				weapon.ObjOwner = nil
			}
			if row.SourceKind == "none" {
				attacker = nil
				weapon.ObjOwner = nil
			}
			if row.SourceKind != "none" && weapon.FindOwnerChainPlayer() != attacker {
				t.Fatal("wrong terminal owner")
			}
			if row.Player {
				v.UpdateDataPlayer().Field76, v.UpdateDataPlayer().Field75 = 99, 77
			} else {
				v.UpdateDataMonster().Field547, v.UpdateDataMonster().Field546 = 99, 77
			}
			r := damageMeleeRuntimeFixture4E17B0(t)
			damageMeleeArmorRuntime4E17B0(&r, armor, .5)
			r.CoopMode = func() bool { return false }
			r.IsEnemy = func(*Object, *Object) bool { return true }
			r.QuestMode = func() bool { return false }
			r.GodMode = func() bool { return false }
			r.GameplayFlag1 = func() bool { return true }
			r.FireProtection = func(*Object) float64 { return 0 }
			r.ElectricArmorScale = func(*Object) float32 { return 1 }
			r.BuffOff = func(*Object, EnchantID) {}
			r.PlayerSetState = func(*Object, PlayerState) bool { return true }
			calls, forward, wear := 0, int32(123456), int32(0)
			r.DefaultDamage = func(target, a, w *Object, d int32, k object.DamageType) bool {
				if target != v || a != attacker || w != weapon || k != row.Type {
					t.Fatal("owned-monster DefaultDamage arguments")
				}
				calls++
				forward = d
				return true
			}
			r.DamageClear = func(target *Object, d int32) {
				if target != v {
					t.Fatal("wrong HP target")
				}
				calls++
				forward = d
			}
			r.DamageArmor = func(item, a, w *Object, d int32, k object.DamageType) bool {
				if item != armor || a != attacker || w != weapon || k != row.Type {
					t.Fatal("owned-monster armor arguments")
				}
				wear += d
				return true
			}
			reason := ""
			r.Unsupported = func(s string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = s }
			h, ok := PlayerDamageNative4E17B0(v, attacker, weapon, row.Raw, row.Type, r)
			marker, kind, carry := damageMeleeMarker4E17B0(v)
			wantWear := int32(0)
			if len(row.Wear) != 0 {
				wantWear = max(int32(row.Wear[0]), 0)
			}
			if !h || ok != (row.Return != 0) || calls != 1 || uint32(forward) != row.Forward || marker != row.Marker || kind != row.Kind || math.Float32bits(carry) != row.Carry || wear != wantWear || reason != "" {
				t.Fatalf("handled/result=%t/%t forward=%d/%d calls=%d marker=%d/%d carry=%g/%g wear=%d/%d reason=%s", h, ok, forward, int32(row.Forward), calls, marker, kind, carry, math.Float32frombits(row.Carry), wear, wantWear, reason)
			}
		})
	}
}
