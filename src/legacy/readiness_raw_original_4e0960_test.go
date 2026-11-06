package legacy

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
	"testing"
	"unsafe"
)

type readinessRawMutation4E0960 struct {
	When, Role, Field string
	Occurrence        int
	Index             *int
	TargetRole        string `json:"target_role"`
	TargetField       string `json:"target_field"`
	TargetIndex       *int   `json:"target_index"`
	Value             any
}

type readinessRawInput4E0960 struct {
	Name      string
	Mutations []readinessRawMutation4E0960
}

type readinessRawAccess4E0960 struct {
	Kind, Role, Field, Address, Base, Instruction string
	Size                                          int
	Index                                         *int
	Value                                         *uint32
	Complete                                      bool
	Mutation                                      *int `json:"source_input_mutation"`
}

type readinessRawOriginal4E0960 struct {
	Name             string
	Roles            map[string]uint32
	InitialHeap      string `json:"initial_heap"`
	FinalHeap        string `json:"final_heap"`
	InitialStack     string `json:"initial_stack"`
	FinalStack       string `json:"final_stack"`
	InitialImage     string `json:"initial_image_sha256"`
	FinalImage       string `json:"final_image_sha256"`
	CallerBefore     string `json:"caller_before"`
	CallerAfter      string `json:"caller_after"`
	Applied          []int  `json:"applied_input_mutations"`
	Accesses, Writes []readinessRawAccess4E0960
	Registers        map[string]uint32
	Calls            []struct {
		Function, Address string
		Args              []uint32
		Depth             int
		External          bool
	}
	Returns []struct {
		Function string
		Result   uint32
		Depth    int
		External bool
	}
	Fault *struct {
		Kind, Address, Instruction string
		Access, Size               int
	}
}

func readinessRawFixture4E0960(t *testing.T, name, encodedSHA, decodedSHA string, target any) {
	t.Helper()
	encoded, err := os.ReadFile("testdata/" + name)
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(encoded)) != encodedSHA {
		t.Fatal("Readiness compressed reference/input digest", name, err)
	}
	z, err := gzip.NewReader(bytes.NewReader(encoded))
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := io.ReadAll(z)
	if err != nil {
		t.Fatal(err)
	}
	if err = z.Close(); err != nil || fmt.Sprintf("%x", sha256.Sum256(decoded)) != decodedSHA {
		t.Fatal("Readiness decoded reference/input digest", name, err)
	}
	if err = json.Unmarshal(decoded, target); err != nil {
		t.Fatal(err)
	}
}

func readinessRawOriginals4E0960(t *testing.T) []readinessRawOriginal4E0960 {
	var rows []readinessRawOriginal4E0960
	readinessRawFixture4E0960(t, "readiness_original_4e0960.json.gz", "140b4b826179835437332b84fe003fcd7bd2c9c0fd1733a077df847e3de13fcf", "b78afdb05c301ba79af06b2c2425adc9ec427a8132a2108c66ccd8bec61ba25b", &rows)
	if len(rows) != 113 {
		t.Fatal("all113 original Readiness records")
	}
	return rows
}

func readinessRawInputs4E0960(t *testing.T) []readinessRawInput4E0960 {
	var rows []readinessRawInput4E0960
	readinessRawFixture4E0960(t, "readiness_independent_inputs_4e0960.json.gz", "cb6330a5622ee63f15cc64b2979d3521fac88df4b3184651517bdc2e88987d4f", "daca8b8c2e4e0ed897a33394dba93657b15cfbfa78dc0bc3ffb3f09f22e8413f", &rows)
	if len(rows) != 113 {
		t.Fatal("all113 independent Readiness inputs")
	}
	return rows
}

func readinessRawHex4E0960(t *testing.T, value string) []byte {
	t.Helper()
	out, err := hex.DecodeString(value)
	if err != nil {
		t.Fatal("complete original hex byte image", err)
	}
	return out
}

func readinessRawAddress4E0960(t *testing.T, value string) uint32 {
	t.Helper()
	n, err := strconv.ParseUint(value, 0, 32)
	if err != nil {
		t.Fatal("original exact address", value, err)
	}
	return uint32(n)
}

func readinessRawCertificate4E0960(t *testing.T, row readinessRawOriginal4E0960) {
	t.Helper()
	heap, stack := readinessRawHex4E0960(t, row.InitialHeap), readinessRawHex4E0960(t, row.InitialStack)
	if len(heap) != 65536 || len(stack) != 65536 {
		t.Fatal("whole original64KiB heap and stack")
	}
	for _, write := range row.Writes {
		address := readinessRawAddress4E0960(t, write.Address)
		if write.Size != 4 || write.Value == nil {
			t.Fatal("original DWORD write value/width")
		}
		if address >= 0x22000000 && uint64(address)+4 <= 0x22010000 {
			if write.Kind != "write" {
				t.Fatal("original private stack write")
			}
			binary.LittleEndian.PutUint32(stack[address-0x22000000:], *write.Value)
		} else if address >= readinessRawPEBase4E0960 && uint64(address)+4 <= 0x21010000 {
			if write.Kind != "stimulus" || write.Mutation == nil {
				t.Fatal("only independent input stimuli may change original heap")
			}
			binary.LittleEndian.PutUint32(heap[address-readinessRawPEBase4E0960:], *write.Value)
		} else {
			t.Fatal("unexpected original memory write domain", write)
		}
	}
	if !bytes.Equal(heap, readinessRawHex4E0960(t, row.FinalHeap)) || !bytes.Equal(stack, readinessRawHex4E0960(t, row.FinalStack)) {
		t.Fatal("complete original ordered-write/heap/stack reconstruction")
	}
	caller := readinessRawHex4E0960(t, row.CallerBefore)
	if len(caller) != 8 || row.CallerBefore != row.CallerAfter || !bytes.Equal(stack[0xe000:0xe008], caller) || binary.LittleEndian.Uint32(caller) != 0x23000100 || binary.LittleEndian.Uint32(caller[4:]) != row.Roles["item"] {
		t.Fatal("complete unchanged original caller and actual root argument")
	}
	if row.InitialImage != "66eb0df5e37c97beb6c9b68fafa5a1d966cd4271347bd6c3453215684ff1a1c5" || row.InitialImage != row.FinalImage {
		t.Fatal("bound unchanged whole original PE image certificate")
	}
	if len(row.Calls) != 1 || row.Calls[0].Function != "readiness" || row.Calls[0].Address != "0x4e0960" || !reflect.DeepEqual(row.Calls[0].Args, []uint32{row.Roles["item"]}) || row.Calls[0].Depth != 0 || row.Calls[0].External {
		t.Fatal("actual root call, zero helpers/external substitutions")
	}
	if row.Fault == nil {
		if len(row.Returns) != 1 || row.Returns[0].Function != "readiness" || row.Returns[0].Depth != 0 || row.Returns[0].External || row.Returns[0].Result != row.Registers["eax"] || row.Registers["eip"] != 0x23000100 || row.Registers["esp"] != 0x2200e004 || row.Registers["ebx"] != 0x12345678 || row.Registers["esi"] != 0x23456789 || row.Registers["edi"] != 0x3456789a || row.Registers["ebp"] != 0x456789ab {
			t.Fatal("complete original successful signed DWORD return/register/stack certificate")
		}
	} else {
		last := row.Accesses[len(row.Accesses)-1]
		if len(row.Returns) != 0 || row.Fault.Kind != "memory" || (row.Fault.Access != 19 && row.Fault.Access != 23) || row.Fault.Size != 4 || last.Kind != "read" || last.Complete || last.Value != nil || last.Address != row.Fault.Address || last.Instruction != row.Fault.Instruction || row.Registers["eip"] != readinessRawAddress4E0960(t, row.Fault.Instruction) {
			t.Fatal("complete original DWORD memory-fault prefix and EIP")
		}
	}
}

type readinessRawRootRecord4E0960 struct {
	Complete bool
	Item     unsafe.Pointer
	Result   int32
}
type readinessRawObserver4E0960 struct {
	m        *readinessRawMemory4E0960
	initial  readinessRawInitialRow4E0960
	input    readinessRawInput4E0960
	accesses []readinessRawAccess4E0960
	roots    []readinessRawRootRecord4E0960
	reads    map[string]int
	applied  []int
	auto     bool
}

func newReadinessRawObserver4E0960(m *readinessRawMemory4E0960, initial readinessRawInitialRow4E0960, input readinessRawInput4E0960) *readinessRawObserver4E0960 {
	if initial.Name != input.Name {
		m.t.Fatal("independent initial/input name binding")
	}
	return &readinessRawObserver4E0960{m: m, initial: initial, input: input, reads: make(map[string]int), auto: true}
}

func readinessRawField4E0960(t *testing.T, role, field string, index *int) (kind, slot int, offset uint32) {
	t.Helper()
	switch field {
	case "init-data":
		if role != "item" {
			t.Fatal("undeclared InitData base")
		}
		return 0, 0, 692
	case "class":
		if role != "item" {
			t.Fatal("undeclared class base")
		}
		return 1, 0, 8
	case "slot":
		if role != "table" && role != "alternate_table" {
			t.Fatal("undeclared modifier table base")
		}
		if index == nil || *index < 0 || *index > 3 {
			t.Fatal("declared modifier slot index required")
		}
		return 2, *index, 4 * uint32(*index)
	case "callback":
		return 3, 0, 40
	case "strength":
		return 4, 0, 48
	default:
		t.Fatal("undeclared Readiness field", role, field)
		return
	}
}

func (o *readinessRawObserver4E0960) mutate(index int) {
	t := o.m.t
	for _, old := range o.applied {
		if old == index {
			t.Fatal("independent mutation applied twice")
		}
	}
	mutation := o.input.Mutations[index]
	kind, slot, offset := readinessRawField4E0960(t, mutation.TargetRole, mutation.TargetField, mutation.TargetIndex)
	role, ok := o.initial.Roles[mutation.TargetRole]
	if !ok {
		t.Fatal("undeclared mutation target role")
	}
	var word uint32
	switch value := mutation.Value.(type) {
	case float64:
		if value < 0 || value > 0xffffffff || float64(uint32(value)) != value {
			t.Fatal("independent DWORD mutation")
		}
		word = uint32(value)
	case string:
		switch value {
		case "null":
			word = 0
		case "bad":
			word = 0x41000000
		default:
			var ok bool
			word, ok = o.initial.Roles[value]
			if !ok {
				t.Fatal("undeclared independent pointer stimulus")
			}
		}
	default:
		t.Fatal("independent mutation value type")
	}
	pointer := unsafe.Add(o.m.fromPE(role), o.m.fieldOffset(kind, slot))
	if kind == 0 || kind == 2 || kind == 3 {
		*(*unsafe.Pointer)(pointer) = o.m.fromPE(word)
	} else {
		*(*uint32)(pointer) = word
	}
	record := readinessRawAccess4E0960{Kind: "stimulus", Role: mutation.TargetRole, Field: mutation.TargetField, Address: fmt.Sprintf("%#x", role+offset), Size: 4, Value: &word, Complete: true, Mutation: &index}
	if kind == 2 {
		copyIndex := slot
		record.Index = &copyIndex
	}
	o.accesses = append(o.accesses, record)
	o.applied = append(o.applied, index)
}

func (o *readinessRawObserver4E0960) observe(complete bool, base unsafe.Pointer, kind, index int, value readinessRawValue4E0960) {
	t := o.m.t
	role, field, offset := "item", "init-data", uint32(692)
	var slot *int
	switch kind {
	case 0:
	case 1:
		field, offset = "class", 8
	case 2:
		role, field, offset = "table", "slot", 4*uint32(index)
		copyIndex := index
		slot = &copyIndex
	case 3:
		role, field, offset = "modifier", "callback", 40
	case 4:
		role, field, offset = "modifier", "strength", 48
	default:
		t.Fatal("undeclared actual typed Readiness read")
	}
	pe := o.m.toPE(uintptr(base))
	address := fmt.Sprintf("%#x", pe+offset)
	if !complete {
		o.accesses = append(o.accesses, readinessRawAccess4E0960{Kind: "read", Role: role, Field: field, Address: address, Base: fmt.Sprintf("%#x", pe), Size: 4, Index: slot})
		return
	}
	last := &o.accesses[len(o.accesses)-1]
	if last.Kind != "read" || last.Complete || last.Address != address || last.Role != role || last.Field != field || !reflect.DeepEqual(last.Index, slot) {
		t.Fatal("reordered or duplicated native Readiness field read")
	}
	word := value.Word
	if kind == 0 || kind == 2 || kind == 3 {
		word = o.m.toPE(uintptr(value.Pointer))
	}
	last.Complete, last.Value = true, &word
	key := role + "/" + field + "/none"
	if slot != nil {
		key = role + "/" + field + "/" + strconv.Itoa(*slot)
	}
	o.reads[key]++
	if o.auto {
		for i, mutation := range o.input.Mutations {
			if mutation.When == "after-read" && mutation.Role == role && mutation.Field == field && reflect.DeepEqual(mutation.Index, slot) && mutation.Occurrence == o.reads[key] {
				o.mutate(i)
			}
		}
	}
}

func (o *readinessRawObserver4E0960) root(complete bool, item unsafe.Pointer, result int32) {
	o.roots = append(o.roots, readinessRawRootRecord4E0960{complete, item, result})
}

func (o *readinessRawObserver4E0960) install() func() {
	oldRead, oldRoot := readinessRawObserve4E0960, readinessRawObserveRoot4E0960
	readinessRawObserve4E0960, readinessRawObserveRoot4E0960 = o.observe, o.root
	return func() { readinessRawObserve4E0960, readinessRawObserveRoot4E0960 = oldRead, oldRoot }
}

func (o *readinessRawObserver4E0960) compareAccesses(want []readinessRawAccess4E0960) {
	expected := append([]readinessRawAccess4E0960(nil), want...)
	for i := range expected {
		expected[i].Instruction = ""
	}
	if !reflect.DeepEqual(o.accesses, expected) {
		o.m.t.Fatal("all original DWORD/pointer values, widths, bases, slot reloads, stimuli and fault prefix", o.accesses, expected)
	}
}

func (o *readinessRawObserver4E0960) compare(want readinessRawOriginal4E0960, fault bool, result int32) {
	t := o.m.t
	if fault != (want.Fault != nil) || (!fault && uint32(result) != want.Returns[0].Result) {
		t.Fatal("actual signed DWORD return/memory-fault contract", fault, result, want.Fault)
	}
	o.compareAccesses(want.Accesses)
	if !bytes.Equal(o.m.project(), readinessRawHex4E0960(t, want.FinalHeap)) {
		t.Fatal("complete native-to-PE32 final object/descriptor/table heap")
	}
	if !reflect.DeepEqual(o.applied, want.Applied) && !(len(o.applied) == 0 && len(want.Applied) == 0) {
		t.Fatal("all and only independent input stimuli applied", o.applied, want.Applied)
	}
	expectedRoots := 1
	if !fault {
		expectedRoots = 2
	}
	if len(o.roots) != expectedRoots || o.roots[0].Complete || o.roots[0].Item != o.m.fromPE(o.initial.Roles["item"]) {
		t.Fatal("actual native root argument/entry")
	}
	if !fault && (!o.roots[1].Complete || o.roots[1].Item != o.roots[0].Item || o.roots[1].Result != result) {
		t.Fatal("actual native signed root return")
	}
}

func TestReadinessRawOriginalCertificates4E0960(t *testing.T) {
	rows := readinessRawOriginals4E0960(t)
	reads, stimuli, returns := 0, 0, 0
	faults := make(map[string]int)
	for _, row := range rows {
		readinessRawCertificate4E0960(t, row)
		for _, access := range row.Accesses {
			if access.Kind == "read" {
				reads++
			} else if access.Kind == "stimulus" {
				stimuli++
			} else {
				t.Fatal("undeclared original external access")
			}
		}
		returns += len(row.Returns)
		if row.Fault != nil {
			faults[row.Fault.Instruction]++
		}
	}
	if reads != 611 || stimuli != 29 || returns != 93 || !reflect.DeepEqual(faults, map[string]int{"0x4e0965": 3, "0x4e0980": 3, "0x4e0986": 4, "0x4e09a0": 10}) {
		t.Fatal("complete original113 root/read/stimulus/return/fault counts", reads, stimuli, returns, faults)
	}
}

func TestReadinessRawConsumersBefore4E0960(t *testing.T) {
	if os.Getenv("OPENNOX_READINESS_RAW_BEFORE") != "1" {
		return
	}
	if readinessRawABI4E0960() != 1 || readinessRawEngine4E0960 != nil {
		t.Fatal("raw integer C ABI and absent raw engine required before whole consumer freeze")
	}
	if _, err := os.Stat("readiness_raw_native_4e0960.go"); !os.IsNotExist(err) {
		t.Fatal("raw native engine must remain absent before consumer freeze")
	}
}

func TestReadinessRawCTransport4E0960(t *testing.T) {
	initial, originals := readinessRawInitials4E0960(t), readinessRawOriginals4E0960(t)
	for i, row := range initial {
		t.Run(row.Name, func(t *testing.T) {
			m := newReadinessRawMemory4E0960(t, row)
			item := m.fromPE(row.Roles["item"])
			readinessRawTracing4E0960(true)
			defer readinessRawTracing4E0960(false)
			result := readinessRawTransport4E0960(item)
			trace := readinessRawTraces4E0960()
			if uint32(result) != 0x87654321 || len(trace) != 2 || trace[0].Phase != 0 || trace[1].Phase != 1 || trace[0].Function == 0 || trace[0].Function != trace[1].Function || trace[0].Item != uintptr(item) || trace[1].Item != uintptr(item) || trace[1].Result != result {
				t.Fatal("actual native-pointer C stub argument and signed return", trace, result)
			}
			if !bytes.Equal(m.buf, m.before) || !bytes.Equal(m.project(), m.pe) {
				t.Fatal("nonalgorithmic C transport changes no native/PE bytes")
			}
			if originals[i].Name != row.Name {
				t.Fatal("canonical row binding")
			}
			if originals[i].Fault == nil {
				word := originals[i].Returns[0].Result
				if uint32(readinessRawReturnTransport4E0960(int32(word))) != word {
					t.Fatal("every original signed return bit survives actual C/Go scalar transport")
				}
			}
		})
	}
}

func TestReadinessRawFrozenRuntime4E0960(t *testing.T) {
	initial, inputs := readinessRawInitials4E0960(t), readinessRawInputs4E0960(t)
	reads, faults := 0, 0
	for i, row := range initial {
		t.Run(row.Name, func(t *testing.T) {
			m := newReadinessRawMemory4E0960(t, row)
			o := newReadinessRawObserver4E0960(m, row, inputs[i])
			o.auto = false
			defer o.install()()
			for _, read := range row.Reads {
				var got readinessRawValue4E0960
				fault := readinessRawRecover4E0960(t, func() { got = readinessRawRead4E0960(m.fromPE(read.Role), read.Kind, read.Index) })
				reads++
				if fault != (read.Want == nil) {
					t.Fatal("actual frozen runtime field fault domain")
				}
				if fault {
					faults++
					continue
				}
				word := got.Word
				if read.Kind == 0 || read.Kind == 2 || read.Kind == 3 {
					word = m.toPE(uintptr(got.Pointer))
				}
				if word != *read.Want {
					t.Fatal("actual frozen runtime independent native field value")
				}
			}
			if !bytes.Equal(m.buf, m.before) || !bytes.Equal(m.project(), m.pe) {
				t.Fatal("whole initial object memory unchanged by typed runtime reads")
			}
		})
	}
	if reads != 1808 || faults != 52 {
		t.Fatal("all independent frozen runtime field/fault domains", reads, faults)
	}
}

func TestReadinessRawWholeTransport4E0960(t *testing.T) {
	initial, inputs, originals := readinessRawInitials4E0960(t), readinessRawInputs4E0960(t), readinessRawOriginals4E0960(t)
	for _, row := range originals {
		readinessRawCertificate4E0960(t, row)
	}
	for _, profile := range []string{"tozero", "nearest"} {
		for i, row := range initial {
			t.Run(profile+"/"+row.Name, func(t *testing.T) {
				m := newReadinessRawMemory4E0960(t, row)
				want := originals[i]
				if row.Name != want.Name || !bytes.Equal(m.pe, readinessRawHex4E0960(t, want.InitialHeap)) {
					t.Fatal("complete independent initial-to-original heap identity")
				}
				o := newReadinessRawObserver4E0960(m, row, inputs[i])
				defer o.install()()
				item := m.fromPE(row.Roles["item"])
				o.root(false, item, 0)
				var result int32
				// Original records select transports, not root gameplay behavior. Input
				// stimuli are still triggered by actual getter completion/occurrences.
				fault := readinessRawRecover4E0960(t, func() {
					for len(o.accesses) < len(want.Accesses) {
						next := want.Accesses[len(o.accesses)]
						if next.Kind != "read" {
							t.Fatal("stimulus must arise from actual independent read occurrence", next)
						}
						kind, index, _ := readinessRawField4E0960(t, next.Role, next.Field, next.Index)
						readinessRawRead4E0960(m.fromPE(readinessRawAddress4E0960(t, next.Base)), kind, index)
						if len(o.accesses) > len(want.Accesses) {
							t.Fatal("extra whole-transport accesses")
						}
						o.compareAccesses(want.Accesses[:len(o.accesses)])
					}
					if want.Fault == nil {
						result = readinessRawReturnTransport4E0960(int32(want.Returns[0].Result))
						o.root(true, item, result)
					}
				})
				o.compare(want, fault, result)
			})
		}
	}
}

func TestReadinessRawOriginal4E0960(t *testing.T) {
	if os.Getenv("OPENNOX_READINESS_RAW_BEFORE") == "1" {
		return
	}
	if readinessRawABI4E0960() != 2 || readinessRawEngine4E0960 == nil {
		t.Fatal("actual native Readiness raw ABI and engine required")
	}
	initial, inputs, originals := readinessRawInitials4E0960(t), readinessRawInputs4E0960(t), readinessRawOriginals4E0960(t)
	for _, profile := range []string{"tozero", "nearest"} {
		for _, backend := range []string{"Go", "C"} {
			for i, row := range initial {
				t.Run(profile+"/"+backend+"/"+row.Name, func(t *testing.T) {
					m := newReadinessRawMemory4E0960(t, row)
					want := originals[i]
					readinessRawCertificate4E0960(t, want)
					if row.Name != want.Name || !bytes.Equal(m.pe, readinessRawHex4E0960(t, want.InitialHeap)) {
						t.Fatal("whole independent initial-original identity")
					}
					o := newReadinessRawObserver4E0960(m, row, inputs[i])
					defer o.install()()
					var result int32
					if backend == "C" {
						readinessRawTracing4E0960(true)
						defer readinessRawTracing4E0960(false)
					}
					fault := readinessRawRecover4E0960(t, func() { result = readinessRawInvoke4E0960(backend == "C", m.fromPE(row.Roles["item"])) })
					o.compare(want, fault, result)
					if backend == "C" {
						trace := readinessRawTraces4E0960()
						count := 2
						if fault {
							count = 1
						}
						if len(trace) != count || trace[0].Phase != 0 || trace[0].Function != readinessRawRootPointer4E0960() || trace[0].Item != uintptr(m.fromPE(row.Roles["item"])) {
							t.Fatal("actual raw C function and full-width entry argument", trace)
						}
						if !fault && (trace[1].Phase != 1 || trace[1].Function != trace[0].Function || trace[1].Item != trace[0].Item || trace[1].Result != result) {
							t.Fatal("actual raw C signed return transport", trace)
						}
					}
				})
			}
		}
	}
}
