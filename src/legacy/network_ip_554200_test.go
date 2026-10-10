package legacy

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/netip"
	"os"
	"os/exec"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/strman"
	"github.com/opennox/opennox/v1/internal/netstr"
	"github.com/opennox/opennox/v1/server"
)

// Prepare native stream input state without opening a socket. The result is
// read only through the actual exported 554200 callback, not an output hook.
func TestNetworkConnectionIP554200(t *testing.T) {
	if os.Getenv("NOX_NETWORK_IP_CGO_CHILD") != "1" {
		cmd := exec.Command(os.Args[0], "-test.run=^TestNetworkConnectionIP554200$", "-test.count=1", "-test.v")
		cmd.Env = append(os.Environ(), "NOX_NETWORK_IP_CGO_CHILD=1")
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("actual connection IP callback: %v\n%s", err, output)
		}
		return
	}
	srv := server.New(nil, nil, strman.New())
	GetServer = func() Server { return &playerTrackingLegacyServer425F10{srv: srv} }
	field := reflect.ValueOf(srv.NetStr).Elem().FieldByName("streams")
	if field.Len() != 128 || field.Type().Elem() != reflect.TypeOf((*netstr.Conn)(nil)) {
		t.Fatal("native stream fixture layout changed")
	}
	streams := (*[128]*netstr.Conn)(unsafe.Pointer(field.UnsafeAddr()))
	indices := make([]uint32, 257, 262)
	for i := range indices {
		indices[i] = uint32(i)
	}
	indices = append(indices, 0xffffffff, 0xffffff80, 0x80000000, 0x7fffffff, 0x10000)
	var transcript bytes.Buffer
	encoder := json.NewEncoder(&transcript)
	for variant := 0; variant < 2; variant++ {
		srv.OwnIP = int2ip(0x80123456 ^ uint32(variant))
		for index := range streams {
			streams[index] = nil
			if index == 0 || variant == 0 || index%3 == 0 {
				continue
			}
			conn := new(netstr.Conn)
			address := reflect.ValueOf(conn).Elem().FieldByName("addr")
			if address.Type() != reflect.TypeOf(netip.AddrPort{}) {
				t.Fatal("native peer address fixture layout changed")
			}
			*(*netip.AddrPort)(unsafe.Pointer(address.UnsafeAddr())) = netip.AddrPortFrom(int2ip(0xabcdef00^uint32(index*0x10203)), 12345)
			streams[index] = conn
		}
		for _, index := range indices {
			want := uint32(0)
			if index == 0 {
				want = 0x80123456 ^ uint32(variant)
			} else if index < 128 && variant != 0 && index%3 != 0 {
				want = 0xabcdef00 ^ (index * 0x10203)
			}
			got := nox_xxx_net_getIP_554200(int32(index))
			if got != want {
				t.Fatalf("variant %d connection %#x: got %#x, want %#x", variant, index, got, want)
			}
			if err := encoder.Encode(struct {
				Variant int    `json:"variant"`
				Index   uint32 `json:"index"`
				Result  uint32 `json:"result"`
			}{variant, index, got}); err != nil {
				t.Fatal(err)
			}
		}
	}
	// Slot 128 is deliberately vacant in the original input table: native
	// Streams cannot allocate it. Do not claim equivalence for an occupied one.
	if got := fmt.Sprintf("%x", sha256.Sum256(transcript.Bytes())); got != "2440e4821898e82d5a920d423edccc43c663e0f844a8a72771bd5aa280886c6a" {
		t.Fatalf("unchanged PE32 524-case transcript changed: %s", got)
	}
	if receipt := os.Getenv("NOX_NETWORK_IP_NATIVE_JSON"); receipt != "" {
		if err := os.WriteFile(receipt, transcript.Bytes(), 0600); err != nil {
			t.Fatal(err)
		}
	}
	t.Log("524 native connection-index/IP cases passed")
}
