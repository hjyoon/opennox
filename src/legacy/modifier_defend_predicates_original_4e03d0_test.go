package legacy

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"os"
	"reflect"
	"runtime/debug"
	"strconv"
	"testing"
	"unsafe"
)

type modifierPredicateOriginal4E03D0 struct {
	Name         string
	InitialHeap  string  `json:"initial_heap"`
	FinalHeap    string  `json:"final_heap"`
	CallerBefore string  `json:"caller_before"`
	CallerAfter  string  `json:"caller_after"`
	Return       *uint32 `json:"return_eax"`
	Fault        *struct {
		Access               int
		Address, Instruction string
	}
	Writes []struct {
		Address string
		Size    int
		Value   uint32
	}
	Calls []string `json:"actual_entry_calls"`
}

type modifierPredicateAccess4E03D0 struct {
	Kind, Address string
	Size          int
	Value         *uint32
	Complete      bool
	Instruction   string
}

type modifierPredicateBoundary4E03D0 struct {
	Name     string
	Accesses []modifierPredicateAccess4E03D0
	Calls    []string `json:"entry_calls"`
}

func modifierPredicateOriginals4E03D0(t *testing.T) ([]modifierPredicateOriginal4E03D0, []modifierPredicateBoundary4E03D0) {
	t.Helper()
	var rows []modifierPredicateOriginal4E03D0
	var bounds []modifierPredicateBoundary4E03D0
	modifierPredicateFixture4E03D0(t, "modifier_defend_predicates_original_4e03d0.json.gz", "f3cc6af63c5eb0a6b0ced5cf5b9af406cd72b14047562d67222a514020ce9f79", "118b828db9ae95e022339f711ec1a218e5574ab0f4ad3e0c79f93aee09c90fe7", &rows)
	modifierPredicateFixture4E03D0(t, "modifier_defend_predicates_accesses_4e03d0.json.gz", "f8b4b5c700b802200b3d03ed286bfd36f1677384c02e5e3dbd48448087ddba59", "58f0b8982be0a65f4a86748769a64b05bfac9ca79d8f823d635962715f29615b", &bounds)
	if len(rows) != 298 || len(bounds) != len(rows) {
		t.Fatal("complete original predicate observations")
	}
	return rows, bounds
}

func modifierPredicateBytes4E03D0(t *testing.T, value string) []byte {
	t.Helper()
	out, err := hex.DecodeString(value)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func modifierPredicateAddress4E03D0(t *testing.T, value string) uint32 {
	t.Helper()
	out, err := strconv.ParseUint(value, 0, 32)
	if err != nil {
		t.Fatal(err)
	}
	return uint32(out)
}

func modifierPredicateOriginalCertificate4E03D0(t *testing.T, row modifierPredicateOriginal4E03D0, bound modifierPredicateBoundary4E03D0) {
	t.Helper()
	before, after := modifierPredicateBytes4E03D0(t, row.InitialHeap), modifierPredicateBytes4E03D0(t, row.FinalHeap)
	caller, callerAfter := modifierPredicateBytes4E03D0(t, row.CallerBefore), modifierPredicateBytes4E03D0(t, row.CallerAfter)
	if len(before) != 0x10000 || len(after) != len(before) || len(caller) != len(callerAfter) || row.Name != bound.Name || !reflect.DeepEqual(row.Calls, bound.Calls) {
		t.Fatal("complete original allocation/caller/call identities")
	}
	for _, write := range row.Writes {
		address := modifierPredicateAddress4E03D0(t, write.Address)
		var target []byte
		if address >= modifierPredicatePEBase4E03D0 && address+uint32(write.Size) <= modifierPredicatePEBase4E03D0+0x10000 {
			target = before[address-modifierPredicatePEBase4E03D0:]
		} else if address >= 0x2200e000 && address+uint32(write.Size) <= 0x2200e000+uint32(len(caller)) {
			target = caller[address-0x2200e000:]
		} else {
			if address < 0x22000000 || address+uint32(write.Size) > 0x22010000 {
				t.Fatal("original write outside declared heap/stack")
			}
			continue
		}
		if write.Size != 4 {
			t.Fatal("original DWORD write width")
		}
		binary.LittleEndian.PutUint32(target, write.Value)
	}
	if !bytes.Equal(before, after) || !bytes.Equal(caller, callerAfter) {
		t.Fatal("complete original heap/caller write-prefix certificate")
	}
	if (row.Return == nil) != (row.Fault != nil) || (row.Return != nil && *row.Return > 1) {
		t.Fatal("original return/fault terminal state")
	}
	if row.Fault != nil && (len(bound.Accesses) == 0 || bound.Accesses[len(bound.Accesses)-1].Complete || bound.Accesses[len(bound.Accesses)-1].Address != row.Fault.Address || bound.Accesses[len(bound.Accesses)-1].Instruction != row.Fault.Instruction) {
		t.Fatal("complete original faulting access prefix and exact original instruction")
	}
}

func (m *modifierPredicateMemory4E03D0) compareWhole(expectedPE []byte, output uint32) {
	m.t.Helper()
	if len(expectedPE) != len(m.pe) {
		m.t.Fatal("complete original heap required")
	}
	expectedNative := bytes.Clone(m.before)
	for pe, value := range expectedPE {
		if value == m.pe[pe] {
			continue
		}
		native := -1
		for _, slot := range m.slots {
			if !slot.pointer && pe >= slot.pe && pe < slot.pe+4 {
				native = slot.native + pe - slot.pe
			}
		}
		if native < 0 && output >= modifierPredicatePEBase4E03D0 && uint32(pe) >= output-modifierPredicatePEBase4E03D0 && uint32(pe) < output-modifierPredicatePEBase4E03D0+4 {
			native = int(uintptr(m.fromPE(output))-uintptr(m.heap)) + pe - int(output-modifierPredicatePEBase4E03D0)
		}
		if native < 0 || native >= len(expectedNative) {
			m.t.Fatal("original change outside declared scalar/output correspondences", pe)
		}
		expectedNative[native] = value
	}
	if !bytes.Equal(m.buf, expectedNative) || !bytes.Equal(m.project(output), expectedPE) {
		m.t.Fatal("complete native allocation guard and full original heap projection")
	}
}

type modifierPredicateObserver4E03D0 struct {
	m        *modifierPredicateMemory4E03D0
	accesses []modifierPredicateAccess4E03D0
}

func (o *modifierPredicateObserver4E03D0) observe(write, complete bool, base unsafe.Pointer, kind, index int, value modifierPredicateValue4E03D0) {
	pe, ok := o.m.reverse[uintptr(base)]
	if !ok && write && kind == 5 {
		trace := modifierPredicateTraceRecords4E03D0()
		if len(trace) > 0 && trace[len(trace)-1].Phase == 2 && trace[len(trace)-1].Args[5] == uintptr(base) {
			return // Exact active callback's caller-owned scratch, checked below.
		}
	}
	if !ok {
		o.m.t.Fatal("undeclared native memory base", base)
	}
	offset := []uint32{96, 8, 692, 4 * uint32(index), 88, 0}[kind]
	address, action := fmt.Sprintf("%#x", pe+offset), "read"
	if write {
		action = "write"
	}
	word := value.Word
	if !write && complete && (kind == 2 || kind == 3 || kind == 4) {
		word = o.m.toPE(uintptr(value.Pointer))
	}
	if !complete {
		row := modifierPredicateAccess4E03D0{Kind: action, Address: address, Size: 4}
		if write {
			row.Value = &word
		}
		o.accesses = append(o.accesses, row)
		return
	}
	if len(o.accesses) == 0 {
		o.m.t.Fatal("completed access without entry")
	}
	last := &o.accesses[len(o.accesses)-1]
	if last.Complete || last.Kind != action || last.Address != address {
		o.m.t.Fatal("reordered or duplicate native memory access")
	}
	last.Complete, last.Value = true, &word
}

func modifierPredicateRecover4E03D0(t *testing.T, fn func()) (fault bool) {
	t.Helper()
	old := debug.SetPanicOnFault(true)
	defer debug.SetPanicOnFault(old)
	defer func() {
		if failure := recover(); failure != nil {
			if _, ok := failure.(interface{ RuntimeError() }); !ok {
				t.Fatalf("unexpected non-memory failure: %v", failure)
			}
			fault = true
		}
	}()
	fn()
	return
}

func TestModifierDefendPredicatesFrozenRuntime4E03D0(t *testing.T) {
	for _, row := range modifierPredicateInitials4E03D0(t) {
		t.Run(row.Name, func(t *testing.T) {
			m := newModifierPredicateMemory4E03D0(t, row)
			observer := &modifierPredicateObserver4E03D0{m: m}
			old := modifierPredicateObserve4E03D0
			modifierPredicateObserve4E03D0 = observer.observe
			defer func() { modifierPredicateObserve4E03D0 = old }()
			for _, read := range row.Reads {
				kind := modifierPredicateReadKinds4E03D0[read.Kind]
				var got modifierPredicateValue4E03D0
				fault := modifierPredicateRecover4E03D0(t, func() { got = modifierPredicateRead4E03D0(m.fromPE(read.Role), kind, read.Index) })
				if fault != (read.Want == nil) {
					t.Fatal("frozen getter's independent memory domain")
				}
				if !fault {
					value := got.Word
					if kind == 2 || kind == 3 || kind == 4 {
						value = m.toPE(uintptr(got.Pointer))
					}
					if value != *read.Want {
						t.Fatal("frozen getter's independent value")
					}
				}
			}
			if len(observer.accesses) != len(row.Reads) {
				t.Fatal("every getter access observed")
			}
			m.compareWhole(m.pe, row.Output)
			if row.Entry != "scan" {
				const probe = uint32(0xa1b2c3d4)
				fault := modifierPredicateRecover4E03D0(t, func() { modifierPredicateStore4E03D0((*int32)(m.out), probe) })
				valid := row.Output >= modifierPredicatePEBase4E03D0 && row.Output <= modifierPredicatePEBase4E03D0+0x10000-4
				if fault == valid {
					t.Fatal("frozen store's independent memory domain")
				}
				expected := bytes.Clone(m.pe)
				if valid {
					binary.LittleEndian.PutUint32(expected[row.Output-modifierPredicatePEBase4E03D0:], probe)
				}
				m.compareWhole(expected, row.Output)
			}
		})
	}
}

func TestModifierDefendPredicatesNativeTransport4E03D0(t *testing.T) {
	row := modifierPredicateInitials4E03D0(t)[0]
	m := newModifierPredicateMemory4E03D0(t, row)
	args := [4]uintptr{}
	for i := range args {
		args[i] = uintptr(unsafe.Add(m.heap, 0x6000+0x100*i))
	}
	modifierPredicateTracing4E03D0(true)
	defer modifierPredicateTracing4E03D0(false)
	result, actual := modifierPredicateTransportControl4E03D0(m.argument, args, (*int32)(m.out))
	want := [6]uintptr{uintptr(m.argument), args[0], args[1], args[2], args[3], uintptr(m.out)}
	if actual != want || result != 0x2468ace0 {
		t.Fatal("independent real C full-width six-argument callback transport", actual, want, result)
	}
	for _, pointer := range actual {
		if unsafe.Sizeof(uintptr(0)) == 8 && pointer <= 0xffffffff {
			t.Fatal("all six native transport arguments must be above 4 GiB")
		}
	}
	trace := modifierPredicateTraceRecords4E03D0()
	if len(trace) != 2 || trace[0].Phase != 2 || trace[1].Phase != 3 || trace[0].Args != want || trace[1].Args != want || trace[1].After != 0x13579bdf || trace[1].Result != result {
		t.Fatal("actual native callback entry/exit/scratch observation")
	}
	expected := bytes.Clone(m.pe)
	binary.LittleEndian.PutUint32(expected[row.Output-modifierPredicatePEBase4E03D0:], 0x13579bdf)
	m.compareWhole(expected, row.Output)
}

func TestModifierDefendPredicatesWholeTransport4E03D0(t *testing.T) {
	initial, originals := modifierPredicateInitials4E03D0(t), []modifierPredicateOriginal4E03D0(nil)
	originals, bounds := modifierPredicateOriginals4E03D0(t)
	for i := range initial {
		modifierPredicateOriginalCertificate4E03D0(t, originals[i], bounds[i])
	}
	for _, profile := range []string{"tozero", "nearest"} {
		for i, row := range initial {
			t.Run(profile+"/"+row.Name, func(t *testing.T) {
				m := newModifierPredicateMemory4E03D0(t, row)
				if row.Name != originals[i].Name || !bytes.Equal(m.pe, modifierPredicateBytes4E03D0(t, originals[i].InitialHeap)) {
					t.Fatal("complete independent initial memory before engine transport")
				}
				m.compareWhole(modifierPredicateBytes4E03D0(t, originals[i].InitialHeap), row.Output)
				// Replay the independently observed CPU accesses as transport data.
				// No predicate or traversal algorithm chooses any of these actions.
				observer := &modifierPredicateObserver4E03D0{m: m}
				old := modifierPredicateObserve4E03D0
				modifierPredicateObserve4E03D0 = observer.observe
				defer func() { modifierPredicateObserve4E03D0 = old }()
				for _, access := range bounds[i].Accesses {
					address := modifierPredicateAddress4E03D0(t, access.Address)
					var fault bool
					if access.Kind == "write" {
						if access.Value == nil {
							t.Fatal("original write value required")
						}
						fault = modifierPredicateRecover4E03D0(t, func() { modifierPredicateStore4E03D0((*int32)(m.fromPE(address)), *access.Value) })
					} else if access.Kind == "read" {
						kind, index, offset := 0, 0, uint32(96)
						switch access.Instruction {
						case "0x4e03d6", "0x4e0486":
						case "0x4e03fd":
							kind, offset = 1, 8
						case "0x4e0406":
							kind, offset = 2, 692
						case "0x4e0427":
							kind, index = 3, -1
							for _, read := range row.Reads {
								if read.Kind == "slot" && read.Role+4*uint32(read.Index) == address {
									index, offset = read.Index, 4*uint32(read.Index)
								}
							}
							if index < 0 {
								t.Fatal("independent original slot address")
							}
						case "0x4e042d":
							kind, offset = 4, 88
						default:
							t.Fatal("undeclared exact original read instruction", access.Instruction)
						}
						fault = modifierPredicateRecover4E03D0(t, func() { modifierPredicateRead4E03D0(m.fromPE(address-offset), kind, index) })
					} else {
						t.Fatal("undeclared original access kind")
					}
					if access.Size != 4 || fault == access.Complete {
						t.Fatal("actual replayed native access width/fault domain", access)
					}
				}
				m.compareWhole(modifierPredicateBytes4E03D0(t, originals[i].FinalHeap), row.Output)
				expected := append([]modifierPredicateAccess4E03D0(nil), bounds[i].Accesses...)
				for j := range expected {
					expected[j].Instruction = ""
				}
				if !reflect.DeepEqual(observer.accesses, expected) {
					t.Fatal("full original access-prefix transport and complete final output", observer.accesses, expected)
				}
			})
		}
	}
}

func TestModifierDefendPredicatesConsumersBefore4E03D0(t *testing.T) {
	if os.Getenv("OPENNOX_MODIFIER_PREDICATES_BEFORE") != "1" {
		return
	}
	for i := 0; i < 3; i++ {
		if modifierPredicatesABI4E03D0(i) != 1 {
			t.Fatal("full consumer freeze must precede all three raw engines")
		}
	}
	if _, err := os.Stat("modifier_defend_predicates_native_4e03d0.go"); !os.IsNotExist(err) {
		t.Fatal("native predicate/traversal engine must be absent before consumer freeze")
	}
}

func TestModifierDefendPredicatesOriginal4E03D0(t *testing.T) {
	if os.Getenv("OPENNOX_MODIFIER_PREDICATES_BEFORE") == "1" {
		return // Compiled full consumer, zero native engine leaves in this phase.
	}
	for i := 0; i < 3; i++ {
		if modifierPredicatesABI4E03D0(i) != 2 {
			t.Fatal("all three actual public C symbols must have native prototypes")
		}
	}
	initial := modifierPredicateInitials4E03D0(t)
	originals, bounds := modifierPredicateOriginals4E03D0(t)
	for _, profile := range []string{"tozero", "nearest"} {
		for i, row := range initial {
			t.Run(profile+"/C/"+row.Name, func(t *testing.T) {
				m := newModifierPredicateMemory4E03D0(t, row)
				m.addresses[0x4e03f0] = modifierPredicateRootPointer4E03D0(2)
				m.reverse[uintptr(m.addresses[0x4e03f0])] = 0x4e03f0
				observer := &modifierPredicateObserver4E03D0{m: m}
				old := modifierPredicateObserve4E03D0
				modifierPredicateObserve4E03D0 = observer.observe
				defer func() { modifierPredicateObserve4E03D0 = old }()
				modifierPredicateTracing4E03D0(true)
				defer modifierPredicateTracing4E03D0(false)
				which := map[string]int{"inversion": 0, "grip": 1, "scan": 2}[row.Entry]
				var result int32
				fault := modifierPredicateRecover4E03D0(t, func() { result = modifierPredicatesTracedPublic4E03D0(which, m.argument, row.OpaqueArgs, m.out) })
				want, bound := originals[i], bounds[i]
				if fault != (want.Fault != nil) || (!fault && uint32(result) != *want.Return) {
					t.Fatal("complete original return/fault", fault, result, want.Return, want.Fault)
				}
				m.compareWhole(modifierPredicateBytes4E03D0(t, want.FinalHeap), row.Output)
				expectedAccesses := append([]modifierPredicateAccess4E03D0(nil), bound.Accesses...)
				for j := range expectedAccesses {
					expectedAccesses[j].Instruction = ""
				}
				if !reflect.DeepEqual(observer.accesses, expectedAccesses) {
					t.Fatal("ordered original read/write values and faulting access prefix", observer.accesses, expectedAccesses)
				}
				modifierPredicateCompareTrace4E03D0(t, m, row, want, bound, fault)
			})
		}
	}
}

func modifierPredicateCompareTrace4E03D0(t *testing.T, m *modifierPredicateMemory4E03D0, row modifierPredicateInitial4E03D0, want modifierPredicateOriginal4E03D0, bound modifierPredicateBoundary4E03D0, fault bool) {
	t.Helper()
	trace := modifierPredicateTraceRecords4E03D0()
	calls, nested := []string{}, []modifierPredicateTrace4E03D0{}
	for _, record := range trace {
		if record.Phase == 0 || record.Phase == 2 {
			calls = append(calls, fmt.Sprintf("%#x", m.toPE(record.Function)))
		}
		if record.Phase == 2 || record.Phase == 3 {
			nested = append(nested, record)
		}
	}
	if !reflect.DeepEqual(calls, bound.Calls) || len(trace) == 0 || trace[0].Phase != 0 || trace[0].Args != [6]uintptr{uintptr(m.argument), uintptr(row.OpaqueArgs[0]), uintptr(row.OpaqueArgs[1]), uintptr(row.OpaqueArgs[2]), uintptr(row.OpaqueArgs[3]), uintptr(m.out)} {
		t.Fatal("actual public C root and ordered native callback invocations", trace, bound.Calls)
	}
	if !fault && (trace[len(trace)-1].Phase != 1 || uint32(trace[len(trace)-1].Result) != *want.Return) {
		t.Fatal("actual C root return observation")
	}
	var outputs []uint32
	for _, write := range want.Writes {
		if write.Address == "0x2200e010" {
			outputs = append(outputs, write.Value)
		}
	}
	var modifiers []uint32
	for _, access := range bound.Accesses {
		if row.Entry == "scan" && access.Instruction == "0x4e0486" {
			modifiers = append(modifiers, modifierPredicateAddress4E03D0(t, access.Address)-96)
		}
	}
	if len(nested) != 2*len(outputs) || len(modifiers) != len(outputs) {
		t.Fatal("every actual original Grip call and scratch output")
	}
	before := row.OpaqueArgs[2]
	for i, word := range outputs {
		enter, leave := nested[2*i], nested[2*i+1]
		if enter.Phase != 2 || leave.Phase != 3 || enter.Args != leave.Args || enter.Args[0] != uintptr(m.fromPE(modifiers[i])) || enter.Args[1] != uintptr(m.argument) || enter.Args[2] != uintptr(row.OpaqueArgs[1]) || enter.Args[3] != uintptr(row.OpaqueArgs[2]) || enter.Args[4] != uintptr(row.OpaqueArgs[0]) || enter.Args[5] == 0 || uint32(enter.Before) != before || uint32(leave.After) != word || uint32(leave.Result) != word {
			t.Fatal("actual native Grip callback arguments, cached context and original scratch words", enter, leave)
		}
		before = word
	}
}
