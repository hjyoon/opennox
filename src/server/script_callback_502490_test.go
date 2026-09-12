package server

import (
	"bytes"
	"encoding/binary"
	"reflect"
	"testing"

	"github.com/opennox/noxscript/ns/asm"
)

func scriptCallbackWriteInt502490(buf *bytes.Buffer, value int32) {
	var raw [4]byte
	binary.LittleEndian.PutUint32(raw[:], uint32(value))
	buf.Write(raw[:])
}

func scriptCallbackWriteString502490(buf *bytes.Buffer, value string) {
	scriptCallbackWriteInt502490(buf, int32(len(value)))
	buf.WriteString(value)
}

func scriptCallbackFunc502490(name string, returns int, code ...uint32) ScriptFunc {
	return ScriptFunc{FuncDef: asm.FuncDef{
		Name:   name,
		Return: returns,
		Code:   code,
	}}
}

func TestScriptCallbackRaw502490PreservesGatesAndOneShotFlag(t *testing.T) {
	s := new(NoxScriptVM)
	s.Init(nil)
	s.vm.funcs = []ScriptFunc{
		scriptCallbackFunc502490("callback", 0, uint32(asm.OpReturn0)),
	}

	for _, tc := range []struct {
		name      string
		block     ScriptCallback
		wantNil   bool
		wantFlags uint32
	}{
		{
			name:      "disabled",
			block:     ScriptCallback{Flags: 0xa5a50001, Func: 0},
			wantNil:   true,
			wantFlags: 0xa5a50001,
		},
		{
			name:      "missing function",
			block:     ScriptCallback{Flags: 0xa5a50000, Func: -1},
			wantNil:   true,
			wantFlags: 0xa5a50000,
		},
		{
			name:      "one shot",
			block:     ScriptCallback{Flags: 0xa5a50002, Func: 0},
			wantFlags: 0xa5a50003,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out := uint32(0x89abcdef)
			got := s.ScriptCallbackRaw(&tc.block, nil, nil, &out)
			if (got == nil) != tc.wantNil {
				t.Fatalf("result = %p, want nil=%v", got, tc.wantNil)
			}
			if !tc.wantNil && got != &out {
				t.Fatalf("result = %p, want output pointer %p", got, &out)
			}
			if out != 0 {
				t.Fatalf("output = %#08x, want zero", out)
			}
			if tc.block.Flags != tc.wantFlags {
				t.Fatalf("flags = %#08x, want %#08x", tc.block.Flags, tc.wantFlags)
			}
		})
	}
}

func TestScriptCallbackRaw502490ReloadsLiveReturnMetadata(t *testing.T) {
	s := new(NoxScriptVM)
	s.Init(nil)
	block := ScriptCallback{Func: 0}
	const value = uint32(0x89abcdef)
	s.vm.funcs = []ScriptFunc{
		scriptCallbackFunc502490(
			"mutating callback",
			1,
			uint32(asm.OpCallBuiltin), uint32(asm.BuiltinWall),
			uint32(asm.OpPushInt), value,
			uint32(asm.OpReturn),
		),
		scriptCallbackFunc502490("live descriptor", 0, uint32(asm.OpReturn0)),
	}
	called := false
	s.CallBuiltinNative = func(asm.Builtin) error {
		called = true
		block.Func = 1
		return nil
	}

	out := uint32(0xfedcba98)
	if got := s.ScriptCallbackRaw(&block, nil, nil, &out); got != &out {
		t.Fatalf("result = %p, want output pointer %p", got, &out)
	}
	if !called {
		t.Fatal("initial callback was not executed")
	}
	if block.Func != 1 {
		t.Fatalf("live function = %d, want 1", block.Func)
	}
	if out != 0 {
		t.Fatalf("output = %#08x, want zero from live non-returning descriptor", out)
	}
	if len(s.vm.stack) != 0 {
		t.Fatalf("stack length = %d, want zero", len(s.vm.stack))
	}
}

func TestScriptCallbackRaw502490RestoresLoadedStringBoundary(t *testing.T) {
	s := new(NoxScriptVM)
	s.Init(nil)
	s.vm.strings = make([]string, 2, 4)
	copy(s.vm.strings, []string{"loaded zero", "loaded one"})
	s.vm.stringsBase = len(s.vm.strings)
	s.vm.funcs = []ScriptFunc{
		scriptCallbackFunc502490(
			"string callback",
			0,
			uint32(asm.OpCallBuiltin), uint32(asm.BuiltinWall),
			uint32(asm.OpReturn0),
		),
	}
	s.CallBuiltinNative = func(asm.Builtin) error {
		if got := s.NewString("temporary zero"); got != 2 {
			t.Fatalf("first temporary string index = %d, want 2", got)
		}
		if got := s.NewString("temporary one"); got != 3 {
			t.Fatalf("second temporary string index = %d, want 3", got)
		}
		return nil
	}

	out := uint32(0x89abcdef)
	block := ScriptCallback{Func: 0}
	s.ScriptCallbackRaw(&block, nil, nil, &out)
	if want := []string{"loaded zero", "loaded one"}; !reflect.DeepEqual(s.vm.strings, want) {
		t.Fatalf("strings = %q, want %q", s.vm.strings, want)
	}
	backing := s.vm.strings[:cap(s.vm.strings)]
	if backing[2] != "" || backing[3] != "" {
		t.Fatalf("released backing strings = %q, want empty", backing[2:])
	}
	if got := s.NewString("reused"); got != 2 {
		t.Fatalf("reused temporary string index = %d, want 2", got)
	}
}

func TestReadScript502490RecordsAndResetsLoadedStringBoundary(t *testing.T) {
	var raw bytes.Buffer
	raw.WriteString("SCRIPT03")
	raw.WriteString("STRG")
	scriptCallbackWriteInt502490(&raw, 2)
	scriptCallbackWriteString502490(&raw, "loaded zero")
	scriptCallbackWriteString502490(&raw, "loaded one")
	raw.WriteString("CODE")
	scriptCallbackWriteInt502490(&raw, 0)
	raw.WriteString("DONE")

	s := new(NoxScriptVM)
	s.Init(nil)
	if err := s.ReadScript(bytes.NewReader(raw.Bytes())); err != nil {
		t.Fatalf("ReadScript: %v", err)
	}
	if s.vm.stringsBase != 2 {
		t.Fatalf("loaded string boundary = %d, want 2", s.vm.stringsBase)
	}
	if got := s.NewString("temporary"); got != 2 {
		t.Fatalf("temporary string index = %d, want 2", got)
	}
	s.resetCallbackStrings()
	if want := []string{"loaded zero", "loaded one"}; !reflect.DeepEqual(s.vm.strings, want) {
		t.Fatalf("strings = %q, want %q", s.vm.strings, want)
	}

	s.Reset()
	if s.vm.stringsBase != 0 || len(s.vm.strings) != 0 {
		t.Fatalf("boundary/strings after reset = %d/%q, want 0/empty", s.vm.stringsBase, s.vm.strings)
	}
}

func TestScriptCallbackQueue502490KeepsOriginalCapacity(t *testing.T) {
	s := new(NoxScriptVM)
	s.Init(nil)
	blocks := make([]ScriptCallback, noxScriptCallbackQueueCap+1)
	callers := make([]Object, len(blocks))
	triggers := make([]Object, len(blocks))
	for i := range blocks {
		s.scriptPushCallback(&blocks[i], &callers[i], &triggers[i])
	}
	if got := len(s.vm.callbacks); got != noxScriptCallbackQueueCap {
		t.Fatalf("callback queue length = %d, want %d", got, noxScriptCallbackQueueCap)
	}
	for i, got := range s.vm.callbacks {
		if got.Block != &blocks[i] || got.Caller != &callers[i] || got.Trigger != &triggers[i] {
			t.Fatalf("callback %d = {%p %p %p}, want {%p %p %p}", i, got.Block, got.Caller, got.Trigger, &blocks[i], &callers[i], &triggers[i])
		}
	}
}

func TestScriptCallbackRaw502490QueuesWhileStackIsBusy(t *testing.T) {
	s := new(NoxScriptVM)
	s.Init(nil)
	s.vm.funcs = []ScriptFunc{
		scriptCallbackFunc502490(
			"must be deferred",
			0,
			uint32(asm.OpCallBuiltin), uint32(asm.BuiltinWall),
			uint32(asm.OpReturn0),
		),
	}
	executed := false
	s.CallBuiltinNative = func(asm.Builtin) error {
		executed = true
		return nil
	}
	s.vm.stack = []uint32{0x89abcdef}
	block := ScriptCallback{Flags: 0xa5a50002, Func: 0}
	caller := new(Object)
	trigger := new(Object)
	out := uint32(0xfedcba98)

	if got := s.ScriptCallbackRaw(&block, caller, trigger, &out); got != &out {
		t.Fatalf("result = %p, want output pointer %p", got, &out)
	}
	if executed {
		t.Fatal("busy callback executed instead of being queued")
	}
	if block.Flags != 0xa5a50003 {
		t.Fatalf("flags = %#08x, want one-shot bit set", block.Flags)
	}
	if out != 0 {
		t.Fatalf("output = %#08x, want zero", out)
	}
	if want := []uint32{0x89abcdef}; !reflect.DeepEqual(s.vm.stack, want) {
		t.Fatalf("stack = %#v, want %#v", s.vm.stack, want)
	}
	if len(s.vm.callbacks) != 1 {
		t.Fatalf("callback queue length = %d, want 1", len(s.vm.callbacks))
	}
	queued := s.vm.callbacks[0]
	if queued.Block != &block || queued.Caller != caller || queued.Trigger != trigger {
		t.Fatalf("queued callback = {%p %p %p}, want {%p %p %p}", queued.Block, queued.Caller, queued.Trigger, &block, caller, trigger)
	}
}

func TestScriptCallbackRaw502490DrainsFirstDeferredCallback(t *testing.T) {
	s := new(NoxScriptVM)
	s.Init(nil)
	s.vm.funcs = []ScriptFunc{
		scriptCallbackFunc502490(
			"outer",
			0,
			uint32(asm.OpCallBuiltin), uint32(asm.BuiltinWall),
			uint32(asm.OpReturn0),
		),
		scriptCallbackFunc502490(
			"deferred",
			0,
			uint32(asm.OpCallBuiltin), uint32(asm.BuiltinWall),
			uint32(asm.OpReturn0),
		),
	}
	outerCaller := new(Object)
	deferredCaller := new(Object)
	var events []string
	s.CallBuiltinNative = func(asm.Builtin) error {
		switch s.Caller() {
		case outerCaller:
			events = append(events, "outer")
		case deferredCaller:
			events = append(events, "deferred")
		default:
			t.Fatalf("unexpected caller %p", s.Caller())
		}
		return nil
	}
	outer := ScriptCallback{Func: 0}
	deferred := ScriptCallback{Func: 1}
	s.scriptPushCallback(&deferred, deferredCaller, nil)
	out := uint32(0x89abcdef)

	s.ScriptCallbackRaw(&outer, outerCaller, nil, &out)
	if want := []string{"outer", "deferred"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	if len(s.vm.callbacks) != 0 {
		t.Fatalf("callback queue length = %d, want zero", len(s.vm.callbacks))
	}
	if out != 0 {
		t.Fatalf("output = %#08x, want zero", out)
	}
}
