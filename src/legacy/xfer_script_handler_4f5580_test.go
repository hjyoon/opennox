package legacy

import (
	"fmt"
	"reflect"
	"testing"
)

type scriptHandlerXferTestHandler4F5580 struct {
	flags uint32
	fn    int32
}

type scriptHandlerXferTestContext4F5580 struct {
	name []byte
}

type scriptHandlerXferTestWorld4F5580 struct {
	version       uint16
	mode          int32
	wireLength    uint32
	wireName      []byte
	wireFlags     uint32
	gameFlag      int32
	indexResult   int32
	callbackNames map[int32][]byte

	events       []string
	lengthArgs   []uint32
	byteValues   [][]byte
	byteNonNil   []bool
	indexArgs    []string
	contextLoads int
	callbackArgs []int32
}

func newScriptHandlerXferTestWorld4F5580() *scriptHandlerXferTestWorld4F5580 {
	return &scriptHandlerXferTestWorld4F5580{
		version:       scriptHandlerXferVersion4F5580,
		wireFlags:     0x11223344,
		indexResult:   17,
		callbackNames: make(map[int32][]byte),
	}
}

func (w *scriptHandlerXferTestWorld4F5580) event(name string) {
	w.events = append(w.events, name)
}

func (w *scriptHandlerXferTestWorld4F5580) deps() scriptHandlerXferDeps4F5580[
	*scriptHandlerXferTestHandler4F5580,
	*scriptHandlerXferTestContext4F5580,
] {
	return scriptHandlerXferDeps4F5580[
		*scriptHandlerXferTestHandler4F5580,
		*scriptHandlerXferTestContext4F5580,
	]{
		rwVersion: func(value uint16) uint16 {
			w.event(fmt.Sprintf("rw-version:%#x", value))
			return w.version
		},
		readOnly: func() int32 {
			w.event("read-only")
			return w.mode
		},
		rwNameLength: func(value uint32) uint32 {
			w.event("rw-name-length")
			w.lengthArgs = append(w.lengthArgs, value)
			if w.mode == 1 {
				return w.wireLength
			}
			return value
		},
		rwNameBytes: func(value []byte) {
			w.event(fmt.Sprintf("rw-name-bytes:%d", len(value)))
			w.byteNonNil = append(w.byteNonNil, value != nil)
			if w.mode == 1 {
				copy(value, w.wireName)
			}
			w.byteValues = append(w.byteValues, append([]byte(nil), value...))
		},
		gameFlagCheck: func(flag uint32) int32 {
			w.event(fmt.Sprintf("game-flag:%#x", flag))
			return w.gameFlag
		},
		storeContext: func(context *scriptHandlerXferTestContext4F5580, name []byte) {
			w.event("store-context")
			context.name = append(append(context.name[:0], name...), 0)
		},
		indexByName: func(name []byte) int32 {
			w.event("index-by-name")
			w.indexArgs = append(w.indexArgs, string(name))
			return w.indexResult
		},
		storeFunc: func(handler *scriptHandlerXferTestHandler4F5580, value int32) {
			w.event("store-func")
			handler.fn = value
		},
		loadFunc: func(handler *scriptHandlerXferTestHandler4F5580) int32 {
			w.event("load-func")
			return handler.fn
		},
		loadContextName: func(context *scriptHandlerXferTestContext4F5580) []byte {
			w.event("load-context-name")
			w.contextLoads++
			return context.name
		},
		callbackName: func(function int32) []byte {
			w.event("callback-name")
			w.callbackArgs = append(w.callbackArgs, function)
			name, ok := w.callbackNames[function]
			if !ok {
				panic("invalid callback index")
			}
			return name
		},
		rwFlags: func(handler *scriptHandlerXferTestHandler4F5580) {
			w.event("rw-flags")
			if w.mode == 1 {
				handler.flags = w.wireFlags
			} else {
				_ = handler.flags
			}
		},
	}
}

func TestScriptHandlerXfer4F5580SignedVersionGate(t *testing.T) {
	for _, tc := range []struct {
		version uint16
		want    int32
	}{
		{version: 0, want: 1},
		{version: 1, want: 1},
		{version: 2, want: 0},
		{version: 0x7fff, want: 0},
		{version: 0x8000, want: 1},
		{version: 0xffff, want: 1},
	} {
		t.Run(fmt.Sprintf("%#x", tc.version), func(t *testing.T) {
			w := newScriptHandlerXferTestWorld4F5580()
			w.version = tc.version
			w.mode = 2
			w.gameFlag = 1
			handler := &scriptHandlerXferTestHandler4F5580{}

			if got := scriptHandlerXfer4F5580(handler, (*scriptHandlerXferTestContext4F5580)(nil), w.deps()); got != tc.want {
				t.Fatalf("result = %d, want %d", got, tc.want)
			}
			if tc.want == 0 {
				wantEvents := []string{"rw-version:0x1"}
				if !reflect.DeepEqual(w.events, wantEvents) {
					t.Fatalf("events = %v, want rejection prefix %v", w.events, wantEvents)
				}
			} else if len(w.events) < 2 || w.events[1] != "read-only" {
				t.Fatalf("accepted-version events = %v, want read-only after version", w.events)
			}
		})
	}
}

func TestScriptHandlerXfer4F5580ReadEmptyNameStillTransfersZeroBytesAndFlags(t *testing.T) {
	w := newScriptHandlerXferTestWorld4F5580()
	w.mode = 1
	handler := &scriptHandlerXferTestHandler4F5580{flags: 7, fn: 9}
	context := &scriptHandlerXferTestContext4F5580{name: []byte("unchanged")}

	if got := scriptHandlerXfer4F5580(handler, context, w.deps()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	wantEvents := []string{
		"rw-version:0x1", "read-only", "rw-name-length", "rw-name-bytes:0", "rw-flags",
	}
	if !reflect.DeepEqual(w.events, wantEvents) {
		t.Fatalf("events = %v, want %v", w.events, wantEvents)
	}
	if len(w.byteValues) != 1 || len(w.byteValues[0]) != 0 || !reflect.DeepEqual(w.byteNonNil, []bool{true}) {
		t.Fatalf("zero-byte transfers/non-nil = %#v/%v, want one non-nil empty transfer", w.byteValues, w.byteNonNil)
	}
	if string(context.name) != "unchanged" || handler.fn != 9 || handler.flags != w.wireFlags {
		t.Fatalf("state = context %q handler {%#x,%d}, want unchanged context/Func and transferred flags",
			context.name, handler.flags, handler.fn)
	}
}

func TestScriptHandlerXfer4F5580ReadNameBranches(t *testing.T) {
	for _, tc := range []struct {
		name     string
		flag     int32
		wantFunc int32
		wantCtx  []byte
		want     []string
	}{
		{
			name: "context", flag: -7, wantFunc: 4, wantCtx: []byte{'O', 'n', 'U', 's', 'e', 0},
			want: []string{"rw-version:0x1", "read-only", "rw-name-length", "rw-name-bytes:5", "game-flag:0x600000", "store-context", "rw-flags"},
		},
		{
			name: "callback", flag: 0, wantFunc: 81, wantCtx: []byte("old"),
			want: []string{"rw-version:0x1", "read-only", "rw-name-length", "rw-name-bytes:5", "game-flag:0x600000", "index-by-name", "store-func", "rw-flags"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := newScriptHandlerXferTestWorld4F5580()
			w.mode = 1
			w.wireLength = 5
			w.wireName = []byte("OnUse")
			w.gameFlag = tc.flag
			w.indexResult = 81
			handler := &scriptHandlerXferTestHandler4F5580{fn: 4}
			context := &scriptHandlerXferTestContext4F5580{name: []byte("old")}

			if got := scriptHandlerXfer4F5580(handler, context, w.deps()); got != 1 {
				t.Fatalf("result = %d, want 1", got)
			}
			if !reflect.DeepEqual(w.events, tc.want) {
				t.Fatalf("events = %v, want %v", w.events, tc.want)
			}
			if handler.fn != tc.wantFunc || !reflect.DeepEqual(context.name, tc.wantCtx) {
				t.Fatalf("state = Func %d/context %q, want %d/%q", handler.fn, context.name, tc.wantFunc, tc.wantCtx)
			}
			if tc.flag == 0 && !reflect.DeepEqual(w.indexArgs, []string{"OnUse"}) {
				t.Fatalf("index arguments = %v, want OnUse", w.indexArgs)
			}
		})
	}
}

func TestScriptHandlerXfer4F5580ReadLengthBoundaryAndFailurePrefix(t *testing.T) {
	for _, tc := range []struct {
		length uint32
		want   int32
	}{
		{length: 1023, want: 1},
		{length: 1024, want: 0},
		{length: ^uint32(0), want: 0},
	} {
		t.Run(fmt.Sprintf("length_%d", tc.length), func(t *testing.T) {
			w := newScriptHandlerXferTestWorld4F5580()
			w.mode = 1
			w.wireLength = tc.length
			if tc.length < scriptHandlerXferMaxName4F5580 {
				w.wireName = make([]byte, tc.length)
				w.gameFlag = 0
			}
			handler := &scriptHandlerXferTestHandler4F5580{flags: 5, fn: 6}

			if got := scriptHandlerXfer4F5580(handler, (*scriptHandlerXferTestContext4F5580)(nil), w.deps()); got != tc.want {
				t.Fatalf("result = %d, want %d", got, tc.want)
			}
			if tc.want == 0 {
				wantEvents := []string{"rw-version:0x1", "read-only", "rw-name-length"}
				if !reflect.DeepEqual(w.events, wantEvents) || handler.flags != 5 || handler.fn != 6 {
					t.Fatalf("events/state = %v/{%#x,%d}, want failure prefix and no mutation", w.events, handler.flags, handler.fn)
				}
			} else if len(w.byteValues) != 1 || len(w.byteValues[0]) != 1023 {
				t.Fatalf("accepted byte transfers = %#v, want one 1023-byte transfer", w.byteValues)
			}
		})
	}
}

func TestScriptHandlerXfer4F5580WriteContextBranches(t *testing.T) {
	for _, tc := range []struct {
		name       string
		context    *scriptHandlerXferTestContext4F5580
		wantLength uint32
		wantBytes  int
	}{
		{name: "present", context: &scriptHandlerXferTestContext4F5580{name: []byte("OnOpen")}, wantLength: 6, wantBytes: 1},
		{name: "present_empty", context: &scriptHandlerXferTestContext4F5580{name: []byte{}}, wantLength: 0, wantBytes: 1},
		{name: "null", context: nil, wantLength: 0, wantBytes: 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := newScriptHandlerXferTestWorld4F5580()
			w.mode = -3
			w.gameFlag = 9
			handler := &scriptHandlerXferTestHandler4F5580{flags: 0xaabbccdd, fn: 44}

			if got := scriptHandlerXfer4F5580(handler, tc.context, w.deps()); got != 1 {
				t.Fatalf("result = %d, want 1", got)
			}
			if !reflect.DeepEqual(w.lengthArgs, []uint32{tc.wantLength}) || len(w.byteValues) != tc.wantBytes {
				t.Fatalf("length/byte calls = %v/%#v, want [%d]/%d", w.lengthArgs, w.byteValues, tc.wantLength, tc.wantBytes)
			}
			if tc.context == nil && w.contextLoads != 0 {
				t.Fatalf("null context loads = %d, want 0", w.contextLoads)
			}
			if handler.flags != 0xaabbccdd || handler.fn != 44 {
				t.Fatalf("write handler = {%#x,%d}, want unchanged", handler.flags, handler.fn)
			}
		})
	}
}

func TestScriptHandlerXfer4F5580WriteCallbackBranches(t *testing.T) {
	for _, tc := range []struct {
		name       string
		function   int32
		callback   []byte
		wantLength uint32
		wantBytes  int
	}{
		{name: "disabled", function: -1, wantLength: 0, wantBytes: 0},
		{name: "named", function: 12, callback: []byte("OnCollide"), wantLength: 9, wantBytes: 1},
		{name: "empty_name", function: 13, callback: []byte{}, wantLength: 0, wantBytes: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := newScriptHandlerXferTestWorld4F5580()
			w.mode = 0
			w.gameFlag = 0
			if tc.function != -1 {
				w.callbackNames[tc.function] = tc.callback
			}
			handler := &scriptHandlerXferTestHandler4F5580{flags: 8, fn: tc.function}

			if got := scriptHandlerXfer4F5580(handler, (*scriptHandlerXferTestContext4F5580)(nil), w.deps()); got != 1 {
				t.Fatalf("result = %d, want 1", got)
			}
			if !reflect.DeepEqual(w.lengthArgs, []uint32{tc.wantLength}) || len(w.byteValues) != tc.wantBytes {
				t.Fatalf("length/byte calls = %v/%#v, want [%d]/%d", w.lengthArgs, w.byteValues, tc.wantLength, tc.wantBytes)
			}
			if tc.function == -1 && len(w.callbackArgs) != 0 {
				t.Fatalf("disabled callback lookups = %v, want none", w.callbackArgs)
			}
		})
	}
}

func TestScriptHandlerXfer4F5580FaultPrefixesAndNoRollback(t *testing.T) {
	t.Run("invalid callback faults before name length", func(t *testing.T) {
		w := newScriptHandlerXferTestWorld4F5580()
		w.mode = 0
		w.gameFlag = 0
		handler := &scriptHandlerXferTestHandler4F5580{flags: 3, fn: 99}

		func() {
			defer func() {
				if got := recover(); got != "invalid callback index" {
					t.Fatalf("panic = %#v, want invalid callback index", got)
				}
			}()
			scriptHandlerXfer4F5580(handler, (*scriptHandlerXferTestContext4F5580)(nil), w.deps())
		}()
		wantEvents := []string{"rw-version:0x1", "read-only", "game-flag:0x600000", "load-func", "callback-name"}
		if !reflect.DeepEqual(w.events, wantEvents) || len(w.lengthArgs) != 0 {
			t.Fatalf("events/lengths = %v/%v, want fault before length transfer", w.events, w.lengthArgs)
		}
	})

	t.Run("context mutation survives final flags fault", func(t *testing.T) {
		w := newScriptHandlerXferTestWorld4F5580()
		w.mode = 1
		w.wireLength = 4
		w.wireName = []byte("Open")
		w.gameFlag = 1
		context := &scriptHandlerXferTestContext4F5580{name: []byte("old")}
		deps := w.deps()
		deps.rwFlags = func(*scriptHandlerXferTestHandler4F5580) {
			w.event("rw-flags")
			panic("flags fault")
		}

		func() {
			defer func() {
				if got := recover(); got != "flags fault" {
					t.Fatalf("panic = %#v, want flags fault", got)
				}
			}()
			scriptHandlerXfer4F5580(&scriptHandlerXferTestHandler4F5580{}, context, deps)
		}()
		if !reflect.DeepEqual(context.name, []byte{'O', 'p', 'e', 'n', 0}) {
			t.Fatalf("context after fault = %q, want committed name and terminator", context.name)
		}
	})

	t.Run("null read context faults before flags", func(t *testing.T) {
		w := newScriptHandlerXferTestWorld4F5580()
		w.mode = 1
		w.wireLength = 1
		w.wireName = []byte("X")
		w.gameFlag = 1
		func() {
			defer func() {
				if recover() == nil {
					t.Fatal("null context store did not fault")
				}
			}()
			scriptHandlerXfer4F5580(&scriptHandlerXferTestHandler4F5580{}, (*scriptHandlerXferTestContext4F5580)(nil), w.deps())
		}()
		if got := w.events[len(w.events)-1]; got != "store-context" {
			t.Fatalf("last event = %q, want store-context fault prefix", got)
		}
	})

	t.Run("null handler write is deferred through zero name", func(t *testing.T) {
		w := newScriptHandlerXferTestWorld4F5580()
		w.mode = 2
		w.gameFlag = 1
		func() {
			defer func() {
				if recover() == nil {
					t.Fatal("null flags transfer did not fault")
				}
			}()
			scriptHandlerXfer4F5580((*scriptHandlerXferTestHandler4F5580)(nil), (*scriptHandlerXferTestContext4F5580)(nil), w.deps())
		}()
		wantEvents := []string{"rw-version:0x1", "read-only", "game-flag:0x600000", "rw-name-length", "rw-flags"}
		if !reflect.DeepEqual(w.events, wantEvents) {
			t.Fatalf("events = %v, want deferred null-handler prefix %v", w.events, wantEvents)
		}
	})
}
