package server

import (
	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
)

// SpellBlinkRuntime530310 owns the map, clock, and visual effects used by the
// two Blink duration callbacks. All object and duration references stay native
// pointers instead of being decoded through the original PE32 dword layout.
type SpellBlinkRuntime530310 struct {
	QuestMode       func() bool
	CoopMode        func() bool
	Frame           func() uint32
	TickRate        func() uint32
	TeleportDelay   func(uint32) float64
	Waypoint        func(*types.Pointf, *Object) int
	PlayerStart     func(*Object) types.Pointf
	RandomReachable func(float32, types.Pointf) types.Pointf
	NewObject       func(string) *Object
	WakeType        *uint32
	LookupType      func(string) uint32
	NewTypedObject  func(uint32) *Object
	CreateAt        func(*Object, *Object, types.Pointf)
	SendPointFX     func(netmsg.Op, types.Pointf)
	CastSound       func(spell.ID) sound.ID
	Audio           func(sound.ID, *Object, int, uint32)
	Teleport        func(*Object, types.Pointf)
	Attribution     func(*Object, *Object)
}

// SpellBlinkCreate530310 follows GAME.EXE 00530310. Quest glyphs finish at
// creation; other Blink durations schedule their update in Frame68.
func SpellBlinkCreate530310(record *DurSpell, runtime SpellBlinkRuntime530310) int32 {
	if runtime.QuestMode() && record.Flag20 == 1 {
		return 1
	}
	if runtime.CoopMode() {
		delay := runtime.TeleportDelay(record.Level - 1)
		record.Frame68 = runtime.Frame() + uint32(playerReadValuesRound4EEDC0(shapeBoxSpill428270(delay)))
	} else {
		record.Frame68 = runtime.Frame() + 1
	}
	return 0
}

// SpellBlinkUpdate530380 follows GAME.EXE 00530380, including the original
// frame-minus-one gate and repeated teleport FX for visible units. A wake's
// destination remains a native Pointf in its collision data.
//
//go:noinline
func SpellBlinkUpdate530380(record *DurSpell, runtime SpellBlinkRuntime530310) int32 {
	target := record.Target48
	if target == nil || uint32(target.ObjFlags)&0x8020 != 0 {
		return 1
	}
	if record.Frame68-1 != runtime.Frame() {
		return 0
	}
	if target.HasEnchant(ENCHANT_ANCHORED) {
		runtime.Audio(sound.ID(231), target, 0, 0)
		return 1
	}

	var destination types.Pointf
	if runtime.QuestMode() && uint8(record.Target48.ObjClass)&4 != 0 {
		target = record.Target48
		if gate := target.UpdateDataPlayer().SoulGate; gate != nil {
			destination = runtime.RandomReachable(60, gate.PosVec)
		} else {
			destination = runtime.PlayerStart(record.Target48)
		}
	} else if runtime.Waypoint(&destination, record.Target48) == 0 {
		destination = runtime.PlayerStart(record.Target48)
	}

	target = record.Target48
	SpellTeleportCreateWake530560(target, &target.PosVec, &destination, runtime)

	runtime.SendPointFX(netmsg.MSG_FX_TELEPORT, record.Target48.PosVec)
	// The target is already on the original call stack before sound lookup.
	play := func(unit *Object, kind int, code uint32) {
		castSound := runtime.CastSound(spell.ID(record.Spell))
		runtime.Audio(castSound, unit, kind, code)
	}
	play(record.Target48, 0, 0)
	if !record.Target48.HasEnchant(ENCHANT_INVISIBLE) {
		runtime.SendPointFX(netmsg.MSG_FX_TELEPORT, record.Target48.PosVec)
		runtime.SendPointFX(netmsg.MSG_FX_TELEPORT, destination)
	}
	runtime.Teleport(record.Target48, destination)
	invisible := record.Target48.HasEnchant(ENCHANT_INVISIBLE)
	target = record.Target48
	if !invisible {
		runtime.SendPointFX(netmsg.MSG_FX_TELEPORT, target.PosVec)
		play(record.Target48, 0, 0)
	} else if uint8(target.ObjClass)&4 != 0 {
		play(target, 2, target.NetCode)
	}
	runtime.Attribution(record.Caster16, record.Target48)
	return 1
}

// SpellTeleportCreateWake530560 keeps both input points live across type
// lookup/allocation. The original caller's target and origin address remain
// fixed while the duration's target may change during those services.
func SpellTeleportCreateWake530560(target *Object, origin, destination *types.Pointf, runtime SpellBlinkRuntime530310) {
	var wake *Object
	if runtime.WakeType != nil {
		kind := *runtime.WakeType
		if kind == 0 {
			kind = runtime.LookupType("TeleportWake")
			*runtime.WakeType = kind
		}
		wake = runtime.NewTypedObject(kind)
	} else {
		wake = runtime.NewObject("TeleportWake")
	}
	if wake == nil || wake.CollideData == nil {
		return
	}
	data := (*TeleportWakeCollideData)(wake.CollideData)
	data.Destination.X = destination.X
	data.Destination.Y = destination.Y
	runtime.CreateAt(wake, target, *origin)
	wake.Field34 = runtime.Frame() + runtime.TickRate()
}
