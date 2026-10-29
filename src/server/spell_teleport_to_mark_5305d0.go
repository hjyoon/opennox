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
	TeleportDelay func(uint32) float64
	NewObject     func(string) *Object
	CreateAt      func(*Object, *Object, types.Pointf)
	Wake          func(*Object, *types.Pointf, *types.Pointf)
	SendPointFX   func(netmsg.Op, types.Pointf)
	OnSound       func(spell.ID) sound.ID
	Audio         func(sound.ID, *Object, int, uint32)
	Teleport      func(*Object, *types.Pointf)
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
		record.Frame68 = rt.Frame() + uint32(playerReadValuesRound4EEDC0(shapeBoxSpill428270(delay)))
	} else {
		record.Frame68 = rt.Frame() + 1
	}
	return 0
}

// SpellTeleportToMarkUpdate530650 mirrors GAME.EXE 00530650 without reading
// PE32 object and duration offsets from native-width records. The original
// source/caster gates test destroyed0x20; dead0x8000 is a separate flag.
func SpellTeleportToMarkUpdate530650(record *DurSpell, rt SpellTeleportToMarkRuntime5305D0) int32 {
	source, target := record.Obj12, record.Target48
	if source == nil || source.ObjFlags.Has(object.FlagDestroyed) || target == nil ||
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
	source = record.Obj12
	if source.ObjClass.Has(object.ClassPlayer) {
		index, ok := teleportMarkIndex530650(spell.ID(record.Spell))
		if !ok || source.UpdateData == nil {
			return 1
		}
		data := source.UpdateDataPlayer()
		marker := data.Field29[index]
		if marker == nil {
			return 1
		}
		target = record.Target48
		origin := target.PosVec
		if rt.Wake != nil {
			rt.Wake(target, &target.PosVec, &marker.PosVec)
		} else {
			SpellTeleportCreateWake530560(target, &target.PosVec, &marker.PosVec, SpellBlinkRuntime530310{
				NewObject: rt.NewObject, CreateAt: rt.CreateAt, Frame: rt.Frame, TickRate: rt.TickRate,
			})
		}
		play := func(unit *Object, kind int, code uint32) {
			onSound := rt.OnSound(spell.ID(record.Spell))
			rt.Audio(onSound, unit, kind, code)
		}
		play(record.Target48, 0, 0)
		rt.Teleport(record.Target48, &data.Field29[index].PosVec)
		if !record.Target48.HasEnchant(ENCHANT_INVISIBLE) {
			rt.SendPointFX(netmsg.MSG_FX_TELEPORT, origin)
			rt.SendPointFX(netmsg.MSG_FX_TELEPORT, record.Target48.PosVec)
			play(record.Target48, 0, 0)
		} else if target = record.Target48; target.ObjClass.Has(object.ClassPlayer) {
			play(target, 2, target.NetCode)
		}
		marker = data.Field29[index]
		marker.Field34 = rt.Frame()
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
	caster := record.Caster16
	if caster != nil && !caster.ObjFlags.Has(object.FlagDestroyed) {
		rt.Attribution(caster, record.Target48)
	}
	return 1
}
