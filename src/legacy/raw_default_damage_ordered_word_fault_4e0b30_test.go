//go:build darwin && arm64

package legacy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"
	"unsafe"
)

// Run exactly the original partial-WORD rows through both actual root entries.
// Its namespace is separate from the complete2076-row roots, so a full suite
// preserves every child without filename collisions or row-name exceptions.
func TestRawDefaultDamageOrderedWordFault4E0B30(t *testing.T) {
	if rddNativeAdapterPhaseGate4E0B30(t) {
		t.Fatal("ordered WORD comparison requires the compiled native root")
	}
	root, err := rddNativeArtifactsRoot4E0B30()
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("OPENNOX_RAW_DEFAULT_DAMAGE_NATIVE_ADAPTER_ARTIFACTS", filepath.Join(root, "ordered-word-focused"))
	fixture, err := ddiLoadFixture4E0B30()
	if err != nil {
		t.Fatal(err)
	}
	bundle, err := rddOpenBundle4E0B30()
	if err != nil {
		t.Fatal(err)
	}
	defer bundle.close()
	attempted := 0
	for _, run := range []string{"baseline1034", "additive4"} {
		err = bundle.rows(bundle.Index.Runs[run], func(raw json.RawMessage) error {
			var row rddRow4E0B30
			if err := json.Unmarshal(raw, &row); err != nil {
				return err
			}
			if row.PartialCount == 0 {
				return nil
			}
			for _, backend := range []string{"Go", "C"} {
				t.Run(row.Mode+"/"+backend, func(t *testing.T) {
					attempted++
					rddBindNativeRowSource4E0B30(&row, raw)
					if err := rddRunSupervisedNativeRow4E0B30(t, bundle, &row, fixture, backend); err != nil {
						t.Fatal(err)
					}
				})
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	if attempted != 4 {
		t.Fatalf("actual original WORD children=%d, want4", attempted)
	}
}

func TestRawDefaultDamageOrderedWordExpectedBankControls4E0B30(t *testing.T) {
	rddNativeAdapterPhaseGate4E0B30(t)
	fixture, err := ddiLoadFixture4E0B30()
	if err != nil {
		t.Fatal(err)
	}
	bundle, err := rddOpenBundle4E0B30()
	if err != nil {
		t.Fatal(err)
	}
	defer bundle.close()
	checked := 0
	for _, run := range []string{"baseline1034", "additive4"} {
		err = bundle.rows(bundle.Index.Runs[run], func(raw json.RawMessage) error {
			var row rddRow4E0B30
			if err := json.Unmarshal(raw, &row); err != nil {
				return err
			}
			if row.PartialCount == 0 {
				return nil
			}
			graph, err := ddiBuildGraph4E0B30(fixture, rddFixtureIndex4E0B30(fixture)[row.Name])
			if err != nil {
				return err
			}
			defer graph.close()
			plan, err := rddNativePlan4E0B30(bundle, &row, fixture, graph)
			if err != nil {
				return err
			}
			for _, step := range plan.Steps {
				if len(step.TerminalFragments) == 0 {
					continue
				}
				checked++
				token := graph.Tokens[step.Base]
				allocation := &graph.Allocations[token.Allocation]
				actual := unsafe.Slice((*byte)(allocation.Map), graph.Page)
				before := append([]byte(nil), actual...)
				if err := plan.stageTerminalFragments(step); err != nil {
					return err
				}
				if !bytes.Equal(before, actual) || bytes.Equal(actual, allocation.Expected) {
					return fmt.Errorf("expected preparation changed actual pages or accepted untouched bytes")
				}
				// This is a declared control buffer mutation, never an AFTER
				// artifact or a replacement result for an actual fault child.
				at := graph.Page - 1
				actual[at] = byte(step.Event.Value) ^ 1
				if bytes.Equal(actual, allocation.Expected) {
					return fmt.Errorf("wrong actual low byte was accepted")
				}
				actual[at] = byte(step.Event.Value)
				if !bytes.Equal(actual, allocation.Expected) {
					return fmt.Errorf("exact original terminal-byte control did not match")
				}
				copy(actual, before)
				copy(allocation.Expected, before)
				wrong := step
				wrong.TerminalFragments = append([]rddEvent4E0B30(nil), step.TerminalFragments...)
				wrong.TerminalFragments[0].WriteOffset = 1
				if plan.stageTerminalFragments(wrong) == nil || !bytes.Equal(actual, before) || !bytes.Equal(allocation.Expected, before) {
					return fmt.Errorf("wrong fragment ordering admitted or modified buffers")
				}
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	if checked != 2 {
		t.Fatalf("expected-bank original controls=%d, want2", checked)
	}
}
