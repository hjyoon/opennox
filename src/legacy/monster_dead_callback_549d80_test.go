package legacy

import (
	"testing"

	"github.com/opennox/opennox/v1/server"
)

func TestMonsterDeadCallbackKind549D80CoversMonsterBinTable(t *testing.T) {
	tests := []struct {
		name string
		kind server.MonsterDeadCallbackKind549D80
	}{
		{name: "EMBERDEMONDEAD", kind: server.MonsterDeadCallbackEmberDemon549D80},
		{name: "DEMONDEAD", kind: server.MonsterDeadCallbackDemon549E00},
		{name: "IMPDEAD", kind: server.MonsterDeadCallbackImp549E70},
		{name: "MECHGOLEMDEAD", kind: server.MonsterDeadCallbackMechGolem549E90},
		{name: "GOLEMDEAD", kind: server.MonsterDeadCallbackGolem549FA0},
		{name: "BOMBERDEAD", kind: server.MonsterDeadCallbackBomber54A150},
		{name: "SPIDERDEAD", kind: server.MonsterDeadCallbackSpider54A250},
		{name: "TROLLDEAD", kind: server.MonsterDeadCallbackTroll54A270},
		{name: "SKELETONDEAD", kind: server.MonsterDeadCallbackSkeleton54A310},
		{name: "SKELETONLORDDEAD", kind: server.MonsterDeadCallbackSkeletonLord54A750},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fnc := monsterDeadFunctions[test.name]
			if fnc == nil {
				t.Fatal("monster.bin DEAD_FUNCTION has no C function pointer")
			}
			got, ok := monsterDeadCallbackKind549D80(fnc)
			if !ok || got != test.kind {
				t.Fatalf("callback kind = %d/%t, want %d/true", got, ok, test.kind)
			}
		})
	}
	if _, ok := monsterDeadCallbackKind549D80(nil); ok {
		t.Fatal("nil DEAD_FUNCTION was accepted")
	}
}
