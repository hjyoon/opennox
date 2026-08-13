package opennox

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/server"
)

func failPlayerLeaveObserverHooks4E6AA0(t *testing.T) playerLeaveObserverHooks_4E6AA0 {
	t.Helper()
	fail := func(name string) { t.Fatalf("unexpected %s call", name) }
	return playerLeaveObserverHooks_4E6AA0{
		isMonsterBot:    func(*server.Object) bool { fail("is-monster-bot"); return false },
		unsetStatus:     func(*server.Player, uint32) { fail("unset-status") },
		disableEnchant:  func(*server.Object, server.EnchantID) { fail("disable-enchant") },
		setPlayerUpdate: func(*server.Object) { fail("set-player-update") },
		markUpdate:      func(*server.Object) { fail("mark-update") },
		gameFlag:        func(noxflags.GameFlag) bool { fail("game-flag"); return false },
		gameplayFlag:    func(noxflags.GameplayFlag) bool { fail("gameplay-flag"); return false },
		teamFlag:        func(*server.Object) *server.Object { fail("team-flag"); return nil },
		pickupTeamFlag:  func(*server.Object, *server.Object) { fail("pickup-team-flag") },
		questListed:     func(*server.Player) int { fail("quest-listed"); return 0 },
		rememberQuest:   func(*server.Player) { fail("remember-quest") },
		firstPlayerUnit: func() *server.Object { fail("first-player-unit"); return nil },
		nextPlayerUnit:  func(*server.Object) *server.Object { fail("next-player-unit"); return nil },
		reportEnchant:   func(ntype.PlayerInd, *server.Object) { fail("report-enchant") },
	}
}

func playerUnitWithQuestState4E6AA0(pl *server.Player) *server.Object {
	ud := &server.PlayerUpdateData{Player: pl}
	return &server.Object{
		ObjClass:   object.ClassPlayer,
		UpdateData: unsafe.Pointer(ud),
	}
}

func TestPlayerLeaveObserver4E6AA0NilGuards(t *testing.T) {
	playerLeaveObserver_4E6AA0(nil, failPlayerLeaveObserverHooks4E6AA0(t))
	playerLeaveObserver_4E6AA0(&server.Player{}, failPlayerLeaveObserverHooks4E6AA0(t))
}

func TestPlayerLeaveObserver4E6AA0MonsterBotGuard(t *testing.T) {
	unit := &server.Object{}
	pl := &server.Player{PlayerUnit: unit}
	h := failPlayerLeaveObserverHooks4E6AA0(t)
	h.isMonsterBot = func(got *server.Object) bool {
		if got != unit {
			t.Fatalf("unit = %p, want %p", got, unit)
		}
		return true
	}
	playerLeaveObserver_4E6AA0(pl, h)
}

func TestPlayerLeaveObserver4E6AA0ReloadsAndOrder(t *testing.T) {
	const keepFlag = object.FlagDestroyed
	cached := &server.Object{ObjFlags: object.FlagNoCollide}
	marked := &server.Object{}
	teamLookupUnit := &server.Object{}
	pickupUnit := &server.Object{}
	flag := &server.Object{}
	pl := &server.Player{PlayerUnit: cached, PlayerInd: 7}

	quest1 := playerUnitWithQuestState4E6AA0(&server.Player{Field4792: 1})
	quest2 := playerUnitWithQuestState4E6AA0(&server.Player{Field4792: 1})
	quest3 := playerUnitWithQuestState4E6AA0(&server.Player{Field4792: 1})
	next := map[*server.Object]*server.Object{quest1: quest2, quest2: quest3}
	name := map[*server.Object]string{quest1: "q1", quest2: "q2", quest3: "q3"}

	var calls []string
	h := failPlayerLeaveObserverHooks4E6AA0(t)
	h.isMonsterBot = func(got *server.Object) bool {
		calls = append(calls, "is-bot")
		return got != cached
	}
	h.unsetStatus = func(got *server.Player, status uint32) {
		calls = append(calls, "unset")
		if got != pl || status != 0x121 {
			t.Fatalf("unset = (%p, %#x), want (%p, %#x)", got, status, pl, uint32(0x121))
		}
		pl.PlayerUnit = marked
	}
	h.disableEnchant = func(got *server.Object, ench server.EnchantID) {
		calls = append(calls, "buff-off")
		if got != cached || ench != server.ENCHANT_INVISIBLE {
			t.Fatalf("buff off = (%p, %d), want (%p, %d)", got, ench, cached, server.ENCHANT_INVISIBLE)
		}
		cached.ObjFlags = object.FlagNoCollide | keepFlag
	}
	h.setPlayerUpdate = func(got *server.Object) {
		calls = append(calls, "set-update")
		if got != cached {
			t.Fatalf("set update unit = %p, want %p", got, cached)
		}
		// GAME.EXE saved the flags immediately before this callback-sized
		// boundary and overwrites this mutation when it clears NoCollide.
		cached.ObjFlags = object.FlagNoCollide
	}
	h.markUpdate = func(got *server.Object) {
		calls = append(calls, "mark")
		if got != marked {
			t.Fatalf("marked unit = %p, want reloaded %p", got, marked)
		}
		pl.PlayerUnit = teamLookupUnit
	}
	h.gameFlag = func(flag noxflags.GameFlag) bool {
		calls = append(calls, "game "+flag.String())
		return true
	}
	h.gameplayFlag = func(flag noxflags.GameplayFlag) bool {
		calls = append(calls, "gameplay")
		return flag == noxflags.GameplayFlag4
	}
	h.teamFlag = func(got *server.Object) *server.Object {
		calls = append(calls, "team-flag")
		if got != teamLookupUnit {
			t.Fatalf("team lookup unit = %p, want %p", got, teamLookupUnit)
		}
		pl.PlayerUnit = pickupUnit
		return flag
	}
	h.pickupTeamFlag = func(gotUnit, gotFlag *server.Object) {
		calls = append(calls, "pickup")
		if gotUnit != pickupUnit || gotFlag != flag {
			t.Fatalf("pickup = (%p, %p), want (%p, %p)", gotUnit, gotFlag, pickupUnit, flag)
		}
	}
	h.questListed = func(got *server.Player) int {
		calls = append(calls, "listed")
		if got != pl {
			t.Fatalf("listed player = %p, want %p", got, pl)
		}
		return 0
	}
	h.rememberQuest = func(got *server.Player) {
		calls = append(calls, "remember")
		if got != pl {
			t.Fatalf("remember player = %p, want %p", got, pl)
		}
	}
	h.firstPlayerUnit = func() *server.Object {
		calls = append(calls, "first")
		return quest1
	}
	h.nextPlayerUnit = func(got *server.Object) *server.Object {
		calls = append(calls, "next "+name[got])
		return next[got]
	}
	h.reportEnchant = func(ind ntype.PlayerInd, got *server.Object) {
		calls = append(calls, "report "+name[got])
		if got == quest1 {
			if ind != 7 {
				t.Fatalf("first report index = %d, want 7", ind)
			}
			pl.PlayerInd = 9
			next[quest1] = quest3
			return
		}
		if got != quest3 || ind != 9 {
			t.Fatalf("later report = (%d, %p), want (9, %p)", ind, got, quest3)
		}
	}

	playerLeaveObserver_4E6AA0(pl, h)

	if got, want := cached.ObjFlags, object.Flags(keepFlag); got != want {
		t.Fatalf("cached flags = %#x, want %#x", got, want)
	}
	wantCalls := []string{
		"is-bot", "unset", "buff-off", "set-update", "mark",
		"game " + noxflags.GameModeKOTR.String(), "gameplay", "team-flag", "pickup",
		"game " + (noxflags.GameFlag15 | noxflags.GameFlag16).String(), "listed", "remember",
		"game " + noxflags.GameModeQuest.String(), "first",
		"report q1", "next q1", "report q3", "next q3",
	}
	if !reflect.DeepEqual(calls, wantCalls) {
		t.Fatalf("calls =\n%v\nwant =\n%v", calls, wantCalls)
	}
}

func TestPlayerLeaveObserver4E6AA0ExactGates(t *testing.T) {
	cached := &server.Object{ObjFlags: object.FlagNoCollide}
	pl := &server.Player{PlayerUnit: cached}
	flagInInventory := &server.Object{InvHolder: &server.Object{}}
	quest := playerUnitWithQuestState4E6AA0(&server.Player{Field4792: 2})

	var pickup, remember, report bool
	h := failPlayerLeaveObserverHooks4E6AA0(t)
	h.isMonsterBot = func(*server.Object) bool { return false }
	h.unsetStatus = func(*server.Player, uint32) {}
	h.disableEnchant = func(*server.Object, server.EnchantID) {}
	h.setPlayerUpdate = func(*server.Object) {}
	h.markUpdate = func(*server.Object) {}
	h.gameFlag = func(noxflags.GameFlag) bool { return true }
	h.gameplayFlag = func(noxflags.GameplayFlag) bool { return true }
	h.teamFlag = func(*server.Object) *server.Object { return flagInInventory }
	h.pickupTeamFlag = func(*server.Object, *server.Object) { pickup = true }
	h.questListed = func(*server.Player) int { return -1 }
	h.rememberQuest = func(*server.Player) { remember = true }
	h.firstPlayerUnit = func() *server.Object { return quest }
	h.nextPlayerUnit = func(*server.Object) *server.Object { return nil }
	h.reportEnchant = func(ntype.PlayerInd, *server.Object) { report = true }

	playerLeaveObserver_4E6AA0(pl, h)
	if pickup || remember || report {
		t.Fatalf("pickup=%v remember=%v report=%v, want all false", pickup, remember, report)
	}
}
