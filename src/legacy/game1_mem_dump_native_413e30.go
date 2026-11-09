package legacy

import (
	"bytes"
	"encoding/binary"
	"runtime"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

const (
	memDumpRowSize413E30  = 84
	memDumpNameSize413E30 = 80
	memDumpRowCount413E30 = 1024
)

// Pointer fields grow with the host. The two intervening scalars and the
// allocation count keep their original DWORD width. See the companion header.
type memDumpNode413E30 struct {
	Reserved0                    unsafe.Pointer
	Next                         *memDumpNode413E30
	Reserved8, Reserved12, Count uint32
	Name                         [memDumpNameSize413E30]byte
}

func memDumpHead413E30() **memDumpNode413E30 {
	return (**memDumpNode413E30)(unsafe.Pointer(memmap.PtrPtr(0x5D4594, 338304)))
}

func memDumpName413E30(b []byte) []byte {
	if i := bytes.IndexByte(b, 0); i >= 0 {
		return b[:i]
	}
	// The old strcpy requires a terminating NUL inside the 80-byte name field.
	// A corrupt unterminated name is bounded here, never copied over its count.
	return b[:len(b)-1]
}

// memDumpAggregate413E30 retains the 84-byte output records and wrapping count
// additions. Only input links use the host pointer width. Repeated links and
// excess distinct names are invalid-list policies: stop a cycle and ignore
// additional new names after the fixed output table fills.
func memDumpAggregate413E30(table []byte, head *memDumpNode413E30) int {
	rows := len(table) / memDumpRowSize413E30
	for i := 0; i < rows; i++ {
		binary.LittleEndian.PutUint32(table[i*memDumpRowSize413E30+memDumpNameSize413E30:], 0)
	}
	count := 0
	seen := make(map[*memDumpNode413E30]struct{})
	for node := head; node != nil; node = node.Next {
		if _, exists := seen[node]; exists {
			break
		}
		seen[node] = struct{}{}
		name := memDumpName413E30(node.Name[:])
		index := 0
		for ; index < count; index++ {
			off := index * memDumpRowSize413E30
			if bytes.Equal(memDumpName413E30(table[off:off+memDumpNameSize413E30]), name) {
				break
			}
		}
		if index == count {
			if count == rows {
				continue
			}
			off := count * memDumpRowSize413E30
			copy(table[off:off+memDumpNameSize413E30], name)
			table[off+len(name)] = 0
			count++
		}
		off := index*memDumpRowSize413E30 + memDumpNameSize413E30
		binary.LittleEndian.PutUint32(table[off:], binary.LittleEndian.Uint32(table[off:])+node.Count)
	}
	if count != 0 {
		// Keep the actual C qsort and existing raw SUB32 comparator. Replacing
		// either with a Go ordering would change tie/wrap behavior.
		memDumpSort413E30(unsafe.Pointer(&table[0]), count)
	}
	runtime.KeepAlive(head)
	return count
}

func memDumpNative413E30() {
	table := memmap.Slice(0x5D4594, 252284)[:memDumpRowCount413E30*memDumpRowSize413E30]
	memDumpAggregate413E30(table, *memDumpHead413E30())
}
