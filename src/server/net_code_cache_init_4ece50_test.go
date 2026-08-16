package server

import (
	"fmt"
	"reflect"
	"testing"
)

func netCodeCacheInitEntries4ECE50() [netCodeCacheInitArrayCapacity4ECE50]string {
	var entries [netCodeCacheInitArrayCapacity4ECE50]string
	for i := range entries {
		entries[i] = fmt.Sprintf("entry-%02d", i)
	}
	return entries
}

func netCodeCacheInitEvents4ECE50(entries [netCodeCacheInitArrayCapacity4ECE50]string) []string {
	events := []string{
		"used-first=", "used-last=", "free-first=", "free-last=",
	}
	for _, entry := range entries {
		events = append(events, "prepend:"+entry)
	}
	return append(events, "clear-needs-init")
}

func TestNetCodeCacheInitArray4ECE50OrderAndReturn(t *testing.T) {
	entries := netCodeCacheInitEntries4ECE50()
	var events []string
	got := netCodeCacheInitArray4ECE50(entries, netCodeCacheInitArrayHooks4ECE50[string, string]{
		storeUsedFirst: func(entry string) {
			events = append(events, "used-first="+entry)
		},
		storeUsedLast: func(entry string) {
			events = append(events, "used-last="+entry)
		},
		storeFreeFirst: func(entry string) {
			events = append(events, "free-first="+entry)
		},
		storeFreeLast: func(entry string) {
			events = append(events, "free-last="+entry)
		},
		prependFree: func(entry string) string {
			events = append(events, "prepend:"+entry)
			return "result:" + entry
		},
		clearNeedsInit: func() {
			events = append(events, "clear-needs-init")
		},
	})

	if got != "result:entry-15" {
		t.Fatalf("result = %q, want final prepend result", got)
	}
	want := netCodeCacheInitEvents4ECE50(entries)
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestNetCodeCacheInitArray4ECE50FaultOrder(t *testing.T) {
	entries := netCodeCacheInitEntries4ECE50()
	want := netCodeCacheInitEvents4ECE50(entries)
	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("event-%02d", faultAt), func(t *testing.T) {
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

			netCodeCacheInitArray4ECE50(entries, netCodeCacheInitArrayHooks4ECE50[string, string]{
				storeUsedFirst: func(entry string) {
					record("used-first=" + entry)
				},
				storeUsedLast: func(entry string) {
					record("used-last=" + entry)
				},
				storeFreeFirst: func(entry string) {
					record("free-first=" + entry)
				},
				storeFreeLast: func(entry string) {
					record("free-last=" + entry)
				},
				prependFree: func(entry string) string {
					record("prepend:" + entry)
					return "result:" + entry
				},
				clearNeedsInit: func() {
					record("clear-needs-init")
				},
			})
		})
	}
}
