package server

import (
	"math"
	"reflect"
	"testing"
)

func defaultFlagPickupCTFHooks4EA490() flagPickupCTFHooks4EA490[int, int, int, int] {
	return flagPickupCTFHooks4EA490[int, int, int, int]{
		loadUpdate:       func(int) int { return 0 },
		flagIndex:        func(int) uint32 { return 0 },
		loadTeamID:       func(int) uint8 { return 0 },
		teamsSame:        func(int, int) int32 { return 0 },
		loadPosX:         func(int) float32 { return 0 },
		loadPosY:         func(int) float32 { return 0 },
		loadHomeX:        func(int) float32 { return 0 },
		loadHomeY:        func(int) float32 { return 0 },
		moveHome:         func(int, int) {},
		loadNetCode:      func(int) uint32 { return 0 },
		informReturn:     func(uint32) {},
		informFlag:       func(uint32, uint32, uint32) {},
		storeFlagState:   func(int, uint32) {},
		flagStatus:       func(uint8, uint8, uint8, uint16) int32 { return 0 },
		firstInventory:   func(int) int { return 0 },
		nextInventory:    func(int) int { return 0 },
		loadClass:        func(int) uint32 { return 0 },
		gameData:         func(uint32) uint16 { return 0 },
		changeScore:      func(int, int32) {},
		reportLesson:     func(int) {},
		hasTeam:          func(int) int32 { return 0 },
		teamByID:         func(uint8) int { return 0 },
		loadTeamScore:    func(int) int32 { return 0 },
		changeTeamScore:  func(int, int32) {},
		observerMode:     func() uint32 { return 0 },
		playerFromUpdate: func(int) int { return 0 },
		observerUpdate:   func(int, int) {},
		detachInventory:  func(int, int) {},
		createAt:         func(int, int, float32, float32) {},
		raise:            func(int, float32) {},
		markMinimap:      func(int, uint32) {},
		firstTeam:        func() int { return 0 },
		nextTeam:         func(int) int { return 0 },
		setGameFlags:     func(uint32) {},
		flagWinner:       func(int, uint32) {},
		inventoryHolder:  func(int) int { return 0 },
		teamEligible:     func(int) int32 { return 0 },
		forceDrop:        func(int, int) {},
		finalizeDelete:   func(int) {},
		inventoryPut:     func(int, int, int32) {},
		markPlayerPickup: func(int, uint32) {},
		reportObject:     func(uint32, int) {},
		unmarkMinimap:    func(int, uint32) {},
		purgeBuffs:       func(int) {},
		priorityMessage:  func(int, string, uint32) {},
	}
}

func TestFlagPickupCTFOutsideHome4EA490(t *testing.T) {
	nan := float32(math.NaN())
	tests := []struct {
		pos, home float32
		want      bool
	}{
		{0, 0, false},
		{5, 0, false},
		{-5, 0, false},
		{math.Nextafter32(5, 6), 0, true},
		{math.Nextafter32(-5, -6), 0, true},
		{nan, 0, false},
		{0, nan, false},
		{float32(math.Inf(1)), 0, true},
	}
	for _, tc := range tests {
		if got := flagPickupCTFOutsideHome4EA490(tc.pos, tc.home); got != tc.want {
			t.Errorf("outside(%v, %v) = %t, want %t", tc.pos, tc.home, got, tc.want)
		}
	}
}

func TestFlagPickupCTF4EA490ReturnHomeCachesAndReloads(t *testing.T) {
	events := make([]string, 0, 16)
	targetTeam := uint8(3)
	hooks := defaultFlagPickupCTFHooks4EA490()
	hooks.loadUpdate = func(obj int) int {
		events = append(events, map[int]string{1: "source-update", 2: "target-update"}[obj])
		return map[int]int{1: 101, 2: 202}[obj]
	}
	hooks.flagIndex = func(obj int) uint32 {
		events = append(events, "source-index")
		if obj != 1 {
			t.Fatalf("index object = %d", obj)
		}
		return 0x102
	}
	hooks.loadTeamID = func(obj int) uint8 {
		if obj == 1 {
			events = append(events, "source-team")
			return 7
		}
		events = append(events, "target-team")
		return targetTeam
	}
	hooks.teamsSame = func(target, source int) int32 {
		events = append(events, "same")
		if target != 2 || source != 1 {
			t.Fatalf("same args = %d/%d", target, source)
		}
		return -1
	}
	hooks.loadPosX = func(obj int) float32 {
		events = append(events, "pos-x")
		return 6
	}
	hooks.loadHomeX = func(update int) float32 {
		events = append(events, "home-x")
		if update != 101 {
			t.Fatalf("home update = %d", update)
		}
		return 0
	}
	hooks.loadPosY = func(int) float32 {
		t.Fatal("outside X loaded Y")
		return 0
	}
	hooks.loadNetCode = func(obj int) uint32 {
		events = append(events, "net")
		if obj != 2 {
			t.Fatalf("net object = %d", obj)
		}
		return 0x12345678
	}
	hooks.moveHome = func(obj, update int) {
		events = append(events, "move")
		if obj != 1 || update != 101 {
			t.Fatalf("move args = %d/%d", obj, update)
		}
		targetTeam = 9
	}
	hooks.informReturn = func(netCode uint32) {
		events = append(events, "inform")
		if netCode != 0x12345678 {
			t.Fatalf("return net code = %#x", netCode)
		}
	}
	hooks.storeFlagState = func(update int, state uint32) {
		events = append(events, "store")
		if update != 101 || state != 0 {
			t.Fatalf("store args = %d/%d", update, state)
		}
	}
	hooks.flagStatus = func(teamID, status, index uint8, carrier uint16) int32 {
		events = append(events, "status")
		if teamID != 9 || status != 0 || index != 2 || carrier != 0 {
			t.Fatalf("status = %d/%d/%d/%#x", teamID, status, index, carrier)
		}
		return -1
	}
	flagPickupCTF4EA490(1, 2, (*uint32)(nil), hooks)
	want := []string{
		"target-update", "source-index", "source-team", "same", "source-update",
		"pos-x", "home-x", "net", "move", "inform", "store", "target-team", "status",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestFlagPickupCTF4EA490ScoresAndFindsWinner(t *testing.T) {
	events := make([]string, 0, 48)
	homeX, homeY := float32(10), float32(20)
	hooks := defaultFlagPickupCTFHooks4EA490()
	hooks.loadUpdate = func(obj int) int {
		events = append(events, map[int]string{1: "update-source", 2: "update-target", 3: "update-item"}[obj])
		return map[int]int{1: 101, 2: 202, 3: 303}[obj]
	}
	hooks.flagIndex = func(obj int) uint32 {
		events = append(events, map[int]string{1: "index-source", 3: "index-item"}[obj])
		return map[int]uint32{1: 10, 3: 0x123}[obj]
	}
	hooks.loadTeamID = func(obj int) uint8 {
		events = append(events, map[int]string{1: "team-source", 2: "team-target", 3: "team-item"}[obj])
		return map[int]uint8{1: 1, 2: 2, 3: 4}[obj]
	}
	hooks.teamsSame = func(int, int) int32 { events = append(events, "same"); return 1 }
	hooks.loadPosX = func(int) float32 { events = append(events, "pos-x"); return 10 }
	hooks.loadHomeX = func(update int) float32 {
		events = append(events, map[int]string{101: "source-home-x", 303: "item-home-x"}[update])
		if update == 303 {
			return homeX
		}
		return 10
	}
	hooks.loadPosY = func(int) float32 { events = append(events, "pos-y"); return 20 }
	hooks.loadHomeY = func(update int) float32 {
		events = append(events, map[int]string{101: "source-home-y", 303: "item-home-y"}[update])
		if update == 303 {
			return homeY
		}
		return 20
	}
	hooks.firstInventory = func(owner int) int {
		events = append(events, "first-item")
		return 3
	}
	hooks.nextInventory = func(int) int {
		t.Fatal("winner path read item successor")
		return 0
	}
	hooks.loadClass = func(item int) uint32 { events = append(events, "class-item"); return 0x10000000 }
	hooks.gameData = func(mode uint32) uint16 {
		events = append(events, "game-data")
		if mode != 32 {
			t.Fatalf("mode = %d", mode)
		}
		return 2
	}
	hooks.changeScore = func(target int, delta int32) {
		events = append(events, "score-player")
		if target != 2 || delta != 1 {
			t.Fatalf("score args = %d/%d", target, delta)
		}
	}
	hooks.reportLesson = func(int) { events = append(events, "report-lesson") }
	hooks.hasTeam = func(int) int32 { events = append(events, "has-team"); return 1 }
	hooks.teamByID = func(id uint8) int {
		events = append(events, "team-by-id")
		if id != 2 {
			t.Fatalf("team id = %d", id)
		}
		return 11
	}
	hooks.loadTeamScore = func(team int) int32 {
		events = append(events, map[int]string{11: "score-own", 21: "score-team-21", 22: "score-team-22"}[team])
		return map[int]int32{11: 4, 21: 1, 22: 2}[team]
	}
	hooks.changeTeamScore = func(team int, score int32) {
		events = append(events, "change-team-score")
		if team != 11 || score != 5 {
			t.Fatalf("team score args = %d/%d", team, score)
		}
	}
	hooks.observerMode = func() uint32 { events = append(events, "observer-mode"); return 1 }
	hooks.playerFromUpdate = func(update int) int {
		events = append(events, "player")
		if update != 202 {
			t.Fatalf("cached target update = %d", update)
		}
		return 55
	}
	hooks.observerUpdate = func(player, other int) {
		events = append(events, "observer")
		if player != 55 || other != 0 {
			t.Fatalf("observer args = %d/%d", player, other)
		}
	}
	hooks.detachInventory = func(owner, item int) {
		events = append(events, "detach")
		if owner != 2 || item != 3 {
			t.Fatalf("detach args = %d/%d", owner, item)
		}
		homeX, homeY = 31, 32
	}
	hooks.createAt = func(item, owner int, x, y float32) {
		events = append(events, "create")
		if item != 3 || owner != 0 || x != 31 || y != 32 {
			t.Fatalf("create args = %d/%d/%v/%v", item, owner, x, y)
		}
	}
	hooks.raise = func(item int, z float32) { events = append(events, "raise") }
	hooks.markMinimap = func(item int, flags uint32) { events = append(events, "mark") }
	hooks.storeFlagState = func(update int, state uint32) { events = append(events, "store") }
	hooks.flagStatus = func(teamID, status, index uint8, carrier uint16) int32 {
		events = append(events, "status")
		if teamID != 4 || status != 0 || index != 0x23 || carrier != 0 {
			t.Fatalf("status = %d/%d/%#x/%#x", teamID, status, index, carrier)
		}
		return 0
	}
	hooks.loadNetCode = func(int) uint32 { events = append(events, "net"); return 0xaabbccdd }
	hooks.informFlag = func(code, netCode, index uint32) {
		events = append(events, "inform")
		if code != 5 || netCode != 0xaabbccdd || index != 0x123 {
			t.Fatalf("inform = %d/%#x/%#x", code, netCode, index)
		}
	}
	hooks.firstTeam = func() int { events = append(events, "first-team"); return 21 }
	hooks.nextTeam = func(team int) int {
		events = append(events, "next-team")
		if team != 21 {
			t.Fatalf("next team = %d", team)
		}
		return 22
	}
	hooks.setGameFlags = func(flags uint32) { events = append(events, "set-game") }
	hooks.flagWinner = func(team int, arg uint32) {
		events = append(events, "winner")
		if team != 22 || arg != 0 {
			t.Fatalf("winner args = %d/%d", team, arg)
		}
	}
	flagPickupCTF4EA490(1, 2, 999, hooks)
	want := []string{
		"update-target", "index-source", "team-source", "same", "update-source",
		"pos-x", "source-home-x", "pos-y", "source-home-y", "first-item", "class-item",
		"game-data", "update-item", "index-item", "team-item", "score-player", "report-lesson",
		"has-team", "team-target", "team-by-id", "score-own", "change-team-score", "observer-mode",
		"player", "observer", "detach", "item-home-y", "item-home-x", "create", "raise", "mark",
		"store", "status", "net", "inform", "first-team", "score-team-21", "next-team",
		"score-team-22", "set-game", "winner",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events =\n%v\nwant =\n%v", events, want)
	}
}

func TestFlagPickupCTF4EA490EnemyPickupUsesCachedAndLiveValues(t *testing.T) {
	events := make([]string, 0, 40)
	sourceTeamLoads := 0
	sourceIndexLoads := 0
	netCode := uint32(0xaaaa5555)
	hooks := defaultFlagPickupCTFHooks4EA490()
	hooks.loadUpdate = func(obj int) int {
		events = append(events, map[int]string{1: "update-source", 2: "update-target"}[obj])
		return map[int]int{1: 101, 2: 202}[obj]
	}
	hooks.flagIndex = func(obj int) uint32 {
		sourceIndexLoads++
		events = append(events, "source-index")
		if sourceIndexLoads == 1 {
			return 0x123
		}
		return 0x456
	}
	hooks.loadTeamID = func(obj int) uint8 {
		sourceTeamLoads++
		events = append(events, "source-team")
		if sourceTeamLoads == 1 {
			return 7
		}
		return 8
	}
	hooks.teamsSame = func(int, int) int32 { events = append(events, "same"); return 0 }
	hooks.inventoryHolder = func(source int) int { events = append(events, "holder"); return 0 }
	hooks.teamByID = func(id uint8) int {
		events = append(events, "team-by-id")
		if id != 8 {
			t.Fatalf("live team id = %d", id)
		}
		return 9
	}
	hooks.teamEligible = func(team int) int32 { events = append(events, "eligible"); return 1 }
	hooks.firstInventory = func(target int) int { events = append(events, "first-item"); return 3 }
	hooks.loadClass = func(item int) uint32 {
		events = append(events, map[int]string{3: "class-3", 4: "class-4"}[item])
		if item == 4 {
			return 0x10000000
		}
		return 0
	}
	hooks.nextInventory = func(item int) int {
		events = append(events, "next-item")
		if item != 3 {
			t.Fatalf("next item = %d", item)
		}
		return 4
	}
	hooks.forceDrop = func(target, item int) {
		events = append(events, "force-drop")
		if target != 2 || item != 4 {
			t.Fatalf("drop args = %d/%d", target, item)
		}
	}
	hooks.finalizeDelete = func(source int) { events = append(events, "finalize") }
	hooks.inventoryPut = func(target, source int, mode int32) {
		events = append(events, "put")
		if target != 2 || source != 1 || mode != 1 {
			t.Fatalf("put args = %d/%d/%d", target, source, mode)
		}
	}
	hooks.markPlayerPickup = func(update int, flags uint32) {
		events = append(events, "mark-player")
		if update != 202 || flags != 1 {
			t.Fatalf("cached target update/flags = %d/%d", update, flags)
		}
	}
	hooks.reportObject = func(recipient uint32, source int) {
		events = append(events, "report-object")
		if recipient != 255 || source != 1 {
			t.Fatalf("report args = %d/%d", recipient, source)
		}
	}
	hooks.loadNetCode = func(target int) uint32 { events = append(events, "net"); return netCode }
	hooks.informFlag = func(code, gotNet, index uint32) {
		events = append(events, "inform")
		if code != 6 || gotNet != 0xaaaa5555 || index != 0x456 {
			t.Fatalf("inform = %d/%#x/%#x", code, gotNet, index)
		}
		netCode = 0x1234beef
	}
	hooks.unmarkMinimap = func(source int, flags uint32) { events = append(events, "unmark") }
	hooks.storeFlagState = func(update int, state uint32) {
		events = append(events, "store")
		if update != 101 || state != 0 {
			t.Fatalf("cached source update/state = %d/%d", update, state)
		}
	}
	hooks.flagStatus = func(teamID, status, index uint8, carrier uint16) int32 {
		events = append(events, "status")
		if teamID != 7 || status != 1 || index != 0x23 || carrier != 0xbeef {
			t.Fatalf("status = %d/%d/%#x/%#x", teamID, status, index, carrier)
		}
		return 0
	}
	hooks.purgeBuffs = func(target int) { events = append(events, "purge") }
	flagPickupCTF4EA490(1, 2, struct{ panicIfRead int }{}, hooks)
	want := []string{
		"update-target", "source-index", "source-team", "same", "update-source", "holder",
		"source-team", "team-by-id", "eligible", "first-item", "class-3", "next-item", "class-4",
		"force-drop", "finalize", "source-index", "put", "mark-player", "report-object", "net",
		"inform", "unmark", "store", "net", "status", "purge",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events =\n%v\nwant =\n%v", events, want)
	}
}

func TestFlagPickupCTF4EA490EnemyEarlyExits(t *testing.T) {
	t.Run("already carried", func(t *testing.T) {
		events := make([]string, 0, 8)
		hooks := defaultFlagPickupCTFHooks4EA490()
		hooks.loadUpdate = func(obj int) int { events = append(events, "update"); return obj }
		hooks.flagIndex = func(int) uint32 { events = append(events, "index"); return 0 }
		hooks.loadTeamID = func(int) uint8 { events = append(events, "team"); return 0 }
		hooks.teamsSame = func(int, int) int32 { events = append(events, "same"); return 0 }
		hooks.inventoryHolder = func(int) int { events = append(events, "holder"); return 99 }
		flagPickupCTF4EA490(1, 2, 0, hooks)
		want := []string{"update", "index", "team", "same", "update", "holder"}
		if !reflect.DeepEqual(events, want) {
			t.Fatalf("events = %v, want %v", events, want)
		}
	})
	for _, tc := range []struct {
		name     string
		team     int
		eligible int32
		wantElig bool
	}{
		{"missing team", 0, 1, false},
		{"empty team", 4, 0, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			hooks := defaultFlagPickupCTFHooks4EA490()
			hooks.loadUpdate = func(obj int) int { return obj }
			hooks.loadTeamID = func(int) uint8 { return 7 }
			hooks.teamByID = func(uint8) int { return tc.team }
			calledEligible := false
			hooks.teamEligible = func(int) int32 { calledEligible = true; return tc.eligible }
			hooks.priorityMessage = func(target int, message string, arg uint32) {
				if target != 2 || message != "objcoll.c:FlagNoTeam" || arg != 0 {
					t.Fatalf("message args = %d/%q/%d", target, message, arg)
				}
			}
			flagPickupCTF4EA490(1, 2, 0, hooks)
			if calledEligible != tc.wantElig {
				t.Fatalf("eligible called = %t, want %t", calledEligible, tc.wantElig)
			}
		})
	}
}

func TestFlagPickupCTF4EA490ZeroThresholdContinuesWithLiveNext(t *testing.T) {
	hooks := defaultFlagPickupCTFHooks4EA490()
	hooks.loadUpdate = func(obj int) int { return obj }
	hooks.loadTeamID = func(int) uint8 { return 1 }
	hooks.teamsSame = func(int, int) int32 { return 1 }
	hooks.firstInventory = func(int) int { return 3 }
	hooks.loadClass = func(item int) uint32 {
		if item == 3 {
			return 0x10000000
		}
		return 0
	}
	next := 4
	hooks.detachInventory = func(int, int) { next = 5 }
	hooks.nextInventory = func(item int) int {
		if item == 3 {
			return next
		}
		return 0
	}
	var classes []int
	hooks.loadClass = func(item int) uint32 {
		classes = append(classes, item)
		if item == 3 {
			return 0x10000000
		}
		return 0
	}
	flagPickupCTF4EA490(1, 2, 0, hooks)
	if !reflect.DeepEqual(classes, []int{3, 5}) {
		t.Fatalf("class objects = %v, want [3 5]", classes)
	}
}
