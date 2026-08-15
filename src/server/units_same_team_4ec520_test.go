package server

import (
	"fmt"
	"reflect"
	"testing"
)

type unitsSameTeamTestObject4EC520 struct {
	name  string
	team  int
	owner *unitsSameTeamTestObject4EC520
}

type unitsSameTeamTestRecord4EC520 struct {
	id int
}

type unitsSameTeamAddressObject4EC520 struct {
	name  string
	team  unitsSameTeamTestRecord4EC520
	owner *unitsSameTeamAddressObject4EC520
}

func TestUnitsHaveSameTeam4EC520NilInputsDoNotReadObjects(t *testing.T) {
	obj := &unitsSameTeamTestObject4EC520{name: "object"}
	hooks := unitsHaveSameTeamHooks4EC520[*unitsSameTeamTestObject4EC520, int]{
		team: func(*unitsSameTeamTestObject4EC520) int { t.Fatal("nil gate read team"); return 0 },
		owner: func(*unitsSameTeamTestObject4EC520) *unitsSameTeamTestObject4EC520 {
			t.Fatal("nil gate read owner")
			return nil
		},
		teamEqual: func(int, int) int32 { t.Fatal("nil gate compared teams"); return 0 },
	}
	if got := unitsHaveSameTeam4EC520(nil, obj, hooks); got != 0 {
		t.Fatalf("nil first result = %d, want 0", got)
	}
	if got := unitsHaveSameTeam4EC520(obj, nil, hooks); got != 0 {
		t.Fatalf("nil second result = %d, want 0", got)
	}
}

func TestUnitsHaveSameTeam4EC520NestedTraversalOrder(t *testing.T) {
	left2 := &unitsSameTeamTestObject4EC520{name: "left2", team: 2}
	left1 := &unitsSameTeamTestObject4EC520{name: "left1", team: 1, owner: left2}
	right2 := &unitsSameTeamTestObject4EC520{name: "right2", team: 4}
	right1 := &unitsSameTeamTestObject4EC520{name: "right1", team: 3, owner: right2}
	var events []string
	got := unitsHaveSameTeam4EC520(left1, right1, unitsHaveSameTeamHooks4EC520[
		*unitsSameTeamTestObject4EC520,
		int,
	]{
		team: func(obj *unitsSameTeamTestObject4EC520) int {
			events = append(events, "team:"+obj.name)
			return obj.team
		},
		owner: func(obj *unitsSameTeamTestObject4EC520) *unitsSameTeamTestObject4EC520 {
			events = append(events, "owner:"+obj.name)
			return obj.owner
		},
		teamEqual: func(left, right int) int32 {
			events = append(events, fmt.Sprintf("compare:%d%d", left, right))
			return 0
		},
	})
	if got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	want := []string{
		"team:left1", "team:right1", "compare:13", "owner:right1",
		"team:right2", "compare:14", "owner:right2", "owner:left1",
		"team:left2", "team:right1", "compare:23", "owner:right1",
		"team:right2", "compare:24", "owner:right2", "owner:left2",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
}

func TestUnitsHaveSameTeam4EC520TeamResultAndIdentityShortCircuit(t *testing.T) {
	left := &unitsSameTeamTestObject4EC520{name: "left", team: 7}
	right := &unitsSameTeamTestObject4EC520{name: "right", team: 7}
	ownerCalls := 0
	if got := unitsHaveSameTeam4EC520(left, right, unitsHaveSameTeamHooks4EC520[
		*unitsSameTeamTestObject4EC520,
		int,
	]{
		team: func(obj *unitsSameTeamTestObject4EC520) int { return obj.team },
		owner: func(obj *unitsSameTeamTestObject4EC520) *unitsSameTeamTestObject4EC520 {
			ownerCalls++
			return obj.owner
		},
		teamEqual: func(int, int) int32 { return -9 },
	}); got != 1 || ownerCalls != 0 {
		t.Fatalf("team match result/owner calls = (%d, %d), want (1, 0)", got, ownerCalls)
	}

	teamCalls := 0
	compareCalls := 0
	if got := unitsHaveSameTeam4EC520(left, left, unitsHaveSameTeamHooks4EC520[
		*unitsSameTeamTestObject4EC520,
		int,
	]{
		team: func(obj *unitsSameTeamTestObject4EC520) int {
			teamCalls++
			return obj.team
		},
		owner: func(obj *unitsSameTeamTestObject4EC520) *unitsSameTeamTestObject4EC520 {
			t.Fatal("identity match read owner")
			return nil
		},
		teamEqual: func(int, int) int32 {
			compareCalls++
			return 0
		},
	}); got != 1 || teamCalls != 2 || compareCalls != 1 {
		t.Fatalf("identity result/team/compare = (%d, %d, %d), want (1, 2, 1)", got, teamCalls, compareCalls)
	}
}

func TestUnitsHaveSameTeam4EC520ReadsMutatedOwnerLinksAfterComparison(t *testing.T) {
	leftReplacement := &unitsSameTeamTestObject4EC520{name: "left-replacement", team: 8}
	rightReplacement := &unitsSameTeamTestObject4EC520{name: "right-replacement", team: 8}
	left := &unitsSameTeamTestObject4EC520{name: "left", team: 1}
	right := &unitsSameTeamTestObject4EC520{name: "right", team: 2}
	comparisons := 0
	got := unitsHaveSameTeam4EC520(left, right, unitsHaveSameTeamHooks4EC520[
		*unitsSameTeamTestObject4EC520,
		int,
	]{
		team:  func(obj *unitsSameTeamTestObject4EC520) int { return obj.team },
		owner: func(obj *unitsSameTeamTestObject4EC520) *unitsSameTeamTestObject4EC520 { return obj.owner },
		teamEqual: func(first, second int) int32 {
			comparisons++
			if comparisons == 1 {
				left.owner = leftReplacement
				right.owner = rightReplacement
			}
			if first == second && first == 8 {
				return 1
			}
			return 0
		},
	})
	if got != 1 || comparisons != 4 {
		t.Fatalf("result/comparisons = (%d, %d), want (1, 4)", got, comparisons)
	}
}

func TestUnitsHaveSameTeam4EC520CachesLeftTeamAddressWithLiveContents(t *testing.T) {
	left := &unitsSameTeamAddressObject4EC520{name: "left", team: unitsSameTeamTestRecord4EC520{id: 1}}
	right2 := &unitsSameTeamAddressObject4EC520{name: "right2", team: unitsSameTeamTestRecord4EC520{id: 3}}
	right1 := &unitsSameTeamAddressObject4EC520{
		name:  "right1",
		team:  unitsSameTeamTestRecord4EC520{id: 2},
		owner: right2,
	}
	var events []string
	leftTeamLoads := 0
	comparisons := 0
	got := unitsHaveSameTeam4EC520(left, right1, unitsHaveSameTeamHooks4EC520[
		*unitsSameTeamAddressObject4EC520,
		*unitsSameTeamTestRecord4EC520,
	]{
		team: func(obj *unitsSameTeamAddressObject4EC520) *unitsSameTeamTestRecord4EC520 {
			events = append(events, "team:"+obj.name)
			if obj == left {
				leftTeamLoads++
			}
			return &obj.team
		},
		owner: func(obj *unitsSameTeamAddressObject4EC520) *unitsSameTeamAddressObject4EC520 {
			events = append(events, "owner:"+obj.name)
			return obj.owner
		},
		teamEqual: func(first, second *unitsSameTeamTestRecord4EC520) int32 {
			comparisons++
			events = append(events, fmt.Sprintf("compare:%d%d", first.id, second.id))
			if comparisons == 1 {
				left.team.id = 3
				return 0
			}
			if first.id != 0 && second.id != 0 && first.id == second.id {
				return 1
			}
			return 0
		},
	})
	if got != 1 || comparisons != 2 || leftTeamLoads != 1 {
		t.Fatalf("result/comparisons/left team loads = (%d, %d, %d), want (1, 2, 1)", got, comparisons, leftTeamLoads)
	}
	want := []string{
		"team:left", "team:right1", "compare:12", "owner:right1",
		"team:right2", "compare:33",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
}

func TestUnitsHaveSameTeam4EC520DoesNotDetectUnmatchedOwnerCycle(t *testing.T) {
	left := &unitsSameTeamTestObject4EC520{name: "left", team: 1}
	right := &unitsSameTeamTestObject4EC520{name: "right", team: 2}
	left.owner = left
	stop := &struct{}{}
	leftOwnerReads := 0
	var recovered any
	returned := false
	func() {
		defer func() { recovered = recover() }()
		_ = unitsHaveSameTeam4EC520(left, right, unitsHaveSameTeamHooks4EC520[
			*unitsSameTeamTestObject4EC520,
			int,
		]{
			team: func(obj *unitsSameTeamTestObject4EC520) int { return obj.team },
			owner: func(obj *unitsSameTeamTestObject4EC520) *unitsSameTeamTestObject4EC520 {
				if obj == right {
					return nil
				}
				leftOwnerReads++
				if leftOwnerReads == 4 {
					panic(stop)
				}
				return obj.owner
			},
			teamEqual: func(int, int) int32 { return 0 },
		})
		returned = true
	}()
	if returned {
		t.Fatal("unmatched owner cycle unexpectedly terminated")
	}
	if recovered != stop {
		t.Fatalf("cycle recovered %#v, want sentinel", recovered)
	}
}
