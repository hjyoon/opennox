package server

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/opennox/libs/types"
)

type playerRespawnTestObject4F7EF0 string
type playerRespawnTestUpdate4F7EF0 string
type playerRespawnTestPlayer4F7EF0 string
type playerRespawnTestSettings4F7EF0 string
type playerRespawnTestTeam4F7EF0 string

type playerRespawnTestWorld4F7EF0 struct {
	events      []string
	flags       map[uint32]int32
	gameplay    map[uint32]int32
	blocker     uint32
	entryPlayer playerRespawnTestPlayer4F7EF0
	livePlayer  playerRespawnTestPlayer4F7EF0
	networkMode uint32
	skeletons   uint32
	soulGate    playerRespawnTestObject4F7EF0
	crown       playerRespawnTestObject4F7EF0
	crownHolder playerRespawnTestObject4F7EF0
	tickRate    uint32
	playerLoads int
	destination types.Pointf
}

func (w *playerRespawnTestWorld4F7EF0) event(format string, args ...any) {
	w.events = append(w.events, fmt.Sprintf(format, args...))
}

func (w *playerRespawnTestWorld4F7EF0) hooks() playerRespawnHooks4F7EF0[
	playerRespawnTestObject4F7EF0,
	playerRespawnTestUpdate4F7EF0,
	playerRespawnTestPlayer4F7EF0,
	playerRespawnTestSettings4F7EF0,
	playerRespawnTestTeam4F7EF0,
] {
	return playerRespawnHooks4F7EF0[
		playerRespawnTestObject4F7EF0,
		playerRespawnTestUpdate4F7EF0,
		playerRespawnTestPlayer4F7EF0,
		playerRespawnTestSettings4F7EF0,
		playerRespawnTestTeam4F7EF0,
	]{
		loadSettings: func() playerRespawnTestSettings4F7EF0 {
			w.event("settings")
			return "settings"
		},
		loadUnitArg: func() playerRespawnTestObject4F7EF0 {
			w.event("unit")
			return "unit"
		},
		loadUpdateData: func(unit playerRespawnTestObject4F7EF0) playerRespawnTestUpdate4F7EF0 {
			w.event("update:%s", unit)
			return "update"
		},
		loadPlayer: func(update playerRespawnTestUpdate4F7EF0) playerRespawnTestPlayer4F7EF0 {
			w.event("player:%s", update)
			w.playerLoads++
			if w.playerLoads == 1 {
				return w.entryPlayer
			}
			return w.livePlayer
		},
		gameFlag: func(flag uint32) int32 {
			w.event("game:%x", flag)
			return w.flags[flag]
		},
		loadQuestBlock: func(update playerRespawnTestUpdate4F7EF0) uint32 {
			w.event("blocker:%s", update)
			return w.blocker
		},
		storePlayerDone: func(player playerRespawnTestPlayer4F7EF0, value uint32) {
			w.event("done:%s=%d", player, value)
		},
		makeDefaultItems: func(unit playerRespawnTestObject4F7EF0, stats, keep int32) {
			w.event("items:%s,%d,%d", unit, stats, keep)
		},
		loadPlayerIndex: func(player playerRespawnTestPlayer4F7EF0) uint8 {
			w.event("index:%s", player)
			return 7
		},
		storeRespawnMarker: func(update playerRespawnTestUpdate4F7EF0, index, value uint8) {
			w.event("marker:%s,%d=%02x", update, index, value)
		},
		priorityMessage: func(unit playerRespawnTestObject4F7EF0, message string, value uint8) {
			w.event("message:%s,%s,%d", unit, message, value)
		},
		audio: func(id uint32, unit playerRespawnTestObject4F7EF0, first, second int32) {
			w.event("audio:%d,%s,%d,%d", id, unit, first, second)
		},
		loadNetworkMode: func() uint32 {
			w.event("network")
			return w.networkMode
		},
		loadSkeletonSetting: func(settings playerRespawnTestSettings4F7EF0) uint32 {
			w.event("skeletons:%s", settings)
			return w.skeletons
		},
		respawnCorpse: func(unit playerRespawnTestObject4F7EF0) {
			w.event("corpse:%s", unit)
		},
		loadPositionX: func(unit playerRespawnTestObject4F7EF0) float32 {
			w.event("posx:%s", unit)
			return 10
		},
		loadPositionY: func(unit playerRespawnTestObject4F7EF0) float32 {
			w.event("posy:%s", unit)
			return 20
		},
		loadSoulGate: func(update playerRespawnTestUpdate4F7EF0) playerRespawnTestObject4F7EF0 {
			w.event("gate:%s", update)
			return w.soulGate
		},
		soulGatePoint: func(gate playerRespawnTestObject4F7EF0, output *types.Pointf) {
			w.event("soul:%s,%.0f,%.0f", gate, output.X, output.Y)
			*output = types.Pointf{X: 30, Y: 40}
		},
		mapPlayerStart: func(output *types.Pointf, unit playerRespawnTestObject4F7EF0) {
			w.event("start:%s,%.0f,%.0f", unit, output.X, output.Y)
			*output = types.Pointf{X: 50, Y: 60}
		},
		move: func(unit playerRespawnTestObject4F7EF0, output *types.Pointf) {
			w.event("move:%s,%.0f,%.0f", unit, output.X, output.Y)
			w.destination = *output
		},
		gameplayFlag: func(flag uint32) int32 {
			w.event("play:%x", flag)
			return w.gameplay[flag]
		},
		loadPlayerUnit: func(player playerRespawnTestPlayer4F7EF0) playerRespawnTestObject4F7EF0 {
			w.event("playerUnit:%s", player)
			return "playerUnit"
		},
		loadTeamID: func(unit playerRespawnTestObject4F7EF0) uint8 {
			w.event("teamID:%s", unit)
			return 9
		},
		teamByID: func(id uint8) playerRespawnTestTeam4F7EF0 {
			w.event("team:%d", id)
			return "team"
		},
		loadTeamCrown: func(team playerRespawnTestTeam4F7EF0) playerRespawnTestObject4F7EF0 {
			w.event("crown:%s", team)
			return w.crown
		},
		loadInventoryHolder: func(crown playerRespawnTestObject4F7EF0) playerRespawnTestObject4F7EF0 {
			w.event("holder:%s", crown)
			return w.crownHolder
		},
		crownPickup: func(unit, crown playerRespawnTestObject4F7EF0, first, second int32) {
			w.event("pickup:%s,%s,%d,%d", unit, crown, first, second)
		},
		loadTickRate: func() uint32 {
			w.event("tick")
			return w.tickRate
		},
		applyBuff: func(unit playerRespawnTestObject4F7EF0, enchant uint32, duration uint16, power uint32) {
			w.event("buff:%s,%d,%d,%d", unit, enchant, duration, power)
		},
	}
}

func TestPlayerRespawn4F7EF0QuestOrderReloadsAndResult(t *testing.T) {
	w := &playerRespawnTestWorld4F7EF0{
		flags: map[uint32]int32{
			playerRespawnQuestFlag4F7EF0:      1,
			playerRespawnCrownFlag4F7EF0:      1,
			playerRespawnProtectionFlag4F7EF0: 1,
		},
		gameplay:    map[uint32]int32{playerRespawnCrownPlayFlag4F7EF0: 1},
		entryPlayer: "entry",
		livePlayer:  "live",
		networkMode: 1,
		skeletons:   1,
		soulGate:    "gate",
		crown:       "crown",
		tickRate:    13,
	}
	result := playerRespawn4F7EF0(w.hooks())
	if result.kind != playerRespawnResultScalar4F7EF0 || result.value != 1 {
		t.Fatalf("unexpected result: %+v", result)
	}
	want := []string{
		"settings", "unit", "update:unit", "player:update", "game:1000", "blocker:update",
		"done:entry=0", "game:1000", "items:unit,1,1", "player:update", "index:live",
		"marker:update,7=fa", "message:unit,GeneralPrint:Respawn,0", "game:1000",
		"audio:1006,unit,0,0", "network", "skeletons:settings", "corpse:unit",
		"posx:unit", "posy:unit", "game:1000", "gate:update", "soul:gate,10,20",
		"move:unit,30,40", "game:10", "play:4", "playerUnit:entry", "teamID:playerUnit",
		"team:9", "crown:team", "holder:crown", "playerUnit:entry",
		"pickup:playerUnit,crown,1,1", "game:2000", "tick", "buff:unit,23,65,5",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("event order mismatch:\n got: %q\nwant: %q", w.events, want)
	}
}

func TestPlayerRespawn4F7EF0EarlyResultsAndShortCircuit(t *testing.T) {
	t.Run("nil unit keeps settings result", func(t *testing.T) {
		w := &playerRespawnTestWorld4F7EF0{}
		h := w.hooks()
		h.loadUnitArg = func() playerRespawnTestObject4F7EF0 {
			w.event("unit")
			return ""
		}
		result := playerRespawn4F7EF0(h)
		if result.kind != playerRespawnResultSettings4F7EF0 || result.settings != "settings" {
			t.Fatalf("unexpected result: %+v", result)
		}
		if want := []string{"settings", "unit"}; !reflect.DeepEqual(w.events, want) {
			t.Fatalf("got %q, want %q", w.events, want)
		}
	})

	t.Run("quest blocker is returned", func(t *testing.T) {
		w := &playerRespawnTestWorld4F7EF0{
			flags:       map[uint32]int32{playerRespawnQuestFlag4F7EF0: 1},
			entryPlayer: "entry",
			blocker:     0xfedcba98,
		}
		result := playerRespawn4F7EF0(w.hooks())
		if result.kind != playerRespawnResultScalar4F7EF0 || result.value != w.blocker {
			t.Fatalf("unexpected result: %+v", result)
		}
		want := []string{"settings", "unit", "update:unit", "player:update", "game:1000", "blocker:update"}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("got %q, want %q", w.events, want)
		}
	})
}

func TestPlayerRespawn4F7EF0NonQuestNetworkShortCircuit(t *testing.T) {
	w := &playerRespawnTestWorld4F7EF0{
		flags:       make(map[uint32]int32),
		gameplay:    make(map[uint32]int32),
		entryPlayer: "entry",
		livePlayer:  "live",
		networkMode: 0,
	}
	result := playerRespawn4F7EF0(w.hooks())
	if result.value != 0 {
		t.Fatalf("unexpected result: %+v", result)
	}
	for _, event := range w.events {
		if event == "skeletons:settings" || event[:min(len(event), len("gate:"))] == "gate:" {
			t.Fatalf("unexpected short-circuited event %q in %q", event, w.events)
		}
	}
	if w.destination != (types.Pointf{X: 50, Y: 60}) {
		t.Fatalf("destination = %+v", w.destination)
	}
}

func TestSoulGateRespawnPoint4F80C0AttemptsAndLiveGate(t *testing.T) {
	type gate struct{ position types.Pointf }
	g := &gate{position: types.Pointf{X: 1, Y: 2}}
	var output types.Pointf
	var randomCalls, allowCalls int
	result := soulGateRespawnPoint4F80C0(g, &output, soulGateRespawnPointHooks4F80C0[*gate]{
		loadPosition: func(gate *gate) types.Pointf { return gate.position },
		randomPoint: func(radius float32, gate *gate, output *types.Pointf) {
			randomCalls++
			if radius != 60 {
				t.Fatalf("radius = %v", radius)
			}
			gate.position.X++
			*output = gate.position
		},
		allowTeleport: func(*types.Pointf) int32 {
			allowCalls++
			return 1
		},
	})
	if result != 1 || randomCalls != 32 || allowCalls != 32 {
		t.Fatalf("result=%d random=%d allow=%d", result, randomCalls, allowCalls)
	}
	if output != (types.Pointf{X: 33, Y: 2}) {
		t.Fatalf("output = %+v", output)
	}

	randomCalls, allowCalls = 0, 0
	result = soulGateRespawnPoint4F80C0(g, &output, soulGateRespawnPointHooks4F80C0[*gate]{
		loadPosition: func(gate *gate) types.Pointf { return gate.position },
		randomPoint: func(_ float32, _ *gate, output *types.Pointf) {
			randomCalls++
			output.Y++
		},
		allowTeleport: func(*types.Pointf) int32 {
			allowCalls++
			if allowCalls == 3 {
				return 0
			}
			return 7
		},
	})
	if result != 0 || randomCalls != 3 || allowCalls != 3 {
		t.Fatalf("early result=%d random=%d allow=%d", result, randomCalls, allowCalls)
	}
}

func TestRespawnPlayerImpl53FBC0OrderFlagsAndNilBreak(t *testing.T) {
	var events []string
	flags := map[int]uint32{1: 0x12345600, 2: 0xabcdef01}
	centers := []types.Pointf{{X: 100, Y: 200}, {X: 300, Y: 400}}
	center := 0
	respawnPlayerImpl53FBC0(int16(-123), respawnPlayerImplHooks53FBC0[int]{
		loadInitialized: func() uint32 { events = append(events, "initialized"); return 0 },
		initialize:      func() { events = append(events, "initialize") },
		directionIndex: func(direction int16) int32 {
			events = append(events, fmt.Sprintf("direction:%d", direction))
			return 6
		},
		loadTypeIndex: func(direction int32, part int) uint32 {
			events = append(events, fmt.Sprintf("type:%d,%d", direction, part))
			return uint32(40 + part)
		},
		newObject: func(typeIndex uint32) int {
			events = append(events, fmt.Sprintf("new:%d", typeIndex))
			if typeIndex == 42 {
				return 0
			}
			return int(typeIndex - 39)
		},
		loadNetworkMode: func() uint32 { events = append(events, "network"); return 1 },
		gameFlag: func(flag uint32) int32 {
			events = append(events, fmt.Sprintf("game:%x", flag))
			return 1
		},
		loadObjectFlags: func(object int) uint32 {
			events = append(events, fmt.Sprintf("flags:%d", object))
			return flags[object]
		},
		storeObjectFlags: func(object int, value uint32) {
			events = append(events, fmt.Sprintf("storeFlags:%d=%08x", object, value))
			flags[object] = value
		},
		loadOffsetY: func(direction int32, part int) float32 {
			events = append(events, fmt.Sprintf("offY:%d,%d", direction, part))
			return float32(part + 1)
		},
		loadCenterY: func() float32 {
			events = append(events, "centerY")
			return centers[center].Y
		},
		loadOffsetX: func(direction int32, part int) float32 {
			events = append(events, fmt.Sprintf("offX:%d,%d", direction, part))
			return float32(part + 10)
		},
		loadCenterX: func() float32 {
			events = append(events, "centerX")
			return centers[center].X
		},
		createAt: func(object int, position types.Pointf) {
			events = append(events, fmt.Sprintf("create:%d,%.0f,%.0f", object, position.X, position.Y))
			center++
		},
		randomInt: func(minimum, maximum int32) int32 {
			events = append(events, fmt.Sprintf("random:%d,%d", minimum, maximum))
			return int32(10 + center)
		},
		loadTickRate: func() uint32 { events = append(events, "tick"); return 30 },
		setDecay: func(object int, delay uint32) {
			events = append(events, fmt.Sprintf("decay:%d,%d", object, delay))
		},
	})
	want := []string{
		"initialized", "initialize", "direction:-123",
		"type:6,0", "new:40", "network", "game:2000", "flags:1", "storeFlags:1=12345640",
		"offY:6,0", "centerY", "offX:6,0", "centerX", "create:1,110,201",
		"random:10,20", "tick", "decay:1,330",
		"type:6,1", "new:41", "network", "game:2000", "flags:2", "storeFlags:2=abcdef41",
		"offY:6,1", "centerY", "offX:6,1", "centerX", "create:2,311,402",
		"random:10,20", "tick", "decay:2,360",
		"type:6,2", "new:42",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("event order mismatch:\n got: %q\nwant: %q", events, want)
	}
}
