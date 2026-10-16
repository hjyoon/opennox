package legacy

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/opennox/opennox/v1/common/memmap"
)

type networkTableOracle256 struct {
	SHA       string `json:"original_sha256"`
	Functions []struct {
		Name  string   `json:"name"`
		Bytes int      `json:"bytes"`
		SHA   string   `json:"sha256"`
		Stock []uint32 `json:"stock"`
	} `json:"functions"`
	Rows []struct {
		Name    string      `json:"name"`
		Variant string      `json:"variant"`
		Table   []uint32    `json:"table"`
		Cases   [][2]uint32 `json:"cases"`
	} `json:"rows"`
}

func networkTableOriginal256(t *testing.T) networkTableOracle256 {
	t.Helper()
	raw, err := os.ReadFile("testdata/network_tables_r256_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var original networkTableOracle256
	if err := json.Unmarshal(raw, &original); err != nil {
		t.Fatal(err)
	}
	if original.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || len(original.Functions) != 3 || len(original.Rows) != 9 {
		t.Fatal("wrong original network table oracle")
	}
	wantSHA := []string{"5a3c58d018faf2c8889bd558f93fe852724eb63760b96cbb12fe53a78c06466e", "745d6c85e1ef1df9323b8cf38aad3857daeaae7ea46f4fcfde3ddf6dbf59ab90", "1406063778cfc47a3e7d22bffb3f404085cf1701727c9f42fa1e27770b2c9a90"}
	for i, function := range original.Functions {
		if function.SHA != wantSHA[i] || function.Bytes != []int{36, 40, 30}[i] || function.Name != []string{"settings", "rate", "mode"}[i] {
			t.Fatal("different original leaf function")
		}
	}
	return original
}

func installNetworkTableStock256(t *testing.T, original networkTableOracle256) {
	t.Helper()
	// The standalone legacy package does not run the application's blob-data
	// initializer. Supply only the original leaf-table bytes and restore them.
	for i, function := range original.Functions {
		offset := []uintptr{4704, 4664, 191836}[i]
		data := memmap.Slice(0x587000, offset)[:4*len(function.Stock)]
		before := bytes.Clone(data)
		t.Cleanup(func() { copy(data, before) })
		for j, value := range function.Stock {
			binary.LittleEndian.PutUint32(data[4*j:], value)
		}
	}
}

func TestNetworkTables256V2CompleteNativeBodies(t *testing.T) {
	original := networkTableOriginal256(t)
	program, err := os.ReadFile("testdata/network_tables_r256_test.c")
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range []struct{ file, first, last, marker string }{
		{"GAME1.c", "//----- (00409A70)", "//----- (00409B30)", "// SETTINGS_BODY"},
		{"GAME1.c", "//----- (0040A710)", "//----- (0040A740)", "// RATE_BODY"},
		{"GAME3_2.c", "//----- (004D0D50)", "//----- (004D0D70)", "// MODE_BODY"},
	} {
		source, err := os.ReadFile(entry.file)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Count(string(source), entry.first) != 1 || strings.Count(string(source), entry.last) != 1 || bytes.Count(program, []byte(entry.marker)) != 1 {
			t.Fatal("ambiguous full network table body")
		}
		_, body, _ := strings.Cut(string(source), entry.first)
		body, _, _ = strings.Cut(body, entry.last)
		program = bytes.Replace(program, []byte(entry.marker), []byte("//"+body), 1)
	}
	var input bytes.Buffer
	word := func(value uint32) {
		if err := binary.Write(&input, binary.LittleEndian, value); err != nil {
			t.Fatal(err)
		}
	}
	word(uint32(len(original.Rows)))
	for _, row := range original.Rows {
		kind := map[string]uint32{"settings": 0, "rate": 1, "mode": 2}[row.Name]
		word(kind)
		word(uint32(len(row.Table)))
		for _, value := range row.Table {
			word(value)
		}
		word(uint32(len(row.Cases)))
		for _, pair := range row.Cases {
			word(pair[0])
			word(pair[1])
		}
	}
	compiler := strings.Fields(os.Getenv("CC"))
	if len(compiler) == 0 {
		compiler = []string{"cc"}
	}
	bin := filepath.Join(t.TempDir(), "network-tables-native")
	args := append(compiler[1:], "-std=c11", "-O2", "-fno-strict-aliasing", "-Wall", "-Wextra", "-Werror", "-Wno-pointer-to-int-cast", "-Wno-sign-compare", "-x", "c", "-", "-o", bin)
	cmd := exec.Command(compiler[0], args...)
	cmd.Stdin = bytes.NewReader(program)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("compile complete C lookup bodies: %v\n%s", err, output)
	}
	for _, name := range []string{"settings", "rate", "mode"} {
		t.Run(name, func(t *testing.T) {
			cmd := exec.Command(bin, name)
			cmd.Stdin = bytes.NewReader(input.Bytes())
			if output, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("native lookup: %v\n%s", err, output)
			} else {
				t.Log(strings.TrimSpace(string(output)))
			}
		})
	}
}

func TestNetworkTables256V2LiveCgoSettingsAndModes(t *testing.T) {
	original := networkTableOriginal256(t)
	installNetworkTableStock256(t, original)
	for _, row := range original.Rows {
		if row.Variant != "stock" {
			continue
		}
		switch row.Name {
		case "settings":
			var expected [0x1800]uint32
			for _, pair := range row.Cases {
				expected[pair[0]] = pair[1]
			}
			for value := 0; value <= 0xffff; value++ {
				if got, want := uint32(Sub_409A70(value)), expected[value&0x17f0]; got != want {
					t.Fatalf("settings %#x = %d, GAME.EXE = %d", value, got, want)
				}
			}
		case "mode":
			live := memmap.Slice(0x5D4594, 1548452)[:24]
			before := bytes.Clone(live)
			t.Cleanup(func() { copy(live, before) })
			for index := 0; index < 6; index++ {
				binary.LittleEndian.PutUint32(live[4*index:], 0x12345670+uint32(index))
			}
			for _, pair := range row.Cases {
				if got, want := uint32(Sub_4D0DE0(pair[0])), 0x12345670+pair[1]; got != want {
					t.Fatalf("mode %#x value = %#x, original slot = %d", pair[0], got, pair[1])
				}
			}
		}
	}
}
