package server

import (
	"reflect"
	"testing"
)

type chestInitTestObject4F0400 struct {
	status uint32
}

func TestChestInit4F0400SetsBitOnlyWhenMaskIsClear(t *testing.T) {
	tests := []struct {
		name      string
		status    uint32
		wantCalls int
	}{
		{name: "zero", status: 0, wantCalls: 1},
		{name: "unrelated low bits", status: 0x000000f1, wantCalls: 1},
		{name: "high bytes ignored", status: 0xffffff01, wantCalls: 1},
		{name: "bit one", status: 0x00000002},
		{name: "bit two", status: 0x00000004},
		{name: "bit three", status: 0x00000008},
		{name: "all masked bits", status: 0xffffff0e},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			unit := &chestInitTestObject4F0400{status: tc.status}
			events := make([]string, 0, 2)
			calls := 0
			chestInit4F0400(unit, chestInitHooks4F0400[*chestInitTestObject4F0400]{
				loadStatusLow: func(got *chestInitTestObject4F0400) uint8 {
					events = append(events, "load-status-low")
					if got != unit {
						t.Fatalf("load object = %p, want %p", got, unit)
					}
					return uint8(got.status)
				},
				setXStatus: func(got *chestInitTestObject4F0400, bit uint32) {
					events = append(events, "set-xstatus")
					calls++
					if got != unit || bit != chestInitStatusBit4F0400 {
						t.Fatalf("set args = %p/%#x, want %p/2", got, bit, unit)
					}
					got.status |= bit
				},
			})

			if calls != tc.wantCalls {
				t.Fatalf("set calls = %d, want %d", calls, tc.wantCalls)
			}
			wantEvents := []string{"load-status-low"}
			wantStatus := tc.status
			if tc.wantCalls != 0 {
				wantEvents = append(wantEvents, "set-xstatus")
				wantStatus |= chestInitStatusBit4F0400
			}
			if !reflect.DeepEqual(events, wantEvents) {
				t.Fatalf("events = %v, want %v", events, wantEvents)
			}
			if unit.status != wantStatus {
				t.Fatalf("status = %#x, want %#x", unit.status, wantStatus)
			}
		})
	}
}

func TestChestInit4F0400FaultPrefixes(t *testing.T) {
	tests := []struct {
		stage      string
		wantEvents []string
	}{
		{stage: "load-status-low", wantEvents: []string{"load-status-low"}},
		{stage: "set-xstatus", wantEvents: []string{"load-status-low", "set-xstatus"}},
	}
	for _, tc := range tests {
		t.Run(tc.stage, func(t *testing.T) {
			unit := &chestInitTestObject4F0400{}
			events := make([]string, 0, 2)
			func() {
				defer func() {
					if got := recover(); got != tc.stage {
						t.Fatalf("panic = %v, want %q", got, tc.stage)
					}
				}()
				chestInit4F0400(unit, chestInitHooks4F0400[*chestInitTestObject4F0400]{
					loadStatusLow: func(*chestInitTestObject4F0400) uint8 {
						events = append(events, "load-status-low")
						if tc.stage == "load-status-low" {
							panic(tc.stage)
						}
						return 0
					},
					setXStatus: func(*chestInitTestObject4F0400, uint32) {
						events = append(events, "set-xstatus")
						panic(tc.stage)
					},
				})
			}()
			if !reflect.DeepEqual(events, tc.wantEvents) {
				t.Fatalf("events = %v, want %v", events, tc.wantEvents)
			}
		})
	}
}
