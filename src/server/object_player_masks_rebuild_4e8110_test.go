package server

import (
	"fmt"
	"reflect"
	"testing"
)

type objectPlayerMasksRebuildObject4E8110 struct {
	name             string
	classLow         uint8
	field35, field36 uint32
	next             *objectPlayerMasksRebuildObject4E8110
}

type objectPlayerMasksRebuildPlayer4E8110 struct {
	name string
	unit *objectPlayerMasksRebuildObject4E8110
}

func objectPlayerMasksRebuildTestHooks4E8110(
	events *[]string,
	player **objectPlayerMasksRebuildPlayer4E8110,
	first **objectPlayerMasksRebuildObject4E8110,
	hostile func(*objectPlayerMasksRebuildObject4E8110, *objectPlayerMasksRebuildObject4E8110) int32,
) objectPlayerMasksRebuildHooks4E8110[*objectPlayerMasksRebuildObject4E8110, *objectPlayerMasksRebuildPlayer4E8110] {
	return objectPlayerMasksRebuildHooks4E8110[*objectPlayerMasksRebuildObject4E8110, *objectPlayerMasksRebuildPlayer4E8110]{
		playerByInd: func(ind int32) *objectPlayerMasksRebuildPlayer4E8110 {
			*events = append(*events, fmt.Sprintf("player:%d", ind))
			return *player
		},
		firstObject: func() *objectPlayerMasksRebuildObject4E8110 {
			*events = append(*events, "first")
			return *first
		},
		loadField36: func(obj *objectPlayerMasksRebuildObject4E8110) uint32 {
			*events = append(*events, "load36:"+obj.name)
			return obj.field36
		},
		loadField35: func(obj *objectPlayerMasksRebuildObject4E8110) uint32 {
			*events = append(*events, "load35:"+obj.name)
			return obj.field35
		},
		loadClassLow: func(obj *objectPlayerMasksRebuildObject4E8110) uint8 {
			*events = append(*events, "class:"+obj.name)
			return obj.classLow
		},
		storeField36: func(obj *objectPlayerMasksRebuildObject4E8110, value uint32) {
			*events = append(*events, fmt.Sprintf("store36:%s:%#x", obj.name, value))
			obj.field36 = value
		},
		storeField35: func(obj *objectPlayerMasksRebuildObject4E8110, value uint32) {
			*events = append(*events, fmt.Sprintf("store35:%s:%#x", obj.name, value))
			obj.field35 = value
		},
		loadPlayerUnit: func(player *objectPlayerMasksRebuildPlayer4E8110) *objectPlayerMasksRebuildObject4E8110 {
			*events = append(*events, "unit:"+player.name)
			return player.unit
		},
		isHostile: func(unit, obj *objectPlayerMasksRebuildObject4E8110) int32 {
			*events = append(*events, "hostile:"+unit.name+":"+obj.name)
			return hostile(unit, obj)
		},
		nextObject: func(obj *objectPlayerMasksRebuildObject4E8110) *objectPlayerMasksRebuildObject4E8110 {
			*events = append(*events, "next:"+obj.name)
			return obj.next
		},
	}
}

func TestObjectPlayerMasksRebuild4E8110LookupAndEmptyListOrder(t *testing.T) {
	t.Run("missing player", func(t *testing.T) {
		var player *objectPlayerMasksRebuildPlayer4E8110
		var first *objectPlayerMasksRebuildObject4E8110
		var events []string
		got := objectPlayerMasksRebuild4E8110(-2147483648, objectPlayerMasksRebuildTestHooks4E8110(
			&events, &player, &first, func(*objectPlayerMasksRebuildObject4E8110, *objectPlayerMasksRebuildObject4E8110) int32 {
				t.Fatal("missing player reached hostility callback")
				return 0
			},
		))
		if got != nil {
			t.Fatalf("result = %p, want nil", got)
		}
		if want := []string{"player:-2147483648"}; !reflect.DeepEqual(events, want) {
			t.Fatalf("events = %v, want %v", events, want)
		}
	})

	t.Run("empty object list", func(t *testing.T) {
		playerValue := &objectPlayerMasksRebuildPlayer4E8110{name: "player"}
		player := playerValue
		var first *objectPlayerMasksRebuildObject4E8110
		var events []string
		got := objectPlayerMasksRebuild4E8110(7, objectPlayerMasksRebuildTestHooks4E8110(
			&events, &player, &first, func(*objectPlayerMasksRebuildObject4E8110, *objectPlayerMasksRebuildObject4E8110) int32 {
				t.Fatal("empty list reached hostility callback")
				return 0
			},
		))
		if got != nil {
			t.Fatalf("result = %p, want nil", got)
		}
		if want := []string{"player:7", "first"}; !reflect.DeepEqual(events, want) {
			t.Fatalf("events = %v, want %v", events, want)
		}
	})
}

func TestObjectPlayerMasksRebuild4E8110MasksShiftCount(t *testing.T) {
	tests := []struct {
		playerInd int32
		want      uint32
	}{
		{playerInd: 0, want: 0xfffffffe},
		{playerInd: 31, want: 0x7fffffff},
		{playerInd: 32, want: 0xfffffffe},
		{playerInd: 34, want: 0xfffffffb},
		{playerInd: 63, want: 0x7fffffff},
		{playerInd: -1, want: 0x7fffffff},
		{playerInd: -2147483648, want: 0xfffffffe},
	}
	for _, tc := range tests {
		t.Run(fmt.Sprintf("index_%d", tc.playerInd), func(t *testing.T) {
			playerValue := &objectPlayerMasksRebuildPlayer4E8110{name: "player"}
			player := playerValue
			obj := &objectPlayerMasksRebuildObject4E8110{name: "obj", field35: ^uint32(0), field36: ^uint32(0)}
			first := obj
			var events []string
			got := objectPlayerMasksRebuild4E8110(tc.playerInd, objectPlayerMasksRebuildTestHooks4E8110(
				&events, &player, &first, func(*objectPlayerMasksRebuildObject4E8110, *objectPlayerMasksRebuildObject4E8110) int32 {
					t.Fatal("non-unit object reached hostility callback")
					return 0
				},
			))
			if got != nil {
				t.Fatalf("result = %p, want nil", got)
			}
			if obj.field35 != tc.want || obj.field36 != tc.want {
				t.Fatalf("masks = (%#x, %#x), want both %#x", obj.field35, obj.field36, tc.want)
			}
		})
	}
}

func TestObjectPlayerMasksRebuild4E8110CachesInitialReadsAndUsesLiveSuccessor(t *testing.T) {
	playerValue := &objectPlayerMasksRebuildPlayer4E8110{name: "player"}
	player := playerValue
	stale := &objectPlayerMasksRebuildObject4E8110{name: "stale", field35: 0xffffffff, field36: 0xffffffff}
	replacement := &objectPlayerMasksRebuildObject4E8110{name: "replacement", field35: 0x16, field36: 0x1e}
	obj := &objectPlayerMasksRebuildObject4E8110{name: "obj", field35: 0x0f, field36: 0x1f, next: stale}
	first := obj
	var events []string
	hooks := objectPlayerMasksRebuildTestHooks4E8110(
		&events, &player, &first, func(*objectPlayerMasksRebuildObject4E8110, *objectPlayerMasksRebuildObject4E8110) int32 {
			t.Fatal("cached non-unit class reached hostility callback")
			return 0
		},
	)
	baseStore36 := hooks.storeField36
	hooks.storeField36 = func(current *objectPlayerMasksRebuildObject4E8110, value uint32) {
		baseStore36(current, value)
		if current == obj {
			current.field35 = 0xffffffff
			current.classLow = 0x6
			current.next = replacement
		}
	}

	got := objectPlayerMasksRebuild4E8110(3, hooks)
	if got != nil {
		t.Fatalf("result = %p, want nil", got)
	}
	if obj.field36 != 0x17 || obj.field35 != 0x07 {
		t.Fatalf("first masks = (%#x, %#x), want (0x7, 0x17)", obj.field35, obj.field36)
	}
	if replacement.field36 != 0x16 || replacement.field35 != 0x16 {
		t.Fatalf("replacement masks = (%#x, %#x), want both 0x16", replacement.field35, replacement.field36)
	}
	if stale.field36 != 0xffffffff || stale.field35 != 0xffffffff {
		t.Fatalf("stale successor was visited: masks = (%#x, %#x)", stale.field35, stale.field36)
	}
	want := []string{
		"player:3", "first",
		"load36:obj", "load35:obj", "class:obj", "store36:obj:0x17", "store35:obj:0x7", "next:obj",
		"load36:replacement", "load35:replacement", "class:replacement", "store36:replacement:0x16", "store35:replacement:0x16", "next:replacement",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestObjectPlayerMasksRebuild4E8110ExactOneAndReloadBranches(t *testing.T) {
	tests := []struct {
		name       string
		hostile    int32
		callback36 uint32
		mutate     bool
		want35     uint32
		want36     uint32
		wantSuffix []string
	}{
		{name: "exact one absent", hostile: 1, want35: 0x18, want36: 0x18, wantSuffix: []string{"load36:obj", "store36:obj:0x18", "load35:obj", "store35:obj:0x18", "next:obj"}},
		{name: "exact one callback present", hostile: 1, callback36: 0x18, mutate: true, want35: 0x10, want36: 0x18, wantSuffix: []string{"load36:obj", "next:obj"}},
		{name: "non-one absent", hostile: 2, want35: 0x10, want36: 0x10, wantSuffix: []string{"load36:obj", "next:obj"}},
		{name: "negative absent", hostile: -1, want35: 0x10, want36: 0x10, wantSuffix: []string{"load36:obj", "next:obj"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			unit := &objectPlayerMasksRebuildObject4E8110{name: "unit"}
			playerValue := &objectPlayerMasksRebuildPlayer4E8110{name: "player", unit: unit}
			player := playerValue
			obj := &objectPlayerMasksRebuildObject4E8110{name: "obj", classLow: 0x2, field35: 0x18, field36: 0x18}
			first := obj
			var events []string
			hooks := objectPlayerMasksRebuildTestHooks4E8110(&events, &player, &first,
				func(gotUnit, gotObj *objectPlayerMasksRebuildObject4E8110) int32 {
					if gotUnit != unit || gotObj != obj {
						t.Fatalf("hostility args = (%p, %p), want (%p, %p)", gotUnit, gotObj, unit, obj)
					}
					if tc.mutate {
						gotObj.field36 = tc.callback36
					}
					return tc.hostile
				},
			)
			objectPlayerMasksRebuild4E8110(3, hooks)
			if obj.field35 != tc.want35 || obj.field36 != tc.want36 {
				t.Fatalf("masks = (%#x, %#x), want (%#x, %#x)", obj.field35, obj.field36, tc.want35, tc.want36)
			}
			prefix := []string{
				"player:3", "first", "load36:obj", "load35:obj", "class:obj",
				"store36:obj:0x10", "store35:obj:0x10", "unit:player", "hostile:unit:obj",
			}
			want := append(prefix, tc.wantSuffix...)
			if !reflect.DeepEqual(events, want) {
				t.Fatalf("events = %v, want %v", events, want)
			}
		})
	}
}

func TestObjectPlayerMasksRebuild4E8110NonOneReloadsField36(t *testing.T) {
	unit := &objectPlayerMasksRebuildObject4E8110{name: "unit"}
	playerValue := &objectPlayerMasksRebuildPlayer4E8110{name: "player", unit: unit}
	player := playerValue
	obj := &objectPlayerMasksRebuildObject4E8110{name: "obj", classLow: 0x4, field35: 0x18, field36: 0x18}
	first := obj
	var events []string
	hooks := objectPlayerMasksRebuildTestHooks4E8110(&events, &player, &first,
		func(*objectPlayerMasksRebuildObject4E8110, *objectPlayerMasksRebuildObject4E8110) int32 {
			obj.field36 = 0x18
			return 0
		},
	)
	baseLoad36 := hooks.loadField36
	loads := 0
	hooks.loadField36 = func(current *objectPlayerMasksRebuildObject4E8110) uint32 {
		value := baseLoad36(current)
		loads++
		if loads == 2 {
			current.field36 = 0x1f
		}
		return value
	}

	objectPlayerMasksRebuild4E8110(3, hooks)
	if loads != 3 {
		t.Fatalf("Field36 loads = %d, want 3", loads)
	}
	if obj.field36 != 0x17 || obj.field35 != 0x18 {
		t.Fatalf("masks = (%#x, %#x), want (0x18, 0x17)", obj.field35, obj.field36)
	}
	wantSuffix := []string{
		"hostile:unit:obj", "load36:obj", "load36:obj", "store36:obj:0x17",
		"load35:obj", "store35:obj:0x18", "next:obj",
	}
	if len(events) < len(wantSuffix) || !reflect.DeepEqual(events[len(events)-len(wantSuffix):], wantSuffix) {
		t.Fatalf("event suffix = %v, want %v", events, wantSuffix)
	}
}

func TestObjectPlayerMasksRebuild4E8110LoadsLiveUnitAfterInitialStores(t *testing.T) {
	stale := &objectPlayerMasksRebuildObject4E8110{name: "stale"}
	replacement := &objectPlayerMasksRebuildObject4E8110{name: "replacement"}
	playerValue := &objectPlayerMasksRebuildPlayer4E8110{name: "player", unit: stale}
	player := playerValue
	obj := &objectPlayerMasksRebuildObject4E8110{name: "obj", classLow: 0x6}
	first := obj
	var events []string
	hooks := objectPlayerMasksRebuildTestHooks4E8110(&events, &player, &first,
		func(unit, _ *objectPlayerMasksRebuildObject4E8110) int32 {
			if unit != replacement {
				t.Fatalf("unit = %p, want live replacement %p", unit, replacement)
			}
			return 0
		},
	)
	baseStore35 := hooks.storeField35
	hooks.storeField35 = func(current *objectPlayerMasksRebuildObject4E8110, value uint32) {
		baseStore35(current, value)
		playerValue.unit = replacement
	}

	objectPlayerMasksRebuild4E8110(3, hooks)
	if !containsObjectPlayerMasksRebuildEvent4E8110(events, "unit:player") ||
		!containsObjectPlayerMasksRebuildEvent4E8110(events, "hostile:replacement:obj") ||
		containsObjectPlayerMasksRebuildEvent4E8110(events, "hostile:stale:obj") {
		t.Fatalf("events = %v, want live replacement only", events)
	}
}

func containsObjectPlayerMasksRebuildEvent4E8110(events []string, want string) bool {
	for _, event := range events {
		if event == want {
			return true
		}
	}
	return false
}
