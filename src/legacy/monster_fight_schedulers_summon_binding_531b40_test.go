package legacy

import (
	"testing"

	"github.com/opennox/opennox/v1/server"
)

func TestFightSchedulersNativeSummonLimitBinding531B40(t *testing.T) {
	f := schedNewFixture531B40(t)
	unit := &f.units[0].Value
	old := Nox_xxx_checkSummonedCreaturesLimit_500D70
	t.Cleanup(func() { Nox_xxx_checkSummonedCreaturesLimit_500D70 = old })
	for _, tc := range []struct {
		name  string
		index int
		allow bool
	}{{"negative", -1, false}, {"allowed", 3, true}, {"rejected", 42, false}} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			Nox_xxx_checkSummonedCreaturesLimit_500D70 = func(got *server.Object, index int32) bool {
				calls++
				if got != unit || index != int32(tc.index) {
					t.Fatalf("native summon limit args=%p/%d", got, index)
				}
				return tc.allow
			}
			if got := monsterFightSchedulerRuntime531B40().CanSummon(unit, tc.index); got != tc.allow || calls != 1 {
				t.Fatalf("native summon limit result/calls=%t/%d", got, calls)
			}
		})
	}
}
