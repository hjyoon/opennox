package server

const teamDefaultNameMax4ECAA0 = int8(16)

type teamDefaultNameEntry4ECAA0 struct {
	text    string
	present bool
}

// teamDefaultNameTable4ECAA0 is the portable form of the 17 pointer cells at
// GAME.EXE 005B91F8. The final cell is deliberately nil in the original.
var teamDefaultNameTable4ECAA0 = [...]teamDefaultNameEntry4ECAA0{
	{text: "NONE", present: true},
	{text: "Team 1", present: true},
	{text: "Team 2", present: true},
	{text: "Team 3", present: true},
	{text: "Team 4", present: true},
	{text: "Team 5", present: true},
	{text: "Team 6", present: true},
	{text: "Team 7", present: true},
	{text: "Team 8", present: true},
	{text: "Team 9", present: true},
	{text: "Team 10", present: true},
	{text: "Team 11", present: true},
	{text: "Team 12", present: true},
	{text: "Team 13", present: true},
	{text: "Team 14", present: true},
	{text: "Team 15", present: true},
	{},
}

type teamDefaultNameHooks4ECAA0[T any] struct {
	load func(int8) T
}

// teamDefaultName4ECAA0 preserves the exact signed-char selection performed
// by GAME.EXE 004ECAA0. Only positive values 17..127 are replaced with zero.
// Zero through 16 and negative values reach the single table load unchanged;
// the loader keeps the original adjacent-memory behavior observable without
// embedding an unsafe 32-bit address table in a native-width runtime.
//
// GAME.EXE and the retained OpenNox history contain no call, jump, or stored
// function-pointer reference to this routine, so no public or CGo edge is
// invented here.
func teamDefaultName4ECAA0[T any](index int8, hooks teamDefaultNameHooks4ECAA0[T]) T {
	if index > teamDefaultNameMax4ECAA0 {
		index = 0
	}
	return hooks.load(index)
}
