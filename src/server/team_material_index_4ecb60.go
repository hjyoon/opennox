package server

// teamMaterialIndexHooks4ECB60 keeps the original pointer-valued name table
// separate from the numeric team value returned for a matching row. This
// models GAME.EXE's 32-bit table without narrowing native pointers to uint32.
type teamMaterialIndexHooks4ECB60[P comparable, V any] struct {
	loadName func(uint32) P
	loadByte func(P, uint32) byte
	loadTeam func(uint32) V
}

// teamMaterialIndex4ECB60 preserves the inlined two-byte strcmp scan at
// GAME.EXE 004ECB60. The first row name is loaded and checked before the input
// string is touched. Within each pair, the candidate byte is loaded before the
// input byte; a matching NUL in the first position skips the second pair.
//
// A mismatch loads and tests the next row name before reading any byte or team
// value from that row. A match reloads the current row's team value. The scan
// has no explicit bound, so writable legacy memory can continue past the
// shipped nil-name sentinel. No public or CGo edge is invented because neither
// GAME.EXE nor retained OpenNox history contains a reference to this routine.
func teamMaterialIndex4ECB60[P comparable, V any](input P, hooks teamMaterialIndexHooks4ECB60[P, V]) V {
	name := hooks.loadName(0)
	var zeroName P
	if name == zeroName {
		var zeroValue V
		return zeroValue
	}

	var index uint32
	for {
		var offset uint32
		for {
			nameByte := hooks.loadByte(name, offset)
			inputByte := hooks.loadByte(input, offset)
			if nameByte != inputByte {
				break
			}
			if nameByte == 0 {
				return hooks.loadTeam(index)
			}

			nameByte = hooks.loadByte(name, offset+1)
			inputByte = hooks.loadByte(input, offset+1)
			if nameByte != inputByte {
				break
			}
			offset += 2
			if nameByte == 0 {
				return hooks.loadTeam(index)
			}
		}

		name = hooks.loadName(index + 1)
		index++
		if name == zeroName {
			var zeroValue V
			return zeroValue
		}
	}
}

func teamMaterialIndexValue4ECB60(input string) uint32 {
	return teamMaterialIndex4ECB60(input, teamMaterialIndexHooks4ECB60[string, uint32]{
		loadName: func(index uint32) string {
			return teamMaterialTable4ECB20[index].name
		},
		loadByte: func(text string, offset uint32) byte {
			if uint64(offset) > uint64(len(text)) {
				panic("read beyond terminating NUL")
			}
			if uint64(offset) == uint64(len(text)) {
				return 0
			}
			return text[offset]
		},
		loadTeam: func(index uint32) uint32 {
			return teamMaterialTable4ECB20[index].team
		},
	})
}
