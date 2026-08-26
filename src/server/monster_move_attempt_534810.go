package server

// MonsterMoveAttemptRecent534810 is GAME.EXE 00534810. The unsigned frame
// subtraction intentionally preserves the original wraparound behavior.
func (s *Server) MonsterMoveAttemptRecent534810(unit *Object) bool {
	if unit == nil || unit.UpdateData == nil {
		return false
	}
	return s.Frame()-unit.UpdateDataMonster().Field127 < 3*s.TickRate()
}
