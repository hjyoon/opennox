package server

import (
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func defaultHomeBaseCollideNativeDeps4EBB80() homeBaseCollideNativeDeps4EBB80 {
	return homeBaseCollideNativeDeps4EBB80{
		lookupType:        func(string) uint32 { return 0 },
		teamByID:          func(uint8) *Team { return nil },
		changeScore:       func(*Object, int32) {},
		reportLesson:      func(*Object) {},
		changeTeamLessons: func(*Team, int32) {},
		observerMode:      func() uint32 { return 0 },
		observerUpdate:    func(*Player, *Player) {},
		audio:             func(uint32, *Object) {},
		pointFX:           func(uint32, types.Pointf) uint32 { return 0 },
		firstObject:       func() *Object { return nil },
		nextObject:        func(*Object) *Object { return nil },
		randomInt:         func(int32, int32) int32 { return 0 },
		clearOwner:        func(*Object) {},
		moveTo:            func(*Object, types.Pointf) {},
	}
}

func TestHomeBaseCollide4EBB80NativeLayout(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantType := uintptr(4)
	wantTeam := uintptr(48)
	wantPos := uintptr(56)
	wantVelocity := uintptr(80)
	wantForce := uintptr(88)
	wantPos24 := uintptr(96)
	wantNext := uintptr(444)
	wantUpdate := uintptr(748)
	wantPlayer := uintptr(276)
	wantTeamSize := uintptr(80)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantType = 8
		wantTeam = 52
		wantPos = 60
		wantVelocity = 84
		wantForce = 92
		wantPos24 = 100
		wantNext = 448
		wantUpdate = 872
		wantPlayer = 336
		wantTeamSize = 88
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.TypeInd", unsafe.Offsetof(Object{}.TypeInd), wantType},
		{"Object.TeamVal", unsafe.Offsetof(Object{}.TeamVal), wantTeam},
		{"Object.PosVec", unsafe.Offsetof(Object{}.PosVec), wantPos},
		{"Object.VelVec", unsafe.Offsetof(Object{}.VelVec), wantVelocity},
		{"Object.ForceVec", unsafe.Offsetof(Object{}.ForceVec), wantForce},
		{"Object.Pos24", unsafe.Offsetof(Object{}.Pos24), wantPos24},
		{"Object.ObjNext", unsafe.Offsetof(Object{}.ObjNext), wantNext},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantUpdate},
		{"ObjectTeam size", unsafe.Sizeof(ObjectTeam{}), 8},
		{"ObjectTeam.ID", unsafe.Offsetof(ObjectTeam{}.ID), 4},
		{"GameBallUpdateData.Carrier", unsafe.Offsetof(GameBallUpdateData4EA800{}.Carrier), 0},
		{"PlayerUpdateData.Player", unsafe.Offsetof(PlayerUpdateData{}.Player), wantPlayer},
		{"Team size", unsafe.Sizeof(Team{}), wantTeamSize},
		{"Team.Lessons", unsafe.Offsetof(Team{}.Lessons), 52},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestHomeBaseCollideNative4EBB80BindsNativeFields(t *testing.T) {
	events := make([]string, 0, 32)
	team := &Team{Lessons: math.MaxInt32}
	player := &Player{Lessons: 41}
	playerUpdate := &PlayerUpdateData{Player: player}
	carrier := &Object{
		ObjClass:   object.ClassPlayer,
		TeamVal:    ObjectTeam{ID: 8},
		UpdateData: unsafe.Pointer(playerUpdate),
	}
	ballUpdate := &GameBallUpdateData4EA800{Carrier: carrier}
	ball := &Object{
		TypeInd:    0x2222,
		PosVec:     types.Ptf(10, 20),
		VelVec:     types.Ptf(1, 2),
		ForceVec:   types.Ptf(3, 4),
		Pos24:      types.Ptf(5, 6),
		UpdateData: unsafe.Pointer(ballUpdate),
	}
	homeBase := &Object{TeamVal: ObjectTeam{ID: 7}}
	plain := &Object{TypeInd: 1}
	marker := &Object{TypeInd: 0x3333, PosVec: types.Ptf(30, 40)}
	plain.ObjNext = marker
	firstCalls := 0
	collision := &types.Pointf{X: 0x123, Y: -0x456}

	deps := defaultHomeBaseCollideNativeDeps4EBB80()
	deps.lookupType = func(name string) uint32 {
		events = append(events, "lookup:"+name)
		if name == homeBaseGameBallName4EBB80 {
			return 0x2222
		}
		return 0x3333
	}
	deps.teamByID = func(id uint8) *Team {
		events = append(events, "team")
		if id != 7 && id != 8 {
			t.Fatalf("team ID = %d", id)
		}
		return team
	}
	deps.changeScore = func(got *Object, delta int32) {
		events = append(events, "score")
		if got != carrier || delta != 1 {
			t.Fatalf("score = %p/%d", got, delta)
		}
		got.changeScore(int(delta))
	}
	deps.reportLesson = func(got *Object) {
		events = append(events, "report")
		if got != carrier || player.Lessons != 42 {
			t.Fatalf("report = %p, lessons = %d", got, player.Lessons)
		}
	}
	deps.changeTeamLessons = func(got *Team, lessons int32) {
		events = append(events, "team-score")
		if got != team || lessons != math.MinInt32 {
			t.Fatalf("team score = %p/%d", got, lessons)
		}
		got.Lessons = lessons
	}
	deps.observerMode = func() uint32 {
		events = append(events, "observer-mode")
		return 1
	}
	deps.observerUpdate = func(got, second *Player) {
		events = append(events, "observer")
		if got != player || second != nil {
			t.Fatalf("observer = %p/%p", got, second)
		}
	}
	deps.audio = func(id uint32, got *Object) {
		events = append(events, "audio")
		if id != homeBaseScoreAudio4EBB80 || got != homeBase {
			t.Fatalf("audio = %d/%p", id, got)
		}
	}
	deps.pointFX = func(code uint32, pos types.Pointf) uint32 {
		if code == homeBaseScorePointFX4EBB80 {
			events = append(events, "score-fx")
			if pos != (types.Pointf{X: 10, Y: 20}) {
				t.Fatalf("score FX position = %v", pos)
			}
			return 0xaaaaaaaa
		}
		events = append(events, "respawn-fx")
		if code != homeBaseRespawnPointFX4EBB80 || pos != marker.PosVec {
			t.Fatalf("respawn FX = %d/%v", code, pos)
		}
		return 0xf1234567
	}
	deps.firstObject = func() *Object {
		firstCalls++
		events = append(events, "first")
		return plain
	}
	deps.nextObject = func(obj *Object) *Object {
		events = append(events, "next")
		return obj.Next()
	}
	deps.randomInt = func(minimum, maximum int32) int32 {
		events = append(events, "random")
		if minimum != 0 || maximum != 0 {
			t.Fatalf("random = %d/%d", minimum, maximum)
		}
		return 0
	}
	deps.clearOwner = func(got *Object) {
		events = append(events, "clear-owner")
		if got != ball {
			t.Fatalf("clear owner = %p", got)
		}
	}
	deps.moveTo = func(got *Object, pos types.Pointf) {
		events = append(events, "move")
		if got != ball || pos != marker.PosVec {
			t.Fatalf("move = %p/%v", got, pos)
		}
		got.PosVec = pos
	}

	got := homeBaseCollideNative4EBB80(homeBase, ball, collision, deps)
	if got != 0xf1234567 {
		t.Fatalf("return = %#x, want %#x", got, uint32(0xf1234567))
	}
	if firstCalls != 2 {
		t.Fatalf("first-object calls = %d, want 2", firstCalls)
	}
	if team.Lessons != math.MinInt32 || player.Lessons != 42 {
		t.Fatalf("scores = team %d/player %d", team.Lessons, player.Lessons)
	}
	if ball.PosVec != marker.PosVec {
		t.Fatalf("ball position = %v, want %v", ball.PosVec, marker.PosVec)
	}
	if ball.VelVec != (types.Pointf{}) || ball.ForceVec.X != 0 || ball.Pos24.Y != 0 {
		t.Fatalf("cleared motion = vel %v/force %v/pos24 %v", ball.VelVec, ball.ForceVec, ball.Pos24)
	}
	if ball.ForceVec.Y != 4 || ball.Pos24.X != 5 {
		t.Fatalf("untouched motion = force.y %v/pos24.x %v", ball.ForceVec.Y, ball.Pos24.X)
	}
	if *collision != (types.Pointf{X: 0x123, Y: -0x456}) {
		t.Fatalf("collision was modified: %v", *collision)
	}
	wantEvents := []string{
		"lookup:GameBall", "lookup:GameBallStart", "team", "team", "score", "report",
		"team-score", "observer-mode", "observer", "audio", "score-fx", "first", "next",
		"next", "random", "first", "next", "clear-owner", "move", "respawn-fx",
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %#v, want %#v", events, wantEvents)
	}
}

func TestHomeBaseCollide4EBB80ServerDepsUsePlayerScoreField(t *testing.T) {
	player := &Player{Lessons: 17}
	carrier := &Object{
		ObjClass: object.ClassPlayer,
		UpdateData: unsafe.Pointer(&PlayerUpdateData{
			Player: player,
		}),
	}
	deps := homeBaseCollideServerDeps4EBB80(&Server{}, HomeBaseCollideRuntime4EBB80{})
	deps.changeScore(carrier, 1)
	if player.Lessons != 18 {
		t.Fatalf("player lessons = %d, want 18", player.Lessons)
	}
}
