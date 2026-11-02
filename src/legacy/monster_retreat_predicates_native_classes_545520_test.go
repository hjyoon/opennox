package legacy

import (
	"bytes"
	"fmt"
	"math"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/server"
)

// These raw entries have no class admission in GAME.EXE. Exercise the
// original golden's low/high-health gate rows with native C-owned records
// carrying unrelated classes, to reject accidental UpdateDataMonster gates.
func TestMonsterRetreatPredicatesNativeRawClasses545520(t *testing.T) {
	selected := 0
	for _, row := range retreatPredicateRows545520(t) {
		if !strings.Contains(row.Input.Name, "/gates/00000020/20000000/") {
			continue
		}
		selected++
		for _, class := range []object.Class{0, object.ClassPlayer, object.ClassMonster, object.ClassImmobile, 0xffffffff} {
			t.Run(fmt.Sprintf("%s/class-%08x", row.Input.Name, uint32(class)), func(t *testing.T) {
				w := newDodgeWorld547C50(t, dodgeRow547C50{})
				w.Unit.HealthData = (*server.HealthData)(unsafe.Pointer(&w.Heap[24000]))
				w.Unit.HealthData.Cur, w.Unit.HealthData.Max = row.Input.Cur, row.Input.Maximum
				w.Unit.ObjClass, w.Unit.Buffs = class, row.Input.Buffs
				w.Data.ResumeLevel = math.Float32frombits(row.Input.Resume)
				w.Data.StatusFlags = object.MonsterStatus(row.Input.Status)
				entry := map[string]int{"545520": 0, "545580": 1, "5455B0": 2}[row.Input.Entry]
				before := bytes.Clone(w.Heap)
				if got := retreatPredicatesOriginalC545520(entry, w.Unit); row.Output.Result == nil || got != int(*row.Output.Result) {
					t.Fatal("raw public C entry adds a class gate or changes original result", got, row.Output.Result)
				}
				if !bytes.Equal(before, w.Heap) {
					t.Fatal("whole native allocation changed")
				}
			})
		}
	}
	if selected != 6 {
		t.Fatal("original low/high-health gate row count", selected)
	}
}
