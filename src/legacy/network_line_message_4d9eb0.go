package legacy

// netLineMessagePlayerIndex4D9EB0 preserves the two dependent pointer reads
// performed by GAME.EXE at UpdateData+276 and Player+2064. Reading them after
// formatting preserves the original evaluation order.
func netLineMessagePlayerIndex4D9EB0[U, P any](
	updateData U,
	player func(U) P,
	index func(P) byte,
) byte {
	return index(player(updateData))
}
