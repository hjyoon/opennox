package opennox

import "github.com/opennox/opennox/v1/server"

type objectDeleteFinish4E5E80Hooks struct {
	transferSlaves  func(*server.Object)
	clearOwner      func(*server.Object)
	clearActivators func(*server.Object)
	decay           func(*server.Object)
	dropAllItems    func(*server.Object)
	finalize        func(*server.Object)
	free            func(*server.Object)
}

// objectDeleteFinish_4E5E80 preserves the original teardown sequence. Each
// callback may invalidate relationships used by the callbacks that follow it.
func objectDeleteFinish_4E5E80(obj *server.Object, hooks objectDeleteFinish4E5E80Hooks) {
	hooks.transferSlaves(obj)
	hooks.clearOwner(obj)
	hooks.clearActivators(obj)
	hooks.decay(obj)
	hooks.dropAllItems(obj)
	hooks.finalize(obj)
	hooks.free(obj)
}
