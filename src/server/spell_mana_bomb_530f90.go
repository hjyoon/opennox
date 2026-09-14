package server

import (
	"math"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/types"
)

// SpellManaBombRuntime530F90 owns effects outside the duration record. The
// charge visual is kept in a native-pointer sidecar instead of the PE32 dword
// at Field76, which also holds non-pointer data for other spells.
type SpellManaBombRuntime530F90 struct {
	Frame         func() uint32
	TickRate      func() uint32
	Balance       func(string) float32
	BalanceLevel  func(string, uint32) float32
	NewObject     func(string) *Object
	CreateAt      func(*Object, *Object, types.Pointf)
	LoadCharge    func(*DurSpell) *Object
	StoreCharge   func(*DurSpell, *Object)
	DelayedDelete func(*Object)
	ApplyBuff     func(*Object, EnchantID, int16, int8)
	BuffOff       func(*Object, EnchantID)
	DamageAround  func(types.Pointf, float32, float32, int, *Object)
	Earthquake    func(types.Pointf, int)
	PointFX       func(netmsg.Op, types.Pointf)
	Audio         func(*Object)
	ManaSub       func(*Object, int32)
}

// SpellManaBombCreate530F90 follows GAME.EXE 00530F90 with native-width
// object references. The original stores the caster's mass as raw float bits.
//
//go:noinline
func SpellManaBombCreate530F90(record *DurSpell, runtime SpellManaBombRuntime530F90) int32 {
	record.Field72 = spellDurationRoundNearestEven(runtime.BalanceLevel("ManaBombInitPower", record.Level-1))
	record.Field76 = 0
	record.Field84 = 0
	caster := record.Caster16
	if caster == nil || uint8(caster.ObjClass)&4 == 0 || record.Flag20 != 0 {
		record.Frame68 = runtime.Frame() + uint32(spellDurationRoundNearestEven(runtime.Balance("ManaBombGlyphDuration")))
	} else {
		duration := int16(10 * uint16(runtime.TickRate()))
		for _, buff := range [...]EnchantID{5, 14, 29} {
			runtime.ApplyBuff(caster, buff, duration, 5)
		}
		record.Field80 = math.Float32bits(caster.Mass)
		caster.Mass = math.Float32frombits(1203982323)
		caster.VelVec = types.Pointf{}
		caster.ForceVec = types.Pointf{}
		caster.Pos24 = types.Pointf{}
	}
	if charge := runtime.NewObject("ManaBombCharge"); charge != nil {
		runtime.CreateAt(charge, nil, record.Pos)
		runtime.StoreCharge(record, charge)
	}
	return 0
}

// SpellManaBombUpdate5310C0 follows GAME.EXE 005310C0. The duration manager
// cancels the record when this callback returns one.
//
//go:noinline
func SpellManaBombUpdate5310C0(record *DurSpell, runtime SpellManaBombRuntime530F90) int32 {
	caster := record.Caster16
	if caster == nil && record.Flag20 == 0 {
		return 1
	}
	frame := runtime.Frame()
	explode := record.Frame68-1 == frame
	if caster != nil && uint8(caster.ObjClass)&4 != 0 && UnitGetOldMana4EEC80(caster) == 0 {
		explode = true
	}
	if charge := runtime.LoadCharge(record); charge != nil {
		keepCharge := false
		if record.Flag20 == 0 && caster != nil && uint8(caster.ObjClass)&4 != 0 {
			keepCharge = UnitGetOldMana4EEC80(caster) >= 15
		} else {
			keepCharge = record.Frame68-frame >= 10
		}
		if !keepCharge {
			runtime.DelayedDelete(charge)
			runtime.StoreCharge(record, nil)
		}
	}
	if explode {
		position := record.Pos
		if record.Flag20 == 0 {
			position = caster.PosVec
		}
		runtime.DamageAround(position, runtime.Balance("ManaBombOutRadius"), runtime.Balance("ManaBombInRadius"), int(record.Field72), caster)
		runtime.Earthquake(position, int(spellDurationRoundNearestEven(runtime.Balance("ManaBombShakeMag"))))
		runtime.PointFX(netmsg.Op(129), position)
		runtime.PointFX(netmsg.Op(154), position)
		runtime.Audio(caster)
		record.Field84 = 1
		return 1
	}
	record.Field72 += spellDurationRoundNearestEven(runtime.BalanceLevel("ManaBombDeltaPower", record.Level-1))
	if record.Flag20 == 0 && caster != nil && uint8(caster.ObjClass)&4 != 0 {
		runtime.ManaSub(caster, int32(record.Level))
	}
	return 0
}

// SpellManaBombDestroy531290 follows GAME.EXE 00531290. A canceled charge
// sends its own FX; an exploded charge already sent the explosion effects.
//
//go:noinline
func SpellManaBombDestroy531290(record *DurSpell, runtime SpellManaBombRuntime530F90) {
	if charge := runtime.LoadCharge(record); charge != nil {
		runtime.DelayedDelete(charge)
	}
	runtime.StoreCharge(record, nil)
	if caster := record.Caster16; record.Flag20 == 0 && caster != nil && uint8(caster.ObjClass)&4 != 0 {
		for _, buff := range [...]EnchantID{5, 14, 29} {
			runtime.BuffOff(caster, buff)
		}
		caster.Mass = math.Float32frombits(record.Field80)
	}
	if record.Field84 == 0 {
		runtime.PointFX(netmsg.Op(163), record.Pos)
	}
}
