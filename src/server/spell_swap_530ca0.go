package server

import (
	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
)

// SpellSwapRuntime530CA0 supplies the game services used by Swap's duration
// callbacks. The record and objects remain native-width Go pointers.
type SpellSwapRuntime530CA0 struct {
	CoopMode      func() bool
	Frame         func() uint32
	TeleportDelay func(uint32) float64
	CanInteract   func(*Object, *Object) bool
	InformNoLOS   func(*Object)
	SendPointFX   func(netmsg.Op, types.Pointf)
	CastSound     func(spell.ID) sound.ID
	Audio         func(sound.ID, *Object, int, uint32)
	Teleport      func(*Object, types.Pointf)
	Attribution   func(*Object, *Object)
}

// SpellSwapCreate530CA0 follows GAME.EXE 00530CA0, including the glyph and
// untargetable-monster gates and the Coop-only TeleportDelay.
func SpellSwapCreate530CA0(record *DurSpell, rt SpellSwapRuntime530CA0) int32 {
	caster, target := record.Caster16, record.Target48
	if caster == nil || byte(record.Flags88)&0x20 != 0 || target == nil ||
		target.ObjFlags.HasAny(object.FlagDead|object.FlagDestroyed) || caster == target ||
		target.ObjClass.Has(object.ClassMonster) && uint32(target.ObjSubClass)&0x4000 != 0 {
		return 1
	}
	if rt.CoopMode() {
		delay := rt.TeleportDelay(record.Level - 1)
		record.Frame68 = rt.Frame() + uint32(playerReadValuesRound4EEDC0(shapeBoxSpill428270(delay)))
	} else {
		record.Frame68 = rt.Frame() + 1
	}
	return 0
}

// spellSwapTargetInPlayerBounds530D30 is the inclusive player rectangle test
// used after the ordinary interaction/LOS check. The source stores each
// computed bound as float32 before testing the target's position. Both the
// arithmetic and spill use startup 53-bit precision toward zero.
func spellSwapTargetInPlayerBounds530D30(caster, target *Object) bool {
	if !caster.ObjClass.Has(object.ClassPlayer) {
		return true
	}
	if caster.UpdateData == nil {
		return false
	}
	data := caster.UpdateDataPlayer()
	if data.Player == nil {
		return false
	}
	halfX := float64(data.Player.Field10)
	halfY := float64(data.Player.Field12)
	x, y := float64(caster.PosVec.X), float64(caster.PosVec.Y)
	minX := shapeBoxSpill428270(shapeBoxAdd53_428270(x, -halfX))
	minY := shapeBoxSpill428270(shapeBoxAdd53_428270(y, -halfY))
	maxX := shapeBoxSpill428270(shapeBoxAdd53_428270(x, halfX))
	maxY := shapeBoxSpill428270(shapeBoxAdd53_428270(y, halfY))
	pos := target.PosVec
	return pos.X >= minX && pos.X <= maxX && pos.Y >= minY && pos.Y <= maxY
}

// SpellSwapUpdate530D30 follows GAME.EXE 00530D30. The target's old position
// is saved before the first teleport; duration links stay live across services.
//
//go:noinline
func SpellSwapUpdate530D30(record *DurSpell, rt SpellSwapRuntime530CA0) int32 {
	target, caster := record.Target48, record.Caster16
	if target == nil || caster == nil ||
		target.ObjFlags.HasAny(object.FlagDead|object.FlagDestroyed) ||
		caster.ObjFlags.HasAny(object.FlagDead|object.FlagDestroyed) {
		return 1
	}
	if record.Frame68-1 != rt.Frame() {
		return 0
	}
	if target.HasEnchant(ENCHANT_ANCHORED) || caster.HasEnchant(ENCHANT_ANCHORED) {
		rt.Audio(sound.ID(231), record.Target48, 0, 0)
		rt.Audio(sound.ID(231), record.Caster16, 0, 0)
		return 1
	}
	if record.Flag20 == 0 &&
		(!rt.CanInteract(record.Caster16, record.Target48) || !spellSwapTargetInPlayerBounds530D30(record.Caster16, record.Target48)) {
		rt.InformNoLOS(record.Caster16)
		return 1
	}
	target = record.Target48
	targetPos := target.PosVec
	rt.Teleport(target, record.Caster16.PosVec)
	rt.Teleport(record.Caster16, targetPos)
	if !record.Target48.HasEnchant(ENCHANT_INVISIBLE) && !record.Caster16.HasEnchant(ENCHANT_INVISIBLE) {
		rt.SendPointFX(netmsg.MSG_FX_TELEPORT, record.Caster16.PosVec)
		rt.SendPointFX(netmsg.MSG_FX_TELEPORT, record.Target48.PosVec)
	}
	// Each audio target is pushed before its corresponding sound lookup.
	play := func(unit *Object) {
		id := rt.CastSound(spell.ID(record.Spell))
		rt.Audio(id, unit, 0, 0)
	}
	play(record.Target48)
	play(record.Caster16)
	rt.Attribution(record.Caster16, record.Target48)
	return 1
}
