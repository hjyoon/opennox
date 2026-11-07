//go:build darwin && arm64

package legacy

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"unsafe"
)

func rddControlRow4E0B30(b *rddBundle4E0B30, name string) (*rddRow4E0B30, error) {
	var found *rddRow4E0B30
	for _, runName := range []string{"baseline1034", "additive4"} {
		e := b.rows(b.Index.Runs[runName], func(raw json.RawMessage) error {
			var row rddRow4E0B30
			if e := json.Unmarshal(raw, &row); e != nil {
				return e
			}
			if row.Name == name && row.Mode == "startup-53-chop" {
				if found != nil {
					return fmt.Errorf("duplicate control row")
				}
				copy := row
				found = &copy
			}
			return nil
		})
		if e != nil {
			return nil, e
		}
	}
	if found == nil {
		return nil, fmt.Errorf("original control row missing %s", name)
	}
	return found, nil
}
func TestRawDefaultDamageNativeAdapterControls4E0B30(t *testing.T) {
	before := rddNativeAdapterPhaseGate4E0B30(t)
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
	t.Run("original-verdict-needs-actual-root", func(t *testing.T) {
		row, e := rddControlRow4E0B30(b, f.Cases[0].Input.Name)
		if e != nil {
			t.Fatal(e)
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
		if e = c.finishReturned(int32(row.RegistersAfter["eax"])); e == nil || !strings.Contains(e.Error(), "no actual raw root") {
			t.Fatal("transport-only root verdict was qualified")
		}
		for _, backend := range []string{"Go", "C"} {
			func() {
				candidateGraph, err := ddiBuildGraph4E0B30(f, index[row.Name])
				if err != nil {
					t.Fatal(err)
				}
				defer candidateGraph.close()
				candidate, err := rddNativePlan4E0B30(b, row, f, candidateGraph)
				if err != nil {
					t.Fatal(err)
				}
				result, caught, err := candidate.execute(backend)
				if before {
					if err == nil || caught != nil || candidate.DispatchedRoot {
						t.Fatal("absent engine/old C ABI was dispatched", backend)
					}
				} else {
					if err != nil || caught != nil {
						t.Fatal("actual postport nil-target dispatch", backend, err, caught)
					}
					if err = candidate.finishReturned(result); err != nil {
						t.Fatal(err)
					}
				}
			}()
		}

		if e = g.checkCanaries(); e != nil {
			t.Fatal(e)
		}
	})
	t.Run("recorded-early-service-and-clock-transport", func(t *testing.T) {
		row, e := rddControlRow4E0B30(b, "early-buff23-00000001-frame-00000000")
		if e != nil {
			t.Fatal(e)
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
		pop := c.install()
		defer pop()
		damage := int32(c.Input.Args[3])
		rawDefaultDamageObserveCaller4E0B30(&damage, 0x4e0b30, true, true)
		target := c.pointer(c.Input.Args[0])
		// Explicit transport control script; no raw parent is called and the
		// plan is never passed to an engine as a gameplay execution schedule.
		first, ok := rawDefaultDamageDispatchObservedService4E0B30(rawDefaultDamageNativeService4E0B30{Site: 0x4e0b46, Arguments: []rawDefaultDamageNativeArgument4E0B30{{Kind: "object", Pointer: target}, {Kind: "word", Word: 23}}})
		if !ok || first.Word != 1 {
			t.Fatal("declared Buff23 service reply")
		}
		frame, ok := rawDefaultDamageObservedFrame4E0B30(0x4e0b52, 1)
		if !ok || frame != 0 {
			t.Fatal("native frame BYTE input")
		}
		_, ok = rawDefaultDamageDispatchObservedService4E0B30(rawDefaultDamageNativeService4E0B30{Site: 0x4e0b66, Arguments: []rawDefaultDamageNativeArgument4E0B30{{Kind: "word", Word: 71}, {Kind: "object", Pointer: target}, {Kind: "word"}, {Kind: "word"}}})
		if !ok || c.Position != len(c.Steps) || c.Pending != nil {
			t.Fatal("full controlled service/read order")
		}
		if e = c.finishReturned(1); e == nil {
			t.Fatal("manual adapter transport was promoted to root engine proof")
		}
		if e = c.Projection.check(); e != nil {
			t.Fatal(e)
		}
	})
	t.Run("callback-admission-uses-declared-endpoints-only", func(t *testing.T) {
		g, e := ddiBuildGraph4E0B30(f, f.Cases[0])
		if e != nil {
			t.Fatal(e)
		}
		defer g.close()
		row, e := rddControlRow4E0B30(b, f.Cases[0].Input.Name)
		if e != nil {
			t.Fatal(e)
		}
		c, e := rddNativePlan4E0B30(b, row, f, g)
		if e != nil {
			t.Fatal(e)
		}
		c.Position = len(c.Steps)
		pop := c.install()
		func() {
			defer pop()
			for _, target := range []unsafe.Pointer{nil, g.Tokens[f.Cases[0].Input.Roles["target"]].Pointer} {
				_, handled := rawDefaultDamageDispatchObservedService4E0B30(rawDefaultDamageNativeService4E0B30{Site: 0x4e0d05, Indirect: true, Callback: target})
				if handled {
					t.Fatal("nil/non-function callback admitted as bounded service")
				}
			}
		}()
		controlled := rawDefaultDamagePushNativeObserver4E0B30(&rawDefaultDamageNativeObserver4E0B30{Service: func(call rawDefaultDamageNativeService4E0B30) (rawDefaultDamageNativeReply4E0B30, bool) {
			if call.Indirect && !rddDeclaredNativeCallback4E0B30(g, call.Callback) {
				return rawDefaultDamageNativeReply4E0B30{}, false
			}
			return rawDefaultDamageNativeReply4E0B30{Word: 0x80000001}, true
		}})
		defer controlled()
		for _, word := range []uint32{0x23001000, 0x23002000} {
			target := g.Tokens[word].Pointer
			if !rddDeclaredNativeCallback4E0B30(g, target) {
				t.Fatal("known fixture endpoint not recognized")
			}
			reply, handled := rawDefaultDamageDispatchObservedService4E0B30(rawDefaultDamageNativeService4E0B30{Site: 0x4e0d05, Indirect: true, Callback: target})
			if !handled || reply.Word != 0x80000001 {
				t.Fatal("known endpoint transport handling")
			}
		}
		if e = g.checkCanaries(); e != nil {
			t.Fatal(e)
		}
		t.Log("RAW_DEFAULT_DAMAGE_CALLBACK_ADMISSION_CONTROL nil=false nonfunction=false declared2=true trace_lookup_for_admission=false actual_callback_invocations=0 native_parent_leaves=0")
	})
	t.Run("actual-C3-C4-C5-inout-positions", func(t *testing.T) {
		inputs := map[string]rddInput4E0B30{}
		for _, runName := range []string{"baseline1034", "additive4"} {
			packed, e := b.member(b.Index.Runs[runName].Inputs)
			if e != nil {
				t.Fatal(e)
			}
			var data struct{ Cases []rddInput4E0B30 }
			if e = rddDecodeGzip4E0B30(packed, &data); e != nil {
				t.Fatal(e)
			}
			for _, in := range data.Cases {
				inputs[in.Name] = in
			}
		}
		counts := map[int]int{}
		for _, fixtureCase := range f.Cases {
			in := inputs[fixtureCase.Input.Name]
			g, e := ddiBuildGraph4E0B30(f, fixtureCase)
			if e != nil {
				t.Fatal(e)
			}
			func() {
				defer g.close()
				allocation, e := g.alloc("actual-C-inout-ABI-control", 256, 0x67)
				if e != nil {
					t.Fatal(e)
				}
				cell := (*int32)(g.base(allocation))
				a := &g.Allocations[allocation]
				for siteText, schedule := range in.Services {
					site := uint32(mustRDDNumber4E0B30(siteText))
					for _, raw := range schedule.Responses {
						var response rddResponse4E0B30
						if e = json.Unmarshal(raw, &response); e != nil {
							t.Fatal(e)
						}
						for _, effect := range response.Effects {
							argc, damageIndex, e := rawDefaultDamageInoutABI4E0B30(site)
							if e != nil || effect.Arg == nil || *effect.Arg != damageIndex || effect.Address != nil || effect.Role != "" || effect.Offset != 0 {
								t.Fatal("declared original damage operand index/site/argc", e)
							}
							data, e := hex.DecodeString(effect.Hex)
							if e != nil || len(data) != 4 || response.EAX == nil {
								t.Fatal("explicit input DWORD mutation")
							}
							objects := [3]unsafe.Pointer{g.Tokens[fixtureCase.Input.Args[0]].Pointer, g.Tokens[fixtureCase.Input.Args[1]].Pointer, g.Tokens[fixtureCase.Input.Args[2]].Pointer}
							if argc == 3 {
								objects[0], objects[1] = objects[1], objects[0]
							}
							mutation := uint32(ddiUnsigned4E0B30(data))
							words, actualArgc, result, e := rawDefaultDamageTransportInoutABI4E0B30(site, objects, cell, fixtureCase.Input.Args[4], *response.EAX, mutation)
							if e != nil {
								t.Fatal(e)
							}
							var want [5]uintptr
							switch argc {
							case 3:
								want = [5]uintptr{uintptr(objects[0]), uintptr(objects[1]), uintptr(unsafe.Pointer(cell))}
							case 4:
								want = [5]uintptr{uintptr(objects[0]), uintptr(unsafe.Pointer(cell)), uintptr(fixtureCase.Input.Args[4]), uintptr(objects[1])}
							case 5:
								want = [5]uintptr{uintptr(objects[0]), uintptr(objects[1]), uintptr(objects[2]), uintptr(unsafe.Pointer(cell)), uintptr(fixtureCase.Input.Args[4])}
							}
							if actualArgc != argc || words != want || result != *response.EAX || uint32(*cell) != mutation {
								t.Fatal("actual exact native C inout ABI capture")
							}
							copy(a.Expected[a.Base:], data)
							counts[argc]++
						}
					}
				}
				if e = g.checkCanaries(); e != nil {
					t.Fatal(e)
				}
			}()
		}
		if counts[3] != 12 || counts[4] != 36 || counts[5] != 54 {
			t.Fatalf("declared actual ABI input counts %+v", counts)
		}
		t.Log("RAW_DEFAULT_DAMAGE_ACTUAL_INOUT_ABI_CONTROLS C3=12 slot2 C4=36 slot1 C5=54 slot3 native_root_leaves=0")
	})
}

func rddNativeAdapterPhaseGate4E0B30(t *testing.T) bool {
	t.Helper()
	flag := os.Getenv("OPENNOX_RAW_DEFAULT_DAMAGE_BEFORE")
	if flag == "1" {
		rddBeforeOnly4E0B30(t)
		if rawDefaultDamageNativeObserverStack4E0B30 != nil {
			t.Fatal("active native observer context")
		}
		return true
	}
	if flag != "" && flag != "0" {
		t.Fatal("unknown raw parent test phase")
	}
	if rawParentABI4E0B30() != 2 || rawDefaultDamageNativeEngine4E0B30 == nil {
		t.Fatal("postport/default phase requires actual typed C ABI2 and native Go raw entry")
	}
	if rawDefaultDamageCurrentContext4E0B30 != nil || rawDefaultDamageContextStack4E0B30 != nil || rawDefaultDamageNativeObserverStack4E0B30 != nil || damagePredicateRawObserve4E1400 != nil {
		t.Fatal("nondefault raw observer state")
	}
	return false
}
