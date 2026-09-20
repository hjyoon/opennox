package server

import (
	"reflect"
	"testing"
)

type kotrMapSetupTestObject4180D0 struct {
	name   string
	flag   bool
	crown  bool
	teamID TeamID
	next   *kotrMapSetupTestObject4180D0
}

type kotrMapSetupTestTeam4180D0 struct {
	name string
	id   TeamID
	next *kotrMapSetupTestTeam4180D0
}

type kotrMapSetupTestWorld4180D0 struct {
	objects []*kotrMapSetupTestObject4180D0
	teams   []*kotrMapSetupTestTeam4180D0
	events  []string
}

func (w *kotrMapSetupTestWorld4180D0) hooks(teamMode bool) kotrMapSetupHooks4180D0[
	*kotrMapSetupTestObject4180D0,
	*kotrMapSetupTestTeam4180D0,
] {
	for i := range w.objects {
		w.objects[i].next = nil
		if i+1 < len(w.objects) {
			w.objects[i].next = w.objects[i+1]
		}
	}
	for i := range w.teams {
		w.teams[i].next = nil
		if i+1 < len(w.teams) {
			w.teams[i].next = w.teams[i+1]
		}
	}
	return kotrMapSetupHooks4180D0[*kotrMapSetupTestObject4180D0, *kotrMapSetupTestTeam4180D0]{
		firstObject: func() *kotrMapSetupTestObject4180D0 {
			if len(w.objects) == 0 {
				return nil
			}
			return w.objects[0]
		},
		nextObject: func(obj *kotrMapSetupTestObject4180D0) *kotrMapSetupTestObject4180D0 {
			return obj.next
		},
		isFlag: func(obj *kotrMapSetupTestObject4180D0) bool {
			return obj.flag
		},
		delayedDelete: func(obj *kotrMapSetupTestObject4180D0) {
			w.events = append(w.events, "delete:"+obj.name)
		},
		firstTeam: func() *kotrMapSetupTestTeam4180D0 {
			if len(w.teams) == 0 {
				return nil
			}
			return w.teams[0]
		},
		nextTeam: func(team *kotrMapSetupTestTeam4180D0) *kotrMapSetupTestTeam4180D0 {
			return team.next
		},
		clearTeamFlag: func(team *kotrMapSetupTestTeam4180D0) {
			w.events = append(w.events, "clear-team:"+team.name)
		},
		isCrown: func(obj *kotrMapSetupTestObject4180D0) bool {
			return obj.crown
		},
		clearPickupTarget: func(obj *kotrMapSetupTestObject4180D0) {
			w.events = append(w.events, "clear-target:"+obj.name)
		},
		teamMode: teamMode,
		teamID: func(obj *kotrMapSetupTestObject4180D0) TeamID {
			return obj.teamID
		},
		respawnRemove: func(obj *kotrMapSetupTestObject4180D0) {
			w.events = append(w.events, "remove-respawn:"+obj.name)
		},
		markMinimapForAll: func(obj *kotrMapSetupTestObject4180D0) {
			w.events = append(w.events, "mark:"+obj.name)
		},
		teamByID: func(id TeamID) *kotrMapSetupTestTeam4180D0 {
			for _, team := range w.teams {
				if team.id == id {
					return team
				}
			}
			return nil
		},
		setTeamFlag: func(team *kotrMapSetupTestTeam4180D0, obj *kotrMapSetupTestObject4180D0) {
			w.events = append(w.events, "set-team:"+team.name+"="+obj.name)
		},
	}
}

func TestMapInfoSetKotr4180D0FreeForAllOrder(t *testing.T) {
	flag := &kotrMapSetupTestObject4180D0{name: "flag", flag: true}
	neutral := &kotrMapSetupTestObject4180D0{name: "neutral", crown: true}
	teamed := &kotrMapSetupTestObject4180D0{name: "teamed", crown: true, teamID: 2}
	world := &kotrMapSetupTestWorld4180D0{
		objects: []*kotrMapSetupTestObject4180D0{flag, neutral, teamed},
		teams: []*kotrMapSetupTestTeam4180D0{
			{name: "red", id: 1},
			{name: "blue", id: 2},
		},
	}
	if !mapInfoSetKotr4180D0(world.hooks(false)) {
		t.Fatal("KOTR setup reported no crowns")
	}
	want := []string{
		"delete:flag",
		"clear-team:red", "clear-team:blue",
		"clear-target:neutral", "mark:neutral",
		"clear-target:teamed", "delete:teamed", "remove-respawn:teamed",
	}
	if !reflect.DeepEqual(world.events, want) {
		t.Fatalf("events = %v, want %v", world.events, want)
	}
}

func TestMapInfoSetKotr4180D0TeamModeBranches(t *testing.T) {
	neutral := &kotrMapSetupTestObject4180D0{name: "neutral", crown: true}
	known := &kotrMapSetupTestObject4180D0{name: "known", crown: true, teamID: 2}
	unknown := &kotrMapSetupTestObject4180D0{name: "unknown", crown: true, teamID: 9}
	world := &kotrMapSetupTestWorld4180D0{
		objects: []*kotrMapSetupTestObject4180D0{neutral, known, unknown},
		teams: []*kotrMapSetupTestTeam4180D0{
			{name: "blue", id: 2},
		},
	}
	if !mapInfoSetKotr4180D0(world.hooks(true)) {
		t.Fatal("KOTR setup reported no crowns")
	}
	want := []string{
		"clear-team:blue",
		"clear-target:neutral", "delete:neutral", "remove-respawn:neutral",
		"clear-target:known", "set-team:blue=known", "mark:known",
		"clear-target:unknown",
	}
	if !reflect.DeepEqual(world.events, want) {
		t.Fatalf("events = %v, want %v", world.events, want)
	}
}

func TestMapInfoSetKotr4180D0NoCrownsStillCleansFlagsAndTeams(t *testing.T) {
	world := &kotrMapSetupTestWorld4180D0{
		objects: []*kotrMapSetupTestObject4180D0{{name: "flag", flag: true}},
		teams:   []*kotrMapSetupTestTeam4180D0{{name: "red", id: 1}},
	}
	if mapInfoSetKotr4180D0(world.hooks(false)) {
		t.Fatal("KOTR setup reported a crown")
	}
	want := []string{"delete:flag", "clear-team:red"}
	if !reflect.DeepEqual(world.events, want) {
		t.Fatalf("events = %v, want %v", world.events, want)
	}
}
