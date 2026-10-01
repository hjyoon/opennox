package server

import (
	"fmt"
	"math"
	"reflect"
	"slices"
	"testing"
)

type playerDieQuestTestPlayer54D2B0 struct {
	stage, generators, monsters, secrets uint32
	index                                uint8
}

type playerDieQuestTestUpdate54D2B0 struct {
	player  *playerDieQuestTestPlayer54D2B0
	lives   uint32
	frame   uint32
	markers [256]byte
}

type playerDieQuestTestUnit54D2B0 struct {
	update *playerDieQuestTestUpdate54D2B0
}

type playerDieQuestTestHooks54D2B0 = playerDieQuestHooks54D2B0[*playerDieQuestTestUnit54D2B0, *playerDieQuestTestUpdate54D2B0, *playerDieQuestTestPlayer54D2B0, *int]

func playerDieQuestTestFixture54D2B0(t *testing.T, trace *[]string, failAt int) (*playerDieQuestTestUnit54D2B0, playerDieQuestTestHooks54D2B0, *int) {
	t.Helper()
	p := &playerDieQuestTestPlayer54D2B0{0x12345678, 0x89abcdef, 0xfedcba98, 0x76543210, 255}
	u := &playerDieQuestTestUnit54D2B0{&playerDieQuestTestUpdate54D2B0{player: p, frame: 0x99887766}}
	result := new(int)
	hit := func(name string) {
		*trace = append(*trace, name)
		if failAt != 0 && len(*trace) == failAt {
			panic(name)
		}
	}
	h := playerDieQuestTestHooks54D2B0{
		gameFlag: func(flag uint32) int32 {
			hit("flag")
			if flag != 0x1000 {
				t.Fatalf("flag = %#x", flag)
			}
			return 1
		},
		loadExtraLives:  func(update *playerDieQuestTestUpdate54D2B0) uint32 { hit("lives"); return update.lives },
		storeExtraLives: func(update *playerDieQuestTestUpdate54D2B0, value uint32) { hit("store-lives"); update.lives = value },
		recordDeath: func(unit *playerDieQuestTestUnit54D2B0) *int {
			hit("record")
			if unit != u {
				t.Fatal("record unit")
			}
			return result
		},
		frame: func() uint32 { hit("frame"); return 0xaabbccdd },
		loadPlayer: func(update *playerDieQuestTestUpdate54D2B0) *playerDieQuestTestPlayer54D2B0 {
			hit("player")
			return update.player
		},
		storeFrame: func(update *playerDieQuestTestUpdate54D2B0, value uint32) { hit("time"); update.frame = value },
		loadStage:  func(player *playerDieQuestTestPlayer54D2B0) uint16 { hit("stage"); return uint16(player.stage) },
		loadGenerators: func(player *playerDieQuestTestPlayer54D2B0) uint16 {
			hit("generators")
			return uint16(player.generators)
		},
		loadMonsters:    func(player *playerDieQuestTestPlayer54D2B0) uint16 { hit("monsters"); return uint16(player.monsters) },
		loadSecrets:     func(player *playerDieQuestTestPlayer54D2B0) uint16 { hit("secrets"); return uint16(player.secrets) },
		loadPlayerIndex: func(player *playerDieQuestTestPlayer54D2B0) uint8 { hit("index"); return player.index },
		sendStats: func(index uint8, packet [14]byte) {
			hit("send")
			want := [14]byte{0xf0, 2, 0xef, 0xcd, 0x10, 0x32, 0x98, 0xba, 0x78, 0x56, 0, 0, 0, 0}
			if index != 255 || packet != want {
				t.Fatalf("packet = %d/%x, want 255/%x", index, packet, want)
			}
		},
		resetPlayer: func(unit *playerDieQuestTestUnit54D2B0) {
			hit("reset")
			if unit != u {
				t.Fatal("reset unit")
			}
		},
		penalty: func(unit *playerDieQuestTestUnit54D2B0) {
			hit("penalty")
			if unit != u {
				t.Fatal("penalty unit")
			}
		},
		balanceFloat: func(key string) float32 {
			hit("balance")
			if key != "QuestGameStartingExtraLives" {
				t.Fatal(key)
			}
			return 3.5
		},
		floatToInt:         func(value float32) int32 { hit("convert"); return playerUnitInitFloatToInt4EFE80(value) },
		loadExtraLivesByte: func(update *playerDieQuestTestUpdate54D2B0) uint8 { hit("life-byte"); return uint8(update.lives) },
		storeRespawnMarker: func(update *playerDieQuestTestUpdate54D2B0, index, value uint8) {
			hit("marker")
			update.markers[index] = value
		},
	}
	return u, h, result
}

var playerDieQuestZeroTrace54D2B0 = []string{
	"flag", "lives", "frame", "player", "time", "stage", "generators", "monsters", "secrets", "index",
	"send", "reset", "penalty", "balance", "convert", "store-lives", "player", "life-byte", "index", "marker",
}

func TestPlayerDieQuest54D2B0ZeroLivesExactOrderPacketAndResult(t *testing.T) {
	var trace []string
	u, h, _ := playerDieQuestTestFixture54D2B0(t, &trace, 0)
	p := u.update.player
	got := playerDieQuest54D2B0(u, u.update, h)
	if !reflect.DeepEqual(trace, playerDieQuestZeroTrace54D2B0) || got.kind != playerDieQuestPlayerReturn54D2B0 || got.player != p || got.recorded != nil {
		t.Fatalf("trace/result = %v/%+v", trace, got)
	}
	if u.update.frame != 0xaabbccdd || u.update.lives != 4 || u.update.markers[255] != 4 {
		t.Fatalf("final update = %+v", u.update)
	}
}

func TestPlayerDieQuest54D2B0UnsignedLivesOnlyDecrementAndRecord(t *testing.T) {
	for _, lives := range []uint32{1, 2, 3, 255, 256, 257, 0x7fffffff, 0x80000000, 0xfffffffe, 0xffffffff} {
		t.Run(fmt.Sprintf("%08x", lives), func(t *testing.T) {
			var trace []string
			u, h, recorded := playerDieQuestTestFixture54D2B0(t, &trace, 0)
			u.update.lives = lives
			before := *u.update
			got := playerDieQuest54D2B0(u, u.update, h)
			before.lives--
			if !reflect.DeepEqual(trace, []string{"flag", "lives", "store-lives", "record"}) || *u.update != before || got.kind != playerDieQuestRecordReturn54D2B0 || got.recorded != recorded || got.player != nil {
				t.Fatalf("trace/result/update = %v/%+v/%+v", trace, got, u.update)
			}
		})
	}
}

func TestPlayerDieQuest54D2B0FlagZeroNeedsNoQuestBindings(t *testing.T) {
	got := playerDieQuest54D2B0(nil, nil, playerDieQuestTestHooks54D2B0{gameFlag: func(uint32) int32 { return 0 }})
	if got.kind != playerDieQuestFlagReturn54D2B0 || got.flag != 0 || got.player != nil || got.recorded != nil {
		t.Fatalf("flag result = %+v", got)
	}
	for _, flag := range []int32{1, 2, -1, math.MinInt32, math.MaxInt32} {
		var trace []string
		u, h, recorded := playerDieQuestTestFixture54D2B0(t, &trace, 0)
		u.update.lives = 1
		h.gameFlag = func(uint32) int32 { return flag }
		got := playerDieQuest54D2B0(u, u.update, h)
		if got.recorded != recorded || u.update.lives != 0 {
			t.Fatalf("nonzero flag %#x rejected", flag)
		}
	}
}

func TestPlayerDieQuest54D2B0LiveLinksAndCachedPacketPlayer(t *testing.T) {
	var trace []string
	u, h, _ := playerDieQuestTestFixture54D2B0(t, &trace, 0)
	entry := u.update
	packetPlayer := &playerDieQuestTestPlayer54D2B0{1, 2, 3, 4, 7}
	latePlayer := &playerDieQuestTestPlayer54D2B0{5, 6, 7, 8, 11}
	finalPlayer := &playerDieQuestTestPlayer54D2B0{9, 10, 11, 12, 31}
	live := &playerDieQuestTestUpdate54D2B0{player: latePlayer, lives: 55, frame: 88}
	h.frame = func() uint32 { entry.player = packetPlayer; return 77 }
	h.storeFrame = func(update *playerDieQuestTestUpdate54D2B0, value uint32) {
		if update != entry {
			t.Fatal("timestamp update reloaded")
		}
		update.frame, update.player = value, latePlayer
	}
	h.loadStage = func(p *playerDieQuestTestPlayer54D2B0) uint16 {
		if p != packetPlayer {
			t.Fatal("packet player reloaded after timestamp store")
		}
		p.generators = 0x12345678
		return uint16(p.stage)
	}
	h.sendStats = func(index uint8, packet [14]byte) {
		want := [14]byte{0xf0, 2, 0x78, 0x56, 4, 0, 3, 0, 1, 0, 0, 0, 0, 0}
		if index != 7 || packet != want || entry.frame != 77 {
			t.Fatalf("cached stats = %d/%x", index, packet)
		}
		u.update = live
	}
	h.resetPlayer = func(unit *playerDieQuestTestUnit54D2B0) {
		if unit != u || unit.update != live {
			t.Fatal("reset must use current unit")
		}
		entry.player = finalPlayer
	}
	h.penalty = func(unit *playerDieQuestTestUnit54D2B0) {
		if unit != u || unit.update != live {
			t.Fatal("penalty must use current unit")
		}
	}
	h.balanceFloat = func(string) float32 { return 257.5 }
	h.storeExtraLives = func(update *playerDieQuestTestUpdate54D2B0, lives uint32) {
		if update != entry || lives != 258 {
			t.Fatalf("cached life store = %p/%d", update, lives)
		}
		update.lives = lives
	}
	h.loadExtraLivesByte = func(update *playerDieQuestTestUpdate54D2B0) uint8 {
		// Player was already loaded; changing the link here must not change
		// the final marker's recipient, but that cached Player's index is live.
		update.player = latePlayer
		finalPlayer.index = 29
		update.lives = 0xabcdef81
		return uint8(update.lives)
	}
	got := playerDieQuest54D2B0(u, entry, h)
	if got.player != finalPlayer || got.kind != playerDieQuestPlayerReturn54D2B0 || entry.markers[29] != 0x81 || entry.lives != 0xabcdef81 || *live != (playerDieQuestTestUpdate54D2B0{player: latePlayer, lives: 55, frame: 88}) {
		t.Fatalf("live/cached result = %+v entry=%+v live=%+v", got, entry, live)
	}
}

func TestPlayerDieQuest54D2B0Binary32ConversionAndByteNarrowing(t *testing.T) {
	for _, tc := range []struct {
		value float32
		want  uint32
	}{
		{0, 0}, {0.5, 0}, {1.5, 2}, {2.5, 2}, {3.5, 4}, {-0.5, 0}, {-1.5, 0xfffffffe},
		{-2.5, 0xfffffffe}, {-3.5, 0xfffffffc}, {255.5, 256}, {256.5, 256}, {257.5, 258},
		{math.MaxFloat32, 0x80000000}, {-math.MaxFloat32, 0x80000000}, {2147483520, 0x7fffff80},
		{2147483648, 0x80000000}, {-2147483648, 0x80000000}, {float32(math.NaN()), 0x80000000},
		{float32(math.Inf(1)), 0x80000000}, {float32(math.Inf(-1)), 0x80000000},
	} {
		t.Run(fmt.Sprintf("%08x", math.Float32bits(tc.value)), func(t *testing.T) {
			var trace []string
			u, h, _ := playerDieQuestTestFixture54D2B0(t, &trace, 0)
			h.balanceFloat = func(string) float32 { return tc.value }
			playerDieQuest54D2B0(u, u.update, h)
			if u.update.lives != tc.want || u.update.markers[255] != byte(tc.want) {
				t.Fatalf("lives/marker = %#x/%#x, want %#x/%#x", u.update.lives, u.update.markers[255], tc.want, byte(tc.want))
			}
		})
	}
}

func TestPlayerDieQuest54D2B0CallbackFaultPrefixes(t *testing.T) {
	for fault := 1; fault <= len(playerDieQuestZeroTrace54D2B0); fault++ {
		t.Run(fmt.Sprint(fault), func(t *testing.T) {
			var trace []string
			u, h, _ := playerDieQuestTestFixture54D2B0(t, &trace, fault)
			var recovered any
			func() { defer func() { recovered = recover() }(); playerDieQuest54D2B0(u, u.update, h) }()
			if recovered == nil || !reflect.DeepEqual(trace, playerDieQuestZeroTrace54D2B0[:fault]) {
				t.Fatalf("fault prefix %v/%v", recovered, trace)
			}
			wantFrame := uint32(0x99887766)
			if fault > 5 {
				wantFrame = 0xaabbccdd
			}
			wantLives := uint32(0)
			if fault > 16 {
				wantLives = 4
			}
			if u.update.frame != wantFrame || u.update.lives != wantLives || u.update.markers != [256]byte{} {
				t.Fatalf("partial state = %+v", u.update)
			}
		})
	}
	for fault := 1; fault <= 4; fault++ {
		var trace []string
		u, h, _ := playerDieQuestTestFixture54D2B0(t, &trace, fault)
		u.update.lives = 1
		func() {
			defer func() {
				if recover() == nil {
					t.Fatal("expected positive-path fault")
				}
			}()
			playerDieQuest54D2B0(u, u.update, h)
		}()
		if !reflect.DeepEqual(trace, []string{"flag", "lives", "store-lives", "record"}[:fault]) {
			t.Fatal(trace)
		}
		wantLives := uint32(1)
		if fault > 3 {
			wantLives = 0
		}
		if u.update.lives != wantLives {
			t.Fatalf("partial decrement = %d", u.update.lives)
		}
	}
}

func TestPlayerDieQuest54D2B0MissingHooksFaultAtOriginalCall(t *testing.T) {
	for name, prefix := range map[string]int{
		"gameFlag": 0, "loadExtraLives": 1, "frame": 2, "loadPlayer": 3, "storeFrame": 4,
		"loadStage": 5, "loadGenerators": 6, "loadMonsters": 7, "loadSecrets": 8, "loadPlayerIndex": 9,
		"sendStats": 10, "resetPlayer": 11, "penalty": 12, "balanceFloat": 13, "floatToInt": 14,
		"storeExtraLives": 15, "loadExtraLivesByte": 17, "storeRespawnMarker": 19,
	} {
		t.Run(name, func(t *testing.T) {
			var trace []string
			u, h, _ := playerDieQuestTestFixture54D2B0(t, &trace, 0)
			switch name {
			case "gameFlag":
				h.gameFlag = nil
			case "loadExtraLives":
				h.loadExtraLives = nil
			case "frame":
				h.frame = nil
			case "loadPlayer":
				h.loadPlayer = nil
			case "storeFrame":
				h.storeFrame = nil
			case "loadStage":
				h.loadStage = nil
			case "loadGenerators":
				h.loadGenerators = nil
			case "loadMonsters":
				h.loadMonsters = nil
			case "loadSecrets":
				h.loadSecrets = nil
			case "loadPlayerIndex":
				h.loadPlayerIndex = nil
			case "sendStats":
				h.sendStats = nil
			case "resetPlayer":
				h.resetPlayer = nil
			case "penalty":
				h.penalty = nil
			case "balanceFloat":
				h.balanceFloat = nil
			case "floatToInt":
				h.floatToInt = nil
			case "storeExtraLives":
				h.storeExtraLives = nil
			case "loadExtraLivesByte":
				h.loadExtraLivesByte = nil
			case "storeRespawnMarker":
				h.storeRespawnMarker = nil
			}
			var recovered any
			func() { defer func() { recovered = recover() }(); playerDieQuest54D2B0(u, u.update, h) }()
			if recovered == nil || !slices.Equal(trace, playerDieQuestZeroTrace54D2B0[:prefix]) {
				t.Fatalf("missing %s fault/prefix = %v/%v", name, recovered, trace)
			}
		})
	}
	var trace []string
	u, h, _ := playerDieQuestTestFixture54D2B0(t, &trace, 0)
	u.update.lives, h.recordDeath = 1, nil
	func() {
		defer func() {
			if recover() == nil || !slices.Equal(trace, []string{"flag", "lives", "store-lives"}) || u.update.lives != 0 {
				t.Fatal("missing record hook did not preserve decrement and fault prefix")
			}
		}()
		playerDieQuest54D2B0(u, u.update, h)
	}()
}
