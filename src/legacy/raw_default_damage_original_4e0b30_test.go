//go:build darwin && arm64

package legacy

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"testing"
	"unsafe"
)

func rddBeforeOnly4E0B30(t *testing.T) {
	t.Helper()
	// Both phases retain every original certificate and transport assertion.
	// Only the entry contract changes after the actual native implementation.
	switch os.Getenv("OPENNOX_RAW_DEFAULT_DAMAGE_BEFORE") {
	case "1":
		if rawParentABI4E0B30() != 1 || rawDefaultDamageNativeEngine4E0B30 != nil {
			t.Fatal("preport requires old C ABI1 and absent native raw parent")
		}
	case "", "0":
		if rawParentABI4E0B30() != 2 || rawDefaultDamageNativeEngine4E0B30 == nil {
			t.Fatal("postport requires typed C ABI2 and registered native raw parent")
		}
	default:
		t.Fatal("unknown raw parent test phase")
	}
	if rawDefaultDamageCurrentContext4E0B30 != nil || rawDefaultDamageContextStack4E0B30 != nil {
		t.Fatal("nonempty serial observer context")
	}
}
func rddFixtureIndex4E0B30(f *ddiFixture4E0B30) map[string]ddiCase4E0B30 {
	out := map[string]ddiCase4E0B30{}
	for _, c := range f.Cases {
		out[c.Input.Name] = c
	}
	return out
}

func TestRawDefaultDamageOriginalCertificates4E0B30(t *testing.T) {
	rddBeforeOnly4E0B30(t)
	fixture, e := ddiLoadFixture4E0B30()
	if e != nil {
		t.Fatal(e)
	}
	cases := rddFixtureIndex4E0B30(fixture)
	bundle, e := rddOpenBundle4E0B30()
	if e != nil {
		t.Fatal(e)
	}
	defer bundle.close()
	counts := rddCertificate4E0B30{}
	seen := map[string]bool{}
	returned, faulted, typed, nested := 0, 0, 0, 0
	for _, name := range []string{"baseline1034", "additive4"} {
		run, ok := bundle.Index.Runs[name]
		if !ok {
			t.Fatal("missing immutable run")
		}
		e = bundle.rows(run, func(raw json.RawMessage) error {
			var row rddRow4E0B30
			if err := json.Unmarshal(raw, &row); err != nil {
				return err
			}
			key := row.Mode + "/" + row.Name
			if seen[key] {
				return fmt.Errorf("duplicate original input/FPU row")
			}
			seen[key] = true
			c, ok := cases[row.Name]
			if !ok || c.InputSHA != row.InputSHA {
				return fmt.Errorf("original/independent initial fixture input identity")
			}
			if !t.Run(key, func(t *testing.T) {
				defer ddiFaultMode4E0B30()()
				proof, err := rddValidateRow4E0B30(bundle, &row)
				if err != nil {
					t.Fatal(err)
				}
				counts.Calls += proof.Calls
				counts.Reads += proof.Reads
				counts.Writes += proof.Writes
				counts.Faults += proof.Faults
				counts.Fragments += proof.Fragments
				counts.FaultingWrites += proof.FaultingWrites
				counts.Services += proof.Services
				counts.UsedReads += proof.UsedReads
				counts.UsedServices += proof.UsedServices
				g, err := ddiBuildGraph4E0B30(fixture, c)
				if err != nil {
					t.Fatal(err)
				}
				defer g.close()
				var pointers [3]unsafe.Pointer
				for n, word := range c.Input.Args[:3] {
					token, found := g.Tokens[word]
					if !found {
						t.Fatal("unbound root argument")
					}
					pointers[n] = token.Pointer
					if word != 0 && token.Bits <= 0xffffffff {
						t.Fatal("root C argument is not high full-width")
					}
				}
				// Returned original EAX is transported as signed int32. A fault has
				// no return: it uses an explicitly independent transport sentinel.
				response := int32(-2147483647)
				if proof.Returned {
					response = int32(proof.Return)
					returned++
				} else {
					faulted++
				}
				got := rawDefaultDamageCTransport4E0B30(pointers[0], pointers[1], pointers[2], int32(c.Input.Args[3]), int32(c.Input.Args[4]), response, true, int32(-123456789))
				want := [5]uintptr{uintptr(pointers[0]), uintptr(pointers[1]), uintptr(pointers[2]), uintptr(c.Input.Args[3]), uintptr(c.Input.Args[4])}
				inner := [5]uintptr{want[1], want[2], want[0], uintptr(c.Input.Args[3] ^ 0x80000000), uintptr(c.Input.Args[4] ^ 0xffffffff)}
				if got.Outer != want || got.Inner != inner || got.Return != response || got.InnerReturn != -123456789 || !got.Restored {
					t.Fatalf("actual typed C transport/reentry: %+v", got)
				}
				typed++
				nested++
				if err = g.checkCanaries(); err != nil {
					t.Fatal(err)
				}
				if _, err = g.roundtrip(); err != nil {
					t.Fatal(err)
				}
			}) {
				return fmt.Errorf("original certificate/transport leaf failed")
			}
			return nil
		})
		if e != nil {
			t.Fatal(e)
		}
	}
	if len(seen) != 2076 || returned != 1902 || faulted != 174 || typed != 2076 || nested != 2076 || counts.Reads != 72874 || counts.Writes != 97472 || counts.Calls != 23642 || counts.Services != 21332 || counts.Faults != 174 || counts.Fragments != 2 || counts.FaultingWrites != 32 || counts.UsedReads != 98 || counts.UsedServices != 60 {
		t.Fatalf("original exact finite-corpus counts rows=%d return=%d fault=%d counts=%+v", len(seen), returned, faulted, counts)
	}
	t.Logf("RAW_DEFAULT_DAMAGE_ORIGINAL_CERTIFICATES_OK rows=2076 returned=1902 faults=174 reads=72874 writes=97472 calls=23642 external_boundaries=21332 used_read_stimuli=98 used_service_stimuli=60 original_partial_fragments=2 faulting_writes=32 typed_C_transports=2076 nested_C_reentries=2076 native_engine_leaves=0 native_partial_WORD_equivalence=false")
}

func TestRawDefaultDamageDeclaredInputTransports4E0B30(t *testing.T) {
	rddBeforeOnly4E0B30(t)
	f, e := ddiLoadFixture4E0B30()
	if e != nil {
		t.Fatal(e)
	}
	bundle, e := rddOpenBundle4E0B30()
	if e != nil {
		t.Fatal(e)
	}
	defer bundle.close()
	inputs := map[string]rddInput4E0B30{}
	for _, name := range []string{"baseline1034", "additive4"} {
		run := bundle.Index.Runs[name]
		packed, err := bundle.member(run.Inputs)
		if err != nil {
			t.Fatal(err)
		}
		var data struct {
			Cases []json.RawMessage `json:"cases"`
		}
		if err = rddDecodeGzip4E0B30(packed, &data); err != nil {
			t.Fatal(err)
		}
		for _, raw := range data.Cases {
			var in rddInput4E0B30
			if err = json.Unmarshal(raw, &in); err != nil {
				t.Fatal(err)
			}
			canonical, err := rddCanonical4E0B30(raw)
			if err != nil {
				t.Fatal(err)
			}
			c, ok := rddFixtureIndex4E0B30(f)[in.Name]
			if !ok || c.InputSHA != ddiSHA4E0B30(canonical) {
				t.Fatal("declared input canonical fixture identity")
			}
			inputs[in.Name] = in
		}
	}
	packets, effects, damageEffects := 0, 0, 0
	for _, c := range f.Cases {
		c := c
		t.Run(c.Input.Name, func(t *testing.T) {
			defer ddiFaultMode4E0B30()()
			in, ok := inputs[c.Input.Name]
			if !ok {
				t.Fatal("missing independent input")
			}
			g, err := ddiBuildGraph4E0B30(f, c)
			if err != nil {
				t.Fatal(err)
			}
			defer g.close()
			p, err := rddNewProjection4E0B30(g, in)
			if err != nil {
				t.Fatal(err)
			}
			// This phase exhausts declared effects once to qualify transport. It is
			// not parent scheduling/reachability credit; actual use is certified in
			// the original ledger above and later observed from the native engine.
			for _, stimuli := range [][]rddStimulus4E0B30{in.AfterReads, in.AfterServices} {
				for _, s := range stimuli {
					for _, ef := range s.Effects {
						if err = p.apply(ef, nil); err != nil {
							t.Fatal(err)
						}
						effects++
					}
				}
			}
			index, err := g.alloc("explicit-input-damage-cell", 256, 0x5c)
			if err != nil {
				t.Fatal(err)
			}
			cell := (*int32)(g.base(index))
			a := &g.Allocations[index]
			*cell = int32(c.Input.Args[3])
			copy(a.Expected[a.Base:], ddiBytes4E0B30(uint64(c.Input.Args[3]), 4))
			for _, schedule := range in.Services {
				for _, raw := range schedule.Responses {
					var response rddResponse4E0B30
					if err = json.Unmarshal(raw, &response); err != nil {
						t.Fatal(err)
					}
					if response.X87 != nil {
						payload, err := hex.DecodeString(response.X87.Hex)
						if err != nil {
							t.Fatal(err)
						}
						got, err := rawDefaultDamageCPacketTransport4E0B30(response.X87.Format, payload)
						if err != nil || !bytes.Equal(got, payload) {
							t.Fatal("actual C raw FPU packet transport", err)
						}
						packets++
					}
					for _, ef := range response.Effects {
						if ef.Arg == nil || ef.Address != nil || ef.Role != "" || ef.Offset != 0 {
							t.Fatal("unexpected declared inout control domain")
						}
						data, err := hex.DecodeString(ef.Hex)
						if err != nil || len(data) != 4 {
							t.Fatal("explicit DWORD inout bytes")
						}
						mutation := uint32(ddiUnsigned4E0B30(data))
						if response.EAX == nil {
							t.Fatal("explicit EAX service response missing")
						}
						pointers := [3]unsafe.Pointer{g.Tokens[c.Input.Args[0]].Pointer, g.Tokens[c.Input.Args[1]].Pointer, g.Tokens[c.Input.Args[2]].Pointer}
						got, word := rawDefaultDamageCServiceTransport4E0B30(pointers[0], pointers[1], pointers[2], cell, c.Input.Args[4], *response.EAX, mutation)
						want := [5]uintptr{uintptr(pointers[0]), uintptr(pointers[1]), uintptr(pointers[2]), uintptr(unsafe.Pointer(cell)), uintptr(c.Input.Args[4])}
						if got != want || word != *response.EAX || uint32(*cell) != mutation {
							t.Fatal("actual typed five-operand C inout transport")
						}
						copy(a.Expected[a.Base:], data)
						damageEffects++
					}
				}
			}
			for _, token := range g.Tokens {
				if rawDefaultDamageCPointerTransport4E0B30(token.Pointer) != token.Pointer {
					t.Fatal("full-width actual C pointer return")
				}
			}
			if err = p.check(); err != nil {
				t.Fatal(err)
			}
			runtime.KeepAlive(cell)
		})
	}
	if len(inputs) != 1038 || packets != 4152 || effects != 87 || damageEffects != 102 {
		t.Fatalf("declared transport domain input=%d packets=%d effects=%d inout=%d", len(inputs), packets, effects, damageEffects)
	}
	t.Log("RAW_DEFAULT_DAMAGE_DECLARED_INPUT_TRANSPORTS_OK inputs=1038 raw_FPU_packets=4152 explicit_absolute_effects=87 typed_C_inout_effects=102 native_engine_leaves=0 input_transport_only=true")
}

func TestRawDefaultDamageObserverContract4E0B30(t *testing.T) {
	rddBeforeOnly4E0B30(t)
	t.Run("cached-byte-then-reloaded-DWORD", func(t *testing.T) {
		p := ddiMap4E0B30(os.Getpagesize())
		if p == nil {
			t.Fatal("mmap")
		}
		defer ddiUnmap4E0B30(p, os.Getpagesize())
		ddiFill4E0B30(p, os.Getpagesize(), 0x69)
		base := unsafe.Add(p, 256)
		*(*uint32)(base) = 0x01020304
		events := []rawDefaultDamageAccess4E0B30{}
		context := &rawDefaultDamageContext4E0B30{AfterRead: func(e rawDefaultDamageAccess4E0B30) {
			events = append(events, e)
			if len(events) == 1 {
				*(*uint32)(base) = 0x89abcdef
			}
		}}
		pop := rawDefaultDamagePushContext4E0B30(context)
		defer pop()
		first := rawDefaultDamageForeignRead4E0B30(base, 0, 1, 0x4e0b52)
		second := rawDefaultDamageForeignRead4E0B30(base, 0, 4, 0x4e0bea)
		if first != 4 || second != 0x89abcdef || len(events) != 2 || events[0].Value != 4 || events[1].Value != second || events[0].NativeWidth != 1 || events[1].NativeWidth != 4 {
			t.Fatal("load-before-afterread mutation/cache/reload order")
		}
	})
	t.Run("overflow-store-observer-is-actual-operand", func(t *testing.T) {
		p := ddiMap4E0B30(os.Getpagesize())
		if p == nil {
			t.Fatal("mmap")
		}
		defer ddiUnmap4E0B30(p, os.Getpagesize())
		base := unsafe.Add(p, 256)
		events := []rawDefaultDamageAccess4E0B30{}
		context := &rawDefaultDamageContext4E0B30{AfterWrite: func(e rawDefaultDamageAccess4E0B30) { events = append(events, e) }}
		pop := rawDefaultDamagePushContext4E0B30(context)
		defer pop()
		for _, width := range []int{1, 2, 4} {
			rawDefaultDamageForeignStore4E0B30(base, 0, width, ^uint64(0), 0x4e0e6d)
			last := events[len(events)-1]
			want := uint64(0xffffffff)
			if width == 1 {
				want = 0xff
			} else if width == 2 {
				want = 0xffff
			}
			if last.Value != want || !last.Write || last.NativeWidth != width || rawParentInitialGoBits4E0B30(base, 0, width) != uintptr(want) {
				t.Fatal("narrow store bit certificate")
			}
		}
	})
	t.Run("same-context-nesting-retains-LIFO-identity", func(t *testing.T) {
		context := &rawDefaultDamageContext4E0B30{}
		outer := rawDefaultDamagePushContext4E0B30(context)
		inner := rawDefaultDamagePushContext4E0B30(context)
		caught := false
		func() { defer func() { caught = recover() != nil }(); outer() }()
		if !caught || rawDefaultDamageCurrentContext4E0B30 != context {
			t.Fatal("out-of-order same-context pop admitted")
		}
		inner()
		outer()
		if rawDefaultDamageCurrentContext4E0B30 != nil || rawDefaultDamageContextStack4E0B30 != nil {
			t.Fatal("nested Go observer context not restored")
		}
	})
	t.Run("protected-load-and-WORD-store-have-no-after-observer", func(t *testing.T) {
		defer ddiFaultMode4E0B30()()
		page := os.Getpagesize()
		p := ddiMap4E0B30(page)
		if p == nil {
			t.Fatal("mmap")
		}
		defer ddiUnmap4E0B30(p, page)
		ddiFill4E0B30(p, page, 0x6a)
		last := unsafe.Add(p, page-1)
		*(*byte)(last) = 0xaa
		reads, writes := 0, 0
		context := &rawDefaultDamageContext4E0B30{AfterRead: func(rawDefaultDamageAccess4E0B30) { reads++ }, AfterWrite: func(rawDefaultDamageAccess4E0B30) { writes++ }}
		pop := rawDefaultDamagePushContext4E0B30(context)
		defer pop()
		readFault := rddRecoverControl4E0B30(func() { rawDefaultDamageForeignRead4E0B30(last, 0, 2, 0x4e0e6d) })
		storeFault := rddRecoverControl4E0B30(func() { rawDefaultDamageForeignStore4E0B30(last, 0, 2, 2, 0x4e0e6d) })
		if readFault == nil || storeFault == nil || reads != 0 || writes != 0 {
			t.Fatal("faulting operand emitted completed observer")
		}
		for _, fault := range []any{readFault, storeFault} {
			addressed, ok := fault.(interface{ Addr() uintptr })
			if !ok || addressed.Addr() < uintptr(last) || addressed.Addr() >= uintptr(last)+2 {
				t.Fatal("actual operand fault address certificate")
			}
		}
		// Record the actual native Go store observation. No original02 fragment
		// is imposed and no platform expectation is inferred from the C probe.
		t.Logf("RAW_DEFAULT_DAMAGE_GO_WORD_FAULT_OBSERVATION first_after=%02x native_original_fragment_equivalence=false", *(*byte)(last))
		*(*byte)(last) = 0xaa
		want := bytes.Repeat([]byte{0x6a}, page)
		want[page-1] = 0xaa
		if !bytes.Equal(ddiSnapshot4E0B30(p, page), want) {
			t.Fatal("whole read/write control buffer changed outside operand")
		}
	})
	t.Log("RAW_DEFAULT_DAMAGE_OBSERVER_CONTRACT_OK controls=4 raw_native_engine_leaves=0 Goheap_GC_qualified=false concurrent_dispatch_qualified=false")
}
func rddRecoverControl4E0B30(fn func()) (value any) {
	defer func() { value = recover() }()
	fn()
	return nil
}
func TestRawDefaultDamageNativeBaselineGate4E0B30(t *testing.T) {
	rddBeforeOnly4E0B30(t)
	if os.Getenv("OPENNOX_RAW_DEFAULT_DAMAGE_BEFORE") == "1" {
		t.Log("RAW_DEFAULT_DAMAGE_NATIVE_BASELINE engine_present=false C_ABI=1 original_engine_comparison_leaves=0 native_fault_process_adapter=false native_whole_memory_equivalence=false")
	} else {
		t.Log("RAW_DEFAULT_DAMAGE_NATIVE_BASELINE engine_present=true C_ABI=2 comparison_leaves_in_this_gate=0 native_whole_memory_equivalence=false")
	}
}
