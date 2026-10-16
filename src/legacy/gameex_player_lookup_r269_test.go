package legacy

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"unsafe"

	playerlib "github.com/opennox/libs/player"
	"github.com/opennox/libs/strman"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// GameEx's optional name messages reach the real C parsers. Keep their
// historical low-byte wide-name conversion and exact, first-active match;
// neither a UTF-8 comparison nor the packed original-name field is equivalent.
// This is a GameEx.c contract test, not a GAME.EXE or GameEx.dll oracle.
func TestGameExPlayerLookupR269(t *testing.T) {
	const child = "OPENNOX_TEST_GAMEEX_PLAYER_LOOKUP_R269"
	for _, parser := range []string{"0", "1"} {
		t.Run("parser"+parser, func(t *testing.T) {
			if os.Getenv(child) != parser {
				cmd := exec.Command(os.Args[0], "-test.run=^TestGameExPlayerLookupR269$/^parser"+parser+"$", "-test.v")
				cmd.Env = append(os.Environ(), child+"="+parser)
				output, err := cmd.CombinedOutput()
				if err != nil {
					t.Fatalf("actual CGo GameEx parser%s: %v\n%s", parser, err, output)
				}
				t.Log(strings.TrimSpace(string(output)))
				return
			}

			var lookupWithTeam func(unsafe.Pointer, *unsafe.Pointer) int
			if parser == "1" {
				var ok bool
				lookupWithTeam, ok = any(PlayerInfoStructParser_1).(func(unsafe.Pointer, *unsafe.Pointer) int)
				if !ok {
					t.Fatal("GameEx parser1 still exports a 32-bit team-pointer output")
				}
			}
			s := server.New(nil, nil, strman.New())
			defer s.Close()
			previous, previousFlags := GetServer, noxflags.GetGame()
			GetServer = func() Server { return &netClientSendTestServer{srv: s} }
			noxflags.ResetGame()
			noxflags.SetGame(noxflags.GameHost)
			defer func() {
				GetServer = previous
				noxflags.ResetGame()
				noxflags.SetGame(previousFlags)
			}()

			names := []string{"Alpha", "inactive", "Beta", strings.Repeat("x", 27), "é", "Ё", "ĀTail", "Alpha"}
			units, freeUnits := alloc.Make([]server.Object{}, len(names))
			defer freeUnits()
			for i, name := range names {
				p := s.Players.NewRaw(0x12340000 + i)
				if p == nil {
					t.Fatal("private player slot unavailable")
				}
				p.SetName(name)
				p.Info().SetName("not-the-approved-name")
				p.Info().SetPlayerClass(playerlib.Class(i % 3))
				units[i].NetCode = uint32(p.NetCode())
				units[i].TeamVal.ID = server.TeamID(0xf0 + i)
				p.PlayerUnit = &units[i]
				if unsafe.Sizeof(uintptr(0)) > 4 && (uintptr(p.C()) <= 0xffffffff || uintptr(units[i].TeamVal.C()) <= 0xffffffff) {
					t.Fatal("native player and object-team must be above 4 GiB")
				}
			}
			// An inactive earlier duplicate must not shadow the later live match.
			inactive := s.Players.ListSlots()[1]
			inactive.SetName("Beta")
			inactive.Active = 0
			playerBytes := func() []byte {
				p := s.Players.ListSlots()
				return unsafe.Slice((*byte)(p[0].C()), len(p)*int(unsafe.Sizeof(*p[0])))
			}
			unitsBytes := unsafe.Slice((*byte)(unsafe.Pointer(&units[0])), len(units)*int(unsafe.Sizeof(units[0])))
			playersBefore, unitsBefore := bytes.Clone(playerBytes()), bytes.Clone(unitsBytes)
			cases := []struct {
				name  string
				query []byte
				index int
			}{
				{"first-and-duplicate", []byte("Alpha"), 0},
				{"inactive-skip-and-next", []byte("Beta"), 2},
				{"maximum-approved-name", []byte(strings.Repeat("x", 27)), 3},
				{"legacy-Latin1-not-UTF8", []byte{0xe9}, 4},
				{"legacy-wide-low-byte", []byte{1}, 5},
				{"legacy-embedded-low-zero", nil, 6},
				{"exact-case-miss", []byte("alpha"), -1},
				{"packed-original-name-miss", []byte("not-the-approved-name"), -1},
				{"UTF8-is-not-legacy-OEM", []byte("é"), -1},
				{"prefix-miss", []byte("Alph"), -1},
				{"missing", []byte("missing"), -1},
			}
			for _, pattern := range []byte{0, 0x5a, 0xff} {
				for _, tc := range cases {
					storage, free := alloc.Make([]byte{}, len(tc.query)+19)
					for i := range storage {
						storage[i] = pattern
					}
					record := storage[8:]
					record[0], record[1] = 0xa5, 0x5a
					copy(record[2:], tc.query)
					record[len(tc.query)+2] = 0
					want := bytes.Clone(storage)
					var wantTeam unsafe.Pointer
					wantResult := 0
					if tc.index >= 0 {
						wantResult = 1
						want[8], want[9] = byte(tc.index%3), byte(0xf0+tc.index)
						wantTeam = units[tc.index].TeamVal.C()
					}
					type teamOutput struct {
						before uintptr
						team   unsafe.Pointer
						after  uintptr
					}
					out, freeOut := alloc.New(teamOutput{})
					*out = teamOutput{before: ^uintptr(0) - 1, team: units[7].TeamVal.C(), after: ^uintptr(0) - 2}
					got := 0
					if parser == "0" {
						got = PlayerInfoStructParser_0(unsafe.Pointer(&record[0]))
					} else {
						got = lookupWithTeam(unsafe.Pointer(&record[0]), &out.team)
						if tc.index < 0 {
							wantTeam = units[7].TeamVal.C()
						}
						if out.team != wantTeam || out.before != ^uintptr(0)-1 || out.after != ^uintptr(0)-2 {
							t.Fatalf("%s guard=%02x output=%+v want native team=%p", tc.name, pattern, *out, wantTeam)
						}
					}
					if got != wantResult || !bytes.Equal(storage, want) || !bytes.Equal(playerBytes(), playersBefore) || !bytes.Equal(unitsBytes, unitsBefore) {
						t.Fatalf("%s guard=%02x result=%d want=%d or record/player/object bytes changed", tc.name, pattern, got, wantResult)
					}
					freeOut()
					free()
				}
			}
			for _, p := range s.Players.ListSlots() {
				p.Active = 0
			}
			record, free := alloc.Make([]byte{}, 4)
			defer free()
			record[0], record[1] = 0xa5, 0x5a
			got := PlayerInfoStructParser_0(unsafe.Pointer(&record[0]))
			if parser == "1" {
				out, free := alloc.New(unsafe.Pointer(nil))
				defer free()
				got = lookupWithTeam(unsafe.Pointer(&record[0]), out)
				if *out != nil {
					t.Fatal("empty roster changed team output")
				}
			}
			if got != 0 || record[0] != 0xa5 || record[1] != 0x5a {
				t.Fatal("empty roster changed result bytes")
			}
			fmt.Printf("R269 actual CGo parser%s: 33 guarded name/class/native-team cases plus empty roster passed\n", parser)
		})
	}
}
