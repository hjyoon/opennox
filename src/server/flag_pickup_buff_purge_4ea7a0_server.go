package server

import (
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/things"
)

// FlagPickupBuffPurgeRuntime4EA7A0 supplies the legacy buff-removal effect.
// The callback returns the original nox_xxx_spellBuffOff_4FF5B0 EAX value.
type FlagPickupBuffPurgeRuntime4EA7A0 struct {
	BuffOff func(*Object, EnchantID) int32
}

type flagPickupBuffPurgeNativeDeps4EA7A0 struct {
	hasBuff       func(*Object, EnchantID) int32
	enchantSpell  func(EnchantID) int32
	spellHasFlags func(int32, uint32) int32
	buffOff       func(*Object, EnchantID) int32
}

func flagPickupBuffPurgeNative4EA7A0(
	obj *Object,
	deps flagPickupBuffPurgeNativeDeps4EA7A0,
) int32 {
	return flagPickupBuffPurge4EA7A0(obj, flagPickupBuffPurgeHooks4EA7A0[*Object]{
		hasBuff: func(obj *Object, enchant uint32) int32 {
			return deps.hasBuff(obj, EnchantID(enchant))
		},
		enchantSpell: func(enchant uint32) int32 {
			return deps.enchantSpell(EnchantID(enchant))
		},
		spellHasFlags: deps.spellHasFlags,
		buffOff: func(obj *Object, enchant uint32) int32 {
			return deps.buffOff(obj, EnchantID(enchant))
		},
	})
}

func flagPickupBuffPurgeServerDeps4EA7A0(
	s *Server,
	runtime FlagPickupBuffPurgeRuntime4EA7A0,
) flagPickupBuffPurgeNativeDeps4EA7A0 {
	return flagPickupBuffPurgeNativeDeps4EA7A0{
		hasBuff: func(obj *Object, enchant EnchantID) int32 {
			if obj.HasEnchant(enchant) {
				return 1
			}
			return 0
		},
		enchantSpell: func(enchant EnchantID) int32 {
			return int32(enchant.Spell())
		},
		spellHasFlags: func(ind int32, flags uint32) int32 {
			if s.Spells.HasFlags(spell.ID(ind), things.SpellFlags(flags)) {
				return 1
			}
			return 0
		},
		buffOff: runtime.BuffOff,
	}
}

// FlagPickupBuffPurge4EA7A0 binds the exact 32-slot purge contract to native
// Object buff storage and server spell definitions.
func (s *Server) FlagPickupBuffPurge4EA7A0(
	obj *Object,
	runtime FlagPickupBuffPurgeRuntime4EA7A0,
) int32 {
	return flagPickupBuffPurgeNative4EA7A0(obj, flagPickupBuffPurgeServerDeps4EA7A0(s, runtime))
}
