package server

import (
	"fmt"
	"math"
	"testing"
)

type homeBaseTestObject4EBB80 struct {
	name         string
	typeIndex    uint16
	update       *homeBaseTestBallData4EBB80
	playerUpdate *homeBaseTestPlayerUpdate4EBB80
	hasTeam      bool
	teamID       uint8
	next         *homeBaseTestObject4EBB80
	velocityX    float32
	velocityY    float32
	forceX       float32
	pos24Y       float32
}

type homeBaseTestBallData4EBB80 struct {
	carrier *homeBaseTestObject4EBB80
}

type homeBaseTestTeam4EBB80 struct {
	name    string
	lessons int32
}

type homeBaseTestPlayerUpdate4EBB80 struct {
	name   string
	player *homeBaseTestPlayer4EBB80
}

type homeBaseTestPlayer4EBB80 struct {
	name string
}

type homeBaseTestState4EBB80 struct {
	events       []string
	gameBallType uint32
	startType    uint32
	teams        map[uint8]*homeBaseTestTeam4EBB80
	first        *homeBaseTestObject4EBB80
	observerMode uint32
	randomResult int32
	pointResults map[uint32]uint32
	carrierLoads []*homeBaseTestObject4EBB80
	onRandom     func()
}

func homeBaseObjectName4EBB80(obj *homeBaseTestObject4EBB80) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func homeBaseTeamName4EBB80(team *homeBaseTestTeam4EBB80) string {
	if team == nil {
		return "nil"
	}
	return team.name
}

func homeBasePlayerName4EBB80(player *homeBaseTestPlayer4EBB80) string {
	if player == nil {
		return "nil"
	}
	return player.name
}

func (s *homeBaseTestState4EBB80) event(format string, args ...any) {
	s.events = append(s.events, fmt.Sprintf(format, args...))
}

func (s *homeBaseTestState4EBB80) hooks() homeBaseCollideHooks4EBB80[
	*homeBaseTestObject4EBB80,
	*homeBaseTestBallData4EBB80,
	*homeBaseTestTeam4EBB80,
	*homeBaseTestPlayerUpdate4EBB80,
	*homeBaseTestPlayer4EBB80,
] {
	return homeBaseCollideHooks4EBB80[
		*homeBaseTestObject4EBB80,
		*homeBaseTestBallData4EBB80,
		*homeBaseTestTeam4EBB80,
		*homeBaseTestPlayerUpdate4EBB80,
		*homeBaseTestPlayer4EBB80,
	]{
		lookupType: func(name string) uint32 {
			s.event("lookup:%s", name)
			if name == homeBaseGameBallName4EBB80 {
				return s.gameBallType
			}
			return s.startType
		},
		loadTypeIndex: func(obj *homeBaseTestObject4EBB80) uint16 {
			s.event("type:%s", homeBaseObjectName4EBB80(obj))
			return obj.typeIndex
		},
		loadUpdate: func(obj *homeBaseTestObject4EBB80) *homeBaseTestBallData4EBB80 {
			s.event("update:%s", homeBaseObjectName4EBB80(obj))
			return obj.update
		},
		loadCarrier: func(update *homeBaseTestBallData4EBB80) *homeBaseTestObject4EBB80 {
			if update == nil {
				s.event("carrier:nil-update")
				panic("nil ball update")
			}
			carrier := update.carrier
			if len(s.carrierLoads) != 0 {
				carrier = s.carrierLoads[0]
				s.carrierLoads = s.carrierLoads[1:]
			}
			s.event("carrier:%s", homeBaseObjectName4EBB80(carrier))
			return carrier
		},
		hasTeam: func(obj *homeBaseTestObject4EBB80) bool {
			s.event("has-team:%s", homeBaseObjectName4EBB80(obj))
			return obj.hasTeam
		},
		loadTeamID: func(obj *homeBaseTestObject4EBB80) uint8 {
			s.event("team-id:%s", homeBaseObjectName4EBB80(obj))
			return obj.teamID
		},
		teamByID: func(id uint8) *homeBaseTestTeam4EBB80 {
			team := s.teams[id]
			s.event("team:%d=%s", id, homeBaseTeamName4EBB80(team))
			return team
		},
		changeScore: func(obj *homeBaseTestObject4EBB80, delta int32) {
			s.event("score:%s:%d", homeBaseObjectName4EBB80(obj), delta)
		},
		reportLesson: func(obj *homeBaseTestObject4EBB80) {
			s.event("report:%s", homeBaseObjectName4EBB80(obj))
		},
		loadTeamLessons: func(team *homeBaseTestTeam4EBB80) int32 {
			s.event("lessons:%s", homeBaseTeamName4EBB80(team))
			return team.lessons
		},
		changeTeamLessons: func(team *homeBaseTestTeam4EBB80, lessons int32) {
			s.event("team-score:%s:%d", homeBaseTeamName4EBB80(team), lessons)
			team.lessons = lessons
		},
		observerMode: func() uint32 {
			s.event("observer-mode:%d", s.observerMode)
			return s.observerMode
		},
		loadPlayerUpdate: func(obj *homeBaseTestObject4EBB80) *homeBaseTestPlayerUpdate4EBB80 {
			s.event("player-update:%s", homeBaseObjectName4EBB80(obj))
			return obj.playerUpdate
		},
		loadPlayer: func(update *homeBaseTestPlayerUpdate4EBB80) *homeBaseTestPlayer4EBB80 {
			s.event("player:%s", update.name)
			return update.player
		},
		observerUpdate: func(first, second *homeBaseTestPlayer4EBB80) {
			s.event("observer:%s:%s", homeBasePlayerName4EBB80(first), homeBasePlayerName4EBB80(second))
		},
		audio: func(id uint32, obj *homeBaseTestObject4EBB80) {
			s.event("audio:%d:%s", id, homeBaseObjectName4EBB80(obj))
		},
		pointFX: func(id uint32, obj *homeBaseTestObject4EBB80) uint32 {
			s.event("fx:%d:%s", id, homeBaseObjectName4EBB80(obj))
			return s.pointResults[id]
		},
		firstObject: func() *homeBaseTestObject4EBB80 {
			s.event("first:%s", homeBaseObjectName4EBB80(s.first))
			return s.first
		},
		nextObject: func(obj *homeBaseTestObject4EBB80) *homeBaseTestObject4EBB80 {
			s.event("next:%s=%s", obj.name, homeBaseObjectName4EBB80(obj.next))
			return obj.next
		},
		randomInt: func(minimum, maximum int32, path string, line int32) int32 {
			s.event("random:%d:%d:%s:%#x", minimum, maximum, path, line)
			if s.onRandom != nil {
				s.onRandom()
			}
			return s.randomResult
		},
		clearOwner: func(obj *homeBaseTestObject4EBB80) {
			s.event("clear-owner:%s", homeBaseObjectName4EBB80(obj))
		},
		moveToMarker: func(obj, marker *homeBaseTestObject4EBB80) {
			s.event("move:%s:%s", homeBaseObjectName4EBB80(obj), homeBaseObjectName4EBB80(marker))
		},
		storeVelocityX: func(obj *homeBaseTestObject4EBB80, value float32) {
			s.event("store:velocity-x:%s:%08x", obj.name, math.Float32bits(value))
			obj.velocityX = value
		},
		storeVelocityY: func(obj *homeBaseTestObject4EBB80, value float32) {
			s.event("store:velocity-y:%s:%08x", obj.name, math.Float32bits(value))
			obj.velocityY = value
		},
		storeForceX: func(obj *homeBaseTestObject4EBB80, value float32) {
			s.event("store:force-x:%s:%08x", obj.name, math.Float32bits(value))
			obj.forceX = value
		},
		storePos24Y: func(obj *homeBaseTestObject4EBB80, value float32) {
			s.event("store:pos24-y:%s:%08x", obj.name, math.Float32bits(value))
			obj.pos24Y = value
		},
	}
}

func assertHomeBaseEvents4EBB80(t *testing.T, got, want []string) {
	t.Helper()
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("events:\n got %q\nwant %q", got, want)
	}
}

func TestHomeBaseCollide4EBB80EntryLookupsAndGuardReturns(t *testing.T) {
	t.Run("nil target keeps full GameBallStart lookup result", func(t *testing.T) {
		state := &homeBaseTestState4EBB80{gameBallType: 0x1234, startType: 0xf1234567}
		got := homeBaseCollide4EBB80(
			(*homeBaseTestObject4EBB80)(nil),
			(*homeBaseTestObject4EBB80)(nil),
			struct{ unread uint32 }{unread: 0xdeadbeef},
			state.hooks(),
		)
		if got != 0xf1234567 {
			t.Fatalf("return = %#x, want %#x", got, uint32(0xf1234567))
		}
		assertHomeBaseEvents4EBB80(t, state.events, []string{
			"lookup:GameBall",
			"lookup:GameBallStart",
		})
	})

	t.Run("type mismatch is zero extended", func(t *testing.T) {
		other := &homeBaseTestObject4EBB80{name: "other", typeIndex: 0xabcd}
		state := &homeBaseTestState4EBB80{gameBallType: 0x1234, startType: 0xf1234567}
		got := homeBaseCollide4EBB80((*homeBaseTestObject4EBB80)(nil), other, 7, state.hooks())
		if got != 0x0000abcd {
			t.Fatalf("return = %#x, want 0xabcd", got)
		}
		assertHomeBaseEvents4EBB80(t, state.events, []string{
			"lookup:GameBall",
			"lookup:GameBallStart",
			"type:other",
		})
	})

	t.Run("high type lookup does not match low object word", func(t *testing.T) {
		other := &homeBaseTestObject4EBB80{name: "other", typeIndex: 4}
		state := &homeBaseTestState4EBB80{gameBallType: 0x00010004, startType: 9}
		if got := homeBaseCollide4EBB80((*homeBaseTestObject4EBB80)(nil), other, 0, state.hooks()); got != 4 {
			t.Fatalf("return = %#x, want 4", got)
		}
		assertHomeBaseEvents4EBB80(t, state.events, []string{
			"lookup:GameBall",
			"lookup:GameBallStart",
			"type:other",
		})
	})
}

func TestHomeBaseCollide4EBB80ScoringUsesLiveCarrierAndRespawns(t *testing.T) {
	homeTeam := &homeBaseTestTeam4EBB80{name: "home-team", lessons: math.MaxInt32}
	homeBase := &homeBaseTestObject4EBB80{name: "home", hasTeam: true, teamID: 7}
	carrierForGate := &homeBaseTestObject4EBB80{name: "gate-carrier", hasTeam: true, teamID: 91}
	carrierForTeam := &homeBaseTestObject4EBB80{name: "team-carrier", teamID: 8}
	carrierForScore := &homeBaseTestObject4EBB80{name: "score-carrier"}
	carrierForReport := &homeBaseTestObject4EBB80{name: "report-carrier"}
	player := &homeBaseTestPlayer4EBB80{name: "observer-player"}
	carrierForObserver := &homeBaseTestObject4EBB80{
		name:         "observer-carrier",
		playerUpdate: &homeBaseTestPlayerUpdate4EBB80{name: "player-data", player: player},
	}
	ball := &homeBaseTestObject4EBB80{
		name:      "ball",
		typeIndex: 0x2222,
		update:    &homeBaseTestBallData4EBB80{carrier: carrierForGate},
		velocityX: 1,
		velocityY: 2,
		forceX:    3,
		pos24Y:    4,
	}
	nonMarker := &homeBaseTestObject4EBB80{name: "plain", typeIndex: 3}
	marker0 := &homeBaseTestObject4EBB80{name: "marker-0", typeIndex: 0x3333}
	marker1 := &homeBaseTestObject4EBB80{name: "marker-1", typeIndex: 0x3333}
	nonMarker.next = marker0
	marker0.next = marker1
	state := &homeBaseTestState4EBB80{
		gameBallType: 0x2222,
		startType:    0x3333,
		teams: map[uint8]*homeBaseTestTeam4EBB80{
			7: homeTeam,
			8: homeTeam,
		},
		first:        nonMarker,
		observerMode: 1,
		randomResult: 1,
		pointResults: map[uint32]uint32{
			homeBaseScorePointFX4EBB80:   0xaaaaaaaa,
			homeBaseRespawnPointFX4EBB80: 0xf1234567,
		},
		carrierLoads: []*homeBaseTestObject4EBB80{
			carrierForGate,
			carrierForTeam,
			carrierForScore,
			carrierForReport,
			carrierForObserver,
		},
	}

	got := homeBaseCollide4EBB80(homeBase, ball, &struct{ unread uint32 }{unread: 0xdeadbeef}, state.hooks())
	if got != 0xf1234567 {
		t.Fatalf("return = %#x, want %#x", got, uint32(0xf1234567))
	}
	if homeTeam.lessons != math.MinInt32 {
		t.Fatalf("wrapped team lessons = %d, want %d", homeTeam.lessons, int32(math.MinInt32))
	}
	if ball.velocityX != 0 || ball.velocityY != 0 || ball.forceX != 0 || ball.pos24Y != 0 {
		t.Fatalf("motion = %v/%v/%v/%v, want zero", ball.velocityX, ball.velocityY, ball.forceX, ball.pos24Y)
	}
	assertHomeBaseEvents4EBB80(t, state.events, []string{
		"lookup:GameBall",
		"lookup:GameBallStart",
		"type:ball",
		"update:ball",
		"has-team:home",
		"team-id:home",
		"team:7=home-team",
		"carrier:gate-carrier",
		"has-team:gate-carrier",
		"carrier:team-carrier",
		"team-id:team-carrier",
		"team:8=home-team",
		"carrier:score-carrier",
		"score:score-carrier:1",
		"carrier:report-carrier",
		"report:report-carrier",
		"lessons:home-team",
		"team-score:home-team:-2147483648",
		"observer-mode:1",
		"carrier:observer-carrier",
		"player-update:observer-carrier",
		"player:player-data",
		"observer:observer-player:nil",
		"audio:929:home",
		"fx:154:ball",
		"first:plain",
		"type:plain",
		"next:plain=marker-0",
		"type:marker-0",
		"next:marker-0=marker-1",
		"type:marker-1",
		"next:marker-1=nil",
		"random:0:1:C:\\NoxPost\\src\\Server\\Object\\collide\\objcoll.c:0xf57",
		"first:plain",
		"type:plain",
		"next:plain=marker-0",
		"type:marker-0",
		"next:marker-0=marker-1",
		"type:marker-1",
		"clear-owner:ball",
		"move:ball:marker-1",
		"fx:129:ball",
		"store:velocity-x:ball:00000000",
		"store:velocity-y:ball:00000000",
		"store:force-x:ball:00000000",
		"store:pos24-y:ball:00000000",
	})
}

func TestHomeBaseCollide4EBB80NilTeamsStillScoreNilCarrier(t *testing.T) {
	ball := &homeBaseTestObject4EBB80{
		name:      "ball",
		typeIndex: 7,
		update:    &homeBaseTestBallData4EBB80{},
	}
	homeBase := &homeBaseTestObject4EBB80{name: "home"}
	state := &homeBaseTestState4EBB80{
		gameBallType: 7,
		startType:    9,
		pointResults: map[uint32]uint32{},
	}
	if got := homeBaseCollide4EBB80(homeBase, ball, 0, state.hooks()); got != 0 {
		t.Fatalf("return = %#x, want 0", got)
	}
	assertHomeBaseEvents4EBB80(t, state.events, []string{
		"lookup:GameBall",
		"lookup:GameBallStart",
		"type:ball",
		"update:ball",
		"has-team:home",
		"carrier:nil",
		"carrier:nil",
		"score:nil:1",
		"carrier:nil",
		"report:nil",
		"first:nil",
		"random:0:-1:C:\\NoxPost\\src\\Server\\Object\\collide\\objcoll.c:0xf57",
		"first:nil",
	})
}

func TestHomeBaseCollide4EBB80SecondTraversalIsLive(t *testing.T) {
	team := &homeBaseTestTeam4EBB80{name: "home"}
	homeBase := &homeBaseTestObject4EBB80{name: "home", hasTeam: true, teamID: 1}
	carrier := &homeBaseTestObject4EBB80{name: "carrier", hasTeam: true, teamID: 2}
	ball := &homeBaseTestObject4EBB80{
		name:      "ball",
		typeIndex: 7,
		update:    &homeBaseTestBallData4EBB80{carrier: carrier},
	}
	oldMarker := &homeBaseTestObject4EBB80{name: "old", typeIndex: 9}
	newMarker := &homeBaseTestObject4EBB80{name: "new", typeIndex: 3}
	oldMarker.next = newMarker
	state := &homeBaseTestState4EBB80{
		gameBallType: 7,
		startType:    9,
		teams: map[uint8]*homeBaseTestTeam4EBB80{
			1: team,
			2: &homeBaseTestTeam4EBB80{name: "other"},
		},
		first:        oldMarker,
		randomResult: 0,
		pointResults: map[uint32]uint32{homeBaseRespawnPointFX4EBB80: 0x89abcdef},
	}
	state.onRandom = func() {
		state.event("mutate-markers")
		oldMarker.typeIndex = 3
		newMarker.typeIndex = 9
	}

	if got := homeBaseCollide4EBB80(homeBase, ball, 0, state.hooks()); got != 0x89abcdef {
		t.Fatalf("return = %#x, want %#x", got, uint32(0x89abcdef))
	}
	wantTail := []string{
		"random:0:0:C:\\NoxPost\\src\\Server\\Object\\collide\\objcoll.c:0xf57",
		"mutate-markers",
		"first:old",
		"type:old",
		"next:old=new",
		"type:new",
		"clear-owner:ball",
		"move:ball:new",
		"fx:129:ball",
		"store:velocity-x:ball:00000000",
		"store:velocity-y:ball:00000000",
		"store:force-x:ball:00000000",
		"store:pos24-y:ball:00000000",
	}
	if len(state.events) < len(wantTail) {
		t.Fatalf("events too short: %q", state.events)
	}
	assertHomeBaseEvents4EBB80(t, state.events[len(state.events)-len(wantTail):], wantTail)
}

func TestHomeBaseCollide4EBB80NilUpdateFaultFollowsHomeTeamLookup(t *testing.T) {
	ball := &homeBaseTestObject4EBB80{name: "ball", typeIndex: 7}
	homeBase := &homeBaseTestObject4EBB80{name: "home", hasTeam: true, teamID: 4}
	state := &homeBaseTestState4EBB80{
		gameBallType: 7,
		startType:    9,
		teams:        map[uint8]*homeBaseTestTeam4EBB80{4: {name: "home-team"}},
	}
	defer func() {
		if recover() == nil {
			t.Fatal("expected nil update fault")
		}
		assertHomeBaseEvents4EBB80(t, state.events, []string{
			"lookup:GameBall",
			"lookup:GameBallStart",
			"type:ball",
			"update:ball",
			"has-team:home",
			"team-id:home",
			"team:4=home-team",
			"carrier:nil-update",
		})
	}()
	homeBaseCollide4EBB80(homeBase, ball, 0, state.hooks())
}
