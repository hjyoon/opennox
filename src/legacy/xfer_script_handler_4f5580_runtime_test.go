package legacy

import (
	"bytes"
	"encoding/binary"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestScriptHandlerXferNativeLayout4F5580(t *testing.T) {
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

func TestScriptHandlerXferExport4F5580PreservesNativePointersAndResult(t *testing.T) {
	handler := new(server.ScriptCallback)
	context := append([]byte("native-context"), 0)
	contextPtr := unsafe.Pointer(&context[0])
	var pin runtime.Pinner
	pin.Pin(handler)
	pin.Pin(&context[0])
	defer pin.Unpin()

	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, pointer := range map[string]unsafe.Pointer{
			"handler": unsafe.Pointer(handler),
			"context": contextPtr,
		} {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("%s pointer = %p, want address above PE32 range", name, pointer)
			}
		}
	}

	old := scriptHandlerXferCall4F5580
	t.Cleanup(func() { scriptHandlerXferCall4F5580 = old })
	calls := 0
	scriptHandlerXferCall4F5580 = func(
		gotHandler *server.ScriptCallback,
		gotContext unsafe.Pointer,
	) int32 {
		calls++
		if gotHandler != handler || gotContext != contextPtr {
			t.Fatalf("pointers = %p/%p, want %p/%p",
				gotHandler, gotContext, handler, contextPtr)
		}
		return math.MinInt32
	}

	if got := scriptHandlerXferExportCall4F5580(handler, contextPtr); got != math.MinInt32 {
		t.Fatalf("result = %d, want %d", got, int32(math.MinInt32))
	}
	if calls != 1 {
		t.Fatalf("calls = %d, want 1", calls)
	}
	runtime.KeepAlive(handler)
	runtime.KeepAlive(context)
}

func setScriptHandlerXferRuntimeGlobals4F5580(t *testing.T, cf *cryptfile.CryptFile) {
	t.Helper()
	oldFile := cryptfile.Global()
	oldFlags := noxflags.GetGame()
	cryptfile.SetGlobal(cf)
	noxflags.ResetGame()
	noxflags.SetGame(noxflags.GameFlag22)
	t.Cleanup(func() {
		cryptfile.SetGlobal(oldFile)
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
	})
}

func TestScriptHandlerXferRuntimeWrite4F5580PreservesWire(t *testing.T) {
	context, freeContext := alloc.CString("wide-context")
	defer freeContext()
	handler, freeHandler := alloc.New(server.ScriptCallback{})
	defer freeHandler()
	handler.Flags = 0xa1b2c3d4
	handler.Func = 0x1020304

	path := filepath.Join(t.TempDir(), "script-handler-write.bin")
	cf, err := cryptfile.OpenFile(path, cryptfile.WriteOnly, -1)
	if err != nil {
		t.Fatal(err)
	}
	setScriptHandlerXferRuntimeGlobals4F5580(t, cf)

	if got := scriptHandlerXferRuntime4F5580(handler, unsafe.Pointer(context)); got != 1 {
		_ = cf.Close()
		t.Fatalf("result = %d, want 1", got)
	}
	if err := cf.Close(); err != nil {
		t.Fatal(err)
	}

	var want bytes.Buffer
	_ = binary.Write(&want, binary.LittleEndian, uint16(scriptHandlerXferVersion4F5580))
	_ = binary.Write(&want, binary.LittleEndian, uint32(len("wide-context")))
	want.WriteString("wide-context")
	_ = binary.Write(&want, binary.LittleEndian, handler.Flags)
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want.Bytes()) {
		t.Fatalf("wire payload = %x, want %x", got, want.Bytes())
	}
}

func TestScriptHandlerXferRuntimeRead4F5580RestoresContextAndFlags(t *testing.T) {
	const (
		name  = "restored-context"
		flags = uint32(0x55667788)
	)
	var payload bytes.Buffer
	_ = binary.Write(&payload, binary.LittleEndian, uint16(scriptHandlerXferVersion4F5580))
	_ = binary.Write(&payload, binary.LittleEndian, uint32(len(name)))
	payload.WriteString(name)
	_ = binary.Write(&payload, binary.LittleEndian, flags)

	path := filepath.Join(t.TempDir(), "script-handler-read.bin")
	if err := os.WriteFile(path, payload.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}
	cf, err := cryptfile.OpenFile(path, cryptfile.ReadOnly, -1)
	if err != nil {
		t.Fatal(err)
	}
	setScriptHandlerXferRuntimeGlobals4F5580(t, cf)
	context, freeContext := alloc.Make([]byte{}, 64)
	defer freeContext()
	handler, freeHandler := alloc.New(server.ScriptCallback{})
	defer freeHandler()
	handler.Flags = 0xffffffff
	handler.Func = 17

	if got := scriptHandlerXferRuntime4F5580(handler, unsafe.Pointer(&context[0])); got != 1 {
		_ = cf.Close()
		t.Fatalf("result = %d, want 1", got)
	}
	if err := cf.Close(); err != nil {
		t.Fatal(err)
	}
	if got := alloc.GoString(&context[0]); got != name {
		t.Fatalf("context = %q, want %q", got, name)
	}
	if handler.Flags != flags || handler.Func != 17 {
		t.Fatalf("handler = {%#x, %d}, want {%#x, 17}",
			handler.Flags, handler.Func, flags)
	}
}
