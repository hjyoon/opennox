package legacy

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"unsafe"
)

type statsRNGOriginal262 struct {
	Original string `json:"original_sha256"`
	Constant string `json:"constant_hex"`
	CRT      uint32 `json:"crt_control"`
	Game     uint32 `json:"game_control"`
	Rows     []struct {
		Pattern   uint32 `json:"pattern"`
		Ready     uint32 `json:"ready"`
		Seed      uint32 `json:"seed"`
		I         uint32 `json:"index_i"`
		J         uint32 `json:"index_j"`
		Draws     uint32 `json:"draws"`
		Initial   string `json:"initial_hex"`
		Final     string `json:"final_hex"`
		FinalSeed uint32 `json:"final_seed"`
		Values    string `json:"values_sha256"`
		States    string `json:"states_sha256"`
		Trace     string `json:"trace_sha256"`
	} `json:"rows"`
}

func TestStatsRNG262CompleteNativeBody(t *testing.T) {
	data, err := os.ReadFile("testdata/stats_rng_r262_original.json")
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(data)) != "eb1406a61a2f0fae6380ed62f0d71c8806c3652cebd5a7daca9634476cddad39" {
		t.Fatal("different unchanged original statistics RNG transcript")
	}
	var original statsRNGOriginal262
	if err = json.Unmarshal(data, &original); err != nil {
		t.Fatal(err)
	}
	if original.Original != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || original.CRT != 0x023f || original.Game != 0x0e3f || len(original.Rows) != 162 {
		t.Fatal("original input scope changed")
	}
	constant, err := hex.DecodeString(original.Constant)
	if err != nil || len(constant) != 8 {
		t.Fatal("invalid original coefficient")
	}
	blob, err := os.ReadFile("../common/memmap/nox/blobdata/blob_581450.dat")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(blob[8368:8376], constant) {
		t.Fatal("startup coefficient differs")
	}
	var input bytes.Buffer
	input.Write(constant)
	binary.Write(&input, binary.LittleEndian, uint32(len(original.Rows)))
	seeds := []uint32{0xffffffff, 0xfffffffe, uint32(0xf65b137a), 0x80000001, 0x80000000, 0, 1, 161803398, 0x7fffffff}
	rowIndex := 0
	for _, pattern := range []uint32{0, 0x5a, 0xff} {
		for _, ready := range []uint32{0, 1, 0x80000000} {
			for _, seed := range seeds {
				for _, indices := range [][2]uint32{{0, 31}, {55, 55}} {
					row := original.Rows[rowIndex]
					rowIndex++
					if row.Pattern != pattern || row.Ready != ready || row.Seed != seed || row.I != indices[0] || row.J != indices[1] || row.Draws != 128 {
						t.Fatal("original input matrix changed")
					}
					initial, err := hex.DecodeString(row.Initial)
					if err != nil || len(initial) != 308 {
						t.Fatal("original state size differs")
					}
					for _, value := range []uint32{row.Pattern, row.Seed, row.Draws} {
						if err = binary.Write(&input, binary.LittleEndian, value); err != nil {
							t.Fatal(err)
						}
					}
					input.Write(initial)
				}
			}
		}
	}
	fixture, err := os.ReadFile("testdata/stats_rng_r262_v1_test.c")
	if err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile("GAME1_2.c")
	if err != nil {
		t.Fatal(err)
	}
	start, end := "//----- (0042AAA0)", "//----- (0042ABF0)"
	if strings.Count(string(source), start) != 1 || strings.Count(string(source), end) != 1 || bytes.Count(fixture, []byte("// STATS_RNG_BODY")) != 1 {
		t.Fatal("ambiguous full statistics RNG body")
	}
	_, body, _ := strings.Cut(string(source), start)
	body, _, _ = strings.Cut(body, end)
	fixture = bytes.Replace(fixture, []byte("// STATS_RNG_BODY"), []byte("//"+body), 1)
	compiler := strings.Fields(os.Getenv("CC"))
	if len(compiler) == 0 {
		compiler = []string{"cc"}
	}
	bin := filepath.Join(t.TempDir(), "stats-rng-native")
	args := append(compiler[1:], "-std=c11", "-O2", "-fno-strict-aliasing", "-fno-strict-overflow", "-Wall", "-Wextra", "-Werror", "-Wno-pointer-to-int-cast", "-x", "c", "-", "-o", bin, "-lm")
	cmd := exec.Command(compiler[0], args...)
	cmd.Stdin = bytes.NewReader(fixture)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("compile complete statistics RNG: %v\n%s", err, out)
	}
	layouts := []string{"ordinary", "signed32-cross", "dword-wrap", "above-signed32", "above-dword", "heap"}
	if unsafe.Sizeof(uintptr(0)) == 4 {
		layouts = layouts[:1]
	}
	for layout, name := range layouts {
		t.Run(name, func(t *testing.T) {
			cmd := exec.Command(bin, fmt.Sprint(layout))
			cmd.Stdin = bytes.NewReader(input.Bytes())
			var diagnostic bytes.Buffer
			cmd.Stderr = &diagnostic
			output, err := cmd.Output()
			if err != nil {
				t.Fatalf("native RNG: %v\n%s", err, diagnostic.String())
			}
			t.Log(strings.TrimSpace(diagnostic.String()))
			if len(output) != len(original.Rows)*128*320 {
				t.Fatalf("trace bytes=%d", len(output))
			}
			for _, mode := range []string{"states", "values"} {
				t.Run(mode, func(t *testing.T) {
					failures := 0
					for index, row := range original.Rows {
						trace := output[index*128*320 : (index+1)*128*320]
						values, states := sha256.New(), sha256.New()
						for offset := 0; offset < len(trace); offset += 320 {
							values.Write(trace[offset : offset+8])
							states.Write(trace[offset+8 : offset+320])
						}
						okay := fmt.Sprintf("%x", values.Sum(nil)) == row.Values
						if mode == "states" {
							final, err := hex.DecodeString(row.Final)
							if err != nil {
								t.Fatal(err)
							}
							okay = fmt.Sprintf("%x", states.Sum(nil)) == row.States && binary.LittleEndian.Uint32(trace[len(trace)-312:]) == row.FinalSeed && bytes.Equal(trace[len(trace)-308:], final)
						}
						if !okay {
							if failures < 4 {
								t.Logf("row=%d pattern=%x ready=%x seed=%x indices=%d/%d", index, row.Pattern, row.Ready, row.Seed, row.I, row.J)
							}
							failures++
						}
					}
					if failures != 0 {
						t.Fatalf("mode=%s rows=%d differing_rows=%d draws_per_row=128", mode, len(original.Rows), failures)
					}
				})
			}
		})
	}
}
