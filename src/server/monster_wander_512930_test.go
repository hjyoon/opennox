package server

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

const (
	monsterWanderHighUnit512930      = uint64(0x7ff71d949570)
	monsterWanderHighUpdate512930    = uint64(0x7ff7200002ec)
	monsterWanderChangedUpdate512930 = uint64(0x7ff7300002ec)
	monsterWanderHighReport512930    = uint64(0x7ff740000020)
	monsterWanderHighRoam512930      = uint64(0x7ff750000010)
)

type monsterWanderTestWorld512930 struct {
	events        []string
	faultAt       int
	classLow      uint8
	flags         uint32
	update        uint64
	reportResult  uint64
	roamResult    uint64
	roamFlags     map[uint64]uint8
	mutateOnClear bool
	storedU32     map[uint64][4]uint32
	storedLow     map[uint64][4]uint8
}

func (w *monsterWanderTestWorld512930) observe(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
}

func (w *monsterWanderTestWorld512930) hooks() monsterWanderHooks512930[uint64, uint64, uint64] {
	return monsterWanderHooks512930[uint64, uint64, uint64]{
		loadClassLow: func(unit uint64) uint8 {
			w.observe(fmt.Sprintf("class:%x", unit))
			return w.classLow
		},
		loadFlags: func(unit uint64) uint32 {
			w.observe(fmt.Sprintf("flags:%x", unit))
			return w.flags
		},
		loadUpdate: func(unit uint64) uint64 {
			w.observe(fmt.Sprintf("update:%x", unit))
			return w.update
		},
		clearActionStack: func(unit uint64) {
			w.observe(fmt.Sprintf("clear:%x", unit))
			if w.mutateOnClear {
				w.update = monsterWanderChangedUpdate512930
			}
		},
		pushAction: func(unit uint64, action uint32) uint64 {
			w.observe(fmt.Sprintf("push:%x:%d", unit, action))
			switch action {
			case monsterWanderReportAction512930:
				return w.reportResult
			case monsterWanderRoamAction512930:
				return w.roamResult
			default:
				panic(fmt.Sprintf("unexpected action %d", action))
			}
		},
		storeArgU32: func(action uint64, index int, value uint32) {
			w.observe(fmt.Sprintf("store-u32:%x:%d:%08x", action, index, value))
			args := w.storedU32[action]
			args[index] = value
			w.storedU32[action] = args
		},
		loadRoamFlagsLow: func(update uint64) uint8 {
			w.observe(fmt.Sprintf("roam-flags:%x", update))
			return w.roamFlags[update]
		},
		storeArgLow: func(action uint64, index int, value uint8) {
			w.observe(fmt.Sprintf("store-low:%x:%d:%02x", action, index, value))
			args := w.storedLow[action]
			args[index] = value
			w.storedLow[action] = args
		},
	}
}

func newMonsterWanderTestWorld512930() *monsterWanderTestWorld512930 {
	return &monsterWanderTestWorld512930{
		classLow:     monsterWanderMonsterClassLow512930,
		update:       monsterWanderHighUpdate512930,
		reportResult: monsterWanderHighReport512930,
		roamResult:   monsterWanderHighRoam512930,
		roamFlags: map[uint64]uint8{
			monsterWanderHighUpdate512930:    0xa5,
			monsterWanderChangedUpdate512930: 0x5a,
		},
		storedU32: make(map[uint64][4]uint32),
		storedLow: make(map[uint64][4]uint8),
	}
}

func TestMonsterWander512930ExactOrderAndNativeWidthHandles(t *testing.T) {
	w := newMonsterWanderTestWorld512930()
	monsterWander512930(monsterWanderHighUnit512930, w.hooks())

	want := []string{
		"class:7ff71d949570",
		"flags:7ff71d949570",
		"update:7ff71d949570",
		"clear:7ff71d949570",
		"push:7ff71d949570:32",
		"store-u32:7ff740000020:0:0000000a",
		"push:7ff71d949570:10",
		"store-u32:7ff750000010:0:00000000",
		"roam-flags:7ff7200002ec",
		"store-low:7ff750000010:2:a5",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want %q", w.events, want)
	}
	if got := w.storedU32[monsterWanderHighReport512930][0]; got != 10 {
		t.Fatalf("REPORT arg 0 = %d, want 10", got)
	}
	if got := w.storedU32[monsterWanderHighRoam512930][0]; got != 0 {
		t.Fatalf("ROAM arg 0 = %d, want 0", got)
	}
	if got := w.storedLow[monsterWanderHighRoam512930][2]; got != 0xa5 {
		t.Fatalf("ROAM arg 2 low byte = %#02x, want 0xa5", got)
	}
}

func TestMonsterWander512930OracleConstants(t *testing.T) {
	if ai.ActionType(monsterWanderReportAction512930) != ai.ACTION_REPORT {
		t.Fatalf("action 32 = %s, want ACTION_REPORT", ai.ActionType(monsterWanderReportAction512930))
	}
	if ai.ActionType(monsterWanderRoamAction512930) != ai.ACTION_ROAM {
		t.Fatalf("action 10 = %s, want ACTION_ROAM", ai.ActionType(monsterWanderRoamAction512930))
	}
}

func TestMonsterWander512930PreservesOriginalBranches(t *testing.T) {
	tests := []struct {
		name         string
		classLow     uint8
		flags        uint32
		reportResult uint64
		roamResult   uint64
		wantEvents   []string
	}{
		{
			name:       "not monster",
			wantEvents: []string{"class:7ff71d949570"},
		},
		{
			name:     "blocked flag",
			classLow: 0x82,
			flags:    monsterWanderBlockedFlag512930,
			wantEvents: []string{
				"class:7ff71d949570", "flags:7ff71d949570",
			},
		},
		{
			name:       "report push failure continues",
			classLow:   0x02,
			roamResult: monsterWanderHighRoam512930,
			wantEvents: []string{
				"class:7ff71d949570", "flags:7ff71d949570",
				"update:7ff71d949570", "clear:7ff71d949570",
				"push:7ff71d949570:32", "push:7ff71d949570:10",
				"store-u32:7ff750000010:0:00000000",
				"roam-flags:7ff7200002ec", "store-low:7ff750000010:2:a5",
			},
		},
		{
			name:         "roam push failure stops",
			classLow:     0x02,
			reportResult: monsterWanderHighReport512930,
			wantEvents: []string{
				"class:7ff71d949570", "flags:7ff71d949570",
				"update:7ff71d949570", "clear:7ff71d949570",
				"push:7ff71d949570:32",
				"store-u32:7ff740000020:0:0000000a",
				"push:7ff71d949570:10",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := newMonsterWanderTestWorld512930()
			w.classLow = tc.classLow
			w.flags = tc.flags
			w.reportResult = tc.reportResult
			w.roamResult = tc.roamResult
			monsterWander512930(monsterWanderHighUnit512930, w.hooks())
			if !reflect.DeepEqual(w.events, tc.wantEvents) {
				t.Fatalf("events = %q, want %q", w.events, tc.wantEvents)
			}
		})
	}
}

func TestMonsterWander512930CachesUpdateBeforeStackCallbacks(t *testing.T) {
	w := newMonsterWanderTestWorld512930()
	w.mutateOnClear = true
	monsterWander512930(monsterWanderHighUnit512930, w.hooks())
	if got := w.storedLow[monsterWanderHighRoam512930][2]; got != 0xa5 {
		t.Fatalf("ROAM arg 2 low byte = %#02x, want cached UpdateData value 0xa5", got)
	}
	if got := w.events[len(w.events)-2]; got != "roam-flags:7ff7200002ec" {
		t.Fatalf("late UpdateData load = %q, want cached handle", got)
	}
}

func TestMonsterWander512930DoesNotAddNullUnitGuard(t *testing.T) {
	w := newMonsterWanderTestWorld512930()
	hooks := w.hooks()
	hooks.loadClassLow = func(unit uint64) uint8 {
		w.observe(fmt.Sprintf("class:%x", unit))
		panic("class load through null unit")
	}
	defer func() {
		if recover() == nil {
			t.Fatal("null unit did not preserve the original class-load fault")
		}
		if want := []string{"class:0"}; !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %q, want %q", w.events, want)
		}
	}()
	monsterWander512930(uint64(0), hooks)
}

func TestMonsterWander512930AllFaultPrefixes(t *testing.T) {
	baseline := newMonsterWanderTestWorld512930()
	monsterWander512930(monsterWanderHighUnit512930, baseline.hooks())
	want := append([]string(nil), baseline.events...)

	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			w := newMonsterWanderTestWorld512930()
			w.faultAt = faultAt
			var recovered any
			func() {
				defer func() { recovered = recover() }()
				monsterWander512930(monsterWanderHighUnit512930, w.hooks())
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
