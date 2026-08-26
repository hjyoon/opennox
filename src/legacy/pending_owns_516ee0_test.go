package legacy

import (
	"reflect"
	"testing"

	"github.com/opennox/opennox/v1/server"
)

func TestPendingOwnStore516EE0Lifecycle(t *testing.T) {
	var store pendingOwnStore516EE0
	if store.add(1, 2) {
		t.Fatal("add succeeded before allocator initialization")
	}
	if !store.alloc() || !store.add(1, 2) || !store.add(3, 4) {
		t.Fatal("initialized store rejected a pending ownership record")
	}
	if got := store.take(); !reflect.DeepEqual(got, []pendingOwn516F90{{1, 2}, {3, 4}}) {
		t.Fatalf("take = %#v", got)
	}
	if got := store.take(); len(got) != 0 {
		t.Fatalf("second take returned %d records", len(got))
	}
	if !store.add(5, 6) {
		t.Fatal("take unexpectedly deactivated the store")
	}
	store.clear()
	if got := store.take(); len(got) != 0 {
		t.Fatalf("clear left %d records", len(got))
	}
	if got := store.free(); got != 0 {
		t.Fatalf("free = %d, want 0", got)
	}
	if store.add(7, 8) {
		t.Fatal("add succeeded after allocator teardown")
	}
}

func TestResolvePendingOwns516FC0PreservesLIFOAndSkipsMissing(t *testing.T) {
	objects := map[int32]*server.Object{
		1: {ScriptIDVal: 1},
		2: {ScriptIDVal: 2},
		3: {ScriptIDVal: 3},
		4: {ScriptIDVal: 4},
	}
	entries := []pendingOwn516F90{
		{ownerScriptID: 1, ownedScriptID: 2},
		{ownerScriptID: 99, ownedScriptID: 2},
		{ownerScriptID: 3, ownedScriptID: 4},
		{ownerScriptID: 1, ownedScriptID: 98},
	}
	var got [][2]int32
	resolvePendingOwns516FC0(entries, func(id int32) *server.Object {
		return objects[id]
	}, func(owner, owned *server.Object) {
		got = append(got, [2]int32{owner.ScriptIDVal, owned.ScriptIDVal})
	})
	want := [][2]int32{{3, 4}, {1, 2}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("resolved ownership = %v, want %v", got, want)
	}
}
