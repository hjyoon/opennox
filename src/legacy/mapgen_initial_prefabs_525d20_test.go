package legacy

import (
	"reflect"
	"strconv"
	"testing"
)

func TestMapgenInitialPrefabs525D20NativePointerOrder(t *testing.T) {
	got := mapgenInitialPrefabsFixture525D20(4, 0b1010, 0)
	if !got.result {
		t.Fatal("initial prefab placement failed")
	}
	if got.calls != 4 {
		t.Fatalf("placement calls = %d, want 4", got.calls)
	}
	if want := []int{1, 3, 0, 2}; !reflect.DeepEqual(got.order[:4], want) {
		t.Fatalf("placement order = %v, want %v", got.order[:4], want)
	}
	if got.placedMask != 0b1111 {
		t.Fatalf("placed mask = %#x, want %#x", got.placedMask, uint32(0b1111))
	}
	if strconv.IntSize == 64 {
		if got.themeAddress <= 1<<32 {
			t.Fatalf("theme fixture address = %#x, want address above PE32 range", got.themeAddress)
		}
		for i, address := range got.prefabAddresses[:4] {
			if address <= 1<<32 {
				t.Fatalf("prefab %d fixture address = %#x, want address above PE32 range", i, address)
			}
		}
	}
}

func TestMapgenInitialPrefabs525D20LimitsAndFailure(t *testing.T) {
	t.Run("optional sixth prefab is left for the next stage", func(t *testing.T) {
		got := mapgenInitialPrefabsFixture525D20(6, 0, 0)
		if !got.result || got.calls != 5 || got.placedMask != 0b011111 {
			t.Fatalf("result=%v calls=%d placed=%#x", got.result, got.calls, got.placedMask)
		}
	})

	t.Run("six required prefabs fail before placing the sixth", func(t *testing.T) {
		got := mapgenInitialPrefabsFixture525D20(6, 0b111111, 0)
		if got.result || got.calls != 5 || got.placedMask != 0b011111 {
			t.Fatalf("result=%v calls=%d placed=%#x", got.result, got.calls, got.placedMask)
		}
	})

	t.Run("placement callback failure is propagated", func(t *testing.T) {
		got := mapgenInitialPrefabsFixture525D20(4, 0b0010, 2)
		if got.result || got.calls != 2 || got.placedMask != 0b0010 {
			t.Fatalf("result=%v calls=%d placed=%#x order=%v", got.result, got.calls, got.placedMask, got.order[:2])
		}
	})
}
