package legacy

import (
	"fmt"
	"reflect"
	"testing"
)

type scriptHostStatusTestUpdate5166A0 struct {
	active bool
}

type scriptHostStatusTestUnit5166A0 struct {
	update *scriptHostStatusTestUpdate5166A0
}

type scriptHostStatusTestPlayer5166A0 struct {
	unit *scriptHostStatusTestUnit5166A0
}

type scriptHostStatusTestWorld5166A0 struct {
	player *scriptHostStatusTestPlayer5166A0
	events []string
	pushed []int32
	fault  string
}

func (w *scriptHostStatusTestWorld5166A0) event(name string) {
	w.events = append(w.events, name)
	if w.fault == name {
		panic(name)
	}
}

func (w *scriptHostStatusTestWorld5166A0) deps() scriptHostStatusDeps5166A0[
	*scriptHostStatusTestPlayer5166A0,
	*scriptHostStatusTestUnit5166A0,
	*scriptHostStatusTestUpdate5166A0,
] {
	return scriptHostStatusDeps5166A0[
		*scriptHostStatusTestPlayer5166A0,
		*scriptHostStatusTestUnit5166A0,
		*scriptHostStatusTestUpdate5166A0,
	]{
		hostPlayer: func() *scriptHostStatusTestPlayer5166A0 {
			w.event("host")
			return w.player
		},
		playerIsNil: func(player *scriptHostStatusTestPlayer5166A0) bool {
			return player == nil
		},
		loadUnit: func(player *scriptHostStatusTestPlayer5166A0) *scriptHostStatusTestUnit5166A0 {
			w.event("unit")
			if player == nil {
				panic("nil player")
			}
			return player.unit
		},
		loadUpdate: func(unit *scriptHostStatusTestUnit5166A0) *scriptHostStatusTestUpdate5166A0 {
			w.event("update")
			if unit == nil {
				panic("nil unit")
			}
			return unit.update
		},
		loadStateNonzero: func(update *scriptHostStatusTestUpdate5166A0) bool {
			w.event("state")
			if update == nil {
				panic("nil update")
			}
			return update.active
		},
		push: func(value int32) {
			w.event("push")
			w.pushed = append(w.pushed, value)
		},
	}
}

func TestScriptHostStatus5166A0CanonicalResults(t *testing.T) {
	for _, tc := range []struct {
		name   string
		player *scriptHostStatusTestPlayer5166A0
		want   int32
		events []string
	}{
		{
			name:   "inactive host",
			want:   0,
			events: []string{"host", "push"},
		},
		{
			name: "nil state",
			player: &scriptHostStatusTestPlayer5166A0{
				unit: &scriptHostStatusTestUnit5166A0{
					update: &scriptHostStatusTestUpdate5166A0{},
				},
			},
			want:   0,
			events: []string{"host", "unit", "update", "state", "push"},
		},
		{
			name: "non-nil state",
			player: &scriptHostStatusTestPlayer5166A0{
				unit: &scriptHostStatusTestUnit5166A0{
					update: &scriptHostStatusTestUpdate5166A0{active: true},
				},
			},
			want:   1,
			events: []string{"host", "unit", "update", "state", "push"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := &scriptHostStatusTestWorld5166A0{player: tc.player}
			if got := scriptHostStatus5166A0(w.deps()); got != 0 {
				t.Fatalf("result = %d, want canonical zero", got)
			}
			if !reflect.DeepEqual(w.pushed, []int32{tc.want}) {
				t.Fatalf("pushed = %v, want [%d]", w.pushed, tc.want)
			}
			if !reflect.DeepEqual(w.events, tc.events) {
				t.Fatalf("events = %v, want %v", w.events, tc.events)
			}
		})
	}
}

func TestScriptHostStatus5166A0CachesEachPointerOnce(t *testing.T) {
	firstUpdate := &scriptHostStatusTestUpdate5166A0{active: true}
	secondUpdate := &scriptHostStatusTestUpdate5166A0{}
	firstUnit := &scriptHostStatusTestUnit5166A0{update: firstUpdate}
	secondUnit := &scriptHostStatusTestUnit5166A0{update: secondUpdate}
	player := &scriptHostStatusTestPlayer5166A0{unit: firstUnit}
	var events []string
	var pushed []int32

	deps := scriptHostStatusDeps5166A0[
		*scriptHostStatusTestPlayer5166A0,
		*scriptHostStatusTestUnit5166A0,
		*scriptHostStatusTestUpdate5166A0,
	]{
		hostPlayer: func() *scriptHostStatusTestPlayer5166A0 {
			events = append(events, "host")
			return player
		},
		playerIsNil: func(player *scriptHostStatusTestPlayer5166A0) bool {
			return player == nil
		},
		loadUnit: func(got *scriptHostStatusTestPlayer5166A0) *scriptHostStatusTestUnit5166A0 {
			events = append(events, "unit")
			unit := got.unit
			got.unit = secondUnit
			return unit
		},
		loadUpdate: func(got *scriptHostStatusTestUnit5166A0) *scriptHostStatusTestUpdate5166A0 {
			events = append(events, "update")
			update := got.update
			got.update = secondUpdate
			return update
		},
		loadStateNonzero: func(got *scriptHostStatusTestUpdate5166A0) bool {
			events = append(events, "state")
			return got.active
		},
		push: func(value int32) {
			events = append(events, "push")
			pushed = append(pushed, value)
		},
	}
	if got := scriptHostStatus5166A0(deps); got != 0 {
		t.Fatalf("result = %d, want canonical zero", got)
	}
	if !reflect.DeepEqual(pushed, []int32{1}) {
		t.Fatalf("pushed = %v, want [1] from the cached chain", pushed)
	}
	wantEvents := []string{"host", "unit", "update", "state", "push"}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
}

func TestScriptHostStatus5166A0FaultPrefixes(t *testing.T) {
	wantEvents := []string{"host", "unit", "update", "state", "push"}
	for i, fault := range wantEvents {
		t.Run(fmt.Sprintf("%d_%s", i, fault), func(t *testing.T) {
			update := &scriptHostStatusTestUpdate5166A0{active: true}
			unit := &scriptHostStatusTestUnit5166A0{update: update}
			player := &scriptHostStatusTestPlayer5166A0{unit: unit}
			w := &scriptHostStatusTestWorld5166A0{player: player, fault: fault}

			var recovered any
			func() {
				defer func() { recovered = recover() }()
				_ = scriptHostStatus5166A0(w.deps())
			}()
			if recovered == nil {
				t.Fatal("expected fault")
			}
			if want := wantEvents[:i+1]; !reflect.DeepEqual(w.events, want) {
				t.Fatalf("events = %v, want prefix %v", w.events, want)
			}
			if len(w.pushed) != 0 {
				t.Fatalf("pushed after fault = %v, want none", w.pushed)
			}
		})
	}
}

func TestScriptHostStatus5166A0DoesNotGuardUnitOrUpdate(t *testing.T) {
	for _, tc := range []struct {
		name   string
		player *scriptHostStatusTestPlayer5166A0
		events []string
	}{
		{
			name:   "nil unit",
			player: &scriptHostStatusTestPlayer5166A0{},
			events: []string{"host", "unit", "update"},
		},
		{
			name: "nil update",
			player: &scriptHostStatusTestPlayer5166A0{
				unit: &scriptHostStatusTestUnit5166A0{},
			},
			events: []string{"host", "unit", "update", "state"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := &scriptHostStatusTestWorld5166A0{player: tc.player}
			var recovered any
			func() {
				defer func() { recovered = recover() }()
				_ = scriptHostStatus5166A0(w.deps())
			}()
			if recovered == nil {
				t.Fatal("expected original eager-dereference fault")
			}
			if !reflect.DeepEqual(w.events, tc.events) {
				t.Fatalf("events = %v, want %v", w.events, tc.events)
			}
			if len(w.pushed) != 0 {
				t.Fatalf("pushed after fault = %v, want none", w.pushed)
			}
		})
	}
}
