package legacy

import (
	"bytes"
	_ "embed"
	"encoding/binary"
	"fmt"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

//go:embed testdata/object_field_original_v1.json.gz
var objectFieldNativeOriginal []byte

type objectFieldNativeCase struct {
	Input struct {
		Index              int
		Entry, Name        string
		Word               uint32
		Head               int8
		Action, A, B, Mode uint32
		Flags, Code        uint32
		Nonnull            bool
		QuerySubclass      uint32 `json:"query_subclass"`
		Nodes              []struct {
			Flags    uint32
			Class    uint32 `json:"obj_class"`
			Subclass uint32
		}
		Words [32]uint32
	}
	ReturnWord *uint32     `json:"return_u32"`
	ReturnRole *string     `json:"return_role"`
	FinalSync  uint32      `json:"final_sync"`
	FinalWords [32]uint32  `json:"final_words"`
	Writes     [][2]uint32 `json:"ordered_writes"`
}

func TestObjectFieldNativeLayout(t *testing.T) {
	var owner server.Object
	var update server.MonsterUpdateData
	var action server.AIStackItem
	var drawable client.Drawable
	pointer := unsafe.Sizeof(uintptr(0))
	prefix := unsafe.Offsetof(owner.ScriptPickup) + unsafe.Sizeof(owner.ScriptPickup)
	want := [35]uintptr{
		pointer, prefix, unsafe.Offsetof(owner.UpdateData), unsafe.Offsetof(owner.Field38), unsafe.Offsetof(owner.Field37), unsafe.Offsetof(owner.Field140), unsafe.Sizeof(owner.Field140),
		unsafe.Offsetof(owner.Buffs), unsafe.Offsetof(owner.ObjClass), unsafe.Offsetof(owner.ObjSubClass), unsafe.Offsetof(owner.ObjFlags), unsafe.Offsetof(owner.InvFirstItem), unsafe.Offsetof(owner.InvNextItem), unsafe.Sizeof(owner.Field37),
		unsafe.Offsetof(update.AIStackInd), unsafe.Offsetof(update.AIStack), unsafe.Sizeof(action), unsafe.Offsetof(action.Action), unsafe.Offsetof(action.Args), unsafe.Offsetof(action.Field5), unsafe.Offsetof(update.Field132), unsafe.Sizeof(update.AIStackInd), unsafe.Sizeof(update.AIStack),
		unsafe.Offsetof(drawable.ObjClass), unsafe.Offsetof(drawable.NetCode32), unsafe.Sizeof(drawable.ObjClass), unsafe.Sizeof(drawable.NetCode32), 4, 2, 4, unsafe.Sizeof(owner.Buffs), pointer, pointer, pointer, pointer,
	}
	if got := objectFieldNativeLayoutC(); got != want {
		t.Fatal("native Object/AI/Drawable and public ABI projections", got, want)
	}
	if unsafe.Sizeof(owner) != prefix+2*pointer || unsafe.Sizeof(owner) > 0x1000 || unsafe.Sizeof(update) > 0x2000 || unsafe.Sizeof(drawable) > 0x5000 {
		t.Fatal("native shared prefix or foreign owner stride")
	}
}

func TestObjectFieldNativeOriginal(t *testing.T) {
	var cases []objectFieldNativeCase
	npcVoiceOriginalDecode(t, objectFieldNativeOriginal, "000dd04a7ef22dcea1fba6edcf7221c45f319f85455b49e0e290bd17e947e32d", "6a65dd0aa8b4270853adbab9137cfba1e53edf34bf614ab3bd2f267460c74789", 5080, &cases)
	for route := 0; route < 3; route++ {
		for _, placement := range []int{0, 64, 512} {
			raw, free := alloc.MakeHigh([]byte{}, 65536)
			t.Cleanup(free)
			if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(&raw[0])) <= math.MaxUint32 {
				t.Fatal("actual foreign Object/AI/Drawable owner above4GiB required")
			}
			owner := (*server.Object)(unsafe.Pointer(&raw[0x1000+placement]))
			update := (*server.MonsterUpdateData)(unsafe.Pointer(&raw[0x3000+placement]))
			item := (*server.Object)(unsafe.Pointer(&raw[0x5000+placement]))
			drawable := (*client.Drawable)(unsafe.Pointer(&raw[0xa000+placement]))
			expected := make([]byte, len(raw))
			for _, c := range cases {
				t.Run(fmt.Sprintf("route%d/place%d/%s-%s", route, placement, c.Input.Entry, c.Input.Name), func(t *testing.T) {
					i := c.Input.Index
					if i < 0 || i >= 8 || c.Input.Head < -1 || c.Input.Head > 23 || len(c.Input.Nodes) > 4 || (c.ReturnWord == nil) != (i == 1) || (c.ReturnRole != nil) != (i == 2 || i == 4) {
						t.Fatal("invalid sealed original owner/return domain")
					}
					for j := range raw {
						raw[j] = byte(j*47 + 23)
					}
					var input [56]uint32
					input[0], input[1], input[2], input[3] = uint32(i), c.Input.Word, uint32(uint8(c.Input.Head)), c.Input.Action
					input[4], input[5], input[6], input[7], input[8] = c.Input.A, c.Input.B, c.Input.Mode, c.Input.Flags, c.Input.Code
					if c.Input.Nonnull {
						input[9] = 1
					}
					input[10], input[11] = c.Input.QuerySubclass, uint32(len(c.Input.Nodes))
					copy(input[12:44], c.Input.Words[:])
					for j, node := range c.Input.Nodes {
						input[44+3*j], input[45+3*j], input[46+3*j] = node.Flags, node.Class, node.Subclass
					}
					objectFieldNativeSeedC(unsafe.Pointer(&raw[0]), placement, &input)
					if owner.UpdateData != unsafe.Pointer(update) || update.AIStackInd != c.Input.Head || owner.Field140 != c.Input.Words || drawable.NetCode32 != c.Input.Code || uint32(drawable.ObjClass) != c.Input.Flags {
						t.Fatal("actual native C owner seed does not match Go fields")
					}
					if (c.Input.Head == -1 && update.Field132 != c.Input.Action) || (c.Input.Head >= 0 && update.AIStack[c.Input.Head].Action != c.Input.Action) {
						t.Fatal("native empty/valid AI head projection")
					}
					copy(expected, raw)
					var dr unsafe.Pointer
					if c.Input.Nonnull {
						dr = unsafe.Pointer(drawable)
					}
					word, pointer := objectFieldNativeCall(route, i, unsafe.Pointer(owner), unsafe.Pointer(item), dr, c.Input.A, c.Input.B, c.Input.Mode, c.Input.Word)
					if c.ReturnRole != nil {
						var want unsafe.Pointer
						switch *c.ReturnRole {
						case "field140_end":
							want = unsafe.Add(unsafe.Pointer(&owner.Field140[0]), unsafe.Sizeof(owner.Field140))
						case "nil":
						case "node0", "node1", "node2", "node3":
							index := int((*c.ReturnRole)[4] - '0')
							if index >= len(c.Input.Nodes) {
								t.Fatal("original return outside inventory chain")
							}
							want = unsafe.Pointer(&raw[0x6000+index*0x1000+placement])
						default:
							t.Fatal("unknown original pointer return role")
						}
						if pointer != want {
							t.Fatalf("original pointer role %s: got %p want %p", *c.ReturnRole, pointer, want)
						}
					} else if c.ReturnWord != nil && (word != *c.ReturnWord || pointer != nil) {
						t.Fatalf("original return %#x/%p want %#x", word, pointer, *c.ReturnWord)
					}
					if owner.Field38 != c.FinalSync || owner.Field140 != c.FinalWords {
						t.Fatal("original final sync word or32 field-mask words differ")
					}
					for _, write := range c.Writes {
						var offset uintptr
						if i == 1 && write[0] == 0x1000+152 {
							offset = 0x1000 + uintptr(placement) + unsafe.Offsetof(owner.Field38)
						} else if i == 2 && write[0] >= 0x1000+560 && write[0] < 0x1000+688 && (write[0]-0x1000-560)%4 == 0 {
							offset = 0x1000 + uintptr(placement) + unsafe.Offsetof(owner.Field140) + uintptr(write[0]-0x1000-560)
						} else {
							t.Fatal("undeclared original scalar write")
						}
						binary.LittleEndian.PutUint32(expected[offset:offset+4], write[1])
					}
					if !bytes.Equal(raw, expected) {
						t.Fatal("whole64KiB native graph changed outside exact original scalar projection")
					}
				})
			}
		}
	}
}
