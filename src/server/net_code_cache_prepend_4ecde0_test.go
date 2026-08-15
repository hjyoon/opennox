package server

import (
	"fmt"
	"reflect"
	"testing"
)

func TestNetCodeCachePrepend4ECDE0ReloadsHead(t *testing.T) {
	loads := []string{"snapshot-head", "live-head"}
	loadIndex := 0
	prev := make(map[string]string)
	next := make(map[string]string)
	first := "snapshot-head"
	last := "tail"
	var events []string

	got := netCodeCachePrepend4ECDE0("entry", netCodeCachePrependHooks4ECDE0[string]{
		storeEntryPrev: func(entry, value string) {
			events = append(events, "prev:"+entry+"="+value)
			prev[entry] = value
		},
		loadFirst: func() string {
			value := loads[loadIndex]
			loadIndex++
			events = append(events, "load-first:"+value)
			return value
		},
		storeEntryNext: func(entry, value string) {
			events = append(events, "next:"+entry+"="+value)
			next[entry] = value
		},
		storeLast: func(value string) {
			t.Fatalf("non-empty second head stored last = %q", value)
		},
		storeFirst: func(value string) {
			events = append(events, "store-first:"+value)
			first = value
		},
	})

	if got != "entry" {
		t.Fatalf("result = %q, want entry", got)
	}
	if prev["entry"] != "" || next["entry"] != "snapshot-head" {
		t.Fatalf("entry links = prev %q next %q, want empty/snapshot-head", prev["entry"], next["entry"])
	}
	if prev["live-head"] != "entry" {
		t.Fatalf("live head prev = %q, want entry", prev["live-head"])
	}
	if first != "entry" || last != "tail" {
		t.Fatalf("list = first %q last %q, want entry/tail", first, last)
	}
	want := []string{
		"prev:entry=",
		"load-first:snapshot-head",
		"next:entry=snapshot-head",
		"load-first:live-head",
		"prev:live-head=entry",
		"store-first:entry",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestNetCodeCachePrepend4ECDE0SecondHeadControlsBranch(t *testing.T) {
	tests := []struct {
		name       string
		loads      []string
		wantEvents []string
	}{
		{
			name:  "first-null-second-live",
			loads: []string{"", "live-head"},
			wantEvents: []string{
				"prev:entry=", "load-first:", "next:entry=", "load-first:live-head",
				"prev:live-head=entry", "store-first:entry",
			},
		},
		{
			name:  "first-live-second-null",
			loads: []string{"snapshot-head", ""},
			wantEvents: []string{
				"prev:entry=", "load-first:snapshot-head", "next:entry=snapshot-head", "load-first:",
				"store-last:entry", "store-first:entry",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			loadIndex := 0
			var events []string
			got := netCodeCachePrepend4ECDE0("entry", netCodeCachePrependHooks4ECDE0[string]{
				storeEntryPrev: func(entry, value string) {
					events = append(events, "prev:"+entry+"="+value)
				},
				loadFirst: func() string {
					value := test.loads[loadIndex]
					loadIndex++
					events = append(events, "load-first:"+value)
					return value
				},
				storeEntryNext: func(entry, value string) {
					events = append(events, "next:"+entry+"="+value)
				},
				storeLast: func(value string) {
					events = append(events, "store-last:"+value)
				},
				storeFirst: func(value string) {
					events = append(events, "store-first:"+value)
				},
			})
			if got != "entry" {
				t.Fatalf("result = %q, want entry", got)
			}
			if !reflect.DeepEqual(events, test.wantEvents) {
				t.Fatalf("events = %v, want %v", events, test.wantEvents)
			}
		})
	}
}

func TestNetCodeCachePrepend4ECDE0FaultOrder(t *testing.T) {
	branches := []struct {
		name       string
		loads      []string
		wantEvents []string
	}{
		{
			name:  "non-empty",
			loads: []string{"snapshot-head", "live-head"},
			wantEvents: []string{
				"prev:entry=", "load-first:snapshot-head", "next:entry=snapshot-head",
				"load-first:live-head", "prev:live-head=entry", "store-first:entry",
			},
		},
		{
			name:  "empty",
			loads: []string{"snapshot-head", ""},
			wantEvents: []string{
				"prev:entry=", "load-first:snapshot-head", "next:entry=snapshot-head",
				"load-first:", "store-last:entry", "store-first:entry",
			},
		},
	}
	for _, branch := range branches {
		for faultAt := 1; faultAt <= len(branch.wantEvents); faultAt++ {
			t.Run(fmt.Sprintf("%s/%d", branch.name, faultAt), func(t *testing.T) {
				var events []string
				loadIndex := 0
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
					want := branch.wantEvents[:faultAt]
					if !reflect.DeepEqual(events, want) {
						t.Fatalf("events = %v, want %v", events, want)
					}
				}()
				netCodeCachePrepend4ECDE0("entry", netCodeCachePrependHooks4ECDE0[string]{
					storeEntryPrev: func(entry, value string) {
						record("prev:" + entry + "=" + value)
					},
					loadFirst: func() string {
						value := branch.loads[loadIndex]
						loadIndex++
						record("load-first:" + value)
						return value
					},
					storeEntryNext: func(entry, value string) {
						record("next:" + entry + "=" + value)
					},
					storeLast: func(value string) {
						record("store-last:" + value)
					},
					storeFirst: func(value string) {
						record("store-first:" + value)
					},
				})
			})
		}
	}
}
