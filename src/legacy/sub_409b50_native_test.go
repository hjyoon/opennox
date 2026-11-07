//go:build darwin && arm64

package legacy

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"testing"
	"unsafe"
)

const sub409B50InputsSHA = "f8b02862d2ffc239243310d8e6ac2a1b7309552f4529973a6a61518f1782eb33"

type sub409B50Case struct {
	ID        string `json:"id"`
	Source    string `json:"source"`
	Hex       string `json:"bytes_hex"`
	Alignment int    `json:"source_alignment_offset"`
	Delta     int    `json:"source_delta"`
	Native    bool   `json:"native_consumer"`
}
type sub409B50Row struct {
	ID           string            `json:"case_id"`
	CW           uint32            `json:"FPU_control"`
	Returned     bool              `json:"returned"`
	Count        uint32            `json:"actual_result_u32"`
	Valid        bool              `json:"certificate_valid"`
	Replay       bool              `json:"memory_replay_exact"`
	Before       map[string]uint32 `json:"registers_before"`
	After        map[string]uint32 `json:"registers_after"`
	FPUBefore    json.RawMessage   `json:"fpu_before"`
	FPUAfter     json.RawMessage   `json:"fpu_after"`
	CallerBefore string            `json:"caller_before_hex"`
	CallerAfter  string            `json:"caller_after_hex"`
	Faults       []json.RawMessage `json:"faults"`
	Events       []struct {
		Kind     string `json:"kind"`
		Address  uint64 `json:"address"`
		Width    int    `json:"width"`
		Value    uint64 `json:"value"`
		Complete bool   `json:"complete"`
	} `json:"events"`
	Memory []struct {
		Name      string `json:"name"`
		Address   uint64 `json:"address"`
		Before    string `json:"before_hex"`
		After     string `json:"after_hex"`
		BeforeSHA string `json:"before_sha256"`
		AfterSHA  string `json:"after_sha256"`
	} `json:"memory"`
}

func sub409B50Read(t *testing.T, path string) ([]byte, []byte) {
	t.Helper()
	packed, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	z, e := gzip.NewReader(bytes.NewReader(packed))
	if e != nil {
		t.Fatal(e)
	}
	defer z.Close()
	raw, e := io.ReadAll(io.LimitReader(z, 32<<20))
	if e != nil || len(raw) == 32<<20 {
		t.Fatal("bounded original fixture", e)
	}
	return packed, raw
}
func sub409B50Load(t *testing.T) ([]sub409B50Case, []sub409B50Row) {
	t.Helper()
	packed, raw := sub409B50Read(t, "testdata/sub_409b50_inputs-v1.json.gz")
	if fmt.Sprintf("%x", sha256.Sum256(packed)) != sub409B50InputsSHA {
		t.Fatal("frozen independent inputs")
	}
	var in struct {
		Cases []sub409B50Case `json:"cases"`
	}
	if e := json.Unmarshal(raw, &in); e != nil {
		t.Fatal(e)
	}
	_, raw = sub409B50Read(t, "testdata/sub_409b50_original-v1.json.gz")
	var ref struct {
		Schema string                             `json:"schema"`
		Source struct{ Exe, Body, Inputs string } `json:"source"`
		Rows   []sub409B50Row                     `json:"rows"`
	}
	if e := json.Unmarshal(raw, &ref); e != nil {
		t.Fatal(e)
	}
	if ref.Schema != "opennox-sub409b50-original-reference-v1" || ref.Source.Exe != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || ref.Source.Body != "39f4f1088bbfd476b125556405887c8220aa5912fb70137f4620cc0ab6bf5ee4" || ref.Source.Inputs != sub409B50InputsSHA || len(in.Cases) != 25 || len(ref.Rows) != 50 {
		t.Fatal("complete41B/input/reference identity")
	}
	return in.Cases, ref.Rows
}
func TestSub409B50Original41ByteReference(t *testing.T) {
	cases, rows := sub409B50Load(t)
	wanted := map[string]bool{}
	for _, c := range cases {
		for _, cw := range []uint32{0xe3f, 0x23f} {
			wanted[fmt.Sprintf("%s/%x", c.ID, cw)] = true
		}
	}
	for _, r := range rows {
		key := fmt.Sprintf("%s/%x", r.ID, r.CW)
		if !wanted[key] || !r.Valid || !r.Replay {
			t.Fatal("original actual case certificate", key)
		}
		delete(wanted, key)
		var f struct {
			Control uint32 `json:"control"`
		}
		if json.Unmarshal(r.FPUBefore, &f) != nil || f.Control != r.CW || !bytes.Equal(r.FPUBefore, r.FPUAfter) || r.CallerBefore != r.CallerAfter {
			t.Fatal("actual FPU/caller preservation", key)
		}
		if r.Returned {
			if r.After["esp"] != r.Before["esp"]+4 || r.After["eip"] != 0x24000100 || r.After["ecx"] != 0 {
				t.Fatal("actual original returned frame", key)
			}
			for _, reg := range []string{"ebx", "ebp", "edx", "esi", "edi"} {
				if r.Before[reg] != r.After[reg] {
					t.Fatal("actual saved register", reg, key)
				}
			}
		} else if len(r.Faults) == 0 {
			t.Fatal("unqualified harness failure", key)
		}
		for _, m := range r.Memory {
			before, e := hex.DecodeString(m.Before)
			if e != nil || len(before) != 4096 || fmt.Sprintf("%x", sha256.Sum256(before)) != m.BeforeSHA {
				t.Fatal("full before memory", key)
			}
			after, e := hex.DecodeString(m.After)
			if e != nil || len(after) != 4096 || fmt.Sprintf("%x", sha256.Sum256(after)) != m.AfterSHA {
				t.Fatal("full after memory", key)
			}
			replay := append([]byte(nil), before...)
			for _, event := range r.Events {
				if event.Kind == "write" && event.Complete && m.Address <= event.Address && event.Address+uint64(event.Width) <= m.Address+4096 {
					if event.Width != 1 && event.Width != 4 {
						t.Fatal("original write width")
					}
					var bits [8]byte
					binary.LittleEndian.PutUint64(bits[:], event.Value)
					copy(replay[event.Address-m.Address:], bits[:event.Width])
				}
			}
			if !bytes.Equal(replay, after) {
				t.Fatal("whole original completed-write reconstruction", key, m.Name)
			}
		}
	}
	if len(wanted) != 0 {
		t.Fatal("missing original contexts")
	}
}
func sub409B50Window(t *testing.T, r sub409B50Row) ([]byte, []byte) {
	t.Helper()
	for _, m := range r.Memory {
		if m.Name == "destination" {
			b, e := hex.DecodeString(m.Before)
			if e != nil {
				t.Fatal(e)
			}
			a, e := hex.DecodeString(m.After)
			if e != nil {
				t.Fatal(e)
			}
			return b[0x300:0x360], a[0x300:0x360]
		}
	}
	t.Fatal("healthy destination snapshot missing")
	return nil, nil
}
func TestSub409B50NativeOrderAndAlias(t *testing.T) {
	cases, rows := sub409B50Load(t)
	references := map[string]sub409B50Row{}
	for _, r := range rows {
		if r.CW == 0xe3f {
			references[r.ID] = r
		} else {
			other, ok := references[r.ID]
			if ok && other.Count != r.Count {
				t.Fatal("nonarithmetic original count differs by inherited FPU")
			}
		}
	}
	for _, c := range cases {
		if !c.Native {
			continue
		}
		t.Run(c.ID, func(t *testing.T) {
			r := references[c.ID]
			if !r.Returned {
				t.Fatal("no healthy original certificate")
			}
			payload, e := hex.DecodeString(c.Hex)
			if e != nil {
				t.Fatal(e)
			}
			nul := bytes.IndexByte(payload, 0)
			if nul < 0 || r.Count != uint32(nul+1) {
				t.Fatal("independent source firstNUL/actual count")
			}
			physical := sub409B80FixtureBlob()
			if uintptr(physical) <= 0xffffffff || sub409B80FixtureC() != physical {
				t.Fatal("actual shared native C-owned alias")
			}
			window := unsafe.Slice((*byte)(unsafe.Add(physical, -16)), 96)
			saved := append([]byte(nil), window...)
			defer func() { copy(window, saved) }()
			before, after := sub409B50Window(t, r)
			chosen := bytes.Repeat([]byte{0xa5}, 96)
			if c.Source == "alias" {
				copy(chosen[16+c.Delta:], payload)
			}
			if !bytes.Equal(chosen, before) {
				t.Fatal("native input initialization must match independent original BEFORE")
			}
			run := func(backend string, exact bool) {
				copy(window, chosen)
				alignment := c.Alignment
				if exact {
					alignment = 0
				}
				var allocation []byte
				var source []byte
				if c.Source == "alias" {
					source = unsafe.Slice((*byte)(unsafe.Add(physical, c.Delta)), len(payload))
				} else {
					allocation = make([]byte, alignment+len(payload))
					copy(allocation[alignment:], payload)
					source = allocation[alignment:]
				}
				var count uint32
				if backend == "Go" {
					count = sub409B50Native(unsafe.Pointer(&source[0]))
				} else {
					count = sub409B80FixtureSet(source)
				}
				runtime.KeepAlive(source)
				runtime.KeepAlive(allocation)
				if count != r.Count || !bytes.Equal(window, after) || sub409B80FixtureC() != physical {
					t.Fatal("actual ordered DWORD/tail/overlap/return/alias", backend, exact)
				}
				if c.Source != "alias" && !bytes.Equal(source, payload) {
					t.Fatal("disjoint source was modified")
				}
				// Destructive forward overlap can erase NUL. Compare fixed raw
				// bytes above, then restore a fresh normal string before GoString.
				if sub409B80FixtureSet([]byte("next\x00")) != 5 || Sub_409B80() != "next" || sub409B80FixtureC() != physical {
					t.Fatal("fresh setter/getter regression after raw snapshot")
				}
				if exact {
					t.Logf("SUB409B50_GO_SOURCE_END_OK case=%s backend=%s bytes=%d", c.ID, backend, len(payload))
				} else {
					t.Logf("SUB409B50_NATIVE_ORDER_OK case=%s backend=%s source=%p destination=%p count=%d raw_snapshot_only=true", c.ID, backend, unsafe.Pointer(&source[0]), physical, count)
				}
			}
			for _, backend := range []string{"Go", "C"} {
				run(backend, false)
			}
			if c.Source == "separate" && (len(payload) == 1 || len(payload) == 16 || len(payload) == 32 || len(payload) == 64) {
				for _, backend := range []string{"Go", "C"} {
					run(backend, true)
				}
			}
		})
	}
}
