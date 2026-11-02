package server

import "encoding/binary"

// SecretWallStateNative4DF120 preserves the unchecked public C entry contract.
// The ordinary campaign sender retains its own admission and player snapshot.
func (wall *Wall) SecretWallStateNative4DF120(open bool, runtime SecretWallPacketRuntime4DF120) int {
	var packet [3]byte
	packet[0] = 59
	if open {
		packet[0] = 60
	}
	binary.LittleEndian.PutUint16(packet[1:], wall.Field10)
	return int(sendWallPacketToPlayersNative4DF0A0(packet[:], runtime))
}
