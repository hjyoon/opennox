package legacy

/*
#include "GAME3_2.h"
#include "GAME5_2.h"
#include "cgo_const.h"
*/
import "C"

import (
	"encoding/binary"
	"math"
	"unsafe"

	"github.com/opennox/libs/noxnet/discover"
	"github.com/opennox/libs/noxnet/netmsg"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/server"
)

const (
	serverDiscoveryHeaderSize  = 3
	serverDiscoveryMapOffset   = 10
	serverDiscoveryMapSize     = 9
	serverDiscoveryTokenOffset = 8
	serverDiscoveryRequestSize = serverDiscoveryTokenOffset + 4
)

func serverDiscoveryInfo554040(token uint32) *discover.MsgServerInfo {
	settings := (*server.Settings)(sub416640Native())
	game := (*server.Settings2)(cliGamedataNative416590(0))
	if sub_43AF30() == 0 || sub_4D6F30() != 0 {
		return nil
	}
	maxPlayers := headSettingsGetLimit409470()
	players := nox_common_playerInfoCount_416F40()
	if noxflags.HasEngine(noxflags.EngineNoRendering) {
		maxPlayers--
		players--
	}
	name := GoString((*C.char)(headSettingsName409470()))
	flags := uint32(noxflags.GetGame())
	var stage uint16
	if nox_xxx_isQuest_4D6F50() != 0 {
		flags = flags&^uint32(noxflags.GameModeChat) | uint32(noxflags.GameModeQuest)
		stage = uint16(nox_game_getQuestStage_4E3CC0())
	}
	video := byte(sub_43BE50_get_video_mode_id())
	mapName := GoString((*C.char)(headSettingsMap409470()))
	return newServerDiscoveryInfo554040(settings, game, byte(players), byte(maxPlayers),
		name, mapName, flags, stage, video, token)
}

func newServerDiscoveryInfo554040(settings *server.Settings, game *server.Settings2,
	players, maxPlayers byte, name, mapName string, flags uint32, stage uint16, video byte, token uint32,
) *discover.MsgServerInfo {
	msg := &discover.MsgServerInfo{
		PlayersCur: players,
		PlayersMax: maxPlayers,
		MapName:    mapName,
		Status1:    byte(settings.Flags102) | video,
		Status2:    byte(settings.Flags100),
		Flags:      uint16(flags),
		Token:      token,
		ServerName: name,
	}
	msg.Unk2[0] = settings.Field101 & 15
	msg.Unk2[1] = settings.Field101 >> 4
	// The map name starts where the fourth weapon byte would land.
	copy(msg.Unk2[2:], game.Field44[:3])
	msg.Unk19[0] = byte(settings.Flags100 & server.ServerClosed)
	binary.LittleEndian.PutUint32(msg.Unk19[3:], settings.Version48)
	binary.LittleEndian.PutUint16(msg.Unk27[:], uint16(flags>>16))
	binary.LittleEndian.PutUint32(msg.Unk29[:], game.Field48)
	copy(msg.Unk29[4:6], settings.RestrictedPingMind105[:])
	copy(msg.Unk29[6:8], settings.RestrictedPingMax107[:])
	binary.LittleEndian.PutUint32(msg.Unk37[:], settings.Field44)
	for i, spells := range game.Field24.Vals {
		binary.LittleEndian.PutUint32(msg.Unk45[4*i:], spells)
	}
	binary.LittleEndian.PutUint16(msg.Unk65[:], stage)
	return msg
}

func encodeServerDiscovery554040(dst []byte, info *discover.MsgServerInfo) int {
	if info == nil || len(dst) < serverDiscoveryHeaderSize+info.EncodeSize() {
		return 0
	}
	length, err := info.Encode(dst[serverDiscoveryHeaderSize:])
	if err != nil {
		return 0
	}
	// The original also used the ninth map byte when the name filled the field.
	copy(dst[serverDiscoveryMapOffset:serverDiscoveryMapOffset+serverDiscoveryMapSize], info.MapName)
	dst[0] = 0
	dst[1] = 0
	dst[2] = byte(netmsg.MSG_SERVER_INFO)
	return serverDiscoveryHeaderSize + length
}

func Nox_server_makeServerInfoPacket_554040(src, dst []byte) int {
	if len(src) < serverDiscoveryRequestSize {
		return 0
	}
	token := binary.LittleEndian.Uint32(src[serverDiscoveryTokenOffset:])
	return encodeServerDiscovery554040(dst, serverDiscoveryInfo554040(token))
}

//export nox_server_makeServerInfoPacket_554040
func nox_server_makeServerInfoPacket_554040(src *C.nox_cgo_const_char, size C.int, dst *C.char) C.uint {
	if size < serverDiscoveryRequestSize || src == nil || dst == nil {
		return 0
	}
	request := unsafe.Slice((*byte)(unsafe.Pointer(src)), serverDiscoveryRequestSize)
	info := serverDiscoveryInfo554040(binary.LittleEndian.Uint32(request[serverDiscoveryTokenOffset:]))
	if info == nil {
		return 0
	}
	output := unsafe.Slice((*byte)(unsafe.Pointer(dst)), serverDiscoveryHeaderSize+info.EncodeSize())
	return C.uint(encodeServerDiscovery554040(output, info))
}

//export sub_554290
func sub_554290() C.uint { return C.uint(Sub_554290()) }

func Sub_554290() uint32 {
	var recordCount int32
	var minimum uint32 = math.MaxUint32
	players := &GetServer().S().Players
	for player := players.First(); player != nil; player = players.Next(player) {
		if player.PlayerInd == byte(server.HostPlayerIndex) || sub_554240(int32(player.PlayerInd)) <= 0 {
			continue
		}
		minimum = min(minimum, uint32(sub_554240(int32(player.PlayerInd))))
		recordCount++
	}
	if recordCount == 0 {
		return 0
	}
	return minimum
}

//export sub_554300
func sub_554300() C.int { return C.int(int32(Sub_554300())) }

func Sub_554300() uint32 {
	var total, count int32
	players := &GetServer().S().Players
	for player := players.First(); player != nil; player = players.Next(player) {
		if player.PlayerInd == byte(server.HostPlayerIndex) || sub_554240(int32(player.PlayerInd)) <= 0 {
			continue
		}
		total += sub_554240(int32(player.PlayerInd))
		count++
	}
	if count == 0 {
		return 0
	}
	return uint32(total / count)
}
