package server

// MapInitializeRuntime4FC590 supplies the event fan-out that remains owned by
// the outer server. BeforeLegacy runs after the state/player gates and before
// the legacy script table is visited. AfterLegacy runs after that visit but
// before the map-initialize state is cleared.
type MapInitializeRuntime4FC590 struct {
	BeforeLegacy func()
	AfterLegacy  func()
}

// MapInitialize4FC590 binds GAME.EXE 004FC590 to the native server state and
// NoxScript VM. Script-table and count reads stay live across callbacks, as in
// the original routine; no C pointer-width boundary is involved.
func (s *Server) MapInitialize4FC590(runtime MapInitializeRuntime4FC590) int32 {
	return mapInitialize4FC590(mapInitializeHooks4FC590[[]ScriptFunc]{
		loadState: s.MapInitState4FC570,
		hasPlayerUnit: func() bool {
			if !s.Players.HasUnits() {
				return false
			}
			if runtime.BeforeLegacy != nil {
				runtime.BeforeLegacy()
			}
			return true
		},
		loadScriptCount: func() int32 {
			return int32(s.NoxScriptVM.FuncsCnt())
		},
		loadScriptTable: s.NoxScriptVM.Funcs,
		loadScriptName: func(table []ScriptFunc, offset uint32) string {
			return table[offset/mapInitializeScriptStride4FC590].Name()
		},
		callScriptByIndex: func(index int32, _, _ uintptr) int32 {
			if err := s.NoxScriptVM.CallByIndex(int(index), nil, nil); err != nil {
				ScriptLog.Println(err)
			}
			return 0
		},
		clearState: func(value int32) int32 {
			if runtime.AfterLegacy != nil {
				runtime.AfterLegacy()
			}
			return s.SetMapInitState4FC570(value)
		},
	})
}
