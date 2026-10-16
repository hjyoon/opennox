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

	"github.com/opennox/opennox/v1/common/memmap"
)

type trigOriginal260 struct {
	SHA       string `json:"original_sha256"`
	Constants string `json:"constants_hex"`
	Table     string `json:"table_hex"`
	TableSHA  string `json:"table_sha256"`
	Rows      []struct {
		Pattern  byte `json:"pattern"`
		Ready    byte `json:"initialized"`
		Returned byte `json:"return_byte"`
		Final    byte `json:"final_flag"`
	} `json:"rows"`
}

func trigOracle260(t *testing.T) trigOriginal260 {
	t.Helper()
	data, err := os.ReadFile("testdata/trig_init_r260_original.json")
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(data)) != "3c1f48dc54111070ccaf6bcdd1b768941a1a7a993c3429326b86451d42b3d65a" {
		t.Fatal("changed original trig transcript")
	}
	var original trigOriginal260
	if err = json.Unmarshal(data, &original); err != nil {
		t.Fatal(err)
	}
	if original.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || len(original.Rows) != 15 || original.TableSHA != "362e011c229b40ea30aca24a974a98f1ddaee6a70c9df9562712f5c174ac805a" {
		t.Fatal("different original trig input")
	}
	return original
}
func TestTrigInit260CompleteNativeBody(t *testing.T) {
	original := trigOracle260(t)
	program, err := os.ReadFile("testdata/trig_init_r260_v3_test.c")
	if err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile("GAME1.c")
	if err != nil {
		t.Fatal(err)
	}
	start, end := "//----- (00414C90)", "//----- (00414D40)"
	if strings.Count(string(source), start) != 1 || strings.Count(string(source), end) != 1 || bytes.Count(program, []byte("// TRIG_BODY")) != 1 {
		t.Fatal("ambiguous trig body")
	}
	_, body, _ := strings.Cut(string(source), start)
	body, _, _ = strings.Cut(body, end)
	program = bytes.Replace(program, []byte("// TRIG_BODY"), []byte("//"+body), 1)
	constants, err := hex.DecodeString(original.Constants)
	if err != nil || len(constants) != 28 {
		t.Fatal("different original constants")
	}
	table, err := hex.DecodeString(original.Table)
	if err != nil || len(table) != 49152 {
		t.Fatal("different original table")
	}
	var input bytes.Buffer
	input.Write(constants)
	input.Write(table)
	if err = binary.Write(&input, binary.LittleEndian, uint32(len(original.Rows))); err != nil {
		t.Fatal(err)
	}
	for _, row := range original.Rows {
		input.Write([]byte{row.Pattern, row.Ready, row.Returned, row.Final})
	}
	compiler := strings.Fields(os.Getenv("CC"))
	if len(compiler) == 0 {
		compiler = []string{"cc"}
	}
	bin := filepath.Join(t.TempDir(), "trig-init-native")
	args := append(compiler[1:], "-std=c11", "-O2", "-fno-strict-aliasing", "-Wall", "-Wextra", "-Werror", "-Wno-pointer-to-int-cast", "-x", "c", "-", "-o", bin, "-lm")
	cmd := exec.Command(compiler[0], args...)
	cmd.Stdin = bytes.NewReader(program)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("compile complete trig initializer: %v\n%s", err, out)
	}
	for i, name := range []string{"bounds", "original-values"} {
		t.Run(name, func(t *testing.T) {
			cmd := exec.Command(bin, fmt.Sprint(i))
			cmd.Stdin = bytes.NewReader(input.Bytes())
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("native trig initializer: %v\n%s", err, out)
			} else {
				t.Log(strings.TrimSpace(string(out)))
			}
		})
	}
}
func TestTrigInit260ActualCGoOriginalTable(t *testing.T) {
	original := trigOracle260(t)
	table, err := hex.DecodeString(original.Table)
	if err != nil {
		t.Fatal(err)
	}
	sine := unsafe.Slice((*byte)(unsafe.Pointer(memmap.PtrUint32(0x85B3FC, 12260))), 16384)
	arc := unsafe.Slice((*byte)(unsafe.Pointer(memmap.PtrUint32(0x5D4594, 338472))), 32768)
	flag := memmap.PtrUint8(0x5D4594, 371240)
	// The real game loads this immutable blob before startup; package-legacy
	// tests do not run that loader. Restore the controlled input afterward.
	constantBytes, err := hex.DecodeString(original.Constants)
	if err != nil || len(constantBytes) != 28 {
		t.Fatal("different original trig constant input")
	}
	constants := memmap.Slice(0x581450, 7184)[:28]
	oldConstants := bytes.Clone(constants)
	copy(constants, constantBytes)
	t.Cleanup(func() { copy(constants, oldConstants) })
	oldSine, oldArc, oldFlag := bytes.Clone(sine), bytes.Clone(arc), *flag
	t.Cleanup(func() { copy(sine, oldSine); copy(arc, oldArc); *flag = oldFlag })
	for _, row := range original.Rows {
		t.Run(fmt.Sprintf("pattern_%02x_ready_%02x", row.Pattern, row.Ready), func(t *testing.T) {
			for i := range sine {
				sine[i] = row.Pattern
			}
			for i := range arc {
				arc[i] = row.Pattern
			}
			*flag = row.Ready
			beforeSine, beforeArc := bytes.Clone(sine), bytes.Clone(arc)
			Nox_xxx_initSinCosTables_414C90()
			if *flag != row.Final {
				t.Fatalf("flag=%x want=%x", *flag, row.Final)
			}
			if row.Ready != 0 {
				if !bytes.Equal(sine, beforeSine) || !bytes.Equal(arc, beforeArc) {
					t.Fatal("ready table modified")
				}
				return
			}
			sinBad, arcBad := 0, 0
			for i := 0; i < 4096; i++ {
				if binary.LittleEndian.Uint32(sine[4*i:]) != binary.LittleEndian.Uint32(table[4*i:]) {
					sinBad++
				}
			}
			for i := 0; i < 8192; i++ {
				if binary.LittleEndian.Uint32(arc[4*i:]) != binary.LittleEndian.Uint32(table[16384+4*i:]) {
					arcBad++
				}
			}
			if sinBad != 0 || arcBad != 0 {
				t.Fatalf("actual CGo differs from original: sine_words=%d acos_words=%d", sinBad, arcBad)
			}
		})
	}
}
