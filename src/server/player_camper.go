package server

import (
	"math"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/sound"
)

type playerCamper struct {
	s         *Server
	LastCheck uint32
	RadiusSq  float64
	StartTime uint32
	FadeTime  uint32
}

func (s *playerCamper) init(srv *Server) {
	s.s = srv
}

func (s *playerCamper) Reset() {
	s.LastCheck = 0
	s.RadiusSq = s.s.Balance.Float("CamperRadiusSq")
	s.StartTime = camperTimer4D7C60(s.s.Balance.Float("CamperStartTime"))
	s.FadeTime = camperTimer4D7C60(s.s.Balance.Float("CamperFadeTime"))
}

// camperTimer4D7C60 preserves the signed DWORD produced by original419A70
// before its bits become an unsigned frame interval. The startup control-word
// call at4014B5 selects chop; invalid FISTP inputs produce integer-indefinite.
func camperTimer4D7C60(value float64) uint32 {
	if math.IsNaN(value) || value < math.MinInt32 || value >= float64(math.MaxInt32)+1 {
		return 0x80000000
	}
	return uint32(int32(value))
}

func (s *playerCamper) Update() {
	s.update4D7CC0(playerCamperRuntime4D7CC0{
		frame:     s.s.Frame,
		firstUnit: s.s.Players.FirstUnit,
		nextUnit:  s.s.Players.NextUnit,
		isCTF:     func() bool { return noxflags.HasGame(noxflags.GameModeCTF) },
		audio:     func(id uint32, unit *Object) { s.s.Audio.EventObj(sound.ID(id), unit, 0, 0) },
		markerOn:  s.s.NetSendInterestingIDOn,
		markerOff: s.s.NetSendInterestingIDOff,
	})
}
