package legacy

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/opennox/opennox/v1/legacy/timer"
	"os"
	"testing"
	"unsafe"
)

func TestAudioSourceNativeOriginalProducersCABI(t *testing.T) {
	raw, err := os.ReadFile("testdata/audio_producer_original_4bd8c0.json")
	if err != nil {
		t.Fatal(err)
	}
	if hash := fmt.Sprintf("%x", sha256.Sum256(raw)); hash != "5107e51e65ffca0f556d894a2e0a5ecbf6e568b5f6c6bdbfa4d0bbb394f4e1a7" {
		t.Fatalf("original producer CPU fixture changed: %s", hash)
	}
	var rows []struct {
		Input struct {
			Kind, Chain              string
			Custom, Supply, Attached bool
			Flags, Repeat, Result    *uint32
		}
		Before, Events, After, Result json.RawMessage
	}
	if err := json.Unmarshal(raw, &rows); err != nil || len(rows) != 116 {
		t.Fatalf("116 original CPU producer cases required: %v / %d", err, len(rows))
	}
	oldSvc, oldCb := audioSourceTestServices, audioNativeFixtureCallbackService
	t.Cleanup(func() { audioSourceTestServices, audioNativeFixtureCallbackService = oldSvc, oldCb })
	oldTicks, oldTimerTicks := PlatformTicks, timer.PlatformTicks
	PlatformTicks = func() uint64 { return 0x100000047 }
	timer.PlatformTicks = PlatformTicks
	t.Cleanup(func() { PlatformTicks, timer.PlatformTicks = oldTicks, oldTimerTicks })
	for owner := 0; owner < 3; owner++ {
		a := newASNArena(t)
		for index, row := range rows {
			t.Run(fmt.Sprintf("foreign%d/%s/%03d", owner, row.Input.Kind, index), func(t *testing.T) {
				a.reset()
				a.take(uintptr(owner*176 + 16))
				v := (*AudioVoiceNative)(a.take(unsafe.Sizeof(AudioVoiceNative{})))
				b := (*AudioBufferNative)(a.take(unsafe.Sizeof(AudioBufferNative{})))
				c1 := (*AudioChunkNative)(a.take(unsafe.Sizeof(AudioChunkNative{})))
				c2 := (*AudioChunkNative)(a.take(unsafe.Sizeof(AudioChunkNative{})))
				methods := (*AudioMethodsNative)(a.take(unsafe.Sizeof(AudioMethodsNative{})))
				ctx := (*AudioContextNative)(a.take(unsafe.Sizeof(AudioContextNative{})))
				data, data2 := (*byte)(a.take(0x1000)), (*byte)(a.take(0x1000))
				asnPointer(19, unsafe.Pointer(b), nil, 0, 0, 0)
				b.Data, b.Size = data, 31
				chain := row.Input.Chain
				if chain == "" {
					chain = "next"
				}
				if chain != "none" {
					asnPointer(20, unsafe.Pointer(c1), unsafe.Pointer(data), 3, 0, 0)
					asnPointer(20, unsafe.Pointer(c2), unsafe.Pointer(data2), 5, 0, 0)
					asnScalar(9, unsafe.Pointer(b), unsafe.Pointer(c1), 0)
					asnScalar(9, unsafe.Pointer(b), unsafe.Pointer(c2), 0)
					v.Chunk = c1
					if chain == "last" {
						v.Chunk = c2
					}
				}
				b.Size = 31
				v.Buffer, v.Cursor, v.Available, v.Remaining = b, (*byte)(unsafe.Add(unsafe.Pointer(data), 5)), 11, 13
				v.Flags, v.Repeat = 0xa5a5a5a5, 7
				if row.Input.Flags != nil {
					v.Flags = *row.Input.Flags
				}
				if row.Input.Repeat != nil {
					v.Repeat = int32(*row.Input.Repeat)
				}
				v.Context, v.Methods = ctx, methods
				methods.Restart, methods.Stop, methods.Play = asnCallback(7, true), asnCallback(8, true), asnCallback(9, false)
				if row.Input.Custom {
					v.OnRefill, v.OnRestart, v.OnEnd, v.OnStop = asnCallback(3, false), asnCallback(4, true), asnCallback(5, false), asnCallback(6, false)
				}
				if (row.Input.Kind == "play" || row.Input.Kind == "select") && !row.Input.Attached {
					v.Buffer = nil
				}
				role := func(p unsafe.Pointer) any {
					switch p {
					case nil:
						return nil
					case unsafe.Pointer(v):
						return "voice"
					case unsafe.Pointer(&v.Priority):
						return "priority"
					case unsafe.Pointer(b):
						return "buffer"
					case unsafe.Pointer(c1):
						return "chunk1"
					case unsafe.Pointer(c2):
						return "chunk2"
					}
					for i, base := range []*byte{data, data2} {
						if uintptr(p) >= uintptr(unsafe.Pointer(base)) && uintptr(p)-uintptr(unsafe.Pointer(base)) < 0x1000 {
							name := "data"
							if i == 1 {
								name = "data2"
							}
							return []any{name, uintptr(p) - uintptr(unsafe.Pointer(base))}
						}
					}
					t.Fatalf("unknown or truncated producer pointer %p", p)
					return nil
				}
				state := func() any {
					return map[string]any{"flags": v.Flags, "repeat": uint32(v.Repeat), "buffer": role(unsafe.Pointer(v.Buffer)), "chunk": role(unsafe.Pointer(v.Chunk)), "cursor": role(unsafe.Pointer(v.Cursor)), "available": v.Available, "remaining": v.Remaining, "buffer_data": role(unsafe.Pointer(b.Data))}
				}
				asnEqualJSON(t, "initial producer state", state(), row.Before)
				events := make([]any, 0)
				audioSourceTestServices = &audioSourceServices{Dependency: func(name string, source *AudioVoiceNative) {
					if source != v {
						t.Fatal("dependency owner pointer changed")
					}
					events = append(events, map[string]any{"service": name, "state": state()})
				}}
				audioNativeFixtureCallbackService = func(kind int, p unsafe.Pointer) int32 {
					if p != unsafe.Pointer(v) || kind < 3 || kind > 9 {
						t.Fatalf("callback kind/owner changed: %d %p", kind, p)
					}
					name := []string{"custom_refill", "custom_restart", "custom_end", "custom_stop", "method_restart", "method_stop", "method_play"}[kind-3]
					events = append(events, map[string]any{"service": name, "state": state()})
					if name == "custom_refill" && row.Input.Supply {
						v.Cursor, v.Available, v.Remaining = data2, 17, 19
					}
					if name == "custom_restart" && row.Input.Supply {
						v.Buffer = b
					}
					if row.Input.Result != nil {
						return int32(*row.Input.Result)
					}
					return 0
				}
				beforeV, beforeB := *v, *b
				var result any
				if op, ok := map[string]int{"refill": 2, "restart": 3, "end": 4, "stop": 5, "play": 6}[row.Input.Kind]; ok {
					result = uint32(asnScalar(op, unsafe.Pointer(v), nil, 0))
				} else {
					switch row.Input.Kind {
					case "set_flag":
						result = role(asnPointer(0, unsafe.Pointer(v), nil, 0, 0, 0))
					case "clear_flag":
						result = role(asnPointer(1, unsafe.Pointer(v), nil, 0, 0, 0))
					case "priority":
						result = role(asnPointer(2, unsafe.Pointer(&v.Priority), nil, 0, 0, 0))
					case "select":
						var p unsafe.Pointer
						if row.Input.Attached {
							p = unsafe.Pointer(b)
						}
						asnPointer(3, unsafe.Pointer(v), p, 0, 0, 0)
					default:
						t.Fatal("unknown original producer case")
					}
				}
				asnEqualJSON(t, "producer result", result, row.Result)
				asnEqualJSON(t, "producer callbacks and state", events, row.Events)
				asnEqualJSON(t, "producer final state", state(), row.After)
				beforeV.Flags, beforeV.Repeat, beforeV.Buffer = v.Flags, v.Repeat, v.Buffer
				beforeV.Chunk, beforeV.Cursor, beforeV.Available, beforeV.Remaining = v.Chunk, v.Cursor, v.Available, v.Remaining
				beforeB.Data = b.Data
				if row.Input.Kind == "end" {
					beforeV.Group = v.Group // delegated timer implementation has separate coverage
				}
				if row.Input.Kind == "play" {
					beforeV.Mixed, beforeV.Group = v.Mixed, v.Group
				}
				if !bytes.Equal(unsafe.Slice((*byte)(unsafe.Pointer(v)), unsafe.Sizeof(*v)), unsafe.Slice((*byte)(unsafe.Pointer(&beforeV)), unsafe.Sizeof(beforeV))) || !bytes.Equal(unsafe.Slice((*byte)(unsafe.Pointer(b)), unsafe.Sizeof(*b)), unsafe.Slice((*byte)(unsafe.Pointer(&beforeB)), unsafe.Sizeof(beforeB))) {
					t.Fatal("producer changed fields outside its recorded write set")
				}
				a.guards()
			})
		}
	}
}
