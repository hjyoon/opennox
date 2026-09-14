package opennox

import (
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func TestExitCollideData4DB600NativeObjectLayout(t *testing.T) {
	if got := exitCollideData4DB600(nil); got != nil {
		t.Fatalf("nil exit object: got %p", got)
	}
	exit := &server.Object{}
	if got := exitCollideData4DB600(unsafe.Pointer(exit)); got != nil {
		t.Fatalf("exit object without collide data: got %p", got)
	}
	data := &server.ExitCollideData{DestinationX: 123, DestinationY: 456}
	copy(data.MapName[:], "Wiz01b.map")
	exit.CollideData = unsafe.Pointer(data)
	if got := exitCollideData4DB600(unsafe.Pointer(exit)); got != data {
		t.Fatalf("exit collide data: got %p, want %p", got, data)
	}
}
