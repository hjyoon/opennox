package legacy

const questJournalReadVersion500B70 = uint16(1)

// questJournalReadHooks500B70 exposes every observable journal, stream, and
// setter operation in the read-mode path through GAME.EXE 00500B70. Buffer is
// a generic comparable token so this semantic contract cannot inherit PE32
// pointer width. Stream errors are the native port's explicit counterpart to
// the original void transfer helper; they stop at the exact completed prefix.
type questJournalReadHooks500B70[Buffer comparable] struct {
	deletePattern       func(string)
	readWriteVersion    func(uint16) (uint16, error)
	readCount           func() (uint32, error)
	readNameLength      func() (uint8, error)
	readName            func(Buffer, uint8) error
	storeNameTerminator func(Buffer, uint8)
	readKind            func() (uint32, error)
	readValue           func() (uint32, error)
	setNumeric          func(Buffer, uint32)
	setBoolean          func(Buffer, uint32)
}

// questJournalReadContract500B70 preserves GAME.EXE 00500B70's destructive
// clear, signed version gate, unsigned count loop, byte-sized name length,
// trailing-NUL store, and kind dispatch. Kinds zero and one consume a value
// before calling their respective setter; every other kind consumes neither a
// value nor a setter call. Setter results are deliberately absent because the
// original ignores both return values. There are no count or name guards
// beyond the original 256-byte buffer and uint8 length.
func questJournalReadContract500B70[Buffer comparable](
	nameBuffer Buffer,
	h questJournalReadHooks500B70[Buffer],
) (int32, error) {
	h.deletePattern("*:*")

	version, err := h.readWriteVersion(questJournalReadVersion500B70)
	if err != nil {
		return 0, err
	}
	if int16(version) > int16(questJournalReadVersion500B70) {
		return 0, nil
	}

	count, err := h.readCount()
	if err != nil {
		return 0, err
	}
	for i := uint32(0); i < count; i++ {
		nameLength, err := h.readNameLength()
		if err != nil {
			return 0, err
		}
		if err := h.readName(nameBuffer, nameLength); err != nil {
			return 0, err
		}
		h.storeNameTerminator(nameBuffer, nameLength)

		kind, err := h.readKind()
		if err != nil {
			return 0, err
		}
		switch kind {
		case 0:
			value, err := h.readValue()
			if err != nil {
				return 0, err
			}
			h.setNumeric(nameBuffer, value)
		case 1:
			value, err := h.readValue()
			if err != nil {
				return 0, err
			}
			h.setBoolean(nameBuffer, value)
		}
	}
	return 1, nil
}
