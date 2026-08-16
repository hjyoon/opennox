package server

import (
	"fmt"
	"reflect"
	"testing"
)

func netCodeCacheNextUnusedEvents4ECEF0(head string) []string {
	if head == "" {
		return []string{"first:"}
	}
	return []string{
		"first:" + head,
		"next:" + head + "=next",
		"store-first:next",
	}
}

func netCodeCacheNextUnusedHooksForTest4ECEF0(head string, record func(string)) netCodeCacheNextUnusedHooks4ECEF0[string] {
	return netCodeCacheNextUnusedHooks4ECEF0[string]{
		loadFirstFree: func() string {
			record("first:" + head)
			return head
		},
		loadEntryNext: func(entry string) string {
			record("next:" + entry + "=next")
			return "next"
		},
		storeFirstFree: func(entry string) {
			record("store-first:" + entry)
		},
	}
}

func TestNetCodeCacheNextUnused4ECEF0OrderAndReturn(t *testing.T) {
	for _, head := range []string{"", "head"} {
		name := "empty"
		wantResult := ""
		if head != "" {
			name = "nonempty"
			wantResult = "head"
		}
		t.Run(name, func(t *testing.T) {
			var events []string
			got := netCodeCacheNextUnused4ECEF0(netCodeCacheNextUnusedHooksForTest4ECEF0(head, func(event string) {
				events = append(events, event)
			}))
			if got != wantResult {
				t.Fatalf("result = %q, want %q", got, wantResult)
			}
			want := netCodeCacheNextUnusedEvents4ECEF0(head)
			if !reflect.DeepEqual(events, want) {
				t.Fatalf("events = %v, want %v", events, want)
			}
		})
	}
}

func TestNetCodeCacheNextUnused4ECEF0FaultOrder(t *testing.T) {
	for _, head := range []string{"", "head"} {
		name := "empty"
		if head != "" {
			name = "nonempty"
		}
		want := netCodeCacheNextUnusedEvents4ECEF0(head)
		for faultAt := 1; faultAt <= len(want); faultAt++ {
			t.Run(fmt.Sprintf("%s/event-%d", name, faultAt), func(t *testing.T) {
				var events []string
				record := func(event string) {
					events = append(events, event)
					if len(events) == faultAt {
						panic(faultAt)
					}
				}
				defer func() {
					if got := recover(); got != faultAt {
						t.Fatalf("panic = %v, want %d", got, faultAt)
					}
					if !reflect.DeepEqual(events, want[:faultAt]) {
						t.Fatalf("events = %v, want %v", events, want[:faultAt])
					}
				}()
				netCodeCacheNextUnused4ECEF0(netCodeCacheNextUnusedHooksForTest4ECEF0(head, record))
			})
		}
	}
}
