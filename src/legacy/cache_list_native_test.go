package legacy

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

//go:embed testdata/cache_list_original_v1.json.gz
var cacheListNativeOriginal []byte

type cacheListNativeCase struct {
	Input struct {
		Name, Entry string
		Seed        int
		ListOwner   int `json:"list_owner"`
		ListIndex   int `json:"list_index"`
		NodeOwner   int `json:"node_owner"`
		NodeIndex   int `json:"node_index"`
		Slots       [128]byte
	}
	ReturnPointer uint32 `json:"return_pointer"`
	OrderedWrites []struct {
		Address uint32
		Bytes   string
	} `json:"ordered_writes"`
	OwnerBefore       [2]string   `json:"owner_sha256_before"`
	OwnerAfter        [2]string   `json:"owner_sha256_after"`
	PointerWordsAfter [128]uint32 `json:"pointer_words_after"`
}

func cacheListNativeCases(t *testing.T) []cacheListNativeCase {
	t.Helper()
	var cases []cacheListNativeCase
	npcVoiceOriginalDecode(t, cacheListNativeOriginal, "ea1e6a21d40c05f14b8d33b6652cb87b7ed58bc16ca360ecd58cfa8f3b482f1e", "e84f82fb86e2d3b1eedb264a73e9b084cbfe9d6c479e25dd21ee7fc24eb89898", 2992, &cases)
	return cases
}

func TestCacheListNativeLayout(t *testing.T) {
	w := unsafe.Sizeof(uintptr(0))
	want := [14]uintptr{w, 4, 3 * w, 0, w, 2 * w, 52 * w, 0, w, 2 * w, 50 * w, 51 * w, 16, w}
	if got := cacheListNativeLayoutC(); got != want {
		t.Fatal("native node/list/cache/callback ABI", got, want)
	}
	cache := cacheListNativeCacheC()
	if len(cache) != int(52*w) || w == 8 && uintptr(unsafe.Pointer(&cache[0])) <= math.MaxUint32 {
		t.Fatal("full native extracted C cache owner required")
	}
}

func cacheListNativeOrigin(role byte) uint32 {
	if role == 0 {
		return 0
	}
	r := uint32(role - 1)
	return 0x21000000 + (r/62)*0x100000 + 4*(r%62)
}

func cacheListNativeInput(c cacheListNativeCase) (out [2][]byte) {
	for owner := range out {
		out[owner] = make([]byte, 65536)
		for j := range out[owner] {
			out[owner][j] = byte(j*53 + c.Input.Seed + owner*71)
		}
		for cell := 0; cell < 64; cell++ {
			binary.LittleEndian.PutUint32(out[owner][4*cell:], cacheListNativeOrigin(c.Input.Slots[64*owner+cell]))
		}
	}
	return
}

func cacheListNativeHash(t *testing.T, raw []byte, want string) {
	t.Helper()
	if got := fmt.Sprintf("%x", sha256.Sum256(raw)); got != want {
		t.Fatalf("original owner SHA got%s want%s", got, want)
	}
}

func cacheListNativeProject(t *testing.T, raw [2][]byte, placement, word int, address uint32) unsafe.Pointer {
	t.Helper()
	if address == 0 {
		return nil
	}
	for owner, base := range [2]uint32{0x21000000, 0x21100000} {
		if address >= base && address < base+256 && (address-base)%4 == 0 {
			return unsafe.Pointer(&raw[owner][placement+word*int((address-base)/4)])
		}
	}
	t.Fatalf("pointer outside declared two C owners: %#x", address)
	return nil
}

func cacheListNativePut(raw []byte, offset, word int, value unsafe.Pointer) {
	if word == 4 {
		binary.LittleEndian.PutUint32(raw[offset:], uint32(uintptr(value)))
	} else {
		binary.LittleEndian.PutUint64(raw[offset:], uint64(uintptr(value)))
	}
}

func cacheListNativeOwners(t *testing.T, word int) (out [2][]byte) {
	t.Helper()
	for owner := range out {
		raw, free := alloc.MakeHigh([]byte{}, 65536+64*(word-4)+512)
		t.Cleanup(free)
		if word == 8 && uintptr(unsafe.Pointer(&raw[0])) <= math.MaxUint32 {
			t.Fatal("C graph owners above4GiB required")
		}
		out[owner] = raw
	}
	return
}

func cacheListNativeInitialize(t *testing.T, c cacheListNativeCase, raw [2][]byte, placement, word int) [2][]byte {
	t.Helper()
	original := cacheListNativeInput(c)
	for owner := range raw {
		cacheListNativeHash(t, original[owner], c.OwnerBefore[owner])
		for j := range raw[owner] {
			raw[owner][j] = byte(j*17 + c.Input.Seed + owner*43 + 7)
		}
		copy(raw[owner][placement+64*word:], original[owner][256:])
		for cell := 0; cell < 64; cell++ {
			value := binary.LittleEndian.Uint32(original[owner][4*cell:])
			cacheListNativePut(raw[owner], placement+cell*word, word, cacheListNativeProject(t, raw, placement, word, value))
		}
	}
	return original
}

func TestCacheListNativeOriginal(t *testing.T) {
	cases := cacheListNativeCases(t)
	word := int(unsafe.Sizeof(uintptr(0)))
	for route := 0; route < 3; route++ {
		for _, placement := range []int{0, 64, 512} {
			raw := cacheListNativeOwners(t, word)
			for _, c := range cases {
				t.Run(fmt.Sprintf("route%d/place%d/%s", route, placement, c.Input.Name), func(t *testing.T) {
					original := cacheListNativeInitialize(t, c, raw, placement, word)
					want := [2][]byte{append([]byte(nil), raw[0]...), append([]byte(nil), raw[1]...)}
					for _, event := range c.OrderedWrites {
						data, err := hex.DecodeString(event.Bytes)
						if err != nil || len(data) != 4 {
							t.Fatal("original pointer store bytes")
						}
						found := false
						for owner, base := range [2]uint32{0x21000000, 0x21100000} {
							if event.Address < base || event.Address >= base+256 || (event.Address-base)%4 != 0 {
								continue
							}
							offset := int(event.Address - base)
							copy(original[owner][offset:], data)
							cacheListNativePut(want[owner], placement+(offset/4)*word, word, cacheListNativeProject(t, raw, placement, word, binary.LittleEndian.Uint32(data)))
							found = true
						}
						if !found {
							t.Fatal("original write outside declared pointer-cell owners")
						}
					}
					for owner := range original {
						cacheListNativeHash(t, original[owner], c.OwnerAfter[owner])
					}
					list := unsafe.Pointer(&raw[c.Input.ListOwner][placement+word*c.Input.ListIndex])
					node := unsafe.Pointer(&raw[c.Input.NodeOwner][placement+word*c.Input.NodeIndex])
					entry := 0
					if c.Input.Entry == "remove" {
						entry = 1
					} else if c.Input.Entry != "prepend" {
						t.Fatal("unknown original entry")
					}
					got := cacheListNativeCall(route, entry, list, node)
					if expected := cacheListNativeProject(t, raw, placement, word, c.ReturnPointer); got != expected {
						t.Fatalf("full native return got%p want%p", got, expected)
					}
					for owner := range raw {
						if !bytes.Equal(raw[owner], want[owner]) {
							t.Fatal("whole C graph, opaque tail or placement guards differ", owner)
						}
						for cell := 0; cell < 64; cell++ {
							value := cacheListNativeProject(t, raw, placement, word, c.PointerWordsAfter[64*owner+cell])
							if got := cacheListNativeReadC(unsafe.Pointer(&raw[owner][placement]), cell); got != value {
								t.Fatal("canonical C pointer reader differs", owner, cell)
							}
						}
					}
				})
			}
		}
	}
}

func TestCacheListNativeActualCacheSequence(t *testing.T) {
	word := int(unsafe.Sizeof(uintptr(0)))
	cache := cacheListNativeCacheC()
	foreign, free := alloc.MakeHigh([]byte{}, 65536+512)
	t.Cleanup(free)
	saved := append([]byte(nil), cache...)
	t.Cleanup(func() { copy(cache, saved) })
	for j := range foreign {
		foreign[j] = byte(j*7 + 19)
	}
	wantForeign := append([]byte(nil), foreign...)
	clear(cache)
	node := func(index int) unsafe.Pointer { return unsafe.Pointer(&cache[(2+3*index)*word]) }
	for i := 0; i < 16; i++ {
		cacheListNativePut(cache, (2+3*i)*word, word, unsafe.Pointer(&foreign[64*i]))
	}
	var order []int
	expect := func(before []byte) []byte {
		want := append([]byte(nil), before...)
		var first, last unsafe.Pointer
		if len(order) > 0 {
			first = node(order[0])
			last = node(order[len(order)-1])
		}
		cacheListNativePut(want, 50*word, word, first)
		cacheListNativePut(want, 51*word, word, last)
		for position, index := range order {
			var prev, next unsafe.Pointer
			if position > 0 {
				prev = node(order[position-1])
			}
			if position+1 < len(order) {
				next = node(order[position+1])
			}
			cacheListNativePut(want, (3+3*index)*word, word, prev)
			cacheListNativePut(want, (4+3*index)*word, word, next)
		}
		return want
	}
	verify := func(want []byte) {
		t.Helper()
		if !bytes.Equal(cache, want) || !bytes.Equal(foreign, wantForeign) {
			t.Fatal("actual extracted C cache or opaque value owners differ")
		}
		for cell := 0; cell < 52; cell++ {
			var bits uint64
			if word == 4 {
				bits = uint64(binary.LittleEndian.Uint32(want[cell*word:]))
			} else {
				bits = binary.LittleEndian.Uint64(want[cell*word:])
			}
			if uint64(uintptr(cacheListNativeReadC(unsafe.Pointer(&cache[0]), cell))) != bits {
				t.Fatal("actual C cache canonical pointer reader", cell)
			}
		}
	}
	for i := 0; i < 16; i++ {
		before := append([]byte(nil), cache...)
		if got := cacheListNativeCall(i%3, 0, unsafe.Pointer(&cache[50*word]), node(i)); got != node(i) {
			t.Fatal("C cache prepend identity", i)
		}
		order = append([]int{i}, order...)
		verify(expect(before))
	}
	for step, index := range []int{7, 15, 0, 8, 3, 12, 1, 14, 4, 11, 2, 13, 5, 10, 6, 9} {
		before := append([]byte(nil), cache...)
		position := -1
		for j, v := range order {
			if v == index {
				position = j
				break
			}
		}
		if position < 0 {
			t.Fatal("sequence membership", index)
		}
		var result unsafe.Pointer
		if position > 0 {
			result = node(index)
		} else if len(order) > 1 {
			result = node(order[1])
		}
		if got := cacheListNativeCall(step%3, 1, unsafe.Pointer(&cache[50*word]), node(index)); got != result {
			t.Fatal("C cache remove return identity", index)
		}
		order = append(order[:position], order[position+1:]...)
		verify(expect(before))
	}
}
