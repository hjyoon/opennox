package server

type unitsHaveSameTeamHooks4EC520[O, T comparable] struct {
	team      func(O) T
	owner     func(O) O
	teamEqual func(T, T) int32
}

// unitsHaveSameTeam4EC520 preserves GAME.EXE 004EC520. The original compares
// every object in the first owner chain with every object in the second owner
// chain. The first object's embedded team-record address is cached once per
// outer iteration, but comparisons still observe that record's live contents.
// Team comparison precedes pointer identity, the second chain restarts at its
// original object for every first-chain object, and successor owner links are
// read live only after a pair fails both comparison tests.
func unitsHaveSameTeam4EC520[O, T comparable](
	first, second O,
	hooks unitsHaveSameTeamHooks4EC520[O, T],
) int32 {
	var zero O
	if first == zero || second == zero {
		return 0
	}
	for left := first; left != zero; left = hooks.owner(left) {
		leftTeam := hooks.team(left)
		for right := second; right != zero; right = hooks.owner(right) {
			if hooks.teamEqual(leftTeam, hooks.team(right)) != 0 || left == right {
				return 1
			}
		}
	}
	return 0
}
