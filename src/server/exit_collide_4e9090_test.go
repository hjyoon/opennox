package server

import (
	"fmt"
	"reflect"
	"slices"
	"testing"
)

type exitCollideTestMap4E9090 struct {
	name string
}

type exitCollideTestPlayer4E9090 struct {
	name  string
	class uint8
	index uint8
	stage uint32
}

type exitCollideTestUpdate4E9090 struct {
	name   string
	player *exitCollideTestPlayer4E9090
	exit   *exitCollideTestObject4E9090
	warp   *exitCollideTestObject4E9090
	traps  uint8
}

type exitCollideTestObject4E9090 struct {
	name       string
	class      uint8
	subclass   uint8
	update     *exitCollideTestUpdate4E9090
	mapName    *exitCollideTestMap4E9090
	firstOwned *exitCollideTestObject4E9090
	nextOwned  *exitCollideTestObject4E9090
	nextPlayer *exitCollideTestObject4E9090
	typeIndex  uint16
	holder     *exitCollideTestObject4E9090
}

type exitCollideTestState4E9090 struct {
	events           []string
	flags            map[uint32]int32
	glyphType        uint32
	warpEnabled      int32
	saveBusy         int32
	exitAllowed      int32
	paused           int32
	questMap         *exitCollideTestMap4E9090
	abilities        map[uint32]int32
	currentStage     uint32
	nextThreshold    uint32
	allExited        int32
	maybeWarp        int32
	frame            uint32
	firstPlayer      *exitCollideTestObject4E9090
	copiedMap        *exitCollideTestMap4E9090
	loadedMap        *exitCollideTestMap4E9090
	onDelete         func(*exitCollideTestObject4E9090)
	onStoreStage     func(*exitCollideTestPlayer4E9090, uint32)
	onAudio          func(*exitCollideTestObject4E9090)
	onSendUnit       func(*exitCollideTestObject4E9090)
	onRecordProgress func(*exitCollideTestObject4E9090)
}

func (s *exitCollideTestState4E9090) event(format string, args ...any) {
	s.events = append(s.events, fmt.Sprintf(format, args...))
}

func (s *exitCollideTestState4E9090) hooks() exitCollideHooks4E9090[
	*exitCollideTestObject4E9090,
	*exitCollideTestUpdate4E9090,
	*exitCollideTestPlayer4E9090,
	*exitCollideTestMap4E9090,
	string,
] {
	return exitCollideHooks4E9090[
		*exitCollideTestObject4E9090,
		*exitCollideTestUpdate4E9090,
		*exitCollideTestPlayer4E9090,
		*exitCollideTestMap4E9090,
		string,
	]{
		glyphType: func() uint32 {
			s.event("glyph")
			return s.glyphType
		},
		loadClassByte: func(obj *exitCollideTestObject4E9090) uint8 {
			s.event("class:%s", obj.name)
			return obj.class
		},
		defaultCollide: func(exit, unit *exitCollideTestObject4E9090, collision string) {
			s.event("default:%s:%s:%s", exit.name, unit.name, collision)
		},
		loadUpdateData: func(obj *exitCollideTestObject4E9090) *exitCollideTestUpdate4E9090 {
			s.event("update:%s", obj.name)
			return obj.update
		},
		loadMap: func(obj *exitCollideTestObject4E9090) *exitCollideTestMap4E9090 {
			s.event("map:%s", obj.name)
			return obj.mapName
		},
		gameFlag: func(flag uint32) int32 {
			s.event("flag:%x", flag)
			return s.flags[flag]
		},
		loadQuestExit: func(update *exitCollideTestUpdate4E9090) *exitCollideTestObject4E9090 {
			s.event("quest-exit:%s", update.name)
			return update.exit
		},
		loadQuestWarpGate: func(update *exitCollideTestUpdate4E9090) *exitCollideTestObject4E9090 {
			s.event("quest-warp:%s", update.name)
			return update.warp
		},
		loadSubclassByte: func(obj *exitCollideTestObject4E9090) uint8 {
			s.event("subclass:%s", obj.name)
			return obj.subclass
		},
		warpEnabled: func() int32 {
			s.event("warp-enabled")
			return s.warpEnabled
		},
		saveBusy: func() int32 {
			s.event("save-busy")
			return s.saveBusy
		},
		exitAllowed: func(obj *exitCollideTestObject4E9090) int32 {
			s.event("allowed:%s", obj.name)
			return s.exitAllowed
		},
		paused: func() int32 {
			s.event("paused")
			return s.paused
		},
		mapFirstByte: func(name *exitCollideTestMap4E9090) byte {
			s.event("map-first:%s", name.name)
			if name.name == "" {
				return 0
			}
			return name.name[0]
		},
		loadPlayer: func(update *exitCollideTestUpdate4E9090) *exitCollideTestPlayer4E9090 {
			s.event("player:%s:%s", update.name, update.player.name)
			return update.player
		},
		loadPlayerClass: func(player *exitCollideTestPlayer4E9090) uint8 {
			s.event("player-class:%s", player.name)
			return player.class
		},
		firstOwned: func(obj *exitCollideTestObject4E9090) *exitCollideTestObject4E9090 {
			s.event("first-owned:%s", obj.name)
			return obj.firstOwned
		},
		nextOwned: func(obj *exitCollideTestObject4E9090) *exitCollideTestObject4E9090 {
			s.event("next-owned:%s", obj.name)
			return obj.nextOwned
		},
		loadTypeIndex: func(obj *exitCollideTestObject4E9090) uint16 {
			s.event("type:%s", obj.name)
			return obj.typeIndex
		},
		loadInventoryHolder: func(obj *exitCollideTestObject4E9090) *exitCollideTestObject4E9090 {
			s.event("holder:%s", obj.name)
			return obj.holder
		},
		delayedDelete: func(obj *exitCollideTestObject4E9090) {
			s.event("delete:%s", obj.name)
			if s.onDelete != nil {
				s.onDelete(obj)
			}
		},
		loadCurTrapsByte: func(update *exitCollideTestUpdate4E9090) uint8 {
			s.event("traps:%s:%d", update.name, update.traps)
			return update.traps
		},
		storeCurTrapsByte: func(update *exitCollideTestUpdate4E9090, value uint8) {
			s.event("store-traps:%s:%d", update.name, value)
			update.traps = value
		},
		setMapLoadRequired: func(value int32) {
			s.event("map-required:%d", value)
		},
		setSaveFileName: func(name string) {
			s.event("save-name:%s", name)
		},
		saveCoop: func(enabled int32, obj *exitCollideTestObject4E9090, value int32) {
			s.event("save-coop:%d:%s:%d", enabled, obj.name, value)
		},
		questMapFile: func() *exitCollideTestMap4E9090 {
			s.event("quest-map")
			return s.questMap
		},
		abilityActive: func(obj *exitCollideTestObject4E9090, ability uint32) int32 {
			s.event("ability:%s:%d", obj.name, ability)
			return s.abilities[ability]
		},
		disableAbility: func(obj *exitCollideTestObject4E9090, ability uint32) {
			s.event("disable:%s:%d", obj.name, ability)
		},
		currentQuestStage: func() uint32 {
			s.event("stage:%d", s.currentStage)
			return s.currentStage
		},
		recordExitProgress: func(obj *exitCollideTestObject4E9090) {
			s.event("record:%s", obj.name)
			if s.onRecordProgress != nil {
				s.onRecordProgress(obj)
			}
		},
		loadQuestStage: func(player *exitCollideTestPlayer4E9090) uint32 {
			s.event("player-stage:%s:%d", player.name, player.stage)
			return player.stage
		},
		storeQuestStage: func(player *exitCollideTestPlayer4E9090, value uint32) {
			s.event("store-stage:%s:%d", player.name, value)
			player.stage = value
			if s.onStoreStage != nil {
				s.onStoreStage(player, value)
			}
		},
		loadPlayerIndex: func(player *exitCollideTestPlayer4E9090) uint8 {
			s.event("index:%s:%d", player.name, player.index)
			return player.index
		},
		sendQuestStage: func(index uint8, stage uint32) {
			s.event("send-stage:%d:%d", index, stage)
		},
		storeQuestExit: func(update *exitCollideTestUpdate4E9090, obj *exitCollideTestObject4E9090) {
			name := "nil"
			if obj != nil {
				name = obj.name
			}
			s.event("store-exit:%s:%s", update.name, name)
			update.exit = obj
		},
		storeQuestWarpGate: func(update *exitCollideTestUpdate4E9090, obj *exitCollideTestObject4E9090) {
			name := "nil"
			if obj != nil {
				name = obj.name
			}
			s.event("store-warp:%s:%s", update.name, name)
			update.warp = obj
		},
		setPlayerState: func(obj *exitCollideTestObject4E9090, state int32) {
			s.event("state:%s:%d", obj.name, state)
		},
		goObserver: func(player *exitCollideTestPlayer4E9090, a2, a3 int32) {
			s.event("observer:%s:%d:%d", player.name, a2, a3)
		},
		broadcastUnitMessage: func(code uint8, obj *exitCollideTestObject4E9090) {
			s.event("broadcast:%d:%s", code, obj.name)
		},
		allPlayersExited: func() int32 {
			s.event("all-exited:%d", s.allExited)
			return s.allExited
		},
		frame: func() uint32 {
			s.event("frame:%d", s.frame)
			return s.frame
		},
		storeWarpFrame: func(frame uint32) {
			s.event("store-warp-frame:%d", frame)
		},
		firstPlayerUnit: func() *exitCollideTestObject4E9090 {
			s.event("first-player")
			return s.firstPlayer
		},
		nextPlayerUnit: func(obj *exitCollideTestObject4E9090) *exitCollideTestObject4E9090 {
			s.event("next-player:%s", obj.name)
			return obj.nextPlayer
		},
		sendUnitMessage: func(code uint8, recipient, subject *exitCollideTestObject4E9090) {
			s.event("send-unit:%d:%s:%s", code, recipient.name, subject.name)
			if s.onSendUnit != nil {
				s.onSendUnit(recipient)
			}
		},
		priorityMessage: func(obj *exitCollideTestObject4E9090, message string, value uint8) {
			s.event("priority:%s:%s:%d", obj.name, message, value)
		},
		maybeWarp: func() int32 {
			s.event("maybe-warp:%d", s.maybeWarp)
			return s.maybeWarp
		},
		audio: func(id uint32, obj *exitCollideTestObject4E9090, a3, a4 int32) {
			s.event("audio:%d:%s:%d:%d", id, obj.name, a3, a4)
			if s.onAudio != nil {
				s.onAudio(obj)
			}
		},
		exitCountdown: func() int32 {
			s.event("countdown")
			return 77
		},
		copyNextMap: func(name *exitCollideTestMap4E9090) {
			s.event("copy-map:%s", name.name)
			s.copiedMap = name
		},
		nextStageThreshold: func(stage uint32) uint32 {
			s.event("threshold:%d", stage)
			return s.nextThreshold
		},
		setCurrentQuestStage: func(stage uint32) {
			s.event("set-current-stage:%d", stage)
			s.currentStage = stage
		},
		setQuestWarping: func(value int32) {
			s.event("set-warping:%d", value)
		},
		resetQuestPlayers: func() {
			s.event("reset-players")
		},
		mapLoad: func(name *exitCollideTestMap4E9090) {
			s.event("map-load:%s", name.name)
			s.loadedMap = name
		},
	}
}

func newExitCollideTest4E9090() (
	*exitCollideTestState4E9090,
	*exitCollideTestObject4E9090,
	*exitCollideTestObject4E9090,
) {
	player := &exitCollideTestPlayer4E9090{name: "player", class: 2, index: 3, stage: 4}
	update := &exitCollideTestUpdate4E9090{name: "update", player: player, traps: 2}
	unit := &exitCollideTestObject4E9090{name: "unit", class: exitCollidePlayerClassByte4E9090, update: update}
	exit := &exitCollideTestObject4E9090{
		name: "exit", subclass: exitCollideSubtypeExit4E9090,
		mapName: &exitCollideTestMap4E9090{name: "base.map"},
	}
	state := &exitCollideTestState4E9090{
		flags:         make(map[uint32]int32),
		glyphType:     7,
		warpEnabled:   1,
		exitAllowed:   1,
		questMap:      &exitCollideTestMap4E9090{name: "quest.map"},
		abilities:     make(map[uint32]int32),
		currentStage:  5,
		nextThreshold: 11,
		allExited:     1,
		maybeWarp:     1,
		frame:         0x80000003,
	}
	return state, exit, unit
}

func requireExitCollideEvents4E9090(t *testing.T, got, want []string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("events mismatch\n got: %#v\nwant: %#v", got, want)
	}
}

func TestExitCollide4E9090GlyphCachePrecedesArgumentGuards(t *testing.T) {
	state, exit, unit := newExitCollideTest4E9090()
	hooks := state.hooks()
	exitCollide4E9090(exit, nil, "collision", hooks)
	requireExitCollideEvents4E9090(t, state.events, []string{"glyph"})

	state.events = nil
	unit.class = 0
	exitCollide4E9090(exit, unit, "collision", hooks)
	requireExitCollideEvents4E9090(t, state.events, []string{
		"glyph", "class:unit", "default:exit:unit:collision",
	})
}

func TestExitCollide4E9090QuestOccupancyShortCircuitsInOrder(t *testing.T) {
	state, exit, unit := newExitCollideTest4E9090()
	state.flags[exitCollideGameQuest4E9090] = 1
	unit.update.exit = &exitCollideTestObject4E9090{name: "prior-exit"}
	exitCollide4E9090(exit, unit, "collision", state.hooks())
	requireExitCollideEvents4E9090(t, state.events, []string{
		"glyph", "class:unit", "update:unit", "map:exit", "flag:1000", "quest-exit:update",
	})

	state.events = nil
	unit.update.exit = nil
	unit.update.warp = &exitCollideTestObject4E9090{name: "prior-warp"}
	exitCollide4E9090(exit, unit, "collision", state.hooks())
	requireExitCollideEvents4E9090(t, state.events, []string{
		"glyph", "class:unit", "update:unit", "map:exit", "flag:1000",
		"quest-exit:update", "quest-warp:update",
	})
}

func TestExitCollide4E9090WarriorGlyphCleanupUsesLiveNextThenCoopSave(t *testing.T) {
	state, exit, unit := newExitCollideTest4E9090()
	state.flags[exitCollideGameCoop4E9090] = 1
	unit.update.player.class = exitCollideWarriorClass4E9090
	glyph := &exitCollideTestObject4E9090{name: "glyph", typeIndex: uint16(state.glyphType)}
	oldTail := &exitCollideTestObject4E9090{name: "old-tail", typeIndex: 99}
	liveTail := &exitCollideTestObject4E9090{name: "live-tail", typeIndex: 98}
	glyph.nextOwned = oldTail
	unit.firstOwned = glyph
	state.onDelete = func(obj *exitCollideTestObject4E9090) {
		obj.nextOwned = liveTail
	}

	exitCollide4E9090(exit, unit, "collision", state.hooks())
	requireExitCollideEvents4E9090(t, state.events, []string{
		"glyph", "class:unit", "update:unit", "map:exit", "flag:1000", "subclass:exit",
		"save-busy", "allowed:unit", "paused", "map-first:base.map", "player:update:player",
		"player-class:player", "first-owned:unit", "type:glyph", "holder:glyph", "delete:glyph",
		"traps:update:2", "store-traps:update:1", "next-owned:glyph", "type:live-tail",
		"next-owned:live-tail", "map-required:1", "flag:800", "save-name:WORKING",
		"save-coop:1:exit:0",
	})
	if unit.update.traps != 1 {
		t.Fatalf("CurTraps byte = %d, want 1", unit.update.traps)
	}
}

func TestExitCollide4E9090NonQuestLoadsLiveBaseMapWithoutQuestSideEffects(t *testing.T) {
	state, exit, unit := newExitCollideTest4E9090()
	exitCollide4E9090(exit, unit, "collision", state.hooks())
	requireExitCollideEvents4E9090(t, state.events, []string{
		"glyph", "class:unit", "update:unit", "map:exit", "flag:1000", "subclass:exit",
		"save-busy", "allowed:unit", "paused", "map-first:base.map", "player:update:player",
		"player-class:player", "map-required:1", "flag:800", "flag:1000", "flag:1000",
		"map-first:base.map", "subclass:exit", "map-load:base.map",
	})
	if state.copiedMap != nil || state.loadedMap != exit.mapName {
		t.Fatalf("maps copied/loaded = (%p,%p), want (nil,%p)", state.copiedMap, state.loadedMap, exit.mapName)
	}
}

func TestExitCollide4E9090QuestExitReloadsPlayerAndCopiesBeforeLoad(t *testing.T) {
	state, exit, unit := newExitCollideTest4E9090()
	state.flags[exitCollideGameQuest4E9090] = 1
	state.abilities[2] = 1
	state.abilities[5] = 2
	second := &exitCollideTestPlayer4E9090{name: "second", class: 2, index: 7, stage: 1}
	third := &exitCollideTestPlayer4E9090{name: "third", class: 2, index: 8, stage: 9}
	state.onRecordProgress = func(*exitCollideTestObject4E9090) {
		unit.update.player = second
	}
	state.onStoreStage = func(*exitCollideTestPlayer4E9090, uint32) {
		unit.update.player = third
	}

	exitCollide4E9090(exit, unit, "collision", state.hooks())
	requireExitCollideEvents4E9090(t, state.events, []string{
		"glyph", "class:unit", "update:unit", "map:exit", "flag:1000", "quest-exit:update",
		"quest-warp:update", "subclass:exit", "save-busy", "allowed:unit", "paused",
		"map-first:base.map", "player:update:player", "player-class:player", "map-required:1",
		"flag:800", "flag:1000", "quest-map", "flag:1000", "ability:unit:1",
		"ability:unit:2", "disable:unit:2", "ability:unit:3", "ability:unit:4",
		"ability:unit:5", "disable:unit:5", "subclass:exit", "stage:5", "record:unit",
		"player:update:second", "player-stage:second:1", "store-stage:second:6",
		"player:update:third", "player-stage:third:9", "index:third:8", "send-stage:8:9",
		"store-exit:update:exit", "store-warp:update:nil", "state:unit:13", "player:update:third",
		"observer:third:0:0", "broadcast:18:unit", "all-exited:1", "audio:1003:exit:0:0",
		"subclass:exit", "copy-map:quest.map", "map-first:quest.map", "subclass:exit",
		"map-load:quest.map",
	})
	if second.stage != 6 || unit.update.exit != exit || unit.update.warp != nil {
		t.Fatalf("Quest exit state = stage %d exit %p warp %p", second.stage, unit.update.exit, unit.update.warp)
	}
}

func TestExitCollide4E9090QuestExitFailureCountsDownAndReturnsAfterCopy(t *testing.T) {
	state, exit, unit := newExitCollideTest4E9090()
	state.flags[exitCollideGameQuest4E9090] = 1
	state.allExited = 0
	exitCollide4E9090(exit, unit, "collision", state.hooks())

	var tail []string
	for i, event := range state.events {
		if event == "all-exited:0" {
			tail = state.events[i:]
			break
		}
	}
	requireExitCollideEvents4E9090(t, tail, []string{
		"all-exited:0", "audio:1003:exit:0:0", "subclass:exit", "countdown",
		"copy-map:quest.map",
	})
	if state.loadedMap != nil || state.copiedMap != state.questMap {
		t.Fatalf("maps copied/loaded = (%p,%p), want (%p,nil)", state.copiedMap, state.loadedMap, state.questMap)
	}
}

func TestExitCollide4E9090WarpUsesLiveTraversalThenTransitionsStage(t *testing.T) {
	state, exit, unit := newExitCollideTest4E9090()
	exit.subclass = exitCollideSubtypeWarp4E9090
	state.flags[exitCollideGameQuest4E9090] = 1
	other := &exitCollideTestObject4E9090{name: "other"}
	liveTail := &exitCollideTestObject4E9090{name: "live-tail"}
	oldTail := &exitCollideTestObject4E9090{name: "old-tail"}
	unit.nextPlayer = other
	other.nextPlayer = oldTail
	state.firstPlayer = unit
	state.onSendUnit = func(obj *exitCollideTestObject4E9090) {
		if obj == other {
			obj.nextPlayer = liveTail
		}
	}

	exitCollide4E9090(exit, unit, "collision", state.hooks())

	var fromStores []string
	for i, event := range state.events {
		if event == "store-exit:update:nil" {
			fromStores = state.events[i:]
			break
		}
	}
	requireExitCollideEvents4E9090(t, fromStores, []string{
		"store-exit:update:nil", "store-warp:update:exit", "frame:2147483651",
		"store-warp-frame:2147483651", "state:unit:13", "player:update:player",
		"observer:player:0:0", "first-player", "next-player:unit",
		"send-unit:19:other:unit", "next-player:other", "send-unit:19:live-tail:unit",
		"next-player:live-tail", "priority:unit:objcoll.c:PlayerEntersWarp:0", "maybe-warp:1",
		"audio:1003:exit:0:0", "subclass:exit", "copy-map:quest.map", "map-first:quest.map",
		"subclass:exit", "stage:5", "threshold:5", "set-current-stage:10", "set-warping:1",
		"reset-players", "map-load:quest.map",
	})
	if unit.update.exit != nil || unit.update.warp != exit || state.currentStage != 10 || state.loadedMap != state.questMap {
		t.Fatalf("warp state = exit %p warp %p stage %d map %p", unit.update.exit, unit.update.warp, state.currentStage, state.loadedMap)
	}
}

func TestExitCollide4E9090ExactOneGuardsAndNilQuestMapFault(t *testing.T) {
	state, exit, unit := newExitCollideTest4E9090()
	state.saveBusy = 2
	state.paused = 2
	exitCollide4E9090(exit, unit, "collision", state.hooks())
	if state.loadedMap != exit.mapName {
		t.Fatal("non-one save/pause results must not block the original path")
	}

	state, exit, unit = newExitCollideTest4E9090()
	state.flags[exitCollideGameQuest4E9090] = 1
	state.questMap = nil
	defer func() {
		if recover() == nil {
			t.Fatal("nil Quest map must fault when GAME.EXE performs strcpy")
		}
	}()
	exitCollide4E9090(exit, unit, "collision", state.hooks())
}

func TestExitCollide4E9090GateFailuresStopAtOriginalBoundary(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*exitCollideTestState4E9090, *exitCollideTestObject4E9090)
		want   []string
	}{
		{
			name: "warp disabled",
			mutate: func(state *exitCollideTestState4E9090, exit *exitCollideTestObject4E9090) {
				exit.subclass = exitCollideSubtypeWarp4E9090
				state.warpEnabled = 0
			},
			want: []string{"glyph", "class:unit", "update:unit", "map:exit", "flag:1000", "subclass:exit", "warp-enabled"},
		},
		{
			name: "save busy exact one",
			mutate: func(state *exitCollideTestState4E9090, _ *exitCollideTestObject4E9090) {
				state.saveBusy = 1
			},
			want: []string{"glyph", "class:unit", "update:unit", "map:exit", "flag:1000", "subclass:exit", "save-busy"},
		},
		{
			name: "exit denied",
			mutate: func(state *exitCollideTestState4E9090, _ *exitCollideTestObject4E9090) {
				state.exitAllowed = 0
			},
			want: []string{"glyph", "class:unit", "update:unit", "map:exit", "flag:1000", "subclass:exit", "save-busy", "allowed:unit"},
		},
		{
			name: "paused exact one",
			mutate: func(state *exitCollideTestState4E9090, _ *exitCollideTestObject4E9090) {
				state.paused = 1
			},
			want: []string{"glyph", "class:unit", "update:unit", "map:exit", "flag:1000", "subclass:exit", "save-busy", "allowed:unit", "paused"},
		},
		{
			name: "empty non Quest map",
			mutate: func(_ *exitCollideTestState4E9090, exit *exitCollideTestObject4E9090) {
				exit.mapName.name = ""
			},
			want: []string{"glyph", "class:unit", "update:unit", "map:exit", "flag:1000", "subclass:exit", "save-busy", "allowed:unit", "paused", "map-first:", "flag:1000"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			state, exit, unit := newExitCollideTest4E9090()
			test.mutate(state, exit)
			exitCollide4E9090(exit, unit, "collision", state.hooks())
			requireExitCollideEvents4E9090(t, state.events, test.want)
		})
	}
}

func TestExitCollide4E9090ExitBitWinsBranchButWarpBitWinsFinalTransition(t *testing.T) {
	state, exit, unit := newExitCollideTest4E9090()
	state.flags[exitCollideGameQuest4E9090] = 1
	exit.subclass = exitCollideSubtypeExit4E9090 | exitCollideSubtypeWarp4E9090

	exitCollide4E9090(exit, unit, "collision", state.hooks())

	for _, event := range []string{
		"store-exit:update:exit",
		"broadcast:18:unit",
		"set-current-stage:10",
		"set-warping:1",
		"reset-players",
		"map-load:quest.map",
	} {
		if !slices.Contains(state.events, event) {
			t.Fatalf("missing event %q in %#v", event, state.events)
		}
	}
	for _, event := range state.events {
		if event == "countdown" || event == "store-warp-frame:2147483651" ||
			event == "priority:unit:objcoll.c:PlayerEntersWarp:0" {
			t.Fatalf("wrong branch event %q in %#v", event, state.events)
		}
	}
	if unit.update.exit != exit || unit.update.warp != nil || state.currentStage != 10 {
		t.Fatalf("combined subtype state = exit %p warp %p stage %d", unit.update.exit, unit.update.warp, state.currentStage)
	}
}
