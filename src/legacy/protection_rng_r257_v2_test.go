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
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
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
	program, err := os.ReadFile("testdata/protection_rng_r257_test.c")
	if err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile("GAME5_2.c")
	if err != nil {
		t.Fatal(err)
	}
	start, end := "//----- (0056FE30)", "//----- (00578B00)"
	if strings.Count(string(source), start) != 1 || strings.Count(string(source), end) != 1 || bytes.Count(program, []byte("// RNG_BODY")) != 1 {
		t.Fatal("ambiguous production RNG body")
	}
	_, body, _ := strings.Cut(string(source), start)
	body, _, _ = strings.Cut(body, end)
	program = bytes.Replace(program, []byte("// RNG_BODY"), []byte("//"+body), 1)
	var input bytes.Buffer
	write := func(value any) {
		if err := binary.Write(&input, binary.LittleEndian, value); err != nil {
			t.Fatal(err)
		}
	}
	constants, err := hex.DecodeString(original.Constants)
	if err != nil || len(constants) != 40 {
		t.Fatal("different RNG constants")
	}
	input.Write(constants)
	write(uint32(len(original.Rows)))
	for _, row := range original.Rows {
		write(row.Seed)
		write(row.Initial)
		write(row.Bounds)
		write(uint32(len(row.Draws)))
		for _, draw := range row.Draws {
			write(draw.Value)
			write(draw.State)
		}
		write(uint32(len(row.Ranges)))
		for _, r := range row.Ranges {
			write(r.Low)
			write(r.High)
			write(r.Value)
			write(r.State)
			write(r.Bounds)
		}
	}
	compiler := strings.Fields(os.Getenv("CC"))
	if len(compiler) == 0 {
		compiler = []string{"cc"}
	}
	bin := filepath.Join(t.TempDir(), "protection-rng-native")
	args := append(compiler[1:], "-std=c11", "-O2", "-fno-strict-aliasing", "-Wall", "-Wextra", "-Werror", "-Wno-pointer-to-int-cast", "-Wno-sign-compare", "-Wno-unused-value", "-x", "c", "-", "-o", bin, "-lm")
	cmd := exec.Command(compiler[0], nativeFixtureCompileArgs(t, args)...)
	cmd.Stdin = bytes.NewReader(program)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("compile unchanged full RNG bodies: %v\n%s", err, out)
	}
	for i, name := range []string{"seed", "next", "range"} {
		t.Run(name, func(t *testing.T) {
			cmd := exec.Command(bin, fmt.Sprint(i))
			cmd.Stdin = bytes.NewReader(input.Bytes())
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("native protection RNG: %v\n%s", err, out)
			} else {
				t.Log(strings.TrimSpace(string(out)))
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
