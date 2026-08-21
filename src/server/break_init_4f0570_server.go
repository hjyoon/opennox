package server

type breakInitNativeDeps4F0570 struct {
	setXStatus func(*Object, uint32)
}

func breakInitNative4F0570(unit *Object, deps breakInitNativeDeps4F0570) {
	breakInit4F0570(unit, breakInitHooks4F0570[*Object]{
		loadStatusLow: func(unit *Object) uint8 {
			return uint8(unit.Field5)
		},
		setXStatus: deps.setXStatus,
	})
}

// BreakInit4F0570 binds GAME.EXE 004F0570 to the native-width Object while
// preserving BreakInit's independent callback identity. Field5 is the xstatus
// dword at original 32-bit offset 20 and native 64-bit offset 24. Only its low
// byte participates in the original mask test. There is deliberately no nil
// guard.
//
//go:noinline
func (s *Server) BreakInit4F0570(unit *Object) {
	breakInitNative4F0570(unit, breakInitNativeDeps4F0570{
		setXStatus: func(unit *Object, bit uint32) {
			unit.SetXStatus(bit)
		},
	})
}

// objectBreakInitParse536910 is the native-width binding for the dedicated
// parser in BreakInit's registration row. Its ObjectType write must still run
// when BreakInit has zero bytes of per-object init data.
func objectBreakInitParse536910(objt *ObjectType, _ []string) error {
	breakInitParse536910(objt, func(current *ObjectType, value uint32) {
		current.Field9 = value
	})
	return nil
}
