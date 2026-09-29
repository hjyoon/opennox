package server

import (
	"encoding/binary"
	"math"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/ntype"
)

const (
	questStatsPacketSize4D6770 = 90
	questStatsMaxPlayers4D6770 = 6
	questStatsMaxScore4D6540   = uint32(999999999)
)

// questStatsFloatToInt4D66E0 models nox_float2int's x87 FISTP conversion.
// The original runtime uses round-to-nearest-even and returns the x87 integer
// indefinite value for NaN and values outside the signed 32-bit range.
func questStatsFloatToInt4D66E0(value float32) int32 {
	v := float64(value)
	if math.IsNaN(v) || v >= 1<<31 || v < -(1<<31) {
		return math.MinInt32
	}
	return int32(math.RoundToEven(v))
}

// QuestScore4D66E0 preserves GAME.EXE 004D66E0's binary32 spills while
// reading the original eight-byte score exponent as a Go float64. This avoids
// interpreting the Win32 data blob as a 16-byte Darwin/arm64 long double.
func QuestScore4D66E0(generators, secrets, monsters, stage uint32, exponent float64) uint32 {
	stageScale := float32(math.Pow(float64(stage), exponent))
	weighted := float64(generators)*10.0 + float64(secrets)*35.0 + float64(monsters)*0.1
	score := float32(float64(stageScale) * weighted)
	return uint32(questStatsFloatToInt4D66E0(score))
}

func (s *Server) questPlayerCount4E3CE0() int {
	count := 0
	for unit := s.Players.FirstUnit(); unit != nil; unit = s.questNextPlayerUnit4DA7F0(unit) {
		player := (*PlayerUpdateData)(unit.UpdateData).Player
		if noxflags.HasGame(noxflags.GameHost) &&
			noxflags.HasEngine(noxflags.EngineNoRendering) &&
			player.PlayerInd == HostPlayerIndex {
			continue
		}
		if player.Field4792 == 1 {
			count++
		}
	}
	return count
}

// QuestPlayerScore4D6540 computes the Quest scoreboard value without routing
// Object, PlayerUpdateData, or Player pointers through the retained 32-bit C
// ABI. Unsigned accumulation and binary32 scaling match GAME.EXE 004D6540.
func (s *Server) QuestPlayerScore4D6540(ind ntype.PlayerInd, exponent float64) uint32 {
	target := s.Players.ByInd(ind)
	if target == nil || target.Field4792 == 0 {
		return 0
	}

	var result uint32
	if s.questPlayerCount4E3CE0() == 1 {
		result = QuestScore4D66E0(
			target.field4668,
			target.field4672,
			target.field4664,
			target.field4688,
			exponent,
		)
	} else {
		var (
			maxBase         uint32
			totalGenerators uint32
			totalSecrets    uint32
			stage           = uint32(1)
		)
		for unit := s.Players.FirstUnit(); unit != nil; unit = s.questNextPlayerUnit4DA7F0(unit) {
			player := (*PlayerUpdateData)(unit.UpdateData).Player
			base := QuestScore4D66E0(player.field4668, player.field4672, 0, player.field4688, exponent)
			if base > maxBase {
				maxBase = base
			}
			totalGenerators += player.field4668
			totalSecrets += player.field4672
			stage = player.field4688
		}
		teamScore := QuestScore4D66E0(totalGenerators, totalSecrets, 0, stage, exponent)
		scale := float32(1)
		if maxBase > 0 {
			scale = float32(float64(teamScore) / float64(maxBase))
		}
		playerScore := QuestScore4D66E0(
			target.field4668,
			target.field4672,
			target.field4664,
			target.field4688,
			exponent,
		)
		result = uint32(questStatsFloatToInt4D66E0(float32(float64(playerScore) * float64(scale))))
	}
	if result > questStatsMaxScore4D6540 {
		return questStatsMaxScore4D6540
	}
	return result
}

// SendQuestStats4D6770 builds the fixed F0/12 Quest scoreboard packet while
// retaining native-width player and unit pointers throughout traversal.
func (s *Server) SendQuestStats4D6770(recipient ntype.PlayerInd, questCounter uint16, exponent float64) int {
	var packet [questStatsPacketSize4D6770]byte
	packet[0] = 0xf0
	packet[1] = 12
	binary.LittleEndian.PutUint16(packet[2:4], questCounter)

	recipientPlayer := s.Players.ByInd(recipient)
	entries := 0
	for unit := s.Players.FirstUnit(); unit != nil; unit = s.questNextPlayerUnit4DA7F0(unit) {
		player := (*PlayerUpdateData)(unit.UpdateData).Player
		if player.Field4792 != 1 || entries >= questStatsMaxPlayers4D6770 {
			continue
		}
		off := 6 + entries*14
		binary.LittleEndian.PutUint16(packet[off:off+2], uint16(unit.NetCode))
		binary.LittleEndian.PutUint16(packet[off+2:off+4], uint16(player.field4668))
		binary.LittleEndian.PutUint16(packet[off+4:off+6], uint16(player.field4672))
		binary.LittleEndian.PutUint16(packet[off+6:off+8], uint16(player.field4680))
		binary.LittleEndian.PutUint16(packet[off+8:off+10], uint16(player.field4664))
		binary.LittleEndian.PutUint32(packet[off+10:off+14], s.QuestPlayerScore4D6540(ntype.PlayerInd(player.PlayerInd), exponent))
		entries++
		if recipientPlayer != nil {
			binary.LittleEndian.PutUint16(packet[4:6], uint16(recipientPlayer.field4688))
		}
	}
	return s.NetSendPacketXxx1(int(recipient), packet[:], nil, 1)
}
