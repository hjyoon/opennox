package opennox

import (
	"reflect"
	"testing"

	"github.com/opennox/opennox/v1/server"
)

func TestPlayerLeaveObsByObserved4E60A0Empty(t *testing.T) {
	var calls []string
	playerLeaveObsByObserved_4E60A0(
		&server.Object{},
		func() *server.Player {
			calls = append(calls, "first")
			return nil
		},
		func(*server.Player) *server.Player {
			calls = append(calls, "next")
			return nil
		},
		func(*server.Player) { calls = append(calls, "leave") },
	)
	if want := []string{"first"}; !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %v, want %v", calls, want)
	}
}

func TestPlayerLeaveObsByObserved4E60A0OrderAndMutationTiming(t *testing.T) {
	target := &server.Object{}
	other := &server.Object{}
	p1 := &server.Player{CameraFollowObj: target}
	p2 := &server.Player{CameraFollowObj: other}
	p3 := &server.Player{CameraFollowObj: target}
	nextByPlayer := map[*server.Player]*server.Player{p1: p2, p2: p3}
	name := map[*server.Player]string{p1: "p1", p2: "p2", p3: "p3"}
	var calls []string

	playerLeaveObsByObserved_4E60A0(
		target,
		func() *server.Player {
			calls = append(calls, "first")
			return p1
		},
		func(pl *server.Player) *server.Player {
			calls = append(calls, "next "+name[pl])
			return nextByPlayer[pl]
		},
		func(pl *server.Player) {
			calls = append(calls, "leave "+name[pl])
			if pl == p1 {
				// The next player's target is read only after arriving there.
				p2.CameraFollowObj = target
			}
		},
	)

	want := []string{
		"first",
		"leave p1", "next p1",
		"leave p2", "next p2",
		"leave p3", "next p3",
	}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %v, want %v", calls, want)
	}
}

func TestPlayerLeaveObsByObserved4E60A0ReadsNextAfterLeave(t *testing.T) {
	target := &server.Object{}
	p1 := &server.Player{CameraFollowObj: target}
	p2 := &server.Player{CameraFollowObj: target}
	p3 := &server.Player{CameraFollowObj: target}
	nextByPlayer := map[*server.Player]*server.Player{p1: p2, p2: p3}
	var left []*server.Player

	playerLeaveObsByObserved_4E60A0(
		target,
		func() *server.Player { return p1 },
		func(pl *server.Player) *server.Player { return nextByPlayer[pl] },
		func(pl *server.Player) {
			left = append(left, pl)
			if pl == p1 {
				// The callback's list mutation controls the successor lookup.
				nextByPlayer[p1] = p3
			}
		},
	)

	if want := []*server.Player{p1, p3}; !reflect.DeepEqual(left, want) {
		t.Fatalf("left = %v, want %v", left, want)
	}
}

func TestPlayerLeaveObsByObserved4E60A0MatchesNilTarget(t *testing.T) {
	p1 := &server.Player{}
	p2 := &server.Player{CameraFollowObj: &server.Object{}}
	var left []*server.Player
	playerLeaveObsByObserved_4E60A0(
		nil,
		func() *server.Player { return p1 },
		func(pl *server.Player) *server.Player {
			if pl == p1 {
				return p2
			}
			return nil
		},
		func(pl *server.Player) { left = append(left, pl) },
	)
	if want := []*server.Player{p1}; !reflect.DeepEqual(left, want) {
		t.Fatalf("left = %v, want %v", left, want)
	}
}
