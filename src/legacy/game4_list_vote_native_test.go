package legacy

import (
	"bytes"
	"fmt"
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/legacy/internal/cfixture"
	"github.com/opennox/opennox/v1/server"
)

func game4SupportRequireHigh(t *testing.T, addresses []uintptr) {
	t.Helper()
	if unsafe.Sizeof(uintptr(0)) != 8 {
		return
	}
	for _, p := range addresses {
		if p <= math.MaxUint32 {
			t.Fatalf("native allocation address %#x is not above 4 GiB", p)
		}
	}
}

func TestGame4TempListsNativeCReferenceFreeOrder(t *testing.T) {
	for _, payloads := range []bool{false, true} {
		t.Run(fmt.Sprint(payloads), func(t *testing.T) {
			want, err := game4SupportFixture.Walls(true, payloads)
			if err != nil {
				t.Fatal(err)
			}
			got, err := game4SupportFixture.Walls(false, payloads)
			if err != nil {
				t.Fatal(err)
			}
			game4SupportRequireHigh(t, got.NativeAddresses)
			if !reflect.DeepEqual(got.Links, want.Links) || !reflect.DeepEqual(got.Freed, want.Freed) {
				t.Fatalf("native links/free order %v/%v, C %v/%v", got.Links, got.Freed, want.Links, want.Freed)
			}
			expected := []int{21, 11, 20, 10}
			if payloads {
				expected = []int{31, 21, 11, 30, 20, 10}
			}
			if !reflect.DeepEqual(got.Freed, expected) || !reflect.DeepEqual(got.Links, []int{11, 10, 11, 10, 0}) {
				t.Fatalf("unexpected bounded constructor/lookup/free result: %+v", got)
			}
		})
	}
}

func TestGame4MapObjectsNativeCReferenceAndPinnedOwners(t *testing.T) {
	want, err := game4SupportFixture.MapObjects(true)
	if err != nil {
		t.Fatal(err)
	}
	got, err := game4SupportFixture.MapObjects(false)
	if err != nil {
		t.Fatal(err)
	}
	game4SupportRequireHigh(t, got.NativeAddresses)
	if !reflect.DeepEqual(got.Links, want.Links) || !reflect.DeepEqual(got.Freed, want.Freed) {
		t.Fatalf("native object links/free order %v/%v, C %v/%v", got.Links, got.Freed, want.Links, want.Freed)
	}
	if !reflect.DeepEqual(got.Freed, []int{13, 12, 11}) {
		t.Fatalf("node free order %v", got.Freed)
	}
	// The C nodes retain real Go owners only while every owner is explicitly
	// pinned. Unlink/free all C nodes before unpinning either Go allocation.
	first, second := new(server.Object), new(server.Object)
	var pin runtime.Pinner
	pin.Pin(first)
	pin.Pin(second)
	defer pin.Unpin()
	ok, cleanup := game4SupportFixture.PinnedMapObjects(unsafe.Pointer(first), unsafe.Pointer(second))
	defer cleanup()
	if !ok || second.ObjNext != first || first.ObjPrev != second {
		t.Fatalf("pinned owner identity/link mismatch")
	}
	runtime.GC()
	runtime.KeepAlive(first)
	runtime.KeepAlive(second)
}

func TestGame4VoteNativeExactCScalarAndUnlink(t *testing.T) {
	for _, in := range []struct {
		voters uint32
		count  uint8
		mask   uint32
	}{{0, 0, 1}, {1, 1, 1}, {0x100, 3, 0x100}, {0, 255, 0x80000000}, {0x80000000, 1, 0x80000000}, {0xffffffff, 255, 0x20}} {
		t.Run(fmt.Sprintf("%08x-%d-%08x", in.voters, in.count, in.mask), func(t *testing.T) {
			ret, body, guard := game4SupportFixture.VoteAdd(false, in.voters, in.count, in.mask)
			want, wbody, wguard := game4SupportFixture.VoteAdd(true, in.voters, in.count, in.mask)
			if ret != want || !bytes.Equal(body, wbody) || !guard || !wguard {
				t.Fatalf("add-voter mismatch or canary damage: return %d/%d", ret, want)
			}
		})
	}
	for i := 0; i < 3; i++ {
		t.Run(fmt.Sprintf("unlink-%d", i), func(t *testing.T) {
			got, err := game4SupportFixture.VoteUnlink(false, i)
			if err != nil {
				t.Fatal(err)
			}
			want, err := game4SupportFixture.VoteUnlink(true, i)
			if err != nil {
				t.Fatal(err)
			}
			game4SupportRequireHigh(t, got.NativeAddresses)
			if !reflect.DeepEqual(got.Links, want.Links) {
				t.Fatalf("unlink result/links %v, C %v", got.Links, want.Links)
			}
		})
	}
	for _, in := range []struct{ a, b []uint16 }{{[]uint16{0}, []uint16{0}}, {[]uint16{'A', 0}, []uint16{'a', 0}}, {[]uint16{0xd83d, 0xde03, 0}, []uint16{0xd83d, 0xde03, 0}}, {[]uint16{1, 0xffff, 0}, []uint16{1, 0xfffe, 0}}} {
		got, want := game4SupportNameEquality(in.a, in.b)
		if got != want {
			t.Fatalf("UTF16 code-unit equality %v/%v", got, want)
		}
	}
}

func TestGame4VoteNativeCReferencePacketsAndLifetime(t *testing.T) {
	handles.InitHigh()
	defer handles.Release()
	oldFrame := gameFrameHook
	gameFrameHook = func() uint32 { return 0x12345678 }
	defer func() { gameFrameHook = oldFrame }()
	want, err := game4SupportFixture.VoteLifecycle(true)
	if err != nil {
		t.Fatal(err)
	}
	got, err := game4SupportFixture.VoteLifecycle(false)
	if err != nil {
		t.Fatal(err)
	}
	game4SupportRequireHigh(t, got.NativeAddresses)
	if !reflect.DeepEqual(got.Packets, want.Packets) || !reflect.DeepEqual(got.Links, want.Links) {
		t.Fatalf("lifecycle packets/fields %v/%v, C %v/%v", got.Packets, got.Links, want.Packets, want.Links)
	}
	expected := []cfixture.Game4SupportPacket{{255, []byte{0xee, 6, 1}}, {1, []byte{0xee, 7}}, {31, []byte{0xee, 7}}, {255, []byte{0xee, 6, 0}}}
	if !reflect.DeepEqual(got.Packets, expected) || !reflect.DeepEqual(got.Links, []int{6, 6, 0x12345678, 0x12345678, 0}) {
		t.Fatalf("bounded vote lifetime sequence %+v", got)
	}
	t.Run("initial-status-callback-retains-target", game4SupportTargetCacheControl)
}

type game4SupportLegacyServer struct {
	Server
	srv *server.Server
}

func (s *game4SupportLegacyServer) S() *server.Server { return s.srv }

func game4SupportTargetCacheControl(t *testing.T) {
	t.Helper()
	srv := server.NewHigh(nil, nil, strman.New())
	defer srv.Close()
	oldGet := GetServer
	GetServer = func() Server { return &game4SupportLegacyServer{srv: srv} }
	defer func() { GetServer = oldGet }()
	voter, targetPlayer := srv.Players.NewRaw(71001), srv.Players.NewRaw(71002)
	if voter == nil || targetPlayer == nil {
		t.Fatal("private player slots unavailable")
	}
	unit, freeUnit := alloc.NewHigh(server.Object{})
	defer freeUnit()
	update, freeUpdate := alloc.NewHigh(server.PlayerUpdateData{})
	defer freeUpdate()
	target, freeTarget := alloc.NewHigh(server.Object{})
	defer freeTarget()
	replacement, freeReplacement := alloc.NewHigh(server.Object{})
	defer freeReplacement()
	*unit = server.Object{}
	*update = server.PlayerUpdateData{}
	*target = server.Object{}
	*replacement = server.Object{}
	unit.ObjClass = object.ClassPlayer
	unit.UpdateData = unsafe.Pointer(update)
	update.Player = voter
	voter.Field4792 = 1
	targetPlayer.Field4792 = 1
	targetPlayer.SetName("cache")
	oldObserver := game4VotePacketObserver
	defer func() { game4VotePacketObserver = oldObserver }()
	for _, reference := range []bool{true, false} {
		targetPlayer.PlayerUnit = target
		calls := 0
		game4VotePacketObserver = func(recipient int32, b []byte) int32 {
			if recipient == 255 && bytes.Equal(b, []byte{0xee, 6, 1}) {
				calls++
				targetPlayer.PlayerUnit = replacement
			}
			return 73
		}
		got, err := game4SupportFixture.VoteTargetCache(reference, unsafe.Pointer(unit), []uint16{'c', 'a', 'c', 'h', 'e', 0})
		if err != nil {
			t.Fatal(err)
		}
		if calls != 1 || got != unsafe.Pointer(target) || targetPlayer.PlayerUnit != replacement {
			t.Fatalf("backend reference=%v cached=%p want=%p mutation=%p calls=%d", reference, got, target, targetPlayer.PlayerUnit, calls)
		}
	}
	targetPlayer.PlayerUnit = nil
}

var game4SupportFixture = cfixture.Game4SupportServices{
	HighMemory:       game4SupportHighFixtureMemory,
	Allocate:         game4SupportAllocate,
	Release:          game4SupportReleaseMemory,
	WallHead:         func() unsafe.Pointer { return unsafe.Pointer(game4WallHead()) },
	SetWallHead:      func(p unsafe.Pointer) { game4SetWallHead(game4FixturePointer(p, game4WallHead())) },
	MapObjectHead:    func() unsafe.Pointer { return unsafe.Pointer(game4MapObjectHead()) },
	SetMapObjectHead: func(p unsafe.Pointer) { game4SetMapObjectHead(game4FixturePointer(p, game4MapObjectHead())) },
	VoteHead:         func() unsafe.Pointer { return unsafe.Pointer(game4VoteHead) },
	SetVoteHead:      func(p unsafe.Pointer) { game4VoteHead = game4FixturePointer(p, game4VoteHead) },
	ObserveFree: func(next func(unsafe.Pointer)) func() {
		previous := game4SupportFreeObserver
		game4SupportFreeObserver = next
		return func() { game4SupportFreeObserver = previous }
	},
	ObservePacket: func(next func(int32, []byte) int32) func() {
		previous := game4VotePacketObserver
		game4VotePacketObserver = next
		return func() { game4VotePacketObserver = previous }
	},
	AddVoter: func(p unsafe.Pointer, mask uint32) int8 {
		return int8(game4VoteAddVoter(game4FixturePointer(p, game4VoteHead), mask))
	},
}

func game4FixturePointer[T any](p unsafe.Pointer, _ *T) *T {
	return (*T)(p)
}

func game4SupportNameEquality(a, b []uint16) (bool, bool) {
	first := (*wchar2_t)(unsafe.Pointer(&a[0]))
	second := (*wchar2_t)(unsafe.Pointer(&b[0]))
	got := game4VoteNamesEqual(first, second)
	want := cfixture.Game4VoteNameEqual(a, b)
	runtime.KeepAlive(a)
	runtime.KeepAlive(b)
	return got, want
}
