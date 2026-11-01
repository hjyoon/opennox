package legacy

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/server"
)

type playerSendOutput518C30 struct {
	Return uint32           `json:"return_word"`
	Events []map[string]any `json:"events"`
	Reads  [][2]any         `json:"global_frame_reads"`
	State  map[string]any   `json:"state"`
	Guards bool             `json:"all_other_heap_bytes_preserved"`
}
type playerSendRow518C30 struct {
	Name   string                   `json:"name"`
	Input  map[string]any           `json:"input"`
	Output []playerSendOutput518C30 `json:"output"`
}

func playerSendNumber518C30(value any) uint32      { return uint32(int64(value.(float64))) }
func playerSendMap518C30(value any) map[string]any { return value.(map[string]any) }
func playerSendLoad518C30(t *testing.T, nearest bool) ([]playerSendRow518C30, []playerSendRow518C30) {
	t.Helper()
	name, compressed, decoded := "player_send_original_startup_518c30.json.gz", "eefc2dca42d31f0f3da476be0eeef95da7e5fe14eacde47d66503b540dc07d40", "23823d7c2d0efac48163ea100318e446df8289389aaf72681d74bd42ed1a0c9f"
	if nearest {
		name, compressed, decoded = "player_send_original_nearest_518c30.json.gz", "59151c7ba200ee569b339cfc2962be97d33c588b2974ffbccd70a6128cf70e33", "b9d0956ee2b478e3340a0bd3fcc677b1fc760f29825f7ab4df37470c8623a4d9"
	}
	blob, err := os.ReadFile(filepath.Join("..", "server", "testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(blob)) != compressed {
		t.Fatal("frozen compressed reference changed")
	}
	z, err := gzip.NewReader(bytes.NewReader(blob))
	if err != nil {
		t.Fatal(err)
	}
	raw, err := io.ReadAll(z)
	if err != nil {
		t.Fatal(err)
	}
	if err = z.Close(); err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(raw)) != decoded {
		t.Fatal("frozen decoded reference changed")
	}
	var data struct {
		Contexts []playerSendRow518C30 `json:"contexts"`
		Faults   []playerSendRow518C30 `json:"expected_fault_contexts"`
	}
	if err = json.Unmarshal(raw, &data); err != nil {
		t.Fatal(err)
	}
	if len(data.Contexts) != 2075 || len(data.Faults) != 9 {
		t.Fatal("original context dimensions")
	}
	return data.Contexts, data.Faults
}

type playerSendGuard518C30[T comparable] struct {
	Before [32]byte
	Value  T
	After  [32]byte
}
type playerSendMemory518C30 struct {
	Pointer  unsafe.Pointer
	Expected []byte
}
type playerSendFixture518C30 struct {
	t       *testing.T
	s       *server.Server
	objects [2]*server.Object
	updates [4]*server.PlayerUpdateData
	players [4]*server.Player
	health  [2]*server.HealthData
	r, u    *server.Object
	cr, cu  *server.PlayerUpdateData
	memory  []playerSendMemory518C30
	input   map[string]any
	want    *playerSendOutput518C30
	events  []map[string]any
	frames  []uint32
	nearest bool
}

func playerSendBytes518C30(pointer unsafe.Pointer, size uintptr) []byte {
	return unsafe.Slice((*byte)(pointer), int(size))
}
func playerSendAllocate518C30[T comparable](f *playerSendFixture518C30) *T {
	f.t.Helper()
	p, free := alloc.New(playerSendGuard518C30[T]{})
	f.t.Cleanup(free)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(p)) <= math.MaxUint32 {
		f.t.Fatal("native allocations must exceed4GiB")
	}
	for i := range p.Before {
		p.Before[i] = 0xa7
		p.After[i] = 0x5c
	}
	f.memory = append(f.memory, playerSendMemory518C30{Pointer: unsafe.Pointer(p), Expected: make([]byte, int(unsafe.Sizeof(*p)))})
	return &p.Value
}
func (f *playerSendFixture518C30) pointer(label any) unsafe.Pointer {
	if label == nil {
		return nil
	}
	switch label.(string) {
	case "receiver_update":
		return unsafe.Pointer(f.updates[0])
	case "unit_update":
		return unsafe.Pointer(f.updates[1])
	case "other_receiver_update":
		return unsafe.Pointer(f.updates[2])
	case "other_unit_update":
		return unsafe.Pointer(f.updates[3])
	case "receiver_player":
		return unsafe.Pointer(f.players[0])
	case "unit_player":
		return unsafe.Pointer(f.players[1])
	case "other_receiver_player":
		return unsafe.Pointer(f.players[2])
	case "other_unit_player":
		return unsafe.Pointer(f.players[3])
	case "health":
		return unsafe.Pointer(f.health[0])
	case "other_health":
		return unsafe.Pointer(f.health[1])
	default:
		f.t.Fatalf("unknown original pointer label %v", label)
		return nil
	}
}
func newPlayerSendFixture518C30(t *testing.T, s *server.Server, input map[string]any, nearest bool) *playerSendFixture518C30 {
	f := &playerSendFixture518C30{t: t, s: s, input: make(map[string]any), nearest: nearest}
	for k, v := range input {
		f.input[k] = v
	}
	for i := range f.objects {
		f.objects[i] = playerSendAllocate518C30[server.Object](f)
	}
	for i := range f.updates {
		f.updates[i] = playerSendAllocate518C30[server.PlayerUpdateData](f)
	}
	for i := range f.players {
		f.players[i] = playerSendAllocate518C30[server.Player](f)
	}
	for i := range f.health {
		f.health[i] = playerSendAllocate518C30[server.HealthData](f)
	}
	f.r, f.u = f.objects[0], f.objects[1]
	if input["same_unit"].(bool) {
		f.u = f.r
	}
	f.cr, f.cu = f.updates[0], f.updates[1]
	if input["shared_update"].(bool) {
		f.cu = f.cr
	}
	f.updates[0].Player = f.players[0]
	f.updates[1].Player = f.players[1]
	if input["shared_player"].(bool) {
		f.updates[1].Player = f.players[0]
	}
	f.updates[2].Player, f.updates[3].Player = f.players[2], f.players[3]
	f.r.UpdateData, f.u.UpdateData = unsafe.Pointer(f.cr), unsafe.Pointer(f.cu)
	if input["nil_receiver_update"] == true {
		f.r.UpdateData = nil
	}
	if input["nil_unit_update"] == true {
		f.u.UpdateData = nil
	}
	if input["nil_receiver_player"] == true {
		f.cr.Player = nil
	}
	if input["nil_unit_player"] == true {
		f.cu.Player = nil
	}
	indices := []uint32{playerSendNumber518C30(input["receiver_index"]), playerSendNumber518C30(input["unit_index"]), 23, 29}
	for i, p := range f.players {
		p.PlayerInd = byte(indices[i])
		for j := range p.PeerUpdateFrames {
			p.PeerUpdateFrames[j] = 0x12345600 + uint32(j)
		}
	}
	f.r.ObjClass, f.u.ObjClass = object.ClassPlayer, object.ClassPlayer
	if input["health_present"].(bool) {
		f.u.HealthData = f.health[0]
	}
	f.health[0].Cur, f.health[1].Cur = uint16(playerSendNumber518C30(input["health"])), 601
	for _, name := range []string{"type_index", "net_code", "direction", "field131", "damage_frame", "state", "anim_frame", "random_flag", "expire"} {
		f.change(name, input[name])
	}
	position := input["position_bits"].([]any)
	f.u.PosVec.X, f.u.PosVec.Y = math.Float32frombits(playerSendNumber518C30(position[0])), math.Float32frombits(playerSendNumber518C30(position[1]))
	f.s.SetFrame(playerSendNumber518C30(input["frame"]))
	for i := range f.cu.HealthSamples {
		f.cu.HealthSamples[i] = 0x2200 + uint16(i)
	}
	f.cu.HealthSamples[indices[0]] = uint16(playerSendNumber518C30(input["sample"]))
	for _, m := range f.memory {
		copy(m.Expected, playerSendBytes518C30(m.Pointer, uintptr(len(m.Expected))))
	}
	return f
}
func (f *playerSendFixture518C30) change(name string, value any) {
	f.t.Helper()
	switch name {
	case "frame":
		f.s.SetFrame(playerSendNumber518C30(value))
	case "health":
		f.health[0].Cur = uint16(playerSendNumber518C30(value))
	case "other_health":
		f.health[1].Cur = uint16(playerSendNumber518C30(value))
	case "health_link":
		f.u.HealthData = nil
		if value != nil {
			f.u.HealthData = f.health[0]
			if value == "other" {
				f.u.HealthData = f.health[1]
			}
		}
	case "receiver_update":
		f.r.UpdateData = unsafe.Pointer(f.updates[0])
		if value == "other" {
			f.r.UpdateData = unsafe.Pointer(f.updates[2])
		}
	case "unit_update":
		f.u.UpdateData = unsafe.Pointer(f.updates[1])
		if value == "other" {
			f.u.UpdateData = unsafe.Pointer(f.updates[3])
		}
	case "cached_receiver_player":
		f.cr.Player = nil
		if value != nil {
			f.cr.Player = f.players[0]
			if value == "other" {
				f.cr.Player = f.players[2]
			}
		}
	case "cached_unit_player":
		f.cu.Player = nil
		if value != nil {
			f.cu.Player = f.players[1]
			if value == "other" {
				f.cu.Player = f.players[3]
			}
		}
	case "receiver_index":
		f.players[0].PlayerInd = byte(playerSendNumber518C30(value))
	case "unit_index":
		index := 1
		if f.input["shared_player"].(bool) {
			index = 0
		}
		f.players[index].PlayerInd = byte(playerSendNumber518C30(value))
	case "type_index":
		f.u.TypeInd = uint16(playerSendNumber518C30(value))
	case "net_code":
		f.u.NetCode = playerSendNumber518C30(value)
	case "position_x":
		f.u.PosVec.X = math.Float32frombits(playerSendNumber518C30(value))
	case "position_y":
		f.u.PosVec.Y = math.Float32frombits(playerSendNumber518C30(value))
	case "direction":
		f.u.Direction1 = server.Dir16(playerSendNumber518C30(value))
	case "field131":
		f.u.Field131 = playerSendNumber518C30(value)
	case "damage_frame":
		f.u.Frame134 = playerSendNumber518C30(value)
	case "state":
		f.cu.State = server.PlayerState(playerSendNumber518C30(value))
	case "anim_frame":
		f.cu.Field59_0 = byte(playerSendNumber518C30(value))
	case "random_flag":
		f.cu.Field40_0 = uint16(playerSendNumber518C30(value))
	case "expire":
		f.cu.Field41 = playerSendNumber518C30(value)
	default:
		f.t.Fatalf("unknown mutation %s", name)
	}
}
func (f *playerSendFixture518C30) mutations(kind string) []any {
	if value, ok := playerSendMap518C30(f.input["mutations"])[kind]; ok {
		return value.([]any)
	}
	return []any{}
}
func (f *playerSendFixture518C30) mutate(kind string) {
	for _, value := range f.mutations(kind) {
		pair := value.([]any)
		f.change(pair[0].(string), pair[1])
	}
}
func (f *playerSendFixture518C30) expect(pointer unsafe.Pointer, value any) {
	f.t.Helper()
	var blob []byte
	switch v := value.(type) {
	case byte:
		blob = []byte{v}
	case uint16:
		blob = make([]byte, 2)
		binary.LittleEndian.PutUint16(blob, v)
	case uint32:
		blob = make([]byte, 4)
		binary.LittleEndian.PutUint32(blob, v)
	case unsafe.Pointer:
		blob = make([]byte, unsafe.Sizeof(v))
		if len(blob) == 8 {
			binary.LittleEndian.PutUint64(blob, uint64(uintptr(v)))
		} else {
			binary.LittleEndian.PutUint32(blob, uint32(uintptr(v)))
		}
	default:
		f.t.Fatalf("unsupported expected type %T", value)
	}
	for _, m := range f.memory {
		if uintptr(pointer) >= uintptr(m.Pointer) && uintptr(pointer)+uintptr(len(blob)) <= uintptr(m.Pointer)+uintptr(len(m.Expected)) {
			offset := uintptr(pointer) - uintptr(m.Pointer)
			copy(m.Expected[offset:offset+uintptr(len(blob))], blob)
			return
		}
	}
	f.t.Fatal("expected field outside native allocations")
}
func (f *playerSendFixture518C30) checkState() {
	f.t.Helper()
	state := f.want.State
	objects := playerSendMap518C30(state["objects"])
	for name, o := range map[string]*server.Object{"recipient": f.r, "unit": f.u} {
		v := playerSendMap518C30(objects[name])
		pos := v["position_bits"].([]any)
		for _, field := range []struct {
			pointer unsafe.Pointer
			value   any
		}{{unsafe.Pointer(&o.TypeInd), uint16(playerSendNumber518C30(v["type_index"]))}, {unsafe.Pointer(&o.NetCode), playerSendNumber518C30(v["net_code"])}, {unsafe.Pointer(&o.PosVec.X), playerSendNumber518C30(pos[0])}, {unsafe.Pointer(&o.PosVec.Y), playerSendNumber518C30(pos[1])}, {unsafe.Pointer(&o.Direction1), uint16(playerSendNumber518C30(v["direction"]))}, {unsafe.Pointer(&o.Field131), playerSendNumber518C30(v["field131"])}, {unsafe.Pointer(&o.Frame134), playerSendNumber518C30(v["damage_frame"])}, {unsafe.Pointer(&o.UpdateData), f.pointer(v["update"])}, {unsafe.Pointer(&o.HealthData), f.pointer(v["health"])}} {
			f.expect(field.pointer, field.value)
		}
	}
	updates := playerSendMap518C30(state["updates"])
	for i, name := range []string{"receiver_update", "unit_update", "other_receiver_update", "other_unit_update"} {
		u := f.updates[i]
		v := playerSendMap518C30(updates[name])
		f.expect(unsafe.Pointer(&u.Player), f.pointer(v["player"]))
		for j, n := range v["samples"].([]any) {
			f.expect(unsafe.Pointer(&u.HealthSamples[j]), uint16(playerSendNumber518C30(n)))
		}
		f.expect(unsafe.Pointer(&u.State), byte(playerSendNumber518C30(v["state"])))
		f.expect(unsafe.Pointer(&u.Field59_0), byte(playerSendNumber518C30(v["anim_frame"])))
		f.expect(unsafe.Pointer(&u.Field40_0), uint16(playerSendNumber518C30(v["random_flag"])))
		f.expect(unsafe.Pointer(&u.Field41), playerSendNumber518C30(v["expire"]))
	}
	players := playerSendMap518C30(state["players"])
	for i, name := range []string{"receiver_player", "unit_player", "other_receiver_player", "other_unit_player"} {
		p := f.players[i]
		v := playerSendMap518C30(players[name])
		f.expect(unsafe.Pointer(&p.PlayerInd), byte(playerSendNumber518C30(v["index"])))
		for j, n := range v["peers"].([]any) {
			f.expect(unsafe.Pointer(&p.PeerUpdateFrames[j]), playerSendNumber518C30(n))
		}
	}
	for i, n := range state["health"].([]any) {
		f.expect(unsafe.Pointer(&f.health[i].Cur), uint16(playerSendNumber518C30(n)))
	}
	if f.s.Frame() != playerSendNumber518C30(state["frame"]) {
		f.t.Fatal("frame differs from original")
	}
	for _, m := range f.memory {
		if !bytes.Equal(m.Expected, playerSendBytes518C30(m.Pointer, uintptr(len(m.Expected)))) {
			f.t.Fatal("complete native allocation or neighbor differs from frozen original")
		}
	}
}

func playerSendCanonical518C30(t *testing.T, value any) any {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var out any
	if err = json.Unmarshal(raw, &out); err != nil {
		t.Fatal(err)
	}
	return out
}
func (f *playerSendFixture518C30) event(kind string, args []any, ret *uint32, out []any) {
	f.t.Helper()
	index := len(f.events)
	if index >= len(f.want.Events) {
		f.t.Fatalf("unexpected child%s args=%v", kind, args)
	}
	got := map[string]any{"service": kind, "arguments": args, "mutations": f.mutations(kind)}
	if kind == "convert" {
		delete(got, "mutations")
	}
	if ret != nil {
		got["return_word"] = *ret
	}
	if out != nil {
		got["out"] = out
	}
	want := make(map[string]any)
	for key, value := range f.want.Events[index] {
		if key != "ignored_high_word_diagnostic" {
			want[key] = value
		}
	}
	if !reflect.DeepEqual(playerSendCanonical518C30(f.t, got), want) {
		f.t.Fatalf("child%d=%v want=%v", index, got, want)
	}
	f.events = append(f.events, got)
	f.mutate(kind)
}
func playerSendReturnPointer518C30(value uint32) *uint32 { return &value }

func (f *playerSendFixture518C30) runtime() server.PlayerObjectSendRuntime518C30 {
	return server.PlayerObjectSendRuntime518C30{
		Frame: func() uint32 { value := f.s.Frame(); f.frames = append(f.frames, value); return value },
		HealthDelta: func(index int32, code uint16, delta int16) {
			f.event("health_delta", []any{uint32(index), int16(code), delta}, nil, nil)
			// This is the actual unchanged portable native C4D8760 child,
			// not a duplicate of its packet-producing implementation.
			first, _ := importantNativeListC()
			if first != nil {
				f.t.Fatal("health queue not empty before actual Cchild")
			}
			complexObjectReportHealthC518960(index, code, delta)
			recipient, packet := complexObjectTakeImportantPacketC518960()
			if packet != nil {
				f.event("health_packet", []any{uint32(recipient), fmt.Sprintf("%x", packet), 5, 0, 1}, playerSendReturnPointer518C30(0x76543210), nil)
			}
			first, _ = importantNativeListC()
			if first != nil {
				f.t.Fatal("actual health packet leaked")
			}
		},
		Self: func(unit *server.Object) {
			if unit != f.r {
				f.t.Fatal("self pointer")
			}
			f.event("self", []any{"recipient"}, playerSendReturnPointer518C30(0x76543210), nil)
		},
		UnitCode: func(unit *server.Object) uint32 {
			if unit != f.u {
				f.t.Fatal("code pointer")
			}
			value := playerSendNumber518C30(f.input["code_return"])
			f.event("unit_code", []any{"unit"}, &value, nil)
			return value
		},
		FloatToInt: func(value float32) int32 {
			var result int32
			if f.nearest {
				// Alternate x87 mode is an explicit converter boundary replay.
				// Production startup0xe3f uses the actual C converter below.
				index := len(f.events)
				if index >= len(f.want.Events) || f.want.Events[index]["service"] != "convert" {
					f.t.Fatal("alternate converter order")
				}
				result = int32(playerSendNumber518C30(f.want.Events[index]["return_word"]))
			} else {
				result = complexObjectFloatToIntC518960(value)
			}
			f.event("convert", []any{math.Float32bits(value)}, playerSendReturnPointer518C30(uint32(result)), nil)
			return result
		},
		Action: func(unit *server.Object) uint32 {
			if unit != f.u {
				f.t.Fatal("action pointer")
			}
			value := playerSendNumber518C30(f.input["action_return"])
			f.event("action", []any{"unit"}, &value, nil)
			return value
		},
		Random: func(min, max int32) int32 {
			value := playerSendNumber518C30(f.input["random_return"])
			// The original immutable debug filename/line are retained in the
			// observed trace; the native415FA0 export consumes min/max only.
			f.event("random", []any{min, max, `C:\NoxPost\src\Server\Network\upclient.c`, 399}, &value, nil)
			return int32(value)
		},
		Frames: func(index int32) (int32, int32) {
			count, delay := playerSendNumber518C30(f.input["range_count"]), playerSendNumber518C30(f.input["range_delay"])
			f.event("range", []any{index}, playerSendReturnPointer518C30(0x76543210), []any{count, delay})
			return int32(count), int32(delay)
		},
		Direction: func(index int32) server.IndexedDirectionVector509E20 {
			words := f.input["direction_out"].([]any)
			x, y := playerSendNumber518C30(words[0]), playerSendNumber518C30(words[1])
			f.event("direction", []any{uint32(index)}, playerSendReturnPointer518C30(0x76543210), []any{x, y})
			return server.IndexedDirectionVector509E20{X: int32(x), Y: int32(y)}
		},
		ServerPacket: func(index int32, packet [12]byte) int32 {
			value := playerSendNumber518C30(f.input["publication_return"])
			f.event("server_packet", []any{uint32(index), fmt.Sprintf("%x", packet[:]), 12}, &value, nil)
			return int32(value)
		},
		ClientPacket: func(index, kind int32, packet [12]byte) int32 {
			value := playerSendNumber518C30(f.input["publication_return"])
			f.event("client_packet", []any{uint32(index), kind, fmt.Sprintf("%x", packet[:]), 12}, &value, nil)
			return int32(value)
		},
	}
}

type playerSendLegacyServer518C30 struct {
	Server
	native *server.Server
}

func (s *playerSendLegacyServer518C30) S() *server.Server { return s.native }
func comparePlayerSendOriginal518C30(t *testing.T, invoke func(*server.Object, *server.Object, int32, int32) int32, nearest bool, faults bool) {
	t.Helper()
	rows, bad := playerSendLoad518C30(t, nearest)
	if faults {
		rows = bad
	}
	handles.Init()
	t.Cleanup(handles.Release)
	s := server.New(nil, nil, strman.New())
	t.Cleanup(s.Close)
	preserveImportantPacketState(t)
	setImportantNativeListC(nil, nil)
	setImportantListHeadsC(0, 0)
	setImportantAllocClassC(nil)
	setImportantRecipientMaskC(0)
	previous := GetServer
	GetServer = func() Server { return &playerSendLegacyServer518C30{native: s} }
	t.Cleanup(func() { GetServer = previous })
	for _, row := range rows {
		t.Run(row.Name, func(t *testing.T) {
			f := newPlayerSendFixture518C30(t, s, row.Input, nearest)
			oldRuntime, oldTransport := playerObjectSendRuntimeNative518C30, s.NetSendPacketXxx
			playerObjectSendRuntimeNative518C30 = f.runtime
			t.Cleanup(func() { playerObjectSendRuntimeNative518C30 = oldRuntime; s.NetSendPacketXxx = oldTransport })
			sequence := f.input["sequence"].([]any)
			if len(row.Output) != len(sequence)+1 {
				t.Fatal("original repeat dimensions")
			}
			for index, want := range row.Output {
				if index != 0 {
					for name, value := range playerSendMap518C30(sequence[index-1]) {
						f.input[name] = value
						switch name {
						case "frame", "state", "anim_frame", "random_flag", "expire", "field131":
							f.change(name, value)
						}
					}
				}
				f.cr, f.cu = (*server.PlayerUpdateData)(f.r.UpdateData), (*server.PlayerUpdateData)(f.u.UpdateData)
				f.want = &want
				f.events = nil
				f.frames = nil
				r, u := f.r, f.u
				if f.input["nil_receiver"] == true {
					r = nil
				}
				if f.input["nil_unit"] == true {
					u = nil
				}
				var returned int32
				var recovered any
				func() {
					defer func() { recovered = recover() }()
					returned = invoke(r, u, int32(playerSendNumber518C30(f.input["stream_flag"])), int32(playerSendNumber518C30(f.input["health_flag"])))
				}()
				if faults {
					if recovered == nil {
						t.Fatal("original pointer/divide fault swallowed")
					}
				} else {
					if recovered != nil {
						t.Fatalf("unexpected native panic: %v", recovered)
					}
					if uint32(returned) != want.Return {
						t.Fatalf("call%d return=%#x want=%#x", index, uint32(returned), want.Return)
					}
				}
				if len(f.events) != len(want.Events) {
					t.Fatalf("call%d child count=%d want=%d", index, len(f.events), len(want.Events))
				}
				if len(f.frames) != len(want.Reads) {
					t.Fatalf("call%d frame read count=%d want=%d", index, len(f.frames), len(want.Reads))
				}
				for i, value := range f.frames {
					if value != playerSendNumber518C30(want.Reads[i][1]) {
						t.Fatalf("frame read%d differs", i)
					}
				}
				if !want.Guards {
					t.Fatal("original complete heap/neighbor proof missing")
				}
				f.checkState()
			}
		})
	}
}
func TestPlayerSendOriginalCGoStartup518C30(t *testing.T) {
	comparePlayerSendOriginal518C30(t, playerSendOriginalPublicC518C30, false, false)
}
func TestPlayerSendOriginalGoStartup518C30(t *testing.T) {
	comparePlayerSendOriginal518C30(t, NetPlayerObjectSend518C30, false, false)
}
func TestPlayerSendOriginalCGoAlternateConverterBoundary518C30(t *testing.T) {
	comparePlayerSendOriginal518C30(t, playerSendOriginalPublicC518C30, true, false)
}
func TestPlayerSendOriginalGoAlternateConverterBoundary518C30(t *testing.T) {
	comparePlayerSendOriginal518C30(t, NetPlayerObjectSend518C30, true, false)
}
func TestPlayerSendOriginalCGoFaults518C30(t *testing.T) {
	comparePlayerSendOriginal518C30(t, playerSendOriginalPublicC518C30, false, true)
}
func TestPlayerSendOriginalGoFaults518C30(t *testing.T) {
	comparePlayerSendOriginal518C30(t, NetPlayerObjectSend518C30, false, true)
}
