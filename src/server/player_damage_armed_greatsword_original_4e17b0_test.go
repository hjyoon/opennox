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

// Independent native fixture for the unchanged PE32 callback-boundary replay.
// External callbacks change the incoming weapon to MISSILE after admission.
// Root/cached-state/action-readiness/inventory/carry calculations are not
// supplied by callbacks. The DefaultDamage hook records arguments, not HP.
func armedGreatSwordFixtureR84(t *testing.T, p [6]int) (
	v, a, w *Object, r PlayerDamageRuntime4E17B0, events *[]uint64,
	collect func(bool) []uint64,
) {
	t.Helper()
	si, ti, st, mode, sub := p[0], p[1], p[2], p[3], p[4]
	player := ti == 1
	states := [][]uint32{{0, 1, 4, 23, 25, 26, 27, 3, 21}, {13, 18, 19, 20, 0, 1, 16, 21}}
	v, a = damageMeleeUnitFixture4E17B0(t, player), damageMeleeUnitFixture4E17B0(t, si == 1)
	v.ObjFlags, a.ObjFlags = object.Flags(16777732), object.Flags(16777732)
	v.ObjClass, v.ObjSubClass = object.Class(2097154), 69650
	a.ObjClass, a.ObjSubClass, a.TypeInd = object.Class(2097154), 69778, 1343
	if player {
		v.ObjClass, v.ObjSubClass = object.Class(2621444), 0
	}
	if si == 1 {
		a.ObjClass, a.ObjSubClass, a.TypeInd = object.Class(2621444), 0, 713
	}
	w = &Object{TypeInd: 1174, ObjClass: object.Class(85458944), ObjSubClass: 256, ObjFlags: object.Flags(17301776),
		InitData: unsafe.Pointer(&ModifierInitData{}), PosVec: types.Ptf(19, 11), PrevPos: types.Ptf(17, 9), ObjOwner: a}
	entry, live := v.UpdateData, damageMeleeUnitFixture4E17B0(t, player).UpdateData
	setup := func(ptr unsafe.Pointer, armor, carry float32, marker, kind uint32) {
		if player {
			u := (*PlayerUpdateData)(ptr)
			u.Field57, u.Field21, u.Field76, u.Field75 = math.Float32bits(armor), math.Float32bits(carry), marker, kind
		} else {
			u := (*MonsterUpdateData)(ptr)
			u.Field518, u.Field1, u.Field547, u.Field546 = math.Float32bits(armor), math.Float32bits(carry), marker, kind
			u.AIStackInd = 0
		}
	}
	setup(entry, .5, .25, 99, 77)
	setup(live, .25, .5, 33, 44)
	setAction := func(ptr unsafe.Pointer, value uint32) {
		(*MonsterUpdateData)(ptr).AIStack[0].Action = value
	}
	weaponFlags := func(value uint32) {
		if player {
			(*PlayerUpdateData)(entry).Player.WeaponEquip = value
		} else {
			(*MonsterUpdateData)(entry).WeaponEquipFlags = value
		}
	}
	weaponFlags(0x400)
	if player {
		(*PlayerUpdateData)(entry).State = PlayerState(states[ti][st])
		(*PlayerUpdateData)(live).State = PlayerState0
		if mode == 14 {
			(*PlayerUpdateData)(entry).Player.Field3680 = 2
			(*PlayerUpdateData)(entry).Player.CameraFollowObj = a
		}
	} else {
		setAction(entry, states[ti][st])
		setAction(live, 3)
	}
	item, second := &Object{ObjSubClass: 0x400, ObjFlags: object.FlagEquipped, HealthData: &HealthData{}},
		&Object{ObjSubClass: 0x400, ObjFlags: object.FlagEquipped, HealthData: &HealthData{}}
	item.InvNextItem, v.InvFirstItem = second, item
	events = new([]uint64)
	appendEvent := func(value uint64) { *events = append(*events, value) }
	resultDamage, calls, wearBits, selected := uint32(0xfeedface), uint32(0), uint32(0), uint32(0)
	r = PlayerDamageRuntime4E17B0{
		Frame:    func() uint32 { return 1400 },
		CoopMode: func() bool { appendEvent(1); return false },
		ObserveClear: func(got *Object) {
			if got != v {
				t.Fatal("observer identity")
			}
			appendEvent(3)
			w.ObjClass |= object.ClassMissile
			w.ObjSubClass = object.SubClass(sub)
		},
		BlockSourceExcluded: func(got *Object) bool {
			if got != w {
				t.Fatal("weapon exclusion identity")
			}
			return false
		},
		BlockDirection: func(got *Object, pos types.Pointf) bool {
			if got != v || pos != types.Ptf(17, 9) {
				t.Fatal("cached previous-position identity")
			}
			appendEvent(4)
			if mode != 2 {
				w.ObjClass |= object.ClassMissile
				w.ObjSubClass = object.SubClass(sub)
			}
			if mode == 5 {
				weaponFlags(0)
			}
			if mode == 12 {
				v.UpdateData = live
			}
			if mode == 13 {
				if player {
					(*PlayerUpdateData)(entry).State = PlayerState13
				} else {
					setAction(entry, 4)
				}
			}
			return mode != 1
		},
		ProjectileReflect: func(got, defender *Object) {
			if got != w || defender != v {
				t.Fatal("reflection identity")
			}
			appendEvent(9)
			if mode == 3 {
				w.ObjClass &^= object.ClassMissile
			}
			if mode == 4 {
				w.ObjSubClass = 2
			}
		},
		ClearOwner: func(got *Object) {
			if got != w {
				t.Fatal("clear-owner identity")
			}
			appendEvent(10)
			w.ObjOwner = nil
		},
		SetOwner: func(owner, got *Object) {
			if owner != v || got != w {
				t.Fatal("set-owner identity")
			}
			appendEvent(11)
			w.ObjOwner = v
		},
		ChangeOwner: func(*Object, *Object) { t.Fatal("GreatSword must not call Reflect Shield owner-change") },
		Audio: func(sound int, got *Object) {
			if sound != 890 || got != v {
				t.Fatal("GreatSword sound identity")
			}
			appendEvent(7)
			if mode == 6 {
				v.InvFirstItem = second
			}
		},
		PlayerSetState: func(got *Object, state PlayerState) bool {
			if got != v || state != PlayerState13 && state != PlayerState19 {
				t.Fatal("GreatSword player state identity")
			}
			if state == PlayerState13 {
				appendEvent(17)
			} else {
				appendEvent(22)
			}
			got.UpdateDataPlayer().State = state
			if mode == 8 {
				v.InvFirstItem = second
			}
			return true
		},
		BlockDamagePercent: func() float64 {
			appendEvent(13)
			if mode == 9 {
				v.InvFirstItem = second
			}
			if mode == 10 {
				item.ObjFlags = 0
			}
			return .20000000000000007
		},
		GodMode:   func() bool { return false },
		QuestMode: func() bool { appendEvent(14); return false },
		DefaultDamage: func(target, source, weapon *Object, amount int32, typ object.DamageType) bool {
			if target != v || source != a || weapon != w || typ != object.DamageExplosion {
				t.Fatal("raw type-7 tail identity")
			}
			appendEvent(20)
			resultDamage, calls = uint32(amount), calls+1
			return true
		},
		Unsupported: func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) {
			t.Fatalf("p=%v rejected %s", p, reason)
		},
	}
	r.Melee = PlayerDamageMeleeRuntime4E17B0{
		RandomInt: func(min, max int) int {
			if min != 18 || max != 20 {
				t.Fatal("GreatSword RNG bounds")
			}
			appendEvent(21)
			if mode == 7 {
				v.InvFirstItem = second
			}
			return 19
		},
		MonsterBlockAction: func(got *Object) {
			if got != v {
				t.Fatal("GreatSword NPC action identity")
			}
			appendEvent(23)
			setAction(v.UpdateData, 23)
			if mode == 7 || mode == 8 {
				v.InvFirstItem = second
			}
		},
		MonsterPopBlockAction: func(got *Object) {
			if got != v {
				t.Fatal("GreatSword NPC pop identity")
			}
			appendEvent(18)
			setAction(v.UpdateData, 4)
		},
		CanDamageBlockWeapon: func(got *Object) bool { return got == item || got == second },
		DamageBlockWeapon: func(got, owner, source, attack *Object, amount float32, typ object.DamageType) bool {
			if owner != v || source != a || attack != w || typ != object.DamageExplosion {
				t.Fatal("GreatSword durability identity/type")
			}
			appendEvent(16)
			wearBits = math.Float32bits(amount)
			switch got {
			case nil:
				selected = 0
			case item:
				selected = 1
			case second:
				selected = 2
			default:
				t.Fatal("GreatSword selected inventory item")
			}
			if mode == 11 {
				got.ObjFlags |= object.FlagDestroyed
			}
			return true
		},
	}
	fields := func(ptr unsafe.Pointer) (uint32, uint32, uint32, uint32) {
		if player {
			u := (*PlayerUpdateData)(ptr)
			return u.Field76, u.Field75, u.Field21, u.Field57
		}
		u := (*MonsterUpdateData)(ptr)
		return u.Field547, u.Field546, u.Field1, u.Field518
	}
	collect = func(result bool) []uint64 {
		cm, ck, cc, _ := fields(entry)
		lm, lk, lc, la := fields(v.UpdateData)
		var state uint32
		if player {
			state = uint32(v.UpdateDataPlayer().State)
		} else {
			state = uint32(v.MonsterActionGet50A020())
		}
		row := make([]uint64, 0, 48)
		for _, value := range p {
			row = append(row, uint64(value))
		}
		var ret, owner, missile uint64
		if result {
			ret = 1
		}
		if w.ObjOwner == v {
			owner = 1
		}
		if w.Class().Has(object.ClassMissile) {
			missile = 1
		}
		row = append(row, ret, uint64(cm), uint64(ck), uint64(cc), uint64(lc), uint64(lm), uint64(lk), uint64(la),
			uint64(state), uint64(resultDamage), uint64(calls), uint64(wearBits), uint64(selected), owner, missile,
			uint64(w.ObjSubClass), uint64(item.ObjFlags), uint64(second.ObjFlags))
		return append(row, *events...)
	}
	return
}

// Golden digest is from 5,904 unchanged-original executions, sealed before
// any production edit. Actual original action-readiness and inventory-loop
// bytes execute; missing-item faults are negative cases outside this digest.
func TestPlayerDamageArmedGreatSword4E17B0OriginalTranscript(t *testing.T) {
	var transcript bytes.Buffer
	damages := []int32{-9, 0, 1, 3, 9, 61}
	rows := 0
	for si := range 2 {
		for ti := range 2 {
			count, modes := 9, 14
			if ti == 1 {
				count, modes = 8, 15
			}
			for st := range count {
				for mode := range modes {
					for _, sub := range []int{0, 2} {
						for di, damage := range damages {
							p := [6]int{si, ti, st, mode, sub, di}
							v, a, w, r, _, collect := armedGreatSwordFixtureR84(t, p)
							h, result := PlayerDamageNative4E17B0(v, a, w, damage, object.DamageExplosion, r)
							if !h {
								t.Fatalf("p=%v unhandled", p)
							}
							if err := json.NewEncoder(&transcript).Encode(collect(result)); err != nil {
								t.Fatal(err)
							}
							rows++
						}
					}
				}
			}
		}
	}
	const expected = "6377f57288267ea72e772d8d97ffb84045bed1943614408828682ee6e9924e63"
	if sum := fmt.Sprintf("%x", sha256.Sum256(transcript.Bytes())); rows != 5904 || sum != expected {
		t.Fatalf("unchanged original GreatSword transcript: rows=%d bytes=%d sha256=%s want=%s", rows, transcript.Len(), sum, expected)
	}
	t.Logf("unchanged original armed GreatSword callback boundary: rows=%d bytes=%d sha256=%s", rows, transcript.Len(), expected)
}

func TestPlayerDamageArmedGreatSword4E17B0NilItemOriginalFault(t *testing.T) {
	for ti := range 2 {
		t.Run(fmt.Sprintf("player-%d", ti), func(t *testing.T) {
			v, a, w, r, events, collect := armedGreatSwordFixtureR84(t, [6]int{0, ti, 0, 0, 0, 4})
			v.InvFirstItem = nil
			reason := ""
			r.Unsupported = func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = why }
			h, result := PlayerDamageNative4E17B0(v, a, w, 9, object.DamageExplosion, r)
			want := []uint64{1, 4, 9, 10, 11, 7, 23, 13, 16}
			if ti == 1 {
				want = []uint64{1, 4, 9, 10, 11, 7, 21, 22, 13, 16}
			}
			row := collect(result)
			if h || result || reason != "original armed EXPLOSION GreatSword nil-item fault" || !slices.Equal(*events, want) ||
				row[7] != 1 || row[8] != 1174 || row[9] != uint64(math.Float32bits(.25)) || row[16] != 0 || row[18] != 0 || row[19] != 1 {
				t.Fatalf("safe original fault: h/result=%t/%t reason=%q events=%v row=%v", h, result, reason, *events, row)
			}
		})
	}
}
