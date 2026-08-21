package server

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestMonsterGeneratorInit4F0590NativeLayout(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantSubClass := uintptr(12)
	wantDirection1 := uintptr(124)
	wantDirection2 := uintptr(126)
	wantObjectUpdate := uintptr(748)
	wantUpdateSize := uintptr(164)
	wantScriptCollision := uintptr(72)
	wantSpawnRate := uintptr(80)
	wantQuestSpawnRate := uintptr(83)
	wantActiveCount := uintptr(86)
	wantMaxActive := uintptr(87)
	wantFrame := uintptr(88)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantSubClass = 16
		wantDirection1 = 128
		wantDirection2 = 130
		wantObjectUpdate = 872
		wantUpdateSize = 216
		wantScriptCollision = 120
		wantSpawnRate = 128
		wantQuestSpawnRate = 131
		wantActiveCount = 134
		wantMaxActive = 135
		wantFrame = 136
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.ObjSubClass", unsafe.Offsetof(Object{}.ObjSubClass), wantSubClass},
		{"Object.Direction1", unsafe.Offsetof(Object{}.Direction1), wantDirection1},
		{"Object.Direction2", unsafe.Offsetof(Object{}.Direction2), wantDirection2},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantObjectUpdate},
		{"MonsterGenUpdateData size", unsafe.Sizeof(MonsterGenUpdateData{}), wantUpdateSize},
		{"MonsterGenUpdateData.ScriptCollision", unsafe.Offsetof(MonsterGenUpdateData{}.ScriptCollision), wantScriptCollision},
		{"MonsterGenUpdateData.SpawnRate", unsafe.Offsetof(MonsterGenUpdateData{}.SpawnRate), wantSpawnRate},
		{"MonsterGenUpdateData.QuestSpawnRate", unsafe.Offsetof(MonsterGenUpdateData{}.QuestSpawnRate), wantQuestSpawnRate},
		{"MonsterGenUpdateData.ActiveCount", unsafe.Offsetof(MonsterGenUpdateData{}.ActiveCount), wantActiveCount},
		{"MonsterGenUpdateData.MaxActive", unsafe.Offsetof(MonsterGenUpdateData{}.MaxActive), wantMaxActive},
		{"MonsterGenUpdateData.Frame88", unsafe.Offsetof(MonsterGenUpdateData{}.Frame88), wantFrame},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestDirectionIndexToAngleNative509E90MatchesSealedTable(t *testing.T) {
	wants := []uint32{160, 192, 224, 128, 0, 0, 96, 64, 32}
	for index, want := range wants {
		if got := directionIndexToAngleNative509E90(uint32(index)); got != want {
			t.Errorf("table[%d] = %d, want %d", index, got, want)
		}
	}
}

func TestDirectionIndexToAngleNative509E90RejectsUnsealedAdjacentData(t *testing.T) {
	for _, index := range []uint32{9, ^uint32(0)} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("index %d did not panic", index)
				}
			}()
			directionIndexToAngleNative509E90(index)
		}()
	}
}

func TestMonsterGeneratorInit4F0590NativeCachesUpdateAndUsesExactFields(t *testing.T) {
	type guardedUpdate struct {
		update MonsterGenUpdateData
		guard  uint32
	}
	entry := &guardedUpdate{guard: 0xa5a5a5a5}
	entry.update.SpawnRate = [3]uint8{0x11, 0x22, 0x33}
	entry.update.QuestSpawnRate = [3]uint8{0, 1, 2}
	entry.update.ActiveCount = 0x44
	entry.update.MaxActive = 0xaa
	entry.update.Frame88 = 0x11223344
	replacement := &guardedUpdate{guard: 0x5a5a5a5a}
	replacement.update.QuestSpawnRate = [3]uint8{3, 3, 3}
	replacement.update.MaxActive = 0x55
	unit := &Object{
		ObjSubClass: object.SubClass(4),
		Direction1:  0xaaaa,
		Direction2:  0xbbbb,
		UpdateData:  unsafe.Pointer(&entry.update),
	}
	groupCalls := 0
	balanceCalls := 0
	got := monsterGeneratorInitNative4F0590(unit, monsterGeneratorInitNativeDeps4F0590{
		currentQuestGroup: func() uint32 {
			groupCalls++
			unit.UpdateData = unsafe.Pointer(&replacement.update)
			return 1
		},
		balanceFloat: func(key string) float64 {
			balanceCalls++
			if key != monsterGeneratorMaxActiveNormalKey4F0590 {
				t.Fatalf("balance key = %q", key)
			}
			// This rounds to exactly 256 at the original FLD-dword boundary.
			// Truncating the unrounded float64 would instead produce 255.
			return 255.999999
		},
	})

	if got != 32 || groupCalls != 1 || balanceCalls != 1 {
		t.Fatalf("return/group/balance calls = %d/%d/%d, want 32/1/1", got, groupCalls, balanceCalls)
	}
	if entry.update.MaxActive != 0 || replacement.update.MaxActive != 0x55 {
		t.Fatalf("cached/replacement MaxActive = %#x/%#x, want 0/0x55", entry.update.MaxActive, replacement.update.MaxActive)
	}
	if unit.UpdateData != unsafe.Pointer(&replacement.update) {
		t.Fatal("current-group mutation of live UpdateData was lost")
	}
	if unit.Direction1 != 32 || unit.Direction2 != 32 {
		t.Fatalf("directions = %d/%d, want 32/32", unit.Direction1, unit.Direction2)
	}
	if entry.update.SpawnRate != [3]uint8{0x11, 0x22, 0x33} || entry.update.QuestSpawnRate != [3]uint8{0, 1, 2} ||
		entry.update.ActiveCount != 0x44 || entry.update.Frame88 != 0x11223344 || entry.guard != 0xa5a5a5a5 ||
		replacement.guard != 0x5a5a5a5a {
		t.Fatalf("adjacent fields changed: entry=%+v replacement=%+v", *entry, *replacement)
	}
}

func TestMonsterGeneratorInit4F0590NativeInvalidSelectorSkipsBalance(t *testing.T) {
	update := &MonsterGenUpdateData{QuestSpawnRate: [3]uint8{4, 4, 4}, MaxActive: 0xaa}
	unit := &Object{
		ObjSubClass: object.SubClass(8), Direction1: 0xaaaa, Direction2: 0xbbbb,
		UpdateData: unsafe.Pointer(update),
	}
	got := monsterGeneratorInitNative4F0590(unit, monsterGeneratorInitNativeDeps4F0590{
		currentQuestGroup: func() uint32 { return 2 },
		balanceFloat: func(string) float64 {
			t.Fatal("selector greater than three reached balance")
			return 0
		},
	})
	if got != 96 || update.MaxActive != 0xaa || unit.Direction1 != 96 || unit.Direction2 != 96 {
		t.Fatalf("return/max/directions = %d/%#x/%d/%d", got, update.MaxActive, unit.Direction1, unit.Direction2)
	}
}

func TestMonsterGeneratorInit4F0590NativeFaultBoundaries(t *testing.T) {
	t.Run("nil-unit-before-group", func(t *testing.T) {
		groupCalls := 0
		defer func() {
			if recover() == nil {
				t.Fatal("nil Object did not preserve UpdateData-load fault")
			}
			if groupCalls != 0 {
				t.Fatalf("group calls = %d, want 0", groupCalls)
			}
		}()
		monsterGeneratorInitNative4F0590(nil, monsterGeneratorInitNativeDeps4F0590{
			currentQuestGroup: func() uint32 { groupCalls++; return 0 },
		})
	})

	t.Run("nil-update-after-group", func(t *testing.T) {
		groupCalls := 0
		balanceCalls := 0
		defer func() {
			if recover() == nil {
				t.Fatal("nil UpdateData did not preserve selector-load fault")
			}
			if groupCalls != 1 || balanceCalls != 0 {
				t.Fatalf("group/balance calls = %d/%d, want 1/0", groupCalls, balanceCalls)
			}
		}()
		monsterGeneratorInitNative4F0590(&Object{}, monsterGeneratorInitNativeDeps4F0590{
			currentQuestGroup: func() uint32 { groupCalls++; return 0 },
			balanceFloat:      func(string) float64 { balanceCalls++; return 0 },
		})
	})

	t.Run("invalid-group-after-callback", func(t *testing.T) {
		update := new(MonsterGenUpdateData)
		groupCalls := 0
		balanceCalls := 0
		defer func() {
			if recover() == nil {
				t.Fatal("out-of-range group did not fault")
			}
			if groupCalls != 1 || balanceCalls != 0 {
				t.Fatalf("group/balance calls = %d/%d, want 1/0", groupCalls, balanceCalls)
			}
		}()
		monsterGeneratorInitNative4F0590(&Object{UpdateData: unsafe.Pointer(update)}, monsterGeneratorInitNativeDeps4F0590{
			currentQuestGroup: func() uint32 { groupCalls++; return 3 },
			balanceFloat:      func(string) float64 { balanceCalls++; return 0 },
		})
	})
}
