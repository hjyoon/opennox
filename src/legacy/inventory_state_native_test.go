package legacy

import (
	"bytes"
	_ "embed"
	"fmt"
	"math"
	"strconv"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

//go:embed testdata/inventory_state_original_v1.json.gz
var inventoryStateNativeOriginal []byte

type inventoryStateNativeCase struct {
	Input struct {
		Index                             int
		Entry, Name                       string
		State, Value, Stored, Query, X, Y uint32
		Window                            bool
		First                             uint32
		Heads                             []int
		Nodes                             []struct {
			Type  uint32
			Class uint32 `json:"obj_class"`
			Next  int
		}
	}
	Return *uint32 `json:"return_u32"`
	Role   *string `json:"return_role"`
	Final  struct {
		State, Stored, Window, First uint32
		Heads                        []uint32
	}
	Writes [][2]uint32 `json:"ordered_writes"`
}

func TestInventoryStateNativeLayout(t *testing.T) {
	var drawable client.Drawable
	var owner server.Object
	var window gui.Window
	pointer := unsafe.Sizeof(uintptr(0))
	prefix := unsafe.Offsetof(owner.ScriptPickup) + unsafe.Sizeof(owner.ScriptPickup)
	// Drawable appends a Go-only clientHandle after its aligned C prefix.
	align := unsafe.Alignof(drawable)
	drawablePrefix := (unsafe.Offsetof(drawable.Field_127) + unsafe.Sizeof(drawable.Field_127) + align - 1) &^ (align - 1)
	want := [20]uintptr{
		pointer, drawablePrefix, unsafe.Offsetof(drawable.TypeIDVal), unsafe.Offsetof(drawable.ObjClass), unsafe.Offsetof(drawable.NextPtr),
		unsafe.Sizeof(drawable.TypeIDVal), unsafe.Sizeof(drawable.ObjClass), unsafe.Sizeof(drawable.NextPtr),
		prefix, unsafe.Offsetof(owner.InvFirstItem), unsafe.Sizeof(owner.InvFirstItem), unsafe.Sizeof(window),
		unsafe.Sizeof(*inventoryStateNativeHeads()), pointer, 9, 4, 4, pointer, 4, 4,
	}
	if got := inventoryStateNativeLayoutC(); got != want {
		t.Fatal("native inventory slots/Drawable/Object/Window ABI", got, want)
	}
	if unsafe.Sizeof(drawable) != drawablePrefix+pointer || unsafe.Sizeof(drawable) > 0x400 || unsafe.Sizeof(owner) != prefix+2*pointer || unsafe.Sizeof(owner) > 0x1000 || unsafe.Sizeof(window) > 0x1000 {
		t.Fatal("native owner prefixes or disjoint foreign node strides")
	}
}

func TestEquippedInventoryZoneNative465990(t *testing.T) {
	// This package's tests run without the application's data loader.
	zones := memmap.PtrT[[9][4]int32](0x587000, 136192)
	savedZones := *zones
	defer func() { *zones = savedZones }()
	*zones = [9][4]int32{
		0: {2, 1, 10, 10},
		1: {8, 5, 12, 15},
		6: {30, 20, 40, 31},
		8: {40, 20, 50, 40},
	}

	items, free := alloc.NewHigh([4]client.Drawable{})
	defer free()
	items[0].ObjClass = object.ClassWeapon
	items[0].ObjSubClass = object.SubClass(object.ArmorShield)
	items[1].ObjClass = object.ClassArmor
	items[1].ObjSubClass = object.SubClass(object.ArmorHelmet)
	items[2].ObjClass = object.ClassArmor
	items[2].ObjSubClass = object.SubClass(object.ArmorShield)
	items[3].ObjClass = object.ClassWeapon
	items[3].ObjSubClass = object.SubClass(object.ArmorShield)
	items[0].NextPtr = &items[1]
	items[1].NextPtr = &items[2]

	heads := inventoryStateNativeHeads()
	saved := *heads
	defer func() { *heads = saved }()
	for _, tc := range []struct {
		name      string
		x, y      int32
		slotZero  *client.Drawable
		slotEight *client.Drawable
		want      int32
	}{
		{name: "outside", x: 9, y: 15, want: -1},
		{name: "empty-slot-zero-continues", x: 19, y: 21, want: 1},
		{name: "occupied-slot-zero-wins-overlap", x: 19, y: 21, slotZero: &items[0], want: 0},
		{name: "no-shield", x: 41, y: 35, want: 5},
		{name: "nonarmor-shield-bit", x: 41, y: 35, slotEight: &items[3], want: 5},
		{name: "shield-after-other-items", x: 41, y: 35, slotEight: &items[0], want: 8},
		{name: "shield-zone-inclusive-end", x: 51, y: 46, want: 5},
		{name: "past-shield-zone", x: 52, y: 46, want: 8},
	} {
		t.Run(tc.name, func(t *testing.T) {
			*heads = [9]*client.Drawable{0: tc.slotZero, 8: tc.slotEight}
			before := *heads
			if got := inventoryStateZoneCallC(tc.x, tc.y); got != tc.want {
				t.Fatalf("equipped slot = %d, want %d", got, tc.want)
			}
			if *heads != before {
				t.Fatal("slot heads changed during lookup")
			}
		})
	}
}

func TestInventoryStateNativeOriginal(t *testing.T) {
	var cases []inventoryStateNativeCase
	npcVoiceOriginalDecode(t, inventoryStateNativeOriginal, "69308b39e1ab752c319515c636d83975286630876261d8f5b7c7c36163c73294", "5053285a0dc7a1ed6d8f4530cf60e258dd1b053ca8d29106a2f32093b5befeec", 4916, &cases)
	headsBytes := unsafe.Slice((*byte)(inventoryStateNativeSlot(0)), 9*int(unsafe.Sizeof(uintptr(0))))
	windowBytes := unsafe.Slice((*byte)(inventoryStateNativeSlot(2)), int(unsafe.Sizeof(uintptr(0))))
	savedHeads, savedWindow := append([]byte(nil), headsBytes...), append([]byte(nil), windowBytes...)
	state := (*uint32)(inventoryStateNativeSlot(1))
	stored := globalStateNativeWord(4)
	savedState, savedStored := *state, *stored
	defer func() {
		copy(headsBytes, savedHeads)
		copy(windowBytes, savedWindow)
		*state, *stored = savedState, savedStored
	}()
	for route := 0; route < 3; route++ {
		for _, placement := range []int{0, 64, 512} {
			raw, free := alloc.MakeHigh([]byte{}, 65536)
			t.Cleanup(free)
			if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(&raw[0])) <= math.MaxUint32 {
				t.Fatal("actual foreign inventory graph above4GiB required")
			}
			var nodes [12]*client.Drawable
			for j := range nodes {
				nodes[j] = (*client.Drawable)(unsafe.Pointer(&raw[0x1000+j*0x400+placement]))
			}
			owner := (*server.Object)(unsafe.Pointer(&raw[0x8000+placement]))
			items := []unsafe.Pointer{nil, unsafe.Pointer(&raw[0x9000+placement]), unsafe.Pointer(&raw[0xa000+placement])}
			window := unsafe.Pointer(&raw[0xc000+placement])
			projectNode := func(word uint32) *client.Drawable {
				if word == 0 {
					return nil
				}
				if word < 0x21001000 || (word-0x21001000)%0x400 != 0 || (word-0x21001000)/0x400 >= 12 {
					t.Fatal("unknown original Drawable role")
				}
				return nodes[(word-0x21001000)/0x400]
			}
			projectItem := func(word uint32) unsafe.Pointer {
				switch word {
				case 0:
					return nil
				case 0x21009000:
					return items[1]
				case 0x2100a000:
					return items[2]
				}
				t.Fatal("unknown original Object inventory role")
				return nil
			}
			before := make([]byte, len(raw))
			for _, c := range cases {
				t.Run(fmt.Sprintf("route%d/place%d/%s-%s", route, placement, c.Input.Entry, c.Input.Name), func(t *testing.T) {
					i := c.Input.Index
					if i < 0 || i > 6 || len(c.Input.Heads) != 9 || len(c.Input.Nodes) != 12 || len(c.Final.Heads) != 9 || c.Input.First > 2 || (c.Return == nil) != (i == 3) || (c.Role != nil) != (i == 0 || i == 4 || i == 6) {
						t.Fatal("invalid sealed original owner/return domain")
					}
					for j := range raw {
						raw[j] = byte(j*37 + 29)
					}
					input := [54]uint32{uint32(i), c.Input.State, c.Input.Value, c.Input.Stored, c.Input.Query, c.Input.X, c.Input.Y, 0, c.Input.First}
					if c.Input.Window {
						input[7] = 1
					}
					for j, head := range c.Input.Heads {
						if head < -1 || head >= len(nodes) {
							t.Fatal("original head outside foreign graph")
						}
						input[9+j] = uint32(int32(head))
					}
					for j, node := range c.Input.Nodes {
						if node.Next < -1 || node.Next >= len(nodes) {
							t.Fatal("original link outside foreign graph")
						}
						input[18+3*j], input[19+3*j], input[20+3*j] = node.Type, node.Class, uint32(int32(node.Next))
					}
					inventoryStateNativeSeedC(unsafe.Pointer(&raw[0]), placement, &input)
					for j, node := range c.Input.Nodes {
						var next *client.Drawable
						if node.Next >= 0 {
							next = nodes[node.Next]
						}
						if nodes[j].TypeIDVal != node.Type || uint32(nodes[j].ObjClass) != node.Class || nodes[j].NextPtr != next {
							t.Fatal("actual C-seeded native Drawable fields differ")
						}
					}
					if unsafe.Pointer(owner.InvFirstItem) != items[c.Input.First] || *state != c.Input.State || *stored != c.Input.Stored {
						t.Fatal("actual C-seeded Object/global owners differ")
					}
					copy(before, raw)
					headBefore, windowBefore := append([]byte(nil), headsBytes...), append([]byte(nil), windowBytes...)
					word, pointer := inventoryStateNativeCall(route, i, unsafe.Pointer(owner), int32(c.Input.Query), int32(c.Input.X), int32(c.Input.Y), int32(c.Input.Value))
					if c.Role != nil {
						var want unsafe.Pointer
						switch role := *c.Role; {
						case role == "nil":
						case role == "window":
							want = window
						case role == "item0":
							want = items[1]
						case role == "item1":
							want = items[2]
						case strings.HasPrefix(role, "node"):
							index, err := strconv.Atoi(strings.TrimPrefix(role, "node"))
							if err != nil || index < 0 || index >= len(nodes) {
								t.Fatal("original return outside Drawable graph")
							}
							want = unsafe.Pointer(nodes[index])
						default:
							t.Fatal("unknown original pointer return role")
						}
						if pointer != want || word != 0 {
							t.Fatalf("original pointer role %s: got %#x/%p want %p", *c.Role, word, pointer, want)
						}
					} else if c.Return != nil && (word != *c.Return || pointer != nil) {
						t.Fatalf("original scalar return: got %#x/%p want %#x", word, pointer, *c.Return)
					}
					if *state != c.Final.State || *stored != c.Final.Stored || unsafe.Pointer(owner.InvFirstItem) != projectItem(c.Final.First) {
						t.Fatal("original final global/Object words differ")
					}
					var expectedWindow unsafe.Pointer
					if c.Final.Window == 0x2100c000 {
						expectedWindow = window
					} else if c.Final.Window != 0 {
						t.Fatal("unknown original shared Window role")
					}
					if unsafe.Pointer(Get_dword_5d4594_1062452()) != expectedWindow {
						t.Fatal("original shared Window identity differs")
					}
					for j, original := range c.Final.Heads {
						if inventoryStateNativeHeads()[j] != projectNode(original) {
							t.Fatal("original nine-head native slot identities differ")
						}
					}
					if !bytes.Equal(raw, before) || !bytes.Equal(headsBytes, headBefore) || !bytes.Equal(windowBytes, windowBefore) {
						t.Fatal("readonly whole64KiB inventory graph, nine shared heads or Window slot changed")
					}
					allowedState, allowedStored := c.Input.State, c.Input.Stored
					for _, write := range c.Writes {
						switch write[0] {
						case 0x6d4a9c:
							if i != 3 {
								t.Fatal("state write outside original reset root")
							}
							allowedState = write[1]
						case 0x6d7bec:
							if i != 5 {
								t.Fatal("store outside original setter root")
							}
							allowedStored = write[1]
						default:
							t.Fatal("undeclared original global write")
						}
					}
					if *state != allowedState || *stored != allowedStored {
						t.Fatal("global words differ outside exact original write projection")
					}
				})
			}
		}
	}
}
