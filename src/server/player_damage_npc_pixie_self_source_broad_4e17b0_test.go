package server

import (
	"math"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func npcPixieSelfSourceFixture4E17B0(t *testing.T) (*Object, *Object, PlayerDamageRuntime4E17B0) {
	t.Helper()
	target := damageMeleeUnitFixture4E17B0(t, false)
	target.TypeInd, target.ObjClass, target.ObjSubClass, target.ObjFlags = 1343, object.Class(2097154), object.SubClass(69778), object.Flags(16777732)
	pixie := &Object{TypeInd: 672, ObjClass: object.Class(1572865), ObjSubClass: 3, ObjFlags: object.Flags(16794116), PrevPos: types.Ptf(31, -17)}
	r := damageMeleeRuntimeFixture4E17B0(t)
	r.CoopMode = func() bool { return false }
	r.GodMode = func() bool { t.Fatal("NPC self-sourced missile queried player GodMode"); return false }
	return target, pixie, r
}

// These numeric expectations are looked up from the unchanged original
// case-11 physical-tail fixture; no candidate arithmetic supplies them.
func TestPlayerDamageNPCPixieSelfSourceImpact4E17B0ArmorCarry(t *testing.T) {
	for _, c := range []struct {
		name         string
		raw          int32
		armor, carry float32
		quest        bool
		scale        float32
	}{
		{"reported-672-to-1343", 8, .5, 0, false, 1},
		{"fraction-first", 3, .25, 0, false, 1},
		{"fraction-next", 3, .25, .25, false, 1},
		{"minimum-one", 1, 1, .125, false, 1},
		{"quest", 5, .5, .5, true, .5},
		{"negative", -3, .25, .125, false, 1},
		{"zero", 0, .25, .125, false, 1},
	} {
		t.Run(c.name, func(t *testing.T) {
			v, p, r := npcPixieSelfSourceFixture4E17B0(t)
			before := *p
			u := v.UpdateDataMonster()
			u.Field518, u.Field1, u.Field547, u.Field546 = math.Float32bits(c.armor), math.Float32bits(c.carry), 99, 88
			original := physicalOriginal4E17B0(t, object.DamageImpact, false, c.raw, c.armor, c.carry, c.quest, c.scale)
			r.QuestMode = func() bool { return c.quest }
			r.QuestDamageScale = func() float32 { return c.scale }
			calls := 0
			r.DefaultDamage = func(got, a, w *Object, d int32, typ object.DamageType) bool {
				calls++
				if got != v || a != p || w != p || d != original.Damage || typ != object.DamageImpact || u.Field1 != original.Carry || u.Field547 != 1 || u.Field546 != 672 {
					t.Fatalf("existing case11 tail: amount=%d/%d carry=%08x/%08x", d, original.Damage, u.Field1, original.Carry)
				}
				return true
			}
			if h, ok := PlayerDamageNative4E17B0(v, p, p, c.raw, object.DamageImpact, r); !h || !ok || calls != 1 || *p != before || p.UpdateData != nil || v.HealthData.Cur != 200 {
				t.Fatal("self-sourced pure missile tail or unchanged source/HP stub boundary")
			}
		})
	}
}

func TestPlayerDamageNPCPixieSelfSourceImpact4E17B0CachedAndLiveUpdate(t *testing.T) {
	v, p, r := npcPixieSelfSourceFixture4E17B0(t)
	cached := v.UpdateDataMonster()
	cached.Field518, cached.Field1, cached.Field547, cached.Field546 = math.Float32bits(.5), math.Float32bits(.125), 99, 88
	live := &MonsterUpdateData{Field518: math.Float32bits(.875), Field1: math.Float32bits(.5), Field547: 55, Field546: 66}
	original := physicalOriginal4E17B0(t, object.DamageImpact, false, 5, .5, .5, true, .5)
	v.Buffs = 1 << playerDamageReflectEnchant4E17B0
	directions, calls := 0, 0
	r.BlockDirection = func(got *Object, _ types.Pointf) bool {
		directions++
		if got != v || v.UpdateDataMonster() != cached {
			t.Fatal("entry target/update before defense")
		}
		v.UpdateData = unsafe.Pointer(live)
		return false
	}
	r.QuestMode = func() bool { return true }
	r.QuestDamageScale = func() float32 { return .5 }
	r.DefaultDamage = func(got, a, w *Object, d int32, typ object.DamageType) bool {
		calls++
		if got != v || a != p || w != p || d != original.Damage || typ != object.DamageImpact || v.UpdateDataMonster() != live || cached.Field547 != 1 || cached.Field546 != 672 || live.Field1 != original.Carry || live.Field547 != 55 || live.Field546 != 66 || cached.Field1 != math.Float32bits(.125) {
			t.Fatal("cached absorption/marker and live carry split")
		}
		return true
	}
	if h, ok := PlayerDamageNative4E17B0(v, p, p, 5, object.DamageImpact, r); !h || !ok || calls != 1 || directions != 1 || p.UpdateData != nil {
		t.Fatal("live replacement call order or accidental source-unit dependency")
	}
}

func TestPlayerDamageNPCPixieSelfSourceImpact4E17B0ArmorWearLiveQuest(t *testing.T) {
	v, p, r := npcPixieSelfSourceFixture4E17B0(t)
	armor := damageMeleeArmorFixture4E17B0(v, .5, .5)
	armor.HealthData.Cur, armor.HealthData.Max = 100, 100
	damageMeleeArmorRuntime4E17B0(&r, armor, .5)
	original := physicalOriginal4E17B0(t, object.DamageImpact, false, 5, .5, .5, true, .5)
	quest := false
	var events []string
	r.DamageArmor = func(item, a, w *Object, d int32, typ object.DamageType) bool {
		if item != armor || a != p || w != p || d != original.Armor || typ != object.DamageImpact || v.UpdateDataMonster().Field546 != 672 {
			t.Fatal("full armor wear/source identity before live Quest")
		}
		events = append(events, "armor")
		quest = true
		return true
	}
	r.QuestMode = func() bool { events = append(events, "quest"); return quest }
	r.QuestDamageScale = func() float32 { events = append(events, "scale"); return .5 }
	r.DefaultDamage = func(got, a, w *Object, d int32, typ object.DamageType) bool {
		if got != v || a != p || w != p || d != original.Damage || typ != object.DamageImpact || v.UpdateDataMonster().Field1 != original.Carry {
			t.Fatal("live Quest amount/carry after armor")
		}
		events = append(events, "default")
		return true
	}
	if h, ok := PlayerDamageNative4E17B0(v, p, p, 5, object.DamageImpact, r); !h || !ok || !slices.Equal(events, []string{"armor", "quest", "scale", "default"}) {
		t.Fatalf("shared tail events=%v", events)
	}
}

func TestPlayerDamageNPCPixieSelfSourceImpact4E17B0AdjacentAdmission(t *testing.T) {
	for _, name := range []string{"mixed-unit", "weapon-class", "wand-class", "non-missile-admission-only", "other-damage-admission-only", "distinct-world-source"} {
		t.Run(name, func(t *testing.T) {
			v, p, r := npcPixieSelfSourceFixture4E17B0(t)
			source := p
			typ := object.DamageImpact
			switch name {
			case "mixed-unit":
				p.ObjClass |= object.ClassMonster
			case "weapon-class":
				p.ObjClass |= object.ClassWeapon
			case "wand-class":
				p.ObjClass |= object.ClassWand
			case "non-missile-admission-only":
				p.ObjClass &^= object.ClassMissile
			case "other-damage-admission-only":
				typ = object.DamageBlade
			case "distinct-world-source":
				source = &Object{ObjClass: object.ClassSimple}
			}
			if source == p && playerDamageSpellMissileImpactShape4E17B0(source, p, typ) {
				t.Fatal("near-miss entered the new self-source admission")
			}
			// Other damage, nonmissile and distinct-source shapes retain existing routes;
			// this control rejects only the new IMPACT admission, not those routes.
			if name == "other-damage-admission-only" || name == "non-missile-admission-only" || name == "distinct-world-source" {
				return
			}
			why := ""
			r.Unsupported = func(s string, _, _, _ *Object, _ int32, _ object.DamageType) { why = s }
			r.DefaultDamage = func(_, _, _ *Object, _ int32, _ object.DamageType) bool {
				t.Fatal("unadmitted neighboring shape reached shared tail")
				return false
			}
			if h, ok := PlayerDamageNative4E17B0(v, source, p, 8, typ, r); h || ok || why == "" || v.HealthData.Cur != 200 {
				t.Fatal("existing unsupported neighboring contract changed")
			}
		})
	}
}

func TestPlayerDamageNPCPixieSelfSourceImpact4E17B0ActualHPAttribution(t *testing.T) {
	for _, name := range []string{"reported-zero-armor", "reported-half-armor"} {
		t.Run(name, func(t *testing.T) {
			v, p, r := npcPixieSelfSourceFixture4E17B0(t)
			before := *p
			amount := int32(8)
			if name == "reported-half-armor" {
				v.UpdateDataMonster().Field518 = math.Float32bits(.5)
				amount = physicalOriginal4E17B0(t, object.DamageImpact, false, 8, .5, 0, false, 1).Damage
			}
			world := damageMeleeWorldRuntime4E0B30(t)
			damaged, enemy := 0, 0
			world.IsEnemy = func(got, a *Object) bool {
				if got != v || a != p {
					t.Fatal("actual nonItem enemy query identity")
				}
				enemy++
				return false
			}
			world.MonsterHasHitSound = func(*Object) bool { t.Fatal("nonunit Pixie entered monster source update"); return false }
			world.DamageClear = func(got *Object, d int32) {
				if got != v || d != amount {
					t.Fatal("actual HP amount")
				}
				v.HealthData.Cur -= uint16(d)
				damaged++
			}
			r.DefaultDamage = func(got, a, w *Object, d int32, typ object.DamageType) bool {
				return DefaultDamageWorld4E0B30(got, a, w, d, typ, world)
			}
			if h, ok := PlayerDamageNative4E17B0(v, p, p, 8, object.DamageImpact, r); !h || !ok || damaged != 1 || enemy != 1 || v.HealthData.Cur != 200-uint16(amount) || v.Obj130 != p || v.Pos132 != p.PrevPos || v.Field131 != 11 || v.Frame134 != 1400 || v.UpdateDataMonster().Field547 != 2 || v.UpdateDataMonster().Field546 != 11 || !v.UpdateDataMonster().StatusFlags.Has(object.MonStatusInjured) || *p != before {
				t.Fatal("actual default HP/attribution or source immutability")
			}
		})
	}
}

func TestPixieCollideNative4EA080NPCSelfSourcePlayerDamage(t *testing.T) {
	v, p, r := npcPixieSelfSourceFixture4E17B0(t)
	p.CollideData = unsafe.Pointer(&ProjectileCollideData{Damage: 8, Field4: 0x12345678})
	deps := defaultPixieCollideNativeDeps4EA080()
	var events []string
	deps.findParent = func(got *Object) *Object {
		if got != p {
			t.Fatal("Pixie parent lookup input")
		}
		events = append(events, "parent")
		return got.FindOwnerChainPlayer()
	}
	deps.targetDamage = func(got, a, w *Object, d int32, typ object.DamageType) int32 {
		if got != v || a != p || w != p || d != 8 || typ != object.DamageImpact {
			t.Fatal("actual unowned Pixie collider callback transport")
		}
		events = append(events, "PlayerDamage")
		h, ok := PlayerDamageNative4E17B0(got, a, w, d, typ, r)
		if !h || !ok {
			t.Fatal("actual PlayerDamage callback rejected self-source Pixie")
		}
		return 1
	}
	deps.audio = func(id uint32, got *Object) {
		if id != 96 || got != p || v.HealthData.Cur != 192 {
			t.Fatal("collision audio preceded actual HP")
		}
		events = append(events, "audio")
	}
	deps.delayedDelete = func(got *Object) {
		if got != p {
			t.Fatal("collision delete identity")
		}
		events = append(events, "delete")
	}
	pixieCollideNative4EA080(p, v, nil, deps)
	if !slices.Equal(events, []string{"parent", "PlayerDamage", "audio", "delete"}) || v.HealthData.Cur != 192 || v.Obj130 != p || v.Field131 != 11 || v.UpdateDataMonster().Field547 != 2 || v.UpdateDataMonster().Field546 != 11 || p.UpdateData != nil || (*ProjectileCollideData)(p.CollideData).Field4 != 0x12345678 {
		t.Fatal("PixieCollide→NPC PlayerDamage→DefaultDamage integration")
	}
}
