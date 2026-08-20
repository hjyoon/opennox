package server

import "encoding/binary"

const (
	playerStatusReportMask4174F0 = uint32(0x00000423)
	playerStatusHostFlag4174F0   = uint32(0x00000001)
	playerStatusChatFlag417530   = uint32(0x00000080)
	playerStatusObserver417530   = uint32(0x00000001)
	playerStatusReportOp417630   = byte(106)
	playerStatusBroadcast417630  = byte(255)
)

type playerNeedStatusHooks4174F0[P any] struct {
	loadPlayerArg func() P
	loadMaskArg   func() uint32
	loadFlags     func(P) uint32
	storeFlags    func(P, uint32)
	gameFlag      func(uint32) int32
	reportStatus  func(P) int32
}

// playerNeedStatus4174F0 preserves GAME.EXE 004174F0. It loads the Player
// argument before the mask, stores the ORed flags before checking GameHost,
// and replaces the host-check result only when reporting is required.
func playerNeedStatus4174F0[P any](hooks playerNeedStatusHooks4174F0[P]) int32 {
	player := hooks.loadPlayerArg()
	mask := hooks.loadMaskArg()
	flags := hooks.loadFlags(player)
	hooks.storeFlags(player, flags|mask)
	result := hooks.gameFlag(playerStatusHostFlag4174F0)
	if result != 0 && mask&playerStatusReportMask4174F0 != 0 {
		result = hooks.reportStatus(player)
	}
	return result
}

type playerUnsetStatusHooks417530[P any] struct {
	loadMaskArg    func() uint32
	loadPlayerArg  func() P
	loadFlags      func(P) uint32
	storeFlags     func(P, uint32)
	gameFlag       func(uint32) int32
	anyPlayers     func() int32
	timerStatus    func() int32
	gameFlagsValue func() uint32
	modeEnabled    func(int16) uint8
	startTimer     func()
	setTimerStatus func(int32) int32
	reportStatus   func(P) int32
}

// playerUnsetStatus417530 preserves GAME.EXE 00417530. Unlike the setter it
// reads the mask before Player. The flag clear precedes every service, the
// observer-bit transition chain is evaluated only for a host, and a reportable
// mask replaces every earlier service result with the report callback result.
func playerUnsetStatus417530[P any](hooks playerUnsetStatusHooks417530[P]) int32 {
	mask := hooks.loadMaskArg()
	player := hooks.loadPlayerArg()
	flags := hooks.loadFlags(player)
	hooks.storeFlags(player, flags&^mask)

	result := hooks.gameFlag(playerStatusHostFlag4174F0)
	if result == 0 {
		return result
	}
	if uint8(mask)&uint8(playerStatusObserver417530) != 0 {
		result = hooks.gameFlag(playerStatusChatFlag417530)
		if result == 0 {
			result = hooks.anyPlayers()
			if result != 0 {
				result = hooks.timerStatus()
				if result == 0 {
					gameFlags := hooks.gameFlagsValue()
					if hooks.modeEnabled(int16(uint16(gameFlags))) != 0 {
						hooks.startTimer()
						result = hooks.setTimerStatus(1)
					}
				}
			}
		}
	}
	if mask&playerStatusReportMask4174F0 != 0 {
		result = hooks.reportStatus(player)
	}
	return result
}

type playerStatusReportHooks417630[P any] struct {
	loadPlayerArg func() P
	loadNetCode   func(P) uint16
	loadFlags     func(P) uint32
	send          func(byte, []byte, int32) int32
}

// playerStatusReport417630 preserves GAME.EXE 00417630. NetCode and status
// flags are cached in that order before the seven-byte broadcast packet is
// formed. The final argument is the original remove-if-disconnected zero.
func playerStatusReport417630[P any](hooks playerStatusReportHooks417630[P]) int32 {
	player := hooks.loadPlayerArg()
	netCode := hooks.loadNetCode(player)
	flags := hooks.loadFlags(player)
	var packet [7]byte
	packet[0] = playerStatusReportOp417630
	binary.LittleEndian.PutUint16(packet[1:], netCode)
	binary.LittleEndian.PutUint32(packet[3:], flags&playerStatusReportMask4174F0)
	return hooks.send(playerStatusBroadcast417630, packet[:], 0)
}
