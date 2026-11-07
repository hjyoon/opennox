//go:build darwin && arm64

package legacy

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unsafe"
)

func TestRawDefaultDamageNativeOperandControls4E0B30(t *testing.T) {
	rddNativeAdapterPhaseGate4E0B30(t)
	f, e := ddiLoadFixture4E0B30()
	if e != nil {
		t.Fatal(e)
	}
	index := rddFixtureIndex4E0B30(f)
	b, e := rddOpenBundle4E0B30()
	if e != nil {
		t.Fatal(e)
	}
	defer b.close()
	t.Run("actual-aligned-unaligned-and-live-mutation-prefixes", func(t *testing.T) {
		for _, name := range []string{"physical-type-00000000-damage-00000014", "unaligned-target-plus-1", "stimulus-early-class-variant-0"} {
			func() {
				row, e := rddControlRow4E0B30(b, name)
				if e != nil {
					t.Fatal(e)
				}
				g, e := ddiBuildGraph4E0B30(f, index[name])
				if e != nil {
					t.Fatal(e)
				}
				defer g.close()
				c, e := rddNativePlan4E0B30(b, row, f, g)
				if e != nil {
					t.Fatal(e)
				}
				pop := c.install()
				defer pop()
				damage := int32(c.Input.Args[3])
				rawDefaultDamageObserveCaller4E0B30(&damage, 0x4e0b30, true, true)
				target := c.pointer(c.Input.Args[0])
				response, ok := rawDefaultDamageDispatchObservedService4E0B30(rawDefaultDamageNativeService4E0B30{Site: 0x4e0b46, Arguments: []rawDefaultDamageNativeArgument4E0B30{{Kind: "object", Pointer: target}, {Kind: "word", Word: 23}}})
				if !ok || response.Word != 0 {
					t.Fatal("controlled prefix initial service input")
				}
				cached := rawDefaultDamageReadNativeOperand4E0B30(target, 1, 1, 0x4e0b7b, false)
				declaredClass := uint32(0)
				declaredFound := false
				for _, field := range index[name].Fields {
					if field.Role == "target" && field.Field == 1 {
						raw, e := hex.DecodeString(field.Hex)
						if e != nil || len(raw) < 1 {
							t.Fatal("independent initial target class bytes")
						}
						declaredClass = uint32(raw[0])
						declaredFound = true
						break
					}
				}
				if !declaredFound || cached.Word != declaredClass {
					t.Fatal("actual cached low BYTE differs from independent initial input")
				}

				if name == "stimulus-early-class-variant-0" && rawParentInitialCBits4E0B30(target, g.Offsets[1], 4) != 4 {
					t.Fatal("independent after-read class mutation was not applied")
				}
				if cached.Word&2 != 0 {
					update := rawDefaultDamageReadNativeOperand4E0B30(target, 18, 8, 0x4e0b86, true)
					if update.Pointer != c.pointer(c.Input.Roles["target_update"]) {
						t.Fatal("actual full high update pointer")
					}
					rawDefaultDamageWriteNativeOperand4E0B30(update.Pointer, 24, 4, 0x4e0b8f, false, rawDefaultDamageNativeValue4E0B30{})
				}
				if e = c.Projection.check(); e != nil {
					t.Fatal(e)
				}
				if e = c.finishReturned(1); e == nil || !strings.Contains(e.Error(), "no actual raw root") {
					t.Fatal("transport prefix was promoted to parent proof")
				}
			}()
		}
		t.Log("RAW_DEFAULT_DAMAGE_NATIVE_OPERAND_PREFIX_CONTROLS cases=3 aligned_and_unaligned_Cowned=true live_after_read_mutation=true actual_parent_leaves=0")
	})
	t.Run("wrong-faulting-store-attempt-is-durably-rejected", func(t *testing.T) {
		var row *rddRow4E0B30
		for _, runName := range []string{"baseline1034", "additive4"} {
			e = b.rows(b.Index.Runs[runName], func(raw json.RawMessage) error {
				var candidate rddRow4E0B30
				if e := json.Unmarshal(raw, &candidate); e != nil {
					return e
				}
				if candidate.PartialCount != 0 && candidate.Mode == "startup-53-chop" {
					if row != nil {
						return fmt.Errorf("expected one original fragmented WORD input")
					}
					copy := candidate
					row = &copy
				}
				return nil
			})
			if e != nil {
				t.Fatal(e)
			}
		}
		if row == nil {
			t.Fatal("original WORD fault case absent")
		}
		g, e := ddiBuildGraph4E0B30(f, index[row.Name])
		if e != nil {
			t.Fatal(e)
		}
		defer g.close()
		c, e := rddNativePlan4E0B30(b, row, f, g)
		if e != nil {
			t.Fatal(e)
		}
		found := -1
		for i, step := range c.Steps {
			if step.Kind == "write" && step.Field == 20 && !step.Event.Completed && step.Event.Value&0xffff == 2 {
				if found != -1 {
					t.Fatal("ambiguous WORD control")
				}
				found = i
			}
		}
		if found < 0 {
			t.Fatal("original attempted WORD2 absent")
		}
		// Position a transport-only negative control at this single original
		// operand. No skipped prefix is credited and no engine is dispatched.
		c.Position = found
		pop := c.install()
		defer pop()
		defer ddiFaultMode4E0B30()()
		root := os.Getenv("OPENNOX_RAW_DEFAULT_DAMAGE_INITIAL_ARTIFACTS")
		if root == "" {
			t.Fatal("preserved initial artifact root required")
		}
		directory := filepath.Join(root, "native-wrong-store-attempt-control-v1")
		if e = os.Mkdir(directory, 0700); e != nil {
			t.Fatal(e)
		}
		path := filepath.Join(directory, "actual-shared-validation-failure.bin")
		file, e := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
		if e != nil {
			t.Fatal(e)
		}
		defer file.Close()
		if e = file.Truncate(int64(g.Page)); e != nil {
			t.Fatal(e)
		}
		shared, e := rawDefaultDamageSharedMap4E0B30(int(file.Fd()), 0, g.Page)
		if e != nil {
			t.Fatal(e)
		}
		defer rawDefaultDamageSharedUnmap4E0B30(shared, g.Page)
		c.ValidationFailure = func(position int, message string) {
			*(*uint64)(shared) = 1
			*(*uint64)(unsafe.Add(shared, 8)) = uint64(position)
			if e := rawDefaultDamageSharedSync4E0B30(shared, g.Page); e != nil {
				panic(e)
			}
		}
		step := c.Steps[found]
		caught := rddRecoverControl4E0B30(func() {
			rawDefaultDamageWriteNativeOperand4E0B30(c.pointer(step.Base), 20, 2, uint32(mustRDDNumber4E0B30(step.Event.PC)), false, rawDefaultDamageNativeValue4E0B30{Word: 3})
		})
		if caught == nil || !strings.Contains(fmt.Sprint(caught), "attempted scalar store") {
			t.Fatal("wrong attempted WORD value was admitted or misclassified", caught)
		}
		if e = file.Sync(); e != nil {
			t.Fatal(e)
		}
		raw, e := os.ReadFile(path)
		if e != nil || len(raw) != g.Page || ddiUnsigned4E0B30(raw[:8]) != 1 || ddiUnsigned4E0B30(raw[8:16]) != uint64(found) {
			t.Fatal("validation failure survived only in private Go memory")
		}
		if e = g.checkCanaries(); e != nil {
			t.Fatal("rejected store modified actual native operand", e)
		}
		if e = ddiSaveJSON4E0B30(filepath.Join(directory, "actual-wrong-store-rejection.json"), map[string]any{"original_input": row.Name, "original_hook_attempt": 2, "actual_control_attempt": 3, "shared_failure_flag": 1, "shared_failure_SHA256": ddiSHA4E0B30(raw), "whole_native_buffers_unchanged": true, "actual_unsafe_store_executed": false, "actual_parent_leaves": 0, "original_fragment_rewritten": false}); e != nil {
			t.Fatal(e)
		}
		t.Log("RAW_DEFAULT_DAMAGE_WRONG_ATTEMPT_CONTROL expected_WORD=2 actual_WORD=3 rejected_before_store=true shared_failure_persisted=true actual_parent_leaves=0")
	})
}
