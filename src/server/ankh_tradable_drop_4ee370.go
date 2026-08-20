package server

// ankhTradableDropHooks4EE370 exposes GAME.EXE 004EE370's exact stack-
// argument load order. All three values are cached before DefaultDrop runs.
type ankhTradableDropHooks4EE370[O, P any] struct {
	loadPointArg func() P
	loadItemArg  func() O
	loadOwnerArg func() O
	defaultDrop  func(O, O, P) int32
}

// ankhTradableDrop4EE370 preserves the original forwarding thunk. It adds no
// nil guard and returns DefaultDrop's whole 32-bit result without
// canonicalizing it.
func ankhTradableDrop4EE370[O, P any](hooks ankhTradableDropHooks4EE370[O, P]) int32 {
	point := hooks.loadPointArg()
	item := hooks.loadItemArg()
	owner := hooks.loadOwnerArg()
	return hooks.defaultDrop(owner, item, point)
}
