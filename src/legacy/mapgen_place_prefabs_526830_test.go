package legacy

import (
	"strconv"
	"testing"
)

func TestMapgenPlacePrefabs526830PreservesNativePointers(t *testing.T) {
	got := mapgenPlacePrefabsFixture526830()
	if !got.result || got.calls != 2 {
		t.Fatalf("finalization result=%v calls=%d, want two placed prefabs", got.result, got.calls)
	}
	for i := range got.callbackThemeAddresses {
		if got.callbackThemeAddresses[i] != got.themeAddress {
			t.Fatalf("callback %d theme=%#x, want %#x", i, got.callbackThemeAddresses[i], got.themeAddress)
		}
	}
	wantPrefab := [2]uintptr{got.prefabAddresses[0], got.prefabAddresses[2]}
	wantForeach := [2]uintptr{got.foreachAddresses[0], got.foreachAddresses[1]}
	wantRoom := [2]uintptr{got.roomAddresses[0], got.roomAddresses[2]}
	for i := range wantPrefab {
		if got.callbackPrefabAddresses[i] != wantPrefab[i] ||
			got.callbackForeachAddresses[i] != wantForeach[i] ||
			got.callbackRoomAddresses[i] != wantRoom[i] {
			t.Fatalf("callback %d = prefab %#x foreach %#x room %#x, want %#x %#x %#x",
				i, got.callbackPrefabAddresses[i], got.callbackForeachAddresses[i], got.callbackRoomAddresses[i],
				wantPrefab[i], wantForeach[i], wantRoom[i])
		}
	}
	if got.foundChoiceAddress != got.choiceAddresses[1] {
		t.Fatalf("FOREACH second-node choices=%#x, want %#x", got.foundChoiceAddress, got.choiceAddresses[1])
	}
	if got.nextChoiceAddress != got.choiceAddresses[1] {
		t.Fatalf("choice next=%#x, want %#x", got.nextChoiceAddress, got.choiceAddresses[1])
	}
	if strconv.IntSize == 64 {
		addresses := map[string]uintptr{
			"theme": got.themeAddress,
		}
		for i, address := range got.prefabAddresses {
			addresses["prefab "+strconv.Itoa(i)] = address
			addresses["room "+strconv.Itoa(i)] = got.roomAddresses[i]
		}
		for i, address := range got.foreachAddresses {
			addresses["foreach "+strconv.Itoa(i)] = address
			addresses["choice "+strconv.Itoa(i)] = got.choiceAddresses[i]
		}
		for name, address := range addresses {
			if address <= 1<<32 {
				t.Fatalf("%s fixture address=%#x, want address above PE32 range", name, address)
			}
		}
	}
}
