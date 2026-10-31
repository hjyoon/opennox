package legacy

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestMissileLeadAliasesOriginal533080(t *testing.T) {
	path := os.Getenv("NOX_MISSILE_LEAD_ALIAS_ORIGINAL_REFERENCE")
	if path == "" {
		path = "../server/testdata/monster_missile_lead_aliases_original_533080.json.gz"
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(data)
	if hex.EncodeToString(digest[:]) != "ca2b239be3a4c861d78b3bb3d4b7a1c19aad4058b084840803a18d2233d746de" {
		t.Fatal("original alias reference changed")
	}
	z, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	var ref struct {
		Control uint32 `json:"control"`
		Rows    []struct {
			missileLeadOriginalRow533080
			Alias         string    `json:"alias"`
			UnitAfter     [2]uint32 `json:"unit_position_after_bits"`
			TargetAfter   [2]uint32 `json:"target_position_after_bits"`
			VelocityAfter [2]uint32 `json:"target_velocity_after_bits"`
		} `json:"rows"`
	}
	if err := json.NewDecoder(z).Decode(&ref); err != nil {
		t.Fatal(err)
	}
	if ref.Control != 0xe3f || len(ref.Rows) != 56 {
		t.Fatal("original alias scope")
	}
	u, freeU := alloc.New(missileLeadOriginalObjectGuard533080{})
	v, freeV := alloc.New(missileLeadOriginalObjectGuard533080{})
	o, freeO := alloc.New(missileLeadOriginalOutputGuard533080{})
	t.Cleanup(freeU)
	t.Cleanup(freeV)
	t.Cleanup(freeO)
	guard := [4]uint64{0x1122334455667788, 0x123456789abcdef0, 0xaabbccddeeff0099, 0x9988776655443322}
	copyBytes := func(p unsafe.Pointer, n uintptr) []byte { return bytes.Clone(unsafe.Slice((*byte)(p), int(n))) }
	bits := func(p types.Pointf) [2]uint32 { return [2]uint32{math.Float32bits(p.X), math.Float32bits(p.Y)} }
	oldServer := GetServer
	GetServer = func() Server { panic("pure native prediction unexpectedly required a server") }
	t.Cleanup(func() { GetServer = oldServer })
	for i, row := range ref.Rows {
		for _, mode := range []string{"Go", "C"} {
			t.Run(fmt.Sprintf("%03d/%s", i, mode), func(t *testing.T) {
				*u = missileLeadOriginalObjectGuard533080{Before: guard, After: guard}
				*v = missileLeadOriginalObjectGuard533080{Before: guard, After: guard}
				*o = missileLeadOriginalOutputGuard533080{Before: guard, After: guard}
				u.Value.PosVec = types.Ptf(math.Float32frombits(row.Position[0]), math.Float32frombits(row.Position[1]))
				v.Value.PosVec = types.Ptf(math.Float32frombits(row.Position[2]), math.Float32frombits(row.Position[3]))
				v.Value.VelVec = types.Ptf(math.Float32frombits(row.Velocity[0]), math.Float32frombits(row.Velocity[1]))
				out := &o.Value
				switch row.Alias {
				case "separate":
				case "unit-position":
					out = &u.Value.PosVec
				case "unit-position-y":
					out = (*types.Pointf)(unsafe.Pointer(&u.Value.PosVec.Y))
				case "target-position":
					out = &v.Value.PosVec
				case "target-position-y":
					out = (*types.Pointf)(unsafe.Pointer(&v.Value.PosVec.Y))
				case "target-velocity":
					out = &v.Value.VelVec
				case "target-velocity-y":
					out = (*types.Pointf)(unsafe.Pointer(&v.Value.VelVec.Y))
				default:
					t.Fatal("unknown original alias")
				}
				pointers := []unsafe.Pointer{unsafe.Pointer(u), unsafe.Pointer(v), unsafe.Pointer(o)}
				sizes := []uintptr{unsafe.Sizeof(*u), unsafe.Sizeof(*v), unsafe.Sizeof(*o)}
				expected := make([][]byte, 3)
				written := false
				for j, p := range pointers {
					if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(p) <= math.MaxUint32 {
						t.Fatal("native alias fixture below4GiB")
					}
					expected[j] = copyBytes(p, sizes[j])
					start, address := uintptr(p), uintptr(unsafe.Pointer(out))
					if start <= address && address+8 <= start+sizes[j] {
						offset := address - start
						binary.LittleEndian.PutUint32(expected[j][offset:], row.Output[0])
						binary.LittleEndian.PutUint32(expected[j][offset+4:], row.Output[1])
						written = true
					}
				}
				if !written {
					t.Fatal("alias outside guarded allocation")
				}
				var returned uintptr
				if mode == "C" {
					returned = missileLeadPublicCBefore533080(u.Value.CObj(), v.Value.CObj(), math.Float32frombits(row.Speed), unsafe.Pointer(out))
				} else {
					returned = uintptr(unsafe.Pointer(server.MonsterProjectileLead533080(&u.Value, &v.Value, math.Float32frombits(row.Speed), out)))
				}
				if returned != uintptr(unsafe.Pointer(&v.Value)) || bits(*out) != row.Output {
					t.Errorf("original alias result differs: pointer=%#x output=%08x expected=%08x", returned, bits(*out), row.Output)
				}
				if bits(u.Value.PosVec) != row.UnitAfter || bits(v.Value.PosVec) != row.TargetAfter || bits(v.Value.VelVec) != row.VelocityAfter {
					t.Error("original sequential alias input mutation differs")
				}
				for j, p := range pointers {
					if !bytes.Equal(copyBytes(p, sizes[j]), expected[j]) {
						t.Errorf("alias changed guarded record%d beyond the original two stores", j)
					}
				}
			})
		}
	}
	t.Log("original alias contexts=56 Go-calls=56 actual-C-calls=56 no-server-no-class-gate=true")
}
