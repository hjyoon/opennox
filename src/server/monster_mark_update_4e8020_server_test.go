package server

import "testing"

func TestMonsterMarkUpdateNative4E8020UsesNativeFields(t *testing.T) {
	obj := &Object{Field34: 0x11223344, Field35: 0x11, Field36: 0x1, Field37: 0x55667788}
	hostileUnit := &Object{TypeInd: 17}
	second := &Player{PlayerInd: 34, PlayerUnit: hostileUnit}
	first := &Player{PlayerInd: 32}
	monsterMarkUpdateNative4E8020(obj, monsterMarkUpdateNativeDeps4E8020{
		firstPlayer: func() *Player { return first },
		nextPlayer: func(player *Player) *Player {
			switch player {
			case first:
				return second
			case second:
				return nil
			default:
				t.Fatalf("unexpected player %p", player)
				return nil
			}
		},
		isHostile: func(unit, marked *Object) int32 {
			if unit != hostileUnit || marked != obj {
				t.Fatalf("hostility args = (%p, %p), want (%p, %p)", unit, marked, hostileUnit, obj)
			}
			return 1
		},
	})
	if obj.Field35 != 0x14 || obj.Field36 != 0x4 {
		t.Fatalf("masks = (%#x, %#x), want (0x14, 0x4)", obj.Field35, obj.Field36)
	}
	if obj.Field34 != 0x11223344 || obj.Field37 != 0x55667788 || first.PlayerInd != 32 || second.PlayerInd != 34 {
		t.Fatal("native adapter changed an unrelated field")
	}
}

func TestServerMonsterMarkUpdate4E8020EmptyListDoesNotTouchNilObject(t *testing.T) {
	s := &Server{}
	s.Nox_xxx_monsterMarkUpdate_4E8020(nil)
}

func TestServerMonsterMarkUpdate4E8020UsesActualPlayerList(t *testing.T) {
	s := &Server{}
	s.Players.list = []Player{{Active: 1, PlayerInd: 63}}
	obj := &Object{Field35: 0xffffffff, Field36: 0xffffffff}
	s.Nox_xxx_monsterMarkUpdate_4E8020(obj)
	if obj.Field35 != 0x7fffffff || obj.Field36 != 0x7fffffff {
		t.Fatalf("masks = (%#x, %#x), want both 0x7fffffff", obj.Field35, obj.Field36)
	}
}
