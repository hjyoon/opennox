package server

import "github.com/opennox/libs/types"

// SpellForceOfNatureRuntime52EF30 contains the game-owned effects used by the
// three Force of Nature duration callbacks. The charge is deliberately kept
// outside Field76: that field was a PE32 object pointer in the original game.
type SpellForceOfNatureRuntime52EF30 struct {
	NewObject      func(string) *Object
	CreateAt       func(*Object, *Object, types.Pointf)
	DelayedDelete  func(*Object)
	LoadCharge     func(*DurSpell) *Object
	StoreCharge    func(*DurSpell, *Object)
	CurrentFrame   func() uint32
	TraceRay       func(types.Pointf, types.Pointf, MapTraceFlags) bool
	SetPlayerState func(*Object, PlayerState)
	EventObj       func(*Object)
}

func forceOfNatureEquippedWand52EF30(caster *Object) *Object {
	if caster == nil || uint8(caster.ObjClass)&4 == 0 {
		return nil
	}
	return caster.UpdateDataPlayer().EquippedWeapon
}

// SpellForceOfNatureCreate52EF30 follows GAME.EXE 0052EF30 without sending a
// duration record or object through the 32-bit C callback ABI.
//
//go:noinline
func SpellForceOfNatureCreate52EF30(record *DurSpell, runtime SpellForceOfNatureRuntime52EF30) int32 {
	var source *Object
	if record.Flag20 != 0 {
		source = record.Obj24
	} else {
		source = record.Caster16
		if wand := forceOfNatureEquippedWand52EF30(source); wand != nil &&
			uint32(wand.ObjSubClass)&0x200000 != 0 && wand.UseData.Ptr != nil && wand.UseData.AsWand().Flags&4 != 0 {
			record.Flags88 |= 2
		}
	}
	record.Field72 = int32(uint32(record.Field72)&0xffff0000 | uint32(source.Direction1))
	if charge := runtime.NewObject("ForceOfNatureCharge"); charge != nil {
		runtime.CreateAt(charge, nil, source.PosVec)
		runtime.StoreCharge(record, charge)
	}
	return 0
}

func forceOfNatureLaunchDistance52EFD0(caster *Object) float32 {
	switch caster.Shape.Kind {
	case ShapeKindCenter:
		return 4
	case ShapeKindCircle:
		return caster.Shape.Circle.R + 4
	case ShapeKindBox:
		width, height := caster.Shape.Box.W, caster.Shape.Box.H
		if width < height {
			width = height
		}
		return width + 4
	default:
		return 24
	}
}

// SpellForceOfNatureUpdate52EFD0 follows GAME.EXE 0052EFD0. The charge is
// removed seven frames before expiry, then a DeathBall is launched one frame
// before expiry. The duration process cancels the record on a nonzero return.
//
//go:noinline
func SpellForceOfNatureUpdate52EFD0(record *DurSpell, runtime SpellForceOfNatureRuntime52EF30) int32 {
	caster := record.Caster16
	if caster != nil && caster.UnitBuffTest4FF350(8) != 0 {
		return 1
	}
	frame := runtime.CurrentFrame()
	if record.Frame68-7 == frame {
		if charge := runtime.LoadCharge(record); charge != nil {
			runtime.DelayedDelete(charge)
			runtime.StoreCharge(record, nil)
		}
	}
	if record.Frame68-1 == frame {
		ball := runtime.NewObject("DeathBall")
		if ball != nil {
			var source types.Pointf
			var direction Dir16
			var distance float32
			if record.Flag20 != 0 {
				source = record.Pos
				direction = Dir16(uint16(record.Field72))
			} else {
				source = caster.PosVec
				direction = caster.Direction1
				distance = forceOfNatureLaunchDistance52EFD0(caster)
			}
			vector := direction.Vec()
			position := types.Pointf{
				X: source.X + vector.X*distance,
				Y: source.Y + vector.Y*distance,
			}
			if !runtime.TraceRay(source, position, MapTraceFlag1|MapTraceFlag3) {
				position = source
			}
			runtime.CreateAt(ball, caster, position)
			if record.Flag20 == 1 {
				ball.VelVec = types.Pointf{}
			} else {
				ball.VelVec = types.Pointf{X: vector.X * ball.SpeedCur, Y: vector.Y * ball.SpeedCur}
			}
			ball.Direction1 = direction
			ball.Direction2 = direction
			runtime.EventObj(caster)
		}
		return 1
	}
	if record.Flag20 == 0 && caster != nil && uint8(caster.ObjClass)&4 != 0 {
		runtime.SetPlayerState(caster, PlayerState(10))
	}
	return 0
}

// SpellForceOfNatureDestroy52F1D0 follows GAME.EXE 0052F1D0. Its original
// pointer-valued return was discarded by the duration manager.
//
//go:noinline
func SpellForceOfNatureDestroy52F1D0(record *DurSpell, runtime SpellForceOfNatureRuntime52EF30) {
	if charge := runtime.LoadCharge(record); charge != nil {
		runtime.DelayedDelete(charge)
	}
	runtime.StoreCharge(record, nil)
	if wand := forceOfNatureEquippedWand52EF30(record.Caster16); wand != nil &&
		uint32(wand.ObjSubClass)&0x200000 != 0 && wand.UseData.Ptr != nil {
		wand.UseData.AsWand().Flags &^= 4
	}
}
