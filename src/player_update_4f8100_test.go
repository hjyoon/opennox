package opennox

import (
	"math"
	"reflect"
	"testing"
)

type playerUpdateHarpoonTarget4F8100 struct {
	name      string
	destroyed bool
}

func TestPlayerUpdateHarpoon4F8100NilTargetStopsAtFirstLoad(t *testing.T) {
	var calls []string
	playerUpdateHarpoon4F8100(playerUpdateHarpoonHooks4F8100[*playerUpdateHarpoonTarget4F8100]{
		loadTarget: func() *playerUpdateHarpoonTarget4F8100 {
			calls = append(calls, "target")
			return nil
		},
		loadForce: func() float64 {
			calls = append(calls, "force")
			return 1
		},
	})
	if want := []string{"target"}; !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %v, want %v", calls, want)
	}
}

func TestPlayerUpdateHarpoon4F8100ReloadsTargetAroundCallbacks(t *testing.T) {
	first := &playerUpdateHarpoonTarget4F8100{name: "first"}
	second := &playerUpdateHarpoonTarget4F8100{name: "second"}
	third := &playerUpdateHarpoonTarget4F8100{name: "third"}
	current := first
	var calls []string
	var applied *playerUpdateHarpoonTarget4F8100
	var appliedForce float64

	playerUpdateHarpoon4F8100(playerUpdateHarpoonHooks4F8100[*playerUpdateHarpoonTarget4F8100]{
		loadTarget: func() *playerUpdateHarpoonTarget4F8100 {
			calls = append(calls, "target:"+current.name)
			return current
		},
		loadForce: func() float64 {
			calls = append(calls, "force")
			current = second
			return math.Float64frombits(0x3ff0000010000000)
		},
		destroyed: func(target *playerUpdateHarpoonTarget4F8100) bool {
			calls = append(calls, "destroyed:"+target.name)
			return target.destroyed
		},
		breakOwner: func() {
			calls = append(calls, "break")
		},
		attribution: func(target *playerUpdateHarpoonTarget4F8100) {
			calls = append(calls, "attribution:"+target.name)
			current = third
		},
		applyForce: func(target *playerUpdateHarpoonTarget4F8100, force float64) {
			calls = append(calls, "apply:"+target.name)
			applied = target
			appliedForce = force
		},
	})

	wantCalls := []string{
		"target:first",
		"force",
		"target:second",
		"destroyed:second",
		"attribution:second",
		"target:third",
		"apply:third",
	}
	if !reflect.DeepEqual(calls, wantCalls) {
		t.Fatalf("calls = %v, want %v", calls, wantCalls)
	}
	if applied != third {
		t.Fatalf("applied target = %p, want third %p", applied, third)
	}
	wantForce := -float64(float32(math.Float64frombits(0x3ff0000010000000)))
	if appliedForce != wantForce {
		t.Fatalf("applied force = %.17g, want rounded %.17g", appliedForce, wantForce)
	}
}

func TestPlayerUpdateHarpoon4F8100DestroyedReloadBreaks(t *testing.T) {
	first := &playerUpdateHarpoonTarget4F8100{name: "first"}
	destroyed := &playerUpdateHarpoonTarget4F8100{name: "destroyed", destroyed: true}
	current := first
	var calls []string

	playerUpdateHarpoon4F8100(playerUpdateHarpoonHooks4F8100[*playerUpdateHarpoonTarget4F8100]{
		loadTarget: func() *playerUpdateHarpoonTarget4F8100 {
			calls = append(calls, "target:"+current.name)
			return current
		},
		loadForce: func() float64 {
			calls = append(calls, "force")
			current = destroyed
			return 2
		},
		destroyed: func(target *playerUpdateHarpoonTarget4F8100) bool {
			calls = append(calls, "destroyed:"+target.name)
			return target.destroyed
		},
		breakOwner: func() {
			calls = append(calls, "break")
		},
		attribution: func(target *playerUpdateHarpoonTarget4F8100) {
			calls = append(calls, "attribution:"+target.name)
		},
		applyForce: func(target *playerUpdateHarpoonTarget4F8100, _ float64) {
			calls = append(calls, "apply:"+target.name)
		},
	})

	want := []string{"target:first", "force", "target:destroyed", "destroyed:destroyed", "break"}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %v, want %v", calls, want)
	}
}

func TestPlayerUpdateHarpoon4F8100DoesNotGuardPostAttributionTarget(t *testing.T) {
	target := &playerUpdateHarpoonTarget4F8100{name: "target"}
	current := target
	var applied *playerUpdateHarpoonTarget4F8100
	playerUpdateHarpoon4F8100(playerUpdateHarpoonHooks4F8100[*playerUpdateHarpoonTarget4F8100]{
		loadTarget: func() *playerUpdateHarpoonTarget4F8100 { return current },
		loadForce:  func() float64 { return 1 },
		destroyed:  func(*playerUpdateHarpoonTarget4F8100) bool { return false },
		breakOwner: func() {},
		attribution: func(*playerUpdateHarpoonTarget4F8100) {
			current = nil
		},
		applyForce: func(target *playerUpdateHarpoonTarget4F8100, _ float64) {
			applied = target
		},
	})
	if applied != nil {
		t.Fatalf("post-attribution target = %p, want unguarded nil", applied)
	}
}
