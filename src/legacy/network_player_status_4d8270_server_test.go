package legacy

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/server"
)

func TestNetReportPlayerStatusNative4D8270(t *testing.T) {
	player := &server.Player{PlayerInd: 0xfe}
	updateData := &server.PlayerUpdateData{Player: player}
	obj := &server.Object{
		ObjClass:   object.ClassPlayer,
		ObjFlags:   object.Flags(0x89abcdef),
		UpdateData: unsafe.Pointer(updateData),
	}
	got := netReportPlayerStatusNativeWithSend4D8270(obj, func(playerInd byte, packet []byte) int {
		if playerInd != 0xfe {
			t.Fatalf("player index = %#x, want 0xfe", playerInd)
		}
		if want := []byte{102, 0xef, 0xcd, 0xab, 0x89}; !reflect.DeepEqual(packet, want) {
			t.Fatalf("packet = % x, want % x", packet, want)
		}
		return 0x12345678
	})
	if got != 0x12345678 {
		t.Fatalf("return = %#x, want send result", got)
	}
}
