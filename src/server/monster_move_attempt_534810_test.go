package server

import (
	"math"
	"testing"
)

func TestMonsterMoveAttemptRecent534810(t *testing.T) {
	s := new(Server)
	s.SetTickRate(30)
	unit := passiveMonsterTestObject547210(t)
	update := unit.UpdateDataMonster()

	s.SetFrame(89)
	if !s.MonsterMoveAttemptRecent534810(unit) {
		t.Fatal("89-frame move attempt was not recent")
	}
	s.SetFrame(90)
	if s.MonsterMoveAttemptRecent534810(unit) {
		t.Fatal("90-frame move attempt remained recent")
	}

	update.Field127 = math.MaxUint32 - 2
	s.SetFrame(5)
	if !s.MonsterMoveAttemptRecent534810(unit) {
		t.Fatal("wrapped seven-frame move attempt was not recent")
	}
}
