package server

import (
	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
)

type SpellTeleportToMarkRuntime5305D0 struct {
	CoopMode      func() bool
	Frame         func() uint32
	TickRate      func() uint32
	TeleportDelay func(uint32) float32
	NewObject     func(string) *Object
	CreateAt      func(*Object, *Object, types.Pointf)
	SendPointFX   func(netmsg.Op, types.Pointf)
	OnSound       func(spell.ID) sound.ID
	Audio         func(sound.ID, *Object, int, uint32)
	Teleport      func(*Object, types.Pointf)
	DelayedDelete func(*Object)
	Attribution   func(*Object, *Object)
}

// The four Teleport Other To Mark IDs precede the four Teleport To Mark IDs.
// Both families use the same four player glyph slots; a raw ID subtraction
// from the latter family would instead address unrelated player update data.
func teleportMarkIndex530650(id spell.ID) (int, bool) {
	switch {
	case id >= spell.SPELL_TELEPORT_OTHER_TO_MARK_1 && id <= spell.SPELL_TELEPORT_OTHER_TO_MARK_4:
		return int(id - spell.SPELL_TELEPORT_OTHER_TO_MARK_1), true
	case id >= spell.SPELL_TELEPORT_TO_MARK_1 && id <= spell.SPELL_TELEPORT_TO_MARK_4:
		return int(id - spell.SPELL_TELEPORT_TO_MARK_1), true
	default:
		return 0, false
	}
}

// SpellTeleportToMarkCreate5305D0 is the native-width scheduling callback.
func SpellTeleportToMarkCreate5305D0(record *DurSpell, rt SpellTeleportToMarkRuntime5305D0) int32 {
	source := record.Obj12
	if source == nil {
		return 1
	}
	if source.ObjClass.Has(object.ClassPlayer) {
		index, ok := teleportMarkIndex530650(spell.ID(record.Spell))
		if !ok || source.UpdateData == nil || source.UpdateDataPlayer().Field29[index] == nil {
			return 1
		}
	}
	if rt.CoopMode() {
		delay := rt.TeleportDelay(record.Level - 1)
		record.Frame68 = rt.Frame() + uint32(spellDurationRoundNearestEven(delay))
	} else {
		record.Frame68 = rt.Frame() + 1
	}
	return 0
}

// SpellTeleportToMarkUpdate530650 mirrors GAME.EXE 00530650 without reading
// PE32 object and duration offsets from native-width records.
func SpellTeleportToMarkUpdate530650(record *DurSpell, rt SpellTeleportToMarkRuntime5305D0) int32 {
	source, target := record.Obj12, record.Target48
	if source == nil || source.ObjFlags.Has(object.FlagDead) || target == nil ||
		target.ObjFlags.HasAny(object.FlagDead|object.FlagDestroyed) ||
		target != source && record.Flag20 == 0 {
		return 1
	}
	if record.Frame68-1 != rt.Frame() {
		return 0
	}
	if target.HasEnchant(ENCHANT_ANCHORED) {
		rt.Audio(sound.ID(231), target, 0, 0)
		return 1
	}
	if source.ObjClass.Has(object.ClassPlayer) {
		index, ok := teleportMarkIndex530650(spell.ID(record.Spell))
		if !ok || source.UpdateData == nil {
			return 1
		}
		data := source.UpdateDataPlayer()
		marker := data.Field29[index]
		if marker == nil || marker.ObjFlags.Has(object.FlagDestroyed) {
			return 1
		}
		origin := target.PosVec
		wake := rt.NewObject("TeleportWake")
		if wake != nil && wake.CollideData != nil {
			(*TeleportWakeCollideData)(wake.CollideData).Destination = marker.PosVec
			rt.CreateAt(wake, target, origin)
			wake.Field34 = rt.Frame() + rt.TickRate()
		}
		rt.Audio(rt.OnSound(spell.ID(record.Spell)), target, 0, 0)
		rt.Teleport(target, data.Field29[index].PosVec)
		if !target.HasEnchant(ENCHANT_INVISIBLE) {
			rt.SendPointFX(netmsg.MSG_FX_TELEPORT, origin)
			rt.SendPointFX(netmsg.MSG_FX_TELEPORT, target.PosVec)
			rt.Audio(rt.OnSound(spell.ID(record.Spell)), target, 0, 0)
		} else if target.ObjClass.Has(object.ClassPlayer) {
			rt.Audio(rt.OnSound(spell.ID(record.Spell)), target, 2, target.NetCode)
		}
		marker.Field34 = rt.Frame()
		shift := uint(index * 8)
		charge := byte(data.Field39 >> shift)
		charge--
		data.Field39 = data.Field39&^(uint32(0xff)<<shift) | uint32(charge)<<shift
		if charge == 0 {
			rt.SendPointFX(netmsg.MSG_FX_BLUE_SPARKS, marker.PosVec)
			rt.DelayedDelete(marker)
			data.Field29[index] = nil
		}
	}
	caster := record.Caster16
	if caster != nil && !caster.ObjFlags.Has(object.FlagDead) {
		rt.Attribution(caster, target)
	}
	return 1
}
