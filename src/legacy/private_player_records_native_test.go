package legacy

import (
	"bytes"
	_ "embed"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/strman"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

//go:embed testdata/private_player_records_original_4259c0_425be0_v2.json.gz
var privatePlayerRecordsOriginal []byte

type privateRecordsBlock struct {
	Label  string `json:"label"`
	Size   int    `json:"size"`
	Freed  bool   `json:"freed"`
	RawHex string `json:"raw_hex"`
}

type privateRecordsSnapshot struct {
	RootHex string                `json:"root_hex"`
	Blocks  []privateRecordsBlock `json:"blocks"`
}

type privateRecordsService struct {
	Service   string `json:"service"`
	Pointer   string `json:"pointer"`
	Size      int    `json:"size"`
	Fill      byte   `json:"fill"`
	BeforeHex string `json:"before_hex"`
}

type privateRecordsOp struct {
	Kind string `json:"kind"`
	Args []struct {
		Alias string `json:"alias"`
		Value uint32 `json:"value"`
	} `json:"args"`
	Alias      string                  `json:"alias"`
	ResultType string                  `json:"result_type"`
	Result     json.RawMessage         `json:"result"`
	Before     privateRecordsSnapshot  `json:"before"`
	After      privateRecordsSnapshot  `json:"after"`
	Services   []privateRecordsService `json:"services"`
	Writes     []struct {
		Owner  string `json:"owner"`
		Offset int    `json:"offset"`
		Bytes  int    `json:"bytes"`
		Value  uint64 `json:"value"`
	} `json:"writes"`
}

type privateRecordsSequence struct {
	Name           string             `json:"name"`
	InitialFlag    uint32             `json:"initial_flag"`
	AllocationFill byte               `json:"allocation_fill"`
	Names          [][]uint16         `json:"names_UTF16"`
	Operations     []privateRecordsOp `json:"operations"`
}

var privateRecordsModes = map[string]int{"init": 0, "free_all": 1, "first": 2, "next": 3, "find": 4, "exists": 5, "create": 6, "add_child": 7, "remove_child": 8, "first_child": 9, "next_child": 10}

func privateRecordsCases(t *testing.T) []privateRecordsSequence {
	t.Helper()
	var rows []privateRecordsSequence
	npcVoiceOriginalDecode(t, privatePlayerRecordsOriginal, "90182dfb88081235cd64c5045651584421ddd17641bf5a51a98dd4885083f265", "c87515b1e95aa4f013a2aaa49386f2c91f4dea887b75f553ada821545b671a69", 63, &rows)
	return rows
}

// Verify the reference's raw writes independently of the native normalization.
func privateRecordsVerifyOriginal(t *testing.T, seq privateRecordsSequence, op privateRecordsOp) {
	t.Helper()
	owners := map[string][]byte{"root": npcVoiceUnhex(t, op.Before.RootHex)}
	freed := make(map[string]bool)
	for _, b := range op.Before.Blocks {
		owners[b.Label], freed[b.Label] = npcVoiceUnhex(t, b.RawHex), b.Freed
	}
	for _, e := range op.Services {
		switch e.Service {
		case "malloc":
			if owners[e.Pointer] != nil || e.Fill != seq.AllocationFill || (e.Size != 52 && e.Size != 16) {
				t.Fatal("invalid sealed malloc service", e)
			}
			owners[e.Pointer] = bytes.Repeat([]byte{e.Fill}, e.Size)
		case "free":
			if owners[e.Pointer] == nil || freed[e.Pointer] {
				t.Fatal("invalid sealed free service", e)
			}
			freed[e.Pointer] = true
		default:
			t.Fatal("unknown sealed service", e)
		}
	}
	for _, w := range op.Writes {
		owner := owners[w.Owner]
		if w.Offset < 0 || w.Bytes < 1 || w.Bytes > 8 || w.Offset+w.Bytes > len(owner) {
			t.Fatal("sealed write escaped declared original owner", w)
		}
		for i := 0; i < w.Bytes; i++ {
			owner[w.Offset+i] = byte(w.Value >> (8 * i))
		}
	}
	if !bytes.Equal(owners["root"], npcVoiceUnhex(t, op.After.RootHex)) || len(owners) != len(op.After.Blocks)+1 {
		t.Fatal("sealed root/allocation transcript differs")
	}
	for _, b := range op.After.Blocks {
		if len(owners[b.Label]) != b.Size || !bytes.Equal(owners[b.Label], npcVoiceUnhex(t, b.RawHex)) || freed[b.Label] != b.Freed {
			t.Fatal("sealed whole block/write/free transcript differs", b.Label)
		}
	}
}

type privateRecordsArena struct {
	raw       []byte
	placement int
	width     int
	seq       privateRecordsSequence
	labels    map[string]int
	aliases   map[string]unsafe.Pointer
	allocated int
	freed     map[string]bool
	services  []privateRecordsService
	serviceAt int
	packed    []byte
	packedOld []byte
}

func privateRecordsArenaNew(t *testing.T, placement int, seq privateRecordsSequence) *privateRecordsArena {
	t.Helper()
	raw, free := alloc.MakeHigh([]byte{}, 8192)
	t.Cleanup(free)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(&raw[0])) <= math.MaxUint32 {
		t.Fatal("C list owners must exercise foreign addresses above4GiB")
	}
	v := &privateRecordsArena{raw: raw, placement: placement, width: int(unsafe.Sizeof(uintptr(0))), seq: seq, labels: map[string]int{"root": 64 + placement}, aliases: make(map[string]unsafe.Pointer), freed: make(map[string]bool)}
	for i := 0; i < 20; i++ {
		v.labels[fmt.Sprintf("block%03d", i)] = 1024 + 256*i + placement
	}
	for i := range seq.Names {
		v.aliases[fmt.Sprintf("name%d", i)] = unsafe.Pointer(&raw[256+128*i+placement])
	}
	v.packed = unsafe.Slice((*byte)(memmap.PtrOff(0x5D4594, 599440)), 64)
	v.packedOld = bytes.Clone(v.packed)
	old := privatePlayerRecords4259C0
	t.Cleanup(func() { privatePlayerRecords4259C0 = old; copy(v.packed, v.packedOld) })
	for i := range v.packed {
		v.packed[i] = 0xa5
	}
	copy(v.raw, v.expected(t, seq.Operations[0].Before))
	binary.LittleEndian.PutUint32(v.packed[32:], seq.InitialFlag)
	privatePlayerRecords4259C0 = privatePlayerRecordStore4259C0{head: (*intrusiveListNode425760)(v.pointer("root")), allocate: v.allocate(t), release: v.release(t)}
	return v
}

func (v *privateRecordsArena) pointer(label string) unsafe.Pointer {
	return unsafe.Pointer(&v.raw[v.labels[label]])
}

func (v *privateRecordsArena) originalPointer(t *testing.T, address uint32) uintptr {
	t.Helper()
	if address == 0 {
		return 0
	}
	if address == 0x666b38 {
		return uintptr(v.pointer("root"))
	}
	for i := 0; i < 20; i++ {
		base := uint32(0x21001000 + 256*i)
		if address == base || address == base+40 {
			p := uintptr(v.pointer(fmt.Sprintf("block%03d", i)))
			if address != base {
				p += unsafe.Offsetof(privatePlayerRecord425AD0{}.children)
			}
			return p
		}
	}
	t.Fatalf("unmapped original pointer %#x", address)
	return 0
}

func (v *privateRecordsArena) header(t *testing.T, dst, source []byte) {
	t.Helper()
	for i := 0; i < 3; i++ {
		value := v.originalPointer(t, binary.LittleEndian.Uint32(source[4*i:]))
		if v.width == 8 {
			binary.LittleEndian.PutUint64(dst[8*i:], uint64(value))
		} else {
			binary.LittleEndian.PutUint32(dst[4*i:], uint32(value))
		}
	}
}

func (v *privateRecordsArena) block(t *testing.T, original []byte) []byte {
	t.Helper()
	size := int(unsafe.Sizeof(privatePlayerChild425B30{}))
	if len(original) == 52 {
		size = int(unsafe.Sizeof(privatePlayerRecord425AD0{}))
	} else if len(original) != 16 {
		t.Fatal("original block size", len(original))
	}
	out := bytes.Repeat([]byte{v.seq.AllocationFill}, size)
	v.header(t, out, original)
	if len(original) == 16 {
		copy(out[unsafe.Offsetof(privatePlayerChild425B30{}.peer):], original[12:16])
	} else {
		copy(out[unsafe.Offsetof(privatePlayerRecord425AD0{}.name):], original[12:32])
		copy(out[unsafe.Offsetof(privatePlayerRecord425AD0{}.id):], original[32:36])
		if binary.LittleEndian.Uint32(original[36:40]) != 0 {
			t.Fatal("original team initialization differs from zero")
		}
		clear(out[unsafe.Offsetof(privatePlayerRecord425AD0{}.team):][:v.width])
		v.header(t, out[unsafe.Offsetof(privatePlayerRecord425AD0{}.children):], original[40:52])
	}
	return out
}

func (v *privateRecordsArena) expected(t *testing.T, snap privateRecordsSnapshot) []byte {
	t.Helper()
	out := bytes.Repeat([]byte{0xcc}, len(v.raw))
	v.header(t, out[v.labels["root"]:], npcVoiceUnhex(t, snap.RootHex))
	for i, name := range v.seq.Names {
		pos := 256 + 128*i + v.placement
		for j, unit := range append(append([]uint16(nil), name...), 0) {
			binary.LittleEndian.PutUint16(out[pos+2*j:], unit)
		}
	}
	for _, b := range snap.Blocks {
		copy(out[v.labels[b.Label]:], v.block(t, npcVoiceUnhex(t, b.RawHex)))
	}
	return out
}

func (v *privateRecordsArena) nextService(t *testing.T, kind string) privateRecordsService {
	t.Helper()
	if v.serviceAt >= len(v.services) || v.services[v.serviceAt].Service != kind {
		t.Fatal("actual allocation/free call order differs", kind, v.serviceAt, v.services)
	}
	e := v.services[v.serviceAt]
	v.serviceAt++
	return e
}

func (v *privateRecordsArena) allocate(t *testing.T) func(uintptr) unsafe.Pointer {
	return func(size uintptr) unsafe.Pointer {
		t.Helper()
		e := v.nextService(t, "malloc")
		want := unsafe.Sizeof(privatePlayerChild425B30{})
		if e.Size == 52 {
			want = unsafe.Sizeof(privatePlayerRecord425AD0{})
		}
		if size != want || e.Pointer != fmt.Sprintf("block%03d", v.allocated) || e.Fill != v.seq.AllocationFill {
			t.Fatal("actual allocation size/identity/fill differs", size, e)
		}
		v.allocated++
		p := v.pointer(e.Pointer)
		copy(unsafe.Slice((*byte)(p), size), bytes.Repeat([]byte{e.Fill}, int(size)))
		return p
	}
}

func (v *privateRecordsArena) release(t *testing.T) func(unsafe.Pointer) {
	return func(p unsafe.Pointer) {
		t.Helper()
		e := v.nextService(t, "free")
		want := v.block(t, npcVoiceUnhex(t, e.BeforeHex))
		if p != v.pointer(e.Pointer) || v.freed[e.Pointer] || !bytes.Equal(unsafe.Slice((*byte)(p), len(want)), want) {
			t.Fatal("actual free identity/order/whole pre-free owner differs", e.Pointer)
		}
		v.freed[e.Pointer] = true
		// Keep provider storage observable after logical free, as the CPU
		// reference does. Default CRT release is exercised separately below.
	}
}

func (v *privateRecordsArena) run(t *testing.T, op privateRecordsOp) {
	t.Helper()
	if !bytes.Equal(v.raw, v.expected(t, op.Before)) {
		t.Fatal("whole native owner/input/canaries differ before C call")
	}
	v.services, v.serviceAt = op.Services, 0
	var p unsafe.Pointer
	var value int32
	var name *uint16
	for i, arg := range op.Args {
		if arg.Alias != "" {
			ptr, ok := v.aliases[arg.Alias]
			if !ok {
				t.Fatal("unknown input alias", arg.Alias)
			}
			if op.Kind == "create" && i == 1 {
				name = (*uint16)(ptr)
			} else {
				p = ptr
			}
		} else {
			value = int32(arg.Value)
		}
	}
	got := privatePlayerRecordsCallC(privateRecordsModes[op.Kind], p, value, name)
	switch op.ResultType {
	case "pointer":
		var result struct {
			Pointer *string `json:"pointer"`
			Offset  int     `json:"offset"`
		}
		if err := json.Unmarshal(op.Result, &result); err != nil {
			t.Fatal(err)
		}
		var want unsafe.Pointer
		if result.Pointer != nil {
			if result.Offset != 0 {
				t.Fatal("unexpected reference return offset", result)
			}
			want = v.pointer(*result.Pointer)
		}
		if got != uintptr(want) {
			t.Fatal("actual C result differs from original pointer identity", got, want)
		}
		if op.Alias != "" {
			v.aliases[op.Alias] = want
		}
	case "int32":
		var want int32
		if err := json.Unmarshal(op.Result, &want); err != nil || int32(got) != want {
			t.Fatal("actual C scalar result differs", got, want, err)
		}
	case "void":
		if string(op.Result) != "null" {
			t.Fatal("incidental original EAX must not become a golden")
		}
		// Retain current decoded-C pointer compatibility independently of
		// the deliberately ignored allocator-dependent original EAX.
		if op.Kind == "remove_child" && got != uintptr(unsafe.Add(p, unsafe.Offsetof(privatePlayerRecord425AD0{}.children))) {
			t.Fatal("current C remove-child return lost its native head pointer")
		}
		if op.Kind == "free_all" {
			original := binary.LittleEndian.Uint32(npcVoiceUnhex(t, op.Before.RootHex))
			var want uintptr
			if original != 0 && original != 0x666b38 {
				want = v.originalPointer(t, original)
			}
			if got != want {
				t.Fatal("current C free-all return lost its native first pointer", got, want)
			}
		}
	default:
		t.Fatal("unknown original return type", op.ResultType)
	}
	if v.serviceAt != len(op.Services) || !bytes.Equal(v.raw, v.expected(t, op.After)) {
		t.Fatal("actual C whole graph/payload/padding/input/canaries or service count differs")
	}
	wantPacked := bytes.Repeat([]byte{0xa5}, 64)
	copy(wantPacked[32:36], npcVoiceUnhex(t, op.After.RootHex)[12:16])
	if !bytes.Equal(v.packed, wantPacked) {
		t.Fatal("native header overwrote retired packed root, raw flag or neighbors")
	}
	for _, b := range op.After.Blocks {
		if v.freed[b.Label] != b.Freed {
			t.Fatal("actual logical release state differs", b.Label)
		}
	}
}

func TestPrivatePlayerRecordsNativeLayoutCABI(t *testing.T) {
	want := [7]uintptr{52, 16, 12, 32, 36, 40, 12}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		want = [7]uintptr{80, 32, 24, 44, 48, 56, 24}
	}
	got := [7]uintptr{unsafe.Sizeof(privatePlayerRecord425AD0{}), unsafe.Sizeof(privatePlayerChild425B30{}), unsafe.Offsetof(privatePlayerRecord425AD0{}.name), unsafe.Offsetof(privatePlayerRecord425AD0{}.id), unsafe.Offsetof(privatePlayerRecord425AD0{}.team), unsafe.Offsetof(privatePlayerRecord425AD0{}.children), unsafe.Offsetof(privatePlayerChild425B30{}.peer)}
	if got != want || privatePlayerRecordsLayoutC() != want {
		t.Fatal("independent32/64bit C/Go owner layout", got, privatePlayerRecordsLayoutC(), want)
	}
}

func TestPrivatePlayerRecordsNativeOriginalCABI(t *testing.T) {
	rows := privateRecordsCases(t)
	for _, row := range rows {
		if len(row.Operations) != 52 {
			t.Fatal("sealed sequence operation count")
		}
		for _, op := range row.Operations {
			privateRecordsVerifyOriginal(t, row, op)
		}
	}
	for _, placement := range []int{0, 8, 24} {
		for _, row := range rows {
			t.Run(fmt.Sprintf("placement-%d/%s", placement, row.Name), func(t *testing.T) {
				v := privateRecordsArenaNew(t, placement, row)
				for i, op := range row.Operations {
					t.Run(fmt.Sprintf("%02d-%s", i, op.Kind), func(t *testing.T) { v.run(t, op) })
				}
				runtime.GC()
			})
		}
	}
	t.Log("63sequences/3276original operations, three foreign placements;655unchanged bytes; original54sequence prefix retained; only declared allocation/free services; root+flag16byte image exception, all other original PE bytes immutable")
}

func privateRecordsDefaultStore(t *testing.T) {
	t.Helper()
	old := privatePlayerRecords4259C0
	flag := memmap.PtrUint32(0x5D4594, 599472)
	oldFlag := *flag
	head, release := alloc.New(intrusiveListNode425760{})
	privatePlayerRecords4259C0 = privatePlayerRecordStore4259C0{head: head}
	*flag = 0
	t.Cleanup(func() {
		privatePlayerRecords4259C0.freeAll()
		release()
		privatePlayerRecords4259C0 = old
		*flag = oldFlag
	})
}

func TestPrivatePlayerRecordsNativeDefaultLifecycleCABI(t *testing.T) {
	privateRecordsDefaultStore(t)
	for cycle := 0; cycle < 3; cycle++ {
		t.Run(fmt.Sprint(cycle), func(t *testing.T) {
			Sub_4259C0()
			name, freeName := alloc.CString16("Guide")
			defer freeName()
			// Use the actual Go registration wrappers and their C entries.
			p := Sub_425A70(0x12345678)
			if p != nil {
				t.Fatal("previous stop/start left a player record")
			}
			p = Sub_425AD0(0x12345678, name)
			if p == nil || Sub_425A70(0x12345678) != p {
				t.Fatal("default calloc registration or lookup")
			}
			Sub_425B30(p, 0)
			Sub_425B30(p, 31)
			privatePlayerRecordsCallC(8, p, 0, nil)
			if privatePlayerRecordsCallC(4, nil, 0x12345678, nil) != uintptr(p) {
				t.Fatal("first disconnect removed a still-used record")
			}
			privatePlayerRecordsCallC(8, p, 31, nil)
			if Sub_425A70(0x12345678) != nil || alloc.TryFreePtr(p) {
				t.Fatal("last C disconnect did not release its record")
			}
			p = Sub_425AD0(-7, name)
			Sub_425B30(p, 4)
			Sub_4259F0()
			if Sub_425A70(-7) != nil || alloc.TryFreePtr(p) || memmap.Uint32(0x5D4594, 599472) != 1 {
				t.Fatal("server teardown did not release/reuse initialized list")
			}
			runtime.GC()
		})
	}
}

func TestPrivatePlayerRecordsNativeFullPlayerNameCABI(t *testing.T) {
	privateRecordsDefaultStore(t)
	Sub_4259C0()
	s := server.NewHigh(nil, nil, strman.New())
	defer s.Close()
	pl := s.Players.ResetInd(0)
	adjacent := unsafe.Slice((*byte)(unsafe.Pointer(&pl.Active)), 4)
	for i := range adjacent {
		adjacent[i] = 0xa5
	}
	pl = s.Players.ResetInd(0)
	if !bytes.Equal(adjacent, []byte{1, 0, 0, 0}) || unsafe.Offsetof(pl.Active)-unsafe.Offsetof(pl.Field2072) != 20 {
		t.Fatal("actual Player.ResetInd adjacent-field termination contract", adjacent)
	}
	alloc.StrCopy16(pl.Field2072[:], "AAAAAAAAAA")
	units := unsafe.Slice(&pl.Field2072[0], 12)
	want := []uint16{65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 1, 0}
	if !reflect.DeepEqual(units, want) {
		t.Fatal("full production field raw UTF16 input", units)
	}
	pl.Field2068 = 0x12345678
	p := Sub_425AD0(int(pl.Field2068), &pl.Field2072[0])
	record := (*privatePlayerRecord425AD0)(p)
	if record.id != 1 || record.name != pl.Field2072 || record.team != nil || Sub_425A70(int(pl.Field2068)) != nil {
		t.Fatal("full field must retain original raw name/ID overlap")
	}
	Sub_425B30(p, 0)
	privatePlayerRecordsCallC(8, p, 0, nil)
	if privatePlayerRecordsCallC(2, nil, 0, nil) != 0 {
		t.Fatal("full-name record did not release")
	}
}

func TestPrivatePlayerRecordsNativeTypedPayloadConsumerCABI(t *testing.T) {
	privateRecordsDefaultStore(t)
	Sub_4259C0()
	name, freeName := alloc.CString16("Team")
	defer freeName()
	p := Sub_425AD0(-23, name)
	team, freeTeam := alloc.MakeHigh([]byte{}, 128)
	defer freeTeam()
	teamPtr := unsafe.Pointer(&team[0])
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(teamPtr) <= math.MaxUint32 {
		t.Fatal("typed C consumer must retain a high team pointer")
	}
	got := privatePlayerRecordsConsumerC(p, teamPtr)
	record := (*privatePlayerRecord425AD0)(p)
	want := [3]uintptr{uintptr(unsafe.Pointer(&record.name[0])), uintptr(uint32(0xffffffe9)), uintptr(teamPtr)}
	if got != want || record.team != teamPtr {
		t.Fatal("native C name/ID/team payload consumer", got, want)
	}
	runtime.GC()
	Sub_4259F0()
}
