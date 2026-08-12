package opennox

import (
	"reflect"
	"testing"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/server"
)

type playerObserverFallbackHooks struct {
	ensureGameBallID func()
	hasFlagBall      func() bool
	findGameBall     func() *server.Object
	firstPlayerUnit  func() *server.Object
	nextPlayerUnit   func(*server.Object) *server.Object
	playerByNetCode  func(uint32) *server.Player
}

func runPlayerObserverFallback4E6150(pl *server.Player, h playerObserverFallbackHooks) *server.Object {
	return playerObserverFallback_4E6150(
		pl,
		h.ensureGameBallID,
		h.hasFlagBall,
		h.findGameBall,
		h.firstPlayerUnit,
		h.nextPlayerUnit,
		h.playerByNetCode,
	)
}

func failPlayerObserverFallbackHooks(t *testing.T) playerObserverFallbackHooks {
	t.Helper()
	unexpected := func(name string) {
		t.Helper()
		t.Fatalf("unexpected %s call", name)
	}
	return playerObserverFallbackHooks{
		ensureGameBallID: func() { unexpected("ensure GameBall ID") },
		hasFlagBall: func() bool {
			unexpected("FlagBall check")
			return false
		},
		findGameBall: func() *server.Object {
			unexpected("find GameBall")
			return nil
		},
		firstPlayerUnit: func() *server.Object {
			unexpected("first player unit")
			return nil
		},
		nextPlayerUnit: func(*server.Object) *server.Object {
			unexpected("next player unit")
			return nil
		},
		playerByNetCode: func(uint32) *server.Player {
			unexpected("player lookup")
			return nil
		},
	}
}

func TestPlayerObserverFallback4E6150SeedsBeforePlayerRead(t *testing.T) {
	h := failPlayerObserverFallbackHooks(t)
	seeded := false
	h.ensureGameBallID = func() { seeded = true }
	defer func() {
		if recover() == nil {
			t.Fatal("nil player returned without a panic")
		}
		if !seeded {
			t.Fatal("GameBall ID was not seeded before the player read")
		}
	}()
	runPlayerObserverFallback4E6150(nil, h)
}

func TestPlayerObserverFallback4E6150StartsAfterCameraPlayer(t *testing.T) {
	camera := &server.Object{ObjClass: object.ClassPlayer}
	dead := &server.Object{ObjClass: object.ClassPlayer, ObjFlags: object.FlagDead, NetCode: 1}
	observing := &server.Object{ObjClass: object.ClassPlayer, NetCode: 2}
	good := &server.Object{ObjClass: object.ClassPlayer, NetCode: 3}
	pl := &server.Player{CameraFollowObj: camera}
	players := map[uint32]*server.Player{
		2: {Field3680: 1},
		3: {},
	}
	next := map[*server.Object]*server.Object{camera: dead, dead: observing, observing: good}
	name := map[*server.Object]string{camera: "camera", dead: "dead", observing: "observing", good: "good"}
	var calls []string
	h := failPlayerObserverFallbackHooks(t)
	h.ensureGameBallID = func() { calls = append(calls, "seed") }
	h.nextPlayerUnit = func(unit *server.Object) *server.Object {
		calls = append(calls, "next "+name[unit])
		return next[unit]
	}
	h.playerByNetCode = func(code uint32) *server.Player {
		calls = append(calls, "player "+name[map[uint32]*server.Object{1: dead, 2: observing, 3: good}[code]])
		return players[code]
	}

	if got := runPlayerObserverFallback4E6150(pl, h); got != good {
		t.Fatalf("result = %p, want %p", got, good)
	}
	want := []string{"seed", "next camera", "player dead", "next dead", "player observing", "next observing", "player good"}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %v, want %v", calls, want)
	}
}

func TestPlayerObserverFallback4E6150FlagBallCandidatePrecedesPlayers(t *testing.T) {
	ball := &server.Object{}
	var calls []string
	h := failPlayerObserverFallbackHooks(t)
	h.ensureGameBallID = func() { calls = append(calls, "seed") }
	h.hasFlagBall = func() bool {
		calls = append(calls, "flag")
		return true
	}
	h.findGameBall = func() *server.Object {
		calls = append(calls, "ball")
		return ball
	}

	if got := runPlayerObserverFallback4E6150(&server.Player{}, h); got != ball {
		t.Fatalf("result = %p, want %p", got, ball)
	}
	if want := []string{"seed", "flag", "ball"}; !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %v, want %v", calls, want)
	}
}

func TestPlayerObserverFallback4E6150NonPlayerCameraStartsAtFirst(t *testing.T) {
	camera := &server.Object{}
	good := &server.Object{ObjClass: object.ClassPlayer, NetCode: 7}
	var calls []string
	h := failPlayerObserverFallbackHooks(t)
	h.ensureGameBallID = func() { calls = append(calls, "seed") }
	h.firstPlayerUnit = func() *server.Object {
		calls = append(calls, "first")
		return good
	}
	h.playerByNetCode = func(code uint32) *server.Player {
		calls = append(calls, "player")
		if code != 7 {
			t.Fatalf("net code = %d, want 7", code)
		}
		return &server.Player{}
	}

	if got := runPlayerObserverFallback4E6150(&server.Player{CameraFollowObj: camera}, h); got != good {
		t.Fatalf("result = %p, want %p", got, good)
	}
	if want := []string{"seed", "first", "player"}; !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %v, want %v", calls, want)
	}
}

func TestPlayerObserverFallback4E6150ExhaustionWrapsThroughBall(t *testing.T) {
	camera := &server.Object{ObjClass: object.ClassPlayer}
	dead := &server.Object{ObjClass: object.ClassPlayer, ObjFlags: object.FlagDead, NetCode: 1}
	good := &server.Object{ObjClass: object.ClassPlayer, NetCode: 2}
	var calls []string
	h := failPlayerObserverFallbackHooks(t)
	h.ensureGameBallID = func() { calls = append(calls, "seed") }
	h.nextPlayerUnit = func(unit *server.Object) *server.Object {
		switch unit {
		case camera:
			calls = append(calls, "next camera")
			return dead
		case dead:
			calls = append(calls, "next dead")
			return nil
		default:
			t.Fatalf("unexpected next unit %p", unit)
			return nil
		}
	}
	h.playerByNetCode = func(code uint32) *server.Player {
		calls = append(calls, "player")
		if code == 1 {
			return nil // The dead flag is read after this lookup and avoids a dereference.
		}
		return &server.Player{}
	}
	h.findGameBall = func() *server.Object {
		calls = append(calls, "ball")
		return nil
	}
	h.firstPlayerUnit = func() *server.Object {
		calls = append(calls, "first")
		return good
	}

	if got := runPlayerObserverFallback4E6150(&server.Player{CameraFollowObj: camera}, h); got != good {
		t.Fatalf("result = %p, want %p", got, good)
	}
	want := []string{"seed", "next camera", "player", "next dead", "ball", "first", "player"}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %v, want %v", calls, want)
	}
}

func TestPlayerObserverFallback4E6150FinalBallIsReturnedDirectly(t *testing.T) {
	dead := &server.Object{ObjClass: object.ClassPlayer, ObjFlags: object.FlagDead, NetCode: 1}
	ball := &server.Object{ObjClass: object.ClassPlayer, ObjFlags: object.FlagDead, NetCode: 2}
	var calls []string
	h := failPlayerObserverFallbackHooks(t)
	h.ensureGameBallID = func() { calls = append(calls, "seed") }
	h.hasFlagBall = func() bool {
		calls = append(calls, "flag")
		return false
	}
	h.firstPlayerUnit = func() *server.Object {
		calls = append(calls, "first")
		return dead
	}
	h.playerByNetCode = func(uint32) *server.Player {
		calls = append(calls, "player")
		return nil
	}
	h.nextPlayerUnit = func(*server.Object) *server.Object {
		calls = append(calls, "next")
		return nil
	}
	h.findGameBall = func() *server.Object {
		calls = append(calls, "ball")
		return ball
	}

	if got := runPlayerObserverFallback4E6150(&server.Player{}, h); got != ball {
		t.Fatalf("result = %p, want direct fallback %p", got, ball)
	}
	want := []string{"seed", "flag", "first", "player", "next", "ball"}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %v, want %v", calls, want)
	}
}

func TestPlayerObserverFallback4E6150RescansFromFirstAfterMissingBall(t *testing.T) {
	stale := &server.Object{ObjClass: object.ClassPlayer, ObjFlags: object.FlagDead, NetCode: 1}
	fresh := &server.Object{ObjClass: object.ClassPlayer, NetCode: 2}
	firstCalls := 0
	var calls []string
	h := failPlayerObserverFallbackHooks(t)
	h.ensureGameBallID = func() { calls = append(calls, "seed") }
	h.hasFlagBall = func() bool {
		calls = append(calls, "flag")
		return false
	}
	h.firstPlayerUnit = func() *server.Object {
		firstCalls++
		calls = append(calls, "first")
		if firstCalls == 1 {
			return stale
		}
		return fresh
	}
	h.playerByNetCode = func(code uint32) *server.Player {
		calls = append(calls, "player")
		if code == 1 {
			return nil
		}
		return &server.Player{}
	}
	h.nextPlayerUnit = func(*server.Object) *server.Object {
		calls = append(calls, "next")
		return nil
	}
	h.findGameBall = func() *server.Object {
		calls = append(calls, "ball")
		return nil
	}

	if got := runPlayerObserverFallback4E6150(&server.Player{}, h); got != fresh {
		t.Fatalf("result = %p, want rescanned %p", got, fresh)
	}
	want := []string{"seed", "flag", "first", "player", "next", "ball", "first", "player"}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %v, want %v", calls, want)
	}
}

func TestPlayerObserverFallback4E6150MissingPlayerInfoPanics(t *testing.T) {
	unit := &server.Object{ObjClass: object.ClassPlayer, NetCode: 9}
	h := failPlayerObserverFallbackHooks(t)
	h.ensureGameBallID = func() {}
	h.hasFlagBall = func() bool { return false }
	h.firstPlayerUnit = func() *server.Object { return unit }
	h.playerByNetCode = func(uint32) *server.Player { return nil }
	defer func() {
		if recover() == nil {
			t.Fatal("missing player info returned without a panic")
		}
	}()
	runPlayerObserverFallback4E6150(&server.Player{}, h)
}
