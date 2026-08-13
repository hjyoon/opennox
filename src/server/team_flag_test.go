package server

import "testing"

func TestServerTeamsTeamFlagNativePointer(t *testing.T) {
	teams := serverTeams{Arr: make([]Team, 3)}
	team := &teams.Arr[1]
	flag := &Object{}

	teams.SetTeamFlag(team, flag)
	if got := teams.TeamFlag(team); got != flag {
		t.Fatalf("team flag = %p, want %p", got, flag)
	}
	teams.SetTeamFlag(team, nil)
	if got := teams.TeamFlag(team); got != nil {
		t.Fatalf("cleared team flag = %p, want nil", got)
	}
}

func TestServerTeamsResetClearsNativeTeamFlags(t *testing.T) {
	teams := serverTeams{Arr: make([]Team, 3)}
	team := &teams.Arr[1]
	teams.SetTeamFlag(team, &Object{})
	teams.Reset()
	if got := teams.TeamFlag(team); got != nil {
		t.Fatalf("reset team flag = %p, want nil", got)
	}
}
