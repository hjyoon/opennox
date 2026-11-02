package legacy

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/server"
)

type retreatPredicateRow545520 struct {
	Input struct {
		Name, Entry           string
		Cur                   uint16
		Maximum               uint16
		Resume, Status, Buffs uint32
		NilUnit               bool `json:"nil_unit"`
		NilHealth             bool `json:"nil_health"`
		NilUpdate             bool `json:"nil_update"`
		Alias                 bool
	}
	Output struct {
		Result *uint32
		Fault  []json.RawMessage
	}
}

func retreatPredicateRows545520(t *testing.T) []retreatPredicateRow545520 {
	t.Helper()
	raw, err := os.ReadFile("../server/testdata/monster_retreat_predicates_original_545520.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(raw)) != "3c8ca86a5bb03ca5ac48a495c98b534fa32cf2d2670db84e1ce87e58c1e175d7" {
		t.Fatal("frozen original gzip changed")
	}
	z, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := io.ReadAll(z)
	if err != nil {
		t.Fatal(err)
	}
	if err = z.Close(); err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(decoded)) != "8d721b4b0a416310851bc2feae147e8a28ed07eba1aebe02b7e6224bc3788db7" {
		t.Fatal("frozen original decoded bytes changed")
	}
	var data struct{ Rows []retreatPredicateRow545520 }
	if err = json.Unmarshal(decoded, &data); err != nil {
		t.Fatal(err)
	}
	if len(data.Rows) != 2073 {
		t.Fatal("original context count", len(data.Rows))
	}
	return data.Rows
}

type retreatPredicateNative545520 interface {
	MonsterCanResumeAttack545520() bool
	MonsterRetreatActive545580() bool
	MonsterCanCast2_5455B0() bool
}

func TestMonsterRetreatPredicatesOriginal545520(t *testing.T) {
	for _, path := range []string{"C", "Go"} {
		t.Run(path, func(t *testing.T) {
			for _, row := range retreatPredicateRows545520(t) {
				t.Run(row.Input.Name, func(t *testing.T) {
					entry := map[string]int{"545520": 0, "545580": 1, "5455B0": 2}[row.Input.Entry]
					if path == "C" && retreatPredicatesOriginalABI545520(entry) != 64 {
						t.Fatal("native C entry pending; actual legacy-int faults qualified independently")
					}
					w := newDodgeWorld547C50(t, dodgeRow547C50{})
					w.Unit.HealthData = (*server.HealthData)(unsafe.Pointer(&w.Heap[24000]))
					if row.Input.Alias {
						w.Unit.HealthData = (*server.HealthData)(unsafe.Pointer(w.Unit))
					}
					w.Unit.HealthData.Cur, w.Unit.HealthData.Max = row.Input.Cur, row.Input.Maximum
					w.Data.ResumeLevel = math.Float32frombits(row.Input.Resume)
					w.Data.StatusFlags = object.MonsterStatus(row.Input.Status)
					w.Unit.Buffs = row.Input.Buffs
					if row.Input.NilHealth {
						w.Unit.HealthData = nil
					}
					if row.Input.NilUpdate {
						w.Unit.UpdateData = nil
					}
					unit := w.Unit
					if row.Input.NilUnit {
						unit = nil
					}
					before := bytes.Clone(w.Heap)
					var fault any
					result := 0
					func() {
						defer func() { fault = recover() }()
						if path == "C" {
							result = retreatPredicatesOriginalC545520(entry, unit)
							return
						}
						native, ok := any(unit).(retreatPredicateNative545520)
						if !ok {
							t.Fatal("native direct-Go methods pending")
						}
						var got bool
						switch entry {
						case 0:
							got = native.MonsterCanResumeAttack545520()
						case 1:
							got = native.MonsterRetreatActive545580()
						case 2:
							got = native.MonsterCanCast2_5455B0()
						}
						if got {
							result = 1
						}
					}()
					if !bytes.Equal(before, w.Heap) {
						t.Fatal("whole native object/health/MUD/alias allocation and canaries changed")
					}
					if (fault != nil) != (len(row.Output.Fault) != 0) {
						t.Fatal("original raw fault boundary", fault, row.Output.Fault)
					}
					if fault == nil && (row.Output.Result == nil || uint32(result) != *row.Output.Result) {
						t.Fatalf("original result got=%d want=%v", result, row.Output.Result)
					}
				})
			}
		})
	}
}
