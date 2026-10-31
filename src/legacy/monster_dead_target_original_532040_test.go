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
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

type deadTargetOriginalRow532040 struct {
	Name  string `json:"name"`
	Input struct {
		NetCode     uint32 `json:"candidate_netcode"`
		ContextCode uint32 `json:"context_netcode"`
		Flags       uint32 `json:"flags"`
		Initial     uint32 `json:"initial_global"`
		Class       uint32 `json:"class"`
		OtherCode   uint32 `json:"other_update_netcode"`
	} `json:"input"`
	Output struct {
		Returned struct {
			Kind string `json:"kind"`
		} `json:"return_value"`
		Global    uint32 `json:"global"`
		Unchanged bool   `json:"complete_candidate_context_other_record_and_guards_unchanged"`
	} `json:"output"`
}

type deadTargetObjectGuard532040 struct {
	Before [64]byte
	Value  server.Object
	After  [64]byte
}

type deadTargetUpdateGuard532040 struct {
	Before [64]byte
	Value  server.MonsterUpdateData
	After  [64]byte
}

func deadTargetSnapshot532040(p unsafe.Pointer, size uintptr) []byte {
	return append([]byte(nil), unsafe.Slice((*byte)(p), int(size))...)
}

func TestDeadTargetOriginalActualC532040(t *testing.T) {
	path := os.Getenv("NOX_DEAD_TARGET_ORIGINAL_REFERENCE")
	if path == "" {
		path = "../server/testdata/monster_dead_target_original_532040.json.gz"
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(raw)
	if hex.EncodeToString(digest[:]) != "3c33673b07ab5f97e070493a3af6d0af720334680336eae6f5fc2297f6505819" {
		t.Fatal("unsealed original dead-target reference")
	}
	z, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	var ref struct {
		Schema   string                        `json:"schema"`
		Original string                        `json:"original_sha256"`
		Entry    uint32                        `json:"entry"`
		Extent   uint32                        `json:"extent"`
		CodeSHA  string                        `json:"code_sha256"`
		Global   uint32                        `json:"global_address"`
		Control  uint16                        `json:"control_word"`
		Rows     []deadTargetOriginalRow532040 `json:"rows"`
	}
	if err := json.NewDecoder(z).Decode(&ref); err != nil {
		t.Fatal(err)
	}
	if ref.Schema != "original-monster-dead-target-532040-v1" || ref.Original != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" ||
		ref.Entry != 0x532040 || ref.Extent != 48 || ref.CodeSHA != "73bc67defadf3daee1a979ef5343bf73b8170bec5b0808f2e8d113a46795a80b" ||
		ref.Global != 0x833c1c || ref.Control != 0xe3f || len(ref.Rows) != 10800 {
		t.Fatal("original provenance/extent/dimensions")
	}
	candidate, freeCandidate := alloc.New(deadTargetObjectGuard532040{})
	t.Cleanup(freeCandidate)
	context, freeContext := alloc.New(deadTargetUpdateGuard532040{})
	t.Cleanup(freeContext)
	other, freeOther := alloc.New(deadTargetUpdateGuard532040{})
	t.Cleanup(freeOther)
	regions := []struct {
		Pointer unsafe.Pointer
		Size    uintptr
	}{{unsafe.Pointer(candidate), unsafe.Sizeof(*candidate)}, {unsafe.Pointer(context), unsafe.Sizeof(*context)}, {unsafe.Pointer(other), unsafe.Sizeof(*other)}}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, region := range regions {
			if uintptr(region.Pointer) <= math.MaxUint32 {
				t.Fatal("native candidate and both context allocations must exceed4GiB")
			}
		}
	}
	flag := memmap.PtrUint32(0x5d4594, 2487944)
	global := unsafe.Slice((*byte)(unsafe.Add(unsafe.Pointer(flag), -64)), 132)
	oldGlobal := append([]byte(nil), global...)
	t.Cleanup(func() { copy(global, oldGlobal) })
	control := os.Getenv("NOX_DEAD_TARGET_CONTROL_ONLY") == "true"
	for _, row := range ref.Rows {
		t.Run(row.Name, func(t *testing.T) {
			candidate.Value = server.Object{ObjClass: object.Class(row.Input.Class), ObjFlags: object.Flags(row.Input.Flags), NetCode: row.Input.NetCode, UpdateData: unsafe.Pointer(&other.Value)}
			context.Value = server.MonsterUpdateData{Field300: row.Input.ContextCode}
			other.Value = server.MonsterUpdateData{Field300: row.Input.OtherCode}
			for i := range candidate.Before {
				candidate.Before[i], candidate.After[i] = 0xa5, 0x5a
				context.Before[i], context.After[i] = 0x91, 0x19
				other.Before[i], other.After[i] = 0xc3, 0x3c
			}
			for i := range global {
				global[i] = 0x6b
			}
			*flag = row.Input.Initial
			before := make([][]byte, len(regions))
			for i, region := range regions {
				before[i] = deadTargetSnapshot532040(region.Pointer, region.Size)
			}
			wantGlobal := append([]byte(nil), global...)
			if row.Output.Returned.Kind != "candidate" || !row.Output.Unchanged {
				t.Fatal("invalid frozen original return/read-only contract")
			}
			if !control {
				result := ccall.CallUPtrPtr2(deadTargetPublicCBefore532040(), candidate.Value.CObj(), unsafe.Pointer(&context.Value))
				if result != uintptr(unsafe.Pointer(&candidate.Value)) {
					t.Errorf("native return=%#x, original candidate role=%p", result, &candidate.Value)
				}
				*(*uint32)(unsafe.Pointer(&wantGlobal[64])) = row.Output.Global
			}
			if !bytes.Equal(global, wantGlobal) {
				t.Errorf("sticky global/neighbor guards differ: got=%#x original=%#x", *flag, row.Output.Global)
			}
			for i, region := range regions {
				if !bytes.Equal(before[i], unsafe.Slice((*byte)(region.Pointer), int(region.Size))) {
					t.Errorf("native candidate/context/other-update region%d or guards changed", i)
				}
			}
		})
	}
	t.Logf("original actual C dead-target contexts=%d control-only=%t pointer-return-and-explicit-context=true", len(ref.Rows), control)
}
