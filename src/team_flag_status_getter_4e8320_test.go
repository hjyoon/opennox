package opennox

import "testing"

func TestTeamFlagStatusGetter4E8320UsesLowByteAndExactStride(t *testing.T) {
	var records [256]teamFlagStatusRecord4E82C0
	tests := []struct {
		teamID uint32
		index  uint8
	}{
		{teamID: 0, index: 0},
		{teamID: 1, index: 1},
		{teamID: 0xff, index: 0xff},
		{teamID: 0x100, index: 0},
		{teamID: 0xa5a50101, index: 1},
		{teamID: 0xffffffff, index: 0xff},
	}
	for _, tc := range tests {
		got := teamFlagStatusGetter4E8320(&records[0], tc.teamID)
		want := &records[tc.index]
		if got != want {
			t.Fatalf("record(%#x) = %p, want index %#x at %p", tc.teamID, got, tc.index, want)
		}
	}
}

func TestTeamFlagStatusGetter4E8320DoesNotMutateSelectedRecord(t *testing.T) {
	var records [256]teamFlagStatusRecord4E82C0
	records[0xfe] = teamFlagStatusRecord4E82C0{
		TeamID:         0xfe,
		FlagIndex:      0xa7,
		Status:         0x81,
		Reserved:       0x7b,
		CarrierNetCode: 0xbcde,
	}
	want := records[0xfe]
	got := teamFlagStatusGetter4E8320(&records[0], 0x123456fe)
	if got != &records[0xfe] {
		t.Fatalf("record pointer = %p, want %p", got, &records[0xfe])
	}
	if records[0xfe] != want {
		t.Fatalf("record = %#v, want unchanged %#v", records[0xfe], want)
	}
}

func TestTeamFlagStatusGetter4E8320AliasesLiveRecord(t *testing.T) {
	var records [256]teamFlagStatusRecord4E82C0
	records[17].Reserved = 0xa5
	got := teamFlagStatusGetter4E8320(&records[0], 17)
	records[17].TeamID = 17
	records[17].FlagIndex = 3
	records[17].Status = 2
	records[17].CarrierNetCode = 0xabcd
	if got.TeamID != 17 || got.FlagIndex != 3 || got.Status != 2 ||
		got.Reserved != 0xa5 || got.CarrierNetCode != 0xabcd {
		t.Fatalf("returned record = %#v, want live source values", *got)
	}
	got.Status = 0x42
	if records[17].Status != 0x42 {
		t.Fatalf("source status = %#x, want returned-pointer write 0x42", records[17].Status)
	}
}

func TestTeamFlagStatusGetter4E8320ReturnsNilForZeroIndexWithoutFault(t *testing.T) {
	if got := teamFlagStatusGetter4E8320(nil, 0x100); got != nil {
		t.Fatalf("nil-base zero-index pointer = %p, want nil", got)
	}
}
