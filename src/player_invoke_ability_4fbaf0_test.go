package opennox

import (
	"fmt"
	"math"
	"reflect"
	"testing"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/server"
)

type playerInvokeAbilityTestUnit4FBAF0 struct {
	name  string
	flags object.Flags
}

type playerInvokeAbilityTestWorld4FBAF0 struct {
	events   []string
	faultAt  int
	duration int32
}

func playerInvokeAbilityTestUnitName4FBAF0(unit *playerInvokeAbilityTestUnit4FBAF0) string {
	if unit == nil {
		return "nil"
	}
	return unit.name
}

func (w *playerInvokeAbilityTestWorld4FBAF0) event(value string) {
	w.events = append(w.events, value)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(value)
	}
}

func (w *playerInvokeAbilityTestWorld4FBAF0) hooks() playerInvokeAbilityHooks4FBAF0[*playerInvokeAbilityTestUnit4FBAF0] {
	return playerInvokeAbilityHooks4FBAF0[*playerInvokeAbilityTestUnit4FBAF0]{
		loadFlags: func(unit *playerInvokeAbilityTestUnit4FBAF0) object.Flags {
			name := playerInvokeAbilityTestUnitName4FBAF0(unit)
			if unit == nil {
				w.event("flags:" + name)
				return unit.flags
			}
			w.event(fmt.Sprintf("flags:%s=%08x", name, uint32(unit.flags)))
			return unit.flags
		},
		berserk: func(unit *playerInvokeAbilityTestUnit4FBAF0) {
			w.event("berserk:" + playerInvokeAbilityTestUnitName4FBAF0(unit))
		},
		warcry: func(unit *playerInvokeAbilityTestUnit4FBAF0) {
			w.event("warcry:" + playerInvokeAbilityTestUnitName4FBAF0(unit))
		},
		harpoon: func(unit *playerInvokeAbilityTestUnit4FBAF0) {
			w.event("harpoon:" + playerInvokeAbilityTestUnitName4FBAF0(unit))
		},
		loadDuration: func(ability server.Ability) int32 {
			w.event(fmt.Sprintf("duration:%d=%08x", ability, uint32(w.duration)))
			return w.duration
		},
		treadLightly: func(unit *playerInvokeAbilityTestUnit4FBAF0, duration int32) {
			w.event(fmt.Sprintf("tread-lightly:%s=%08x", playerInvokeAbilityTestUnitName4FBAF0(unit), uint32(duration)))
		},
		infravis: func(unit *playerInvokeAbilityTestUnit4FBAF0, duration int32) {
			w.event(fmt.Sprintf("infravis:%s=%08x", playerInvokeAbilityTestUnitName4FBAF0(unit), uint32(duration)))
		},
	}
}

func playerInvokeAbilityWantTrace4FBAF0(ability server.Ability, duration int32) []string {
	want := []string{"flags:unit=40004000"}
	switch ability {
	case server.AbilityBerserk:
		return append(want, "berserk:unit")
	case server.AbilityWarcry:
		return append(want, "warcry:unit")
	case server.AbilityHarpoon:
		return append(want, "harpoon:unit")
	case server.AbilityTreadLightly:
		return append(want,
			fmt.Sprintf("duration:4=%08x", uint32(duration)),
			fmt.Sprintf("tread-lightly:unit=%08x", uint32(duration)),
		)
	case server.AbilityInfravis:
		return append(want,
			fmt.Sprintf("duration:5=%08x", uint32(duration)),
			fmt.Sprintf("infravis:unit=%08x", uint32(duration)),
		)
	default:
		return want
	}
}

func TestPlayerInvokeAbility4FBAF0ExactSignedDispatchAndDurationOrder(t *testing.T) {
	duration := int32(math.MinInt32 + 0x123456)
	unit := &playerInvokeAbilityTestUnit4FBAF0{
		name:  "unit",
		flags: object.FlagAirborne | object.FlagSelected,
	}
	abilities := []server.Ability{
		server.Ability(math.MinInt32), -1, server.AbilityInvalid,
		server.AbilityBerserk, server.AbilityWarcry, server.AbilityHarpoon,
		server.AbilityTreadLightly, server.AbilityInfravis,
		server.AbilityMax, server.Ability(math.MaxInt32),
	}
	for _, ability := range abilities {
		t.Run(fmt.Sprintf("ability-%08x", uint32(ability)), func(t *testing.T) {
			world := &playerInvokeAbilityTestWorld4FBAF0{duration: duration}
			playerInvokeAbility4FBAF0(unit, ability, world.hooks())
			want := playerInvokeAbilityWantTrace4FBAF0(ability, duration)
			if !reflect.DeepEqual(world.events, want) {
				t.Fatalf("events = %v, want %v", world.events, want)
			}
		})
	}
}

func TestPlayerInvokeAbility4FBAF0DeadDestroyedGatePrecedesDispatch(t *testing.T) {
	for _, flags := range []object.Flags{
		object.FlagDestroyed,
		object.FlagDead,
		object.FlagDestroyed | object.FlagDead,
		object.FlagMarked | object.FlagDestroyed,
		object.FlagMarked | object.FlagDead,
	} {
		t.Run(fmt.Sprintf("flags-%08x", uint32(flags)), func(t *testing.T) {
			unit := &playerInvokeAbilityTestUnit4FBAF0{name: "unit", flags: flags}
			world := &playerInvokeAbilityTestWorld4FBAF0{duration: math.MaxInt32}
			playerInvokeAbility4FBAF0(unit, server.AbilityTreadLightly, world.hooks())
			want := []string{fmt.Sprintf("flags:unit=%08x", uint32(flags))}
			if !reflect.DeepEqual(world.events, want) {
				t.Fatalf("events = %v, want %v", world.events, want)
			}
		})
	}
}

func TestPlayerInvokeAbility4FBAF0EveryObservableFaultPrefix(t *testing.T) {
	duration := int32(math.MinInt32 + 0x654321)
	unit := &playerInvokeAbilityTestUnit4FBAF0{
		name:  "unit",
		flags: object.FlagAirborne | object.FlagSelected,
	}
	for _, ability := range []server.Ability{
		server.AbilityInvalid,
		server.AbilityBerserk,
		server.AbilityWarcry,
		server.AbilityHarpoon,
		server.AbilityTreadLightly,
		server.AbilityInfravis,
	} {
		want := playerInvokeAbilityWantTrace4FBAF0(ability, duration)
		for faultAt := 1; faultAt <= len(want); faultAt++ {
			t.Run(fmt.Sprintf("ability-%d-fault-%d", ability, faultAt), func(t *testing.T) {
				world := &playerInvokeAbilityTestWorld4FBAF0{duration: duration, faultAt: faultAt}
				defer func() {
					if got := recover(); got != want[faultAt-1] {
						t.Fatalf("panic = %v, want %q", got, want[faultAt-1])
					}
					if prefix := want[:faultAt]; !reflect.DeepEqual(world.events, prefix) {
						t.Fatalf("events = %v, want prefix %v", world.events, prefix)
					}
				}()
				playerInvokeAbility4FBAF0(unit, ability, world.hooks())
			})
		}
	}
}

func TestPlayerInvokeAbility4FBAF0NilUnitFaultsAtFirstFlagRead(t *testing.T) {
	world := &playerInvokeAbilityTestWorld4FBAF0{}
	defer func() {
		if recover() == nil {
			t.Fatal("nil unit did not fault at flag read")
		}
		want := []string{"flags:nil"}
		if !reflect.DeepEqual(world.events, want) {
			t.Fatalf("events = %v, want %v", world.events, want)
		}
	}()
	playerInvokeAbility4FBAF0(
		(*playerInvokeAbilityTestUnit4FBAF0)(nil),
		server.AbilityBerserk,
		world.hooks(),
	)
}
