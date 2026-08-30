package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/server"
)

type playerAdjustStaminaLegacyServer4F7DB0 struct {
	Server
	srv *server.Server
}

func (s *playerAdjustStaminaLegacyServer4F7DB0) S() *server.Server {
	return s.srv
}

func TestPlayerAdjustStaminaExport4F7DB0PreservesNativePointersAndLowByte(t *testing.T) {
	calls := 0
	srv := &server.Server{
		NetSendPacketXxx: func(recipient int, packet []byte, related *server.Object, remove, sequence int) int {
			calls++
			if recipient != 0xfe || len(packet) != 2 || packet[0] != byte(netmsg.MSG_REPORT_STAMINA) || packet[1] != 6 || related != nil || remove != 1 || sequence != 1 {
				t.Fatalf("send = recipient:%d packet:%v related:%p remove:%d sequence:%d", recipient, packet, related, remove, sequence)
			}
			return -1
		},
	}
	oldGetServer := GetServer
	GetServer = func() Server { return &playerAdjustStaminaLegacyServer4F7DB0{srv: srv} }
	t.Cleanup(func() { GetServer = oldGetServer })

	player := &server.Player{PlayerInd: 0xfe}
	update := &server.PlayerUpdateData{Stamina: 5, Player: player}
	unit := &server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
	var pin runtime.Pinner
	pin.Pin(player)
	pin.Pin(update)
	pin.Pin(unit)
	defer pin.Unpin()
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, pointer := range map[string]uintptr{
			"unit":   uintptr(unsafe.Pointer(unit)),
			"update": uintptr(unsafe.Pointer(update)),
			"player": uintptr(unsafe.Pointer(player)),
		} {
			if pointer <= math.MaxUint32 {
				t.Fatalf("%s pointer = %#x, want native address above 4 GiB", name, pointer)
			}
		}
	}

	playerAdjustStaminaExportCall4F7DB0(unit, 0xff)
	if update.Stamina != 6 || calls != 1 {
		t.Fatalf("stamina=%d sends=%d, want 6/1", update.Stamina, calls)
	}
	runtime.KeepAlive(player)
	runtime.KeepAlive(update)
	runtime.KeepAlive(unit)
}
