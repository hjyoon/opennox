package server

import (
	"fmt"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestMonsterActionGetUp534A90PopsOnlyAfterAnimationCompletes(t *testing.T) {
	for _, done := range []uint8{0, 1, 255} {
		t.Run(fmt.Sprintf("terminal_%d", done), func(t *testing.T) {
			unit := monsterActionTestObject50A910(t)
			unit.UpdateDataMonster().Field120_3 = done
			pops := 0
			if !monsterActionGetUp534A90(unit, func() int {
				pops++
				return 0
			}) {
				t.Fatal("get-up action was not handled")
			}
			want := 0
			if done != 0 {
				want = 1
			}
			if pops != want {
				t.Fatalf("terminal byte/pop count = %d/%d, want %d/%d", done, pops, done, want)
			}
		})
	}
}

func TestMonsterActionGetUp534A90RejectsInvalidMetadataBeforePop(t *testing.T) {
	for _, unit := range []*Object{
		nil,
		{},
		{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(new(MonsterUpdateData))},
	} {
		if monsterActionGetUp534A90(unit, func() int {
			t.Fatal("invalid metadata popped action")
			return 0
		}) {
			t.Fatalf("invalid unit %#v was handled", unit)
		}
	}
}

func TestMonsterActionGetUp534A90PreservesNativeObjectPointer(t *testing.T) {
	unit := monsterActionTestObject50A910(t)
	got := uintptr(unsafe.Pointer(unit))
	if unsafe.Sizeof(uintptr(0)) == 8 && got <= uintptr(^uint32(0)) {
		t.Fatalf("unit pointer = %#x, want value above PE32 range", got)
	}
	if !monsterActionGetUp534A90(unit, func() int { return 0 }) {
		t.Fatal("native-width unit was not handled")
	}
}
