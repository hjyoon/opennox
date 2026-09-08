package legacy

import "strings"

// questJournalQualifyHooks5009B0 exposes each C-string scan and scratch
// mutation performed by GAME.EXE 005009B0. Lengths exclude the terminating
// NUL; copy counts include it, matching the IA-32 strlen/memcpy sequence.
type questJournalQualifyHooks5009B0 struct {
	scanInputColon    func() bool
	scanInputLength   func() int
	scanMapLength     func() int
	copyInput         func(int)
	copyMap           func(int)
	scanScratchLength func() int
	writeSeparator    func(int)
	appendInput       func(int, int)
}

// questJournalQualifyContract5009B0 preserves GAME.EXE 005009B0's branch,
// access, copy, and return order for valid NUL-terminated strings. Any colon
// makes the input qualified: the input is scanned again, copied with its NUL,
// and its byte count is returned. Otherwise the current map is copied with its
// NUL, that terminator is replaced by ":\0", the input is appended with its
// NUL, and zero is returned.
func questJournalQualifyContract5009B0(h questJournalQualifyHooks5009B0) uint32 {
	if h.scanInputColon() {
		count := h.scanInputLength() + 1
		h.copyInput(count)
		return uint32(count)
	}

	mapCount := h.scanMapLength() + 1
	h.copyMap(mapCount)
	h.writeSeparator(h.scanScratchLength())
	inputCount := h.scanInputLength() + 1
	h.appendInput(h.scanScratchLength(), inputCount)
	return 0
}

// questJournalQualifyString5009B0 applies the exact contract to a Go-owned
// scratch value. loadMapName remains lazy so already-qualified inputs never
// touch the current-map slot, as in GAME.EXE.
func questJournalQualifyString5009B0(name string, loadMapName func() string) (string, uint32) {
	var mapName string
	var scratch string
	result := questJournalQualifyContract5009B0(questJournalQualifyHooks5009B0{
		scanInputColon: func() bool {
			return strings.IndexByte(name, ':') >= 0
		},
		scanInputLength: func() int {
			return len(name)
		},
		scanMapLength: func() int {
			mapName = loadMapName()
			return len(mapName)
		},
		copyInput: func(count int) {
			if count != len(name)+1 {
				panic("quest journal qualified copy count mismatch")
			}
			scratch = name
		},
		copyMap: func(count int) {
			if count != len(mapName)+1 {
				panic("quest journal map copy count mismatch")
			}
			scratch = mapName
		},
		scanScratchLength: func() int {
			return len(scratch)
		},
		writeSeparator: func(at int) {
			if at != len(scratch) {
				panic("quest journal separator offset mismatch")
			}
			scratch += ":"
		},
		appendInput: func(at, count int) {
			if at != len(scratch) || count != len(name)+1 {
				panic("quest journal append range mismatch")
			}
			scratch += name
		},
	})
	return scratch, result
}
