package opennox

import (
	"fmt"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/server"
)

func noopPlayerGoObserverHooks4E6860() playerGoObserverHooks_4E6860 {
	return playerGoObserverHooks_4E6860{
		abilityActive:     func(*server.Object) int { return 0 },
		isMonsterBot:      func(*server.Object) bool { return false },
		gameFlag:          func(noxflags.GameFlag) bool { return false },
		ensureCrownID:     func() {},
		ensureGameBallID:  func() {},
		crownID:           func() uint32 { return 0 },
		gameBallID:        func() uint32 { return 0 },
		dropCrown:         func(*server.Object, *server.Object, *types.Pointf) {},
		clearOwner:        func(*server.Object) {},
		gameBallDropped:   func() {},
		dropFlag:          func(*server.Object, *server.Object, *types.Pointf) {},
		getPossess:        func(*server.Object) *server.Object { return nil },
		clearObserve:      func(*server.Object) {},
		needTimestamp:     func(*server.Player) {},
		anyPlayers:        func() int { return 1 },
		resetState:        func() {},
		forceLessons:      func() {},
		resetTeams:        func() {},
		finishReset:       func() {},
		inform:            func(ntype.PlayerInd, byte, uint32) {},
		applyInvisible:    func(*server.Object) {},
		unlockCamera:      func(*server.Object) {},
		leaveObserver:     func(*server.Player) {},
		removeSpawned:     func(*server.Object) {},
		setObserverUpdate: func(*server.Object) {},
		resetCamping:      func(*server.Object) {},
	}
}

func failPlayerGoObserverHooks4E6860(t *testing.T) playerGoObserverHooks_4E6860 {
	t.Helper()
	h := noopPlayerGoObserverHooks4E6860()
	fail := func(name string) {
		t.Helper()
		t.Fatalf("unexpected %s call", name)
	}
	h.abilityActive = func(*server.Object) int { fail("ability"); return 0 }
	h.isMonsterBot = func(*server.Object) bool { fail("monster-bot"); return false }
	h.gameFlag = func(noxflags.GameFlag) bool { fail("game-flag"); return false }
	h.ensureCrownID = func() { fail("ensure-crown-id") }
	h.ensureGameBallID = func() { fail("ensure-game-ball-id") }
	h.crownID = func() uint32 { fail("crown-id"); return 0 }
	h.gameBallID = func() uint32 { fail("game-ball-id"); return 0 }
	h.dropCrown = func(*server.Object, *server.Object, *types.Pointf) { fail("drop-crown") }
	h.clearOwner = func(*server.Object) { fail("clear-owner") }
	h.gameBallDropped = func() { fail("game-ball-dropped") }
	h.dropFlag = func(*server.Object, *server.Object, *types.Pointf) { fail("drop-flag") }
	h.getPossess = func(*server.Object) *server.Object { fail("get-possess"); return nil }
	h.clearObserve = func(*server.Object) { fail("clear-observe") }
	h.needTimestamp = func(*server.Player) { fail("timestamp") }
	h.anyPlayers = func() int { fail("any-players"); return 0 }
	h.resetState = func() { fail("reset-state") }
	h.forceLessons = func() { fail("force-lessons") }
	h.resetTeams = func() { fail("reset-teams") }
	h.finishReset = func() { fail("finish-reset") }
	h.inform = func(ntype.PlayerInd, byte, uint32) { fail("inform") }
	h.applyInvisible = func(*server.Object) { fail("enchant") }
	h.unlockCamera = func(*server.Object) { fail("unlock") }
	h.leaveObserver = func(*server.Player) { fail("leave") }
	h.removeSpawned = func(*server.Object) { fail("remove-spawned") }
	h.setObserverUpdate = func(*server.Object) { fail("set-update") }
	h.resetCamping = func(*server.Object) { fail("reset-camping") }
	return h
}

func TestPlayerGoObserver4E6860Guards(t *testing.T) {
	if got := playerGoObserver_4E6860(nil, 0, 0, failPlayerGoObserverHooks4E6860(t)); got != 1 {
		t.Fatalf("nil player result = %d, want 1", got)
	}
	if got := playerGoObserver_4E6860(&server.Player{}, 0, 0, failPlayerGoObserverHooks4E6860(t)); got != 1 {
		t.Fatalf("nil unit result = %d, want 1", got)
	}

	unit := &server.Object{}
	pl := &server.Player{PlayerUnit: unit}
	h := failPlayerGoObserverHooks4E6860(t)
	h.abilityActive = func(got *server.Object) int {
		if got != unit {
			t.Fatalf("ability unit = %p, want %p", got, unit)
		}
		return 1
	}
	if got := playerGoObserver_4E6860(pl, 0, 0, h); got != 0 {
		t.Fatalf("active ability result = %d, want 0", got)
	}

	h = failPlayerGoObserverHooks4E6860(t)
	h.isMonsterBot = func(got *server.Object) bool {
		if got != unit {
			t.Fatalf("monster-bot unit = %p, want %p", got, unit)
		}
		return true
	}
	if got := playerGoObserver_4E6860(pl, 0, 1, h); got != 0 {
		t.Fatalf("monster-bot result = %d, want 0", got)
	}
}

func TestPlayerGoObserver4E6860OrderReloadsAndWidths(t *testing.T) {
	oldUD := &server.PlayerUpdateData{CurTraps: 0xaabbcc07}
	newUD := &server.PlayerUpdateData{CurTraps: 9}
	unit := &server.Object{
		ObjClass:   object.ClassPlayer,
		ObjFlags:   object.FlagActive,
		PosVec:     types.Pointf{X: 1, Y: 2},
		UpdateData: unsafe.Pointer(oldUD),
	}
	ownedUnit := &server.Object{PosVec: types.Pointf{X: 10, Y: 20}}
	crownUnit := &server.Object{PosVec: types.Pointf{X: 30, Y: 40}}
	possessUnit := &server.Object{}
	clearUnit := &server.Object{}
	stale := &server.Object{TypeInd: 99}
	crown := &server.Object{TypeInd: 11, Field128: stale}
	ballOwner := &server.Object{}
	ballMarker := &server.Object{}
	ball := &server.Object{
		TypeInd:  22,
		ObjFlags: object.FlagActive | object.FlagNoCollide,
		ObjOwner: ballOwner,
		Obj130:   ballMarker,
		Field128: stale,
	}
	flag := &server.Object{TypeInd: 33, ObjClass: object.ClassFlag}
	ownedUnit.Field129 = crown
	pl := &server.Player{
		PlayerUnit:      unit,
		PlayerInd:       0xe1,
		CameraFollowObj: &server.Object{},
	}

	var calls []string
	add := func(name string) { calls = append(calls, name) }
	h := noopPlayerGoObserverHooks4E6860()
	h.abilityActive = func(got *server.Object) int {
		add("ability")
		if got != unit {
			t.Fatalf("ability unit = %p, want %p", got, unit)
		}
		unit.UpdateData = unsafe.Pointer(newUD)
		return 2 // GAME.EXE rejects exactly 1, not arbitrary nonzero values.
	}
	h.isMonsterBot = func(got *server.Object) bool {
		add("monster-bot")
		return got != unit
	}
	h.gameFlag = func(flag noxflags.GameFlag) bool {
		add(fmt.Sprintf("flag:%d", flag))
		switch flag {
		case noxflags.GameModeKOTR | noxflags.GameModeCTF | noxflags.GameModeFlagBall:
			return true
		case noxflags.GameModeQuest:
			return false
		case noxflags.GameModeCoop:
			return false
		case noxflags.GameModeFlagBall:
			return true
		default:
			t.Fatalf("unexpected game flag %d", flag)
			return false
		}
	}
	h.ensureCrownID = func() {
		add("ensure-crown-id")
	}
	h.ensureGameBallID = func() {
		add("ensure-game-ball-id")
		pl.PlayerUnit = ownedUnit
	}
	h.crownID = func() uint32 {
		add("crown-id-cached")
		return 11
	}
	ballID := uint32(99)
	h.gameBallID = func() uint32 {
		add("game-ball-id-cached")
		return ballID
	}
	h.dropCrown = func(owner, item *server.Object, pos *types.Pointf) {
		add("drop-crown")
		if owner != ownedUnit || item != crown || pos != &ownedUnit.PosVec || *pos != ownedUnit.PosVec {
			t.Fatalf("crown drop = (%p, %p, %#v)", owner, item, pos)
		}
		crown.Field128 = ball
		pl.PlayerUnit = crownUnit
		ballID = 22
	}
	h.clearOwner = func(item *server.Object) {
		add("clear-owner")
		if item != ball || item.Obj130 != nil || item.ObjFlags != object.FlagActive {
			t.Fatalf("ball state before clear = %#v", item)
		}
		item.ObjOwner = nil
		ball.Field128 = flag
	}
	h.gameBallDropped = func() {
		add("game-ball-dropped")
		if ball.ObjOwner != nil {
			t.Fatal("game-ball notification preceded owner clear")
		}
	}
	h.dropFlag = func(owner, item *server.Object, pos *types.Pointf) {
		add("drop-flag")
		if owner != crownUnit || item != flag || pos != &crownUnit.PosVec || *pos != crownUnit.PosVec {
			t.Fatalf("flag drop = (%p, %p, %#v)", owner, item, pos)
		}
		pl.PlayerUnit = possessUnit
	}
	h.getPossess = func(got *server.Object) *server.Object {
		add("get-possess")
		if got != possessUnit {
			t.Fatalf("possess unit = %p, want %p", got, possessUnit)
		}
		pl.PlayerUnit = clearUnit
		return &server.Object{}
	}
	h.clearObserve = func(got *server.Object) {
		add("clear-observe")
		if got != clearUnit {
			t.Fatalf("clear unit = %p, want %p", got, clearUnit)
		}
	}
	h.needTimestamp = func(got *server.Player) {
		add("timestamp")
		if got != pl {
			t.Fatalf("timestamp player = %p, want %p", got, pl)
		}
	}
	h.anyPlayers = func() int { add("any-players"); return 0 }
	h.resetState = func() { add("reset-state") }
	h.forceLessons = func() { add("force-lessons") }
	h.resetTeams = func() { add("reset-teams") }
	h.finishReset = func() { add("finish-reset") }
	h.inform = func(ind ntype.PlayerInd, code byte, value uint32) {
		add("inform")
		if ind != 0xe1 || code != 12 || value != 0x89abcdef {
			t.Fatalf("inform = (%d, %d, %#x)", ind, code, value)
		}
	}
	h.applyInvisible = func(got *server.Object) {
		add("enchant")
		if got != unit {
			t.Fatalf("enchant unit = %p, want %p", got, unit)
		}
		unit.ObjFlags = object.Flags(0x80000000)
		unit.PosVec = types.Pointf{X: 123.5, Y: -456.25}
	}
	h.unlockCamera = func(got *server.Object) {
		add("unlock")
		if got != unit || unit.ObjFlags != object.Flags(0x80000040) || pl.Pos3632Vec != unit.PosVec {
			t.Fatalf("unlock state: unit=%p flags=%#x player=%#v unit-pos=%#v", got, unit.ObjFlags, pl.Pos3632Vec, unit.PosVec)
		}
		pl.CameraFollowObj = nil
	}
	h.leaveObserver = func(got *server.Player) {
		add("leave")
		if got != pl {
			t.Fatalf("leave player = %p, want %p", got, pl)
		}
	}
	h.removeSpawned = func(got *server.Object) {
		add("remove-spawned")
		if got != unit {
			t.Fatalf("remove-spawned unit = %p, want %p", got, unit)
		}
		oldUD.CurTraps = 0xaabbcc2a
	}
	h.setObserverUpdate = func(got *server.Object) {
		add("set-update")
		if got != unit || oldUD.CurTraps != 0xaabbcc00 || newUD.CurTraps != 9 {
			t.Fatalf("cached update data was not cleared: old=%d new=%d", oldUD.CurTraps, newUD.CurTraps)
		}
	}
	h.resetCamping = func(got *server.Object) {
		add("reset-camping")
		if got != unit {
			t.Fatalf("reset-camping unit = %p, want %p", got, unit)
		}
	}

	notify := uint32(0x89abcdef)
	if got := playerGoObserver_4E6860(pl, int(notify), 0, h); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	want := []string{
		"ability", "monster-bot", "flag:112", "ensure-crown-id", "ensure-game-ball-id",
		"crown-id-cached", "drop-crown", "crown-id-cached", "game-ball-id-cached",
		"clear-owner", "game-ball-dropped", "crown-id-cached", "game-ball-id-cached", "drop-flag",
		"get-possess", "clear-observe", "timestamp", "any-players", "flag:4096",
		"reset-state", "force-lessons", "reset-teams", "finish-reset", "inform",
		"enchant", "unlock", "flag:2048", "flag:64", "leave", "remove-spawned",
		"set-update", "reset-camping",
	}
	if fmt.Sprint(calls) != fmt.Sprint(want) {
		t.Fatalf("calls = %v, want %v", calls, want)
	}
}

func TestPlayerGoObserver4E6860CoopAndShortCircuits(t *testing.T) {
	ud := &server.PlayerUpdateData{CurTraps: 0x11223301}
	unit := &server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(ud)}
	pl := &server.Player{PlayerUnit: unit, CameraFollowObj: &server.Object{}}
	h := noopPlayerGoObserverHooks4E6860()
	var flags []noxflags.GameFlag
	h.gameFlag = func(flag noxflags.GameFlag) bool {
		flags = append(flags, flag)
		switch flag {
		case noxflags.GameModeKOTR | noxflags.GameModeCTF | noxflags.GameModeFlagBall:
			return false
		case noxflags.GameModeCoop:
			return true
		default:
			t.Fatalf("unexpected non-short-circuited flag %d", flag)
			return false
		}
	}
	h.abilityActive = func(*server.Object) int {
		t.Fatal("keep did not skip ability check")
		return 0
	}
	h.anyPlayers = func() int { return -1 }
	h.resetState = func() { t.Fatal("players-present path reset state") }
	h.leaveObserver = func(*server.Player) { t.Fatal("Coop path entered FlagBall leave") }

	if got := playerGoObserver_4E6860(pl, 0, 7, h); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	wantFlags := []noxflags.GameFlag{
		noxflags.GameModeKOTR | noxflags.GameModeCTF | noxflags.GameModeFlagBall,
		noxflags.GameModeCoop,
	}
	if fmt.Sprint(flags) != fmt.Sprint(wantFlags) {
		t.Fatalf("flags = %v, want %v", flags, wantFlags)
	}
	if pl.Field3672 != 1 || pl.CameraFollowObj != nil || ud.CurTraps != 0x11223300 {
		t.Fatalf("Coop final state = field3672:%d camera:%p traps:%d", pl.Field3672, pl.CameraFollowObj, ud.CurTraps)
	}
}
