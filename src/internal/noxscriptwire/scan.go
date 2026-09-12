// Package noxscriptwire scans the fixed-width map-script callback wire format.
package noxscriptwire

import (
	"encoding/binary"
	"io"
)

// Metadata describes the original 268-byte function records. The argument
// count is queried after each argument, as in GAME.EXE 00502790.
type Metadata interface {
	ArgCount(function byte) byte
	ArgKind(function byte, argument byte) uint32
}

// Scan preserves the read/seek sequence and signed result of GAME.EXE
// 00502790 for valid map-script records. The original reads the name into a
// 1024-byte stack buffer regardless of its length; chunking that ignored name
// avoids a stack overwrite without allocating according to untrusted input.
// A negative name length is malformed and safely stops before modifying flags.
func Scan(r io.ReadSeeker, meta Metadata, flags *uint32) int32 {
	var word [4]byte
	_, _ = r.Read(word[:])
	nameLen := int32(binary.LittleEndian.Uint32(word[:]))
	if nameLen < 0 {
		return 0
	}
	var name [1024]byte
	for left := uint32(nameLen); left != 0; {
		size := len(name)
		if left < uint32(size) {
			size = int(left)
		}
		n, err := r.Read(name[:size])
		if n <= 0 || err != nil {
			break
		}
		left -= uint32(n)
		if n != size {
			break
		}
	}

	binary.LittleEndian.PutUint32(word[:], *flags)
	_, _ = r.Read(word[:])
	*flags = binary.LittleEndian.Uint32(word[:])
	word = [4]byte{}
	_, _ = r.Read(word[:])
	count := int32(binary.LittleEndian.Uint32(word[:]))
	for i := int32(0); i < count; i++ {
		var function [1]byte
		if n, _ := r.Read(function[:]); n != 1 {
			// The original loops over uninitialized stack bytes on a short
			// record. Stop instead of spinning on a malformed huge count.
			break
		}
		_, _ = r.Seek(1, io.SeekCurrent)
		for argument := byte(0); argument < meta.ArgCount(function[0]); argument++ {
			switch meta.ArgKind(function[0], argument) {
			case 0, 3, 4, 5, 6:
				_, _ = r.Seek(4, io.SeekCurrent)
			case 1:
				_, _ = r.Seek(8, io.SeekCurrent)
			case 2, 7:
				var length [1]byte
				_, _ = r.Read(length[:])
				_, _ = r.Seek(int64(length[0]), io.SeekCurrent)
			}
		}
	}
	return count
}
