package server

// UnitsHaveSameTeam4EC520 binds the original nested owner-chain traversal to
// native-width Object pointers and the embedded fixed-width team records.
func UnitsHaveSameTeam4EC520(first, second *Object) bool {
	return unitsHaveSameTeam4EC520(first, second, unitsHaveSameTeamHooks4EC520[
		*Object,
		*ObjectTeam,
	]{
		team: func(obj *Object) *ObjectTeam {
			return &obj.TeamVal
		},
		owner: func(obj *Object) *Object {
			return obj.ObjOwner
		},
		teamEqual: func(first, second *ObjectTeam) int32 {
			if first.SameAs(second) {
				return 1
			}
			return 0
		},
	}) != 0
}
