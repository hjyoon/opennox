package server

import "testing"

func TestMonsterSightNativeList528560(t *testing.T) {
	a := &Object{NetCode: 10}
	b := &Object{NetCode: 20}
	c := &Object{NetCode: 30}
	ud := &MonsterUpdateData{}
	for _, obj := range []*Object{a, b, c} {
		if !ud.MonsterAppendSeenEnemy5287B0(obj) {
			t.Fatalf("append %d failed", obj.NetCode)
		}
	}
	ud.CurrentEnemy = b

	if !ud.MonsterHasSeenEnemy528950(b) || ud.MonsterSeenEnemyIndex528950(b) != 1 {
		t.Fatal("middle target was not found")
	}
	removed, ok := ud.MonsterRemoveSeenEnemyAt528560(1)
	if !ok || removed != b {
		t.Fatalf("removed = (%p, %v), want (%p, true)", removed, ok, b)
	}
	if ud.Field282_1 != 2 || ud.SeenEnemies[0] != a || ud.SeenEnemies[1] != c || ud.SeenEnemies[2] != nil {
		t.Fatalf("compacted list = count %d, %p, %p, %p", ud.Field282_1, ud.SeenEnemies[0], ud.SeenEnemies[1], ud.SeenEnemies[2])
	}
	if ud.CurrentEnemy != nil || ud.Field300 != b.NetCode {
		t.Fatalf("focus bookkeeping = (%p, %d), want (nil, %d)", ud.CurrentEnemy, ud.Field300, b.NetCode)
	}

	ud.PreferredEnemy = c
	ud.MonsterClearSeenEnemies528560()
	if ud.Field282_1 != 0 || ud.CurrentEnemy != nil || ud.SeenEnemies[0] != nil || ud.PreferredEnemy != c {
		t.Fatal("clear changed fields outside the sight list/current focus")
	}
}

func TestMonsterSightNativeListBounds528560(t *testing.T) {
	ud := &MonsterUpdateData{Field282_1: 255}
	if got := ud.MonsterSeenEnemyCount528560(); got != monsterSeenEnemyCapacity528560 {
		t.Fatalf("bounded count = %d, want %d", got, monsterSeenEnemyCapacity528560)
	}
	if ud.MonsterAppendSeenEnemy5287B0(&Object{}) {
		t.Fatal("append succeeded for a full/corrupt list")
	}
}
