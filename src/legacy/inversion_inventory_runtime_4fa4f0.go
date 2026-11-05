package legacy

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

// Frozen typed reads and actual call boundaries precede the raw checker
// engine. Optional observers are nil in ordinary production.
var inversionInventoryObserveMemory4FA4F0 func(complete bool, base unsafe.Pointer, kind int, value modifierPredicateValue4E03D0)

func inversionInventoryRead4FA4F0(base unsafe.Pointer, kind, index int) (value modifierPredicateValue4E03D0) {
	if kind < 6 {
		return modifierPredicateRead4E03D0(base, kind, index)
	}
	if inversionInventoryObserveMemory4FA4F0 != nil {
		inversionInventoryObserveMemory4FA4F0(false, base, kind, value)
	}
	o := (*server.Object)(base)
	switch kind {
	case 6:
		value.Pointer = unsafe.Pointer(o.InvFirstItem)
	case 7:
		value.Pointer = unsafe.Pointer(o.InvNextItem)
	case 8:
		value.Pointer = unsafe.Pointer(o.ObjOwner)
	case 9:
		value.Word = uint32(o.ObjFlags)
	case 10:
		value.Word = uint32(uint8(o.ObjClass))
	default:
		panic("undeclared raw inventory field")
	}
	if inversionInventoryObserveMemory4FA4F0 != nil {
		inversionInventoryObserveMemory4FA4F0(true, base, kind, value)
	}
	return
}

type inversionInventoryCall4FA4F0 struct {
	Function      unsafe.Pointer
	Kind          string
	Complete      bool
	Args          [6]uintptr
	Before, After int32
	Word          int32
	Pointer       unsafe.Pointer
}

var inversionInventoryObserveCall4FA4F0 func(inversionInventoryCall4FA4F0)

func inversionInventoryTracedPublic4FA4F0(root, target unsafe.Pointer) int32 {
	call := inversionInventoryCall4FA4F0{Function: inversionInventoryPointer4FA4F0(false), Kind: "checker", Args: [6]uintptr{uintptr(root), uintptr(target)}}
	if inversionInventoryObserveCall4FA4F0 != nil {
		inversionInventoryObserveCall4FA4F0(call)
	}
	result := inversionInventoryPublic4FA4F0(root, target)
	if inversionInventoryObserveCall4FA4F0 != nil {
		call.Complete, call.Word = true, result
		inversionInventoryObserveCall4FA4F0(call)
	}
	return result
}

func inversionInventoryParentRuntime4FA4F0(target unsafe.Pointer) unsafe.Pointer {
	call := inversionInventoryCall4FA4F0{Function: inversionInventoryPointer4FA4F0(true), Kind: "parent", Args: [6]uintptr{uintptr(target)}}
	if inversionInventoryObserveCall4FA4F0 != nil {
		inversionInventoryObserveCall4FA4F0(call)
	}
	result := inversionInventoryParent4FA4F0(target)
	if inversionInventoryObserveCall4FA4F0 != nil {
		call.Complete, call.Pointer = true, result
		inversionInventoryObserveCall4FA4F0(call)
	}
	return result
}

func inversionInventoryInvoke4FA4F0(fn, effect unsafe.Pointer, args [4]uintptr, output *int32) int32 {
	call := inversionInventoryCall4FA4F0{Function: fn, Kind: "inversion", Args: [6]uintptr{uintptr(effect), args[0], args[1], args[2], args[3], uintptr(unsafe.Pointer(output))}}
	if inversionInventoryObserveCall4FA4F0 != nil {
		call.Before = *output
		inversionInventoryObserveCall4FA4F0(call)
	}
	result := modifierPredicateInvoke4E03D0(fn, effect, args, output)
	if inversionInventoryObserveCall4FA4F0 != nil {
		call.Complete, call.Word, call.After = true, result, *output
		inversionInventoryObserveCall4FA4F0(call)
	}
	return result
}
