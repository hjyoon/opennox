package opennox

import (
	"math"
	"reflect"
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

func TestPlayerObserverPrepare4E62F0(t *testing.T) {
	a := &server.Object{ObjFlags: object.FlagDestroyed, PosVec: types.Pointf{X: 1, Y: 2}}
	b := &server.Object{ObjFlags: object.FlagEnabled, PosVec: types.Pointf{X: 3, Y: 4}}
	c := &server.Object{ObjFlags: object.FlagDestroyed | object.FlagEnabled}
	markers := [4]*server.Object{a, b, nil, c}
	targets := []*server.Object{a, b}
	var calls []string
	var positions []types.Pointf
	playerObserverPrepare_4E62F0(
		&markers,
		func() {
			calls = append(calls, "sync")
			if markers[0] != nil || markers[1] != b || markers[2] != nil || markers[3] != nil {
				t.Fatalf("markers were not cleaned before sync: %#v", markers)
			}
		},
		true,
		func() *server.Object {
			calls = append(calls, "camera")
			out := targets[0]
			targets = targets[1:]
			return out
		},
		func(pos types.Pointf) {
			calls = append(calls, "set")
			positions = append(positions, pos)
		},
	)
	if want := []string{"sync", "camera", "set", "camera", "set"}; !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %v, want %v", calls, want)
	}
	if want := []types.Pointf{a.PosVec, b.PosVec}; !reflect.DeepEqual(positions, want) {
		t.Fatalf("positions = %v, want %v", positions, want)
	}
}

func TestPlayerObserverPrepare4E62F0NonReplayReadsCameraOnce(t *testing.T) {
	markers := [4]*server.Object{}
	reads := 0
	playerObserverPrepare_4E62F0(&markers, func() {}, false, func() *server.Object {
		reads++
		return nil
	}, func(types.Pointf) {})
	if reads != 1 {
		t.Fatalf("camera reads = %d, want 1", reads)
	}
}

func TestPlayerObserverClosestOwnedMonster4E6800(t *testing.T) {
	owner := &server.Object{}
	otherOwner := &server.Object{}
	nonMonster := &server.Object{ObjOwner: owner, PosVec: types.Pointf{X: 1}}
	noOwner := &server.Object{ObjClass: object.ClassMonster, PosVec: types.Pointf{X: 1}}
	other := &server.Object{ObjClass: object.ClassMonster, ObjOwner: otherOwner, PosVec: types.Pointf{X: 1}}
	far := &server.Object{ObjClass: object.ClassMonster, ObjOwner: owner, PosVec: types.Pointf{X: 8}}
	tie := &server.Object{ObjClass: object.ClassMonster, ObjOwner: owner, PosVec: types.Pointf{X: -8}}
	near := &server.Object{ObjClass: object.ClassMonster, ObjOwner: owner, PosVec: types.Pointf{X: 2}}
	objects := []*server.Object{nonMonster, noOwner, other, far, tie, near}
	center := types.Pointf{X: 0, Y: 0}
	got := playerObserverClosestOwnedMonster_4E6800(owner, &center, func(rect types.Rectf, fn func(*server.Object) bool) {
		want := types.RectFromPointsf(types.Pointf{X: -100, Y: -100}, types.Pointf{X: 100, Y: 100})
		if rect != want {
			t.Fatalf("rect = %#v, want %#v", rect, want)
		}
		for _, obj := range objects {
			if !fn(obj) {
				t.Fatal("callback stopped iteration")
			}
		}
	})
	if got != near {
		t.Fatalf("closest = %p, want near %p", got, near)
	}

	got = playerObserverClosestOwnedMonster_4E6800(owner, &center, func(_ types.Rectf, fn func(*server.Object) bool) {
		fn(far)
		fn(tie)
	})
	if got != far {
		t.Fatalf("strict tie selected %p, want first %p", got, far)
	}
}

func TestPlayerObserverClosestOwnedMonster4E6800Unordered(t *testing.T) {
	owner := &server.Object{}
	unordered := &server.Object{
		ObjClass: object.ClassMonster,
		ObjOwner: owner,
		PosVec:   types.Pointf{X: float32(math.NaN())},
	}
	center := types.Pointf{}
	got := playerObserverClosestOwnedMonster_4E6800(owner, &center, func(_ types.Rectf, fn func(*server.Object) bool) {
		fn(unordered)
	})
	if got != unordered {
		t.Fatalf("unordered x87 distance selected %p, want %p", got, unordered)
	}
}

func TestPlayerObserverClosestOwnedMonster4E6800ReloadsCenter(t *testing.T) {
	owner := &server.Object{}
	first := &server.Object{ObjClass: object.ClassMonster, ObjOwner: owner, PosVec: types.Pointf{X: 10}}
	second := &server.Object{ObjClass: object.ClassMonster, ObjOwner: owner, PosVec: types.Pointf{X: 21}}
	center := types.Pointf{}
	got := playerObserverClosestOwnedMonster_4E6800(owner, &center, func(_ types.Rectf, fn func(*server.Object) bool) {
		fn(first)
		center.X = 20
		fn(second)
	})
	if got != second {
		t.Fatalf("closest after center reload = %p, want %p", got, second)
	}
}

func TestPlayerObserverPanCamera4E62F0(t *testing.T) {
	pos := types.Pointf{X: 100, Y: -100}
	var checked types.Pointf
	got, ok := playerObserverPanCamera_4E62F0(pos, 0, 0, func(next types.Pointf) bool {
		checked = next
		return true
	})
	if !ok {
		t.Fatal("valid pan was rejected")
	}
	scale := float64(float32(0.1))
	want := types.Pointf{
		X: float32(100 - (100-30)*scale),
		Y: float32(-100 - (30-100)*scale),
	}
	if got != want || checked != want {
		t.Fatalf("pan = %#v, checked %#v, want %#v", got, checked, want)
	}

	got, ok = playerObserverPanCamera_4E62F0(pos, 100, -100, func(types.Pointf) bool { return false })
	if ok || got != pos {
		t.Fatalf("invalid pan = %#v, %v; want original, false", got, ok)
	}
}

func TestPlayerObserverPanCamera4E62F0UnorderedX(t *testing.T) {
	pos := types.Pointf{X: float32(math.NaN()), Y: 7}
	var checked types.Pointf
	_, _ = playerObserverPanCamera_4E62F0(pos, 0, 0, func(next types.Pointf) bool {
		checked = next
		return false
	})
	if !math.IsNaN(float64(checked.X)) || checked.Y != 7 {
		t.Fatalf("unordered pan candidate = %#v", checked)
	}
}

func TestPlayerObserverPanCamera4E62F0SpillsYDelta(t *testing.T) {
	pos := types.Pointf{Y: math.Float32frombits(0x25dbfac1)}
	got, ok := playerObserverPanCamera_4E62F0(pos, 0, 1264358700, func(types.Pointf) bool { return true })
	if !ok {
		t.Fatal("valid pan was rejected")
	}
	dy := float64(float32(float64(pos.Y) - 1264358700))
	wantY := float32(float64(pos.Y) - (30+dy)*float64(float32(0.1)))
	if got.Y != wantY || got.Y != 126435864 {
		t.Fatalf("pan Y = %v (%08x), want %v (%08x)", got.Y, math.Float32bits(got.Y), wantY, math.Float32bits(wantY))
	}
}

func TestPlayerObserverMustLeaveForGameState4E62F0Order(t *testing.T) {
	pl := &server.Player{Field3680: 0x100}
	tests := []struct {
		name       string
		restricted bool
		flag16     bool
		any        bool
		allow      bool
		want       bool
		wantCalls  []string
	}{
		{name: "restricted short circuit", restricted: true, allow: false, want: true, wantCalls: []string{"restricted", "allow"}},
		{name: "flag short circuit", flag16: true, allow: true, want: false, wantCalls: []string{"restricted", "flag16", "allow"}},
		{name: "any players", any: true, allow: false, want: true, wantCalls: []string{"restricted", "flag16", "any", "allow"}},
		{name: "not restricted", allow: false, want: false, wantCalls: []string{"restricted", "flag16", "any"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var calls []string
			got := playerObserverMustLeaveForGameState_4E62F0(
				pl,
				func() bool { calls = append(calls, "restricted"); return tc.restricted },
				func() bool { calls = append(calls, "flag16"); return tc.flag16 },
				func() bool { calls = append(calls, "any"); return tc.any },
				func(gotPl *server.Player) bool {
					calls = append(calls, "allow")
					if gotPl != pl {
						t.Fatalf("allow player = %p, want %p", gotPl, pl)
					}
					return tc.allow
				},
			)
			if got != tc.want || !reflect.DeepEqual(calls, tc.wantCalls) {
				t.Fatalf("got %v calls %v, want %v calls %v", got, calls, tc.want, tc.wantCalls)
			}
		})
	}
}

func TestPlayerObserverMustLeaveForGameState4E62F0StatusGate(t *testing.T) {
	pl := &server.Player{}
	var calls []string
	got := playerObserverMustLeaveForGameState_4E62F0(
		pl,
		func() bool { calls = append(calls, "restricted"); return false },
		func() bool { calls = append(calls, "flag16"); return false },
		func() bool { calls = append(calls, "any"); return true },
		func(*server.Player) bool { calls = append(calls, "allow"); return false },
	)
	if got || !reflect.DeepEqual(calls, []string{"restricted", "flag16"}) {
		t.Fatalf("got %v calls %v", got, calls)
	}
}

func TestPlayerObserverHandleQuest4E62F0ReloadsPlayer(t *testing.T) {
	unit := &server.Object{}
	cached := &server.Player{}
	afterJoin := &server.Player{}
	afterJoined := &server.Player{}
	ud := &server.PlayerUpdateData{Player: cached}
	event := &server.PlayerCtrl{Active: true}
	var calls []string
	handled := playerObserverHandleQuest_4E62F0(unit, ud, cached, event, playerObserverQuestHooks_4E62F0{
		loading: func(*server.Object) { calls = append(calls, "loading") },
		join: func() uint32 {
			calls = append(calls, "join")
			ud.Player = afterJoin
			return 1
		},
		joined: func(*server.Object) {
			calls = append(calls, "joined")
			ud.Player = afterJoined
		},
		full:    func(*server.Object) { calls = append(calls, "full") },
		field79: func(*server.Object) { calls = append(calls, "field79") },
		leave: func(pl *server.Player) {
			calls = append(calls, "leave")
			if pl != cached {
				t.Fatalf("leave player = %p, want cached %p", pl, cached)
			}
		},
	})
	if !handled || event.Active || afterJoin.Field4792 != 1 || afterJoined.Field4792 != 0 {
		t.Fatalf("handled=%v active=%v joined fields=%d/%d", handled, event.Active, afterJoin.Field4792, afterJoined.Field4792)
	}
	want := []string{"join", "joined", "leave"}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %v, want %v", calls, want)
	}
}

func TestPlayerObserverHandleQuest4E62F0Field79KeepsEvent(t *testing.T) {
	unit := &server.Object{}
	pl := &server.Player{Field4792: 1}
	ud := &server.PlayerUpdateData{Player: pl, Field79: 1, Field78: 1}
	event := &server.PlayerCtrl{Active: true}
	var calls []string
	handled := playerObserverHandleQuest_4E62F0(unit, ud, pl, event, playerObserverQuestHooks_4E62F0{
		loading: func(*server.Object) { calls = append(calls, "loading") },
		join:    func() uint32 { calls = append(calls, "join"); return 0 },
		joined:  func(*server.Object) { calls = append(calls, "joined") },
		full:    func(*server.Object) { calls = append(calls, "full") },
		field79: func(got *server.Object) {
			calls = append(calls, "field79")
			if got != unit {
				t.Fatalf("field79 unit = %p, want %p", got, unit)
			}
		},
		leave: func(*server.Player) { calls = append(calls, "leave") },
	})
	if !handled || !event.Active || !reflect.DeepEqual(calls, []string{"field79"}) {
		t.Fatalf("handled=%v active=%v calls=%v", handled, event.Active, calls)
	}
}
