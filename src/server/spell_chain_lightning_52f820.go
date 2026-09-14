package server

import (
	"math"

	"github.com/opennox/libs/types"
)

// SpellChainLightningRuntime52F820 supplies the game services used by the
// three duration callbacks. Object and duration references remain native-width.
type SpellChainLightningRuntime52F820 struct {
	Frame           func() uint32
	TickRate        func() uint32
	Balance         func(string) float32
	SpellLevel      func(uint32) uint32
	ObjectsInCircle func(types.Pointf, float32, func(*Object) bool)
	CanInteract     func(*Object, *Object) bool
	IsEnemy         func(*Object, *Object) bool
	TraceRay        func(types.Pointf, types.Pointf, MapTraceFlags) bool
	PositionDelta   func(*Object, *types.Pointf) int32
	CancelSpell     func(int32, *Object)
	FreeDuration    func(*DurSpell)
	NewLightningSub func(*DurSpell, *Object, *Object)
	StartRay        func(*DurSpell)
	StopRay         func(*DurSpell, *Object)
	PointFX         func(uint8, types.Pointf)
	Damage          func(*Object, *Object, int32)
	Audio           func(uint16, *Object)
	CastSound       func() uint16
	SetPlayerState  func(*Object, PlayerState)
	ManaSub         func(*Object, int32)
	ReportCharges   func(*Object, *Object, uint8, uint8)
	LoadWeapon      func(*DurSpell) *Object
	StoreWeapon     func(*DurSpell, *Object)
}

// EachChainLightningObject52F8A0 uses the same spatial object index as the
// legacy unitsGetInCircle callback, without passing object pointers through C.
func (s *Server) EachChainLightningObject52F8A0(pos types.Pointf, radius float32, visit func(*Object) bool) {
	s.Map.EachObjInCircle(pos, radius, visit)
}

// SpellChainLightningCreate52F820 follows GAME.EXE 0052F820. The equipped
// wand is stored in a sidecar: Field72 was only a PE32 pointer slot.
func SpellChainLightningCreate52F820(record *DurSpell, rt SpellChainLightningRuntime52F820) int32 {
	record.Field72 = 0
	record.Field76 = 0
	caster := record.Caster16
	if caster != nil {
		rt.CancelSpell(24, caster) // Lightning and Chain Lightning are exclusive.
	}
	if record.Flag20 == 0 && caster != nil && uint8(caster.ObjClass)&4 != 0 {
		if update := caster.UpdateDataPlayer(); update != nil {
			if wand := update.EquippedWeapon; wand != nil && uint32(wand.ObjSubClass)&0x40000 != 0 {
				if data := wand.UseData.AsWand(); data != nil && data.Flags&4 != 0 {
					rt.StoreWeapon(record, wand)
					record.Flags88 |= 2
				}
			}
		}
	}
	rt.PointFX(129, record.Pos)
	return 0
}

func chainLightningFreeList530100(head *DurSpell, rt SpellChainLightningRuntime52F820) {
	for cur := head; cur != nil; {
		next := cur.Next
		rt.FreeDuration(cur)
		cur = next
	}
}

// SpellChainLightningDestroy530100 frees both ray generations and releases
// the wand mode. A pointer to the wand never passes through Field72.
func SpellChainLightningDestroy530100(record *DurSpell, rt SpellChainLightningRuntime52F820) {
	chainLightningFreeList530100(record.Sub108, rt)
	record.Sub108 = nil
	chainLightningFreeList530100(record.Sub104, rt)
	record.Sub104 = nil
	if wand := rt.LoadWeapon(record); wand != nil {
		if data := wand.UseData.AsWand(); data != nil {
			data.Flags &^= 4
		}
	}
	rt.StoreWeapon(record, nil)
}

func chainLightningClosest52FF10(center types.Pointf, radius float32, source, owner *Object, selected []*Object, rt SpellChainLightningRuntime52F820) *Object {
	bestDistance := radius * radius
	var best *Object
	rt.ObjectsInCircle(center, radius, func(target *Object) bool {
		if target == source || target == owner || uint32(target.ObjClass)&0x20006 == 0 ||
			uint32(target.ObjFlags)&0x8020 != 0 ||
			uint8(target.ObjClass)&2 != 0 && uint32(target.ObjSubClass)&0x8000 != 0 {
			return true
		}
		if owner != nil && !rt.IsEnemy(owner, target) {
			return true
		}
		for _, previous := range selected {
			if target == previous {
				return true
			}
		}
		if !rt.CanInteract(source, target) {
			return true
		}
		dx := target.PosVec.X - source.PosVec.X
		dy := target.PosVec.Y - source.PosVec.Y
		distance := dx*dx + dy*dy
		if distance < bestDistance {
			bestDistance = distance
			best = target
		}
		return true
	})
	return best
}

func chainLightningTrapEffect530020(record *DurSpell, radius float32, rt SpellChainLightningRuntime52F820) {
	owner := record.Caster16
	rt.ObjectsInCircle(record.Pos, radius, func(target *Object) bool {
		if target == owner || uint32(target.ObjClass)&6 == 0 || uint32(target.ObjFlags)&0x8020 != 0 ||
			uint8(target.ObjClass)&2 != 0 && uint32(target.ObjSubClass)&0x8000 != 0 ||
			owner != nil && !rt.IsEnemy(owner, target) || !rt.TraceRay(record.Pos, target.PosVec, MapTraceFlags(9)) {
			return true
		}
		damage := spellDurationRoundNearestEven(rt.Balance("LightningGlyphDamage"))
		rt.Damage(target, nil, damage)
		rt.PointFX(129, target.PosVec)
		rt.Audio(rt.CastSound(), target)
		return true
	})
}

func chainLightningBuildTargets52F8A0(record *DurSpell, radius float32, rt SpellChainLightningRuntime52F820) [5]*Object {
	caster := record.Caster16
	var targets [5]*Object
	if uint8(caster.ObjClass)&4 != 0 {
		if update := caster.UpdateDataPlayer(); update != nil {
			if cursor := update.CursorObj; cursor != nil && rt.IsEnemy(caster, cursor) {
				dx := cursor.PosVec.X - caster.PosVec.X
				dy := cursor.PosVec.Y - caster.PosVec.Y
				if dx*dx+dy*dy <= radius*radius {
					targets[0] = cursor
				}
			}
		}
	}
	if targets[0] == nil {
		targets[0] = chainLightningClosest52FF10(record.Pos, radius, caster, caster, nil, rt)
	}
	if targets[0] == nil {
		return targets
	}
	level := rt.SpellLevel(record.Spell)
	count := 1
	search := func(source *Object, scale float32) {
		if source == nil || count >= len(targets) {
			return
		}
		if target := chainLightningClosest52FF10(source.PosVec, radius*scale, source, caster, targets[:count], rt); target != nil {
			targets[count] = target
			count++
		}
	}
	if level > 1 {
		search(targets[0], 0.94999999)
	}
	if level > 2 {
		search(targets[0], 0.89999998)
	}
	if level > 3 && targets[1] != nil {
		search(targets[1], 0.85000002)
	}
	if level > 4 && targets[2] != nil {
		search(targets[2], 0.80000001)
	}
	return targets
}

func chainLightningSyncRays52F8A0(record *DurSpell, rt SpellChainLightningRuntime52F820) {
	old := record.Sub104
	damageFloat := rt.Balance("LightningDamage") + math.Float32frombits(uint32(record.Field76))
	damage := spellDurationRoundNearestEven(damageFloat)
	record.Field76 = uintptr(math.Float32bits(damageFloat - float32(damage)))
	for ray := record.Sub108; ray != nil; ray = ray.Next {
		if old != nil {
			if old.Target48 != ray.Target48 || old.Caster16 != ray.Caster16 {
				if old.Target48 != nil {
					rt.StopRay(old, old.Target48)
				}
				rt.StartRay(ray)
			}
			old = old.Next
		} else {
			rt.StartRay(ray)
		}
		if damage > 0 {
			rt.Damage(ray.Target48, record.Caster16, damage)
		}
		if uint32(ray.Target48.ObjFlags)&0x8020 != 0 {
			rt.PointFX(129, ray.Target48.PosVec)
		}
	}
	for ; old != nil; old = old.Next {
		if old.Target48 != nil {
			rt.StopRay(old, old.Target48)
		}
	}
}

// SpellChainLightningUpdate52F8A0 follows GAME.EXE 0052F8A0 without the
// decoded C float-argument ABI mismatch or 32-bit object references.
func SpellChainLightningUpdate52F8A0(record *DurSpell, rt SpellChainLightningRuntime52F820) int32 {
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
		chainLightningTrapEffect530020(record, radius, rt)
		return 1
	}
	frame := rt.Frame()
	if uint8(caster.ObjClass)&4 != 0 && UnitGetOldMana4EEC80(caster) == 0 {
		return 1
	}
	if frame-record.Frame60 > 2 && caster.HealthData != nil && frame-caster.Frame134 <= 1 {
		return 1
	}
	if uint8(caster.ObjClass)&2 != 0 && rt.PositionDelta(caster, &record.Pos) != 0 {
		return 1
	}
	chainLightningFreeList530100(record.Sub104, rt)
	record.Sub104 = record.Sub108
	record.Sub108 = nil
	targets := chainLightningBuildTargets52F8A0(record, radius, rt)
	if targets[0] == nil {
		for ray := record.Sub104; ray != nil; ray = ray.Next {
			if ray.Target48 != nil {
				rt.StopRay(ray, ray.Target48)
			}
		}
		chainLightningFreeList530100(record.Sub104, rt)
		record.Sub104 = nil
		return 0
	}
	level := rt.SpellLevel(record.Spell)
	rt.NewLightningSub(record, caster, targets[0])
	if level > 1 && targets[1] != nil {
		rt.NewLightningSub(record, targets[0], targets[1])
	}
	if level > 2 && targets[2] != nil {
		rt.NewLightningSub(record, targets[0], targets[2])
	}
	if level > 3 && targets[3] != nil {
		from := targets[1]
		if from == nil {
			from = targets[2]
		}
		if from != nil {
			rt.NewLightningSub(record, from, targets[3])
		}
	}
	if level > 4 && targets[4] != nil {
		from := targets[2]
		if from == nil {
			from = targets[1]
		}
		if from != nil {
			rt.NewLightningSub(record, from, targets[4])
		}
	}
	chainLightningSyncRays52F8A0(record, rt)
	if uint8(caster.ObjClass)&4 != 0 {
		if wand := rt.LoadWeapon(record); wand != nil {
			data := wand.UseData.AsWand()
			if data.Charge == 0 {
				return 1
			}
			data.Charge--
			if data.MaxCharge != 0 {
				data.Progress = 100 * uint32(data.Charge) / uint32(data.MaxCharge)
			}
			rt.SetPlayerState(caster, 22)
			rt.ReportCharges(caster, wand, data.Charge, data.MaxCharge)
			if data.Charge == 0 {
				return 1
			}
		} else {
			rt.SetPlayerState(caster, 10)
			rt.ManaSub(caster, 1)
			if UnitGetOldMana4EEC80(caster) == 0 {
				return 1
			}
		}
	}
	if period := rt.TickRate() / 3; period != 0 && frame%period == 0 {
		rt.Audio(78, caster)
		rt.Audio(78, targets[0])
	}
	record.Frame68 = frame + uint32(spellDurationRoundNearestEven(rt.Balance("LightningSearchTime")))
	return 0
}
