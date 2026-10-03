package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestPlayerDamageItems4E2180OriginalReadAndCallbackOrder(t *testing.T) {
	type record struct {
		name         string
		class, flags uint32
		subclass     uint8
		armor        float32
		coefficient  float64
		next, first  *record
	}
	target := &record{name: "target", class: 4, armor: 0.5}
	source, effective := &record{name: "source"}, &record{name: "effective"}
	first := &record{name: "first", class: 0x2000000, flags: 0x100, coefficient: 0.25}
	stale := &record{name: "stale", class: 0x2000000, flags: 0x100}
	weapon := &record{name: "weapon", class: 0x1000000, flags: 0x100}
	later := &record{name: "later", class: 0x2000000, coefficient: 0.125}
	unequipped := &record{name: "unequipped", class: 0x2000000}
	first.next, stale.next, weapon.next, later.next = stale, weapon, later, unequipped
	target.first = first
	var events []string
	appendEvent := func(event string, rec *record) { events = append(events, event+":"+rec.name) }
	hooks := playerDamageItemsHooks4E2180[*record]{
		loadClass: func(rec *record) uint32 { appendEvent("class", rec); return rec.class },
		loadSubclassLow: func(rec *record) uint8 {
			t.Fatal("Player branch read Monster subclass")
			return rec.subclass
		},
		loadPlayerArmor: func(rec *record) float32 { appendEvent("armor", rec); return rec.armor },
		loadMonsterArmor: func(*record) float32 {
			t.Fatal("Player branch read Monster armor")
			return 0
		},
		loadFirstItem: func(rec *record) *record { appendEvent("first", rec); return rec.first },
		loadFlags:     func(rec *record) uint32 { appendEvent("flags", rec); return rec.flags },
		itemArmorValue: func(rec *record) float64 {
			appendEvent("lookup", rec)
			if rec == first {
				// The denominator is already cached. No target class/update
				// reread or inventory snapshot may move across this callback.
				target.class, target.armor, target.first = 2, 1, stale
			}
			return rec.coefficient
		},
		equipDamage: func(rec, owner, gotSource, gotEffective *record, amount float32, typ int32) {
			appendEvent("wear", rec)
			if owner != target || gotSource != source || gotEffective != effective || typ != 3 {
				t.Fatal("EquipDamage argument order")
			}
			want := float32(4)
			if rec == later {
				want = 2
			}
			if amount != want {
				t.Fatalf("%s amount=%g, want %g from cached armor", rec.name, amount, want)
			}
			if rec == first {
				first.next = weapon // next is loaded only after wear returns
				later.flags = 0x100 // later flags are not planned at entry
			}
		},
		loadNextItem: func(rec *record) *record { appendEvent("next", rec); return rec.next },
	}
	playerDamageItems4E2180(target, source, effective, 8, 3, hooks)
	want := []string{
		"class:target", "armor:target", "first:target",
		"class:first", "flags:first", "lookup:first", "wear:first", "next:first",
		"class:weapon", "next:weapon",
		"class:later", "flags:later", "lookup:later", "wear:later", "next:later",
		"class:unequipped", "flags:unequipped", "next:unequipped",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events=%v, want %v", events, want)
	}
}

func TestPlayerDamageItems4E2180ClassSelectionAndFaultBoundaries(t *testing.T) {
	for _, tc := range []struct {
		name       string
		class      object.Class
		subclass   object.SubClass
		player     bool
		qualifying bool
	}{
		{"Player", object.ClassPlayer, 0, true, true},
		{"Player before Monster", object.ClassPlayer | object.ClassMonster, 0, true, true},
		{"NPC", object.ClassMonster, 0x10, false, true},
		{"NPC low subclass byte", object.ClassMonster, 0x10010, false, true},
		{"ordinary Monster", object.ClassMonster, 0x10000, false, false},
		{"other class", object.ClassArmor, 0x10, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			target := &Object{ObjClass: tc.class, ObjSubClass: tc.subclass}
			item := &Object{ObjClass: object.ClassArmor, ObjFlags: object.FlagEquipped}
			target.InvFirstItem = item
			if tc.qualifying {
				if tc.player {
					target.UpdateData = unsafe.Pointer(&PlayerUpdateData{Field57: math.Float32bits(0.5)})
				} else {
					target.UpdateData = unsafe.Pointer(&MonsterUpdateData{Field518: math.Float32bits(0.25)})
				}
			}
			calls := 0
			PlayerDamageItems4E2180(target, nil, nil, 4, object.DamageFlame, PlayerDamageItemsRuntime4E2180{
				ItemArmorValue: func(*Object) float64 { return 0.25 },
				EquipDamage: func(gotItem, owner, source, effective *Object, amount float32, typ object.DamageType) {
					calls++
					want := float32(4)
					if tc.player {
						want = 2
					}
					if gotItem != item || owner != target || source != nil || effective != nil || amount != want || typ != object.DamageFlame {
						t.Fatal("wrong selected armor layout or raw arguments")
					}
				},
			})
			if (calls != 0) != tc.qualifying {
				t.Fatalf("calls=%d, qualifying=%t", calls, tc.qualifying)
			}
		})
	}
	for _, target := range []*Object{nil, {ObjClass: object.ClassPlayer}, {ObjClass: object.ClassMonster, ObjSubClass: 0x10}} {
		t.Run(fmt.Sprintf("original unguarded %p", target), func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("target/selected update was silently accepted")
				}
			}()
			PlayerDamageItems4E2180(target, nil, nil, 0, object.DamageImpact, PlayerDamageItemsRuntime4E2180{})
		})
	}
	// Empty inventory does not suppress the update read, but needs no callee.
	target := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(&PlayerUpdateData{})}
	PlayerDamageItems4E2180(target, nil, nil, 0, object.DamageImpact, PlayerDamageItemsRuntime4E2180{})
}

func TestPlayerDamageItems4E2180SignedAmountsAndZeroArmor(t *testing.T) {
	for _, tc := range []struct {
		name        string
		armor       float32
		coefficient float64
		damage      int32
		want        float32
		nan         bool
	}{
		{name: "positive", armor: 0.5, coefficient: 0.25, damage: 7, want: 3.5},
		{name: "zero is still dispatched", armor: 0.5, coefficient: 0.25, damage: 0, want: 0},
		{name: "negative", armor: 0.5, coefficient: 0.25, damage: -7, want: -3.5},
		{name: "signed maximum", armor: 0.5, coefficient: 0.25, damage: math.MaxInt32, want: 1073741824},
		{name: "signed minimum", armor: 0.5, coefficient: 0.25, damage: math.MinInt32, want: -1073741824},
		{name: "spill after multiply", armor: 0.5, coefficient: 1 + math.Ldexp(1, -24), damage: 3, want: math.Float32frombits(0x40c00001)},
		{name: "zero armor positive", coefficient: 0.25, damage: 1, want: float32(math.Inf(1))},
		{name: "zero armor negative", coefficient: 0.25, damage: -1, want: float32(math.Inf(-1))},
		{name: "zero armor and damage", coefficient: 0.25, nan: true},
		{name: "zero ratio", armor: 0, coefficient: 0, damage: 4, nan: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			target := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(&PlayerUpdateData{Field57: math.Float32bits(tc.armor)})}
			// HealthData is deliberately nil: 004E2180 still looks up the
			// coefficient and calls 004E16D0, which owns the health guard.
			item := &Object{ObjClass: object.ClassArmor, ObjFlags: object.FlagEquipped}
			target.InvFirstItem = item
			var got float32
			calls := 0
			PlayerDamageItems4E2180(target, nil, nil, tc.damage, object.DamageType(0x11223344), PlayerDamageItemsRuntime4E2180{
				ItemArmorValue: func(*Object) float64 { return tc.coefficient },
				EquipDamage: func(_, _, _, _ *Object, amount float32, typ object.DamageType) {
					calls++
					got = amount
					if typ != object.DamageType(0x11223344) {
						t.Fatalf("raw type=%#x", typ)
					}
				},
			})
			if calls != 1 || (tc.nan && !math.IsNaN(float64(got))) || (!tc.nan && math.Float32bits(got) != math.Float32bits(tc.want)) {
				t.Fatalf("calls=%d portion=%g/%#x, want %g/%#x nan=%t", calls, got, math.Float32bits(got), tc.want, math.Float32bits(tc.want), tc.nan)
			}
		})
	}
}

func TestPlayerDamageItems4E2180OrderedEquipDamageUsesLiveRecords(t *testing.T) {
	ud := &PlayerUpdateData{Field57: math.Float32bits(0.5)}
	target := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(ud)}
	source, effective := &Object{}, &Object{}
	first, firstHP, oldCarry, firstInit := durabilityItem4E1560(object.ClassArmor, 0, 10)
	stale, _, _, _ := durabilityItem4E1560(object.ClassArmor, 0, 10)
	later, laterHP, laterCarry, _ := durabilityItem4E1560(object.ClassArmor, 0, 20)
	noHealth := &Object{ObjClass: object.ClassArmor, ObjFlags: object.FlagEquipped}
	first.ObjFlags, stale.ObjFlags = object.FlagEquipped, object.FlagEquipped
	first.InvNextItem, stale.InvNextItem, later.InvNextItem = stale, later, noHealth
	target.InvFirstItem = first
	modifier := &ModifierEff{Defend76: ModifierEffFnc{Fnc: unsafe.Pointer(new(byte))}}
	firstInit.Modifiers[1] = modifier
	liveCarry := &WeaponArmorUpdateData{Field0: math.Float32bits(0.25)}
	liveHP := &HealthData{Cur: 6, Max: 10}
	var events []string
	durability := ItemDurabilityDamageRuntime4E1560{
		ApplyDefend: func(got *ModifierEff, item, owner, gotEffective, gotSource *Object, value *float32) bool {
			events = append(events, "defend:first")
			if got != modifier || item != first || owner != target || gotEffective != effective || gotSource != source || *value != 4 {
				t.Fatal("original defend argument order/amount")
			}
			*value *= 0.25
			first.UpdateData = unsafe.Pointer(liveCarry)
			ud.Field57 = math.Float32bits(1) // denominator must not be recaptured
			return true
		},
		Damage: func(item, gotSource, gotEffective *Object, damage int32, typ object.DamageType) bool {
			if gotSource != source || gotEffective != effective || typ != object.DamageImpale {
				t.Fatal("damage argument order")
			}
			if item == first {
				events = append(events, "damage:first")
				if damage != 1 || liveCarry.Field0 != math.Float32bits(0.25) || oldCarry.Field0 != 0 {
					t.Fatal("durability did not reload carry after modifier")
				}
				first.HealthData = liveHP
				first.InvNextItem = later
				later.ObjFlags = object.FlagEquipped
			} else if item == later {
				events = append(events, "damage:later")
				if damage != 2 || laterCarry.Field0 != 0 {
					t.Fatal("later item did not use captured denominator")
				}
				laterHP.Cur -= uint16(damage)
			} else {
				t.Fatal("damage reached stale or health-less item")
			}
			return false // EquipDamage ignores the original damage callback's result
		},
		ReportHealth: func(owner, item *Object, before, after uint16) {
			if owner != target {
				t.Fatal("report owner")
			}
			if item == first && before == 10 && after == 6 {
				events = append(events, "report:first")
			} else if item == later && before == 20 && after == 18 {
				events = append(events, "report:later")
			} else {
				t.Fatalf("report did not reload health: %p %d -> %d", item, before, after)
			}
		},
	}
	PlayerDamageItems4E2180(target, source, effective, 8, object.DamageImpale, PlayerDamageItemsRuntime4E2180{
		ItemArmorValue: func(item *Object) float64 {
			if item == first {
				events = append(events, "lookup:first")
				return 0.25
			}
			if item == later {
				events = append(events, "lookup:later")
				return 0.125
			}
			if item != noHealth {
				t.Fatal("lookup reached stale item")
			}
			events = append(events, "lookup:no-health")
			return 0.25
		},
		EquipDamage: func(item, owner, source, effective *Object, amount float32, typ object.DamageType) {
			if !EquipDamageNative4E16D0(item, owner, source, effective, amount, typ, durability) {
				t.Fatal("real EquipDamage rejected the fixture")
			}
		},
	})
	want := []string{"lookup:first", "defend:first", "damage:first", "report:first", "lookup:later", "damage:later", "report:later", "lookup:no-health"}
	if !reflect.DeepEqual(events, want) || firstHP.Cur != 10 || laterHP.Cur != 18 || oldCarry.Field0 != 0 {
		t.Fatalf("events=%v, want %v; oldHP=%d laterHP=%d", events, want, firstHP.Cur, laterHP.Cur)
	}
}

func TestPlayerDamageItems4E2180ZeroDamageFlushesExistingCarry(t *testing.T) {
	target := &Object{ObjClass: object.ClassMonster, ObjSubClass: 0x10, UpdateData: unsafe.Pointer(&MonsterUpdateData{Field518: math.Float32bits(0.5)})}
	item, hp, carry, init := durabilityItem4E1560(object.ClassArmor, 0, 4)
	item.ObjFlags = object.FlagEquipped
	target.InvFirstItem = item
	init.Modifiers[1] = &ModifierEff{Defend76: ModifierEffFnc{Fnc: unsafe.Pointer(new(byte))}}
	carry.Field0 = math.Float32bits(0.75)
	defends, wears, reports := 0, 0, 0
	runtime := ItemDurabilityDamageRuntime4E1560{
		ApplyDefend: func(_ *ModifierEff, _, _, _, _ *Object, amount *float32) bool {
			defends++
			if *amount != 0 {
				t.Fatal("zero portion was lost")
			}
			return true
		},
		Damage: func(got, _, _ *Object, damage int32, typ object.DamageType) bool {
			wears++
			if got != item || damage != 1 || typ != object.DamageImpale || carry.Field0 != math.Float32bits(-0.25) {
				t.Fatal("zero incoming wear did not flush carry before damage")
			}
			hp.Cur--
			return true
		},
		ReportHealth: func(_, _ *Object, _, _ uint16) { reports++ },
	}
	PlayerDamageItems4E2180(target, nil, nil, 0, object.DamageImpale, PlayerDamageItemsRuntime4E2180{
		ItemArmorValue: func(*Object) float64 { return 0.25 },
		EquipDamage: func(item, owner, source, effective *Object, amount float32, typ object.DamageType) {
			if !EquipDamageNative4E16D0(item, owner, source, effective, amount, typ, runtime) {
				t.Fatal("EquipDamage rejected zero")
			}
		},
	})
	if defends != 1 || wears != 1 || reports != 0 || hp.Cur != 3 {
		t.Fatalf("defends=%d wears=%d reports=%d HP=%d", defends, wears, reports, hp.Cur)
	}
}
