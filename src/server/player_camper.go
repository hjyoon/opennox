package server

import (
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
	s.StartTime = uint32(s.s.Balance.Float("CamperStartTime"))
	s.FadeTime = uint32(s.s.Balance.Float("CamperFadeTime"))
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
