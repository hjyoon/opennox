package legacy

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"testing"
	"unsafe"
)

type damagePredicateRawMutation4E1400 struct {
	When, Role, Field string
	Occurrence        int
	Value             any
}

type damagePredicateRawInput4E1400 struct {
	Name, Entry string
	Mutations   []damagePredicateRawMutation4E1400
}

type damagePredicateRawAccess4E1400 struct {
	Kind, Role, Field, Address, Instruction string
	Size                                    int
	Value                                   *uint32
	Complete                                bool
	Mutation                                *int `json:"source_input_mutation"`
}

type damagePredicateRawOriginal4E1400 struct {
	Name, Entry                    string
	Roles                          map[string]uint32
	InitialHeap                    string `json:"initial_heap"`
	FinalHeap                      string `json:"final_heap"`
	InitialStack                   string `json:"initial_stack"`
	FinalStack                     string `json:"final_stack"`
	InitialImage                   string `json:"initial_image_sha256"`
	FinalImage                     string `json:"final_image_sha256"`
	CallerBefore                   string `json:"caller_before"`
	CallerAfter                    string `json:"caller_after"`
	Applied                        []int  `json:"applied_input_mutations"`
	Initializers, Accesses, Writes []damagePredicateRawAccess4E1400
	StackReads                     []damagePredicateRawAccess4E1400 `json:"stack_reads"`
	Registers                      map[string]uint32
	Calls                          []struct {
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

func damagePredicateRawOriginals4E1400(t *testing.T) []damagePredicateRawOriginal4E1400 {
	t.Helper()
	var rows []damagePredicateRawOriginal4E1400
	// Generic checksum/gzip reader is already part of the immutable baseline.
	modifierPredicateFixture4E03D0(t, "damage_predicates_original_4e1400.json.gz", "1148775df123f822dcc2d5b8b2d97b311edba63eeaa3fbfecb7bb07d548247b7", "81f7a5a71486d1ee3904ca9087f72b8aa4bc40d48bfeb670e592e881c97f74f3", &rows)
	if len(rows) != 487 {
		t.Fatal("all487 complete original damage-predicate records")
	}
	return rows
}

func damagePredicateRawInputs4E1400(t *testing.T) []damagePredicateRawInput4E1400 {
	t.Helper()
	var rows []damagePredicateRawInput4E1400
	modifierPredicateFixture4E03D0(t, "damage_predicates_independent_inputs_4e1400.json.gz", "1f96b7f8b86e8fd9f37cd9ee8c71993b8f1c9a2093ca0e5ba9afb0586f6fdac6", "a6f5a269759c38d3e850436c5acea2b61b629a17e18fd2eb6d4c27026971be55", &rows)
	if len(rows) != 487 {
		t.Fatal("all487 independently chosen damage-predicate inputs")
	}
	return rows
}

func damagePredicateRawRecover4E1400(t *testing.T, fn func()) (fault bool) {
	t.Helper()
	previous := debug.SetPanicOnFault(true)
	defer debug.SetPanicOnFault(previous)
	defer func() {
		if failure := recover(); failure != nil {
			err, ok := failure.(runtime.Error)
			if !ok || !strings.Contains(err.Error(), "invalid memory address") {
				t.Fatalf("non-memory failure is not an original memory-fault outcome: %v", failure)
			}
			fault = true
		}
	}()
	fn()
	return
}

func damagePredicateRawEntry4E1400(t *testing.T, entry string) int {
	t.Helper()
	switch entry {
	case "attack":
		return 0
	case "friendly":
		return 1
	default:
		t.Fatal("undeclared original damage predicate", entry)
		return 0
	}
}

func damagePredicateRawHex4E1400(t *testing.T, value string) []byte {
	t.Helper()
	out, err := hex.DecodeString(value)
	if err != nil {
		t.Fatal("complete original hex memory/caller certificate", err)
	}
	return out
}

func damagePredicateRawAddress4E1400(t *testing.T, address string) uint32 {
	t.Helper()
	n, err := strconv.ParseUint(address, 0, 32)
	if err != nil {
		t.Fatal("original PE32 address", address, err)
	}
	return uint32(n)
}

func damagePredicateRawField4E1400(t *testing.T, role, field string) (kind int, offset uint32, width int) {
	t.Helper()
	switch role + "/" + field {
	case "source/class":
		return 0, 8, 4
	case "source/subclass":
		return 1, 12, 1
	case "weapon/class":
		return 2, 8, 4
	case "weapon/subclass":
		return 3, 12, 4
	case "weapon/use":
		return 4, 736, 4
	case "use/flags":
		return 5, 96, 1
	default:
		t.Fatal("undeclared original damage field", role, field)
		return
	}
}

func damagePredicateRawKind4E1400(t *testing.T, kind int) (role, field string, offset uint32, width int) {
	t.Helper()
	roles := [6]string{"source", "source", "weapon", "weapon", "weapon", "use"}
	fields := [6]string{"class", "subclass", "class", "subclass", "use", "flags"}
	if kind < 0 || kind >= len(roles) {
		t.Fatal("undeclared native damage read kind", kind)
	}
	role, field = roles[kind], fields[kind]
	_, offset, width = damagePredicateRawField4E1400(t, role, field)
	return
}

func damagePredicateRawWord4E1400(t *testing.T, data []byte, offset uint32, width int) uint32 {
	t.Helper()
	if uint64(offset)+uint64(width) > uint64(len(data)) || (width != 1 && width != 4) {
		t.Fatal("original BYTE/DWORD memory extent", offset, width)
	}
	if width == 1 {
		return uint32(data[offset])
	}
	return binary.LittleEndian.Uint32(data[offset:])
}

func damagePredicateRawReadIP4E1400(t *testing.T, entry int, access damagePredicateRawAccess4E1400) {
	t.Helper()
	kind, _, _ := damagePredicateRawField4E1400(t, access.Role, access.Field)
	allowed := map[int][]uint32{
		0: {0x4e140c}, 1: {0x4e141d}, 2: {0x4e142c},
		3: {0x4e1434, 0x4e1456}, 4: {0x4e143d}, 5: {0x4e1443},
	}
	if entry == 1 {
		allowed = map[int][]uint32{2: {0x4e1478}, 3: {0x4e1481}}
	}
	instruction := damagePredicateRawAddress4E1400(t, access.Instruction)
	for _, valid := range allowed[kind] {
		if valid == instruction {
			return
		}
	}
	t.Fatal("original field read instruction identity", entry, access)
}

func damagePredicateRawCertificate4E1400(t *testing.T, row damagePredicateRawOriginal4E1400) {
	t.Helper()
	entry := damagePredicateRawEntry4E1400(t, row.Entry)
	heap, stack := damagePredicateRawHex4E1400(t, row.InitialHeap), damagePredicateRawHex4E1400(t, row.InitialStack)
	if len(heap) != 65536 || len(stack) != 65536 {
		t.Fatal("complete original64KiB heap/stack")
	}
	for _, initializer := range row.Initializers {
		address := damagePredicateRawAddress4E1400(t, initializer.Address)
		if address < damagePredicatePEBase4E1400 || initializer.Value == nil || damagePredicateRawWord4E1400(t, heap, address-damagePredicatePEBase4E1400, initializer.Size) != *initializer.Value {
			t.Fatal("retained complete original initializer value/width", initializer)
		}
	}
	var writes []damagePredicateRawAccess4E1400
	for _, access := range row.Accesses {
		_, offset, width := damagePredicateRawField4E1400(t, access.Role, access.Field)
		address := damagePredicateRawAddress4E1400(t, access.Address)
		if access.Size != width || address != row.Roles[access.Role]+offset {
			t.Fatal("original exact field role/address/width", access)
		}
		if access.Kind == "stimulus" {
			if access.Mutation == nil || access.Value == nil || !access.Complete || address < damagePredicatePEBase4E1400 || uint64(address)+uint64(width) > 0x21010000 {
				t.Fatal("only declared input stimuli write memory", access)
			}
			if width == 1 {
				if *access.Value > 255 {
					t.Fatal("original stimulus BYTE bits")
				}
				heap[address-damagePredicatePEBase4E1400] = byte(*access.Value)
			} else {
				binary.LittleEndian.PutUint32(heap[address-damagePredicatePEBase4E1400:], *access.Value)
			}
			writes = append(writes, access)
		} else if access.Kind == "read" {
			damagePredicateRawReadIP4E1400(t, entry, access)
			if access.Complete {
				if access.Value == nil || address < damagePredicatePEBase4E1400 || damagePredicateRawWord4E1400(t, heap, address-damagePredicatePEBase4E1400, width) != *access.Value {
					t.Fatal("original ordered complete field value", access)
				}
			} else if access.Value != nil {
				t.Fatal("original incomplete field has no value")
			}
		} else {
			t.Fatal("original leaf has no hidden memory write/helper access", access)
		}
	}
	if !reflect.DeepEqual(writes, row.Writes) && !(len(writes) == 0 && len(row.Writes) == 0) {
		t.Fatal("all and only original declared stimulus writes")
	}
	if !bytes.Equal(heap, damagePredicateRawHex4E1400(t, row.FinalHeap)) || !bytes.Equal(stack, damagePredicateRawHex4E1400(t, row.FinalStack)) {
		t.Fatal("complete original heap reconstruction and unchanged stack")
	}
	args, root, end := []uint32{row.Roles["source"], row.Roles["weapon"]}, uint32(0x4e1400), uint32(0x4e1470)
	if entry == 1 {
		args, root, end = []uint32{row.Roles["weapon"]}, 0x4e1470, 0x4e1492
	}
	caller := damagePredicateRawHex4E1400(t, row.CallerBefore)
	if len(caller) != 4*(len(args)+1) || row.CallerBefore != row.CallerAfter || !bytes.Equal(stack[0xe000:0xe000+len(caller)], caller) || binary.LittleEndian.Uint32(caller) != 0x23000100 {
		t.Fatal("complete unchanged original caller words")
	}
	for i, arg := range args {
		if binary.LittleEndian.Uint32(caller[4*(i+1):]) != arg {
			t.Fatal("complete actual original argument order/count")
		}
	}
	for _, read := range row.StackReads {
		address, instruction := damagePredicateRawAddress4E1400(t, read.Address), damagePredicateRawAddress4E1400(t, read.Instruction)
		if read.Kind != "read" || !read.Complete || read.Value == nil || read.Size != 4 || address < 0x22000000 || instruction < root || instruction >= end || damagePredicateRawWord4E1400(t, stack, address-0x22000000, 4) != *read.Value {
			t.Fatal("all original ordered stack read addresses/widths/values/IPs", read)
		}
	}
	if row.InitialImage != "66eb0df5e37c97beb6c9b68fafa5a1d966cd4271347bd6c3453215684ff1a1c5" || row.InitialImage != row.FinalImage {
		t.Fatal("bound unchanged complete original PE image certificate")
	}
	if len(row.Calls) != 1 || row.Calls[0].Function != row.Entry || row.Calls[0].Address != fmt.Sprintf("%#x", root) || !reflect.DeepEqual(row.Calls[0].Args, args) || row.Calls[0].Depth != 0 || row.Calls[0].External {
		t.Fatal("actual two-argument/one-argument original root call and zero helpers")
	}
	if len(row.Registers) != 10 {
		t.Fatal("all original integer register certificates retained")
	}
	for _, name := range []string{"eax", "ecx", "edx", "ebx", "esi", "edi", "ebp", "esp", "eip", "eflags"} {
		if _, ok := row.Registers[name]; !ok {
			t.Fatal("missing original integer register", name)
		}
	}
	if row.Registers["ebx"] != 0x12345678 || row.Registers["esi"] != 0x23456789 || row.Registers["edi"] != 0x3456789a || row.Registers["ebp"] != 0x456789ab {
		t.Fatal("complete original nonvolatile integer registers")
	}
	if row.Fault == nil {
		if len(row.Returns) != 1 || row.Returns[0].Function != row.Entry || row.Returns[0].Depth != 0 || row.Returns[0].External || row.Returns[0].Result > 1 || row.Returns[0].Result != row.Registers["eax"] || row.Registers["eip"] != 0x23000100 || row.Registers["esp"] != 0x2200e004 {
			t.Fatal("canonical original signed int return, stack balance and return IP")
		}
	} else {
		if len(row.Accesses) == 0 {
			t.Fatal("original memory fault needs a field-read prefix")
		}
		last := row.Accesses[len(row.Accesses)-1]
		if len(row.Returns) != 0 || row.Fault.Kind != "memory" || (row.Fault.Access != 19 && row.Fault.Access != 23) || last.Kind != "read" || last.Complete || last.Value != nil || last.Address != row.Fault.Address || last.Instruction != row.Fault.Instruction || last.Size != row.Fault.Size || row.Registers["eip"] != damagePredicateRawAddress4E1400(t, row.Fault.Instruction) || row.Registers["esp"] != 0x2200e000 {
			t.Fatal("original exact memory-fault role/field/width/EIP/register/read prefix")
		}
	}
}

func damagePredicateRawOriginalProofs4E1400(t *testing.T) {
	t.Helper()
	for _, spec := range []struct {
		name, sha string
		control   uint32
	}{
		{"damage_predicates_original_4e1400_tozero.proof.json", "34f56551e531573eaaa9db64f7aa23618a597f7e79ce96123ee6ededc580f940", 0xe3f},
		{"damage_predicates_original_4e1400_nearest.proof.json", "a75e14683e5dc65b0344caebf135e715119ea6e13e8ccc34921f5917bf37e03c", 0x23f},
	} {
		data, err := os.ReadFile("testdata/" + spec.name)
		if err != nil || fmt.Sprintf("%x", sha256.Sum256(data)) != spec.sha {
			t.Fatal("retained independently executed original FPU proof", spec.name, err)
		}
		var proof struct {
			Cases              int
			FPUControl         uint32 `json:"FPU_control"`
			FPUStatusUnchanged bool   `json:"FPU_status_unchanged"`
			AllBytes           bool   `json:"all146_original_bytes_reached"`
			Instructions       int    `json:"instruction_addresses"`
			NoHelpers          bool   `json:"no_helpers_substituted"`
			OriginalWrites     int    `json:"actual_original_memory_writes"`
			Stimuli            int    `json:"declared_input_stimulus_writes"`
			StackReads         int    `json:"stack_read_records"`
			Image              string `json:"original_image_sha256"`
			Inputs             string `json:"independent_inputs_sha256"`
			Output             string `json:"compressed_output_sha256"`
		}
		if err := json.Unmarshal(data, &proof); err != nil || proof.Cases != 487 || proof.FPUControl != spec.control || !proof.FPUStatusUnchanged || !proof.AllBytes || proof.Instructions != 49 || !proof.NoHelpers || proof.OriginalWrites != 0 || proof.Stimuli != 9 || proof.StackReads != 1050 || proof.Image != "66eb0df5e37c97beb6c9b68fafa5a1d966cd4271347bd6c3453215684ff1a1c5" || proof.Inputs != "1f96b7f8b86e8fd9f37cd9ee8c71993b8f1c9a2093ca0e5ba9afb0586f6fdac6" || proof.Output != "1148775df123f822dcc2d5b8b2d97b311edba63eeaa3fbfecb7bb07d548247b7" {
			t.Fatal("complete original two-FPU execution-certificate scope", proof, err)
		}
	}
}

type damagePredicateRawRoot4E1400 struct {
	Complete bool
	Entry    int
	Args     [2]unsafe.Pointer
	Result   int32
}

type damagePredicateRawObserver4E1400 struct {
	m        *damagePredicateMemory4E1400
	initial  damagePredicateInitial4E1400
	input    damagePredicateRawInput4E1400
	accesses []damagePredicateRawAccess4E1400
	roots    []damagePredicateRawRoot4E1400
	reads    map[string]int
	applied  []int
	auto     bool
}

func newDamagePredicateRawObserver4E1400(m *damagePredicateMemory4E1400, initial damagePredicateInitial4E1400, input damagePredicateRawInput4E1400) *damagePredicateRawObserver4E1400 {
	if initial.Name != input.Name || initial.Entry != input.Entry {
		m.t.Fatal("independent damage initial/input entry and name binding")
	}
	return &damagePredicateRawObserver4E1400{m: m, initial: initial, input: input, reads: make(map[string]int), auto: true}
}

func (o *damagePredicateRawObserver4E1400) mutate(index int) {
	t := o.m.t
	if index < 0 || index >= len(o.input.Mutations) {
		t.Fatal("undeclared independent damage stimulus index")
	}
	for _, old := range o.applied {
		if old == index {
			t.Fatal("independent damage stimulus applied twice")
		}
	}
	change := o.input.Mutations[index]
	kind, offset, width := damagePredicateRawField4E1400(t, change.Role, change.Field)
	pe, ok := o.initial.Roles[change.Role]
	if !ok || change.When != "after-read" {
		t.Fatal("undeclared independent stimulus role/timing")
	}
	var word uint32
	switch value := change.Value.(type) {
	case float64:
		if value < 0 || value > 0xffffffff || float64(uint32(value)) != value {
			t.Fatal("exact independent BYTE/DWORD stimulus")
		}
		word = uint32(value)
	case string:
		if kind != 4 {
			t.Fatal("pointer-valued stimulus requires actual UseData field")
		}
		switch value {
		case "null":
			word = 0
		case "bad":
			word = 0x41000000
		default:
			var ok bool
			word, ok = o.initial.Roles[value]
			if !ok {
				t.Fatal("undeclared stimulus pointer role", value)
			}
		}
	default:
		t.Fatal("undeclared stimulus value type")
	}
	if width == 1 {
		word &= 0xff
	}
	address := unsafe.Add(o.m.fromPE(pe), o.m.layout[kind])
	nativeOffset := int(uintptr(address) - uintptr(o.m.heap))
	nativeWidth, bits := width, uint64(word)
	if kind == 4 {
		nativeWidth = int(o.m.layout[6])
		bits = uint64(uintptr(o.m.fromPE(word)))
	}
	if nativeOffset < 0 || nativeOffset+nativeWidth > len(o.m.buf) {
		t.Fatal("declared input stimulus must target readable native storage")
	}
	declared := false
	for _, slot := range o.m.slots {
		if slot.native == nativeOffset && nativeWidth <= slot.width {
			declared = true
		}
	}
	if !declared {
		t.Fatal("stimulus outside independent native field projection")
	}
	// All of this fixture's memory/pointees are C-owned; byte writes preserve
	// unaligned pointer storage and never encode a Go-owned graph pointer.
	for i := 0; i < nativeWidth; i++ {
		o.m.buf[nativeOffset+i] = byte(bits >> (8 * i))
	}
	o.accesses = append(o.accesses, damagePredicateRawAccess4E1400{Kind: "stimulus", Role: change.Role, Field: change.Field, Address: fmt.Sprintf("%#x", pe+offset), Size: width, Value: &word, Complete: true, Mutation: &index})
	o.applied = append(o.applied, index)
}

func (o *damagePredicateRawObserver4E1400) observe(complete bool, base unsafe.Pointer, kind int, value damagePredicateRawValue4E1400) {
	t := o.m.t
	role, field, offset, width := damagePredicateRawKind4E1400(t, kind)
	pe := o.m.toPE(uintptr(base))
	address := fmt.Sprintf("%#x", pe+offset)
	if !complete {
		o.accesses = append(o.accesses, damagePredicateRawAccess4E1400{Kind: "read", Role: role, Field: field, Address: address, Size: width})
		return
	}
	if len(o.accesses) == 0 {
		t.Fatal("actual damage field read lacks before event")
	}
	last := &o.accesses[len(o.accesses)-1]
	if last.Kind != "read" || last.Complete || last.Role != role || last.Field != field || last.Address != address || last.Size != width {
		t.Fatal("reordered/duplicated native damage read completion")
	}
	word := value.Word
	if kind == 4 {
		word = o.m.toPE(uintptr(value.Pointer))
	}
	last.Complete, last.Value = true, &word
	key := role + "/" + field
	o.reads[key]++
	if o.auto {
		for i, change := range o.input.Mutations {
			if change.When == "after-read" && change.Role == role && change.Field == field && change.Occurrence == o.reads[key] {
				o.mutate(i)
			}
		}
	}
}

func (o *damagePredicateRawObserver4E1400) root(complete bool, entry int, args [2]unsafe.Pointer, result int32) {
	o.roots = append(o.roots, damagePredicateRawRoot4E1400{complete, entry, args, result})
}

func (o *damagePredicateRawObserver4E1400) install() func() {
	oldRead, oldRoot := damagePredicateRawObserve4E1400, damagePredicateRawObserveRoot4E1400
	damagePredicateRawObserve4E1400, damagePredicateRawObserveRoot4E1400 = o.observe, o.root
	return func() { damagePredicateRawObserve4E1400, damagePredicateRawObserveRoot4E1400 = oldRead, oldRoot }
}

func (o *damagePredicateRawObserver4E1400) compareAccesses(want []damagePredicateRawAccess4E1400) {
	expected := append([]damagePredicateRawAccess4E1400(nil), want...)
	for i := range expected {
		// Original EIP is retained/validated in the original certificate. Go
		// physical fault PCs/addresses are not the PE instruction namespace.
		expected[i].Instruction = ""
	}
	if !reflect.DeepEqual(o.accesses, expected) && !(len(o.accesses) == 0 && len(expected) == 0) {
		o.m.t.Fatal("all ordered original field values/widths/lazy reads/full pointers/stimuli/fault prefix", o.accesses, expected)
	}
}

func (o *damagePredicateRawObserver4E1400) compare(want damagePredicateRawOriginal4E1400, fault bool, result int32) {
	t := o.m.t
	if fault != (want.Fault != nil) || (!fault && (len(want.Returns) != 1 || uint32(result) != want.Returns[0].Result)) {
		t.Fatal("actual signed raw damage return or memory-fault domain", fault, result, want.Fault)
	}
	o.compareAccesses(want.Accesses)
	if !bytes.Equal(o.m.project(), damagePredicateRawHex4E1400(t, want.FinalHeap)) {
		t.Fatal("complete native-to-PE final64KiB object/payload heap and native guard bytes")
	}
	if !reflect.DeepEqual(o.applied, want.Applied) && !(len(o.applied) == 0 && len(want.Applied) == 0) {
		t.Fatal("all and only independent damage input stimuli applied", o.applied, want.Applied)
	}
	entry := damagePredicateRawEntry4E1400(t, o.initial.Entry)
	args := damagePredicateRawArgs4E1400(entry, o.m.fromPE(o.initial.Roles["source"]), o.m.fromPE(o.initial.Roles["weapon"]))
	count := 2
	if fault {
		count = 1
	}
	if len(o.roots) != count || o.roots[0].Complete || o.roots[0].Entry != entry || o.roots[0].Args != args {
		t.Fatal("actual native root entry and respective two/one full-width arguments")
	}
	if !fault && (!o.roots[1].Complete || o.roots[1].Entry != entry || o.roots[1].Args != args || o.roots[1].Result != result) {
		t.Fatal("actual native root signed return and complete args")
	}
}

func damagePredicateRawCheckCTrace4E1400(t *testing.T, trace []damagePredicateRawTrace4E1400, entry int, function uintptr, source, weapon unsafe.Pointer, fault bool, result int32) {
	t.Helper()
	pointers := damagePredicateRawArgs4E1400(entry, source, weapon)
	args := [2]uintptr{uintptr(pointers[0]), uintptr(pointers[1])}
	argc, count := 2, 2
	if entry == 1 {
		argc = 1
	}
	if fault {
		count = 1
	}
	if function == 0 || len(trace) != count || trace[0].Function != function || trace[0].Entry != entry || trace[0].Argc != argc || trace[0].Phase != 0 || trace[0].Args != args {
		t.Fatal("actual C function, argument count/order and complete native pointer transport", trace, function, args)
	}
	if !fault && (trace[1].Function != function || trace[1].Entry != entry || trace[1].Argc != argc || trace[1].Phase != 1 || trace[1].Args != args || trace[1].Result != result) {
		t.Fatal("actual C signed return and complete exit arguments", trace, result)
	}
}

func TestDamagePredicatesRawOriginalCertificates4E1400(t *testing.T) {
	damagePredicateRawOriginalProofs4E1400(t)
	rows, inputs := damagePredicateRawOriginals4E1400(t), damagePredicateRawInputs4E1400(t)
	reads, stackReads, stimuli, returns := 0, 0, 0, 0
	families, faults, outcomes := make(map[string]int), make(map[string]int), make(map[string]int)
	for i, row := range rows {
		if row.Name != inputs[i].Name || row.Entry != inputs[i].Entry || len(row.Applied) != len(inputs[i].Mutations) {
			t.Fatal("original entry/input binding and all declared stimulus IDs")
		}
		for j, index := range row.Applied {
			if j != index {
				t.Fatal("all independently chosen stimulus indices in order")
			}
		}
		damagePredicateRawCertificate4E1400(t, row)
		families[row.Entry]++
		for _, access := range row.Accesses {
			if access.Kind == "read" {
				reads++
			} else if access.Kind == "stimulus" {
				stimuli++
			}
		}
		stackReads += len(row.StackReads)
		returns += len(row.Returns)
		if row.Fault != nil {
			faults[row.Fault.Instruction]++
		} else {
			outcomes[row.Entry+"/"+strconv.FormatUint(uint64(row.Returns[0].Result), 10)]++
		}
	}
	if reads != 865 || stackReads != 1050 || stimuli != 9 || returns != 442 || !reflect.DeepEqual(families, map[string]int{"attack": 317, "friendly": 170}) || !reflect.DeepEqual(outcomes, map[string]int{"attack/0": 128, "attack/1": 156, "friendly/0": 140, "friendly/1": 18}) || !reflect.DeepEqual(faults, map[string]int{"0x4e140c": 16, "0x4e141d": 2, "0x4e142c": 4, "0x4e1434": 1, "0x4e143d": 1, "0x4e1443": 8, "0x4e1456": 1, "0x4e1478": 8, "0x4e1481": 4}) {
		t.Fatal("complete487 original BYTE/DWORD reads/stimuli/stack/caller/returns/fault domains", reads, stackReads, stimuli, returns, families, outcomes, faults)
	}
	t.Logf("DAMAGE_PREDICATES_RAW_CERTIFICATES=487 reads=%d stack-reads=%d stimuli=%d returns=%d faults=45 helpers=0", reads, stackReads, stimuli, returns)
}

func TestDamagePredicatesRawConsumersBefore4E1400(t *testing.T) {
	if os.Getenv("OPENNOX_DAMAGE_PREDICATES_BEFORE") != "1" {
		return
	}
	if damagePredicateABI4E1400(0) != 1 || damagePredicateABI4E1400(1) != 1 || damagePredicateRawAttackEngine4E1400 != nil || damagePredicateRawFriendlyEngine4E1470 != nil {
		t.Fatal("both original int C ABIs and nil raw engines required before consumer freeze")
	}
	for _, name := range []string{"damage_predicates_raw_native_4e1400.go", "damage_predicates_native_4e1400.go", "damage_predicates_raw_native_4e1470.go", "damage_predicates_native_4e1470.go"} {
		if _, err := os.Stat(name); !os.IsNotExist(err) {
			t.Fatal("both raw engines must remain absent during consumer freeze", name)
		}
	}
}

func TestDamagePredicatesRawCTransport4E1400(t *testing.T) {
	initial, originals := damagePredicateInitials4E1400(t), damagePredicateRawOriginals4E1400(t)
	attacks, friendlies, originalReturns := 0, 0, 0
	for _, word := range []uint32{0, 0x80000000, 0x87654321, 0x7fffffff, 0xffffffff} {
		if uint32(damagePredicateRawReturn4E1400(int32(word))) != word {
			t.Fatal("nonalgorithmic C signed scalar transport bit patterns", word)
		}
	}
	for i, row := range initial {
		t.Run(row.Name, func(t *testing.T) {
			m := newDamagePredicateMemory4E1400(t, row)
			entry := damagePredicateRawEntry4E1400(t, row.Entry)
			source, weapon := m.fromPE(row.Roles["source"]), m.fromPE(row.Roles["weapon"])
			damagePredicateRawTracing4E1400(true)
			defer damagePredicateRawTracing4E1400(false)
			result := damagePredicateRawTransport4E1400(entry, source, weapon)
			want := int32(-2023406815)
			if entry == 1 {
				want = -19088743
			}
			if result != want {
				t.Fatal("separate actual attack/friendly C signed-return stubs", result, want)
			}
			damagePredicateRawCheckCTrace4E1400(t, damagePredicateRawTraces4E1400(), entry, damagePredicateRawStubPointer4E1400(entry), source, weapon, false, result)
			if !bytes.Equal(m.buf, m.before) || !bytes.Equal(m.project(), m.pe) {
				t.Fatal("nonalgorithmic C argument/return transport changes no native/PE byte")
			}
			if originals[i].Name != row.Name || originals[i].Entry != row.Entry {
				t.Fatal("canonical original row/entry binding")
			}
			if originals[i].Fault == nil {
				word := originals[i].Returns[0].Result
				if uint32(damagePredicateRawReturn4E1400(int32(word))) != word {
					t.Fatal("every actual original return survives C/Go scalar transport")
				}
				originalReturns++
			}
			if entry == 0 {
				attacks++
			} else {
				friendlies++
			}
		})
	}
	if attacks != 317 || friendlies != 170 || originalReturns != 442 {
		t.Fatal("all actual two/one-argument stubs and original return scalar echoes", attacks, friendlies, originalReturns)
	}
	t.Logf("DAMAGE_PREDICATES_RAW_C_TRANSPORT attack=%d friendly=%d original-returns=%d scalar-bit-patterns=5 engine-calls=0", attacks, friendlies, originalReturns)
}

func TestDamagePredicatesRawFrozenRuntime4E1400(t *testing.T) {
	initial, inputs := damagePredicateInitials4E1400(t), damagePredicateRawInputs4E1400(t)
	reads, faults := 0, 0
	for i, row := range initial {
		t.Run(row.Name, func(t *testing.T) {
			m := newDamagePredicateMemory4E1400(t, row)
			o := newDamagePredicateRawObserver4E1400(m, row, inputs[i])
			o.auto = false
			defer o.install()()
			for _, read := range row.Reads {
				var got damagePredicateRawValue4E1400
				fault := damagePredicateRawRecover4E1400(t, func() { got = damagePredicateRawRead4E1400(m.fromPE(read.Role), read.Kind) })
				reads++
				if fault != (read.Want == nil) {
					t.Fatal("actual typed runtime partial/unaligned/nil field-fault domain", row.Name, read.Kind)
				}
				if fault {
					faults++
					continue
				}
				word := got.Word
				if read.Kind == 4 {
					word = m.toPE(uintptr(got.Pointer))
				}
				if word != *read.Want {
					t.Fatal("actual runtime independent BYTE/DWORD/full native pointer", row.Name, read.Kind, word, *read.Want)
				}
			}
			if !bytes.Equal(m.buf, m.before) || !bytes.Equal(m.project(), m.pe) {
				t.Fatal("complete native/PE initial heap and guards unchanged by runtime reads")
			}
		})
	}
	if reads != 2922 || faults != 1222 {
		t.Fatal("all independent frozen runtime domains", reads, faults)
	}
	t.Logf("DAMAGE_PREDICATES_RAW_RUNTIME_PROBES=%d unreadable=%d full-pointer-width=true partial-and-unaligned=true", reads, faults)
}

func TestDamagePredicatesRawWholeTransport4E1400(t *testing.T) {
	initial, inputs, originals := damagePredicateInitials4E1400(t), damagePredicateRawInputs4E1400(t), damagePredicateRawOriginals4E1400(t)
	cases, reads, stimuli, faults := 0, 0, 0, 0
	for _, row := range originals {
		damagePredicateRawCertificate4E1400(t, row)
	}
	for _, profile := range []string{"tozero", "nearest"} {
		for i, row := range initial {
			t.Run(profile+"/"+row.Name, func(t *testing.T) {
				m := newDamagePredicateMemory4E1400(t, row)
				want := originals[i]
				if row.Name != want.Name || row.Entry != want.Entry || !bytes.Equal(m.pe, damagePredicateRawHex4E1400(t, want.InitialHeap)) {
					t.Fatal("complete independent initial-to-original heap identity")
				}
				o := newDamagePredicateRawObserver4E1400(m, row, inputs[i])
				defer o.install()()
				entry := damagePredicateRawEntry4E1400(t, row.Entry)
				source, weapon := m.fromPE(row.Roles["source"]), m.fromPE(row.Roles["weapon"])
				args := damagePredicateRawArgs4E1400(entry, source, weapon)
				o.root(false, entry, args, 0)
				var result int32
				// Original records select field transports, not branch behavior.
				// The independent input's stimuli arise from actual read completion.
				fault := damagePredicateRawRecover4E1400(t, func() {
					for len(o.accesses) < len(want.Accesses) {
						next := want.Accesses[len(o.accesses)]
						if next.Kind != "read" {
							t.Fatal("stimulus must arise from independent actual read occurrence")
						}
						kind, offset, _ := damagePredicateRawField4E1400(t, next.Role, next.Field)
						base := damagePredicateRawAddress4E1400(t, next.Address) - offset
						damagePredicateRawRead4E1400(m.fromPE(base), kind)
						if len(o.accesses) > len(want.Accesses) {
							t.Fatal("extra whole-transport field/stimulus accesses")
						}
						o.compareAccesses(want.Accesses[:len(o.accesses)])
					}
					if want.Fault == nil {
						result = damagePredicateRawReturn4E1400(int32(want.Returns[0].Result))
						o.root(true, entry, args, result)
					}
				})
				o.compare(want, fault, result)
				cases++
				if fault {
					faults++
				}
				for _, access := range o.accesses {
					if access.Kind == "read" {
						reads++
					} else if access.Kind == "stimulus" {
						stimuli++
					}
				}
			})
		}
	}
	if cases != 974 || reads != 1730 || stimuli != 18 || faults != 90 {
		t.Fatal("both original-label whole transports actual read/stimulus/fault counts", cases, reads, stimuli, faults)
	}
	t.Logf("DAMAGE_PREDICATES_RAW_WHOLE_TRANSPORT reference-labels=2 cases=%d reads=%d stimuli=%d memory-faults=%d reference-guided=true engine-calls=0", cases, reads, stimuli, faults)
}

func TestDamagePredicatesRawOriginal4E1400(t *testing.T) {
	if os.Getenv("OPENNOX_DAMAGE_PREDICATES_BEFORE") == "1" {
		t.Log("DAMAGE_PREDICATES_RAW_ENGINE_LEAVES=0 before-mode=true")
		return
	}
	if damagePredicateABI4E1400(0) != 2 || damagePredicateABI4E1400(1) != 2 || damagePredicateRawAttackEngine4E1400 == nil || damagePredicateRawFriendlyEngine4E1470 == nil {
		t.Fatal("both actual native raw ABIs and separate attack/friendly engines required")
	}
	initial, inputs, originals := damagePredicateInitials4E1400(t), damagePredicateRawInputs4E1400(t), damagePredicateRawOriginals4E1400(t)
	for _, profile := range []string{"tozero", "nearest"} {
		for _, backend := range []string{"Go", "C"} {
			for i, row := range initial {
				t.Run(profile+"/"+backend+"/"+row.Name, func(t *testing.T) {
					m := newDamagePredicateMemory4E1400(t, row)
					want := originals[i]
					damagePredicateRawCertificate4E1400(t, want)
					if row.Name != want.Name || row.Entry != want.Entry || !bytes.Equal(m.pe, damagePredicateRawHex4E1400(t, want.InitialHeap)) {
						t.Fatal("whole independent initial/original row identity")
					}
					o := newDamagePredicateRawObserver4E1400(m, row, inputs[i])
					defer o.install()()
					entry := damagePredicateRawEntry4E1400(t, row.Entry)
					source, weapon := m.fromPE(row.Roles["source"]), m.fromPE(row.Roles["weapon"])
					if backend == "C" {
						damagePredicateRawTracing4E1400(true)
						defer damagePredicateRawTracing4E1400(false)
					}
					var result int32
					fault := damagePredicateRawRecover4E1400(t, func() { result = damagePredicateRawInvoke4E1400(backend == "C", entry, source, weapon) })
					o.compare(want, fault, result)
					if backend == "C" {
						damagePredicateRawCheckCTrace4E1400(t, damagePredicateRawTraces4E1400(), entry, damagePredicateRawRootPointer4E1400(entry), source, weapon, fault, result)
					}
				})
			}
		}
	}
}
