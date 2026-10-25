package server

import (
	"bytes"
	"encoding/json"
	"math"
	"os"
	"testing"
	"unsafe"

	"github.com/opennox/libs/balance"
	noxflags "github.com/opennox/opennox/v1/common/flags"
)

func TestClassStatsPrecisionOriginal57B200Native(t *testing.T) {
	raw, err := os.ReadFile("testdata/class_stats_precision_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Contexts int      `json:"contexts"`
		Keys     []string `json:"keys"`
		Cases    []struct {
			Name     string       `json:"name"`
			Group    string       `json:"group"`
			Flags    uint         `json:"game_flags"`
			Tag      string       `json:"selected_tag"`
			Values   [16]uint32   `json:"values_bits32"`
			Mult     [3][4]uint32 `json:"multipliers_bits32"`
			Missing  []string     `json:"missing_keys"`
			Expected [16]uint32   `json:"expected_stats_bits32"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Contexts != 458 || len(fixture.Cases) != 458 || len(fixture.Keys) != 16 {
		t.Fatal("incomplete original class-stat contexts")
	}
	savedFlags := noxflags.GetGame()
	defer func() { noxflags.ResetGame(); noxflags.SetGame(savedFlags) }()
	groups := []struct {
		name  string
		count int
	}{{"stock-derived", 6}, {"finite-edges", 432}, {"missing-controls", 16}, {"zero-controls", 4}}
	total := 0
	for _, group := range groups {
		t.Run(group.name, func(t *testing.T) {
			count := 0
			for _, tc := range fixture.Cases {
				if tc.Group != group.name {
					continue
				}
				count++
				total++
				t.Run(tc.Name, func(t *testing.T) {
					noxflags.ResetGame()
					noxflags.SetGame(noxflags.GameFlag(tc.Flags))
					cfg := balance.Config{}
					for i, key := range fixture.Keys {
						missing := false
						for _, v := range tc.Missing {
							if key == v {
								missing = true
							}
						}
						if !missing {
							cfg[key] = balance.Float(float64(math.Float32frombits(tc.Values[i])))
						}
					}
					srv := new(Server)
					srv.Balance.file = &balance.File{Global: balance.Config{}, Tags: map[balance.Tag]balance.Config{balance.Tag(tc.Tag): cfg}}
					stat := func(b [4]uint32) ClassStats {
						return ClassStats{Health: math.Float32frombits(b[0]), Mana: math.Float32frombits(b[1]), Speed: math.Float32frombits(b[2]), Strength: math.Float32frombits(b[3])}
					}
					srv.Players.Mult.Warrior = stat(tc.Mult[0])
					srv.Players.Mult.Wizard = stat(tc.Mult[1])
					srv.Players.Mult.Conjurer = stat(tc.Mult[2])
					srv.Players.Mult2.Warrior = ClassStats{Health: -1, Mana: -2, Speed: -3, Strength: -4}
					srv.Players.Mult2.Wizard = ClassStats{Health: -5, Mana: -6, Speed: -7, Strength: -8}
					srv.Players.Mult2.Conjurer = ClassStats{Health: -9, Mana: -10, Speed: -11, Strength: -12}
					if unsafe.Sizeof(srv.Players.Stats) != 64 || unsafe.Sizeof(ClassStats{}) != 16 {
						t.Fatal("class stats native layout")
					}
					frame := unsafe.Slice((*byte)(unsafe.Pointer(&srv.Players)), int(unsafe.Sizeof(srv.Players)))
					outputOffset := int(uintptr(unsafe.Pointer(&srv.Players.Stats)) - uintptr(unsafe.Pointer(&srv.Players)))
					for i := 0; i < 64; i++ {
						frame[outputOffset+i] = 0x6b
					}
					before := append([]byte(nil), frame...)
					srv.CalcClassStats()
					values := unsafe.Slice((*float32)(unsafe.Pointer(&srv.Players.Stats)), 16)
					var got [16]uint32
					for i, v := range values {
						got[i] = math.Float32bits(v)
					}
					if !bytes.Equal(before[:outputOffset], frame[:outputOffset]) || !bytes.Equal(before[outputOffset+64:], frame[outputOffset+64:]) {
						t.Fatal("nonoutput native player frame or multiplier bytes changed")
					}
					if got != tc.Expected {
						t.Errorf("actual native CalcClassStats bits=%08x; unchanged original=%08x", got, tc.Expected)
					}
				})
			}
			if count != group.count {
				t.Fatalf("group contexts=%d want%d", count, group.count)
			}
		})
	}
	if total != 458 {
		t.Fatalf("total contexts=%d", total)
	}
}
