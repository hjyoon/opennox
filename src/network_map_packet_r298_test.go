package opennox

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/opennox/libs/maps"
	netmapsend "github.com/opennox/libs/noxnet/mapsend"
	"github.com/opennox/libs/noxnet/netmsg"

	"github.com/opennox/opennox/v1/client"
)

// These exercise the production packet switch and downloader without a socket,
// renderer, OS window, stock assets, or personal configuration.
func callMapPacketR298(t *testing.T, c *Client, packet []byte) (n int) {
	t.Helper()
	defer func() {
		if p := recover(); p != nil {
			t.Fatalf("map packet len=%d cap=%d panicked: %v", len(packet), cap(packet), p)
		}
	}()
	frame, frame16 := uint32(0xaabbccdd), uint16(0xeeff)
	n = c.nox_xxx_netOnPacketRecvCli48EA70_switch(0, netmsg.MSG_MAP_SEND_PACKET, packet, &frame, &frame16)
	if frame != 0xaabbccdd || frame16 != 0xeeff {
		t.Fatal("map packet changed frame state")
	}
	return n
}

func mapPacketR298(t *testing.T, block uint16, payload, trailing []byte) []byte {
	t.Helper()
	p := netmapsend.MsgMapSendPacket{Unk: 0x93, Block: block, Data: payload}
	packet := make([]byte, 1+p.EncodeSize()+len(trailing))
	packet[0] = byte(p.NetOp())
	n, err := p.Encode(packet[1:])
	if err != nil || n != p.EncodeSize() {
		t.Fatalf("production map packet encoder: n=%d err=%v", n, err)
	}
	copy(packet[1+n:], trailing)
	return packet
}

func mapDownloadClientR298(t *testing.T, total uint) (*Client, string, *int) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "fixture", "fixture.nxz")
	download := maps.NewNativeDownloader(dir)
	if err := download.Start(path, total); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(download.Reset)
	c := &Client{Client: &client.Client{}}
	c.mapsend.ndl = download
	c.mapsend.inProgress = true
	acks := new(int)
	c.mapsend.iface.sendReceivedMap = func() int { *acks++; return 1 }
	return c, path, acks
}

func TestMapSendPacketBoundaryR298(t *testing.T) {
	for _, tc := range []struct {
		name     string
		size     int
		declared uint16
		spare    int
	}{
		{name: "truncated_payload", size: 6, declared: 1},
		{name: "truncated_stock_chunk", size: 517, declared: 512},
		{name: "spare_capacity_is_not_datagram_data", size: 6, declared: 8, spare: 32},
		{name: "wrapped_end", size: 6, declared: 65535},
	} {
		t.Run(tc.name, func(t *testing.T) {
			backing := bytes.Repeat([]byte{0xac}, tc.size+tc.spare)
			packet := backing[:tc.size]
			packet[0] = byte(netmsg.MSG_MAP_SEND_PACKET)
			binary.LittleEndian.PutUint16(packet[2:4], 0xffff)
			binary.LittleEndian.PutUint16(packet[4:6], tc.declared)
			before := append([]byte(nil), backing...)
			c := &Client{Client: &client.Client{}}
			c.mapsend.inProgress, c.mapsend.downloadOK = true, true
			state := c.mapsend
			if got := callMapPacketR298(t, c, packet); got != -1 {
				t.Fatalf("truncated packet consumed %d bytes, want -1", got)
			}
			if !reflect.DeepEqual(c.mapsend, state) || !bytes.Equal(backing, before) {
				t.Fatal("rejected packet touched downloader or input backing")
			}
		})
	}
	for size := 0; size < 6; size++ {
		t.Run(fmt.Sprintf("short_header_%d", size), func(t *testing.T) {
			packet := make([]byte, size)
			if size != 0 {
				packet[0] = byte(netmsg.MSG_MAP_SEND_PACKET)
			}
			want := -1
			if size == 0 {
				want = 0
			} // Existing empty packet switch contract.
			c := &Client{Client: &client.Client{}}
			if got := callMapPacketR298(t, c, packet); got != want {
				t.Fatalf("short header consumed %d bytes, want %d", got, want)
			}
		})
	}
}

func TestMapSendPacketAllTruncatedLengthsR298(t *testing.T) {
	packet := []byte{byte(netmsg.MSG_MAP_SEND_PACKET), 0x93, 1, 0, 0, 0}
	c := &Client{Client: &client.Client{}}
	for declared := 1; declared <= 65535; declared++ {
		binary.LittleEndian.PutUint16(packet[4:6], uint16(declared))
		if got := callMapPacketR298(t, c, packet); got != -1 {
			t.Fatalf("declared=%d returned %d, want -1", declared, got)
		}
	}
}

func TestMapSendPacketValidPayloadsR298(t *testing.T) {
	for _, size := range []int{0, 1, 511, 512, 513, 65529, 65530, 65535} {
		t.Run(fmt.Sprint(size), func(t *testing.T) {
			payload := make([]byte, size)
			for i := range payload {
				payload[i] = byte(i*31 + 17)
			}
			packet := mapPacketR298(t, 1, payload, []byte{0xfa, 0xfb, 0xfc})
			before := append([]byte(nil), packet...)
			total := size
			if total == 0 {
				total = 1
			}
			c, path, acks := mapDownloadClientR298(t, uint(total))
			if got := callMapPacketR298(t, c, packet); got != 6+size {
				t.Fatalf("consumed=%d, want %d without trailing bytes", got, 6+size)
			}
			got, err := os.ReadFile(path)
			if err != nil || !bytes.Equal(got, payload) || !bytes.Equal(packet, before) {
				t.Fatalf("actual download bytes/input changed: got=%d want=%d err=%v", len(got), size, err)
			}
			if size == 0 {
				if *acks != 0 || !c.mapsend.inProgress || c.mapsend.downloadOK {
					t.Fatal("zero payload changed transfer state")
				}
			} else if *acks != 1 || c.mapsend.inProgress || !c.mapsend.downloadOK {
				t.Fatal("complete map did not reach exactly one existing received-map callback")
			}
		})
	}
}

func TestMapSendPacketReorderedTransferR298(t *testing.T) {
	c, path, acks := mapDownloadClientR298(t, 6)
	for _, part := range []struct {
		block uint16
		data  string
	}{{2, "def"}, {1, "abc"}} {
		packet := mapPacketR298(t, part.block, []byte(part.data), []byte{0xff})
		if got := callMapPacketR298(t, c, packet); got != 9 {
			t.Fatalf("block=%d consumed=%d", part.block, got)
		}
		if part.block == 2 && (*acks != 0 || !c.mapsend.inProgress) {
			t.Fatal("out-of-order block finished early")
		}
	}
	got, err := os.ReadFile(path)
	if err != nil || string(got) != "abcdef" || *acks != 1 || c.mapsend.inProgress || !c.mapsend.downloadOK {
		t.Fatalf("reordered transfer got=%q acks=%d err=%v", got, *acks, err)
	}
}
