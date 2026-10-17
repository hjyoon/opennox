package legacy

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
	"unsafe"

	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

// Compile the complete production functions against controlled services. These
// cases verify signed row translation, fixed wire widths and native list links;
// the companion CGo test below uses the real cryptfile service.
func TestFloorMapNativeRegionDataR300(t *testing.T) {
	if unsafe.Sizeof(uintptr(0)) != 8 {
		t.Skip("64-bit native pointer regression")
	}
	fixture, err := os.ReadFile("testdata/floor_map_region_r300_test.c")
	if err != nil {
		t.Fatal(err)
	}
	var bodies strings.Builder
	for _, part := range []struct{ file, symbol string }{
		{"GAME1_1.c", "nox_server_mapRWFloorMap_422230"},
		{"GAME1_1.c", "nox_xxx_tileReadOne_422A40"},
		{"GAME1_1.c", "nox_xxx_tile_422C10"},
		{"GAME1_2.c", "sub_428170"},
	} {
		data, err := os.ReadFile(part.file)
		if err != nil {
			t.Fatal(err)
		}
		starts := regexp.MustCompile(`(?m)^(?:int|unsigned char) `+part.symbol+`\([^;\n]*\) \{`).FindAllIndex(data, -1)
		if len(starts) != 1 {
			t.Fatalf("nonunique production function %s", part.symbol)
		}
		body, _, ok := strings.Cut(string(data[starts[0][0]:]), "\n}\n")
		if !ok {
			t.Fatalf("missing production end for %s", part.symbol)
		}
		bodies.WriteString(body + "\n}\n")
	}
	program := strings.Replace(string(fixture), "// PRODUCTION_FLOOR_REGION_BODIES", bodies.String(), 1)
	compiler := strings.Fields(os.Getenv("CC"))
	if len(compiler) == 0 {
		compiler = []string{"cc"}
	}
	bin := filepath.Join(t.TempDir(), "floor-map-region")
	args := append(compiler[1:], "-std=c11", "-O2", "-Wno-int-conversion", "-Wno-incompatible-pointer-types",
		"-Wno-pointer-to-int-cast", "-Wno-int-to-pointer-cast", "-I.", "-x", "c", "-", "-o", bin)
	cmd := exec.Command(compiler[0], args...)
	cmd.Stdin = strings.NewReader(program)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("compile actual FloorMap functions: %v\n%s", err, out)
	}
	for _, name := range []string{"v4-negative-row-delta-read", "v3-region-two-column-read", "v4-region-write-native-links"} {
		t.Run(name, func(t *testing.T) {
			if out, err := exec.Command(bin, name).CombinedOutput(); err != nil {
				t.Fatalf("native FloorMap data contract: %v\n%s", err, out)
			} else {
				t.Log(strings.TrimSpace(string(out)))
			}
		})
	}
}

// Exercise the actual Go wrapper and C FloorMap entry/delegate with a region
// above 4GiB. Isolate C crashes so a truncated pointer fails this regression
// without terminating the rest of the legacy test suite.
func TestFloorMapReadNativeRegionR300(t *testing.T) {
	if unsafe.Sizeof(uintptr(0)) != 8 {
		t.Skip("64-bit native pointer regression")
	}
	const selector = "OPENNOX_FLOOR_ABI_R300_CHILD"
	if child := os.Getenv(selector); child != "" {
		var version uint16
		switch child {
		case "3":
			version = 3
		case "4":
			version = 4
		default:
			t.Fatalf("unexpected child selector %q", child)
		}
		floorReadNativeRegionR300(t, version)
		return
	}
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, version := range []uint16{3, 4} {
		t.Run(fmt.Sprint(version), func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, self, "-test.v", "-test.run=^TestFloorMapReadNativeRegionR300$", "-test.count=1", "-test.timeout=20s")
			for _, entry := range os.Environ() {
				if !strings.HasPrefix(entry, selector+"=") {
					cmd.Env = append(cmd.Env, entry)
				}
			}
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%d", selector, version))
			cmd.WaitDelay = 2 * time.Second
			output, err := cmd.CombinedOutput()
			if err != nil || ctx.Err() != nil {
				t.Fatalf("actual FloorMap CGo child: err=%v context=%v\n%s", err, ctx.Err(), output)
			}
			marker := fmt.Sprintf("R300_FLOOR_CGO_OK version=%d bytes=%d", version, 18+2*int(version-3))
			if !bytes.Contains(output, []byte(marker)) {
				t.Fatalf("child exited without completing the actual callback: %s", output)
			}
		})
	}
}

func floorReadNativeRegionR300(t *testing.T, version uint16) {
	t.Helper()
	wire := binary.LittleEndian.AppendUint16(nil, version)
	for i := 0; i < 4; i++ {
		wire = binary.LittleEndian.AppendUint32(wire, 0)
	}
	if version == 4 {
		wire = binary.LittleEndian.AppendUint16(wire, math.MaxUint16)
	}
	consumed := len(wire)
	tail := [4]byte{0xa7, byte(version), 0x5c, 0xd9}
	wire = append(wire, tail[:]...)
	path := filepath.Join(t.TempDir(), "native-floor.bin")
	if err := os.WriteFile(path, wire, 0o600); err != nil {
		t.Fatal(err)
	}
	cf, err := cryptfile.OpenFile(path, cryptfile.ReadOnly, -1)
	if err != nil {
		t.Fatal(err)
	}
	previous := cryptfile.Global()
	cryptfile.SetGlobal(cf)
	defer func() {
		cryptfile.SetGlobal(previous)
		if err := cf.Close(); err != nil {
			t.Error(err)
		}
	}()
	type regionBox struct {
		Prefix [4]uint32
		Region [8]uint32
		Suffix [4]uint32
	}
	box, free := alloc.New(regionBox{})
	defer free()
	*box = regionBox{
		Prefix: [4]uint32{0x11223344, 0x55667788, 0x99aabbcc, 0xddeeff00},
		Region: [8]uint32{0x13243546, 23, 46, 0x24354657, 92, 0x35465768, 0x46576879, 115},
		Suffix: [4]uint32{0xa7b8c9da, 0x1b2c3d4e, 0x5f607182, 0x93a4b5c6},
	}
	before := *box
	region := unsafe.Pointer(&box.Region[0])
	if uintptr(region) <= math.MaxUint32 {
		t.Fatalf("C allocation did not exercise a pointer above 4GiB: %p", region)
	}
	if err := Nox_server_mapRWFloorMap_422230(cf, region); err != nil {
		t.Fatal(err)
	}
	if *box != before {
		t.Fatal("FloorMap changed region or allocation canaries")
	}
	pos, err := cf.File.Seek(0, io.SeekCurrent)
	if err != nil || pos != int64(consumed) {
		t.Fatalf("actual file position=%d want=%d err=%v", pos, consumed, err)
	}
	var got [4]byte
	if _, err := io.ReadFull(cf, got[:]); err != nil || got != tail {
		t.Fatalf("trailing canary=%x want=%x err=%v", got, tail, err)
	}
	var extra [1]byte
	if n, err := cf.Read(extra[:]); n != 0 || !errors.Is(err, io.EOF) {
		t.Fatalf("expected EOF after exact canary, got=%d err=%v", n, err)
	}
	t.Logf("R300_FLOOR_CGO_OK version=%d bytes=%d", version, consumed)
}
