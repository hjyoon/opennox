package server

import (
	"math"

	"github.com/opennox/libs/types"
)

// SpellPlasmaRuntime531580 binds the original Plasma callbacks to game
// services. Field72 and Field36 were PE32 pointers, not integer values: the
// weapon and previous ray target therefore live in native-width sidecars.
type SpellPlasmaRuntime531580 struct {
	HecubahType, HecubahOrbType *uint32
	LookupType                  func(string) uint32
	Frame, TickRate             func() uint32
	Balance                     func(string) float32
	ObjectsInCircle             func(types.Pointf, float32, func(*Object) bool)
	IsEnemy, CanInteract        func(*Object, *Object) bool
	Facing                      func(*Object, *Object) int
	Distance                    func(*Object, *Object) float64
	PositionDelta               func(*Object, *types.Pointf) int32
	StartRay                    func(*DurSpell)
	StopRay                     func(*DurSpell, *Object)
	PointFX                     func(uint8, types.Pointf)
	Damage                      func(*Object, *Object, int32)
	Audio                       func(uint16, *Object)
	SetPlayerState              func(*Object, PlayerState)
	ReportCharges               func(*PlayerUpdateData, *Object, uint8, uint8)
	LoadWeapon, LoadRayTarget   func(*DurSpell) *Object
	StoreWeapon, StoreRayTarget func(*DurSpell, *Object)
}

// SpellPlasmaCreate531580 restores GAME.EXE 00531580 without reading native
// objects through the old +8/+748 offsets or narrowing the equipped weapon.
func SpellPlasmaCreate531580(record *DurSpell, rt SpellPlasmaRuntime531580) int32 {
	record.Field72 = 0
	rt.StoreWeapon(record, nil)
	record.Field76 = 0
	record.Target48 = nil
	rt.PointFX(131, record.Pos)
	// The point-effect callback precedes the caster and equipment reads.
	caster := record.Caster16
	if caster == nil || uint8(caster.ObjClass)&4 == 0 {
		return 0
	}
	wand := caster.UpdateDataPlayer().EquippedWeapon
	if wand == nil || uint32(wand.ObjSubClass)&0x4000000 == 0 {
		return 1
	}
	if wand.UseData.AsWand().Flags&4 != 0 {
		rt.StoreWeapon(record, wand)
		record.Flags88 |= 2
	}
	return 0
}

func plasmaBalanceInt531600(value float32) int32 {
	// 00419A77 FISTPL runs in the gameplay x87 round-toward-zero mode.
	// Masked invalid conversions produce the original integer-indefinite word.
	if math.IsNaN(float64(value)) || value >= 2147483648 || value < -2147483648 {
		return math.MinInt32
	}
	return int32(value)
}

func plasmaClosest531920(caster *Object, rt SpellPlasmaRuntime531580) *Object {
	nearest := float32(160000) // Original 481C4000 DWORD, 400 squared.
	var best *Object
	rt.ObjectsInCircle(caster.PosVec, 400, func(target *Object) bool {
		class := uint32(target.ObjClass)
		if class&0x20006 == 0 || uint32(target.ObjFlags)&0x8020 != 0 || target == caster ||
			class&2 != 0 && uint32(target.ObjSubClass)&0x8000 != 0 || !rt.IsEnemy(caster, target) {
			return true
		}
		// 00531979 AND EAX,1 / OR AL,12 still calls the facing service, but
		// its result cannot reject a candidate. Do not turn this into '&13'.
		_ = rt.Facing(caster, target)
		if !rt.CanInteract(caster, target) {
			return true
		}
		dx := monsterMoveToRunAddChop53_544434(float64(target.PosVec.X), -float64(caster.PosVec.X))
		dy := monsterMoveToRunAddChop53_544434(float64(target.PosVec.Y), -float64(caster.PosVec.Y))
		square := monsterMoveToRunAddChop53_544434(monsterMoveToRunSquareChop53_544434(dy), monsterMoveToRunSquareChop53_544434(dx))
		// FCOM tests C0 only: less or unordered selects. Compare the live
		// binary64 register with the previous binary32 spill, before spilling.
		if !(square >= float64(nearest)) {
			nearest = float32(monsterMoveToRunSpill544440(square))
			best = target
		}
		return true
	})
	return best
}

// SpellPlasmaUpdate531600 restores the retained-target grace period, ray
// changes, damage and per-hit wand-charge consumption of GAME.EXE 00531600.
func SpellPlasmaUpdate531600(record *DurSpell, rt SpellPlasmaRuntime531580) int32 {
	if *rt.HecubahType == 0 {
		*rt.HecubahType = rt.LookupType("Hecubah")
		*rt.HecubahOrbType = rt.LookupType("HecubahWithOrb")
	}
	caster := record.Caster16
	if caster == nil || uint8(record.Flags88)&0x20 != 0 ||
		uint8(caster.ObjClass)&2 != 0 && rt.PositionDelta(caster, &record.Pos) != 0 {
		return 1
	}
	if target := record.Target48; target != nil {
		if uint32(target.ObjFlags)&0x8020 != 0 {
			record.Target48 = nil
		} else {
			if rt.Facing(record.Caster16, target)&2 != 0 {
				if uint32(record.Field76) == 0 {
					record.Field76 = uintptr(3 * rt.TickRate())
				}
			} else {
				record.Field76 = 0
			}
			// The original range is collision-surface distance, not centers.
			if rt.Distance(record.Target48, record.Caster16) > 400 ||
				!rt.CanInteract(record.Caster16, record.Target48) {
				record.Target48 = nil
			}
		}
	}
	if record.Target48 == nil {
		caster = record.Caster16
		var selected *Object
		// +288 is CursorObj only in PlayerUpdateData; the monster DWORD at
		// the same PE32 offset is not a pointer in the native monster record.
		if uint8(caster.ObjClass)&4 != 0 {
			cursor := caster.UpdateDataPlayer().CursorObj
			if cursor != nil && rt.IsEnemy(caster, cursor) && rt.Distance(caster, cursor) <= 400 {
				selected = cursor
			}
		}
		if selected == nil {
			selected = plasmaClosest531920(record.Caster16, rt)
		}
		record.Target48 = selected
		record.Field76 = 0
	}
	if grace := uint32(record.Field76); grace != 0 {
		grace--
		record.Field76 = uintptr(grace)
		if grace == 0 {
			record.Target48 = nil
		}
	}
	target, previous := record.Target48, rt.LoadRayTarget(record)
	if target == nil {
		if previous != nil {
			rt.StopRay(record, previous)
			rt.StoreRayTarget(record, nil)
		}
		return 0
	}
	if target != previous {
		if previous != nil {
			rt.StopRay(record, previous)
		}
		rt.StartRay(record)
	}
	key := "PlasmaDamage"
	if typ := uint32(record.Target48.TypeInd); typ == *rt.HecubahType || typ == *rt.HecubahOrbType {
		key = "PlasmaDamageHecubah"
	}
	damage := plasmaBalanceInt531600(rt.Balance(key))
	rt.Damage(record.Target48, record.Caster16, damage)
	if uint32(record.Target48.ObjFlags)&0x8020 != 0 {
		rt.PointFX(131, record.Target48.PosVec)
	}
	rt.StoreRayTarget(record, record.Target48)
	rt.SetPlayerState(record.Caster16, 22)
	if rt.Frame()%(rt.TickRate()/3) == 0 {
		rt.Audio(98, record.Caster16)
		rt.Audio(98, record.Target48)
	}
	if uint32(record.Field76) == 0 {
		searchTime := plasmaBalanceInt531600(rt.Balance("PlasmaSearchTime"))
		record.Frame68 = rt.Frame() + uint32(searchTime)
	}
	if wand := rt.LoadWeapon(record); wand != nil {
		data := wand.UseData.AsWand()
		if data.Charge == 0 {
			return 1
		}
		data.Charge--
		data.Progress = 100 * uint32(data.Charge) / uint32(data.MaxCharge)
		if caster = record.Caster16; caster != nil && uint8(caster.ObjClass)&4 != 0 {
			update := caster.UpdateDataPlayer()
			rt.SetPlayerState(caster, 22)
			rt.ReportCharges(update, rt.LoadWeapon(record), data.Charge, data.MaxCharge)
		}
		if data.Charge == 0 {
			return 1
		}
	}
	return 0
}

// SpellPlasmaDestroy5319E0 releases the active wand bit, preserving all
// unrelated DWORD flags. Ray stop delivery remains with the duration engine.
func SpellPlasmaDestroy5319E0(record *DurSpell, rt SpellPlasmaRuntime531580) {
	if wand := rt.LoadWeapon(record); wand != nil {
		wand.UseData.AsWand().Flags &^= 4
	}
	rt.StoreWeapon(record, nil)
}
