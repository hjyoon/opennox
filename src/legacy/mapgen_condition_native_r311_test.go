package legacy

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
	"unsafe"

	playerlib "github.com/opennox/libs/player"
	"github.com/opennox/libs/strman"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/internal/binfile"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/server"
)

// Exercise actual encrypted files, conditional recursion and native player
// iteration. Child isolation also keeps the production percentage RNG state
// and binfile-handle lifecycle independent of every pre-existing test.
func TestMapgenConditionNativeEncryptedR311(t *testing.T) {
	const child = "OPENNOX_TEST_MAPGEN_CONDITION_R311"
	if os.Getenv(child) != "1" {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestMapgenConditionNativeEncryptedR311$", "-test.v")
		cmd.Env = append(os.Environ(), child+"=1")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("actual encrypted conditional parser: %v\n%s", err, out)
		}
		t.Log(strings.TrimSpace(string(out)))
		return
	}
	srv := server.New(nil, nil, strman.New())
	t.Cleanup(srv.Close)
	oldServer := GetServer
	GetServer = func() Server { return &questCanJoinLegacyServer4E4100{srv: srv} }
	t.Cleanup(func() { GetServer = oldServer })
	handles.Init()
	t.Cleanup(handles.Release)
	t.Chdir(t.TempDir())
	if err := os.Mkdir("mapgen", 0o700); err != nil {
		t.Fatal(err)
	}
	native64 := unsafe.Sizeof(uintptr(0)) == 8
	buffer := unsafe.Slice((*byte)(memmap.PtrOff(0x5D4594, 2487264)), 256)
	oldBuffer := bytes.Clone(buffer)
	lineCounter := memmap.PtrUint32(0x5D4594, 2487520)
	oldCounter := *lineCounter
	t.Cleanup(func() { copy(buffer, oldBuffer); *lineCounter = oldCounter })
	cases := 0
	write := func(contents string) string {
		t.Helper()
		name := fmt.Sprintf("case%03d", cases)
		path := filepath.Join("mapgen", name+".thm")
		file, err := binfile.BinfileOpen(path, binfile.WriteOnly)
		if err != nil {
			t.Fatal(err)
		}
		if err := file.SetKey(1); err != nil {
			t.Fatal(err)
		}
		if n, err := file.Write([]byte(contents)); err != nil || n != len(contents) {
			t.Fatalf("write encrypted fixture: %d/%d, %v", n, len(contents), err)
		}
		if err := file.Close(); err != nil {
			t.Fatal(err)
		}
		return path
	}
	playerBytes := func() []byte {
		all := srv.Players.ListSlots()
		return unsafe.Slice((*byte)(all[0].C()), len(all)*int(unsafe.Sizeof(*all[0])))
	}
	setPlayers := func(classes []byte, inactive int) {
		t.Helper()
		srv.Players.ResetAll()
		for i, class := range classes {
			p := srv.Players.NewRaw(100 + i)
			if p == nil {
				t.Fatal("private native player allocation failed")
			}
			p.Info().SetPlayerClass(playerlib.Class(class))
			if native64 {
				if uintptr(p.C()) <= 0xffffffff {
					t.Fatal("player address did not exercise native pointer width")
				}
				// The original class offset must not determine native class.
				*(*byte)(unsafe.Add(p.C(), 2251)) = 0xff
			}
		}
		if inactive >= 0 {
			srv.Players.ListSlots()[inactive].Active = 0
		}
	}
	check := func(contents string, mode int, legacy, success bool, value uint32, tokens []string) {
		t.Helper()
		before := bytes.Clone(playerBytes())
		path := write(contents)
		got := mapgenConditionFixtureR311(path, mode, legacy)
		if got.result != success || !got.guardsPreserved || got.value != value || !reflect.DeepEqual(got.tokens, tokens) ||
			!bytes.Equal(before, playerBytes()) || got.fileAddress == 0 || native64 && got.fileAddress <= 0xffffffff || legacy && !got.highToken {
			t.Fatalf("mode=%d legacy=%t input=%q result=%+v want=%t/%#x/%v or player/output mutation", mode, legacy, contents, got, success, value, tokens)
		}
		files.RLock()
		defer files.RUnlock()
		found := 0
		for _, file := range files.byHandle {
			if file.Name() == path {
				found++
				var probe [1]byte
				if _, err := file.Read(probe[:]); !errors.Is(err, os.ErrClosed) {
					t.Fatalf("encrypted fixture handle remained open: %v", err)
				}
			}
		}
		if found != 1 {
			t.Fatalf("file closure observations=%d, want one", found)
		}
		cases++
	}
	const unchanged = uint32(0xA5A5A5A5)
	for _, legacy := range []bool{false, true} {
		for i, operator := range []string{"<", ">", "==", "!="} {
			check(operator+"\n", 1, legacy, true, uint32(i), nil)
		}
		for _, invalid := range []string{"", "<=\n", "?\n"} {
			check(invalid, 1, legacy, false, unchanged, nil)
		}
		for _, pattern := range []struct {
			classes  []byte
			inactive int
			want     [3]uint32
		}{
			{nil, -1, [3]uint32{0, 0, 0}},
			{[]byte{0}, -1, [3]uint32{1, 0, 0}},
			{[]byte{1}, -1, [3]uint32{0, 1, 0}},
			{[]byte{2}, -1, [3]uint32{0, 0, 1}},
			{[]byte{255, 2, 1, 0}, -1, [3]uint32{1, 1, 1}},
			{[]byte{1, 255, 0, 2}, 0, [3]uint32{1, 0, 1}},
		} {
			setPlayers(pattern.classes, pattern.inactive)
			for i, name := range []string{"warrior", "WIZARD", "Conjurer"} {
				check(name+"\n", 0, legacy, true, pattern.want[i], nil)
			}
		}
		for _, count := range []int{0, 1, 3} {
			setPlayers(bytes.Repeat([]byte{0}, count), -1)
			for _, row := range []struct {
				threshold  int
				experience [4]uint32
			}{
				{-1, [4]uint32{0, 1, 0, 1}},
				{2, [4]uint32{0, 1, 0, 1}},
				{3, [4]uint32{0, 0, 1, 0}},
				{4, [4]uint32{1, 0, 0, 1}},
			} {
				truth := []bool{count < row.threshold, count > row.threshold, count == row.threshold, count != row.threshold}
				for i, operator := range []string{"<", ">", "==", "!="} {
					var want uint32
					if truth[i] {
						want = 1
					}
					check(fmt.Sprintf("NUMPLAYERS %s %d\n", operator, row.threshold), 0, legacy, true, want, nil)
					check(fmt.Sprintf("EXPERIENCE_LEVEL %s %d\n", operator, row.threshold), 0, legacy, true, row.experience[i], nil)
				}
			}
		}
		for _, invalid := range []string{"", "UNKNOWN\n", "NUMPLAYERS ? 1\n", "NUMPLAYERS <\n", "EXPERIENCE_LEVEL\n", "EXPERIENCE_LEVEL <\n"} {
			check(invalid, 0, legacy, false, unchanged, nil)
		}
		check("0%\n", 0, legacy, true, 0, nil)
		check("100%\n", 0, legacy, true, 1, nil)
	}
	setPlayers([]byte{0, 1, 2}, -1)
	for _, row := range []struct {
		text string
		want []string
	}{
		{"IF WIZARD YES ELSE NO ENDIF TAIL\n", []string{"YES", "TAIL"}},
		{"IF 0% NO ELSE YES ENDIF TAIL\n", []string{"YES", "TAIL"}},
		{"IF WARRIOR IF CONJURER BOTH ELSE WRONG ENDIF ELSE WRONG ENDIF TAIL\n", []string{"BOTH", "TAIL"}},
		{"IF 0% IF 100% WRONG ELSE WRONG ENDIF ELSE OUTER ENDIF TAIL\n", []string{"OUTER", "TAIL"}},
		{"// comment\nIF NUMPLAYERS == 3 THREE ENDIF TAIL\n", []string{"THREE", "TAIL"}},
		{"IF 0% OMIT ENDIF TAIL\n", []string{"TAIL"}},
		{"IF UNKNOWN\n", nil},
		{"IF 0% OMIT\n", nil},
	} {
		check(row.text, 2, false, len(row.want) != 0, unchanged, row.want)
	}
	for _, class := range []byte{0, 1} {
		setPlayers([]byte{class}, -1)
		for _, legacy := range []bool{false, true} {
			path := write("ALGORITHM_DATA\nIF WIZARD\nmidHallLength\n17\nELSE\nmidHallLength\n5\nENDIF\nEND\nDECOR\nROOM\nRoomBase\nEND\nDECOR\nHALL\nHallBase\nEND\n")
			got := mapgenThemeFixtureR310(strings.TrimSuffix(filepath.Base(path), ".thm"), legacy)
			want := uint32(5)
			if class == 1 {
				want = 17
			}
			if !got.result || got.words[1] != want || !got.cleanupCleared || !got.guardsPreserved {
				t.Fatalf("conditional whole theme class=%d legacy=%t got=%+v", class, legacy, got)
			}
			cases++
		}
	}
	t.Logf("R311 actual encrypted condition/comparison/token/theme cases: %d; native players, handles, adapters, nesting and unchanged output/player bytes", cases)
}
