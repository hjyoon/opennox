package server

import (
	"fmt"
	"math"
	"slices"
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

func damageGreatSwordFixture4E17B0(t *testing.T, player, splash bool) (
	target, source, weapon, missile, sword *Object,
	cached playerDamageGreatSwordContext4E17B0, r PlayerDamageRuntime4E17B0, events *[]string,
) {
	t.Helper()
	target, source, weapon, missile = damageFlameFixture4E17B0(t, !player, player, splash)
	sword = &Object{ObjClass: object.ClassWeapon, ObjSubClass: 0x400, ObjFlags: object.FlagEquipped,
		HealthData: &HealthData{Cur: 25, Max: 25}}
	target.InvFirstItem = sword
	if player {
		ud := target.UpdateDataPlayer()
		ud.Player.WeaponEquip = 0x400
		ud.Field76, ud.Field75, ud.Field21 = 88, 77, math.Float32bits(0.4)
		cached = playerDamageGreatSwordContext4E17B0{
			weaponFlags: ud.Player.WeaponEquip, armorFlags: ud.Player.ArmorEquip,
			state: &ud.State, marker: &ud.Field76, markerType: &ud.Field75,
		}
	} else {
		ud := target.UpdateDataMonster()
		ud.WeaponEquipFlags, ud.AIStack[0].Action = 0x400, uint32(ai.ACTION_GUARD)
		ud.Field547, ud.Field546, ud.Field1 = 88, 77, math.Float32bits(0.4)
		cached = playerDamageGreatSwordContext4E17B0{
			weaponFlags: ud.WeaponEquipFlags, armorFlags: ud.ArmorEquipFlags,
			marker: &ud.Field547, markerType: &ud.Field546,
		}
	}
	events = new([]string)
	r = damageMeleeRuntimeFixture4E17B0(t)
	r.BlockSourceExcluded = func(got *Object) bool {
		if got != missile {
			t.Fatal("wrong GreatSword exclusion object")
		}
		return false
	}
	r.BlockSourceOnlyExcluded = r.BlockSourceExcluded
	r.BlockDirection = func(got *Object, pos types.Pointf) bool {
		if got != target || pos != missile.PrevPos {
			t.Fatal("wrong GreatSword facing input")
		}
		return true
	}
	r.ProjectileReflect = func(got, defender *Object) {
		if got != missile || defender != target {
			t.Fatal("wrong GreatSword reflection")
		}
		wantMarker := uint32(1)
		if splash {
			wantMarker = 0
		}
		if *cached.marker != wantMarker || !splash && *cached.markerType != uint32(missile.TypeInd) {
			t.Fatal("reflection preceded cached marker prefix")
		}
		*events = append(*events, "reflect")
	}
	r.ClearOwner = func(got *Object) { *events = append(*events, "clear-owner"); got.ObjOwner = nil }
	r.SetOwner = func(owner, got *Object) { *events = append(*events, "set-owner"); got.ObjOwner = owner }
	r.ChangeOwner = func(*Object, *Object) { t.Fatal("GreatSword called Reflect Shield's ChangeOwner") }
	r.Audio = func(sound int, unit *Object) {
		if sound != 890 || unit != target {
			t.Fatal("wrong GreatSword sound")
		}
		*events = append(*events, "audio")
	}
	r.Melee.RandomInt = func(min, max int) int {
		if min != 18 || max != 20 {
			t.Fatal("wrong GreatSword RNG bounds")
		}
		*events = append(*events, "rng")
		return 19
	}
	r.PlayerSetState = func(unit *Object, state PlayerState) bool {
		*events = append(*events, fmt.Sprintf("state-%d", state))
		unit.UpdateDataPlayer().State = state
		return true
	}
	r.Melee.MonsterBlockAction = func(*Object) { *events = append(*events, "push-23") }
	r.Melee.MonsterPopBlockAction = func(*Object) { *events = append(*events, "pop") }
	r.BlockDamagePercent = func() float64 { *events = append(*events, "balance"); return 0.2 }
	r.Melee.CanDamageBlockWeapon = func(*Object) bool { return true }
	r.Melee.DamageBlockWeapon = func(item, owner, attacker, effective *Object, amount float32, typ object.DamageType) bool {
		if item != sword || owner != target || attacker != source || effective != missile || amount != float32(1.8) {
			t.Fatalf("wrong GreatSword wear: item=%p amount=%g", item, amount)
		}
		*events = append(*events, "wear")
		return true
	}
	r.DefaultDamage = func(_, _, _ *Object, _ int32, _ object.DamageType) bool {
		t.Fatal("reflected missile reached HP damage")
		return false
	}
	return
}

func TestPlayerDamageGreatSwordMissile4E17B0Order(t *testing.T) {
	for _, player := range []bool{false, true} {
		for _, splash := range []bool{false, true} {
			for _, retainOwner := range []bool{false, true} {
				for _, destroyed := range []bool{false, true} {
					t.Run(fmt.Sprintf("player-%t/splash-%t/owner-bit2-%t/broken-%t", player, splash, retainOwner, destroyed), func(t *testing.T) {
						target, source, weapon, missile, sword, cached, r, events := damageGreatSwordFixture4E17B0(t, player, splash)
						owner := missile.ObjOwner
						if retainOwner {
							missile.ObjSubClass = 2
							r.ClearOwner, r.SetOwner = nil, nil
						}
						wear := r.Melee.DamageBlockWeapon
						r.Melee.DamageBlockWeapon = func(i, o, s, w *Object, amount float32, typ object.DamageType) bool {
							ok := wear(i, o, s, w, amount, typ)
							if destroyed {
								i.ObjFlags |= object.FlagDestroyed
							}
							return ok
						}
						a, h, result := playerDamageGreatSwordMissileBlock4E17B0(target, source, weapon, cached, 9, object.DamageFlame, r)
						want := []string{"reflect"}
						if !retainOwner {
							want = append(want, "clear-owner", "set-owner")
							owner = target
						}
						want = append(want, "audio")
						if player {
							want = append(want, "rng", "state-19")
						} else {
							want = append(want, "push-23")
						}
						want = append(want, "balance", "wear")
						if destroyed {
							if player {
								want = append(want, "state-13")
							} else {
								want = append(want, "pop")
							}
						}
						_, _, carry := damageMeleeMarker4E17B0(target)
						if !a || !h || result || target.HealthData.Cur != 200 || sword.HealthData.Cur != 25 ||
							carry != 0.4 || missile.ObjOwner != owner || !slices.Equal(*events, want) {
							t.Fatalf("block=%t/%t/%t owner=%p carry=%g events=%v want=%v", a, h, result, missile.ObjOwner, carry, *events, want)
						}
					})
				}
			}
		}
	}
}

func TestPlayerDamageNative4E17B0NPCGreatSwordMissileTypes(t *testing.T) {
	for _, typ := range []object.DamageType{object.DamageFlame, object.DamageExplosion, object.DamageImpale,
		object.DamageImpact, object.DamageElectric, object.DamageAirborneElectric, object.DamageZapRay} {
		t.Run(fmt.Sprint(typ), func(t *testing.T) {
			target, source, weapon, _, _, _, r, events := damageGreatSwordFixture4E17B0(t, false, false)
			if h, result := PlayerDamageNative4E17B0(target, source, weapon, 9, typ, r); !h || result ||
				!slices.Equal(*events, []string{"reflect", "clear-owner", "set-owner", "audio", "push-23", "balance", "wear"}) {
				t.Fatalf("NPC block=%t/%t events=%v", h, result, *events)
			}
		})
	}
}

func TestPlayerDamageGreatSwordMissile4E17B0Admission(t *testing.T) {
	for _, player := range []bool{false, true} {
		for _, tc := range []struct {
			name   string
			change func(*Object, *Object, *PlayerDamageRuntime4E17B0)
		}{
			{"missing item", func(v, _ *Object, _ *PlayerDamageRuntime4E17B0) { v.InvFirstItem = nil }},
			{"unequipped item", func(v, _ *Object, _ *PlayerDamageRuntime4E17B0) { v.InvFirstItem.ObjFlags = 0 }},
			{"reflect", func(_, _ *Object, r *PlayerDamageRuntime4E17B0) { r.ProjectileReflect = nil }},
			{"direction", func(_, _ *Object, r *PlayerDamageRuntime4E17B0) { r.BlockDirection = nil }},
			{"ownership", func(_, _ *Object, r *PlayerDamageRuntime4E17B0) { r.SetOwner = nil }},
			{"audio", func(_, _ *Object, r *PlayerDamageRuntime4E17B0) { r.Audio = nil }},
			{"wear", func(_, _ *Object, r *PlayerDamageRuntime4E17B0) { r.Melee.DamageBlockWeapon = nil }},
			{"wear admission", func(_, _ *Object, r *PlayerDamageRuntime4E17B0) {
				r.Melee.CanDamageBlockWeapon = func(*Object) bool { return false }
			}},
		} {
			t.Run(fmt.Sprintf("player-%t/%s", player, tc.name), func(t *testing.T) {
				target, source, weapon, missile, _, cached, r, events := damageGreatSwordFixture4E17B0(t, player, false)
				tc.change(target, missile, &r)
				reason := ""
				r.Unsupported = func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = why }
				a, h, result := playerDamageGreatSwordMissileBlock4E17B0(target, source, weapon, cached, 9, object.DamageFlame, r)
				if !a || h || result || reason == "" || *cached.marker != 88 || *cached.markerType != 77 ||
					len(*events) != 0 || target.HealthData.Cur != 200 {
					t.Fatalf("admission=%t/%t/%t reason=%s events=%v", a, h, result, reason, *events)
				}
			})
		}
	}
}

func TestPlayerDamageGreatSwordMissile4E17B0LiveReloads(t *testing.T) {
	for _, mutation := range []string{"class", "subclass", "inventory"} {
		t.Run(mutation, func(t *testing.T) {
			target, source, weapon, _, sword, cached, r, events := damageGreatSwordFixture4E17B0(t, true, false)
			reflect := r.ProjectileReflect
			replacement := &Object{ObjClass: object.ClassWeapon, ObjSubClass: 0x400, ObjFlags: object.FlagEquipped}
			r.ProjectileReflect = func(m, v *Object) {
				reflect(m, v)
				if mutation == "class" {
					m.ObjClass = 0
				}
				if mutation == "subclass" {
					m.ObjSubClass = 2
				}
			}
			setState := r.PlayerSetState
			r.PlayerSetState = func(v *Object, state PlayerState) bool {
				if mutation == "inventory" {
					v.InvFirstItem = replacement
				}
				return setState(v, state)
			}
			wear := r.Melee.DamageBlockWeapon
			r.Melee.DamageBlockWeapon = func(i, o, s, w *Object, amount float32, typ object.DamageType) bool {
				if mutation == "inventory" {
					if i != replacement {
						t.Fatal("GreatSword wore eager inventory item")
					}
					*events = append(*events, "wear")
					return true
				}
				return wear(i, o, s, w, amount, typ)
			}
			a, h, result := playerDamageGreatSwordMissileBlock4E17B0(target, source, weapon, cached, 9, object.DamageFlame, r)
			transferred := slices.Contains(*events, "set-owner")
			if !a || !h || result || transferred != (mutation == "inventory") || sword.Flags().Has(object.FlagDestroyed) {
				t.Fatalf("live reload=%t/%t/%t events=%v", a, h, result, *events)
			}
		})
	}
}
