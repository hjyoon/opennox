package legacy

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"os"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

type protectionRNGOracle257 struct {
	SHA         string `json:"original_sha256"`
	Constants   string `json:"constants_hex"`
	CRTControl  uint32 `json:"crt_control"`
	GameControl uint32 `json:"gameplay_control"`
	Rows        []struct {
		Seed    uint32    `json:"seed"`
		Initial [5]uint64 `json:"initial"`
		Bounds  [3]uint32 `json:"initial_bounds"`
		Draws   []struct {
			Value uint64    `json:"value"`
			State [5]uint64 `json:"state"`
		} `json:"draws"`
		Ranges []struct {
			Low, High, Value uint32
			State            [5]uint64
			Bounds           [3]uint32
		} `json:"ranges"`
	} `json:"rows"`
}

func protectionRNGOriginal257(t *testing.T) protectionRNGOracle257 {
	t.Helper()
	raw, err := os.ReadFile("testdata/protection_rng_r257_original.json")
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(raw)) != "3e8d293f60e34183665edeef6d24caa59ded3b188a25d29d5a440c6bca48cd1e" {
		t.Fatal("changed original transcript")
	}
	var original protectionRNGOracle257
	if err := json.Unmarshal(raw, &original); err != nil {
		t.Fatal(err)
	}
	if original.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || len(original.Rows) != 9 || original.CRTControl != 0x023f || original.GameControl != 0x0e3f {
		t.Fatal("different original protection RNG input")
	}
	// Unicorn omits reserved x87 control-word bit 6; both live setup routines
	// were executed, rather than assigning an assumed floating-point mode.
	for _, row := range original.Rows {
		if len(row.Draws) != 128 || len(row.Ranges) != 40 {
			t.Fatal("truncated original stream")
		}
	}
	return original
}

func TestProtectionRNG257CompleteNativeBodies(t *testing.T) {
	original := protectionRNGOriginal257(t)
	raw, err := hex.DecodeString(original.Constants)
	if err != nil || len(raw) != 40 {
		t.Fatal("different RNG constants")
	}
	const stateOffset = 2516388
	blob := memmap.BlobByAddr(0x5D4594)
	ordinary := blob.Data
	window := ordinary[stateOffset-24 : stateOffset+48]
	oldWindow := [72]byte(window)
	oldWidth, oldMinimum := *playerProtectionRandomWidth, *playerProtectionRandomMinimum
	constants := memmap.Slice(0x581450, 11344)[:40]
	oldConstants := [40]byte(constants)
	copy(constants, raw)
	t.Cleanup(func() {
		blob.Data = ordinary
		copy(window, oldWindow[:])
		copy(constants, oldConstants[:])
		*playerProtectionRandomWidth, *playerProtectionRandomMinimum = oldWidth, oldMinimum
	})

	for layout := range 2 {
		t.Run([]string{"ordinary", "native-reallocated"}[layout], func(t *testing.T) {
			active := ordinary
			if layout == 1 {
				high, release := alloc.MakeHigh[byte](nil, len(ordinary))
				t.Cleanup(release)
				active = high
				if unsafe.Sizeof(uintptr(0)) == 8 && uint64(uintptr(unsafe.Pointer(&active[stateOffset]))) <= math.MaxUint32 {
					t.Fatal("RNG fixture did not use a native address above 4 GiB")
				}
			}
			blob.Data = active
			t.Cleanup(func() { blob.Data = ordinary })
			state := active[stateOffset : stateOffset+40]
			before := active[stateOffset-24 : stateOffset-16]
			after := active[stateOffset+40 : stateOffset+48]
			guard := [8]byte{0x12, 0x49, 0x38, 0xbf, 0xc1, 0xfe, 0xe2, 0x65}
			checkState := func(t *testing.T, expected [5]uint64) {
				t.Helper()
				for i, word := range expected {
					if actual := binary.LittleEndian.Uint64(state[8*i:]); actual != word {
						t.Fatalf("state word %d = %016x, original %016x", i, actual, word)
					}
				}
			}
			bounds := func() [3]uint32 {
				return [3]uint32{*playerProtectionRandomWidth, memmap.Uint32(0x5D4594, 2516376), *playerProtectionRandomMinimum}
			}
			for _, row := range original.Rows {
				t.Run(fmt.Sprintf("seed-%08x", row.Seed), func(t *testing.T) {
					copy(before, guard[:])
					copy(after, guard[:])
					sub_56FF00(int32(row.Seed))
					checkState(t, row.Initial)
					if bounds() != row.Bounds {
						t.Fatalf("seed bounds %08x, original %08x", bounds(), row.Bounds)
					}
					for step, draw := range row.Draws {
						if got := math.Float64bits(nox_xxx_unkDoubleSmth_56FE30()); got != draw.Value {
							t.Fatalf("draw %d = %016x, original %016x", step, got, draw.Value)
						}
						checkState(t, draw.State)
					}
					for step, draw := range row.Ranges {
						if got := uint32(sub_56FF80(int32(draw.Low), int32(draw.High))); got != draw.Value {
							t.Fatalf("range %d = %08x, original %08x", step, got, draw.Value)
						}
						checkState(t, draw.State)
						if bounds() != draw.Bounds {
							t.Fatalf("range %d bounds %08x, original %08x", step, bounds(), draw.Bounds)
						}
					}
					if !bytes.Equal(before, guard[:]) || !bytes.Equal(after, guard[:]) {
						t.Fatal("protection RNG overwrote bytes outside the state")
					}
				})
			}
		})
	}
}

// Independent arithmetic model uses big.Float, never the C residual helpers.
// Seeded states are finite binary fractions. The original repeats state[3]
// for both final coefficients; it does NOT read the oldest pre-shift state.
func protectionRNGReferenceNext257(state [5]float64, constants [5]float64) [5]float64 {
	op := func(a, b float64, multiply bool) float64 {
		x := new(big.Float).SetPrec(53).SetMode(big.ToZero).SetFloat64(a)
		y := new(big.Float).SetPrec(53).SetMode(big.ToZero).SetFloat64(b)
		if multiply {
			x.Mul(x, y)
		} else {
			x.Add(x, y)
		}
		v, _ := x.Float64()
		return v
	}
	state[3], state[2], state[1] = state[2], state[1], state[0]
	v := op(state[0], constants[4], true)
	v = op(v, op(state[2], constants[3], true), false)
	v = op(v, op(state[3], constants[2], true), false)
	v = op(v, op(state[3], constants[1], true), false)
	v = op(v, state[4], false)
	whole := math.Floor(v)
	state[0], state[4] = op(v, -whole, false), op(whole, constants[0], true)
	return state
}

func TestProtectionRNG257IndependentOriginalArithmetic(t *testing.T) {
	original := protectionRNGOriginal257(t)
	raw, err := hex.DecodeString(original.Constants)
	if err != nil {
		t.Fatal(err)
	}
	var constants [5]float64
	for i := range constants {
		constants[i] = math.Float64frombits(binary.LittleEndian.Uint64(raw[8*i:]))
	}
	for _, row := range original.Rows {
		t.Run(fmt.Sprintf("seed-%08x", row.Seed), func(t *testing.T) {
			seed := row.Seed
			if seed == 0 {
				seed = 0xffffffff
			}
			var state [5]float64
			for i := range state {
				seed ^= seed << 13
				seed ^= seed >> 17
				seed ^= seed << 5
				state[i] = float64(seed) * constants[0]
			}
			for i := 0; i < 19; i++ {
				state = protectionRNGReferenceNext257(state, constants)
			}
			check := func(want [5]uint64) {
				t.Helper()
				for i, v := range state {
					if math.Float64bits(v) != want[i] {
						t.Fatalf("state[%d]=%016x GAME.EXE=%016x", i, math.Float64bits(v), want[i])
					}
				}
			}
			check(row.Initial)
			for _, draw := range row.Draws {
				state = protectionRNGReferenceNext257(state, constants)
				check(draw.State)
				if math.Float64bits(state[0]) != draw.Value {
					t.Fatal("wrong original floating return")
				}
			}
			for _, r := range row.Ranges {
				state = protectionRNGReferenceNext257(state, constants)
				check(r.State)
				width := r.High - r.Low + 1
				x := new(big.Float).SetPrec(53).SetMode(big.ToZero).SetFloat64(state[0])
				x.Mul(x, new(big.Float).SetUint64(uint64(width)))
				integer, _ := x.Int64()
				value := uint32(integer)
				if value >= width {
					value = width
				}
				if r.Value != r.Low+value || r.Bounds != [3]uint32{width, r.High, r.Low} {
					t.Fatal("wrong original unsigned range result/state")
				}
			}
		})
	}
}
