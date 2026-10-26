package legacy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func testKOTRMapExit216(t *testing.T, first, last int) {
	t.Helper()
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Fatal(err)
	}
	fixture, err := os.ReadFile("testdata/kotr_map_exit_4d2160.c")
	if err != nil {
		t.Fatal(err)
	}
	source := strings.ReplaceAll(string(fixture), "/* COMPLETE_KOTR_MAP_EXIT_BODY */",
		teamReportBody210(t, "server__system__server.c", "void sub_4D2160("))
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	temp := t.TempDir()
	file, binary := filepath.Join(temp, "kotr.c"), filepath.Join(temp, "kotr")
	if err := os.WriteFile(file, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	args := []string{"-std=gnu11", "-O2", "-g", "-fshort-wchar", "-fsigned-char", "-fno-strict-aliasing", "-fno-strict-overflow", "-Wno-int-conversion", "-Wno-incompatible-pointer-types", "-I", dir, file, "-o", binary}
	if runtime.GOARCH == "386" {
		args = append([]string{"-m32"}, args...)
	}
	if output, err := exec.Command(cc, nativeFixtureCompileArgs(t, args)...).CombinedOutput(); err != nil {
		t.Fatalf("compile complete KOTR body: %v\n%s", err, output)
	}
	data, err := os.ReadFile("testdata/kotr_map_exit_original_4d2160_v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var original struct {
		OriginalSHA256 string `json:"original_sha256"`
		BodySHA256     string `json:"body_sha256"`
		BodyBytes      int    `json:"body_bytes"`
		Guards         bool   `json:"code_and_input_guards_unchanged"`
		Registers      bool   `json:"successful_callee_saved_registers_unchanged"`
		Cases          []struct {
			Index   int      `json:"index"`
			Name    string   `json:"name"`
			Pattern int      `json:"pattern"`
			Events  []string `json:"events"`
			Fault   bool     `json:"fault"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(data, &original); err != nil {
		t.Fatal(err)
	}
	if original.OriginalSHA256 != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || original.BodySHA256 != "4e636081e9b36e3bb9532c6fefec50c6941e30eb2e35fb936afb87c181415e63" || original.BodyBytes != 208 || !original.Guards || !original.Registers || len(original.Cases) != 48 {
		t.Fatal("sealed complete original provenance missing")
	}
	count := 0
	for _, c := range original.Cases {
		if c.Index < first || c.Index >= last {
			continue
		}
		output, err := exec.Command(binary, fmt.Sprint(c.Index), fmt.Sprint(c.Pattern)).CombinedOutput()
		if c.Fault {
			if exit, ok := err.(*exec.ExitError); !ok || exit.ExitCode() != 86 {
				t.Fatalf("%s/%02x: original fault not observed: %v\n%s", c.Name, c.Pattern, err, output)
			}
		} else if err != nil {
			t.Fatalf("%s/%02x: native execution: %v\n%s", c.Name, c.Pattern, err, output)
		}
		want := strings.Join(c.Events, "\n") + "\n"
		if string(output) != want {
			t.Fatalf("%s/%02x: native calls differ from original\ngot=%s\nwant=%s", c.Name, c.Pattern, output, want)
		}
		count++
	}
	if count != (last-first)*3 {
		t.Fatal("incomplete original case selection")
	}
	t.Logf("complete native body: %d original count/iterator/RNG/pickup/guard/fault conditions", count)
}

func TestKOTRMapExit4D2160SelectionMatchesOriginal(t *testing.T)     { testKOTRMapExit216(t, 0, 9) }
func TestKOTRMapExit4D2160ReloadAndFaultsMatchOriginal(t *testing.T) { testKOTRMapExit216(t, 9, 16) }

func TestKOTRMapExit4D2160CGoReadsNativeCrownSideTable(t *testing.T) {
	const helper = "OPENNOX_TEST_KOTR_MAP_EXIT_4D2160"
	if os.Getenv(helper) != "1" {
		cmd := exec.Command(os.Args[0], "-test.run=^TestKOTRMapExit4D2160CGoReadsNativeCrownSideTable$", "-test.v")
		cmd.Env = append(os.Environ(), helper+"=1")
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("actual CGo native crown storage: %v\n%s", err, output)
		} else {
			t.Log(strings.TrimSpace(string(output)))
		}
		return
	}
	oldServer, oldGame := GetServer, noxflags.GetGame()
	defer func() { GetServer = oldServer; noxflags.ResetGame(); noxflags.SetGame(oldGame) }()
	noxflags.SetGame(noxflags.GameModeCoopTeam)
	cases := 0
	for _, count := range []int{0, 1, 3} {
		for _, pattern := range []uint32{0, 0x5a5a5a5a, 0xffffffff} {
			for _, held := range []bool{false, true} {
				srv := new(server.Server)
				teams, freeTeams := alloc.Make([]server.Team{}, 17)
				objects, freeObjects := alloc.Make([]server.Object{}, 4)
				srv.Teams.Arr = teams
				GetServer = func() Server { return &netClientSendTestServer{srv: srv} }
				for i := 1; i <= count; i++ {
					team := srv.Teams.Create(server.TeamID(i))
					if team == nil {
						t.Fatal("native team creation failed")
					}
					// Scalar legacy prefix stays poison; native crown is only in side table.
					*(*uint32)(unsafe.Add(team.C(), 76+unsafe.Sizeof(uintptr(0))-4)) = pattern
					crown := &objects[i-1]
					if held {
						crown.InvHolder = &objects[3]
					}
					srv.Teams.SetTeamFlag(team, crown)
					if unsafe.Sizeof(uintptr(0)) > 4 && (uintptr(team.C()) <= math.MaxUint32 || uintptr(unsafe.Pointer(crown)) <= math.MaxUint32) {
						t.Fatal("native fixtures not above 4 GiB")
					}
				}
				rawTeams := unsafe.Slice((*byte)(unsafe.Pointer(&teams[0])), len(teams)*int(unsafe.Sizeof(teams[0])))
				rawObjects := unsafe.Slice((*byte)(unsafe.Pointer(&objects[0])), len(objects)*int(unsafe.Sizeof(objects[0])))
				beforeTeams, beforeObjects := bytes.Clone(rawTeams), bytes.Clone(rawObjects)
				Sub_4D2160()
				if !bytes.Equal(rawTeams, beforeTeams) || !bytes.Equal(rawObjects, beforeObjects) {
					t.Fatal("held or empty-member branch changed native input")
				}
				for i := 1; i <= count; i++ {
					if srv.Teams.TeamFlag(&teams[i]) != &objects[i-1] {
						t.Fatal("native crown side table changed")
					}
				}
				srv.Teams.Reset()
				freeObjects()
				freeTeams()
				cases++
			}
		}
	}
	t.Logf("actual CGo team/crown enumeration with poisoned legacy scalar: %d conditions", cases)
}
