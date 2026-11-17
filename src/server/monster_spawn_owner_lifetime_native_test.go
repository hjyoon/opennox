package server

import (
	"runtime"
	"testing"

	"github.com/opennox/libs/strman"
)

func TestMonsterSpawnLinkPinLifetime(t *testing.T) {
	s := NewHigh(nil, nil, strman.New())
	defer s.Close()
	s.MonsterSpawnInit50D780()
	generator, _ := monsterSpawnTestGenerator50D780(0)
	var nodes [3]*Object
	for i := range nodes {
		nodes[i], _ = monsterSpawnTestMonster50D780(0)
		if !s.MonsterSpawnRegister50E030(generator, nodes[i], MonsterSpawnRegisterRuntime50E030{}) {
			t.Fatal("registration failed", i)
		}
	}
	if got := len(s.monsterSpawns50D780.pins); got != 3 {
		t.Fatal("active pin owners", got)
	}
	runtime.GC()
	s.MonsterSpawnCleanup50E140(nodes[1])
	if got := len(s.monsterSpawns50D780.pins); got != 2 {
		t.Fatal("detached pin remains owned", got)
	}
	s.MonsterSpawnReset50D7E0()
	if got := len(s.monsterSpawns50D780.pins); got != 0 {
		t.Fatal("reset retained pins", got)
	}
	if !s.MonsterSpawnRegister50E030(generator, nodes[0], MonsterSpawnRegisterRuntime50E030{}) {
		t.Fatal("registration after reset failed")
	}
	s.Close()
	if got := len(s.monsterSpawns50D780.pins); got != 0 || s.monsterSpawns50D780.initialized {
		t.Fatal("close retained pin/allocator ownership", got)
	}
	runtime.GC()
}
