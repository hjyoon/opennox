package legacy

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/internal/netlist"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type heightProductionLegacyServer4D9020 struct {
	Server
	core *server.Server
}

func (s *heightProductionLegacyServer4D9020) S() *server.Server { return s.core }

type heightProductionInput4D9020 struct {
	Name      string `json:"name"`
	Recipient uint32 `json:"recipient"`
	Flags     uint32 `json:"field5"`
	Code      uint32 `json:"net_code"`
	Extent    uint32 `json:"extent"`
	Height    uint32 `json:"height_bits"`
	Velocity  uint32 `json:"velocity_bits"`
	Tail      uint32 `json:"tail_bits"`
	Class     uint32 `json:"unit_class"`
	QueueFull bool   `json:"queue_full"`
}
type heightProductionRow4D9020 struct {
	Input  heightProductionInput4D9020 `json:"input"`
	Output struct {
		Events []struct {
			Service   string `json:"service"`
			Packet    string `json:"packet"`
			Recipient uint32 `json:"recipient"`
			Kind      uint32 `json:"kind"`
		} `json:"events"`
		Result uint32 `json:"result"`
		Heap   bool   `json:"only_declared_service_heap_writes"`
		CW     bool   `json:"CW_TOP_preserved"`
	} `json:"output"`
}
type heightProductionGuard4D9020 struct {
	Before [32]byte
	Unit   server.Object
	After  [32]byte
}

func TestUnitHeightProductionOriginal4D9020(t *testing.T) {
	file := os.Getenv("NOX_HEIGHT_PRODUCTION_REFERENCE")
	if file == "" {
		file = filepath.Join("..", "server", "testdata", "object_height_production_4d9020.json.gz")
	}
	encoded, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(encoded)) != "efffd2dc94547a837240af0de68e9e970dc8df8b3aa4a333d48a3049e7d3fe3b" {
		t.Fatal("original production gzip seal")
	}
	z, err := gzip.NewReader(bytes.NewReader(encoded))
	if err != nil {
		t.Fatal(err)
	}
	raw, err := io.ReadAll(z)
	if err != nil {
		t.Fatal(err)
	}
	if err = z.Close(); err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(raw)) != "17f2d7c1bd41df592f18df9c1eb59136f189d9d1a040ed60db5584dd065892eb" {
		t.Fatal("original production decoded seal")
	}
	var data struct {
		Contexts []heightProductionRow4D9020 `json:"contexts"`
	}
	if err = json.Unmarshal(raw, &data); err != nil {
		t.Fatal(err)
	}
	if len(data.Contexts) != 544 {
		t.Fatal("original production context count")
	}
	for _, row := range data.Contexts {
		t.Run(row.Input.Name, func(t *testing.T) {
			in := row.Input
			if !row.Output.Heap || !row.Output.CW || len(row.Output.Events) != 1 {
				t.Fatal("original producer invariants")
			}
			event := row.Output.Events[0]
			if event.Service != "netlist" || event.Kind != 1 || event.Recipient != in.Recipient {
				t.Fatal("original actual-code transport boundary")
			}
			want, err := hex.DecodeString(event.Packet)
			if err != nil {
				t.Fatal(err)
			}
			if len(want) != 4 && len(want) != 6 {
				t.Fatal("original fixed packet length")
			}
			g, free := alloc.New(heightProductionGuard4D9020{})
			defer free()
			for i := range g.Before {
				g.Before[i] = 0xa5
				g.After[i] = 0x5a
			}
			unit := &g.Unit
			if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
				t.Fatal("production native object must exceed4GiB")
			}
			unit.ObjClass = object.Class(in.Class)
			unit.Field5 = in.Flags
			unit.NetCode = in.Code
			unit.Extent = in.Extent
			unit.ZVal = math.Float32frombits(in.Height)
			unit.Field27 = math.Float32frombits(in.Velocity)
			unit.Field29 = in.Tail
			memory := unsafe.Slice((*byte)(unsafe.Pointer(g)), unsafe.Sizeof(*g))
			before := bytes.Clone(memory)
			list := netlist.New()
			list.Init()
			defer list.Free()
			core := &server.Server{NetList: list}
			index := ntype.PlayerInd(in.Recipient)
			var retained []byte
			if in.QueueFull {
				for len(retained) < 65536 && list.AddToMsgListCli(index, netlist.Kind1, []byte{0x55}) {
					retained = append(retained, 0x55)
				}
				if len(retained) == 0 || len(retained) >= 65536 {
					t.Fatal("bounded native queue-full prerequisite")
				}
			}
			var result int32
			old := GetServer
			GetServer = func() Server { return &heightProductionLegacyServer4D9020{core: core} }
			defer func() { GetServer = old }()
			result = heightOriginalCEntry4D9020(int32(in.Recipient), unit)
			if uint32(result) != row.Output.Result {
				t.Fatal("full public C transport result", uint32(result), row.Output.Result)
			}
			queued := list.CopyPacketsA(index, netlist.Kind1)
			if in.QueueFull {
				if row.Output.Result != 0 || !bytes.Equal(queued, retained) {
					t.Fatal("full queue must preserve existing packets and reject height", len(queued), len(retained))
				}
			} else if row.Output.Result != 1 || !bytes.Equal(queued, want) {
				t.Fatalf("actual production packet%x, original%x", queued, want)
			}
			if got := list.CopyPacketsA((index+1)%32, netlist.Kind1); len(got) != 0 {
				t.Fatal("unintended recipient", got)
			}
			if !bytes.Equal(memory, before) {
				t.Fatal("production Object/neighbor guards changed")
			}
		})
	}
}
