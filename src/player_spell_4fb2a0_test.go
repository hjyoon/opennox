package opennox

import (
	"fmt"
	"reflect"
	"testing"
)

type playerSpellTestObject4FB2A0 struct {
	name string
}

type playerSpellTestLeaf4FB2A0 struct {
	name string
	id   int32
	root bool
}

type playerSpellTestPlayer4FB2A0 struct {
	name   string
	index  uint8
	levels map[int32]uint32
	target *playerSpellTestObject4FB2A0
	x      int32
	y      int32
}

type playerSpellTestUpdate4FB2A0 struct {
	leaf   *playerSpellTestLeaf4FB2A0
	player *playerSpellTestPlayer4FB2A0
	cursor *playerSpellTestObject4FB2A0
	state  uint8
}

type playerSpellTestEnv4FB2A0 struct {
	update       *playerSpellTestUpdate4FB2A0
	quest        bool
	offensive    map[int32]bool
	noReport     map[int32]bool
	events       []string
	precheck     int32
	castCheck    int32
	mana         int32
	castSucceeds bool
}

func playerSpellTestObjectName4FB2A0(v *playerSpellTestObject4FB2A0) string {
	if v == nil {
		return "nil"
	}
	return v.name
}

func playerSpellTestLeafName4FB2A0(v *playerSpellTestLeaf4FB2A0) string {
	if v == nil {
		return "nil"
	}
	return v.name
}

func playerSpellTestPlayerName4FB2A0(v *playerSpellTestPlayer4FB2A0) string {
	if v == nil {
		return "nil"
	}
	return v.name
}

func playerSpellTestHooks4FB2A0(env *playerSpellTestEnv4FB2A0) playerSpellHooks4FB2A0[
	*playerSpellTestObject4FB2A0,
	*playerSpellTestUpdate4FB2A0,
	*playerSpellTestPlayer4FB2A0,
	*playerSpellTestLeaf4FB2A0,
	string,
] {
	record := func(format string, args ...any) {
		env.events = append(env.events, fmt.Sprintf(format, args...))
	}
	return playerSpellHooks4FB2A0[
		*playerSpellTestObject4FB2A0,
		*playerSpellTestUpdate4FB2A0,
		*playerSpellTestPlayer4FB2A0,
		*playerSpellTestLeaf4FB2A0,
		string,
	]{
		loadUpdateData: func(unit *playerSpellTestObject4FB2A0) *playerSpellTestUpdate4FB2A0 {
			record("update:%s", playerSpellTestObjectName4FB2A0(unit))
			return env.update
		},
		loadLeaf: func(update *playerSpellTestUpdate4FB2A0) *playerSpellTestLeaf4FB2A0 {
			record("leaf:%s", playerSpellTestLeafName4FB2A0(update.leaf))
			return update.leaf
		},
		isRootLeaf: func(leaf *playerSpellTestLeaf4FB2A0) bool {
			root := leaf != nil && leaf.root
			record("root:%s:%t", playerSpellTestLeafName4FB2A0(leaf), root)
			return root
		},
		loadSpellID: func(leaf *playerSpellTestLeaf4FB2A0) int32 {
			record("id:%s:%d", leaf.name, leaf.id)
			return leaf.id
		},
		hasGameFlag: func(flag uint32) bool {
			record("game:%#x:%t", flag, env.quest)
			return env.quest
		},
		loadCursorObj: func(update *playerSpellTestUpdate4FB2A0) *playerSpellTestObject4FB2A0 {
			record("cursor-object:%s", playerSpellTestObjectName4FB2A0(update.cursor))
			return update.cursor
		},
		hasSpellFlags: func(id int32, flag uint32) bool {
			var value bool
			switch flag {
			case playerSpellOffensive4FB2A0:
				value = env.offensive[id]
			case playerSpellNoReport4FB2A0:
				value = env.noReport[id]
			}
			record("flags:%d:%#x:%t", id, flag, value)
			return value
		},
		isEnemy: func(unit, target *playerSpellTestObject4FB2A0) bool {
			record("enemy:%s:%s:true", playerSpellTestObjectName4FB2A0(unit), playerSpellTestObjectName4FB2A0(target))
			return true
		},
		loadPlayer: func(update *playerSpellTestUpdate4FB2A0) *playerSpellTestPlayer4FB2A0 {
			record("player:%s", playerSpellTestPlayerName4FB2A0(update.player))
			return update.player
		},
		loadSpellLevel: func(player *playerSpellTestPlayer4FB2A0, id int32) uint32 {
			value := player.levels[id]
			record("level:%s:%d:%d", player.name, id, value)
			return value
		},
		precheck: func(unit *playerSpellTestObject4FB2A0, id int32) int32 {
			record("precheck:%s:%d:%d", unit.name, id, env.precheck)
			return env.precheck
		},
		checkCantCast: func(unit *playerSpellTestObject4FB2A0, id, bypass int32) int32 {
			record("cast-check:%s:%d:%d:%d", unit.name, id, bypass, env.castCheck)
			return env.castCheck
		},
		loadPlayerInd: func(player *playerSpellTestPlayer4FB2A0) uint8 {
			record("player-index:%s:%d", player.name, player.index)
			return player.index
		},
		informResult: func(index, code uint8, value int32) {
			record("inform-result:%d:%d:%d", index, code, value)
		},
		informSpell: func(index, code uint8, leaf *playerSpellTestLeaf4FB2A0) {
			record("inform-spell:%d:%d:%s:%d", index, code, leaf.name, leaf.id)
		},
		audioEvent: func(sound int32, unit *playerSpellTestObject4FB2A0, kind, code int32) {
			record("audio:%d:%s:%d:%d", sound, unit.name, kind, code)
		},
		chargeMana: func(unit *playerSpellTestObject4FB2A0, id, amount int32) int32 {
			record("mana:%s:%d:%d:%d", unit.name, id, amount, env.mana)
			return env.mana
		},
		loadCastTarget: func(player *playerSpellTestPlayer4FB2A0) *playerSpellTestObject4FB2A0 {
			record("target:%s:%s", player.name, playerSpellTestObjectName4FB2A0(player.target))
			return player.target
		},
		loadCursorPos: func(player *playerSpellTestPlayer4FB2A0) (int32, int32) {
			record("cursor:%s:%d:%d", player.name, player.x, player.y)
			return player.x, player.y
		},
		castSpell: func(id int32, unit *playerSpellTestObject4FB2A0, arg playerSpellArg4FB2A0[*playerSpellTestObject4FB2A0]) bool {
			record("cast:%d:%s:%s:%g:%g:%t", id, unit.name, playerSpellTestObjectName4FB2A0(arg.target), arg.posX, arg.posY, env.castSucceeds)
			return env.castSucceeds
		},
		refundMana: func(unit *playerSpellTestObject4FB2A0, mana int32) {
			record("refund:%s:%d", unit.name, mana)
		},
		loadState: func(update *playerSpellTestUpdate4FB2A0) uint8 {
			record("state:%d", update.state)
			return update.state
		},
		setState: func(unit *playerSpellTestObject4FB2A0, state uint8) {
			record("set-state:%s:%d", unit.name, state)
			env.update.state = state
		},
		unknownMessage: func() string {
			record("unknown-message")
			return "SpellUnknown"
		},
		lineMessage: func(unit *playerSpellTestObject4FB2A0, message string) {
			record("line:%s:%s", unit.name, message)
		},
		reportSpell: func(index uint8, id int32, status uint8) {
			record("report:%d:%d:%d", index, id, status)
		},
	}
}

func TestPlayerSpell4FB2A0UnknownAndRoot(t *testing.T) {
	unit := &playerSpellTestObject4FB2A0{name: "caster"}
	t.Run("unknown", func(t *testing.T) {
		env := &playerSpellTestEnv4FB2A0{
			update: &playerSpellTestUpdate4FB2A0{state: playerSpellReadyState4FB2A0},
		}
		playerSpell4FB2A0(unit, playerSpellTestHooks4FB2A0(env))
		want := []string{
			"update:caster", "leaf:nil", "root:nil:false", "leaf:nil", "state:2",
			"set-state:caster:13", "unknown-message", "line:caster:SpellUnknown",
		}
		if !reflect.DeepEqual(env.events, want) {
			t.Fatalf("events:\n got %v\nwant %v", env.events, want)
		}
	})

	t.Run("root", func(t *testing.T) {
		root := &playerSpellTestLeaf4FB2A0{name: "root", id: 4, root: true}
		player := &playerSpellTestPlayer4FB2A0{name: "player", index: 9}
		env := &playerSpellTestEnv4FB2A0{
			update: &playerSpellTestUpdate4FB2A0{leaf: root, player: player, state: 7},
		}
		playerSpell4FB2A0(unit, playerSpellTestHooks4FB2A0(env))
		want := []string{
			"update:caster", "leaf:root", "root:root:true", "state:7",
			"leaf:root", "player:player", "id:root:4", "player-index:player:9", "report:9:4:0",
		}
		if !reflect.DeepEqual(env.events, want) {
			t.Fatalf("events:\n got %v\nwant %v", env.events, want)
		}
	})
}

func TestPlayerSpell4FB2A0FriendlyOffensiveTargetReturnsImmediately(t *testing.T) {
	unit := &playerSpellTestObject4FB2A0{name: "caster"}
	target := &playerSpellTestObject4FB2A0{name: "friendly"}
	leaf := &playerSpellTestLeaf4FB2A0{name: "spell", id: 7}
	env := &playerSpellTestEnv4FB2A0{
		update:    &playerSpellTestUpdate4FB2A0{leaf: leaf, cursor: target, state: 2},
		offensive: map[int32]bool{7: true},
	}
	hooks := playerSpellTestHooks4FB2A0(env)
	hooks.isEnemy = func(gotUnit, gotTarget *playerSpellTestObject4FB2A0) bool {
		env.events = append(env.events, "enemy:caster:friendly:false")
		return false
	}

	playerSpell4FB2A0(unit, hooks)
	want := []string{
		"update:caster", "leaf:spell", "root:spell:false", "leaf:spell", "id:spell:7",
		"game:0x1000:false", "leaf:spell", "cursor-object:friendly", "id:spell:7",
		"flags:7:0x20:true", "enemy:caster:friendly:false",
	}
	if !reflect.DeepEqual(env.events, want) {
		t.Fatalf("events:\n got %v\nwant %v", env.events, want)
	}
	if env.update.state != 2 {
		t.Fatalf("state = %d, want unchanged 2", env.update.state)
	}
}

func TestPlayerSpell4FB2A0ReloadsLeafAndPlayerAcrossCallbacks(t *testing.T) {
	unit := &playerSpellTestObject4FB2A0{name: "caster"}
	target2 := &playerSpellTestObject4FB2A0{name: "target-2"}
	target3 := &playerSpellTestObject4FB2A0{name: "target-3"}
	leaves := []*playerSpellTestLeaf4FB2A0{
		{name: "leaf-10", id: 10},
		{name: "leaf-11", id: 11},
		{name: "leaf-12", id: 12},
		{name: "leaf-13", id: 13},
		{name: "leaf-14", id: 14},
	}
	players := []*playerSpellTestPlayer4FB2A0{
		{name: "player-1", index: 1, levels: map[int32]uint32{10: 1}},
		{name: "player-2", index: 2, target: target2},
		{name: "player-3", index: 3, target: target3},
		{name: "player-4", index: 4, x: -2147483648, y: 16777217},
		{name: "player-5", index: 5},
		{name: "player-6", index: 6},
	}
	env := &playerSpellTestEnv4FB2A0{
		update:       &playerSpellTestUpdate4FB2A0{leaf: leaves[0], player: players[0], state: 2},
		quest:        true,
		offensive:    map[int32]bool{12: true},
		noReport:     make(map[int32]bool),
		mana:         37,
		castSucceeds: true,
	}
	hooks := playerSpellTestHooks4FB2A0(env)
	hooks.precheck = func(_ *playerSpellTestObject4FB2A0, id int32) int32 {
		env.events = append(env.events, fmt.Sprintf("precheck:caster:%d:0", id))
		env.update.leaf = leaves[1]
		return 0
	}
	hooks.checkCantCast = func(_ *playerSpellTestObject4FB2A0, id, bypass int32) int32 {
		env.events = append(env.events, fmt.Sprintf("cast-check:caster:%d:%d:0", id, bypass))
		env.update.leaf = leaves[2]
		return 0
	}
	hooks.chargeMana = func(_ *playerSpellTestObject4FB2A0, id, amount int32) int32 {
		env.events = append(env.events, fmt.Sprintf("mana:caster:%d:%d:37", id, amount))
		env.update.player = players[1]
		return 37
	}
	hooks.hasSpellFlags = func(id int32, flag uint32) bool {
		value := flag == playerSpellOffensive4FB2A0 && id == 12
		env.events = append(env.events, fmt.Sprintf("flags:%d:%#x:%t", id, flag, value))
		if flag == playerSpellOffensive4FB2A0 {
			env.update.player = players[2]
		} else if flag == playerSpellNoReport4FB2A0 {
			env.update.player = players[5]
			env.update.leaf = leaves[4]
		}
		return value
	}
	hooks.isEnemy = func(_, target *playerSpellTestObject4FB2A0) bool {
		env.events = append(env.events, "enemy:caster:"+target.name+":true")
		env.update.player = players[3]
		return true
	}
	hooks.castSpell = func(id int32, _ *playerSpellTestObject4FB2A0, arg playerSpellArg4FB2A0[*playerSpellTestObject4FB2A0]) bool {
		env.events = append(env.events, fmt.Sprintf("cast:%d:caster:%s:%g:%g:true", id, arg.target.name, arg.posX, arg.posY))
		env.update.player = players[4]
		env.update.leaf = leaves[3]
		return true
	}

	playerSpell4FB2A0(unit, hooks)
	want := []string{
		"update:caster", "leaf:leaf-10", "root:leaf-10:false", "leaf:leaf-10", "id:leaf-10:10",
		"game:0x1000:true", "leaf:leaf-10", "player:player-1", "id:leaf-10:10", "level:player-1:10:1",
		"precheck:caster:10:0", "leaf:leaf-11", "id:leaf-11:11", "cast-check:caster:11:0:0",
		"leaf:leaf-12", "id:leaf-12:12", "mana:caster:12:1:37",
		"player:player-2", "target:player-2:target-2", "game:0x1000:true", "leaf:leaf-12", "id:leaf-12:12",
		"flags:12:0x20:true", "player:player-3", "target:player-3:target-3", "enemy:caster:target-3:true",
		"player:player-4", "cursor:player-4:-2147483648:16777217", "leaf:leaf-12", "id:leaf-12:12",
		"cast:12:caster:target-2:-2.1474836e+09:1.6777216e+07:true", "player:player-5", "leaf:leaf-13",
		"player-index:player-5:5", "inform-spell:5:1:leaf-13:13", "state:2", "set-state:caster:13",
		"leaf:leaf-13", "id:leaf-13:13", "flags:13:0x100000:false", "leaf:leaf-14", "player:player-6",
		"id:leaf-14:14", "player-index:player-6:6", "report:6:14:15",
	}
	if !reflect.DeepEqual(env.events, want) {
		t.Fatalf("events:\n got %v\nwant %v", env.events, want)
	}
}

func TestPlayerSpell4FB2A0RejectedAndManaFailure(t *testing.T) {
	unit := &playerSpellTestObject4FB2A0{name: "caster"}
	for _, tc := range []struct {
		name       string
		castCheck  int32
		mana       int32
		wantResult int32
		wantSound  int32
	}{
		{name: "cast check", castCheck: 6, mana: 9, wantResult: 6, wantSound: playerSpellFizzleSound4FB2A0},
		{name: "mana", mana: -1, wantResult: playerSpellNoMana4FB2A0, wantSound: playerSpellNoManaSound4FB2A0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			leaf := &playerSpellTestLeaf4FB2A0{name: "spell", id: 20}
			player := &playerSpellTestPlayer4FB2A0{name: "player", index: 7, levels: map[int32]uint32{20: 1}}
			env := &playerSpellTestEnv4FB2A0{
				update:    &playerSpellTestUpdate4FB2A0{leaf: leaf, player: player},
				quest:     true,
				castCheck: tc.castCheck,
				mana:      tc.mana,
			}
			playerSpell4FB2A0(unit, playerSpellTestHooks4FB2A0(env))
			inform := fmt.Sprintf("inform-result:7:0:%d", tc.wantResult)
			audio := fmt.Sprintf("audio:%d:caster:0:0", tc.wantSound)
			for _, want := range []string{inform, audio, "report:7:20:0"} {
				if !containsPlayerSpellEvent4FB2A0(env.events, want) {
					t.Fatalf("events %v do not contain %q", env.events, want)
				}
			}
			if containsPlayerSpellPrefix4FB2A0(env.events, "cast:") {
				t.Fatalf("rejected spell was cast: %v", env.events)
			}
		})
	}
}

func TestPlayerSpell4FB2A0CastFailureRefundsAndNoReportFlagSuppresses(t *testing.T) {
	unit := &playerSpellTestObject4FB2A0{name: "caster"}
	newEnv := func() *playerSpellTestEnv4FB2A0 {
		leaf := &playerSpellTestLeaf4FB2A0{name: "spell", id: 25}
		player := &playerSpellTestPlayer4FB2A0{name: "player", index: 8, levels: map[int32]uint32{25: 1}}
		return &playerSpellTestEnv4FB2A0{
			update:   &playerSpellTestUpdate4FB2A0{leaf: leaf, player: player},
			quest:    true,
			mana:     44,
			noReport: make(map[int32]bool),
		}
	}

	t.Run("cast failure", func(t *testing.T) {
		env := newEnv()
		playerSpell4FB2A0(unit, playerSpellTestHooks4FB2A0(env))
		for _, want := range []string{"refund:caster:44", "report:8:25:0"} {
			if !containsPlayerSpellEvent4FB2A0(env.events, want) {
				t.Fatalf("events %v do not contain %q", env.events, want)
			}
		}
	})

	t.Run("successful no-report spell", func(t *testing.T) {
		env := newEnv()
		env.castSucceeds = true
		env.noReport[25] = true
		playerSpell4FB2A0(unit, playerSpellTestHooks4FB2A0(env))
		if !containsPlayerSpellEvent4FB2A0(env.events, "inform-spell:8:1:spell:25") {
			t.Fatalf("events %v do not contain success inform", env.events)
		}
		if containsPlayerSpellPrefix4FB2A0(env.events, "report:") {
			t.Fatalf("no-report spell was reported: %v", env.events)
		}
	})
}

func containsPlayerSpellEvent4FB2A0(events []string, want string) bool {
	for _, event := range events {
		if event == want {
			return true
		}
	}
	return false
}

func containsPlayerSpellPrefix4FB2A0(events []string, prefix string) bool {
	for _, event := range events {
		if len(event) >= len(prefix) && event[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}
