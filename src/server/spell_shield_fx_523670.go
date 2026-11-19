package server

import "github.com/opennox/libs/types"

func shieldHitVector523670(target, source types.Pointf) types.Pointf {
	difference := func(a, b float32) float32 {
		return experienceSpill32TowardZero(objectForceSub53_52DF80(objectForceMemory32_52DF80(a), objectForceMemory32_52DF80(b)))
	}
	// The original stores both coordinates to binary32 before taking the angle.
	return types.Ptf(difference(target.X, source.X), difference(target.Y, source.Y))
}

func shieldFXPacket523670(code uint16, position types.Pointf, facing int16, source *types.Pointf) [4]byte {
	direction := int32(facing)
	if source != nil {
		direction = int32(DirFromVec(shieldHitVector523670(position, *source)))
	}
	return [4]byte{0x80, byte(code), byte(code >> 8), byte(DirectionOctant509EA0(direction))}
}

// SendShieldFX523670 broadcasts the Shield hit in range of the native target.
func (s *Server) SendShieldFX523670(target *Object, source *types.Pointf) {
	if target == nil {
		return
	}
	code := uint16(s.GetUnitNetCode(target))
	packet := shieldFXPacket523670(code, target.PosVec, int16(target.Direction1), source)
	s.Nox_xxx_netSendFxAllCli_523030(target.PosVec, packet[:])
}
