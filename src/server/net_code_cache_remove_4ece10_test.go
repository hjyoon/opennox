package server

import (
	"fmt"
	"reflect"
	"testing"
)

func TestNetCodeCacheRemove4ECE10ReloadsLinks(t *testing.T) {
	tests := []struct {
		name       string
		nextLoads  []string
		prevLoads  []string
		wantResult string
		wantEvents []string
	}{
		{
			name:       "successor-and-predecessor",
			nextLoads:  []string{"first-next", "live-next"},
			prevLoads:  []string{"first-prev", "live-prev"},
			wantResult: "entry",
			wantEvents: []string{
				"load-next:first-next", "load-prev:first-prev", "prev:first-next=first-prev",
				"load-prev:live-prev", "load-next:live-next", "next:live-prev=live-next",
			},
		},
		{
			name:       "successor-then-live-head",
			nextLoads:  []string{"first-next", "live-next"},
			prevLoads:  []string{"first-prev", ""},
			wantResult: "live-next",
			wantEvents: []string{
				"load-next:first-next", "load-prev:first-prev", "prev:first-next=first-prev",
				"load-prev:", "load-next:live-next", "store-first:live-next",
			},
		},
		{
			name:       "tail-then-live-predecessor",
			nextLoads:  []string{"", "live-next"},
			prevLoads:  []string{"first-prev", "live-prev"},
			wantResult: "entry",
			wantEvents: []string{
				"load-next:", "load-prev:first-prev", "store-last:first-prev",
				"load-prev:live-prev", "load-next:live-next", "next:live-prev=live-next",
			},
		},
		{
			name:       "tail-then-live-head",
			nextLoads:  []string{"", "live-next"},
			prevLoads:  []string{"first-prev", ""},
			wantResult: "live-next",
			wantEvents: []string{
				"load-next:", "load-prev:first-prev", "store-last:first-prev",
				"load-prev:", "load-next:live-next", "store-first:live-next",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var events []string
			nextIndex := 0
			prevIndex := 0
			got := netCodeCacheRemove4ECE10("entry", netCodeCacheRemoveHooks4ECE10[string]{
				loadEntryNext: func(entry string) string {
					if entry != "entry" {
						t.Fatalf("next load entry = %q, want entry", entry)
					}
					value := test.nextLoads[nextIndex]
					nextIndex++
					events = append(events, "load-next:"+value)
					return value
				},
				loadEntryPrev: func(entry string) string {
					if entry != "entry" {
						t.Fatalf("prev load entry = %q, want entry", entry)
					}
					value := test.prevLoads[prevIndex]
					prevIndex++
					events = append(events, "load-prev:"+value)
					return value
				},
				storeEntryPrev: func(entry, prev string) {
					events = append(events, "prev:"+entry+"="+prev)
				},
				storeLast: func(entry string) {
					events = append(events, "store-last:"+entry)
				},
				storeEntryNext: func(entry, next string) {
					events = append(events, "next:"+entry+"="+next)
				},
				storeFirst: func(entry string) {
					events = append(events, "store-first:"+entry)
				},
			})
			if got != test.wantResult {
				t.Fatalf("result = %q, want %q", got, test.wantResult)
			}
			if nextIndex != 2 || prevIndex != 2 {
				t.Fatalf("loads = next %d prev %d, want 2/2", nextIndex, prevIndex)
			}
			if !reflect.DeepEqual(events, test.wantEvents) {
				t.Fatalf("events = %v, want %v", events, test.wantEvents)
			}
		})
	}
}

func TestNetCodeCacheRemove4ECE10FaultOrder(t *testing.T) {
	branches := []struct {
		name       string
		nextLoads  []string
		prevLoads  []string
		wantEvents []string
	}{
		{
			name:      "successor-and-predecessor",
			nextLoads: []string{"first-next", "live-next"},
			prevLoads: []string{"first-prev", "live-prev"},
			wantEvents: []string{
				"load-next:first-next", "load-prev:first-prev", "prev:first-next=first-prev",
				"load-prev:live-prev", "load-next:live-next", "next:live-prev=live-next",
			},
		},
		{
			name:      "successor-then-head",
			nextLoads: []string{"first-next", "live-next"},
			prevLoads: []string{"first-prev", ""},
			wantEvents: []string{
				"load-next:first-next", "load-prev:first-prev", "prev:first-next=first-prev",
				"load-prev:", "load-next:live-next", "store-first:live-next",
			},
		},
		{
			name:      "tail-then-predecessor",
			nextLoads: []string{"", "live-next"},
			prevLoads: []string{"first-prev", "live-prev"},
			wantEvents: []string{
				"load-next:", "load-prev:first-prev", "store-last:first-prev",
				"load-prev:live-prev", "load-next:live-next", "next:live-prev=live-next",
			},
		},
		{
			name:      "tail-then-head",
			nextLoads: []string{"", "live-next"},
			prevLoads: []string{"first-prev", ""},
			wantEvents: []string{
				"load-next:", "load-prev:first-prev", "store-last:first-prev",
				"load-prev:", "load-next:live-next", "store-first:live-next",
			},
		},
	}
	for _, branch := range branches {
		for faultAt := 1; faultAt <= len(branch.wantEvents); faultAt++ {
			t.Run(fmt.Sprintf("%s/%d", branch.name, faultAt), func(t *testing.T) {
				var events []string
				nextIndex := 0
				prevIndex := 0
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
				netCodeCacheRemove4ECE10("entry", netCodeCacheRemoveHooks4ECE10[string]{
					loadEntryNext: func(string) string {
						value := branch.nextLoads[nextIndex]
						nextIndex++
						record("load-next:" + value)
						return value
					},
					loadEntryPrev: func(string) string {
						value := branch.prevLoads[prevIndex]
						prevIndex++
						record("load-prev:" + value)
						return value
					},
					storeEntryPrev: func(entry, prev string) {
						record("prev:" + entry + "=" + prev)
					},
					storeLast: func(entry string) {
						record("store-last:" + entry)
					},
					storeEntryNext: func(entry, next string) {
						record("next:" + entry + "=" + next)
					},
					storeFirst: func(entry string) {
						record("store-first:" + entry)
					},
				})
			})
		}
	}
}
