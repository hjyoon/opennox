package server

import (
	"image"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func playerChargeFixture4E17B0(t *testing.T) (*Object, *Object, PlayerDamageRuntime4E17B0, *[]int32) {
	t.Helper()
	target, _, sound := playerDamageFixture4E17B0(t)
	target.HealthData = &HealthData{Cur: 2000, Field2: 2000, Max: 2000}
	source := &Object{
		ObjClass: object.ClassPlayer, PrevPos: types.Ptf(44, 19),
		UpdateData: unsafe.Pointer(&PlayerUpdateData{Player: &Player{}, State: PlayerState13}),
	}
	damages := new([]int32)
	runtime := playerDamageRuntime4E17B0(t, sound, damages)
	runtime.GameplayFlag1 = func() bool { return true }
	runtime.PlayerSetState = func(got *Object, state PlayerState) bool {
		if got != target || state != PlayerState30 {
			t.Fatalf("charge hurt state = %p/%d, want %p/%d", got, state, target, PlayerState30)
		}
		got.UpdateDataPlayer().State = state
		return true
	}
	return target, source, runtime, damages
}

func TestPlayerDamageNative4E17B0BerserkerChargeFractionalCarry(t *testing.T) {
	target, source, runtime, damages := playerChargeFixture4E17B0(t)
	for i, want := range []float32{-0.25, -0.5, 0.25} {
		if handled, result := PlayerDamageNative4E17B0(target, source, source, 150, object.DamageCrush, runtime); !handled || !result {
			t.Fatalf("charge %d = %t/%t", i, handled, result)
		}
		if got := math.Float32frombits(target.UpdateDataPlayer().Field21); got != want {
			t.Fatalf("charge %d fractional carry = %g, want %g", i, got, want)
		}
	}
	if !reflect.DeepEqual(*damages, []int32{148, 148, 147}) || target.HealthData.Cur != 1557 {
		t.Fatalf("fractional charges = %v hp:%d", *damages, target.HealthData.Cur)
	}
}

func TestPlayerDamageNative4E17B0BerserkerChargeDefenseTail(t *testing.T) {
	target, source, runtime, damages := playerChargeFixture4E17B0(t)
	target.UpdateDataPlayer().Field57 = math.Float32bits(0.4)
	source.Buffs = 1 << damageVampirismEnchant4E0B30
	source.BuffsPower[damageVampirismEnchant4E0B30] = 2
	source.PosVec, target.PosVec = types.Ptf(30.5, 31.5), types.Ptf(40.5, 41.5)
	target.Buffs = 1<<playerDamageInvisibleEnchant4E17B0 | 1<<playerDamageShieldEnchant4E17B0 | 1<<ENCHANT_SHOCK
	modifier := &ModifierEff{Defend76: ModifierEffFnc{Fnc: unsafe.Pointer(new(byte))}}
	carry := damageArmorCarryFixture4E17B0(0.25)
	armor := &Object{
		ObjClass: object.ClassArmor, ObjFlags: object.FlagEquipped,
		Damage:     unsafe.Pointer(new(byte)),
		HealthData: &HealthData{Cur: 100, Max: 100}, UpdateData: unsafe.Pointer(carry),
		InitData: unsafe.Pointer(&ModifierInitData{Modifiers: [4]*ModifierEff{nil, nil, modifier, nil}}),
	}
	target.InvFirstItem = armor
	var events []string
	runtime.ItemArmorValue = func(*Object) float32 { return 0.4 }
	runtime.CanDamageArmor = func(item *Object) bool { return item == armor }
	runtime.DamageArmor = func(item, gotSource, weapon *Object, damage int32, typ object.DamageType) bool {
		if item != armor || gotSource != source || weapon != source || damage != 30 || typ != object.DamageCrush {
			t.Fatalf("charge armor damage = %p/%p/%p/%d/%d", item, gotSource, weapon, damage, typ)
		}
		events = append(events, "armor")
		item.HealthData.Cur -= uint16(damage)
		return true
	}
	runtime.ReportArmorHealth = func(owner, item *Object, before, after uint16) {
		if owner != target || item != armor || before != 100 || after != 70 {
			t.Fatalf("armor report = %p/%p/%d/%d", owner, item, before, after)
		}
		events = append(events, "report")
	}
	runtime.QuestMode = func() bool { return true }
	runtime.QuestDamageScale = func() float32 { events = append(events, "quest"); return 0.5 }
	runtime.BuffOff = func(got *Object, enchant EnchantID) {
		if got != target || enchant != playerDamageInvisibleEnchant4E17B0 {
			t.Fatalf("charge BuffOff = %p/%d", got, enchant)
		}
		got.Buffs &^= 1 << enchant
		events = append(events, "invisible-off")
	}
	runtime.CanApplyLateDefend = func(got *ModifierEff) bool { return got == modifier }
	runtime.ApplyLateDefend = func(got *ModifierEff, item, victim, weapon, attacker *Object, damage int32, typ object.DamageType) int32 {
		if got != modifier || item != armor || victim != target || weapon != source || attacker != source || damage != 60 || typ != object.DamageCrush {
			t.Fatalf("charge late defend = %p/%p/%p/%p/%p/%d/%d", got, item, victim, weapon, attacker, damage, typ)
		}
		events = append(events, "late-defend")
		return damage - 10
	}
	runtime.PlayerDamageSound = func(*Object, *Object) { events = append(events, "sound") }
	runtime.Audio = func(id int, obj *Object) {
		if id != damageVampirismSound4E0B30 || obj != source {
			t.Fatalf("charge Vampirism sound = %d/%p", id, obj)
		}
		events = append(events, "vampirism-sound")
	}
	runtime.BalanceFloatInd = func(key string, index int) float64 {
		if key != damageVampirismBalance4E0B30 || index != 1 {
			t.Fatalf("charge Vampirism coefficient = %q/%d", key, index)
		}
		return 0.5
	}
	runtime.AdjustHP = func(obj *Object, amount int32) {
		if obj != source || amount != 25 {
			t.Fatalf("charge Vampirism heal = %p/%d", obj, amount)
		}
		events = append(events, "heal")
	}
	runtime.VampirismFX = func(id int, from, to image.Point, amount uint16) {
		if id != damageVampirismFX4E0B30 || from != image.Pt(30, 32) || to != image.Pt(40, 42) || amount != 25 {
			t.Fatalf("charge Vampirism FX = %d/%v/%v/%d", id, from, to, amount)
		}
		events = append(events, "fx")
	}
	setState := runtime.PlayerSetState
	runtime.PlayerSetState = func(obj *Object, state PlayerState) bool {
		events = append(events, "hurt")
		return setState(obj, state)
	}
	runtime.ShieldReduce = func(obj *Object, damage *int32, typ object.DamageType, weapon *Object) {
		if obj != target || *damage != 50 || typ != object.DamageCrush || weapon != source {
			t.Fatalf("charge Shield = %p/%d/%d/%p", obj, *damage, typ, weapon)
		}
		*damage -= 10
		events = append(events, "shield")
	}
	damageClear := runtime.DamageClear
	runtime.DamageClear = func(obj *Object, damage int32) { events = append(events, "damage"); damageClear(obj, damage) }
	if handled, result := PlayerDamageNative4E17B0(target, source, source, 150, object.DamageCrush, runtime); !handled || !result {
		t.Fatalf("defended charge = %t/%t", handled, result)
	}
	want := []string{"armor", "report", "quest", "invisible-off", "late-defend", "sound", "vampirism-sound", "heal", "fx", "hurt", "shield", "damage"}
	if !reflect.DeepEqual(events, want) || !reflect.DeepEqual(*damages, []int32{40}) ||
		target.HealthData.Cur != 1960 || armor.HealthData.Cur != 70 || *carry != 0.25 ||
		target.HasEnchant(playerDamageInvisibleEnchant4E17B0) || !target.HasEnchant(ENCHANT_SHOCK) {
		t.Fatalf("defense tail = events:%v damages:%v hp:%d armor:%d carry:%g buffs:%#x", events, *damages, target.HealthData.Cur, armor.HealthData.Cur, *carry, target.Buffs)
	}
}

func TestPlayerDamageNative4E17B0BerserkerChargeGates(t *testing.T) {
	for _, tc := range []struct {
		name                                                                 string
		gameplay, enemy, quest, self, coop, god, invulnerable, blockedPlayer bool
		flags                                                                object.Flags
		state                                                                PlayerState
		shieldAbsorbs                                                        bool
		wantResult, wantDamage, wantMutation                                 bool
	}{
		{name: "enemy", gameplay: true, enemy: true, wantResult: true, wantDamage: true, wantMutation: true},
		{name: "friendly gameplay passes non-melee gate", gameplay: true, wantResult: true, wantDamage: true, wantMutation: true},
		{name: "friendly owner gate", wantResult: true, wantMutation: true},
		{name: "campaign enemy", enemy: true, quest: true, wantResult: true, wantDamage: true, wantMutation: true},
		{name: "self outside coop", self: true, wantResult: true, wantDamage: true, wantMutation: true},
		{name: "quest self gate", self: true, quest: true, wantResult: true, wantMutation: true},
		{name: "coop self gate", self: true, coop: true},
		{name: "NoUpdate", flags: object.FlagNoUpdate},
		{name: "dead", flags: object.FlagDead},
		{name: "invulnerable", invulnerable: true, wantResult: true},
		{name: "observer", blockedPlayer: true},
		{name: "god", god: true, wantResult: true, wantMutation: true},
		{name: "shield absorbs", gameplay: true, enemy: true, shieldAbsorbs: true, wantMutation: true},
		{name: "charging target keeps state", gameplay: true, enemy: true, state: PlayerState1, wantResult: true, wantDamage: true, wantMutation: true},
		{name: "protected state keeps state", gameplay: true, enemy: true, state: PlayerState15, wantResult: true, wantDamage: true, wantMutation: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			target, source, runtime, damages := playerChargeFixture4E17B0(t)
			update := target.UpdateDataPlayer()
			update.Field76, update.Field75 = 99, 98
			if tc.self {
				source = target
			}
			target.ObjFlags |= tc.flags
			if tc.invulnerable {
				target.Buffs |= 1 << playerDamageInvulnerableEnchant4E17B0
			}
			if tc.blockedPlayer {
				update.Player.Field3680 = 1
			}
			if tc.state != 0 {
				update.State = tc.state
			}
			if tc.shieldAbsorbs {
				target.Buffs |= 1 << playerDamageShieldEnchant4E17B0
				runtime.ShieldReduce = func(_ *Object, damage *int32, _ object.DamageType, _ *Object) { *damage = 0 }
			}
			runtime.GameplayFlag1 = func() bool { return tc.gameplay }
			runtime.IsEnemy = func(*Object, *Object) bool { return tc.enemy }
			runtime.QuestMode = func() bool { return tc.quest }
			runtime.CoopMode = func() bool { return tc.coop }
			runtime.GodMode = func() bool { return tc.god }
			before, beforeUpdate := *target, *update
			if handled, result := PlayerDamageNative4E17B0(target, source, source, 150, object.DamageCrush, runtime); !handled || result != tc.wantResult {
				t.Fatalf("charge gate = %t/%t, want true/%t", handled, result, tc.wantResult)
			}
			if (len(*damages) != 0) != tc.wantDamage || (update.Field76 != 99) != tc.wantMutation {
				t.Fatalf("gate damages:%v marker:%d, want damage:%t mutation:%t", *damages, update.Field76, tc.wantDamage, tc.wantMutation)
			}
			if !tc.wantMutation && (*target != before || *update != beforeUpdate) {
				t.Fatal("early exit changed player state")
			}
			if tc.wantDamage && tc.state != 0 && update.State != tc.state {
				t.Fatal("charge changed protected player state")
			}
			if !tc.wantDamage && target.HealthData.Cur != 2000 {
				t.Fatal("immune player lost health")
			}
		})
	}
}

func TestPlayerDamageNative4E17B0BerserkerChargePreflight(t *testing.T) {
	for _, tc := range []struct {
		name, reason string
		change       func(*Object, *Object, *PlayerDamageRuntime4E17B0)
	}{
		{"friendly service", "missing friendly-fire service", func(_, _ *Object, r *PlayerDamageRuntime4E17B0) { r.GameplayFlag1 = nil }},
		{"hurt service", "missing player hurt-state service", func(_, _ *Object, r *PlayerDamageRuntime4E17B0) { r.PlayerSetState = nil }},
		{"GameBall drop", "GameBall drop", func(target, _ *Object, r *PlayerDamageRuntime4E17B0) {
			r.GameBallType = 77
			target.Field129 = &Object{TypeInd: 77}
		}},
		{"Vampirism", "missing Vampirism service", func(_, source *Object, _ *PlayerDamageRuntime4E17B0) {
			source.Buffs |= 1 << damageVampirismEnchant4E0B30
		}},
		{"monster-like weapon", "unsupported player damage shape", func(_, source *Object, _ *PlayerDamageRuntime4E17B0) { source.ObjClass |= object.ClassMonster }},
		{"weapon-like source", "unsupported player damage shape", func(_, source *Object, _ *PlayerDamageRuntime4E17B0) { source.ObjClass |= object.ClassWeapon }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			target, source, runtime, damages := playerChargeFixture4E17B0(t)
			tc.change(target, source, &runtime)
			before, beforeUpdate, beforeSource := *target, *target.UpdateDataPlayer(), *source.UpdateDataPlayer()
			var reason string
			runtime.Unsupported = func(got string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = got }
			if handled, result := PlayerDamageNative4E17B0(target, source, source, 150, object.DamageCrush, runtime); handled || result || reason != tc.reason {
				t.Fatalf("charge preflight = %t/%t reason:%q, want %q", handled, result, reason, tc.reason)
			}
			if tc.reason != "unsupported player damage shape" {
				// The supported charge entry now owns 004E18C4's unconditional
				// cached marker clear before admitting its unblocked HP services.
				// All other fields, carry, HP and attacker data remain untouched.
				beforeUpdate.Field76 = 0
			}
			if *target != before || *target.UpdateDataPlayer() != beforeUpdate || *source.UpdateDataPlayer() != beforeSource || len(*damages) != 0 {
				t.Fatal("unsupported charge changed player state")
			}
		})
	}
}

func TestPlayerDamageNative4E17B0BerserkerChargeShieldBlock(t *testing.T) {
	target, source, runtime, damages := playerChargeFixture4E17B0(t)
	update := target.UpdateDataPlayer()
	update.State, update.Player.ArmorEquip = PlayerState16, 0x1000000
	shield := &Object{ObjClass: object.ClassArmor, ObjSubClass: 2, ObjFlags: object.FlagEquipped, HealthData: &HealthData{Cur: 100, Max: 100}}
	target.InvFirstItem = shield
	runtime.BlockSourceExcluded = func(*Object) bool { return false }
	runtime.BlockDirection = func(*Object, types.Pointf) bool { return true }
	runtime.CanDamageBlockItem = func(item *Object) bool { return item == shield }
	runtime.BlockDamagePercent = func() float64 { return 0.25 }
	runtime.Audio = func(id int, obj *Object) {
		if id != 878 || obj != target {
			t.Fatalf("shield sound = %d/%p", id, obj)
		}
	}
	runtime.DamageBlockItem = func(item, owner, attacker, weapon *Object, amount float32, typ object.DamageType) bool {
		if item != shield || owner != target || attacker != source || weapon != source || amount != 37.5 || typ != object.DamageCrush {
			t.Fatal("wrong charge shield arguments")
		}
		item.HealthData.Cur -= uint16(math.RoundToEven(float64(amount)))
		return true
	}
	if handled, result := PlayerDamageNative4E17B0(target, source, source, 150, object.DamageCrush, runtime); !handled || result {
		t.Fatalf("shield-blocked charge = %t/%t", handled, result)
	}
	if len(*damages) != 0 || target.HealthData.Cur != 2000 || shield.HealthData.Cur != 62 || update.State != PlayerState16 || update.Field76 != 0 {
		t.Fatalf("shield-blocked charge = damage:%v hp:%d shield:%d state:%d marker:%d", *damages, target.HealthData.Cur, shield.HealthData.Cur, update.State, update.Field76)
	}
}

func TestPlayerDamageNative4E17B0BerserkerCharge(t *testing.T) {
	target, source, runtime, damages := playerChargeFixture4E17B0(t)
	update := target.UpdateDataPlayer()
	update.Field57 = math.Float32bits(0.4)
	beforeSource := *source.UpdateDataPlayer()
	var events []string
	runtime.BuffOff = func(got *Object, enchant EnchantID) {
		if got != target || enchant != playerDamageInvisibleEnchant4E17B0 {
			t.Fatalf("BuffOff(%p,%d)", got, enchant)
		}
		events = append(events, "buff-off")
	}
	runtime.PlayerDamageSound = func(got, weapon *Object) {
		if got != target || weapon != source {
			t.Fatalf("charge sound = %p/%p", got, weapon)
		}
		events = append(events, "sound")
	}
	setState := runtime.PlayerSetState
	runtime.PlayerSetState = func(got *Object, state PlayerState) bool {
		events = append(events, "hurt")
		return setState(got, state)
	}
	damageClear := runtime.DamageClear
	runtime.DamageClear = func(got *Object, damage int32) {
		events = append(events, "damage")
		damageClear(got, damage)
	}
	if handled, result := PlayerDamageNative4E17B0(target, source, source, 150, object.DamageCrush, runtime); !handled || !result {
		t.Fatalf("charge = handled:%t result:%t", handled, result)
	}
	// GAME.EXE 004E1EE8 halves armor absorption, not the resulting damage.
	// 004E1F42 writes the integer 2 into both player damage-marker DWORDs.
	if !reflect.DeepEqual(*damages, []int32{120}) || target.HealthData.Cur != 1880 ||
		update.Field76 != 2 || update.Field75 != 2 || update.Field21 != 0 || update.State != PlayerState30 ||
		target.Obj130 != source || target.Pos132 != source.PrevPos ||
		target.Field131 != uint32(object.DamageCrush) || target.Frame134 != 700 {
		t.Fatalf("charge state = damage:%v hp:%d marker:%d/%d carry:%#x state:%d source:%p pos:%v type/frame:%d/%d",
			*damages, target.HealthData.Cur, update.Field76, update.Field75, update.Field21, update.State,
			target.Obj130, target.Pos132, target.Field131, target.Frame134)
	}
	if want := []string{"buff-off", "sound", "hurt", "damage"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("charge events = %v, want %v", events, want)
	}
	if *source.UpdateDataPlayer() != beforeSource {
		t.Fatal("charge interpreted or mutated its PLAYER source as monster update data")
	}
}
