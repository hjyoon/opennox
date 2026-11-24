package legacy

import (
	"bytes"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

func TestQuestStagePacketsRetainNamesFlagsAndStorage(t *testing.T) {
	objectStatusTestServer(t, 2)
	setImportantRecipientMaskC(1 << 6)
	firstName := memmap.Slice(0x973F18, questStageFirstNameOffset)[:32]
	secondName := memmap.Slice(0x973F18, questStageSecondNameOffset)[:32]
	firstSaved, secondSaved := bytes.Clone(firstName), bytes.Clone(secondName)
	oldStage, oldComplete := nox_game_getQuestStage_4E3CC0(), sub_51A950()
	t.Cleanup(func() {
		copy(firstName, firstSaved)
		copy(secondName, secondSaved)
		nox_game_setQuestStage_4E3CD0(oldStage)
		sub_51A940(oldComplete)
	})
	copy(firstName, "0123456789ABCDEFGHIJKLMNOPQRSTUV")
	clear(secondName)
	copy(secondName, "QuestCave")
	nox_game_setQuestStage_4E3CD0(0x7123a456)
	sub_51A940(-1)
	if unsafe.Pointer(sub_4D6940()) != unsafe.Pointer(&firstName[0]) ||
		unsafe.Pointer(sub_4D6950()) != unsafe.Pointer(&secondName[0]) {
		t.Fatal("quest name exports lost their native storage addresses")
	}
	if got := sub_4D6880(6, -1); got != 1 {
		t.Fatalf("quest state report returned %d", got)
	}
	if got := nox_game_sendQuestStage_4D6960(6); got != 1 {
		t.Fatalf("quest stage report returned %d", got)
	}
	want := [69]byte{0xf0, 14, 0x56, 0xa4, 2}
	copy(want[5:36], firstName[:31])
	copy(want[37:69], "QuestCave")
	first, _ := importantNativeListC()
	checkObjectStatusPacket(t, first, 6, want[:], 1, 1)
	want[1], want[4] = 13, 3
	checkObjectStatusPacket(t, importantPacketNextC(first), 6, want[:], 1, 1)
}
