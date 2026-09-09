package server

import (
	"fmt"
	"reflect"
	"testing"
)

const (
	creatureMonitoredHighOwner500CC0  = uint64(0x7f03ec213490)
	creatureMonitoredHighUnit500CC0   = uint64(0x7ff71d949570)
	creatureMonitoredHighUpdate500CC0 = uint64(0x7fb3078c7430)
)

type creatureMonitoredTestWorld500CC0 struct {
	events    []string
	faultAt   int
	classLow  uint8
	flags     uint32
	zombie    bool
	statusLow uint8
	hasOwner  bool
}

func (w *creatureMonitoredTestWorld500CC0) observe(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
}

func (w *creatureMonitoredTestWorld500CC0) hooks() creatureMonitoredHooks500CC0[uint64, uint64] {
	return creatureMonitoredHooks500CC0[uint64, uint64]{
		loadClassLow: func(unit uint64) uint8 {
			w.observe(fmt.Sprintf("class:%x", unit))
			return w.classLow
		},
		loadFlags: func(unit uint64) uint32 {
			w.observe(fmt.Sprintf("flags:%x", unit))
			return w.flags
		},
		isZombie: func(unit uint64) bool {
			w.observe(fmt.Sprintf("zombie:%x", unit))
			return w.zombie
		},
		loadUpdate: func(unit uint64) uint64 {
			w.observe(fmt.Sprintf("update:%x", unit))
			return creatureMonitoredHighUpdate500CC0
		},
		loadStatusLow: func(update uint64) uint8 {
			w.observe(fmt.Sprintf("status:%x", update))
			return w.statusLow
		},
		hasOwner: func(unit, owner uint64) bool {
			w.observe(fmt.Sprintf("owner:%x:%x", unit, owner))
			return w.hasOwner
		},
	}
}

func newCreatureMonitoredTestWorld500CC0() *creatureMonitoredTestWorld500CC0 {
	return &creatureMonitoredTestWorld500CC0{
		classLow:  creatureMonitoredMonsterClassLow500CC0,
		statusLow: creatureMonitoredSummonedStatus500CC0,
		hasOwner:  true,
	}
}

func TestCreatureIsMonitored500CC0AliveMonsterOrderAndNativeWidthHandles(t *testing.T) {
	w := newCreatureMonitoredTestWorld500CC0()
	if !creatureIsMonitored500CC0(creatureMonitoredHighOwner500CC0, creatureMonitoredHighUnit500CC0, w.hooks()) {
		t.Fatal("alive summoned Monster with the requested owner was rejected")
	}
	want := []string{
		"class:7ff71d949570",
		"flags:7ff71d949570",
		"update:7ff71d949570",
		"status:7fb3078c7430",
		"owner:7ff71d949570:7f03ec213490",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want exact oracle order %q", w.events, want)
	}
}

func TestCreatureIsMonitored500CC0ExactBranches(t *testing.T) {
	tests := []struct {
		name       string
		classLow   uint8
		flags      uint32
		zombie     bool
		statusLow  uint8
		hasOwner   bool
		want       bool
		wantEvents []string
	}{
		{
			name:     "alive Monster bypasses zombie",
			classLow: 0xfe, statusLow: 0x80, hasOwner: true, want: true,
			wantEvents: []string{"class", "flags", "update", "status", "owner"},
		},
		{
			name:     "dead Monster rejected when not zombie",
			classLow: 0x02, flags: 0xffff8000,
			wantEvents: []string{"class", "flags", "zombie"},
		},
		{
			name:     "dead Monster zombie accepted",
			classLow: 0x02, flags: 0x00008000, zombie: true, statusLow: 0xff, hasOwner: true, want: true,
			wantEvents: []string{"class", "flags", "zombie", "update", "status", "owner"},
		},
		{
			name:     "non-Monster skips flags and rejects when not zombie",
			classLow: 0xfd, flags: 0xffffffff,
			wantEvents: []string{"class", "zombie"},
		},
		{
			name:     "non-Monster zombie is accepted by original expression",
			classLow: 0xfd, flags: 0xffffffff, zombie: true, statusLow: 0x80, hasOwner: true, want: true,
			wantEvents: []string{"class", "zombie", "update", "status", "owner"},
		},
		{
			name:     "summoned bit absent",
			classLow: 0x02, statusLow: 0x7f, hasOwner: true,
			wantEvents: []string{"class", "flags", "update", "status"},
		},
		{
			name:     "different owner",
			classLow: 0x02, statusLow: 0x80,
			wantEvents: []string{"class", "flags", "update", "status", "owner"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := &creatureMonitoredTestWorld500CC0{
				classLow: tc.classLow, flags: tc.flags, zombie: tc.zombie,
				statusLow: tc.statusLow, hasOwner: tc.hasOwner,
			}
			got := creatureIsMonitored500CC0(creatureMonitoredHighOwner500CC0, creatureMonitoredHighUnit500CC0, w.hooks())
			if got != tc.want {
				t.Fatalf("result = %t, want %t", got, tc.want)
			}
			gotEvents := eventKinds500CC0(w.events)
			if !reflect.DeepEqual(gotEvents, tc.wantEvents) {
				t.Fatalf("event kinds = %q, want %q (events %q)", gotEvents, tc.wantEvents, w.events)
			}
		})
	}
}

func eventKinds500CC0(events []string) []string {
	kinds := make([]string, len(events))
	for i, event := range events {
		for j, c := range event {
			if c == ':' {
				kinds[i] = event[:j]
				break
			}
		}
	}
	return kinds
}

func TestCreatureIsMonitored500CC0DoesNotPreflightNullUnit(t *testing.T) {
	w := newCreatureMonitoredTestWorld500CC0()
	hooks := w.hooks()
	hooks.loadClassLow = func(unit uint64) uint8 {
		w.observe(fmt.Sprintf("class:%x", unit))
		panic("original null class read")
	}
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		creatureIsMonitored500CC0(creatureMonitoredHighOwner500CC0, 0, hooks)
	}()
	if recovered == nil {
		t.Fatal("zero unit was incorrectly returned before the original class read")
	}
	if want := []string{"class:0"}; !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want %q", w.events, want)
	}
}

func TestCreatureIsMonitored500CC0AllFaultPrefixes(t *testing.T) {
	baseline := newCreatureMonitoredTestWorld500CC0()
	baseline.flags = creatureMonitoredDeadFlag500CC0
	baseline.zombie = true
	creatureIsMonitored500CC0(creatureMonitoredHighOwner500CC0, creatureMonitoredHighUnit500CC0, baseline.hooks())
	want := append([]string(nil), baseline.events...)

	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			w := newCreatureMonitoredTestWorld500CC0()
			w.flags = creatureMonitoredDeadFlag500CC0
			w.zombie = true
			w.faultAt = faultAt
			var recovered any
			func() {
				defer func() { recovered = recover() }()
				creatureIsMonitored500CC0(creatureMonitoredHighOwner500CC0, creatureMonitoredHighUnit500CC0, w.hooks())
			}()
			if recovered == nil {
				t.Fatal("fault sentinel was not recovered")
			}
			if prefix := want[:faultAt]; !reflect.DeepEqual(w.events, prefix) {
				t.Fatalf("events = %q, want fault prefix %q", w.events, prefix)
			}
		})
	}
}
