package legacy

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type secretWallRawContract4DF120 interface {
	SecretWallStateNative4DF120(bool, server.SecretWallPacketRuntime4DF120) int
}

type secretWallOriginalInput4DF120 struct {
	Name, Entry string
	WallID      uint16 `json:"wall_id"`
	Seed        uint32
	WallKind    int `json:"wall_kind"`
	Route       []int
	Updates     [4]int
	Players     [4]int
	Indices     [4]byte
	Classes     [4]uint32
	Edits       []struct {
		After, Field string
		Slot, Value  int
	}
}

func TestSecretWallStateOriginal4DF120(t *testing.T) {
	TestSecretWallStateNativeLayout4DF120(t)
	raw, err := os.ReadFile("../server/testdata/secret_wall_state_original_4df120.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(raw)) != "c20df57cec10adf344255c8bba6316a6c6292520ddf655fab73d866150cb8f50" {
		t.Fatal("frozen original gzip changed")
	}
	z, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := io.ReadAll(z)
	if err != nil {
		t.Fatal(err)
	}
	if err = z.Close(); err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(decoded)) != "f5051b573ac657881e21adaae6090f2f9257095cc5653bb2ca63686c025fb25b" {
		t.Fatal("frozen decoded original changed")
	}
	var data struct {
		Rows []struct {
			Input  secretWallOriginalInput4DF120
			Output struct {
				Events []map[string]any
				Fault  []json.RawMessage
				Return *int `json:"return_value"`
				Guard  bool `json:"whole_heap_and_guarded_caller_preserved"`
			}
		}
	}
	if err = json.Unmarshal(decoded, &data); err != nil {
		t.Fatal(err)
	}
	if len(data.Rows) != 1188 {
		t.Fatal("original context count", len(data.Rows))
	}
	layout := secretWallOriginalLayout4DF120()
	for _, mode := range []string{"C", "Go"} {
		t.Run(mode, func(t *testing.T) {
			for _, row := range data.Rows {
				t.Run(row.Input.Name, func(t *testing.T) {
					in := row.Input
					if !row.Output.Guard {
						t.Fatal("original guards not certified")
					}
					invalidNative := originalInvalidNativePointer32(t)
					unmapped := func() unsafe.Pointer {
						if invalidNative != nil {
							return invalidNative
						}
						return secretWallOriginalUnmapped4DF120()
					}
					buf, free := alloc.MakeHigh([]byte{}, 0x10000)
					defer free()
					for i := 0; i < len(buf); i += 4 {
						binary.LittleEndian.PutUint32(buf[i:], in.Seed)
					}
					wall := (*server.Wall)(unsafe.Pointer(&buf[0x200]))
					var units [4]*server.Object
					var updates [4]*server.PlayerUpdateData
					var players [4]*server.Player
					for i := range units {
						units[i] = (*server.Object)(unsafe.Pointer(&buf[0x1000+i*0x800]))
						updates[i] = (*server.PlayerUpdateData)(unsafe.Pointer(&buf[0x4000+i*0x800]))
						players[i] = (*server.Player)(unsafe.Pointer(&buf[0x8000+i*0x2000]))
					}
					unitAt := func(index int) *server.Object {
						if index == -1 {
							return nil
						}
						if index == -2 {
							return (*server.Object)(unmapped())
						}
						return units[index]
					}
					updateAt := func(index int) unsafe.Pointer {
						if index == -1 {
							return nil
						}
						if index == -2 {
							return unmapped()
						}
						return unsafe.Pointer(updates[index])
					}
					playerAt := func(index int) *server.Player {
						if index == -1 {
							return nil
						}
						if index == -2 {
							return (*server.Player)(unmapped())
						}
						return players[index]
					}
					wall.Field10 = in.WallID
					for i := range units {
						units[i].ObjClass = object.Class(in.Classes[i])
						storeWallFixturePointerBits(buf, 0x1000+i*0x800+int(unsafe.Offsetof(server.Object{}.UpdateData)), updateAt(in.Updates[i]))
						storeWallFixturePointerBits(buf, 0x4000+i*0x800+int(unsafe.Offsetof(server.PlayerUpdateData{}.Player)), unsafe.Pointer(playerAt(in.Players[i])))
						players[i].PlayerInd = in.Indices[i]
					}
					want := bytes.Clone(buf)
					writeExpected := func(offset, n int, value uint64) {
						for i := 0; i < n; i++ {
							want[offset+i] = byte(value >> (8 * i))
						}
					}
					position, sends := 0, 0
					var payload []byte
					var packetIdentity unsafe.Pointer
					edit := func(after string) {
						for _, e := range in.Edits {
							if e.After != after {
								continue
							}
							switch e.Field {
							case "wall":
								wall.Field10 = uint16(e.Value)
								writeExpected(0x200+int(layout[0]), 2, uint64(e.Value))
							case "update":
								storeWallFixturePointerBits(buf, 0x1000+e.Slot*0x800+int(unsafe.Offsetof(server.Object{}.UpdateData)), updateAt(e.Value))
								writeExpected(0x1000+e.Slot*0x800+int(layout[3]), int(unsafe.Sizeof(uintptr(0))), uint64(uintptr(updateAt(e.Value))))
							case "player":
								storeWallFixturePointerBits(buf, 0x4000+e.Slot*0x800+int(unsafe.Offsetof(server.PlayerUpdateData{}.Player)), unsafe.Pointer(playerAt(e.Value)))
								writeExpected(0x4000+e.Slot*0x800+int(layout[5]), int(unsafe.Sizeof(uintptr(0))), uint64(uintptr(unsafe.Pointer(playerAt(e.Value)))))
							case "index":
								players[e.Slot].PlayerInd = byte(e.Value)
								writeExpected(0x8000+e.Slot*0x2000+int(layout[6]), 1, uint64(e.Value))
							case "packet":
								if len(payload) != 3 {
									t.Fatal("packet mutation without live payload")
								}
								for i := range payload {
									payload[i] = byte(e.Value >> (8 * i))
								}
							default:
								t.Fatal("unknown original mutation", e)
							}
						}
					}
					events := make([]map[string]any, 0)
					runtime := server.SecretWallPacketRuntime4DF120{
						FirstUnit: func() *server.Object {
							index := -1
							if len(in.Route) != 0 {
								index = in.Route[0]
							}
							events = append(events, map[string]any{"kind": "first", "unit": index})
							edit("first")
							return unitAt(index)
						},
						NextUnit: func(current *server.Object) *server.Object {
							if position >= len(in.Route) || current != unitAt(in.Route[position]) {
								t.Fatal("live current-unit argument changed")
							}
							index := -1
							if position+1 < len(in.Route) {
								index = in.Route[position+1]
							}
							events = append(events, map[string]any{"kind": "next", "unit": in.Route[position], "next": index})
							edit("next" + fmt.Sprint(position))
							position++
							return unitAt(index)
						},
						Send: func(recipient int, packet []byte, related *server.Object, remove int) int {
							if len(packet) != 3 || related != nil || remove != 1 {
								t.Fatal("packet service arguments", len(packet), related, remove)
							}
							ptr := unsafe.Pointer(unsafe.SliceData(packet))
							if packetIdentity == nil {
								packetIdentity = ptr
							}
							if ptr != packetIdentity {
								t.Fatal("original packet allocation identity changed within traversal")
							}
							payload = packet
							events = append(events, map[string]any{"kind": "send", "recipient": recipient, "packet": fmt.Sprintf("%x", packet), "size": len(packet), "related": 0, "remove": remove})
							edit("send" + fmt.Sprint(sends))
							sends++
							return int(int32(in.Seed ^ (uint32(sends) * 0x12345678)))
						},
					}
					arg := wall
					if in.WallKind == -1 {
						arg = nil
					}
					if in.WallKind == -2 {
						arg = (*server.Wall)(unmapped())
					}
					var result int
					var fault any
					oldFault := debug.SetPanicOnFault(true)
					defer debug.SetPanicOnFault(oldFault)
					oldRuntime := secretWallNativeRuntimeOverride4DF120
					secretWallNativeRuntimeOverride4DF120 = &runtime
					defer func() { secretWallNativeRuntimeOverride4DF120 = oldRuntime }()
					func() {
						defer func() { fault = recover() }()
						if mode == "C" {
							entry := 0
							if in.Entry == "4DF180" {
								entry = 1
							}
							if !secretWallOriginalTyped4DF120(entry) {
								t.Fatal("raw C port is not implemented")
							}
							result = secretWallOriginalC4DF120(entry, arg)
							return
						}
						native, ok := any(arg).(secretWallRawContract4DF120)
						if !ok {
							t.Fatal("raw native original contract is not implemented")
						}
						result = native.SecretWallStateNative4DF120(in.Entry == "4DF180", runtime)
					}()
					if (fault != nil) != (len(row.Output.Fault) != 0) {
						t.Fatal("original fault boundary", fault, row.Output.Fault)
					}
					if row.Output.Return != nil && result != *row.Output.Return {
						t.Fatal("original return", result, *row.Output.Return)
					}
					gotEvents, _ := json.Marshal(events)
					wantEvents, _ := json.Marshal(row.Output.Events)
					if !bytes.Equal(gotEvents, wantEvents) {
						t.Fatalf("original ordered event prefix differs\ngot %s\nwant %s", gotEvents, wantEvents)
					}
					if !bytes.Equal(buf, want) {
						for i := range buf {
							if buf[i] != want[i] {
								t.Fatalf("whole native64KiB record/canary differs at%x got%02x want%02x", i, buf[i], want[i])
							}
						}
					}
				})
			}
		})
	}
}
