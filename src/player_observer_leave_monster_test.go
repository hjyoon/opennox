package opennox

import (
	"reflect"
	"testing"

	"github.com/opennox/opennox/v1/server"
)

type playerLeaveMonsterObserverHooks struct {
	getPossess   func(*server.Object) *server.Object
	findGood     func(*server.Player) *server.Object
	clearObserve func(*server.Object)
	findFallback func(*server.Player) *server.Object
	unlock       func(*server.Object)
	follow       func(*server.Object, *server.Object)
}

func runPlayerLeaveMonsterObserver4E60E0(pl *server.Player, h playerLeaveMonsterObserverHooks) {
	playerLeaveMonsterObserver_4E60E0(
		pl,
		h.getPossess,
		h.findGood,
		h.clearObserve,
		h.findFallback,
		h.unlock,
		h.follow,
	)
}

func failPlayerLeaveMonsterObserverHooks(t *testing.T) playerLeaveMonsterObserverHooks {
	t.Helper()
	unexpected := func(name string) {
		t.Helper()
		t.Fatalf("unexpected %s call", name)
	}
	return playerLeaveMonsterObserverHooks{
		getPossess: func(*server.Object) *server.Object { unexpected("get possess"); return nil },
		findGood: func(*server.Player) *server.Object {
			unexpected("find good slave")
			return nil
		},
		clearObserve: func(*server.Object) { unexpected("clear observe") },
		findFallback: func(*server.Player) *server.Object {
			unexpected("find fallback")
			return nil
		},
		unlock: func(*server.Object) { unexpected("unlock") },
		follow: func(*server.Object, *server.Object) {
			unexpected("follow")
		},
	}
}

func TestPlayerLeaveMonsterObserver4E60E0NilUnit(t *testing.T) {
	runPlayerLeaveMonsterObserver4E60E0(&server.Player{}, failPlayerLeaveMonsterObserverHooks(t))
}

func TestPlayerLeaveMonsterObserver4E60E0DoesNotGuardNilPlayer(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("nil player returned without a panic")
		}
	}()
	runPlayerLeaveMonsterObserver4E60E0(nil, failPlayerLeaveMonsterObserverHooks(t))
}

func TestPlayerLeaveMonsterObserver4E60E0ClearsAfterMissingGoodSlave(t *testing.T) {
	u1 := &server.Object{}
	u2 := &server.Object{}
	possess := &server.Object{}
	pl := &server.Player{PlayerUnit: u1}
	var calls []string
	h := failPlayerLeaveMonsterObserverHooks(t)
	h.getPossess = func(got *server.Object) *server.Object {
		if got != u1 {
			t.Fatalf("get possess unit = %p, want %p", got, u1)
		}
		calls = append(calls, "possess")
		return possess
	}
	h.findGood = func(got *server.Player) *server.Object {
		if got != pl {
			t.Fatalf("find good player = %p, want %p", got, pl)
		}
		calls = append(calls, "good")
		pl.PlayerUnit = u2
		return nil
	}
	h.clearObserve = func(got *server.Object) {
		if got != u2 {
			t.Fatalf("clear unit = %p, want reloaded %p", got, u2)
		}
		calls = append(calls, "clear")
	}

	runPlayerLeaveMonsterObserver4E60E0(pl, h)
	if want := []string{"possess", "good", "clear"}; !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %v, want %v", calls, want)
	}
}

func TestPlayerLeaveMonsterObserver4E60E0GoodSlaveReloadsUnit(t *testing.T) {
	u1 := &server.Object{}
	u2 := &server.Object{}
	u3 := &server.Object{}
	possess := &server.Object{}
	target := &server.Object{}
	pl := &server.Player{PlayerUnit: u1}
	var calls []string
	h := failPlayerLeaveMonsterObserverHooks(t)
	h.getPossess = func(got *server.Object) *server.Object {
		if got != u1 {
			t.Fatalf("get possess unit = %p, want %p", got, u1)
		}
		calls = append(calls, "possess")
		return possess
	}
	h.findGood = func(got *server.Player) *server.Object {
		if got != pl {
			t.Fatalf("find good player = %p, want %p", got, pl)
		}
		calls = append(calls, "good")
		pl.PlayerUnit = u2
		return target
	}
	h.unlock = func(got *server.Object) {
		if got != u2 {
			t.Fatalf("unlock unit = %p, want reloaded %p", got, u2)
		}
		calls = append(calls, "unlock")
		pl.PlayerUnit = u3
	}
	h.follow = func(gotUnit, gotTarget *server.Object) {
		if gotUnit != u3 || gotTarget != target {
			t.Fatalf("follow = (%p, %p), want (%p, %p)", gotUnit, gotTarget, u3, target)
		}
		calls = append(calls, "follow")
	}

	runPlayerLeaveMonsterObserver4E60E0(pl, h)
	if want := []string{"possess", "good", "unlock", "follow"}; !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %v, want %v", calls, want)
	}
}

func TestPlayerLeaveMonsterObserver4E60E0FallbackAndNilTarget(t *testing.T) {
	u1 := &server.Object{}
	u2 := &server.Object{}
	u3 := &server.Object{}
	pl := &server.Player{PlayerUnit: u1}
	var calls []string
	h := failPlayerLeaveMonsterObserverHooks(t)
	h.getPossess = func(got *server.Object) *server.Object {
		if got != u1 {
			t.Fatalf("get possess unit = %p, want %p", got, u1)
		}
		calls = append(calls, "possess")
		pl.PlayerUnit = u2
		return nil
	}
	h.findFallback = func(got *server.Player) *server.Object {
		if got != pl {
			t.Fatalf("fallback player = %p, want %p", got, pl)
		}
		calls = append(calls, "fallback")
		return nil
	}
	h.unlock = func(got *server.Object) {
		if got != u2 {
			t.Fatalf("unlock unit = %p, want reloaded %p", got, u2)
		}
		calls = append(calls, "unlock")
		pl.PlayerUnit = u3
	}
	h.follow = func(gotUnit, gotTarget *server.Object) {
		if gotUnit != u3 || gotTarget != nil {
			t.Fatalf("follow = (%p, %p), want (%p, nil)", gotUnit, gotTarget, u3)
		}
		calls = append(calls, "follow")
	}

	runPlayerLeaveMonsterObserver4E60E0(pl, h)
	if want := []string{"possess", "fallback", "unlock", "follow"}; !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %v, want %v", calls, want)
	}
}
