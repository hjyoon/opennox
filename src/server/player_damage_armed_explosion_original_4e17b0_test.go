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

	"github.com/opennox/opennox/v1/common/unit/ai"
)

// This digest is the 7,194-row transcript from unchanged GAME.EXE 004E17B0,
// 004E20F0 (FISTP32 included), 004E1400, 004EC580 and the real 004E2330
// first-equipped-shield loop. Buff/mode/observer/facing/audio/balance,
// armor-wear and DefaultDamage are declared external services in that replay;
// their hooks do not replace prefix/carry/selection calculations. Direct-read
// observer/action getters have input values but no callback event numbers.
// The nil-shield PE32 fault is deliberately excluded from this positive
// transcript and checked separately. This is not a real-HP or GUI certificate.
func TestPlayerDamageArmedExplosion4E17B0OriginalTranscript(t *testing.T) {
	var transcript bytes.Buffer
	rows := 0
	armors := []float32{0, .25, .5, 1}
	carries := []float32{-.25, 0, .25, .5}
	damages := []int32{-9, 0, 1, 3, 9, 61}
	scales := []float32{0, 0, .5}
	run := func(p [9]int) {
		si, ti, wi, ai0, ci, di, qi, gi, mode := p[0], p[1], p[2], p[3], p[4], p[5], p[6], p[7], p[8]
		player := ti == 1
		v, a := damageMeleeUnitFixture4E17B0(t, player), damageMeleeUnitFixture4E17B0(t, si == 1)
		v.ObjFlags, a.ObjFlags = object.Flags(16777732), object.Flags(16777732)
		v.ObjClass, v.ObjSubClass = object.Class(2097154), 69650
		a.ObjClass, a.ObjSubClass, a.TypeInd = object.Class(2097154), 69778, 1343
		if player {
			v.ObjClass, v.ObjSubClass = object.Class(2621444), 0
		}
		if si == 1 {
			a.ObjClass, a.ObjSubClass, a.TypeInd = object.Class(2621444), 0, 713
		}
		w := &Object{TypeInd: 1174, ObjClass: object.Class(85458944), ObjSubClass: 256, ObjFlags: object.Flags(17301776), InitData: unsafe.Pointer(&ModifierInitData{}), PosVec: types.Ptf(19, 11), PrevPos: types.Ptf(17, 9)}
		if wi == 1 {
			w.TypeInd, w.ObjClass, w.ObjSubClass, w.UseData.Ptr = 777, object.ClassWand, 0x10000, unsafe.Pointer(&WandUseData{Flags: 2})
		}
		if wi == 2 {
			w.TypeInd, w.ObjClass, w.ObjSubClass = 777, object.ClassWeapon, 0x4000
		}
		entry, live := v.UpdateData, damageMeleeUnitFixture4E17B0(t, player).UpdateData
		setup := func(ptr unsafe.Pointer, absorption, carry float32, marker, kind uint32) {
			if player {
				u := (*PlayerUpdateData)(ptr)
				u.Field57, u.Field21, u.Field76, u.Field75 = math.Float32bits(absorption), math.Float32bits(carry), marker, kind
				u.State = PlayerState13
			} else {
				u := (*MonsterUpdateData)(ptr)
				u.Field518, u.Field1, u.Field547, u.Field546 = math.Float32bits(absorption), math.Float32bits(carry), marker, kind
				u.AIStackInd, u.AIStack[0].Action = 0, 0
			}
		}
		setup(entry, armors[ai0], carries[ci], 99, 77)
		setup(live, .25, .5, 33, 44)
		marker := func(ptr unsafe.Pointer) (*uint32, *uint32) {
			if player {
				u := (*PlayerUpdateData)(ptr)
				return &u.Field76, &u.Field75
			}
			u := (*MonsterUpdateData)(ptr)
			return &u.Field547, &u.Field546
		}
		shieldMode := map[int]bool{7: true, 8: true, 9: true, 10: true, 12: true, 19: true, 20: true, 21: true, 22: true, 26: true}[mode]
		item, second := &Object{ObjSubClass: 2, ObjFlags: object.FlagEquipped, HealthData: &HealthData{}}, &Object{ObjSubClass: 2, ObjFlags: object.FlagEquipped, HealthData: &HealthData{}}
		if shieldMode {
			v.InvFirstItem = item
			if player {
				(*PlayerUpdateData)(entry).Player.ArmorEquip, (*PlayerUpdateData)(entry).State = 0x1000000, PlayerState16
			} else {
				(*MonsterUpdateData)(entry).ArmorEquipFlags, (*MonsterUpdateData)(entry).AIStack[0].Action = 0x1000000, uint32(ai.ACTION_BLOCK_ATTACK)
				(*MonsterUpdateData)(live).AIStack[0].Action = uint32(ai.ACTION_BLOCK_ATTACK)
			}
		}
		if mode == 12 {
			item.HealthData = nil
		}
		if mode == 1 || mode == 2 {
			v.Buffs |= 1 << playerDamageInvulnerableEnchant4E17B0
		}
		if mode == 3 {
			v.ObjFlags |= object.FlagNoUpdate
		}
		if mode == 4 {
			v.ObjFlags |= object.FlagDead
		}
		if mode == 5 {
			a.ObjOwner = v
		}
		if mode == 6 && player {
			(*PlayerUpdateData)(entry).Player.Field3680 = 1
		}
		if mode == 13 || mode == 23 || mode == 25 {
			v.Buffs |= 1 << playerDamageReflectEnchant4E17B0
		}
		if mode == 14 || mode == 23 || mode == 25 {
			(*PlayerUpdateData)(entry).Player.Field3680 = 2
			(*PlayerUpdateData)(entry).Player.CameraFollowObj = a
		}
		var events []uint64
		resultDamage, calls, wearBits, selected, remaining := uint32(0xfeedface), uint32(0), uint32(0), uint32(0), uint32(0)
		quest, god := qi, gi != 0
		r := PlayerDamageRuntime4E17B0{
			Frame: func() uint32 {
				if mode == 2 {
					return 1401
				}
				return 1400
			},
			CoopMode: func() bool { events = append(events, 1); return mode == 5 },
			ObserveClear: func(got *Object) {
				if got != v {
					t.Fatal("observer target")
				}
				events = append(events, 3)
				if mode == 14 {
					v.UpdateData = live
				} else {
					w.ObjClass |= object.ClassMissile
					w.ObjSubClass = 0
				}
			},
			BlockSourceExcluded: func(got *Object) bool {
				if got != w {
					t.Fatal("exclusion attack")
				}
				return mode == 9
			},
			BlockDirection: func(got *Object, pos types.Pointf) bool {
				want := types.Ptf(17, 9)
				if mode == 23 || mode == 25 {
					want = types.Ptf(19, 11)
				}
				if got != v || pos != want {
					t.Fatalf("direction p=%v got=%v want=%v", p, pos, want)
				}
				events = append(events, 4)
				if mode == 15 {
					v.UpdateData = live
				}
				if mode == 16 {
					m, _ := marker(entry)
					*m = 0
				}
				return map[int]bool{7: true, 9: true, 10: true, 12: true, 19: true, 20: true, 21: true, 22: true, 23: true, 25: true, 26: true}[mode]
			},
			Audio: func(sound int, got *Object) {
				if got != v {
					t.Fatal("audio target")
				}
				event, ok := map[int]uint64{71: 6, 878: 7, 122: 8}[sound]
				if !ok {
					t.Fatalf("unexpected audio %d", sound)
				}
				events = append(events, event)
				if sound == 878 {
					if mode == 19 {
						v.InvFirstItem = second
					}
					if mode == 21 || mode == 22 {
						w.ObjClass |= object.ClassMissile
						w.ObjSubClass = 0
					}
				}
			},
			ProjectileReflect: func(got, target *Object) {
				if got != w || target != v {
					t.Fatal("reflect args")
				}
				events = append(events, 9)
				if mode == 22 || mode == 25 {
					w.ObjClass &^= object.ClassMissile
				}
				if mode == 23 {
					w.ObjSubClass = 0x42
				}
				if mode == 25 {
					w.ObjSubClass = 0
				}
			},
			ClearOwner: func(got *Object) {
				if got != w {
					t.Fatal("clear owner")
				}
				events = append(events, 10)
				got.ObjOwner = nil
			},
			SetOwner: func(owner, got *Object) {
				if owner != v || got != w {
					t.Fatal("set owner")
				}
				events = append(events, 11)
				got.ObjOwner = owner
			},
			ChangeOwner: func(got, target *Object) {
				if got != w || target != v {
					t.Fatal("change owner")
				}
				events = append(events, 12)
			},
			BlockDamagePercent: func() float64 {
				events = append(events, 13)
				if mode == 20 {
					v.InvFirstItem = second
				}
				if mode == 26 {
					v.UpdateData = live
				}
				return .25
			},
			CanDamageBlockItem: func(got *Object) bool { return got == item || got == second },
			DamageBlockItem: func(got, owner, source, weapon *Object, amount float32, typ object.DamageType) bool {
				if owner != v || source != a || weapon != w || typ != object.DamageExplosion {
					t.Fatal("shield durability args")
				}
				events = append(events, 16)
				wearBits = math.Float32bits(amount)
				switch got {
				case item:
					selected = 1
				case second:
					selected = 2
				default:
					t.Fatal("unexpected shield")
				}
				if mode == 10 {
					got.ObjFlags |= object.FlagDestroyed
				}
				return true
			},
			PlayerSetState: func(got *Object, state PlayerState) bool {
				if got != v || state != PlayerState13 {
					t.Fatal("player state")
				}
				events = append(events, 17)
				got.UpdateDataPlayer().State = state
				return true
			},
			GodMode:          func() bool { return god },
			QuestMode:        func() bool { events = append(events, 14); return quest != 0 },
			QuestDamageScale: func() float32 { events = append(events, 15); return scales[quest] },
			DefaultDamage: func(target, source, weapon *Object, amount int32, typ object.DamageType) bool {
				if target != v || source != a || weapon != w || typ != object.DamageExplosion {
					t.Fatal("default args")
				}
				events = append(events, 20)
				resultDamage, calls = uint32(amount), calls+1
				return true
			},
			Unsupported: func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) {
				t.Fatalf("p=%v rejected %s", p, reason)
			},
		}
		r.Melee.MonsterPopBlockAction = func(got *Object) {
			if got != v {
				t.Fatal("pop action")
			}
			events = append(events, 18)
			got.UpdateDataMonster().AIStack[0].Action = 4
		}
		if mode == 17 || mode == 18 || mode == 24 {
			armor := damageMeleeArmorFixture4E17B0(v, .5, .25)
			armor.UpdateData = unsafe.Pointer(&WeaponArmorUpdateData{})
			r.ItemArmorValue = func(got *Object) float32 {
				if got != armor {
					t.Fatal("armor lookup")
				}
				return .5
			}
			r.CanDamageArmor = func(got *Object) bool { return got == armor }
			r.DamageArmor = func(got, source, weapon *Object, amount int32, typ object.DamageType) bool {
				if got != armor || source != a || weapon != w || typ != object.DamageExplosion {
					t.Fatal("armor wear args")
				}
				events = append(events, 19)
				remaining = uint32(amount)
				m, k := marker(entry)
				if mode == 17 {
					*m = 0
					quest = 2
				}
				if mode == 18 {
					*m, *k = 88, 89
				}
				if mode == 24 {
					god = true
				}
				return true
			}
		}
		handled, result := PlayerDamageNative4E17B0(v, a, w, damages[di], object.DamageExplosion, r)
		if !handled {
			t.Fatalf("p=%v unhandled", p)
		}
		fields := func(ptr unsafe.Pointer) (uint32, uint32, uint32, uint32) {
			if player {
				u := (*PlayerUpdateData)(ptr)
				return u.Field76, u.Field75, u.Field21, u.Field57
			}
			u := (*MonsterUpdateData)(ptr)
			return u.Field547, u.Field546, u.Field1, u.Field518
		}
		cm, ck, cc, _ := fields(entry)
		lm, lk, lc, la := fields(v.UpdateData)
		var state uint32
		if player {
			state = uint32(v.UpdateDataPlayer().State)
		} else {
			state = uint32(v.MonsterActionGet50A020())
		}
		var resultInt, ownerInt uint64
		if result {
			resultInt = 1
		}
		if w.ObjOwner == v {
			ownerInt = 1
		}
		row := make([]uint64, 0, 32)
		for _, value := range p {
			row = append(row, uint64(value))
		}
		row = append(row, resultInt, uint64(cm), uint64(ck), uint64(cc), uint64(lc), uint64(lm), uint64(lk), uint64(la), uint64(state), uint64(resultDamage), uint64(calls), uint64(wearBits), uint64(selected), ownerInt, uint64(remaining))
		row = append(row, events...)
		if err := json.NewEncoder(&transcript).Encode(row); err != nil {
			t.Fatal(err)
		}
		rows++
	}
	for si := range 2 {
		for ti := range 2 {
			for wi := range 3 {
				for ai0 := range 4 {
					for ci := range 4 {
						for di := range 6 {
							for qi := range 3 {
								for gi := range 2 {
									run([9]int{si, ti, wi, ai0, ci, di, qi, gi, 0})
								}
							}
						}
					}
				}
			}
		}
	}
	for mode := 1; mode <= 26; mode++ {
		for si := range 2 {
			for ti := range 2 {
				if mode == 11 || (mode == 14 || mode == 23 || mode == 25) && ti == 0 {
					continue
				}
				for wi := range 3 {
					run([9]int{si, ti, wi, 2, 2, 4, 0, 0, mode})
				}
			}
		}
	}
	const nearestControl = "3a7b83ab9d810c8a763452e0db49d8eda01edf8bd2285c5d833bbed06f74c2ac"
	expected, expectedBytes := originalArmedExplosionStartupSeal4E17B0(t, nearestControl)
	if sum := fmt.Sprintf("%x", sha256.Sum256(transcript.Bytes())); rows != 7194 || sum != expected || transcript.Len() != expectedBytes {
		t.Fatalf("unchanged original prefix transcript: rows=%d bytes=%d sha256=%s want=%s", rows, transcript.Len(), sum, expected)
	}
	t.Logf("unchanged original armed EXPLOSION prefix: rows=%d bytes=%d sha256=%s", rows, transcript.Len(), expected)
}
