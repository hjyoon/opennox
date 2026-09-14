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
	TeleportDelay func(uint32) float32
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
		record.Frame68 = rt.Frame() + uint32(spellDurationRoundNearestEven(delay))
	} else {
		record.Frame68 = rt.Frame() + 1
	}
	return 0
}

// spellSwapTargetInPlayerBounds530D30 is the inclusive player rectangle test
// used after the ordinary interaction/LOS check. The source stores each
// computed bound as float32 before testing the target's position.
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
	minX, minY := float32(x-halfX), float32(y-halfY)
	maxX, maxY := float32(x+halfX), float32(y+halfY)
	pos := target.PosVec
	return pos.X >= minX && pos.X <= maxX && pos.Y >= minY && pos.Y <= maxY
}

// SpellSwapUpdate530D30 follows GAME.EXE 00530D30. Both teleport destinations
// are captured before moving either object, so the two locations are swapped.
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
		rt.Audio(sound.ID(231), target, 0, 0)
		rt.Audio(sound.ID(231), caster, 0, 0)
		return 1
	}
	if record.Flag20 == 0 &&
		(!rt.CanInteract(caster, target) || !spellSwapTargetInPlayerBounds530D30(caster, target)) {
		rt.InformNoLOS(caster)
		return 1
	}
	casterPos, targetPos := caster.PosVec, target.PosVec
	rt.Teleport(target, casterPos)
	rt.Teleport(caster, targetPos)
	if !target.HasEnchant(ENCHANT_INVISIBLE) && !caster.HasEnchant(ENCHANT_INVISIBLE) {
		rt.SendPointFX(netmsg.MSG_FX_TELEPORT, caster.PosVec)
		rt.SendPointFX(netmsg.MSG_FX_TELEPORT, target.PosVec)
	}
	rt.Audio(rt.CastSound(spell.ID(record.Spell)), target, 0, 0)
	rt.Audio(rt.CastSound(spell.ID(record.Spell)), caster, 0, 0)
	rt.Attribution(caster, target)
	return 1
}
