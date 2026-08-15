package server

import "unsafe"

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
// the Go-owned CTF path and retained typed legacy callers.
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
