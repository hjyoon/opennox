package legacy

import (
	"bytes"
	"encoding/binary"
	"os"
	"os/exec"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/server"
)

// Each child invokes the actual public C entry with its default C iterator,
// packet sender and allocator. No runtime override supplies a target result.
func TestWallDestroyedActualCProduction4DF0A0(t *testing.T) {
	modes := []string{"broadcast", "direct0", "direct31", "direct255", "empty-broadcast", "empty-direct", "inactive-first", "last-slot", "unsigned255", "cached-wall", "live-disconnect", "direct-poison", "direct-excluded"}
	if mode := os.Getenv("OPENNOX_WALL_DESTROYED_PRODUCTION_CHILD"); mode != "" {
		if !slices.Contains(modes, mode) || secretWallNativeRuntimeOverride4DF120 != nil || !wallDestroyedOriginalTyped4DF0A0() {
			t.Fatal("known mode, native public C entry and default runtime required", mode)
		}
		TestSecretWallStateNativeLayout4DF120(t)
		handles.InitHigh()
		t.Cleanup(handles.Release)
		preserveImportantPacketState(t)
		setImportantNativeListC(nil, nil)
		setImportantListHeadsC(0, 0)
		pool := alloc.NewClassHigh("wall-destroyed-real-C-production", importantPacketSizeC(), 16)
		setImportantAllocClassC(pool.UPtr())
		mask := uint32(0x80000005)
		s, wall, units := secretWallLivePlayersFixture4DF120(t)
		s.SetFrame(0x12345678)
		firstPlayer := (*server.PlayerUpdateData)(units[0].UpdateData).Player
		lastPlayer := (*server.PlayerUpdateData)(units[1].UpdateData).Player
		recipient, wantReturn := int32(32), int32(0)
		wantRecipients := []byte{0, 2}
		switch mode {
		case "direct0", "direct-poison":
			recipient, wantReturn, wantRecipients = 0, 1, []byte{0}
		case "direct31", "empty-direct":
			recipient, wantReturn, wantRecipients = 31, 1, []byte{31}
		case "direct255":
			recipient, wantReturn, wantRecipients = 255, 1, []byte{255}
		case "direct-excluded":
			// Real C Send accepts the all-except-player-0 mode without allocating
			// when its recipient mask contains only player 0. Its result is still 1.
			recipient, wantReturn, wantRecipients, mask = 128, 1, nil, 1
		case "empty-broadcast":
			wantRecipients = nil
		case "inactive-first":
			firstPlayer.Active = 0
			wantRecipients = []byte{2}
		case "last-slot":
			lastPlayer.Active = 0
			lastPlayer = s.Players.ResetInd(31)
			lastPlayer.PlayerUnit = units[1]
			(*server.PlayerUpdateData)(units[1].UpdateData).Player = lastPlayer
			wantRecipients = []byte{0, 31}
		case "unsigned255":
			firstPlayer.PlayerInd = 255
			wantRecipients = []byte{255}
		case "live-disconnect":
			wantRecipients = []byte{0}
		}
		if mode == "empty-broadcast" || mode == "empty-direct" {
			firstPlayer.PlayerUnit, lastPlayer.PlayerUnit = nil, nil
		}
		setImportantRecipientMaskC(mask)
		type record struct {
			ptr  unsafe.Pointer
			size uintptr
		}
		records := []record{{unsafe.Pointer(wall), unsafe.Sizeof(*wall)}}
		for _, unit := range units {
			update := (*server.PlayerUpdateData)(unit.UpdateData)
			records = append(records, record{unsafe.Pointer(unit), unsafe.Sizeof(*unit)}, record{unit.UpdateData, unsafe.Sizeof(*update)})
		}
		playerRecords := make(map[*server.Player]int)
		for _, player := range s.Players.ListSlots() {
			playerRecords[player] = len(records)
			records = append(records, record{unsafe.Pointer(player), unsafe.Sizeof(*player)})
		}
		if len(playerRecords) != 32 {
			t.Fatal("all native player slots must be guarded")
		}
		if mode == "direct-poison" {
			// Prepare after capturing the valid update addresses. The direct path
			// must send without dereferencing these deliberately absent records.
			for _, unit := range units {
				unit.ObjClass, unit.UpdateData = 0, nil
			}
		}
		want := make([][]byte, len(records))
		for i, r := range records {
			if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(r.ptr) <= 0xffffffff {
				t.Fatal("native production records must exceed4GiB")
			}
			want[i] = bytes.Clone(unsafe.Slice((*byte)(r.ptr), int(r.size)))
		}
		frames := 0
		gameFrameHook = func() uint32 {
			frames++
			if frames == 1 {
				switch mode {
				case "cached-wall":
					wall.Field10 = 0x55aa
					binary.LittleEndian.PutUint16(want[0][unsafe.Offsetof(wall.Field10):], 0x55aa)
				case "live-disconnect":
					lastPlayer.Active = 0
					want[playerRecords[lastPlayer]][unsafe.Offsetof(lastPlayer.Active)] = 0
				}
			}
			return s.Frame()
		}
		counters := *memmap.PtrT[[32]uint16](0x5D4594, 1565524)
		if got := wallDestroyedOriginalC4DF0A0(wall, recipient); got != wantReturn {
			t.Fatal("actual direct-Send/broadcast return", got, wantReturn)
		}
		var actualRecipients []byte
		first, last := importantNativeListC()
		previous := unsafe.Pointer(nil)
		for p := first; p != nil; p = importantPacketNextC(p) {
			if len(actualRecipients) >= 2 || importantPacketPrevC(p) != previous {
				t.Fatal("actual C important queue links/cycle")
			}
			packet := (*importantPacketLegacy)(p)
			if packet.PayloadSize != 3 || !bytes.Equal(packet.Payload[:3], []byte{58, 0x34, 0x92}) || packet.RemoveIfDisconnected != 1 || packet.SequenceEnabled != 0 || packet.Sequence != [32]uint16{} || packet.CreatedFrame != s.Frame() || importantPacketRelatedObjectC(p) != nil || packet.RecipientMask != mask {
				t.Fatalf("actual C wall-destroyed packet bytes/routing/frame/sequence: %+v", packet)
			}
			actualRecipients = append(actualRecipients, byte(packet.Recipient))
			previous = p
		}
		slices.Reverse(actualRecipients)
		if !bytes.Equal(actualRecipients, wantRecipients) || previous != last || frames != len(wantRecipients) {
			t.Fatal("actual C queue/traversal", actualRecipients, wantRecipients, previous, last, frames)
		}
		if *memmap.PtrT[[32]uint16](0x5D4594, 1565524) != counters || s.Frame() != 0x12345678 {
			t.Fatal("unsequenced send modified counters or the real frame")
		}
		for i, r := range records {
			if !bytes.Equal(want[i], unsafe.Slice((*byte)(r.ptr), int(r.size))) {
				t.Fatal("whole native wall/unit/update/all32player record changed outside declared mutation", i)
			}
		}
		t.Logf("WALL_DESTROYED_ACTUAL_C_PRODUCTION mode=%s recipient=%d return=%d queued=%v native-wall=%p actual-C-first-next-send-allocator=true whole-records=true", mode, recipient, wantReturn, actualRecipients, wall)
		return
	}
	for _, mode := range modes {
		t.Run(mode, func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.v", "-test.run=^TestWallDestroyedActualCProduction4DF0A0$", "-test.timeout=60s")
			cmd.Env = append(os.Environ(), "OPENNOX_WALL_DESTROYED_PRODUCTION_CHILD="+mode)
			out, err := cmd.CombinedOutput()
			t.Logf("%s", out)
			if err != nil || !bytes.Contains(out, []byte("WALL_DESTROYED_ACTUAL_C_PRODUCTION")) {
				t.Fatal("actual default C pipeline child failed", err)
			}
		})
	}
}
