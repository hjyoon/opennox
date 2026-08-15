package legacy

import "testing"

func TestSub4EC5B0DefaultState(t *testing.T) {
	// The package defaults model GAME.EXE's valid nil allocator/list state.
	// This also keeps the complete native CGo binding live in cross-linked
	// legacy test products instead of allowing dead-code elimination to hide it.
	Sub_4EC5B0()
}
