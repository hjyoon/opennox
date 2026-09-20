package server

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/player"
	"github.com/opennox/libs/types"
)

func obeliskTestDeps53C580(units ...*Object) obeliskUpdateNativeDeps53C580 {
	return obeliskUpdateNativeDeps53C580{
		playerUnits: func() []*Object { return units },
		mapTraceVision: func(*Object, *Object) bool {
			return true
		},
		frame:    func() uint32 { return 20 },
		tickRate: func() uint32 { return 30 },
		balanceFloat: func(string) float32 {
			return 0
		},
		questManaMultiplier: func(player.Class) float32 {
			return 1
		},
	}
}

func TestObeliskUpdate53C580NormalRechargeSequence(t *testing.T) {
	replenishment := byte(1)
	effect := &ModifierEff{Attack40: ModifierEffFnc{
		Fnc: unsafe.Pointer(&replenishment),
		Val: 25,
	}}
	attrs := &ModifierInitData{}
	attrs.Modifiers[2] = effect
	wand := &WandUseData{Charge: 8, MaxCharge: 10, Progress: 80}
	weapon := &Object{
		ObjClass:    object.ClassWand,
		ObjSubClass: object.SubClass(object.WeaponStaffLightning),
		InitData:    unsafe.Pointer(attrs),
		UseData:     UseDataPtr{Ptr: unsafe.Pointer(wand)},
	}
	pl := &Player{PlayerInd: 7, ProtUnitManaCur: 0x12345678}
	update := &PlayerUpdateData{
		ManaCur:        8,
		ManaMax:        10,
		EquippedWeapon: weapon,
		Player:         pl,
	}
	unit := &Object{PosVec: types.Ptf(1, 1), UpdateData: unsafe.Pointer(update)}
	data := &ObeliskUpdateData{Mana: 2}
	obelisk := &Object{Field34: 1, UpdateData: unsafe.Pointer(data)}

	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, ptr := range map[string]unsafe.Pointer{
			"obelisk": unsafe.Pointer(obelisk),
			"unit":    unsafe.Pointer(unit),
			"weapon":  unsafe.Pointer(weapon),
			"update":  unsafe.Pointer(update),
		} {
			if uintptr(ptr) <= 0xffffffff {
				t.Fatalf("%s pointer = %p, want native address above 4 GiB", name, ptr)
			}
		}
	}

	var events []string
	deps := obeliskTestDeps53C580(unit)
	deps.replenishmentEffect = unsafe.Pointer(&replenishment)
	deps.reportCharges = func(ind uint8, gotWeapon *Object, charge, max uint8) {
		if ind != 7 || gotWeapon != weapon || charge != 10 || max != 10 {
			t.Fatalf("charge report = %d/%p/%d/%d", ind, gotWeapon, charge, max)
		}
		events = append(events, "report")
	}
	deps.protectMana = func(token uint32, delta int16) {
		if token != pl.ProtUnitManaCur || delta != 1 {
			t.Fatalf("protect mana = %#x/%d", token, delta)
		}
		events = append(events, "protect")
	}
	deps.protectPlayerHPMana = func(uint32, uint16) {
		t.Fatal("normal recharge should not clamp player mana")
	}
	deps.needSync = func(got *Object) {
		if got != obelisk {
			t.Fatalf("sync object = %p, want %p", got, obelisk)
		}
		events = append(events, "sync")
	}
	deps.audioManaRecharge = func(got *Object) {
		if got != obelisk {
			t.Fatalf("audio object = %p, want %p", got, obelisk)
		}
		events = append(events, "audio")
	}

	if got := obeliskUpdateNative53C580(obelisk, deps); got != 0 {
		t.Fatalf("result = %d, want 0 with a nearby player", got)
	}
	if data.Mana != 0 || wand.Progress != 100 || wand.Charge != 10 || update.ManaCur != 9 {
		t.Fatalf("mana/wand/player = %d/%d/%d/%d, want 0/100/10/9", data.Mana, wand.Progress, wand.Charge, update.ManaCur)
	}
	if obelisk.Field34 != 20 {
		t.Fatalf("audio frame = %d, want 20", obelisk.Field34)
	}
	if want := []string{"report", "protect", "sync", "audio"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestObeliskUpdate53C580QuestDoesNotConsumeMana(t *testing.T) {
	replenishment := byte(1)
	effect := &ModifierEff{Attack40: ModifierEffFnc{
		Fnc: unsafe.Pointer(&replenishment),
		Val: 25,
	}}
	attrs := &ModifierInitData{}
	attrs.Modifiers[3] = effect
	wand := &WandUseData{Charge: 8, MaxCharge: 10, Progress: 80}
	weapon := &Object{
		ObjClass:    object.ClassWand,
		ObjSubClass: object.SubClass(object.WeaponStaffForceOfNature),
		InitData:    unsafe.Pointer(attrs),
		UseData:     UseDataPtr{Ptr: unsafe.Pointer(wand)},
	}
	pl := &Player{PlayerInd: 3, ProtUnitManaCur: 0x89abcdef}
	pl.Info().SetPlayerClass(player.Wizard)
	update := &PlayerUpdateData{
		ManaCur:        9,
		ManaMax:        10,
		EquippedWeapon: weapon,
		Player:         pl,
	}
	unit := &Object{UpdateData: unsafe.Pointer(update)}
	data := &ObeliskUpdateData{Mana: 2}
	obelisk := &Object{Field34: 11, UpdateData: unsafe.Pointer(data)}

	var events []string
	deps := obeliskTestDeps53C580(unit)
	deps.questMode = true
	deps.frame = func() uint32 { return 30 }
	deps.replenishmentEffect = unsafe.Pointer(&replenishment)
	deps.questManaMultiplier = func(class player.Class) float32 {
		if class != player.Wizard {
			t.Fatalf("quest class = %d, want wizard", class)
		}
		return 2.5
	}
	deps.audioManaRecharge = func(*Object) { events = append(events, "audio") }
	deps.reportCharges = func(ind uint8, gotWeapon *Object, charge, max uint8) {
		if ind != 3 || gotWeapon != weapon || charge != 10 || max != 10 {
			t.Fatalf("charge report = %d/%p/%d/%d", ind, gotWeapon, charge, max)
		}
		events = append(events, "report")
	}
	deps.protectMana = func(token uint32, delta int16) {
		if token != pl.ProtUnitManaCur || delta != 2 {
			t.Fatalf("protect mana = %#x/%d, want %#x/2", token, delta, pl.ProtUnitManaCur)
		}
		events = append(events, "protect")
	}
	deps.protectPlayerHPMana = func(token uint32, value uint16) {
		if token != pl.ProtUnitManaCur || value != 10 {
			t.Fatalf("protect max = %#x/%d, want %#x/10", token, value, pl.ProtUnitManaCur)
		}
		events = append(events, "clamp")
	}
	deps.needSync = func(*Object) { t.Fatal("quest recharge should not sync unchanged obelisk mana") }

	if got := obeliskUpdateNative53C580(obelisk, deps); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	if data.Mana != 2 || wand.Progress != 100 || wand.Charge != 10 || update.ManaCur != 10 {
		t.Fatalf("mana/wand/player = %d/%d/%d/%d, want 2/100/10/10", data.Mana, wand.Progress, wand.Charge, update.ManaCur)
	}
	if obelisk.Field34 != 11 {
		t.Fatalf("quest mode changed audio throttle frame to %d", obelisk.Field34)
	}
	if want := []string{"audio", "report", "protect", "clamp", "audio"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestObeliskUpdate53C580OnlineForcesWandRecharge(t *testing.T) {
	wand := &WandUseData{Charge: 9, MaxCharge: 10, Progress: 99}
	weapon := &Object{
		ObjClass:    object.ClassWand,
		ObjSubClass: object.SubClass(object.WeaponStaffDeathRay),
		UseData:     UseDataPtr{Ptr: unsafe.Pointer(wand)},
	}
	pl := &Player{PlayerInd: 5}
	update := &PlayerUpdateData{ManaCur: 10, ManaMax: 10, EquippedWeapon: weapon, Player: pl}
	unit := &Object{UpdateData: unsafe.Pointer(update)}
	data := &ObeliskUpdateData{Mana: 2}
	obelisk := &Object{UpdateData: unsafe.Pointer(data)}

	reports := 0
	deps := obeliskTestDeps53C580(unit)
	deps.onlineMode = true
	deps.reportCharges = func(ind uint8, gotWeapon *Object, charge, max uint8) {
		if ind != 5 || gotWeapon != weapon || charge != 10 || max != 10 {
			t.Fatalf("charge report = %d/%p/%d/%d", ind, gotWeapon, charge, max)
		}
		reports++
	}
	obeliskUpdateNative53C580(obelisk, deps)
	if data.Mana != 1 || wand.Progress != 100 || wand.Charge != 10 || reports != 1 {
		t.Fatalf("mana/progress/charge/reports = %d/%d/%d/%d, want 1/100/10/1", data.Mana, wand.Progress, wand.Charge, reports)
	}
}

func TestObeliskUpdate53C580PassiveRegenerationAndNearbySuppression(t *testing.T) {
	data := &ObeliskUpdateData{Mana: 8}
	obelisk := &Object{UpdateData: unsafe.Pointer(data)}
	deps := obeliskTestDeps53C580()
	deps.frame = func() uint32 { return 15 }
	syncs := 0
	deps.needSync = func(got *Object) {
		if got != obelisk {
			t.Fatalf("sync object = %p, want %p", got, obelisk)
		}
		syncs++
	}
	if got := obeliskUpdateNative53C580(obelisk, deps); got != 8 {
		t.Fatalf("passive result = %d, want previous mana 8", got)
	}
	if data.Mana != 9 || syncs != 1 {
		t.Fatalf("passive mana/syncs = %d/%d, want 9/1", data.Mana, syncs)
	}

	data.Mana = 8
	update := &PlayerUpdateData{ManaCur: 10, ManaMax: 10, Player: new(Player)}
	unit := &Object{UpdateData: unsafe.Pointer(update)}
	deps = obeliskTestDeps53C580(unit)
	deps.frame = func() uint32 { return 15 }
	deps.needSync = func(*Object) { t.Fatal("nearby full player should suppress passive regeneration") }
	if got := obeliskUpdateNative53C580(obelisk, deps); got != 0 {
		t.Fatalf("nearby result = %d, want 0", got)
	}
	if data.Mana != 8 {
		t.Fatalf("nearby player allowed passive mana to become %d", data.Mana)
	}
}

func TestObeliskRechargeRate53C940(t *testing.T) {
	replenishment := byte(1)
	other := byte(2)
	attrs := &ModifierInitData{}
	attrs.Modifiers[2] = &ModifierEff{Attack40: ModifierEffFnc{Fnc: unsafe.Pointer(&other), Val: 17}}
	attrs.Modifiers[3] = &ModifierEff{Attack40: ModifierEffFnc{Fnc: unsafe.Pointer(&replenishment), Val: 37}}
	item := &Object{InitData: unsafe.Pointer(attrs)}
	deps := obeliskTestDeps53C580()
	deps.replenishmentEffect = unsafe.Pointer(&replenishment)
	if got := obeliskRechargeRate53C940(item, deps); got != 37 {
		t.Fatalf("modifier recharge rate = %d, want 37", got)
	}

	item = &Object{
		ObjClass:    object.ClassWand,
		ObjSubClass: object.SubClass(object.WeaponStaffOblivionOrb),
	}
	deps.balanceFloat = func(name string) float32 {
		if name != "OblivionStaffRechargeRate" {
			t.Fatalf("balance key = %q", name)
		}
		return 2.5
	}
	if got := obeliskRechargeRate53C940(item, deps); got != 2 {
		t.Fatalf("Oblivion recharge rate = %d, want round-to-even 2", got)
	}
}
