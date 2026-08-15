package server

import (
	"fmt"
	"reflect"
	"testing"
)

type netCodeCacheLookupTestObject4ECD90 struct {
	name string
	code uint32
}

type netCodeCacheLookupTestEntry4ECD90 struct {
	name   string
	object *netCodeCacheLookupTestObject4ECD90
	next   *netCodeCacheLookupTestEntry4ECD90
}

func TestNetCodeCacheLookup4ECD90LazyInitBeforeHead(t *testing.T) {
	for _, needsInit := range []bool{false, true} {
		t.Run(fmt.Sprintf("needs-init-%t", needsInit), func(t *testing.T) {
			var events []string
			got := netCodeCacheLookupObject4ECD90(uint32(7), netCodeCacheLookupHooks4ECD90[
				*netCodeCacheLookupTestEntry4ECD90,
				*netCodeCacheLookupTestObject4ECD90,
			]{
				loadNeedsInit: func() bool {
					events = append(events, "needs-init")
					return needsInit
				},
				initCache: func() {
					events = append(events, "init")
				},
				loadFirstUsed: func() *netCodeCacheLookupTestEntry4ECD90 {
					events = append(events, "first")
					return nil
				},
				loadObject: func(*netCodeCacheLookupTestEntry4ECD90) *netCodeCacheLookupTestObject4ECD90 {
					t.Fatal("empty cache loaded an object")
					return nil
				},
				loadNetCode: func(*netCodeCacheLookupTestObject4ECD90) uint32 { t.Fatal("empty cache loaded a net code"); return 0 },
				loadNext: func(*netCodeCacheLookupTestEntry4ECD90) *netCodeCacheLookupTestEntry4ECD90 {
					t.Fatal("empty cache loaded next")
					return nil
				},
				removeUsed:  func(*netCodeCacheLookupTestEntry4ECD90) { t.Fatal("empty cache removed an entry") },
				prependUsed: func(*netCodeCacheLookupTestEntry4ECD90) { t.Fatal("empty cache prepended an entry") },
			})
			if got != nil {
				t.Fatalf("result = %p, want nil", got)
			}
			want := []string{"needs-init", "first"}
			if needsInit {
				want = []string{"needs-init", "init", "first"}
			}
			if !reflect.DeepEqual(events, want) {
				t.Fatalf("events = %v, want %v", events, want)
			}
		})
	}
}

func TestNetCodeCacheLookup4ECD90SearchPromoteAndLiveReload(t *testing.T) {
	firstObject := &netCodeCacheLookupTestObject4ECD90{name: "first-object", code: 3}
	matchedObject := &netCodeCacheLookupTestObject4ECD90{name: "matched-object", code: 0xffffffff}
	liveObject := &netCodeCacheLookupTestObject4ECD90{name: "live-object", code: 9}
	matchedEntry := &netCodeCacheLookupTestEntry4ECD90{name: "matched", object: matchedObject}
	firstEntry := &netCodeCacheLookupTestEntry4ECD90{name: "first", object: firstObject, next: matchedEntry}
	var events []string
	hooks := netCodeCacheLookupHooks4ECD90[
		*netCodeCacheLookupTestEntry4ECD90,
		*netCodeCacheLookupTestObject4ECD90,
	]{
		loadNeedsInit: func() bool {
			events = append(events, "needs-init")
			return false
		},
		initCache: func() { t.Fatal("initialized an initialized cache") },
		loadFirstUsed: func() *netCodeCacheLookupTestEntry4ECD90 {
			events = append(events, "first")
			return firstEntry
		},
		loadObject: func(entry *netCodeCacheLookupTestEntry4ECD90) *netCodeCacheLookupTestObject4ECD90 {
			events = append(events, "object:"+entry.name)
			return entry.object
		},
		loadNetCode: func(obj *netCodeCacheLookupTestObject4ECD90) uint32 {
			events = append(events, "code:"+obj.name)
			return obj.code
		},
		loadNext: func(entry *netCodeCacheLookupTestEntry4ECD90) *netCodeCacheLookupTestEntry4ECD90 {
			events = append(events, "next:"+entry.name)
			return entry.next
		},
		removeUsed: func(entry *netCodeCacheLookupTestEntry4ECD90) {
			events = append(events, "remove:"+entry.name)
		},
		prependUsed: func(entry *netCodeCacheLookupTestEntry4ECD90) {
			events = append(events, "prepend:"+entry.name)
			entry.object = liveObject
		},
	}
	got := netCodeCacheLookupObject4ECD90(^uint32(0), hooks)
	if got != liveObject {
		t.Fatalf("result = %p, want live reloaded object %p", got, liveObject)
	}
	want := []string{
		"needs-init",
		"first",
		"object:first",
		"code:first-object",
		"next:first",
		"object:matched",
		"code:matched-object",
		"remove:matched",
		"prepend:matched",
		"object:matched",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestNetCodeCacheLookup4ECD90MissStopsAtNullNext(t *testing.T) {
	second := &netCodeCacheLookupTestEntry4ECD90{
		name:   "second",
		object: &netCodeCacheLookupTestObject4ECD90{name: "second-object", code: 2},
	}
	first := &netCodeCacheLookupTestEntry4ECD90{
		name:   "first",
		object: &netCodeCacheLookupTestObject4ECD90{name: "first-object", code: 1},
		next:   second,
	}
	var events []string
	got := netCodeCacheLookupObject4ECD90(uint32(3), netCodeCacheLookupHooks4ECD90[
		*netCodeCacheLookupTestEntry4ECD90,
		*netCodeCacheLookupTestObject4ECD90,
	]{
		loadNeedsInit: func() bool { events = append(events, "needs-init"); return false },
		initCache:     func() { t.Fatal("unexpected init") },
		loadFirstUsed: func() *netCodeCacheLookupTestEntry4ECD90 { events = append(events, "first"); return first },
		loadObject: func(entry *netCodeCacheLookupTestEntry4ECD90) *netCodeCacheLookupTestObject4ECD90 {
			events = append(events, "object:"+entry.name)
			return entry.object
		},
		loadNetCode: func(obj *netCodeCacheLookupTestObject4ECD90) uint32 {
			events = append(events, "code:"+obj.name)
			return obj.code
		},
		loadNext: func(entry *netCodeCacheLookupTestEntry4ECD90) *netCodeCacheLookupTestEntry4ECD90 {
			events = append(events, "next:"+entry.name)
			return entry.next
		},
		removeUsed:  func(*netCodeCacheLookupTestEntry4ECD90) { t.Fatal("miss removed an entry") },
		prependUsed: func(*netCodeCacheLookupTestEntry4ECD90) { t.Fatal("miss prepended an entry") },
	})
	if got != nil {
		t.Fatalf("result = %p, want nil", got)
	}
	want := []string{
		"needs-init", "first",
		"object:first", "code:first-object", "next:first",
		"object:second", "code:second-object", "next:second",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestNetCodeCacheLookup4ECD90MutationFaultOrder(t *testing.T) {
	const fault = "remove fault"
	entry := &netCodeCacheLookupTestEntry4ECD90{
		name:   "match",
		object: &netCodeCacheLookupTestObject4ECD90{name: "object", code: 4},
	}
	var events []string
	defer func() {
		if got := recover(); got != fault {
			t.Fatalf("panic = %v, want %q", got, fault)
		}
		want := []string{"needs-init", "first", "object", "code", "remove"}
		if !reflect.DeepEqual(events, want) {
			t.Fatalf("events = %v, want %v", events, want)
		}
	}()
	netCodeCacheLookupObject4ECD90(uint32(4), netCodeCacheLookupHooks4ECD90[
		*netCodeCacheLookupTestEntry4ECD90,
		*netCodeCacheLookupTestObject4ECD90,
	]{
		loadNeedsInit: func() bool { events = append(events, "needs-init"); return false },
		initCache:     func() { t.Fatal("unexpected init") },
		loadFirstUsed: func() *netCodeCacheLookupTestEntry4ECD90 { events = append(events, "first"); return entry },
		loadObject: func(*netCodeCacheLookupTestEntry4ECD90) *netCodeCacheLookupTestObject4ECD90 {
			events = append(events, "object")
			return entry.object
		},
		loadNetCode: func(*netCodeCacheLookupTestObject4ECD90) uint32 { events = append(events, "code"); return 4 },
		loadNext: func(*netCodeCacheLookupTestEntry4ECD90) *netCodeCacheLookupTestEntry4ECD90 {
			t.Fatal("match loaded next")
			return nil
		},
		removeUsed: func(*netCodeCacheLookupTestEntry4ECD90) {
			events = append(events, "remove")
			panic(fault)
		},
		prependUsed: func(*netCodeCacheLookupTestEntry4ECD90) { t.Fatal("remove fault prepended entry") },
	})
}
