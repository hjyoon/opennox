package server

import (
	"testing"

	"github.com/opennox/libs/spell"
)

func TestFlagPickupBuffPurgeNative4EA7A0UsesNativeBuffsAndEnchantIDs(t *testing.T) {
	obj := &Object{Buffs: uint32(1) << ENCHANT_ANTI_MAGIC}
	var first, second int32
	lookups := 0
	got := flagPickupBuffPurgeNative4EA7A0(obj, flagPickupBuffPurgeNativeDeps4EA7A0{
		hasBuff: func(got *Object, enchant EnchantID) int32 {
			if got != obj {
				t.Fatal("wrong object")
			}
			if got.HasEnchant(enchant) {
				return 1
			}
			return 0
		},
		enchantSpell: func(enchant EnchantID) int32 {
			lookups++
			if enchant != ENCHANT_ANTI_MAGIC {
				t.Fatalf("enchant = %v", enchant)
			}
			if lookups == 1 {
				first = int32(spell.SPELL_NULLIFY)
				return first
			}
			second = int32(spell.SPELL_NULLIFY) + 1
			return second
		},
		spellHasFlags: func(ind int32, flags uint32) int32 {
			if ind != second || flags != 0x80000 {
				t.Fatalf("HasFlags args = %d/%#x", ind, flags)
			}
			return 1
		},
		buffOff: func(got *Object, enchant EnchantID) int32 {
			if got != obj || enchant != ENCHANT_ANTI_MAGIC {
				t.Fatalf("BuffOff args = %p/%v", got, enchant)
			}
			return 0x5a5a
		},
	})
	if first == 0 || second == first || lookups != 2 {
		t.Fatalf("lookup results/count = %d/%d/%d", first, second, lookups)
	}
	if got != 0 {
		t.Fatalf("slot 31 result = %#x", got)
	}
}

func TestFlagPickupBuffPurge4EA7A0ServerBindingNilObject(t *testing.T) {
	s := &Server{}
	if got := s.FlagPickupBuffPurge4EA7A0(nil, FlagPickupBuffPurgeRuntime4EA7A0{}); got != 0 {
		t.Fatalf("result = %#x", got)
	}
}
