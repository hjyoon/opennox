package legacy

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

type worldRegionOracle267 struct {
	OriginalSHA string  `json:"original_sha256"`
	Entry       string  `json:"entry"`
	End         string  `json:"end"`
	Bytes       int     `json:"bytes"`
	CodeSHA     string  `json:"code_sha256"`
	Stock       []int16 `json:"stock"`
	Cases       int     `json:"cases"`
	Rows        []struct {
		Name  string     `json:"name"`
		Table []int16    `json:"table"`
		Cases [][3]int32 `json:"cases"`
	} `json:"rows"`
}

func worldRegionOriginal267(t *testing.T) worldRegionOracle267 {
	t.Helper()
	raw, err := os.ReadFile("testdata/world_region_r267_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var oracle worldRegionOracle267
	if err := json.Unmarshal(raw, &oracle); err != nil {
		t.Fatal(err)
	}
	if oracle.OriginalSHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || oracle.Entry != "0x437860" || oracle.End != "0x4378a1" || oracle.Bytes != 65 || oracle.CodeSHA != "7139c7351cec00f518a035fcad680538ae8c2d625fbfd986cc956dd54a0314b4" || len(oracle.Stock) != 16 || len(oracle.Rows) != 4 || oracle.Cases != 1504 {
		t.Fatal("different original world-region function/input")
	}
	count := 0
	for i, row := range oracle.Rows {
		if len(row.Table) != 16 || row.Name != []string{"stock", "overlapping-first-match", "signed-int16", "reversed-and-degenerate"}[i] || len(row.Cases) != []int{225, 693, 361, 225}[i] {
			t.Fatal("different declared original rectangle cases")
		}
		for _, test := range row.Cases {
			if test[2] < 0 || test[2] > 3 {
				t.Fatal("invalid original region result")
			}
		}
		count += len(row.Cases)
	}
	if count != oracle.Cases {
		t.Fatal("missing original region cases")
	}
	return oracle
}

func TestWorldRegion437860NativeRelocationR267(t *testing.T) {
	original := worldRegionOriginal267(t)
	program, err := os.ReadFile("testdata/world_region_r267_test.c")
	if err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile("GAME1_2.c")
	if err != nil {
		t.Fatal(err)
	}
	const first, last, marker = "//----- (00437860)", "//----- (004379C0)", "// PRODUCTION_BODY"
	if strings.Count(string(source), first) != 1 || strings.Count(string(source), last) != 1 || bytes.Count(program, []byte(marker)) != 1 {
		t.Fatal("ambiguous complete world-region body")
	}
	_, body, _ := strings.Cut(string(source), first)
	body, _, _ = strings.Cut(body, last)
	program = bytes.Replace(program, []byte(marker), []byte("//"+body), 1)
	var input bytes.Buffer
	word := func(value uint32) {
		if err := binary.Write(&input, binary.LittleEndian, value); err != nil {
			t.Fatal(err)
		}
	}
	word(uint32(len(original.Rows)))
	for _, row := range original.Rows {
		for _, value := range row.Table {
			word(uint32(int32(value)))
		}
		word(uint32(len(row.Cases)))
		for _, test := range row.Cases {
			for _, value := range test {
				word(uint32(value))
			}
		}
	}
	compiler := strings.Fields(os.Getenv("CC"))
	if len(compiler) == 0 {
		compiler = []string{"cc"}
	}
	bin := filepath.Join(t.TempDir(), "world-region-native")
	args := append(compiler[1:], "-std=c11", "-O2", "-fno-strict-aliasing", "-Wall", "-Wextra", "-Werror", "-Wno-pointer-to-int-cast", "-x", "c", "-", "-o", bin)
	cmd := exec.Command(compiler[0], args...)
	cmd.Stdin = bytes.NewReader(program)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("compile complete world-region C body: %v\n%s", err, output)
	}
	layouts := []string{"ordinary"}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		layouts = append(layouts, "signed-minus32", "signed-minus24", "signed-minus16", "signed-minus8", "wrap-minus16", "signed-plus16")
	}
	for i, name := range layouts {
		t.Run(name, func(t *testing.T) {
			cmd := exec.Command(bin, strconv.Itoa(i))
			cmd.Stdin = bytes.NewReader(input.Bytes())
			if output, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("native region layout=%s: %v\n%s", name, err, output)
			} else {
				t.Log(strings.TrimSpace(string(output)))
			}
		})
	}
}

func TestWorldRegion437860LiveCgoR267(t *testing.T) {
	original := worldRegionOriginal267(t)
	// The standalone legacy package has not run the application's blob-data
	// initialization. Install only these declared original rectangle inputs,
	// restore every byte, and call the real CGo entry rather than a Go model.
	data := memmap.Slice(0x587000, 87528)[:32]
	before := bytes.Clone(data)
	t.Cleanup(func() { copy(data, before) })
	for _, row := range original.Rows {
		t.Run(row.Name, func(t *testing.T) {
			for i, value := range row.Table {
				binary.LittleEndian.PutUint16(data[2*i:], uint16(value))
			}
			declared := bytes.Clone(data)
			for _, test := range row.Cases {
				if got := Sub_437860(int(test[0]), int(test[1])); got != int(test[2]) {
					t.Fatalf("region(%d,%d)=%d, GAME.EXE=%d", test[0], test[1], got, test[2])
				}
			}
			if !bytes.Equal(data, declared) {
				t.Fatal("region lookup changed input rectangles")
			}
			t.Log(fmt.Sprintf("actual CGo original cases=%d native_pointer_bytes=%d", len(row.Cases), unsafe.Sizeof(uintptr(0))))
		})
	}
}
