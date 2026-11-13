package legacy

import (
	"bytes"
	"fmt"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// Complement the independent root comparison with production PopAction/Reset
// composition. End handlers are inactive here; this is not full End proof.
func TestMonsterPopAttacksProduction5471B0(t *testing.T) {
	for _, path := range []string{"C", "Go"} {
		t.Run(path, func(t *testing.T) {
			for index := -1; index < 24; index++ {
				for _, action := range []uint32{16, 17, 18, 19, 20, 25, 26, 27, 28} {
					t.Run(fmt.Sprintf("%d/%d", index, action), func(t *testing.T) {
						s := server.NewHigh(nil, nil, strman.New())
						t.Cleanup(s.Close)
						if !s.Objs.InitHigh(1) {
							t.Fatal("native object allocator")
						}
						t.Cleanup(s.Objs.FreeObjects)
						unit := s.Objs.NewObject(&server.ObjectType{})
						update, free := alloc.NewHigh(server.MonsterUpdateData{})
						t.Cleanup(free)
						unit.ObjClass, unit.UpdateData = object.ClassMonster, unsafe.Pointer(update)
						s.SetFrame(100)
						update.AIStackInd = int8(index)
						update.Field2, update.Field67, update.Field74, update.Field91 = 2, 67, 74, 91
						update.Field120_1, update.Field120_2, update.Field120_3 = 1, 2, 3
						update.Field124, update.Field137 = 124, 137
						if index == -1 {
							update.Field132 = action
						}
						for i := range update.AIStack {
							a := action
							if i == 0 {
								a = uint32(ai.ACTION_GUARD)
							} else if i < index && i%2 != 0 {
								a = uint32(ai.DEPENDENCY_TIME)
							}
							update.AIStack[i] = server.AIStackItem{Action: a, Args: [4]uintptr{1, 2, 3, 4}}
						}
						wantUnit, wantUpdate := *unit, *update
						changed := index != 0
						if changed {
							wantUpdate.AIStackInd = 0
							if index == -1 {
								wantUpdate.AIStack[0].Action = 0
							}
							wantUpdate.Field2, wantUpdate.Field67, wantUpdate.Field74, wantUpdate.Field91 = 0, 0, 0, 0
							wantUpdate.Field120_1, wantUpdate.Field120_2, wantUpdate.Field120_3 = 0, 0, 0
							wantUpdate.Field124, wantUpdate.Field137 = 100, 100
						}
						if monsterPopAttacksOverride5471B0 != nil {
							t.Fatal("production Pop must remain bound")
						}
						if path == "C" {
							if popOriginalABI5471B0() != 64 {
								t.Fatal("typed pointer public ABI required")
							}
							popOriginalC5471B0(unit)
						} else {
							unit.MonsterPopAttackActionsNative5471B0(func(obj *server.Object) int { return obj.MonsterPopAction() })
						}
						if s.AI.StackChanged != changed {
							t.Fatal("production stack-change marker")
						}
						for _, pair := range []struct {
							got, want unsafe.Pointer
							size      uintptr
						}{
							{unsafe.Pointer(unit), unsafe.Pointer(&wantUnit), unsafe.Sizeof(wantUnit)},
							{unsafe.Pointer(update), unsafe.Pointer(&wantUpdate), unsafe.Sizeof(wantUpdate)},
						} {
							if !bytes.Equal(unsafe.Slice((*byte)(pair.got), int(pair.size)), unsafe.Slice((*byte)(pair.want), int(pair.size))) {
								t.Fatal("whole production records/padding differ")
							}
						}
					})
				}
			}
		})
	}
}
