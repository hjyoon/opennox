package server

import (
	"fmt"
	"reflect"
	"testing"
)

func TestTeamMaterialName4ECB20Table(t *testing.T) {
	want := []teamMaterialEntry4ECB20{
		{name: "MaterialTeamRed", team: 1},
		{name: "MaterialTeamGreen", team: 3},
		{name: "MaterialTeamBlue", team: 2},
		{name: "MaterialTeamYellow", team: 5},
		{name: "MaterialTeamCyan", team: 4},
		{name: "MaterialTeamViolet", team: 6},
		{name: "MaterialTeamBlack", team: 7},
		{name: "MaterialTeamWhite", team: 8},
		{name: "MaterialTeamOrange", team: 9},
		{},
	}
	if !reflect.DeepEqual(teamMaterialTable4ECB20[:], want) {
		t.Fatalf("table = %#v, want %#v", teamMaterialTable4ECB20, want)
	}
}

func TestTeamMaterialName4ECB20PortableValues(t *testing.T) {
	tests := []struct {
		team uint32
		want string
	}{
		{team: 0, want: "NO TEAM"},
		{team: 1, want: "MaterialTeamRed"},
		{team: 2, want: "MaterialTeamBlue"},
		{team: 3, want: "MaterialTeamGreen"},
		{team: 4, want: "MaterialTeamCyan"},
		{team: 5, want: "MaterialTeamYellow"},
		{team: 6, want: "MaterialTeamViolet"},
		{team: 7, want: "MaterialTeamBlack"},
		{team: 8, want: "MaterialTeamWhite"},
		{team: 9, want: "MaterialTeamOrange"},
		{team: 10, want: ""},
		{team: ^uint32(0), want: ""},
	}
	for _, tc := range tests {
		t.Run(fmt.Sprintf("%d", tc.team), func(t *testing.T) {
			if got := teamMaterialNameValue4ECB20(tc.team); got != tc.want {
				t.Fatalf("name = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestTeamMaterialName4ECB20ZeroBypassesTable(t *testing.T) {
	var events []string
	got := teamMaterialName4ECB20(0, teamMaterialNameHooks4ECB20[int, int]{
		noTeam: func() int {
			events = append(events, "no-team")
			return 77
		},
		loadName: func(uint32) int {
			t.Fatal("name table touched for zero team")
			return 0
		},
		loadTeam: func(uint32) int {
			t.Fatal("team table touched for zero team")
			return 0
		},
	})
	if got != 77 {
		t.Fatalf("name = %d, want 77", got)
	}
	if want := []string{"no-team"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
}

func TestTeamMaterialName4ECB20EmptyHeadSkipsTeamLoad(t *testing.T) {
	var events []string
	got := teamMaterialName4ECB20(7, teamMaterialNameHooks4ECB20[int, int]{
		noTeam: func() int { return 99 },
		loadName: func(index uint32) int {
			events = append(events, fmt.Sprintf("name:%d", index))
			return 0
		},
		loadTeam: func(uint32) int {
			t.Fatal("team loaded after nil head")
			return 0
		},
	})
	if got != 0 {
		t.Fatalf("name = %d, want zero", got)
	}
	if want := []string{"name:0"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
}

func TestTeamMaterialName4ECB20ScanOrder(t *testing.T) {
	names := []int{10, 20, 0}
	teams := []int{1, 7, 0}
	var events []string
	got := teamMaterialName4ECB20(7, teamMaterialNameHooks4ECB20[int, int]{
		noTeam: func() int { return 99 },
		loadName: func(index uint32) int {
			events = append(events, fmt.Sprintf("name:%d", index))
			return names[index]
		},
		loadTeam: func(index uint32) int {
			events = append(events, fmt.Sprintf("team:%d", index))
			return teams[index]
		},
	})
	if got != 20 {
		t.Fatalf("name = %d, want 20", got)
	}
	want := []string{"name:0", "team:0", "name:1", "team:1", "name:1"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
}

func TestTeamMaterialName4ECB20MatchReloadsLiveName(t *testing.T) {
	loads := 0
	var events []string
	got := teamMaterialName4ECB20(1, teamMaterialNameHooks4ECB20[int, int]{
		noTeam: func() int { return 99 },
		loadName: func(index uint32) int {
			events = append(events, fmt.Sprintf("name:%d", index))
			loads++
			if loads == 1 {
				return 10
			}
			return 11
		},
		loadTeam: func(index uint32) int {
			events = append(events, fmt.Sprintf("team:%d", index))
			return 1
		},
	})
	if got != 11 {
		t.Fatalf("name = %d, want reloaded value 11", got)
	}
	want := []string{"name:0", "team:0", "name:0"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
}

func TestTeamMaterialName4ECB20GateDoesNotCacheMatchName(t *testing.T) {
	nameOneLoads := 0
	got := teamMaterialName4ECB20(7, teamMaterialNameHooks4ECB20[int, int]{
		noTeam: func() int { return 99 },
		loadName: func(index uint32) int {
			if index == 0 {
				return 10
			}
			nameOneLoads++
			if nameOneLoads == 1 {
				return 20
			}
			return 0
		},
		loadTeam: func(index uint32) int {
			return []int{1, 7}[index]
		},
	})
	if got != 0 {
		t.Fatalf("name = %d, want live reloaded zero", got)
	}
	if nameOneLoads != 2 {
		t.Fatalf("row-one name loads = %d, want 2", nameOneLoads)
	}
}

func TestTeamMaterialName4ECB20SentinelStopsBeforeTeamLoad(t *testing.T) {
	var events []string
	got := teamMaterialName4ECB20(7, teamMaterialNameHooks4ECB20[int, int]{
		noTeam: func() int { return 99 },
		loadName: func(index uint32) int {
			events = append(events, fmt.Sprintf("name:%d", index))
			return []int{10, 0}[index]
		},
		loadTeam: func(index uint32) int {
			events = append(events, fmt.Sprintf("team:%d", index))
			if index == 1 {
				t.Fatal("sentinel team value loaded")
			}
			return 1
		},
	})
	if got != 0 {
		t.Fatalf("name = %d, want zero", got)
	}
	want := []string{"name:0", "team:0", "name:1"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
}

func TestTeamMaterialName4ECB20FaultOrder(t *testing.T) {
	stages := []struct {
		name       string
		faultEvent string
		want       []string
	}{
		{name: "head", faultEvent: "name:0", want: []string{"name:0"}},
		{name: "team", faultEvent: "team:0", want: []string{"name:0", "team:0"}},
		{name: "next-name", faultEvent: "name:1", want: []string{"name:0", "team:0", "name:1"}},
	}
	for _, tc := range stages {
		t.Run(tc.name, func(t *testing.T) {
			stop := &struct{}{}
			var events []string
			var recovered any
			func() {
				defer func() { recovered = recover() }()
				teamMaterialName4ECB20(7, teamMaterialNameHooks4ECB20[int, int]{
					noTeam: func() int { return 99 },
					loadName: func(index uint32) int {
						event := fmt.Sprintf("name:%d", index)
						events = append(events, event)
						if event == tc.faultEvent {
							panic(stop)
						}
						return 10
					},
					loadTeam: func(index uint32) int {
						event := fmt.Sprintf("team:%d", index)
						events = append(events, event)
						if event == tc.faultEvent {
							panic(stop)
						}
						return 1
					},
				})
			}()
			if recovered != stop {
				t.Fatalf("panic = %#v, want sentinel", recovered)
			}
			if !reflect.DeepEqual(events, tc.want) {
				t.Fatalf("events = %#v, want %#v", events, tc.want)
			}
		})
	}
}

func TestTeamMaterialName4ECB20NoTeamFaultPrecedesTable(t *testing.T) {
	stop := &struct{}{}
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		teamMaterialName4ECB20(0, teamMaterialNameHooks4ECB20[int, int]{
			noTeam: func() int { panic(stop) },
			loadName: func(uint32) int {
				t.Fatal("name table touched after NO TEAM fault")
				return 0
			},
			loadTeam: func(uint32) int {
				t.Fatal("team table touched after NO TEAM fault")
				return 0
			},
		})
	}()
	if recovered != stop {
		t.Fatalf("panic = %#v, want sentinel", recovered)
	}
}
