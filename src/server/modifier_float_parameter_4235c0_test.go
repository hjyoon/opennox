package server

import (
	"bytes"
	"encoding/json"
	"math"
	"os"
	"strconv"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

type modifierFloatParameterFrame4235C0 struct {
	Before [4]uint64
	Effect ModifierEff
	After  [4]uint64
}

func TestModifierFloatParameterOriginal4235C0(t *testing.T) {
	if strconv.IntSize != 64 {
		t.Skip("native 64-bit modifier float parameter context")
	}
	raw, err := os.ReadFile("testdata/modifier_float_parameter_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		OriginalSHA256    string `json:"original_sha256"`
		OriginalUnchanged bool   `json:"original_unchanged"`
		CompletedContexts int    `json:"completed_contexts"`
		RawRows           int    `json:"raw_rows"`
		Failure           any    `json:"failure"`
		Rows              []struct {
			Name, Mode   string
			Inputs       struct{ Attribute, Argument string }
			OriginalBits uint32 `json:"original_bits"`
		}
	}
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.OriginalSHA256 != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || !fixture.OriginalUnchanged || fixture.Failure != nil || fixture.CompletedContexts != 63 || fixture.RawRows != 252 || len(fixture.Rows) != 252 {
		t.Fatal("incomplete original parser/CRT observations")
	}
	count := 0
	for _, row := range fixture.Rows {
		if row.Mode != "actual-startup-inherited" {
			continue
		}
		count++
		t.Run(row.Name, func(t *testing.T) {
			frame, free := alloc.New(modifierFloatParameterFrame4235C0{})
			defer free()
			for i := range frame.Before {
				frame.Before[i] = 0x123456789abcdef0
				frame.After[i] = 0xfedcba9876543210
			}
			effect := &frame.Effect
			fields := map[string]*float32{"ATTACKEFFECT": &effect.Attack40.Valf, "ATTACKPREHITEFFECT": &effect.AttackPreHit52.Valf, "ATTACKPREDAMAGEEFFECT": &effect.AttackPreDmg64.Valf, "DEFENDEFFECT": &effect.Defend76.Valf, "DEFENDCOLLIDEEFFECT": &effect.DefendCollide88.Valf, "UPDATEEFFECT": &effect.Update100.Valf, "ENGAGEEFFECT": &effect.EngageFloat120, "DISENGAGEEFFECT": &effect.DisengageFloat128}
			for _, field := range fields {
				*field = math.Float32frombits(0x4f123456)
			}
			target := fields[row.Inputs.Attribute]
			if target == nil {
				t.Fatal("unknown original target attribute")
			}
			if uintptr(unsafe.Pointer(frame)) <= math.MaxUint32 || uintptr(unsafe.Pointer(effect)) <= math.MaxUint32 || uintptr(unsafe.Pointer(target)) <= math.MaxUint32 {
				t.Fatal("native modifier pointers not above UINT32_MAX")
			}
			view := unsafe.Slice((*byte)(unsafe.Pointer(frame)), int(unsafe.Sizeof(*frame)))
			before := append([]byte(nil), view...)
			if !ModEffectParseFloat(effect, ModParseTarg{Float: target}, row.Inputs.Argument) {
				t.Fatal("actual Go parser changed original true return")
			}
			got := math.Float32bits(*target)
			*target = math.Float32frombits(0x4f123456)
			if !bytes.Equal(before, view) {
				t.Fatal("native frame/modifier nonoutput bytes changed")
			}
			if got != row.OriginalBits {
				t.Fatalf("actual modifier float parser=%08x original=%08x argument=%q", got, row.OriginalBits, row.Inputs.Argument)
			}
		})
	}
	if count != 63 {
		t.Fatalf("original startup contexts=%d", count)
	}
}
