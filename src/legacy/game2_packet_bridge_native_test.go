package legacy

import (
	"bytes"
	"fmt"
	"github.com/opennox/opennox/v1/common/ntype"
	"runtime"
	"testing"
	"unsafe"
)

func TestGame2PacketBridgeConsumesOneNativeMessage(t *testing.T) {
	old := Nox_xxx_netOnPacketRecvCli_48EA70
	defer func() { Nox_xxx_netOnPacketRecvCli_48EA70 = old }()
	for _, tc := range []struct{ op, length int32 }{{47, 9}, {48, 11}, {195, 12}} {
		for _, viaSwitch := range []bool{false, true} {
			t.Run(fmt.Sprintf("%d/switch_%t", tc.op, viaSwitch), func(t *testing.T) {
				p := accessListNativeFixtureAlloc416720(64)
				if p == nil {
					t.Fatal("C packet allocation")
				}
				defer accessListNativeFixtureFree416720(p)
				buf := unsafe.Slice((*byte)(p), 64)
				for i := range buf {
					buf[i] = byte(0x80 + i)
				}
				buf[0] = byte(tc.op)
				before := bytes.Clone(buf)
				calls := 0
				Nox_xxx_netOnPacketRecvCli_48EA70 = func(player ntype.PlayerInd, data *byte, size int) int {
					calls++
					if player != 3 || unsafe.Pointer(data) != p || size != int(tc.length) {
						t.Fatal("native identity or one-message framing changed")
					}
					runtime.GC()
					if !bytes.Equal(unsafe.Slice(data, size), before[:size]) {
						t.Fatal("native packet bytes changed")
					}
					return 1
				}
				if got := game2PacketBridgeFixture(3, tc.op, p, 64, viaSwitch); got != tc.length || calls != 1 {
					t.Fatalf("consumed=%d calls=%d", got, calls)
				}
				if !bytes.Equal(buf, before) {
					t.Fatal("borrowed packet or following message mutated")
				}
				if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(p) <= 0xffffffff {
					t.Fatal("need high native packet address")
				}
			})
		}
	}
}
func TestGame2PacketBridgeRejectsIncompleteAndUnregistered(t *testing.T) {
	old := Nox_xxx_netOnPacketRecvCli_48EA70
	defer func() { Nox_xxx_netOnPacketRecvCli_48EA70 = old }()
	p := accessListNativeFixtureAlloc416720(32)
	if p == nil {
		t.Fatal("C packet")
	}
	defer accessListNativeFixtureFree416720(p)
	for _, tc := range []struct{ op, length int32 }{{47, 9}, {48, 11}, {195, 12}} {
		*(*byte)(p) = byte(tc.op)
		calls := 0
		Nox_xxx_netOnPacketRecvCli_48EA70 = func(ntype.PlayerInd, *byte, int) int { calls++; return 1 }
		for _, sz := range []int32{-1, 0, tc.length - 1} {
			if game2PacketBridgeFixture(0, tc.op, p, sz, false) != -1 {
				t.Fatal("incomplete message accepted")
			}
		}
		if game2PacketBridgeFixture(0, tc.op, nil, tc.length, false) != -1 || calls != 0 {
			t.Fatal("nil message executed")
		}
		*(*byte)(p) = 0
		if game2PacketBridgeFixture(0, tc.op, p, tc.length, false) != -1 || calls != 0 {
			t.Fatal("opcode mismatch executed")
		}
		*(*byte)(p) = byte(tc.op)
		Nox_xxx_netOnPacketRecvCli_48EA70 = nil
		if game2PacketBridgeFixture(0, tc.op, p, tc.length, false) != -1 {
			t.Fatal("unregistered native handler")
		}
		for _, result := range []int{0, -1, 2} {
			Nox_xxx_netOnPacketRecvCli_48EA70 = func(ntype.PlayerInd, *byte, int) int { return result }
			if game2PacketBridgeFixture(0, tc.op, p, tc.length, false) != -1 {
				t.Fatal("failure interpreted as consumed bytes")
			}
		}
	}
	if game2PacketBridgeFixture(0, 99, p, 32, false) != -1 {
		t.Fatal("unsupported opcode accepted")
	}
}
