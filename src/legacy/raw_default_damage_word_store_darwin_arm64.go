package legacy

import "unsafe"

// The retained original WORD fault records expose a completed low byte before
// the high byte reaches an inaccessible page. An ARM64 STRH faults without
// that progress. Issue two real ordered stores, without looking at a fixture,
// observer, page map or expected result. Separate noninline calls prevent the
// compiler from combining the writes back into a single STRH.
//
//go:noinline
func rawDefaultDamageWriteWord4E0B30(address unsafe.Pointer, value uint16) {
	rawDefaultDamageWriteWordByte4E0B30(address, byte(value))
	rawDefaultDamageWriteWordByte4E0B30(unsafe.Add(address, 1), byte(value>>8))
}

//go:noinline
func rawDefaultDamageWriteWordByte4E0B30(address unsafe.Pointer, value byte) {
	*(*byte)(address) = value
}
