package server

import "github.com/opennox/libs/object"

// decayState511660 owns native-width links for the original decay list. The
// ABI32 object slot at offset 468 is shared with the pending-object list and
// cannot hold a pointer on 64-bit targets. Keeping decay links here preserves
// both pointer width and the original mutual exclusion through ObjFlags.
type decayState511660 struct {
	head *Object
	next map[*Object]*Object
}

func (s *Server) decayHooks511660(obj *Object, delay uint32, delayedDelete func(*Object)) decayHooks511660[*Object] {
	return decayHooks511660[*Object]{
		loadSetObjectArg: func() *Object {
			return obj
		},
		loadSetDelayArg: func() uint32 {
			return delay
		},
		loadObjectFlags: func(obj *Object) uint32 {
			return uint32(obj.ObjFlags)
		},
		storeObjectFlags: func(obj *Object, value uint32) {
			obj.ObjFlags = object.Flags(value)
		},
		loadFrame: s.Frame,
		loadDeadline: func(obj *Object) uint32 {
			return obj.Field34
		},
		storeDeadline: func(obj *Object, value uint32) {
			obj.Field34 = value
		},
		loadHead: func() *Object {
			return s.decay.head
		},
		storeHead: func(obj *Object) {
			s.decay.head = obj
		},
		loadNext: func(obj *Object) *Object {
			return s.decay.next[obj]
		},
		storeNext: func(obj, next *Object) {
			if s.decay.next == nil {
				s.decay.next = make(map[*Object]*Object)
			}
			s.decay.next[obj] = next
		},
		loadHolder: func(obj *Object) *Object {
			return obj.InvHolder
		},
		loadDeleteFlags: func(obj *Object) uint32 {
			return obj.Field5
		},
		storeDeleteFlags: func(obj *Object, value uint32) {
			obj.Field5 = value
		},
		delayedDelete: delayedDelete,
	}
}

// DecaySetTime511660 schedules obj after delay frames and returns the same
// 32-bit flags value left in EAX by GAME.EXE 00511660.
func (s *Server) DecaySetTime511660(obj *Object, delay uint32) uint32 {
	return decaySetTime511660(s.decayHooks511660(obj, delay, nil))
}

// DecayRemove5116F0 removes obj from the native decay list. All active callers
// discard the original mixed word/pointer EAX result, so it stays inside the
// typed contract instead of being narrowed at a public boundary.
func (s *Server) DecayRemove5116F0(obj *Object) {
	decayRemove5116F0(obj, s.decayHooks511660(nil, 0, nil))
}

// DecayTick511750 processes held and expired decay entries in original list
// order. delayedDelete is injected by the outer server to avoid a package
// cycle while preserving the native object pointer.
func (s *Server) DecayTick511750(delayedDelete func(*Object)) {
	decayTick511750(s.decayHooks511660(nil, 0, delayedDelete))
}

// DecayDestroy5117B0 removes every native decay entry during session teardown.
func (s *Server) DecayDestroy5117B0() {
	decayDestroy5117B0(s.decayHooks511660(nil, 0, nil))
}
