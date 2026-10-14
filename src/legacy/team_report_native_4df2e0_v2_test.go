package legacy

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

// Extract full active bodies without rewriting them. Fixture services match
// those declared by the original PE execution, but native addresses stay wide.
func teamReportBody210(t *testing.T, path, signature string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	start := strings.Index(text, signature)
	if start < 0 || strings.Count(text, signature) != 1 {
		t.Fatal("ambiguous production function", path, signature)
	}
	brace := strings.Index(text[start:], "{") + start
	depth := 0
	for i := brace; i < len(text); i++ {
		switch text[i] {
		case '{':
			depth++
		case '}':
			depth--
		}
		if depth == 0 {
			return text[start : i+1]
		}
	}
	t.Fatal("unterminated production body")
	return ""
}

func testTeamReport210(t *testing.T, mode string, first int) {
	t.Helper()
	compiler, err := exec.LookPath("cc")
	if err != nil {
		t.Fatal(err)
	}
	fixture, err := os.ReadFile("testdata/team_report_native_4df2e0.c")
	if err != nil {
		t.Fatal(err)
	}
	text := string(fixture)
	text = strings.ReplaceAll(text, "/* PRODUCTION_MEMBER_BODY */",
		teamReportBody210(t, "GAME1_1.c", "void sub_4198A0("))
	text = strings.ReplaceAll(text, "/* PRODUCTION_ENUMERATION_BODY */",
		teamReportBody210(t, "GAME3_2.c", "void sub_4DF2E0("))
	directory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	temp := t.TempDir()
	source, binary := filepath.Join(temp, "team-report.c"), filepath.Join(temp, "team-report")
	if err := os.WriteFile(source, []byte(text), 0600); err != nil {
		t.Fatal(err)
	}
	args := []string{"-std=gnu11", "-O2", "-g", "-fshort-wchar", "-fsigned-char", "-fno-strict-aliasing",
		"-Wno-int-conversion", "-I", directory, source, "-o", binary}
	if runtime.GOARCH == "386" {
		args = append([]string{"-m32"}, args...)
	}
	if output, err := exec.Command(compiler, args...).CombinedOutput(); err != nil {
		t.Fatalf("compile exact production bodies and headers: %v\n%s", err, output)
	}
	data, err := os.ReadFile("testdata/team_report_original_4df2e0_v2.json")
	if err != nil {
		t.Fatal(err)
	}
	var original struct {
		OriginalSHA256 string            `json:"original_sha256"`
		Cases          []json.RawMessage `json:"cases"`
		MemoryFaults   []any             `json:"memory_faults"`
		Guards         bool              `json:"code_and_input_guards_unchanged"`
	}
	if err := json.Unmarshal(data, &original); err != nil {
		t.Fatal(err)
	}
	if original.OriginalSHA256 != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || len(original.Cases) != 10 || len(original.MemoryFaults) != 0 || !original.Guards {
		t.Fatal("sealed original provenance or complete case set missing")
	}
	var want []any
	selected, err := json.Marshal(original.Cases[first : first+5])
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(selected, &want); err != nil {
		t.Fatal(err)
	}
	for _, pattern := range []int{0, 0x5a, 0xff} {
		output, err := exec.Command(binary, mode, fmt.Sprint(pattern)).CombinedOutput()
		if err != nil {
			t.Fatalf("actual native %s body pattern=%02x failed: %v\n%s", mode, pattern, err, output)
		}
		var got []any
		if err := json.Unmarshal(output, &got); err != nil {
			t.Fatalf("native transcript: %v\n%s", err, output)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("native %s pattern=%02x differs from original\ngot=%s\nwant=%s", mode, pattern, output, selected)
		}
	}
}

func TestTeamReportEnumerationNative4DF2E0MatchesOriginalV2(t *testing.T) {
	testTeamReport210(t, "enumeration", 0)
}

func TestTeamMemberPacketNative4198A0MatchesOriginalV2(t *testing.T) {
	testTeamReport210(t, "packet", 5)
}
