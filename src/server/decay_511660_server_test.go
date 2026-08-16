package server

import (
	"testing"
	"unsafe"
)

func TestDecayNativeLayout511660(t *testing.T) {
	want := map[string]uintptr{
		"Object.ObjFlags":  16,
		"Object.Field5":    20,
		"Object.Field34":   136,
		"Object.Field117":  468,
		"Object.InvHolder": 492,
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		want = map[string]uintptr{
			"Object.ObjFlags":  20,
			"Object.Field5":    24,
			"Object.Field34":   140,
			"Object.Field117":  484,
			"Object.InvHolder": 520,
		}
	}
	got := map[string]uintptr{
		"Object.ObjFlags":  unsafe.Offsetof(Object{}.ObjFlags),
		"Object.Field5":    unsafe.Offsetof(Object{}.Field5),
		"Object.Field34":   unsafe.Offsetof(Object{}.Field34),
		"Object.Field117":  unsafe.Offsetof(Object{}.Field117),
		"Object.InvHolder": unsafe.Offsetof(Object{}.InvHolder),
	}
	for name, expected := range want {
		if got[name] != expected {
			t.Fatalf("%s offset = %d, want %d", name, got[name], expected)
		}
	}
}

func TestDecayNativeSetTime511660UsesSidecarAndStableEqualOrder(t *testing.T) {
	s := &Server{}
	s.SetFrame(100)
	first := &Object{Field117: 0x11111111}
	second := &Object{Field117: 0x22222222}
	equal := &Object{Field117: 0x33333333}

	if got := s.DecaySetTime511660(first, 5); got != decayListedFlag511660 {
		t.Fatalf("first result = %08x", got)
	}
	if got := s.DecaySetTime511660(second, 3); got != decayListedFlag511660 {
		t.Fatalf("second result = %08x", got)
	}
	if got := s.DecaySetTime511660(equal, 5); got != decayListedFlag511660 {
		t.Fatalf("equal result = %08x", got)
	}
	if s.decay.head != second || s.decay.next[second] != first || s.decay.next[first] != equal || s.decay.next[equal] != nil {
		t.Fatalf("order = %p -> %p -> %p -> %p", s.decay.head, s.decay.next[second], s.decay.next[first], s.decay.next[equal])
	}
	if first.Field34 != 105 || second.Field34 != 103 || equal.Field34 != 105 {
		t.Fatalf("deadlines = %d/%d/%d", first.Field34, second.Field34, equal.Field34)
	}
	if first.Field117 != 0x11111111 || second.Field117 != 0x22222222 || equal.Field117 != 0x33333333 {
		t.Fatalf("ABI32 shared fields changed: %08x/%08x/%08x", first.Field117, second.Field117, equal.Field117)
	}

	s.SetFrame(102)
	s.DecaySetTime511660(equal, 0)
	if s.decay.head != equal || s.decay.next[equal] != second || s.decay.next[second] != first || s.decay.next[first] != nil {
		t.Fatalf("rescheduled order = %p -> %p -> %p -> %p", s.decay.head, s.decay.next[equal], s.decay.next[second], s.decay.next[first])
	}
	if equal.Field34 != 102 || uint32(equal.ObjFlags)&decayListedFlag511660 == 0 {
		t.Fatalf("rescheduled state = deadline %d flags %08x", equal.Field34, uint32(equal.ObjFlags))
	}
}

func TestDecayNativeSetTime511660PendingObjectIsUntouched(t *testing.T) {
	s := &Server{}
	s.SetFrame(50)
	obj := &Object{ObjFlags: 0x00010040, Field34: 17, Field117: 0xdeadbeef}
	if got := s.DecaySetTime511660(obj, 100); got != decayPendingFlag511660|0x40 {
		t.Fatalf("result = %08x", got)
	}
	if s.decay.head != nil || obj.Field34 != 17 || obj.Field117 != 0xdeadbeef {
		t.Fatalf("pending object changed: head %p deadline %d field117 %08x", s.decay.head, obj.Field34, obj.Field117)
	}
}

func TestDecayNativeTick511750RemovesHeldDeletesDueAndStopsAtFuture(t *testing.T) {
	s := &Server{}
	s.SetFrame(10)
	owner := &Object{}
	held := &Object{InvHolder: owner, Field117: 0x11111111}
	due := &Object{Field5: 0x12340001, Field117: 0x22222222}
	future := &Object{Field117: 0x33333333}

	s.DecaySetTime511660(held, 0)
	s.DecaySetTime511660(due, 0)
	s.DecaySetTime511660(future, 5)
	var deleted []*Object
	s.DecayTick511750(func(obj *Object) {
		deleted = append(deleted, obj)
		// The scan must continue through the next pointer cached before this
		// callback, not through a callback mutation of the sidecar.
		s.decay.next[obj] = held
	})

	if len(deleted) != 1 || deleted[0] != due {
		t.Fatalf("deleted = %p, want [%p]", deleted, due)
	}
	if uint32(held.ObjFlags)&decayListedFlag511660 != 0 || uint32(due.ObjFlags)&decayListedFlag511660 != 0 {
		t.Fatalf("removed flags = %08x/%08x", uint32(held.ObjFlags), uint32(due.ObjFlags))
	}
	if due.Field5 != 0x12340081 {
		t.Fatalf("due delete flags = %08x", due.Field5)
	}
	if s.decay.head != future || uint32(future.ObjFlags)&decayListedFlag511660 == 0 {
		t.Fatalf("future state = head %p flags %08x", s.decay.head, uint32(future.ObjFlags))
	}
	if held.Field117 != 0x11111111 || due.Field117 != 0x22222222 || future.Field117 != 0x33333333 {
		t.Fatalf("ABI32 shared fields changed: %08x/%08x/%08x", held.Field117, due.Field117, future.Field117)
	}
}

func TestDecayNativeDestroy5117B0ClearsListAndListedBits(t *testing.T) {
	s := &Server{}
	first := &Object{Field117: 0x11111111}
	second := &Object{Field117: 0x22222222}
	s.DecaySetTime511660(first, 1)
	s.DecaySetTime511660(second, 2)

	s.DecayDestroy5117B0()
	if s.decay.head != nil || uint32(first.ObjFlags)&decayListedFlag511660 != 0 || uint32(second.ObjFlags)&decayListedFlag511660 != 0 {
		t.Fatalf("destroy state = head %p flags %08x/%08x", s.decay.head, uint32(first.ObjFlags), uint32(second.ObjFlags))
	}
	if first.Field117 != 0x11111111 || second.Field117 != 0x22222222 {
		t.Fatalf("ABI32 shared fields changed: %08x/%08x", first.Field117, second.Field117)
	}
}
