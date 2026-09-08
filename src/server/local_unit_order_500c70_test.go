package server

import (
	"reflect"
	"testing"

	"github.com/opennox/opennox/v1/common/ntype"
)

func TestLocalUnitOrder500C70PreservesFullOrderAndNarrowsPacketByte(t *testing.T) {
	const (
		owner   = ntype.PlayerInd(0x1357)
		player  = uint64(0x123456789abcdef0)
		wantRet = -0x1234567
	)
	orderType := uint32(0x89abcdef)
	var events []string
	hooks := localUnitOrderHooks500C70[uint64]{
		playerByIndex: func(gotOwner ntype.PlayerInd) uint64 {
			events = append(events, "lookup")
			if gotOwner != owner {
				t.Fatalf("lookup owner = %#x, want %#x", gotOwner, owner)
			}
			return player
		},
		storeOrder: func(gotPlayer uint64, gotOrder uint32) {
			events = append(events, "store")
			if gotPlayer != player || gotOrder != orderType {
				t.Fatalf("store args = (%#x, %#x), want (%#x, %#x)", gotPlayer, gotOrder, player, orderType)
			}
		},
		sendCreatureCommand: func(gotOwner ntype.PlayerInd, command byte) int {
			events = append(events, "send")
			if gotOwner != owner || command != byte(orderType) {
				t.Fatalf("send args = (%#x, %#x), want (%#x, %#x)", gotOwner, command, owner, byte(orderType))
			}
			return wantRet
		},
	}

	if got := localUnitOrder500C70(owner, orderType, hooks); got != wantRet {
		t.Fatalf("result = %d, want %d", got, wantRet)
	}
	if want := []string{"lookup", "store", "send"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestLocalUnitOrder500C70FaultPrefixes(t *testing.T) {
	tests := []struct {
		name       string
		panicAt    string
		wantEvents []string
	}{
		{name: "lookup", panicAt: "lookup", wantEvents: []string{"lookup"}},
		{name: "store", panicAt: "store", wantEvents: []string{"lookup", "store"}},
		{name: "send", panicAt: "send", wantEvents: []string{"lookup", "store", "send"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var events []string
			hooks := localUnitOrderHooks500C70[uint64]{
				playerByIndex: func(ntype.PlayerInd) uint64 {
					events = append(events, "lookup")
					if tc.panicAt == "lookup" {
						panic("lookup")
					}
					return 0x1_0000_0001
				},
				storeOrder: func(uint64, uint32) {
					events = append(events, "store")
					if tc.panicAt == "store" {
						panic("store")
					}
				},
				sendCreatureCommand: func(ntype.PlayerInd, byte) int {
					events = append(events, "send")
					panic("send")
				},
			}
			func() {
				defer func() {
					if recover() == nil {
						t.Fatal("call did not panic")
					}
				}()
				localUnitOrder500C70(3, 0xfedcba98, hooks)
			}()
			if !reflect.DeepEqual(events, tc.wantEvents) {
				t.Fatalf("events = %v, want %v", events, tc.wantEvents)
			}
		})
	}
}
