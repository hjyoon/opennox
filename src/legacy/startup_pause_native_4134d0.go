package legacy

import (
	"runtime"
	"unsafe"
)

// Scalar state stays DWORD-sized; addresses always refer to live native owners.
type startupPauseOperand4134D0 struct {
	Site                                uint32
	Field                               int
	Base, Address                       unsafe.Pointer
	Offset                              uintptr
	Width, Value                        uint32
	Write, Complete, ValueValid, Policy bool
}

type startupPauseArg4134D0 struct {
	Word    uint32
	Pointer unsafe.Pointer
	Text    string
	Kind    string
}

type startupPauseObserver4134D0 struct {
	Operand func(startupPauseOperand4134D0)
	// true replaces only a declared external service, never an entry algorithm.
	Service func(site uint32, args []startupPauseArg4134D0) (uint32, bool)
	Cells   [6]unsafe.Pointer
	Buffer  unsafe.Pointer
}

type startupPauseFrame4134D0 struct {
	Observer *startupPauseObserver4134D0
	Previous *startupPauseFrame4134D0
}

var startupPauseHead4134D0 *startupPauseFrame4134D0

func pushStartupPauseObserver4134D0(observer *startupPauseObserver4134D0) func() {
	frame := &startupPauseFrame4134D0{observer, startupPauseHead4134D0}
	startupPauseHead4134D0 = frame
	return func() {
		if startupPauseHead4134D0 != frame {
			panic("startup/pause observer pop order")
		}
		startupPauseHead4134D0 = frame.Previous
	}
}

func startupPauseObserverCurrent4134D0() *startupPauseObserver4134D0 {
	if startupPauseHead4134D0 == nil {
		return nil
	}
	return startupPauseHead4134D0.Observer
}

func startupPauseCell4134D0(field int) unsafe.Pointer {
	if o := startupPauseObserverCurrent4134D0(); o != nil && o.Cells[field] != nil {
		return o.Cells[field]
	}
	return startupPauseDefaultCell4134D0(field)
}

func startupPauseObserve4134D0(e startupPauseOperand4134D0) {
	if o := startupPauseObserverCurrent4134D0(); o != nil && o.Operand != nil {
		o.Operand(e)
	}
}

//go:noinline
func startupPauseLoadWord4134D0(p unsafe.Pointer) uint32 { return *(*uint32)(p) }

func startupPauseRead4134D0(site uint32, field int, base unsafe.Pointer, off uintptr) uint32 {
	var address unsafe.Pointer
	if base != nil {
		address = unsafe.Add(base, off)
	}
	e := startupPauseOperand4134D0{Site: site, Field: field, Base: base, Address: address, Offset: off, Width: 4}
	startupPauseObserve4134D0(e)
	v := startupPauseLoadWord4134D0(address)
	e.Value, e.Complete, e.ValueValid = v, true, true
	startupPauseObserve4134D0(e)
	runtime.KeepAlive(base)
	return v
}

func startupPauseWrite4134D0(site uint32, field int, value uint32, policy bool) {
	p := startupPauseCell4134D0(field)
	e := startupPauseOperand4134D0{Site: site, Field: field, Base: p, Address: p, Width: 4, Value: value, ValueValid: true, Write: true, Policy: policy}
	startupPauseObserve4134D0(e)
	*(*uint32)(p) = value
	e.Complete = true
	startupPauseObserve4134D0(e)
	runtime.KeepAlive(p)
}

func startupPauseService4134D0(site uint32, args ...startupPauseArg4134D0) uint32 {
	if o := startupPauseObserverCurrent4134D0(); o != nil && o.Service != nil {
		if reply, handled := o.Service(site, args); handled {
			return reply
		}
	}
	return startupPauseDefaultService4134D0(site, args)
}

// Native startup policy: the four current C owners, not removed CD/AIL state.
func sub4134D0Native() int32 {
	for field := 0; field < 4; field++ {
		startupPauseWrite4134D0(0, field, 0, true)
	}
	return 0
}

func sub4137E0Native() {
	if startupPauseRead4134D0(0x4137e0, 3, startupPauseCell4134D0(3), 0) != 0 {
		startupPauseWrite4134D0(0x4137e9, 3, 0, false)
		startupPauseService4134D0(0x4137f3)
	}
}

// Native music uses configured paths. The existing disabled drive fallback
// clears exactly two bytes and returns nil; it does not restore a DOS drive.
func sub413890Native() unsafe.Pointer {
	p := startupPauseDefaultBuffer4134D0()
	if o := startupPauseObserverCurrent4134D0(); o != nil && o.Buffer != nil {
		p = o.Buffer
	}
	for off := uintptr(0); off < 2; off++ {
		a := unsafe.Add(p, off)
		e := startupPauseOperand4134D0{Field: 6, Base: p, Address: a, Offset: off, Width: 1, Write: true, ValueValid: true, Policy: true}
		startupPauseObserve4134D0(e)
		*(*byte)(a) = 0
		e.Complete = true
		startupPauseObserve4134D0(e)
	}
	runtime.KeepAlive(p)
	return nil
}

func sub4138E0Native(_ int32) {
	v := startupPauseService4134D0(0x4138e0)
	startupPauseWrite4134D0(0x4138e7, 4, v, false)
	startupPauseService4134D0(0x4138ec, startupPauseArg4134D0{Kind: "word", Word: 1})
}

func sub413900Native(_ int32) {
	if startupPauseService4134D0(0x413900) != 0 {
		return
	}
	if startupPauseRead4134D0(0x413909, 4, startupPauseCell4134D0(4), 0) == 0 {
		startupPauseService4134D0(0x413914, startupPauseArg4134D0{Kind: "word"})
	}
}

func sub413920Native() int32 {
	for i, site := range []uint32{0x41392e, 0x413941} {
		kind := uint32(i + 1)
		callback, text := startupPauseCallback4134D0(kind), startupPauseString4134D0(kind)
		startupPauseService4134D0(site,
			startupPauseArg4134D0{Kind: "word", Word: kind},
			startupPauseArg4134D0{Kind: "callback", Pointer: callback},
			startupPauseArg4134D0{Kind: "word"},
			startupPauseArg4134D0{Kind: "cstring", Pointer: text, Text: "Pause"})
	}
	startupPauseWrite4134D0(0x413949, 5, 0, false)
	return 1
}

func sub4139B0Native() int32 {
	if startupPauseRead4134D0(0x4139b0, 5, startupPauseCell4134D0(5), 0) != 0 {
		return 1
	}
	return 0
}

func sub413F60Native(a, b unsafe.Pointer) int32 {
	defer func() { runtime.KeepAlive(a); runtime.KeepAlive(b) }()
	x := startupPauseRead4134D0(0x413f68, 7, a, 80)
	y := startupPauseRead4134D0(0x413f6b, 8, b, 80)
	return int32(x - y) // raw SUB32, including wrap; do not normalize ordering.
}
