package legacy

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

// Original 00409B80 returns the live map-stem storage token005D5310. The
// registered C-owned blob supplies its relocated native address. This accessor
// allocates nothing and establishes neither string capacity nor termination.
func sub409B80Native() unsafe.Pointer {
	return memmap.PtrOff(0x5D4594, 3452)
}
