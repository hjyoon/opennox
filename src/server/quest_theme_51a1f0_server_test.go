package server

import (
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/balance"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/prand"
	"github.com/opennox/libs/types"
)

func TestQuestTheme51A1F0NativeGroupLayoutAndBinary32Balance(t *testing.T) {
	for group := int32(0); group < 3; group++ {
		t.Run(string(rune('0'+group)), func(t *testing.T) {
			srv := new(Server)
			srv.Balance.file = &balance.File{Global: balance.Config{
				"questhardcorestage":              balance.Float(10),
				"generatormaxactivecreatureshigh": balance.Float(255.999999),
			}, Tags: make(map[balance.Tag]balance.Config)}
			selected, replacementCreature, neighbor := new(Object), new(Object), new(Object)
			update := &MonsterGenUpdateData{
				SpawnRate: [3]uint8{11, 22, 33}, QuestSpawnRate: [3]uint8{0, 0, 0},
				ActiveCount: 44, MaxActive: 170, Frame88: 0x12345678,
			}
			update.Field0[4*group], update.Field0[4*group+1] = selected, neighbor
			otherUpdate := &MonsterGenUpdateData{MaxActive: 171}
			unit := &Object{ObjClass: object.ClassMonsterGenerator, TypeInd: 9, UpdateData: unsafe.Pointer(update)}
			srv.Objs.SetObjects(unit)
			if unsafe.Sizeof(uintptr(0)) > 4 {
				for _, p := range []unsafe.Pointer{unsafe.Pointer(unit), unsafe.Pointer(update), unsafe.Pointer(selected), unsafe.Pointer(replacementCreature)} {
					if uintptr(p) <= math.MaxUint32 {
						t.Fatalf("native pointer %p is not above 4 GiB", p)
					}
				}
			}
			calls, mapped := 0, 0
			var minions []int32
			srv.QuestTheme51A1F0(group, QuestThemeRuntime51A1F0{
				QuestStage: func() uint32 {
					calls++
					if calls == 1 {
						return 1
					}
					unit.UpdateData = unsafe.Pointer(otherUpdate)
					update.Field0[4*group] = replacementCreature
					return 10
				},
				HecubahType: func() uint32 { return 101 }, NecroType: func() uint32 { return 102 },
				GeneratorType: func(creature *Object) int32 {
					mapped++
					if creature != replacementCreature {
						t.Fatalf("mapped creature=%p, want %p", creature, replacementCreature)
					}
					return 0x1ABCD
				},
				SetMinions: func(value int32) { minions = append(minions, value) },
			})
			// FLD-dword rounds 255.999999 to 256, then truncation keeps low
			// byte zero. Native update/creature pointers survive callback changes.
			if update.MaxActive != 0 || otherUpdate.MaxActive != 171 || unit.TypeInd != 0xABCD || calls != 2 || mapped != 1 ||
				!reflect.DeepEqual(minions, []int32{0}) || update.Field0[4*group+1] != neighbor ||
				update.SpawnRate != [3]uint8{11, 22, 33} || update.ActiveCount != 44 || update.Frame88 != 0x12345678 {
				t.Fatalf("native theme changed wrong fields: update=%+v other=%+v type=%x calls=%d/%d minions=%v", update, otherUpdate, unit.TypeInd, calls, mapped, minions)
			}
		})
	}
}

func TestQuestTheme51A1F0NativeSpawnReceivesLivePositionAddress(t *testing.T) {
	srv := new(Server)
	srv.Rand.Logic = prand.New(0)
	srv.Balance.file = &balance.File{Global: balance.Config{"minionsalwaysstage": balance.Float(10)}, Tags: make(map[balance.Tag]balance.Config)}
	unit := &Object{TypeInd: 101, PosVec: types.Pointf{X: 23, Y: 37}}
	srv.Objs.SetObjects(unit)
	var spawned *types.Pointf
	srv.QuestTheme51A1F0(0, QuestThemeRuntime51A1F0{
		QuestStage:  func() uint32 { return 5 },
		HecubahType: func() uint32 { return 101 }, NecroType: func() uint32 { return 102 },
		SetMinions:   func(int32) {},
		SpawnHecubah: func(position *types.Pointf) { spawned = position; position.X = 41 },
		DelayedDelete: func(got *Object) {
			if got != unit {
				t.Fatalf("delete identity=%p", got)
			}
		},
	})
	if spawned != &unit.PosVec || unit.PosVec != (types.Pointf{X: 41, Y: 37}) {
		t.Fatalf("spawn position=%p, want %p, position=%v", spawned, &unit.PosVec, unit.PosVec)
	}
}

func TestQuestTheme51A1F0NativeNullUpdateAndInvalidGroupFault(t *testing.T) {
	for _, group := range []int32{-1, 0, 3} {
		t.Run(string(rune('1'+group)), func(t *testing.T) {
			srv := new(Server)
			unit := &Object{ObjClass: object.ClassMonsterGenerator}
			if group != 0 {
				unit.UpdateData = unsafe.Pointer(new(MonsterGenUpdateData))
			}
			srv.Objs.SetObjects(unit)
			defer func() {
				if recover() == nil {
					t.Fatal("invalid native update/group did not fault")
				}
			}()
			srv.QuestTheme51A1F0(group, QuestThemeRuntime51A1F0{
				QuestStage:  func() uint32 { return 1 },
				HecubahType: func() uint32 { return 101 }, NecroType: func() uint32 { return 102 },
			})
		})
	}
}
