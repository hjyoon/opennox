package server

// MapEntryRuntime4FC600 supplies the event fan-out that remains owned by the
// outer server. BeforeLegacy runs after the state/player gates and before the
// legacy script table is visited. AfterLegacy runs after that visit but before
// the map-entry state is cleared.
type MapEntryRuntime4FC600 struct {
	BeforeLegacy func()
	AfterLegacy  func()
}

// MapEntry4FC600 binds GAME.EXE 004FC600 to the native server state and
// NoxScript VM. Script-table and count reads stay live across callbacks, as in
// the original routine; no C pointer-width boundary is involved.
func (s *Server) MapEntry4FC600(runtime MapEntryRuntime4FC600) int32 {
	return mapEntry4FC600(mapEntryHooks4FC600[[]ScriptFunc]{
		loadState: s.MapEntryState4FC580,
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
			return table[offset/mapEntryScriptStride4FC600].Name()
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
			return s.SetMapEntryState4FC580(value)
		},
	})
}
