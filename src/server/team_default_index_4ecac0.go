package server

const teamDefaultIndexMax4ECAC0 = uint8(16)

// teamDefaultIndexHooks4ECAC0 models the two pointer-valued operations in
// GAME.EXE 004ECAC0 without embedding the original 32-bit address table in a
// native-width runtime. The input and table entries have the same pointer
// type, just as the original const char pointers do.
type teamDefaultIndexHooks4ECAC0[P any] struct {
	loadName func(uint8) P
	loadByte func(P, uint32) byte
}

// teamDefaultIndex4ECAC0 preserves the inlined two-byte strcmp loop in
// GAME.EXE 004ECAC0. Each candidate name pointer is loaded before the input is
// read. Within each pair, the input byte is loaded before the candidate byte;
// a matching NUL in the first position skips the second pair of loads.
//
// Candidate 16 is included. In the shipped table that cell is nil, so an
// unmatched valid input faults while comparing against it: the input byte at
// offset zero is read before the nil candidate is dereferenced. Returning zero
// after exhausting candidate 16 remains possible if writable legacy memory
// supplies a non-nil seventeenth string. No public or CGo edge is invented,
// because GAME.EXE and retained OpenNox history contain no reference to this
// routine.
func teamDefaultIndex4ECAC0[P any](input P, hooks teamDefaultIndexHooks4ECAC0[P]) uint8 {
	var result uint8
	for {
		name := hooks.loadName(result)
		var offset uint32
		for {
			inputByte := hooks.loadByte(input, offset)
			nameByte := hooks.loadByte(name, offset)
			if inputByte != nameByte {
				break
			}
			if inputByte == 0 {
				return result
			}

			inputByte = hooks.loadByte(input, offset+1)
			nameByte = hooks.loadByte(name, offset+1)
			if inputByte != nameByte {
				break
			}
			offset += 2
			if inputByte == 0 {
				return result
			}
		}

		result++
		if result > teamDefaultIndexMax4ECAC0 {
			return 0
		}
	}
}
