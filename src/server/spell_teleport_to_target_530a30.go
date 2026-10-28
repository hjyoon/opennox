package server

import (
	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
)

// SpellTeleportToTargetRuntime530A30 keeps game services separate from the
// native-width duration and object references used by the two callbacks.
type SpellTeleportToTargetRuntime530A30 struct {
	CoopMode           func() bool
	Frame              func() uint32
	TickRate           func() uint32
	TeleportDelay      func(uint32) float64
	TileBlocksTeleport func(*types.Pointf) bool
	TraceRay9          func(types.Pointf, types.Pointf) bool
	LookupUnseenTarget func() string
	SendUnseenTarget   func(*Object, string)
	InformNoLOS        func(*PlayerUpdateData)
	Wake               func(*Object, *types.Pointf, *types.Pointf)
	NewObject          func(string) *Object
	CreateAt           func(*Object, *Object, types.Pointf)
	SendPointFX        func(netmsg.Op, types.Pointf)
	CastSound          func(spell.ID) sound.ID
	Audio              func(sound.ID, *Object, int, uint32)
	Teleport           func(*Object, types.Pointf)
	Attribution        func(*Object, *Object)
}

// SpellTeleportToTargetCreate530A30 follows ExecDur.c 00530A30. The source
// record's PE32 offset 48 is a pointer, not the native record's Pos.X field.
func SpellTeleportToTargetCreate530A30(record *DurSpell, rt SpellTeleportToTargetRuntime530A30) int32 {
	if record.Target48 == nil {
		record.Target48 = record.Caster16
	}
	caster, target := record.Caster16, record.Target48
	if caster == nil || target == nil {
		return 1
	}
	var playerData *PlayerUpdateData
	if caster.ObjClass.Has(object.ClassPlayer) && caster.UpdateData != nil {
		playerData = caster.UpdateDataPlayer()
	}
	if rt.TileBlocksTeleport(&record.Pos2) {
		message := rt.LookupUnseenTarget()
		rt.SendUnseenTarget(record.Target48, message)
		rt.Audio(sound.ID(231), record.Target48, 0, 0)
		return 1
	}
	// ExecDur.c caches the original player's update data before the tile
	// service, but reloads the target, destination, and caster for the ray.
	if record.Caster16 != nil && !rt.TraceRay9(record.Target48.PosVec, record.Pos2) {
		if playerData != nil {
			rt.InformNoLOS(playerData)
		}
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

// SpellTeleportToTargetUpdate530B70 follows GAME.EXE 00530B70, including
// its frame-minus-one gate and repeated destination effects for visible units.
//
//go:noinline
func SpellTeleportToTargetUpdate530B70(record *DurSpell, rt SpellTeleportToTargetRuntime530A30) int32 {
	target, caster := record.Target48, record.Caster16
	if target == nil || caster == nil ||
		target.ObjFlags.HasAny(object.FlagDead|object.FlagDestroyed) ||
		caster.ObjFlags.HasAny(object.FlagDead|object.FlagDestroyed) {
		return 1
	}
	if record.Frame68-1 != rt.Frame() {
		return 0
	}
	if target.HasEnchant(ENCHANT_ANCHORED) {
		rt.Audio(sound.ID(231), target, 0, 0)
		return 1
	}
	rt.SendPointFX(netmsg.MSG_FX_TELEPORT, record.Target48.PosVec)
	// Each original sound lookup receives the current spell. Its audio
	// target and network arguments are already fixed on the caller's stack.
	play := func(unit *Object, kind int, code uint32) {
		castSound := rt.CastSound(spell.ID(record.Spell))
		rt.Audio(castSound, unit, kind, code)
	}
	play(record.Target48, 0, 0)
	target = record.Target48
	if rt.Wake != nil {
		rt.Wake(target, &target.PosVec, &record.Pos2)
	} else {
		SpellTeleportCreateWake530560(target, &target.PosVec, &record.Pos2, SpellBlinkRuntime530310{
			NewObject: rt.NewObject, CreateAt: rt.CreateAt, Frame: rt.Frame, TickRate: rt.TickRate,
		})
	}
	rt.Teleport(record.Target48, record.Pos2)
	target = record.Target48
	if !target.HasEnchant(ENCHANT_INVISIBLE) {
		rt.SendPointFX(netmsg.MSG_FX_TELEPORT, target.PosVec)
		play(record.Target48, 0, 0)
	} else if target.ObjClass.Has(object.ClassPlayer) {
		play(target, 2, target.NetCode)
	}
	rt.Attribution(record.Caster16, record.Target48)
	return 1
}
