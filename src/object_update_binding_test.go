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

func TestNewServerBindsNativeWidthWorldUpdatesToGo(t *testing.T) {
	s := NewServer(nil, nil, strman.New())
	t.Cleanup(s.Close)
	s.SetFrame(31)
	s.SetTickRate(30)

	tests := []struct {
		name string
		size uintptr
		make func(unsafe.Pointer) (*server.Object, func(*testing.T))
	}{
		{
			name: "ElevatorUpdate",
			size: unsafe.Sizeof(server.ElevatorUpdateData{}),
			make: func(callback unsafe.Pointer) (*server.Object, func(*testing.T)) {
				data := new(server.ElevatorUpdateData)
				obj := &server.Object{
					ObjClass:   object.ClassElevator,
					ObjFlags:   object.FlagEnabled,
					Update:     callback,
					UpdateData: unsafe.Pointer(data),
				}
				return obj, func(t *testing.T) {
					if data.Field_3 != 3 {
						t.Fatalf("elevator state = %d, want 3", data.Field_3)
					}
				}
			},
		},
		{
			name: "ElevatorShaftUpdate",
			size: unsafe.Sizeof(server.ElevatorShaftUpdateData{}),
			make: func(callback unsafe.Pointer) (*server.Object, func(*testing.T)) {
				data := &server.ElevatorShaftUpdateData{Field_3: 7}
				obj := &server.Object{
					ObjClass:   object.ClassElevatorShaft,
					Update:     callback,
					UpdateData: unsafe.Pointer(data),
				}
				return obj, func(t *testing.T) {
					if data.Field_3 != 7 {
						t.Fatalf("unlinked shaft state = %d, want 7", data.Field_3)
					}
				}
			},
		},
		{
			name: "SwitchUpdate",
			make: func(callback unsafe.Pointer) (*server.Object, func(*testing.T)) {
				obj := &server.Object{ObjFlags: object.FlagActive, Update: callback}
				return obj, func(t *testing.T) {
					if obj.Field33 != 1 || obj.Field38 != ^uint32(0) || !obj.ObjFlags.Has(object.FlagNoCollide) {
						t.Fatalf("switch state = field33:%d field38:%#x flags:%#x", obj.Field33, obj.Field38, obj.ObjFlags)
					}
				}
			},
		},
		{
			name: "TrapDoorUpdate",
			make: func(callback unsafe.Pointer) (*server.Object, func(*testing.T)) {
				data := new(server.TrapDoorCollideData)
				obj := &server.Object{
					ObjClass:    object.ClassImmobile,
					ObjFlags:    object.FlagActive | object.FlagEnabled,
					Field5:      0x102,
					CollideData: unsafe.Pointer(data),
					Update:      callback,
				}
				return obj, func(t *testing.T) {
					if obj.Field5 != 0x108 {
						t.Fatalf("trap-door xstatus = %#x, want 0x108", obj.Field5)
					}
				}
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			callback, size, ok := server.ObjectUpdateHandler(tc.name)
			if !ok || callback == nil || size != tc.size {
				t.Fatalf("registration = %p/%d/%t, want non-nil/%d/true", callback, size, ok, tc.size)
			}
			for _, dispatch := range []struct {
				name string
				call func(*server.Object)
			}{
				{"outer object", func(obj *server.Object) { asObjectS(obj).CallUpdate() }},
				{"server object", func(obj *server.Object) { obj.CallUpdate() }},
			} {
				t.Run(dispatch.name, func(t *testing.T) {
					obj, verify := tc.make(callback)
					dispatch.call(obj)
					verify(t)
				})
			}
		})
	}
}
