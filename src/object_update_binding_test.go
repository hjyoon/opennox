package opennox

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"

	"github.com/opennox/opennox/v1/server"
)

func TestNewServerBindsTriggerAndToggleUpdatesToGo(t *testing.T) {
	s := NewServer(nil, nil, strman.New())
	t.Cleanup(s.Close)

	for _, name := range []string{"TriggerUpdate", "ToggleUpdate"} {
		t.Run(name, func(t *testing.T) {
			callback, size, ok := server.ObjectUpdateHandler(name)
			if !ok || callback == nil || size != unsafe.Sizeof(server.ToggleUpdateData{}) {
				t.Fatalf("registration = %p/%d/%t", callback, size, ok)
			}
			for _, dispatch := range []struct {
				name string
				call func(*server.Object)
			}{
				{"game loop", func(obj *server.Object) { asObjectS(obj).CallUpdate() }},
				{"server", func(obj *server.Object) { obj.CallUpdate() }},
			} {
				t.Run(dispatch.name, func(t *testing.T) {
					update := &server.ToggleUpdateData{Flags: 0xf}
					obj := &server.Object{
						ObjClass:   object.ClassTrigger,
						Update:     callback,
						UpdateData: unsafe.Pointer(update),
					}
					dispatch.call(obj)
					if update.Flags != 0x6 {
						t.Fatalf("disabled %s flags = %#x, want 0x6", name, update.Flags)
					}
				})
			}
		})
	}
}
