package legacy

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/server"
)

func TestVote5066D0PreservesNativeAllocatorRecordsLinksAndPlayerPointers(t *testing.T) {
	handles.Init()
	// The production shutdown path is intentionally idempotent so this test is
	// independent of whether a prior test initialized the global vote class.
	Sub_506720()
	if got := Nox_xxx_allocVoteArray_5066D0(); got != 1 {
		t.Fatalf("alloc vote array = %d, want 1", got)
	}

	unit, freeUnit := alloc.New(server.Object{})
	update, freeUpdate := alloc.New(server.PlayerUpdateData{})
	player, freePlayer := alloc.New(server.Player{})
	unit2, freeUnit2 := alloc.New(server.Object{})
	update2, freeUpdate2 := alloc.New(server.PlayerUpdateData{})
	player2, freePlayer2 := alloc.New(server.Player{})
	oldFrameHook := gameFrameHook
	minimumVotes := memmap.PtrUint32(0x587000, 229980)
	oldMinimumVotes := *minimumVotes
	t.Cleanup(func() {
		Sub_506720()
		gameFrameHook = oldFrameHook
		*minimumVotes = oldMinimumVotes
		freePlayer()
		freeUpdate()
		freeUnit()
		freePlayer2()
		freeUpdate2()
		freeUnit2()
		handles.Release()
	})

	const frame = uint32(0x89ABCDEF)
	gameFrameHook = func() uint32 { return frame }
	*minimumVotes = 7
	unit.ObjClass = object.ClassPlayer
	unit.UpdateData = unsafe.Pointer(update)
	update.Player = player
	player.PlayerInd = 5
	unit2.ObjClass = object.ClassPlayer
	unit2.UpdateData = unsafe.Pointer(update2)
	update2.Player = player2
	player2.PlayerInd = 7

	wantSize := uintptr(52)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantSize = 72
	}
	if got := voteRecordSize5066D0(); got != wantSize {
		t.Fatalf("native Vote size = %d, want %d", got, wantSize)
	}

	sentinel := voteRecordAlloc5066D0()
	if sentinel == nil || voteRecordAdd506AD0(sentinel) != sentinel {
		t.Fatalf("sentinel allocation/add = %p/%p", sentinel, voteHead5066D0())
	}
	created := voteRecordCreate506A20(0, unit)
	if created == nil {
		t.Fatal("sub_506A20 returned nil")
	}
	if got := voteHead5066D0(); got != created {
		t.Fatalf("head = %p, want created %p", got, created)
	}
	if got := voteNext5066D0(created); got != sentinel {
		t.Fatalf("created.next = %p, want sentinel %p", got, sentinel)
	}
	if got := votePrevious5066D0(sentinel); got != created {
		t.Fatalf("sentinel.previous = %p, want created %p", got, created)
	}
	if got := votePrevious5066D0(created); got != nil {
		t.Fatalf("created.previous = %p, want nil", got)
	}
	wantTeam := unsafe.Add(unsafe.Pointer(unit), unsafe.Offsetof(server.Object{}.TeamVal))
	if got := voteTeam5066D0(created); got != wantTeam {
		t.Fatalf("created.team = %p, want embedded team %p", got, wantTeam)
	}
	if got := voteFrame5066D0(created); got != frame {
		t.Fatalf("created.frame = %#x, want %#x", got, frame)
	}
	if got := voteThreshold5066D0(created); got != 7 {
		t.Fatalf("created.threshold = %d, want 7", got)
	}

	const initialMask = uint32(1<<5 | 1<<7)
	const remainingMask = uint32(1 << 7)
	voteSetVoters5066D0(created, initialMask, 2)
	voteSetVoters5066D0(sentinel, initialMask, 2)
	Sub_506740(unit)
	for name, vote := range map[string]unsafe.Pointer{"created": created, "sentinel": sentinel} {
		if got := voteVoters5066D0(vote); got != remainingMask {
			t.Fatalf("%s voters = %#x, want %#x", name, got, remainingMask)
		}
		if got := voteCount5066D0(vote); got != 1 {
			t.Fatalf("%s count = %d, want 1", name, got)
		}
	}

	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, pointer := range map[string]uintptr{
			"allocator": uintptr(voteAllocator5066D0()),
			"sentinel":  uintptr(sentinel),
			"created":   uintptr(created),
			"unit":      uintptr(unsafe.Pointer(unit)),
			"update":    uintptr(unsafe.Pointer(update)),
			"player":    uintptr(unsafe.Pointer(player)),
			"unit2":     uintptr(unsafe.Pointer(unit2)),
			"update2":   uintptr(unsafe.Pointer(update2)),
			"player2":   uintptr(unsafe.Pointer(player2)),
		} {
			if pointer <= math.MaxUint32 {
				t.Fatalf("%s pointer = %#x, want native address above 4 GiB", name, pointer)
			}
		}
	}

	Sub_506740(unit2)
	if voteHead5066D0() != nil || Sub_5071C0() {
		t.Fatalf("removing the final voter left vote state head=%p active=%v",
			voteHead5066D0(), Sub_5071C0())
	}
	emptyVote := voteRecordAlloc5066D0()
	if emptyVote == nil || voteRecordAdd506AD0(emptyVote) != emptyVote {
		t.Fatalf("empty fixture allocation/add = %p/%p", emptyVote, voteHead5066D0())
	}
	Sub_506740(unit)
	if voteHead5066D0() != nil || Sub_5071C0() {
		t.Fatalf("disconnect cleanup retained an already-empty vote: head=%p active=%v",
			voteHead5066D0(), Sub_5071C0())
	}
	resetVote := voteRecordAlloc5066D0()
	if resetVote == nil || voteRecordAdd506AD0(resetVote) != resetVote {
		t.Fatalf("reset fixture allocation/add = %p/%p", resetVote, voteHead5066D0())
	}
	allocator := voteAllocator5066D0()
	Sub_506700()
	if voteHead5066D0() != nil || voteAllocator5066D0() != allocator || Sub_5071C0() {
		t.Fatalf("reset left vote state head=%p allocator=%p active=%v, want nil/%p/false",
			voteHead5066D0(), voteAllocator5066D0(), Sub_5071C0(), allocator)
	}
	reused := voteRecordAlloc5066D0()
	if reused == nil || voteRecordAdd506AD0(reused) != reused || !Sub_5071C0() {
		t.Fatalf("allocator was not reusable after reset: vote=%p head=%p active=%v",
			reused, voteHead5066D0(), Sub_5071C0())
	}

	Sub_506720()
	if voteHead5066D0() != nil || voteAllocator5066D0() != nil || Sub_5071C0() {
		t.Fatalf("shutdown left vote state head=%p allocator=%p active=%v",
			voteHead5066D0(), voteAllocator5066D0(), Sub_5071C0())
	}
}
