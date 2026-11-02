package server

import "encoding/binary"

// WallSendDestroyedNative4DF0A0 implements the unchecked public C contract.
// Direct sends preserve the original signed C-int result; broadcasts return
// zero after live traversal, independently of the individual Send results.
func (wall *Wall) WallSendDestroyedNative4DF0A0(recipient int32, runtime SecretWallPacketRuntime4DF120) int32 {
	var packet [3]byte
	packet[0] = 58
	binary.LittleEndian.PutUint16(packet[1:], wall.Field10)
	if recipient != 32 {
		return int32(runtime.Send(int(recipient), packet[:], nil, 1))
	}
	return sendWallPacketToPlayersNative4DF0A0(packet[:], runtime)
}

func sendWallPacketToPlayersNative4DF0A0(packet []byte, runtime SecretWallPacketRuntime4DF120) int32 {
	for unit := runtime.FirstUnit(); unit != nil; unit = runtime.NextUnit(unit) {
		// Raw packet entries neither classify the object nor guard its pointers.
		player := (*PlayerUpdateData)(unit.UpdateData).Player
		runtime.Send(int(player.PlayerInd), packet, nil, 1)
	}
	return 0
}
