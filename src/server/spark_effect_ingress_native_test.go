package server

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"math"
	"os"
	"reflect"
	"strconv"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/noxscript/ns/v4/effect"

	"github.com/opennox/libs/common"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/internal/netlist"
)

type sparkEffectObservedPosition514426 struct {
	value types.Pointf
	name  string
	reads *[]string
}

func (p *sparkEffectObservedPosition514426) Pos() types.Pointf {
	*p.reads = append(*p.reads, p.name)
	return p.value
}

func TestSparkExplosionEffectOriginal514426Native(t *testing.T) {
	raw, err := os.ReadFile("testdata/spark_effect_ingress_514426_5231b0.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		OriginalSHA string `json:"original_sha256"`
		Unchanged   bool   `json:"original_unchanged"`
		Startup     uint32 `json:"actual_startup_control_word"`
		Rows        []struct {
			Name  string `json:"name"`
			Group string `json:"group"`
			Mode  string `json:"mode"`
			X     string `json:"x_bits"`
			Y     string `json:"y_bits"`
			Count string `json:"count_bits"`
			Send  struct {
				Packet string `json:"packet"`
			} `json:"send"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	if !fixture.Unchanged || fixture.OriginalSHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || fixture.Startup != 0xe3f || len(fixture.Rows) != 284 {
		t.Fatal("incomplete independent original spark-effect observations")
	}
	fromBits := func(value string) float32 {
		t.Helper()
		bits, err := strconv.ParseUint(value, 16, 32)
		if err != nil {
			t.Fatal(err)
		}
		return math.Float32frombits(uint32(bits))
	}
	nl := netlist.New()
	nl.Init()
	t.Cleanup(nl.Free)
	s := &Server{NetList: nl}
	s.Players.list = make([]Player, common.MaxPlayers)
	const index = ntype.PlayerInd(0)
	pl := &s.Players.list[index]
	pl.Active = 1
	pl.PlayerInd = byte(index)
	pl.Field10 = 4096
	pl.Field12 = 4096
	update := &PlayerUpdateData{Player: pl}
	unit := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
	pl.PlayerUnit = unit
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, ptr := range []unsafe.Pointer{unsafe.Pointer(pl), unsafe.Pointer(unit), unsafe.Pointer(update)} {
			if uintptr(ptr) <= math.MaxUint32 {
				t.Fatalf("native player/unit/update pointer %p is not above UINT32_MAX", ptr)
			}
		}
	}
	selected := 0
	for _, group := range []string{"count", "x", "y"} {
		t.Run(group, func(t *testing.T) {
			for _, row := range fixture.Rows {
				if row.Mode != "actual-startup-inherited" || row.Group != group {
					continue
				}
				selected++
				t.Run(row.Name, func(t *testing.T) {
					position := types.Ptf(fromBits(row.X), fromBits(row.Y))
					unit.PosVec = position
					nl.ResetByInd(index, netlist.Kind1)
					var reads []string
					first := &sparkEffectObservedPosition514426{position, "first", &reads}
					second := &sparkEffectObservedPosition514426{types.Ptf(fromBits(row.Count), -7), "second", &reads}
					s.NoxScriptNS().Effect(effect.Effect("SPARK_EXPLOSION"), first, second)
					if !reflect.DeepEqual(reads, []string{"first", "second"}) || unit.PosVec != position {
						t.Fatalf("source reads=%v position=%v", reads, unit.PosVec)
					}
					want, err := hex.DecodeString(row.Send.Packet)
					if err != nil || len(want) != 6 || want[0] != 0x93 {
						t.Fatal("invalid independent original spark packet")
					}
					if got := nl.CopyPacketsA(index, netlist.Kind1); !bytes.Equal(got, want) {
						t.Fatalf("actual Effect->packet->view filter->NetList=%x, original=%x", got, want)
					}
				})
			}
		})
	}
	if selected != 71 {
		t.Fatalf("selected startup rows=%d", selected)
	}
}
