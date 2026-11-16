package legacy

import (
	"bytes"
	_ "embed"
	"encoding/binary"
	"fmt"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

//go:embed testdata/object_trace_original_v1.json.gz
var objectTraceNativeOriginal []byte

type objectTraceNativeCase struct {
	Input struct {
		Index          int
		Entry, Name    string
		Owner          bool
		NetCode        uint32 `json:"net_code"`
		TypeInd        uint32 `json:"type_ind"`
		Flags          uint32
		ID, Init, Next int
		Ready, X, Y    uint32
	}
	Return *uint32 `json:"return_u32"`
	Byte   *uint8  `json:"return_u8"`
	Role   *string `json:"return_role"`
	Final  struct {
		ID, Flags, Next, Init, Ready uint32
		TypeInd                      uint32 `json:"type_ind"`
		NetCode                      uint32 `json:"net_code"`
		Point                        [2]uint32
	}
	Writes [][3]uint32 `json:"ordered_writes"`
}

func TestObjectTraceNativeLayout(t *testing.T) {
	var owner server.Object
	var point ntype.Point32
	pointer := unsafe.Sizeof(uintptr(0))
	prefix := unsafe.Offsetof(owner.ScriptPickup) + unsafe.Sizeof(owner.ScriptPickup)
	want := [20]uintptr{
		pointer, prefix, unsafe.Offsetof(owner.IDPtr), unsafe.Offsetof(owner.TypeInd), unsafe.Offsetof(owner.ObjFlags),
		unsafe.Offsetof(owner.NetCode), unsafe.Offsetof(owner.InvNextItem), unsafe.Offsetof(owner.InitData),
		unsafe.Sizeof(owner.IDPtr), unsafe.Sizeof(owner.TypeInd), unsafe.Sizeof(owner.ObjFlags), unsafe.Sizeof(owner.NetCode),
		unsafe.Sizeof(owner.InvNextItem), unsafe.Sizeof(owner.InitData), unsafe.Sizeof(point), unsafe.Offsetof(point.X), unsafe.Offsetof(point.Y), 4, 1, unsafe.Sizeof(point.X),
	}
	if got := objectTraceNativeLayoutC(); got != want {
		t.Fatal("native Object/Point32 fields and declared scalar ABI", got, want)
	}
	if unsafe.Sizeof(owner) != prefix+2*pointer || unsafe.Sizeof(owner) > 0x4000 {
		t.Fatal("shared Object prefix and two existing Go-only handles")
	}
}

func TestObjectTraceNativeOriginal(t *testing.T) {
	var cases []objectTraceNativeCase
	npcVoiceOriginalDecode(t, objectTraceNativeOriginal, "061282c31e0413e3e12e52506147cb8e85d73ddd09844b822c4bc1dd928a911e", "3d255fba0c080112c4cc6a036be1121e7a523bb89bc942e2621aea7987bc4217", 1780, &cases)
	point := memmap.PtrT[ntype.Point32](0x5D4594, 2488612)
	savedPoint, savedReady := *point, Get_dword_5d4594_2488620()
	defer func() {
		*point = savedPoint
		Set_dword_5d4594_2488620(savedReady)
	}()
	entries := [7]string{"nox_xxx_objectGetNetCode_4E4C00", "nox_xxx_objectGetTypeInd_4E4C10", "nox_object_getInitData_4E4C30", "nox_xxx_objectGetID_4E4C80", "nox_xxx_inventoryGetNext_4E7990", "sub_537760", "sub_53B300"}
	for route := 0; route < 3; route++ {
		for _, placement := range []int{0, 64, 512} {
			raw, free := alloc.MakeHigh([]byte{}, 65536)
			t.Cleanup(free)
			if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(&raw[0])) <= math.MaxUint32 || uintptr(unsafe.Pointer(point)) <= math.MaxUint32) {
				t.Fatal("actual native Object/opaque/Point32 owners above4GiB required")
			}
			owner := (*server.Object)(unsafe.Pointer(&raw[0x1000+placement]))
			pointers := map[string][3]unsafe.Pointer{
				"id":   {nil, unsafe.Pointer(&raw[0x6000+placement]), unsafe.Pointer(&raw[0x7000+placement])},
				"init": {nil, unsafe.Pointer(&raw[0x8000+placement]), unsafe.Pointer(&raw[0x9000+placement])},
				"next": {nil, unsafe.Pointer(&raw[0xa000+placement]), unsafe.Pointer(&raw[0xb000+placement])},
			}
			roles := map[string]unsafe.Pointer{"nil": nil, "trace": unsafe.Pointer(point)}
			for kind, values := range pointers {
				roles[kind+"0"], roles[kind+"1"] = values[1], values[2]
			}
			project := func(kind string, original uint32) unsafe.Pointer {
				if original == 0 {
					return nil
				}
				base := map[string]uint32{"id": 0x21006000, "init": 0x21008000, "next": 0x2100a000}[kind]
				if original < base || (original-base)%0x1000 != 0 || (original-base)/0x1000 > 1 {
					t.Fatal("unknown original opaque/Object role", kind, original)
				}
				return pointers[kind][1+(original-base)/0x1000]
			}
			expected := make([]byte, len(raw))
			for _, c := range cases {
				t.Run(fmt.Sprintf("route%d/place%d/%s-%s", route, placement, c.Input.Entry, c.Input.Name), func(t *testing.T) {
					i := c.Input.Index
					if i < 0 || i >= len(entries) || c.Input.Entry != entries[i] || c.Input.TypeInd > math.MaxUint16 || c.Input.ID < 0 || c.Input.ID > 2 || c.Input.Init < 0 || c.Input.Init > 2 || c.Input.Next < 0 || c.Input.Next > 2 || (i == 6 && !c.Input.Owner) || (c.Return == nil) != (i == 6) || (c.Byte != nil) != (i == 6) || (c.Role != nil) != (i >= 2 && i <= 5) {
						t.Fatal("invalid sealed Object/trace owner, entry or declared return domain")
					}
					for j := range raw {
						raw[j] = byte(j*37 + 29)
					}
					input := [11]uint32{uint32(i), 0, c.Input.NetCode, c.Input.TypeInd, c.Input.Flags, uint32(c.Input.ID), uint32(c.Input.Init), uint32(c.Input.Next), c.Input.Ready, c.Input.X, c.Input.Y}
					if c.Input.Owner {
						input[1] = 1
					}
					objectTraceNativeSeedC(unsafe.Pointer(&raw[0]), placement, &input)
					if owner.NetCode != c.Input.NetCode || uint32(owner.TypeInd) != c.Input.TypeInd || uint32(owner.ObjFlags) != c.Input.Flags || owner.IDPtr != pointers["id"][c.Input.ID] || owner.InitData != pointers["init"][c.Input.Init] || unsafe.Pointer(owner.InvNextItem) != pointers["next"][c.Input.Next] || Get_dword_5d4594_2488620() != c.Input.Ready || uint32(point.X) != c.Input.X || uint32(point.Y) != c.Input.Y {
						t.Fatal("actual C-seeded shared native Object/Point32 fields differ")
					}
					copy(expected, raw)
					var argument unsafe.Pointer
					if c.Input.Owner {
						argument = unsafe.Pointer(owner)
					}
					word, pointer := objectTraceNativeCall(route, i, argument)
					if c.Role != nil {
						want, valid := roles[*c.Role]
						if !valid || pointer != want || word != 0 {
							t.Fatalf("original full native pointer role %s: got %#x/%p want %p", *c.Role, word, pointer, want)
						}
					} else {
						want := uint32(0)
						if c.Byte != nil {
							want = uint32(*c.Byte)
							// Original upper EAX residue is deliberately uncredited.
						} else {
							want = *c.Return
						}
						if word != want || pointer != nil {
							t.Fatalf("original declared scalar return: got %#x/%p want %#x", word, pointer, want)
						}
					}
					if owner.IDPtr != project("id", c.Final.ID) || owner.InitData != project("init", c.Final.Init) || unsafe.Pointer(owner.InvNextItem) != project("next", c.Final.Next) || owner.NetCode != c.Final.NetCode || uint32(owner.TypeInd) != c.Final.TypeInd || uint32(owner.ObjFlags) != c.Final.Flags || Get_dword_5d4594_2488620() != c.Final.Ready || uint32(point.X) != c.Final.Point[0] || uint32(point.Y) != c.Final.Point[1] {
						t.Fatal("original final native Object and trace words differ")
					}
					allowedFlags := c.Input.Flags
					if (len(c.Writes) == 1) != (i == 6) || len(c.Writes) > 1 {
						t.Fatal("original exact single-store flags ledger")
					}
					for _, write := range c.Writes {
						if write[0] != 0x21001010 || write[1] != 4 || i != 6 {
							t.Fatal("write outside original ObjFlags word")
						}
						allowedFlags = write[2]
					}
					if uint32(owner.ObjFlags) != allowedFlags {
						t.Fatal("flags differ outside original write projection")
					}
					if i == 6 {
						at := 0x1000 + placement + int(unsafe.Offsetof(owner.ObjFlags))
						binary.LittleEndian.PutUint32(expected[at:at+4], allowedFlags)
					}
					if !bytes.Equal(raw, expected) {
						t.Fatal("whole64KiB Object and opaque graph differs outside exact original flags word")
					}
				})
			}
		}
	}
}
