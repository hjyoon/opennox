package opennox

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"math"
	"os"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

// Exercise the actual common production entry, including its real movable
// gate and C wakeup path. Empty callback fields make wakeup leave no queue.
// Expectations execute the complete unchanged original arithmetic and getters.
func TestObjectApplyForce52DF80Original(t *testing.T) {
	objectApplyForceOriginalFixture52DF80(t, "testdata/object_force_startup_original.json", 97)
}

func TestObjectApplyForce52DF80Compatibility(t *testing.T) {
	objectApplyForceOriginalFixture52DF80(t, "testdata/object_force_compatibility_original.json", 307)
}

// Enter through the actual registered C RecoilEffect, real CGo force export
// and common production path. No force implementation is substituted.
func TestObjectApplyForce52DF80ActualRecoil(t *testing.T) {
	old := legacy.GetServer
	legacy.GetServer = func() legacy.Server { return new(Server) }
	t.Cleanup(func() { legacy.GetServer = old })
	for _, class := range []object.Class{object.ClassMissile, object.ClassPlayer} {
		for _, row := range []struct {
			name             string
			position, origin types.Pointf
			want             [2]uint32
		}{
			{"three-four-five", types.Ptf(3, 4), types.Pointf{}, [2]uint32{0x4096c6c6, 0x40b83c3c}},
			{"map-horizontal", types.Ptf(212, 200), types.Ptf(100, 200), [2]uint32{0x40f962ca, 0xbe000000}},
			{"map-diagonal", types.Ptf(4363, 2036), types.Ptf(4016, 1984), [2]uint32{0x40f6e1d2, 0x3f7c770b}},
		} {
			t.Run(row.name+"/"+class.String(), func(t *testing.T) {
				unit, freeUnit := alloc.New(server.Object{})
				weapon, freeWeapon := alloc.New(server.Object{})
				effect, freeEffect := alloc.New(server.ModifierEff{})
				t.Cleanup(freeUnit)
				t.Cleanup(freeWeapon)
				t.Cleanup(freeEffect)
				*unit = server.Object{ObjClass: class, PosVec: row.position, Mass: 20, ForceVec: types.Ptf(.3, -.125)}
				*weapon = server.Object{PosVec: row.origin}
				*effect = server.ModifierEff{AttackPreHit52: server.ModifierEffFnc{Valf: 15}}
				for _, p := range []unsafe.Pointer{unsafe.Pointer(unit), unsafe.Pointer(weapon), unsafe.Pointer(effect)} {
					if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(p) <= math.MaxUint32 {
						t.Fatal("actual C force input below 4 GiB")
					}
				}
				beforeWeapon, beforeEffect := *weapon, *effect
				ccall.CallVoidPtr5(legacy.RecoilEffectPointer4E0640(), effect.C(), weapon.CObj(), nil, unit.CObj(), nil)
				got := [2]uint32{math.Float32bits(unit.ForceVec.X), math.Float32bits(unit.ForceVec.Y)}
				if got != row.want || *weapon != beforeWeapon || *effect != beforeEffect {
					t.Fatalf("actual C Recoil force=%08x,%08x; original=%08x,%08x", got[0], got[1], row.want[0], row.want[1])
				}
			})
		}
	}
}

func objectApplyForceOriginalFixture52DF80(t *testing.T, path string, contexts int) {
	t.Helper()
	blob, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		SHA               string            `json:"original_sha256"`
		Unchanged         bool              `json:"original_unchanged"`
		CW                uint16            `json:"actual_startup_control_word"`
		Contexts          int               `json:"contexts"`
		Faults            []json.RawMessage `json:"memory_faults"`
		Reference         string            `json:"expected_startup_result_source"`
		OracleDifferences []string          `json:"signaling_NaN_oracle_disagreements"`
		Excluded          int               `json:"excluded_or_weakened_conditions"`
		Rows              []struct {
			Mode  string `json:"mode"`
			Input struct {
				Name     string    `json:"name"`
				Class    uint32    `json:"target_class"`
				Flags    uint32    `json:"target_flags"`
				Position [2]uint32 `json:"position_bits32"`
				Origin   [2]uint32 `json:"origin_bits32"`
				Mass     uint32    `json:"mass_bits32"`
				Force    uint32    `json:"force_bits32"`
				Initial  [2]uint32 `json:"initial_force_bits32"`
			} `json:"input"`
			Final     [2]uint32 `json:"final_force_bits32"`
			Preserved bool      `json:"nonoutput_bytes_preserved"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(blob, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || !fixture.Unchanged || fixture.CW != 0xe3f || fixture.Contexts != contexts || len(fixture.Rows) != 2*contexts || len(fixture.Faults) != 0 {
		t.Fatal("incomplete unchanged original force observations")
	}
	if contexts == 307 && (fixture.Reference == "" || len(fixture.OracleDifferences) != 10 || fixture.Excluded != 0) {
		t.Fatal("supplemental signaling NaN expectations require independent x87 evidence")
	}
	point := func(bits [2]uint32) types.Pointf {
		return types.Ptf(math.Float32frombits(bits[0]), math.Float32frombits(bits[1]))
	}
	count := 0
	for _, row := range fixture.Rows {
		if row.Mode != "actual-startup" {
			continue
		}
		count++
		t.Run(row.Input.Name, func(t *testing.T) {
			i := row.Input
			unit, free := alloc.New(server.Object{})
			t.Cleanup(free)
			*unit = server.Object{ObjClass: object.Class(i.Class), ObjFlags: object.Flags(i.Flags), PosVec: point(i.Position), Mass: math.Float32frombits(i.Mass), ForceVec: point(i.Initial), Field29: 0xa55a1234}
			if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
				t.Fatal("native force object below 4 GiB")
			}
			raw := func() []byte { return unsafe.Slice((*byte)(unsafe.Pointer(unit)), int(unsafe.Sizeof(*unit))) }
			want := bytes.Clone(raw())
			offset := int(unsafe.Offsetof(unit.ForceVec))
			binary.LittleEndian.PutUint32(want[offset:], row.Final[0])
			binary.LittleEndian.PutUint32(want[offset+4:], row.Final[1])
			asObjectS(unit).applyForce(point(i.Origin), float64(math.Float32frombits(i.Force)))
			got := [2]uint32{math.Float32bits(unit.ForceVec.X), math.Float32bits(unit.ForceVec.Y)}
			if !row.Preserved || got != row.Final || !bytes.Equal(raw(), want) {
				t.Fatalf("actual force=%08x,%08x; original=%08x,%08x; nonoutput bytes unchanged=%t", got[0], got[1], row.Final[0], row.Final[1], bytes.Equal(raw(), want))
			}
		})
	}
	if count != contexts {
		t.Fatalf("selected startup contexts=%d", count)
	}
}
