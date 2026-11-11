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
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type dispatchInput519410 struct {
	Index         byte                `json:"index"`
	Same          bool                `json:"same_object"`
	Class         uint32              `json:"object_class"`
	Flags         uint32              `json:"flags"`
	Field5        uint32              `json:"field5"`
	FriendPending uint32              `json:"friend_pending"`
	FriendValue   uint32              `json:"friend_value"`
	Seen          uint32              `json:"seen"`
	Pending       uint32              `json:"pending"`
	State         uint32              `json:"state"`
	Camera        [2]uint32           `json:"camera"`
	Position      [2]uint32           `json:"position"`
	Owner         uint32              `json:"owner_return"`
	Game          uint32              `json:"game_return"`
	Tracks        uint32              `json:"tracks_return"`
	Ray           uint32              `json:"ray_return"`
	Send          uint32              `json:"send_return"`
	Special       uint32              `json:"special_return"`
	Mutations     map[string][][2]any `json:"mutations"`
	Sequence      []json.RawMessage   `json:"sequence"`
	NilRecipient  bool                `json:"nil_recipient"`
	NilUpdate     bool                `json:"nil_update"`
	NilPlayer     bool                `json:"nil_player"`
	NilObject     bool                `json:"nil_object"`
}
type dispatchOutput519410 struct {
	Return byte             `json:"return_byte"`
	Events []map[string]any `json:"events"`
	State  map[string]any   `json:"state"`
	Writes [][3]uint32      `json:"heap_writes"`
	Guards bool             `json:"all_other_heap_bytes_nonvolatile_and_FP_preserved"`
	Fault  json.RawMessage  `json:"fault"`
}
type dispatchRow519410 struct {
	Name   string                 `json:"name"`
	Input  dispatchInput519410    `json:"input"`
	Output []dispatchOutput519410 `json:"output"`
}
type dispatchGuard519410[T any] struct {
	Before [64]byte
	Value  T
	After  [64]byte
}
type dispatchRegion519410 struct {
	Pointer unsafe.Pointer
	Size    uintptr
}
type dispatchFixture519410 struct {
	t                   *testing.T
	recipient, object   *dispatchGuard519410[server.Object]
	update, otherUpdate *dispatchGuard519410[server.PlayerUpdateData]
	player, otherPlayer *dispatchGuard519410[server.Player]
	regions             []dispatchRegion519410
	input               dispatchInput519410
	runtime             server.ObjectDispatchNativeRuntime519410
	events              []map[string]any
}

func dispatchAllocate519410[T comparable](f *dispatchFixture519410) *dispatchGuard519410[T] {
	g, free := alloc.NewHigh(dispatchGuard519410[T]{})
	f.t.Cleanup(free)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(g)) <= math.MaxUint32 {
		f.t.Fatal("native allocation must exceed4GiB")
	}
	for i := range g.Before {
		g.Before[i] = 0x6d
		g.After[i] = 0xf3
	}
	f.regions = append(f.regions, dispatchRegion519410{unsafe.Pointer(g), unsafe.Sizeof(*g)})
	return g
}
func dispatchCanonical519410(value any) any {
	b, e := json.Marshal(value)
	if e != nil {
		panic(e)
	}
	var out any
	if e = json.Unmarshal(b, &out); e != nil {
		panic(e)
	}
	return out
}
func dispatchNewFixture519410(t *testing.T) *dispatchFixture519410 {
	f := &dispatchFixture519410{t: t}
	f.recipient = dispatchAllocate519410[server.Object](f)
	f.object = dispatchAllocate519410[server.Object](f)
	f.update = dispatchAllocate519410[server.PlayerUpdateData](f)
	f.otherUpdate = dispatchAllocate519410[server.PlayerUpdateData](f)
	f.player = dispatchAllocate519410[server.Player](f)
	f.otherPlayer = dispatchAllocate519410[server.Player](f)
	checkObject := func(unit *server.Object) {
		if unit != f.unit() {
			f.t.Fatal("child lost native target pointer")
		}
	}
	checkRecipient := func(recipient *server.Object) {
		if recipient != &f.recipient.Value {
			f.t.Fatal("child lost native recipient pointer")
		}
	}
	checkIndex := func(index int32) {
		if index != int32(f.input.Index) {
			f.t.Fatalf("cached player index=%d want%d", index, f.input.Index)
		}
	}
	indexed := func(kind string, index int32, unit *server.Object) uint32 {
		checkIndex(index)
		checkObject(unit)
		return f.event(kind, []any{index, "object"})
	}
	paired := func(kind string, recipient, unit *server.Object) uint32 {
		checkRecipient(recipient)
		checkObject(unit)
		return f.event(kind, []any{"recipient", "object"})
	}
	f.runtime = server.ObjectDispatchNativeRuntime519410{
		Friend: func(index int32, unit *server.Object, add uint32) uint32 {
			checkIndex(index)
			checkObject(unit)
			return f.event("friend", []any{index, "object", add})
		},
		Owner: func(unit, recipient *server.Object) uint32 {
			checkObject(unit)
			checkRecipient(recipient)
			return f.event("owner", []any{"object", "recipient"})
		},
		Game: func(flag uint32) uint32 {
			if flag != 512 {
				f.t.Fatal("wrong game flag")
			}
			return f.event("game", []any{flag})
		},
		Tracks: func(index int32, unit *server.Object) uint32 { return indexed("tracks", index, unit) },
		Ray: func(bits [4]uint32, zero1, zero2, flags int32) uint32 {
			return f.event("ray", []any{bits, zero1, zero2, flags})
		},
		Sprite: func(recipient *server.Object, index int32, unit *server.Object) uint32 {
			checkRecipient(recipient)
			checkIndex(index)
			checkObject(unit)
			return f.event("sprite", []any{"recipient", index, "object"})
		},
		Complex: func(index int32, unit *server.Object, report int32) uint32 {
			checkIndex(index)
			checkObject(unit)
			return f.event("complex", []any{index, "object", report})
		},
		Player: func(recipient, unit *server.Object, stream, health int32) uint32 {
			checkRecipient(recipient)
			checkObject(unit)
			return f.event("player", []any{"recipient", "object", stream, health})
		},
		Phantom: func(index int32, unit *server.Object) uint32 { return indexed("phantom", index, unit) },
		Simple:  func(index int32, unit *server.Object) uint32 { return indexed("simple", index, unit) },
		Special: func(recipient, unit *server.Object) uint32 { return paired("special", recipient, unit) },
		Out:     func(index int32, unit *server.Object) uint32 { return indexed("out", index, unit) },
		Shadow:  func(index int32, unit *server.Object) uint32 { return indexed("shadow", index, unit) },
	}
	old := objectDispatchRuntimeNative519410
	objectDispatchRuntimeNative519410 = func() server.ObjectDispatchNativeRuntime519410 { return f.runtime }
	t.Cleanup(func() { objectDispatchRuntimeNative519410 = old })
	return f
}
func (f *dispatchFixture519410) unit() *server.Object {
	if f.input.Same {
		return &f.recipient.Value
	}
	return &f.object.Value
}
func (f *dispatchFixture519410) change(field string, value uint32) {
	u := f.unit()
	switch field {
	case "object_class":
		u.ObjClass = object.Class(value)
	case "flags":
		u.ObjFlags = object.Flags(value)
	case "field5":
		u.Field5 = value
	case "friend_pending":
		u.Field35 = value
	case "friend_value":
		u.Field36 = value
	case "seen":
		u.Field37 = value
	case "pending":
		u.Field38 = value
	case "state":
		u.Field140[f.input.Index] = value
	case "position_x":
		u.PosVec.X = math.Float32frombits(value)
	case "cached_update_player":
		f.update.Value.Player = &f.player.Value
		if value != 0 {
			f.update.Value.Player = &f.otherPlayer.Value
		}
	case "recipient_update":
		f.recipient.Value.UpdateData = unsafe.Pointer(&f.update.Value)
		if value != 0 {
			f.recipient.Value.UpdateData = unsafe.Pointer(&f.otherUpdate.Value)
		}
	default:
		f.t.Fatal("unknown original mutation", field)
	}
}

var dispatchServices519410 = map[string]string{"friend": "0x4d97a0", "owner": "0x4ec4f0", "game": "0x40a5c0", "tracks": "0x4173d0", "ray": "0x535250", "sprite": "0x518ae0", "complex": "0x518960", "player": "0x518c30", "phantom": "0x5187e0", "simple": "0x5188a0", "special": "0x527e50", "out": "0x528a60", "shadow": "0x528a90"}

func (f *dispatchFixture519410) event(kind string, args []any) uint32 {
	mutations := f.input.Mutations[kind]
	if mutations == nil {
		mutations = [][2]any{}
	}
	for _, m := range mutations {
		f.change(m[0].(string), uint32(m[1].(float64)))
	}
	var result uint32
	switch kind {
	case "owner":
		result = f.input.Owner
	case "game":
		result = f.input.Game
	case "tracks":
		result = f.input.Tracks
	case "ray":
		result = f.input.Ray
	case "special":
		result = f.input.Special
	case "sprite", "complex", "player", "phantom", "simple":
		result = f.input.Send
	default:
		result = 0x76543210
	}
	f.events = append(f.events, map[string]any{"service": kind, "address": dispatchServices519410[kind], "arguments": args, "return_word": result, "mutations": mutations})
	return result
}
func (f *dispatchFixture519410) reset(input dispatchInput519410) {
	f.input = input
	f.events = []map[string]any{}
	f.recipient.Value = server.Object{}
	f.object.Value = server.Object{}
	f.update.Value = server.PlayerUpdateData{}
	f.otherUpdate.Value = server.PlayerUpdateData{}
	f.player.Value = server.Player{}
	f.otherPlayer.Value = server.Player{}
	f.recipient.Value.ObjClass = object.ClassPlayer
	f.recipient.Value.UpdateData = unsafe.Pointer(&f.update.Value)
	if input.NilUpdate {
		f.recipient.Value.UpdateData = nil
	}
	f.update.Value.Player = &f.player.Value
	if input.NilPlayer {
		f.update.Value.Player = nil
	}
	f.otherUpdate.Value.Player = &f.otherPlayer.Value
	f.player.Value.PlayerInd = input.Index
	f.otherPlayer.Value.PlayerInd = 11
	f.player.Value.Pos3632Vec = types.Pointf{X: math.Float32frombits(input.Camera[0]), Y: math.Float32frombits(input.Camera[1])}
	f.otherPlayer.Value.Pos3632Vec = types.Pointf{X: math.Float32frombits(0x43aa0000), Y: math.Float32frombits(0x43880000)}
	for field, value := range map[string]uint32{"object_class": input.Class, "flags": input.Flags, "field5": input.Field5, "friend_pending": input.FriendPending, "friend_value": input.FriendValue, "seen": input.Seen, "pending": input.Pending, "state": input.State} {
		f.change(field, value)
	}
	f.unit().PosVec = types.Pointf{X: math.Float32frombits(input.Position[0]), Y: math.Float32frombits(input.Position[1])}
}
func (f *dispatchFixture519410) snapshot() [][]byte {
	var copies [][]byte
	for _, r := range f.regions {
		copies = append(copies, append([]byte(nil), unsafe.Slice((*byte)(r.Pointer), int(r.Size))...))
	}
	return copies
}
func (f *dispatchFixture519410) check(want [][]byte) {
	for i, r := range f.regions {
		if !bytes.Equal(want[i], unsafe.Slice((*byte)(r.Pointer), int(r.Size))) {
			f.t.Fatalf("complete native allocation or neighbor guard changed beyond original writes: region%d", i)
		}
	}
}
func (f *dispatchFixture519410) expectedWrites(want [][]byte, writes [][3]uint32) {
	u := f.unit()
	base := uint32(0x2000)
	if f.input.Same {
		base = 0x1000
	}
	for _, w := range writes {
		if w[1] != 4 {
			f.t.Fatal("unexpected original write width", w)
		}
		var target unsafe.Pointer
		value := uint64(w[2])
		size := 4
		switch {
		case w[0] == 0x1000+748:
			target = unsafe.Pointer(&f.recipient.Value.UpdateData)
			value = uint64(uintptr(unsafe.Pointer(&f.update.Value)))
			if w[2] == 0x21005000 {
				value = uint64(uintptr(unsafe.Pointer(&f.otherUpdate.Value)))
			} else if w[2] != 0x21004000 {
				f.t.Fatal("unknown original update pointer")
			}
			size = int(unsafe.Sizeof(uintptr(0)))
		case w[0] == 0x4000+276:
			target = unsafe.Pointer(&f.update.Value.Player)
			value = uint64(uintptr(unsafe.Pointer(&f.player.Value)))
			if w[2] == 0x21008000 {
				value = uint64(uintptr(unsafe.Pointer(&f.otherPlayer.Value)))
			} else if w[2] != 0x21006000 {
				f.t.Fatal("unknown original player pointer")
			}
			size = int(unsafe.Sizeof(uintptr(0)))
		case w[0] >= base+560 && w[0] < base+688:
			target = unsafe.Pointer(&u.Field140[(w[0]-base-560)/4])
		default:
			switch w[0] - base {
			case 8:
				target = unsafe.Pointer(&u.ObjClass)
			case 16:
				target = unsafe.Pointer(&u.ObjFlags)
			case 20:
				target = unsafe.Pointer(&u.Field5)
			case 56:
				target = unsafe.Pointer(&u.PosVec.X)
			case 140:
				target = unsafe.Pointer(&u.Field35)
			case 144:
				target = unsafe.Pointer(&u.Field36)
			case 148:
				target = unsafe.Pointer(&u.Field37)
			case 152:
				target = unsafe.Pointer(&u.Field38)
			default:
				f.t.Fatal("unmapped original write", w)
			}
		}
		found := false
		for i, r := range f.regions {
			address, start := uintptr(target), uintptr(r.Pointer)
			if address >= start && address+uintptr(size) <= start+r.Size {
				var b [8]byte
				binary.LittleEndian.PutUint64(b[:], value)
				copy(want[i][address-start:], b[:size])
				found = true
				break
			}
		}
		if !found {
			f.t.Fatal("expected native write outside allocations")
		}
	}
}
func (f *dispatchFixture519410) state() map[string]any {
	u := f.unit()
	labelUpdate := func(p unsafe.Pointer) any {
		switch p {
		case nil:
			return nil
		case unsafe.Pointer(&f.update.Value):
			return "update"
		case unsafe.Pointer(&f.otherUpdate.Value):
			return "other_update"
		default:
			f.t.Fatal("unrecognized native update")
			return nil
		}
	}
	labelPlayer := func(p *server.Player) any {
		switch p {
		case nil:
			return nil
		case &f.player.Value:
			return "player"
		case &f.otherPlayer.Value:
			return "other_player"
		default:
			f.t.Fatal("unrecognized native player")
			return nil
		}
	}
	return map[string]any{"fields": map[string]uint32{"object_class": uint32(u.ObjClass), "flags": uint32(u.ObjFlags), "field5": u.Field5, "position_x": math.Float32bits(u.PosVec.X), "friend_pending": u.Field35, "friend_value": u.Field36, "seen": u.Field37, "pending": u.Field38}, "state_words": u.Field140, "recipient_update": labelUpdate(f.recipient.Value.UpdateData), "cached_update_player": labelPlayer(f.update.Value.Player)}
}
func (f *dispatchFixture519410) invoke(native bool) (byte, bool) {
	recipient, unit := &f.recipient.Value, f.unit()
	if f.input.NilRecipient {
		recipient = nil
	}
	if f.input.NilObject {
		unit = nil
	}
	if native {
		return objectDispatchPublicC519410(recipient, unit), false
	}
	return server.ObjectDispatchNative519410(recipient, unit, f.runtime)
}
func dispatchLoad519410(t *testing.T) ([]dispatchRow519410, []dispatchRow519410) {
	t.Helper()
	path := os.Getenv("NOX_OBJECT_DISPATCH_ORIGINAL_FIXTURE")
	if path == "" {
		path = filepath.Join("..", "server", "testdata", "object_update_dispatch_original_519410.json.gz")
	}
	file, e := os.Open(path)
	if e != nil {
		t.Fatal(e)
	}
	defer file.Close()
	z, e := gzip.NewReader(file)
	if e != nil {
		t.Fatal(e)
	}
	defer z.Close()
	b, e := io.ReadAll(z)
	if e != nil {
		t.Fatal(e)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(b)); got != "103d8a315d8b0fb5e25c66b81798fcedefc1a8f198d4a60a8de14f3020588638" {
		t.Fatal("sealed original fixture hash", got)
	}
	var rows struct {
		Contexts []dispatchRow519410 `json:"contexts"`
		Faults   []dispatchRow519410 `json:"expected_fault_contexts"`
	}
	if e = json.Unmarshal(b, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows.Contexts) != 5918 || len(rows.Faults) != 4 {
		t.Fatal("incomplete original contexts")
	}
	return rows.Contexts, rows.Faults
}
func dispatchRunOriginal519410(t *testing.T, native bool) {
	rows, faults := dispatchLoad519410(t)
	f := dispatchNewFixture519410(t)
	for index, row := range rows {
		t.Run(fmt.Sprintf("%06d/%s", index, row.Name), func(t *testing.T) {
			f.t = t
			f.reset(row.Input)
			steps := append([]json.RawMessage{json.RawMessage(`{}`)}, row.Input.Sequence...)
			if len(steps) != len(row.Output) {
				t.Fatal("incomplete original repeated calls")
			}
			for call, changes := range steps {
				if e := json.Unmarshal(changes, &f.input); e != nil {
					t.Fatal(e)
				}
				var fields map[string]uint32
				if e := json.Unmarshal(changes, &fields); e != nil {
					t.Fatal(e)
				}
				for field, value := range fields {
					switch field {
					case "object_class", "flags", "field5", "friend_pending", "friend_value", "seen", "pending", "state", "position_x":
						f.change(field, value)
					}
				}
				want := row.Output[call]
				before := f.snapshot()
				f.events = []map[string]any{}
				value, sent := f.invoke(native)
				f.expectedWrites(before, want.Writes)
				f.check(before)
				wantSent := false
				for _, event := range want.Events {
					switch event["service"] {
					case "sprite", "complex", "player", "phantom", "simple":
						wantSent = event["return_word"].(float64) != 0
					}
				}
				if !want.Guards || value != want.Return || (!native && sent != wantSent) || !reflect.DeepEqual(dispatchCanonical519410(f.events), dispatchCanonical519410(want.Events)) || !reflect.DeepEqual(dispatchCanonical519410(f.state()), dispatchCanonical519410(want.State)) {
					t.Fatalf("native%t call%d return%#x want%#x sent%t want%t\nevents=%#v\noriginal=%#v\nstate=%#v\noriginal=%#v", native, call, value, want.Return, sent, wantSent, f.events, want.Events, f.state(), want.State)
				}
			}
		})
	}
	for index, row := range faults {
		t.Run(fmt.Sprintf("fault/%d/%s", index, row.Name), func(t *testing.T) {
			f.t = t
			f.reset(row.Input)
			before := f.snapshot()
			panicked := func() (panicked bool) { defer func() { panicked = recover() != nil }(); f.invoke(native); return false }()
			f.check(before)
			if !panicked || !reflect.DeepEqual(dispatchCanonical519410(f.events), dispatchCanonical519410(row.Output[0].Events)) {
				t.Fatalf("original pointer fault not preserved;panicked%t events%#v", panicked, f.events)
			}
		})
	}
}
func TestObjectDispatchOriginalNative519410(t *testing.T) { dispatchRunOriginal519410(t, true) }
func TestObjectDispatchOriginalGo519410(t *testing.T)     { dispatchRunOriginal519410(t, false) }
