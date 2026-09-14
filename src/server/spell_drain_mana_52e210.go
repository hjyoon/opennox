package server

import (
	"math"

	"github.com/opennox/libs/types"
)

// SpellDrainManaRuntime52E210 owns the effects around the native duration
// record. In particular, the previous ray target cannot occupy PE32 Field36.
type SpellDrainManaRuntime52E210 struct {
	Frame           func() uint32
	TickRate        func() uint32
	Balance         func(string) float32
	BalanceLevel    func(string, uint32) float32
	ObjectsInCircle func(types.Pointf, float32, func(*Object) bool)
	IsEnemy         func(*Object, *Object) bool
	SameTeam        func(*Object, *Object) bool
	TraceRay        func(types.Pointf, types.Pointf, MapTraceFlags) bool
	PositionDelta   func(*Object, *types.Pointf) int32
	QuestMode       func() bool
	QuestManaScale  func(*Object) float32
	AddMana         func(*Object, int16)
	SubMana         func(*Object, int32)
	StartRay        func(*DurSpell)
	StopRay         func(*DurSpell, *Object)
	Audio           func(uint16, *Object)
	LoadRayTarget   func(*DurSpell) *Object
	StoreRayTarget  func(*DurSpell, *Object)
}

func spellDrainManaTargetHasMana52E7C0(target *Object) bool {
	class := uint32(target.ObjClass)
	if class&0x400000 != 0 && uint8(target.ObjSubClass)&0x18 != 0 {
		return target.UpdateDataObelisk().Mana > 0
	}
	if class&2 != 0 {
		return uint32(target.UpdateDataMonster().StatusFlags)&0x20 != 0
	}
	return class&4 != 0 && UnitGetOldMana4EEC80(target) != 0
}

func spellDrainManaClosest52E610(center types.Pointf, caster *Object, rt SpellDrainManaRuntime52E210) *Object {
	radius := rt.Balance("ManaDrainRange")
	bestScore := float32(100000000)
	var best *Object
	rt.ObjectsInCircle(center, radius, func(target *Object) bool {
		if target == caster || uint32(target.ObjFlags)&0x8020 != 0 || !spellDrainManaTargetHasMana52E7C0(target) {
			return true
		}
		class := uint32(target.ObjClass)
		weight := float32(1)
		switch {
		case class&2 != 0:
			if caster != nil && !rt.IsEnemy(caster, target) {
				return true
			}
		case class&0x400000 != 0 && uint8(target.ObjSubClass)&0x18 != 0:
			// Mana obelisks can be selected regardless of hostility.
		case class&4 != 0:
			if caster != nil && !rt.IsEnemy(caster, target) ||
				rt.SameTeam(target, caster) || target.UpdateDataPlayer().Player.Field3680&1 != 0 {
				return true
			}
			weight = 0.5
		default:
			return true
		}
		dx := center.X - target.PosVec.X
		dy := center.Y - target.PosVec.Y
		score := weight * (dx*dx + dy*dy)
		if score < bestScore && rt.TraceRay(center, target.PosVec, MapTraceFlags(5)) {
			bestScore = score
			best = target
		}
		return true
	})
	return best
}

func spellDrainManaTransfer52E450(caster, target *Object, amount int32, rt SpellDrainManaRuntime52E210) bool {
	if caster == nil {
		return false
	}
	if uint8(caster.ObjClass)&4 != 0 && UnitGetOldMana4EEC80(caster) >= PlayerGetMaxMana4EECB0(caster) {
		return false
	}
	class := uint32(target.ObjClass)
	transferred := amount
	switch {
	case class&0x400000 != 0 && uint8(target.ObjSubClass)&0x18 != 0:
		if target.HasTeam() && !target.TeamVal.SameAs(&caster.TeamVal) {
			return false
		}
		data := target.UpdateDataObelisk()
		if transferred > data.Mana {
			transferred = data.Mana
		}
		if transferred <= 0 {
			return false
		}
		if !rt.QuestMode() {
			data.Mana -= transferred
			target.NeedSync()
		}
	case class&2 != 0:
		if uint32(target.UpdateDataMonster().StatusFlags)&0x20 != 0 {
			transferred = 1
		} else if transferred <= 0 {
			return false
		}
	case class&4 != 0:
		current := int32(UnitGetOldMana4EEC80(target))
		if transferred > current {
			transferred = current
		}
		rt.SubMana(target, transferred)
		if transferred <= 0 {
			return false
		}
	default:
		return false
	}
	if rt.QuestMode() && uint8(caster.ObjClass)&4 != 0 {
		transferred = spellDurationRoundNearestEven(float32(transferred) * rt.QuestManaScale(caster))
	}
	rt.AddMana(caster, int16(transferred))
	return true
}

// SpellDrainManaUpdate52E210 replaces the C callback's float-argument ABI and
// PE32 duration/object loads. Glyphs drain once; ordinary casts track a ray.
func SpellDrainManaUpdate52E210(record *DurSpell, rt SpellDrainManaRuntime52E210) int32 {
	caster := record.Caster16
	if caster != nil {
		if caster.UnitBuffTest4FF350(8) != 0 {
			return 1
		}
	} else if record.Flag20 == 0 {
		return 1
	}
	if record.Flag20 != 0 {
		center := record.Pos
		if caster != nil {
			center = caster.PosVec
		}
		if target := spellDrainManaClosest52E610(center, caster, rt); target != nil {
			rt.SubMana(target, 50)
		}
		return 1
	}
	if uint8(caster.ObjClass)&4 != 0 && UnitGetOldMana4EEC80(caster) >= PlayerGetMaxMana4EECB0(caster) {
		return 1
	}
	if uint8(caster.ObjClass)&2 != 0 && rt.PositionDelta(caster, &record.Pos) != 0 {
		return 1
	}
	frame := rt.Frame()
	if caster.HealthData != nil && frame-caster.Frame134 <= 1 ||
		math.Abs(float64(record.Pos.X-caster.PosVec.X)) >= 5 ||
		math.Abs(float64(record.Pos.Y-caster.PosVec.Y)) >= 5 {
		return 1
	}
	target := spellDrainManaClosest52E610(caster.PosVec, caster, rt)
	record.Target48 = target
	if target == nil {
		if previous := rt.LoadRayTarget(record); previous != nil {
			rt.StopRay(record, previous)
			rt.StoreRayTarget(record, nil)
		}
		return 1
	}
	value := rt.BalanceLevel("ManaDrainCoeff", record.Level-1) + math.Float32frombits(uint32(record.Field72))
	amount := spellDurationRoundNearestEven(value)
	record.Field72 = int32(math.Float32bits(value - float32(amount)))
	if previous := rt.LoadRayTarget(record); previous != target {
		if previous != nil {
			rt.StopRay(record, previous)
		}
		rt.StartRay(record)
	}
	if spellDrainManaTransfer52E450(caster, target, amount, rt) {
		if period := rt.TickRate() >> 1; period != 0 && frame%period == 0 {
			rt.Audio(230, caster)
			rt.Audio(229, target)
		}
	}
	rt.StoreRayTarget(record, target)
	return 0
}
