package server

import (
	"math"

	"github.com/opennox/libs/types"
)

// SpellEnergyBoltRuntime52E820 keeps the external services used by the
// Lightning duration callbacks separate from their native-width record logic.
type SpellEnergyBoltRuntime52E820 struct {
	Frame          func() uint32
	TickRate       func() uint32
	Balance        func(string) float32
	BalanceLevel   func(string, uint32) float32
	Balance64      func(string) float64
	BalanceLevel64 func(string, uint32) float64
	Distance       func(*Object, *Object) float64
	Selection      *SpellEnergyBoltSelection52EC60
	// The standalone API retains its nearest profile. Production explicitly
	// selects the original gameplay PC53/toward-zero context.
	RoundTowardZero bool
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

type SpellEnergyBoltSelection52EC60 struct {
	Mode    uint32
	Center  types.Pointf
	Target  *Object
	Nearest float32
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

func energyBoltSpill52E850(value float64, rt SpellEnergyBoltRuntime52E820) float32 {
	if rt.RoundTowardZero {
		return float32(monsterMoveToRunSpill544440(value))
	}
	return float32(value)
}

//go:noinline
func energyBoltAdd52E850(a, b float64, rt SpellEnergyBoltRuntime52E820) float64 {
	if rt.RoundTowardZero {
		return monsterMoveToRunAddChop53_544434(a, b)
	}
	return a + b
}

//go:noinline
func energyBoltSquare52EC60(v float64, rt SpellEnergyBoltRuntime52E820) float64 {
	if rt.RoundTowardZero {
		return monsterMoveToRunSquareChop53_544434(v)
	}
	return v * v
}

func energyBoltConvert52E850(value float32, rt SpellEnergyBoltRuntime52E820) int32 {
	v := float64(value)
	if !rt.RoundTowardZero {
		v = math.RoundToEven(v)
	}
	if math.IsNaN(v) || v >= 2147483648 || v < -2147483648 {
		return math.MinInt32
	}
	return int32(v)
}

func energyBoltBalance52E850(key string, rt SpellEnergyBoltRuntime52E820) float64 {
	if rt.Balance64 != nil {
		return rt.Balance64(key)
	}
	return float64(rt.Balance(key))
}

func energyBoltBalanceLevel52E850(level uint32, rt SpellEnergyBoltRuntime52E820) float64 {
	if rt.BalanceLevel64 != nil {
		return rt.BalanceLevel64("EnergyBoltDamage", level)
	}
	return float64(rt.BalanceLevel("EnergyBoltDamage", level))
}

func energyBoltDistance52E850(a, b *Object, rt SpellEnergyBoltRuntime52E820) float64 {
	if rt.Distance != nil {
		return rt.Distance(a, b)
	}
	return ObjectDistance4E6C00(a, b)
}

func energyBoltEnumerate52EC60(caster *Object, point types.Pointf, radius float32, rt SpellEnergyBoltRuntime52E820) {
	s := rt.Selection
	rt.ObjectsInCircle(point, radius, func(target *Object) bool {
		class := uint32(target.ObjClass)
		if class&0x20006 == 0 || uint32(target.ObjFlags)&0x8020 != 0 || target == caster || class&2 != 0 && uint32(target.ObjSubClass)&0x8000 != 0 {
			return true
		}
		if caster != nil {
			if !rt.IsEnemy(caster, target) {
				return true
			}
			// Mode is read after enemy testing. Nil source bypasses all three
			// services; glyph mode bypasses only facing and visibility.
			if s.Mode == 0 && (!rt.InFront(caster, target) || !rt.CanInteract(caster, target)) {
				return true
			}
		}
		dx := energyBoltAdd52E850(float64(target.PosVec.X), -float64(s.Center.X), rt)
		dy := energyBoltAdd52E850(float64(target.PosVec.Y), -float64(s.Center.Y), rt)
		square := energyBoltAdd52E850(energyBoltSquare52EC60(dy, rt), energyBoltSquare52EC60(dx, rt), rt)
		// FCOM's C0 test admits less and unordered before the binary32 spill.
		if !(square >= float64(s.Nearest)) {
			s.Nearest = energyBoltSpill52E850(square, rt)
			s.Target = target
		}
		return true
	})
}

// SpellEnergyBoltUpdate52E850 executes Lightning's original glyph and sustained
// ray paths with native record links. Both original PC53 rounding contexts are
// supported; production supplies the gameplay toward-zero context explicitly.
func SpellEnergyBoltUpdate52E850(record *DurSpell, rt SpellEnergyBoltRuntime52E820) int32 {
	if caster := record.Caster16; caster != nil {
		if caster.UnitBuffTest4FF350(8) != 0 {
			return 1
		}
	} else if record.Flag20 == 0 {
		return 1
	}
	if rt.Selection == nil {
		rt.Selection = new(SpellEnergyBoltSelection52EC60)
	}
	radius := energyBoltSpill52E850(energyBoltBalance52E850("LightningRange", rt), rt)
	s := rt.Selection
	if record.Flag20 != 0 {
		point := record.Pos
		s.Center, s.Target, s.Nearest, s.Mode = point, nil, energyBoltSpill52E850(float64(radius)*float64(radius), rt), 1
		energyBoltEnumerate52EC60(record.Caster16, point, radius, rt)
		if s.Target != nil {
			damage := energyBoltConvert52E850(energyBoltSpill52E850(energyBoltBalance52E850("EnergyBoltGlyphDamage", rt), rt), rt)
			rt.Damage(s.Target, record.Obj12, damage)
			// The audio object is cached before cast-sound lookup; FX reloads
			// the shared selected target after the audio callback.
			target := s.Target
			id := rt.CastSound()
			rt.Audio(id, target)
			rt.PointFX(130, s.Target.PosVec)
		}
		return 1
	}
	caster := record.Caster16
	if caster != nil && uint8(caster.ObjClass)&2 != 0 && rt.PositionDelta(caster, &record.Pos) != 0 {
		return 1
	}
	if rt.Frame()-record.Frame60 > 2 {
		caster = record.Caster16
		if caster.HealthData != nil && rt.Frame()-caster.Frame134 <= 1 {
			return 1
		}
	}
	if target := record.Target48; target != nil {
		if uint32(target.ObjFlags)&0x8020 != 0 || !rt.InFront(record.Caster16, target) ||
			energyBoltDistance52E850(record.Target48, record.Caster16, rt) > float64(radius) || !rt.CanInteract(record.Caster16, record.Target48) {
			record.Target48 = nil
		}
	}
	if record.Target48 == nil {
		caster = record.Caster16
		if uint8(caster.ObjClass)&4 != 0 {
			if cursor := caster.UpdateDataPlayer().CursorObj; cursor != nil && rt.IsEnemy(caster, cursor) && !(energyBoltDistance52E850(record.Caster16, cursor, rt) > float64(radius)) {
				record.Target48 = cursor
			}
		}
		if record.Target48 == nil {
			caster = record.Caster16
			s.Mode, s.Target, s.Center, s.Nearest = 0, nil, caster.PosVec, energyBoltSpill52E850(float64(radius)*float64(radius), rt)
			energyBoltEnumerate52EC60(caster, caster.PosVec, radius, rt)
			record.Target48 = s.Target
		}
	}
	if record.Target48 == nil {
		if previous := rt.LoadRayTarget(record); previous != nil {
			rt.StopRay(record, previous)
			rt.StoreRayTarget(record, nil)
		}
		return 0
	}
	// The original saves carry before the table callback, then spills the
	// PC53 sum before both FISTP calls. Keep integer-indefinite deterministic.
	carry := math.Float32frombits(uint32(record.Field72))
	value := energyBoltSpill52E850(energyBoltAdd52E850(energyBoltBalanceLevel52E850(record.Level-1, rt), float64(carry), rt), rt)
	damage := energyBoltConvert52E850(value, rt)
	target, previous := record.Target48, rt.LoadRayTarget(record)
	record.Field72 = int32(math.Float32bits(energyBoltSpill52E850(energyBoltAdd52E850(float64(value), -float64(damage), rt), rt)))
	if target != previous {
		if previous != nil {
			rt.StopRay(record, previous)
		}
		rt.StartRay(record)
	}
	rt.Damage(record.Target48, record.Caster16, energyBoltConvert52E850(value, rt))
	if uint32(record.Target48.ObjFlags)&0x8020 != 0 {
		rt.PointFX(130, record.Target48.PosVec)
	}
	rt.StoreRayTarget(record, record.Target48)
	if caster = record.Caster16; uint8(caster.ObjClass)&4 != 0 {
		rt.SetPlayerState(caster, 10)
	}
	period := rt.TickRate() / 3
	if rt.Frame()%period == 0 {
		rt.Audio(32, record.Caster16)
		rt.Audio(32, record.Target48)
	}
	search := energyBoltConvert52E850(energyBoltSpill52E850(energyBoltBalance52E850("LightningSearchTime", rt), rt), rt)
	record.Frame68 = rt.Frame() + uint32(search)
	if caster = record.Caster16; uint8(caster.ObjClass)&4 != 0 {
		rt.SetPlayerState(caster, 10)
		rt.ManaSub(record.Caster16, 1)
		if UnitGetOldMana4EEC80(record.Caster16) == 0 {
			return 1
		}
	}
	if uint32(record.Target48.ObjFlags)&0x8000 != 0 {
		record.Frame68 = rt.Frame() + 1
	}
	return 0
}
