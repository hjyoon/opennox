package server

type chestInitNativeDeps4F0400 struct {
	setXStatus func(*Object, uint32)
}

func chestInitNative4F0400(unit *Object, deps chestInitNativeDeps4F0400) {
	chestInit4F0400(unit, chestInitHooks4F0400[*Object]{
		loadStatusLow: func(unit *Object) uint8 {
			return uint8(unit.Field5)
		},
		setXStatus: deps.setXStatus,
	})
}

// ChestInit4F0400 binds GAME.EXE 004F0400 to the native-width Object.
// Field5 is the xstatus dword at original 32-bit offset 20 and native 64-bit
// offset 24. Only its low byte participates in the original mask test. There
// is deliberately no nil guard.
//
//go:noinline
func (s *Server) ChestInit4F0400(unit *Object) {
	chestInitNative4F0400(unit, chestInitNativeDeps4F0400{
		setXStatus: func(unit *Object, bit uint32) {
			unit.SetXStatus(bit)
		},
	})
}
