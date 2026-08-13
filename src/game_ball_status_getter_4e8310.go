package opennox

// gameBallStatusGetter4E8310 preserves GAME.EXE 004E8310. The original
// function returns the address of the four-byte GameBall status record without
// reading it. In particular, a nil test pointer must be returned without a
// fault; production always supplies the mapped 0x0075318C record.
func gameBallStatusGetter4E8310(record *gameBallStatusRecord4E8290) *gameBallStatusRecord4E8290 {
	return record
}
