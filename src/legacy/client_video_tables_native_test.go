package legacy

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func clientVideoTablesNative200(t *testing.T, mode string) {
	t.Helper()
	var original struct {
		SHA  string `json:"original_sha256"`
		Code []struct {
			Entry string `json:"entry"`
			Bytes int    `json:"bytes"`
			SHA   string `json:"sha256"`
		} `json:"code"`
		Calls     [][2]json.RawMessage `json:"calls"`
		Trig      []int32              `json:"trig_words"`
		Constants map[string]string    `json:"constants"`
	}
	raw, err := os.ReadFile("testdata/video_tables_4a9c80_4aee30_original.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &original); err != nil {
		t.Fatal(err)
	}
	if original.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || len(original.Calls) != 36 || len(original.Trig) != 320 || len(original.Code) != 3 ||
		original.Code[0].SHA != "77fb768c6d7f2d775d2f3c40b19a909dde71339720a7d35a61f5ae645853da47" ||
		original.Code[1].SHA != "c4c3aed4857ed3354964da85547e69cf852ccd97d7927f426a73cbbee321220f" ||
		original.Code[2].SHA != "d0a2aca68648a0ae9d5cbfae1b276d4a64d4f797fc0c60f28b64096cdfa3148c" {
		t.Fatal("wrong sealed original entries")
	}
	var names [36][32]byte
	var tokens [36]uint32
	for i, call := range original.Calls {
		var name string
		if err := json.Unmarshal(call[0], &name); err != nil {
			t.Fatal(err)
		}
		if len(name) >= 32 {
			t.Fatal("original name does not fit fixture")
		}
		copy(names[i][:], name)
		if err := json.Unmarshal(call[1], &tokens[i]); err != nil {
			t.Fatal(err)
		}
	}
	var input bytes.Buffer
	for _, v := range []any{names, tokens, original.Trig} {
		if err := binary.Write(&input, binary.LittleEndian, v); err != nil {
			t.Fatal(err)
		}
	}
	for _, address := range []string{"0x583a78", "0x583a70", "0x5839a0"} {
		value, err := hex.DecodeString(original.Constants[address])
		if err != nil || len(value) != 8 {
			t.Fatal("missing original double bytes")
		}
		input.Write(value)
	}
	program, err := os.ReadFile("testdata/client_video_tables_test.c")
	if err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile("GAME3.c")
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range []struct{ first, last, marker string }{
		{"//----- (004A9C80)", "//----- (004AA6B0)", "// COMPASS_BODY"},
		{"//----- (004AEE30)", "typedef struct nox_slider_data_4B4860", "// TRIG_BODY"},
	} {
		if strings.Count(string(source), entry.first) != 1 || strings.Count(string(source), entry.last) != 1 {
			t.Fatal("ambiguous complete video body")
		}
		_, body, _ := strings.Cut(string(source), entry.first)
		body, _, _ = strings.Cut(body, entry.last)
		program = []byte(strings.Replace(string(program), entry.marker, "//"+body, 1))
	}
	compiler := strings.Fields(os.Getenv("CC"))
	if len(compiler) == 0 {
		compiler = []string{"cc"}
	}
	bin := filepath.Join(t.TempDir(), "client-video-tables")
	args := append(compiler[1:], "-std=c11", "-O2", "-ffp-contract=off", "-fshort-wchar", "-fsigned-char", "-fno-strict-aliasing", "-Wall", "-Wextra", "-Werror", "-Itestdata", "-x", "c", "-", "-lm", "-o", bin)
	cmd := exec.Command(compiler[0], args...)
	cmd.Stdin = bytes.NewReader(program)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("compile complete video bodies: %v\n%s", err, output)
	}
	cmd = exec.Command(bin, mode)
	cmd.Stdin = bytes.NewReader(input.Bytes())
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("actual video tables %s: %v\n%s", mode, err, output)
	} else {
		t.Log(strings.TrimSpace(string(output)))
	}
}

func TestClientVideoCompass4A9C80NativeBounds(t *testing.T) { clientVideoTablesNative200(t, "compass") }
func TestClientVideoTrig4AEE30OriginalNativeBounds(t *testing.T) {
	clientVideoTablesNative200(t, "trig")
}
