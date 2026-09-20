package server

import "math"

// MonsterGeneratorCreateNative54CA90 restores GAME.EXE 0054CA90 against the
// native-width MonsterGenUpdateData layout. The original C callback addresses
// these fields through PE32 byte offsets and therefore cannot initialize the
// widened leading Object pointer array on 64-bit hosts.
func MonsterGeneratorCreateNative54CA90(obj *Object) {
	update := (*MonsterGenUpdateData)(obj.UpdateData)
	update.Field92 = 2
	update.FuncInd52 = math.MaxUint32
	update.FuncInd60 = math.MaxUint32
	update.ScriptCollision.Func = -1
	update.FuncInd68 = math.MaxUint32
}
