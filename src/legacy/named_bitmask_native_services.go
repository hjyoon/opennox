package legacy

import "unsafe"

// Serial observation of the three original string operations.
type namedBitmaskObserver423930 struct {
	Compare func(unsafe.Pointer, unsafe.Pointer) int32
	Prefix  func(unsafe.Pointer) int32
	Token   func(unsafe.Pointer) unsafe.Pointer
}

var namedBitmaskObserved423930 *namedBitmaskObserver423930

func namedBitmaskObserve423930(o *namedBitmaskObserver423930) func() {
	previous := namedBitmaskObserved423930
	namedBitmaskObserved423930 = o
	return func() {
		if namedBitmaskObserved423930 != o {
			panic("named-bitmask observer restore order")
		}
		namedBitmaskObserved423930 = previous
	}
}

func namedBitmaskCompare4239C0(a, b unsafe.Pointer) int32 {
	if o := namedBitmaskObserved423930; o != nil && o.Compare != nil {
		return o.Compare(a, b)
	}
	return int32(compareLegacyStringBytes(a, b, -1))
}

func namedBitmaskPrefix423930(input unsafe.Pointer) int32 {
	if o := namedBitmaskObserved423930; o != nil && o.Prefix != nil {
		return o.Prefix(input)
	}
	const prefix = "NULL"
	for i := range len(prefix) {
		if value := *(*byte)(unsafe.Add(input, i)); value != prefix[i] {
			return int32(value) - int32(prefix[i])
		}
	}
	return 0
}

func namedBitmaskToken423930(tokens *legacyByteTokens, input unsafe.Pointer) unsafe.Pointer {
	if o := namedBitmaskObserved423930; o != nil && o.Token != nil {
		return o.Token(input)
	}
	return unsafe.Pointer(tokens.take("+"))
}
