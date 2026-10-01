package opennox

import (
	"fmt"
	"unsafe"

	"github.com/opennox/libs/player"
	"github.com/opennox/libs/spell"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func e2ePlayerStatusSpell(kind string) (id spell.ID, animation, durationKey string, ok bool) {
	switch kind {
	case "confused":
		return spell.SPELL_CONFUSE, "confused", "ConfuseEnchantDuration", true
	case "stun":
		return spell.SPELL_STUN, "stun", "StunEnchantDuration", true
	case "stun-warrior":
		return spell.SPELL_STUN, "slow", "StunEnchantDuration", true
	default:
		return 0, "", "", false
	}
}

func e2ePlayerStatusSpellArg(unit *server.Object) (*server.SpellAcceptArg, func()) {
	// alloc.New is C-heap new, not a struct initializer.
	arg, free := alloc.New(server.SpellAcceptArg{})
	*arg = server.SpellAcceptArg{Obj: unit, Pos: unit.Pos()}
	return arg, free
}

// CheckPlayerStatusSpell enters the regular spell selector, including the
// actual six-argument C cast entry and buff/packet/render/expiry pipeline.
// Only the self-targeted, level-one server cast is a fixture: this does not
// simulate an enemy hit, player incantation input, or mana cost. Class comes
// from the real host menu; balance, client buffs and timers are not changed.
func (sc *e2eScenario) CheckPlayerStatusSpell(kind, name string) {
	id, animation, key, ok := e2ePlayerStatusSpell(kind)
	if !ok {
		e2eError(fmt.Errorf("unknown player status spell %q", kind))
		return
	}
	sc.checkPlayerStatusAnimation(animation, name, "player status spell "+kind, 3000, func(unit *server.Object, buff server.EnchantID) uint32 {
		p := unit.ControllingPlayer()
		if p == nil {
			e2eError(fmt.Errorf("status spell requires the actual host player"))
			return 0
		}
		class := p.PlayerClass()
		if id == spell.SPELL_STUN && (kind == "stun-warrior") != (class == player.Warrior) {
			e2eError(fmt.Errorf("status spell %s has unexpected actual player class %s", kind, class))
			return 0
		}
		balance := noxServer.Balance.Float(key)
		want := uint16(int16(lesserHealFloatToInt52DD50(float32(balance))))
		if want <= 24 || want > 2900 {
			e2eError(fmt.Errorf("stock status duration is outside the bounded visual scenario: key=%s balance=%g timer=%d", key, balance, want))
			return 0
		}
		arg, free := e2ePlayerStatusSpellArg(unit)
		defer free()
		before := *arg
		got := noxServer.SpellAccept4FD400(id, unit, unit, unit, arg, 1)
		if got != 1 || !unit.HasEnchant(buff) || unit.EnchantDur(buff) != int(want) || unit.EnchantPower(buff) != 1 || *arg != before {
			e2eError(fmt.Errorf("status spell dispatch mismatch: spell=%s class=%s result=%d buff=%d buffs=%#x timer=%d/%d power=%d arg-changed=%t", id, class, got, buff, unit.Buffs, unit.EnchantDur(buff), want, unit.EnchantPower(buff), *arg != before))
			return 0
		}
		e2eLog.Printf("STATUS SPELL CAST: kind=%s spell=%s class=%s target=%p native-high=%t enchant=%d balance=%g duration=%d power=%d", kind, id, class, unit, uintptr(unsafe.Pointer(unit)) > 0xffffffff, buff, balance, want, unit.EnchantPower(buff))
		return uint32(want)
	})
}
