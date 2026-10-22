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

// Only the explicitly declared service boundaries are observed here. The
// independent unchanged PE32 replay executes DefaultDamage, owner/qualifier
// code, metadata stores and item-effect loops. HP/death are not emulated by
// this transcript; the real C-owned collision and HP path is tested in legacy.
func turnUndeadDamageTranscript4E0B30(t *testing.T) []byte {
	t.Helper()
	subclasses := []uint32{70210, 0x440, 0x840, 0x11052}
	damages := []int32{-9, 0, 1, 7, 20, 5413}
	var transcript bytes.Buffer
	for si := range 3 {
		for ti := range subclasses {
			for di := range damages {
				for mode := range 8 {
					target := damageMeleeUnitFixture4E17B0(t, false)
					target.ObjClass, target.ObjSubClass, target.ObjFlags = object.Class(2097154), object.SubClass(subclasses[ti]), object.Flags(16777732)
					u := target.UpdateDataMonster()
					*u = MonsterUpdateData{Field547: 99, Field546: 77, Field523_2: 88, Field1: math.Float32bits(0.125)}
					var source *Object
					if si != 2 {
						source = damageMeleeUnitFixture4E17B0(t, si == 0)
						source.ObjClass, source.ObjSubClass, source.ObjFlags = object.Class(2621444), 0, object.Flags(16777732)
						if si == 1 {
							source.ObjClass, source.ObjSubClass = object.Class(2097154), 69778
						}
						source.PrevPos = types.Ptf(32.5, -18.25)
					}
					// Stock UndeadKiller is LIGHT|CLIENT_PREDICT, NOT MISSILE.
					// Type 777 is a declared synthetic attribution ID, not stock.
					weapon := &Object{TypeInd: 777, ObjClass: object.Class(0x40080000), PrevPos: types.Ptf(44, 7)}
					target.Buffs = 1<<defaultDamageShockEnchant4E0B30 | 1<<defaultDamageInvisibleEnchant4E0B30
					switch mode {
					case 3:
						target.ObjFlags |= object.FlagNoUpdate
					case 4:
						target.Buffs |= 1 << defaultDamageInvulnerableEnchant4E0B30
					case 5:
						target.ObjFlags |= object.FlagDead
					case 6, 7:
						target.Buffs |= 1 << defaultDamageShieldEnchant4E0B30
					}
					var events []int64
					resultDamage, hpCalls := uint32(0xfeedface), int64(0)
					r := damageMeleeWorldRuntime4E0B30(t)
					r.GameplayFlag1 = func() bool { events = append(events, 1); return mode != 1 }
					r.QuestMode = func() bool { return false }
					r.IsZombie = func(v *Object) bool {
						if v != target {
							t.Fatal("Turn Undead zombie identity")
						}
						return false
					}
					r.IsEnemy = func(v, a *Object) bool {
						if v != target || a != source || source == nil {
							t.Fatal("Turn Undead enemy identity")
						}
						events = append(events, 2)
						return mode != 1 && mode != 2
					}
					r.Audio = func(id int, v *Object) {
						if id != 71 || v != target {
							t.Fatal("Turn Undead invulnerability audio")
						}
						events = append(events, 19)
					}
					r.BuffOff = func(v *Object, id EnchantID) {
						if v != target || id != 0 {
							t.Fatal("Turn Undead invisibility identity")
						}
						v.Buffs &^= 1 << id
						events = append(events, 9)
					}
					r.MonsterHasHitSound = func(a *Object) bool {
						if a != source || si != 1 {
							t.Fatal("Turn Undead source sound identity")
						}
						events = append(events, 17)
						return false
					}
					r.DefaultDamageSound = func(v, w *Object) {
						if v != target || w != weapon {
							t.Fatal("Turn Undead weapon sound identity")
						}
						events = append(events, 18)
					}
					r.AdjustFieldGuide = func(a, v *Object, damage int32) int32 {
						if a != source || v != target {
							t.Fatal("Turn Undead field-guide identity")
						}
						events = append(events, 14)
						return damage
					}
					r.ShieldReduce = func(v *Object, damage *int32, typ object.DamageType, w *Object) {
						if v != target || typ != object.DamageDispelUndead || w != weapon {
							t.Fatal("Turn Undead Shield identity")
						}
						*damage = 0
						if mode == 7 {
							*damage = 5
						}
						events = append(events, 15)
					}
					r.DamageClear = func(v *Object, damage int32) {
						if v != target {
							t.Fatal("Turn Undead HP identity")
						}
						resultDamage, hpCalls = uint32(damage), hpCalls+1
						events = append(events, 16)
					}
					result := DefaultDamageWorld4E0B30(target, source, weapon, damages[di], object.DamageDispelUndead, r)
					resultInt, weaponInt, sourceFrame := int64(0), int64(0), uint32(0)
					if result {
						resultInt = 1
					}
					if target.Obj130 == weapon {
						weaponInt = 1
					}
					if si == 1 {
						sourceFrame = source.UpdateDataMonster().Field130
					}
					row := []int64{int64(si), int64(ti), int64(di), int64(mode), resultInt,
						int64(u.Field547), int64(u.Field546), int64(u.StatusFlags), int64(target.Buffs), weaponInt,
						int64(math.Float32bits(target.Pos132.X)), int64(math.Float32bits(target.Pos132.Y)), int64(target.Field131), int64(target.Frame134),
						int64(u.Field523_2), int64(u.Field1), int64(resultDamage), hpCalls, int64(sourceFrame)}
					row = append(row, events...)
					encoded, err := json.Marshal(row)
					if err != nil {
						t.Fatal(err)
					}
					transcript.Write(encoded)
					transcript.WriteByte('\n')
				}
			}
		}
	}
	return transcript.Bytes()
}

func TestDefaultDamageTurnUndeadOriginalTranscript4E0B30(t *testing.T) {
	got := turnUndeadDamageTranscript4E0B30(t)
	// 576 independent unchanged GAME.EXE cases: PLAYER/NPC/nil source,
	// fire/electric-immune subclasses, signed raw inputs and all early gates.
	const originalSHA = "98d9f58f6d869d60a99d9c4bd9eb9c8a22e51371ef4b15543389911d63025587"
	if bytes.Count(got, []byte{'\n'}) != 576 || fmt.Sprintf("%x", sha256.Sum256(got)) != originalSHA {
		t.Fatalf("Turn Undead original transcript: rows=%d SHA=%x want=%s", bytes.Count(got, []byte{'\n'}), sha256.Sum256(got), originalSHA)
	}
	t.Logf("unchanged PE32 Turn Undead comparison rows=576 bytes=%d sha256=%x; declared external services, not full Windows world fidelity", len(got), sha256.Sum256(got))
}

func TestDefaultDamageTurnUndeadAdmissionServices4E0B30(t *testing.T) {
	for _, name := range []string{"nil-health", "unowned", "class-zero-owner", "nil-source", "missing-enemy", "missing-buff", "missing-HP", "missing-NPC-sound", "player-without-NPC-sound", "no-update-without-tail", "invulnerable-without-tail", "missile-weapon", "unit-weapon", "weapon-class", "wand-class", "source-without-update", "player-target", "non-undead-direct-tail"} {
		t.Run(name, func(t *testing.T) {
			v, a := damageMeleeUnitFixture4E17B0(t, false), damageMeleeUnitFixture4E17B0(t, true)
			v.ObjSubClass = 70210
			w := &Object{TypeInd: 777, ObjClass: object.Class(0x40080000), PrevPos: types.Ptf(44, 7)}
			calls, enemyCalls, reason := 0, 0, ""
			wantCalls, wantEnemy, wantReason := 1, 1, ""
			r := damageMeleeWorldRuntime4E0B30(t)
			r.FireProtection, r.ElectricProtection = nil, nil
			r.IsEnemy = func(*Object, *Object) bool { enemyCalls++; return true }
			r.DamageClear = func(unit *Object, damage int32) {
				if unit != v || damage != -9 {
					t.Fatal("type 6 gained elemental rounding/minimum")
				}
				calls++
			}
			r.Unsupported = func(why string, target, source, weapon *Object, damage int32, typ object.DamageType) {
				if target != v || source != a || weapon != w || damage != -9 || typ != object.DamageDispelUndead {
					t.Fatal("Turn Undead rejection identity")
				}
				reason = why
			}
			switch name {
			case "nil-health":
				v.HealthData = nil
			case "unowned":
				a = w
			case "class-zero-owner":
				a = &Object{}
			case "nil-source":
				a, r.IsEnemy, r.BuffOff, r.MonsterHasHitSound, wantEnemy = nil, nil, nil, nil, 0
			case "missing-enemy":
				r.IsEnemy, wantCalls, wantEnemy, wantReason = nil, 0, 0, "missing Turn Undead enemy service"
			case "missing-buff":
				r.BuffOff, wantCalls, wantReason = nil, 0, "missing Turn Undead tail service"
			case "missing-HP":
				r.DamageClear, wantCalls, wantReason = nil, 0, "missing Turn Undead tail service"
			case "missing-NPC-sound":
				a = damageMeleeUnitFixture4E17B0(t, false)
				r.MonsterHasHitSound, wantCalls, wantReason = nil, 0, "missing Turn Undead tail service"
			case "player-without-NPC-sound":
				r.MonsterHasHitSound = nil
			case "no-update-without-tail":
				v.ObjFlags |= object.FlagNoUpdate
				r.DamageClear, r.BuffOff, r.MonsterHasHitSound, wantCalls = nil, nil, nil, 0
			case "invulnerable-without-tail":
				v.Buffs |= 1 << ENCHANT_INVULNERABLE
				r.DamageClear, r.BuffOff, r.MonsterHasHitSound, r.IsEnemy, wantCalls, wantEnemy = nil, nil, nil, nil, 0, 0
			case "missile-weapon":
				w.ObjClass = object.ClassMissile
			case "unit-weapon":
				w.ObjClass = object.ClassPlayer
			case "weapon-class":
				w.ObjClass = object.ClassWeapon
			case "wand-class":
				w.ObjClass = object.ClassWand
			case "source-without-update":
				a.UpdateData = nil
			case "player-target":
				v = damageMeleeUnitFixture4E17B0(t, true)
			case "non-undead-direct-tail":
				v.ObjSubClass &^= 0x40
			}
			switch name {
			case "unit-weapon", "weapon-class", "wand-class", "source-without-update":
				wantCalls, wantEnemy, wantReason = 0, 0, "unsupported monster damage shape"
			case "player-target":
				wantCalls, wantEnemy, wantReason = 0, 0, "non-monster unit target"
			}
			if !DefaultDamageWorld4E0B30(v, a, w, -9, object.DamageDispelUndead, r) || calls != wantCalls || enemyCalls != wantEnemy || reason != wantReason {
				t.Fatalf("tail calls=%d/%d enemy=%d/%d reason=%q/%q", calls, wantCalls, enemyCalls, wantEnemy, reason, wantReason)
			}
			if name == "missile-weapon" {
				if v.Obj130 != w || v.Pos132 != w.PrevPos || v.Frame134 != 1400 || v.Field131 != 6 || v.UpdateDataMonster().Field547 != 1 || v.UpdateDataMonster().Field546 != 777 {
					t.Fatal("original-confirmed type6 missile attribution")
				}
			}
		})
	}
}

func TestDefaultDamageTurnUndeadLiveEnemyQualifier4E0B30(t *testing.T) {
	for _, hammer := range []bool{false, true} {
		t.Run(fmt.Sprintf("callback-makes-Hammer-%t", hammer), func(t *testing.T) {
			v, a := damageMeleeUnitFixture4E17B0(t, false), damageMeleeUnitFixture4E17B0(t, true)
			v.ObjSubClass = 70210
			// The callback introduces WEAPON class, whose original later
			// pre-damage loop dereferences InitData. Supply that live record
			// as an input; do not turn the unguarded loop into a no-op.
			w := &Object{TypeInd: 777, ObjClass: object.Class(0x40080000), InitData: unsafe.Pointer(&ModifierInitData{})}
			r := damageMeleeWorldRuntime4E0B30(t)
			enemy, hpCalls := 0, 0
			r.IsEnemy = func(*Object, *Object) bool {
				enemy++
				w.ObjClass, w.ObjSubClass = object.ClassWeapon, object.SubClass(object.WeaponSword)
				if hammer {
					w.ObjSubClass = object.SubClass(object.WeaponHammer)
				}
				return false
			}
			r.DamageClear = func(*Object, int32) { hpCalls++ }
			if !DefaultDamageWorld4E0B30(v, a, w, 7, object.DamageDispelUndead, r) || enemy != 1 || hpCalls != map[bool]int{false: 0, true: 1}[hammer] {
				t.Fatalf("live qualifier enemy=%d HP=%d hammer=%t", enemy, hpCalls, hammer)
			}
		})
	}
}
