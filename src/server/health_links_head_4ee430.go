package server

// healthLinksHead4EE430 preserves the single global load performed by
// GAME.EXE 004EE430. The value returned by that load is returned unchanged.
func healthLinksHead4EE430[O any](loadHead func() O) O {
	return loadHead()
}
