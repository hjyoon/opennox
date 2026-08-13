package opennox

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

const (
	gameBallStatusBlob4E8290   uintptr = 0x5D4594
	gameBallStatusOffset4E8290 uintptr = 1567736
)

func (s *Server) setGameBallStatus4E8290(state uint8, netCode uint16) int32 {
	record := memmap.PtrT[gameBallStatusRecord4E8290](gameBallStatusBlob4E8290, gameBallStatusOffset4E8290)
	return gameBallStatusNative4E8290(record, state, netCode, func(recipient int32, state uint8, netCode uint16) int32 {
		return s.Nox_xxx_netSendBallStatus_4D95F0(recipient, state, netCode)
	})
}

func sub_4E8290(state uint8, netCode uint16) int32 {
	return noxServer.setGameBallStatus4E8290(state, netCode)
}

var (
	_ = [1]struct{}{}[4-unsafe.Sizeof(gameBallStatusRecord4E8290{})]
	_ = [1]struct{}{}[0-unsafe.Offsetof(gameBallStatusRecord4E8290{}.State)]
	_ = [1]struct{}{}[1-unsafe.Offsetof(gameBallStatusRecord4E8290{}.Reserved)]
	_ = [1]struct{}{}[2-unsafe.Offsetof(gameBallStatusRecord4E8290{}.NetCode)]
)
