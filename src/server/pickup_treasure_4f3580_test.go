package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type pickupTreasureTestObject4F3580 struct {
	name     string
	classLow uint8
	update   *pickupTreasureTestUpdate4F3580
	teamID   uint8
	score    int32
	deaths   uint32
}

type pickupTreasureTestUpdate4F3580 struct {
	name   string
	player *pickupTreasureTestPlayer4F3580
}

type pickupTreasureTestPlayer4F3580 struct {
	name    string
	count   uint32
	maximum uint32
}

type pickupTreasureTestTeam4F3580 struct {
	name string
	id   uint8
}

type pickupTreasureTestWorld4F3580 struct {
	arg3, arg4       int32
	defaultResult    int32
	gameFlagResult   int32
	hasTeamResult    int32
	maximums         []uint32
	maximumCall      int
	team             *pickupTreasureTestTeam4F3580
	players          []*pickupTreasureTestObject4F3580
	events           []string
	faultAt          int
	afterDefault     func(*pickupTreasureTestWorld4F3580)
	afterGameFlag    func(*pickupTreasureTestWorld4F3580)
	afterMaximum     func(*pickupTreasureTestWorld4F3580, int)
	afterReport      func(*pickupTreasureTestWorld4F3580)
	afterHasTeam     func(*pickupTreasureTestWorld4F3580)
	afterTeamCompare func(*pickupTreasureTestWorld4F3580, *pickupTreasureTestObject4F3580)
}

func pickupTreasureTestObjectName4F3580(obj *pickupTreasureTestObject4F3580) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func pickupTreasureTestUpdateName4F3580(update *pickupTreasureTestUpdate4F3580) string {
	if update == nil {
		return "nil"
	}
	return update.name
}

func pickupTreasureTestPlayerName4F3580(player *pickupTreasureTestPlayer4F3580) string {
	if player == nil {
		return "nil"
	}
	return player.name
}

func pickupTreasureTestTeamName4F3580(team *pickupTreasureTestTeam4F3580) string {
	if team == nil {
		return "nil"
	}
	return team.name
}

func (w *pickupTreasureTestWorld4F3580) event(value string) {
	w.events = append(w.events, value)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(value)
	}
}

func (w *pickupTreasureTestWorld4F3580) hooks() pickupTreasureHooks4F3580[
	*pickupTreasureTestObject4F3580,
	*pickupTreasureTestUpdate4F3580,
	*pickupTreasureTestPlayer4F3580,
	*pickupTreasureTestTeam4F3580,
] {
	return pickupTreasureHooks4F3580[
		*pickupTreasureTestObject4F3580,
		*pickupTreasureTestUpdate4F3580,
		*pickupTreasureTestPlayer4F3580,
		*pickupTreasureTestTeam4F3580,
	]{
		loadArg4: func() int32 {
			w.event(fmt.Sprintf("arg4=%08x", uint32(w.arg4)))
			return w.arg4
		},
		loadArg3: func() int32 {
			w.event(fmt.Sprintf("arg3=%08x", uint32(w.arg3)))
			return w.arg3
		},
		defaultPickup: func(owner, item *pickupTreasureTestObject4F3580, arg3, arg4 int32) int32 {
			w.event(fmt.Sprintf("default:%s:%s:%08x:%08x", pickupTreasureTestObjectName4F3580(owner), pickupTreasureTestObjectName4F3580(item), uint32(arg3), uint32(arg4)))
			if w.afterDefault != nil {
				w.afterDefault(w)
			}
			return w.defaultResult
		},
		loadClassLow: func(obj *pickupTreasureTestObject4F3580) uint8 {
			w.event(fmt.Sprintf("class:%s=%02x", pickupTreasureTestObjectName4F3580(obj), obj.classLow))
			return obj.classLow
		},
		loadUpdate: func(obj *pickupTreasureTestObject4F3580) *pickupTreasureTestUpdate4F3580 {
			w.event("update:" + pickupTreasureTestObjectName4F3580(obj) + "=" + pickupTreasureTestUpdateName4F3580(obj.update))
			return obj.update
		},
		gameFlag: func(flag uint32) int32 {
			w.event(fmt.Sprintf("game-flag:%08x=%08x", flag, uint32(w.gameFlagResult)))
			if w.afterGameFlag != nil {
				w.afterGameFlag(w)
			}
			return w.gameFlagResult
		},
		audio: func(id uint32, obj *pickupTreasureTestObject4F3580, kind int32, code uint32) {
			w.event(fmt.Sprintf("audio:%d:%s:%d:%08x", id, pickupTreasureTestObjectName4F3580(obj), kind, code))
		},
		loadPlayer: func(update *pickupTreasureTestUpdate4F3580) *pickupTreasureTestPlayer4F3580 {
			w.event("player:" + pickupTreasureTestUpdateName4F3580(update) + "=" + pickupTreasureTestPlayerName4F3580(update.player))
			return update.player
		},
		loadCount: func(player *pickupTreasureTestPlayer4F3580) uint32 {
			w.event(fmt.Sprintf("count:%s=%08x", pickupTreasureTestPlayerName4F3580(player), player.count))
			return player.count
		},
		storeCount: func(player *pickupTreasureTestPlayer4F3580, value uint32) {
			w.event(fmt.Sprintf("store-count:%s=%08x", pickupTreasureTestPlayerName4F3580(player), value))
			player.count = value
		},
		treasureMax: func() uint32 {
			call := w.maximumCall
			value := w.maximums[call]
			w.maximumCall++
			w.event(fmt.Sprintf("maximum:%d=%08x", call, value))
			if w.afterMaximum != nil {
				w.afterMaximum(w, call)
			}
			return value
		},
		storeMax: func(player *pickupTreasureTestPlayer4F3580, value uint32) {
			w.event(fmt.Sprintf("store-max:%s=%08x", pickupTreasureTestPlayerName4F3580(player), value))
			player.maximum = value
		},
		report: func(obj *pickupTreasureTestObject4F3580) {
			w.event("report:" + pickupTreasureTestObjectName4F3580(obj))
			if w.afterReport != nil {
				w.afterReport(w)
			}
		},
		hasTeam: func(obj *pickupTreasureTestObject4F3580) int32 {
			w.event(fmt.Sprintf("has-team:%s=%08x", pickupTreasureTestObjectName4F3580(obj), uint32(w.hasTeamResult)))
			if w.afterHasTeam != nil {
				w.afterHasTeam(w)
			}
			return w.hasTeamResult
		},
		loadObjectTeam: func(obj *pickupTreasureTestObject4F3580) uint8 {
			w.event(fmt.Sprintf("object-team:%s=%02x", pickupTreasureTestObjectName4F3580(obj), obj.teamID))
			return obj.teamID
		},
		findTeam: func(id uint8) *pickupTreasureTestTeam4F3580 {
			w.event(fmt.Sprintf("find-team:%02x=%s", id, pickupTreasureTestTeamName4F3580(w.team)))
			return w.team
		},
		loadTeamID: func(team *pickupTreasureTestTeam4F3580) uint8 {
			w.event(fmt.Sprintf("team-id:%s=%02x", pickupTreasureTestTeamName4F3580(team), team.id))
			return team.id
		},
		teamContains: func(obj *pickupTreasureTestObject4F3580, id uint8) int32 {
			result := int32(0)
			if obj.teamID == id && id != 0 {
				result = 1
			}
			w.event(fmt.Sprintf("team-contains:%s:%02x=%08x", pickupTreasureTestObjectName4F3580(obj), id, uint32(result)))
			if w.afterTeamCompare != nil {
				w.afterTeamCompare(w, obj)
			}
			return result
		},
		firstPlayer: func() *pickupTreasureTestObject4F3580 {
			var first *pickupTreasureTestObject4F3580
			if len(w.players) != 0 {
				first = w.players[0]
			}
			w.event("first-player=" + pickupTreasureTestObjectName4F3580(first))
			return first
		},
		nextPlayer: func(current *pickupTreasureTestObject4F3580) *pickupTreasureTestObject4F3580 {
			var next *pickupTreasureTestObject4F3580
			for i, obj := range w.players {
				if obj == current && i+1 < len(w.players) {
					next = w.players[i+1]
					break
				}
			}
			w.event("next-player:" + pickupTreasureTestObjectName4F3580(current) + "=" + pickupTreasureTestObjectName4F3580(next))
			return next
		},
		setGameFlags: func(flags uint32) {
			w.event(fmt.Sprintf("set-flags:%08x", flags))
		},
		changeScore: func(obj *pickupTreasureTestObject4F3580, value int32) {
			w.event(fmt.Sprintf("change-score:%s=%d", pickupTreasureTestObjectName4F3580(obj), value))
			obj.score += value
		},
		reportLesson: func(obj *pickupTreasureTestObject4F3580) {
			w.event("report-lesson:" + pickupTreasureTestObjectName4F3580(obj))
		},
		incrementDeaths: func(obj *pickupTreasureTestObject4F3580) {
			w.event("increment-deaths:" + pickupTreasureTestObjectName4F3580(obj))
			obj.deaths++
		},
	}
}

func verifyPickupTreasureFaultPrefixes4F3580(
	t *testing.T,
	want []string,
	build func() (*pickupTreasureTestWorld4F3580, *pickupTreasureTestObject4F3580, *pickupTreasureTestObject4F3580),
) {
	t.Helper()
	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			w, owner, item := build()
			w.faultAt = faultAt
			defer func() {
				if got := recover(); got != want[faultAt-1] {
					t.Fatalf("panic = %v, want %q", got, want[faultAt-1])
				}
				if !reflect.DeepEqual(w.events, want[:faultAt]) {
					t.Fatalf("events = %v, want %v", w.events, want[:faultAt])
				}
			}()
			pickupTreasure4F3580(owner, item, w.hooks())
		})
	}
}

func TestPickupTreasure4F3580DefaultGateLoadsArg4First(t *testing.T) {
	owner := &pickupTreasureTestObject4F3580{name: "owner"}
	item := &pickupTreasureTestObject4F3580{name: "item"}
	w := &pickupTreasureTestWorld4F3580{
		arg3:          math.MinInt32,
		arg4:          math.MaxInt32,
		defaultResult: 0,
	}
	if got := pickupTreasure4F3580(owner, item, w.hooks()); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	want := []string{
		"arg4=7fffffff",
		"arg3=80000000",
		"default:owner:item:80000000:7fffffff",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
}

func TestPickupTreasure4F3580CanonicalSuccessAndCachedUpdateBeforeFlag(t *testing.T) {
	oldUpdate := &pickupTreasureTestUpdate4F3580{name: "old"}
	newUpdate := &pickupTreasureTestUpdate4F3580{name: "new"}
	owner := &pickupTreasureTestObject4F3580{name: "owner", classLow: 0x84, update: oldUpdate}
	item := &pickupTreasureTestObject4F3580{name: "item"}
	w := &pickupTreasureTestWorld4F3580{defaultResult: math.MinInt32}
	w.afterGameFlag = func(*pickupTreasureTestWorld4F3580) { owner.update = newUpdate }
	if got := pickupTreasure4F3580(owner, item, w.hooks()); got != 1 {
		t.Fatalf("result = %d, want canonical 1", got)
	}
	want := []string{
		"arg4=00000000",
		"arg3=00000000",
		"default:owner:item:00000000:00000000",
		"class:owner=84",
		"update:owner=old",
		"game-flag:00000040=00000000",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}

	owner.classLow = 0x80
	owner.update = nil
	w = &pickupTreasureTestWorld4F3580{defaultResult: 1}
	if got := pickupTreasure4F3580(owner, item, w.hooks()); got != 1 {
		t.Fatalf("non-Player result = %d, want 1", got)
	}
	if got := w.events[len(w.events)-1]; got != "class:owner=80" {
		t.Fatalf("non-Player last event = %q; all = %v", got, w.events)
	}
}

func pickupTreasureSoloBuild4F3580() (*pickupTreasureTestWorld4F3580, *pickupTreasureTestObject4F3580, *pickupTreasureTestObject4F3580) {
	firstPlayer := &pickupTreasureTestPlayer4F3580{name: "first", count: 6}
	maxPlayer := &pickupTreasureTestPlayer4F3580{name: "max"}
	completionPlayer := &pickupTreasureTestPlayer4F3580{name: "completion", count: 9}
	update := &pickupTreasureTestUpdate4F3580{name: "cached", player: firstPlayer}
	owner := &pickupTreasureTestObject4F3580{name: "owner", classLow: 4, update: update}
	other1 := &pickupTreasureTestObject4F3580{name: "other1"}
	other2 := &pickupTreasureTestObject4F3580{name: "other2"}
	item := &pickupTreasureTestObject4F3580{name: "item"}
	w := &pickupTreasureTestWorld4F3580{
		arg3:           -3,
		arg4:           -4,
		defaultResult:  math.MinInt32,
		gameFlagResult: math.MinInt32,
		maximums:       []uint32{9, 9},
		players:        []*pickupTreasureTestObject4F3580{owner, other1, other2},
	}
	w.afterMaximum = func(_ *pickupTreasureTestWorld4F3580, call int) {
		if call == 0 {
			update.player = maxPlayer
		} else {
			update.player = completionPlayer
		}
	}
	return w, owner, item
}

func pickupTreasureSoloTrace4F3580() []string {
	return []string{
		"arg4=fffffffc",
		"arg3=fffffffd",
		"default:owner:item:fffffffd:fffffffc",
		"class:owner=04",
		"update:owner=cached",
		"game-flag:00000040=80000000",
		"audio:307:owner:0:00000000",
		"player:cached=first",
		"count:first=00000006",
		"store-count:first=00000007",
		"maximum:0=00000009",
		"player:cached=max",
		"store-max:max=00000009",
		"report:owner",
		"has-team:owner=00000000",
		"maximum:1=00000009",
		"player:cached=completion",
		"count:completion=00000009",
		"set-flags:00000008",
		"change-score:owner=1",
		"report-lesson:owner",
		"first-player=owner",
		"next-player:owner=other1",
		"increment-deaths:other1",
		"report-lesson:other1",
		"next-player:other1=other2",
		"increment-deaths:other2",
		"report-lesson:other2",
		"next-player:other2=nil",
	}
}

func TestPickupTreasure4F3580SoloCompletionOrderReloadsAndFaults(t *testing.T) {
	w, owner, item := pickupTreasureSoloBuild4F3580()
	if got := pickupTreasure4F3580(owner, item, w.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	want := pickupTreasureSoloTrace4F3580()
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events =\n%v\nwant =\n%v", w.events, want)
	}
	if owner.score != 1 || w.players[1].deaths != 1 || w.players[2].deaths != 1 {
		t.Fatalf("score/deaths = %d/%d/%d", owner.score, w.players[1].deaths, w.players[2].deaths)
	}
	verifyPickupTreasureFaultPrefixes4F3580(t, want, pickupTreasureSoloBuild4F3580)
}

func TestPickupTreasure4F3580SoloMismatchStopsBeforeCompletionEffects(t *testing.T) {
	w, owner, item := pickupTreasureSoloBuild4F3580()
	w.maximums[1] = 10
	if got := pickupTreasure4F3580(owner, item, w.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if got := w.events[len(w.events)-1]; got != "count:completion=00000009" {
		t.Fatalf("last event = %q; all = %v", got, w.events)
	}
}

func pickupTreasureTeamBuild4F3580() (*pickupTreasureTestWorld4F3580, *pickupTreasureTestObject4F3580, *pickupTreasureTestObject4F3580) {
	ownerPlayer := &pickupTreasureTestPlayer4F3580{name: "owner-player", count: 1}
	ownerUpdate := &pickupTreasureTestUpdate4F3580{name: "owner-update", player: ownerPlayer}
	owner := &pickupTreasureTestObject4F3580{name: "owner", classLow: 4, update: ownerUpdate, teamID: 3}
	item := &pickupTreasureTestObject4F3580{name: "item"}
	memberOld := &pickupTreasureTestUpdate4F3580{name: "member-old", player: &pickupTreasureTestPlayer4F3580{name: "ignored", count: 99}}
	memberNew := &pickupTreasureTestUpdate4F3580{name: "member-new", player: &pickupTreasureTestPlayer4F3580{name: "member", count: math.MaxUint32}}
	member := &pickupTreasureTestObject4F3580{name: "member", update: memberOld, teamID: 7}
	outsider := &pickupTreasureTestObject4F3580{name: "outsider", update: nil, teamID: 9}
	last := &pickupTreasureTestObject4F3580{name: "last", update: &pickupTreasureTestUpdate4F3580{name: "last-update", player: &pickupTreasureTestPlayer4F3580{name: "last-player", count: 2}}, teamID: 8}
	team := &pickupTreasureTestTeam4F3580{name: "team", id: 7}
	w := &pickupTreasureTestWorld4F3580{
		defaultResult:  1,
		gameFlagResult: 1,
		hasTeamResult:  math.MinInt32,
		maximums:       []uint32{5, 1},
		team:           team,
		players:        []*pickupTreasureTestObject4F3580{member, outsider, last},
	}
	w.afterHasTeam = func(*pickupTreasureTestWorld4F3580) { owner.teamID = 7 }
	w.afterTeamCompare = func(_ *pickupTreasureTestWorld4F3580, obj *pickupTreasureTestObject4F3580) {
		if obj == member {
			obj.update = memberNew
			team.id = 8
		}
	}
	return w, owner, item
}

func TestPickupTreasure4F3580TeamUsesLiveIDsDelayedLoadsAndWrappingSum(t *testing.T) {
	w, owner, item := pickupTreasureTeamBuild4F3580()
	if got := pickupTreasure4F3580(owner, item, w.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	wantTail := []string{
		"has-team:owner=80000000",
		"object-team:owner=07",
		"find-team:07=team",
		"first-player=member",
		"team-id:team=07",
		"team-contains:member:07=00000001",
		"update:member=member-new",
		"player:member-new=member",
		"count:member=ffffffff",
		"next-player:member=outsider",
		"team-id:team=08",
		"team-contains:outsider:08=00000000",
		"next-player:outsider=last",
		"team-id:team=08",
		"team-contains:last:08=00000001",
		"update:last=last-update",
		"player:last-update=last-player",
		"count:last-player=00000002",
		"next-player:last=nil",
		"maximum:1=00000001",
		"set-flags:00000008",
	}
	if len(w.events) < len(wantTail) || !reflect.DeepEqual(w.events[len(w.events)-len(wantTail):], wantTail) {
		t.Fatalf("tail =\n%v\nwant =\n%v\nall =\n%v", w.events[len(w.events)-len(wantTail):], wantTail, w.events)
	}
}

func TestPickupTreasure4F3580MissingTeamReturnsWithoutSecondMaximum(t *testing.T) {
	w, owner, item := pickupTreasureTeamBuild4F3580()
	w.team = nil
	if got := pickupTreasure4F3580(owner, item, w.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if w.maximumCall != 1 {
		t.Fatalf("maximum calls = %d, want only initial update", w.maximumCall)
	}
	if got := w.events[len(w.events)-1]; got != "find-team:07=nil" {
		t.Fatalf("last event = %q; all = %v", got, w.events)
	}
}

func TestPickupTreasure4F3580CountIncrementWraps(t *testing.T) {
	player := &pickupTreasureTestPlayer4F3580{name: "player", count: math.MaxUint32}
	owner := &pickupTreasureTestObject4F3580{
		name:     "owner",
		classLow: 4,
		update:   &pickupTreasureTestUpdate4F3580{name: "update", player: player},
	}
	item := &pickupTreasureTestObject4F3580{name: "item"}
	w := &pickupTreasureTestWorld4F3580{
		defaultResult:  1,
		gameFlagResult: 1,
		maximums:       []uint32{3, 4},
	}
	if got := pickupTreasure4F3580(owner, item, w.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if player.count != 0 {
		t.Fatalf("wrapped count = %08x, want 00000000", player.count)
	}
}
