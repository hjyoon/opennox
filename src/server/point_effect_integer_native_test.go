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

	"github.com/opennox/libs/common"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/noxscript/ns/v4/effect"

	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/internal/netlist"
)

type pointEffectObservedPosition5143F0 struct {
	value types.Pointf
	reads *int
}

func (p *pointEffectObservedPosition5143F0) Pos() types.Pointf {
	*p.reads++
	return p.value
}

func TestPointEffectOriginal5143F0Native(t *testing.T) {
	raw, err := os.ReadFile("testdata/point_effect_integer_5143f0_522ff0.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		OriginalSHA string   `json:"original_sha256"`
		Unchanged   bool     `json:"original_unchanged"`
		Startup     uint32   `json:"actual_startup_control_word"`
		Names       []string `json:"effect_names"`
		Rows        []struct {
			Name   string `json:"name"`
			Effect string `json:"effect"`
			Opcode byte   `json:"opcode"`
			Mode   string `json:"mode"`
			X      string `json:"x_bits"`
			Y      string `json:"y_bits"`
			Send   struct {
				Packet string `json:"packet"`
			} `json:"send"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	if !fixture.Unchanged || fixture.OriginalSHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || fixture.Startup != 0xe3f || len(fixture.Names) != 15 || len(fixture.Rows) != 1920 {
		t.Fatal("incomplete independent original point-effect observations")
	}
	fromBits := func(value string) float32 {
		t.Helper()
		bits, err := strconv.ParseUint(value, 16, 32)
		if err != nil {
			t.Fatal(err)
		}
		return math.Float32frombits(uint32(bits))
	}
	pointBits := func(p types.Pointf) [2]uint32 {
		return [2]uint32{math.Float32bits(p.X), math.Float32bits(p.Y)}
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
	for _, name := range fixture.Names {
		t.Run(name, func(t *testing.T) {
			for _, row := range fixture.Rows {
				if row.Mode != "actual-startup-inherited" || row.Effect != name {
					continue
				}
				selected++
				t.Run(row.Name, func(t *testing.T) {
					position := types.Ptf(fromBits(row.X), fromBits(row.Y))
					unit.PosVec = position
					nl.ResetByInd(index, netlist.Kind1)
					reads := 0
					first := &pointEffectObservedPosition5143F0{position, &reads}
					s.NoxScriptNS().Effect(effect.Effect(name), first, nil)
					if reads != 1 || !reflect.DeepEqual(pointBits(unit.PosVec), pointBits(position)) || pointBits(first.value) != pointBits(position) {
						t.Fatalf("source reads=%d position=%v", reads, unit.PosVec)
					}
					want, err := hex.DecodeString(row.Send.Packet)
					if err != nil || len(want) != 5 || want[0] != row.Opcode {
						t.Fatal("invalid independent original point-effect packet")
					}
					if got := nl.CopyPacketsA(index, netlist.Kind1); !bytes.Equal(got, want) {
						t.Fatalf("actual Effect->packet->view filter->NetList=%x, original=%x", got, want)
					}
				})
			}
		})
	}
	if selected != 480 {
		t.Fatalf("selected startup rows=%d", selected)
	}
}
