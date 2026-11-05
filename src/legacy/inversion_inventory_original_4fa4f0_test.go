package legacy

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

type inversionInventoryOriginal4FA4F0 struct {
	Name         string
	InitialHeap  string `json:"initial_heap"`
	FinalHeap    string `json:"final_heap"`
	InitialStack string `json:"initial_stack"`
	FinalStack   string `json:"final_stack"`
	Caller       string
	Result       *uint32 `json:"return_eax"`
	Fault        *struct {
		Access               int
		Address, Instruction string
	}
	Accesses []modifierPredicateAccess4E03D0
	Writes   []struct {
		Address string
		Size    int
		Value   uint32
	}
	Calls []struct {
		Function, Address string
		Args              []uint32
		Depth             int
	}
	Returns []struct {
		Function string
		Result   uint32
		Depth    int
	}
	FinalRegisters map[string]uint32 `json:"final_registers"`
}

func inversionInventoryOriginals4FA4F0(t *testing.T) []inversionInventoryOriginal4FA4F0 {
	t.Helper()
	var rows []inversionInventoryOriginal4FA4F0
	modifierPredicateFixture4E03D0(t, "inversion_inventory_original_4fa4f0.json.gz", "0bb44fa16475b2311e8d1ec5aa862fbeb87b7b94141719efe681e44310269853", "df0ec5b5b7bb81bc8c36e7e7ff2f7e936f3d741949926d91c2c4f86e6dcea609", &rows)
	if len(rows) != 207 {
		t.Fatal("all complete original raw checker observations")
	}
	return rows
}

func inversionInventoryCertificate4FA4F0(t *testing.T, row inversionInventoryOriginal4FA4F0) {
	t.Helper()
	before, after := modifierPredicateBytes4E03D0(t, row.InitialHeap), modifierPredicateBytes4E03D0(t, row.FinalHeap)
	stack, finalStack := modifierPredicateBytes4E03D0(t, row.InitialStack), modifierPredicateBytes4E03D0(t, row.FinalStack)
	if len(before) != 0x10000 || len(stack) != 0x10000 || len(finalStack) != len(stack) || !bytes.Equal(before, after) {
		t.Fatal("complete unchanged original external heap and full stack extent")
	}
	for _, write := range row.Writes {
		address := modifierPredicateAddress4E03D0(t, write.Address)
		if address < 0x22000000 || address+uint32(write.Size) > 0x22010000 || write.Size != 4 {
			t.Fatal("original checker/callee write outside private stack DWORDs")
		}
		binary.LittleEndian.PutUint32(stack[address-0x22000000:], write.Value)
	}
	caller := modifierPredicateBytes4E03D0(t, row.Caller)
	if !bytes.Equal(stack, finalStack) || len(caller) != 12 || !bytes.Equal(finalStack[0xe000:0xe00c], caller) {
		t.Fatal("complete original stack/write/caller certificate")
	}
	if (row.Fault == nil) != (row.Result != nil) {
		t.Fatal("original return/fault terminal state")
	}
	if row.Result != nil {
		if *row.Result > 1 || row.FinalRegisters["eax"] != *row.Result || row.FinalRegisters["esp"] != 0x2200e004 || row.FinalRegisters["eip"] != 0x23000100 || row.FinalRegisters["ebx"] != 0x12345678 || row.FinalRegisters["esi"] != 0x23456789 || row.FinalRegisters["edi"] != 0x3456789a || row.FinalRegisters["ebp"] != 0x456789ab {
			t.Fatal("complete original returned register/caller state")
		}
	} else {
		last := row.Accesses[len(row.Accesses)-1]
		if last.Complete || last.Address != row.Fault.Address || last.Instruction != row.Fault.Instruction || row.FinalRegisters["eip"] != modifierPredicateAddress4E03D0(t, row.Fault.Instruction) {
			t.Fatal("complete original fault prefix and register instruction")
		}
	}
}

type inversionInventoryObserver4FA4F0 struct {
	m        *modifierPredicateMemory4E03D0
	accesses []modifierPredicateAccess4E03D0
	calls    []inversionInventoryCall4FA4F0
}

func (o *inversionInventoryObserver4FA4F0) oldMemory(write, complete bool, base unsafe.Pointer, kind, index int, value modifierPredicateValue4E03D0) {
	old := modifierPredicateObserver4E03D0{m: o.m, accesses: o.accesses}
	old.observe(write, complete, base, kind, index, value)
	o.accesses = old.accesses
}

func (o *inversionInventoryObserver4FA4F0) memory(complete bool, base unsafe.Pointer, kind int, value modifierPredicateValue4E03D0) {
	pe, ok := o.m.reverse[uintptr(base)]
	if !ok {
		o.m.t.Fatal("undeclared inventory/owner memory base", base)
	}
	offset, size := uint32(0), 4
	switch kind {
	case 6:
		offset = 504
	case 7:
		offset = 496
	case 8:
		offset = 508
	case 9:
		offset = 16
	case 10:
		offset, size = 8, 1
	default:
		o.m.t.Fatal("undeclared inventory/owner field")
	}
	address := fmt.Sprintf("%#x", pe+offset)
	if !complete {
		o.accesses = append(o.accesses, modifierPredicateAccess4E03D0{Kind: "read", Address: address, Size: size})
		return
	}
	valueWord := value.Word
	if kind == 6 || kind == 7 || kind == 8 {
		valueWord = o.m.toPE(uintptr(value.Pointer))
	}
	if len(o.accesses) == 0 {
		o.m.t.Fatal("completed inventory read without entry")
	}
	last := &o.accesses[len(o.accesses)-1]
	if last.Complete || last.Kind != "read" || last.Address != address || last.Size != size {
		o.m.t.Fatal("reordered inventory/owner memory access")
	}
	last.Complete, last.Value = true, &valueWord
}

func (o *inversionInventoryObserver4FA4F0) parent(read server.OwnerChainPlayerRead4EC580) {
	kind, value := 8, modifierPredicateValue4E03D0{Pointer: unsafe.Pointer(read.Owner)}
	if read.Field == 1 {
		kind, value = 10, modifierPredicateValue4E03D0{Word: uint32(read.Class)}
	} else if read.Field != 0 {
		o.m.t.Fatal("undeclared actual parent field")
	}
	o.memory(read.Complete, unsafe.Pointer(read.Object), kind, value)
}

func (o *inversionInventoryObserver4FA4F0) install() func() {
	oldMemory, oldInventory, oldCall, oldParent := modifierPredicateObserve4E03D0, inversionInventoryObserveMemory4FA4F0, inversionInventoryObserveCall4FA4F0, server.ObserveOwnerChainPlayerRead4EC580
	modifierPredicateObserve4E03D0 = o.oldMemory
	inversionInventoryObserveMemory4FA4F0 = o.memory
	inversionInventoryObserveCall4FA4F0 = func(call inversionInventoryCall4FA4F0) { o.calls = append(o.calls, call) }
	server.ObserveOwnerChainPlayerRead4EC580 = o.parent
	return func() {
		modifierPredicateObserve4E03D0, inversionInventoryObserveMemory4FA4F0, inversionInventoryObserveCall4FA4F0, server.ObserveOwnerChainPlayerRead4EC580 = oldMemory, oldInventory, oldCall, oldParent
	}
}

func inversionInventoryCompareAccesses4FA4F0(t *testing.T, got, expected []modifierPredicateAccess4E03D0) {
	t.Helper()
	want := append([]modifierPredicateAccess4E03D0(nil), expected...)
	for i := range want {
		want[i].Instruction = ""
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatal("complete ordered original memory values/widths/fault prefix", got, want)
	}
}

func inversionInventoryReplayRead4FA4F0(t *testing.T, m *modifierPredicateMemory4E03D0, row inversionInventoryInitial4FA4F0, access modifierPredicateAccess4E03D0) {
	t.Helper()
	if access.Kind != "read" {
		t.Fatal("original declared external heap is read-only")
	}
	address := modifierPredicateAddress4E03D0(t, access.Address)
	kind, index, offset := 0, 0, uint32(0)
	switch access.Instruction {
	case "0x4fa4f9":
		kind, offset = 6, 504
	case "0x4fa507":
		kind, offset = 9, 16
	case "0x4fa50f":
		kind, offset = 1, 8
	case "0x4fa518":
		kind, offset = 2, 692
	case "0x4fa526":
		kind, index = 3, -1
		for _, read := range row.Reads {
			if read.Kind == "slot" && read.Role+4*uint32(read.Index) == address {
				index, offset = read.Index, 4*uint32(read.Index)
			}
		}
		if index < 0 {
			t.Fatal("independent slot address required")
		}
	case "0x4fa534", "0x4fa561":
		kind, offset = 4, 88
	case "0x4fa57d":
		kind, offset = 7, 496
	case "0x4ec588", "0x4ec59b":
		kind, offset = 8, 508
	case "0x4ec594":
		kind, offset = 10, 8
	case "0x4e03d6":
		kind, offset = 0, 96
	default:
		t.Fatal("undeclared exact original read instruction", access.Instruction)
	}
	fault := modifierPredicateRecover4E03D0(t, func() { inversionInventoryRead4FA4F0(m.fromPE(address-offset), kind, index) })
	if fault == access.Complete {
		t.Fatal("actual native replay fault domain", access)
	}
}

func TestInversionInventoryFrozenRuntime4FA4F0(t *testing.T) {
	for _, row := range inversionInventoryInitials4FA4F0(t) {
		t.Run(row.Name, func(t *testing.T) {
			m := newInversionInventoryMemory4FA4F0(t, row)
			o := &inversionInventoryObserver4FA4F0{m: m}
			defer o.install()()
			for _, read := range row.Reads {
				kind := inversionInventoryReadKinds4FA4F0[read.Kind]
				var value modifierPredicateValue4E03D0
				fault := modifierPredicateRecover4E03D0(t, func() { value = inversionInventoryRead4FA4F0(m.fromPE(read.Role), kind, read.Index) })
				if fault != (read.Want == nil) {
					t.Fatal("frozen native getter independent fault domain")
				}
				if !fault {
					got := value.Word
					if inversionInventoryPointerKind4FA4F0(kind) {
						got = m.toPE(uintptr(value.Pointer))
					}
					if got != *read.Want {
						t.Fatal("frozen native getter independent value")
					}
				}
			}
			if len(o.accesses) != len(row.Reads) {
				t.Fatal("every frozen getter observed")
			}
			m.compareWhole(m.pe, 0)
		})
	}
}

func TestInversionInventoryActualParentTrace4FA4F0(t *testing.T) {
	initial, originals := inversionInventoryInitials4FA4F0(t), inversionInventoryOriginals4FA4F0(t)
	for i, row := range initial {
		t.Run(row.Name, func(t *testing.T) {
			m := newInversionInventoryMemory4FA4F0(t, row)
			o := &inversionInventoryObserver4FA4F0{m: m}
			defer o.install()()
			for _, call := range originals[i].Calls {
				if call.Function == "parent" {
					modifierPredicateRecover4E03D0(t, func() { inversionInventoryParentRuntime4FA4F0(m.fromPE(row.Target)) })
				}
			}
			var want []modifierPredicateAccess4E03D0
			for _, access := range originals[i].Accesses {
				if access.Instruction == "0x4ec588" || access.Instruction == "0x4ec594" || access.Instruction == "0x4ec59b" {
					want = append(want, access)
				}
			}
			inversionInventoryCompareAccesses4FA4F0(t, o.accesses, want)
			m.compareWhole(m.pe, 0)
		})
	}
}

func TestInversionInventoryWholeTransport4FA4F0(t *testing.T) {
	initial, originals := inversionInventoryInitials4FA4F0(t), inversionInventoryOriginals4FA4F0(t)
	for _, row := range originals {
		inversionInventoryCertificate4FA4F0(t, row)
	}
	for _, profile := range []string{"tozero", "nearest"} {
		for i, row := range initial {
			t.Run(profile+"/"+row.Name, func(t *testing.T) {
				m := newInversionInventoryMemory4FA4F0(t, row)
				o := &inversionInventoryObserver4FA4F0{m: m}
				defer o.install()()
				if row.Name != originals[i].Name || !bytes.Equal(m.pe, modifierPredicateBytes4E03D0(t, originals[i].InitialHeap)) {
					t.Fatal("independent complete original initial heap identity")
				}
				for _, access := range originals[i].Accesses {
					inversionInventoryReplayRead4FA4F0(t, m, row, access)
				}
				m.compareWhole(modifierPredicateBytes4E03D0(t, originals[i].FinalHeap), 0)
				inversionInventoryCompareAccesses4FA4F0(t, o.accesses, originals[i].Accesses)
			})
		}
	}
}

func TestInversionInventoryConsumersBefore4FA4F0(t *testing.T) {
	if os.Getenv("OPENNOX_INVERSION_INVENTORY_BEFORE") != "1" {
		return
	}
	if inversionInventoryABI4FA4F0() != 1 {
		t.Fatal("complete checker consumers must precede its native C prototype")
	}
	for i := 0; i < 3; i++ {
		if modifierPredicatesABI4E03D0(i) != 1 {
			t.Fatal("all four coupled raw C engines must remain absent")
		}
	}
	if _, err := os.Stat("inversion_inventory_native_4fa4f0.go"); !os.IsNotExist(err) {
		t.Fatal("checker engine must be absent before full consumer freeze")
	}
}

func TestInversionInventoryOriginal4FA4F0(t *testing.T) {
	if os.Getenv("OPENNOX_INVERSION_INVENTORY_BEFORE") == "1" {
		return // Full compiled consumer; zero raw checker engine leaves yet.
	}
	if inversionInventoryABI4FA4F0() != 2 {
		t.Fatal("actual raw checker must accept two native object pointers")
	}
	for i := 0; i < 3; i++ {
		if modifierPredicatesABI4E03D0(i) != 2 {
			t.Fatal("all three coupled raw modifier callbacks must be native")
		}
	}
	initial, originals := inversionInventoryInitials4FA4F0(t), inversionInventoryOriginals4FA4F0(t)
	for _, profile := range []string{"tozero", "nearest"} {
		for i, row := range initial {
			t.Run(profile+"/C/"+row.Name, func(t *testing.T) {
				m := newInversionInventoryMemory4FA4F0(t, row)
				o := &inversionInventoryObserver4FA4F0{m: m}
				defer o.install()()
				modifierPredicateTracing4E03D0(true)
				defer modifierPredicateTracing4E03D0(false)
				var result int32
				fault := modifierPredicateRecover4E03D0(t, func() { result = inversionInventoryTracedPublic4FA4F0(m.argument, m.fromPE(row.Target)) })
				want := originals[i]
				if fault != (want.Fault != nil) || (!fault && uint32(result) != *want.Result) {
					t.Fatal("complete original checker return/fault", fault, result, want.Result, want.Fault)
				}
				m.compareWhole(modifierPredicateBytes4E03D0(t, want.FinalHeap), 0)
				inversionInventoryCompareAccesses4FA4F0(t, o.accesses, want.Accesses)
				inversionInventoryCompareCalls4FA4F0(t, m, o.calls, want)
			})
		}
	}
}

func inversionInventoryCompareCalls4FA4F0(t *testing.T, m *modifierPredicateMemory4E03D0, calls []inversionInventoryCall4FA4F0, want inversionInventoryOriginal4FA4F0) {
	t.Helper()
	entered, returned, nativeCallbacks := 0, 0, 0
	trace := modifierPredicateTraceRecords4E03D0()
	for _, call := range calls {
		if !call.Complete {
			if entered >= len(want.Calls) {
				t.Fatal("extra native call")
			}
			old := want.Calls[entered]
			if call.Kind != old.Function || m.toPE(uintptr(call.Function)) != modifierPredicateAddress4E03D0(t, old.Address) {
				t.Fatal("actual native call identity/order", call, old)
			}
			for i, argument := range old.Args {
				if i == 5 {
					if call.Args[i] == 0 || (unsafe.Sizeof(uintptr(0)) == 8 && call.Args[i] <= 0xffffffff) || call.Before != 0 {
						t.Fatal("native caller-owned DWORD scratch initialized exactly zero")
					}
				} else if call.Args[i] != uintptr(m.fromPE(argument)) {
					t.Fatal("full native callback/caller argument", i, call, old)
				}
			}
			if call.Kind == "inversion" {
				if nativeCallbacks >= len(trace) {
					t.Fatal("missing actual dynamic C callback entry")
				}
				enter := trace[nativeCallbacks]
				if enter.Phase != 2 || enter.Function != uintptr(call.Function) || enter.Args != call.Args || enter.Before != 0 {
					t.Fatal("actual C callback entry, including a faulting callback", enter, call)
				}
				nativeCallbacks++
			}
			entered++
			continue
		}
		if returned >= len(want.Returns) || call.Kind != want.Returns[returned].Function {
			t.Fatal("actual native return order")
		}
		expected := want.Returns[returned].Result
		if call.Kind == "parent" {
			if m.toPE(uintptr(call.Pointer)) != expected {
				t.Fatal("actual parent pointer return")
			}
		} else if uint32(call.Word) != expected {
			t.Fatal("actual DWORD return")
		}
		if call.Kind == "inversion" {
			if nativeCallbacks >= len(trace) {
				t.Fatal("missing actual dynamic C callback return")
			}
			leave := trace[nativeCallbacks]
			if leave.Phase != 3 || leave.Function != uintptr(call.Function) || leave.Args != call.Args || leave.Before != 0 || uint32(leave.After) != expected || uint32(leave.Result) != expected || uint32(call.After) != expected {
				t.Fatal("actual C callback return and original scratch value", leave, call)
			}
			nativeCallbacks++
		}
		returned++
	}
	if entered != len(want.Calls) || returned != len(want.Returns) || nativeCallbacks != len(trace) {
		t.Fatal("every original call/return and actual C callback must be consumed")
	}
}
