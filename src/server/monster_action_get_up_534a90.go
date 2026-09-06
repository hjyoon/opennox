package server

import "github.com/opennox/libs/object"

// monsterActionGetUp534A90 restores GAME.EXE 00534A90 without passing the
// native-width monster pointer through its original PE32 callback ABI. The
// original byte-sized return value is ignored by the action dispatcher; its
// only observable behavior is the conditional action pop.
func monsterActionGetUp534A90(unit *Object, pop func() int) bool {
	if unit == nil || unit.UpdateData == nil || !unit.Class().Has(object.ClassMonster) || pop == nil {
		return false
	}
	if unit.UpdateDataMonster().Field120_3 != 0 {
		pop()
	}
	return true
}

// MonsterActionGetUp534A90 binds the restored get-up update to the native
// action stack.
func (s *Server) MonsterActionGetUp534A90(unit *Object) bool {
	if unit == nil {
		return false
	}
	return monsterActionGetUp534A90(unit, unit.MonsterPopAction)
}
