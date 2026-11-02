package legacy

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/server"
)

// Each child uses the default public-entry bindings: real typed C first/next,
// the actual C packet wrapper and allocator, and independently inspected queues.
func TestSecretWallStateActualCProduction4DF120(t *testing.T) {
	modes := []string{"ordinary", "empty", "inactive-first", "last-slot", "unsigned255", "cached-wall", "live-disconnect"}
	if child := os.Getenv("OPENNOX_SECRET_WALL_PRODUCTION_CHILD"); child != "" {
		parts := strings.Split(child, "/")
		if len(parts) != 2 || !slices.Contains(modes, parts[1]) {
			t.Fatal("invalid production child", child)
		}
		entry, err := strconv.Atoi(parts[0])
		if err != nil || entry < 0 || entry > 1 {
			t.Fatal("invalid public entry", child)
		}
		TestSecretWallStateNativeLayout4DF120(t)
		if secretWallNativeRuntimeOverride4DF120 != nil || !secretWallOriginalTyped4DF120(entry) {
			t.Fatal("actual native entry and default runtime are required")
		}
		handles.Init()
		t.Cleanup(handles.Release)
		preserveImportantPacketState(t)
		setImportantNativeListC(nil, nil)
		setImportantListHeadsC(0, 0)
		pool := alloc.NewClass("secret-wall-real-C-production", importantPacketSizeC(), 16)
		setImportantAllocClassC(pool.UPtr())
		setImportantRecipientMaskC(0x80000005)
		s, wall, units := secretWallLivePlayersFixture4DF120(t)
		s.SetFrame(0x12345678)
		firstPlayer := (*server.PlayerUpdateData)(units[0].UpdateData).Player
		lastPlayer := (*server.PlayerUpdateData)(units[1].UpdateData).Player
		wantRecipients := []byte{0, 2}
		switch parts[1] {
		case "empty":
			firstPlayer.PlayerUnit, lastPlayer.PlayerUnit = nil, nil
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
		type record struct {
			ptr  unsafe.Pointer
			size uintptr
		}
		records := []record{{unsafe.Pointer(wall), unsafe.Sizeof(*wall)}}
		for _, unit := range units {
			update := (*server.PlayerUpdateData)(unit.UpdateData)
			records = append(records, record{unsafe.Pointer(unit), unsafe.Sizeof(*unit)}, record{unit.UpdateData, unsafe.Sizeof(*update)}, record{unsafe.Pointer(update.Player), unsafe.Sizeof(*update.Player)})
		}
		want := make([][]byte, len(records))
		for i, r := range records {
			if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(r.ptr) <= 0xffffffff {
				t.Fatal("actual production records must retain native high pointers")
			}
			want[i] = bytes.Clone(unsafe.Slice((*byte)(r.ptr), int(r.size)))
		}
		frames := 0
		gameFrameHook = func() uint32 {
			frames++
			if frames == 1 {
				switch parts[1] {
				case "cached-wall":
					wall.Field10 = 0x55aa
					binary.LittleEndian.PutUint16(want[0][unsafe.Offsetof(wall.Field10):], 0x55aa)
				case "live-disconnect":
					lastPlayer.Active = 0
					want[6][unsafe.Offsetof(lastPlayer.Active)] = 0
				}
			}
			return s.Frame() // Observe the real frame getter; do not supply a fake result.
		}
		counters := *memmap.PtrT[[32]uint16](0x5D4594, 1565524)
		if got := secretWallOriginalC4DF120(entry, wall); got != 0 {
			t.Fatal("terminating iterator return changed", got)
		}
		var actualRecipients []byte
		first, last := importantNativeListC()
		previous := unsafe.Pointer(nil)
		opcode := byte(59 + entry)
		for p := first; p != nil; p = importantPacketNextC(p) {
			if len(actualRecipients) >= 2 || importantPacketPrevC(p) != previous {
				t.Fatal("actual C important queue links/cycle")
			}
			packet := (*importantPacketLegacy)(p)
			if packet.PayloadSize != 3 || !bytes.Equal(packet.Payload[:3], []byte{opcode, 0x34, 0x92}) || packet.RemoveIfDisconnected != 1 || packet.SequenceEnabled != 0 || packet.Sequence != [32]uint16{} || packet.CreatedFrame != s.Frame() || importantPacketRelatedObjectC(p) != nil || packet.RecipientMask != 0x80000005 {
				t.Fatalf("actual C packet bytes/routing/frame/sequence changed: %+v", packet)
			}
			actualRecipients = append(actualRecipients, byte(packet.Recipient))
			previous = p
		}
		slices.Reverse(actualRecipients) // The actual allocator prepends new packets.
		if !bytes.Equal(actualRecipients, wantRecipients) || previous != last || frames != len(wantRecipients) {
			t.Fatal("actual live C traversal/queue recipients", actualRecipients, wantRecipients, previous, last, frames)
		}
		if *memmap.PtrT[[32]uint16](0x5D4594, 1565524) != counters || s.Frame() != 0x12345678 {
			t.Fatal("unsequenced sender modified counters or the server frame")
		}
		for i, r := range records {
			if !bytes.Equal(want[i], unsafe.Slice((*byte)(r.ptr), int(r.size))) {
				t.Fatal("whole native production record changed outside declared frame-observer mutation", i)
			}
		}
		t.Logf("SECRET_WALL_ACTUAL_C_PRODUCTION entry=%d mode=%s recipients=%v native-wall=%p actual-C-first-next-send-allocator=true whole-records=true", entry, parts[1], actualRecipients, wall)
		return
	}
	for entry := 0; entry < 2; entry++ {
		for _, mode := range modes {
			child := fmt.Sprintf("%d/%s", entry, mode)
			t.Run(child, func(t *testing.T) {
				cmd := exec.Command(os.Args[0], "-test.v", "-test.run=^TestSecretWallStateActualCProduction4DF120$", "-test.timeout=60s")
				cmd.Env = append(os.Environ(), "OPENNOX_SECRET_WALL_PRODUCTION_CHILD="+child)
				out, err := cmd.CombinedOutput()
				t.Logf("%s", out)
				if err != nil || !bytes.Contains(out, []byte("SECRET_WALL_ACTUAL_C_PRODUCTION")) {
					t.Fatal("actual C/default runtime production child failed", err)
				}
			})
		}
	}
}
