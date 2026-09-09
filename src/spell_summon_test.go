package opennox

import (
	"math"
	"testing"
)

func TestCheckSummonedCreaturesLimit500D70CheatBypassesOracleCalls(t *testing.T) {
	old := cheatSummonNoLimit
	t.Cleanup(func() { cheatSummonNoLimit = old })
	cheatSummonNoLimit = true
	if !nox_xxx_checkSummonedCreaturesLimit_500D70(nil, math.MinInt32) {
		t.Fatal("summon-limit cheat rejected a call that must bypass guide and owner loads")
	}
}
