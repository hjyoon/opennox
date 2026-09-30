package legacy

import (
	"bytes"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestNetSendSimpleObject4DF360CEntryPreservesNativePointerAndX87Coordinates(t *testing.T) {
	obj, free := alloc.New(server.Object{})
	defer free()
	*obj = server.Object{TypeInd: 0xabc, NetCode: 0x1234, PosVec: types.Pointf{X: 2.5, Y: -3.5}}
	if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(unsafe.Pointer(obj)) <= math.MaxUint32 {
		t.Fatalf("object=%p, want above 4 GiB", obj)
	}
	s := new(server.Server)
	s.NetSendPacketXxx = func(recipient int, packet []byte, related *server.Object, remove, sequence int) int {
		want := []byte{0x2f, 0x34, 0x12, 0xbc, 0x0a, 2, 0, 0xfc, 0xff}
		if recipient != -0x123456 || !bytes.Equal(packet, want) || related != nil || remove != 1 || sequence != 1 {
			t.Fatalf("routing=%d/%p/%d/%d packet=%x", recipient, related, remove, sequence, packet)
		}
		return -17
	}
	oldServer := GetServer
	GetServer = func() Server { return &netClientSendTestServer{srv: s} }
	defer func() { GetServer = oldServer }()
	if got := netSendSimpleObjectCEntry4DF360(-0x123456, obj); got != -17 {
		t.Fatalf("C return=%d, want -17", got)
	}
}

func TestNetSendSimpleObject4DF360X87Conversion(t *testing.T) {
	for _, tc := range []struct {
		value float32
		want  int32
	}{
		{2.5, 2}, {3.5, 4}, {-2.5, -2}, {-3.5, -4},
		{0.75, 1}, {-0.75, -1}, {65536.5, 65536},
		{float32(math.NaN()), math.MinInt32},
		{float32(math.Inf(1)), math.MinInt32},
		{float32(math.Inf(-1)), math.MinInt32},
		{2147483648, math.MinInt32}, {-2147483648, math.MinInt32},
	} {
		if got := netSendSimpleObjectFloatToInt4DF360(tc.value); got != tc.want {
			t.Errorf("round(%v)=%d, want %d", tc.value, got, tc.want)
		}
	}
}
