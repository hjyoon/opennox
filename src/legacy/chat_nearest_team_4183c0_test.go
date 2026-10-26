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

type chatNearestOriginal215 struct {
	OriginalSHA string `json:"original_sha256"`
	Function    struct {
		Address string
		Bytes   int
		SHA256  string
	} `json:"function"`
	ControlWord string `json:"control_word"`
	Cases       []struct {
		Name           string
		Teams          []struct{ ID, X, Y, Flag uint32 }
		Units          []struct{ X, Y, Team, Status, Code, Update uint32 }
		Client, Engine uint32
		Events         []any
	}
	Executions int
	Faults     []any
	Guards     bool `json:"code_and_input_guards_unchanged"`
}

// The unchanged original PE32 body supplies the transcript; complete native
// production bodies and real layout headers supply the implementation under
// test. Only enumeration, engine flag storage and attach/change are controlled.
func testChatNearestTeam215(t *testing.T, boundary bool) {
	t.Helper()
	data, err := os.ReadFile("testdata/chat_nearest_original_4183c0_v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var original chatNearestOriginal215
	if err := json.Unmarshal(data, &original); err != nil {
		t.Fatal(err)
	}
	if original.OriginalSHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || original.Function.Address != "0x4183c0" || original.Function.Bytes != 272 || original.Function.SHA256 != "3201fec59d0de584d3bbf363ef55e9a03d1549629400e26666f50468eea3ed47" || original.ControlWord != "0x0e7f" || len(original.Cases) != 56 || original.Executions != 168 || len(original.Faults) != 0 || !original.Guards {
		t.Fatal("sealed original transcript scope changed")
	}
	fixture, err := os.ReadFile("testdata/chat_nearest_team_4183c0.c")
	if err != nil {
		t.Fatal(err)
	}
	production, err := os.ReadFile("GAME1.c")
	if err != nil {
		t.Fatal(err)
	}
	helpers := ""
	for _, signature := range []string{"static double nox_chat_team_add_chop53_4183C0(", "static double nox_chat_team_square_chop53_4183C0(", "static float nox_chat_team_spill_chop32_4183C0("} {
		if strings.Contains(string(production), signature) {
			helpers += teamReportBody210(t, "GAME1.c", signature) + "\n"
		}
	}
	source := strings.ReplaceAll(string(fixture), "/* CHAT_MATH_HELPERS */", helpers)
	source = strings.ReplaceAll(source, "/* COMPLETE_CHAT_BODY */", teamReportBody210(t, "GAME1.c", "int sub_4183C0("))
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Fatal(err)
	}
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	temp := t.TempDir()
	file, binary := filepath.Join(temp, "chat-nearest.c"), filepath.Join(temp, "chat-nearest")
	if err := os.WriteFile(file, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	args := []string{"-std=gnu11", "-O2", "-g", "-fshort-wchar", "-fsigned-char", "-fno-strict-aliasing", "-fno-strict-overflow", "-Wno-int-conversion", "-Wno-incompatible-pointer-types", "-I", dir, file, "-o", binary, "-lm"}
	if runtime.GOARCH == "386" {
		args = append([]string{"-m32"}, args...)
	}
	if out, err := exec.Command(cc, nativeFixtureCompileArgs(t, args)...).CombinedOutput(); err != nil {
		t.Fatalf("compile complete native Chat body: %v\n%s", err, out)
	}
	count := 0
	for index, row := range original.Cases {
		if isBoundary := index >= 8 && index < 11 || index >= 36; isBoundary != boundary {
			continue
		}
		for _, pattern := range []int{0, 0x5a, 0xff} {
			var input strings.Builder
			fmt.Fprintf(&input, "%d %d %d %d\n", len(row.Teams), len(row.Units), row.Client, row.Engine)
			for _, team := range row.Teams {
				fmt.Fprintf(&input, "%d %d %d %d\n", team.ID, team.X, team.Y, team.Flag)
			}
			for _, unit := range row.Units {
				fmt.Fprintf(&input, "%d %d %d %d %d %d\n", unit.X, unit.Y, unit.Team, unit.Status, unit.Code, unit.Update)
			}
			cmd := exec.Command(binary, fmt.Sprint(pattern))
			cmd.Stdin = strings.NewReader(input.String())
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("native Chat %s guard=%02x: %v\n%s", row.Name, pattern, err, out)
			}
			var events []any
			if err := json.Unmarshal(out, &events); err != nil {
				t.Fatalf("native Chat transcript: %v\n%s", err, out)
			}
			if !reflect.DeepEqual(events, row.Events) {
				t.Fatalf("native Chat %s guard=%02x differs from original\ngot=%s\nwant=%v", row.Name, pattern, out, row.Events)
			}
			count++
		}
	}
	t.Logf("complete native Chat body: %d original pointer/eligibility/distance/guard cases passed", count)
}

func TestChatNearestTeam4183C0NativePointersAndEligibility(t *testing.T) {
	testChatNearestTeam215(t, false)
}
func TestChatNearestTeam4183C0OriginalDistanceAndTieBoundaries(t *testing.T) {
	testChatNearestTeam215(t, true)
}
