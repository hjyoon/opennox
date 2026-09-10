package legacy

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/server"
)

func TestCharmQuestSpawnCleanup5013E0UsesNativeSpawnLink(t *testing.T) {
	generatorData := &server.MonsterGenUpdateData{}
	generator := &server.Object{
		ObjClass:   object.ClassMonsterGenerator,
		UpdateData: unsafe.Pointer(generatorData),
	}
	monsterData := &server.MonsterUpdateData{}
	monster := &server.Object{
		ObjClass:   object.ClassMonster,
		UpdateData: unsafe.Pointer(monsterData),
	}
	s := new(server.Server)
	if !s.MonsterSpawnInit50D780() {
		t.Fatal("MonsterSpawnInit50D780 failed")
	}
	if !s.MonsterSpawnRegister50E030(generator, monster, server.MonsterSpawnRegisterRuntime50E030{}) {
		t.Fatal("MonsterSpawnRegister50E030 failed")
	}
	link := monsterData.Field549
	if link == nil {
		t.Fatal("registration did not create a native link")
	}
	if link.Object != monster || generatorData.ActiveCount != 1 {
		t.Fatalf("registration = link %p object %p count %d", link, link.Object, generatorData.ActiveCount)
	}

	charmQuestSpawnCleanupNative5013E0(s, monster)
	if monsterData.Field548 != nil || monsterData.Field549 != nil {
		t.Fatalf("spawn back-references = %p/%p, want nil/nil", monsterData.Field548, monsterData.Field549)
	}
	if generatorData.ActiveCount != 0 {
		t.Fatalf("active count = %d, want 0", generatorData.ActiveCount)
	}
	if link.Object != nil || link.Older != nil || link.Newer != nil {
		t.Fatalf("detached link = %#v", link)
	}
}
