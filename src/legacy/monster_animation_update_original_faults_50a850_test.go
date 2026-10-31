package legacy

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

// These four original faulting inputs are deliberately separate from the
// valid-pointer return/memory comparator. Lookup clears the cached table,
// the actual child returns address 0x80, and its caller faults on a read.
func animationOriginalFaultDiagnostics50A850(t *testing.T, actualC bool) {
	path := os.Getenv("NOX_ANIMATION_UPDATE_ORIGINAL_REFERENCE")
	if path == "" {
		path = "../server/testdata/monster_animation_update_original_50A850.json.gz"
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(raw)
	if hex.EncodeToString(sum[:]) != "cfc6b738e4886bb2a83046bceb301d9e625780517ede7624845d43c5d6f3ba17" {
		t.Fatal("original fault fixture seal")
	}
	z, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	var ref struct {
		Diagnostics []json.RawMessage `json:"diagnostics"`
	}
	if err = json.NewDecoder(z).Decode(&ref); err != nil || len(ref.Diagnostics) != 4 {
		t.Fatal("original separate fault dimensions", err)
	}
	f := animationNewFixture50A850(t)
	for _, data := range ref.Diagnostics {
		var row animationRow50A850
		var fault struct {
			Output struct {
				Fault struct {
					Faults []struct {
						Address uint32 `json:"address"`
						Size    uint32 `json:"size"`
						PC      uint32 `json:"pc"`
					} `json:"faults"`
				} `json:"fault"`
			} `json:"output"`
		}
		if err := json.Unmarshal(data, &row); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(data, &fault); err != nil {
			t.Fatal(err)
		}
		if len(fault.Output.Fault.Faults) != 1 || fault.Output.Fault.Faults[0].Address != 0x89 || fault.Output.Fault.Faults[0].Size != 1 || fault.Output.Fault.Faults[0].PC != 0x50a89e || len(row.Output.Returns) != 0 || len(row.Output.ChildReturns) != 1 || row.Output.ChildReturns[0].Pointer != (NPCAnimationReturn50A850{Role: "nil", Offset: 0x80}) {
			t.Fatal("original fault contract differs")
		}
		t.Run(row.Name, func(t *testing.T) {
			f.t = t
			f.prepare(&row)
			f.call, f.recording = 0, true
			var recovered any
			func() {
				defer func() { recovered = recover() }()
				if actualC {
					animationUpdatePublicC50A850(&f.unit.Value)
				} else {
					server.MonsterAnimationUpdateNative50A850(&f.unit.Value, f.root)
				}
			}()
			f.recording = false
			runtimeError, ok := recovered.(runtime.Error)
			if !ok || !strings.Contains(runtimeError.Error(), "invalid memory address") {
				t.Fatalf("native original fault missing or different: %v", recovered)
			}
			if !reflect.DeepEqual(f.children, []animationUpdateNativeChild50A850{{Call: 0, Pointer: 0x80}}) || !reflect.DeepEqual(f.predicates, row.Output.Predicates) || !reflect.DeepEqual(f.lookups, row.Output.Lookups) || !reflect.DeepEqual(f.mutations, row.Output.Mutations) || !reflect.DeepEqual(f.helpers, row.Output.Helpers) || f.caches() != row.Output.Caches {
				t.Fatal("native fault did not follow original actual child/predicate/lookup path")
			}
			animationPatchValue50A850(t, f.before, f.regions, &f.updates[0].Value.Field119, (*[16]server.MonsterAnim)(nil))
			for i := range f.cache.Value {
				animationPatchValue50A850(t, f.before, f.regions, &f.cache.Value[i], row.Output.Caches[i+1])
			}
			*(*uint32)(unsafe.Pointer(&f.mimicGuard[64])) = row.Output.Caches[0]
			for i, region := range f.regions {
				if !bytes.Equal(f.before[i], unsafe.Slice((*byte)(region.Ptr), int(region.Size))) {
					t.Errorf("fault guard region%d changed beyond original lookup effects", i)
				}
			}
			if !bytes.Equal(f.scratchBefore, animationSnapshot50A850(unsafe.Add(memmap.PtrOff(0x5D4594, 2487964), -32), 80)) || !bytes.Equal(f.mimicGuard, animationSnapshot50A850(unsafe.Add(statePredicateOriginalMimicCachePointer534840(), -64), 132)) {
				t.Error("fault scratch/cache neighbor guards changed")
			}
		})
	}
	t.Logf("original separate animation-update faults=4 actual-C=%t actual-child-pointer=0x80 original-read-address=0x89 native-read-fault-and-full-guards=true", actualC)
}

func TestAnimationUpdateOriginalFaultActualC50A850(t *testing.T) {
	animationOriginalFaultDiagnostics50A850(t, true)
}

func TestAnimationUpdateOriginalFaultGo50A850(t *testing.T) {
	animationOriginalFaultDiagnostics50A850(t, false)
}
