package server

import "testing"

func TestUnusedHealthLinksRemove4EE3E0NativeMiddleUsesPointerWidthSidecar(t *testing.T) {
	s := &Server{}
	removedHealth := &HealthData{field8: 0x11111111, field12: 0x22222222}
	nextHealth := &HealthData{field8: 0x33333333, field12: 0x44444444}
	previousHealth := &HealthData{field8: 0x55555555, field12: 0x66666666}
	removed := &Object{HealthData: removedHealth}
	next := &Object{HealthData: nextHealth}
	previous := &Object{HealthData: previousHealth}
	oldHead := &Object{}
	s.healthLinks.head = oldHead
	s.healthLinks.storeNext(removedHealth, next)
	s.healthLinks.storePrevious(removedHealth, previous)
	s.healthLinks.storePrevious(nextHealth, removed)
	s.healthLinks.storeNext(previousHealth, removed)

	s.UnusedHealthLinksRemove4EE3E0(removed)

	if s.healthLinks.previous[nextHealth] != previous || s.healthLinks.next[previousHealth] != next {
		t.Fatalf("repairs = next.previous %p, previous.next %p", s.healthLinks.previous[nextHealth], s.healthLinks.next[previousHealth])
	}
	if s.healthLinks.next[removedHealth] != next || s.healthLinks.previous[removedHealth] != previous {
		t.Fatalf("removed sidecar links changed: next %p previous %p", s.healthLinks.next[removedHealth], s.healthLinks.previous[removedHealth])
	}
	if s.healthLinks.head != oldHead {
		t.Fatalf("head = %p, want %p", s.healthLinks.head, oldHead)
	}
	if removedHealth.field8 != 0x11111111 || removedHealth.field12 != 0x22222222 ||
		nextHealth.field8 != 0x33333333 || nextHealth.field12 != 0x44444444 ||
		previousHealth.field8 != 0x55555555 || previousHealth.field12 != 0x66666666 {
		t.Fatal("native removal changed fixed-width ABI32 link words")
	}
}

func TestUnusedHealthLinksRemove4EE3E0NativeHeadRepairsSuccessor(t *testing.T) {
	s := &Server{}
	removedHealth := &HealthData{field8: 0x11111111, field12: 0x22222222}
	nextHealth := &HealthData{field8: 0x33333333, field12: 0x44444444}
	removed := &Object{HealthData: removedHealth}
	next := &Object{HealthData: nextHealth}
	s.healthLinks.head = removed
	s.healthLinks.storeNext(removedHealth, next)
	s.healthLinks.storePrevious(nextHealth, removed)

	s.UnusedHealthLinksRemove4EE3E0(removed)

	if s.healthLinks.head != next || s.healthLinks.previous[nextHealth] != nil {
		t.Fatalf("head removal = head %p next.previous %p", s.healthLinks.head, s.healthLinks.previous[nextHealth])
	}
	if s.healthLinks.next[removedHealth] != next {
		t.Fatalf("removed next = %p, want %p", s.healthLinks.next[removedHealth], next)
	}
	if removedHealth.field8 != 0x11111111 || removedHealth.field12 != 0x22222222 ||
		nextHealth.field8 != 0x33333333 || nextHealth.field12 != 0x44444444 {
		t.Fatal("native head removal changed fixed-width ABI32 link words")
	}
}

func TestUnusedHealthLinksRemove4EE3E0NativeNullGuards(t *testing.T) {
	s := &Server{}
	head := &Object{}
	s.healthLinks.head = head

	s.UnusedHealthLinksRemove4EE3E0(nil)
	s.UnusedHealthLinksRemove4EE3E0(&Object{})

	if s.healthLinks.head != head || len(s.healthLinks.next) != 0 || len(s.healthLinks.previous) != 0 {
		t.Fatalf("null guards changed state: head %p next %d previous %d", s.healthLinks.head, len(s.healthLinks.next), len(s.healthLinks.previous))
	}
}

func TestUnusedHealthLinksRemove4EE3E0NativePreservesIntermediateNullHealthFault(t *testing.T) {
	t.Run("next", func(t *testing.T) {
		s := &Server{}
		health := &HealthData{}
		removed := &Object{HealthData: health}
		next := &Object{}
		s.healthLinks.storeNext(health, next)

		defer func() {
			if recover() == nil {
				t.Fatal("expected nil successor HealthData panic")
			}
		}()
		s.UnusedHealthLinksRemove4EE3E0(removed)
	})

	t.Run("previous", func(t *testing.T) {
		s := &Server{}
		health := &HealthData{}
		nextHealth := &HealthData{}
		removed := &Object{HealthData: health}
		next := &Object{HealthData: nextHealth}
		previous := &Object{}
		s.healthLinks.storeNext(health, next)
		s.healthLinks.storePrevious(health, previous)

		defer func() {
			if recover() == nil {
				t.Fatal("expected nil predecessor HealthData panic")
			}
			if s.healthLinks.previous[nextHealth] != previous {
				t.Fatalf("successor repair before fault = %p, want %p", s.healthLinks.previous[nextHealth], previous)
			}
		}()
		s.UnusedHealthLinksRemove4EE3E0(removed)
	})
}

func TestUnusedHealthLinksReset4EE390ServerBindingClearsNativeSidecar(t *testing.T) {
	s := &Server{}
	health := &HealthData{field8: 0x11223344, field12: 0x55667788}
	obj := &Object{HealthData: health}
	s.healthLinks.storeNext(health, &Object{})
	s.healthLinks.storePrevious(health, &Object{})

	if got := s.UnusedHealthLinksReset4EE390(obj); got != health {
		t.Fatalf("result = %p, want %p", got, health)
	}
	if s.healthLinks.next[health] != nil || s.healthLinks.previous[health] != nil {
		t.Fatalf("native links = next %p previous %p, want nil", s.healthLinks.next[health], s.healthLinks.previous[health])
	}
	if health.field8 != 0 || health.field12 != 0 {
		t.Fatalf("ABI32 compatibility links = %08x/%08x, want zero", health.field8, health.field12)
	}
}

func TestUnusedHealthLinksHead4EE430ServerBindingPreservesPointerIdentity(t *testing.T) {
	s := &Server{}
	if got := s.UnusedHealthLinksHead4EE430(); got != nil {
		t.Fatalf("empty head = %p, want nil", got)
	}

	head := &Object{}
	s.healthLinks.head = head
	if got := s.UnusedHealthLinksHead4EE430(); got != head {
		t.Fatalf("head = %p, want %p", got, head)
	}
}

func TestUnusedHealthLinksNext4EE440ServerBindingPreservesPointerIdentity(t *testing.T) {
	s := &Server{}
	if got := s.UnusedHealthLinksNext4EE440(nil); got != nil {
		t.Fatalf("null object result = %p, want nil", got)
	}

	health := &HealthData{field8: 0x11223344}
	obj := &Object{HealthData: health}
	next := &Object{}
	s.healthLinks.storeNext(health, next)
	if got := s.UnusedHealthLinksNext4EE440(obj); got != next {
		t.Fatalf("next = %p, want %p", got, next)
	}
	if health.field8 != 0x11223344 {
		t.Fatalf("ABI32 next word = %08x, want unchanged", health.field8)
	}
}

func TestUnusedHealthLinksNext4EE440ServerBindingPreservesNullHealthFault(t *testing.T) {
	s := &Server{}
	defer func() {
		if recover() == nil {
			t.Fatal("expected nil HealthData panic")
		}
	}()
	s.UnusedHealthLinksNext4EE440(&Object{})
}
