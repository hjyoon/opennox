package legacy

import (
	"bytes"
	"encoding/binary"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestPlayerQuestKeysClientReadNativeBytesAndDoesNotWrite(t *testing.T) {
	if got := Nox_client_playerQuestKeysReceived(nil); got != [2]byte{} {
		t.Fatalf("nil client player=%v", got)
	}
	player, free := alloc.New(server.Player{})
	t.Cleanup(free)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(player)) <= math.MaxUint32 {
		t.Fatal("client player fixture must exceed 4 GiB")
	}
	// Independently locate the remaining Go tail after the five native Ankh
	// pointers. The reader uses the C field; their layout must agree.
	tailOffset := unsafe.Offsetof(server.Player{}.QuestAnkhs) + unsafe.Sizeof(player.QuestAnkhs)
	if tailOffset != unsafe.Offsetof(nox_playerInfo{}.tail_padding) || unsafe.Sizeof(*player) != unsafe.Sizeof(nox_playerInfo{}) {
		t.Fatal("native Go/C player tail layouts disagree")
	}
	raw := unsafe.Slice((*byte)(unsafe.Pointer(player)), int(unsafe.Sizeof(*player)))
	tail := raw[int(tailOffset) : int(tailOffset)+12]
	binary.LittleEndian.PutUint32(tail, 0x11223344)
	binary.LittleEndian.PutUint32(tail[4:], 0x55667788)
	before := make([]byte, len(raw))
	for silver := 0; silver < 256; silver++ {
		for gold := 0; gold < 256; gold++ {
			binary.LittleEndian.PutUint32(tail[8:], 0xaabb0000|uint32(gold)<<8|uint32(silver))
			copy(before, raw)
			if got := Nox_client_playerQuestKeysReceived(player); got != [2]byte{byte(silver), byte(gold)} {
				t.Fatalf("client tail=%x received=%v want=%d/%d", tail, got, silver, gold)
			}
			if !bytes.Equal(raw, before) {
				t.Fatal("client key reader changed the player record")
			}
		}
	}
}
