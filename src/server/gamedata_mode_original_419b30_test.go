package server

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/opennox/libs/balance"
	"github.com/opennox/noxcrypt"
	noxflags "github.com/opennox/opennox/v1/common/flags"
)

func TestGamedataModeOriginal419B30(t *testing.T) {
	const fixturePath = "testdata/gamedata_mode_original.json"
	raw, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Contexts int `json:"contexts"`
		Modes    []struct {
			Name  string `json:"name"`
			Flags uint   `json:"flags"`
			Tag   string `json:"tag"`
		} `json:"modes"`
		Cases []struct {
			Mode     string `json:"flag_context"`
			Flags    uint   `json:"game_flags"`
			Key      string `json:"key"`
			Index    int    `json:"index"`
			Name     string `json:"name"`
			Expected uint64 `json:"expected_bits64"`
		} `json:"cases"`
		Stock []struct {
			Tag    string   `json:"tag"`
			Key    string   `json:"original_key"`
			Tokens []string `json:"tokens"`
		} `json:"stock_records"`
	}
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Contexts != 2568 || len(fixture.Cases) != 2568 || len(fixture.Modes) != 8 || len(fixture.Stock) != 215 {
		t.Fatal("incomplete original mode contexts")
	}
	var text strings.Builder
	for _, record := range fixture.Stock {
		if record.Tag != "GLOBAL" {
			fmt.Fprintf(&text, "%s ", record.Tag)
		}
		fmt.Fprintf(&text, "%s = %s;\n", record.Key, strings.Join(record.Tokens, ", "))
	}
	var encrypted bytes.Buffer
	w, err := crypt.NewWriter(&encrypted, crypt.GameDataBin)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte(text.String())); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), balance.GamedataFile)
	if err := os.WriteFile(path, encrypted.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	f, err := balance.ReadBalance(path)
	if err != nil {
		t.Fatal(err)
	}
	if f.Parent != nil {
		t.Fatal("unexpected overlay on stock input")
	}
	s := serverBalance{file: f}
	oldFlags := noxflags.GetGame()
	t.Cleanup(func() { noxflags.ResetGame(); noxflags.SetGame(oldFlags) })
	setFlags := func(flags uint) { noxflags.ResetGame(); noxflags.SetGame(noxflags.GameFlag(flags)) }
	for _, mode := range fixture.Modes {
		t.Run(mode.Name, func(t *testing.T) {
			setFlags(mode.Flags)
			t.Run("selected_tag", func(t *testing.T) {
				if got := s.Tag(); string(got) != mode.Tag {
					t.Fatalf("actual native gamedata mode=%s want original=%s", got, mode.Tag)
				}
			})
			count := 0
			for _, c := range fixture.Cases {
				if c.Mode != mode.Name {
					continue
				}
				count++
				t.Run(strings.TrimPrefix(c.Name, mode.Name+"/"), func(t *testing.T) {
					if c.Flags != mode.Flags {
						t.Fatal("fixture flag mismatch")
					}
					got := s.FloatInd(strings.ToUpper(c.Key), c.Index)
					if math.Float64bits(got) != c.Expected {
						t.Fatalf("actual native gamedata mode value=%016x want original=%016x flags=%#x", math.Float64bits(got), c.Expected, c.Flags)
					}
					if c.Index == 0 && math.Float64bits(s.Float(c.Key)) != c.Expected {
						t.Fatal("scalar mode differs from original getter")
					}
				})
			}
			if count != 321 {
				t.Fatalf("incomplete mode values: %d", count)
			}
		})
	}
	t.Run("compatibility", func(t *testing.T) {
		for _, c := range []struct {
			name  string
			flags uint
		}{
			{"no_flags_arena", 0}, {"quest_without_online_arena", 0x1000}, {"arena_mode_only", 0x100}, {"unrelated_flag23_arena", 0x400000},
		} {
			t.Run(c.name, func(t *testing.T) {
				setFlags(c.flags)
				if s.Tag() != balance.TagArena {
					t.Fatal("existing fallback changed")
				}
			})
		}
		t.Run("preview_global_and_bounds", func(t *testing.T) {
			setFlags(0x200000)
			if s.Float("BASEHEALTH") != 20 || s.FloatInd("xpTABLE", 2) != 5000 || s.Float("absent") != 0 || s.FloatInd("xpTable", -1) != 0 || s.FloatInd("xpTable", 32) != 0 {
				t.Fatal("global/case/missing/bounds behavior changed")
			}
		})
		t.Run("preview_in_memory_binary64_tags", func(t *testing.T) {
			setFlags(0x200000)
			x := serverBalance{file: &balance.File{Tags: map[balance.Tag]balance.Config{balance.TagSolo: {"scalar": balance.Float(0.1), "array": balance.Array{0.2, 0.3}}, balance.TagArena: {"scalar": balance.Float(0.4), "array": balance.Array{0.5, 0.6}}}}}
			if math.Float64bits(x.Float("scalar")) != math.Float64bits(0.1) || math.Float64bits(x.FloatInd("array", 1)) != math.Float64bits(0.3) {
				t.Fatal("preview tag selection or in-memory precision differs")
			}
		})
	})
	after, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(after, encrypted.Bytes()) {
		t.Fatal("encrypted input changed", err)
	}
	fixtureAfter, err := os.ReadFile(fixturePath)
	if err != nil || sha256.Sum256(raw) != sha256.Sum256(fixtureAfter) {
		t.Fatal("original fixture changed", err)
	}
}
