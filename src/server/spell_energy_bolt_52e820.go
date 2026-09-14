package server

import (
	"math"

	"github.com/opennox/libs/types"
)

// SpellEnergyBoltRuntime52E820 keeps the external services used by the
// Lightning duration callbacks separate from their native-width record logic.
type SpellEnergyBoltRuntime52E820 struct {
	Frame           func() uint32
	TickRate        func() uint32
	Balance         func(string) float32
	BalanceLevel    func(string, uint32) float32
	ObjectsInCircle func(types.Pointf, float32, func(*Object) bool)
	CanInteract     func(*Object, *Object) bool
	IsEnemy         func(*Object, *Object) bool
	InFront         func(*Object, *Object) bool
	PositionDelta   func(*Object, *types.Pointf) int32
	CancelSpell     func(int32, *Object)
	StartRay        func(*DurSpell)
	StopRay         func(*DurSpell, *Object)
	PointFX         func(uint8, types.Pointf)
	Damage          func(*Object, *Object, int32)
	Audio           func(uint16, *Object)
	CastSound       func() uint16
	SetPlayerState  func(*Object, PlayerState)
	ManaSub         func(*Object, int32)
	LoadRayTarget   func(*DurSpell) *Object
	StoreRayTarget  func(*DurSpell, *Object)
}

// SpellEnergyBoltCreate52E820 cancels Chain Lightning for the same caster and
// emits Lightning's initial point effect. The old C callback read PE32 offsets.
func SpellEnergyBoltCreate52E820(record *DurSpell, rt SpellEnergyBoltRuntime52E820) int32 {
	if caster := record.Caster16; caster != nil {
		rt.CancelSpell(43, caster)
	}
	rt.PointFX(130, record.Pos)
	return 0
}

func energyBoltDistanceSquared52EC60(a, b types.Pointf) float32 {
	dx := a.X - b.X
	dy := a.Y - b.Y
	return dx*dx + dy*dy
}

// The original 0052EC60 callback selects the closest eligible object. The
// glyph variant skips the facing and interaction checks, but still requires an
// enemy when a caster is present.
func energyBoltClosest52EC60(record *DurSpell, center types.Pointf, radius float32, rt SpellEnergyBoltRuntime52E820) *Object {
	caster := record.Caster16
	bestDistance := radius * radius
	var best *Object
	rt.ObjectsInCircle(center, radius, func(target *Object) bool {
		if target == caster || uint32(target.ObjClass)&0x20006 == 0 ||
			uint32(target.ObjFlags)&0x8020 != 0 ||
			uint8(target.ObjClass)&2 != 0 && uint32(target.ObjSubClass)&0x8000 != 0 {
			return true
		}
		if caster != nil && !rt.IsEnemy(caster, target) {
			return true
		}
		if record.Flag20 == 0 && (!rt.InFront(caster, target) || !rt.CanInteract(caster, target)) {
			return true
		}
		if distance := energyBoltDistanceSquared52EC60(center, target.PosVec); distance < bestDistance {
			bestDistance = distance
			best = target
		}
		return true
	})
	return best
}

// SpellEnergyBoltUpdate52E850 follows the two paths of GAME.EXE 0052E850:
// a one-shot glyph strike and a sustained Lightning ray. Both the C callback's
// float argument ABI and its PE32 object/record offsets are bypassed.
func SpellEnergyBoltUpdate52E850(record *DurSpell, rt SpellEnergyBoltRuntime52E820) int32 {
	caster := record.Caster16
	if caster != nil {
		if caster.UnitBuffTest4FF350(8) != 0 {
			return 1
		}
	} else if record.Flag20 == 0 {
		return 1
	}
	radius := rt.Balance("LightningRange")
	if record.Flag20 != 0 {
		if target := energyBoltClosest52EC60(record, record.Pos, radius, rt); target != nil {
			damage := spellDurationRoundNearestEven(rt.Balance("EnergyBoltGlyphDamage"))
			rt.Damage(target, record.Obj12, damage)
			rt.Audio(rt.CastSound(), target)
			rt.PointFX(130, target.PosVec)
		}
		return 1
	}
	frame := rt.Frame()
	if uint8(caster.ObjClass)&2 != 0 && rt.PositionDelta(caster, &record.Pos) != 0 {
		return 1
	}
	if frame-record.Frame60 > 2 && caster.HealthData != nil && frame-caster.Frame134 <= 1 {
		return 1
	}
	target := record.Target48
	if target != nil {
		if uint32(target.ObjFlags)&0x8020 != 0 || !rt.InFront(caster, target) ||
			energyBoltDistanceSquared52EC60(caster.PosVec, target.PosVec) > radius*radius ||
			!rt.CanInteract(caster, target) {
			target = nil
			record.Target48 = nil
		}
	}
	if target == nil && uint8(caster.ObjClass)&4 != 0 {
		if update := caster.UpdateDataPlayer(); update != nil {
			if cursor := update.CursorObj; cursor != nil && rt.IsEnemy(caster, cursor) &&
				energyBoltDistanceSquared52EC60(caster.PosVec, cursor.PosVec) <= radius*radius {
				target = cursor
			}
		}
	}
	if target == nil {
		target = energyBoltClosest52EC60(record, caster.PosVec, radius, rt)
	}
	record.Target48 = target
	if target == nil {
		if previous := rt.LoadRayTarget(record); previous != nil {
			rt.StopRay(record, previous)
			rt.StoreRayTarget(record, nil)
		}
		return 0
	}
	damageFloat := rt.BalanceLevel("EnergyBoltDamage", record.Level-1) + math.Float32frombits(uint32(record.Field72))
	damage := spellDurationRoundNearestEven(damageFloat)
	record.Field72 = int32(math.Float32bits(damageFloat - float32(damage)))
	if previous := rt.LoadRayTarget(record); target != previous {
		if previous != nil {
			rt.StopRay(record, previous)
		}
		rt.StartRay(record)
	}
	rt.Damage(target, caster, damage)
	if uint32(target.ObjFlags)&0x8020 != 0 {
		rt.PointFX(130, target.PosVec)
	}
	rt.StoreRayTarget(record, target)
	if uint8(caster.ObjClass)&4 != 0 {
		rt.SetPlayerState(caster, 10)
	}
	if period := rt.TickRate() / 3; period != 0 && frame%period == 0 {
		rt.Audio(32, caster)
		rt.Audio(32, target)
	}
	record.Frame68 = frame + uint32(spellDurationRoundNearestEven(rt.Balance("LightningSearchTime")))
	if uint8(caster.ObjClass)&4 != 0 {
		rt.SetPlayerState(caster, 10)
		rt.ManaSub(caster, 1)
		if UnitGetOldMana4EEC80(caster) == 0 {
			return 1
		}
	}
	if uint32(target.ObjFlags)&0x8000 != 0 {
		record.Frame68 = frame + 1
	}
	return 0
}
