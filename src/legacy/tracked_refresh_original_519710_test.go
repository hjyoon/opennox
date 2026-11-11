package legacy

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type refreshInput519710 struct {
	Count     int32             `json:"count"`
	Last      uint32            `json:"last"`
	Frame     uint32            `json:"frame"`
	Index     byte              `json:"index"`
	Mutations [][2]any          `json:"mutations"`
	Sequence  []json.RawMessage `json:"sequence"`
	NilUpdate bool              `json:"nil_update"`
	NilPlayer bool              `json:"nil_player"`
}
type refreshEvent519710 struct {
	Service   string   `json:"service"`
	Index     uint32   `json:"index"`
	Return    uint32   `json:"return_word"`
	Mutations [][2]any `json:"mutations"`
}
type refreshState519710 struct {
	Last       uint32  `json:"last"`
	Frame      uint32  `json:"frame"`
	Player     *string `json:"player"`
	Index      byte    `json:"index"`
	OtherIndex byte    `json:"other_index"`
}
type refreshOutput519710 struct {
	Return uint32               `json:"return_word"`
	Events []refreshEvent519710 `json:"events"`
	Reads  [][2]any             `json:"reads"`
	State  refreshState519710   `json:"state"`
	Guards bool                 `json:"all_other_heap_bytes_preserved"`
}
type refreshCase519710 struct {
	Name   string                `json:"name"`
	Input  refreshInput519710    `json:"input"`
	Output []refreshOutput519710 `json:"output"`
}

func refreshOriginalCases519710(t *testing.T) []refreshCase519710 {
	t.Helper()
	blob, err := os.ReadFile(filepath.Join("..", "server", "testdata", "tracked_refresh_original_519710.json.gz"))
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(blob)) != "bb8195ad55dbea82058220f37af0996a19ef1fa3e22ae2f722da4ced5fa810f8" {
		t.Fatal("frozen compressed reference changed")
	}
	z, err := gzip.NewReader(bytes.NewReader(blob))
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := io.ReadAll(z)
	if err != nil {
		t.Fatal(err)
	}
	if err := z.Close(); err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(decoded)) != "afce8985033666560a40efb098041d7f5097b10385f319209b28a6678083a560" {
		t.Fatal("frozen decoded reference changed")
	}
	var data struct {
		Contexts []refreshCase519710 `json:"contexts"`
	}
	if err := json.Unmarshal(decoded, &data); err != nil {
		t.Fatal(err)
	}
	if len(data.Contexts) != 1975 {
		t.Fatalf("contexts=%d", len(data.Contexts))
	}
	return data.Contexts
}

type refreshGuardUpdate519710 struct {
	Before [32]byte
	Value  server.PlayerUpdateData
	After  [32]byte
}
type refreshGuardPlayer519710 struct {
	Before [32]byte
	Value  server.Player
	After  [32]byte
}
type refreshMemory519710 struct {
	Pointer  unsafe.Pointer
	Expected []byte
}
type refreshFixture519710 struct {
	Update        *refreshGuardUpdate519710
	Player, Other *refreshGuardPlayer519710
	Memory        []refreshMemory519710
	Frame         uint32
}

func refreshBytes519710(pointer unsafe.Pointer, size uintptr) []byte {
	return unsafe.Slice((*byte)(pointer), int(size))
}
func newRefreshFixture519710(t *testing.T, in refreshInput519710) *refreshFixture519710 {
	t.Helper()
	u, fu := alloc.NewHigh(refreshGuardUpdate519710{})
	p, fp := alloc.NewHigh(refreshGuardPlayer519710{})
	o, fo := alloc.NewHigh(refreshGuardPlayer519710{})
	t.Cleanup(fu)
	t.Cleanup(fp)
	t.Cleanup(fo)
	f := &refreshFixture519710{Update: u, Player: p, Other: o, Frame: in.Frame}
	u.Value.Field67 = in.Last
	u.Value.Player = &p.Value
	if in.NilPlayer {
		u.Value.Player = nil
	}
	p.Value.PlayerInd = in.Index
	o.Value.PlayerInd = 11
	for _, guards := range [][2]*[32]byte{{&u.Before, &u.After}, {&p.Before, &p.After}, {&o.Before, &o.After}} {
		for i := range guards[0] {
			guards[0][i] = 0xa7
			guards[1][i] = 0x5c
		}
	}
	for _, m := range []struct {
		ptr  unsafe.Pointer
		size uintptr
	}{{unsafe.Pointer(u), unsafe.Sizeof(*u)}, {unsafe.Pointer(p), unsafe.Sizeof(*p)}, {unsafe.Pointer(o), unsafe.Sizeof(*o)}} {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(m.ptr) <= math.MaxUint32 {
			t.Fatal("native allocations must exceed4GiB")
		}
		f.Memory = append(f.Memory, refreshMemory519710{Pointer: m.ptr, Expected: bytes.Clone(refreshBytes519710(m.ptr, m.size))})
	}
	return f
}
func (f *refreshFixture519710) patch(t *testing.T, pointer unsafe.Pointer, size uintptr) {
	t.Helper()
	for _, m := range f.Memory {
		if uintptr(pointer) >= uintptr(m.Pointer) && uintptr(pointer)+size <= uintptr(m.Pointer)+uintptr(len(m.Expected)) {
			off := uintptr(pointer) - uintptr(m.Pointer)
			copy(m.Expected[off:off+size], refreshBytes519710(pointer, size))
			return
		}
	}
	t.Fatal("test mutation outside allocated fields")
}
func (f *refreshFixture519710) mutate(t *testing.T, name string, value any) {
	t.Helper()
	switch name {
	case "last":
		f.Update.Value.Field67 = uint32(value.(float64))
		f.patch(t, unsafe.Pointer(&f.Update.Value.Field67), 4)
	case "frame":
		f.Frame = uint32(value.(float64))
	case "index":
		f.Player.Value.PlayerInd = byte(value.(float64))
		f.patch(t, unsafe.Pointer(&f.Player.Value.PlayerInd), 1)
	case "player":
		f.Update.Value.Player = nil
		if value == "other" {
			f.Update.Value.Player = &f.Other.Value
		} else if value == "player" {
			f.Update.Value.Player = &f.Player.Value
		} else if value != nil {
			t.Fatalf("unknown player %v", value)
		}
		f.patch(t, unsafe.Pointer(&f.Update.Value.Player), unsafe.Sizeof(f.Update.Value.Player))
	default:
		t.Fatalf("unknown mutation %q", name)
	}
}
func (f *refreshFixture519710) check(t *testing.T, want refreshState519710) {
	t.Helper()
	for _, m := range f.Memory {
		if !bytes.Equal(m.Expected, refreshBytes519710(m.Pointer, uintptr(len(m.Expected)))) {
			t.Fatal("complete native allocation or neighbor changed")
		}
	}
	var label *string
	if p := f.Update.Value.Player; p != nil {
		name := "player"
		if p == &f.Other.Value {
			name = "other"
		} else if p != &f.Player.Value {
			t.Fatalf("unrecognized Player pointer %p", p)
		}
		label = &name
	}
	got := refreshState519710{Last: f.Update.Value.Field67, Frame: f.Frame, Player: label, Index: f.Player.Value.PlayerInd, OtherIndex: f.Other.Value.PlayerInd}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("state=%+v want=%+v", got, want)
	}
}
func compareRefreshOriginal519710(t *testing.T, call func(*server.PlayerUpdateData) int32) {
	for _, c := range refreshOriginalCases519710(t) {
		t.Run(c.Name, func(t *testing.T) {
			current := c.Input
			f := newRefreshFixture519710(t, current)
			var events []refreshEvent519710
			var frameCalls int
			previous := trackedRefreshRuntimeNative519710
			trackedRefreshRuntimeNative519710 = func() server.TrackedRefreshRuntime519710 {
				return server.TrackedRefreshRuntime519710{
					Count: func(index int32) int32 {
						for _, m := range current.Mutations {
							f.mutate(t, m[0].(string), m[1])
						}
						events = append(events, refreshEvent519710{Service: "tracked_count", Index: uint32(index), Return: uint32(current.Count), Mutations: current.Mutations})
						return current.Count
					},
					Frame: func() uint32 { frameCalls++; return f.Frame },
				}
			}
			t.Cleanup(func() { trackedRefreshRuntimeNative519710 = previous })
			if len(c.Output) != 1+len(current.Sequence) {
				t.Fatal("reference sequence length")
			}
			for i, want := range c.Output {
				if i != 0 {
					change := current.Sequence[i-1]
					if err := json.Unmarshal(change, &current); err != nil {
						t.Fatal(err)
					}
					var fields map[string]any
					if err := json.Unmarshal(change, &fields); err != nil {
						t.Fatal(err)
					}
					for name, value := range fields {
						if name != "count" {
							f.mutate(t, name, value)
						}
					}
				}
				events = nil
				frameCalls = 0
				got := call(&f.Update.Value)
				if uint32(got) != want.Return {
					t.Fatalf("call%d return=%d want=%d", i, got, want.Return)
				}
				if !reflect.DeepEqual(events, want.Events) {
					t.Fatalf("call%d events=%+v want=%+v", i, events, want.Events)
				}
				if frameCalls*2 != len(want.Reads) {
					t.Fatalf("call%d Frame calls=%d original reads=%v", i, frameCalls, want.Reads)
				}
				if !want.Guards {
					t.Fatal("original guard evidence missing")
				}
				f.check(t, want.State)
			}
		})
	}
}
func TestTrackedRefreshOriginalCGo519710(t *testing.T) {
	compareRefreshOriginal519710(t, trackedRefreshOriginalPublicC519710)
}
func TestTrackedRefreshOriginalGo519710(t *testing.T) {
	compareRefreshOriginal519710(t, TrackedObjectRefresh519710)
}
func TestTrackedRefreshOriginalNilFaults519710(t *testing.T) {
	for _, mode := range []struct {
		name string
		call func(*server.PlayerUpdateData) int32
	}{{"CGo", trackedRefreshOriginalPublicC519710}, {"Go", TrackedObjectRefresh519710}} {
		for _, field := range []string{"update", "player"} {
			t.Run(mode.name+"/"+field, func(t *testing.T) {
				f := newRefreshFixture519710(t, refreshInput519710{NilPlayer: true})
				update := &f.Update.Value
				if field == "update" {
					update = nil
				}
				previous := trackedRefreshRuntimeNative519710
				trackedRefreshRuntimeNative519710 = func() server.TrackedRefreshRuntime519710 {
					return server.TrackedRefreshRuntime519710{Count: func(int32) int32 { t.Fatal("fault must precede child"); return 0 }, Frame: func() uint32 { t.Fatal("fault must precede frame"); return 0 }}
				}
				t.Cleanup(func() { trackedRefreshRuntimeNative519710 = previous })
				defer func() {
					if recover() == nil {
						t.Fatal("original nil fault was swallowed")
					}
				}()
				mode.call(update)
			})
		}
	}
}
