package legacy

import (
	"bytes"
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/internal/netlist"
	"github.com/opennox/opennox/v1/server"
)

func TestClientTalkPacket42E7B0(t *testing.T) {
	dr := &client.Drawable{NetCode32: 0x89ABCDEF}
	tests := []struct {
		name           string
		dr             *client.Drawable
		playerStatus   uint32
		inventoryState int
		quitMenuState  int
		want           [4]byte
		ok             bool
	}{
		{name: "nil drawable"},
		{name: "status bit zero", dr: dr, playerStatus: 1},
		{name: "status bit one", dr: dr, playerStatus: 2},
		{name: "both status bits", dr: dr, playerStatus: 3},
		{name: "inventory active", dr: dr, inventoryState: 1},
		{name: "quit menu active", dr: dr, quitMenuState: 1},
		{name: "noncanonical inventory state remains allowed", dr: dr, inventoryState: 2, want: [4]byte{0xD0, 0x01, 0xEF, 0xCD}, ok: true},
		{name: "noncanonical quit state remains allowed", dr: dr, quitMenuState: 2, want: [4]byte{0xD0, 0x01, 0xEF, 0xCD}, ok: true},
		{name: "wire net code uses low word", dr: dr, want: [4]byte{0xD0, 0x01, 0xEF, 0xCD}, ok: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, ok := clientTalkPacket42E7B0(test.dr, test.playerStatus, test.inventoryState, test.quitMenuState)
			if got != test.want || ok != test.ok {
				t.Fatalf("packet = % X, %t; want % X, %t", got, ok, test.want, test.ok)
			}
		})
	}
}

func TestClientTalk42E7B0PreservesNativePointerAndQueuesPacket(t *testing.T) {
	list := netlist.New()
	list.Init()
	t.Cleanup(list.Free)
	oldGetServer := GetServer
	GetServer = func() Server {
		return &netClientSendTestServer{srv: &server.Server{NetList: list}}
	}
	t.Cleanup(func() { GetServer = oldGetServer })

	dr := new(client.Drawable)
	var pin runtime.Pinner
	pin.Pin(dr)
	defer pin.Unpin()
	dr.NetCode32 = 0x89ABCDEF
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(dr)) <= math.MaxUint32 {
		t.Fatalf("drawable pointer = %p, want native address above 4 GiB", dr)
	}

	clientTalk42E7B0(dr, 0, 0, 0)
	want := []byte{0xD0, 0x01, 0xEF, 0xCD}
	if got := list.CopyPacketsA(ntype.PlayerInd(31), netlist.Kind0); !bytes.Equal(got, want) {
		t.Fatalf("queued packet = % X, want % X", got, want)
	}

	for name, gate := range map[string]func(){
		"player status": func() { clientTalk42E7B0(dr, 1, 0, 0) },
		"inventory":     func() { clientTalk42E7B0(dr, 0, 1, 0) },
		"quit menu":     func() { clientTalk42E7B0(dr, 0, 0, 1) },
	} {
		t.Run(name, func(t *testing.T) {
			gate()
			if got := list.CopyPacketsA(ntype.PlayerInd(31), netlist.Kind0); len(got) != 0 {
				t.Fatalf("gated queue = % X, want empty", got)
			}
		})
	}
	runtime.KeepAlive(dr)
}
