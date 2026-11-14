package legacy

import "unsafe"

type bitCursorNative4254A0 struct {
	data *byte
	bit  uint8
}

func bitCursorReadInit4254A0(cursor *bitCursorNative4254A0, data *byte) int32 {
	cursor.data = data
	cursor.bit = 0
	return int32(*data & 1)
}

func bitCursorReadNext4254C0(cursor *bitCursorNative4254A0) bool {
	cursor.bit++
	if cursor.bit == 8 {
		cursor.bit = 0
		cursor.data = (*byte)(unsafe.Add(unsafe.Pointer(cursor.data), 1))
	}
	// The original SHL masks its count to five bits, even for a cursor whose
	// stored index is outside the ordinary 0..7 range.
	return uint32(*cursor.data)&(uint32(1)<<(cursor.bit&31)) != 0
}

func bitCursorWriteInit425500(cursor *bitCursorNative4254A0, data *byte, value byte) *byte {
	cursor.data = data
	cursor.bit = 0
	*data = value
	return data
}

func bitCursorWriteNext425520(cursor *bitCursorNative4254A0, value byte) byte {
	cursor.bit++
	if cursor.bit == 8 {
		cursor.bit = 0
		cursor.data = (*byte)(unsafe.Add(unsafe.Pointer(cursor.data), 1))
		*cursor.data = 0
	}
	result := value << (cursor.bit & 31)
	*cursor.data |= result
	return result
}

func bitstreamEncodeNative425550(source, target *byte, bits int32) int32 {
	var reader, writer bitCursorNative4254A0
	first := bitCursorReadInit4254A0(&reader, source)
	bitCursorWriteInit425500(&writer, target, byte(first))
	count := int32(1)
	if bits == 1 {
		return count
	}
	remaining := uint32(bits) - 1
	var index uint32
	for {
		index++
		if index%7 == 0 {
			bitCursorWriteNext425520(&writer, 1)
			count++
		}
		var value byte
		if bitCursorReadNext4254C0(&reader) {
			value = 1
		}
		bitCursorWriteNext425520(&writer, value)
		remaining--
		if remaining == 0 {
			return count
		}
	}
}
