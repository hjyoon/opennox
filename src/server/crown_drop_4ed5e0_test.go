package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type crownDropTestPoint4ED5E0 struct {
	name string
}

type crownDropTestCrownData4ED5E0 struct {
	name   string
	target *crownDropTestObject4ED5E0
}

type crownDropTestPlayerData4ED5E0 struct {
	name  string
	frame uint32
}

type crownDropTestObject4ED5E0 struct {
	name       string
	teamID     uint8
	netCode    uint32
	crownData  *crownDropTestCrownData4ED5E0
	playerData *crownDropTestPlayerData4ED5E0
	next       *crownDropTestObject4ED5E0
}

type crownDropTestWorld4ED5E0 struct {
	owner, crown *crownDropTestObject4ED5E0
	point        *crownDropTestPoint4ED5E0
	first        *crownDropTestObject4ED5E0

	ownerArg *crownDropTestObject4ED5E0
	crownArg *crownDropTestObject4ED5E0
	pointArg *crownDropTestPoint4ED5E0

	gameResult     int32
	gameplayResult int32
	defaultResult  int32
	hasTeamResult  int32
	frame          uint32

	events  []string
	faultAt int

	afterFirst       func(*crownDropTestWorld4ED5E0)
	afterPlayerData  func(*crownDropTestWorld4ED5E0, *crownDropTestObject4ED5E0)
	afterTeamCompare func(*crownDropTestWorld4ED5E0, *crownDropTestObject4ED5E0)
	afterBuffOff     func(*crownDropTestWorld4ED5E0)
	afterNetCode     func(*crownDropTestWorld4ED5E0)
	afterHasTeam     func(*crownDropTestWorld4ED5E0)
}

func (w *crownDropTestWorld4ED5E0) event(name string) {
	w.events = append(w.events, name)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(name)
	}
}

func crownDropObjectName4ED5E0(obj *crownDropTestObject4ED5E0) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func newCrownDropTestWorld4ED5E0() *crownDropTestWorld4ED5E0 {
	owner := &crownDropTestObject4ED5E0{name: "owner", teamID: 7, netCode: 0x11223344}
	crownData := &crownDropTestCrownData4ED5E0{name: "crown-data"}
	crown := &crownDropTestObject4ED5E0{name: "crown", teamID: 7, crownData: crownData}
	second := &crownDropTestObject4ED5E0{
		name:       "second",
		teamID:     7,
		playerData: &crownDropTestPlayerData4ED5E0{name: "second-data", frame: 90},
	}
	first := &crownDropTestObject4ED5E0{
		name:       "first",
		teamID:     7,
		playerData: &crownDropTestPlayerData4ED5E0{name: "first-data", frame: 80},
		next:       second,
	}
	point := &crownDropTestPoint4ED5E0{name: "point"}
	return &crownDropTestWorld4ED5E0{
		owner: owner, crown: crown, point: point, first: first,
		ownerArg: owner, crownArg: crown, pointArg: point,
		gameResult: 1, gameplayResult: 1, defaultResult: 1, hasTeamResult: 1,
		frame: 100,
	}
}

func (w *crownDropTestWorld4ED5E0) hooks() crownDropHooks4ED5E0[
	*crownDropTestObject4ED5E0,
	*crownDropTestCrownData4ED5E0,
	*crownDropTestPlayerData4ED5E0,
	*crownDropTestPoint4ED5E0,
] {
	return crownDropHooks4ED5E0[
		*crownDropTestObject4ED5E0,
		*crownDropTestCrownData4ED5E0,
		*crownDropTestPlayerData4ED5E0,
		*crownDropTestPoint4ED5E0,
	]{
		gameFlag: func(flag uint32) int32 {
			w.event(fmt.Sprintf("game:%d", flag))
			return w.gameResult
		},
		gameplayFlag: func(flag uint32) int32 {
			w.event(fmt.Sprintf("gameplay:%d", flag))
			return w.gameplayResult
		},
		loadCrownArg: func() *crownDropTestObject4ED5E0 {
			w.event("crown-arg:" + crownDropObjectName4ED5E0(w.crownArg))
			return w.crownArg
		},
		loadFrame: func() uint32 {
			w.event(fmt.Sprintf("frame:%#x", w.frame))
			return w.frame
		},
		loadTeamID: func(obj *crownDropTestObject4ED5E0) uint8 {
			w.event(fmt.Sprintf("team:%s:%d", crownDropObjectName4ED5E0(obj), obj.teamID))
			return obj.teamID
		},
		loadCrownUpdate: func(obj *crownDropTestObject4ED5E0) *crownDropTestCrownData4ED5E0 {
			w.event("crown-data:" + obj.crownData.name)
			return obj.crownData
		},
		firstPlayer: func() *crownDropTestObject4ED5E0 {
			w.event("first-player:" + crownDropObjectName4ED5E0(w.first))
			first := w.first
			if w.afterFirst != nil {
				w.afterFirst(w)
			}
			return first
		},
		loadPlayerData: func(obj *crownDropTestObject4ED5E0) *crownDropTestPlayerData4ED5E0 {
			data := obj.playerData
			name := "nil"
			if data != nil {
				name = data.name
			}
			w.event("player-data:" + obj.name + ":" + name)
			if w.afterPlayerData != nil {
				w.afterPlayerData(w, obj)
			}
			return data
		},
		teamContains: func(obj *crownDropTestObject4ED5E0, teamID uint8) int32 {
			w.event(fmt.Sprintf("team-contains:%s:%d", obj.name, teamID))
			if w.afterTeamCompare != nil {
				w.afterTeamCompare(w, obj)
			}
			if obj.teamID == teamID {
				return -1
			}
			return 0
		},
		loadPickupFrame: func(data *crownDropTestPlayerData4ED5E0) uint32 {
			w.event(fmt.Sprintf("pickup-frame:%s:%#x", data.name, data.frame))
			return data.frame
		},
		nextPlayer: func(obj *crownDropTestObject4ED5E0) *crownDropTestObject4ED5E0 {
			w.event("next:" + obj.name + ":" + crownDropObjectName4ED5E0(obj.next))
			return obj.next
		},
		loadOwnerArg: func() *crownDropTestObject4ED5E0 {
			w.event("owner-arg:" + crownDropObjectName4ED5E0(w.ownerArg))
			return w.ownerArg
		},
		storePickupTarget: func(data *crownDropTestCrownData4ED5E0, target *crownDropTestObject4ED5E0) {
			w.event("store-target:" + data.name + ":" + crownDropObjectName4ED5E0(target))
			data.target = target
		},
		loadPointArg: func() *crownDropTestPoint4ED5E0 {
			w.event("point-arg:" + w.pointArg.name)
			return w.pointArg
		},
		defaultDrop: func(owner, crown *crownDropTestObject4ED5E0, point *crownDropTestPoint4ED5E0) int32 {
			w.event(fmt.Sprintf("default:%s:%s:%s", owner.name, crown.name, point.name))
			return w.defaultResult
		},
		clearOwner: func(crown *crownDropTestObject4ED5E0) {
			w.event("clear-owner:" + crown.name)
		},
		buffOff: func(owner *crownDropTestObject4ED5E0, enchant uint32) {
			w.event(fmt.Sprintf("buff-off:%s:%d", owner.name, enchant))
			if w.afterBuffOff != nil {
				w.afterBuffOff(w)
			}
		},
		loadNetCode: func(owner *crownDropTestObject4ED5E0) uint32 {
			value := owner.netCode
			w.event(fmt.Sprintf("net-code:%s:%#x", owner.name, value))
			if w.afterNetCode != nil {
				w.afterNetCode(w)
			}
			return value
		},
		hasTeam: func(owner *crownDropTestObject4ED5E0) int32 {
			w.event("has-team:" + owner.name)
			result := w.hasTeamResult
			if w.afterHasTeam != nil {
				w.afterHasTeam(w)
			}
			return result
		},
		informDrop: func(code uint8, netCode, teamID uint32) {
			w.event(fmt.Sprintf("inform:%d:%#x:%#x", code, netCode, teamID))
		},
		markMinimap: func(crown *crownDropTestObject4ED5E0, flags uint32) {
			w.event(fmt.Sprintf("minimap:%s:%d", crown.name, flags))
		},
	}
}

func crownDropSuccessEvents4ED5E0() []string {
	return []string{
		"game:16", "gameplay:4", "crown-arg:crown", "frame:0x64", "team:crown:7",
		"crown-data:crown-data", "first-player:first", "crown-arg:crown",
		"crown-arg:crown", "player-data:first:first-data", "team:crown:7",
		"team-contains:first:7", "pickup-frame:first-data:0x50", "next:first:second",
		"crown-arg:crown", "player-data:second:second-data", "team:crown:7",
		"team-contains:second:7", "pickup-frame:second-data:0x5a", "next:second:nil",
		"crown-arg:crown", "owner-arg:owner", "store-target:crown-data:first",
		"point-arg:point", "default:owner:crown:point", "clear-owner:crown",
		"buff-off:owner:30", "net-code:owner:0x11223344", "has-team:owner",
		"team:owner:7", "inform:11:0x11223344:0x7", "minimap:crown:1",
	}
}

func TestCrownDrop4ED5E0ExactSuccessTraceAndFaultPrefixes(t *testing.T) {
	w := newCrownDropTestWorld4ED5E0()
	if got := crownDrop4ED5E0(w.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	want := crownDropSuccessEvents4ED5E0()
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %#v, want %#v", w.events, want)
	}
	if w.crown.crownData.target != w.first {
		t.Fatalf("target = %p, want first %p", w.crown.crownData.target, w.first)
	}

	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			w := newCrownDropTestWorld4ED5E0()
			w.faultAt = faultAt
			defer func() {
				if got := recover(); got != want[faultAt-1] {
					t.Fatalf("panic = %v, want %q", got, want[faultAt-1])
				}
				if !reflect.DeepEqual(w.events, want[:faultAt]) {
					t.Fatalf("events = %#v, want %#v", w.events, want[:faultAt])
				}
			}()
			crownDrop4ED5E0(w.hooks())
		})
	}
}

func TestCrownDrop4ED5E0FlagAndTeamGates(t *testing.T) {
	tests := []struct {
		name           string
		game, gameplay int32
		teamID         uint8
		prefix         []string
	}{
		{"game-off", 0, 1, 7, []string{"game:16", "crown-arg:crown", "owner-arg:owner"}},
		{"gameplay-off", -1, 0, 7, []string{"game:16", "gameplay:4", "crown-arg:crown", "owner-arg:owner"}},
		{"no-crown-team", 1, 1, 0, []string{
			"game:16", "gameplay:4", "crown-arg:crown", "frame:0x64", "team:crown:0", "owner-arg:owner",
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := newCrownDropTestWorld4ED5E0()
			w.gameResult, w.gameplayResult, w.crown.teamID = tc.game, tc.gameplay, tc.teamID
			if got := crownDrop4ED5E0(w.hooks()); got != 1 {
				t.Fatalf("result = %d, want 1", got)
			}
			if !reflect.DeepEqual(w.events[:len(tc.prefix)], tc.prefix) {
				t.Fatalf("prefix = %#v, want %#v", w.events[:len(tc.prefix)], tc.prefix)
			}
			for _, event := range w.events {
				if event == "crown-data:crown-data" || event == "first-player:first" || event == "store-target:crown-data:first" {
					t.Fatalf("disabled selection emitted %q", event)
				}
			}
		})
	}
}

func TestCrownDrop4ED5E0DefaultAndHasTeamUseFullResults(t *testing.T) {
	for _, result := range []int32{0, 1, -1, math.MinInt32} {
		t.Run(fmt.Sprintf("default-%08x", uint32(result)), func(t *testing.T) {
			w := newCrownDropTestWorld4ED5E0()
			w.gameResult = 0
			w.defaultResult = result
			got := crownDrop4ED5E0(w.hooks())
			if result == 0 {
				if got != 0 || w.events[len(w.events)-1] != "default:owner:crown:point" {
					t.Fatalf("result/events = %d/%#v", got, w.events)
				}
				return
			}
			if got != 1 || w.events[len(w.events)-1] != "minimap:crown:1" {
				t.Fatalf("result/events = %d/%#v", got, w.events)
			}
		})
	}

	for _, result := range []int32{0, 1, -1, math.MinInt32} {
		t.Run(fmt.Sprintf("team-%08x", uint32(result)), func(t *testing.T) {
			w := newCrownDropTestWorld4ED5E0()
			w.gameResult = 0
			w.hasTeamResult = result
			if got := crownDrop4ED5E0(w.hooks()); got != 1 {
				t.Fatalf("result = %d, want 1", got)
			}
			wantInform := "inform:11:0x11223344:0x0"
			if result != 0 {
				wantInform = "inform:11:0x11223344:0x7"
			}
			if !containsCrownDropEvent4ED5E0(w.events, wantInform) {
				t.Fatalf("events = %#v, missing %q", w.events, wantInform)
			}
		})
	}
}

func containsCrownDropEvent4ED5E0(events []string, want string) bool {
	for _, event := range events {
		if event == want {
			return true
		}
	}
	return false
}

func TestCrownDrop4ED5E0NoPlayersKeepsCrownCandidate(t *testing.T) {
	w := newCrownDropTestWorld4ED5E0()
	w.first = nil
	if got := crownDrop4ED5E0(w.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if w.crown.crownData.target != w.crown {
		t.Fatalf("target = %p, want crown %p", w.crown.crownData.target, w.crown)
	}
	if containsCrownDropEvent4ED5E0(w.events, "crown-arg:crown") {
		count := 0
		for _, event := range w.events {
			if event == "crown-arg:crown" {
				count++
			}
		}
		if count != 2 {
			t.Fatalf("crown arg loads = %d, want 2; events %#v", count, w.events)
		}
	}
	if containsCrownDropEvent4ED5E0(w.events, "next:first:second") {
		t.Fatalf("nil player iterated: %#v", w.events)
	}

	w = newCrownDropTestWorld4ED5E0()
	w.first = nil
	w.ownerArg = w.crown
	crownDrop4ED5E0(w.hooks())
	if w.crown.crownData.target != nil {
		t.Fatalf("owner-equals-candidate target = %p, want nil", w.crown.crownData.target)
	}
}

func TestCrownDrop4ED5E0ComparesPickupFramesAsSignedInt32(t *testing.T) {
	tests := []struct {
		name        string
		frame       uint32
		playerFrame uint32
		wantPlayer  bool
	}{
		{"negative-best-rejects-zero", 0x80000000, 0, false},
		{"negative-player-beats-positive", 0x7fffffff, 0xffffffff, true},
		{"equal-does-not-replace", 80, 80, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := newCrownDropTestWorld4ED5E0()
			w.frame = tc.frame
			w.first.playerData.frame = tc.playerFrame
			w.first.next = nil
			crownDrop4ED5E0(w.hooks())
			want := w.crown
			if tc.wantPlayer {
				want = w.first
			}
			if w.crown.crownData.target != want {
				t.Fatalf("target = %s, want %s", crownDropObjectName4ED5E0(w.crown.crownData.target), want.name)
			}
		})
	}
}

func TestCrownDrop4ED5E0CachesUpdatesButReloadsCrownAndOwnerFields(t *testing.T) {
	w := newCrownDropTestWorld4ED5E0()
	w.first.next = nil
	w.first.teamID = 9
	oldData := w.crown.crownData
	replacementData := &crownDropTestCrownData4ED5E0{name: "replacement-data"}
	replacementCrown := &crownDropTestObject4ED5E0{name: "replacement", teamID: 9, crownData: replacementData}
	finalCrown := &crownDropTestObject4ED5E0{name: "final", teamID: 9}
	oldPlayerData := w.first.playerData
	replacementPlayerData := &crownDropTestPlayerData4ED5E0{name: "replacement-player-data", frame: 0}
	w.afterFirst = func(w *crownDropTestWorld4ED5E0) {
		w.crownArg = replacementCrown
	}
	w.afterPlayerData = func(w *crownDropTestWorld4ED5E0, player *crownDropTestObject4ED5E0) {
		player.playerData = replacementPlayerData
	}
	w.afterTeamCompare = func(w *crownDropTestWorld4ED5E0, _ *crownDropTestObject4ED5E0) {
		w.crownArg = finalCrown
	}
	w.afterBuffOff = func(w *crownDropTestWorld4ED5E0) {
		w.owner.netCode = 0xaabbccdd
	}
	w.afterNetCode = func(w *crownDropTestWorld4ED5E0) {
		w.owner.teamID = 12
	}
	w.afterHasTeam = func(w *crownDropTestWorld4ED5E0) {
		w.owner.teamID = 13
	}

	if got := crownDrop4ED5E0(w.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if oldData.target != w.first || replacementData.target != nil {
		t.Fatalf("cached targets = old:%s replacement:%s", crownDropObjectName4ED5E0(oldData.target), crownDropObjectName4ED5E0(replacementData.target))
	}
	if w.first.playerData != replacementPlayerData || oldPlayerData.frame != 80 {
		t.Fatalf("player data cache was not isolated")
	}
	if !containsCrownDropEvent4ED5E0(w.events, "pickup-frame:first-data:0x50") {
		t.Fatalf("cached player frame missing: %#v", w.events)
	}
	if !containsCrownDropEvent4ED5E0(w.events, "default:owner:final:point") {
		t.Fatalf("final Crown was not reloaded: %#v", w.events)
	}
	if !containsCrownDropEvent4ED5E0(w.events, "inform:11:0xaabbccdd:0xd") {
		t.Fatalf("live owner fields missing: %#v", w.events)
	}
}

func TestCrownDrop4ED5E0TeamRejectSkipsCachedPlayerDereference(t *testing.T) {
	w := newCrownDropTestWorld4ED5E0()
	w.first.playerData = nil
	w.first.teamID = 8
	w.first.next = nil
	if got := crownDrop4ED5E0(w.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if containsCrownDropEvent4ED5E0(w.events, "pickup-frame:nil:0x0") {
		t.Fatalf("rejected player data was dereferenced: %#v", w.events)
	}
	if w.crown.crownData.target != w.crown {
		t.Fatalf("target = %s, want crown", crownDropObjectName4ED5E0(w.crown.crownData.target))
	}
}
