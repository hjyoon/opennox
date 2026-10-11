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
	"reflect"
	"strconv"
	"strings"
	"testing"
)

// The immutable original corpus executes 45EF50 and 4281F0 unchanged, with
// explicitly declared window/drop/cast/repaint services, not a whole GUI oracle.
func TestQuickbarDragOut45EF50NativeOriginal(t *testing.T) {
	data, err := os.ReadFile("testdata/quickbar_drag_out_45ef50.json")
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(data)) != "2ab9d21d12fa7c8ddf4ad3f87742b25b498b10f9e86554535c0d26c2baca1c73" {
		t.Fatal("frozen unchanged-original corpus differs")
	}
	var fixture struct {
		Original string `json:"original_sha256"`
		Rows     []struct {
			Case struct {
				Name        string `json:"name"`
				Bar         int    `json:"bar"`
				Row         int    `json:"row"`
				Slot        int    `json:"slot"`
				Flag        int    `json:"flag"`
				Event       int    `json:"event"`
				X           int    `json:"x"`
				Y           int    `json:"y"`
				Target      int    `json:"target"`
				Dropped     uint32 `json:"dropped"`
				Mode        int    `json:"mode"`
				PlayerClass int    `json:"player_class"`
			} `json:"case"`
			Changes       [][3]int `json:"changed_bytes"`
			Calls         []string `json:"calls"`
			Press         int      `json:"press"`
			Release       int      `json:"release"`
			Captured      bool     `json:"captured"`
			LastButtonNil bool     `json:"last_button_nil"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Original != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || len(fixture.Rows) != 245 {
		t.Fatalf("unexpected original identity/case count: %s %d", fixture.Original, len(fixture.Rows))
	}
	csource, err := os.ReadFile("testdata/quickbar_drag_out_45ef50_test.c")
	if err != nil {
		t.Fatal(err)
	}
	program := string(csource)
	for _, part := range []struct{ file, signature, marker string }{
		{"GAME1_2.c", "int nox_xxx_pointInRect_4281F0(int2* a1, int4* a2) {", "// PRODUCTION_RECTANGLE"},
		{"GAME2.c", "int nox_xxx_quickBarWnd_45EF50(nox_window* win, int event, unsigned int pos) {", "// PRODUCTION_EVENT"},
	} {
		body, err := os.ReadFile(part.file)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Count(string(body), part.signature) != 1 || strings.Count(program, part.marker) != 1 {
			t.Fatal("ambiguous production/fixture anchor")
		}
		_, tail, _ := strings.Cut(string(body), part.signature)
		tail, _, ok := strings.Cut(tail, "\n}")
		if !ok {
			t.Fatal("missing complete production body")
		}
		program = strings.Replace(program, part.marker, part.signature+tail+"\n}", 1)
	}
	cc := strings.Fields(os.Getenv("CC"))
	if len(cc) == 0 {
		cc = []string{"cc"}
	}
	binaryPath := filepath.Join(t.TempDir(), "quickbar-drag-out")
	args := append(cc[1:], "-std=c11", "-O2", "-Wall", "-Wextra", "-Werror", "-Itestdata", "-x", "c", "-", "-o", binaryPath)
	cmd := exec.Command(cc[0], args...)
	cmd.Stdin = strings.NewReader(program)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("compile native production handlers: %v\n%s", err, output)
	}
	for _, row := range fixture.Rows {
		t.Run(row.Case.Name, func(t *testing.T) {
			c := row.Case
			var args []string
			for _, v := range []int64{int64(c.Bar), int64(c.Row), int64(c.Slot), int64(c.Flag), int64(c.Event), int64(c.X), int64(c.Y), int64(c.Target), int64(c.Dropped), int64(c.Mode), int64(c.PlayerClass)} {
				args = append(args, strconv.FormatInt(v, 10))
			}
			output, err := exec.Command(binaryPath, args...).CombinedOutput()
			if err != nil {
				t.Fatalf("C-owned native event boundary: %v\n%s", err, output)
			}
			lines := strings.Split(strings.TrimSpace(string(output)), "\n")
			if len(lines) != 4 {
				t.Fatalf("native record receipt: %s", output)
			}
			want := [2][]byte{bytes.Repeat([]byte{0xA5}, 201), bytes.Repeat([]byte{0xA5}, 201)}
			for b := range want {
				for r := 0; r < 5; r++ {
					for s := 0; s < 5; s++ {
						at := r*40 + s*8
						binary.LittleEndian.PutUint32(want[b][at:], uint32(27+r*5+s))
						want[b][at+4] = byte((r + s) & 1)
					}
				}
				want[b][200] = 0
			}
			want[c.Bar][200] = byte(c.Row)
			want[c.Bar][c.Row*40+c.Slot*8+4] = byte(c.Flag)
			for _, change := range row.Changes {
				want[change[0]][change[1]] = byte(change[2])
			}
			for b := range want {
				got, err := hex.DecodeString(lines[b])
				if err != nil || !bytes.Equal(got, want[b]) {
					t.Errorf("bar %d numeric spell/flag/padding/selected row: got %s want %x error=%v", b, lines[b], want[b], err)
				}
			}
			var calls []string
			if lines[2] != "" {
				calls = strings.Split(lines[2], ",")
			}
			if len(calls) != len(row.Calls) || len(calls) != 0 && !reflect.DeepEqual(calls, row.Calls) {
				t.Errorf("original service order: got %v want %v", calls, row.Calls)
			}
			bit := func(v bool) int {
				if v {
					return 1
				}
				return 0
			}
			if want := fmt.Sprintf("%d %d %d %d", row.Press, row.Release, bit(row.Captured), bit(row.LastButtonNil)); lines[3] != want {
				t.Errorf("original capture/cleanup/return: got %s want %s", lines[3], want)
			}
		})
	}
}
