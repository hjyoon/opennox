package legacy

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/client/audio/ail"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestAudioSourceNativeCLayouts(t *testing.T) {
	for kind, value := range []any{AudioVoiceNative{}, AudioContextNative{}, AudioBufferNative{}, AudioChunkNative{}, AudioDeviceNative{}, AudioClassNative{}, AudioMethodsNative{}, audioGlobalNative{}, AudioSample{}, audioCacheNative{}, audioCachedSampleNative{}} {
		typ := reflect.TypeOf(value)
		if got := asnLayout(kind, 0); got != typ.Size() {
			t.Fatalf("%s C size %d != Go %d", typ, got, typ.Size())
		}
		for i := 0; i < typ.NumField(); i++ {
			if got := asnLayout(kind, i+1); got != typ.Field(i).Offset {
				t.Fatalf("%s.%s C offset %d != Go %d", typ, typ.Field(i).Name, got, typ.Field(i).Offset)
			}
		}
	}
}

func TestAudioSourceNativeDriverCABI(t *testing.T) {
	a := newASNArena(t)
	h := ail.Driver(uintptr(a.take(16)))
	old := Sub_43F130
	t.Cleanup(func() { Sub_43F130 = old })
	for _, handle := range []ail.Driver{0, h} {
		Sub_43F130 = func() ail.Driver { return handle }
		if got := asnDriver(); got != uintptr(handle) {
			t.Fatalf("public driver getter truncated: %x != %x", got, handle)
		}
	}
}

type asnArena struct {
	t     *testing.T
	data  []byte
	pos   int
	spans [][2]int
}

func newASNArena(t *testing.T) *asnArena {
	t.Helper()
	b, free := alloc.MakeHigh([]byte{}, 1<<20)
	t.Cleanup(free)
	a := &asnArena{t: t, data: b}
	a.reset()
	return a
}
func (a *asnArena) reset() {
	for i := range a.data {
		a.data[i] = 0xa7
	}
	a.pos, a.spans = 0, nil
}
func (a *asnArena) take(n uintptr) unsafe.Pointer {
	a.t.Helper()
	start := (a.pos+15)&^15 + 32
	end := start + int(n)
	if end+32 > len(a.data) {
		a.t.Fatal("fixture arena exhausted")
	}
	a.pos = end + 32
	a.spans = append(a.spans, [2]int{start, end})
	clear(a.data[start:end])
	p := unsafe.Pointer(&a.data[start])
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(p)) <= math.MaxUint32 {
		a.t.Fatal("foreign owner must exceed 4 GiB")
	}
	return p
}
func (a *asnArena) guards() {
	a.t.Helper()
	for _, span := range a.spans {
		if !bytes.Equal(a.data[span[0]-32:span[0]], bytes.Repeat([]byte{0xa7}, 32)) || !bytes.Equal(a.data[span[1]:span[1]+32], bytes.Repeat([]byte{0xa7}, 32)) {
			a.t.Fatal("native owner/buffer canary changed")
		}
	}
}

func asnEqualJSON(t *testing.T, label string, actual any, expected json.RawMessage) {
	t.Helper()
	b, err := json.Marshal(actual)
	if err != nil {
		t.Fatal(err)
	}
	var a, e any
	if json.Unmarshal(b, &a) != nil || json.Unmarshal(expected, &e) != nil {
		t.Fatal("invalid comparison JSON")
	}
	if !reflect.DeepEqual(a, e) {
		eb, _ := json.Marshal(e)
		t.Fatalf("%s\nactual: %s\noriginal: %s", label, b, eb)
	}
}

func TestAudioSourceNativeOriginalCPU43EDB0CABI(t *testing.T) {
	raw, err := os.ReadFile("testdata/audio_source_original_43edb0.json")
	if err != nil {
		t.Fatal(err)
	}
	if hash := fmt.Sprintf("%x", sha256.Sum256(raw)); hash != "1e744e85304d3b2b805fbdaa2eb18c1a10bbac46af63c0cfa026845bd57454ee" {
		t.Fatalf("original CPU fixture changed: %s", hash)
	}
	var rows []struct {
		Input struct {
			Kind          string
			Lengths       []uint32
			Ready         []int
			Flag, EOF     uint32
			Remaining     *uint32
			RestartLength uint32 `json:"restart_length"`
			EndOnRefill   bool   `json:"end_on_refill"`
			EndAfterLoad  bool   `json:"end_after_load"`
		}
		Events  json.RawMessage
		After   json.RawMessage
		Scratch []string `json:"scratch_sha256"`
	}
	if err := json.Unmarshal(raw, &rows); err != nil || len(rows) != 82 {
		t.Fatalf("82 original CPU cases required: %v / %d", err, len(rows))
	}
	oldServices, oldCallbacks := audioSourceTestServices, audioNativeFixtureCallbackService
	t.Cleanup(func() { audioSourceTestServices, audioNativeFixtureCallbackService = oldServices, oldCallbacks })
	for owner := 0; owner < 3; owner++ {
		a := newASNArena(t)
		for index, row := range rows {
			t.Run(fmt.Sprintf("foreign%d/%s/%03d", owner, row.Input.Kind, index), func(t *testing.T) {
				a.reset()
				// Vary owner placement without changing the sealed input bytes.
				a.take(uintptr(16 + owner*112))
				v := (*AudioVoiceNative)(a.take(unsafe.Sizeof(AudioVoiceNative{})))
				s := (*AudioSample)(a.take(unsafe.Sizeof(AudioSample{})))
				handle := ail.Sample(uintptr(a.take(16)))
				buffers := [2]*byte{(*byte)(a.take(0x4000)), (*byte)(a.take(0x4000))}
				for _, b := range buffers {
					for i := range unsafe.Slice(b, 0x4000) {
						unsafe.Slice(b, 0x4000)[i] = 0xa7
					}
				}
				segments := make([]*byte, len(row.Input.Lengths))
				for i, n := range row.Input.Lengths {
					segments[i] = (*byte)(a.take(uintptr(n) + 1))
					for j := uint32(0); j < n; j++ {
						unsafe.Slice(segments[i], int(n))[j] = byte(j*13 + uint32(i)*47 + 19)
					}
				}
				restart := (*byte)(a.take(0x4001))
				for j := range unsafe.Slice(restart, 0x4000) {
					unsafe.Slice(restart, 0x4000)[j] = byte(j*5 + 71)
				}
				s.Field1, s.Smp, s.Field3, s.Flag7 = unsafe.Pointer(v), handle, row.Input.EOF, row.Input.Flag
				s.Data1, s.Data2 = buffers[0], buffers[1]
				v.Refill, v.Restart, v.Finish = asnCallback(0, false), asnCallback(1, false), asnCallback(2, false)
				install := func(i int) {
					if i < len(segments) {
						v.Cursor, v.Available = segments[i], row.Input.Lengths[i]
						v.Remaining = v.Available
						if row.Input.Remaining != nil {
							v.Remaining = *row.Input.Remaining
						}
					} else {
						v.Available = 0
					}
				}
				if len(segments) != 0 {
					install(0)
				}
				pointer := func(p *byte) any {
					if p == nil {
						return nil
					}
					for i, b := range buffers {
						if uintptr(unsafe.Pointer(p)) >= uintptr(unsafe.Pointer(b)) && uintptr(unsafe.Pointer(p))-uintptr(unsafe.Pointer(b)) <= 0x4000 {
							return []any{"scratch", i, uintptr(unsafe.Pointer(p)) - uintptr(unsafe.Pointer(b))}
						}
					}
					for i, b := range segments {
						if uintptr(unsafe.Pointer(p)) >= uintptr(unsafe.Pointer(b)) && uintptr(unsafe.Pointer(p))-uintptr(unsafe.Pointer(b)) <= uintptr(row.Input.Lengths[i]) {
							return []any{"source", i, uintptr(unsafe.Pointer(p)) - uintptr(unsafe.Pointer(b))}
						}
					}
					if row.Input.RestartLength != 0 && uintptr(unsafe.Pointer(p)) >= uintptr(unsafe.Pointer(restart)) && uintptr(unsafe.Pointer(p))-uintptr(unsafe.Pointer(restart)) <= uintptr(row.Input.RestartLength) {
						return []any{"restart", uintptr(unsafe.Pointer(p)) - uintptr(unsafe.Pointer(restart))}
					}
					t.Fatalf("unknown or truncated native pointer: %p", p)
					return nil
				}
				state := func() any {
					return map[string]any{"eof": s.Field3, "flag": s.Flag7, "cursor": pointer(v.Cursor), "available": v.Available, "remaining": v.Remaining, "sample": uint32(0x12345678)}
				}
				var events []any
				readyPos, next, restarted := 0, 1, false
				audioSourceTestServices = &audioSourceServices{
					UserData: func(h ail.Sample) *AudioSample {
						if h != handle {
							t.Fatalf("native handle truncated: %x != %x", h, handle)
						}
						events = append(events, map[string]any{"service": "userdata", "state": state()})
						return s
					},
					Ready: func(h ail.Sample) int {
						if h != handle || readyPos > len(row.Input.Ready) {
							t.Fatal("handle truncation or unexpected ready loop")
						}
						n := -1
						if readyPos < len(row.Input.Ready) {
							n = row.Input.Ready[readyPos]
						}
						readyPos++
						events = append(events, map[string]any{"service": "ready", "result": n, "state": state()})
						return n
					},
					Load: func(h ail.Sample, slot uint32, p unsafe.Pointer, size uint32) {
						if h != handle {
							t.Fatal("load handle truncated")
						}
						hash := fmt.Sprintf("%x", sha256.Sum256(unsafe.Slice((*byte)(p), int(size))))
						events = append(events, map[string]any{"service": "load", "index": slot, "pointer": pointer((*byte)(p)), "size": size, "sha256": hash, "state": state()})
						if row.Input.EndAfterLoad {
							s.Field3 = 1
						}
					},
				}
				audioNativeFixtureCallbackService = func(kind int, owner unsafe.Pointer) int32 {
					if owner != unsafe.Pointer(v) {
						t.Fatalf("native callback argument truncated: %p", owner)
					}
					names := []string{"refill", "restart", "finish"}
					events = append(events, map[string]any{"service": names[kind], "state": state()})
					switch kind {
					case 0:
						install(next)
						next++
						if row.Input.EndOnRefill {
							s.Field3 = 1
						}
					case 1:
						if row.Input.RestartLength != 0 && !restarted {
							restarted = true
							v.Cursor, v.Available, v.Remaining = restart, row.Input.RestartLength, row.Input.RestartLength
						}
					case 2:
						s.Flag7 = 0x55555555
						return 0xabc
					}
					return 0
				}
				beforeV, beforeS := *v, *s
				if row.Input.Kind == "end" {
					asnScalar(1, unsafe.Pointer(uintptr(handle)), nil, 0)
				} else if result := asnScalar(0, unsafe.Pointer(s), nil, 0); result != -1 {
					t.Fatalf("pump result %d != -1", result)
				}
				asnEqualJSON(t, "service order/payload/state", events, row.Events)
				asnEqualJSON(t, "final state", state(), row.After)
				for i, b := range buffers {
					if hash := fmt.Sprintf("%x", sha256.Sum256(unsafe.Slice(b, 0x4000))); hash != row.Scratch[i] {
						t.Fatalf("scratch %d: %s != %s", i, hash, row.Scratch[i])
					}
				}
				beforeV.Cursor, beforeV.Available, beforeV.Remaining = v.Cursor, v.Available, v.Remaining
				beforeS.Field3, beforeS.Flag7 = s.Field3, s.Flag7
				if !bytes.Equal(unsafe.Slice((*byte)(unsafe.Pointer(v)), unsafe.Sizeof(*v)), unsafe.Slice((*byte)(unsafe.Pointer(&beforeV)), unsafe.Sizeof(beforeV))) || !bytes.Equal(unsafe.Slice((*byte)(unsafe.Pointer(s)), unsafe.Sizeof(*s)), unsafe.Slice((*byte)(unsafe.Pointer(&beforeS)), unsafe.Sizeof(beforeS))) {
					t.Fatal("fields outside the original write set changed")
				}
				a.guards()
			})
		}
	}
}
