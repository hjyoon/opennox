package legacy

import (
	"testing"

	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
)

func TestFriendListNative495980(t *testing.T) {
	handles.Init()
	defer handles.Release()

	if got := Nox_xxx_allocClassListFriends_495980(); got == 0 {
		t.Fatal("friend-list allocation failed")
	}
	defer Sub_4959D0()

	if head := friendListHead495980(); head != nil {
		t.Fatalf("initial head = %p, want nil", head)
	}
	if got := Sub_495A80(0x1234); got != 0 {
		t.Fatalf("empty lookup = %d, want 0", got)
	}

	first := Nox_xxx_cliAddObjFriend_4959F0(0x1234)
	if first == nil {
		t.Fatal("first friend allocation failed")
	}
	if head := friendListHead495980(); head != first {
		t.Fatalf("first head = %p, want full native pointer %p", head, first)
	}
	if got := Sub_495A80(0x1234); got != 1 {
		t.Fatalf("first lookup = %d, want 1", got)
	}

	second := Nox_xxx_cliAddObjFriend_4959F0(0x5678)
	if second == nil {
		t.Fatal("second friend allocation failed")
	}
	if head := friendListHead495980(); head != second {
		t.Fatalf("second head = %p, want full native pointer %p", head, second)
	}
	if got := Sub_495A80(0x1234); got != 1 {
		t.Fatalf("tail lookup = %d, want 1", got)
	}

	Sub_495A20(0xffff)
	if head := friendListHead495980(); head != second {
		t.Fatalf("missing removal changed head to %p, want %p", head, second)
	}
	Sub_495A20(0x5678)
	if head := friendListHead495980(); head != first {
		t.Fatalf("head removal produced %p, want %p", head, first)
	}
	if got := Sub_495A80(0x5678); got != 0 {
		t.Fatalf("removed lookup = %d, want 0", got)
	}

	duplicate := Nox_xxx_cliAddObjFriend_4959F0(0x1234)
	if duplicate == nil {
		t.Fatal("duplicate friend allocation failed")
	}
	Sub_495A20(0x1234)
	if got := Sub_495A80(0x1234); got != 1 {
		t.Fatalf("one duplicate removal lookup = %d, want 1", got)
	}
	Sub_495A20(0x1234)
	if got := Sub_495A80(0x1234); got != 0 {
		t.Fatalf("two duplicate removals lookup = %d, want 0", got)
	}

	if node := Nox_xxx_cliAddObjFriend_4959F0(0x7ffe); node == nil {
		t.Fatal("pre-reset friend allocation failed")
	}
	Sub_4959B0()
	if head := friendListHead495980(); head != nil {
		t.Fatalf("reset head = %p, want nil", head)
	}
	if got := Sub_495A80(0x7ffe); got != 0 {
		t.Fatalf("reset lookup = %d, want 0", got)
	}
}
