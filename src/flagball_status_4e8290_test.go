package opennox

import (
	"fmt"
	"math"
	"reflect"
	"testing"
	"unsafe"
)

func TestGameBallStatus4E8290StoresInOrderThenBroadcasts(t *testing.T) {
	var events []string
	state := uint8(0x11)
	netCode := uint16(0x2233)
	wantResult := int32(math.MinInt32 + 17)
	got := gameBallStatus4E8290(0xfe, 0xabcd, gameBallStatusHooks4E8290{
		storeState: func(value uint8) {
			events = append(events, fmt.Sprintf("state:%02x", value))
			state = value
		},
		storeNetCode: func(value uint16) {
			events = append(events, fmt.Sprintf("code:%04x", value))
			netCode = value
		},
		send: func(recipient int32, sentState uint8, sentNetCode uint16) int32 {
			events = append(events, fmt.Sprintf("send:%d:%02x:%04x", recipient, sentState, sentNetCode))
			if state != 0xfe || netCode != 0xabcd {
				t.Fatalf("state at send = %02x/%04x, want fe/abcd", state, netCode)
			}
			return wantResult
		},
	})
	if got != wantResult {
		t.Fatalf("result = %d, want %d", got, wantResult)
	}
	wantEvents := []string{"state:fe", "code:abcd", "send:255:fe:abcd"}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
}

func TestGameBallStatus4E8290PreservesRecordGapAndFullWidths(t *testing.T) {
	record := gameBallStatusRecord4E8290{
		State:    0x12,
		Reserved: 0x7b,
		NetCode:  0x3456,
	}
	const wantResult int32 = -1
	got := gameBallStatusNative4E8290(&record, math.MaxUint8, math.MaxUint16,
		func(recipient int32, state uint8, netCode uint16) int32 {
			if recipient != 255 || state != math.MaxUint8 || netCode != math.MaxUint16 {
				t.Fatalf("send = (%d, %#x, %#x), want (255, 0xff, 0xffff)", recipient, state, netCode)
			}
			return wantResult
		})
	if got != wantResult {
		t.Fatalf("result = %d, want %d", got, wantResult)
	}
	if record.State != math.MaxUint8 || record.NetCode != math.MaxUint16 {
		t.Fatalf("record = %#v, want full-width values", record)
	}
	if record.Reserved != 0x7b {
		t.Fatalf("reserved byte = %#x, want 0x7b", record.Reserved)
	}
	if got, want := unsafe.Sizeof(record), uintptr(4); got != want {
		t.Fatalf("record size = %d, want %d", got, want)
	}
	if got, want := unsafe.Offsetof(record.NetCode), uintptr(2); got != want {
		t.Fatalf("net-code offset = %d, want %d", got, want)
	}
}

func TestGameBallStatus4E8290ZeroStillBroadcasts(t *testing.T) {
	record := gameBallStatusRecord4E8290{
		State:    7,
		Reserved: 0xa5,
		NetCode:  9,
	}
	called := false
	got := gameBallStatusNative4E8290(&record, 0, 0, func(recipient int32, state uint8, netCode uint16) int32 {
		called = true
		if recipient != 255 || state != 0 || netCode != 0 {
			t.Fatalf("send = (%d, %d, %d), want (255, 0, 0)", recipient, state, netCode)
		}
		return math.MaxInt32
	})
	if got != math.MaxInt32 {
		t.Fatalf("result = %d, want %d", got, int32(math.MaxInt32))
	}
	if !called {
		t.Fatal("zero status was not broadcast")
	}
	if record.State != 0 || record.NetCode != 0 || record.Reserved != 0xa5 {
		t.Fatalf("record = %#v, want zero state/code and preserved reserved byte", record)
	}
}

func TestGameBallStatus4E8290NilRecordFaultsBeforeSend(t *testing.T) {
	sent := false
	defer func() {
		if recover() == nil {
			t.Fatal("nil record did not panic")
		}
		if sent {
			t.Fatal("send occurred after nil-record state-store fault")
		}
	}()
	gameBallStatusNative4E8290(nil, 1, 2, func(int32, uint8, uint16) int32 {
		sent = true
		return 0
	})
}

func TestGameBallStatus4E8290StoreFaultsShortCircuit(t *testing.T) {
	t.Run("state", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("state store did not panic")
			}
		}()
		gameBallStatus4E8290(1, 2, gameBallStatusHooks4E8290{
			storeState: func(uint8) { panic("state") },
			storeNetCode: func(uint16) {
				t.Fatal("net-code store after state fault")
			},
			send: func(int32, uint8, uint16) int32 {
				t.Fatal("send after state fault")
				return 0
			},
		})
	})

	t.Run("net-code", func(t *testing.T) {
		storedState := false
		defer func() {
			if recover() == nil {
				t.Fatal("net-code store did not panic")
			}
			if !storedState {
				t.Fatal("state was not stored before net-code fault")
			}
		}()
		gameBallStatus4E8290(1, 2, gameBallStatusHooks4E8290{
			storeState: func(uint8) { storedState = true },
			storeNetCode: func(uint16) {
				panic("net-code")
			},
			send: func(int32, uint8, uint16) int32 {
				t.Fatal("send after net-code fault")
				return 0
			},
		})
	})
}
