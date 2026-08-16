package server

import (
	"fmt"
	"reflect"
	"testing"
)

func netCodeCacheAddEvents4ECEA0(next string) []string {
	if next != "" {
		return []string{
			"next:" + next,
			"object:value",
			"store:" + next + "=value",
			"prepend:" + next,
		}
	}
	return []string{
		"next:",
		"object:value",
		"last:used-tail",
		"store:used-tail=value",
		"remove:used-tail",
		"prepend:used-tail",
	}
}

func netCodeCacheAddHooksForTest4ECEA0(next string, record func(string)) netCodeCacheAddHooks4ECEA0[string, string, string] {
	return netCodeCacheAddHooks4ECEA0[string, string, string]{
		nextUnused: func() string {
			record("next:" + next)
			return next
		},
		loadObject: func() string {
			record("object:value")
			return "value"
		},
		loadLastUsed: func() string {
			record("last:used-tail")
			return "used-tail"
		},
		storeObject: func(entry, obj string) {
			record("store:" + entry + "=" + obj)
		},
		removeUsed: func(entry string) {
			record("remove:" + entry)
		},
		prependUsed: func(entry string) string {
			record("prepend:" + entry)
			return "result:" + entry
		},
	}
}

func TestNetCodeCacheAddObject4ECEA0OrderAndReturn(t *testing.T) {
	tests := []struct {
		name       string
		next       string
		wantResult string
	}{
		{name: "free-entry", next: "free", wantResult: "result:free"},
		{name: "reuse-used-tail", wantResult: "result:used-tail"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var events []string
			got := netCodeCacheAddObject4ECEA0(netCodeCacheAddHooksForTest4ECEA0(test.next, func(event string) {
				events = append(events, event)
			}))
			if got != test.wantResult {
				t.Fatalf("result = %q, want %q", got, test.wantResult)
			}
			want := netCodeCacheAddEvents4ECEA0(test.next)
			if !reflect.DeepEqual(events, want) {
				t.Fatalf("events = %v, want %v", events, want)
			}
		})
	}
}

func TestNetCodeCacheAddObject4ECEA0FaultOrder(t *testing.T) {
	for _, next := range []string{"free", ""} {
		name := "free-entry"
		if next == "" {
			name = "reuse-used-tail"
		}
		want := netCodeCacheAddEvents4ECEA0(next)
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
				netCodeCacheAddObject4ECEA0(netCodeCacheAddHooksForTest4ECEA0(next, record))
			})
		}
	}
}
