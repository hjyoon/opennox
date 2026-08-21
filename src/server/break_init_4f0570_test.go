package server

import (
	"reflect"
	"testing"
)

type breakInitTestObject4F0570 struct {
	status uint32
}

func TestBreakInit4F0570SetsBitOnlyWhenLowMaskIsClear(t *testing.T) {
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
			unit := &breakInitTestObject4F0570{status: tc.status}
			events := make([]string, 0, 2)
			calls := 0
			breakInit4F0570(unit, breakInitHooks4F0570[*breakInitTestObject4F0570]{
				loadStatusLow: func(got *breakInitTestObject4F0570) uint8 {
					events = append(events, "load-status-low")
					if got != unit {
						t.Fatalf("load object = %p, want %p", got, unit)
					}
					return uint8(got.status)
				},
				setXStatus: func(got *breakInitTestObject4F0570, bit uint32) {
					events = append(events, "set-xstatus")
					calls++
					if got != unit || bit != breakInitStatusBit4F0570 {
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
				wantStatus |= breakInitStatusBit4F0570
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

func TestBreakInit4F0570CachesObjectAndDoesNotReloadStatus(t *testing.T) {
	entry := &breakInitTestObject4F0570{}
	replacement := &breakInitTestObject4F0570{status: 0xffffffff}
	live := entry
	loads := 0
	breakInit4F0570(entry, breakInitHooks4F0570[*breakInitTestObject4F0570]{
		loadStatusLow: func(got *breakInitTestObject4F0570) uint8 {
			loads++
			live = replacement
			return uint8(got.status)
		},
		setXStatus: func(got *breakInitTestObject4F0570, bit uint32) {
			if got != entry || got == live || bit != 2 {
				t.Fatalf("set args = %p/%#x, entry/live = %p/%p", got, bit, entry, live)
			}
			got.status |= bit
		},
	})
	if loads != 1 || entry.status != 2 || replacement.status != 0xffffffff {
		t.Fatalf("loads/entry/replacement = %d/%#x/%#x", loads, entry.status, replacement.status)
	}
}

func TestBreakInit4F0570FaultPrefixes(t *testing.T) {
	tests := []struct {
		stage      string
		wantEvents []string
	}{
		{stage: "load-status-low", wantEvents: []string{"load-status-low"}},
		{stage: "set-xstatus", wantEvents: []string{"load-status-low", "set-xstatus"}},
	}
	for _, tc := range tests {
		t.Run(tc.stage, func(t *testing.T) {
			unit := &breakInitTestObject4F0570{}
			events := make([]string, 0, 2)
			func() {
				defer func() {
					if got := recover(); got != tc.stage {
						t.Fatalf("panic = %v, want %q", got, tc.stage)
					}
				}()
				breakInit4F0570(unit, breakInitHooks4F0570[*breakInitTestObject4F0570]{
					loadStatusLow: func(*breakInitTestObject4F0570) uint8 {
						events = append(events, "load-status-low")
						if tc.stage == "load-status-low" {
							panic(tc.stage)
						}
						return 0
					},
					setXStatus: func(*breakInitTestObject4F0570, uint32) {
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

func TestBreakInitParse536910IgnoresTextAndStoresFullDword(t *testing.T) {
	type objectType struct {
		field9 uint32
		guard  uint32
	}
	for _, definition := range []string{"", "ignored", "2 3", "\x00\xff"} {
		got := &objectType{field9: 0xffffffff, guard: 0xa5a5a5a5}
		_ = definition // The original parser never observes definition text.
		result := breakInitParse536910(got, func(current *objectType, value uint32) {
			if current != got || value != 2 {
				t.Fatalf("store args = %p/%#x, want %p/2", current, value, got)
			}
			current.field9 = value
		})
		if result != 1 || got.field9 != 2 || got.guard != 0xa5a5a5a5 {
			t.Fatalf("definition %q: result/record = %d/%+v", definition, result, *got)
		}
	}
}

func TestBreakInitParse536910StoreFaultPreventsReturn(t *testing.T) {
	const fault = "store-field9"
	defer func() {
		if got := recover(); got != fault {
			t.Fatalf("panic = %v, want %q", got, fault)
		}
	}()
	breakInitParse536910(struct{}{}, func(struct{}, uint32) {
		panic(fault)
	})
}
