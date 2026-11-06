package legacy

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"os"
	"reflect"
	"runtime"
	"runtime/debug"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

type regenerationRawMutation4E01D0 struct {
	When, Function, Role, Field string
	Occurrence                  int
	TargetRole                  string `json:"target_role"`
	TargetField                 string `json:"target_field"`
	Value                       any
}

type regenerationRawInput4E01D0 struct {
	Name          string
	Frame, FPS    uint32
	ArmorFlags    uint32 `json:"armor_flags"`
	ServiceReturn uint32 `json:"service_return"`
	Mutations     []regenerationRawMutation4E01D0
}

type regenerationRawAccess4E01D0 struct {
	Kind, Role, Field, Address, Instruction string
	Size                                    int
	Value                                   *uint32
	Complete                                bool
	Mutation                                *int `json:"source_input_mutation"`
}

type regenerationRawOriginal4E01D0 struct {
	Name             string
	InitialHeap      string `json:"initial_heap"`
	FinalHeap        string `json:"final_heap"`
	InitialStack     string `json:"initial_stack"`
	FinalStack       string `json:"final_stack"`
	InitialImage     string `json:"initial_image_sha256"`
	FinalImage       string `json:"final_image_sha256"`
	CallerBefore     string `json:"caller_before"`
	CallerAfter      string `json:"caller_after"`
	Roles            map[string]uint32
	Registers        map[string]uint32
	Applied          []int `json:"applied_input_mutations"`
	Accesses, Writes []regenerationRawAccess4E01D0
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
		Access, Size, Interrupt    int
	}
}

func regenerationRawInputs4E01D0(t *testing.T) []regenerationRawInput4E01D0 {
	t.Helper()
	var rows []regenerationRawInput4E01D0
	modifierPredicateFixture4E03D0(t, "regeneration_raw_independent_inputs_4e01d0.json.gz", "e7a9dc92b938fd023aef24b062e14b0e7127259350ec67cd2bdf8b14c030b713", "ceea31a4b90206d5f2979049f1d28f0aaa1ba7250f6ecf8e825360bbf462fa66", &rows)
	if len(rows) != 198 {
		t.Fatal("all independent Regeneration inputs")
	}
	return rows
}

func regenerationRawOriginals4E01D0(t *testing.T) []regenerationRawOriginal4E01D0 {
	t.Helper()
	var rows []regenerationRawOriginal4E01D0
	modifierPredicateFixture4E03D0(t, "regeneration_raw_original_4e01d0.json.gz", "19aa7f1b50b6958762d06a912e7c82e2adfe055b6b79fc237845c6743fb2554e", "e3bf75e4030f35a57e20155abc2d6f1d18780f0d2a0ba000978d6040d0db6268", &rows)
	if len(rows) != 198 {
		t.Fatal("all complete original Regeneration observations")
	}
	return rows
}

var regenerationRawPEFields4E01D0 = []struct {
	role, field string
	offset      uint32
	size        int
}{
	{"item", "holder", 492, 4}, {"item", "class", 8, 4},
	{"owner", "health", 556, 4}, {"owner", "last-damage", 536, 4},
	{"owner", "flags", 16, 4}, {"effect", "rate", 108, 4},
	{"health", "current", 0, 2}, {"health", "maximum", 4, 2},
	{"world", "frame", 0x84ea04, 4}, {"world", "fps", 0x85b3fc, 4},
}

func regenerationRawFieldKind4E01D0(t *testing.T, role, field string) int {
	t.Helper()
	for kind, spec := range regenerationRawPEFields4E01D0 {
		if spec.role == role && spec.field == field {
			return kind
		}
	}
	t.Fatal("undeclared original raw Regeneration field", role, field)
	return -1
}

func regenerationRawCertificate4E01D0(t *testing.T, row regenerationRawOriginal4E01D0) {
	t.Helper()
	heap := modifierPredicateBytes4E03D0(t, row.InitialHeap)
	stack := modifierPredicateBytes4E03D0(t, row.InitialStack)
	if len(heap) != 65536 || len(stack) != 65536 {
		t.Fatal("complete original heap/stack extents")
	}
	for _, write := range row.Writes {
		address := modifierPredicateAddress4E03D0(t, write.Address)
		if write.Value == nil || (write.Size != 2 && write.Size != 4) {
			t.Fatal("exact original write value/width")
		}
		var destination []byte
		switch {
		case address >= 0x22000000 && uint64(address)+uint64(write.Size) <= 0x22010000:
			if write.Kind != "write" || write.Size != 4 {
				t.Fatal("original stack DWORD write")
			}
			destination = stack[address-0x22000000:]
		case address >= regenerationRawPEBase4E01D0 && uint64(address)+uint64(write.Size) <= 0x21010000:
			if write.Kind != "stimulus" || write.Mutation == nil {
				t.Fatal("only declared stimuli change original heap")
			}
			destination = heap[address-regenerationRawPEBase4E01D0:]
		case address == 0x84ea04 || address == 0x85b3fc:
			if write.Kind != "stimulus" || write.Size != 4 || write.Mutation == nil {
				t.Fatal("declared original world stimulus")
			}
			continue
		default:
			t.Fatal("original write outside declared whole-memory domains", write)
		}
		for j := 0; j < write.Size; j++ {
			destination[j] = byte(*write.Value >> (8 * j))
		}
	}
	if !bytes.Equal(heap, modifierPredicateBytes4E03D0(t, row.FinalHeap)) || !bytes.Equal(stack, modifierPredicateBytes4E03D0(t, row.FinalStack)) {
		t.Fatal("complete original ordered write reconstruction")
	}
	caller := modifierPredicateBytes4E03D0(t, row.CallerBefore)
	if len(caller) != 12 || row.CallerBefore != row.CallerAfter || !bytes.Equal(stack[0xe000:0xe00c], caller) || binary.LittleEndian.Uint32(caller) != 0x23000100 || binary.LittleEndian.Uint32(caller[4:]) != row.Roles["effect"] || binary.LittleEndian.Uint32(caller[8:]) != row.Roles["item"] {
		t.Fatal("complete unchanged original caller certificate")
	}
	for _, hash := range []string{row.InitialImage, row.FinalImage} {
		decoded, err := hex.DecodeString(hash)
		if err != nil || len(decoded) != 32 {
			t.Fatal("original image hash certificate")
		}
	}
	rootReturns := 0
	for _, call := range row.Calls {
		depth := 1
		if call.Function == "regeneration" {
			depth = 0
		}
		if call.Depth != depth || call.External != (call.Function == "armor-flags" || call.Function == "adjust-hp") {
			t.Fatal("original call depth/external boundaries")
		}
	}
	for _, result := range row.Returns {
		depth := 1
		if result.Function == "regeneration" {
			depth = 0
			rootReturns++
		}
		if result.Depth != depth || result.External != (result.Function == "armor-flags" || result.Function == "adjust-hp") {
			t.Fatal("original return depth/external boundaries")
		}
		if result.Function == "regeneration" && result.Result != row.Registers["eax"] {
			t.Fatal("hardware EAX original certificate, without defined void return")
		}
	}
	if row.Fault == nil {
		if rootReturns != 1 || row.Registers["esp"] != 0x2200e004 || row.Registers["ebx"] != 0x12345678 || row.Registers["esi"] != 0x23456789 || row.Registers["edi"] != 0x3456789a || row.Registers["ebp"] != 0x456789ab {
			t.Fatal("complete original successful caller/register certificate")
		}
	} else {
		if rootReturns != 0 {
			t.Fatal("fault cannot have a root return")
		}
		if row.Fault.Kind == "memory" {
			last := row.Accesses[len(row.Accesses)-1]
			if last.Complete || last.Address != row.Fault.Address || last.Size != row.Fault.Size || last.Instruction != row.Fault.Instruction || (row.Fault.Access != 19 && row.Fault.Access != 23) {
				t.Fatal("complete original memory fault prefix")
			}
		} else if row.Fault.Kind != "integer-divide" || row.Fault.Interrupt != 0 || (row.Fault.Instruction != "0x4e0279" && row.Fault.Instruction != "0x4e0287") {
			t.Fatal("exact original divide-fault instruction")
		}
	}
}

type regenerationRawObserver4E01D0 struct {
	m              *regenerationRawMemory4E01D0
	input          regenerationRawInput4E01D0
	initial        regenerationRawInitial4E01D0
	frame, fps     uint32
	accesses       []regenerationRawAccess4E01D0
	calls          []regenerationRawCall4E01D0
	reads, entries map[string]int
	applied        []int
	auto           bool
	arithmetic     int
}

func newRegenerationRawObserver4E01D0(m *regenerationRawMemory4E01D0, input regenerationRawInput4E01D0, initial regenerationRawInitial4E01D0) *regenerationRawObserver4E01D0 {
	if input.Name != initial.Name {
		m.t.Fatal("independent input identity")
	}
	return &regenerationRawObserver4E01D0{m: m, input: input, initial: initial, frame: input.Frame, fps: input.FPS, reads: make(map[string]int), entries: make(map[string]int), auto: true}
}

func (o *regenerationRawObserver4E01D0) mutate(index int) {
	t := o.m.t
	for _, old := range o.applied {
		if old == index {
			t.Fatal("input mutation applied twice")
		}
	}
	mutation := o.input.Mutations[index]
	role, field := mutation.TargetRole, mutation.TargetField
	if role == "" {
		role = mutation.Role
	}
	if field == "" {
		field = mutation.Field
	}
	kind := regenerationRawFieldKind4E01D0(t, role, field)
	spec := regenerationRawPEFields4E01D0[kind]
	var value uint32
	switch input := mutation.Value.(type) {
	case float64:
		if input < 0 || input > 0xffffffff || input != float64(uint32(input)) {
			t.Fatal("independent uint32 stimulus")
		}
		value = uint32(input)
	case string:
		switch input {
		case "null":
			value = 0
		case "bad":
			value = 0x41000000
		default:
			var ok bool
			value, ok = o.initial.Roles[input]
			if !ok {
				t.Fatal("undeclared stimulus pointer role")
			}
		}
	default:
		t.Fatal("independent stimulus input type")
	}
	if spec.size == 2 {
		value &= 0xffff
	}
	address := spec.offset
	if kind >= 8 {
		if kind == 8 {
			o.frame = value
		} else {
			o.fps = value
		}
	} else {
		base := o.initial.Roles[role]
		address += base
		pointer := unsafe.Add(o.m.fromPE(base), o.m.layout[kind])
		if kind == 0 || kind == 2 {
			*(*unsafe.Pointer)(pointer) = o.m.fromPE(value)
		} else if spec.size == 2 {
			*(*uint16)(pointer) = uint16(value)
		} else {
			*(*uint32)(pointer) = value
		}
	}
	o.accesses = append(o.accesses, regenerationRawAccess4E01D0{Kind: "stimulus", Role: role, Field: field, Address: fmt.Sprintf("%#x", address), Size: spec.size, Value: &value, Complete: true, Mutation: &index})
	o.applied = append(o.applied, index)
}

func (o *regenerationRawObserver4E01D0) read(complete bool, base unsafe.Pointer, kind int, value regenerationRawValue4E01D0) {
	t := o.m.t
	if kind < 0 || kind >= len(regenerationRawPEFields4E01D0) {
		t.Fatal("observed undeclared typed field")
	}
	spec := regenerationRawPEFields4E01D0[kind]
	address := spec.offset
	if kind < 8 {
		address += o.m.toPE(uintptr(base))
	}
	if !complete {
		o.accesses = append(o.accesses, regenerationRawAccess4E01D0{Kind: "read", Role: spec.role, Field: spec.field, Address: fmt.Sprintf("%#x", address), Size: spec.size})
		return
	}
	last := &o.accesses[len(o.accesses)-1]
	if last.Kind != "read" || last.Complete || last.Address != fmt.Sprintf("%#x", address) || last.Size != spec.size {
		t.Fatal("reordered native Regeneration read")
	}
	word := value.Word
	if kind == 0 || kind == 2 {
		word = o.m.toPE(uintptr(value.Pointer))
	}
	last.Complete, last.Value = true, &word
	key := spec.role + "/" + spec.field
	o.reads[key]++
	if o.auto {
		for i, mutation := range o.input.Mutations {
			if mutation.When == "after-read" && mutation.Role == spec.role && mutation.Field == spec.field && mutation.Occurrence == o.reads[key] {
				o.mutate(i)
			}
		}
	}
}

func (o *regenerationRawObserver4E01D0) call(call regenerationRawCall4E01D0) {
	if !call.Complete {
		o.entries[call.Kind]++
		if o.auto {
			for i, mutation := range o.input.Mutations {
				if mutation.When == "before-call" && mutation.Function == call.Kind && mutation.Occurrence == o.entries[call.Kind] {
					o.mutate(i)
				}
			}
		}
	}
	o.calls = append(o.calls, call)
}

func (o *regenerationRawObserver4E01D0) install() func() {
	oldRead, oldCall, oldArithmetic := regenerationRawObserveRead4E01D0, regenerationRawObserveCall4E01D0, regenerationRawObserveArithmetic4E01D0
	oldFrame, oldFPS := regenerationRawFrame4E01D0, regenerationRawFPS4E01D0
	oldArmor, oldAdjust, oldHP := regenerationRawArmorService4E01D0, regenerationRawAdjustService4E01D0, server.ObserveUnitHPRead4EE780
	regenerationRawObserveRead4E01D0, regenerationRawObserveCall4E01D0 = o.read, o.call
	regenerationRawObserveArithmetic4E01D0 = func(complete bool, stage int, dividend, divisor, result uint32) {
		if complete {
			o.arithmetic = 0
		} else {
			o.arithmetic = stage + 1
		}
	}
	regenerationRawFrame4E01D0 = func() uint32 { return o.frame }
	regenerationRawFPS4E01D0 = func() uint32 { return o.fps }
	regenerationRawArmorService4E01D0 = func(*server.Object) uint32 { return o.input.ArmorFlags }
	regenerationRawAdjustService4E01D0 = func(*server.Object, int32) {}
	server.ObserveUnitHPRead4EE780 = func(read server.UnitHPRead4EE780) {
		kind, base, value := 2, unsafe.Pointer(read.Unit), regenerationRawValue4E01D0{Pointer: unsafe.Pointer(read.ValueHealth)}
		if read.Field == 1 {
			kind, base, value = 6, unsafe.Pointer(read.Health), regenerationRawValue4E01D0{Word: uint32(read.Value)}
		} else if read.Field == 2 {
			kind, base, value = 7, unsafe.Pointer(read.Health), regenerationRawValue4E01D0{Word: uint32(read.Value)}
		} else if read.Field != 0 {
			o.m.t.Fatal("undeclared actual HP helper read")
		}
		o.read(read.Complete, base, kind, value)
	}
	return func() {
		regenerationRawObserveRead4E01D0, regenerationRawObserveCall4E01D0, regenerationRawObserveArithmetic4E01D0 = oldRead, oldCall, oldArithmetic
		regenerationRawFrame4E01D0, regenerationRawFPS4E01D0 = oldFrame, oldFPS
		regenerationRawArmorService4E01D0, regenerationRawAdjustService4E01D0, server.ObserveUnitHPRead4EE780 = oldArmor, oldAdjust, oldHP
	}
}

func regenerationRawCapture4E01D0(t *testing.T, fn func()) (failure runtime.Error) {
	t.Helper()
	old := debug.SetPanicOnFault(true)
	defer debug.SetPanicOnFault(old)
	defer func() {
		if caught := recover(); caught != nil {
			var ok bool
			failure, ok = caught.(runtime.Error)
			if !ok {
				t.Fatalf("unexpected non-runtime failure: %v", caught)
			}
		}
	}()
	fn()
	return
}

func (o *regenerationRawObserver4E01D0) compareAccesses(want []regenerationRawAccess4E01D0) {
	expected := append([]regenerationRawAccess4E01D0(nil), want...)
	for i := range expected {
		expected[i].Instruction = ""
	}
	if !reflect.DeepEqual(o.accesses, expected) {
		o.m.t.Fatal("complete ordered Regeneration values/widths/stimuli/fault prefix", o.accesses, expected)
	}
}

func (o *regenerationRawObserver4E01D0) compare(want regenerationRawOriginal4E01D0, failure runtime.Error) {
	t := o.m.t
	if (failure != nil) != (want.Fault != nil) {
		t.Fatal("complete native return/fault agreement", failure, want.Fault)
	}
	if failure != nil {
		divide := strings.Contains(failure.Error(), "integer divide by zero")
		if want.Fault.Kind == "integer-divide" {
			stage := 1
			if want.Fault.Instruction == "0x4e0287" {
				stage = 2
			}
			if !divide || o.arithmetic != stage {
				t.Fatal("actual unsigned divide-fault stage", failure, o.arithmetic, want.Fault)
			}
		} else if divide || o.arithmetic != 0 {
			t.Fatal("actual native memory fault required", failure, o.arithmetic)
		}
	}
	o.compareAccesses(want.Accesses)
	frame, fps := o.input.Frame, o.input.FPS
	for _, access := range want.Accesses {
		if access.Kind == "stimulus" && access.Role == "world" {
			if access.Field == "frame" {
				frame = *access.Value
			} else if access.Field == "fps" {
				fps = *access.Value
			} else {
				t.Fatal("undeclared world stimulus")
			}
		}
	}
	if o.frame != frame || o.fps != fps {
		t.Fatal("complete declared native world state", o.frame, o.fps, frame, fps)
	}
	if !bytes.Equal(o.m.project(), modifierPredicateBytes4E03D0(t, want.FinalHeap)) {
		t.Fatal("complete native-to-PE32 final heap")
	}
	if !reflect.DeepEqual(o.applied, want.Applied) && !(len(o.applied) == 0 && len(want.Applied) == 0) {
		t.Fatal("all and only declared input mutations applied", o.applied, want.Applied)
	}
	entered, returned := 0, 0
	for _, call := range o.calls {
		if !call.Complete {
			if entered >= len(want.Calls) {
				t.Fatal("extra native call")
			}
			old := want.Calls[entered]
			if call.Kind != old.Function {
				t.Fatal("native call identity/order", call, old)
			}
			for i, arg := range old.Args {
				expected := uintptr(0)
				if call.Kind == "adjust-hp" && i == 1 {
					expected = uintptr(arg)
				} else {
					expected = uintptr(o.m.fromPE(arg))
				}
				if call.Args[i] != expected {
					t.Fatal("native full-width call arguments", call, old)
				}
			}
			if len(old.Args) == 1 && call.Args[1] != 0 {
				t.Fatal("unexpected helper second argument")
			}
			if call.Kind == "regeneration" && call.Function == 0 {
				t.Fatal("actual public raw C identity")
			}
			entered++
			continue
		}
		if returned >= len(want.Returns) || call.Kind != want.Returns[returned].Function {
			t.Fatal("native return order", call)
		}
		expected := want.Returns[returned].Result
		if call.Kind == "maximum-hp" || call.Kind == "current-hp" {
			if call.Word != uint32(uint16(expected)) {
				t.Fatal("actual HP WORD return", call, expected)
			}
		} else if call.Kind == "armor-flags" && call.Word != expected {
			t.Fatal("actual armor DWORD return", call, expected)
		}
		returned++ // Root and adjustment have no defined native return value.
	}
	if entered != len(want.Calls) || returned != len(want.Returns) {
		t.Fatal("every original call/return consumed", entered, returned)
	}
}

func TestRegenerationRawConsumersBefore4E01D0(t *testing.T) {
	if os.Getenv("OPENNOX_REGENERATION_RAW_BEFORE") != "1" {
		return
	}
	if regenerationRawABI4E01D0() != 1 || regenerationRawEngine4E01D0 != nil {
		t.Fatal("old integer C and absent raw native engine required before consumer freeze")
	}
	if _, err := os.Stat("regeneration_raw_native_4e01d0.go"); !os.IsNotExist(err) {
		t.Fatal("raw native engine absent before full consumer freeze")
	}
}

func TestRegenerationRawOriginalCertificates4E01D0(t *testing.T) {
	rows := regenerationRawOriginals4E01D0(t)
	reads, stimuli := 0, 0
	calls := make(map[string]int)
	faults := make(map[string]int)
	for _, row := range rows {
		regenerationRawCertificate4E01D0(t, row)
		for _, access := range row.Accesses {
			if access.Kind == "read" {
				reads++
			} else if access.Kind == "stimulus" {
				stimuli++
			} else {
				t.Fatal("undeclared external access")
			}
		}
		for _, call := range row.Calls {
			calls[call.Function]++
		}
		if row.Fault != nil {
			faults[row.Fault.Kind+"/"+row.Fault.Instruction]++
		}
	}
	if reads != 2822 || stimuli != 22 || !reflect.DeepEqual(calls, map[string]int{"regeneration": 198, "maximum-hp": 341, "current-hp": 175, "armor-flags": 61, "adjust-hp": 97}) || !reflect.DeepEqual(faults, map[string]int{"memory/0x4e01df": 1, "memory/0x4e01ed": 1, "memory/0x4e0242": 2, "memory/0x4ee792": 1, "memory/0x4ee7b2": 4, "integer-divide/0x4e0279": 3, "integer-divide/0x4e0287": 22}) {
		t.Fatal("all complete original read/stimulus/call/fault contracts", reads, stimuli, calls, faults)
	}
}

func TestRegenerationRawCTransport4E01D0(t *testing.T) {
	row := regenerationRawInitials4E01D0(t)[0]
	m := newRegenerationRawMemory4E01D0(t, row)
	effect, item := m.fromPE(row.Roles["effect"]), m.fromPE(row.Roles["item"])
	if got := regenerationRawTransportControl4E01D0(effect, item); got != [2]uintptr{uintptr(effect), uintptr(item)} {
		t.Fatal("compiled two-native-pointer void C transport", got)
	}
	if !bytes.Equal(m.buf, m.before) || !bytes.Equal(m.project(), m.pe) {
		t.Fatal("C transport changes no native or PE32 bytes")
	}
}

func TestRegenerationRawFrozenRuntime4E01D0(t *testing.T) {
	initial, inputs := regenerationRawInitials4E01D0(t), regenerationRawInputs4E01D0(t)
	faults := 0
	for i, row := range initial {
		t.Run(row.Name, func(t *testing.T) {
			m := newRegenerationRawMemory4E01D0(t, row)
			o := newRegenerationRawObserver4E01D0(m, inputs[i], row)
			o.auto = false
			defer o.install()()
			for _, read := range row.Reads {
				var got regenerationRawValue4E01D0
				failure := regenerationRawCapture4E01D0(t, func() { got = regenerationRawRead4E01D0(m.fromPE(read.Role), read.Kind) })
				if (failure != nil) != (read.Want == nil) {
					t.Fatal("independent frozen getter fault domain")
				}
				if failure != nil {
					faults++
					continue
				}
				word := got.Word
				if read.Kind == 0 || read.Kind == 2 {
					word = m.toPE(uintptr(got.Pointer))
				}
				if word != *read.Want {
					t.Fatal("independent frozen getter value")
				}
			}
			if !bytes.Equal(m.buf, m.before) || !bytes.Equal(m.project(), m.pe) {
				t.Fatal("independent whole heaps unchanged during frozen getters")
			}
		})
	}
	if faults != 36 {
		t.Fatal("all independent native typed-getter fault domains", faults)
	}
}

// Replay only transports specified by the complete original records. This
// deliberately contains no root gameplay algorithm and earns no engine credit.
func regenerationRawReplay4E01D0(o *regenerationRawObserver4E01D0, want regenerationRawOriginal4E01D0) runtime.Error {
	t := o.m.t
	root := regenerationRawCallBegin4E01D0("regeneration", uintptr(o.m.fromPE(o.initial.Roles["effect"])), uintptr(o.m.fromPE(o.initial.Roles["item"])))
	entered := 1
	invoke := func() {
		if entered >= len(want.Calls) {
			t.Fatal("missing recorded helper transport")
		}
		call := want.Calls[entered]
		entered++
		first := (*server.Object)(o.m.fromPE(call.Args[0]))
		switch call.Function {
		case "maximum-hp":
			regenerationRawMaximum4E01D0(first)
		case "current-hp":
			regenerationRawCurrent4E01D0(first)
		case "armor-flags":
			regenerationRawArmor4E01D0(first)
		case "adjust-hp":
			regenerationRawAdjust4E01D0(first, int32(call.Args[1]))
		default:
			t.Fatal("unknown original helper transport")
		}
	}
	failure := regenerationRawCapture4E01D0(t, func() {
		for len(o.accesses) < len(want.Accesses) {
			next := want.Accesses[len(o.accesses)]
			helper := strings.HasPrefix(next.Instruction, "0x4ee7")
			if helper {
				for entered < len(want.Calls) && want.Calls[entered].Function == "armor-flags" {
					invoke()
				}
				invoke()
			} else if next.Kind == "stimulus" && (next.Instruction == "0x415c70" || next.Instruction == "0x4ee460") {
				invoke()
			} else {
				if next.Kind != "read" {
					t.Fatal("recorded stimulus was not triggered by actual read/call occurrence", next)
				}
				kind := regenerationRawFieldKind4E01D0(t, next.Role, next.Field)
				var base unsafe.Pointer
				if kind < 8 {
					base = o.m.fromPE(modifierPredicateAddress4E03D0(t, next.Address) - regenerationRawPEFields4E01D0[kind].offset)
				}
				regenerationRawRead4E01D0(base, kind)
			}
			if len(o.accesses) > len(want.Accesses) {
				t.Fatal("extra transport accesses")
			}
			o.compareAccesses(want.Accesses[:len(o.accesses)])
		}
		if want.Fault != nil && want.Fault.Kind == "integer-divide" {
			stage := 0
			if want.Fault.Instruction == "0x4e0287" {
				stage = 1
			}
			regenerationRawDivide4E01D0(stage, want.Registers["eax"], want.Registers["ecx"])
		}
		for entered < len(want.Calls) {
			invoke()
		}
		if want.Fault == nil {
			regenerationRawCallEnd4E01D0(root, 0)
		}
	})
	return failure
}

func TestRegenerationRawWholeTransport4E01D0(t *testing.T) {
	initial, inputs, originals := regenerationRawInitials4E01D0(t), regenerationRawInputs4E01D0(t), regenerationRawOriginals4E01D0(t)
	for _, row := range originals {
		regenerationRawCertificate4E01D0(t, row)
	}
	for _, profile := range []string{"tozero", "nearest"} {
		for i, row := range initial {
			t.Run(profile+"/"+row.Name, func(t *testing.T) {
				m := newRegenerationRawMemory4E01D0(t, row)
				want := originals[i]
				if row.Name != want.Name || !bytes.Equal(m.pe, modifierPredicateBytes4E03D0(t, want.InitialHeap)) {
					t.Fatal("complete independent initial-original heap identity")
				}
				o := newRegenerationRawObserver4E01D0(m, inputs[i], row)
				defer o.install()()
				failure := regenerationRawReplay4E01D0(o, want)
				o.compare(want, failure)
			})
		}
	}
}

func TestRegenerationRawOriginal4E01D0(t *testing.T) {
	if os.Getenv("OPENNOX_REGENERATION_RAW_BEFORE") == "1" {
		return
	}
	if regenerationRawABI4E01D0() != 2 || regenerationRawEngine4E01D0 == nil {
		t.Fatal("actual native raw C ABI and engine required")
	}
	initial, inputs, originals := regenerationRawInitials4E01D0(t), regenerationRawInputs4E01D0(t), regenerationRawOriginals4E01D0(t)
	for _, profile := range []string{"tozero", "nearest"} {
		for _, backend := range []string{"Go", "C"} {
			for i, row := range initial {
				t.Run(profile+"/"+backend+"/"+row.Name, func(t *testing.T) {
					m := newRegenerationRawMemory4E01D0(t, row)
					want := originals[i]
					regenerationRawCertificate4E01D0(t, want)
					if row.Name != want.Name || !bytes.Equal(m.pe, modifierPredicateBytes4E03D0(t, want.InitialHeap)) {
						t.Fatal("whole independent initial-original heap identity")
					}
					o := newRegenerationRawObserver4E01D0(m, inputs[i], row)
					defer o.install()()
					failure := regenerationRawCapture4E01D0(t, func() {
						regenerationRawInvoke4E01D0(backend == "C", m.fromPE(row.Roles["effect"]), m.fromPE(row.Roles["item"]))
					})
					o.compare(want, failure)
				})
			}
		}
	}
}
