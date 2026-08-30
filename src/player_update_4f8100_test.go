package opennox

import (
	"math"
	"reflect"
	"testing"

	"github.com/opennox/opennox/v1/common/sound"
)

func TestPlayerUpdateHurt4F8100ReloadsPlayerAndPreservesLoadOrder(t *testing.T) {
	type testPlayer struct{ name string }
	stale := &testPlayer{name: "stale"}
	live := &testPlayer{name: "live"}
	player := stale
	damageType := uint32(5)
	var calls []string
	var gotSound sound.ID

	// Models NeedSync mutating the cached update record before the hurt branch.
	player = live
	playerUpdateHurt4F8100(71, playerUpdateHurtHooks4F8100[*testPlayer]{
		loadPlayer: func() *testPlayer {
			calls = append(calls, "player:"+player.name)
			return player
		},
		isFemale: func(got *testPlayer) bool {
			calls = append(calls, "female:"+got.name)
			// Proves Field131 is loaded after the gender byte.
			damageType = 0
			return true
		},
		loadDamageType: func() uint32 {
			calls = append(calls, "damage-type")
			return damageType
		},
		audio: func(id sound.ID) {
			calls = append(calls, "audio")
			gotSound = id
		},
	})

	wantCalls := []string{"player:live", "female:live", "damage-type", "audio"}
	if !reflect.DeepEqual(calls, wantCalls) {
		t.Fatalf("calls = %v, want %v", calls, wantCalls)
	}
	if gotSound != sound.SoundHumanFemaleHurtMedium {
		t.Fatalf("sound = %v, want female medium", gotSound)
	}
}

func TestPlayerUpdateHurt4F8100NonPositiveSkipsLoads(t *testing.T) {
	for _, tc := range []struct {
		name   string
		damage int32
	}{{name: "zero"}, {name: "negative", damage: -1}} {
		t.Run(tc.name, func(t *testing.T) {
			called := false
			playerUpdateHurt4F8100(tc.damage, playerUpdateHurtHooks4F8100[int]{
				loadPlayer: func() int { called = true; return 0 },
			})
			if called {
				t.Fatalf("damage %d loaded player", tc.damage)
			}
		})
	}
}

func TestPlayerUpdateHurtSound4F8100(t *testing.T) {
	tests := []struct {
		name       string
		damage     int32
		damageType uint32
		female     bool
		want       sound.ID
	}{
		{name: "male light boundary", damage: 70, want: sound.SoundHumanMaleHurtLight},
		{name: "male medium low", damage: 71, want: sound.SoundHumanMaleHurtMedium},
		{name: "male medium high", damage: 450, want: sound.SoundHumanMaleHurtMedium},
		{name: "male heavy", damage: 451, want: sound.SoundHumanMaleHurtHeavy},
		{name: "male poison overrides heavy", damage: 451, damageType: 5, want: sound.SoundHumanMaleHurtPoison},
		{name: "female light boundary", damage: 70, female: true, want: sound.SoundHumanFemaleHurtLight},
		{name: "female medium low", damage: 71, female: true, want: sound.SoundHumanFemaleHurtMedium},
		{name: "female medium high", damage: 450, female: true, want: sound.SoundHumanFemaleHurtMedium},
		{name: "female heavy", damage: 451, female: true, want: sound.SoundHumanFemaleHurtHeavy},
		{name: "female poison overrides heavy", damage: 451, damageType: 5, female: true, want: sound.SoundHumanFemaleHurtPoison},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := playerUpdateHurtSound4F8100(tc.damage, tc.damageType, tc.female); got != tc.want {
				t.Fatalf("sound = %v, want %v", got, tc.want)
			}
		})
	}
}

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
