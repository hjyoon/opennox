package server

const objectForceDropRadius4ED930 = float32(50)

// objectForceDropHooks4ED930 exposes the exact wrapper boundary around the
// random reachable-point helper at 004ED970 and the drop dispatcher at
// 004ED790. The owner is cached before the helper. The item remains delayed
// until the helper has initialized the local target.
type objectForceDropHooks4ED930[O, P any] struct {
	loadOwnerArg    func() O
	randomReachable func(radius float32, owner O, output *P) *P
	loadItemArg     func() O
	dispatch        func(owner, item O, point *P) int32
}

// objectForceDrop4ED930 preserves GAME.EXE 004ED930. The return pointer from
// 004ED970 is deliberately ignored; 004ED790 receives the address of the same
// local output object passed to the helper and its full signed result escapes.
func objectForceDrop4ED930[O, P any](hooks objectForceDropHooks4ED930[O, P]) int32 {
	owner := hooks.loadOwnerArg()
	var point P
	_ = hooks.randomReachable(objectForceDropRadius4ED930, owner, &point)
	item := hooks.loadItemArg()
	return hooks.dispatch(owner, item, &point)
}
