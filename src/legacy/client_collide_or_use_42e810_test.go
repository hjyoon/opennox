package legacy

import (
	"bytes"
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/internal/netlist"
	"github.com/opennox/opennox/v1/server"
)

func TestClientWireUnitCode578B00(t *testing.T) {
	tests := []struct {
		name string
		dr   *client.Drawable
		want uint16
	}{
		{name: "nil", want: 0},
		{name: "dynamic", dr: &client.Drawable{NetCode32: 0x1234}, want: 0x1234},
		{name: "first static class bit", dr: &client.Drawable{NetCode32: 0x1234, ObjClass: object.Class(0x00400000)}, want: 0x9234},
		{name: "second static class bit", dr: &client.Drawable{NetCode32: 0x1234, ObjClass: object.Class(0x20000000)}, want: 0x9234},
		{name: "unrelated class bit", dr: &client.Drawable{NetCode32: 0x1234, ObjClass: object.Class(0x00800000)}, want: 0x1234},
		{name: "largest valid static code", dr: &client.Drawable{NetCode32: 0x7FFF, ObjClass: object.Class(clientStaticUnitMask578B00)}, want: 0xFFFF},
		{name: "first invalid code", dr: &client.Drawable{NetCode32: 0x8000, ObjClass: object.Class(clientStaticUnitMask578B00)}, want: 0},
		{name: "largest invalid code", dr: &client.Drawable{NetCode32: math.MaxUint32, ObjClass: object.Class(clientStaticUnitMask578B00)}, want: 0},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := clientWireUnitCode578B00(test.dr); got != test.want {
				t.Fatalf("wire unit code = %#x, want %#x", got, test.want)
			}
		})
	}
}

func TestClientCollideOrUsePacket42E810(t *testing.T) {
	dr := &client.Drawable{NetCode32: 0x1234, ObjClass: object.Class(clientStaticUnitMask578B00)}
	invalid := &client.Drawable{NetCode32: 0x8000, ObjClass: object.Class(clientStaticUnitMask578B00)}
	tests := []struct {
		name   string
		dr     *client.Drawable
		status uint32
		want   [3]byte
		ok     bool
	}{
		{name: "nil drawable"},
		{name: "status bit zero", dr: dr, status: 1},
		{name: "status bit one", dr: dr, status: 2},
		{name: "both status bits", dr: dr, status: 3},
		{name: "unrelated status", dr: dr, status: 0xFFFFFFFC, want: [3]byte{0x7B, 0x34, 0x92}, ok: true},
		{name: "invalid code still sends", dr: invalid, want: [3]byte{0x7B, 0, 0}, ok: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, ok := clientCollideOrUsePacket42E810(test.dr, test.status)
			if got != test.want || ok != test.ok {
				t.Fatalf("packet = % X, %t; want % X, %t", got, ok, test.want, test.ok)
			}
		})
	}
}

func TestClientCollideOrUse42E810PreservesNativePointerAndQueuesPacket(t *testing.T) {
	list := netlist.New()
	list.Init()
	t.Cleanup(list.Free)
	oldGetServer := GetServer
	GetServer = func() Server {
		return &netClientSendTestServer{srv: &server.Server{NetList: list}}
	}
	t.Cleanup(func() { GetServer = oldGetServer })

	oldPlayer := Get_dword_8531A0_2576()
	Set_dword_8531A0_2576(nil)

	dr := new(client.Drawable)
	player := new(server.Player)
	var pin runtime.Pinner
	pin.Pin(dr)
	pin.Pin(player)
	t.Cleanup(func() {
		Set_dword_8531A0_2576(oldPlayer)
		pin.Unpin()
	})
	dr.NetCode32 = 0x1234
	dr.ObjClass = object.Class(clientStaticUnitMask578B00)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(dr)) <= math.MaxUint32 {
		t.Fatalf("drawable pointer = %p, want native address above 4 GiB", dr)
	}

	Nox_xxx_clientCollideOrUse_42E810(dr)
	want := []byte{0x7B, 0x34, 0x92}
	if got := list.CopyPacketsA(ntype.PlayerInd(31), netlist.Kind0); !bytes.Equal(got, want) {
		t.Fatalf("queued packet = % X, want % X", got, want)
	}

	player.Field3680 = 2
	Set_dword_8531A0_2576(player)
	Nox_xxx_clientCollideOrUse_42E810(dr)
	if got := list.CopyPacketsA(ntype.PlayerInd(31), netlist.Kind0); len(got) != 0 {
		t.Fatalf("status-gated queue = % X, want empty", got)
	}
	Set_dword_8531A0_2576(nil)
	runtime.KeepAlive(dr)
	runtime.KeepAlive(player)
}
