package server

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func playerArmedExplosionFixture4E0B30(t *testing.T, playerSource bool) (*Object, *Object, *Object, DefaultDamageWorldRuntime4E0B30) {
	t.Helper()
	v, a := damageMeleeUnitFixture4E17B0(t, true), damageMeleeUnitFixture4E17B0(t, playerSource)
	v.TypeInd, v.ObjClass, v.ObjFlags = 713, object.Class(2621444), object.Flags(16777732)
	v.HealthData = &HealthData{Cur: 2000, Max: 2000, Field2: 2000}
	u := v.UpdateDataPlayer()
	u.Field76, u.Field75, u.Field21, u.Field57 = 99, 77, math.Float32bits(.125), math.Float32bits(.5)
	a.ObjFlags = object.Flags(16777732)
	if playerSource {
		a.TypeInd, a.ObjClass = 713, object.Class(2621444)
	} else {
		a.TypeInd, a.ObjClass, a.ObjSubClass = 1343, object.Class(2097154), 69778
	}
	w := &Object{TypeInd: 1174, ObjClass: object.Class(85458944), ObjSubClass: 256, ObjFlags: object.Flags(17301776), PrevPos: types.Ptf(44, 7), InitData: unsafe.Pointer(&ModifierInitData{})}
	r := damageMeleeWorldRuntime4E0B30(t)
	r.FireProtection = func(*Object) float64 { return 0 }
	return v, a, w, r
}

// DefaultDamage receives an already calculated integer from PlayerDamage.
// It must not repeat the armor, carry, hit-marker or GodMode/Quest prefix.
func TestDefaultDamagePlayerArmedExplosion4E0B30RepeatedHP(t *testing.T) {
	for _, fromPlayer := range []bool{false, true} {
		for _, attack := range []string{"Sword", "Wand", "Hammer"} {
			t.Run(fmt.Sprintf("player-source-%t/%s", fromPlayer, attack), func(t *testing.T) {
				v, a, w, r := playerArmedExplosionFixture4E0B30(t, fromPlayer)
				if attack == "Wand" {
					w.ObjClass, w.ObjSubClass = object.ClassWand, 0x10000
					w.UseData.Ptr = unsafe.Pointer(&WandUseData{Flags: 2})
				}
				if attack == "Hammer" {
					w.ObjSubClass = 0x4000
				}
				before := *v.UpdateDataPlayer()
				for hit := range 3 {
					if !DefaultDamageWorld4E0B30(v, a, w, 3, object.DamageExplosion, r) || v.HealthData.Cur != uint16(1997-3*hit) || *v.UpdateDataPlayer() != before || v.Obj130 != w || v.Pos132 != a.PrevPos || v.Field131 != 7 || v.Frame134 != 1400 {
						t.Fatalf("hit=%d HP=%d; EXPLOSION attribution or independent prefix changed", hit, v.HealthData.Cur)
					}
				}
			})
		}
	}
}

// Match the independent 774-row unchanged GAME.EXE replay. Only declared
// external services are replaced; the root, qualifiers and live effect loops
// are production Go. The real C-owned HP entry is verified separately.
func TestDefaultDamagePlayerArmedExplosion4E0B30OriginalTranscript(t *testing.T) {
	states := []PlayerState{PlayerState13, PlayerState1, PlayerState15, PlayerState30}
	damages := []int32{-9, 0, 1, 3, 20, 61}
	protections := []float64{0, .25, float64(math.Float32frombits(0x3defaf0d)), 1}
	var transcript bytes.Buffer
	run := func(si, ti, wi, di, pi, mode int) {
		v, a, w, r := playerArmedExplosionFixture4E0B30(t, si == 0)
		u := v.UpdateDataPlayer()
		u.State = states[ti]
		if wi == 3 {
			w.TypeInd, w.ObjClass, w.ObjSubClass = 777, object.ClassWand, 0x10000
			w.UseData.Ptr = unsafe.Pointer(&WandUseData{Flags: 2})
		}
		if wi == 4 {
			w.TypeInd, w.ObjClass, w.ObjSubClass = 777, object.ClassWeapon, 0x4000
		}
		v.Buffs = 1<<22 | 1
		switch mode {
		case 3:
			v.ObjFlags |= object.FlagNoUpdate
		case 4:
			v.Buffs |= 1 << 23
		case 5:
			v.ObjFlags |= object.FlagDead
		case 6, 7:
			v.Buffs |= 1 << 26
		}
		var events []int64
		amount, calls := uint32(0xfeedface), int64(0)
		r.GameplayFlag1 = func() bool { events = append(events, 1); return mode != 1 }
		r.QuestMode = func() bool { return false }
		r.IsZombie = func(*Object) bool { return false }
		r.IsEnemy = func(target, source *Object) bool {
			if target != v || source != a {
				t.Fatal("armed player EXPLOSION enemy identity")
			}
			events = append(events, 2)
			if mode == 9 {
				w.ObjSubClass = 0x10
				if wi == 3 {
					w.ObjSubClass, w.UseData.AsWand().Flags = 0x10000, 0
				}
			}
			return mode != 1 && mode != 2
		}
		r.Audio = func(id int, target *Object) {
			want := v
			if id == 135 {
				want = a
			}
			if target != want {
				t.Fatal("armed player EXPLOSION audio identity")
			}
			events = append(events, map[int]int64{135: 5, 104: 8, 71: 19}[id])
		}
		r.BuffOff = func(target *Object, id EnchantID) {
			if target != v || (id != 0 && id != 22) {
				t.Fatal("armed player EXPLOSION buff identity")
			}
			v.Buffs &^= 1 << id
			event := int64(9)
			if id == 22 {
				event = 4
			}
			events = append(events, event)
		}
		r.BalanceFloatInd = func(key string, index int) float64 {
			if key != "ShockDamage" || index != 4 {
				t.Fatal("Shock balance identity")
			}
			return 80
		}
		r.CallDamage = func(target, source, weapon *Object, damage int32, typ object.DamageType) bool {
			if target != a || source != v || weapon != nil || damage != 80 || typ != object.DamageElectric {
				t.Fatal("Shock return identity")
			}
			events = append(events, 3)
			return false
		}
		r.PlayerSetState = func(target *Object, state PlayerState) bool {
			if target == a {
				if si != 0 || state != PlayerState23 {
					t.Fatal("Shock source state")
				}
				events = append(events, 6)
			} else {
				if target != v || state != PlayerState30 {
					t.Fatal("late player hurt state")
				}
				events = append(events, 20)
			}
			target.UpdateDataPlayer().State = state
			return true
		}
		r.FireProtection = func(target *Object) float64 {
			if target != v {
				t.Fatal("fire protection identity")
			}
			events = append(events, 7)
			return protections[pi]
		}
		r.MonsterHasHitSound = func(source *Object) bool {
			if source != a || si != 1 {
				t.Fatal("monster sound identity")
			}
			events = append(events, 17)
			return false
		}
		r.DefaultDamageSound = func(target, weapon *Object) {
			if target != v || weapon != w {
				t.Fatal("damage sound identity")
			}
			events = append(events, 18)
		}
		r.ShieldReduce = func(target *Object, damage *int32, typ object.DamageType, weapon *Object) {
			if target != v || typ != object.DamageExplosion || weapon != w {
				t.Fatal("late Shield identity")
			}
			*damage = 5
			if mode == 6 {
				*damage = 0
			}
			events = append(events, 15)
		}
		r.DamageClear = func(target *Object, damage int32) {
			if target != v {
				t.Fatal("HP identity")
			}
			amount, calls = uint32(damage), calls+1
			events = append(events, 16)
		}
		if mode == 8 {
			first, second, preFirst, preLast := &ModifierEff{}, &ModifierEff{}, &ModifierEff{}, &ModifierEff{}
			first.Defend76.Fnc, second.Defend76.Fnc = unsafe.Pointer(new(byte)), unsafe.Pointer(new(byte))
			preFirst.AttackPreDmg64.Fnc, preLast.AttackPreDmg64.Fnc = unsafe.Pointer(new(byte)), unsafe.Pointer(new(byte))
			item := &Object{ObjFlags: object.FlagEquipped, InitData: unsafe.Pointer(&ModifierInitData{Modifiers: [4]*ModifierEff{nil, nil, first, second}})}
			v.InvFirstItem = item
			w.InitDataModifier().Modifiers = [4]*ModifierEff{preFirst, nil, nil, preLast}
			r.CanApplyLateDefend = func(*ModifierEff) bool { return true }
			r.ApplyLateDefend = func(mod *ModifierEff, gotItem, target, weapon, source *Object, damage int32, typ object.DamageType) int32 {
				if gotItem != item || target != v || weapon != w || source != a || typ != object.DamageExplosion {
					t.Fatal("late Defend identity")
				}
				if mod == first {
					events = append(events, 10)
					return damage + 5
				}
				if mod != second {
					t.Fatal("late Defend slot")
				}
				events = append(events, 11)
				return damage - 2
			}
			r.CanApplyPreDamage = func(*ModifierEff) bool { return true }
			r.ApplyPreDamage = func(mod *ModifierEff, weapon, source, target *Object, damage *int32) {
				if weapon != w || source != a || target != v {
					t.Fatal("pre-Damage identity")
				}
				if mod == preFirst {
					events = append(events, 12)
					*damage *= 2
				} else {
					if mod != preLast {
						t.Fatal("pre-Damage slot")
					}
					events = append(events, 13)
					*damage--
				}
			}
		}
		result := DefaultDamageWorld4E0B30(v, a, w, damages[di], object.DamageExplosion, r)
		bit := func(value bool) int64 {
			if value {
				return 1
			}
			return 0
		}
		hitFrame := uint32(0)
		if si == 1 {
			hitFrame = a.UpdateDataMonster().Field130
		}
		row := []int64{int64(si), int64(ti), int64(wi), 7, int64(di), int64(pi), int64(mode), bit(result), int64(u.Field76), int64(u.Field75), int64(v.Buffs), bit(v.Obj130 == w), int64(math.Float32bits(v.Pos132.X)), int64(math.Float32bits(v.Pos132.Y)), int64(v.Field131), int64(v.Frame134), int64(u.State), int64(u.Field21), int64(amount), calls, int64(hitFrame)}
		line, err := json.Marshal(append(row, events...))
		if err != nil {
			t.Fatal(err)
		}
		transcript.Write(line)
		transcript.WriteByte('\n')
	}
	for si := range 2 {
		for ti := range 4 {
			for _, wi := range []int{0, 3, 4} {
				for di := range 6 {
					for pi := range 4 {
						run(si, ti, wi, di, pi, 0)
					}
				}
			}
		}
	}
	for si := range 2 {
		for ti := range 4 {
			for _, wi := range []int{0, 3, 4} {
				for mode := 1; mode < 10; mode++ {
					if mode != 8 || ti == 3 {
						run(si, ti, wi, 3, 1, mode)
					}
				}
			}
		}
	}
	// Startup HP arguments are projected from unchanged original arithmetic;
	// all 774 actor/event rows and the old control digest remain recorded.
	const want = "e0fa7b58811727144cce69869ac6be8e9dbf05721575c2c56618716c507cfd9a"
	if sum := fmt.Sprintf("%x", sha256.Sum256(transcript.Bytes())); sum != want || transcript.Len() != 78268 || bytes.Count(transcript.Bytes(), []byte{'\n'}) != 774 {
		t.Fatalf("unchanged PE32 player armed EXPLOSION rows=%d bytes=%d SHA=%s want=%s", bytes.Count(transcript.Bytes(), []byte{'\n'}), transcript.Len(), sum, want)
	}
	t.Logf("unchanged PE32 player armed EXPLOSION rows=774 bytes=78268 SHA=%s", want)
}

func TestDefaultDamagePlayerArmedExplosion4E0B30ServiceBoundaries(t *testing.T) {
	for _, missing := range []string{"enemy", "buff", "HP", "monster-sound", "fire", "hurt"} {
		t.Run(missing, func(t *testing.T) {
			v, a, w, r := playerArmedExplosionFixture4E0B30(t, false)
			var events []string
			raw := int32(3)
			r.FireProtection = func(*Object) float64 { events = append(events, "fire"); return 0 }
			r.BuffOff = func(*Object, EnchantID) { events = append(events, "buff") }
			r.DamageClear = func(*Object, int32) { t.Fatal("unsupported service reached HP") }
			r.DefaultDamageSound = func(*Object, *Object) { events = append(events, "sound") }
			r.Unsupported = func(reason string, target, source, weapon *Object, damage int32, typ object.DamageType) {
				if target != v || source != a || weapon != w || damage != raw || typ != object.DamageExplosion || reason == "" {
					t.Fatal("unsupported identity")
				}
				events = append(events, "unsupported")
			}
			switch missing {
			case "enemy":
				r.IsEnemy = nil
			case "buff":
				r.BuffOff = nil
			case "HP":
				r.DamageClear = nil
			case "monster-sound":
				r.MonsterHasHitSound = nil
			case "fire":
				r.FireProtection = nil
			case "hurt":
				raw, r.PlayerSetState = 20, nil
			}
			want := []string{"unsupported"}
			if missing == "hurt" {
				want = []string{"fire", "buff", "sound", "unsupported"}
			}
			if !DefaultDamageWorld4E0B30(v, a, w, raw, object.DamageExplosion, r) || !slices.Equal(events, want) || v.HealthData.Cur != 2000 || v.UpdateDataPlayer().Field76 != 99 || v.UpdateDataPlayer().Field75 != 77 {
				t.Fatalf("service=%s events=%v HP=%d", missing, events, v.HealthData.Cur)
			}
		})
	}
	for _, noUpdate := range []bool{false, true} {
		t.Run(fmt.Sprintf("nil-health/update-%t", noUpdate), func(t *testing.T) {
			v, a, w, r := playerArmedExplosionFixture4E0B30(t, false)
			v.HealthData = nil
			if noUpdate {
				v.UpdateData = nil
			}
			calls := 0
			r.DamageClear = func(target *Object, damage int32) {
				if target != v || damage != 3 {
					t.Fatal("late guarded HP identity")
				}
				calls++
			}
			if !DefaultDamageWorld4E0B30(v, a, w, 3, object.DamageExplosion, r) || calls != 1 || v.Obj130 != w || v.Field131 != 7 {
				t.Fatal("early nil-health guard replaced original late HP guard")
			}
		})
	}
}
