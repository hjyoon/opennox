package netalias

import "encoding/binary"

const (
	Count      = 255
	RecordSize = 8
	TableSize  = Count * RecordSize
	NoAlias    = 0xff
)

type Entry struct {
	Code     uint16
	TypeID   uint16
	Deadline uint32
}

func (e Entry) PutRecord(dst []byte) {
	binary.LittleEndian.PutUint16(dst, e.Code)
	binary.LittleEndian.PutUint16(dst[2:], e.TypeID)
	binary.LittleEndian.PutUint32(dst[4:], e.Deadline)
}

type Table []byte

func (t Table) Lookup(ind byte) Entry {
	entry := t[int(ind)*RecordSize:]
	return Entry{
		Code:     binary.LittleEndian.Uint16(entry),
		TypeID:   binary.LittleEndian.Uint16(entry[2:]),
		Deadline: binary.LittleEndian.Uint32(entry[4:]),
	}
}

func (t Table) Store(ind byte, entry Entry) {
	entry.PutRecord(t[int(ind)*RecordSize:])
}

// FindAvailable selects the first matching or expired slot. The original
// selector compares the full arguments to the stored 16-bit words.
func FindAvailable(code, typeID int32, now uint32, lookup func(byte) Entry) byte {
	ind := byte(code)
	if ind == 0 || ind == NoAlias {
		ind = 1
	}
	start := ind
	for {
		entry := lookup(ind)
		if (int32(entry.Code) == code && int32(entry.TypeID) == typeID) || entry.Deadline < now {
			return ind
		}
		ind++
		if ind == NoAlias {
			ind = 1
		}
		if ind == start {
			return NoAlias
		}
	}
}
