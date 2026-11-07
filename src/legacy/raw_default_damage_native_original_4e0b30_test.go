//go:build darwin && arm64

package legacy

import (
	"encoding/json"
	"fmt"
	"testing"
	"unsafe"
)

func rddNativeOriginalCorpus4E0B30(t *testing.T, backend string) {
	before := rddNativeAdapterPhaseGate4E0B30(t)
	f, e := ddiLoadFixture4E0B30()
	if e != nil {
		t.Fatal(e)
	}
	if before {
		// A compiled pre-engine root proves actual typed C transport and the
		// engine absence gate. No per-original native leaves are fabricated.
		g, e := ddiBuildGraph4E0B30(f, f.Cases[0])
		if e != nil {
			t.Fatal(e)
		}
		defer g.close()
		roles := f.Cases[0].Input.Roles
		objects := [3]unsafe.Pointer{g.Tokens[roles["target"]].Pointer, g.Tokens[roles["source"]].Pointer, g.Tokens[roles["weapon"]].Pointer}
		for _, p := range objects {
			if uintptr(p) <= 0xffffffff {
				t.Fatal("actual high C-owned transport object")
			}
		}
		got := rawDefaultDamageCTransport4E0B30(objects[0], objects[1], objects[2], -305419896, 2023406814, -123, true, -456)
		want := [5]uintptr{uintptr(objects[0]), uintptr(objects[1]), uintptr(objects[2]), uintptr(uint32(0xedcba988)), uintptr(uint32(2023406814))}
		if got.Outer != want || got.Return != -123 || got.InnerReturn != -456 || !got.Restored {
			t.Fatal("compiled native consumer pre-engine transport")
		}
		if e = g.checkCanaries(); e != nil {
			t.Fatal(e)
		}
		t.Logf("RAW_DEFAULT_DAMAGE_NATIVE_ORIGINAL_BEFORE backend=%s actual_C_transport=true native_engine_leaves=0", backend)
		return
	}
	b, e := rddOpenBundle4E0B30()
	if e != nil {
		t.Fatal(e)
	}
	defer b.close()
	attempted, passed, failed := 0, 0, 0
	seen := map[string]bool{}
	for _, runName := range []string{"baseline1034", "additive4"} {
		e = b.rows(b.Index.Runs[runName], func(raw json.RawMessage) error {
			var row rddRow4E0B30
			if e := json.Unmarshal(raw, &row); e != nil {
				return e
			}
			rddBindNativeRowSource4E0B30(&row, raw)
			name := row.Mode + "/" + row.Name
			if seen[name] {
				return fmt.Errorf("duplicate actual native row")
			}
			seen[name] = true
			attempted++
			if t.Run(name, func(t *testing.T) {
				if e := rddRunSupervisedNativeRow4E0B30(t, b, &row, f, backend); e != nil {
					t.Fatal(e)
				}
			}) {
				passed++
			} else {
				failed++
			}
			return nil
		})
		if e != nil {
			t.Logf("RAW_DEFAULT_DAMAGE_NATIVE_ORIGINAL_SUMMARY backend=%s attempted_original_rows=%d passed_original_rows=%d failed_original_rows=%d stream_complete=false", backend, attempted, passed, failed)
			t.Fatal(e)
		}
	}
	t.Logf("RAW_DEFAULT_DAMAGE_NATIVE_ORIGINAL_SUMMARY backend=%s attempted_original_rows=%d passed_original_rows=%d failed_original_rows=%d stream_complete=true", backend, attempted, passed, failed)
	if attempted != 2076 {
		t.Fatalf("actual native %s root leaves: attempted=%d passed=%d failed=%d, want attempted=2076", backend, attempted, passed, failed)
	}
	if failed != 0 {
		t.Fatalf("actual native %s root failures: attempted=%d passed=%d failed=%d", backend, attempted, passed, failed)
	}
	t.Logf("RAW_DEFAULT_DAMAGE_NATIVE_ORIGINAL_OK backend=%s actual_original_rows=2076 actual_native_root_leaves=2076", backend)
}
func TestRawDefaultDamageNativeGoOriginal4E0B30(t *testing.T) { rddNativeOriginalCorpus4E0B30(t, "Go") }
func TestRawDefaultDamageNativeCOriginal4E0B30(t *testing.T)  { rddNativeOriginalCorpus4E0B30(t, "C") }

func TestRawDefaultDamageNativePlans4E0B30(t *testing.T) {
	rddNativeAdapterPhaseGate4E0B30(t)
	fixture, e := ddiLoadFixture4E0B30()
	if e != nil {
		t.Fatal(e)
	}
	index := rddFixtureIndex4E0B30(fixture)
	bundle, e := rddOpenBundle4E0B30()
	if e != nil {
		t.Fatal(e)
	}
	defer bundle.close()
	rows, originalReads, derived, readNoHook, readHook, writeHook, fetch, returned := 0, 0, 0, 0, 0, 0, 0, 0
	for _, runName := range []string{"baseline1034", "additive4"} {
		e = bundle.rows(bundle.Index.Runs[runName], func(raw json.RawMessage) error {
			var row rddRow4E0B30
			if e := json.Unmarshal(raw, &row); e != nil {
				return e
			}
			g, e := ddiBuildGraph4E0B30(fixture, index[row.Name])
			if e != nil {
				return e
			}
			defer g.close()
			plan, e := rddNativePlan4E0B30(bundle, &row, fixture, g)
			if e != nil {
				return e
			}
			rows++
			derived += plan.DerivedReadAttempts
			for _, rawEvent := range row.Events {
				var event rddEvent4E0B30
				if e = json.Unmarshal(rawEvent, &event); e != nil {
					return e
				}
				if event.Kind == "read" {
					originalReads++
				}
			}
			for _, step := range plan.Steps {
				if step.DerivedAttempt && (step.Kind != "read" || step.Event.Completed || step.Event.ValueHex != nil || step.Event.Source != "qualified-original-fault-operand-no-read-hook") {
					return fmt.Errorf("derived attempt was counted as original completed read")
				}
			}
			if row.Status == "returned" {
				returned++
				if plan.DerivedReadAttempts != 0 {
					return fmt.Errorf("derived attempt in returned original")
				}
			} else {
				if len(row.Faults) != 1 {
					return fmt.Errorf("original fault shape count")
				}
				var fault rddEvent4E0B30
				if e = json.Unmarshal(row.Faults[0], &fault); e != nil {
					return e
				}
				if len(plan.Steps) == 0 || plan.Steps[len(plan.Steps)-1].Kind != "fault" {
					return fmt.Errorf("terminal original fault step absent")
				}
				switch fault.Access {
				case 19, 22:
					if plan.DerivedReadAttempts == 1 {
						readNoHook++
					} else if plan.DerivedReadAttempts == 0 {
						readHook++
					} else {
						return fmt.Errorf("multiple derived read attempts")
					}
				case 20, 23:
					writeHook++
					if plan.DerivedReadAttempts != 0 {
						return fmt.Errorf("derived read for store fault")
					}
				case 21, 24:
					fetch++
					if plan.DerivedReadAttempts != 0 {
						return fmt.Errorf("derived read for fetch fault")
					}
				default:
					return fmt.Errorf("unclassified original fault")
				}
			}
			return g.checkCanaries()
		})
		if e != nil {
			t.Fatal(e)
		}
	}
	if rows != 2076 || originalReads != 72874 || derived != 106 || readNoHook != 106 || readHook != 16 || writeHook != 32 || fetch != 20 || returned != 1902 {
		t.Fatalf("all native observation plan shapes rows%d reads%d derived%d nohook%d hooked%d stores%d fetch%d returned%d", rows, originalReads, derived, readNoHook, readHook, writeHook, fetch, returned)
	}
	t.Log("RAW_DEFAULT_DAMAGE_ALL_NATIVE_PLANS_OK original_rows=2076 original_read_hooks=72874 derived_incomplete_read_attempts=106 observed_read_fault_hooks=16 observed_store_fault_hooks=32 observed_fetch_faults=20 fresh_native_graphs=2076 native_parent_leaves=0 original_records_modified=false")
}
