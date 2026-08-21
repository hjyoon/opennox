package server

import "unsafe"

// FrogInitUpdateData is the exact three-byte prefix touched by GAME.EXE
// 004F03B0. FrogInit may be registered against any update procedure whose
// record supplies this prefix, so the remaining update record is deliberately
// not assumed here.
type FrogInitUpdateData struct {
	Delay  uint8
	Field1 uint8
	Field2 uint8
}

type frogInitNativeDeps4F03B0 struct {
	randomInt func(int32, int32, string, int32) int32
}

func frogInitNative4F03B0(unit *Object, deps frogInitNativeDeps4F03B0) int32 {
	return frogInit4F03B0(unit, frogInitHooks4F03B0[*Object, *FrogInitUpdateData]{
		loadUpdateData: func(unit *Object) *FrogInitUpdateData {
			return (*FrogInitUpdateData)(unit.UpdateData)
		},
		randomInt: deps.randomInt,
		storeDelay: func(update *FrogInitUpdateData, value uint8) {
			update.Delay = value
		},
		storeByte1: func(update *FrogInitUpdateData, value uint8) {
			update.Field1 = value
		},
		storeByte2: func(update *FrogInitUpdateData, value uint8) {
			update.Field2 = value
		},
		storeDirection: func(unit *Object, value uint16) {
			unit.Direction2 = Dir16(value)
		},
	})
}

// FrogInit4F03B0 binds GAME.EXE 004F03B0 to native-width Object and
// UpdateData pointers. The debug path and line arguments remain in the
// semantic call contract even though the original logic RNG callee ignores
// its extra cdecl arguments. There are deliberately no nil guards.
//
//go:noinline
func (s *Server) FrogInit4F03B0(unit *Object) int32 {
	return frogInitNative4F03B0(unit, frogInitNativeDeps4F03B0{
		randomInt: func(minimum, maximum int32, _ string, _ int32) int32 {
			return int32(s.Rand.Logic.IntClamp(int(minimum), int(maximum)))
		},
	})
}

var (
	_ = [1]struct{}{}[3-unsafe.Sizeof(FrogInitUpdateData{})]
	_ = [1]struct{}{}[0-unsafe.Offsetof(FrogInitUpdateData{}.Delay)]
	_ = [1]struct{}{}[1-unsafe.Offsetof(FrogInitUpdateData{}.Field1)]
	_ = [1]struct{}{}[2-unsafe.Offsetof(FrogInitUpdateData{}.Field2)]
)
