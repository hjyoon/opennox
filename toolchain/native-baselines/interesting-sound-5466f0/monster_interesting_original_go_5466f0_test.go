package legacy

import (
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
	"testing"
)

// A dynamic interface permits compiling the unchanged-C baseline before
// the native Go method exists. This comparison uses the same frozen heap,
// services and original rows; no expected value comes from the new method.
type interestingNativeEntry5466F0 interface {
	MonsterInterestingSoundNative5466F0(server.MonsterReactiveRuntime545DA0,
		func(*types.Pointf) int, func(*server.Object, *types.Pointf) bool,
		func(types.Pointf, types.Pointf) bool) int
}

func TestMonsterInterestingOriginalGo5466F0(t *testing.T) {
	for _, row := range interestingLoad5466F0(t) {
		t.Run(row.Name, func(t *testing.T) {
			w := interestingWorldNew5466F0(t, row)
			unit := w.Unit
			if row.Input.UnitNil {
				unit = nil
			}
			entry, ok := any(unit).(interestingNativeEntry5466F0)
			if !ok {
				t.Fatal("native Go entry not implemented")
			}
			call := func() uint32 {
				return uint32(entry.MonsterInterestingSoundNative5466F0(w.Runtime.Reactive, w.Runtime.Tile, w.Runtime.Path, w.Runtime.Trace))
			}
			if row.Fault != nil {
				panicked := false
				func() { defer func() { panicked = recover() != nil }(); call() }()
				if !panicked {
					t.Fatal("original raw nil fault must remain")
				}
			} else if result := call(); result != row.Result {
				t.Fatal("original DWORD result", result, row.Result)
			}
			w.check(t, row)
		})
	}
}
