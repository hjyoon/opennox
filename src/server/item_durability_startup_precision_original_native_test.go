package server

import (
	"bytes"
	"encoding/json"
	"math"
	"os"
	"reflect"
	"sync"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

type itemDurabilityOriginalReport struct {
	Before uint16 `json:"before"`
	After  uint16 `json:"after"`
}

type itemDurabilityOriginalCase struct {
	Name             string                         `json:"name"`
	WeaponRules      bool                           `json:"weapon_rules"`
	Amount           uint32                         `json:"amount_bits32"`
	Carry            uint32                         `json:"carry_bits32"`
	NewCarry         uint32                         `json:"new_carry_bits32"`
	ItemClass        object.Class                   `json:"item_class"`
	ItemSubclass     object.SubClass                `json:"item_subclass"`
	OwnerClass       object.Class                   `json:"owner_class"`
	DefendSwap       bool                           `json:"defend_swap"`
	NilItem          bool                           `json:"nil_item"`
	NilHealth        bool                           `json:"nil_health"`
	HealthAfter      uint16                         `json:"health_after"`
	Type             object.DamageType              `json:"damage_type"`
	ExpectedOldCarry uint32                         `json:"expected_old_carry_bits32"`
	ExpectedNewCarry uint32                         `json:"expected_new_carry_bits32"`
	ExpectedHealth   uint16                         `json:"expected_health"`
	ExpectedSwap     bool                           `json:"expected_update_pointer_replaced"`
	ExpectedDamage   []int32                        `json:"expected_damage"`
	ExpectedReports  []itemDurabilityOriginalReport `json:"expected_reports"`
}

func itemDurabilityOriginalFrame(p unsafe.Pointer, n uintptr) []byte {
	return unsafe.Slice((*byte)(p), int(n))
}

func observeItemDurabilityOriginalNative(t *testing.T, c itemDurabilityOriginalCase) map[string]any {
	t.Helper()
	var frees []func()
	defer func() {
		for i := len(frees) - 1; i >= 0; i-- {
			frees[i]()
		}
	}()
	item, f := alloc.New(Object{})
	frees = append(frees, f)
	old, f := alloc.New(WeaponArmorUpdateData{})
	frees = append(frees, f)
	*old = WeaponArmorUpdateData{Field0: c.Carry}
	live, f := alloc.New(WeaponArmorUpdateData{})
	frees = append(frees, f)
	*live = WeaponArmorUpdateData{Field0: c.NewCarry}
	init, f := alloc.New(ModifierInitData{})
	frees = append(frees, f)
	*init = ModifierInitData{}
	hp, f := alloc.New(HealthData{})
	frees = append(frees, f)
	*hp = HealthData{Cur: 100, Field2: 100, Max: 100}
	owner, f := alloc.New(Object{})
	frees = append(frees, f)
	ownerUpdate, f := alloc.New(PlayerUpdateData{})
	frees = append(frees, f)
	ownerInfo, f := alloc.New(Player{})
	frees = append(frees, f)
	*ownerInfo = Player{PlayerInd: 8}
	*ownerUpdate = PlayerUpdateData{Player: ownerInfo}
	*owner = Object{ObjClass: c.OwnerClass, UpdateData: unsafe.Pointer(ownerUpdate)}
	source, f := alloc.New(Object{})
	frees = append(frees, f)
	*source = Object{}
	effective, f := alloc.New(Object{})
	frees = append(frees, f)
	*effective = Object{}
	modifier, f := alloc.New(ModifierEff{})
	frees = append(frees, f)
	callback, f := alloc.New(byte(0))
	frees = append(frees, f)
	*modifier = ModifierEff{Defend76: ModifierEffFnc{Fnc: unsafe.Pointer(callback)}}
	if c.DefendSwap {
		init.Modifiers[1] = modifier
	}
	*item = Object{ObjClass: c.ItemClass, ObjSubClass: c.ItemSubclass, HealthData: hp, UpdateData: unsafe.Pointer(old), InitData: unsafe.Pointer(init), Damage: unsafe.Pointer(callback)}
	if c.NilHealth {
		item.HealthData = nil
	}
	frames := [][]byte{
		itemDurabilityOriginalFrame(unsafe.Pointer(item), unsafe.Sizeof(*item)),
		itemDurabilityOriginalFrame(unsafe.Pointer(old), unsafe.Sizeof(*old)),
		itemDurabilityOriginalFrame(unsafe.Pointer(live), unsafe.Sizeof(*live)),
		itemDurabilityOriginalFrame(unsafe.Pointer(hp), unsafe.Sizeof(*hp)),
		itemDurabilityOriginalFrame(unsafe.Pointer(init), unsafe.Sizeof(*init)),
		itemDurabilityOriginalFrame(unsafe.Pointer(owner), unsafe.Sizeof(*owner)),
		itemDurabilityOriginalFrame(unsafe.Pointer(ownerUpdate), unsafe.Sizeof(*ownerUpdate)),
		itemDurabilityOriginalFrame(unsafe.Pointer(ownerInfo), unsafe.Sizeof(*ownerInfo)),
		itemDurabilityOriginalFrame(unsafe.Pointer(source), unsafe.Sizeof(*source)),
		itemDurabilityOriginalFrame(unsafe.Pointer(effective), unsafe.Sizeof(*effective)),
		itemDurabilityOriginalFrame(unsafe.Pointer(modifier), unsafe.Sizeof(*modifier)),
		itemDurabilityOriginalFrame(unsafe.Pointer(callback), unsafe.Sizeof(*callback)),
	}
	before := make([][]byte, len(frames))
	for i, frame := range frames {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(&frame[0])) <= 0xffffffff {
			t.Fatal("expected C-owned native input frames above4GiB")
		}
		before[i] = append([]byte(nil), frame...)
	}
	damage := []int32{}
	reports := []itemDurabilityOriginalReport{}
	unsupported := []string{}
	defends := 0
	r := ItemDurabilityDamageRuntime4E1560{
		ApplyDefend: func(got *ModifierEff, v, by, eff, src *Object, amount *float32) bool {
			if got != modifier || v != item || by != owner || eff != effective || src != source || math.Float32bits(*amount) != c.Amount || item.UpdateData != unsafe.Pointer(old) {
				t.Fatal("changed Defend arguments or pre-callback update identity")
			}
			defends++
			item.UpdateData = unsafe.Pointer(live)
			return true // Declared integer pointer effect; no FP output substitution.
		},
		Damage: func(v, src, eff *Object, amount int32, typ object.DamageType) bool {
			if v != item || src != source || eff != effective || typ != c.Type || amount <= 0 {
				t.Fatal("changed item damage arguments")
			}
			damage = append(damage, amount)
			hp.Cur = c.HealthAfter
			return false // The original ignores this callback's return value.
		},
		ReportHealth: func(by, v *Object, was, now uint16) {
			if by != owner || v != item || ownerInfo.PlayerInd != 8 {
				t.Fatal("changed health report identities")
			}
			reports = append(reports, itemDurabilityOriginalReport{Before: was, After: now})
		},
		Unsupported: func(reason string, _, _, _, _ *Object, _ float32, _ object.DamageType) {
			unsupported = append(unsupported, reason)
		},
	}
	arg := item
	if c.NilItem {
		arg = nil
	}
	var handled bool
	if c.WeaponRules {
		handled = PlayerDamageWeaponNative4E1560(arg, owner, source, effective, math.Float32frombits(c.Amount), c.Type, r)
	} else {
		handled = EquipDamageNative4E16D0(arg, owner, source, effective, math.Float32frombits(c.Amount), c.Type, r)
	}
	expectedUpdate := unsafe.Pointer(old)
	if c.ExpectedSwap {
		expectedUpdate = unsafe.Pointer(live)
	}
	off := unsafe.Offsetof(item.UpdateData)
	copy(before[0][off:off+unsafe.Sizeof(item.UpdateData)], frames[0][off:off+unsafe.Sizeof(item.UpdateData)])
	copy(before[1][:4], frames[1][:4])
	copy(before[3][:2], frames[3][:2])
	preserved := true
	for i, frame := range frames {
		preserved = preserved && bytes.Equal(before[i], frame)
	}
	wantDefends := 0
	if c.ExpectedSwap {
		wantDefends = 1
	}
	match := handled && len(unsupported) == 0 && old.Field0 == c.ExpectedOldCarry && live.Field0 == c.ExpectedNewCarry && hp.Cur == c.ExpectedHealth && item.UpdateData == expectedUpdate && defends == wantDefends && reflect.DeepEqual(damage, c.ExpectedDamage) && reflect.DeepEqual(reports, c.ExpectedReports) && preserved
	return map[string]any{"name": c.Name, "accepted_native_adapter": handled, "damage": damage, "old_carry_bits32": old.Field0, "new_carry_bits32": live.Field0, "health": hp.Cur, "reports": reports, "defends": defends, "unsupported": unsupported, "all_native_nonoutput_frames_preserved": preserved, "matches_original": match}
}

func TestPlayerDamageItemsDurabilityPrecisionOriginalNative(t *testing.T) {
	b, err := os.ReadFile("testdata/item_durability_startup_precision_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []itemDurabilityOriginalCase `json:"cases"`
	}
	if err = json.Unmarshal(b, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Cases) != 600 {
		t.Fatal("incomplete independently executed original item durability contexts")
	}
	historical := itemDurabilityHistoricalCases(t)
	if len(historical) != 391 {
		t.Fatal("incomplete original historical numeric contexts")
	}
	fixture.Cases = append(fixture.Cases, historical...)
	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			got := observeItemDurabilityOriginalNative(t, c)
			if !got["matches_original"].(bool) {
				t.Fatalf("actual native item durability mismatch=%v original damage=%v oldCarry=%08x newCarry=%08x", got, c.ExpectedDamage, c.ExpectedOldCarry, c.ExpectedNewCarry)
			}
		})
	}
}

// These inputs are raw post-Defend amount and captured carry bytes. Expected
// results come from complete GAME.EXE 004E1560/004E16D0 executions, including
// its real 00419A70 converter. No native arithmetic computes the expectations.
var itemDurabilityHistoricalFixture struct {
	once  sync.Once
	cases []itemDurabilityOriginalCase
	err   error
}

func itemDurabilityHistoricalCases(t *testing.T) []itemDurabilityOriginalCase {
	t.Helper()
	itemDurabilityHistoricalFixture.once.Do(func() {
		b, err := os.ReadFile("testdata/item_durability_historical_precision_original.json")
		if err != nil {
			itemDurabilityHistoricalFixture.err = err
			return
		}
		var fixture struct {
			Cases []itemDurabilityOriginalCase `json:"cases"`
		}
		itemDurabilityHistoricalFixture.err = json.Unmarshal(b, &fixture)
		itemDurabilityHistoricalFixture.cases = fixture.Cases
	})
	if err := itemDurabilityHistoricalFixture.err; err != nil {
		t.Fatal(err)
	}
	return itemDurabilityHistoricalFixture.cases
}

func itemDurabilityHistoricalExpected(t *testing.T, amount, carry float32) (int32, uint32) {
	t.Helper()
	found := false
	var damage int32
	var remainder uint32
	for _, c := range itemDurabilityHistoricalCases(t) {
		if c.Amount != math.Float32bits(amount) || c.Carry != math.Float32bits(carry) {
			continue
		}
		var d int32
		if len(c.ExpectedDamage) != 0 {
			d = c.ExpectedDamage[0]
		}
		if found && (damage != d || remainder != c.ExpectedOldCarry) {
			t.Fatal("ambiguous original numeric boundary")
		}
		found, damage, remainder = true, d, c.ExpectedOldCarry
	}
	if !found {
		t.Fatalf("missing original numeric input amount=%08x carry=%08x", math.Float32bits(amount), math.Float32bits(carry))
	}
	return damage, remainder
}
