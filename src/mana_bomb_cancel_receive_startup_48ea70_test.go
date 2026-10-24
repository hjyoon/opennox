package opennox

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"
	"unsafe"

	"github.com/opennox/libs/datapath"
	"github.com/opennox/libs/strman"
	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/internal/binfile"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestManaBombCancelProductionReceiveStartupContextV1(t *testing.T) {
	raw, err := os.ReadFile("testdata/mana_bomb_cancel_receive_startup_48ea70.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Unchanged bool `json:"original_unchanged"`
		Rows      []struct {
			Radius    float64  `json:"radius"`
			Mode      string   `json:"mode"`
			Draws     int      `json:"actual_rng_draws"`
			Positions [][2]int `json:"positions"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(raw, &fixture); err != nil || !fixture.Unchanged || len(fixture.Rows) != 14 {
		t.Fatalf("original full receive fixture: %v", err)
	}
	base := server.New(nil, nil, strman.New())
	t.Cleanup(base.Close)
	data := t.TempDir()
	oldData := datapath.Data()
	datapath.SetData(data)
	t.Cleanup(func() { datapath.SetData(oldData) })
	connected := memmap.PtrUint32(0x5D4594, 815764)
	oldConnected := *connected
	*connected = 1
	t.Cleanup(func() { *connected = oldConnected })
	for _, row := range fixture.Rows {
		if row.Mode != "actual-startup-inherited" {
			continue
		}
		t.Run(fmt.Sprintf("radius-%g", row.Radius), func(t *testing.T) {
			cli := client.NewClient(nil, nil, base)
			t.Cleanup(cli.Close)
			t.Cleanup(func() { alloc.Free(cli.Viewport()) })
			cli.Things.Nox_things_free_44C580_B()
			t.Cleanup(cli.Things.Nox_things_free_44C580_B)
			cli.Objs.Init(151)
			t.Cleanup(cli.Objs.Free)
			// A minimal parsed effect type isolates geometry without invoking a renderer.
			bytes := append([]byte{9}, []byte("CyanSpark")...)
			bytes = append(bytes, 0)
			native, _ := alloc.Make([]byte{}, len(bytes))
			copy(native, bytes)
			file := binfile.NewMemFile(unsafe.Pointer(&native[0]), len(native))
			if err := cli.Things.ReadType(file, make([]byte, 512)); err != nil {
				file.Free()
				t.Fatal(err)
			}
			file.Free()
			if cli.Things.IndByID("CyanSpark") != 1 {
				t.Fatal("actual parsed native type")
			}
			c := &Client{Client: cli, srv: &Server{Server: base}}
			if err := os.WriteFile(filepath.Join(data, "gamedata.yml"), []byte("ManaBombOutRadius: "+strconv.FormatFloat(row.Radius, 'g', -1, 64)+"\n"), 0600); err != nil {
				t.Fatal(err)
			}
			if err := base.Balance.Read(); err != nil {
				t.Fatal(err)
			}
			if base.Balance.Float("ManaBombOutRadius") != row.Radius || base.Frame() != 0 {
				t.Fatal("native balance/frame setup")
			}
			base.Rand.Other.Reset(0)
			packet := []byte{0xa3, 0, 0, 0, 0}
			binary.LittleEndian.PutUint16(packet[1:3], 1000)
			binary.LittleEndian.PutUint16(packet[3:5], 1000)
			if got := c.handleManaBombCancelPacketNative48EA70(packet); got != 5 {
				t.Fatalf("actual packet branch consumed %d", got)
			}
			var positions [][2]int
			for dr := cli.Objs.List1; dr != nil; dr = dr.NextPtr {
				if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(unsafe.Pointer(dr)) <= math.MaxUint32 {
					t.Fatal("native drawable did not exceed4GiB")
				}
				positions = append(positions, [2]int{dr.PosVec.X, dr.PosVec.Y})
			}
			for i, j := 0, len(positions)-1; i < j; i, j = i+1, j-1 {
				positions[i], positions[j] = positions[j], positions[i]
			}
			if base.Rand.Other.Index() != row.Draws || len(positions) != 150 {
				t.Fatalf("actual native RNG/spawn count %d/%d", base.Rand.Other.Index(), len(positions))
			}
			if !reflect.DeepEqual(positions, row.Positions) {
				t.Fatalf("actual production receive positions differ from original startup path at radius%g: native first%v original first%v", row.Radius, positions[:3], row.Positions[:3])
			}
		})
	}
}
