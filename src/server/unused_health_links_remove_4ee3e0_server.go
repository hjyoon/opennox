package server

// healthLinksState4EE390 is the native-width form of the unused health-link
// list shared by GAME.EXE 004EE390..004EE440. The original HealthData +8/+12
// slots are fixed ABI32 dwords and must never receive truncated Go pointers.
type healthLinksState4EE390 struct {
	head     *Object
	next     map[*HealthData]*Object
	previous map[*HealthData]*Object
}

func (s *healthLinksState4EE390) storeNext(health *HealthData, next *Object) {
	if health == nil {
		panic("GAME.EXE health-link next access through nil HealthData")
	}
	if s.next == nil {
		s.next = make(map[*HealthData]*Object)
	}
	s.next[health] = next
}

func (s *healthLinksState4EE390) storePrevious(health *HealthData, previous *Object) {
	if health == nil {
		panic("GAME.EXE health-link previous access through nil HealthData")
	}
	if s.previous == nil {
		s.previous = make(map[*HealthData]*Object)
	}
	s.previous[health] = previous
}

func (s *Server) unusedHealthLinksRemoveHooks4EE3E0() unusedHealthLinksRemoveHooks4EE3E0[*Object, *HealthData] {
	return unusedHealthLinksRemoveHooks4EE3E0[*Object, *HealthData]{
		loadHealth: func(obj *Object) *HealthData {
			return obj.HealthData
		},
		loadNext: func(health *HealthData) *Object {
			if health == nil {
				panic("GAME.EXE 004EE3E0 next load through nil HealthData")
			}
			return s.healthLinks.next[health]
		},
		loadPrevious: func(health *HealthData) *Object {
			if health == nil {
				panic("GAME.EXE 004EE3E0 previous load through nil HealthData")
			}
			return s.healthLinks.previous[health]
		},
		storePrevious: s.healthLinks.storePrevious,
		storeNext:     s.healthLinks.storeNext,
		storeHead: func(obj *Object) {
			s.healthLinks.head = obj
		},
	}
}

// UnusedHealthLinksRemove4EE3E0 exposes the unreferenced original routine
// without inventing a C caller. Native object links live only in the Server
// sidecar; the fixed-width HealthData ABI32 words remain untouched.
func (s *Server) UnusedHealthLinksRemove4EE3E0(obj *Object) {
	unusedHealthLinksRemove4EE3E0(obj, s.unusedHealthLinksRemoveHooks4EE3E0())
}
