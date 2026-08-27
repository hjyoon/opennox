package opennox

import "github.com/opennox/opennox/v1/common/memmap"

// sub_467430 reads the client item-light override byte. Keep this tiny state
// accessor available to server-tag builds because the shared packet decoder is
// compiled there even though the client packet path is dormant.
func sub_467430() byte {
	return memmap.Uint8(0x5D4594, 1062536)
}
