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

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type missileLeadOriginalRow533080 struct {
	Index        int       `json:"index"`
	Position     [4]uint32 `json:"position_bits"`
	Velocity     [2]uint32 `json:"velocity_bits"`
	Speed        uint32    `json:"speed_bits"`
	Output       [2]uint32 `json:"output_bits"`
	ReturnRole   string    `json:"return_role"`
}

type missileLeadOriginalObjectGuard533080 struct {
	Before [4]uint64
	Value  server.Object
	After  [4]uint64
}

type missileLeadOriginalOutputGuard533080 struct {
	Before [4]uint64
	Value  types.Pointf
	After  [4]uint64
}

func TestMissileLeadOriginalActualC533080(t *testing.T) {
	path := os.Getenv("NOX_MONSTER_MISSILE_LEAD_ORIGINAL_FIXTURE")
	if path == "" {
		path = filepath.Join("..", "server", "testdata", "monster_missile_lead_original_533080.json.gz")
	}
	blob, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(blob)
	if hex.EncodeToString(digest[:]) != "d07ab9e0373bf9e075714de6804e1f43091f1169ca7b112efe5d7cef87996fe9" {
		t.Fatal("frozen original missile prediction reference changed")
	}
	z, err := gzip.NewReader(bytes.NewReader(blob))
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
	var ref struct {
		Entry   string                         `json:"entry"`
		Control uint32                         `json:"control"`
		Rows    []missileLeadOriginalRow533080 `json:"rows"`
	}
	if err := json.Unmarshal(raw, &ref); err != nil {
		t.Fatal(err)
	}
	if ref.Entry != "0x533080" || ref.Control != 0xe3f || len(ref.Rows) != 2844 {
		t.Fatal("unexpected frozen original prediction scope")
	}
	unit, freeUnit := alloc.New(missileLeadOriginalObjectGuard533080{})
	target, freeTarget := alloc.New(missileLeadOriginalObjectGuard533080{})
	out, freeOutput := alloc.New(missileLeadOriginalOutputGuard533080{})
	t.Cleanup(freeUnit)
	t.Cleanup(freeTarget)
	t.Cleanup(freeOutput)
	guard := [4]uint64{0x0123456789abcdef, 0xfedcba9876543210, 0x1122334455667788, 0xaabbccddeeff0099}
	control := os.Getenv("NOX_MISSILE_LEAD_CONTROL_ONLY") == "1"
	copyBytes := func(p unsafe.Pointer, size uintptr) []byte {
		return bytes.Clone(unsafe.Slice((*byte)(p), int(size)))
	}
	for index, row := range ref.Rows {
		t.Run(fmt.Sprintf("%04d", index), func(t *testing.T) {
			if row.Index != index || row.ReturnRole != "target" {
				t.Fatal("original input or return identity changed")
			}
			*unit, *target = missileLeadOriginalObjectGuard533080{Before: guard, After: guard}, missileLeadOriginalObjectGuard533080{Before: guard, After: guard}
			*out = missileLeadOriginalOutputGuard533080{Before: guard, After: guard, Value: types.Pointf{X: -9, Y: 11}}
			unit.Value.PosVec = types.Pointf{X: math.Float32frombits(row.Position[0]), Y: math.Float32frombits(row.Position[1])}
			target.Value.PosVec = types.Pointf{X: math.Float32frombits(row.Position[2]), Y: math.Float32frombits(row.Position[3])}
			target.Value.VelVec = types.Pointf{X: math.Float32frombits(row.Velocity[0]), Y: math.Float32frombits(row.Velocity[1])}
			for _, p := range []unsafe.Pointer{unsafe.Pointer(&unit.Value), unsafe.Pointer(&target.Value), unsafe.Pointer(&out.Value)} {
				if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(p) <= math.MaxUint32 {
					t.Fatal("C-owned native arguments must exceed4GiB")
				}
			}
			beforeUnit := copyBytes(unsafe.Pointer(unit), unsafe.Sizeof(*unit))
			beforeTarget := copyBytes(unsafe.Pointer(target), unsafe.Sizeof(*target))
			if !control {
				result := missileLeadPublicCBefore533080(unit.Value.CObj(), target.Value.CObj(), math.Float32frombits(row.Speed), unsafe.Pointer(&out.Value))
				if result != uintptr(unsafe.Pointer(&target.Value)) {
					t.Errorf("native target pointer return = %#x, want %p", result, &target.Value)
				}
				got := [2]uint32{math.Float32bits(out.Value.X), math.Float32bits(out.Value.Y)}
				if got != row.Output {
					t.Errorf("original prediction output = %08x, want %08x", got, row.Output)
				}
			} else if out.Value != (types.Pointf{X: -9, Y: 11}) {
				t.Fatal("allocation-only control changed its output")
			}
			if !bytes.Equal(beforeUnit, copyBytes(unsafe.Pointer(unit), unsafe.Sizeof(*unit))) ||
				!bytes.Equal(beforeTarget, copyBytes(unsafe.Pointer(target), unsafe.Sizeof(*target))) || out.Before != guard || out.After != guard {
				t.Fatal("original read-only argument or output neighbor changed")
			}
		})
	}
	t.Logf("original actual C missile-lead contexts=%d control-only=%v native-target-return=true", len(ref.Rows), control)
}
