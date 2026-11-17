package server

import (
	"math"
	"runtime"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

// MonsterSpawnScreenRecord50D960 replaces the 148-byte PE32 scratch record.
// Its owner remains pinned while a public native C query can return it.
type MonsterSpawnScreenRecord50D960 struct {
	Object    *Object
	Average   float32
	Mask      uint32
	Distances [32]float32
	Older     *MonsterSpawnScreenRecord50D960
	Newer     *MonsterSpawnScreenRecord50D960
}

type monsterSpawnScreenOwner50D960 struct {
	pin  *runtime.Pinner
	free func(*MonsterSpawnScreenRecord50D960)
}

func (s *Server) monsterSpawnRuntime50D890(rt MonsterSpawnDeleteRuntime50E210) MonsterSpawnDeleteRuntime50E210 {
	if rt.FirstPlayerUnit == nil {
		rt.FirstPlayerUnit = s.Players.FirstUnit
	}
	if rt.NextPlayerUnit == nil {
		rt.NextPlayerUnit = s.Players.NextUnit
	}
	if rt.MaxOnscreen == nil {
		rt.MaxOnscreen = func() float64 { return s.Balance.Float("MaxOnscreenMonsterCount") }
	}
	if rt.PointInRect == nil {
		rt.PointInRect = monsterSpawnPointInRect50DE80
	}
	if rt.Distance == nil {
		rt.Distance = ObjectDistance4E6C00
	}
	if rt.Trace == nil {
		rt.Trace = func(from, to types.Pointf, flags MapTraceFlags) int {
			if s.MapTraceRay(from, to, flags) {
				return 1
			}
			return 0
		}
	}
	if rt.EachUnitInRect == nil {
		rt.EachUnitInRect = func(rect types.Rectf, visit func(*Object)) {
			s.Map.EachObjInRect(rect, func(obj *Object) bool { visit(obj); return true })
		}
	}
	if rt.IsZombie == nil {
		rt.IsZombie = func(obj *Object) int {
			if s.IsZombie(obj) {
				return 1
			}
			return 0
		}
	}
	if rt.NewScreenRecord == nil {
		rt.NewScreenRecord = func() *MonsterSpawnScreenRecord50D960 { return new(MonsterSpawnScreenRecord50D960) }
	}
	return rt
}

func monsterSpawnFloat32Chop50D960(value float64) float32 {
	return float32(monsterMoveToRunSpill544440(value))
}

func monsterSpawnLimit50D960(value float64) int32 {
	value = float64(monsterSpawnFloat32Chop50D960(value))
	if math.IsNaN(value) || value >= 2147483648 || value < -2147483648 {
		return math.MinInt32
	}
	return int32(value) // the original 17-byte FISTP helper uses ToZero
}

func monsterSpawnPlayerRect50DE80(player *Player, unit *Object) types.Rectf {
	x, y := float64(unit.PosVec.X), float64(unit.PosVec.Y)
	w, h := float64(player.Field10), float64(player.Field12)
	add := monsterMoveToRunAddChop53_544434
	return types.Rectf{
		Min: types.Ptf(monsterSpawnFloat32Chop50D960(add(add(x, -w), -100)), monsterSpawnFloat32Chop50D960(add(add(y, -h), -100))),
		Max: types.Ptf(monsterSpawnFloat32Chop50D960(add(add(w, x), 100)), monsterSpawnFloat32Chop50D960(add(add(h, y), 100))),
	}
}

func monsterSpawnPointInRect50DE80(point types.Pointf, rect types.Rectf) bool {
	return point.X >= rect.Min.X && point.X <= rect.Max.X && point.Y >= rect.Min.Y && point.Y <= rect.Max.Y
}

func (s *Server) newMonsterSpawnScreen50D960(obj *Object, rt MonsterSpawnDeleteRuntime50E210) *MonsterSpawnScreenRecord50D960 {
	record := rt.NewScreenRecord()
	if record == nil {
		return nil
	}
	pin := new(runtime.Pinner)
	pin.Pin(record)
	if s.monsterSpawns50D780.screenPins == nil {
		s.monsterSpawns50D780.screenPins = make(map[*MonsterSpawnScreenRecord50D960]monsterSpawnScreenOwner50D960)
	}
	s.monsterSpawns50D780.screenPins[record] = monsterSpawnScreenOwner50D960{pin: pin, free: rt.FreeScreenRecord}
	*record = MonsterSpawnScreenRecord50D960{Object: obj, Older: s.monsterSpawns50D780.screenHead}
	if record.Older != nil {
		record.Older.Newer = record
	}
	s.monsterSpawns50D780.screenHead = record
	s.monsterSpawns50D780.screenCount++
	return record
}

func (s *Server) freeMonsterSpawnScreen50D960(record *MonsterSpawnScreenRecord50D960) {
	if record.Older != nil {
		record.Older.Newer = record.Newer
	}
	if record.Newer != nil {
		record.Newer.Older = record.Older
	} else {
		s.monsterSpawns50D780.screenHead = record.Older
	}
	owner := s.monsterSpawns50D780.screenPins[record]
	delete(s.monsterSpawns50D780.screenPins, record)
	record.Object, record.Older, record.Newer = nil, nil, nil
	if owner.pin != nil {
		owner.pin.Unpin()
	}
	if owner.free != nil {
		owner.free(record)
	}
	s.monsterSpawns50D780.screenCount--
}

// MonsterSpawnScratchClear50DE10 releases scratch owners in list order.
func (s *Server) MonsterSpawnScratchClear50DE10() int {
	for s.monsterSpawns50D780.screenHead != nil {
		s.freeMonsterSpawnScreen50D960(s.monsterSpawns50D780.screenHead)
	}
	return 0
}

func (s *Server) MonsterSpawnScratchFind50DE60(obj *Object) *MonsterSpawnScreenRecord50D960 {
	for record := s.monsterSpawns50D780.screenHead; record != nil; record = record.Older {
		if record.Object == obj {
			return record
		}
	}
	return nil
}

// MonsterSpawnCullDistant50D8D0 caches the next link before delayed deletion,
// which can immediately invoke generator cleanup and release the current link.
func (s *Server) MonsterSpawnCullDistant50D8D0(rt MonsterSpawnDeleteRuntime50E210) {
	rt = s.monsterSpawnRuntime50D890(rt)
	for link := s.monsterSpawns50D780.head; link != nil; {
		next, keep := link.Older, false
		for unit := rt.FirstPlayerUnit(); unit != nil; unit = rt.NextPlayerUnit(unit) {
			update := unit.UpdateDataPlayer()
			player := update.Player
			if player != nil && player.Field4792 == 1 && update.QuestExit == nil && rt.Distance(unit, link.Object) < 700 {
				keep = true
			}
		}
		if !keep && rt.DelayedDelete != nil {
			rt.DelayedDelete(link.Object)
		}
		link = next
	}
}

func (s *Server) monsterSpawnVisible50DFB0(obj, player *Object, rt MonsterSpawnDeleteRuntime50E210) {
	if !obj.Class().Has(object.ClassMonster) || obj.Flags().Has(object.FlagDestroyed) {
		return
	}
	if obj.Flags().Has(object.FlagDead) && rt.IsZombie(obj) == 0 {
		return
	}
	if rt.Trace(player.PosVec, obj.PosVec, 69) == 1 {
		s.monsterSpawns50D780.visibleCount++
	}
}

func (s *Server) MonsterSpawnVisible50DFB0(obj, player *Object, rt MonsterSpawnDeleteRuntime50E210) {
	s.monsterSpawnVisible50DFB0(obj, player, s.monsterSpawnRuntime50D890(rt))
}

func (s *Server) MonsterSpawnCanPlace50DE80(point types.Pointf) bool {
	return s.MonsterSpawnCanPlaceWithRuntime50DE80(point, MonsterSpawnDeleteRuntime50E210{})
}

func (s *Server) MonsterSpawnCanPlaceWithRuntime50DE80(point types.Pointf, rt MonsterSpawnDeleteRuntime50E210) bool {
	rt = s.monsterSpawnRuntime50D890(rt)
	maximum := monsterSpawnLimit50D960(rt.MaxOnscreen())
	for unit := rt.FirstPlayerUnit(); unit != nil; unit = rt.NextPlayerUnit(unit) {
		player := unit.UpdateDataPlayer().Player
		if player == nil || player.Field4792 == 0 {
			continue
		}
		rect := monsterSpawnPlayerRect50DE80(player, unit)
		if rt.PointInRect(point, rect) {
			s.monsterSpawns50D780.visibleCount = 0
			rt.EachUnitInRect(rect, func(obj *Object) { s.monsterSpawnVisible50DFB0(obj, unit, rt) })
			if int32(s.monsterSpawns50D780.visibleCount) >= maximum {
				return false
			}
		}
	}
	return true
}

// MonsterSpawnCullVisible50D960 retains the PE's scratch insertion, pairwise
// swaps and player selection. Equal averages must follow that list order;
// collecting a map into a stable sort would choose different monsters.
func (s *Server) MonsterSpawnCullVisible50D960(rt MonsterSpawnDeleteRuntime50E210) int {
	rt = s.monsterSpawnRuntime50D890(rt)
	maximum := monsterSpawnLimit50D960(rt.MaxOnscreen())
	s.monsterSpawns50D780.screenCounts = [32]int32{}
	counts := &s.monsterSpawns50D780.screenCounts
	over := false
	for unit := rt.FirstPlayerUnit(); unit != nil; unit = rt.NextPlayerUnit(unit) {
		player := unit.UpdateDataPlayer().Player
		if player == nil || player.PlayerInd >= 32 {
			continue
		}
		index := player.PlayerInd
		counts[index] = 0
		if player.Field4792 == 0 {
			continue
		}
		rect := monsterSpawnPlayerRect50DE80(player, unit)
		for link := s.monsterSpawns50D780.head; link != nil; link = link.Older {
			spawned := link.Object
			if spawned == nil || !rt.PointInRect(spawned.PosVec, rect) || rt.Trace(unit.PosVec, spawned.PosVec, 69) == 0 {
				continue
			}
			record := s.MonsterSpawnScratchFind50DE60(spawned)
			if record == nil {
				record = s.newMonsterSpawnScreen50D960(spawned, rt)
				if record == nil {
					break
				}
			}
			record.Mask |= uint32(1) << index
			counts[index]++
			if counts[index] > maximum {
				over = true
			}
			record.Distances[index] = monsterSpawnFloat32Chop50D960(rt.Distance(spawned, unit))
		}
	}
	if over {
		var list []*MonsterSpawnScreenRecord50D960
		for record := s.monsterSpawns50D780.screenHead; record != nil; record = record.Older {
			var sum float64
			var n int
			for _, distance := range record.Distances {
				if distance != 0 {
					sum = monsterMoveToRunAddChop53_544434(sum, float64(distance))
					n++
				}
			}
			if n == 0 {
				record.Average = math.Float32frombits(0xffc00000)
			} else {
				record.Average = monsterSpawnFloat32Chop50D960(monsterMoveForceDivChop53_50D581(sum, float64(n)))
			}
			list = append(list, record)
		}
		for j := range list {
			for k := j + 1; k < len(list); k++ {
				if list[k].Average > list[j].Average {
					list[j], list[k] = list[k], list[j]
				}
			}
		}
		for j, record := range list {
			record.Older, record.Newer = nil, nil
			if j+1 < len(list) {
				record.Older = list[j+1]
			}
			if j != 0 {
				record.Newer = list[j-1]
			}
		}
		if len(list) != 0 {
			s.monsterSpawns50D780.screenHead = list[0]
		}
		for {
			var highest int32
			var index uint
			for i, count := range counts {
				if count > highest {
					highest, index = count, uint(i)
				}
			}
			if highest <= maximum || s.monsterSpawns50D780.screenHead == nil {
				break
			}
			for record := s.monsterSpawns50D780.screenHead; record != nil; {
				next := record.Older
				if record.Mask&(uint32(1)<<index) != 0 {
					if rt.DelayedDelete != nil {
						rt.DelayedDelete(record.Object)
					}
					for i := range counts {
						if record.Mask&(uint32(1)<<uint(i)) != 0 {
							counts[i]--
						}
					}
					s.freeMonsterSpawnScreen50D960(record)
					if counts[index] <= maximum {
						break
					}
				}
				record = next
			}
		}
	}
	return s.MonsterSpawnScratchClear50DE10()
}

// MonsterSpawnTick50D890 also preserves the public scalar return on frames
// which skip the screen cleanup. Callers may ignore that original value.
func (s *Server) MonsterSpawnTick50D890(rt MonsterSpawnDeleteRuntime50E210) uint32 {
	frame := s.Frame()
	if rate := s.TickRate(); rate != 0 && frame%(5*rate) == 0 {
		s.MonsterSpawnCullDistant50D8D0(rt)
	}
	if frame%15 == 0 {
		return uint32(s.MonsterSpawnCullVisible50D960(rt))
	}
	return frame / 15
}
