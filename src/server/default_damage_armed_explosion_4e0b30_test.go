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
func armedExplosionTranscript4E0B30(t *testing.T) []byte {
	t.Helper()
	subclasses := []uint32{70210, 66050, 458753, 69778, 0x400, 0xc00}
	damages := []int32{-9, 0, 1, 3, 5413}
	protections := []float64{0, 0.25, float64(math.Float32frombits(0x3defaf0d)), 1}
	var transcript bytes.Buffer
	run := func(si, ti, wi, typ, di, pi, mode int) {
		target, source := damageMeleeUnitFixture4E17B0(t, false), damageMeleeUnitFixture4E17B0(t, si == 0)
		target.ObjClass, target.ObjSubClass = object.Class(2097154), object.SubClass(subclasses[ti])
		target.ObjFlags = object.Flags([]uint32{17302020, 151519748, 17302164, 0, 0, 0}[ti])
		ud := target.UpdateDataMonster()
		*ud = MonsterUpdateData{Field547: 99, Field546: 77, Field523_2: 88, Field1: math.Float32bits(0.125)}
		source.ObjClass, source.ObjFlags = object.Class(2621444), object.Flags(16777732)
		if si == 1 {
			source.ObjClass, source.ObjSubClass = object.Class(2097154), 69778
		}
		source.PrevPos = types.Ptf(32.5, -18.25)
		weapon := &Object{TypeInd: 777, ObjClass: object.Class([]uint32{85458944, 0x1000000, 0x1000, 0x1000, 0x1000000}[wi]),
			ObjSubClass: object.SubClass([]uint32{0x100, 0x10, 0x10000, 0x10000, 0x4000}[wi]), ObjFlags: object.Flags(17301776), PrevPos: types.Ptf(44, 7), InitData: unsafe.Pointer(&ModifierInitData{})}
		if wi == 0 {
			weapon.TypeInd = 1174
		}
		if wi == 4 {
			weapon.ObjSubClass = 0x4000
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
				t.Fatal("armed EXPLOSION enemy identity")
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
				t.Fatal("armed EXPLOSION audio identity")
			}
			events = append(events, map[int]int64{135: 5, 104: 8, 71: 19}[id])
		}
		r.BuffOff = func(v *Object, id EnchantID) {
			if v != target || (id != 0 && id != 22) {
				t.Fatal("armed EXPLOSION buff identity")
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
				t.Fatal("armed EXPLOSION Shock balance")
			}
			return 80
		}
		r.CallDamage = func(v, a, w *Object, d int32, dt object.DamageType) bool {
			if v != source || a != target || w != nil || d != 80 || dt != object.DamageElectric {
				t.Fatal("armed EXPLOSION Shock callback identity")
			}
			events = append(events, 3)
			return false
		}
		r.PlayerSetState = func(v *Object, state PlayerState) bool {
			if v != source || si != 0 || state != PlayerState23 {
				t.Fatal("armed EXPLOSION retaliation state")
			}
			events = append(events, 6)
			v.UpdateDataPlayer().State = state
			return true
		}
		r.FireProtection = func(v *Object) float64 {
			if v != target {
				t.Fatal("armed EXPLOSION protection identity")
			}
			events = append(events, 7)
			return protections[pi]
		}
		r.MonsterHasHitSound = func(v *Object) bool {
			if v != source || si != 1 {
				t.Fatal("armed EXPLOSION monster sound identity")
			}
			events = append(events, 17)
			return false
		}
		r.DefaultDamageSound = func(v, w *Object) {
			if v != target || w != weapon {
				t.Fatal("armed EXPLOSION damage sound identity")
			}
			events = append(events, 18)
		}
		r.AdjustFieldGuide = func(a, v *Object, d int32) int32 {
			if a != source || v != target {
				t.Fatal("armed EXPLOSION field guide identity")
			}
			events = append(events, 14)
			return d
		}
		r.ShieldReduce = func(v *Object, d *int32, dt object.DamageType, w *Object) {
			if v != target || dt != object.DamageType(typ) || w != weapon {
				t.Fatal("armed EXPLOSION Shield identity")
			}
			*d = 0
			if mode == 7 {
				*d = 5
			}
			events = append(events, 15)
		}
		r.DamageClear = func(v *Object, d int32) {
			if v != target {
				t.Fatal("armed EXPLOSION HP identity")
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
					t.Fatal("armed EXPLOSION Defend identity")
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
					t.Fatal("armed EXPLOSION pre-damage identity")
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
			for _, wi := range []int{0, 3, 4} {
				for _, typ := range []int{7} {
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
						run(si, ti, wi, 7, 3, 1, mode)
					}
				}
			}
		}
	}
	return transcript.Bytes()
}

func TestDefaultDamageArmedExplosion4E0B30OriginalTranscript(t *testing.T) {
	got := armedExplosionTranscript4E0B30(t)
	const want = "bf472f152a13c5365a9c47ba20161bc6e0f3b6754ebeb04bcf9d73765e84314b"
	if sum := fmt.Sprintf("%x", sha256.Sum256(got)); sum != want {
		t.Fatalf("independent original EXPLOSION transcript bytes=%d SHA=%s want=%s", len(got), sum, want)
	}
	t.Logf("unchanged PE32 armed EXPLOSION transcript rows=%d bytes=%d SHA=%s", bytes.Count(got, []byte{'\n'}), len(got), want)
}

func TestDefaultDamageArmedExplosion4E0B30AdmissionAndServices(t *testing.T) {
	fixture := func(t *testing.T) (*Object, *Object, *Object, DefaultDamageWorldRuntime4E0B30) {
		v, a := damageMeleeUnitFixture4E17B0(t, false), damageMeleeUnitFixture4E17B0(t, false)
		v.TypeInd, v.ObjClass, v.ObjSubClass, v.ObjFlags = 1314, object.Class(2097154), 70210, object.Flags(17302020)
		a.TypeInd, a.ObjClass, a.ObjSubClass, a.ObjFlags = 1343, object.Class(2097154), 69778, object.Flags(16777732)
		w := &Object{TypeInd: 1174, ObjClass: object.Class(85458944), ObjSubClass: 256, ObjFlags: object.Flags(17301776), InitData: unsafe.Pointer(&ModifierInitData{})}
		r := damageMeleeWorldRuntime4E0B30(t)
		r.FireProtection = func(*Object) float64 { return 0 }
		return v, a, w, r
	}
	for _, shape := range []string{"non-unit-source", "weapon-source", "nil-source-update", "unit-weapon", "world-weapon", "ranged-weapon", "ranged-wand", "missing-ranged-wand-data", "poison-type"} {
		t.Run("reject-"+shape, func(t *testing.T) {
			v, a, w, r := fixture(t)
			typ := object.DamageExplosion
			switch shape {
			case "non-unit-source":
				a.ObjClass = object.ClassSimple
			case "weapon-source":
				a.ObjClass |= object.ClassWeapon
			case "nil-source-update":
				a.UpdateData = nil
			case "unit-weapon":
				w.ObjClass |= object.ClassMonster
			case "world-weapon":
				w.ObjClass = object.ClassSimple
			case "ranged-weapon":
				w.ObjSubClass = 0x10
			case "ranged-wand", "missing-ranged-wand-data":
				w.ObjClass, w.ObjSubClass = object.ClassWand, 0x10000
				if shape == "ranged-wand" {
					w.UseData = UseDataPtr{Ptr: unsafe.Pointer(&WandUseData{})}
				}
			case "poison-type":
				typ = object.DamagePoison
			}
			reason := ""
			r.Unsupported = func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = why }
			ok := DefaultDamageWorld4E0B30(v, a, w, 3, typ, r)
			if shape == "poison-type" {
				// The unchanged original exits through poison immunity before attribution/HP.
				u := v.UpdateDataMonster()
				if !ok || reason != "" || v.HealthData.Cur != 200 || v.Obj130 != nil || v.Frame134 != 0 || u.Field547 != 0 || u.Field546 != 0 || u.StatusFlags != 0 {
					t.Fatal("original poison immunity control")
				}
				return
			}
			if shape == "world-weapon" || shape == "non-unit-source" || shape == "ranged-weapon" {
				// Independent original replays forward raw3 for each exact historical input.
				u := v.UpdateDataMonster()
				if !ok || reason != "" || v.HealthData.Cur != 197 || v.Obj130 != w || v.Field131 != uint32(typ) || v.Frame134 != 1400 || (shape != "world-weapon" && v.Pos132 != a.PrevPos) || u.Field547 != 1 || u.Field546 != uint32(w.TypeInd) {
					t.Fatal("world explosion control")
				}
				return
			}
			if !ok || reason != "unsupported monster damage shape" || v.HealthData.Cur != 200 || v.Obj130 != nil {
				t.Fatalf("unrelated shape admitted: reason=%q HP=%d", reason, v.HealthData.Cur)
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
				r.FireProtection = nil
			case "buff-off":
				r.BuffOff = nil
			case "HP":
				r.DamageClear = nil
			case "monster-sound":
				r.MonsterHasHitSound = nil
			}
			reason := ""
			r.Unsupported = func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = why }
			if !DefaultDamageWorld4E0B30(v, a, w, 3, object.DamageExplosion, r) || reason == "" || reason == "unsupported monster damage shape" || v.HealthData.Cur != 200 || v.Obj130 != nil {
				t.Fatalf("missing service not reported before tail: reason=%q", reason)
			}
		})
	}
	t.Run("nil-health-reaches-ordinary-HP-boundary", func(t *testing.T) {
		v, a, w, r := fixture(t)
		v.HealthData = nil
		calls := 0
		r.DamageClear = func(target *Object, damage int32) {
			if target != v || damage != 3 || target.Obj130 != w || target.Frame134 != 1400 {
				t.Fatal("nil-health EXPLOSION HP identity/prefix")
			}
			calls++
		}
		if !DefaultDamageWorld4E0B30(v, a, w, 3, object.DamageExplosion, r) || calls != 1 {
			t.Fatal("UnitDamageClear owns the original late nil-health boundary")
		}
	})
}

func TestDefaultDamageArmedExplosion4E0B30LiveFireAndRepeatedHP(t *testing.T) {
	for _, installImmunity := range []bool{false, true} {
		t.Run(fmt.Sprintf("Shock-installs-fire-immunity-%t", installImmunity), func(t *testing.T) {
			v, a := damageMeleeUnitFixture4E17B0(t, false), damageMeleeUnitFixture4E17B0(t, false)
			v.ObjSubClass, a.ObjSubClass = 70210, 69778
			v.Buffs = 1 << defaultDamageShockEnchant4E0B30
			if !installImmunity {
				v.ObjSubClass |= 0x400
			}
			w := &Object{TypeInd: 1174, ObjClass: object.Class(85458944), ObjSubClass: 256, InitData: unsafe.Pointer(&ModifierInitData{})}
			r := damageMeleeWorldRuntime4E0B30(t)
			shocks, soundLookups, damageSounds := 0, 0, 0
			r.Audio = func(int, *Object) {}
			r.BuffOff = func(v *Object, id EnchantID) { v.Buffs &^= 1 << id }
			r.BalanceFloatInd = func(string, int) float64 { return 80 }
			r.CallDamage = func(target, attacker, weapon *Object, damage int32, typ object.DamageType) bool {
				if target != a || attacker != v || weapon != nil || damage != 80 || typ != object.DamageElectric || v.HasEnchant(defaultDamageShockEnchant4E0B30) {
					t.Fatal("Shock callback identity/order")
				}
				shocks++
				v.ObjSubClass &^= 0x400
				if installImmunity {
					v.ObjSubClass |= 0x400
				}
				return false
			}
			r.FireProtection = func(*Object) float64 { return 0 }
			r.MonsterHasHitSound = func(source *Object) bool {
				if source != a {
					t.Fatal("NPC hit-sound source")
				}
				soundLookups++
				return installImmunity
			}
			r.DefaultDamageSound = func(target, weapon *Object) {
				if target != v || weapon != w || installImmunity {
					t.Fatal("duplicate NPC hit sound or wrong weapon")
				}
				damageSounds++
			}
			for hit := range 3 {
				perHit := 3
				if installImmunity {
					perHit = 1
				}
				if !DefaultDamageWorld4E0B30(v, a, w, 3, object.DamageExplosion, r) || int(v.HealthData.Cur) != 200-perHit*(hit+1) || v.Obj130 != w || v.Pos132 != a.PrevPos || v.Field131 != 7 || v.Frame134 != 1400 || v.UpdateDataMonster().Field547 != 1 || v.UpdateDataMonster().Field546 != 1174 || a.UpdateDataMonster().Field130 != 1400 {
					t.Fatalf("repeated EXPLOSION hit=%d HP=%d", hit, v.HealthData.Cur)
				}
			}
			if shocks != 1 || soundLookups != 3 || (installImmunity && damageSounds != 0) || (!installImmunity && damageSounds != 3) {
				t.Fatalf("live callback counts: Shock=%d lookups=%d damage-sounds=%d", shocks, soundLookups, damageSounds)
			}
		})
	}
}
