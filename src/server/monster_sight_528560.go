package server

const monsterSeenEnemyCapacity528560 = 16

// MonsterSeenEnemyCount528560 returns the bounded number of valid entries in
// the original GAME.EXE 16-element sight list. Keeping the byte count preserves
// the PE32 layout while SeenEnemies itself uses native-width pointers.
func (ud *MonsterUpdateData) MonsterSeenEnemyCount528560() int {
	if ud == nil {
		return 0
	}
	n := int(ud.Field282_1)
	if n > len(ud.SeenEnemies) {
		return len(ud.SeenEnemies)
	}
	return n
}

func (ud *MonsterUpdateData) MonsterSeenEnemyIndex528950(target *Object) int {
	if ud == nil || target == nil {
		return -1
	}
	for i := 0; i < ud.MonsterSeenEnemyCount528560(); i++ {
		if ud.SeenEnemies[i] == target {
			return i
		}
	}
	return -1
}

func (ud *MonsterUpdateData) MonsterHasSeenEnemy528950(target *Object) bool {
	return ud.MonsterSeenEnemyIndex528950(target) >= 0
}

// MonsterRemoveSeenEnemyAt528560 mirrors GAME.EXE 0x528560's list compaction
// and current-enemy bookkeeping. Script dispatch remains at the legacy bridge
// because it is an engine side effect rather than list state.
func (ud *MonsterUpdateData) MonsterRemoveSeenEnemyAt528560(index int) (*Object, bool) {
	if ud == nil {
		return nil, false
	}
	n := ud.MonsterSeenEnemyCount528560()
	if index < 0 || index >= n {
		return nil, false
	}
	removed := ud.SeenEnemies[index]
	if ud.CurrentEnemy == removed {
		ud.CurrentEnemy = nil
		if removed != nil {
			ud.Field300 = removed.NetCode
		}
	}
	copy(ud.SeenEnemies[index:n-1], ud.SeenEnemies[index+1:n])
	ud.SeenEnemies[n-1] = nil
	ud.Field282_1 = uint8(n - 1)
	return removed, true
}

func (ud *MonsterUpdateData) MonsterAppendSeenEnemy5287B0(target *Object) bool {
	if ud == nil || target == nil {
		return false
	}
	n := ud.MonsterSeenEnemyCount528560()
	if n >= monsterSeenEnemyCapacity528560 {
		return false
	}
	ud.SeenEnemies[n] = target
	ud.Field282_1 = uint8(n + 1)
	return true
}

func (ud *MonsterUpdateData) MonsterClearSeenEnemies528560() {
	if ud == nil {
		return
	}
	for i := range ud.SeenEnemies {
		ud.SeenEnemies[i] = nil
	}
	ud.Field282_1 = 0
	ud.CurrentEnemy = nil
}
