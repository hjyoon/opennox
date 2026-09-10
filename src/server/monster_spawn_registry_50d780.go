package server

import (
	"math"
	"sort"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

// MonsterSpawnLink is the native-width replacement for GAME.EXE SpawnClass.
// Older points toward the previous list head; Newer points back toward the
// current head. MonsterUpdateData.Field549 owns the link while it is active.
type MonsterSpawnLink struct {
	Object *Object
	Older  *MonsterSpawnLink
	Newer  *MonsterSpawnLink
}

type monsterSpawnState50D780 struct {
	head        *MonsterSpawnLink
	initialized bool
}

// MonsterSpawnDeleteRuntime50E210 supplies the outer-server delayed-delete
// operation used for tracked monsters and their private Glyph inventory item.
type MonsterSpawnDeleteRuntime50E210 struct {
	DelayedDelete func(*Object)
}

// MonsterSpawnRegisterRuntime50E030 supplies the two operations that still
// cross the root server boundary when a generator creates a Bomber-like unit.
type MonsterSpawnRegisterRuntime50E030 struct {
	CreateGlyph  func() *Object
	InventoryPut func(owner, item *Object, report int32)
}

// MonsterSpawnInit50D780 replaces both fixed-capacity PE32 allocation classes.
// Go links and temporary slices grow naturally and retain full native pointers.
func (s *Server) MonsterSpawnInit50D780() bool {
	s.MonsterSpawnReset50D7E0()
	s.monsterSpawns50D780.initialized = true
	return true
}

// MonsterSpawnReset50D7E0 drops the session list and clears live back-references.
func (s *Server) MonsterSpawnReset50D7E0() {
	for link := s.monsterSpawns50D780.head; link != nil; {
		next := link.Older
		if obj := link.Object; obj != nil && obj.UpdateData != nil && obj.Class().Has(object.ClassMonster) {
			update := (*MonsterUpdateData)(obj.UpdateData)
			if update.Field549 == link {
				if generator := update.Field548; generator != nil && generator.UpdateData != nil {
					generatorUpdate := (*MonsterGenUpdateData)(generator.UpdateData)
					generatorUpdate.ActiveCount--
				}
				update.Field548 = nil
				update.Field549 = nil
			}
		}
		link.Object = nil
		link.Older = nil
		link.Newer = nil
		link = next
	}
	s.monsterSpawns50D780.head = nil
}

// MonsterSpawnFree50D820 releases the logical session allocator.
func (s *Server) MonsterSpawnFree50D820() {
	s.MonsterSpawnReset50D7E0()
	s.monsterSpawns50D780.initialized = false
}

func (s *Server) unlinkMonsterSpawn50E1B0(link *MonsterSpawnLink) {
	if link == nil {
		return
	}
	if link.Older != nil {
		link.Older.Newer = link.Newer
	}
	if link.Newer != nil {
		link.Newer.Older = link.Older
	} else if s.monsterSpawns50D780.head == link {
		s.monsterSpawns50D780.head = link.Older
	}
	link.Older = nil
	link.Newer = nil
}

// MonsterSpawnRegister50E030 records a generator-created monster exactly once.
func (s *Server) MonsterSpawnRegister50E030(generator, spawned *Object, runtime MonsterSpawnRegisterRuntime50E030) bool {
	if !s.monsterSpawns50D780.initialized || generator == nil || spawned == nil {
		return false
	}
	generatorUpdate := (*MonsterGenUpdateData)(generator.UpdateData)
	spawnedUpdate := (*MonsterUpdateData)(spawned.UpdateData)
	if spawnedUpdate.Field549 != nil {
		return true
	}
	link := &MonsterSpawnLink{Object: spawned, Older: s.monsterSpawns50D780.head}
	if link.Older != nil {
		link.Older.Newer = link
	}
	s.monsterSpawns50D780.head = link
	generatorUpdate.ActiveCount++
	spawnedUpdate.Field548 = generator
	spawnedUpdate.Field549 = link

	if spawned.SubClass().AsMonster().Has(object.MonsterBomber) && runtime.CreateGlyph != nil {
		glyph := runtime.CreateGlyph()
		if glyph != nil {
			data := glyph.InitDataGlyph()
			spells := [...]uint32{spawnedUpdate.Field511, spawnedUpdate.Field512, spawnedUpdate.Field513}
			copy(data.Spells[:3], spells[:])
			data.SpellsCnt = 0
			for i := 0; i < 3; i++ {
				if data.Spells[i] != 0 {
					data.SpellsCnt++
				}
			}
			data.SpellArg = SpellAcceptArg{Pos: spawned.PosVec}
			if runtime.InventoryPut != nil {
				runtime.InventoryPut(spawned, glyph, 1)
			}
		}
	}
	return true
}

// MonsterSpawnCleanup50E140 removes generator ownership and the native link.
func (s *Server) MonsterSpawnCleanup50E140(obj *Object) {
	if obj == nil || obj.UpdateData == nil {
		return
	}
	update := (*MonsterUpdateData)(obj.UpdateData)
	if generator := update.Field548; generator != nil {
		if generator.UpdateData != nil {
			(*MonsterGenUpdateData)(generator.UpdateData).ActiveCount--
		}
		update.Field548 = nil
	}
	if link := update.Field549; link != nil {
		s.unlinkMonsterSpawn50E1B0(link)
		link.Object = nil
		update.Field549 = nil
	}
}

// MonsterSpawnCleanupUnlessZombie50E1E0 preserves the original zombie exemption.
func (s *Server) MonsterSpawnCleanupUnlessZombie50E1E0(obj *Object) {
	if obj != nil && !s.IsZombie(obj) {
		s.MonsterSpawnCleanup50E140(obj)
	}
}

// MonsterSpawnDelete50E210 deletes a tracked Bomber Glyph before unlinking.
func (s *Server) MonsterSpawnDelete50E210(obj *Object, runtime MonsterSpawnDeleteRuntime50E210) {
	if obj == nil {
		return
	}
	if obj.Class().Has(object.ClassMonster) && obj.SubClass().AsMonster().Has(object.MonsterBomber) {
		update := (*MonsterUpdateData)(obj.UpdateData)
		if update.Field549 != nil && runtime.DelayedDelete != nil {
			glyphID := s.Types.GlyphID()
			for item := obj.FirstItem(); item != nil; {
				next := item.NextItem()
				if int(item.TypeInd) == glyphID {
					runtime.DelayedDelete(item)
				}
				item = next
			}
		}
	}
	s.MonsterSpawnCleanup50E140(obj)
}

func monsterSpawnPlayerRect50DE80(player *Player, unit *Object) types.Rectf {
	halfWidth := float32(player.Field10) + 100
	halfHeight := float32(player.Field12) + 100
	return types.Rectf{
		Min: types.Ptf(unit.PosVec.X-halfWidth, unit.PosVec.Y-halfHeight),
		Max: types.Ptf(unit.PosVec.X+halfWidth, unit.PosVec.Y+halfHeight),
	}
}

func monsterSpawnPointInRect50DE80(point types.Pointf, rect types.Rectf) bool {
	return point.X >= rect.Min.X && point.X <= rect.Max.X && point.Y >= rect.Min.Y && point.Y <= rect.Max.Y
}

func (s *Server) monsterSpawnVisibleCount50DFB0(player *Object, rect types.Rectf) int {
	count := 0
	s.Map.EachObjInRect(rect, func(obj *Object) bool {
		if !obj.Class().Has(object.ClassMonster) || obj.Flags().Has(object.FlagDestroyed) {
			return true
		}
		if obj.Flags().Has(object.FlagDead) && !s.IsZombie(obj) {
			return true
		}
		if s.MapTraceRay(player.PosVec, obj.PosVec, MapTraceFlags(69)) {
			count++
		}
		return true
	})
	return count
}

// MonsterSpawnCanPlace50DE80 enforces the per-player on-screen spawn cap.
func (s *Server) MonsterSpawnCanPlace50DE80(point types.Pointf) bool {
	maximum := int(math.RoundToEven(s.Balance.Float("MaxOnscreenMonsterCount")))
	for unit := s.Players.FirstUnit(); unit != nil; unit = s.Players.NextUnit(unit) {
		player := unit.UpdateDataPlayer().Player
		if player == nil || player.Field4792 == 0 {
			continue
		}
		rect := monsterSpawnPlayerRect50DE80(player, unit)
		if monsterSpawnPointInRect50DE80(point, rect) && s.monsterSpawnVisibleCount50DFB0(unit, rect) >= maximum {
			return false
		}
	}
	return true
}

type monsterSpawnScreenRecord50D960 struct {
	object    *Object
	mask      uint32
	distances [32]float32
	average   float32
}

// MonsterSpawnTick50D890 performs the two Quest cleanup cadences without the
// PE32 MonsterListClass scratch allocator.
func (s *Server) MonsterSpawnTick50D890(runtime MonsterSpawnDeleteRuntime50E210) {
	frame := s.Frame()
	if rate := s.TickRate(); rate != 0 && frame%(5*rate) == 0 {
		for link := s.monsterSpawns50D780.head; link != nil; {
			next := link.Older
			keep := false
			for unit := s.Players.FirstUnit(); unit != nil; unit = s.Players.NextUnit(unit) {
				player := unit.UpdateDataPlayer().Player
				if player != nil && player.Field4792 == 1 && unit.UpdateDataPlayer().QuestExit == nil &&
					unit.PosVec.Sub(link.Object.PosVec).Len() < 700 {
					keep = true
				}
			}
			if !keep && runtime.DelayedDelete != nil {
				runtime.DelayedDelete(link.Object)
			}
			link = next
		}
	}
	if frame%15 != 0 {
		return
	}

	maximum := int(math.RoundToEven(s.Balance.Float("MaxOnscreenMonsterCount")))
	counts := [32]int{}
	records := make(map[*Object]*monsterSpawnScreenRecord50D960)
	for unit := s.Players.FirstUnit(); unit != nil; unit = s.Players.NextUnit(unit) {
		player := unit.UpdateDataPlayer().Player
		if player == nil || player.Field4792 == 0 || player.PlayerInd >= 32 {
			continue
		}
		index := int(player.PlayerInd)
		rect := monsterSpawnPlayerRect50DE80(player, unit)
		for link := s.monsterSpawns50D780.head; link != nil; link = link.Older {
			spawned := link.Object
			if spawned == nil || !monsterSpawnPointInRect50DE80(spawned.PosVec, rect) ||
				!s.MapTraceRay(unit.PosVec, spawned.PosVec, MapTraceFlags(69)) {
				continue
			}
			record := records[spawned]
			if record == nil {
				record = &monsterSpawnScreenRecord50D960{object: spawned}
				records[spawned] = record
			}
			record.mask |= uint32(1) << uint(index)
			record.distances[index] = float32(unit.PosVec.Sub(spawned.PosVec).Len())
			counts[index]++
		}
	}
	list := make([]*monsterSpawnScreenRecord50D960, 0, len(records))
	for _, record := range records {
		var sum float32
		var n int
		for _, distance := range record.distances {
			if distance != 0 {
				sum += distance
				n++
			}
		}
		if n != 0 {
			record.average = sum / float32(n)
		}
		list = append(list, record)
	}
	sort.SliceStable(list, func(i, j int) bool { return list[i].average > list[j].average })
	for {
		playerIndex := 0
		for i := 1; i < len(counts); i++ {
			if counts[i] > counts[playerIndex] {
				playerIndex = i
			}
		}
		if counts[playerIndex] <= maximum {
			return
		}
		removed := false
		for i, record := range list {
			if record.mask&(uint32(1)<<uint(playerIndex)) == 0 {
				continue
			}
			if runtime.DelayedDelete != nil {
				runtime.DelayedDelete(record.object)
			}
			for p := range counts {
				if record.mask&(uint32(1)<<uint(p)) != 0 {
					counts[p]--
				}
			}
			list = append(list[:i], list[i+1:]...)
			removed = true
			break
		}
		if !removed {
			return
		}
	}
}
