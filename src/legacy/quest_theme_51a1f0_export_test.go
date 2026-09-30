package legacy

import (
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestQuestTheme51A1F0CEntryNativeObjectsAndThreeGroups(t *testing.T) {
	for group := int32(0); group < 3; group++ {
		t.Run(string(rune('0'+group)), func(t *testing.T) {
			srv := server.New(nil, nil, strman.New())
			t.Cleanup(srv.Close)
			unit, freeUnit := alloc.New(server.Object{})
			t.Cleanup(freeUnit)
			removed, freeRemoved := alloc.New(server.Object{})
			t.Cleanup(freeRemoved)
			creature, freeCreature := alloc.New(server.Object{})
			t.Cleanup(freeCreature)
			update, freeUpdate := alloc.New(server.MonsterGenUpdateData{})
			t.Cleanup(freeUpdate)
			empty, freeEmpty := alloc.New(server.MonsterGenUpdateData{})
			t.Cleanup(freeEmpty)
			*unit = server.Object{ObjClass: object.ClassMonsterGenerator, TypeInd: 9, ObjNext: removed, UpdateData: unsafe.Pointer(update)}
			*removed = server.Object{ObjClass: object.ClassMonsterGenerator, TypeInd: 10, UpdateData: unsafe.Pointer(empty)}
			update.Field0[4*group] = creature
			update.QuestSpawnRate = [3]uint8{3, 3, 3}
			srv.Objs.SetObjects(unit)
			t.Cleanup(func() { srv.Objs.SetObjects(nil) })
			if unsafe.Sizeof(uintptr(0)) > 4 {
				for _, p := range []unsafe.Pointer{unsafe.Pointer(unit), unsafe.Pointer(removed), unsafe.Pointer(update), unsafe.Pointer(creature)} {
					if uintptr(p) <= math.MaxUint32 {
						t.Fatalf("pointer=%p, want above 4 GiB", p)
					}
				}
			}
			var deleted []*server.Object
			var minions []int32
			old := questThemeCall51A1F0
			questThemeCall51A1F0 = func(got int32) {
				if got != group {
					t.Fatalf("C group=%d, want %d", got, group)
				}
				srv.QuestTheme51A1F0(got, server.QuestThemeRuntime51A1F0{
					QuestStage:  func() uint32 { return 1 },
					HecubahType: func() uint32 { return 101 }, NecroType: func() uint32 { return 102 },
					GeneratorType: func(got *server.Object) int32 {
						if got != creature {
							t.Fatalf("C creature identity=%p, want %p", got, creature)
						}
						return 0x12345
					},
					DelayedDelete: func(got *server.Object) { deleted = append(deleted, got) },
					SetMinions:    func(value int32) { minions = append(minions, value) },
				})
			}
			defer func() { questThemeCall51A1F0 = old }()
			Sub_51A1F0(int(group))
			if unit.TypeInd != 0x2345 || unit.ObjNext != removed || unit.UpdateData != unsafe.Pointer(update) ||
				update.Field0[4*group] != creature || !reflect.DeepEqual(deleted, []*server.Object{removed}) || !reflect.DeepEqual(minions, []int32{0}) {
				t.Fatalf("native C entry type=%x deleted=%p minions=%v", unit.TypeInd, deleted, minions)
			}
		})
	}
}
