package server

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

func TestMonsterActionHunt5449D0StoresAggressionBeforeRoamPush(t *testing.T) {
	unit := monsterActionTestObject50A910(t)
	item := &AIStackItem{Args: [4]uintptr{1, 2, ^uintptr(0), 4}}
	calls := 0
	if !monsterActionHunt5449D0(unit, func(got *Object, action ai.ActionType) *AIStackItem {
		calls++
		if got != unit || action != ai.ACTION_ROAM {
			t.Fatalf("push = (%p, %s), want (%p, %s)", got, action, unit, ai.ACTION_ROAM)
		}
		if bits := math.Float32bits(got.UpdateDataMonster().Aggression); bits != monsterHuntAggressionBits5449D0 {
			t.Fatalf("aggression at push = %#08x, want %#08x", bits, monsterHuntAggressionBits5449D0)
		}
		return item
	}) {
		t.Fatal("hunt action was not handled")
	}
	if calls != 1 {
		t.Fatalf("push calls = %d, want 1", calls)
	}
	wantMask := ^uintptr(0)&^uintptr(0xff) | uintptr(0x80)
	if item.Args != [4]uintptr{0, 2, wantMask, 4} {
		t.Fatalf("roam args = %#v, want waypoint zero and mask low byte %#x", item.Args, wantMask)
	}
}

func TestMonsterActionHunt5449D0PreservesAggressionWhenPushFails(t *testing.T) {
	unit := monsterActionTestObject50A910(t)
	if !monsterActionHunt5449D0(unit, func(*Object, ai.ActionType) *AIStackItem { return nil }) {
		t.Fatal("hunt action was not handled")
	}
	if bits := math.Float32bits(unit.UpdateDataMonster().Aggression); bits != monsterHuntAggressionBits5449D0 {
		t.Fatalf("aggression after failed push = %#08x, want %#08x", bits, monsterHuntAggressionBits5449D0)
	}
}

func TestMonsterActionHunt5449D0RejectsInvalidMetadataBeforePush(t *testing.T) {
	for _, unit := range []*Object{
		nil,
		{},
		{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(new(MonsterUpdateData))},
	} {
		if monsterActionHunt5449D0(unit, func(*Object, ai.ActionType) *AIStackItem {
			t.Fatal("invalid metadata reached push")
			return nil
		}) {
			t.Fatalf("invalid unit %#v was handled", unit)
		}
	}
	unit := monsterActionTestObject50A910(t)
	if monsterActionHunt5449D0(unit, nil) {
		t.Fatal("nil push hook was handled")
	}
}

func TestMonsterActionHunt5449D0PreservesNativeObjectPointer(t *testing.T) {
	unit := monsterActionTestObject50A910(t)
	want := uintptr(unsafe.Pointer(unit))
	if unsafe.Sizeof(uintptr(0)) == 8 && want <= uintptr(^uint32(0)) {
		t.Fatalf("unit pointer = %#x, want value above PE32 range", want)
	}
	if !monsterActionHunt5449D0(unit, func(got *Object, action ai.ActionType) *AIStackItem {
		if uintptr(unsafe.Pointer(got)) != want || action != ai.ACTION_ROAM {
			t.Fatalf("push = (%#x, %s), want (%#x, %s)", uintptr(unsafe.Pointer(got)), action, want, ai.ACTION_ROAM)
		}
		return nil
	}) {
		t.Fatal("native-width unit was not handled")
	}
}
