package legacy

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	_ "embed"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

//go:embed testdata/buff_iteration_original_424cb0_424d20.json.gz
var buffIterationOriginal []byte

type buffIterationOriginalCase struct {
	Kind     string   `json:"kind"`
	Count    int32    `json:"count"`
	Table    []uint32 `json:"table"`
	Present  bool     `json:"present"`
	BuffMask uint32   `json:"buff_mask"`
	Arg      uint32   `json:"arg"`
	Result   uint32   `json:"result"`
}

func buffIterationOriginalCases(t *testing.T) []buffIterationOriginalCase {
	t.Helper()
	if got := fmt.Sprintf("%x", sha256.Sum256(buffIterationOriginal)); got != "4e7974e825e3e654509d2907565622fcb752c2d9ba709ab6828286bba0844f9f" {
		t.Fatal("sealed compressed buff reference changed", got)
	}
	z, err := gzip.NewReader(bytes.NewReader(buffIterationOriginal))
	if err != nil {
		t.Fatal(err)
	}
	raw, err := io.ReadAll(z)
	if err != nil {
		t.Fatal(err)
	}
	if err := z.Close(); err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(raw)); got != "4ead174b4e191af68e06557c886ec3f192d6afb4625944099b4ca6317207f7ce" {
		t.Fatal("sealed raw buff reference changed", got)
	}
	var cases []buffIterationOriginalCase
	if err := json.Unmarshal(raw, &cases); err != nil {
		t.Fatal(err)
	}
	kinds := make(map[string]int)
	for _, c := range cases {
		kinds[c.Kind]++
	}
	if len(cases) != 3339 || !reflect.DeepEqual(kinds, map[string]int{"count": 2592, "first": 27, "next": 720}) {
		t.Fatal("original buff operation counts changed", len(cases), kinds)
	}
	return cases
}

func buffIterationPreserveOwners(t *testing.T) (*uint32, []byte, []byte) {
	t.Helper()
	count, table, blobCount := buffIterationOwnersC()
	oldCount := *count
	oldTable, oldBlob := append([]byte(nil), table...), append([]byte(nil), blobCount...)
	t.Cleanup(func() { *count = oldCount; copy(table, oldTable); copy(blobCount, oldBlob) })
	if unsafe.Pointer(count) == unsafe.Pointer(unsafe.SliceData(blobCount)) ||
		unsafe.Pointer(unsafe.SliceData(table)) != memmap.PtrOff(0x587000, 66000) {
		t.Fatal("extracted count and packed table ownership disagree")
	}
	return count, table, blobCount
}

func buffIterationInstallCase(c buffIterationOriginalCase, count *uint32, table, blobCount []byte) {
	*count = uint32(c.Count)
	for i, value := range c.Table {
		binary.LittleEndian.PutUint32(table[i*4:], value)
	}
	// The same original address in the blob is a different owner. An
	// accidental raw-blob count read must predict a different bounded loop.
	decoy := uint32(1)
	if c.Count > 0 {
		decoy = 0
	}
	binary.LittleEndian.PutUint32(blobCount, decoy)
}

func TestBuffIterationNativeLayoutCABI(t *testing.T) {
	size, buffs := buffIterationLayoutC()
	wantSize, wantBuffs := uintptr(772), uintptr(340)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantSize, wantBuffs = 912, 344
	}
	if size != wantSize || buffs != wantBuffs || buffs != unsafe.Offsetof(server.Object{}.Buffs) ||
		size != unsafe.Offsetof(server.Object{}.ScriptPickup)+unsafe.Sizeof(server.Object{}.ScriptPickup) {
		t.Fatal("C/Go native buff layout disagrees", size, buffs, wantSize, wantBuffs)
	}
	count, table, raw := buffIterationPreserveOwners(t)
	t.Logf("extracted count=%p packed table=%p raw count decoy=%p", count, unsafe.SliceData(table), unsafe.SliceData(raw))
}

func TestBuffIterationNativeOriginalCABI(t *testing.T) {
	cases := buffIterationOriginalCases(t)
	count, table, blobCount := buffIterationPreserveOwners(t)
	t.Logf("unchanged original buff iteration cases=3339, instruction bytes=181, declared services=0")
	for _, placement := range []int{0, 8, 24} {
		t.Run(fmt.Sprintf("placement-%d", placement), func(t *testing.T) {
			owner, free := alloc.MakeHigh([]byte{}, int(unsafe.Sizeof(server.Object{}))+128)
			defer free()
			if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(&owner[0])) <= math.MaxUint32 {
				t.Fatal("public C unit must be foreign memory above 4GiB")
			}
			start := 32 + placement
			unit := (*server.Object)(unsafe.Pointer(&owner[start]))
			for index, c := range cases {
				t.Run(fmt.Sprintf("%04d-%s", index, c.Kind), func(t *testing.T) {
					for i := range owner {
						owner[i] = 0xa5
					}
					buffIterationInstallCase(c, count, table, blobCount)
					if unsafe.Sizeof(uintptr(0)) == 8 {
						binary.LittleEndian.PutUint32(owner[start+340:], ^c.BuffMask)
					}
					unit.Buffs = c.BuffMask
					before, beforeTable, beforeBlob := append([]byte(nil), owner...), append([]byte(nil), table...), append([]byte(nil), blobCount...)
					var got uint32
					switch c.Kind {
					case "count":
						var input unsafe.Pointer
						if c.Present {
							input = unsafe.Pointer(unit)
						}
						got = uint32(buffIterationCountC(input))
					case "first":
						got = buffIterationFirstC()
					case "next":
						got = buffIterationNextC(c.Arg)
					default:
						t.Fatal("unknown original operation", c.Kind)
					}
					if got != c.Result {
						t.Fatalf("public C %s result=%#x want=%#x", c.Kind, got, c.Result)
					}
					if *count != uint32(c.Count) || !bytes.Equal(owner, before) || !bytes.Equal(table, beforeTable) || !bytes.Equal(blobCount, beforeBlob) {
						t.Fatal("read-only entry changed native object, table, count, decoy or canaries")
					}
				})
			}
		})
	}
}

func TestBuffIterationNativePlayerSaveOrder(t *testing.T) {
	cases := buffIterationOriginalCases(t)
	count, table, blobCount := buffIterationPreserveOwners(t)
	for index, c := range cases {
		if c.Kind != "first" {
			continue
		}
		t.Run(fmt.Sprintf("%04d", index), func(t *testing.T) {
			// Build the expected chain from actual first/next CPU results.
			next := make(map[uint32]uint32)
			for _, candidate := range cases {
				if candidate.Kind == "next" && candidate.Count == c.Count && reflect.DeepEqual(candidate.Table, c.Table) {
					next[candidate.Arg] = candidate.Result
				}
			}
			var want []server.EnchantID
			seen := make(map[uint32]bool)
			for value := c.Result; value != math.MaxUint32; value = next[value] {
				if seen[value] {
					t.Fatal("original chain would cycle outside admitted save-order corpus")
				}
				seen[value] = true
				if _, ok := next[value]; !ok {
					t.Fatal("original next reference missing", value)
				}
				if int32(value) >= 0 && value < 32 {
					want = append(want, server.EnchantID(value))
				}
			}
			buffIterationInstallCase(c, count, table, blobCount)
			beforeTable, beforeBlob := append([]byte(nil), table...), append([]byte(nil), blobCount...)
			if got := playerEnchantIDs41B9C0(); !reflect.DeepEqual(got, want) {
				t.Fatal("native player-save order differs from original C iteration", got, want)
			}
			if *count != uint32(c.Count) || !bytes.Equal(table, beforeTable) || !bytes.Equal(blobCount, beforeBlob) {
				t.Fatal("player-save iteration changed shared owners")
			}
		})
	}
}
