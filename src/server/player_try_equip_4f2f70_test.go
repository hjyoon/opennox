package server

import (
	"fmt"
	"reflect"
	"testing"
)

func TestPlayerTryEquip4F2F70WeaponShortCircuitsAndNormalizes(t *testing.T) {
	owner := new(Object)
	item := new(Object)
	for _, result := range []int32{1, -1, 0x7fffffff, -0x80000000} {
		t.Run(fmt.Sprintf("result_%d", result), func(t *testing.T) {
			calls := 0
			got := PlayerTryEquip4F2F70(
				owner,
				item,
				func(gotOwner, gotItem *Object, flag1, flag2 int32) int32 {
					calls++
					if gotOwner != owner || gotItem != item {
						t.Fatalf("weapon pointers = (%p, %p), want (%p, %p)", gotOwner, gotItem, owner, item)
					}
					if flag1 != 1 || flag2 != 1 {
						t.Fatalf("weapon flags = (%d, %d), want (1, 1)", flag1, flag2)
					}
					return result
				},
				func(*Object, *Object, int32, int32) int32 {
					t.Fatal("armor callback ran after nonzero weapon result")
					return 0
				},
			)
			if got != 1 {
				t.Fatalf("result = %d, want normalized 1", got)
			}
			if calls != 1 {
				t.Fatalf("weapon calls = %d, want 1", calls)
			}
		})
	}
}

func TestPlayerTryEquip4F2F70ArmorFallbackOrderAndNormalization(t *testing.T) {
	owner := new(Object)
	item := new(Object)
	for _, tc := range []struct {
		name   string
		armor  int32
		result int32
	}{
		{name: "zero", armor: 0, result: 0},
		{name: "positive", armor: 37, result: 1},
		{name: "negative", armor: -37, result: 1},
		{name: "minimum", armor: -0x80000000, result: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var events []string
			got := PlayerTryEquip4F2F70(
				owner,
				item,
				func(gotOwner, gotItem *Object, flag1, flag2 int32) int32 {
					events = append(events, "weapon")
					if gotOwner != owner || gotItem != item || flag1 != 1 || flag2 != 1 {
						t.Fatalf("weapon arguments = (%p, %p, %d, %d)", gotOwner, gotItem, flag1, flag2)
					}
					return 0
				},
				func(gotOwner, gotItem *Object, flag1, flag2 int32) int32 {
					events = append(events, "armor")
					if gotOwner != owner || gotItem != item || flag1 != 1 || flag2 != 1 {
						t.Fatalf("armor arguments = (%p, %p, %d, %d)", gotOwner, gotItem, flag1, flag2)
					}
					return tc.armor
				},
			)
			if got != tc.result {
				t.Fatalf("result = %d, want %d", got, tc.result)
			}
			if want := []string{"weapon", "armor"}; !reflect.DeepEqual(events, want) {
				t.Fatalf("callback order = %v, want %v", events, want)
			}
		})
	}
}

func TestPlayerTryEquip4F2F70ForwardsNilPointersWithoutGuard(t *testing.T) {
	var events []string
	got := PlayerTryEquip4F2F70(
		nil,
		nil,
		func(owner, item *Object, flag1, flag2 int32) int32 {
			events = append(events, "weapon")
			if owner != nil || item != nil || flag1 != 1 || flag2 != 1 {
				t.Fatalf("weapon arguments = (%p, %p, %d, %d), want nil, nil, 1, 1", owner, item, flag1, flag2)
			}
			return 0
		},
		func(owner, item *Object, flag1, flag2 int32) int32 {
			events = append(events, "armor")
			if owner != nil || item != nil || flag1 != 1 || flag2 != 1 {
				t.Fatalf("armor arguments = (%p, %p, %d, %d), want nil, nil, 1, 1", owner, item, flag1, flag2)
			}
			return 0
		},
	)
	if got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	if want := []string{"weapon", "armor"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("callback order = %v, want %v", events, want)
	}
}
