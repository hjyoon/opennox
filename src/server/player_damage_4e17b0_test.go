package server

import (
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func playerDamageFixture4E17B0(t *testing.T) (*Object, *Object, unsafe.Pointer) {
	t.Helper()
	sound := unsafe.Pointer(new(byte))
	player := &Player{ArmorEquip: 0x405, WeaponEquip: 0x100}
	update := &PlayerUpdateData{Player: player, Field57: math.Float32bits(0.03), State: PlayerState13}
	target := &Object{
		ObjClass:    object.ClassPlayer,
		ObjFlags:    object.FlagActive | object.FlagEnabled,
		UpdateData:  unsafe.Pointer(update),
		HealthData:  &HealthData{Cur: 20, Field2: 20, Max: 20},
		DamageSound: sound,
	}
	source := monsterActionTestObject50A910(t)
	source.ObjClass = object.ClassMonster
	source.PrevPos.X = 44
	return target, source, sound
}

func playerDamageRuntime4E17B0(t *testing.T, sound unsafe.Pointer, damages *[]int32) PlayerDamageRuntime4E17B0 {
	t.Helper()
	return PlayerDamageRuntime4E17B0{
		Frame:              func() uint32 { return 700 },
		QuestMode:          func() bool { return false },
		QuestDamageScale:   func() float32 { return 1 },
		GodMode:            func() bool { return false },
		IsEnemy:            func(*Object, *Object) bool { return true },
		BuffOff:            func(*Object, EnchantID) {},
		ItemArmorValue:     func(*Object) float32 { return 0.01 },
		FireProtection:     func(*Object) float64 { return 0 },
		PlayerDamageSoundC: sound,
		DamageClear: func(target *Object, damage int32) {
			*damages = append(*damages, damage)
			if int32(target.HealthData.Cur) <= damage {
				target.HealthData.Cur = 0
			} else {
				target.HealthData.Cur -= uint16(damage)
			}
		},
		Unsupported: func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) {
			t.Fatalf("unexpected unsupported branch: %s", reason)
		},
	}
}

func TestPlayerDamageNative4E17B0SourceLessLava(t *testing.T) {
	target, _, sound := playerDamageFixture4E17B0(t)
	target.UpdateDataPlayer().Field57 = 0
	target.Pos132.X = 99
	var damages []int32
	var events []string
	runtime := playerDamageRuntime4E17B0(t, sound, &damages)
	runtime.QuestMode = func() bool { return true }
	runtime.QuestDamageScale = func() float32 {
		events = append(events, "quest-scale")
		return 0.5
	}
	runtime.FireProtection = func(got *Object) float64 {
		if got != target {
			t.Fatalf("FireProtection(%p), want %p", got, target)
		}
		events = append(events, "fire-protection")
		return 0.25
	}
	runtime.Audio = func(id int, got *Object) {
		if id != 104 || got != target {
			t.Fatalf("Audio(%d,%p), want (104,%p)", id, got, target)
		}
		events = append(events, "fire-sound")
	}
	runtime.BuffOff = func(got *Object, enchant EnchantID) {
		if got != target || enchant != playerDamageInvisibleEnchant4E17B0 {
			t.Fatalf("BuffOff(%p,%d)", got, enchant)
		}
		events = append(events, "buff-off")
	}
	runtime.PlayerDamageSound = func(gotTarget, gotSource *Object) {
		if gotTarget != target || gotSource != nil {
			t.Fatalf("PlayerDamageSound(%p,%p), want (%p,nil)", gotTarget, gotSource, target)
		}
		events = append(events, "damage-sound")
	}
	if handled, result := PlayerDamageNative4E17B0(target, nil, nil, 5, object.DamageLava, runtime); !handled || !result {
		t.Fatalf("source-less LAVA = handled:%t result:%t", handled, result)
	}
	if len(damages) != 1 || damages[0] != 2 {
		t.Fatalf("LAVA damages = %v, want [2]", damages)
	}
	wantEvents := []string{"quest-scale", "fire-protection", "fire-sound", "buff-off", "damage-sound"}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
	update := target.UpdateDataPlayer()
	if update.Field76 != 2 || update.Field75 != math.Float32bits(float32(object.DamageLava)) ||
		target.Pos132 != (types.Pointf{}) || target.Obj130 != nil || target.Field131 != uint32(object.DamageLava) || target.Frame134 != 700 {
		t.Fatalf("LAVA metadata = marker:%#x/%#x pos:%v source:%p type:%d frame:%d",
			update.Field75, update.Field76, target.Pos132, target.Obj130, target.Field131, target.Frame134)
	}
}

func TestPlayerDamageNative4E17B0LavaDamagesEquippedArmor(t *testing.T) {
	target, _, sound := playerDamageFixture4E17B0(t)
	target.UpdateDataPlayer().Field57 = math.Float32bits(0.5)
	carry := float32(0.4)
	armor := &Object{
		ObjClass:   object.ClassArmor,
		ObjFlags:   object.FlagEquipped,
		HealthData: &HealthData{Cur: 10, Max: 10},
		UpdateData: unsafe.Pointer(&carry),
		InitData:   unsafe.Pointer(&ModifierInitData{}),
	}
	target.InvFirstItem = armor
	var damages []int32
	runtime := playerDamageRuntime4E17B0(t, sound, &damages)
	runtime.ItemArmorValue = func(got *Object) float32 {
		if got != armor {
			t.Fatalf("ItemArmorValue(%p), want %p", got, armor)
		}
		return 0.5
	}
	runtime.CanDamageArmor = func(got *Object) bool { return got == armor }
	var reported bool
	runtime.DamageArmor = func(got, source, weapon *Object, damage int32, typ object.DamageType) bool {
		if got != armor || source != nil || weapon != nil || damage != 2 || typ != object.DamageLava {
			t.Fatalf("DamageArmor(%p,%p,%p,%d,%d)", got, source, weapon, damage, typ)
		}
		got.HealthData.Cur -= uint16(damage)
		return true
	}
	runtime.ReportArmorHealth = func(owner, got *Object, before, after uint16) {
		if owner != target || got != armor || before != 10 || after != 8 {
			t.Fatalf("ReportArmorHealth(%p,%p,%d,%d)", owner, got, before, after)
		}
		reported = true
	}
	if handled, result := PlayerDamageNative4E17B0(target, nil, nil, 2, object.DamageLava, runtime); !handled || !result {
		t.Fatalf("armored LAVA = handled:%t result:%t", handled, result)
	}
	wantCarry := float32(float64(float32(0.5)) / float64(float32(0.5)) * float64(int32(2)))
	wantCarry += float32(0.4)
	wantCarry -= 2
	if armor.HealthData.Cur != 8 || math.Float32bits(carry) != math.Float32bits(wantCarry) || !reported {
		t.Fatalf("armor state = hp:%d carry:%v reported:%t", armor.HealthData.Cur, carry, reported)
	}
	if !reflect.DeepEqual(damages, []int32{2}) {
		t.Fatalf("player damages = %v, want [2]", damages)
	}
}

func TestPlayerDamageNative4E17B0LavaUsesBinary64BeforeFloatSpill(t *testing.T) {
	target, _, sound := playerDamageFixture4E17B0(t)
	target.UpdateDataPlayer().Field57 = 0
	var damages []int32
	runtime := playerDamageRuntime4E17B0(t, sound, &damages)
	protection := math.Float32frombits(0x3defaf0d)
	runtime.FireProtection = func(*Object) float64 { return float64(protection) }
	if handled, result := PlayerDamageNative4E17B0(target, nil, nil, 5413, object.DamageLava, runtime); !handled || !result {
		t.Fatalf("source-less LAVA = handled:%t result:%t", handled, result)
	}
	if !reflect.DeepEqual(damages, []int32{4780}) {
		t.Fatalf("binary64 LAVA damages = %v, want [4780]", damages)
	}
}

func TestPlayerDamageNative4E17B0SourceLessPoisonSkipsEquipment(t *testing.T) {
	target, _, sound := playerDamageFixture4E17B0(t)
	target.UpdateDataPlayer().Player.ArmorEquip = 0x3000000
	target.Buffs |= 1 << playerDamageShieldEnchant4E17B0
	target.Pos132.X = 99
	carry := float32(0.4)
	armor := &Object{
		ObjClass:   object.ClassArmor,
		ObjFlags:   object.FlagEquipped,
		HealthData: &HealthData{Cur: 10, Max: 10},
		UpdateData: unsafe.Pointer(&carry),
		InitData:   unsafe.Pointer(&ModifierInitData{}),
	}
	target.InvFirstItem = armor
	var damages []int32
	var events []string
	runtime := playerDamageRuntime4E17B0(t, sound, &damages)
	runtime.QuestMode = func() bool { return true }
	runtime.QuestDamageScale = func() float32 {
		events = append(events, "quest-scale")
		return 0.25
	}
	runtime.ItemArmorValue = func(*Object) float32 { t.Fatal("poison damaged armor"); return 0 }
	runtime.FireProtection = func(*Object) float64 { t.Fatal("poison checked fire protection"); return 0 }
	runtime.BuffOff = func(*Object, EnchantID) { t.Fatal("source-less poison removed invisibility") }
	runtime.PlayerDamageSound = func(gotTarget, gotSource *Object) {
		if gotTarget != target || gotSource != nil {
			t.Fatalf("PlayerDamageSound(%p,%p), want (%p,nil)", gotTarget, gotSource, target)
		}
		events = append(events, "damage-sound")
	}
	if handled, result := PlayerDamageNative4E17B0(target, nil, nil, 1, object.DamagePoison, runtime); !handled || !result {
		t.Fatalf("source-less POISON = handled:%t result:%t", handled, result)
	}
	if !reflect.DeepEqual(damages, []int32{1}) || target.HealthData.Cur != 19 {
		t.Fatalf("POISON damages = %v, player health = %d", damages, target.HealthData.Cur)
	}
	if !reflect.DeepEqual(events, []string{"quest-scale", "damage-sound"}) {
		t.Fatalf("events = %v", events)
	}
	if armor.HealthData.Cur != 10 || math.Float32bits(carry) != math.Float32bits(0.4) {
		t.Fatalf("armor changed: hp:%d carry:%v", armor.HealthData.Cur, carry)
	}
	update := target.UpdateDataPlayer()
	if update.Field76 != 2 || update.Field75 != math.Float32bits(float32(object.DamagePoison)) ||
		target.Pos132 != (types.Pointf{}) || target.Obj130 != nil || target.Field131 != uint32(object.DamagePoison) || target.Frame134 != 700 {
		t.Fatalf("POISON metadata = marker:%#x/%#x pos:%v source:%p type:%d frame:%d",
			update.Field75, update.Field76, target.Pos132, target.Obj130, target.Field131, target.Frame134)
	}
}

func TestPlayerDamageNative4E17B0SpiderBiteSequence(t *testing.T) {
	target, source, sound := playerDamageFixture4E17B0(t)
	for i := 0; i < 3; i++ {
		carry := new(float32)
		item := &Object{
			ObjClass:    object.ClassArmor,
			ObjFlags:    object.FlagEquipped,
			HealthData:  &HealthData{Cur: 10, Max: 10},
			UpdateData:  unsafe.Pointer(carry),
			InitData:    unsafe.Pointer(&ModifierInitData{}),
			InvNextItem: target.InvFirstItem,
		}
		target.InvFirstItem = item
	}
	var damages []int32
	runtime := playerDamageRuntime4E17B0(t, sound, &damages)
	for i := 0; i < 7; i++ {
		handled, result := PlayerDamageNative4E17B0(target, source, source, 3, object.DamageBite, runtime)
		if !handled || !result {
			t.Fatalf("hit %d = handled:%t result:%t", i, handled, result)
		}
	}
	want := []int32{3, 3, 3, 3, 3, 2, 3}
	if len(damages) != len(want) {
		t.Fatalf("damages = %v, want %v", damages, want)
	}
	for i := range want {
		if damages[i] != want[i] {
			t.Fatalf("damages = %v, want %v", damages, want)
		}
	}
	if target.HealthData.Cur != 0 || target.Obj130 != source || target.Field131 != uint32(object.DamageBite) ||
		target.Frame134 != 700 || target.Pos132.X != 44 {
		t.Fatalf("final target state = hp:%d source:%p type:%d frame:%d pos:%v",
			target.HealthData.Cur, target.Obj130, target.Field131, target.Frame134, target.Pos132)
	}
	update := target.UpdateDataPlayer()
	if update.Field76 != 2 || update.Field75 != math.Float32bits(float32(object.DamageBite)) {
		t.Fatalf("damage marker = %#x/%#x", update.Field75, update.Field76)
	}
}

func TestPlayerDamageNative4E17B0SpiderBiteShieldBlock(t *testing.T) {
	target, source, sound := playerDamageFixture4E17B0(t)
	update := target.UpdateDataPlayer()
	update.State = PlayerState16
	update.Field76 = 7
	update.Player.ArmorEquip = 0x1000000
	shield := &Object{
		ObjClass:    object.ClassArmor,
		ObjSubClass: object.SubClass(2),
		ObjFlags:    object.FlagEquipped,
		HealthData:  &HealthData{Cur: 10, Max: 10},
	}
	target.InvFirstItem = shield
	var damages []int32
	var events []string
	runtime := playerDamageRuntime4E17B0(t, sound, &damages)
	runtime.BlockSourceExcluded = func(got *Object) bool {
		if got != source {
			t.Fatalf("block source = %p, want %p", got, source)
		}
		events = append(events, "source")
		return false
	}
	runtime.BlockDirection = func(got *Object, pos types.Pointf) bool {
		if got != target || pos != source.PrevPos {
			t.Fatalf("block direction = (%p,%v), want (%p,%v)", got, pos, target, source.PrevPos)
		}
		events = append(events, "direction")
		return true
	}
	runtime.CanDamageBlockItem = func(got *Object) bool {
		if got != shield {
			t.Fatalf("block item = %p, want %p", got, shield)
		}
		events = append(events, "preflight")
		return true
	}
	runtime.PlayerSetState = func(*Object, PlayerState) bool {
		t.Fatal("intact shield changed player state")
		return false
	}
	runtime.Audio = func(id int, got *Object) {
		if id != 878 || got != target {
			t.Fatalf("block audio = (%d,%p)", id, got)
		}
		events = append(events, "audio")
	}
	runtime.BlockDamagePercent = func() float64 {
		events = append(events, "percent")
		return 0.25
	}
	runtime.DamageBlockItem = func(item, owner, gotSource, effective *Object, amount float32, typ object.DamageType) bool {
		if item != shield || owner != target || gotSource != source || effective != source ||
			amount != 2.5 || typ != object.DamageBite {
			t.Fatalf("block damage = (%p,%p,%p,%p,%v,%d)", item, owner, gotSource, effective, amount, typ)
		}
		events = append(events, "durability")
		shield.HealthData.Cur -= 2
		return true
	}
	if handled, result := PlayerDamageNative4E17B0(target, source, source, 10, object.DamageBite, runtime); !handled || result {
		t.Fatalf("shield block = handled:%t result:%t", handled, result)
	}
	if !reflect.DeepEqual(events, []string{"source", "direction", "preflight", "audio", "percent", "durability"}) {
		t.Fatalf("block events = %v", events)
	}
	if len(damages) != 0 || target.HealthData.Cur != 20 || shield.HealthData.Cur != 8 || update.Field76 != 0 || update.State != PlayerState16 {
		t.Fatalf("block state = damage:%v health:%d shield:%d marker:%d state:%d",
			damages, target.HealthData.Cur, shield.HealthData.Cur, update.Field76, update.State)
	}
}

func TestPlayerDamageNative4E17B0ShieldBreakChangesState(t *testing.T) {
	target, source, sound := playerDamageFixture4E17B0(t)
	update := target.UpdateDataPlayer()
	update.State = PlayerState16
	update.Player.ArmorEquip = 0x1000000
	shield := &Object{ObjSubClass: object.SubClass(2), ObjFlags: object.FlagEquipped}
	target.InvFirstItem = shield
	var damages []int32
	runtime := playerDamageRuntime4E17B0(t, sound, &damages)
	runtime.BlockSourceExcluded = func(*Object) bool { return false }
	runtime.BlockDirection = func(*Object, types.Pointf) bool { return true }
	runtime.BlockDamagePercent = func() float64 { return 1 }
	runtime.CanDamageBlockItem = func(*Object) bool { return true }
	runtime.DamageBlockItem = func(*Object, *Object, *Object, *Object, float32, object.DamageType) bool {
		shield.ObjFlags |= object.FlagDestroyed
		return true
	}
	runtime.PlayerSetState = func(got *Object, state PlayerState) bool {
		if got != target || state != PlayerState13 {
			t.Fatalf("PlayerSetState(%p,%d)", got, state)
		}
		update.State = state
		return true
	}
	runtime.Audio = func(int, *Object) {}
	if handled, result := PlayerDamageNative4E17B0(target, source, source, 3, object.DamageBite, runtime); !handled || result {
		t.Fatalf("broken shield block = %t/%t", handled, result)
	}
	if update.State != PlayerState13 || len(damages) != 0 {
		t.Fatalf("broken shield state = %d, damages = %v", update.State, damages)
	}
}

func TestPlayerDamageNative4E17B0ShieldStanceNeedsFrontHit(t *testing.T) {
	for _, test := range []struct {
		name     string
		state    PlayerState
		front    bool
		weapon   uint32
		wantTest bool
	}{
		{name: "rear hit", state: PlayerState16, front: false, wantTest: true},
		{name: "ordinary stance", state: PlayerState13, front: true},
		{name: "sword does not block bite", state: PlayerState13, front: true, weapon: 0x400 | 0x7ff8000},
	} {
		t.Run(test.name, func(t *testing.T) {
			target, source, sound := playerDamageFixture4E17B0(t)
			update := target.UpdateDataPlayer()
			update.State = test.state
			update.Player.ArmorEquip = 0x1000000
			update.Player.WeaponEquip = test.weapon
			var damages []int32
			runtime := playerDamageRuntime4E17B0(t, sound, &damages)
			runtime.BlockSourceExcluded = func(*Object) bool { return false }
			runtime.BlockDirection = func(*Object, types.Pointf) bool {
				if !test.wantTest {
					t.Fatal("direction checked outside shield stance")
				}
				return test.front
			}
			if handled, result := PlayerDamageNative4E17B0(target, source, source, 3, object.DamageBite, runtime); !handled || !result {
				t.Fatalf("unblocked bite = %t/%t", handled, result)
			}
			if !reflect.DeepEqual(damages, []int32{3}) {
				t.Fatalf("unblocked bite damages = %v", damages)
			}
		})
	}
}

func TestPlayerDamageNative4E17B0ShieldExcludesFistType(t *testing.T) {
	target, source, sound := playerDamageFixture4E17B0(t)
	target.UpdateDataPlayer().State = PlayerState16
	target.UpdateDataPlayer().Player.ArmorEquip = 0x1000000
	var damages []int32
	runtime := playerDamageRuntime4E17B0(t, sound, &damages)
	runtime.BlockSourceExcluded = func(got *Object) bool {
		if got != source {
			t.Fatalf("excluded source = %p, want %p", got, source)
		}
		return true
	}
	runtime.BlockDirection = func(*Object, types.Pointf) bool {
		t.Fatal("excluded source checked shield direction")
		return false
	}
	if handled, result := PlayerDamageNative4E17B0(target, source, source, 3, object.DamageBite, runtime); !handled || !result {
		t.Fatalf("excluded source bite = %t/%t", handled, result)
	}
	if !reflect.DeepEqual(damages, []int32{3}) {
		t.Fatalf("excluded source damages = %v", damages)
	}
}

func TestPlayerDamageNative4E17B0ShieldPreflightDoesNotMutate(t *testing.T) {
	target, source, sound := playerDamageFixture4E17B0(t)
	update := target.UpdateDataPlayer()
	update.State = PlayerState16
	update.Field76 = 9
	update.Player.ArmorEquip = 0x1000000
	target.InvFirstItem = &Object{ObjSubClass: object.SubClass(2), ObjFlags: object.FlagEquipped}
	beforeTarget := *target
	beforeUpdate := *update
	var reason string
	runtime := playerDamageRuntime4E17B0(t, sound, new([]int32))
	runtime.Unsupported = func(got string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = got }
	runtime.BlockSourceExcluded = func(*Object) bool { return false }
	runtime.BlockDirection = func(*Object, types.Pointf) bool { return true }
	runtime.BlockDamagePercent = func() float64 { return 1 }
	runtime.Audio = func(int, *Object) { t.Fatal("unsupported block played audio") }
	runtime.DamageBlockItem = func(*Object, *Object, *Object, *Object, float32, object.DamageType) bool {
		t.Fatal("unsupported block damaged shield")
		return false
	}
	runtime.CanDamageBlockItem = func(*Object) bool { return false }
	runtime.PlayerSetState = func(*Object, PlayerState) bool { return true }
	if handled, result := PlayerDamageNative4E17B0(target, source, source, 3, object.DamageBite, runtime); handled || result || reason != "shield durability callback" {
		t.Fatalf("unsupported shield = handled:%t result:%t reason:%q", handled, result, reason)
	}
	if *target != beforeTarget || *update != beforeUpdate {
		t.Fatal("unsupported shield changed player state")
	}
}

func TestPlayerDamageNative4E17B0RejectsLateDefendBeforeMutation(t *testing.T) {
	target, source, sound := playerDamageFixture4E17B0(t)
	marker := unsafe.Pointer(new(byte))
	modifier := &ModifierEff{Defend76: ModifierEffFnc{Fnc: marker}}
	target.InvFirstItem = &Object{
		ObjClass: object.ClassWeapon,
		ObjFlags: object.FlagEquipped,
		InitData: unsafe.Pointer(&ModifierInitData{Modifiers: [4]*ModifierEff{nil, nil, modifier}}),
	}
	beforeTarget := *target
	beforeUpdate := *target.UpdateDataPlayer()
	var reason string
	runtime := playerDamageRuntime4E17B0(t, sound, new([]int32))
	runtime.Unsupported = func(got string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = got }
	handled, result := PlayerDamageNative4E17B0(target, source, source, 3, object.DamageBite, runtime)
	if handled || result || reason != "late equipped-item defend effect" {
		t.Fatalf("result = %t/%t reason %q", handled, result, reason)
	}
	if *target != beforeTarget || *target.UpdateDataPlayer() != beforeUpdate {
		t.Fatal("unsupported branch mutated player state")
	}
}

func TestPlayerDamageNative4E17B0EntryGates(t *testing.T) {
	target, source, sound := playerDamageFixture4E17B0(t)
	var damages []int32
	runtime := playerDamageRuntime4E17B0(t, sound, &damages)
	target.ObjFlags |= object.FlagDead
	if handled, result := PlayerDamageNative4E17B0(target, source, source, 3, object.DamageBite, runtime); !handled || result {
		t.Fatalf("dead gate = %t/%t", handled, result)
	}
	target.ObjFlags &^= object.FlagDead
	target.Buffs |= 1 << playerDamageInvulnerableEnchant4E17B0
	if handled, result := PlayerDamageNative4E17B0(target, source, source, 3, object.DamageBite, runtime); !handled || !result {
		t.Fatalf("invulnerability gate = %t/%t", handled, result)
	}
	if len(damages) != 0 {
		t.Fatalf("entry gate applied damage: %v", damages)
	}
}
