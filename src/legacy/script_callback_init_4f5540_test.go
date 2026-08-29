package legacy

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

func TestScriptCallbackInit4F5540ModeGateReturnsExactValue(t *testing.T) {
	for _, mode := range []int32{0, -1, 2, math.MinInt32, math.MaxInt32} {
		t.Run(fmt.Sprintf("mode_%d", mode), func(t *testing.T) {
			reads := 0
			got := scriptCallbackInit4F5540((*struct{})(nil), scriptCallbackInitDeps4F5540[*struct{}, int]{
				readOnly: func() int32 {
					reads++
					return mode
				},
				mapgenFile: func() int {
					t.Fatal("mapgen file read after mode rejection")
					return 0
				},
				makeScript: func(int, *struct{}) int32 {
					t.Fatal("script parser called after mode rejection")
					return 0
				},
				gameFlagCheck: func(uint32) int32 {
					t.Fatal("game flag read after mode rejection")
					return 0
				},
				storeFunc: func(*struct{}, int32) {
					t.Fatal("nil handler written after mode rejection")
				},
			})
			if got != mode {
				t.Fatalf("result = %d, want exact mode %d", got, mode)
			}
			if reads != 1 {
				t.Fatalf("mode reads = %d, want 1", reads)
			}
		})
	}
}

func TestScriptCallbackInit4F5540OrderIdentityAndExactFlagResult(t *testing.T) {
	type handler struct{ Func int32 }
	type file struct{ ID int }
	h := &handler{Func: 17}
	f := &file{ID: 23}
	mode := int32(1)
	reads := 0
	var events []string

	got := scriptCallbackInit4F5540(h, scriptCallbackInitDeps4F5540[*handler, *file]{
		readOnly: func() int32 {
			reads++
			events = append(events, "mode")
			result := mode
			mode = 0
			return result
		},
		mapgenFile: func() *file {
			events = append(events, "file")
			return f
		},
		makeScript: func(gotFile *file, gotHandler *handler) int32 {
			events = append(events, "parse")
			if gotFile != f || gotHandler != h {
				t.Fatalf("parser identity = (%p, %p), want (%p, %p)", gotFile, gotHandler, f, h)
			}
			return math.MinInt32
		},
		gameFlagCheck: func(flag uint32) int32 {
			events = append(events, fmt.Sprintf("flag:%#x", flag))
			return -7
		},
		storeFunc: func(*handler, int32) {
			t.Fatal("Func written for nonzero game-flag result")
		},
	})

	if got != -7 {
		t.Fatalf("result = %d, want exact game-flag result -7", got)
	}
	if reads != 1 {
		t.Fatalf("mode reads = %d, want 1", reads)
	}
	want := []string{"mode", "file", "parse", "flag:0x400000"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
	if h.Func != 17 {
		t.Fatalf("Func = %d, want unchanged 17", h.Func)
	}
}

func TestScriptCallbackInit4F5540ZeroFlagOverwritesOnlyFunc(t *testing.T) {
	type handler struct {
		Flags uint32
		Func  int32
	}
	h := &handler{Flags: 0x11, Func: 3}
	var events []string
	fileBits := uint64(0x123456789)
	fileValue := uintptr(fileBits)

	got := scriptCallbackInit4F5540(h, scriptCallbackInitDeps4F5540[*handler, uintptr]{
		readOnly: func() int32 {
			events = append(events, "mode")
			return 1
		},
		mapgenFile: func() uintptr {
			events = append(events, "file")
			return fileValue
		},
		makeScript: func(file uintptr, gotHandler *handler) int32 {
			events = append(events, fmt.Sprintf("parse:%#x", file))
			if gotHandler != h {
				t.Fatalf("parser handler = %p, want %p", gotHandler, h)
			}
			gotHandler.Flags = 0xfeedbeef
			gotHandler.Func = 99
			return math.MaxInt32
		},
		gameFlagCheck: func(flag uint32) int32 {
			events = append(events, "flag")
			if flag != scriptCallbackInitGameFlag4F5540 {
				t.Fatalf("game flag = %#x, want %#x", flag, scriptCallbackInitGameFlag4F5540)
			}
			if h.Flags != 0xfeedbeef || h.Func != 99 {
				t.Fatalf("flag check saw handler {%#x, %d}, want parser mutations", h.Flags, h.Func)
			}
			return 0
		},
		storeFunc: func(gotHandler *handler, value int32) {
			events = append(events, fmt.Sprintf("store:%d", value))
			if gotHandler != h {
				t.Fatalf("store handler = %p, want %p", gotHandler, h)
			}
			gotHandler.Func = value
		},
	})

	if got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	wantEvents := []string{"mode", "file", fmt.Sprintf("parse:%#x", fileValue), "flag", "store:-1"}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %#v, want %#v", events, wantEvents)
	}
	if h.Flags != 0xfeedbeef || h.Func != -1 {
		t.Fatalf("handler = {%#x, %d}, want {0xfeedbeef, -1}", h.Flags, h.Func)
	}
}

func TestScriptCallbackInit4F5540FaultPrefixes(t *testing.T) {
	tests := []struct {
		fault    string
		want     []string
		wantFunc int32
	}{
		{fault: "mode", want: []string{"mode"}, wantFunc: 10},
		{fault: "file", want: []string{"mode", "file"}, wantFunc: 10},
		{fault: "parse", want: []string{"mode", "file", "parse"}, wantFunc: 10},
		{fault: "flag", want: []string{"mode", "file", "parse", "flag"}, wantFunc: 20},
		{fault: "store", want: []string{"mode", "file", "parse", "flag", "store"}, wantFunc: 20},
	}
	for _, tc := range tests {
		t.Run(tc.fault, func(t *testing.T) {
			type handler struct{ Func int32 }
			h := &handler{Func: 10}
			var events []string
			step := func(name string) {
				events = append(events, name)
				if tc.fault == name {
					panic("fault:" + name)
				}
			}
			func() {
				defer func() {
					if got := recover(); got != "fault:"+tc.fault {
						t.Fatalf("panic = %#v, want %q", got, "fault:"+tc.fault)
					}
				}()
				scriptCallbackInit4F5540(h, scriptCallbackInitDeps4F5540[*handler, int]{
					readOnly: func() int32 {
						step("mode")
						return 1
					},
					mapgenFile: func() int {
						step("file")
						return 7
					},
					makeScript: func(int, *handler) int32 {
						step("parse")
						h.Func = 20
						return 0
					},
					gameFlagCheck: func(uint32) int32 {
						step("flag")
						return 0
					},
					storeFunc: func(*handler, int32) {
						step("store")
						h.Func = -1
					},
				})
			}()
			if !reflect.DeepEqual(events, tc.want) {
				t.Fatalf("events = %#v, want fault prefix %#v", events, tc.want)
			}
			if h.Func != tc.wantFunc {
				t.Fatalf("Func after fault = %d, want %d", h.Func, tc.wantFunc)
			}
		})
	}
}

func TestScriptCallbackInit4F5540DoesNotGuardNilHandler(t *testing.T) {
	type handler struct{ Func int32 }
	var h *handler
	var events []string

	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("nil Func store did not fault")
			}
		}()
		scriptCallbackInit4F5540(h, scriptCallbackInitDeps4F5540[*handler, int]{
			readOnly: func() int32 {
				events = append(events, "mode")
				return 1
			},
			mapgenFile: func() int {
				events = append(events, "file")
				return 0
			},
			makeScript: func(_ int, gotHandler *handler) int32 {
				events = append(events, "parse:nil")
				if gotHandler != nil {
					t.Fatalf("parser handler = %p, want nil", gotHandler)
				}
				return 0
			},
			gameFlagCheck: func(uint32) int32 {
				events = append(events, "flag")
				return 0
			},
			storeFunc: func(gotHandler *handler, value int32) {
				events = append(events, "store")
				gotHandler.Func = value
			},
		})
	}()
	want := []string{"mode", "file", "parse:nil", "flag", "store"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
}
