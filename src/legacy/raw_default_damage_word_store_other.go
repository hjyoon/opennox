//go:build !darwin || !arm64

package legacy

import "unsafe"

func rawDefaultDamageWriteWord4E0B30(address unsafe.Pointer, value uint16) {
	*(*uint16)(address) = value
}
