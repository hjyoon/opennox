package opennox

import (
	"encoding/binary"
	"math"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/client"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/internal/netlist"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

// playerObjectVisualState518C30 is the nine-byte payload shared by the
// original MSG_PLAYER_OBJ sender and receiver after the three identifying
// fields have been decoded.
type playerObjectVisualState518C30 struct {
	X, Y      uint16
	Direction byte
	Frame     byte
	Animation byte
}

func decodePlayerObjectVisualState518C30(data []byte) (playerObjectVisualState518C30, bool) {
	if len(data) < 12 {
		return playerObjectVisualState518C30{}, false
	}
	direction := (data[9] >> 4) & 7
	if direction > 3 {
		direction++
	}
	return playerObjectVisualState518C30{
		X:         binary.LittleEndian.Uint16(data[5:]),
		Y:         binary.LittleEndian.Uint16(data[7:]),
		Direction: direction,
		Frame:     data[10],
		Animation: data[11],
	}, true
}

func applyPlayerObjectVisualState518C30(dr *client.Drawable, state playerObjectVisualState518C30, frame uint32) {
	dr.Field_72 = frame
	dr.SetFrameMB(int(state.Frame))
	dr.AnimDir = state.Direction
	if dr.AnimInd != uint32(state.Animation) {
		dr.AnimInd = uint32(state.Animation)
		dr.AnimStart = frame
	}
}

func (c *Client) handlePlayerObjectPacketNative518C30(data []byte) int {
	state, ok := decodePlayerObjectVisualState518C30(data)
	if !ok {
		return -1
	}
	code := binary.LittleEndian.Uint16(data[1:])
	typeID := binary.LittleEndian.Uint16(data[3:])
	if code != 0 || typeID != 0 {
		dr := c.Nox_xxx_spriteCreate_48E970(int(typeID), code, int(state.X), int(state.Y))
		if dr != nil {
			applyPlayerObjectVisualState518C30(dr, state, c.srv.Frame())
		}
		if nox_xxx_netClearHighBit_578B30(code) == uint16(legacy.ClientPlayerNetCode()) && sub_416120(8) {
			nox_xxx_cliUpdateCameraPos_435600(int(state.X), int(state.Y))
		}
		return 12
	}
	nox_xxx_cliUpdateCameraPos_435600(int(state.X), int(state.Y))
	inputSetKeyTimeoutLegacy(8)
	return 12
}

func playerObjectDirection518C30(direction server.Dir16) byte {
	vec := direction.Vec()
	x := int(math.RoundToEven(float64(vec.X) * 16))
	y := int(math.RoundToEven(float64(vec.Y) * 16))
	classify := func(v int) int {
		if v > 6 {
			return 1
		}
		if v < -6 {
			return -1
		}
		return 0
	}
	ind := classify(x) + 3*classify(y) + 4
	if ind > 3 {
		ind--
	}
	return byte(ind)
}

func (s *Server) playerAnimationStateNative4FA2B0(unit *server.Object) byte {
	return byte(s.Server.PlayerActionState4FA2B0(unit))
}

func (s *Server) playerObjectPacketNative518C30(unit *server.Object) [12]byte {
	update := unit.UpdateDataPlayer()
	state := s.playerAnimationStateNative4FA2B0(unit)
	animFrame := byte(0xff)
	switch update.State {
	case server.PlayerState1, server.PlayerState2, server.PlayerState10,
		server.PlayerState15, server.PlayerState16, server.PlayerState17,
		server.PlayerState14, server.PlayerState20, server.PlayerState18,
		server.PlayerState19, server.PlayerState21, server.PlayerState22,
		server.PlayerState24, server.PlayerStateShakeFist, server.PlayerState27,
		server.PlayerStatePoint, server.PlayerState29, server.PlayerStateLaugh,
		server.PlayerState30, server.PlayerState32:
		animFrame = update.Field59_0
	}
	if update.Field40_0 != 0 && s.Rand.Logic.IntClamp(0, 10) >= 8 {
		state = 50
		animFrame = 0xff
	}
	if (update.State == server.PlayerState3 || update.State == server.PlayerState4) && unit.Field131 == 16 {
		state = 51
		animFrame = 0xff
	}
	if update.State != server.PlayerState30 && update.Field41 != 0 {
		first, count := s.PlayerAnimFrames(52)
		elapsed := s.Frame() - update.Field41
		frame := elapsed / uint32(count+1)
		if int(frame) >= first || elapsed >= 4 {
			update.Field41 = 0
		} else {
			state = 52
			animFrame = byte(frame)
		}
	}

	var out [12]byte
	out[0] = byte(netmsg.MSG_PLAYER_OBJ)
	binary.LittleEndian.PutUint16(out[1:], uint16(s.GetUnitNetCode(unit)))
	binary.LittleEndian.PutUint16(out[3:], unit.TypeInd)
	binary.LittleEndian.PutUint16(out[5:], uint16(int32(math.RoundToEven(float64(unit.PosVec.X)))))
	binary.LittleEndian.PutUint16(out[7:], uint16(int32(math.RoundToEven(float64(unit.PosVec.Y)))))
	out[9] = playerObjectDirection518C30(unit.Direction1) << 4
	out[10] = animFrame
	out[11] = state
	return out
}

// playerReportVitalsNative4D9900 restores the immediate health and mana
// change-reporting tail of GAME.EXE 004D9900. The reported caches are loaded
// again after each callback because the original does the same after sending.
func playerReportVitalsNative4D9900(
	unit *server.Object,
	reportHealth func(byte, *server.Object),
	reportMana func(byte, *server.Object),
) {
	if unit == nil || !unit.ObjClass.Has(object.ClassPlayer) || unit.UpdateData == nil {
		return
	}
	update := (*server.PlayerUpdateData)(unit.UpdateData)
	if update.Player == nil {
		return
	}
	if health := unit.HealthData; health != nil && health.Cur != update.Field2_1 {
		reportHealth(update.Player.PlayerInd, unit)
		update.Field2_1 = health.Cur
	}
	if update.ManaCur != update.ManaPrev {
		if update.Player == nil {
			return
		}
		reportMana(update.Player.PlayerInd, unit)
		update.ManaPrev = update.ManaCur
	}
}

// playerHealthReportNative4D86E0 reports exact current health to the owning
// client and preserves the original Quest party-health percentage report.
func (s *Server) playerHealthReportNative4D86E0(playerInd byte, unit *server.Object) {
	if unit == nil || unit.HealthData == nil {
		return
	}
	health := unit.HealthData
	var packet [3]byte
	packet[0] = byte(netmsg.MSG_REPORT_PLAYER_HEALTH)
	binary.LittleEndian.PutUint16(packet[1:], health.Cur)
	s.Server.NetSendPacketXxx1(int(playerInd), packet[:], nil, 1)

	if noxflags.HasGame(noxflags.GameModeQuest) && health.Max != 0 {
		var teamPacket [5]byte
		teamPacket[0] = 0xc4
		teamPacket[1] = 12
		binary.LittleEndian.PutUint16(teamPacket[2:], uint16(s.Server.GetUnitNetCode(unit)))
		teamPacket[4] = byte(100 * uint32(health.Cur) / uint32(health.Max))
		s.Server.NetSendPacketXxx1(int(playerInd)|0x80, teamPacket[:], nil, 1)
	}
}

// netPlayerObjectSendNative518C30 restores the packet-producing part of
// GAME.EXE 00518C30 without interpreting Object, PlayerUpdateData, or Player
// through their Win32 byte offsets.
func (s *Server) netPlayerObjectSendNative518C30(recipient, unit *server.Object, updateStream bool) bool {
	playerReportSelf518CAF(recipient, unit, func(unit *server.Object) {
		s.Server.PlayerGoldReportSync4D9900(unit)
		playerReportVitalsNative4D9900(unit, s.playerHealthReportNative4D86E0, func(playerInd byte, unit *server.Object) {
			legacy.NetReportManaNative4D8930(s.Server, playerInd, unit)
		})
	})
	packet := s.playerObjectPacketNative518C30(unit)
	player := recipient.ControllingPlayer()
	if updateStream {
		return nox_netlist_addToMsgListSrv(player.PlayerIndex(), packet[:])
	}
	return s.NetList.AddToMsgListCli(player.PlayerIndex(), netlist.Kind1, packet[:])
}

func playerReportSelf518CAF[O comparable](recipient, unit O, report func(O)) {
	if recipient == unit {
		report(recipient)
	}
}
