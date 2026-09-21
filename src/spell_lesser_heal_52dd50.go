package opennox

import (
	"math"

	"github.com/opennox/libs/player"
	"github.com/opennox/libs/spell"

	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

const lesserHealAmountKey52DD50 = "LesserHealAmount"

type lesserHealHooks52DD50[O comparable, A any] struct {
	loadTarget    func(A) O
	getHP         func(O) uint16
	getMaxHP      func(O) uint16
	manaCost      func(spell.ID, int) int
	refundMana    func(O, int)
	balance       func(string) float32
	loadClassLow  func(O) uint8
	loadPlayerCls func(O) uint8
	classHealth   func(uint8) float32
	floatToInt    func(float32) int32
	adjustHP      func(O, int32)
	audio         func(spell.ID, O)
}

// lesserHeal52DD50 preserves GAME.EXE 0052DD50 while allowing object handles
// to use the host pointer width. In particular, SpellAcceptArg.Obj must never
// be decoded through the original PE32 uint32 slot on a 64-bit host.
func lesserHeal52DD50[O comparable, A any](
	spellID spell.ID,
	second, third O,
	arg A,
	hooks lesserHealHooks52DD50[O, A],
) int {
	var nilObject O

	// The original function caches the first target load for MaxHP, but then
	// independently reloads the live SpellAcceptArg target for every other use.
	maxTarget := hooks.loadTarget(arg)
	if hooks.loadTarget(arg) == nilObject {
		return 0
	}
	current := hooks.getHP(hooks.loadTarget(arg))
	if current == hooks.getMaxHP(maxTarget) && second == hooks.loadTarget(arg) {
		hooks.refundMana(third, hooks.manaCost(spellID, 1))
		return 1
	}

	amount := hooks.balance(lesserHealAmountKey52DD50)
	if third != nilObject && hooks.loadClassLow(third)&0x04 != 0 {
		class := hooks.loadPlayerCls(third)
		if class <= uint8(player.Conjurer) {
			// Both operands originate as float32. The x87 multiply is exact at
			// double precision before the result is spilled back to float32.
			amount = float32(float64(amount) * float64(hooks.classHealth(class)))
		}
	}

	hooks.adjustHP(hooks.loadTarget(arg), hooks.floatToInt(amount))
	hooks.audio(spellID, hooks.loadTarget(arg))
	return 1
}

// lesserHealFloatToInt52DD50 models nox_float2int's x87 FISTP conversion:
// round-to-nearest-even, with the signed integer-indefinite value for invalid
// or out-of-range inputs.
func lesserHealFloatToInt52DD50(value float32) int32 {
	if math.IsNaN(float64(value)) || value >= 2147483648 || value < -2147483648 {
		return math.MinInt32
	}
	return int32(math.RoundToEven(float64(value)))
}

func castLesserHeal52DD50(
	spellID spell.ID,
	second, third, _ *server.Object,
	arg *server.SpellAcceptArg,
	_ int,
) int {
	s := noxServer
	return lesserHeal52DD50(spellID, second, third, arg, lesserHealHooks52DD50[*server.Object, *server.SpellAcceptArg]{
		loadTarget: func(arg *server.SpellAcceptArg) *server.Object {
			return arg.Obj
		},
		getHP:    server.UnitGetHP4EE780,
		getMaxHP: server.UnitGetMaxHP4EE7A0,
		manaCost: func(id spell.ID, level int) int {
			return s.Spells.ManaCost(id, level)
		},
		refundMana: func(unit *server.Object, amount int) {
			sub_4FD030(unit, amount)
		},
		balance: func(key string) float32 {
			return float32(s.Balance.Float(key))
		},
		loadClassLow: func(unit *server.Object) uint8 {
			return uint8(unit.ObjClass)
		},
		loadPlayerCls: func(unit *server.Object) uint8 {
			return uint8(unit.UpdateDataPlayer().Player.PlayerClass())
		},
		classHealth: func(class uint8) float32 {
			return s.Players.ClassStatsMult(player.Class(class)).Health
		},
		floatToInt: lesserHealFloatToInt52DD50,
		adjustHP: func(target *server.Object, amount int32) {
			legacy.Nox_xxx_unitAdjustHP_4EE460(target, int(amount))
		},
		audio: func(id spell.ID, target *server.Object) {
			s.Audio.EventObj(s.Spells.DefByInd(id).GetOnSound(), target, 0, 0)
		},
	})
}
