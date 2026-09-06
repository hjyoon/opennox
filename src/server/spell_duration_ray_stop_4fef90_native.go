package server

type durationRayStopNativeDeps4FEF90 struct {
	unitCode      func(*Object) uint32
	sendPacket    func(int32, [7]byte, *Object, int32) int32
	unmarkMinimap func(*Object, uint32)
}

func durationRayStopNative4FEF90(
	record *DurSpell,
	who *Object,
	deps durationRayStopNativeDeps4FEF90,
) {
	DurationRayStop4FEF90(record, who, DurationRayStopHooks4FEF90[*DurSpell, *Object]{
		LoadCaster: func(record *DurSpell) *Object {
			return record.Caster16
		},
		LoadSpell: func(record *DurSpell) uint32 {
			return record.Spell
		},
		LoadLevelLow: func(record *DurSpell) byte {
			return byte(record.Level)
		},
		LoadDirectionLow: func(object *Object) byte {
			return byte(object.Direction1)
		},
		LoadTarget: func(record *DurSpell) *Object {
			return record.Target48
		},
		LoadSub108: func(record *DurSpell) *DurSpell {
			return record.Sub108
		},
		LoadNext: func(record *DurSpell) *DurSpell {
			return record.Next
		},
		UnitCode:      deps.unitCode,
		SendPacket:    deps.sendPacket,
		UnmarkMinimap: deps.unmarkMinimap,
	})
}

func durationRayStopServerDeps4FEF90(s *Server) durationRayStopNativeDeps4FEF90 {
	return durationRayStopNativeDeps4FEF90{
		unitCode: func(object *Object) uint32 {
			return uint32(s.GetUnitNetCode(object))
		},
		sendPacket: func(recipient int32, packet [7]byte, related *Object, remove int32) int32 {
			return int32(s.NetSendPacketXxx1(int(recipient), packet[:], related, int(remove)))
		},
		unmarkMinimap: s.Players.Nox_xxx_netUnmarkMinimapSpec_417470,
	}
}

// DurationRayStop4FEF90 binds GAME.EXE 004FEF90 to native-width duration and
// object pointers. Its legacy export already accepts the duration record as a
// void pointer; decoded C callers that produce that pointer remain separate
// ABI restoration targets.
//
//go:noinline
func (s *Server) DurationRayStop4FEF90(record *DurSpell, who *Object) {
	durationRayStopNative4FEF90(record, who, durationRayStopServerDeps4FEF90(s))
}

// NetStopRaySpell preserves the historical Go-facing name while routing every
// caller through the restored 004FEF90 implementation.
func (s *Server) NetStopRaySpell(record *DurSpell, who *Object) {
	s.DurationRayStop4FEF90(record, who)
}
