package server

import "unsafe"

const teamMaterialFlagClass4ECBD0 = uint32(0x10000000)

// teamMaterialObjectIndexHooks4ECBD0 separates the original object, init-data,
// modifier, and lookup pointer domains. This keeps the GAME.EXE ABI32 loads
// observable without narrowing any native pointer to uint32.
type teamMaterialObjectIndexHooks4ECBD0[O, D, M comparable, V any] struct {
	loadClass          func(O) uint32
	loadInitData       func(O) D
	loadSecondModifier func(D) M
	lookup             func(M) V
}

// teamMaterialObjectIndex4ECBD0 preserves GAME.EXE 004ECBD0. The full class
// dword is tested first. A non-Flag returns zero without loading init-data;
// a Flag live-loads init-data, then its second modifier pointer, and passes
// that pointer (including nil) to 004ECC00.
func teamMaterialObjectIndex4ECBD0[O, D, M comparable, V any](
	obj O,
	hooks teamMaterialObjectIndexHooks4ECBD0[O, D, M, V],
) V {
	if hooks.loadClass(obj)&teamMaterialFlagClass4ECBD0 == 0 {
		var zero V
		return zero
	}
	data := hooks.loadInitData(obj)
	material := hooks.loadSecondModifier(data)
	return hooks.lookup(material)
}

// teamMaterialModifierIndexHooks4ECC00 keeps the pointer-to-pointer input,
// input C string, candidate names, and numeric team values in distinct type
// domains. Candidate table pointers are never represented by ABI32 integers.
type teamMaterialModifierIndexHooks4ECC00[M, I, C comparable, V any] struct {
	loadName          func(uint32) C
	loadInputName     func(M) I
	loadInputByte     func(I, uint32) byte
	loadCandidateByte func(C, uint32) byte
	loadTeam          func(uint32) V
}

// teamMaterialModifierIndex4ECC00 preserves the inlined two-byte strcmp scan
// at GAME.EXE 004ECC00. A nil modifier returns before the table is touched.
// Row zero is loaded and nil-gated before dereferencing the modifier's name.
// Each input byte is loaded before its candidate byte, and a matching NUL in
// the first position skips the second pair.
//
// A mismatch loads and tests the next row name before reading bytes or the ID
// from that row. A match live-loads the current ID. The scan has no explicit
// row bound, so writable legacy memory may continue beyond the shipped nil
// sentinel.
func teamMaterialModifierIndex4ECC00[M, I, C comparable, V any](
	modifier M,
	hooks teamMaterialModifierIndexHooks4ECC00[M, I, C, V],
) V {
	var zeroModifier M
	if modifier == zeroModifier {
		var zero V
		return zero
	}

	name := hooks.loadName(0)
	var zeroName C
	if name == zeroName {
		var zero V
		return zero
	}
	input := hooks.loadInputName(modifier)

	var index uint32
	for {
		var offset uint32
		for {
			inputByte := hooks.loadInputByte(input, offset)
			nameByte := hooks.loadCandidateByte(name, offset)
			if inputByte != nameByte {
				break
			}
			if inputByte == 0 {
				return hooks.loadTeam(index)
			}

			inputByte = hooks.loadInputByte(input, offset+1)
			nameByte = hooks.loadCandidateByte(name, offset+1)
			if inputByte != nameByte {
				break
			}
			offset += 2
			if inputByte == 0 {
				return hooks.loadTeam(index)
			}
		}

		name = hooks.loadName(index + 1)
		index++
		if name == zeroName {
			var zero V
			return zero
		}
	}
}

func teamMaterialModifierIndexValue4ECC00(material *ModifierEff) uint32 {
	return teamMaterialModifierIndex4ECC00(material, teamMaterialModifierIndexHooks4ECC00[
		*ModifierEff, *byte, string, uint32,
	]{
		loadName: func(index uint32) string {
			return teamMaterialTable4ECB20[index].name
		},
		loadInputName: func(material *ModifierEff) *byte {
			return material.name0
		},
		loadInputByte: func(text *byte, offset uint32) byte {
			return *(*byte)(unsafe.Add(unsafe.Pointer(text), uintptr(offset)))
		},
		loadCandidateByte: func(text string, offset uint32) byte {
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

// TeamMaterialObjectIndex4ECBD0 is the native-width runtime adapter used by
// the Go CTF path and the typed legacy export for the seven original callers.
func TeamMaterialObjectIndex4ECBD0(obj *Object) uint32 {
	return teamMaterialObjectIndex4ECBD0(obj, teamMaterialObjectIndexHooks4ECBD0[
		*Object, *ModifierInitData, *ModifierEff, uint32,
	]{
		loadClass: func(obj *Object) uint32 {
			return uint32(obj.ObjClass)
		},
		loadInitData: func(obj *Object) *ModifierInitData {
			return (*ModifierInitData)(obj.InitData)
		},
		loadSecondModifier: func(data *ModifierInitData) *ModifierEff {
			return data.Modifiers[1]
		},
		lookup: teamMaterialModifierIndexValue4ECC00,
	})
}
