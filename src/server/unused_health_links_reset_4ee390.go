package server

type unusedHealthLinksResetResultKind4EE390 uint8

const (
	unusedHealthLinksResetObject4EE390 unusedHealthLinksResetResultKind4EE390 = iota
	unusedHealthLinksResetHealth4EE390
)

// unusedHealthLinksResetResult4EE390 separates the two pointer domains that
// GAME.EXE leaves in EAX. A null object and the otherwise unreachable
// null-HealthData tail return the object argument; the normal path returns a
// live reloaded HealthData pointer.
type unusedHealthLinksResetResult4EE390[O, H any] struct {
	kind   unusedHealthLinksResetResultKind4EE390
	object O
	health H
}

type unusedHealthLinksResetHooks4EE390[O, H any] struct {
	loadHealth                func(O) H
	storeHealthPrevious       func(H, O)
	storeHealthNext           func(H, O)
	storeAbsoluteNullPrevious func()
	loadHead                  func() O
	storeObjectPrevious       func(O, O)
	storeHead                 func(O)
}

// unusedHealthLinksReset4EE390 preserves the exact instruction order of the
// unreferenced GAME.EXE routine at 004EE390. A valid HealthData record has its
// +12 link cleared before the object HealthData pointer is reloaded and that
// record's +8 link is cleared.
//
// The original null-HealthData branch first writes zero to absolute address
// 0x0000000C. That access faults in the supported process memory model before
// any list-head access. The remaining hooks intentionally describe the bytes
// after that fault as well, so tests can keep the complete original stream
// sealed without pretending that the branch is a usable list insertion API.
func unusedHealthLinksReset4EE390[O, H comparable](
	obj O,
	hooks unusedHealthLinksResetHooks4EE390[O, H],
) unusedHealthLinksResetResult4EE390[O, H] {
	result := unusedHealthLinksResetResult4EE390[O, H]{
		kind:   unusedHealthLinksResetObject4EE390,
		object: obj,
	}
	var nilObject O
	if obj == nilObject {
		return result
	}

	health := hooks.loadHealth(obj)
	var nilHealth H
	if health != nilHealth {
		hooks.storeHealthPrevious(health, nilObject)
		health = hooks.loadHealth(obj)
		hooks.storeHealthNext(health, nilObject)
		return unusedHealthLinksResetResult4EE390[O, H]{
			kind:   unusedHealthLinksResetHealth4EE390,
			health: health,
		}
	}

	hooks.storeAbsoluteNullPrevious()
	health = hooks.loadHealth(obj)
	head := hooks.loadHead()
	hooks.storeHealthNext(health, head)
	head = hooks.loadHead()
	if head != nilObject {
		hooks.storeObjectPrevious(head, obj)
	}
	hooks.storeHead(obj)
	return result
}
