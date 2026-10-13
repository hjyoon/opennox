//go:build !server

package legacy

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

const videoCallbackOriginalSHA48A210 = "40c5134c7136151d13086b73e79f4bf70c158248dcb597c48d21946e620ed097"

type videoCallbackOriginal48A210 struct {
	ExecutableSHA256    string            `json:"executable_sha256"`
	UnhookedEntries     map[string]int    `json:"unhooked_entries"`
	DeclaredServices    map[string]int    `json:"declared_services"`
	MemoryFaults        []json.RawMessage `json:"memory_faults"`
	ExecutableUnchanged bool              `json:"executable_bytes_unchanged"`
	InputsUnchanged     bool              `json:"guarded_inputs_unchanged"`
	SetterRows          []struct {
		Initial            uint32 `json:"initial"`
		Value              uint32 `json:"value"`
		Result             uint32 `json:"result"`
		Stored             uint32 `json:"stored"`
		NeighborsUnchanged bool   `json:"neighbors_unchanged"`
	} `json:"setter_rows"`
	DescriptorRows []struct {
		Fields         []uint32 `json:"fields"`
		Ignored        uint32   `json:"ignored_arg"`
		Result         int32    `json:"result"`
		Count          uint32   `json:"count"`
		TableHex       string   `json:"table_hex"`
		InputUnchanged bool     `json:"input_unchanged"`
	} `json:"descriptor_rows"`
}

func loadVideoCallbackOriginal48A210(t *testing.T) videoCallbackOriginal48A210 {
	t.Helper()
	b, err := os.ReadFile("testdata/video_callback_setter_48a210.json")
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(b)
	if hex.EncodeToString(sum[:]) != videoCallbackOriginalSHA48A210 {
		t.Fatal("unchanged original transcript hash differs")
	}
	var original videoCallbackOriginal48A210
	if err := json.Unmarshal(b, &original); err != nil {
		t.Fatal(err)
	}
	return original
}

func TestVideoCallbackSetterOriginal48A210(t *testing.T) {
	x := loadVideoCallbackOriginal48A210(t)
	if x.ExecutableSHA256 != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" ||
		x.UnhookedEntries["0x48a210"] != 24 || x.UnhookedEntries["0x47fce0"] != 6 ||
		len(x.DeclaredServices) != 0 || len(x.MemoryFaults) != 0 || !x.ExecutableUnchanged || !x.InputsUnchanged {
		t.Fatal("original execution scope differs")
	}
	if len(x.SetterRows) != 24 || len(x.DescriptorRows) != 6 {
		t.Fatal("original cases differ")
	}
	for _, row := range x.SetterRows {
		if row.Result != row.Value || row.Stored != row.Value || !row.NeighborsUnchanged {
			t.Fatalf("original did not preserve the pointer argument: %+v", row)
		}
	}
	for _, row := range x.DescriptorRows {
		if len(row.Fields) != 27 || !row.InputUnchanged || row.Result != 1 {
			t.Fatal("descriptor contract differs")
		}
	}
}

func TestVideoCallbackSetterNativeDeclaration48A210(t *testing.T) {
	if !videoCallbackSetterNativeDeclaration48A210() {
		t.Fatal("video callback setter still uses a pointer-truncating ABI")
	}
}

func TestVideoCallbackSetterNativeStorage48A210(t *testing.T) {
	owned, free := alloc.Make([]byte{}, 64)
	defer free()
	data := unsafe.Pointer(&owned[32])
	callback := Get_sub_47FCE0()
	if unsafe.Sizeof(uintptr(0)) > 4 && runtime.GOOS == "darwin" && runtime.GOARCH == "arm64" {
		if uintptr(callback) <= 0xffffffff || uintptr(data) <= 0xffffffff {
			t.Fatalf("native-address premise missing: callback=%p data=%p", callback, data)
		}
	}
	const base, offset = uintptr(0x5D4594), uintptr(1193504)
	slot := memmap.PtrPtr(base, offset)
	saved := *slot
	packed := memmap.Slice(base, offset-4)[:16]
	before := append([]byte(nil), packed...)
	defer func() { copy(packed, before); *slot = saved }()
	cases := []struct {
		name  string
		value unsafe.Pointer
	}{
		{"nil", nil}, {"real-video-callback", callback},
		{"other-C-symbol", Get_nox_thing_trigger_draw()}, {"C-owned-opaque-data", data},
	}
	for _, route := range []string{"Go-wrapper", "C-return"} {
		for pattern := 0; pattern < 2; pattern++ {
			for _, tc := range cases {
				t.Run(fmt.Sprintf("%s/pattern-%d/%s", route, pattern, tc.name), func(t *testing.T) {
					for i := range packed {
						packed[i] = byte(0x17 + i*19 + pattern*37)
					}
					*slot = Get_nox_thing_trigger_draw()
					guards := append([]byte(nil), packed...)
					if route == "Go-wrapper" {
						Nox_xxx_setSomeFunc_48A210(tc.value)
					} else if got := videoCallbackSetterReturnC48A210(tc.value); got != uintptr(tc.value) {
						t.Errorf("return=%#x, want full native argument %#x", got, uintptr(tc.value))
					}
					if got := *slot; got != tc.value {
						t.Errorf("stored=%p, want %p", got, tc.value)
					}
					if unsafe.Sizeof(uintptr(0)) > 4 {
						if !bytes.Equal(packed, guards) {
							t.Errorf("packed PE32 bytes changed: %x -> %x", guards, packed)
						}
					} else if !bytes.Equal(packed[:4], guards[:4]) || !bytes.Equal(packed[8:], guards[8:]) {
						t.Errorf("PE32 neighboring fields changed: %x -> %x", guards, packed)
					}
				})
			}
		}
	}
}

func TestVideoCallbackSetterRealConsumer48A210(t *testing.T) {
	original := loadVideoCallbackOriginal48A210(t)
	descriptor, free := alloc.Make([]uint32{}, 35)
	defer free()
	const base, offset = uintptr(0x5D4594), uintptr(1193504)
	slot := memmap.PtrPtr(base, offset)
	saved := *slot
	packed := memmap.Slice(base, offset-4)[:16]
	packedBefore := append([]byte(nil), packed...)
	countBefore := *videoCallbackDescriptorCountC48A210
	table := memmap.Slice(0x973F18, 6088-16)[:160]
	tableBefore := append([]byte(nil), table...)
	defer func() {
		copy(table, tableBefore)
		*videoCallbackDescriptorCountC48A210 = countBefore
		copy(packed, packedBefore)
		*slot = saved
	}()
	for _, route := range []string{"direct-existing-C-control", "stored-native-callback"} {
		t.Run(route, func(t *testing.T) {
			for i := range table {
				table[i] = 0xa7
			}
			for i := 0; i < 16; i++ {
				table[i] = 0x3c
				table[144+i] = 0xc3
			}
			*videoCallbackDescriptorCountC48A210 = 0
			*slot = nil
			for i, row := range original.DescriptorRows {
				t.Run(fmt.Sprintf("row-%d", i), func(t *testing.T) {
					for j := range descriptor {
						descriptor[j] = 0xdecafbad
					}
					copy(descriptor[4:31], row.Fields)
					input := append([]uint32(nil), descriptor...)
					var result int32
					if route == "stored-native-callback" {
						Nox_xxx_setSomeFunc_48A210(Get_sub_47FCE0())
						if *slot != Get_sub_47FCE0() {
							t.Fatal("real C callback address not preserved")
						}
						result = videoCallbackStoredCall48A210(&descriptor[4], row.Ignored)
					} else {
						result = videoCallbackDirectCall47FCE0(&descriptor[4], row.Ignored)
					}
					want, err := hex.DecodeString(row.TableHex)
					if err != nil {
						t.Fatal(err)
					}
					if result != row.Result || *videoCallbackDescriptorCountC48A210 != row.Count || !bytes.Equal(table[16:144], want) {
						t.Fatalf("descriptor result=%d count=%d table=%x; want %d/%d/%x", result, *videoCallbackDescriptorCountC48A210, table[16:144], row.Result, row.Count, want)
					}
					for j := range descriptor {
						if descriptor[j] != input[j] {
							t.Fatalf("descriptor changed at word %d", j)
						}
					}
					for j := 0; j < 16; j++ {
						if table[j] != 0x3c || table[144+j] != 0xc3 {
							t.Fatal("descriptor table bounds changed")
						}
					}
				})
			}
		})
	}
}
