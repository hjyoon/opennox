package server

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	playerlib "github.com/opennox/libs/player"
	"github.com/opennox/libs/spell"
)

func TestSpellShieldDamage52F690UsesNativeDurationPointers(t *testing.T) {
	target := &Object{}
	otherTarget := &Object{}
	match := &DurSpell{
		Spell:    uint32(spell.SPELL_SHIELD),
		Target48: target,
		Field72:  11,
	}
	first := &DurSpell{
		Spell:    uint32(spell.SPELL_SHIELD),
		Target48: otherTarget,
		Field72:  99,
		Next:     match,
	}
	SpellShieldDamage52F690(first, target, 4, SpellShieldDamageRuntime52F690{
		Cancel:  func(*DurSpell) { t.Fatal("live Shield duration was cancelled") },
		BuffOff: func(*Object, EnchantID) { t.Fatal("live Shield enchant was removed") },
	})
	if first.Field72 != 99 || match.Field72 != 7 {
		t.Fatalf("duration health = first:%d match:%d, want 99/7", first.Field72, match.Field72)
	}
}

func TestSpellShieldDamage52F690CancelsOrRemovesStaleEnchant(t *testing.T) {
	t.Run("depleted duration", func(t *testing.T) {
		target := &Object{}
		record := &DurSpell{
			Spell:    uint32(spell.SPELL_SHIELD),
			Target48: target,
			Field72:  5,
		}
		var cancelled *DurSpell
		SpellShieldDamage52F690(record, target, 5, SpellShieldDamageRuntime52F690{
			Cancel:  func(got *DurSpell) { cancelled = got },
			BuffOff: func(*Object, EnchantID) { t.Fatal("matched Shield enchant was removed directly") },
		})
		if cancelled != record || record.Field72 != 5 {
			t.Fatalf("cancelled=%p health=%d, want %p/5", cancelled, record.Field72, record)
		}
	})

	t.Run("missing duration", func(t *testing.T) {
		target := &Object{}
		var gotTarget *Object
		var gotEnchant EnchantID
		SpellShieldDamage52F690(nil, target, 3, SpellShieldDamageRuntime52F690{
			Cancel: func(*DurSpell) { t.Fatal("missing duration was cancelled") },
			BuffOff: func(got *Object, enchant EnchantID) {
				gotTarget, gotEnchant = got, enchant
			},
		})
		if gotTarget != target || gotEnchant != ENCHANT_SHIELD {
			t.Fatalf("BuffOff(%p,%d), want (%p,%d)", gotTarget, gotEnchant, target, ENCHANT_SHIELD)
		}
	})
}

func TestSpellShieldReduceDamage52F710Nonlethal(t *testing.T) {
	target := &Object{}
	source := &Object{}
	damage := int32(9)
	var sound int
	var fxTarget, fxSource *Object
	var shieldDamage int32
	SpellShieldReduceDamage52F710(target, &damage, object.DamageBlade, source, SpellShieldReduceDamageRuntime52F710{
		Audio: func(id int, got *Object) {
			if got != target {
				t.Fatalf("Audio target=%p, want %p", got, target)
			}
			sound = id
		},
		ShieldFX:  func(gotTarget, gotSource *Object) { fxTarget, fxSource = gotTarget, gotSource },
		CurrentHP: func(*Object) uint16 { return 10 },
		SetHP:     func(*Object, uint16) { t.Fatal("nonlethal hit changed HP directly") },
		ShieldDamage: func(got *Object, amount int32) {
			if got != target {
				t.Fatalf("ShieldDamage target=%p, want %p", got, target)
			}
			shieldDamage = amount
		},
	})
	if damage != 4 || shieldDamage != 4 || sound != spellShieldHitSound52F710 ||
		fxTarget != target || fxSource != source {
		t.Fatalf("reduction = damage:%d shield:%d sound:%d fx:%p/%p", damage, shieldDamage, sound, fxTarget, fxSource)
	}
}

func TestSpellShieldReduceDamage52F710MinimumOne(t *testing.T) {
	damage := int32(1)
	var shieldDamage int32
	SpellShieldReduceDamage52F710(&Object{}, &damage, object.DamageBlade, nil, SpellShieldReduceDamageRuntime52F710{
		CurrentHP:    func(*Object) uint16 { return 2 },
		ShieldDamage: func(_ *Object, amount int32) { shieldDamage = amount },
	})
	if damage != 1 || shieldDamage != 1 {
		t.Fatalf("minimum reduction = damage:%d shield:%d, want 1/1", damage, shieldDamage)
	}
}

func TestSpellShieldReduceDamage52F710LethalHit(t *testing.T) {
	target := &Object{}
	source := &Object{}
	damage := int32(5)
	var setHP uint16
	var shieldDamage int32
	SpellShieldReduceDamage52F710(target, &damage, object.DamageElectric, source, SpellShieldReduceDamageRuntime52F710{
		CurrentHP: func(*Object) uint16 { return 5 },
		SetHP: func(got *Object, hp uint16) {
			if got != target {
				t.Fatalf("SetHP target=%p, want %p", got, target)
			}
			setHP = hp
		},
		ShieldDamage: func(_ *Object, amount int32) { shieldDamage = amount },
		Frame:        func() uint32 { return 0x12345678 },
	})
	if damage != 0 || setHP != 2 || shieldDamage != 999999 || target.Obj130 != source ||
		target.Field131 != uint32(object.DamageElectric) || target.Frame134 != 0x12345678 {
		t.Fatalf("lethal reduction = damage:%d hp:%d shield:%d source:%p type:%d frame:%#x",
			damage, setHP, shieldDamage, target.Obj130, target.Field131, target.Frame134)
	}
}

func TestSpellShieldReduceDamage52F710WarriorFrontBlock(t *testing.T) {
	player := &Player{}
	player.Info().SetPlayerClass(playerlib.Warrior)
	update := &PlayerUpdateData{Player: player, State: PlayerState16}
	target := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
	source := &Object{}
	damage := int32(5)
	var shieldDamage int32
	SpellShieldReduceDamage52F710(target, &damage, object.DamageBlade, source, SpellShieldReduceDamageRuntime52F710{
		CurrentHP: func(*Object) uint16 { return 5 },
		SetHP:     func(*Object, uint16) { t.Fatal("front-blocking warrior had HP forced to two") },
		FrontBlock: func(gotTarget, gotSource *Object) bool {
			return gotTarget == target && gotSource == source
		},
		ShieldDamage: func(_ *Object, amount int32) { shieldDamage = amount },
	})
	if damage != 1 || shieldDamage != 999999 || target.Obj130 != source {
		t.Fatalf("warrior reduction = damage:%d shield:%d source:%p", damage, shieldDamage, target.Obj130)
	}
}

func TestDefaultDamageWorld4E0B30AppliesShieldReduction(t *testing.T) {
	target := &Object{ObjClass: object.ClassObstacle, Buffs: uint32(1) << ENCHANT_SHIELD}
	source := &Object{ObjClass: object.ClassPlayer}
	weapon := &Object{ObjClass: object.ClassWeapon}
	var cleared int32
	runtime := DefaultDamageWorldRuntime4E0B30{
		GameplayFlag1: func() bool { return true },
		BuffOff:       func(*Object, EnchantID) {},
		ShieldReduce: func(gotTarget *Object, gotDamage *int32, typ object.DamageType, gotSource *Object) {
			if gotTarget != target || *gotDamage != 17 || typ != object.DamageBlade || gotSource != weapon {
				t.Fatalf("ShieldReduce(%p,%d,%d,%p)", gotTarget, *gotDamage, typ, gotSource)
			}
			*gotDamage = 8
		},
		DamageClear: func(got *Object, damage int32) {
			if got != target {
				t.Fatalf("DamageClear target=%p, want %p", got, target)
			}
			cleared = damage
		},
		Unsupported: func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) {
			t.Fatalf("shielded world damage rejected: %s", reason)
		},
	}
	if !DefaultDamageWorld4E0B30(target, source, weapon, 17, object.DamageBlade, runtime) {
		t.Fatal("shielded world damage returned false")
	}
	if cleared != 8 {
		t.Fatalf("cleared damage=%d, want 8", cleared)
	}
}

func TestPlayerDamageNative4E17B0AppliesShieldReduction(t *testing.T) {
	target, _, sound := playerDamageFixture4E17B0(t)
	target.UpdateDataPlayer().Field57 = 0
	target.Buffs = uint32(1) << ENCHANT_SHIELD
	var damages []int32
	runtime := playerDamageRuntime4E17B0(t, sound, &damages)
	runtime.ShieldReduce = func(gotTarget *Object, gotDamage *int32, typ object.DamageType, gotSource *Object) {
		if gotTarget != target || *gotDamage != 5 || typ != object.DamageLava || gotSource != nil {
			t.Fatalf("ShieldReduce(%p,%d,%d,%p)", gotTarget, *gotDamage, typ, gotSource)
		}
		*gotDamage = 2
	}
	if handled, result := PlayerDamageNative4E17B0(target, nil, nil, 5, object.DamageLava, runtime); !handled || !result {
		t.Fatalf("shielded LAVA = handled:%t result:%t", handled, result)
	}
	if len(damages) != 1 || damages[0] != 2 {
		t.Fatalf("shielded player damages=%v, want [2]", damages)
	}
}
