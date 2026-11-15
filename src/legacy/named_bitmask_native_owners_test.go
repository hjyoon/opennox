package legacy

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

//go:embed testdata/named_bitmask_original_423930_4239c0_v2.json.gz
var namedBitmaskOriginal []byte

//go:embed testdata/named_bitmask_monster_owner_original.json
var namedBitmaskStockOwnerOriginal []byte

func namedBitmaskStockOwner(t *testing.T) {
	t.Helper()
	var original struct {
		Table   uintptr   `json:"table_offset"`
		Offsets []uintptr `json:"pointer_offsets"`
		Start   uintptr   `json:"string_offset"`
		Size    int       `json:"string_bytes"`
		SHA     string    `json:"string_sha256"`
	}
	if err := json.Unmarshal(namedBitmaskStockOwnerOriginal, &original); err != nil {
		t.Fatal(err)
	}
	if original.Table != 247536 || len(original.Offsets) != 22 || original.Start != 247936 || original.Size != 256 {
		t.Fatal("original stock owner metadata")
	}
	blob, err := os.ReadFile("../common/memmap/nox/blobdata/blob_587000.dat")
	if err != nil {
		t.Fatal(err)
	}
	strings := blob[original.Start : original.Start+uintptr(original.Size)]
	if fmt.Sprintf("%x", sha256.Sum256(strings)) != original.SHA {
		t.Fatal("shipped stock strings differ from original PE")
	}
	owner := memmap.Slice(0x587000, original.Table-16)[:672]
	old := bytes.Clone(owner)
	var oldPtrs [23]unsafe.Pointer
	for i := range oldPtrs {
		oldPtrs[i] = *memmap.PtrPtr(0x587000, original.Table+4*uintptr(i))
	}
	t.Cleanup(func() {
		copy(owner, old)
		for i, p := range oldPtrs {
			*memmap.PtrPtr(0x587000, original.Table+4*uintptr(i)) = p
		}
	})
	// Unit tests do not run game startup. Hydrate exactly this immutable
	// string region and the original4-byte pointer slots, then restore them.
	copy(memmap.Slice(0x587000, original.Start), strings)
	for i, off := range original.Offsets {
		*memmap.PtrPtr(0x587000, original.Table+4*uintptr(i)) = memmap.PtrOff(0x587000, off)
	}
	*memmap.PtrPtr(0x587000, original.Table+4*uintptr(len(original.Offsets))) = nil
}

type namedBitmaskOriginalCase struct {
	Kind       string           `json:"kind"`
	Names      []string         `json:"names_hex"`
	Input      string           `json:"input_hex"`
	BeforeMask uint32           `json:"before_mask"`
	AfterMask  uint32           `json:"after_mask"`
	Result     *int32           `json:"result"`
	Before     string           `json:"owner_before"`
	After      string           `json:"owner_after"`
	Trace      []map[string]any `json:"trace"`
}

func namedBitmaskCases(t *testing.T) []namedBitmaskOriginalCase {
	t.Helper()
	var out []namedBitmaskOriginalCase
	npcVoiceOriginalDecode(t, namedBitmaskOriginal, "ebd1021f867af31d59b821f35a2726174b6f239686fe6d400aa394c2ae94ed74", "df7c30a1186768e14f9f022aeacae62c42ab755defd3e17ff4ff85979c074042", 1600, &out)
	return out
}

type namedBitmaskOwners struct {
	names, array, input, data []byte
	start                     int
}

func namedBitmaskOwnersNew(t *testing.T, placement int) *namedBitmaskOwners {
	t.Helper()
	v := &namedBitmaskOwners{start: 32 + placement}
	for _, r := range []struct {
		out *[]byte
		n   int
	}{{&v.names, 32768}, {&v.array, 4096}, {&v.input, 4096}, {&v.data, 128}} {
		b, free := alloc.MakeHigh([]byte{}, r.n)
		t.Cleanup(free)
		*r.out = b
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(&b[0])) <= math.MaxUint32 {
			t.Fatal("actual C owner must be foreign memory above4GiB")
		}
	}
	return v
}

func (v *namedBitmaskOwners) reset(t *testing.T, c namedBitmaskOriginalCase) {
	t.Helper()
	for _, b := range [][]byte{v.names, v.array, v.input, v.data} {
		for i := range b {
			b[i] = 0xa5
		}
	}
	for i, s := range c.Names {
		name := npcVoiceUnhex(t, s)
		start := v.start + i*64
		if len(name) >= 64 {
			t.Fatal("fixture name too long")
		}
		copy(v.names[start:], name)
		v.names[start+len(name)] = 0
		*(*unsafe.Pointer)(unsafe.Pointer(&v.array[v.start+i*int(unsafe.Sizeof(uintptr(0)))])) = unsafe.Pointer(&v.names[start])
	}
	*(*unsafe.Pointer)(unsafe.Pointer(&v.array[v.start+len(c.Names)*int(unsafe.Sizeof(uintptr(0)))])) = nil
	input := npcVoiceUnhex(t, c.Input)
	copy(v.input[v.start:], input)
	v.input[v.start+len(input)] = 0
	before := npcVoiceUnhex(t, c.Before)
	if len(before) != 20 {
		t.Fatal("sealed original owner size")
	}
	copy(v.data[v.start:], before)
}

func (v *namedBitmaskOwners) pointers() (unsafe.Pointer, unsafe.Pointer, unsafe.Pointer) {
	return unsafe.Pointer(&v.input[v.start]), unsafe.Pointer(&v.data[v.start+8]), unsafe.Pointer(&v.array[v.start])
}

func (v *namedBitmaskOwners) expect(t *testing.T, c namedBitmaskOriginalCase) func() {
	t.Helper()
	names, array, input := bytes.Clone(v.names), bytes.Clone(v.array), bytes.Clone(v.input)
	want := bytes.Clone(v.data)
	copy(want[v.start:], npcVoiceUnhex(t, c.After))
	return func() {
		t.Helper()
		if !bytes.Equal(names, v.names) || !bytes.Equal(array, v.array) || !bytes.Equal(input, v.input) || !bytes.Equal(want, v.data) {
			t.Fatal("C changed nonoutputs, lost native pointers or changed whole mask owner", c.Kind, c.BeforeMask, c.AfterMask)
		}
	}
}

func namedBitmaskHexCString(p unsafe.Pointer) string {
	return hex.EncodeToString([]byte(alloc.GoString((*byte)(p))))
}

func namedBitmaskCaseName(i int, c namedBitmaskOriginalCase) string {
	return fmt.Sprintf("%04d-%s", i, c.Kind)
}
