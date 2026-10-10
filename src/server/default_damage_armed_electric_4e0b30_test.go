package server

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

// These scalar services are the same declared boundaries as the independent
// unchanged PE32 replay. The root, qualifiers, metadata, effect loops and state
// are production code; actual C-owned HP is checked in legacy separately.
func armedElectricTranscript4E0B30(t *testing.T) []byte {
	t.Helper()
	subclasses := []uint32{66050, 458753, 131073, 0x11012, 0x800, 0xc00}
	damages := []int32{-9, 0, 1, 12, 5413}
	protections := []float64{0, 0.25, float64(math.Float32frombits(0x3defaf0d)), 1}
	var transcript bytes.Buffer
	run := func(si, ti, wi, typ, di, pi, mode int) {
		target, source := damageMeleeUnitFixture4E17B0(t, false), damageMeleeUnitFixture4E17B0(t, si == 0)
		target.ObjClass, target.ObjSubClass = object.Class(2097154), object.SubClass(subclasses[ti])
		target.ObjFlags = object.Flags([]uint32{151519748, 17302164, 17302164, 0, 0, 0}[ti])
		ud := target.UpdateDataMonster()
		*ud = MonsterUpdateData{Field547: 99, Field546: 77, Field523_2: 88, Field1: math.Float32bits(0.125)}
		source.ObjClass, source.ObjFlags = object.Class(2621444), object.Flags(16777732)
		if si == 1 {
			source.ObjClass, source.ObjSubClass = object.Class(2097154), 0x11012
		}
		source.PrevPos = types.Ptf(32.5, -18.25)
		weapon := &Object{TypeInd: 777, ObjClass: object.Class([]uint32{85458944, 0x1000000, 0x1000, 0x1000, 0x1000000}[wi]),
			ObjSubClass: object.SubClass([]uint32{0x1000000, 0x10, 0x10000, 0x10000, 0}[wi]), ObjFlags: object.Flags(16777488), PrevPos: types.Ptf(44, 7), InitData: unsafe.Pointer(&ModifierInitData{})}
		if wi == 0 {
			weapon.TypeInd = 223
		}
		if wi == 4 {
			weapon.ObjSubClass = 0
		}
		if wi == 2 || wi == 3 {
			wand := &WandUseData{}
			if wi == 3 {
				wand.Flags = 2
			}
			weapon.UseData = UseDataPtr{Ptr: unsafe.Pointer(wand)}
		}
		target.Buffs = 1<<defaultDamageShockEnchant4E0B30 | 1<<defaultDamageInvisibleEnchant4E0B30
		if mode == 3 {
			target.ObjFlags |= object.FlagNoUpdate
		}
		if mode == 4 {
			target.Buffs |= 1 << defaultDamageInvulnerableEnchant4E0B30
		}
		if mode == 5 {
			target.ObjFlags |= object.FlagDead
		}
		if mode == 6 || mode == 7 {
			target.Buffs |= 1 << defaultDamageShieldEnchant4E0B30
		}
		var events []int64
		resultDamage, hpCalls := uint32(0xfeedface), int64(0)
		r := damageMeleeWorldRuntime4E0B30(t)
		r.GameplayFlag1 = func() bool { events = append(events, 1); return mode != 1 }
		r.QuestMode = func() bool { return false }
		r.IsZombie = func(*Object) bool { return false }
		r.IsEnemy = func(v, a *Object) bool {
			if v != target || a != source {
				t.Fatal("armed electric enemy identity")
			}
			events = append(events, 2)
			if mode == 9 {
				weapon.ObjSubClass = 0x10
				if weapon.Class().Has(object.ClassWand) {
					weapon.ObjSubClass, weapon.UseData.AsWand().Flags = 0x10000, 0
				}
			}
			return mode != 1 && mode != 2
		}
		r.Audio = func(id int, v *Object) {
			want := target
			if id == 135 {
				want = source
			}
			if v != want {
				t.Fatal("armed electric audio identity")
			}
			events = append(events, map[int]int64{135: 5, 108: 8, 71: 19}[id])
		}
		r.BuffOff = func(v *Object, id EnchantID) {
			if v != target || (id != 0 && id != 22) {
				t.Fatal("armed electric buff identity")
			}
			v.Buffs &^= 1 << id
			if id == 22 {
				events = append(events, 4)
			} else {
				events = append(events, 9)
			}
		}
		r.BalanceFloatInd = func(key string, index int) float64 {
			if key != "ShockDamage" || index != 4 {
				t.Fatal("armed electric Shock balance")
			}
			return 80
		}
		r.CallDamage = func(v, a, w *Object, d int32, dt object.DamageType) bool {
			if v != source || a != target || w != nil || d != 80 || dt != object.DamageElectric {
				t.Fatal("armed electric Shock callback identity")
			}
			events = append(events, 3)
			return false
		}
		r.PlayerSetState = func(v *Object, state PlayerState) bool {
			if v != source || si != 0 || state != PlayerState23 {
				t.Fatal("armed electric retaliation state")
			}
			events = append(events, 6)
			v.UpdateDataPlayer().State = state
			return true
		}
		r.ElectricProtection = func(v *Object) float64 {
			if v != target {
				t.Fatal("armed electric protection identity")
			}
			events = append(events, 7)
			return protections[pi]
		}
		r.MonsterHasHitSound = func(v *Object) bool {
			if v != source || si != 1 {
				t.Fatal("armed electric monster sound identity")
			}
			events = append(events, 17)
			return false
		}
		r.DefaultDamageSound = func(v, w *Object) {
			if v != target || w != weapon {
				t.Fatal("armed electric damage sound identity")
			}
			events = append(events, 18)
		}
		r.AdjustFieldGuide = func(a, v *Object, d int32) int32 {
			if a != source || v != target {
				t.Fatal("armed electric field guide identity")
			}
			events = append(events, 14)
			return d
		}
		r.ShieldReduce = func(v *Object, d *int32, dt object.DamageType, w *Object) {
			if v != target || dt != object.DamageType(typ) || w != weapon {
				t.Fatal("armed electric Shield identity")
			}
			*d = 0
			if mode == 7 {
				*d = 5
			}
			events = append(events, 15)
		}
		r.DamageClear = func(v *Object, d int32) {
			if v != target {
				t.Fatal("armed electric HP identity")
			}
			resultDamage, hpCalls = uint32(d), hpCalls+1
			events = append(events, 16)
		}
		if mode == 8 {
			if ti != 3 {
				t.Fatal("Defend must use the original NPC subclass")
			}
			m1, m2, m3, m4 := &ModifierEff{}, &ModifierEff{}, &ModifierEff{}, &ModifierEff{}
			m1.Defend76.Fnc, m2.Defend76.Fnc = unsafe.Pointer(new(byte)), unsafe.Pointer(new(byte))
			m3.AttackPreDmg64.Fnc, m4.AttackPreDmg64.Fnc = unsafe.Pointer(new(byte)), unsafe.Pointer(new(byte))
			init := &ModifierInitData{}
			init.Modifiers[2], init.Modifiers[3] = m1, m2
			item := &Object{ObjFlags: object.FlagEquipped, InitData: unsafe.Pointer(init)}
			target.InvFirstItem = item
			(*ModifierInitData)(weapon.InitData).Modifiers[0], (*ModifierInitData)(weapon.InitData).Modifiers[3] = m3, m4
			r.CanApplyLateDefend = func(m *ModifierEff) bool { return m == m1 || m == m2 }
			r.ApplyLateDefend = func(m *ModifierEff, i, v, w, a *Object, d int32, dt object.DamageType) int32 {
				if i != item || v != target || w != weapon || a != source || dt != object.DamageType(typ) {
					t.Fatal("armed electric Defend identity")
				}
				if m == m1 {
					events = append(events, 10)
					return d + 5
				}
				events = append(events, 11)
				return d - 2
			}
			r.CanApplyPreDamage = func(m *ModifierEff) bool { return m == m3 || m == m4 }
			r.ApplyPreDamage = func(m *ModifierEff, w, a, v *Object, d *int32) {
				if w != weapon || a != source || v != target {
					t.Fatal("armed electric pre-damage identity")
				}
				if m == m3 {
					events = append(events, 12)
					*d *= 2
				} else {
					events = append(events, 13)
					*d--
				}
			}
		}
		result := DefaultDamageWorld4E0B30(target, source, weapon, damages[di], object.DamageType(typ), r)
		ret, attribution := int64(0), int64(0)
		if result {
			ret = 1
		}
		if target.Obj130 == weapon {
			attribution = 1
		}
		latch := uint32(0)
		if si == 1 {
			latch = source.UpdateDataMonster().Field130
		}
		row := []int64{int64(si), int64(ti), int64(wi), int64(typ), int64(di), int64(pi), int64(mode), ret,
			int64(ud.Field547), int64(ud.Field546), int64(ud.StatusFlags), int64(target.Buffs), attribution,
			int64(math.Float32bits(target.Pos132.X)), int64(math.Float32bits(target.Pos132.Y)), int64(target.Field131), int64(target.Frame134),
			int64(ud.Field523_2), int64(ud.Field1), int64(resultDamage), hpCalls, int64(latch)}
		row = append(row, events...)
		data, err := json.Marshal(row)
		if err != nil {
			t.Fatal(err)
		}
		transcript.Write(data)
		transcript.WriteByte('\n')
	}
	for si := range 2 {
		for ti := range 6 {
			for wi := range 5 {
				for _, typ := range []int{9, 17} {
					for di := range 5 {
						for pi := range 4 {
							run(si, ti, wi, typ, di, pi, 0)
						}
					}
				}
			}
		}
	}
	for si := range 2 {
		for _, ti := range []int{0, 3, 4, 5} {
			for _, wi := range []int{0, 3, 4} {
				for mode := 1; mode < 10; mode++ {
					if mode != 8 || ti == 3 {
						run(si, ti, wi, 9, 3, 1, mode)
					}
				}
			}
		}
	}
	return transcript.Bytes()
}

func TestDefaultDamageArmedElectric4E0B30OriginalTranscript(t *testing.T) {
	got := armedElectricTranscript4E0B30(t)
	const want = "b6f3a51bed04c19e992cc6053fc36ced61281d2c5c49b4ac4b558c7a5285355f"
	if sum := fmt.Sprintf("%x", sha256.Sum256(got)); sum != want {
		t.Fatalf("independent original electric transcript bytes=%d SHA=%s want=%s", len(got), sum, want)
	}
	t.Logf("unchanged PE32 armed electric transcript rows=%d bytes=%d SHA=%s", bytes.Count(got, []byte{'\n'}), len(got), want)
}

func TestDefaultDamageArmedElectric4E0B30AdmissionAndServices(t *testing.T) {
	fixture := func(t *testing.T) (*Object, *Object, *Object, DefaultDamageWorldRuntime4E0B30) {
		target, source := damageMeleeUnitFixture4E17B0(t, false), damageMeleeUnitFixture4E17B0(t, true)
		target.ObjSubClass = 0
		weapon := &Object{ObjClass: object.ClassWeapon, ObjSubClass: 0x1000000, InitData: unsafe.Pointer(&ModifierInitData{})}
		r := damageMeleeWorldRuntime4E0B30(t)
		r.ElectricProtection = func(*Object) float64 { return 0 }
		return target, source, weapon, r
	}
	for _, shape := range []string{"non-unit-source", "weapon-source", "nil-source-update", "missile-weapon", "monster-weapon", "player-wand", "world-weapon", "poison-type"} {
		t.Run("reject-"+shape, func(t *testing.T) {
			v, a, w, r := fixture(t)
			typ := object.DamageElectric
			switch shape {
			case "non-unit-source":
				a.ObjClass = object.ClassSimple
			case "weapon-source":
				a.ObjClass |= object.ClassWeapon
			case "nil-source-update":
				a.UpdateData = nil
			case "missile-weapon":
				w.ObjClass |= object.ClassMissile
			case "monster-weapon":
				w.ObjClass |= object.ClassMonster
			case "player-wand":
				w.ObjClass = object.ClassPlayer | object.ClassWand
				w.ObjSubClass = 0
			case "world-weapon":
				w.ObjClass = object.ClassSimple
			case "poison-type":
				typ = object.DamagePoison
			}
			rejected := false
			r.Unsupported = func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) {
				rejected = reason == "unsupported monster damage shape"
			}
			if !DefaultDamageWorld4E0B30(v, a, w, 12, typ, r) || !rejected || v.HealthData.Cur != 200 || v.Obj130 != nil {
				t.Fatal("armed admission swallowed an unrelated shape")
			}
		})
	}
	for _, service := range []string{"enemy", "protection", "buff-off", "HP", "monster-sound"} {
		t.Run("missing-"+service, func(t *testing.T) {
			v, a, w, r := fixture(t)
			switch service {
			case "enemy":
				r.IsEnemy = nil
			case "protection":
				r.ElectricProtection = nil
			case "buff-off":
				r.BuffOff = nil
			case "HP":
				r.DamageClear = nil
			case "monster-sound":
				a = damageMeleeUnitFixture4E17B0(t, false)
				r.MonsterHasHitSound = nil
			}
			reason := ""
			r.Unsupported = func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = why }
			if !DefaultDamageWorld4E0B30(v, a, w, 12, object.DamageElectric, r) || reason == "" || v.HealthData.Cur != 200 || v.Obj130 != nil {
				t.Fatal("missing required service must remain visible before HP")
			}
		})
	}
	t.Run("immune-before-tail-services-and-nil-health", func(t *testing.T) {
		v, a, w, r := fixture(t)
		v.ObjSubClass, v.HealthData = 0x800, nil
		r.BuffOff, r.DamageClear, r.ElectricProtection, r.MonsterHasHitSound = nil, nil, nil, nil
		if !DefaultDamageWorld4E0B30(v, a, w, 12, object.DamageElectric, r) || v.Obj130 != nil || v.Frame134 != 0 {
			t.Fatal("electric immunity requires no ordinary tail")
		}
	})
	t.Run("nil-health-reaches-ordinary-HP-boundary", func(t *testing.T) {
		v, a, w, r := fixture(t)
		v.HealthData = nil
		calls := 0
		r.DamageClear = func(target *Object, damage int32) {
			if target != v || damage != 12 {
				t.Fatal("nil-health HP service identity")
			}
			calls++
		}
		if !DefaultDamageWorld4E0B30(v, a, w, 12, object.DamageElectric, r) || calls != 1 || v.Obj130 != w || v.Frame134 != 1400 {
			t.Fatal("nil health must be handled by UnitDamageClear after attribution")
		}
	})
}

func TestDefaultDamageArmedElectric4E0B30LiveImmunityAndFrame(t *testing.T) {
	for _, immune := range []bool{true, false} {
		t.Run(fmt.Sprintf("Shock-installs-immunity-%t", immune), func(t *testing.T) {
			v, a := damageMeleeUnitFixture4E17B0(t, false), damageMeleeUnitFixture4E17B0(t, true)
			w := &Object{TypeInd: 223, ObjClass: object.ClassWeapon, ObjSubClass: 0x1000000, InitData: unsafe.Pointer(&ModifierInitData{})}
			v.Buffs = 1<<defaultDamageShockEnchant4E0B30 | 1<<defaultDamageInvisibleEnchant4E0B30
			if !immune {
				v.ObjSubClass |= 0x800
			}
			r := damageMeleeWorldRuntime4E0B30(t)
			frame, protections, shock, state := uint32(1400), 0, 0, 0
			r.Frame = func() uint32 { return frame }
			r.Audio = func(id int, target *Object) {
				if id != 135 || target != a {
					t.Fatal("frame must be re-read after protection; no stale resistance audio")
				}
			}
			r.BuffOff = func(target *Object, id EnchantID) {
				if target != v {
					t.Fatal("live BuffOff identity")
				}
				target.Buffs &^= 1 << id
				if id == defaultDamageInvisibleEnchant4E0B30 {
					frame = 1504
				}
			}
			r.BalanceFloatInd = func(string, int) float64 { return 80 }
			r.CallDamage = func(target, source, weapon *Object, damage int32, typ object.DamageType) bool {
				if target != a || source != v || weapon != nil || damage != 80 || typ != object.DamageElectric {
					t.Fatal("live Shock identity")
				}
				shock++
				v.ObjSubClass &^= 0x800
				if immune {
					v.ObjSubClass |= 0x800
				}
				frame = 1404
				return false
			}
			r.PlayerSetState = func(target *Object, value PlayerState) bool {
				if target != a || value != PlayerState23 {
					t.Fatal("live Shock player state")
				}
				state++
				frame = 1440
				return true
			}
			r.ElectricProtection = func(target *Object) float64 {
				if target != v {
					t.Fatal("live protection identity")
				}
				protections++
				frame = 1441
				return 0.25
			}
			if !DefaultDamageWorld4E0B30(v, a, w, 25, object.DamageElectric, r) || shock != 1 || state != 1 || v.HasEnchant(defaultDamageShockEnchant4E0B30) {
				t.Fatal("Shock must finish before the live immunity reload")
			}
			if immune {
				if protections != 0 || v.HealthData.Cur != 200 || v.Obj130 != nil || v.Frame134 != 0 || !v.HasEnchant(defaultDamageInvisibleEnchant4E0B30) {
					t.Fatal("new immunity must stop the ordinary tail")
				}
			} else if protections != 1 || v.HealthData.Cur != 181 || v.Obj130 != w || v.Frame134 != 1504 || v.HasEnchant(defaultDamageInvisibleEnchant4E0B30) || v.UpdateDataMonster().Field523_2 != 2 {
				t.Fatal("removed immunity must use live protection, BuffOff and timestamp")
			}
		})
	}
}
