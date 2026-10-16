package opennox

import (
	"bytes"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// These are actual spellWall/NetSendPacketXxx0 calls with a synthetic external
// sender. They do not run a socket, C important-packet service, map, or client.
type magicWallPacketR281 struct {
	recipient int
	payload   []byte
	related   *server.Object
	remove    int
	sequence  int
}

func magicWallPlayerBytesR281(pl *server.Player) []byte {
	return append([]byte(nil), unsafe.Slice((*byte)(unsafe.Pointer(pl)), int(unsafe.Sizeof(*pl)))...)
}

func magicWallReadyPlayerR281(t *testing.T, index byte, flags uint32, oldByte byte) *server.Player {
	t.Helper()
	pl, free := alloc.New(server.Player{})
	original := magicWallPlayerBytesR281(pl)
	t.Cleanup(free)
	t.Cleanup(func() {
		copy(unsafe.Slice((*byte)(unsafe.Pointer(pl)), int(unsafe.Sizeof(*pl))), original)
	})
	// alloc.New's argument supplies the type, not the initializer.
	*pl = server.Player{PlayerInd: index, Field3680: flags}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		if unsafe.Offsetof(server.Player{}.Field3680) != 4976 ||
			unsafe.Offsetof(server.Player{}.Armor)+unsafe.Sizeof(server.EquipmentData{}) != 3680 ||
			unsafe.Pointer(&pl.Armor[1].Field0) != unsafe.Add(unsafe.Pointer(pl), 3680) {
			t.Fatal("native Player status/old-offset control layout differs")
		}
		// Only a C-owned record is poisoned. The typed scalar also proves this
		// fixture does not fabricate a Go pointer in a scanned Player object.
		pl.Armor[1].Field0 = uint32(oldByte)
		if *(*byte)(unsafe.Add(unsafe.Pointer(pl), 3680)) != oldByte || pl.Field3680 != flags {
			t.Fatal("legacy-offset byte and native status were not independently initialized")
		}
	} else if unsafe.Sizeof(uintptr(0)) == 4 {
		if unsafe.Offsetof(server.Player{}.Field3680) != 3680 {
			t.Fatal("PE32 Player status layout differs")
		}
		// The two addresses alias on32. Do not poison the actual status field;
		// these cases retain the same named-field control expectations.
	} else {
		t.Fatal("unsupported Player pointer width")
	}
	return pl
}

func magicWallSenderR281(result int, afterSend func()) (*spellWall, *[]magicWallPacketR281) {
	var packets []magicWallPacketR281
	native := &server.Server{
		NetSendPacketXxx: func(recipient int, payload []byte, related *server.Object, remove, sequence int) int {
			packets = append(packets, magicWallPacketR281{
				recipient: recipient,
				payload:   append([]byte(nil), payload...),
				related:   related,
				remove:    remove,
				sequence:  sequence,
			})
			if afterSend != nil {
				afterSend()
			}
			return result
		},
	}
	return &spellWall{s: &Server{Server: native}}, &packets
}

func TestMagicWallPlayerReadyNativeLayoutR281(t *testing.T) {
	for _, tc := range []struct {
		name    string
		flags   uint32
		oldByte byte
		sends   bool
	}{
		{"ready-versus-clear-old-offset", 0x10, 0, true},
		{"unready-versus-set-old-offset", 0, 0x10, false},
		{"upper-bits-only-versus-set-old-offset", 0x80010000, 0x10, false},
		{"ready-and-upper-bits-versus-clear-old-offset", 0x80010010, 0, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pl := magicWallReadyPlayerR281(t, 3, tc.flags, tc.oldByte)
			before := magicWallPlayerBytesR281(pl)
			wl := &server.Wall{Tile1: 0x91, Dir0: 0xa2, Field2: 0xb3, X5: 0xc4, Y6: 0xd5}
			oldWall := *wl
			mw := &server.MagicWall{Wall8: wl, PlayerBits16: 0x80000000}
			sp, packets := magicWallSenderR281(1, nil)
			sp.list = mw
			sp.changeOrAddRemoteWalls(pl)
			wantBits := uint32(0x80000000)
			if tc.sends {
				want := []magicWallPacketR281{{recipient: 3, payload: []byte{61, 0x91, 0xa2, 0xb3, 0xc4, 0xd5}, remove: 1, sequence: 0}}
				if !reflect.DeepEqual(*packets, want) {
					t.Errorf("actual helper/forwarder sends = %+v, want %+v", *packets, want)
				}
				wantBits |= 8
			} else if len(*packets) != 0 {
				t.Errorf("unready helper sent %+v", *packets)
			}
			if mw.PlayerBits16 != wantBits {
				t.Errorf("delivery bits = %#x, want %#x", mw.PlayerBits16, wantBits)
			}
			if *wl != oldWall || !bytes.Equal(magicWallPlayerBytesR281(pl), before) {
				t.Fatal("wall payload source or full native Player record changed")
			}
		})
	}
}

func TestMagicWallRemotePayloadHostAndDedupR281(t *testing.T) {
	t.Run("host-is-excluded-before-wall-read", func(t *testing.T) {
		pl := magicWallReadyPlayerR281(t, 31, 0x10, 0x10)
		before := magicWallPlayerBytesR281(pl)
		mw := &server.MagicWall{PlayerBits16: 0x12345678}
		sp, packets := magicWallSenderR281(1, nil)
		sp.list = mw // Wall8 is deliberately nil: host exclusion precedes use.
		sp.changeOrAddRemoteWalls(pl)
		if len(*packets) != 0 || mw.PlayerBits16 != 0x12345678 || !bytes.Equal(magicWallPlayerBytesR281(pl), before) {
			t.Fatal("host exclusion changed sends, wall mask, or native Player")
		}
	})
	t.Run("empty-list-is-inert", func(t *testing.T) {
		pl := magicWallReadyPlayerR281(t, 3, 0x10, 0x10)
		before := magicWallPlayerBytesR281(pl)
		sp, packets := magicWallSenderR281(1, nil)
		sp.changeOrAddRemoteWalls(pl)
		if len(*packets) != 0 || !bytes.Equal(magicWallPlayerBytesR281(pl), before) {
			t.Fatal("empty list changed sends or native Player")
		}
	})
	t.Run("ordered-payload-dedup-and-failed-send-publication", func(t *testing.T) {
		pl := magicWallReadyPlayerR281(t, 3, 0x10, 0x10)
		before := magicWallPlayerBytesR281(pl)
		first := &server.MagicWall{Wall8: &server.Wall{Tile1: 0x11, Dir0: 0x22, Field2: 0x33, X5: 0x44, Y6: 0x55}, PlayerBits16: 1}
		seen := &server.MagicWall{PlayerBits16: 8} // Nil Wall8 must remain unread.
		last := &server.MagicWall{Wall8: &server.Wall{Tile1: 0x66, Dir0: 0x77, Field2: 0x88, X5: 0x99, Y6: 0xaa}, PlayerBits16: 4}
		first.Next24, seen.Next24 = seen, last
		sp, packets := magicWallSenderR281(0, nil)
		sp.list = first
		sp.changeOrAddRemoteWalls(pl)
		want := []magicWallPacketR281{
			{recipient: 3, payload: []byte{61, 0x11, 0x22, 0x33, 0x44, 0x55}, remove: 1, sequence: 0},
			{recipient: 3, payload: []byte{61, 0x66, 0x77, 0x88, 0x99, 0xaa}, remove: 1, sequence: 0},
		}
		if !reflect.DeepEqual(*packets, want) || first.PlayerBits16 != 9 || seen.PlayerBits16 != 8 || last.PlayerBits16 != 12 {
			t.Fatalf("ordered send/mask contract differs: packets=%+v masks=%#x,%#x,%#x", *packets, first.PlayerBits16, seen.PlayerBits16, last.PlayerBits16)
		}
		sp.changeOrAddRemoteWalls(pl)
		if !reflect.DeepEqual(*packets, want) || first.PlayerBits16 != 9 || seen.PlayerBits16 != 8 || last.PlayerBits16 != 12 || !bytes.Equal(magicWallPlayerBytesR281(pl), before) {
			t.Fatal("repeat delivery was not deduped or native Player changed")
		}
	})
}

func TestMagicWallReadyReadRemainsLiveR281(t *testing.T) {
	pl := magicWallReadyPlayerR281(t, 3, 0x10, 0x10)
	before := magicWallPlayerBytesR281(pl)
	first := &server.MagicWall{Wall8: &server.Wall{Tile1: 1, Dir0: 2, Field2: 3, X5: 4, Y6: 5}, PlayerBits16: 1}
	last := &server.MagicWall{Wall8: &server.Wall{Tile1: 6, Dir0: 7, Field2: 8, X5: 9, Y6: 10}, PlayerBits16: 2}
	first.Next24 = last
	calls := 0
	sp, packets := magicWallSenderR281(1, func() {
		calls++
		if calls == 1 {
			pl.Field3680 = 0
		}
	})
	sp.list = first
	sp.changeOrAddRemoteWalls(pl)
	want := []magicWallPacketR281{{recipient: 3, payload: []byte{61, 1, 2, 3, 4, 5}, remove: 1, sequence: 0}}
	if calls != 1 || !reflect.DeepEqual(*packets, want) || first.PlayerBits16 != 9 || last.PlayerBits16 != 2 {
		t.Errorf("live readiness reload differs: calls=%d packets=%+v masks=%#x,%#x", calls, *packets, first.PlayerBits16, last.PlayerBits16)
	}
	status := int(unsafe.Offsetof(server.Player{}.Field3680))
	wantPlayer := append([]byte(nil), before...)
	clear(wantPlayer[status : status+4])
	if !bytes.Equal(magicWallPlayerBytesR281(pl), wantPlayer) {
		t.Fatal("sender callback changed Player bytes beyond the declared status mutation")
	}
}
