package legacy

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
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

func TestScriptCallbackInitRuntime4F5540ReadsOriginalParserWire(t *testing.T) {
	handles.Init()
	t.Cleanup(handles.Release)

	const parsedFlags = uint32(0xa1b2c3d4)
	payload := make([]byte, 12)
	binary.LittleEndian.PutUint32(payload[0:4], 0) // empty script name
	binary.LittleEndian.PutUint32(payload[4:8], parsedFlags)
	binary.LittleEndian.PutUint32(payload[8:12], 0) // no argument records

	for _, tc := range []struct {
		name        string
		gameFlag23  bool
		wantResult  int32
		initialFunc int32
		wantFunc    int32
	}{
		{
			name:        "flag_absent_clears_function",
			gameFlag23:  false,
			wantResult:  0,
			initialFunc: 17,
			wantFunc:    -1,
		},
		{
			name:        "flag_present_preserves_function",
			gameFlag23:  true,
			wantResult:  1,
			initialFunc: 23,
			wantFunc:    23,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "script-callback.bin")
			if err := os.WriteFile(path, payload, 0o600); err != nil {
				t.Fatal(err)
			}
			cf, err := cryptfile.OpenFile(path, cryptfile.ReadOnly, -1)
			if err != nil {
				t.Fatal(err)
			}
			handle := NewFileHandle(cf.File.File)
			oldFile := cryptfile.Global()
			oldFlag23 := noxflags.HasGame(noxflags.GameFlag23)
			cryptfile.SetGlobal(cf)
			if tc.gameFlag23 {
				noxflags.SetGame(noxflags.GameFlag23)
			} else {
				noxflags.UnsetGame(noxflags.GameFlag23)
			}
			t.Cleanup(func() {
				cryptfile.SetGlobal(oldFile)
				if oldFlag23 {
					noxflags.SetGame(noxflags.GameFlag23)
				} else {
					noxflags.UnsetGame(noxflags.GameFlag23)
				}
				nox_fs_close(handle)
			})

			handler := &server.ScriptCallback{Flags: 0x11223344, Func: tc.initialFunc}
			if got := scriptCallbackInitRuntime4F5540(handler); got != tc.wantResult {
				t.Fatalf("result = %d, want %d", got, tc.wantResult)
			}
			if handler.Flags != parsedFlags || handler.Func != tc.wantFunc {
				t.Fatalf("handler = {%#08x, %d}, want {%#08x, %d}",
					handler.Flags, handler.Func, parsedFlags, tc.wantFunc)
			}
			off, err := cf.File.Seek(0, io.SeekCurrent)
			if err != nil {
				t.Fatal(err)
			}
			if off != int64(len(payload)) {
				t.Fatalf("parser file offset = %d, want %d", off, len(payload))
			}
		})
	}
}
