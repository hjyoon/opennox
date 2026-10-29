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
	LoadTypeID     func() uint32
	StoreTypeID    func(uint32)
	NameID         func(string) uint32
	AllocateCharge func(uint32) *Object
	BeginDamage    func()
	Frame          func() uint32
	TickRate       func() uint32
	Balance        func(string) float32
	BalanceLevel   func(string, uint32) float32
	NewObject      func(string) *Object
	CreateAt       func(*Object, *Object, types.Pointf)
	LoadCharge     func(*DurSpell) *Object
	StoreCharge    func(*DurSpell, *Object)
	DelayedDelete  func(*Object)
	ApplyBuff      func(*Object, EnchantID, int16, int8)
	BuffOff        func(*Object, EnchantID)
	DamageAround   func(types.Pointf, float32, float32, int, *Object)
	Earthquake     func(types.Pointf, int)
	PointFX        func(netmsg.Op, types.Pointf)
	Audio          func(*Object)
	ManaSub        func(*Object, int32)
}

// SpellManaBombCreate530F90 follows GAME.EXE 00530F90 with native-width
// object references. The original stores the caster's mass as raw float bits.
// Its FISTP conversions use the startup toward-zero mode without changing
// host floating-point state. World callbacks may replace live record fields.
//
//go:noinline
func SpellManaBombCreate530F90(record *DurSpell, runtime SpellManaBombRuntime530F90) int32 {
	if runtime.LoadTypeID != nil && runtime.LoadTypeID() == 0 {
		runtime.StoreTypeID(runtime.NameID("ManaBombCharge"))
	}
	record.Field72 = playerReadValuesRound4EEDC0(runtime.BalanceLevel("ManaBombInitPower", record.Level-1))
	record.Field76 = 0
	if runtime.StoreCharge != nil {
		runtime.StoreCharge(record, nil)
	}
	record.Field84 = 0
	caster := record.Caster16
	if caster != nil && uint8(caster.ObjClass)&4 == 0 || record.Flag20 != 0 {
		duration := playerReadValuesRound4EEDC0(runtime.Balance("ManaBombGlyphDuration"))
		record.Frame68 = runtime.Frame() + uint32(duration)
	} else {
		for _, buff := range [...]EnchantID{5, 14, 29} {
			duration := int16(10 * runtime.TickRate())
			runtime.ApplyBuff(record.Caster16, buff, duration, 5)
		}
		record.Field80 = math.Float32bits(record.Caster16.Mass)
		record.Caster16.Mass = math.Float32frombits(1203982323)
		record.Caster16.VelVec = types.Pointf{}
		record.Caster16.ForceVec = types.Pointf{}
		record.Caster16.Pos24 = types.Pointf{}
	}
	var charge *Object
	if runtime.AllocateCharge != nil {
		charge = runtime.AllocateCharge(runtime.LoadTypeID())
	} else {
		charge = runtime.NewObject("ManaBombCharge")
	}
	if charge != nil {
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
	power := record.Field72
	if explode {
		position := record.Pos
		if record.Flag20 == 0 {
			if record.Caster16 == nil {
				return 1
			}
			position = record.Caster16.PosVec
		}
		if runtime.BeginDamage != nil {
			runtime.BeginDamage()
		}
		owner := record.Caster16
		inner := runtime.Balance("ManaBombInRadius")
		outer := runtime.Balance("ManaBombOutRadius")
		runtime.DamageAround(position, outer, inner, int(power), owner)
		runtime.Earthquake(position, int(playerReadValuesRound4EEDC0(runtime.Balance("ManaBombShakeMag"))))
		runtime.PointFX(netmsg.Op(129), position)
		runtime.PointFX(netmsg.Op(154), position)
		runtime.Audio(record.Caster16)
		record.Field84 = 1
		return 1
	}
	record.Field72 = power + playerReadValuesRound4EEDC0(runtime.BalanceLevel("ManaBombDeltaPower", record.Level-1))
	if caster = record.Caster16; record.Flag20 == 0 && caster != nil && uint8(caster.ObjClass)&4 != 0 {
		runtime.ManaSub(record.Caster16, int32(record.Level))
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
			runtime.BuffOff(record.Caster16, buff)
		}
		record.Caster16.Mass = math.Float32frombits(record.Field80)
	}
	if record.Field84 == 0 {
		runtime.PointFX(netmsg.Op(163), record.Pos)
	}
}
