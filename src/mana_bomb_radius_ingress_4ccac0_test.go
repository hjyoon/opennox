package opennox

import (
	"encoding/json"
	"fmt"
	"image"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/opennox/libs/datapath"
	"github.com/opennox/libs/strman"
	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/server"
)

func TestManaBombRadiusIngressOriginal4CCAC0(t *testing.T) {
	raw, err := os.ReadFile("testdata/mana_bomb_radius_ingress_4ccac0.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Unchanged bool `json:"original_unchanged"`
		Rows      []struct {
			Bits   uint32 `json:"bits"`
			CW     uint32 `json:"fpcw"`
			Result int32  `json:"result"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(raw, &fixture); err != nil || !fixture.Unchanged || len(fixture.Rows) != 60 {
		t.Fatalf("original ingress fixture: %v", err)
	}
	base := server.New(nil, nil, strman.New())
	t.Cleanup(base.Close)
	c := &Client{Client: &client.Client{}, srv: &Server{Server: base}}
	data := t.TempDir()
	oldData := datapath.Data()
	datapath.SetData(data)
	t.Cleanup(func() { datapath.SetData(oldData) })
	for _, row := range fixture.Rows {
		if row.CW != 0xe7f {
			continue
		}
		t.Run(fmt.Sprintf("f32-%08x", row.Bits), func(t *testing.T) {
			value := float64(math.Float32frombits(row.Bits))
			text := strconv.FormatFloat(value, 'g', -1, 64)
			switch {
			case math.IsNaN(value):
				text = ".nan"
			case math.IsInf(value, 1):
				text = ".inf"
			case math.IsInf(value, -1):
				text = "-.inf"
			}
			if err := os.WriteFile(filepath.Join(data, "gamedata.yml"), []byte("ManaBombOutRadius: "+text+"\n"), 0600); err != nil {
				t.Fatal(err)
			}
			if err := c.srv.Balance.Read(); err != nil {
				t.Fatal(err)
			}
			loaded := c.srv.Balance.Float("ManaBombOutRadius")
			if math.IsNaN(value) {
				if !math.IsNaN(loaded) {
					t.Fatal("native balance NaN fixture")
				}
			} else if loaded != value {
				t.Fatalf("native balance %v want%v", loaded, value)
			}
			hooks := c.manaBombDrawableHooks4CCAC0() // execute the actual production radius closure
			hooks.typeID = func(string) int { return 1 }
			hooks.frame = func() uint32 { return 2 }
			var radius int32
			seen := false
			hooks.random = func(lo, hi int) int {
				if !seen {
					radius, seen = int32(hi), true
				}
				return 0
			}
			hooks.spawn = func(int, image.Point) *client.Drawable { return nil }
			if got := updateManaBombCharge4CCAC0(&client.Drawable{}, hooks); got != 1 || !seen {
				t.Fatal("actual charge branch did not reach radius RNG")
			}
			if radius != row.Result {
				t.Fatalf("actual native radius=%d original=%d", radius, row.Result)
			}
		})
	}
}
