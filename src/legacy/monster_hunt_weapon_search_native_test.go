package legacy

import (
	"fmt"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/player"
	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/server"
)

func TestMonsterHuntNativeCABI5449D0(t *testing.T) {
	s := playerVitalReportWorld(t)
	if !s.Objs.InitHigh(8) {
		t.Fatal("high-address world objects")
	}
	t.Cleanup(s.Objs.FreeObjects)
	s.FreeObjectTypes()
	t.Cleanup(s.FreeObjectTypes)
	const name = "NativeHuntEntryFixture"
	if err := s.Types.ReadObjectType(&things.Thing{Name: name}); err != nil {
		t.Fatal(err)
	}
	for route := range 2 {
		t.Run(fmt.Sprintf("route%d", route), func(t *testing.T) {
			unit := s.NewObjectByTypeID(name)
			if unit == nil {
				t.Fatal("fixture monster allocation")
			}
			defer func() {
				unit.UpdateData = nil
				s.Objs.FreeObject(unit)
			}()
			data := game4ConsumerValue(t, server.MonsterUpdateData{})
			unit.ObjClass, unit.UpdateData = object.ClassMonster, unsafe.Pointer(data)
			data.AIStack[0].Action = uint32(ai.ACTION_HUNT)
			data.AIStack[1].Args = [4]uintptr{0x4321, 0x937e, ^uintptr(0), 0xa17b}
			got := callMonsterHuntNative5449D0(route, unit)
			if got != &data.AIStack[1] || got.Type() != ai.ACTION_ROAM || data.AIStackInd != 1 {
				t.Fatalf("hunt returned %p and stack index %d", got, data.AIStackInd)
			}
			if cgoABIPointerSize == 8 && uintptr(unsafe.Pointer(got)) <= 0xffffffff {
				t.Fatal("hunt action address must exceed 4 GiB")
			}
			wantArgs := [4]uintptr{0, 0x937e, ^uintptr(0xff) | 0x80, 0xa17b}
			if got.Args != wantArgs || math.Float32bits(data.Aggression) != 0x3f547ae1 {
				t.Fatalf("hunt changed preserved arguments or aggression: %v %#x", got.Args, math.Float32bits(data.Aggression))
			}
			data.AIStackInd = int8(len(data.AIStack) - 1)
			data.Aggression = 0
			if got := callMonsterHuntNative5449D0(route, unit); got != nil || math.Float32bits(data.Aggression) != 0x3f547ae1 {
				t.Fatal("full stack must still receive the hunt aggression")
			}
		})
	}
}

func TestMonsterWeaponSearchNativeCABI544AE0(t *testing.T) {
	s := playerVitalReportWorld(t)
	s.Map.Init()
	t.Cleanup(s.Map.Free)
	if s.Walls.InitHigh() == 0 {
		t.Fatal("fixture walls")
	}
	t.Cleanup(s.Walls.Free)
	oldClass, oldState := Nox_xxx_playerClassCanUseItem_57B3D0, monsterWeaponSearchState544AE0
	t.Cleanup(func() {
		Nox_xxx_playerClassCanUseItem_57B3D0 = oldClass
		monsterWeaponSearchState544AE0 = oldState
	})
	actor := game4ConsumerValue(t, server.Object{ObjClass: object.ClassMonster})
	item := game4ConsumerValue(t, server.Object{ObjClass: object.ClassWeapon, ObjFlags: object.FlagActive})
	actor.PosVec, actor.NewPos = types.Ptf(1615, 2398), types.Ptf(1615, 2398)
	item.PosVec, item.NewPos = types.Ptf(1613.2247, 2399.6528), types.Ptf(1613.2247, 2399.6528)
	item.Shape.Kind, item.Shape.Circle.R, item.Shape.Circle.R2 = server.ShapeKindCircle, 3, 9
	actor.Shape.Kind, actor.Shape.Circle.R, actor.Shape.Circle.R2 = server.ShapeKindCircle, 8, 64
	if !s.CanInteract(actor, item, 0) {
		t.Fatal("fixture should place the weapon in direct view")
	}
	s.Map.AddObjectToIndex(item)
	defer s.Map.RemoveObjectFromIndex(item)
	for route := range 2 {
		t.Run(fmt.Sprintf("route%d", route), func(t *testing.T) {
			var calls int
			available := true
			Nox_xxx_playerClassCanUseItem_57B3D0 = func(candidate *server.Object, cl player.Class) bool {
				calls++
				if candidate != item || cl != player.Warrior {
					t.Fatalf("eligibility argument %p/%v, want %p/Warrior", candidate, cl, item)
				}
				return available
			}
			if got := callMonsterWeaponSearchNative544AE0(route, actor, 75); got != item ||
				monsterWeaponSearchState544AE0.BestObject != item || calls != 1 {
				t.Fatalf("search winner %p, scratch %p, %d eligibility calls", got, monsterWeaponSearchState544AE0.BestObject, calls)
			}
			winningDistance := math.Float32bits(monsterWeaponSearchState544AE0.BestDistance)
			if winningDistance != 0x40bc4513 {
				t.Fatalf("stored distance %#08x, want the stock x87 spill %#08x", winningDistance, uint32(0x40bc4513))
			}
			available = false
			if got := callMonsterWeaponSearchNative544AE0(route, actor, 75); got != nil ||
				monsterWeaponSearchState544AE0.BestObject != nil || monsterWeaponSearchState544AE0.BestDistance != 10000000 {
				t.Fatal("rejected search retained the prior winner or distance")
			}
			available = true
			callMonsterWeaponCandidateNative544B20(route, item, actor)
			if got := monsterWeaponSearchState544AE0; got.BestObject != item || math.Float32bits(got.BestDistance) != winningDistance || calls != 3 {
				t.Fatalf("direct callback did not match the search: %+v, %d calls", got, calls)
			}
		})
	}
}
