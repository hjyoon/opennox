package server

import (
	"math"

	"github.com/opennox/libs/types"
)

// SpellDrainManaRuntime52E210 supplies the world services around native-width
// duration/object links. The standalone API retains nearest rounding; the
// production adapter explicitly selects the original startup chop context.
type SpellDrainManaRuntime52E210 struct {
	Frame           func() uint32
	TickRate        func() uint32
	Balance         func(string) float32
	BalanceLevel    func(string, uint32) float32
	Balance64       func(string) float64
	BalanceLevel64  func(string, uint32) float64
	Selection       *SpellDrainManaSelection52E610
	RoundTowardZero bool
	ObjectsInCircle func(types.Pointf, float32, func(*Object) bool)
	IsEnemy         func(*Object, *Object) bool
	SameTeam        func(*Object, *Object) bool
	TraceRay        func(types.Pointf, types.Pointf, MapTraceFlags) bool
	PositionDelta   func(*Object, *types.Pointf) int32
	QuestMode       func() bool
	QuestManaScale  func(*Object) float32
	AddMana         func(*Object, int16)
	AddMana32       func(*Object, int32)
	SubMana         func(*Object, int32)
	StartRay        func(*DurSpell)
	StopRay         func(*DurSpell, *Object)
	Audio           func(uint16, *Object)
	LoadRayTarget   func(*DurSpell) *Object
	StoreRayTarget  func(*DurSpell, *Object)
}

type SpellDrainManaSelection52E610 struct {
	Center  types.Pointf
	Target  *Object
	Nearest float32
}

func drainManaSpill52E210(value float64, rt SpellDrainManaRuntime52E210) float32 {
	if rt.RoundTowardZero {
		return float32(monsterMoveToRunSpill544440(value))
	}
	return float32(value)
}

//go:noinline
func drainManaAdd52E210(a, b float64, rt SpellDrainManaRuntime52E210) float64 {
	if rt.RoundTowardZero {
		return monsterMoveToRunAddChop53_544434(a, b)
	}
	return a + b
}

//go:noinline
func drainManaSubtract52E210(a, b float64, rt SpellDrainManaRuntime52E210) float64 {
	// FSUB propagates a NaN operand without negating its sign. Arithmetic
	// addition of -b would change that observed sign before the FST spill.
	if rt.RoundTowardZero && !math.IsNaN(a) && !math.IsNaN(b) {
		return monsterMoveToRunAddChop53_544434(a, -b)
	}
	return a - b
}

//go:noinline
func drainManaMultiply52E450(a, b float64, rt SpellDrainManaRuntime52E210) float64 {
	product := float64(a * b)
	if rt.RoundTowardZero {
		error := math.FMA(a, b, -product)
		if product > 0 && error < 0 || product < 0 && error > 0 {
			return math.Nextafter(product, 0)
		}
	}
	return product
}

func drainManaConvert52E210(value float32, rt SpellDrainManaRuntime52E210) int32 {
	v := float64(value)
	if !rt.RoundTowardZero {
		v = math.RoundToEven(v)
	}
	if math.IsNaN(v) || v >= 2147483648 || v < -2147483648 {
		return math.MinInt32
	}
	return int32(v)
}

func drainManaBalance52E610(rt SpellDrainManaRuntime52E210) float64 {
	if rt.Balance64 != nil {
		return rt.Balance64("ManaDrainRange")
	}
	return float64(rt.Balance("ManaDrainRange"))
}

func drainManaCoefficient52E210(level uint32, rt SpellDrainManaRuntime52E210) float64 {
	if rt.BalanceLevel64 != nil {
		return rt.BalanceLevel64("ManaDrainCoeff", level)
	}
	return float64(rt.BalanceLevel("ManaDrainCoeff", level))
}

func spellDrainManaTargetHasMana52E7C0(target *Object) bool {
	class := uint32(target.ObjClass)
	if class&0x400000 != 0 && uint8(target.ObjSubClass)&0x18 != 0 {
		return (*ObeliskUpdateData)(target.UpdateData).Mana > 0
	}
	if class&2 != 0 {
		return uint32((*MonsterUpdateData)(target.UpdateData).StatusFlags)&0x20 != 0
	}
	return class&4 != 0 && UnitGetOldMana4EEC80(target) != 0
}

func spellDrainManaClosest52E610(center types.Pointf, caster *Object, rt SpellDrainManaRuntime52E210) *Object {
	return spellDrainManaClosestPoint52E610(&center, caster, rt)
}

// The selector copies scratch before balance lookup, but enumeration receives
// the original point pointer's live value after that lookup. The caster
// callback context stays fixed. Candidate update pointers are cached before
// enemy/same-team callbacks, while player links and flags are then reloaded.
func spellDrainManaClosestPoint52E610(point *types.Pointf, caster *Object, rt SpellDrainManaRuntime52E210) *Object {
	s := rt.Selection
	if s == nil {
		s = new(SpellDrainManaSelection52E610)
	}
	*s = SpellDrainManaSelection52E610{Center: *point, Nearest: 100000000}
	radius := drainManaSpill52E210(drainManaBalance52E610(rt), rt)
	rt.ObjectsInCircle(*point, radius, func(target *Object) bool {
		if target == caster || !spellDrainManaTargetHasMana52E7C0(target) || uint32(target.ObjFlags)&0x8020 != 0 {
			return true
		}
		class := uint32(target.ObjClass)
		weight := float64(1)
		if class&2 != 0 {
			data := (*MonsterUpdateData)(target.UpdateData)
			if caster != nil && !rt.IsEnemy(caster, target) || uint32(data.StatusFlags)&0x20 == 0 {
				return true
			}
		} else if class&0x400000 == 0 || uint8(target.ObjSubClass)&0x18 == 0 {
			if class&4 == 0 {
				return true
			}
			data := (*PlayerUpdateData)(target.UpdateData)
			if caster != nil && !rt.IsEnemy(caster, target) {
				return true
			}
			if rt.SameTeam(target, caster) || data.Player.Field3680&1 != 0 {
				return true
			}
			weight = .5
		}
		dx := drainManaSubtract52E210(float64(s.Center.X), float64(target.PosVec.X), rt)
		dy := drainManaSubtract52E210(float64(s.Center.Y), float64(target.PosVec.Y), rt)
		ySquare := drainManaMultiply52E450(dy, dy, rt)
		xSquare := drainManaMultiply52E450(dx, dx, rt)
		score := drainManaMultiply52E450(drainManaAdd52E210(ySquare, xSquare, rt), weight, rt)
		// FST saves the binary32 score before the unspilled PC53 comparison
		// and trace. The saved score survives callback scratch mutations.
		stored := drainManaSpill52E210(score, rt)
		if !(score >= float64(s.Nearest)) && rt.TraceRay(s.Center, target.PosVec, MapTraceFlags(5)) {
			s.Target, s.Nearest = target, stored
		}
		return true
	})
	return s.Target
}

func drainManaCredit52E450(caster *Object, amount int32, rt SpellDrainManaRuntime52E210) {
	if rt.AddMana32 != nil {
		rt.AddMana32(caster, amount)
	} else {
		rt.AddMana(caster, int16(amount))
	}
}

func spellDrainManaTransfer52E450(caster, target *Object, amount int32, rt SpellDrainManaRuntime52E210) bool {
	if caster == nil {
		return false
	}
	if uint8(caster.ObjClass)&4 != 0 {
		data := (*PlayerUpdateData)(caster.UpdateData)
		if UnitGetOldMana4EEC80(caster) >= data.ManaMax {
			return false
		}
	}
	class := uint32(target.ObjClass)
	transferred := amount
	switch {
	case class&0x400000 != 0 && uint8(target.ObjSubClass)&0x18 != 0:
		data := (*ObeliskUpdateData)(target.UpdateData)
		if target.HasTeam() && !target.TeamVal.SameAs(&caster.TeamVal) {
			return false
		}
		if available := data.Mana; available <= amount {
			transferred, amount = available, available
			if !rt.QuestMode() {
				data.Mana = 0
			}
		} else if !rt.QuestMode() {
			data.Mana -= amount
		}
		if transferred == 0 {
			return false
		}
		if !rt.QuestMode() {
			target.NeedSync()
		}
	case class&2 != 0:
		if uint32((*MonsterUpdateData)(target.UpdateData).StatusFlags)&0x20 != 0 {
			transferred, amount = 1, 1
		}
	case class&4 != 0:
		current := int32((*PlayerUpdateData)(target.UpdateData).ManaCur)
		if current <= amount {
			transferred, amount = current, current
		}
		rt.SubMana(target, amount)
	}
	if transferred <= 0 {
		return false
	}
	if rt.QuestMode() && uint8(caster.ObjClass)&4 != 0 {
		class := uint8((*PlayerUpdateData)(caster.UpdateData).Player.PlayerClass())
		if class <= 2 {
			value := drainManaMultiply52E450(float64(amount), float64(rt.QuestManaScale(caster)), rt)
			transferred = drainManaConvert52E210(drainManaSpill52E210(value, rt), rt)
		}
	}
	drainManaCredit52E450(caster, transferred, rt)
	return true
}

// SpellDrainManaUpdate52E210 replaces the public float-argument ABI and PE32
// field accesses while retaining original live links and startup arithmetic.
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
		point := record.Pos
		if caster != nil {
			point = caster.PosVec
		}
		if target := spellDrainManaClosestPoint52E610(&point, caster, rt); target != nil {
			rt.SubMana(target, 50)
		}
		return 1
	}
	if uint8(caster.ObjClass)&4 != 0 && UnitGetOldMana4EEC80(caster) >= PlayerGetMaxMana4EECB0(caster) {
		return 1
	}
	caster = record.Caster16
	if uint8(caster.ObjClass)&2 != 0 && rt.PositionDelta(caster, &record.Pos) != 0 {
		return 1
	}
	caster = record.Caster16
	dx := drainManaSubtract52E210(float64(record.Pos.X), float64(caster.PosVec.X), rt)
	if !(dx >= 0) {
		dx = -dx
	}
	x := drainManaSpill52E210(dx, rt)
	dy := drainManaSubtract52E210(float64(record.Pos.Y), float64(caster.PosVec.Y), rt)
	if !(dy >= 0) {
		dy = -dy
	}
	y := drainManaSpill52E210(dy, rt)
	if caster.HealthData != nil && rt.Frame()-caster.Frame134 <= 1 || x >= 5 || y >= 5 {
		return 1
	}
	caster = record.Caster16
	record.Target48 = spellDrainManaClosestPoint52E610(&caster.PosVec, caster, rt)
	if record.Target48 == nil {
		if previous := rt.LoadRayTarget(record); previous != nil {
			rt.StopRay(record, previous)
			// Unlike the original dying record's numeric +36 link, a native
			// private sidecar has its own owner lifetime and is released here.
			rt.StoreRayTarget(record, nil)
		}
		return 1
	}
	carry := math.Float32frombits(uint32(record.Field72))
	value := drainManaSpill52E210(drainManaAdd52E210(drainManaCoefficient52E210(record.Level-1, rt), float64(carry), rt), rt)
	amount := drainManaConvert52E210(value, rt)
	target, previous := record.Target48, rt.LoadRayTarget(record)
	record.Field72 = int32(math.Float32bits(drainManaSpill52E210(drainManaAdd52E210(float64(value), -float64(amount), rt), rt)))
	if target != previous {
		if previous != nil {
			rt.StopRay(record, previous)
		}
		rt.StartRay(record)
	}
	if spellDrainManaTransfer52E450(record.Caster16, record.Target48, drainManaConvert52E210(value, rt), rt) && rt.Frame()%(rt.TickRate()>>1) == 0 {
		rt.Audio(230, record.Caster16)
		rt.Audio(229, record.Target48)
	}
	rt.StoreRayTarget(record, record.Target48)
	return 0
}
