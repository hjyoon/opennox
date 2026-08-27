package opennox

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/client"
)

func TestDecodeObjectOutOfSightState48EA70PacketWidth(t *testing.T) {
	state, ok := decodeObjectOutOfSightState48EA70([]byte{0x32, 0x34, 0x92, 0xaa, 0xbb, 0xcc})
	if !ok || state.Code != 0x9234 {
		t.Fatalf("decoded state = %+v, ok=%t", state, ok)
	}
	if _, ok := decodeObjectOutOfSightState48EA70([]byte{0x32, 0x34}); ok {
		t.Fatal("short object-out-of-sight packet was accepted")
	}
}

func TestFindObjectOutOfSightDrawable48EA70SelectsCodeNamespace(t *testing.T) {
	dynamic := &client.Drawable{NetCode32: 0x1234, ObjClass: object.ClassMonster}
	static := &client.Drawable{NetCode32: 0x1234, ObjClass: 0x20400000}
	dynamic.NextPtr = static

	if got := findObjectOutOfSightDrawable48EA70(dynamic, 0x1234); got != dynamic {
		t.Fatalf("dynamic lookup = %p, want %p", got, dynamic)
	}
	if got := findObjectOutOfSightDrawable48EA70(dynamic, 0x9234); got != static {
		t.Fatalf("static lookup = %p, want %p", got, static)
	}
	if got := findObjectOutOfSightDrawable48EA70(dynamic, 0x9235); got != nil {
		t.Fatalf("missing lookup = %p, want nil", got)
	}
}

func TestHandleObjectOutOfSightNative48EA70OrderAndAction(t *testing.T) {
	dr := &client.Drawable{NetCode32: 0x1234, ObjClass: 0x20400000, ObjFlags: object.FlagActive | object.FlagEnabled}
	cli := &client.Client{}
	var calls []string
	hooks := objectOutOfSightHooks48EA70{
		connected: func() bool {
			calls = append(calls, "connected")
			return true
		},
		firstDrawable: func() *client.Drawable {
			calls = append(calls, "first")
			return dr
		},
		localDrawable: func() *client.Drawable {
			calls = append(calls, "local")
			return nil
		},
		deactivate: func(got *client.Drawable) {
			calls = append(calls, "deactivate")
			cli.Nox_xxx_cliDestroyObj_45A9A0(got)
		},
		deleteDrawable: func(*client.Drawable) {
			calls = append(calls, "delete")
		},
	}
	if got := handleObjectOutOfSightNative48EA70([]byte{0x32, 0x34, 0x92, 0x31, 0, 0}, hooks); got != 3 {
		t.Fatalf("consumed bytes = %d, want 3", got)
	}
	if want := []string{"connected", "first", "local", "deactivate"}; !reflect.DeepEqual(calls, want) {
		t.Fatalf("callback order = %v, want %v", calls, want)
	}
	if dr.ObjFlags.Has(object.FlagActive) || !dr.ObjFlags.Has(object.FlagEnabled) {
		t.Fatalf("deactivated flags = %#x", dr.ObjFlags)
	}
}

func TestHandleObjectOutOfSightNative48EA70ProtectionOrder(t *testing.T) {
	var drawData [4]uint32
	drawData[3] = 1
	animate := unsafe.Pointer(&drawData)
	dr := &client.Drawable{
		NetCode32:   7,
		ObjClass:    object.ClassMonster,
		DrawFuncPtr: animate,
		DrawData:    unsafe.Pointer(&drawData[0]),
	}
	var calls []string
	hooks := objectOutOfSightHooks48EA70{
		connected:     func() bool { calls = append(calls, "connected"); return true },
		firstDrawable: func() *client.Drawable { calls = append(calls, "first"); return dr },
		localDrawable: func() *client.Drawable { calls = append(calls, "local"); return dr },
		deactivate:    func(*client.Drawable) { calls = append(calls, "deactivate") },
		deleteDrawable: func(*client.Drawable) {
			calls = append(calls, "delete")
		},
		animateDrawFunc: animate,
	}
	if dr.DrawFuncPtr != hooks.animateDrawFunc || *(*uint32)(unsafe.Add(dr.DrawData, 12)) != 1 {
		t.Fatalf("one-shot fixture is invalid: draw=%p animate=%p state=%d", dr.DrawFuncPtr, hooks.animateDrawFunc, *(*uint32)(unsafe.Add(dr.DrawData, 12)))
	}
	if got := handleObjectOutOfSightNative48EA70([]byte{0x32, 7, 0}, hooks); got != 3 {
		t.Fatalf("consumed bytes = %d, want 3", got)
	}
	if want := []string{"connected", "first"}; !reflect.DeepEqual(calls, want) {
		t.Fatalf("one-shot protection order = %v, want %v", calls, want)
	}

	dr.DrawFuncPtr = nil
	calls = nil
	if got := handleObjectOutOfSightNative48EA70([]byte{0x32, 7, 0}, hooks); got != 3 {
		t.Fatalf("consumed bytes = %d, want 3", got)
	}
	if want := []string{"connected", "first", "local"}; !reflect.DeepEqual(calls, want) {
		t.Fatalf("local-player protection order = %v, want %v", calls, want)
	}
}

func TestHandleObjectInShadowsNative48EA70Fade(t *testing.T) {
	dr := &client.Drawable{NetCode32: 0x1234, ObjClass: object.ClassMonster}
	var calls []string
	hooks := objectInShadowsHooks48EA70{
		connected:     func() bool { calls = append(calls, "connected"); return true },
		firstDrawable: func() *client.Drawable { calls = append(calls, "first"); return dr },
		localDrawable: func() *client.Drawable { calls = append(calls, "local"); return nil },
		fadeObjects:   func() bool { calls = append(calls, "fade-enabled"); return true },
		tickRate:      func() uint32 { calls = append(calls, "tick-rate"); return 30 },
		transparentFade: func(got *client.Drawable, ticks int) {
			calls = append(calls, "transparent-fade")
			if got != dr || ticks != 30 {
				t.Fatalf("transparent fade = (%p, %d), want (%p, 30)", got, ticks, dr)
			}
		},
		deactivate:     func(*client.Drawable) { calls = append(calls, "deactivate") },
		deleteDrawable: func(*client.Drawable) { calls = append(calls, "delete") },
	}
	if got := handleObjectInShadowsNative48EA70([]byte{0x33, 0x34, 0x12}, hooks); got != 3 {
		t.Fatalf("consumed bytes = %d, want 3", got)
	}
	if dr.Field_120 != 1 || dr.Field_121 != 1 || dr.Field_122 != 1 {
		t.Fatalf("visibility counters = (%d, %d, %d), want all one", dr.Field_120, dr.Field_121, dr.Field_122)
	}
	if want := []string{"connected", "first", "fade-enabled", "local", "tick-rate", "transparent-fade"}; !reflect.DeepEqual(calls, want) {
		t.Fatalf("callback order = %v, want %v", calls, want)
	}
}

func TestHandleObjectInShadowsNative48EA70ImmediateRemoval(t *testing.T) {
	dynamic := &client.Drawable{NetCode32: 0x1234, ObjClass: object.ClassMonster}
	static := &client.Drawable{NetCode32: 0x1234, ObjClass: 0x20400000}
	dynamic.NextPtr = static
	var calls []string
	hooks := objectInShadowsHooks48EA70{
		connected:     func() bool { return true },
		firstDrawable: func() *client.Drawable { return dynamic },
		localDrawable: func() *client.Drawable { return nil },
		fadeObjects:   func() bool { return false },
		deactivate: func(got *client.Drawable) {
			calls = append(calls, "deactivate")
			if got != static {
				t.Fatalf("deactivated drawable = %p, want %p", got, static)
			}
		},
		deleteDrawable: func(got *client.Drawable) {
			calls = append(calls, "delete")
			if got != dynamic {
				t.Fatalf("deleted drawable = %p, want %p", got, dynamic)
			}
		},
	}
	if got := handleObjectInShadowsNative48EA70([]byte{0x33, 0x34, 0x12}, hooks); got != 3 {
		t.Fatalf("dynamic consumed bytes = %d, want 3", got)
	}
	if got := handleObjectInShadowsNative48EA70([]byte{0x33, 0x34, 0x92}, hooks); got != 3 {
		t.Fatalf("static consumed bytes = %d, want 3", got)
	}
	if want := []string{"delete", "deactivate"}; !reflect.DeepEqual(calls, want) {
		t.Fatalf("removal calls = %v, want %v", calls, want)
	}
}

func TestHandleObjectInShadowsNative48EA70Guards(t *testing.T) {
	if got := handleObjectInShadowsNative48EA70([]byte{0x33, 1}, objectInShadowsHooks48EA70{}); got != -1 {
		t.Fatalf("short packet consumed bytes = %d, want -1", got)
	}

	var drawData [4]uint32
	drawData[3] = 1
	animate := unsafe.Pointer(&drawData[0])
	dr := &client.Drawable{
		NetCode32:   7,
		ObjClass:    object.ClassMonster,
		DrawFuncPtr: animate,
		DrawData:    unsafe.Pointer(&drawData[0]),
	}
	removed := false
	hooks := objectInShadowsHooks48EA70{
		connected:       func() bool { return true },
		firstDrawable:   func() *client.Drawable { return dr },
		localDrawable:   func() *client.Drawable { return nil },
		fadeObjects:     func() bool { return false },
		deactivate:      func(*client.Drawable) { removed = true },
		deleteDrawable:  func(*client.Drawable) { removed = true },
		animateDrawFunc: animate,
	}
	if got := handleObjectInShadowsNative48EA70([]byte{0x33, 7, 0}, hooks); got != 3 {
		t.Fatalf("one-shot consumed bytes = %d, want 3", got)
	}
	if removed {
		t.Fatal("one-shot drawable was removed")
	}
}
