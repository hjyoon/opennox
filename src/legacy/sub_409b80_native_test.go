//go:build darwin && arm64

package legacy

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"testing"
	"unsafe"
)

const sub409B80InputsSHA = "b3313d14cf5fec847639901104d012b282c330467102319681c7595cd2624d46"

type sub409B80Inputs struct {
	Original []struct {
		ID string `json:"id"`
	} `json:"original_accessor_cases"`
	Controls []uint32 `json:"FPU_controls"`
	Strings  []struct {
		ID  string `json:"id"`
		Hex string `json:"bytes_hex"`
	} `json:"native_strings"`
}

func sub409B80ReadGzip(t *testing.T, path string) ([]byte, []byte) {
	t.Helper()
	packed, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	z, err := gzip.NewReader(bytes.NewReader(packed))
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	raw, err := io.ReadAll(io.LimitReader(z, 4<<20))
	if err != nil || len(raw) == 4<<20 {
		t.Fatalf("bounded reference bytes: %v", err)
	}
	return packed, raw
}

func sub409B80LoadInputs(t *testing.T) sub409B80Inputs {
	t.Helper()
	packed, raw := sub409B80ReadGzip(t, "testdata/sub_409b80_inputs-v1.json.gz")
	if fmt.Sprintf("%x", sha256.Sum256(packed)) != sub409B80InputsSHA {
		t.Fatal("independent input byte binding")
	}
	var in sub409B80Inputs
	if err := json.Unmarshal(raw, &in); err != nil {
		t.Fatal(err)
	}
	if len(in.Original) != 8 || len(in.Controls) != 2 || len(in.Strings) != 6 {
		t.Fatal("exact independent domains")
	}
	return in
}

func TestSub409B80OriginalSixByteReference(t *testing.T) {
	in := sub409B80LoadInputs(t)
	_, raw := sub409B80ReadGzip(t, "testdata/sub_409b80_original-v1.json.gz")
	var ref struct {
		Schema string                             `json:"schema"`
		Source struct{ Exe, Body, Inputs string } `json:"source"`
		Rows   []struct {
			ID           string                                      `json:"case_id"`
			Control      uint32                                      `json:"FPU_control"`
			Returned     bool                                        `json:"returned"`
			Before       map[string]uint32                           `json:"registers_before"`
			After        map[string]uint32                           `json:"registers_after"`
			FPUBefore    json.RawMessage                             `json:"fpu_before"`
			FPUAfter     json.RawMessage                             `json:"fpu_after"`
			CallerBefore string                                      `json:"caller_before_hex"`
			CallerAfter  string                                      `json:"caller_after_hex"`
			Instructions []uint32                                    `json:"executed_instructions"`
			Memory       []struct{ Name, Before, After, SHA string } `json:"memory"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(raw, &ref); err != nil {
		t.Fatal(err)
	}
	if ref.Schema != "opennox-sub409b80-original-reference-v1" || ref.Source.Exe != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || ref.Source.Body != "2171b26dc3b3aa41e6ec1720367d741f96c119ebcb5d114d8851f6e0c3f8ff34" || ref.Source.Inputs != sub409B80InputsSHA {
		t.Fatal("original six-byte source/input certificate")
	}
	want := map[string]bool{}
	for _, c := range in.Original {
		for _, cw := range in.Controls {
			want[fmt.Sprintf("%s/%x", c.ID, cw)] = true
		}
	}
	if len(ref.Rows) != len(want) {
		t.Fatal("exact16 original calls")
	}
	for _, row := range ref.Rows {
		key := fmt.Sprintf("%s/%x", row.ID, row.Control)
		if !want[key] {
			t.Fatalf("unexpected/duplicate original row %s", key)
		}
		delete(want, key)
		if !row.Returned || row.After["eax"] != 0x5D5310 || row.After["esp"] != row.Before["esp"]+4 || row.After["eip"] != 0x22000100 {
			t.Fatalf("actual original return/caller %s", key)
		}
		for _, reg := range []string{"ebx", "ecx", "edx", "esi", "edi", "ebp", "eflags"} {
			if row.Before[reg] != row.After[reg] {
				t.Fatalf("preserved register %s/%s", key, reg)
			}
		}
		var actualFPU struct {
			Control uint32 `json:"control"`
		}
		if err := json.Unmarshal(row.FPUBefore, &actualFPU); err != nil || actualFPU.Control != row.Control {
			t.Fatal("actual inherited FPU control")
		}
		if !bytes.Equal(row.FPUBefore, row.FPUAfter) || row.CallerBefore != row.CallerAfter || len(row.Instructions) != 2 || row.Instructions[0] != 0x409B80 || row.Instructions[1] != 0x409B85 {
			t.Fatalf("FPU/caller/complete original instruction certificate %s", key)
		}
		if len(row.Memory) < 3 || len(row.Memory) > 4 {
			t.Fatal("bounded whole mapped memory")
		}
		for _, region := range row.Memory {
			b, e := hex.DecodeString(region.Before)
			if e != nil || region.Before != region.After || fmt.Sprintf("%x", sha256.Sum256(b)) != region.SHA {
				t.Fatalf("whole memory %s/%s", key, region.Name)
			}
		}
	}
}

func TestSub409B80NativeConsumers(t *testing.T) {
	in := sub409B80LoadInputs(t)
	// These process-local globals are tested serially. The24-byte fixture
	// window is not a declaration of the production string's capacity.
	for _, c := range in.Strings {
		t.Run(c.ID, func(t *testing.T) {
			payload, err := hex.DecodeString(c.Hex)
			if err != nil {
				t.Fatal(err)
			}
			nul := bytes.IndexByte(payload, 0)
			if nul < 0 || nul+1 > 15 {
				t.Fatal("short independent NUL domain")
			}
			physical := sub409B80FixtureBlob()
			if uintptr(physical) <= 0xffffffff {
				t.Fatal("native C-owned storage must be above4GiB")
			}
			window := unsafe.Slice((*byte)(unsafe.Add(physical, -8)), 24)
			saved := append([]byte(nil), window...)
			defer func() { copy(window, saved) }()
			for i := range window {
				window[i] = 0xa5
			}
			if n := sub409B80FixtureSet(payload); n != uint32(nul+1) {
				t.Fatal("actual sibling setter return")
			}
			want := bytes.Repeat([]byte{0xa5}, 24)
			copy(want[8:], payload[:nul+1])
			if !bytes.Equal(window, want) {
				t.Fatal("actual setter firstNUL/prefix/suffix bytes")
			}
			first, second := sub409B80FixtureC(), sub409B80FixtureC()
			if first != physical || second != first || sub409B80Native() != physical {
				t.Fatal("actual C/native/live buffer alias")
			}
			if uint32(uintptr(physical)) != sub409B80FixtureBadU32() || uintptr(sub409B80FixtureBadU32()) == uintptr(physical) {
				t.Fatal("non-dereferenced uint32 loss control")
			}
			oldCopy := Sub_409B80()
			if oldCopy != string(payload[:nul]) {
				t.Fatal("Go firstNUL/case/rawUTF8 copy")
			}
			out := bytes.Repeat([]byte{0x7b}, 24)
			n := sub409B80FixtureCCopy(out)
			if n != nul+1 || !bytes.Equal(out[:n], payload[:nul+1]) || !bytes.Equal(out[n:], bytes.Repeat([]byte{0x7b}, 24-n)) {
				t.Fatal("actual compiled C caller char* and NUL copy")
			}
			replacement := []byte("next\x00")
			if sub409B80FixtureSet(replacement) != uint32(len(replacement)) || sub409B80FixtureC() != first || GoStringP(first) != "next" || Sub_409B80() != "next" || oldCopy != string(payload[:nul]) {
				t.Fatal("later writer visible through old C alias; old Go copy immutable")
			}
			copy(want[8:], replacement)
			if !bytes.Equal(window, want) {
				t.Fatal("only actual setter bytes changed")
			}
			t.Logf("SUB409B80_NATIVE_CONSUMER_OK case=%s native=%p C=%p NULbytes=%d old-C-already-pointer-typed=true truncation-control-not-dereferenced=true", c.ID, physical, first, nul+1)
		})
	}
}
