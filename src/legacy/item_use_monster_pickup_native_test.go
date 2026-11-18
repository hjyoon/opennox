package legacy

import (
	_ "embed"
	"fmt"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/server"
)

//go:embed testdata/item_use_pickup_original_v1.json.gz
var itemUsePickupStock []byte

type itemUsePickupCase struct {
	Input struct {
		Name           string
		Path           string
		Item           *string
		Callback       *string
		Owner          *string
		Gate           uint32
		CallbackResult uint32 `json:"callback_result"`
		Replacement    *string
		Origin         [2]uint32
		TargetPos      [2]uint32 `json:"target_pos"`
		StackSlot      int       `json:"stack_slot"`
	}
	Observed struct {
		ReturnWord uint32 `json:"return_word"`
		Events     []struct{ Child string }
	}
}

type itemUsePickupLegacyServer struct {
	Server
	world *server.Server
}

func (s *itemUsePickupLegacyServer) S() *server.Server { return s.world }

func itemUsePickupWorld(t *testing.T) *server.Server {
	t.Helper()
	world := server.NewHigh(nil, nil, strman.New())
	t.Cleanup(world.Close)
	bridge := &itemUsePickupLegacyServer{world: world}
	old := GetServer
	GetServer = func() Server { return bridge }
	t.Cleanup(func() { GetServer = old })
	return world
}

func itemUsePickupOriginalRows(t *testing.T) []itemUsePickupCase {
	t.Helper()
	return tradeOfferOriginalRows[itemUsePickupCase](t, itemUsePickupStock,
		"290a685360bef99a1363d49013aae5406a772d2eed2aa7b35670ed37dceec264", 216)
}

func TestNativeItemUseCABI53F8E0(t *testing.T) {
	rows, world := itemUsePickupOriginalRows(t), itemUsePickupWorld(t)
	for route := range 2 {
		for _, row := range rows {
			in := row.Input
			if in.Path != "use" || in.Gate > 1 || (in.Owner == nil && in.Gate == 1) || in.Replacement != nil {
				continue
			}
			t.Run(fmt.Sprintf("route%d-%s", route, in.Name), func(t *testing.T) {
				var owner, item *server.Object
				if in.Owner != nil {
					if in.Item != nil && in.Callback != nil {
						player, info := game4ConsumerPlayer(t, 0)
						info.PlayerInd = 19
						owner = player
						world.Players.SetXxx(owner, int32(in.Gate))
					} else {
						// A missing item or callback must return before reading player data.
						owner = game4ConsumerValue(t, server.Object{ObjClass: object.ClassPlayer})
					}
				}
				if in.Item != nil {
					item = game4ConsumerValue(t, server.Object{NetCode: 0xfedcad12})
					if in.Callback != nil {
						item.Use.Ptr = itemUseFixtureCallback53F8E0()
					}
				}
				wantCalls := 0
				for _, event := range row.Observed.Events {
					if event.Child == "callback" {
						wantCalls++
					}
				}
				got, calls, gotOwner, gotItem := callItemUseNative53F8E0(route, owner, item, in.CallbackResult)
				if got != row.Observed.ReturnWord || calls != wantCalls {
					t.Fatalf("entry returned %#08x and called %d times, want %#08x and %d",
						got, calls, row.Observed.ReturnWord, wantCalls)
				}
				if calls != 0 && (gotOwner != owner || gotItem != item) {
					t.Fatalf("callback received %p/%p, want %p/%p", gotOwner, gotItem, owner, item)
				}
			})
		}
	}
}

func TestNativeMonsterPickupCABI544B90(t *testing.T) {
	rows, world := itemUsePickupOriginalRows(t), itemUsePickupWorld(t)
	if !world.Objs.InitHigh(8) {
		t.Fatal("fixture object pool")
	}
	t.Cleanup(world.Objs.FreeObjects)
	world.FreeObjectTypes()
	t.Cleanup(world.FreeObjectTypes)
	const fixtureName = "NativePickupActorFixture"
	if err := world.Types.ReadObjectType(&things.Thing{Name: fixtureName}); err != nil {
		t.Fatal(err)
	}
	for route := range 2 {
		for _, row := range rows {
			if row.Input.Path != "pickup" || len(row.Observed.Events) != 1 || row.Observed.Events[0].Child != "pop" {
				continue
			}
			in := row.Input
			t.Run(fmt.Sprintf("route%d-%s", route, in.Name), func(t *testing.T) {
				actor := world.NewObjectByTypeID(fixtureName)
				if actor == nil || (unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(actor)) <= 0xffffffff) {
					t.Fatal("fixture did not create a high-address monster")
				}
				defer func() {
					actor.UpdateData = nil
					world.Objs.FreeObject(actor)
				}()
				data := game4ConsumerValue(t, server.MonsterUpdateData{})
				actor.ObjClass = object.ClassMonster
				actor.UpdateData = unsafe.Pointer(data)
				actor.PosVec = types.Ptf(math.Float32frombits(in.Origin[0]), math.Float32frombits(in.Origin[1]))
				data.AIStackInd = int8(in.StackSlot)
				head := data.AIStackHead()
				head.Action = uint32(ai.ACTION_PICKUP_OBJECT)
				if in.Item != nil {
					item := game4ConsumerValue(t, server.Object{
						ObjClass: object.ClassFood, ObjSubClass: 0x10,
						PosVec: types.Ptf(math.Float32frombits(in.TargetPos[0]), math.Float32frombits(in.TargetPos[1])),
					})
					head.SetArgs(item)
				}
				wantPop := int8(in.StackSlot - 1)
				if got := callMonsterPickupNative544B90(route, actor); got != wantPop {
					t.Fatalf("entry pop result = %d, want %d", got, wantPop)
				}
				if data.AIStackInd != max(wantPop, 0) || !world.AI.StackChanged {
					t.Fatal("the native pickup action was not popped")
				}
			})
		}
	}
}
