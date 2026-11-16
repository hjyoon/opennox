package legacy

import (
	"bytes"
	_ "embed"
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

//go:embed testdata/pointer_chain_original_v1.json.gz
var pointerChainNativeOriginal []byte

type pointerChainNativeCase struct {
	Input struct {
		Index         int
		Entry, Name   string
		Nonnull       bool
		Value, Delta  uint32
		ClassBits     uint32 `json:"class_bits"`
		ObjectFlags   uint32 `json:"object_flags"`
		Gate, Counter uint32
		StatusBits    uint32 `json:"status_bits"`
		TradeRole     int    `json:"trade_role"`
		ListCount     int    `json:"list_count"`
		Codes         [3]uint32
		Query         uint32
	}
	ReturnWord uint32                            `json:"return_u32"`
	Writes     []struct{ Address, Value uint32 } `json:"heap_writes"`
}

func pointerChainNativeCases(t *testing.T) []pointerChainNativeCase {
	t.Helper()
	var cases []pointerChainNativeCase
	npcVoiceOriginalDecode(t, pointerChainNativeOriginal, "2ce2a72f821998c56c2ca847820b8fc1b5ae6b070c9df4883353daf31c174bf0", "5b6ec3d7c85bcfedab14e067a7338ba6728899403c4b6b44231736cc9ccb8f7f", 2037, &cases)
	return cases
}

func pointerChainNativePlayerField(name string) uintptr {
	f, ok := reflect.TypeOf(server.Player{}).FieldByName(name)
	if !ok {
		panic("unknown original Player field")
	}
	return f.Offset
}

func TestPointerChainNativeLayout(t *testing.T) {
	var obj server.Object
	var update server.PlayerUpdateData
	var player server.Player
	var session server.TradeSession
	var node server.TradeItem
	objectPrefix := unsafe.Offsetof(obj.ScriptPickup) + unsafe.Sizeof(obj.ScriptPickup)
	updatePrefix := unsafe.Offsetof(update.QuestWarpGate) + unsafe.Sizeof(update.QuestWarpGate)
	pointer := unsafe.Sizeof(uintptr(0))
	want := [38]uintptr{pointer, objectPrefix, unsafe.Offsetof(obj.ObjClass), 4, unsafe.Offsetof(obj.ObjFlags), 4, unsafe.Offsetof(obj.UpdateData), pointer, updatePrefix, unsafe.Offsetof(update.Player), pointer, unsafe.Offsetof(update.Trade70), pointer, unsafe.Sizeof(player), unsafe.Offsetof(player.Lessons), 4, pointerChainNativePlayerField("field4652"), 4, pointerChainNativePlayerField("field4664"), 4, pointerChainNativePlayerField("field4692"), 4, unsafe.Offsetof(player.Field4792), 4, unsafe.Sizeof(session), unsafe.Offsetof(session.Field0), 4, unsafe.Offsetof(session.Field20), pointer, unsafe.Sizeof(node), unsafe.Offsetof(node.Item0), pointer, unsafe.Offsetof(node.Field8), pointer, unsafe.Offsetof(obj.NetCode), 4, 4, pointer}
	if got := pointerChainNativeLayoutC(); got != want {
		t.Fatal("native pointer chain layouts", got, want)
	}
	if unsafe.Sizeof(obj) != objectPrefix+2*pointer || unsafe.Sizeof(update) < updatePrefix || unsafe.Sizeof(obj) > 0x400 || unsafe.Sizeof(update) > 0x1000 || unsafe.Sizeof(player) > 0x2000 || unsafe.Sizeof(session) > 0x200 || unsafe.Sizeof(node) > 0x100 {
		t.Fatal("native owner/shared prefix bounds")
	}
}

type pointerChainNativeArena struct {
	raw, far, before, expected, farBefore []byte
	owner                                 *server.Object
	update                                *server.PlayerUpdateData
	player                                *server.Player
	session, other                        *server.TradeSession
	nodes                                 [3]*server.TradeItem
	items                                 [3]*server.Object
}

func pointerChainNativeArenaAt(raw, far []byte, placement int) *pointerChainNativeArena {
	a := &pointerChainNativeArena{raw: raw, far: far, before: make([]byte, len(raw)), expected: make([]byte, len(raw)), farBefore: make([]byte, len(far))}
	a.owner = (*server.Object)(unsafe.Pointer(&raw[0x1000+placement]))
	a.update = (*server.PlayerUpdateData)(unsafe.Pointer(&raw[0x2000+placement]))
	a.player = (*server.Player)(unsafe.Pointer(&raw[0x3000+placement]))
	a.session = (*server.TradeSession)(unsafe.Pointer(&raw[0x6000+placement]))
	a.other = (*server.TradeSession)(unsafe.Pointer(&raw[0x6200+placement]))
	if far != nil {
		a.other = (*server.TradeSession)(unsafe.Pointer(&far[0x6000+placement]))
	}
	for i := range a.nodes {
		a.nodes[i] = (*server.TradeItem)(unsafe.Pointer(&raw[0x7000+i*0x100+placement]))
		a.items[i] = (*server.Object)(unsafe.Pointer(&raw[0x8000+i*0x400+placement]))
	}
	return a
}

func (a *pointerChainNativeArena) originalPointer(t *testing.T, word uint32) unsafe.Pointer {
	t.Helper()
	switch word {
	case 0:
		return nil
	case 0x21001000:
		return unsafe.Pointer(a.owner)
	case 0x21002000:
		return unsafe.Pointer(a.update)
	case 0x21003000:
		return unsafe.Pointer(a.player)
	case 0x21006000:
		return unsafe.Pointer(a.session)
	case 0x21006200:
		return unsafe.Pointer(a.other)
	case 0x21008000:
		return unsafe.Pointer(a.items[0])
	case 0x21008400:
		return unsafe.Pointer(a.items[1])
	case 0x21008800:
		return unsafe.Pointer(a.items[2])
	}
	t.Fatalf("unmapped original pointer %#x", word)
	return nil
}

func (a *pointerChainNativeArena) expectWord(t *testing.T, address, value uint32) {
	t.Helper()
	var field unsafe.Pointer
	pointer := false
	switch address {
	case 0x21006000:
		field = unsafe.Pointer(&a.session.Field0)
	case 0x21002000 + 280:
		field = unsafe.Pointer(&a.update.Trade70)
		pointer = true
	case 0x21003000 + 2136:
		field = unsafe.Pointer(&a.player.Lessons)
	case 0x21003000 + 4652:
		field = unsafe.Add(unsafe.Pointer(a.player), pointerChainNativePlayerField("field4652"))
	case 0x21003000 + 4664:
		field = unsafe.Add(unsafe.Pointer(a.player), pointerChainNativePlayerField("field4664"))
	case 0x21003000 + 4692:
		field = unsafe.Add(unsafe.Pointer(a.player), pointerChainNativePlayerField("field4692"))
	default:
		t.Fatalf("unexpected original write %#x", address)
	}
	offset := uintptr(field) - uintptr(unsafe.Pointer(&a.raw[0]))
	size := uintptr(4)
	if pointer {
		size = unsafe.Sizeof(uintptr(0))
	}
	if offset+size > uintptr(len(a.expected)) {
		t.Fatal("original write outside native owner")
	}
	if pointer {
		word := uintptr(a.originalPointer(t, value))
		if size == 8 {
			binary.LittleEndian.PutUint64(a.expected[offset:], uint64(word))
		} else {
			binary.LittleEndian.PutUint32(a.expected[offset:], uint32(word))
		}
	} else {
		binary.LittleEndian.PutUint32(a.expected[offset:], value)
	}
}

func (a *pointerChainNativeArena) run(t *testing.T, route int, c pointerChainNativeCase) {
	t.Helper()
	if c.Input.Index < 2 || c.Input.Index > 8 || c.Input.TradeRole < 0 || c.Input.TradeRole > 2 || c.Input.ListCount < 0 || c.Input.ListCount > 3 {
		t.Fatal("invalid sealed original pointer chain input")
	}
	for _, raw := range [][]byte{a.raw, a.far} {
		for i := range raw {
			raw[i] = byte(i*47 + 23)
		}
	}
	a.owner.ObjClass = object.Class(c.Input.ClassBits)
	a.owner.ObjFlags = object.Flags(c.Input.ObjectFlags)
	globalStateNativeStorePointer(unsafe.Pointer(&a.owner.UpdateData), unsafe.Pointer(a.update))
	globalStateNativeStorePointer(unsafe.Pointer(&a.update.Player), unsafe.Pointer(a.player))
	var active *server.TradeSession
	if c.Input.TradeRole == 1 {
		active = a.session
	} else if c.Input.TradeRole == 2 {
		active = a.other
	}
	globalStateNativeStorePointer(unsafe.Pointer(&a.update.Trade70), unsafe.Pointer(active))
	a.session.Field0 = c.Input.Value
	pointerChainNativeSeedPlayer(a.player, c.Input.Value, c.Input.Counter, c.Input.StatusBits, c.Input.Gate)
	var head *server.TradeItem
	if c.Input.ListCount > 0 {
		head = a.nodes[0]
	}
	globalStateNativeStorePointer(unsafe.Pointer(&a.session.Field20), unsafe.Pointer(head))
	globalStateNativeStorePointer(unsafe.Pointer(&a.other.Field20), unsafe.Pointer(head))
	for i, node := range a.nodes {
		globalStateNativeStorePointer(unsafe.Pointer(&node.Item0), unsafe.Pointer(a.items[i]))
		var next *server.TradeItem
		if i+1 < c.Input.ListCount {
			next = a.nodes[i+1]
		}
		globalStateNativeStorePointer(unsafe.Pointer(&node.Field8), unsafe.Pointer(next))
		a.items[i].NetCode = c.Input.Codes[i]
	}
	copy(a.before, a.raw)
	copy(a.expected, a.raw)
	copy(a.farBefore, a.far)
	for _, write := range c.Writes {
		a.expectWord(t, write.Address, write.Value)
	}
	owner := a.owner
	if !c.Input.Nonnull {
		owner = nil
	}
	value := c.Input.Delta
	if c.Input.Index == 6 {
		value = c.Input.Query
	}
	if got, want := pointerChainNativeCall(route, c.Input.Index, a.session, owner, value), uintptr(a.originalPointer(t, c.ReturnWord)); got != want {
		t.Fatalf("native pointer return %#x want %#x", got, want)
	}
	if !bytes.Equal(a.raw, a.expected) || !bytes.Equal(a.far, a.farBefore) {
		t.Fatal("whole native owner bytes differ from original declared stores")
	}
}

func TestPointerChainNativeOriginal(t *testing.T) {
	cases := pointerChainNativeCases(t)
	for route := 0; route < 3; route++ {
		for _, placement := range []int{0, 64, 512} {
			raw, free := alloc.MakeHigh([]byte{}, 65536)
			t.Cleanup(free)
			if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(&raw[0])) <= math.MaxUint32 {
				t.Fatal("actual foreign owner above4GiB required")
			}
			a := pointerChainNativeArenaAt(raw, nil, placement)
			for _, c := range cases {
				t.Run(fmt.Sprintf("route%d/place%d/%s-%s", route, placement, c.Input.Entry, c.Input.Name), func(t *testing.T) { a.run(t, route, c) })
			}
		}
	}
}

func TestPointerChainNativeSessionIdentity(t *testing.T) {
	near, far, free := tradeNotifyItemNativeOwnerPair()
	defer free()
	a := pointerChainNativeArenaAt(near, far, 0)
	first, second := uintptr(unsafe.Pointer(a.session)), uintptr(unsafe.Pointer(a.other))
	if first == second {
		t.Fatal("distinct native sessions required")
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && (first <= math.MaxUint32 || uint64(second-first) != uint64(1)<<32 || uint32(first) != uint32(second)) {
		t.Fatal("owned native session pair needs equal low words and4GiB gap")
	}
	for _, c := range pointerChainNativeCases(t) {
		if c.Input.Index == 3 && c.Input.ClassBits&4 != 0 && c.Input.TradeRole == 2 {
			for route := 0; route < 3; route++ {
				a.run(t, route, c)
			}
			return
		}
	}
	t.Fatal("sealed original distinct-session identity case absent")
}
