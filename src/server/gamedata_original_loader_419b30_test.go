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

func TestGamedataOriginalLoader419B30(t *testing.T) {
	raw, err := os.ReadFile("testdata/gamedata_original_loader.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Contexts int `json:"contexts"`
		Cases    []struct {
			Group    string `json:"group"`
			Tag      string `json:"tag"`
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
		Edges []struct {
			Key   string `json:"key"`
			Token string `json:"token"`
		} `json:"edge_records"`
	}
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Contexts != 423 || len(fixture.Cases) != 423 || len(fixture.Stock) != 215 || len(fixture.Edges) != 36 {
		t.Fatal("incomplete independently executed original contexts")
	}
	writeBalance := func(t *testing.T, plain, overlay string) string {
		t.Helper()
		path := filepath.Join(t.TempDir(), balance.GamedataFile)
		var encrypted bytes.Buffer
		w, err := crypt.NewWriter(&encrypted, crypt.GameDataBin)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(plain)); err != nil {
			t.Fatal(err)
		}
		if err := w.Close(); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, encrypted.Bytes(), 0600); err != nil {
			t.Fatal(err)
		}
		if overlay != "" {
			if err := os.WriteFile(filepath.Join(filepath.Dir(path), "gamedata.yml"), []byte(overlay), 0600); err != nil {
				t.Fatal(err)
			}
		}
		return path
	}
	var text strings.Builder
	for _, record := range fixture.Stock {
		if record.Tag != "GLOBAL" {
			fmt.Fprintf(&text, "%s ", record.Tag)
		}
		fmt.Fprintf(&text, "%s = %s;\n", record.Key, strings.Join(record.Tokens, ", "))
	}
	for _, edge := range fixture.Edges {
		fmt.Fprintf(&text, "%s = %s;\n", edge.Key, edge.Token)
	}
	path := writeBalance(t, text.String(), "")
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	f, err := balance.ReadBalance(path)
	if err != nil {
		t.Fatal(err)
	}
	if f.Parent != nil {
		t.Fatal("unexpected stock YAML overlay")
	}
	s := serverBalance{file: f}
	oldFlags := noxflags.GetGame()
	t.Cleanup(func() { noxflags.ResetGame(); noxflags.SetGame(oldFlags) })
	mode := func(tag string) {
		noxflags.UnsetGame(noxflags.GameModeCoop)
		if tag != "ARENA" {
			noxflags.SetGame(noxflags.GameModeCoop)
		}
	}
	for _, group := range []string{"stock", "decimal_edges"} {
		t.Run(group, func(t *testing.T) {
			for _, c := range fixture.Cases {
				if c.Group != group {
					continue
				}
				t.Run(strings.TrimPrefix(c.Name, group+"/"), func(t *testing.T) {
					mode(c.Tag)
					got := s.FloatInd(strings.ToUpper(c.Key), c.Index)
					if math.Float64bits(got) != c.Expected {
						t.Fatalf("actual native gamedata=%016x want original=%016x", math.Float64bits(got), c.Expected)
					}
					tag := balance.Tag(c.Tag)
					if c.Tag == "GLOBAL" {
						tag = ""
					}
					values := f.Array(tag, c.Key)
					if c.Index >= len(values) || math.Float64bits(values[c.Index]) != c.Expected {
						t.Fatal("array loader changed original slot/precision")
					}
					if c.Index == 0 {
						if math.Float64bits(s.Float(c.Key)) != c.Expected {
							t.Fatal("scalar getter differs from original index zero")
						}
						first, ok := f.Float(tag, c.Key)
						if !ok || math.Float64bits(first) != c.Expected {
							t.Fatal("balance scalar selection differs from original")
						}
					}
				})
			}
		})
	}
	t.Run("compatibility", func(t *testing.T) {
		t.Run("missing_and_bounds", func(t *testing.T) {
			mode("SOLO")
			if s.Float("absent") != 0 || s.FloatInd("absent", 0) != 0 || s.FloatInd("xpTable", -1) != 0 || s.FloatInd("xpTable", 32) != 0 {
				t.Fatal("missing/bounds behavior changed")
			}
		})
		t.Run("global_fallback_and_case", func(t *testing.T) {
			for _, tag := range []string{"SOLO", "ARENA"} {
				mode(tag)
				if s.Float("BASEHEALTH") != 20 || s.FloatInd("xpTABLE", 2) != 5000 {
					t.Fatal("mode/global lookup changed")
				}
			}
		})
		t.Run("YAML_scalar_binary64", func(t *testing.T) {
			p := writeBalance(t, "legacy = 2;\n", "legacy: 0.1\n")
			x, e := balance.ReadBalance(p)
			if e != nil {
				t.Fatal(e)
			}
			if math.Float64bits(x.FloatDef("", "legacy", 0)) != math.Float64bits(0.1) {
				t.Fatal("YAML scalar precision changed")
			}
		})
		t.Run("YAML_array_tags_binary64", func(t *testing.T) {
			p := writeBalance(t, "legacy = 2;\n", "custom:\n  SOLO: [0.1, 0.2]\n  ARENA: [0.3, 0.4]\n")
			x, e := balance.ReadBalance(p)
			if e != nil {
				t.Fatal(e)
			}
			for tag, want := range map[balance.Tag][]float64{balance.TagSolo: {0.1, 0.2}, balance.TagArena: {0.3, 0.4}} {
				got := x.Array(tag, "custom")
				if len(got) != 2 {
					t.Fatal(got)
				}
				for i := range want {
					if math.Float64bits(got[i]) != math.Float64bits(want[i]) {
						t.Fatal("YAML array precision changed")
					}
				}
			}
		})
		t.Run("YAML_parent_fallback", func(t *testing.T) {
			p := writeBalance(t, "legacy = 8.5;\n", "custom: 0.1\n")
			x, e := balance.ReadBalance(p)
			if e != nil {
				t.Fatal(e)
			}
			if x.Parent == nil || x.FloatDef(balance.TagSolo, "legacy", 0) != 8.5 {
				t.Fatal("YAML parent lookup changed")
			}
		})
		t.Run("invalid_decimal_error", func(t *testing.T) {
			p := writeBalance(t, "invalid = 1.25tail;\n", "")
			if _, e := balance.ReadBalance(p); e == nil || !strings.Contains(e.Error(), "expected: number") {
				t.Fatal("strict legacy validation changed", e)
			}
		})
		t.Run("in_memory_binary64", func(t *testing.T) {
			x := &balance.File{Global: balance.Config{"scalar": balance.Float(0.1), "array": balance.Array{0.2, 0.3}}}
			if math.Float64bits(x.FloatDef("", "scalar", 0)) != math.Float64bits(0.1) || math.Float64bits(x.Array("", "array")[1]) != math.Float64bits(0.3) {
				t.Fatal("in-memory config precision changed")
			}
		})
	})
	after, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("native encrypted input changed", err)
	}
	fixtureAfter, err := os.ReadFile("testdata/gamedata_original_loader.json")
	if err != nil || sha256.Sum256(raw) != sha256.Sum256(fixtureAfter) {
		t.Fatal("original fixture changed", err)
	}
}
