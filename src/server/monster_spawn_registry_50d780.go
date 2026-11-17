package server

import (
	"runtime"

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
	head         *MonsterSpawnLink
	initialized  bool
	pins         map[*MonsterSpawnLink]*runtime.Pinner
	screenHead   *MonsterSpawnScreenRecord50D960
	screenCount  uint32
	screenPins   map[*MonsterSpawnScreenRecord50D960]monsterSpawnScreenOwner50D960
	screenCounts [32]int32
	visibleCount uint32
	glyphType    int
}

// Native MonsterUpdateData records retain Go links through Field549. Keep each
// link pinned until its native backreference and neighboring links are cleared.
func (s *Server) pinMonsterSpawnLink50E030(link *MonsterSpawnLink) {
	if s.monsterSpawns50D780.pins[link] != nil {
		return
	}
	if s.monsterSpawns50D780.pins == nil {
		s.monsterSpawns50D780.pins = make(map[*MonsterSpawnLink]*runtime.Pinner)
	}
	pin := new(runtime.Pinner)
	pin.Pin(link)
	s.monsterSpawns50D780.pins[link] = pin
}

func (s *Server) unpinMonsterSpawnLink50E140(link *MonsterSpawnLink) {
	if pin := s.monsterSpawns50D780.pins[link]; pin != nil {
		pin.Unpin()
		delete(s.monsterSpawns50D780.pins, link)
	}
}

// MonsterSpawnDeleteRuntime50E210 supplies external object, visibility and
// scratch-storage operations for the original spawn component. Missing query
// services use the server's native players, map, distance and balance owners.
type MonsterSpawnDeleteRuntime50E210 struct {
	DelayedDelete    func(*Object)
	GlyphID          func() int
	FirstPlayerUnit  func() *Object
	NextPlayerUnit   func(*Object) *Object
	MaxOnscreen      func() float64
	PointInRect      func(types.Pointf, types.Rectf) bool
	Trace            func(types.Pointf, types.Pointf, MapTraceFlags) int
	Distance         func(*Object, *Object) float64
	EachUnitInRect   func(types.Rectf, func(*Object))
	IsZombie         func(*Object) int
	NewScreenRecord  func() *MonsterSpawnScreenRecord50D960
	FreeScreenRecord func(*MonsterSpawnScreenRecord50D960)
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
	// Raw unlink leaves the detached node owned. Include pinned nodes that no
	// longer occur in the active list before releasing any foreign reference.
	owned := make(map[*MonsterSpawnLink]struct{}, len(s.monsterSpawns50D780.pins))
	for link := range s.monsterSpawns50D780.pins {
		owned[link] = struct{}{}
	}
	for link := s.monsterSpawns50D780.head; link != nil; link = link.Older {
		owned[link] = struct{}{}
	}
	for link := range owned {
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
	}
	s.monsterSpawns50D780.head = nil
	// A remaining older link can still refer to an already cleared newer link.
	// Release pins only after all native backreferences and list edges are gone.
	for link := range s.monsterSpawns50D780.pins {
		s.unpinMonsterSpawnLink50E140(link)
	}
	s.MonsterSpawnScratchClear50DE10()
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
	s.MonsterSpawnUnlink50E1B0(link)
	link.Older = nil
	link.Newer = nil
}

// MonsterSpawnPrepend50E110 retains a node until cleanup or reset, including
// the interval after a raw unlink which preserves the node's own edges.
func (s *Server) MonsterSpawnPrepend50E110(link *MonsterSpawnLink) *MonsterSpawnLink {
	s.pinMonsterSpawnLink50E030(link)
	link.Newer = nil
	link.Older = s.monsterSpawns50D780.head
	if link.Older != nil {
		link.Older.Newer = link
	}
	s.monsterSpawns50D780.head = link
	return link
}

// MonsterSpawnUnlink50E1B0 preserves the original pointer return and detached
// edges. Unlink transfers no storage ownership and releases no pin.
func (s *Server) MonsterSpawnUnlink50E1B0(link *MonsterSpawnLink) *MonsterSpawnLink {
	result := link
	if link.Older != nil {
		link.Older.Newer = link.Newer
	}
	if link.Newer != nil {
		result = link.Older
		link.Newer.Older = link.Older
	} else {
		s.monsterSpawns50D780.head = link.Older
	}
	return result
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
	link := s.MonsterSpawnPrepend50E110(&MonsterSpawnLink{Object: spawned})
	generatorUpdate.ActiveCount++
	spawnedUpdate.Field548 = generator
	spawnedUpdate.Field549 = link

	if spawned.SubClass().AsMonster().Has(object.MonsterBomber) {
		var glyph *Object
		if runtime.CreateGlyph != nil {
			glyph = runtime.CreateGlyph()
		}
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
		}
		if runtime.InventoryPut != nil {
			runtime.InventoryPut(spawned, glyph, 1)
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
		// Detached nodes can retain stale edges after the public raw unlink.
		for other := range s.monsterSpawns50D780.pins {
			if other.Older == link {
				other.Older = nil
			}
			if other.Newer == link {
				other.Newer = nil
			}
		}
		s.unpinMonsterSpawnLink50E140(link)
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
	var glyphID int
	if runtime.GlyphID == nil {
		// Reuse the Types-owned cache so a type-table reload also invalidates
		// this lookup. The original lookup still precedes the class filter.
		glyphID = s.Types.GlyphID()
	} else {
		if s.monsterSpawns50D780.glyphType == 0 {
			s.monsterSpawns50D780.glyphType = runtime.GlyphID()
		}
		glyphID = s.monsterSpawns50D780.glyphType
	}
	if obj.Class().Has(object.ClassMonster) && obj.SubClass().AsMonster().Has(object.MonsterBomber) {
		update := (*MonsterUpdateData)(obj.UpdateData)
		if update.Field549 != nil && runtime.DelayedDelete != nil {
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
