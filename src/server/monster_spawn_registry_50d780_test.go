package server

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func monsterSpawnTestGenerator50D780(active uint8) (*Object, *MonsterGenUpdateData) {
	data := &MonsterGenUpdateData{ActiveCount: active}
	return &Object{ObjClass: object.ClassMonsterGenerator, UpdateData: unsafe.Pointer(data)}, data
}

func monsterSpawnTestMonster50D780(subclass object.SubClass) (*Object, *MonsterUpdateData) {
	data := new(MonsterUpdateData)
	return &Object{ObjClass: object.ClassMonster, ObjSubClass: subclass, UpdateData: unsafe.Pointer(data)}, data
}

func TestMonsterSpawnRegistry50D780NativeLinksAndCleanup(t *testing.T) {
	s := new(Server)
	s.MonsterSpawnInit50D780()
	generator, generatorData := monsterSpawnTestGenerator50D780(0)
	first, firstData := monsterSpawnTestMonster50D780(0)
	second, secondData := monsterSpawnTestMonster50D780(0)
	third, thirdData := monsterSpawnTestMonster50D780(0)

	for _, monster := range []*Object{first, second, third} {
		if !s.MonsterSpawnRegister50E030(generator, monster, MonsterSpawnRegisterRuntime50E030{}) {
			t.Fatalf("register %p failed", monster)
		}
	}
	if generatorData.ActiveCount != 3 {
		t.Fatalf("active count = %d, want 3", generatorData.ActiveCount)
	}
	if s.monsterSpawns50D780.head != thirdData.Field549 ||
		thirdData.Field549.Older != secondData.Field549 ||
		secondData.Field549.Older != firstData.Field549 ||
		firstData.Field549.Newer != secondData.Field549 ||
		secondData.Field549.Newer != thirdData.Field549 {
		t.Fatal("native-width spawn list does not preserve head/older/newer links")
	}

	middleLink := secondData.Field549
	s.MonsterSpawnCleanup50E140(second)
	if generatorData.ActiveCount != 2 || secondData.Field548 != nil || secondData.Field549 != nil {
		t.Fatalf("middle cleanup count/backrefs = %d/%p/%p", generatorData.ActiveCount, secondData.Field548, secondData.Field549)
	}
	if thirdData.Field549.Older != firstData.Field549 || firstData.Field549.Newer != thirdData.Field549 {
		t.Fatal("middle cleanup did not reconnect adjacent native links")
	}
	if middleLink.Object != nil || middleLink.Older != nil || middleLink.Newer != nil {
		t.Fatalf("middle link remains live: %#v", middleLink)
	}

	s.MonsterSpawnReset50D7E0()
	if generatorData.ActiveCount != 0 || firstData.Field549 != nil || thirdData.Field549 != nil || s.monsterSpawns50D780.head != nil {
		t.Fatalf("reset count/links/head = %d/%p/%p/%p", generatorData.ActiveCount, firstData.Field549, thirdData.Field549, s.monsterSpawns50D780.head)
	}
}

func TestMonsterSpawnRegister50E030RequiresSessionAllocator(t *testing.T) {
	s := new(Server)
	generator, _ := monsterSpawnTestGenerator50D780(0)
	monster, update := monsterSpawnTestMonster50D780(0)
	if s.MonsterSpawnRegister50E030(generator, monster, MonsterSpawnRegisterRuntime50E030{}) {
		t.Fatal("registration succeeded before MonsterSpawnInit50D780")
	}
	if update.Field548 != nil || update.Field549 != nil {
		t.Fatalf("failed registration wrote back-references %p/%p", update.Field548, update.Field549)
	}
}

func TestMonsterSpawnRegister50E030BuildsBomberGlyph(t *testing.T) {
	s := new(Server)
	s.MonsterSpawnInit50D780()
	generator, _ := monsterSpawnTestGenerator50D780(0)
	spawned, update := monsterSpawnTestMonster50D780(object.SubClass(object.MonsterBomber))
	spawned.PosVec = types.Ptf(123.5, 456.25)
	update.Field511 = 7
	update.Field512 = 0
	update.Field513 = 11
	glyphData := new(GlyphInitData)
	glyph := &Object{InitData: unsafe.Pointer(glyphData)}
	var inventoryOwner, inventoryItem *Object
	var report int32

	if !s.MonsterSpawnRegister50E030(generator, spawned, MonsterSpawnRegisterRuntime50E030{
		CreateGlyph: func() *Object { return glyph },
		InventoryPut: func(owner, item *Object, value int32) {
			inventoryOwner, inventoryItem, report = owner, item, value
		},
	}) {
		t.Fatal("register failed")
	}
	if glyphData.Spells[:3][0] != 7 || glyphData.Spells[:3][1] != 0 || glyphData.Spells[:3][2] != 11 {
		t.Fatalf("glyph spells = %v", glyphData.Spells[:3])
	}
	if glyphData.SpellsCnt != 2 || glyphData.SpellArg.Obj != nil || glyphData.SpellArg.Pos != spawned.PosVec {
		t.Fatalf("glyph count/arg = %d/%#v", glyphData.SpellsCnt, glyphData.SpellArg)
	}
	if inventoryOwner != spawned || inventoryItem != glyph || report != 1 {
		t.Fatalf("inventory call = %p/%p/%d", inventoryOwner, inventoryItem, report)
	}
}

func TestMonsterSpawnTick50D890CachesNextAcrossImmediateCleanup(t *testing.T) {
	s := new(Server)
	s.SetTickRate(30)
	s.SetFrame(150)
	s.MonsterSpawnInit50D780()
	generator, _ := monsterSpawnTestGenerator50D780(0)
	first, firstData := monsterSpawnTestMonster50D780(0)
	second, secondData := monsterSpawnTestMonster50D780(0)
	s.MonsterSpawnRegister50E030(generator, first, MonsterSpawnRegisterRuntime50E030{})
	s.MonsterSpawnRegister50E030(generator, second, MonsterSpawnRegisterRuntime50E030{})
	var deleted []*Object
	s.MonsterSpawnTick50D890(MonsterSpawnDeleteRuntime50E210{DelayedDelete: func(obj *Object) {
		deleted = append(deleted, obj)
		s.MonsterSpawnCleanup50E140(obj)
	}})
	if len(deleted) != 2 || deleted[0] != second || deleted[1] != first {
		t.Fatalf("deleted order = %p, want second then first", deleted)
	}
	if firstData.Field549 != nil || secondData.Field549 != nil || s.monsterSpawns50D780.head != nil {
		t.Fatal("immediate cleanup left a tracked spawn")
	}
}
