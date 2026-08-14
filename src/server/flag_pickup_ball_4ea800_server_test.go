package server

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func defaultFlagPickupBallNativeDeps4EA800() flagPickupBallNativeDeps4EA800 {
	return flagPickupBallNativeDeps4EA800{
		loadBallCache:    func() uint32 { return 1 },
		lookupType:       func(string) uint32 { return 0 },
		storeBallCache:   func(uint32) {},
		unitIsGameBall:   func(*Object) int32 { return 0 },
		gameData:         func(uint32) uint16 { return 0 },
		teamByID:         func(uint8) *Team { return nil },
		nextTeam:         func(*Team) *Team { return nil },
		firstTeam:        func() *Team { return nil },
		reportLesson:     func(*Object) {},
		changeTeamScore:  func(*Team, int32) {},
		observerMode:     func() uint32 { return 0 },
		observerUpdate:   func(*Player, *Player) {},
		audio:            func(uint32, *Object) {},
		informScore:      func(uint32, uint32, uint32) {},
		pointFX:          func(uint32, types.Pointf) {},
		setGameFlags:     func(uint32) {},
		flagBallWinner:   func(*Team) {},
		loadStartCache:   func() uint32 { return 1 },
		storeStartCache:  func(uint32) {},
		firstObject:      func() *Object { return nil },
		nextObject:       func(*Object) *Object { return nil },
		randomInt:        func(int32, int32) int32 { return 0 },
		clearOwner:       func(*Object) {},
		dropBall:         func(*Object, *Object) {},
		changeObjectTeam: func(*ObjectTeam, uint32) {},
		setHPMax:         func(*Object) {},
		ticks:            func() uint64 { return 0 },
		moveTo:           func(*Object, types.Pointf) {},
		ballStatus:       func(uint8, uint16) int32 { return 0 },
	}
}

func TestFlagPickupBall4EA800NativeLayout(t *testing.T) {
	wantUpdateSize := uintptr(16)
	wantUpdateField4 := uintptr(4)
	wantUpdateTicks := uintptr(8)
	wantType := uintptr(4)
	wantClass := uintptr(8)
	wantFlags := uintptr(16)
	wantNet := uintptr(36)
	wantTeam := uintptr(48)
	wantPos := uintptr(56)
	wantVelocity := uintptr(80)
	wantForce := uintptr(88)
	wantPos24 := uintptr(96)
	wantNextObject := uintptr(444)
	wantNextOwned := uintptr(512)
	wantFirstOwned := uintptr(516)
	wantUpdate := uintptr(748)
	wantPlayerUpdatePlayer := uintptr(276)
	wantPlayerLessons := uintptr(2136)
	wantTeamLessons := uintptr(52)
	wantTeamID := uintptr(57)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantUpdateSize = 24
		wantUpdateField4 = 8
		wantUpdateTicks = 16
		wantType = 8
		wantClass = 12
		wantFlags = 20
		wantNet = 40
		wantTeam = 52
		wantPos = 60
		wantVelocity = 84
		wantForce = 92
		wantPos24 = 100
		wantNextObject = 448
		wantNextOwned = 560
		wantFirstOwned = 568
		wantUpdate = 872
		wantPlayerUpdatePlayer = 320
		wantPlayerLessons = 2140
		wantTeamLessons = 56
		wantTeamID = 65
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"GameBallUpdateData size", unsafe.Sizeof(GameBallUpdateData4EA800{}), wantUpdateSize},
		{"GameBallUpdateData.Carrier", unsafe.Offsetof(GameBallUpdateData4EA800{}.Carrier), 0},
		{"GameBallUpdateData.Field4", unsafe.Offsetof(GameBallUpdateData4EA800{}.Field4), wantUpdateField4},
		{"GameBallUpdateData.Ticks", unsafe.Offsetof(GameBallUpdateData4EA800{}.Ticks), wantUpdateTicks},
		{"Object.TypeInd", unsafe.Offsetof(Object{}.TypeInd), wantType},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantClass},
		{"Object.ObjFlags", unsafe.Offsetof(Object{}.ObjFlags), wantFlags},
		{"Object.NetCode", unsafe.Offsetof(Object{}.NetCode), wantNet},
		{"Object.TeamVal", unsafe.Offsetof(Object{}.TeamVal), wantTeam},
		{"Object.PosVec", unsafe.Offsetof(Object{}.PosVec), wantPos},
		{"Object.VelVec", unsafe.Offsetof(Object{}.VelVec), wantVelocity},
		{"Object.ForceVec", unsafe.Offsetof(Object{}.ForceVec), wantForce},
		{"Object.Pos24", unsafe.Offsetof(Object{}.Pos24), wantPos24},
		{"Object.ObjNext", unsafe.Offsetof(Object{}.ObjNext), wantNextObject},
		{"Object.Field128", unsafe.Offsetof(Object{}.Field128), wantNextOwned},
		{"Object.Field129", unsafe.Offsetof(Object{}.Field129), wantFirstOwned},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantUpdate},
		{"ObjectTeam.ID", unsafe.Offsetof(ObjectTeam{}.ID), 4},
		{"PlayerUpdateData.Player", unsafe.Offsetof(PlayerUpdateData{}.Player), wantPlayerUpdatePlayer},
		{"Player.Lessons", unsafe.Offsetof(Player{}.Lessons), wantPlayerLessons},
		{"Team.Lessons", unsafe.Offsetof(Team{}.Lessons), wantTeamLessons},
		{"Team.IDVal", unsafe.Offsetof(Team{}.IDVal), wantTeamID},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestFlagPickupBallNative4EA800ResolvesOwnedListThroughNamedFields(t *testing.T) {
	ballUpdate := &GameBallUpdateData4EA800{}
	ball := &Object{TypeInd: 7, UpdateData: unsafe.Pointer(ballUpdate)}
	nonBall := &Object{TypeInd: 6, Field128: ball}
	target := &Object{ObjClass: object.ClassPlayer, Field129: nonBall}
	source := &Object{TeamVal: ObjectTeam{ID: 3}}
	previousTeam := &Team{}
	nextTeam := &Team{}
	deps := defaultFlagPickupBallNativeDeps4EA800()
	deps.loadBallCache = func() uint32 { return 7 }
	deps.unitIsGameBall = func(owner *Object) int32 {
		if owner != target {
			t.Fatal("unitIsGameBall received wrong owner")
		}
		return 1
	}
	deps.teamByID = func(id uint8) *Team {
		if id != 3 {
			t.Fatalf("source team = %d", id)
		}
		return previousTeam
	}
	deps.nextTeam = func(team *Team) *Team {
		if team != previousTeam {
			t.Fatal("wrong previous team")
		}
		return nextTeam
	}
	deps.firstObject = func() *Object {
		t.Fatal("nil carrier reached respawn")
		return nil
	}
	flagPickupBallNative4EA800(source, target, nil, deps)
}

func TestFlagPickupBallNative4EA800ScoresRespawnsAndClearsExactFields(t *testing.T) {
	events := make([]string, 0, 32)
	player := &Player{Lessons: 10}
	carrierPlayerUpdate := &PlayerUpdateData{Player: player}
	inactiveCarrier := &Object{ObjFlags: object.Flags(flagPickupBallInactiveCarrier4EA800)}
	carrier := &Object{
		ObjClass:   object.ClassPlayer,
		NetCode:    0x11111111,
		TeamVal:    ObjectTeam{ID: 5},
		UpdateData: unsafe.Pointer(carrierPlayerUpdate),
	}
	ballUpdate := &GameBallUpdateData4EA800{
		Carrier: inactiveCarrier,
		Field4:  0xaabbccdd,
		Ticks:   7,
	}
	ball := &Object{
		NetCode:    0x22222222,
		TeamVal:    ObjectTeam{ID: 8},
		PosVec:     types.Pointf{X: 10, Y: 20},
		VelVec:     types.Pointf{X: 1, Y: 2},
		ForceVec:   types.Pointf{X: 3, Y: 4},
		Pos24:      types.Pointf{X: 5, Y: 6},
		UpdateData: unsafe.Pointer(ballUpdate),
	}
	source := &Object{TeamVal: ObjectTeam{ID: 4}}
	previousTeam := &Team{}
	scoringTeam := &Team{Lessons: 2, IDVal: 5}
	marker := &Object{TypeInd: 9, PosVec: types.Pointf{X: 30, Y: 40}}
	firstObjectCalls := 0
	deps := defaultFlagPickupBallNativeDeps4EA800()
	deps.loadBallCache = func() uint32 { events = append(events, "ball-cache"); return 7 }
	deps.teamByID = func(id uint8) *Team {
		events = append(events, "team-by-id")
		if id != 4 {
			t.Fatalf("source team ID = %d", id)
		}
		return previousTeam
	}
	deps.nextTeam = func(team *Team) *Team {
		events = append(events, "next-team")
		if team != previousTeam || ballUpdate.Carrier != nil {
			t.Fatalf("team selection = %p, carrier = %p", team, ballUpdate.Carrier)
		}
		ballUpdate.Carrier = carrier
		return scoringTeam
	}
	deps.gameData = func(mode uint32) uint16 {
		events = append(events, "game-data")
		if mode != 64 {
			t.Fatalf("game-data mode = %d", mode)
		}
		return 0
	}
	deps.reportLesson = func(got *Object) {
		events = append(events, "report-lesson")
		if got != carrier {
			t.Fatal("reported wrong carrier")
		}
	}
	deps.changeTeamScore = func(team *Team, score int32) {
		events = append(events, "change-team-score")
		if team != scoringTeam || score != 3 {
			t.Fatalf("team score = %p/%d", team, score)
		}
		team.Lessons = int(score)
	}
	deps.observerMode = func() uint32 { events = append(events, "observer-mode"); return 1 }
	deps.observerUpdate = func(got, other *Player) {
		events = append(events, "observer-update")
		if got != player || other != nil {
			t.Fatalf("observer = %p/%p", got, other)
		}
	}
	deps.audio = func(id uint32, got *Object) {
		events = append(events, "audio")
		if id != 929 || got != source {
			t.Fatalf("audio = %d/%p", id, got)
		}
		carrier.NetCode = 0x33333333
		scoringTeam.IDVal = 6
	}
	deps.informScore = func(code, netCode, teamID uint32) {
		events = append(events, "inform-score")
		if code != 9 || netCode != 0x33333333 || teamID != 6 {
			t.Fatalf("inform = %d/%#x/%d", code, netCode, teamID)
		}
	}
	deps.pointFX = func(code uint32, pos types.Pointf) {
		if code == 154 {
			events = append(events, "score-fx")
			if pos != (types.Pointf{X: 10, Y: 20}) {
				t.Fatalf("score FX position = %v", pos)
			}
			return
		}
		events = append(events, "respawn-fx")
		if code != 129 || pos != marker.PosVec {
			t.Fatalf("respawn FX = %d/%v", code, pos)
		}
	}
	deps.loadStartCache = func() uint32 { events = append(events, "start-cache"); return 9 }
	deps.firstObject = func() *Object {
		firstObjectCalls++
		events = append(events, map[int]string{1: "first-count", 2: "first-select"}[firstObjectCalls])
		return marker
	}
	deps.nextObject = func(got *Object) *Object {
		events = append(events, "next-count")
		if firstObjectCalls != 1 || got != marker {
			t.Fatal("selected marker successor was loaded")
		}
		return nil
	}
	deps.randomInt = func(min, max int32) int32 {
		events = append(events, "random")
		if min != 0 || max != 0 {
			t.Fatalf("random = %d/%d", min, max)
		}
		return 0
	}
	deps.clearOwner = func(got *Object) {
		events = append(events, "clear-owner")
		if got != ball {
			t.Fatal("cleared wrong ball owner")
		}
	}
	deps.dropBall = func(got, owner *Object) {
		events = append(events, "drop-ball")
		if got != ball || owner != nil {
			t.Fatalf("drop = %p/%p", got, owner)
		}
		ball.NetCode = 0x44444444
	}
	deps.changeObjectTeam = func(team *ObjectTeam, netCode uint32) {
		events = append(events, "change-object-team")
		if team != &ball.TeamVal || netCode != 0x44444444 {
			t.Fatalf("change team = %p/%#x", team, netCode)
		}
	}
	deps.setHPMax = func(got *Object) {
		events = append(events, "hp-max")
		if got != ball {
			t.Fatal("set HP on wrong object")
		}
	}
	deps.ticks = func() uint64 {
		events = append(events, "ticks")
		return 0x1122334455667788
	}
	deps.moveTo = func(got *Object, pos types.Pointf) {
		events = append(events, "move")
		if got != ball || pos != marker.PosVec {
			t.Fatalf("move = %p/%v", got, pos)
		}
		got.PosVec = pos
	}
	deps.ballStatus = func(state uint8, netCode uint16) int32 {
		events = append(events, "status")
		if state != 0 || netCode != 0 {
			t.Fatalf("status = %d/%d", state, netCode)
		}
		return -1
	}

	flagPickupBallNative4EA800(source, ball, nil, deps)
	want := []string{
		"ball-cache", "team-by-id", "next-team", "game-data", "report-lesson",
		"change-team-score", "observer-mode", "observer-update", "audio", "inform-score",
		"score-fx", "start-cache", "first-count", "start-cache", "next-count", "random",
		"first-select", "start-cache", "clear-owner", "drop-ball", "change-object-team",
		"hp-max", "ticks", "move", "status", "respawn-fx",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events =\n%v\nwant =\n%v", events, want)
	}
	if player.Lessons != 11 || scoringTeam.Lessons != 3 {
		t.Fatalf("scores = player %d/team %d", player.Lessons, scoringTeam.Lessons)
	}
	if ballUpdate.Carrier != carrier || ballUpdate.Field4 != 0xaabbccdd || ballUpdate.Ticks != 0x1122334455667788 {
		t.Fatalf("ball update = %+v", ballUpdate)
	}
	if ball.VelVec != (types.Pointf{}) {
		t.Fatalf("velocity = %v", ball.VelVec)
	}
	if ball.ForceVec != (types.Pointf{X: 0, Y: 4}) {
		t.Fatalf("force = %v", ball.ForceVec)
	}
	if ball.Pos24 != (types.Pointf{X: 5, Y: 0}) {
		t.Fatalf("Pos24 = %v", ball.Pos24)
	}
}

func TestFlagPickupBallServer4EA800KeepsFourGameBallCachesIndependent(t *testing.T) {
	s := &Server{}
	s.Types.fast.ball = 11
	s.Types.fast.flagCollideGameBall = 22
	s.Types.fast.flagPickupGameBall = 33
	s.Types.fast.flagPickupBallStart = 44
	deps := flagPickupBallServerDeps4EA800(s, FlagPickupBallRuntime4EA800{})
	if deps.loadBallCache() != 33 || deps.loadStartCache() != 44 {
		t.Fatalf("pickup caches = %d/%d", deps.loadBallCache(), deps.loadStartCache())
	}
	deps.storeBallCache(55)
	deps.storeStartCache(66)
	if s.Types.fast.ball != 11 || s.Types.fast.flagCollideGameBall != 22 ||
		s.Types.fast.flagPickupGameBall != 55 || s.Types.fast.flagPickupBallStart != 66 {
		t.Fatalf("caches = common %d/router %d/pickup %d/start %d",
			s.Types.fast.ball, s.Types.fast.flagCollideGameBall,
			s.Types.fast.flagPickupGameBall, s.Types.fast.flagPickupBallStart)
	}
}

func TestFlagPickupBallUnitIsGameBall4EA800UsesCommonOwnedTypeCache(t *testing.T) {
	s := &Server{}
	s.Types.fast.ball = 7
	s.Types.fast.flagCollideGameBall = 20
	s.Types.fast.flagPickupGameBall = 21
	s.Types.fast.flagPickupBallStart = 22
	ball := &Object{TypeInd: 7}
	owner := &Object{Field129: &Object{TypeInd: 6, Field128: ball}}
	if got := flagPickupBallUnitIsGameBall4EA800(s, owner); got != 1 {
		t.Fatalf("unitIsGameBall = %d", got)
	}
	if s.Types.fast.flagCollideGameBall != 20 || s.Types.fast.flagPickupGameBall != 21 || s.Types.fast.flagPickupBallStart != 22 {
		t.Fatal("common owned-type lookup changed collision-local caches")
	}
}
