package server

import (
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/player"
	"github.com/opennox/libs/spell"
)

const (
	spellShieldEnchant52F690  = EnchantID(26)
	spellShieldHitSound52F710 = 131
)

// SpellShieldDamageRuntime52F690 supplies the two state-changing operations
// reached by GAME.EXE 0052F690. Duration and Object pointers remain native
// width throughout the list walk.
type SpellShieldDamageRuntime52F690 struct {
	Cancel  func(*DurSpell)
	BuffOff func(*Object, EnchantID)
}

// SpellShieldDamage52F690 consumes health from target's Shield duration. The
// spell record is cancelled when depleted or when its target is already being
// destroyed. A stale Shield enchant without a duration record is removed.
func SpellShieldDamage52F690(
	first *DurSpell,
	target *Object,
	amount int32,
	runtime SpellShieldDamageRuntime52F690,
) {
	for record := first; record != nil; record = record.Next {
		if record.Target48 != target || record.Spell != uint32(spell.SPELL_SHIELD) {
			continue
		}
		if record.Target48 == nil || record.Target48.ObjFlags.HasAny(object.FlagDestroyed|object.FlagDead) ||
			record.Field72-amount <= 0 {
			if runtime.Cancel != nil {
				runtime.Cancel(record)
			}
			return
		}
		record.Field72 -= amount
		return
	}
	if runtime.BuffOff != nil {
		runtime.BuffOff(target, spellShieldEnchant52F690)
	}
}

// SpellShieldReduceDamageRuntime52F710 supplies services outside the native
// Object and Player layouts used by GAME.EXE 0052F710.
type SpellShieldReduceDamageRuntime52F710 struct {
	Audio        func(int, *Object)
	ShieldFX     func(target, source *Object)
	CurrentHP    func(*Object) uint16
	SetHP        func(*Object, uint16)
	ShieldDamage func(*Object, int32)
	FrontBlock   func(target, source *Object) bool
	Frame        func() uint32
}

// SpellShieldReduceDamage52F710 applies the Shield enchant's damage reduction
// without narrowing target, source, Player, or duration pointers. Nonlethal
// damage is halved (with a minimum of one). A lethal hit consumes the Shield,
// leaves the target at two HP, and suppresses the pending damage; the original
// warrior shield-stance exception instead lets one point through.
func SpellShieldReduceDamage52F710(
	target *Object,
	damage *int32,
	typ object.DamageType,
	source *Object,
	runtime SpellShieldReduceDamageRuntime52F710,
) {
	if target == nil || damage == nil {
		return
	}
	if runtime.Audio != nil {
		runtime.Audio(spellShieldHitSound52F710, target)
	}
	if runtime.ShieldFX != nil {
		runtime.ShieldFX(target, source)
	}

	currentHP := UnitGetHP4EE780(target)
	if runtime.CurrentHP != nil {
		currentHP = runtime.CurrentHP(target)
	}
	incoming := *damage
	if int32(currentHP) > incoming {
		reduced := incoming / 2
		if reduced == 0 {
			reduced = 1
		}
		*damage = reduced
		if runtime.ShieldDamage != nil {
			runtime.ShieldDamage(target, reduced)
		}
		return
	}

	warriorFrontBlock := false
	if target.ObjClass.Has(object.ClassPlayer) && target.UpdateData != nil && source != nil {
		update := target.UpdateDataPlayer()
		warriorFrontBlock = update.Player != nil && update.Player.PlayerClass() == player.Warrior &&
			update.State == PlayerState16 && runtime.FrontBlock != nil && runtime.FrontBlock(target, source)
	}
	if warriorFrontBlock {
		*damage = 1
	} else {
		if runtime.SetHP != nil {
			runtime.SetHP(target, 2)
		}
		*damage = 0
	}
	if runtime.ShieldDamage != nil {
		runtime.ShieldDamage(target, 999999)
	}
	target.Obj130 = source
	target.Field131 = uint32(typ)
	if runtime.Frame != nil {
		target.Frame134 = runtime.Frame()
	} else {
		target.Frame134 = 0
	}
}
