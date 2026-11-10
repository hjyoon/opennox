package legacy

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/opennox/opennox/v1/common/memmap"
	"os"
	"testing"
	"unsafe"
)

// This root uses only the actual public C entry and private preserved C copy.
// It can run before installing the new Go algorithm files. After installation
// the same assertions exercise the public C -> Go path against that reference.
func TestGame2ServerBeforeCCurrentReference(t *testing.T) {
	type observation struct {
		Entry         uint32 `json:"entry"`
		A, B          int32
		Return        uint32 `json:"return_word"`
		PointerResult bool   `json:"pointer_result"`
		NativeReturn  uint64 `json:"native_return_bits"`
		Bytes         []byte `json:"after"`
	}
	var observed []observation
	type input struct {
		entry         uint32
		a, b          int32
		pointerResult bool
	}
	var cases []input
	for _, mask := range []int32{1, 0x80, -1, -2147483648} {
		for _, set := range []int32{0, 1} {
			cases = append(cases, input{0x453620, mask, set, true}, input{0x453640, mask, set, true})
		}
	}
	for _, index := range []int32{0, 31, 32, 63, 127, 136, 159} {
		for _, set := range []int32{0, 1} {
			cases = append(cases, input{0x453FA0, index, set, false})
		}
	}
	for i, item := range cases {
		t.Run(fmt.Sprintf("%02d-%x-%08x-%d", i, item.entry, uint32(item.a), item.b), func(t *testing.T) {
			const size = 64
			var after [2][]byte
			var returns [2]uint32
			var nativeReturn uintptr
			for backend := 0; backend < 2; backend++ {
				p := game2ServerFixtureAlloc(size)
				if p == nil {
					t.Fatal("C control allocation")
				}
				defer game2ServerFixtureFree(p)
				storage := unsafe.Slice((*byte)(p), size)
				for j := range storage {
					storage[j] = byte(0xA9 + j*13)
				}
				arg := unsafe.Add(p, 16)
				got := game2ServerFixtureCall(backend == 0, item.entry, arg, item.a, item.b)
				if item.pointerResult {
					if got != uintptr(arg) {
						t.Fatal("actual complete pointer result")
					}
				} else {
					returns[backend] = uint32(got)
				}
				if backend == 1 {
					nativeReturn = got
				}
				after[backend] = append([]byte(nil), storage...)
				if !bytes.Equal(storage[:16], after[0][:16]) || !bytes.Equal(storage[36:], after[0][36:]) {
					t.Fatal("guard bytes outside5DWORD window")
				}
			}
			if !bytes.Equal(after[0], after[1]) || returns[0] != returns[1] {
				t.Fatalf("current C/private-C mismatch entry=%x", item.entry)
			}
			// Independent addressed-width assertions distinguish BYTE1 from DWORD4.
			initial := make([]byte, size)
			for j := range initial {
				initial[j] = byte(0xA9 + j*13)
			}
			offset, width := 16, 4
			if item.entry == 0x453620 {
				width = 1
			}
			if item.entry == 0x453FA0 {
				offset += 4 * int(item.a/32)
			}
			for j := range initial {
				if j < offset || j >= offset+width {
					if after[1][j] != initial[j] {
						t.Fatal("operation escaped its declared scalar width")
					}
				}
			}
			observed = append(observed, observation{Entry: item.entry, A: item.a, B: item.b, Return: returns[1], PointerResult: item.pointerResult, NativeReturn: uint64(nativeReturn), Bytes: after[1]})
		})
	}
	saved := append([]byte(nil), memmap.Slice(0x5D4594, 1045440)[:96]...)
	defer copy(memmap.Slice(0x5D4594, 1045440)[:96], saved)
	for _, word := range []uint32{0, 0x01020304, 0x80000000, 0xFFFFFFFF} {
		var actual [2][96]byte
		for backend := 0; backend < 2; backend++ {
			copy(memmap.Slice(0x5D4594, 1045440)[:96], saved)
			p := game2ServerFixtureAlloc(4)
			if p == nil {
				t.Fatal("C word allocation")
			}
			binary.NativeEndian.PutUint32(unsafe.Slice((*byte)(p), 4), word)
			got := game2ServerFixtureCall(backend == 0, 0x4535E0, p, 0, 0)
			if got != uintptr(p) {
				t.Fatal("set mask keeps full source pointer")
			}
			if uint32(game2ServerFixtureCall(backend == 0, 0x453610, nil, 0, 0)) != binary.NativeEndian.Uint32(saved[16:20]) {
				t.Fatal("unrelatedDWORD changed")
			}
			if game2ServerFixtureCall(backend == 0, 0x453600, nil, 0, 0) != uintptr(memmap.PtrOff(0x5D4594, 1045452)) {
				t.Fatal("actual mask owner address")
			}
			copy(actual[backend][:], memmap.Slice(0x5D4594, 1045440)[:96])
			game2ServerFixtureFree(p)
		}
		if actual[0] != actual[1] || binary.NativeEndian.Uint32(actual[1][12:]) != word {
			t.Fatal("global DWORD current-C reference")
		}
	}
	if path := os.Getenv("OPENNOX_GAME2_SERVER_BEFORE_C_OUTPUT"); path != "" {
		b, e := json.MarshalIndent(struct {
			Schema       string        `json:"schema"`
			PointerBytes int           `json:"pointer_bytes"`
			Phase        string        `json:"phase_root_attested"`
			Observations []observation `json:"observations"`
		}{"opennox-game2-server-current-C-controls-v1", int(unsafe.Sizeof(uintptr(0))), os.Getenv("OPENNOX_GAME2_SERVER_CONTROL_PHASE"), observed}, "", "  ")
		if e != nil {
			t.Fatal(e)
		}
		if e = os.WriteFile(path, append(b, '\n'), 0600); e != nil {
			t.Fatal(e)
		}
	}
}
