package server

import (
	"math"
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func npcMonsterStoneFixture4E17B0(t *testing.T) (*Object, *Object, *Object, PlayerDamageRuntime4E17B0) {
	t.Helper()
	v, _, r := npcChargeFixture4E17B0(t)
	a := damageMeleeUnitFixture4E17B0(t, false)
	a.ObjClass, a.ObjSubClass = object.ClassMonster|object.ClassComplex, 0x70001
	w := &Object{TypeInd: 525, ObjClass: object.ClassSimple | object.ClassMissile, ObjOwner: a, PrevPos: types.Ptf(20, 7)}
	return v, a, w, r
}

// Original case 11 shares the full-armor PIERCE carry, not half-armor CRUSH.
// This caller is the naturally observed Urchin -> ThrowingStone -> NPC hit.
func TestPlayerDamageNPCMonsterSimpleMissileImpact4E17B0Carry(t *testing.T) {
	v, a, w, r := npcMonsterStoneFixture4E17B0(t)
	u := v.UpdateDataMonster()
	u.Field518, u.Field547, u.Field546 = math.Float32bits(0.25), 99, 77
	beforeSource, beforeSourceUD, beforeWeapon := *a, *a.UpdateDataMonster(), *w
	for i, want := range []struct {
		hp    uint16
		carry float32
	}{{198, .25}, {196, .5}, {193, -.25}} {
		if handled, result := PlayerDamageNative4E17B0(v, a, w, 3, object.DamageImpact, r); !handled || !result {
			t.Fatalf("hit%d handled/result=%t/%t", i, handled, result)
		}
		if v.HealthData.Cur != want.hp || math.Float32frombits(u.Field1) != want.carry ||
			u.Field547 != 1 || u.Field546 != 525 || v.Obj130 != w || v.Pos132 != w.PrevPos ||
			v.Field131 != uint32(object.DamageImpact) || v.Frame134 != 1400 {
			t.Fatalf("hit%d HP=%d carry=%g marker=%d/%d attribution=%p pos=%v", i, v.HealthData.Cur, math.Float32frombits(u.Field1), u.Field547, u.Field546, v.Obj130, v.Pos132)
		}
	}
	// The real DefaultDamage enemy-hit tail records the attacking monster's
	// successful-hit frame. All other source/update/projectile fields stay put.
	beforeSourceUD.Field130 = 1400
	if *a != beforeSource || *a.UpdateDataMonster() != beforeSourceUD || *w != beforeWeapon {
		t.Fatal("NPC stone damage changed source/projectile fields beyond the original hit timestamp")
	}
}

func TestPlayerDamageNPCMonsterSimpleMissileImpact4E17B0ArmorQuest(t *testing.T) {
	v, a, w, r := npcMonsterStoneFixture4E17B0(t)
	armor := damageMeleeArmorFixture4E17B0(v, .5, 0)
	armor.HealthData.Cur, armor.HealthData.Max = 100, 100
	damageMeleeArmorRuntime4E17B0(&r, armor, .5)
	var events []string
	r.DamageArmor = func(item, attacker, projectile *Object, amount int32, typ object.DamageType) bool {
		if item != armor || attacker != a || projectile != w || amount != 40 || typ != object.DamageImpact || v.UpdateDataMonster().Field546 != 525 {
			t.Fatal("stone full-armor wear arguments or marker order")
		}
		events = append(events, "armor")
		item.HealthData.Cur -= uint16(amount)
		return true
	}
	r.QuestMode = func() bool { return true }
	r.QuestDamageScale = func() float32 { events = append(events, "quest"); return .5 }
	world := damageMeleeWorldRuntime4E0B30(t)
	world.DamageClear = func(target *Object, amount int32) {
		if target != v || amount != 20 {
			t.Fatal("native default tail amount")
		}
		events = append(events, "HP")
		target.HealthData.Cur -= uint16(amount)
	}
	r.DefaultDamage = func(target, source, weapon *Object, amount int32, typ object.DamageType) bool {
		return DefaultDamageWorld4E0B30(target, source, weapon, amount, typ, world)
	}
	if h, result := PlayerDamageNative4E17B0(v, a, w, 80, object.DamageImpact, r); !h || !result || v.HealthData.Cur != 180 || armor.HealthData.Cur != 60 {
		t.Fatalf("handled/result=%t/%t HP=%d armor=%d", h, result, v.HealthData.Cur, armor.HealthData.Cur)
	}
	if len(events) != 3 || events[0] != "armor" || events[1] != "quest" || events[2] != "HP" {
		t.Fatalf("callback order=%v", events)
	}
}
