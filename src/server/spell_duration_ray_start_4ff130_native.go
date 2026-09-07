package server

type durationRayStartNativeDeps4FF130 struct {
	unitCode      func(*Object) uint32
	sendPacket    func(int32, [7]byte, *Object, int32) int32
	unmarkMinimap func(*Object, uint32)
}

func durationRayStartNative4FF130(
	record *DurSpell,
	deps durationRayStartNativeDeps4FF130,
) {
	DurationRayStart4FF130(record, DurationRayStartHooks4FF130[*DurSpell, *Object]{
		LoadSpell: func(record *DurSpell) uint32 {
			return record.Spell
		},
		LoadLevelLow: func(record *DurSpell) byte {
			return byte(record.Level)
		},
		LoadCaster: func(record *DurSpell) *Object {
			return record.Caster16
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

func durationRayStartServerDeps4FF130(s *Server) durationRayStartNativeDeps4FF130 {
	return durationRayStartNativeDeps4FF130{
		unitCode: func(object *Object) uint32 {
			return uint32(s.GetUnitNetCode(object))
		},
		sendPacket: func(recipient int32, packet [7]byte, related *Object, remove int32) int32 {
			return int32(s.NetSendPacketXxx1(int(recipient), packet[:], related, int(remove)))
		},
		unmarkMinimap: s.Players.Nox_xxx_netUnmarkMinimapSpec_417470,
	}
}

// DurationRayStart4FF130 binds GAME.EXE 004FF130 to native-width duration and
// object pointers. Its retained C export receives the duration record as a
// void pointer; decoded callers whose own storage is still dword-sized remain
// separate ABI restoration targets.
//
//go:noinline
func (s *Server) DurationRayStart4FF130(record *DurSpell) {
	durationRayStartNative4FF130(record, durationRayStartServerDeps4FF130(s))
}

// NetStartDurationRaySpell preserves the historical Go-facing name while
// routing every caller through the restored 004FF130 implementation.
func (s *Server) NetStartDurationRaySpell(record *DurSpell) {
	s.DurationRayStart4FF130(record)
}
