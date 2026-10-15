package legacy

import (
	"bytes"
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

func TestRegisteredTeamCount417DE0CompleteNativeBody(t *testing.T) {
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Fatal(err)
	}
	fixture, err := os.ReadFile("testdata/team_registered_count_417de0.c")
	if err != nil {
		t.Fatal(err)
	}
	source := strings.ReplaceAll(string(fixture), "/* COMPLETE_COUNT_BODY */",
		teamReportBody210(t, "GAME1.c", "char sub_417DE0("))
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	temp := t.TempDir()
	file, binary := filepath.Join(temp, "team-count.c"), filepath.Join(temp, "team-count")
	if err := os.WriteFile(file, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	args := []string{"-std=gnu11", "-O2", "-g", "-fshort-wchar", "-fsigned-char", "-fno-strict-aliasing", "-fno-strict-overflow", "-Wno-incompatible-pointer-types", "-Wno-int-conversion", "-I", dir, file, "-o", binary}
	if runtime.GOARCH == "386" {
		args = append([]string{"-m32"}, args...)
	}
	if output, err := exec.Command(cc, args...).CombinedOutput(); err != nil {
		t.Fatalf("compile complete team count body: %v\n%s", err, output)
	}
	if output, err := exec.Command(binary).CombinedOutput(); err != nil {
		t.Fatalf("native team count: %v\n%s", err, output)
	} else {
		t.Log(strings.TrimSpace(string(output)))
	}
}

func TestRegisteredTeamCount417DE0CGoUsesNativeServerTeams(t *testing.T) {
	const helper = "OPENNOX_TEST_REGISTERED_TEAM_COUNT_417DE0"
	if os.Getenv(helper) != "1" {
		cmd := exec.Command(os.Args[0], "-test.run=^TestRegisteredTeamCount417DE0CGoUsesNativeServerTeams$", "-test.v")
		cmd.Env = append(os.Environ(), helper+"=1")
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("actual CGo team enumeration: %v\n%s", err, output)
		} else {
			t.Log(strings.TrimSpace(string(output)))
		}
		return
	}
	oldServer, oldGame := GetServer, noxflags.GetGame()
	defer func() {
		GetServer = oldServer
		noxflags.ResetGame()
		noxflags.SetGame(oldGame)
	}()
	// Suppress the existing Create announcement in this isolated server fixture;
	// the counting function neither consumes nor changes gameplay flags.
	noxflags.SetGame(noxflags.GameModeCoopTeam)
	for _, count := range []int{0, 1, 3, 16} {
		srv := new(server.Server)
		teams, free := alloc.Make([]server.Team{}, 17)
		srv.Teams.Arr = teams
		GetServer = func() Server { return &netClientSendTestServer{srv: srv} }
		for i := 1; i <= count; i++ {
			team := srv.Teams.Create(server.TeamID(i))
			if team == nil {
				t.Fatal("native team creation failed")
			}
			if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(team.C()) <= math.MaxUint32 {
				t.Fatal("native team pointer is not above 4 GiB")
			}
		}
		for mode := 0; mode < 4; mode++ {
			want := 0
			for i := 1; i <= count; i++ {
				value := uint32(0)
				if mode == 1 && i%2 == 1 {
					value = 1
				} else if mode == 2 {
					value = 0x80000000
				} else if mode == 3 {
					value = 0xffffffff
				}
				// Fixed scalar ABI input at byte60, also read through Team.Ind60.
				*(*uint32)(unsafe.Add(teams[i].C(), 60)) = value
				if teams[i].Ind60() != 0 {
					want++
				}
			}
			raw := unsafe.Slice((*byte)(unsafe.Pointer(&teams[0])), len(teams)*int(unsafe.Sizeof(teams[0])))
			before := bytes.Clone(raw)
			if got := Sub_417DE0(); got != want {
				t.Fatalf("native CGo count=%d, want %d, active=%d mode=%d", got, want, count, mode)
			}
			if !bytes.Equal(raw, before) {
				t.Fatal("read-only team count mutated native team data")
			}
		}
		free()
	}
	t.Log("actual CGo First/Next and native server team storage: 16 cases passed")
}
