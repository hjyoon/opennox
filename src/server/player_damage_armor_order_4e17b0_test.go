package server

import (
	"fmt"
	"math"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

// Exercise the 004E17B0 callers, not just the standalone 004E2180 helper.
// Its four switch calls at 004E1DE0/004E1E1E/004E1F25/004E1FBE must follow
// the hit-prefix/carry stores and complete each item's callbacks before
// reading the next inventory link. HP/report callbacks below deliberately
// replace records so an eager durability plan cannot satisfy this test.
func TestPlayerDamageNative4E17B0OrderedArmorCallers(t *testing.T) {
	for _, tc := range []struct {
		name   string
		player bool
		typ    object.DamageType
	}{
		{"player PIERCE", true, object.DamageImpale},
		{"NPC PIERCE", false, object.DamageImpale},
		{"player ELECTRIC", true, object.DamageElectric},
		{"NPC ELECTRIC", false, object.DamageElectric},
		{"player BLADE", true, object.DamageBlade},
		{"NPC BLADE", false, object.DamageBlade},
		{"player LAVA", true, object.DamageLava},
		{"player BITE", true, object.DamageBite},
		{"player IMPACT", true, object.DamageImpact},
		{"player charge", true, object.DamageCrush},
	} {
		t.Run(tc.name, func(t *testing.T) {
			target := damageMeleeUnitFixture4E17B0(t, tc.player)
			source := damageMeleeUnitFixture4E17B0(t, tc.typ == object.DamageCrush)
			weapon := &Object{TypeInd: 701, ObjClass: object.ClassMissile}
			marker, markerType := uint32(1), uint32(701)
			residual, portion := float32(-0.25), float32(0.5)
			switch tc.typ {
			case object.DamageElectric:
				weapon, marker, markerType, portion = source, 0, 77, 1.5
			case object.DamageBlade:
				weapon.ObjClass = object.ClassWeapon
			case object.DamageLava:
				source, weapon, marker, markerType, residual, portion = nil, nil, 0, 77, 0.25, 1.5
			case object.DamageBite:
				weapon, marker, markerType = source, 0, 77
			case object.DamageCrush:
				weapon, marker, markerType, residual, portion = source, 0, 77, 0.5, 0.5
			}
			setUnit := func(armor float32) {
				if tc.player {
					ud := target.UpdateDataPlayer()
					ud.Field57 = math.Float32bits(armor)
				} else {
					ud := target.UpdateDataMonster()
					ud.Field518 = math.Float32bits(armor)
				}
			}
			setUnit(0.5)
			if tc.player {
				target.UpdateDataPlayer().Field76, target.UpdateDataPlayer().Field75 = 99, 77
				target.UpdateDataPlayer().Field21 = math.Float32bits(0.25)
			} else {
				target.UpdateDataMonster().Field547, target.UpdateDataMonster().Field546 = 99, 77
				target.UpdateDataMonster().Field1 = math.Float32bits(0.25)
			}
			newItem := func(ind uint16) *Object {
				return &Object{TypeInd: ind, ObjClass: object.ClassArmor, ObjFlags: object.FlagEquipped,
					Damage: unsafe.Pointer(new(byte)), HealthData: &HealthData{Cur: 20, Max: 20},
					InitData: unsafe.Pointer(&ModifierInitData{}), UpdateData: unsafe.Pointer(&WeaponArmorUpdateData{})}
			}
			first, stale, next := newItem(81), newItem(82), newItem(83)
			first.InvNextItem, target.InvFirstItem = stale, first
			modifier := &ModifierEff{Defend76: ModifierEffFnc{Fnc: unsafe.Pointer(new(byte))}}
			first.InitDataModifier().Modifiers[1] = modifier
			oldCarry := first.UpdateDataWeaponArmor()
			liveCarry := &WeaponArmorUpdateData{Field0: math.Float32bits(0.75)}
			liveHP := &HealthData{Cur: 9, Max: 20}
			var hpDamages []int32
			r := playerDamageRuntime4E17B0(t, nil, &hpDamages)
			r.GameplayFlag1 = func() bool { return true }
			r.ElectricArmorScale = func(*Object) float32 { return 0.5 }
			r.CanApplyArmorDefend = func(m *ModifierEff) bool { return m == modifier }
			r.CanDamageArmor = func(item *Object) bool { return item.Damage != nil }
			var events []string
			r.ItemArmorValue = func(item *Object) float32 {
				gotMarker, gotType, gotCarry := damageMeleeMarker4E17B0(target)
				if gotMarker != marker || gotType != markerType || gotCarry != residual {
					t.Fatalf("lookup preceded prefix/carry: %d/%d/%g, want %d/%d/%g", gotMarker, gotType, gotCarry, marker, markerType, residual)
				}
				if item == first {
					events = append(events, "lookup-first")
					setUnit(2) // Helper denominator must already be captured as 0.5.
				} else if item == next {
					events = append(events, "lookup-next")
				} else {
					t.Fatal("eager inventory snapshot visited the detached item")
				}
				return 0.25
			}
			r.ApplyArmorDefend = func(m *ModifierEff, item, owner, effective, attacker *Object, amount *float32) bool {
				if m != modifier || item != first || owner != target || effective != weapon || attacker != source || *amount != portion {
					t.Fatalf("defend argument order/portion: %g, want %g", *amount, portion)
				}
				events = append(events, "defend-first")
				first.UpdateData = unsafe.Pointer(liveCarry)
				return true
			}
			r.DamageArmor = func(item, attacker, effective *Object, damage int32, typ object.DamageType) bool {
				if attacker != source || effective != weapon || typ != tc.typ {
					t.Fatal("wear lost source/effective/raw type")
				}
				if item == first {
					if damage != itemDurabilityRound4E1560(portion+0.75) || oldCarry.Field0 != 0 {
						t.Fatal("wear used stale update data")
					}
					events = append(events, "damage-first")
					first.HealthData = liveHP
					if !tc.player {
						first.InvNextItem = next
					}
				} else if item == next {
					events = append(events, "damage-next")
					item.HealthData.Cur -= uint16(damage)
				} else {
					t.Fatal("wear visited detached item")
				}
				return false // EquipDamage ignores the item damage callback's result.
			}
			r.ReportArmorHealth = func(owner, item *Object, before, after uint16) {
				if !tc.player || owner != target || before != 20 {
					t.Fatal("NPC report or wrong owner/old HP")
				}
				if item == first {
					if after != 9 {
						t.Fatal("report read the replaced HP record")
					}
					events = append(events, "report-first")
					first.InvNextItem = next // Load InvNext after the report too.
				} else {
					events = append(events, "report-next")
				}
			}
			r.DefaultDamage = func(_, _, _ *Object, _ int32, _ object.DamageType) bool { return true }
			if h, result := PlayerDamageNative4E17B0(target, source, weapon, 3, tc.typ, r); !h || !result {
				t.Fatalf("ordered armor call failed: %t/%t", h, result)
			}
			want := []string{"lookup-first", "defend-first", "damage-first"}
			if tc.player {
				want = append(want, "report-first")
			}
			want = append(want, "lookup-next")
			if itemDurabilityRound4E1560(portion) > 0 {
				want = append(want, "damage-next")
				if tc.player {
					want = append(want, "report-next")
				}
			}
			if !slices.Equal(events, want) || stale.HealthData.Cur != 20 || first.HealthData != liveHP {
				t.Fatalf("events=%v, want %v", events, want)
			}
			wantMarker, wantType := marker, markerType
			if marker == 0 {
				wantMarker, wantType = 2, uint32(tc.typ)
			}
			if gotMarker, gotType, _ := damageMeleeMarker4E17B0(target); gotMarker != wantMarker || gotType != wantType {
				t.Fatalf("tail marker=%d/%d, want %d/%d", gotMarker, gotType, wantMarker, wantType)
			}
		})
	}
}

func TestPlayerDamageNative4E17B0ZeroSignedAndNoHealthArmor(t *testing.T) {
	for _, player := range []bool{false, true} {
		for _, tc := range []struct {
			name                    string
			damage, hp, wear        int32
			armor, carry, wantCarry float32
			portion, hpCarry        float32
			noHealth                bool
		}{
			{name: "zero flushes carry", armor: 0.5, carry: 0.75, wear: 1, wantCarry: -0.25},
			{name: "negative portion updates carry", damage: -3, hp: -2, armor: 0.5, carry: 0.1, portion: -0.5, wantCarry: -0.4, hpCarry: 0.5},
			{name: "zero denominator preserves NaN", carry: 0.75, portion: float32(math.NaN()), wantCarry: float32(math.NaN())},
			{name: "no health still looks up armor", damage: 3, hp: 2, armor: 0.5, portion: 0.5, hpCarry: -0.5, noHealth: true},
		} {
			t.Run(fmt.Sprintf("player-%t/%s", player, tc.name), func(t *testing.T) {
				target := damageMeleeUnitFixture4E17B0(t, player)
				source := damageMeleeUnitFixture4E17B0(t, !player)
				weapon := &Object{TypeInd: 701, ObjClass: object.ClassWeapon}
				item := damageMeleeArmorFixture4E17B0(target, tc.armor, 0)
				itemUD := &WeaponArmorUpdateData{Field0: math.Float32bits(tc.carry)}
				item.UpdateData = unsafe.Pointer(itemUD)
				modifier := &ModifierEff{Defend76: ModifierEffFnc{Fnc: unsafe.Pointer(new(byte))}}
				item.InitDataModifier().Modifiers[1] = modifier
				if tc.noHealth {
					item.HealthData, item.InitData, item.UpdateData, item.Damage = nil, nil, nil, nil
				}
				r := damageMeleeRuntimeFixture4E17B0(t)
				var events []string
				r.ItemArmorValue = func(got *Object) float32 {
					marker, markerType, carry := damageMeleeMarker4E17B0(target)
					if got != item || marker != 1 || markerType != 701 || carry != tc.hpCarry {
						t.Fatal("zero/signed/no-health lookup preceded the native prefix")
					}
					events = append(events, "lookup")
					return 0.25
				}
				r.CanApplyArmorDefend = func(m *ModifierEff) bool { return m == modifier }
				r.ApplyArmorDefend = func(m *ModifierEff, armor, owner, effective, attacker *Object, amount *float32) bool {
					if tc.noHealth || m != modifier || armor != item || owner != target || effective != weapon || attacker != source ||
						!(*amount == tc.portion || (math.IsNaN(float64(*amount)) && math.IsNaN(float64(tc.portion)))) {
						t.Fatalf("zero/signed portion=%g, want %g", *amount, tc.portion)
					}
					events = append(events, "defend")
					return true
				}
				r.CanDamageArmor = func(got *Object) bool { return got == item }
				r.DamageArmor = func(got, attacker, effective *Object, wear int32, typ object.DamageType) bool {
					if got != item || attacker != source || effective != weapon || wear != tc.wear || typ != object.DamageBlade {
						t.Fatal("zero/signed wear arguments")
					}
					events = append(events, "wear")
					item.HealthData.Cur -= uint16(wear)
					return false
				}
				r.ReportArmorHealth = func(*Object, *Object, uint16, uint16) {
					if !player {
						t.Fatal("NPC item-health report")
					}
					events = append(events, "report")
				}
				r.DefaultDamage = func(got, attacker, effective *Object, hp int32, typ object.DamageType) bool {
					if got != target || attacker != source || effective != weapon || hp != tc.hp || typ != object.DamageBlade {
						t.Fatalf("signed HP damage=%d, want %d", hp, tc.hp)
					}
					events = append(events, "default")
					return true
				}
				if h, result := PlayerDamageNative4E17B0(target, source, weapon, tc.damage, object.DamageBlade, r); !h || !result {
					t.Fatalf("zero/signed call=%t/%t", h, result)
				}
				want := []string{"lookup"}
				if !tc.noHealth {
					want = append(want, "defend")
				}
				if tc.wear > 0 {
					want = append(want, "wear")
					if player {
						want = append(want, "report")
					}
				}
				want = append(want, "default")
				carry := math.Float32frombits(itemUD.Field0)
				if !slices.Equal(events, want) ||
					!(carry == tc.wantCarry || (math.IsNaN(float64(carry)) && math.IsNaN(float64(tc.wantCarry)))) ||
					(!tc.noHealth && item.HealthData.Cur != uint16(25-tc.wear)) {
					t.Fatalf("events=%v carry=%g, want %v/%g", events, carry, want, tc.wantCarry)
				}
			})
		}
	}
}

func TestPlayerDamageNative4E17B0LiveArmorCallbackFailsClosed(t *testing.T) {
	target := damageMeleeUnitFixture4E17B0(t, true)
	source := damageMeleeUnitFixture4E17B0(t, false)
	weapon := &Object{TypeInd: 701, ObjClass: object.ClassWeapon}
	item := damageMeleeArmorFixture4E17B0(target, 0.5, 0)
	knownDamage := item.Damage
	modifier := &ModifierEff{Defend76: ModifierEffFnc{Fnc: unsafe.Pointer(new(byte))}}
	item.InitDataModifier().Modifiers[1] = modifier
	r := damageMeleeRuntimeFixture4E17B0(t)
	var events []string
	r.ItemArmorValue = func(*Object) float32 { events = append(events, "lookup"); return 0.5 }
	r.CanApplyArmorDefend = func(m *ModifierEff) bool { return m == modifier }
	r.CanDamageArmor = func(got *Object) bool { return got.Damage == knownDamage }
	r.ApplyArmorDefend = func(*ModifierEff, *Object, *Object, *Object, *Object, *float32) bool {
		events = append(events, "defend")
		item.Damage = unsafe.Pointer(new(byte)) // Unknown live callback, not PE32 fallback authority.
		return true
	}
	r.DamageArmor = func(*Object, *Object, *Object, int32, object.DamageType) bool {
		t.Fatal("unknown live item entered a damage callback")
		return false
	}
	r.Unsupported = func(reason string, owner, attacker, effective *Object, remaining int32, typ object.DamageType) {
		if reason != "unsupported live armor damage callback" || owner != target || attacker != source || effective != weapon || remaining != 1 || typ != object.DamageBlade {
			t.Fatal("live rejection context")
		}
		events = append(events, "reject")
	}
	r.DefaultDamage = func(*Object, *Object, *Object, int32, object.DamageType) bool {
		events = append(events, "default")
		return true
	}
	if h, result := PlayerDamageNative4E17B0(target, source, weapon, 3, object.DamageBlade, r); !h || !result ||
		!slices.Equal(events, []string{"lookup", "defend", "reject", "default"}) || item.HealthData.Cur != 25 || math.Abs(float64(math.Float32frombits(item.UpdateDataWeaponArmor().Field0)-0.4)) > 1e-6 {
		t.Fatalf("live rejection=%t/%t events=%v HP=%d", h, result, events, item.HealthData.Cur)
	}
}
