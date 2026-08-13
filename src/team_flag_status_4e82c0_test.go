package opennox

import (
	"fmt"
	"math"
	"reflect"
	"testing"
	"unsafe"
)

func TestTeamFlagStatus4E82C0StoresInOrderThenBroadcasts(t *testing.T) {
	var events []string
	wantResult := int32(math.MinInt32 + 29)
	got := teamFlagStatus4E82C0(0xfe, 0x81, 0xa7, 0xbcde, teamFlagStatusHooks4E82C0{
		storeTeamID: func(value uint8) {
			events = append(events, fmt.Sprintf("team:%02x", value))
		},
		storeFlagIndex: func(value uint8) {
			events = append(events, fmt.Sprintf("flag:%02x", value))
		},
		storeStatus: func(value uint8) {
			events = append(events, fmt.Sprintf("status:%02x", value))
		},
		storeCarrierNetCode: func(value uint16) {
			events = append(events, fmt.Sprintf("code:%04x", value))
		},
		send: func(recipient int32, teamID, status, flagIndex uint8, carrierNetCode uint16) int32 {
			events = append(events, fmt.Sprintf("send:%d:%02x:%02x:%02x:%04x",
				recipient, teamID, status, flagIndex, carrierNetCode))
			return wantResult
		},
	})
	if got != wantResult {
		t.Fatalf("result = %d, want %d", got, wantResult)
	}
	wantEvents := []string{
		"team:fe", "flag:a7", "status:81", "code:bcde", "send:255:fe:81:a7:bcde",
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
}

func TestTeamFlagStatus4E82C0PreservesRecordGapAndFullWidths(t *testing.T) {
	record := teamFlagStatusRecord4E82C0{
		TeamID:         1,
		FlagIndex:      2,
		Status:         3,
		Reserved:       0x7b,
		CarrierNetCode: 4,
	}
	got := teamFlagStatusNative4E82C0(&record, math.MaxUint8, math.MaxUint8, math.MaxUint8, math.MaxUint16,
		func(recipient int32, teamID, status, flagIndex uint8, carrierNetCode uint16) int32 {
			if recipient != 255 || teamID != math.MaxUint8 || status != math.MaxUint8 ||
				flagIndex != math.MaxUint8 || carrierNetCode != math.MaxUint16 {
				t.Fatalf("send = (%d, %#x, %#x, %#x, %#x), want all full-width values",
					recipient, teamID, status, flagIndex, carrierNetCode)
			}
			return -1
		})
	if got != -1 {
		t.Fatalf("result = %d, want -1", got)
	}
	if record.TeamID != math.MaxUint8 || record.FlagIndex != math.MaxUint8 ||
		record.Status != math.MaxUint8 || record.CarrierNetCode != math.MaxUint16 {
		t.Fatalf("record = %#v, want full-width values", record)
	}
	if record.Reserved != 0x7b {
		t.Fatalf("reserved byte = %#x, want 0x7b", record.Reserved)
	}
	if got, want := unsafe.Sizeof(record), uintptr(6); got != want {
		t.Fatalf("record size = %d, want %d", got, want)
	}
	if got, want := unsafe.Offsetof(record.CarrierNetCode), uintptr(4); got != want {
		t.Fatalf("carrier net-code offset = %d, want %d", got, want)
	}
}

func TestTeamFlagStatus4E82C0UsesLowByteIndexStride(t *testing.T) {
	tests := []struct {
		teamID uint8
		want   uintptr
	}{
		{teamID: 0, want: teamFlagStatusBaseOffset4E82C0},
		{teamID: 1, want: teamFlagStatusBaseOffset4E82C0 + 6},
		{teamID: math.MaxUint8, want: teamFlagStatusBaseOffset4E82C0 + 6*math.MaxUint8},
	}
	for _, tc := range tests {
		if got := teamFlagStatusRecordOffset4E82C0(tc.teamID); got != tc.want {
			t.Fatalf("offset(%d) = %d, want %d", tc.teamID, got, tc.want)
		}
	}
}

func TestTeamFlagStatus4E82C0ZeroStillBroadcasts(t *testing.T) {
	record := teamFlagStatusRecord4E82C0{Reserved: 0xa5}
	called := false
	got := teamFlagStatusNative4E82C0(&record, 0, 0, 0, 0,
		func(recipient int32, teamID, status, flagIndex uint8, carrierNetCode uint16) int32 {
			called = true
			if recipient != 255 || teamID != 0 || status != 0 || flagIndex != 0 || carrierNetCode != 0 {
				t.Fatalf("zero send = (%d, %d, %d, %d, %d)",
					recipient, teamID, status, flagIndex, carrierNetCode)
			}
			return math.MaxInt32
		})
	if got != math.MaxInt32 || !called {
		t.Fatalf("result/called = %d/%v, want %d/true", got, called, int32(math.MaxInt32))
	}
	if record != (teamFlagStatusRecord4E82C0{Reserved: 0xa5}) {
		t.Fatalf("record = %#v, want zero fields and preserved reserved byte", record)
	}
}

func TestTeamFlagStatus4E82C0NilRecordFaultsBeforeSend(t *testing.T) {
	sent := false
	defer func() {
		if recover() == nil {
			t.Fatal("nil record did not panic")
		}
		if sent {
			t.Fatal("send occurred after nil-record team-ID store fault")
		}
	}()
	teamFlagStatusNative4E82C0(nil, 1, 2, 3, 4,
		func(int32, uint8, uint8, uint8, uint16) int32 {
			sent = true
			return 0
		})
}

func TestTeamFlagStatus4E82C0StoreFaultsShortCircuit(t *testing.T) {
	steps := []string{"team", "flag", "status", "code"}
	for faultAt := range steps {
		faultAt := faultAt
		t.Run(steps[faultAt], func(t *testing.T) {
			completed := 0
			store := func(step int) {
				if step == faultAt {
					panic(steps[step])
				}
				if step != completed {
					t.Fatalf("store step = %d, want %d", step, completed)
				}
				completed++
			}
			defer func() {
				if recover() == nil {
					t.Fatalf("%s store did not panic", steps[faultAt])
				}
				if completed != faultAt {
					t.Fatalf("completed stores = %d, want %d", completed, faultAt)
				}
			}()
			teamFlagStatus4E82C0(1, 2, 3, 4, teamFlagStatusHooks4E82C0{
				storeTeamID:         func(uint8) { store(0) },
				storeFlagIndex:      func(uint8) { store(1) },
				storeStatus:         func(uint8) { store(2) },
				storeCarrierNetCode: func(uint16) { store(3) },
				send: func(int32, uint8, uint8, uint8, uint16) int32 {
					t.Fatal("send after store fault")
					return 0
				},
			})
		})
	}
}
