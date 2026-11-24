package legacy

import "testing"

func TestLegacyQuestStateWordsRetainSignedValuesAndPreviousStage(t *testing.T) {
	for _, offset := range []uintptr{1556104, 1556108, 1556116, 1556124, 1556132, 1556160, 1556164, 1556300} {
		cell := legacyQuestStateWord(offset)
		previous := *cell
		t.Cleanup(func() { *cell = previous })
	}
	for _, result := range []struct {
		name     string
		set, get int32
		want     int32
	}{
		{"quest mode", int32(nox_xxx_setQuest_4D6F60(-17)), int32(nox_xxx_isQuest_4D6F50()), -17},
		{"quest selection", int32(sub_4D6F80(0x71345678)), int32(sub_4D6F70()), 0x71345678},
		{"retry setting", int32(sub_4D7440(-201)), int32(sub_4D7430()), -201},
		{"warp setting", int32(sub_4D76E0(-1073741824)), int32(sub_4D76F0()), -1073741824},
	} {
		if result.set != result.want || result.get != result.want {
			t.Errorf("%s setter=%#x getter=%#x, want %#x", result.name, result.set, result.get, result.want)
		}
	}
	*legacyQuestStateWord(1556104) = 0xc3470001
	if got := uint32(sub_4D6FA0()); got != 0xc3470001 {
		t.Errorf("quest startup word %#x, want %#x", got, uint32(0xc3470001))
	}
	if got := sub_4D75F0(-2); got != -2 || *legacyQuestStateWord(1556108) != 0xfffffffe {
		t.Errorf("quest timer return %d, word %#x", got, *legacyQuestStateWord(1556108))
	}

	initialStage := sub_4D72D0(-52763)
	t.Cleanup(func() { sub_4D72D0(initialStage) })
	if got := sub_4D7300(); got != initialStage {
		t.Fatalf("previous-stage word %d, want %d", got, initialStage)
	}
	if got := sub_4D72D0(0x7c987654); got != -52763 || sub_4D7300() != -52763 {
		t.Fatalf("stage rotation return %d, previous word %d", got, sub_4D7300())
	}

	players := legacyQuestStateWord(1556300)
	*players = ^uint32(0)
	if mask := uint32(sub_4D79A0(-1)); mask != 0x7fffffff || *players != 0x7fffffff {
		t.Fatalf("client 31 mask %#x, stored %#x", mask, *players)
	}
	if mask := uint32(sub_4D79A0(33)); mask != 0xfffffffd || *players != 0x7ffffffd {
		t.Fatalf("wrapped client 1 mask %#x, stored %#x", mask, *players)
	}
}

func TestQuestStageAndWarpPacketsRetainTheirQueueModes(t *testing.T) {
	objectStatusTestServer(t, 2)
	setImportantRecipientMaskC(1 << 6)
	if result := sub_4D7280(6, -5); result != 1 {
		t.Fatalf("warp status result %d", result)
	}
	if result := sub_4D7450(6, -260); result != 1 {
		t.Fatalf("quest stage result %d", result)
	}
	first, _ := importantNativeListC()
	checkObjectStatusPacket(t, first, 6, []byte{0xf0, 29, 0xfc, 0xfe}, 0, 1)
	checkObjectStatusPacket(t, importantPacketNextC(first), 6, []byte{0xf0, 24, 0xfb}, 1, 1)
}
