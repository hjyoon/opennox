package server

type doorCloseHooks4E8340[O, U, P any] struct {
	class         func(O) uint32
	updateData    func(O) U
	tileX         func(U) int32
	targetX       func(P) int32
	tileY         func(U) int32
	targetY       func(P) int32
	storeLockCode func(U, uint8)
	quest         func() int32
	questSync     func(O) int32
}

// doorClose4E8340 preserves GAME.EXE 004E8340. The original tests only the
// low class byte, compares X before reading Y, clears the lock before checking
// Quest mode, and ignores the synchronization result.
func doorClose4E8340[O, U, P any](obj O, target P, hooks doorCloseHooks4E8340[O, U, P]) {
	if uint8(hooks.class(obj))&0x80 == 0 {
		return
	}
	update := hooks.updateData(obj)
	tileX := hooks.tileX(update)
	targetX := hooks.targetX(target)
	if tileX != targetX {
		return
	}
	tileY := hooks.tileY(update)
	targetY := hooks.targetY(target)
	if tileY != targetY {
		return
	}
	hooks.storeLockCode(update, 0)
	if hooks.quest() != 0 {
		_ = hooks.questSync(obj)
	}
}

type doorQuestSyncHooks4E8390[O, U any] struct {
	updateData    func(O) U
	storeSyncByte func(U, uint8)
	sendExtent    func(int32, O) int32
}

// doorQuestSync4E8390 preserves GAME.EXE 004E8390. The status byte is stored
// before recipient 255 is sent, and the signed 32-bit send result is returned
// without normalization.
func doorQuestSync4E8390[O, U any](obj O, hooks doorQuestSyncHooks4E8390[O, U]) int32 {
	update := hooks.updateData(obj)
	hooks.storeSyncByte(update, 1)
	return hooks.sendExtent(255, obj)
}

type doorExtentPacketHooks4D6A20[O any] struct {
	extent func(O) uint16
	send   func(recipient int32, packet [4]byte, relatedObject uintptr, removeIfDisconnected int32) int32
}

// doorExtentPacket4D6A20 preserves GAME.EXE 004D6A20. The packet is the
// little-endian opcode 0x0ff0 followed by the object's low 16-bit extent.
func doorExtentPacket4D6A20[O any](recipient int32, obj O, hooks doorExtentPacketHooks4D6A20[O]) int32 {
	extent := hooks.extent(obj)
	packet := [4]byte{0xf0, 0x0f, byte(extent), byte(extent >> 8)}
	return hooks.send(recipient, packet, 0, 1)
}
