package legacy

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type killableRow528190 struct {
	Name  string `json:"name"`
	Input struct {
		Cur    uint16 `json:"current"`
		Max    uint16 `json:"maximum"`
		Middle uint16 `json:"middle"`
		Tail   uint16 `json:"tail"`
		Class  uint32 `json:"unit_class"`
		Health bool   `json:"health_link"`
		Nil    bool   `json:"nil_unit"`
	} `json:"input"`
	Output struct {
		Return      uint32           `json:"return_word"`
		Heap        bool             `json:"whole_heap_unchanged"`
		Stack       bool             `json:"whole_caller_stack_unchanged"`
		CW          bool             `json:"CW_TOP_preserved"`
		Nonvolatile bool             `json:"nonvolatile_preserved"`
		Fault       []map[string]any `json:"fault"`
		State       struct {
			Health string `json:"health_hex"`
		} `json:"state"`
	} `json:"output"`
}

func killableLoad528190(t *testing.T) ([]killableRow528190, []killableRow528190) {
	t.Helper()
	file := filepath.Join("..", "server", "testdata", "object_killable_original_528190.json.gz")
	if supplied := os.Getenv("NOX_KILLABLE_ORIGINAL_REFERENCE"); supplied != "" {
		file = supplied
	}
	b, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	if fmt.Sprintf("%x", sha256.Sum256(b)) != "43c192733569c7b8e413805f8e8a20bbbdd34c7665eda62f8a893e2e864468af" {
		t.Fatal("compressed original seal")
	}
	z, e := gzip.NewReader(bytes.NewReader(b))
	if e != nil {
		t.Fatal(e)
	}
	raw, e := io.ReadAll(z)
	if e != nil {
		t.Fatal(e)
	}
	if e = z.Close(); e != nil {
		t.Fatal(e)
	}
	if fmt.Sprintf("%x", sha256.Sum256(raw)) != "2734e78743632c23916542c1e8c4da35ac5c390fbfa06358d728b78773c81f7c" {
		t.Fatal("decoded original seal")
	}
	var d struct {
		Contexts []killableRow528190 `json:"contexts"`
		Faults   []killableRow528190 `json:"expected_fault_contexts"`
	}
	if e = json.Unmarshal(raw, &d); e != nil {
		t.Fatal(e)
	}
	if len(d.Contexts) != 147 || len(d.Faults) != 1 {
		t.Fatal("original dimensions")
	}
	return d.Contexts, d.Faults
}

type killableObjectGuard528190 struct {
	Before [32]byte
	Value  server.Object
	After  [32]byte
}
type killableHealthGuard528190 struct {
	Before [32]byte
	Value  server.HealthData
	After  [32]byte
}

func killableBytes528190(p unsafe.Pointer, n uintptr) []byte {
	return append([]byte(nil), unsafe.Slice((*byte)(p), int(n))...)
}

func killableRun528190(t *testing.T, path int, fault bool) {
	rows, faults := killableLoad528190(t)
	if fault {
		rows = faults
	}
	for _, row := range rows {
		t.Run(row.Name, func(t *testing.T) {
			if !row.Output.Heap || !row.Output.Stack || !row.Output.CW {
				t.Fatal("original guards")
			}
			g, fg := alloc.NewHigh(killableObjectGuard528190{})
			h, fh := alloc.NewHigh(killableHealthGuard528190{})
			t.Cleanup(fg)
			t.Cleanup(fh)
			for i := 0; i < 32; i++ {
				g.Before[i] = byte(0x91 + i)
				g.After[i] = byte(0x51 + i)
				h.Before[i] = byte(0x71 + i)
				h.After[i] = byte(0x31 + i)
			}
			if unsafe.Sizeof(uintptr(0)) == 8 {
				for _, p := range []unsafe.Pointer{unsafe.Pointer(g), unsafe.Pointer(h)} {
					if uintptr(p) <= math.MaxUint32 {
						t.Fatal("guarded native records must exceed4GiB")
					}
				}
			}
			hp := &h.Value
			for i := range unsafe.Slice((*byte)(unsafe.Pointer(hp)), int(unsafe.Sizeof(*hp))) {
				unsafe.Slice((*byte)(unsafe.Pointer(hp)), int(unsafe.Sizeof(*hp)))[i] = 0xa5
			}
			hp.Cur, hp.Field2, hp.Max = row.Input.Cur, row.Input.Middle, row.Input.Max
			*(*uint16)(unsafe.Add(unsafe.Pointer(hp), 6)) = row.Input.Tail
			u := &g.Value
			u.ObjClass = object.Class(row.Input.Class)
			if row.Input.Health {
				u.HealthData = hp
			}
			reference, e := hex.DecodeString(row.Output.State.Health)
			if e != nil {
				t.Fatal(e)
			}
			if len(reference) != 64 || !bytes.Equal(killableBytes528190(unsafe.Pointer(hp), unsafe.Sizeof(*hp)), reference[:unsafe.Sizeof(*hp)]) {
				t.Fatal("native input health layout versus original words")
			}
			beforeG, beforeH := killableBytes528190(unsafe.Pointer(g), unsafe.Sizeof(*g)), killableBytes528190(unsafe.Pointer(h), unsafe.Sizeof(*h))
			if row.Input.Nil {
				u = nil
			}
			var got int32
			var caught any
			func() {
				defer func() { caught = recover() }()
				switch path {
				case 0:
					got = server.ObjectKillableNative528190(u)
				case 1:
					got = killableOriginalCEntry528190(u)
				case 2:
					got = int32(Nox_xxx_checkIsKillable_528190(u))
				default:
					t.Fatal("consumer path")
				}
			}()
			if fault {
				if caught == nil || len(row.Output.Fault) == 0 {
					t.Fatal("expected original raw nil-unit fault", caught)
				}
			} else if caught != nil || uint32(got) != row.Output.Return || !row.Output.Nonvolatile {
				t.Fatal("original root result", got, caught, row.Output.Return)
			}
			if !bytes.Equal(beforeG, killableBytes528190(unsafe.Pointer(g), unsafe.Sizeof(*g))) || !bytes.Equal(beforeH, killableBytes528190(unsafe.Pointer(h), unsafe.Sizeof(*h))) {
				t.Fatal("complete native record/neighbor mutation")
			}
		})
	}
}
func TestKillableOriginal528190RawGo(t *testing.T) { killableRun528190(t, 0, false) }
func TestKillableOriginal528190RawC(t *testing.T)  { killableRun528190(t, 1, false) }
func TestKillableOriginal528190GuardedGo(t *testing.T) {
	killableRun528190(t, 2, false)
	t.Run("existing_nil_policy", func(t *testing.T) {
		if Nox_xxx_checkIsKillable_528190(nil) != 0 {
			t.Fatal("existing guarded nil-unit result")
		}
	})
}
func TestKillableOriginalFault528190RawGo(t *testing.T) { killableRun528190(t, 0, true) }
func TestKillableOriginalFault528190RawC(t *testing.T)  { killableRun528190(t, 1, true) }
