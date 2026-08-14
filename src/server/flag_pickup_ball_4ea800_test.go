package server

import (
	"reflect"
	"testing"
)

func defaultFlagPickupBallHooks4EA800() flagPickupBallHooks4EA800[int, int, int, int] {
	return flagPickupBallHooks4EA800[int, int, int, int]{
		loadBallCache:    func() uint32 { return 1 },
		lookupType:       func(string) uint32 { return 0 },
		storeBallCache:   func(uint32) {},
		loadClassLow:     func(int) uint8 { return 0 },
		unitIsGameBall:   func(int) int32 { return 0 },
		firstOwned:       func(int) int { return 0 },
		nextOwned:        func(int) int { return 0 },
		loadTypeInd:      func(int) uint16 { return 0 },
		loadUpdate:       func(int) int { return 0 },
		loadCarrier:      func(int) int { return 0 },
		loadFlagsLow:     func(int) uint8 { return 0 },
		storeCarrier:     func(int, int) {},
		loadTeamID:       func(int) uint8 { return 0 },
		teamByID:         func(uint8) int { return 0 },
		nextTeam:         func(int) int { return 0 },
		firstTeam:        func() int { return 0 },
		loadTeamIDValue:  func(int) uint8 { return 0 },
		gameData:         func(uint32) uint16 { return 0 },
		changeScore:      func(int, int32) {},
		reportLesson:     func(int) {},
		loadTeamScore:    func(int) int32 { return 0 },
		changeTeamScore:  func(int, int32) {},
		observerMode:     func() uint32 { return 0 },
		playerFromUpdate: func(int) int { return 0 },
		observerUpdate:   func(int, int) {},
		audio:            func(uint32, int) {},
		loadNetCode:      func(int) uint32 { return 0 },
		informScore:      func(uint32, uint32, uint32) {},
		pointFX:          func(uint32, int) {},
		setGameFlags:     func(uint32) {},
		flagBallWinner:   func(int) {},
		loadStartCache:   func() uint32 { return 1 },
		storeStartCache:  func(uint32) {},
		firstObject:      func() int { return 0 },
		nextObject:       func(int) int { return 0 },
		randomInt:        func(int32, int32) int32 { return 0 },
		clearOwner:       func(int) {},
		dropBall:         func(int, int) {},
		changeObjectTeam: func(int, uint32) {},
		setHPMax:         func(int) {},
		ticks:            func() uint64 { return 0 },
		storeTicks:       func(int, uint64) {},
		moveToMarker:     func(int, int) {},
		ballStatus:       func(uint32, uint32) {},
		clearMotion:      func(int) {},
	}
}

func TestFlagPickupBallResolve4EA800CachesZeroBeforePlayerGate(t *testing.T) {
	events := make([]string, 0, 5)
	hooks := defaultFlagPickupBallHooks4EA800()
	hooks.loadBallCache = func() uint32 {
		events = append(events, "cache")
		return 0
	}
	hooks.lookupType = func(name string) uint32 {
		events = append(events, "lookup")
		if name != flagPickupBallGameBallName4EA800 {
			t.Fatalf("lookup name = %q", name)
		}
		return 0
	}
	hooks.storeBallCache = func(ind uint32) {
		events = append(events, "store")
		if ind != 0 {
			t.Fatalf("stored type = %d", ind)
		}
	}
	hooks.loadClassLow = func(target int) uint8 {
		events = append(events, "class")
		if target != 2 {
			t.Fatalf("class target = %d", target)
		}
		return flagPickupBallPlayerClass4EA800
	}
	hooks.unitIsGameBall = func(target int) int32 {
		events = append(events, "unit-is-ball")
		return 0
	}
	if got := flagPickupBallResolve4EA800(2, hooks); got != 0 {
		t.Fatalf("resolved object = %d", got)
	}
	want := []string{"cache", "lookup", "store", "class", "unit-is-ball"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestFlagPickupBallResolve4EA800ReloadsCacheForEachOwnedObject(t *testing.T) {
	events := make([]string, 0, 12)
	cache := uint32(7)
	cacheLoads := 0
	hooks := defaultFlagPickupBallHooks4EA800()
	hooks.loadBallCache = func() uint32 {
		cacheLoads++
		events = append(events, "cache")
		return cache
	}
	hooks.lookupType = func(string) uint32 {
		t.Fatal("nonzero cache triggered lookup")
		return 0
	}
	hooks.loadClassLow = func(target int) uint8 {
		events = append(events, "class")
		return flagPickupBallPlayerClass4EA800
	}
	hooks.unitIsGameBall = func(target int) int32 {
		events = append(events, "unit-is-ball")
		return 1
	}
	hooks.firstOwned = func(target int) int {
		events = append(events, "first-owned")
		return 3
	}
	hooks.loadTypeInd = func(obj int) uint16 {
		events = append(events, map[int]string{3: "type-3", 4: "type-4"}[obj])
		return map[int]uint16{3: 8, 4: 9}[obj]
	}
	hooks.nextOwned = func(obj int) int {
		events = append(events, "next-owned")
		if obj != 3 {
			t.Fatalf("next object = %d", obj)
		}
		cache = 9
		return 4
	}
	if got := flagPickupBallResolve4EA800(2, hooks); got != 4 {
		t.Fatalf("resolved object = %d", got)
	}
	if cacheLoads != 3 {
		t.Fatalf("cache loads = %d, want 3", cacheLoads)
	}
	want := []string{
		"cache", "class", "unit-is-ball", "first-owned",
		"cache", "type-3", "next-owned", "cache", "type-4",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestFlagPickupBallScore4EA800ReloadsCarrierFieldsAndFindsWinner(t *testing.T) {
	events := make([]string, 0, 36)
	carrier := 10
	netCode := uint32(0x100)
	firstTeamCalls := 0
	teamIDLoads := 0
	hooks := defaultFlagPickupBallHooks4EA800()
	hooks.loadUpdate = func(obj int) int {
		if obj == 2 {
			events = append(events, "update-ball")
			return 100
		}
		events = append(events, "update-carrier")
		if obj != 11 {
			t.Fatalf("carrier update object = %d", obj)
		}
		return 111
	}
	carrierLoads := 0
	hooks.loadCarrier = func(update int) int {
		carrierLoads++
		events = append(events, map[int]string{1: "carrier-first", 2: "carrier-reload"}[carrierLoads])
		if update != 100 {
			t.Fatalf("carrier update = %d", update)
		}
		return carrier
	}
	hooks.loadFlagsLow = func(obj int) uint8 {
		events = append(events, "flags")
		if obj != 10 {
			t.Fatalf("flag object = %d", obj)
		}
		return flagPickupBallInactiveCarrier4EA800
	}
	hooks.storeCarrier = func(update, obj int) {
		events = append(events, "clear-carrier")
		if update != 100 || obj != 0 {
			t.Fatalf("store carrier = %d/%d", update, obj)
		}
		carrier = 0
	}
	hooks.loadTeamID = func(obj int) uint8 {
		if obj == 1 {
			events = append(events, "source-team")
			return 3
		}
		events = append(events, "carrier-team")
		if obj != 11 {
			t.Fatalf("team object = %d", obj)
		}
		return 5
	}
	hooks.teamByID = func(id uint8) int {
		events = append(events, "team-by-id")
		if id != 3 {
			t.Fatalf("source team ID = %d", id)
		}
		return 20
	}
	hooks.nextTeam = func(team int) int {
		switch team {
		case 20:
			events = append(events, "next-source-team")
			carrier = 11
			return 0
		case 30:
			events = append(events, "next-winner-team")
			return 31
		default:
			t.Fatalf("next team = %d", team)
			return 0
		}
	}
	hooks.firstTeam = func() int {
		firstTeamCalls++
		if firstTeamCalls == 1 {
			events = append(events, "first-wrap-team")
			return 21
		}
		events = append(events, "first-winner-team")
		return 30
	}
	hooks.loadTeamIDValue = func(team int) uint8 {
		teamIDLoads++
		if team != 21 {
			t.Fatalf("team ID object = %d", team)
		}
		if teamIDLoads == 1 {
			events = append(events, "team-id-compare")
			return 5
		}
		events = append(events, "team-id-inform")
		return 6
	}
	hooks.gameData = func(mode uint32) uint16 {
		events = append(events, "game-data")
		if mode != flagPickupBallScoreMode4EA800 {
			t.Fatalf("game-data mode = %d", mode)
		}
		return 2
	}
	hooks.changeScore = func(obj int, delta int32) {
		events = append(events, "change-score")
		if obj != 11 || delta != 1 {
			t.Fatalf("score = %d/%d", obj, delta)
		}
	}
	hooks.reportLesson = func(obj int) {
		events = append(events, "report-lesson")
	}
	hooks.loadTeamScore = func(team int) int32 {
		switch team {
		case 21:
			events = append(events, "score-own-team")
			return 10
		case 30:
			events = append(events, "score-team-30")
			return -1
		case 31:
			events = append(events, "score-team-31")
			return 2
		default:
			t.Fatalf("score team = %d", team)
			return 0
		}
	}
	hooks.changeTeamScore = func(team int, score int32) {
		events = append(events, "change-team-score")
		if team != 21 || score != 11 {
			t.Fatalf("team score = %d/%d", team, score)
		}
	}
	hooks.observerMode = func() uint32 {
		events = append(events, "observer-mode")
		return 1
	}
	hooks.playerFromUpdate = func(update int) int {
		events = append(events, "player-from-update")
		if update != 111 {
			t.Fatalf("player update = %d", update)
		}
		return 77
	}
	hooks.observerUpdate = func(player, other int) {
		events = append(events, "observer-update")
		if player != 77 || other != 0 {
			t.Fatalf("observer = %d/%d", player, other)
		}
	}
	hooks.audio = func(id uint32, source int) {
		events = append(events, "audio")
		if id != flagPickupBallScoreAudio4EA800 || source != 1 {
			t.Fatalf("audio = %d/%d", id, source)
		}
		netCode = 0xabcdef01
	}
	hooks.loadNetCode = func(obj int) uint32 {
		events = append(events, "net-code")
		if obj != 11 {
			t.Fatalf("net object = %d", obj)
		}
		return netCode
	}
	hooks.informScore = func(code, net, teamID uint32) {
		events = append(events, "inform-score")
		if code != 9 || net != 0xabcdef01 || teamID != 6 {
			t.Fatalf("inform = %d/%#x/%d", code, net, teamID)
		}
	}
	hooks.pointFX = func(code uint32, obj int) {
		events = append(events, "score-fx")
		if code != 154 || obj != 2 {
			t.Fatalf("point FX = %d/%d", code, obj)
		}
	}
	hooks.setGameFlags = func(flags uint32) {
		events = append(events, "set-game-flags")
		if flags != 8 {
			t.Fatalf("game flags = %d", flags)
		}
	}
	hooks.flagBallWinner = func(team int) {
		events = append(events, "winner")
		if team != 31 {
			t.Fatalf("winner team = %d", team)
		}
	}

	if !flagPickupBallScore4EA800(1, 2, hooks) {
		t.Fatal("scoring branch did not reach respawn tail")
	}
	want := []string{
		"update-ball", "carrier-first", "flags", "clear-carrier",
		"source-team", "team-by-id", "next-source-team", "first-wrap-team",
		"carrier-reload", "carrier-team", "team-id-compare", "game-data",
		"change-score", "report-lesson", "score-own-team", "change-team-score",
		"observer-mode", "update-carrier", "player-from-update", "observer-update",
		"audio", "net-code", "team-id-inform", "inform-score", "score-fx",
		"first-winner-team", "score-team-30", "next-winner-team", "score-team-31",
		"set-game-flags", "winner",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events =\n%v\nwant =\n%v", events, want)
	}
}

func TestFlagPickupBallScore4EA800SelectsTeamBeforeNilCarrierReturn(t *testing.T) {
	events := make([]string, 0, 8)
	hooks := defaultFlagPickupBallHooks4EA800()
	hooks.loadUpdate = func(int) int { events = append(events, "update"); return 100 }
	hooks.loadCarrier = func(int) int { events = append(events, "carrier"); return 0 }
	hooks.loadTeamID = func(int) uint8 { events = append(events, "team-id"); return 4 }
	hooks.teamByID = func(uint8) int { events = append(events, "team-by-id"); return 20 }
	hooks.nextTeam = func(int) int { events = append(events, "next-team"); return 21 }
	hooks.firstTeam = func() int {
		t.Fatal("non-nil next team wrapped")
		return 0
	}
	hooks.gameData = func(uint32) uint16 {
		t.Fatal("nil carrier reached score effects")
		return 0
	}
	if flagPickupBallScore4EA800(1, 2, hooks) {
		t.Fatal("nil carrier reached respawn tail")
	}
	want := []string{"update", "carrier", "team-id", "team-by-id", "next-team", "carrier"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestFlagPickupBallRespawn4EA800CountsSelectsAndPlacesWithLiveCaches(t *testing.T) {
	events := make([]string, 0, 40)
	startCache := uint32(0)
	pass := 0
	netCode := uint32(33)
	hooks := defaultFlagPickupBallHooks4EA800()
	hooks.loadStartCache = func() uint32 {
		events = append(events, map[int]string{0: "cache-initial", 1: "cache-count", 2: "cache-select"}[pass])
		return startCache
	}
	hooks.lookupType = func(name string) uint32 {
		events = append(events, "lookup-start")
		if name != flagPickupBallStartName4EA800 {
			t.Fatalf("lookup name = %q", name)
		}
		return 7
	}
	hooks.storeStartCache = func(ind uint32) {
		events = append(events, "store-start")
		startCache = ind
	}
	hooks.firstObject = func() int {
		pass++
		events = append(events, map[int]string{1: "first-count", 2: "first-select"}[pass])
		return 1
	}
	hooks.loadTypeInd = func(obj int) uint16 {
		events = append(events, map[int]map[int]string{
			1: {1: "type-count-1", 2: "type-count-2", 3: "type-count-3"},
			2: {1: "type-select-1", 2: "type-select-2", 3: "type-select-3"},
		}[pass][obj])
		return map[int]uint16{1: 7, 2: 8, 3: 7}[obj]
	}
	hooks.nextObject = func(obj int) int {
		events = append(events, map[int]map[int]string{
			1: {1: "next-count-1", 2: "next-count-2", 3: "next-count-3"},
			2: {1: "next-select-1", 2: "next-select-2", 3: "next-select-3"},
		}[pass][obj])
		return map[int]int{1: 2, 2: 3, 3: 0}[obj]
	}
	hooks.randomInt = func(min, max int32) int32 {
		events = append(events, "random")
		if min != 0 || max != 1 {
			t.Fatalf("random bounds = %d/%d", min, max)
		}
		return 1
	}
	hooks.loadUpdate = func(obj int) int {
		events = append(events, "update-ball")
		if obj != 9 {
			t.Fatalf("update object = %d", obj)
		}
		return 900
	}
	hooks.clearOwner = func(obj int) {
		events = append(events, "clear-owner")
	}
	hooks.dropBall = func(obj, owner int) {
		events = append(events, "drop-ball")
		if obj != 9 || owner != 0 {
			t.Fatalf("drop = %d/%d", obj, owner)
		}
		netCode = 44
	}
	hooks.loadNetCode = func(obj int) uint32 {
		events = append(events, "net-code")
		return netCode
	}
	hooks.changeObjectTeam = func(obj int, net uint32) {
		events = append(events, "change-object-team")
		if obj != 9 || net != 44 {
			t.Fatalf("change team = %d/%d", obj, net)
		}
	}
	hooks.setHPMax = func(obj int) {
		events = append(events, "hp-max")
	}
	hooks.ticks = func() uint64 {
		events = append(events, "ticks")
		return 0x1122334455667788
	}
	hooks.storeTicks = func(update int, ticks uint64) {
		events = append(events, "store-ticks")
		if update != 900 || ticks != 0x1122334455667788 {
			t.Fatalf("ticks = %d/%#x", update, ticks)
		}
	}
	hooks.moveToMarker = func(ball, marker int) {
		events = append(events, "move")
		if ball != 9 || marker != 3 {
			t.Fatalf("move = %d/%d", ball, marker)
		}
	}
	hooks.ballStatus = func(a1, a2 uint32) {
		events = append(events, "status")
		if a1 != 0 || a2 != 0 {
			t.Fatalf("status = %d/%d", a1, a2)
		}
	}
	hooks.pointFX = func(code uint32, obj int) {
		events = append(events, "respawn-fx")
		if code != 129 || obj != 9 {
			t.Fatalf("FX = %d/%d", code, obj)
		}
	}
	hooks.clearMotion = func(obj int) {
		events = append(events, "clear-motion")
	}

	flagPickupBallRespawn4EA800(9, hooks)
	want := []string{
		"cache-initial", "lookup-start", "store-start", "first-count",
		"cache-count", "type-count-1", "next-count-1",
		"cache-count", "type-count-2", "next-count-2",
		"cache-count", "type-count-3", "next-count-3", "random", "first-select",
		"cache-select", "type-select-1", "next-select-1",
		"cache-select", "type-select-2", "next-select-2",
		"cache-select", "type-select-3", "update-ball", "clear-owner", "drop-ball",
		"net-code", "change-object-team", "hp-max", "ticks", "store-ticks", "move",
		"status", "respawn-fx", "clear-motion",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events =\n%v\nwant =\n%v", events, want)
	}
}

func TestFlagPickupBallRespawn4EA800CallsRandomForZeroMarkers(t *testing.T) {
	events := make([]string, 0, 4)
	firstCalls := 0
	hooks := defaultFlagPickupBallHooks4EA800()
	hooks.loadStartCache = func() uint32 { events = append(events, "cache"); return 7 }
	hooks.lookupType = func(string) uint32 {
		t.Fatal("nonzero start cache triggered lookup")
		return 0
	}
	hooks.firstObject = func() int {
		firstCalls++
		events = append(events, map[int]string{1: "first-count", 2: "first-select"}[firstCalls])
		return 0
	}
	hooks.randomInt = func(min, max int32) int32 {
		events = append(events, "random")
		if min != 0 || max != -1 {
			t.Fatalf("zero-count random bounds = %d/%d", min, max)
		}
		return 0
	}
	hooks.clearOwner = func(int) { t.Fatal("zero marker count placed ball") }
	flagPickupBallRespawn4EA800(9, hooks)
	want := []string{"cache", "first-count", "random", "first-select"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestFlagPickupBall4EA800DoesNotRespawnAfterCarrierDisappears(t *testing.T) {
	events := make([]string, 0, 9)
	hooks := defaultFlagPickupBallHooks4EA800()
	hooks.loadBallCache = func() uint32 { events = append(events, "ball-cache"); return 7 }
	hooks.loadClassLow = func(int) uint8 { events = append(events, "class"); return 0 }
	hooks.loadUpdate = func(int) int { events = append(events, "update"); return 100 }
	hooks.loadCarrier = func(int) int { events = append(events, "carrier"); return 0 }
	hooks.loadTeamID = func(int) uint8 { events = append(events, "source-team"); return 3 }
	hooks.teamByID = func(uint8) int { events = append(events, "team-by-id"); return 20 }
	hooks.nextTeam = func(int) int { events = append(events, "next-team"); return 21 }
	hooks.loadStartCache = func() uint32 {
		t.Fatal("nil carrier reached respawn")
		return 0
	}
	flagPickupBall4EA800(1, 2, struct{ untouched uint32 }{untouched: 0xdeadbeef}, hooks)
	want := []string{
		"ball-cache", "class", "update", "carrier", "source-team", "team-by-id", "next-team", "carrier",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}
