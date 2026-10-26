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
)

func clientCenterMessages202(t *testing.T, mode string) {
	t.Helper()
	var original struct {
		SHA  string `json:"original_sha256"`
		Code []struct {
			Entry string `json:"entry"`
			Bytes int    `json:"bytes"`
			SHA   string `json:"sha256"`
		} `json:"code"`
		Offsets [8]int32 `json:"shadow_offsets"`
		Clear   []struct {
			Input struct {
				Index int64    `json:"index"`
				Text  []uint16 `json:"text"`
			} `json:"input"`
			Expected struct {
				Events   [][]json.RawMessage `json:"events"`
				Returned uint32              `json:"returned_slot"`
			} `json:"expected"`
		} `json:"clear_rows"`
		Draw []struct {
			Input struct {
				Index    int64 `json:"index"`
				Mask     int64 `json:"mask"`
				Frame    int64 `json:"frame"`
				Width    int64 `json:"width"`
				Height   int64 `json:"height"`
				Top      int64 `json:"top"`
				Measured int64 `json:"measured"`
				Font     int64 `json:"font"`
			} `json:"input"`
			Expected struct {
				Events   [][]json.RawMessage `json:"events"`
				Returned uint32              `json:"returned"`
			} `json:"expected"`
		} `json:"draw_rows"`
	}
	data, err := os.ReadFile("testdata/center_messages_445450_445530_original.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(data, &original); err != nil {
		t.Fatal(err)
	}
	wantCode := []struct {
		Entry string
		Bytes int
		SHA   string
	}{
		{"0x445450", 62, "781c952cefae34561c4aa25169c026acfa3295d07b9ddf9cee1ed503f31a072d"},
		{"0x445530", 287, "8a53fdc7b4eeb177882165d694d845f2af0942d22c142fab03aae32d8218b830"},
		{"0x565ca3", 37, "f99f7f2e1f82d316e97935cfa2609fac9456e62a80aee9256a381f843031b294"},
	}
	if original.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || len(original.Code) != 3 || len(original.Clear) != 9 || len(original.Draw) != 288 {
		t.Fatal("wrong original centered HUD entries")
	}
	for i, want := range wantCode {
		got := original.Code[i]
		if got.Entry != want.Entry || got.Bytes != want.Bytes || got.SHA != want.SHA {
			t.Fatal("original code differs")
		}
	}
	var input bytes.Buffer
	write := func(v any) {
		if err := binary.Write(&input, binary.LittleEndian, v); err != nil {
			t.Fatal(err)
		}
	}
	write(original.Offsets)
	write(uint32(len(original.Clear) + len(original.Draw)))
	emit := func(fields [10]int64, events [][]json.RawMessage, result uint32) {
		var trace [32][5]int64
		ids := map[string]int64{"viewport": 1, "color": 2, "measure": 3, "draw": 4, "height": 5, "copy": 6}
		if len(events) > len(trace) {
			t.Fatal("trace exceeds fixture")
		}
		for i, event := range events {
			var name string
			if err := json.Unmarshal(event[0], &name); err != nil {
				t.Fatal(err)
			}
			id, ok := ids[name]
			if !ok || len(event) > 5 {
				t.Fatal("unknown original event")
			}
			trace[i][0] = id
			for j, arg := range event[1:] {
				if err := json.Unmarshal(arg, &trace[i][j+1]); err != nil {
					t.Fatal(err)
				}
			}
		}
		write(fields)
		write(uint32(len(events)))
		write(trace)
		write(result)
	}
	for _, row := range original.Clear {
		kind := int64(0)
		if len(row.Input.Text) == 1 {
			kind = 1
		} else if len(row.Input.Text) == 4 {
			kind = 2
		} else if len(row.Input.Text) != 0 {
			t.Fatal("unknown original UTF-16 source")
		}
		emit([10]int64{0, row.Input.Index, 0, 0, 0, 0, 0, 0, 0, kind}, row.Expected.Events, row.Expected.Returned)
	}
	for _, row := range original.Draw {
		x := row.Input
		emit([10]int64{1, x.Index, x.Mask, x.Frame, x.Width, x.Height, x.Top, x.Measured, x.Font, 0}, row.Expected.Events, row.Expected.Returned)
	}
	source, err := os.ReadFile("GAME1_3.c")
	if err != nil {
		t.Fatal(err)
	}
	first, last := "//----- (00445450)", "//----- (004456E0)"
	if strings.Count(string(source), first) != 1 || strings.Count(string(source), last) != 1 {
		t.Fatal("ambiguous complete HUD bodies")
	}
	_, tail, _ := strings.Cut(string(source), first)
	body, _, _ := strings.Cut(tail, last)
	fixture, err := os.ReadFile("testdata/client_center_messages_test.c")
	if err != nil {
		t.Fatal(err)
	}
	program := strings.Replace(string(fixture), "// CENTER_BODIES", first+body, 1)
	compiler := strings.Fields(os.Getenv("CC"))
	if len(compiler) == 0 {
		compiler = []string{"cc"}
	}
	bin := filepath.Join(t.TempDir(), "center-messages")
	args := append(compiler[1:], "-std=c11", "-O2", "-fshort-wchar", "-fsigned-char", "-fno-strict-aliasing", "-fno-strict-overflow", "-Wall", "-Wextra", "-Werror", "-Itestdata", "-x", "c", "-", "-o", bin)
	cmd := exec.Command(compiler[0], nativeFixtureCompileArgs(t, args)...)
	cmd.Stdin = strings.NewReader(program)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("compile complete HUD bodies: %v\n%s", err, out)
	}
	cmd = exec.Command(bin, mode)
	cmd.Stdin = bytes.NewReader(input.Bytes())
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("actual centered %s body: %v\n%s", mode, err, out)
	} else {
		t.Log(strings.TrimSpace(string(out)))
	}
}

func TestClientCenterClear445450NativeOriginal(t *testing.T) { clientCenterMessages202(t, "clear") }
func TestClientCenterDraw445530NativeOriginal(t *testing.T)  { clientCenterMessages202(t, "draw") }
