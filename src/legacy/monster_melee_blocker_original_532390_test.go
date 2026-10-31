package legacy

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"math"
	"os"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

type meleeBlockerOriginalRow532390 struct {
	Name  string `json:"name"`
	Input struct {
		Family           string `json:"family"`
		Direction        int16  `json:"direction"`
		UnitX            uint32 `json:"unit_x"`
		UnitY            uint32 `json:"unit_y"`
		CandidateX       uint32 `json:"candidate_x"`
		CandidateY       uint32 `json:"candidate_y"`
		Threshold        uint32 `json:"threshold"`
		Class            uint32 `json:"class"`
		Flags            uint32 `json:"flags"`
		InitialSelection uint32 `json:"initial_selection"`
	} `json:"input"`
	Output struct {
		Selection string `json:"selection"`
		Distance  uint32 `json:"distance"`
	} `json:"output"`
}

type meleeBlockerObjectGuard532390 struct {
	Before [64]byte
	Value  server.Object
	After  [64]byte
}

func TestMeleeBlockerOriginalActualC532390(t *testing.T) {
	path := os.Getenv("NOX_MELEE_BLOCKER_ORIGINAL_REFERENCE")
	if path == "" {
		path = "../server/testdata/monster_melee_blocker_original_532390.json.gz"
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(raw)
	if hex.EncodeToString(digest[:]) != "bac07bf7a76134589ee1ffd8da09ba9e5fd427971577eb494e4d3badfc11e905" {
		t.Fatal("unsealed original blocker reference")
	}
	z, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	var ref struct {
		Schema   string                          `json:"schema"`
		Original string                          `json:"original_sha256"`
		Control  uint16                          `json:"control_word"`
		Entry    uint32                          `json:"entry"`
		Extent   uint32                          `json:"extent"`
		Services []string                        `json:"declared_services"`
		Rows     []meleeBlockerOriginalRow532390 `json:"rows"`
	}
	if err := json.NewDecoder(z).Decode(&ref); err != nil {
		t.Fatal(err)
	}
	if ref.Schema != "original-monster-melee-blocker-532390-v1" || ref.Original != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || ref.Control != 0xe3f || ref.Entry != 0x532390 || ref.Extent != 176 || len(ref.Services) != 0 || len(ref.Rows) != 14112 {
		t.Fatal("original provenance and dimensions")
	}
	objects := make([]*meleeBlockerObjectGuard532390, 3)
	for i := range objects {
		value, free := alloc.New(meleeBlockerObjectGuard532390{})
		t.Cleanup(free)
		objects[i] = value
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(value)) <= math.MaxUint32 {
			t.Fatal("every native allocation must exceed4GiB")
		}
	}
	unit, candidate, prior := &objects[0].Value, &objects[1].Value, &objects[2].Value
	selected := memmap.PtrPtr(0x5d4594, 2487948)
	distance := memmap.PtrFloat32(0x5d4594, 2487952)
	global := unsafe.Slice((*byte)(unsafe.Add(unsafe.Pointer(distance), -68)), 136)
	oldSelected, oldGlobal := *selected, append([]byte(nil), global...)
	t.Cleanup(func() { copy(global, oldGlobal); *selected = oldSelected })
	control := os.Getenv("NOX_MELEE_BLOCKER_CONTROL_ONLY") == "true"
	for _, row := range ref.Rows {
		t.Run(row.Name, func(t *testing.T) {
			for i, guard := range objects {
				*guard = meleeBlockerObjectGuard532390{}
				for j := range guard.Before {
					guard.Before[j], guard.After[j] = byte(0xa1+i), byte(0x51+i)
				}
			}
			unit.PosVec = types.Pointf{X: math.Float32frombits(row.Input.UnitX), Y: math.Float32frombits(row.Input.UnitY)}
			unit.Direction1 = server.Dir16(uint16(row.Input.Direction))
			candidate.ObjClass, candidate.ObjFlags = object.Class(row.Input.Class), object.Flags(row.Input.Flags)
			candidate.PosVec = types.Pointf{X: math.Float32frombits(row.Input.CandidateX), Y: math.Float32frombits(row.Input.CandidateY)}
			for i := range global {
				global[i] = 0x6b
			}
			*selected = nil
			if row.Input.InitialSelection != 0 {
				*selected = unsafe.Pointer(prior)
			}
			*distance = math.Float32frombits(row.Input.Threshold)
			before := make([][]byte, len(objects))
			for i, value := range objects {
				before[i] = append([]byte(nil), unsafe.Slice((*byte)(unsafe.Pointer(value)), int(unsafe.Sizeof(*value)))...)
			}
			wantSelected, wantDistance := *selected, row.Input.Threshold
			wantGlobal := append([]byte(nil), global...)
			if !control {
				switch row.Output.Selection {
				case "zero":
					wantSelected = nil
				case "prior":
					wantSelected = unsafe.Pointer(prior)
				case "candidate":
					wantSelected = unsafe.Pointer(candidate)
				default:
					t.Fatal("invalid frozen original selected-pointer role")
				}
				wantDistance = row.Output.Distance
				binary.LittleEndian.PutUint32(wantGlobal[68:], wantDistance)
				if unsafe.Sizeof(uintptr(0)) == 4 {
					binary.LittleEndian.PutUint32(wantGlobal[64:], uint32(uintptr(wantSelected)))
				}
				ccall.CallVoidPtr2(meleePublicCBefore532130(0x532390), candidate.CObj(), unit.CObj())
			}
			if *selected != wantSelected || math.Float32bits(*distance) != wantDistance {
				t.Errorf("native selected/distance=%p/%08x; original=%p/%08x", *selected, math.Float32bits(*distance), wantSelected, wantDistance)
			}
			if !bytes.Equal(global, wantGlobal) {
				t.Error("packed global neighbor bytes changed")
			}
			for i, value := range objects {
				if !bytes.Equal(before[i], unsafe.Slice((*byte)(unsafe.Pointer(value)), int(unsafe.Sizeof(*value)))) {
					t.Errorf("native object%d or guards changed", i)
				}
			}
		})
	}
	t.Logf("original actual C melee-blocker contexts=%d control-only=%v native-selection-and-distance=true", len(ref.Rows), control)
}
