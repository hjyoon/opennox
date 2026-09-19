package legacy

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/player"

	"github.com/opennox/opennox/v1/server"
)

func TestQuestPlayerIdentityStore509C30PreservesOriginalPredicates(t *testing.T) {
	var store questPlayerIdentityStore509C30
	store.clear()
	if store.initialized {
		t.Fatal("cleanup initialized an untouched store")
	}

	if !store.allows("Alice", player.Wizard, 7) {
		t.Fatal("empty store rejected an identity")
	}
	if store.contains("Alice") {
		t.Fatal("empty store reported a name")
	}

	store.remember("Alice", player.Wizard, 7)
	store.remember("Bob", player.Conjurer, math.MaxUint32)

	for _, tc := range []struct {
		name  string
		class player.Class
		id    uint32
		want  bool
	}{
		{name: "Alice", class: player.Wizard, id: 7, want: true},
		{name: "Alice", class: player.Conjurer, id: 7, want: false},
		{name: "Alice", class: player.Wizard, id: 8, want: false},
		{name: "Alice", class: player.Conjurer, id: 8, want: false},
		{name: "Bob", class: player.Conjurer, id: math.MaxUint32, want: true},
		{name: "Carol", class: player.Warrior, id: 0, want: true},
	} {
		if got := store.allows(tc.name, tc.class, tc.id); got != tc.want {
			t.Fatalf("allows(%q, %v, %d) = %t, want %t", tc.name, tc.class, tc.id, got, tc.want)
		}
	}

	if !store.contains("Alice") || !store.contains("Bob") || store.contains("Carol") {
		t.Fatal("name membership does not match the remembered records")
	}
	store.clear()
	if !store.initialized {
		t.Fatal("cleanup reset the original one-time initialization state")
	}
	if len(store.entries) != 0 || store.contains("Alice") || !store.allows("Alice", player.Conjurer, 8) {
		t.Fatal("cleanup did not leave an initialized empty store")
	}
}

func TestQuestPlayerIdentityStore509C30PreservesPrependOrderAndDuplicates(t *testing.T) {
	var store questPlayerIdentityStore509C30
	store.remember("first", player.Warrior, 1)
	store.remember("second", player.Wizard, 2)
	store.remember("first", player.Conjurer, 3)

	store.Lock()
	got := append([]questPlayerIdentity509C30(nil), store.entries...)
	store.Unlock()
	want := []questPlayerIdentity509C30{
		{name: "first", class: player.Conjurer, id: 3},
		{name: "second", class: player.Wizard, id: 2},
		{name: "first", class: player.Warrior, id: 1},
	}
	if len(got) != len(want) {
		t.Fatalf("entry count = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("entry %d = %+v, want %+v", i, got[i], want[i])
		}
	}
	if store.allows("first", player.Warrior, 1) {
		t.Fatal("a duplicate name with a different identity was not rejected")
	}
}

func TestQuestPlayerIdentityWrappers509C30CopyNativePlayerFields(t *testing.T) {
	questPlayerIdentities509C30.Lock()
	oldInitialized := questPlayerIdentities509C30.initialized
	oldEntries := append([]questPlayerIdentity509C30(nil), questPlayerIdentities509C30.entries...)
	questPlayerIdentities509C30.initialized = false
	questPlayerIdentities509C30.entries = nil
	questPlayerIdentities509C30.Unlock()
	t.Cleanup(func() {
		questPlayerIdentities509C30.Lock()
		questPlayerIdentities509C30.initialized = oldInitialized
		questPlayerIdentities509C30.entries = oldEntries
		questPlayerIdentities509C30.Unlock()
	})

	pl := new(server.Player)
	pl.SetField2096("GuideOwner")
	pl.Field2068 = math.MaxUint32
	pl.Info().SetPlayerClass(player.Wizard)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(pl)) <= math.MaxUint32 {
		t.Fatalf("test Player address %#x does not exercise a native high pointer", uintptr(unsafe.Pointer(pl)))
	}

	Sub_509C30(pl)
	pl.SetField2096("mutated")
	pl.Field2068 = 1
	pl.Info().SetPlayerClass(player.Warrior)

	if got := Sub_509CF0("GuideOwner", player.Wizard, math.MaxUint32); got != 1 {
		t.Fatalf("same identity result = %d, want 1", got)
	}
	if got := Sub_509CF0("GuideOwner", player.Warrior, math.MaxUint32); got != 0 {
		t.Fatalf("conflicting identity result = %d, want 0", got)
	}
	probe := new(server.Player)
	probe.SetField2096("GuideOwner")
	if got := Sub_509D80(probe); got != 1 {
		t.Fatalf("remembered name result = %d, want 1", got)
	}
	Sub_509CB0()
	if got := Sub_509D80(probe); got != 0 {
		t.Fatalf("cleared name result = %d, want 0", got)
	}
}
