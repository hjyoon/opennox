package server

import (
	"math"
	"reflect"
	"testing"
)

type playerQuestSpawnTestUnit4E8210 struct {
	name string
	gate *playerQuestSpawnTestGate4E8210
}

type playerQuestSpawnTestGate4E8210 struct {
	name string
	data *playerQuestSpawnTestData4E8210
	pos  playerQuestSpawnTestPoint4E8210
}

type playerQuestSpawnTestData4E8210 struct {
	name  string
	frame uint32
}

type playerQuestSpawnTestPoint4E8210 struct {
	x uint32
	y uint32
}

func TestPlayerQuestSpawn4E8210SelectsStrictUnsignedMaximum(t *testing.T) {
	zero := &playerQuestSpawnTestGate4E8210{
		name: "zero",
		data: &playerQuestSpawnTestData4E8210{name: "zero", frame: 0},
	}
	firstFive := &playerQuestSpawnTestGate4E8210{
		name: "five-a",
		data: &playerQuestSpawnTestData4E8210{name: "five-a", frame: 5},
	}
	secondFive := &playerQuestSpawnTestGate4E8210{
		name: "five-b",
		data: &playerQuestSpawnTestData4E8210{name: "five-b", frame: 5},
	}
	high := &playerQuestSpawnTestGate4E8210{
		name: "high",
		data: &playerQuestSpawnTestData4E8210{name: "high", frame: math.MaxUint32},
		pos:  playerQuestSpawnTestPoint4E8210{x: 0x80000000, y: 0x7fffffff},
	}
	highTie := &playerQuestSpawnTestGate4E8210{
		name: "high-tie",
		data: &playerQuestSpawnTestData4E8210{name: "high-tie", frame: math.MaxUint32},
	}
	units := []*playerQuestSpawnTestUnit4E8210{
		{name: "nil"},
		{name: "zero", gate: zero},
		{name: "five-a", gate: firstFive},
		{name: "five-b", gate: secondFive},
		{name: "high", gate: high},
		{name: "high-tie", gate: highTie},
	}
	joining := &playerQuestSpawnTestUnit4E8210{name: "joining", gate: zero}
	next := make(map[*playerQuestSpawnTestUnit4E8210]*playerQuestSpawnTestUnit4E8210)
	for i := 0; i+1 < len(units); i++ {
		next[units[i]] = units[i+1]
	}
	var events []string
	wantPoint := playerQuestSpawnTestPoint4E8210{x: 0x11223344, y: 0xaabbccdd}
	got, ok := playerQuestSpawn4E8210(joining,
		playerQuestSpawnHooks4E8210[*playerQuestSpawnTestUnit4E8210, *playerQuestSpawnTestGate4E8210, *playerQuestSpawnTestData4E8210, playerQuestSpawnTestPoint4E8210]{
			firstUnit: func() *playerQuestSpawnTestUnit4E8210 {
				events = append(events, "first")
				return units[0]
			},
			nextUnit: func(unit *playerQuestSpawnTestUnit4E8210) *playerQuestSpawnTestUnit4E8210 {
				events = append(events, "next:"+unit.name)
				return next[unit]
			},
			loadSoulGate: func(unit *playerQuestSpawnTestUnit4E8210) *playerQuestSpawnTestGate4E8210 {
				events = append(events, "gate:"+unit.name)
				return unit.gate
			},
			loadCollideData: func(gate *playerQuestSpawnTestGate4E8210) *playerQuestSpawnTestData4E8210 {
				events = append(events, "data:"+gate.name)
				return gate.data
			},
			loadLastUsedFrame: func(data *playerQuestSpawnTestData4E8210) uint32 {
				events = append(events, "frame:"+data.name)
				return data.frame
			},
			storeSoulGate: func(unit *playerQuestSpawnTestUnit4E8210, gate *playerQuestSpawnTestGate4E8210) {
				events = append(events, "store:"+unit.name+":"+gate.name)
				unit.gate = gate
			},
			loadSoulGatePos: func(gate *playerQuestSpawnTestGate4E8210) playerQuestSpawnTestPoint4E8210 {
				events = append(events, "pos:"+gate.name)
				return gate.pos
			},
			randomReachablePos: func(radius float32, pos playerQuestSpawnTestPoint4E8210) playerQuestSpawnTestPoint4E8210 {
				events = append(events, "random")
				if math.Float32bits(radius) != 0x42700000 {
					t.Fatalf("radius bits = %#x, want 0x42700000", math.Float32bits(radius))
				}
				if pos != high.pos {
					t.Fatalf("source position = %#v, want %#v", pos, high.pos)
				}
				return wantPoint
			},
		})
	if !ok || got != wantPoint || joining.gate != high {
		t.Fatalf("result/ok/gate = (%#v, %v, %p), want (%#v, true, %p)", got, ok, joining.gate, wantPoint, high)
	}
	wantEvents := []string{
		"first",
		"gate:nil", "next:nil",
		"gate:zero", "data:zero", "frame:zero", "next:zero",
		"gate:five-a", "data:five-a", "frame:five-a", "next:five-a",
		"gate:five-b", "data:five-b", "frame:five-b", "next:five-b",
		"gate:high", "data:high", "frame:high", "next:high",
		"gate:high-tie", "data:high-tie", "frame:high-tie", "next:high-tie",
		"store:joining:high", "pos:high", "random",
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
}

func TestPlayerQuestSpawn4E8210FailureDoesNotInspectJoiningPlayer(t *testing.T) {
	for _, tc := range []struct {
		name  string
		first func() *playerQuestSpawnTestUnit4E8210
	}{
		{name: "empty", first: func() *playerQuestSpawnTestUnit4E8210 { return nil }},
		{name: "frame zero", first: func() *playerQuestSpawnTestUnit4E8210 {
			return &playerQuestSpawnTestUnit4E8210{gate: &playerQuestSpawnTestGate4E8210{
				data: &playerQuestSpawnTestData4E8210{},
			}}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := playerQuestSpawn4E8210((*playerQuestSpawnTestUnit4E8210)(nil),
				playerQuestSpawnHooks4E8210[*playerQuestSpawnTestUnit4E8210, *playerQuestSpawnTestGate4E8210, *playerQuestSpawnTestData4E8210, playerQuestSpawnTestPoint4E8210]{
					firstUnit: tc.first,
					nextUnit:  func(*playerQuestSpawnTestUnit4E8210) *playerQuestSpawnTestUnit4E8210 { return nil },
					loadSoulGate: func(unit *playerQuestSpawnTestUnit4E8210) *playerQuestSpawnTestGate4E8210 {
						return unit.gate
					},
					loadCollideData: func(gate *playerQuestSpawnTestGate4E8210) *playerQuestSpawnTestData4E8210 {
						return gate.data
					},
					loadLastUsedFrame: func(data *playerQuestSpawnTestData4E8210) uint32 { return data.frame },
					storeSoulGate: func(*playerQuestSpawnTestUnit4E8210, *playerQuestSpawnTestGate4E8210) {
						t.Fatal("failure stored the joining SoulGate")
					},
					loadSoulGatePos: func(*playerQuestSpawnTestGate4E8210) playerQuestSpawnTestPoint4E8210 {
						t.Fatal("failure loaded a SoulGate position")
						return playerQuestSpawnTestPoint4E8210{}
					},
					randomReachablePos: func(float32, playerQuestSpawnTestPoint4E8210) playerQuestSpawnTestPoint4E8210 {
						t.Fatal("failure searched for a reachable point")
						return playerQuestSpawnTestPoint4E8210{}
					},
				})
			if ok || got != (playerQuestSpawnTestPoint4E8210{}) {
				t.Fatalf("result = (%#v, %v), want zero/false", got, ok)
			}
		})
	}
}

func TestPlayerQuestSpawn4E8210UsesLiveSuccessorAndPostStorePosition(t *testing.T) {
	gate1 := &playerQuestSpawnTestGate4E8210{name: "one", data: &playerQuestSpawnTestData4E8210{frame: 1}}
	gate2 := &playerQuestSpawnTestGate4E8210{name: "skipped", data: &playerQuestSpawnTestData4E8210{frame: 100}}
	gate3 := &playerQuestSpawnTestGate4E8210{name: "five", data: &playerQuestSpawnTestData4E8210{frame: 5}}
	u1 := &playerQuestSpawnTestUnit4E8210{name: "one", gate: gate1}
	u2 := &playerQuestSpawnTestUnit4E8210{name: "skipped", gate: gate2}
	u3 := &playerQuestSpawnTestUnit4E8210{name: "five", gate: gate3}
	joining := &playerQuestSpawnTestUnit4E8210{name: "joining"}
	next := map[*playerQuestSpawnTestUnit4E8210]*playerQuestSpawnTestUnit4E8210{u1: u2, u2: u3}
	postStore := playerQuestSpawnTestPoint4E8210{x: 17, y: 23}
	var visited []string
	got, ok := playerQuestSpawn4E8210(joining,
		playerQuestSpawnHooks4E8210[*playerQuestSpawnTestUnit4E8210, *playerQuestSpawnTestGate4E8210, *playerQuestSpawnTestData4E8210, playerQuestSpawnTestPoint4E8210]{
			firstUnit: func() *playerQuestSpawnTestUnit4E8210 { return u1 },
			nextUnit: func(unit *playerQuestSpawnTestUnit4E8210) *playerQuestSpawnTestUnit4E8210 {
				return next[unit]
			},
			loadSoulGate: func(unit *playerQuestSpawnTestUnit4E8210) *playerQuestSpawnTestGate4E8210 {
				visited = append(visited, unit.name)
				return unit.gate
			},
			loadCollideData: func(gate *playerQuestSpawnTestGate4E8210) *playerQuestSpawnTestData4E8210 {
				return gate.data
			},
			loadLastUsedFrame: func(data *playerQuestSpawnTestData4E8210) uint32 {
				if data == gate1.data {
					next[u1] = u3
				}
				return data.frame
			},
			storeSoulGate: func(unit *playerQuestSpawnTestUnit4E8210, gate *playerQuestSpawnTestGate4E8210) {
				unit.gate = gate
				gate.pos = postStore
			},
			loadSoulGatePos: func(gate *playerQuestSpawnTestGate4E8210) playerQuestSpawnTestPoint4E8210 {
				return gate.pos
			},
			randomReachablePos: func(_ float32, pos playerQuestSpawnTestPoint4E8210) playerQuestSpawnTestPoint4E8210 {
				return pos
			},
		})
	if !ok || got != postStore || joining.gate != gate3 {
		t.Fatalf("result/ok/gate = (%#v, %v, %p), want (%#v, true, %p)", got, ok, joining.gate, postStore, gate3)
	}
	if want := []string{"one", "five"}; !reflect.DeepEqual(visited, want) {
		t.Fatalf("visited = %v, want %v", visited, want)
	}
}

func TestPlayerQuestSpawn4E8210NilCollideDataFaultsBeforeNext(t *testing.T) {
	unit := &playerQuestSpawnTestUnit4E8210{gate: &playerQuestSpawnTestGate4E8210{}}
	nextCalls := 0
	defer func() {
		if recover() == nil {
			t.Fatal("nil SoulGate collide data did not fault")
		}
		if nextCalls != 0 {
			t.Fatalf("next calls = %d, want 0", nextCalls)
		}
	}()
	playerQuestSpawn4E8210(unit,
		playerQuestSpawnHooks4E8210[*playerQuestSpawnTestUnit4E8210, *playerQuestSpawnTestGate4E8210, *playerQuestSpawnTestData4E8210, playerQuestSpawnTestPoint4E8210]{
			firstUnit:       func() *playerQuestSpawnTestUnit4E8210 { return unit },
			nextUnit:        func(*playerQuestSpawnTestUnit4E8210) *playerQuestSpawnTestUnit4E8210 { nextCalls++; return nil },
			loadSoulGate:    func(unit *playerQuestSpawnTestUnit4E8210) *playerQuestSpawnTestGate4E8210 { return unit.gate },
			loadCollideData: func(gate *playerQuestSpawnTestGate4E8210) *playerQuestSpawnTestData4E8210 { return gate.data },
			loadLastUsedFrame: func(data *playerQuestSpawnTestData4E8210) uint32 {
				return data.frame
			},
		})
}
