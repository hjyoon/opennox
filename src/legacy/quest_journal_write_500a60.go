package legacy

const (
	questJournalWriteVersion500A60  = uint16(1)
	questJournalWriteCoopMask500A60 = uint32(0x800)
)

// questJournalWriteHooks500A60 exposes every observable stream operation,
// global read, and journal-node access in GAME.EXE 00500A60. Entry remains a
// generic comparable token so the semantic contract cannot inherit PE32
// pointer width. Stream errors are the native port's explicit counterpart to
// the original void transfer helper; they stop at the exact completed prefix.
type questJournalWriteHooks500A60[Entry comparable] struct {
	readWriteVersion    func(uint16) (uint16, error)
	loadHead            func() Entry
	loadNext            func(Entry) Entry
	checkGameFlags      func(uint32) int32
	readWriteCount      func(uint32) error
	scanNameLength      func(Entry) uint32
	readWriteNameLength func(uint8) (uint8, error)
	readWriteName       func(Entry, uint8) error
	readWriteKind       func(Entry) error
	loadKind            func(Entry) uint32
	readWriteValue      func(Entry) error
}

// questJournalWriteContract500A60 preserves GAME.EXE 00500A60's signed
// version gate and precise access order. The list is counted before the Coop
// flag check, even when the stream will receive zero. The Coop path writes the
// cached count and then reloads the global head. Each name length is truncated
// to its low byte, while kind and next are reloaded after their preceding
// transfers so callback mutations remain visible. There are deliberately no
// nil-node, cycle, or string-bounds guards beyond the original zero token.
func questJournalWriteContract500A60[Entry comparable](
	h questJournalWriteHooks500A60[Entry],
) (int32, error) {
	version, err := h.readWriteVersion(questJournalWriteVersion500A60)
	if err != nil {
		return 0, err
	}
	if int16(version) > int16(questJournalWriteVersion500A60) {
		return 0, nil
	}

	var nilEntry Entry
	entry := h.loadHead()
	var count uint32
	for entry != nilEntry {
		next := h.loadNext(entry)
		count++
		entry = next
	}

	if h.checkGameFlags(questJournalWriteCoopMask500A60) == 0 {
		if err := h.readWriteCount(0); err != nil {
			return 0, err
		}
		return 1, nil
	}
	if err := h.readWriteCount(count); err != nil {
		return 0, err
	}

	for entry = h.loadHead(); entry != nilEntry; {
		nameLength := uint8(h.scanNameLength(entry))
		nameLength, err = h.readWriteNameLength(nameLength)
		if err != nil {
			return 0, err
		}
		if err := h.readWriteName(entry, nameLength); err != nil {
			return 0, err
		}
		if err := h.readWriteKind(entry); err != nil {
			return 0, err
		}
		kind := h.loadKind(entry)
		if kind == 0 || kind == 1 {
			if err := h.readWriteValue(entry); err != nil {
				return 0, err
			}
		}
		entry = h.loadNext(entry)
	}
	return 1, nil
}
