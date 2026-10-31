package legacy

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math"
	"os"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

type hitLatchRow532880 struct {
	Name  string `json:"name"`
	Input struct {
		Frame   uint32 `json:"frame"`
		Initial uint32 `json:"initial"`
		Class   uint32 `json:"class"`
		Flags   uint32 `json:"flags"`
		Slot    int    `json:"record_slot"`
		Other   uint32 `json:"other_initial"`
	} `json:"input"`
	Output struct {
		Kind      string   `json:"return_kind"`
		Timestamp uint32   `json:"timestamp"`
		Reads     []uint32 `json:"frame_reads"`
		Writes    []struct {
			Size  int    `json:"size"`
			Value uint32 `json:"value"`
		} `json:"timestamp_writes"`
		Unchanged bool `json:"complete_other_heap_stack_frame_guards_code_nonvolatile_and_FPU_unchanged"`
	} `json:"output"`
}
type hitLatchGuard532880[T any] struct {
	Before [64]byte
	Value  T
	After  [64]byte
}

func TestHitLatchOriginalActualC532880(t *testing.T) {
	path := os.Getenv("NOX_HIT_LATCH_ORIGINAL_REFERENCE")
	if path == "" {
		path = "../server/testdata/monster_hit_latch_original_532880.json.gz"
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(raw)
	if hex.EncodeToString(digest[:]) != "9b347f3e68f376d342b77e7e9426b2458baad964ed10b923cb707dcb9f19fc0d" {
		t.Fatal("unsealed original hit-latch reference")
	}
	z, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	var ref struct {
		Schema       string              `json:"schema"`
		Original     string              `json:"original_sha256"`
		Entry        uint32              `json:"entry"`
		Extent       int                 `json:"extent"`
		Body         int                 `json:"body_bytes"`
		Code         string              `json:"code_sha256"`
		FrameAddress uint32              `json:"frame_address"`
		Control      uint16              `json:"control_word"`
		Services     []string            `json:"declared_services"`
		Rows         []hitLatchRow532880 `json:"rows"`
	}
	if err := json.NewDecoder(z).Decode(&ref); err != nil {
		t.Fatal(err)
	}
	if ref.Schema != "original-monster-hit-latch-532880-v1" || ref.Original != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || ref.Entry != 0x532880 || ref.Extent != 48 || ref.Body != 33 || ref.Code != "94c767def73db06645ceed1b9f0114c90cc104c1d3fd7d79a220ac95eb7b52ed" || ref.FrameAddress != 0x84ea04 || ref.Control != 0xe3f || len(ref.Services) != 0 || len(ref.Rows) != 3072 {
		t.Fatal("original provenance/extent/dimensions")
	}
	unit, freeUnit := alloc.New(hitLatchGuard532880[server.Object]{})
	t.Cleanup(freeUnit)
	var updates [2]*hitLatchGuard532880[server.MonsterUpdateData]
	for i := range updates {
		v, free := alloc.New(hitLatchGuard532880[server.MonsterUpdateData]{})
		updates[i] = v
		t.Cleanup(free)
	}
	regions := []struct {
		P unsafe.Pointer
		N uintptr
	}{{unsafe.Pointer(unit), unsafe.Sizeof(*unit)}, {unsafe.Pointer(updates[0]), unsafe.Sizeof(*updates[0])}, {unsafe.Pointer(updates[1]), unsafe.Sizeof(*updates[1])}}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, r := range regions {
			if uintptr(r.P) <= math.MaxUint32 {
				t.Fatal("all native C allocations must exceed4GiB")
			}
		}
	}
	oldFrame := gameFrameHook
	t.Cleanup(func() { gameFrameHook = oldFrame })
	control := os.Getenv("NOX_HIT_LATCH_CONTROL_ONLY") == "true"
	for _, row := range ref.Rows {
		t.Run(row.Name, func(t *testing.T) {
			if row.Input.Slot < 0 || row.Input.Slot >= 2 || row.Output.Kind != "selected_update" || !row.Output.Unchanged {
				t.Fatal("original output/slot contract")
			}
			for _, r := range regions {
				for i := range unsafe.Slice((*byte)(r.P), int(r.N)) {
					unsafe.Slice((*byte)(r.P), int(r.N))[i] = 0xa5
				}
			}
			unit.Value = server.Object{ObjClass: object.Class(row.Input.Class), ObjFlags: object.Flags(row.Input.Flags), UpdateData: unsafe.Pointer(&updates[row.Input.Slot].Value)}
			updates[row.Input.Slot].Value = server.MonsterUpdateData{Field130: row.Input.Initial}
			updates[1-row.Input.Slot].Value = server.MonsterUpdateData{Field130: row.Input.Other}
			snapshots := make([][]byte, len(regions))
			for i, r := range regions {
				snapshots[i] = append([]byte(nil), unsafe.Slice((*byte)(r.P), int(r.N))...)
			}
			reads := []uint32{}
			gameFrameHook = func() uint32 { reads = append(reads, row.Input.Frame); return row.Input.Frame }
			if !control {
				got := ccall.CallUPtrPtr(hitLatchPublicC532880(), unit.Value.CObj())
				if want := uintptr(unsafe.Pointer(&updates[row.Input.Slot].Value)); got != want {
					t.Errorf("native return=%#x original selected update=%#x", got, want)
				}
				if !equalHitLatchReads532880(reads, row.Output.Reads) {
					t.Errorf("conditional frame reads=%v original=%v", reads, row.Output.Reads)
				}
				if updates[row.Input.Slot].Value.Field130 != row.Output.Timestamp {
					t.Errorf("timestamp=%#x original=%#x", updates[row.Input.Slot].Value.Field130, row.Output.Timestamp)
				}
				offset := unsafe.Offsetof(updates[row.Input.Slot].Value) + unsafe.Offsetof(updates[row.Input.Slot].Value.Field130)
				*(*uint32)(unsafe.Pointer(&snapshots[1+row.Input.Slot][offset])) = row.Output.Timestamp
			} else if len(reads) != 0 {
				t.Fatal("control invoked frame service")
			}
			for i, r := range regions {
				if !bytes.Equal(snapshots[i], unsafe.Slice((*byte)(r.P), int(r.N))) {
					t.Errorf("complete native object/update/other record/guards region%d differs", i)
				}
			}
		})
	}
	t.Logf("original actual C hit-latch contexts=%d control-only=%t native-cached-update-and-conditional-clock=true", len(ref.Rows), control)
}
func equalHitLatchReads532880(a, b []uint32) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
