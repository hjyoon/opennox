package legacy

import (
	"fmt"
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func TestScriptCallbackInitNativeLayout4F5540(t *testing.T) {
	if got := unsafe.Sizeof(server.ScriptCallback{}); got != 8 {
		t.Fatalf("ScriptCallback size = %d, want 8", got)
	}
	if got := unsafe.Offsetof(server.ScriptCallback{}.Flags); got != 0 {
		t.Fatalf("ScriptCallback.Flags offset = %d, want 0", got)
	}
	if got := unsafe.Offsetof(server.ScriptCallback{}.Func); got != 4 {
		t.Fatalf("ScriptCallback.Func offset = %d, want 4", got)
	}
}

func TestScriptCallbackInitNative4F5540PreservesWideIdentities(t *testing.T) {
	handler := &server.ScriptCallback{Flags: 0x11223344, Func: 17}
	fileMarker := new(byte)
	file := unsafe.Pointer(fileMarker)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, pointer := range map[string]unsafe.Pointer{
			"handler": unsafe.Pointer(handler),
			"file":    file,
		} {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("%s pointer = %p, want address above PE32 range", name, pointer)
			}
		}
	}

	var events []string
	got := scriptCallbackInitNative4F5540(handler, scriptCallbackInitNativeDeps4F5540{
		readOnly: func() int32 {
			events = append(events, "mode")
			return 1
		},
		mapgenFile: func() unsafe.Pointer {
			events = append(events, fmt.Sprintf("file:%p", file))
			return file
		},
		makeScript: func(gotFile unsafe.Pointer, gotHandler *server.ScriptCallback) int32 {
			events = append(events, fmt.Sprintf("parse:%p:%p", gotFile, gotHandler))
			if gotFile != file || gotHandler != handler {
				t.Fatalf("parser identity = (%p, %p), want (%p, %p)", gotFile, gotHandler, file, handler)
			}
			gotHandler.Flags = 0xaabbccdd
			gotHandler.Func = 99
			return math.MinInt32
		},
		gameFlagCheck: func(flag uint32) int32 {
			events = append(events, fmt.Sprintf("flag:%#x", flag))
			return 0
		},
		storeFunc: func(gotHandler *server.ScriptCallback, value int32) {
			events = append(events, fmt.Sprintf("store:%p:%d", gotHandler, value))
			if gotHandler != handler {
				t.Fatalf("store handler = %p, want %p", gotHandler, handler)
			}
			gotHandler.Func = value
		},
	})

	if got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	want := []string{
		"mode",
		fmt.Sprintf("file:%p", file),
		fmt.Sprintf("parse:%p:%p", file, handler),
		"flag:0x400000",
		fmt.Sprintf("store:%p:-1", handler),
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
	if *handler != (server.ScriptCallback{Flags: 0xaabbccdd, Func: -1}) {
		t.Fatalf("handler = %+v, want parser flags and Func -1", *handler)
	}
	runtime.KeepAlive(fileMarker)
}
