package server

import (
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/internal/netlist"
)

type predictLinearDeps523530 struct {
	netCode func(*Object) uint32
	first   func() *Object
	next    func(*Object) *Object
	send    func(uint8, []byte)
}

// predictLinear523530 keeps the existing typed compatibility entry and hooks.
func predictLinear523530(obj *Object, h predictLinearDeps523530) int32 {
	return PredictLinearNative523530(obj, PredictLinearNativeRuntime523530{
		NetCode: h.netCode, First: h.first, Next: h.next, Send: h.send,
	})
}

func (s *Server) NetClientPredictLinear523530(obj *Object) int32 {
	return PredictLinearNative523530(obj, s.PredictLinearDefaultRuntime523530())
}

func (s *Server) PredictLinearDefaultRuntime523530() PredictLinearNativeRuntime523530 {
	return PredictLinearNativeRuntime523530{
		NetCode: func(obj *Object) uint32 { return uint32(s.GetUnitNetCode(obj)) },
		First:   s.Players.FirstUnit,
		Next:    s.questNextPlayerUnit4DA7F0,
		Send: func(index uint8, packet []byte) {
			s.NetList.AddToMsgListCli(ntype.PlayerInd(index), netlist.Kind1, packet)
		},
	}
}
