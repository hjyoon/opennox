package server

const teamMaterialNoTeam4ECB20 = "NO TEAM"

type teamMaterialEntry4ECB20 struct {
	name string
	team uint32
}

// teamMaterialTable4ECB20 is the portable form of the ten eight-byte rows at
// GAME.EXE 005B91A8. The final row is the original nil-name sentinel.
var teamMaterialTable4ECB20 = [...]teamMaterialEntry4ECB20{
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

// teamMaterialNameHooks4ECB20 separates the value compared in each row from
// the name returned by that row. Both were raw 32-bit words in GAME.EXE, but
// keeping their semantic types distinct avoids encoding a pointer as an
// integer in native-width builds.
type teamMaterialNameHooks4ECB20[V, N comparable] struct {
	noTeam   func() N
	loadName func(uint32) N
	loadTeam func(uint32) V
}

// teamMaterialName4ECB20 preserves the exact row scan at GAME.EXE 004ECB20.
// A zero team returns the separate NO TEAM value before the table is touched.
// For a nonzero team, row zero's name is loaded and checked before any team
// value. Each mismatch loads and tests the next row's name before that row's
// team is read. A match reloads the current name instead of reusing an earlier
// name load, so writable legacy-table mutations remain observable.
//
// GAME.EXE and the retained OpenNox history contain no call, jump, or stored
// function-pointer reference to this routine, so no public or CGo edge is
// invented here.
func teamMaterialName4ECB20[V, N comparable](team V, hooks teamMaterialNameHooks4ECB20[V, N]) N {
	var zeroValue V
	if team == zeroValue {
		return hooks.noTeam()
	}

	name := hooks.loadName(0)
	var zeroName N
	if name == zeroName {
		return zeroName
	}

	var index uint32
	for {
		if hooks.loadTeam(index) == team {
			return hooks.loadName(index)
		}
		name = hooks.loadName(index + 1)
		index++
		if name == zeroName {
			return zeroName
		}
	}
}

func teamMaterialNameValue4ECB20(team uint32) string {
	return teamMaterialName4ECB20(team, teamMaterialNameHooks4ECB20[uint32, string]{
		noTeam: func() string {
			return teamMaterialNoTeam4ECB20
		},
		loadName: func(index uint32) string {
			return teamMaterialTable4ECB20[index].name
		},
		loadTeam: func(index uint32) uint32 {
			return teamMaterialTable4ECB20[index].team
		},
	})
}
