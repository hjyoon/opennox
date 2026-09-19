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

func TestClientTradePacket42E850(t *testing.T) {
	dr := &client.Drawable{NetCode32: 0x1234, ObjClass: object.Class(clientStaticUnitMask578B00)}
	tests := []struct {
		name           string
		dr             *client.Drawable
		playerStatus   uint32
		npcDialogState int
		quitMenuState  int
		want           [4]byte
		ok             bool
	}{
		{name: "nil drawable"},
		{name: "status bit zero", dr: dr, playerStatus: 1},
		{name: "status bit one", dr: dr, playerStatus: 2},
		{name: "both status bits", dr: dr, playerStatus: 3},
		{name: "npc dialog active", dr: dr, npcDialogState: 1},
		{name: "quit menu active", dr: dr, quitMenuState: 1},
		{name: "noncanonical npc dialog state remains allowed", dr: dr, npcDialogState: 2, want: [4]byte{0xC9, 0x15, 0x34, 0x92}, ok: true},
		{name: "noncanonical quit state remains allowed", dr: dr, quitMenuState: 2, want: [4]byte{0xC9, 0x15, 0x34, 0x92}, ok: true},
		{name: "wire code preserves static high bit", dr: dr, want: [4]byte{0xC9, 0x15, 0x34, 0x92}, ok: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, ok := clientTradePacket42E850(test.dr, test.playerStatus, test.npcDialogState, test.quitMenuState)
			if got != test.want || ok != test.ok {
				t.Fatalf("packet = % X, %t; want % X, %t", got, ok, test.want, test.ok)
			}
		})
	}
}

func TestClientTrade42E850PreservesNativePointersAndQueuesPacket(t *testing.T) {
	list := netlist.New()
	list.Init()
	t.Cleanup(list.Free)
	oldGetServer := GetServer
	GetServer = func() Server {
		return &netClientSendTestServer{srv: &server.Server{NetList: list}}
	}
	t.Cleanup(func() { GetServer = oldGetServer })

	oldPlayer := Get_dword_8531A0_2576()
	oldDialogState := Sub_47A260()
	Set_dword_5d4594_1123520(0)

	dr := new(client.Drawable)
	player := new(server.Player)
	var pin runtime.Pinner
	pin.Pin(dr)
	pin.Pin(player)
	t.Cleanup(func() {
		Set_dword_8531A0_2576(oldPlayer)
		Set_dword_5d4594_1123520(oldDialogState)
		pin.Unpin()
	})
	dr.NetCode32 = 0x1234
	if unsafe.Sizeof(uintptr(0)) == 8 {
		if uintptr(unsafe.Pointer(dr)) <= math.MaxUint32 {
			t.Fatalf("drawable pointer = %p, want native address above 4 GiB", dr)
		}
		if uintptr(unsafe.Pointer(player)) <= math.MaxUint32 {
			t.Fatalf("player pointer = %p, want native address above 4 GiB", player)
		}
	}

	Set_dword_8531A0_2576(player)
	Nox_xxx_clientTrade_42E850(dr)
	want := []byte{0xC9, 0x15, 0x34, 0x12}
	if got := list.CopyPacketsA(ntype.PlayerInd(31), netlist.Kind0); !bytes.Equal(got, want) {
		t.Fatalf("queued packet = % X, want % X", got, want)
	}

	player.Field3680 = 2
	Nox_xxx_clientTrade_42E850(dr)
	if got := list.CopyPacketsA(ntype.PlayerInd(31), netlist.Kind0); len(got) != 0 {
		t.Fatalf("status-gated queue = % X, want empty", got)
	}
	runtime.KeepAlive(dr)
	runtime.KeepAlive(player)
}
