package server

import (
	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
)

type SpellTeleportPopRuntime530820 struct {
	CoopMode      func() bool
	Frame         func() uint32
	TickRate      func() uint32
	TeleportDelay func(uint32) float32
	RandomIndex   func() int
	NewObject     func(string) *Object
	CreateAt      func(*Object, *Object, types.Pointf)
	SendPointFX   func(netmsg.Op, types.Pointf)
	CastSound     func(spell.ID) sound.ID
	Audio         func(sound.ID, *Object, int, uint32)
	Teleport      func(*Object, types.Pointf)
	DelayedDelete func(*Object)
	Attribution   func(*Object, *Object)
}

// SpellTeleportPopCreate530820 is GAME.EXE 00530820's scheduling callback.
func SpellTeleportPopCreate530820(record *DurSpell, rt SpellTeleportPopRuntime530820) int32 {
	if rt.CoopMode() {
		delay := rt.TeleportDelay(record.Level - 1)
		record.Frame68 = rt.Frame() + uint32(spellDurationRoundNearestEven(delay))
	} else {
		record.Frame68 = rt.Frame() + 1
	}
	return 0
}

// SpellTeleportPopUpdate530880 follows GAME.EXE 00530880. Marker pointers
// are read from the native PlayerUpdateData array; its four charges remain
// packed as bytes in Field39, matching the original fixed-width count field.
//
//go:noinline
func SpellTeleportPopUpdate530880(record *DurSpell, rt SpellTeleportPopRuntime530820) int32 {
	caster, target := record.Caster16, record.Target48
	if caster == nil || target == nil ||
		caster.ObjFlags.HasAny(object.FlagDead|object.FlagDestroyed) ||
		target.ObjFlags.HasAny(object.FlagDead|object.FlagDestroyed) {
		return 1
	}
	if record.Frame68-1 != rt.Frame() {
		return 0
	}
	if target.HasEnchant(ENCHANT_ANCHORED) {
		rt.Audio(sound.ID(231), target, 0, 0)
		return 1
	}
	if caster.ObjClass.Has(object.ClassPlayer) {
		if caster.UpdateData == nil {
			return 1
		}
		data := caster.UpdateDataPlayer()
		var available bool
		for _, marker := range data.Field29 {
			available = available || marker != nil
		}
		if !available {
			return 1
		}
		index := rt.RandomIndex()
		for index >= 0 && index < len(data.Field29) && data.Field29[index] == nil {
			index = rt.RandomIndex()
		}
		if index < 0 || index >= len(data.Field29) {
			return 1
		}
		marker := data.Field29[index]
		origin := target.PosVec
		wake := rt.NewObject("TeleportWake")
		if wake != nil && wake.CollideData != nil {
			(*TeleportWakeCollideData)(wake.CollideData).Destination = marker.PosVec
			rt.CreateAt(wake, target, origin)
			wake.Field34 = rt.Frame() + rt.TickRate()
		}
		// GAME.EXE reloads the live marker position after creating the wake.
		rt.Teleport(target, data.Field29[index].PosVec)
		if !target.HasEnchant(ENCHANT_INVISIBLE) {
			rt.SendPointFX(netmsg.MSG_FX_TELEPORT, origin)
			rt.SendPointFX(netmsg.MSG_FX_TELEPORT, target.PosVec)
		}
		rt.Audio(rt.CastSound(spell.ID(record.Spell)), target, 0, 0)
		shift := uint(index * 8)
		charge := byte(data.Field39 >> shift)
		charge--
		data.Field39 = data.Field39&^(uint32(0xff)<<shift) | uint32(charge)<<shift
		if charge == 0 {
			rt.SendPointFX(netmsg.MSG_FX_BLUE_SPARKS, data.Field29[index].PosVec)
			rt.DelayedDelete(data.Field29[index])
			data.Field29[index] = nil
		}
	}
	rt.Attribution(caster, target)
	return 1
}
