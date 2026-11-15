package legacy

import (
	"bytes"
	_ "embed"
	"encoding/binary"
	"fmt"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

//go:embed testdata/trace_ray9_original_5374b0_v1.json.gz
var traceRay9Original []byte

type traceRay9Case struct {
	Input struct {
		Vector    int    `json:"vector"`
		MapReturn uint32 `json:"map_return"`
	} `json:"input"`
	Events []struct {
		Args   [4]int32 `json:"args"`
		Result uint32   `json:"result"`
	} `json:"events"`
	Return uint32 `json:"return_u32"`
}

func TestTraceRay9NativeLayout(t *testing.T) {
	if got := traceRay9LayoutC(); got != [5]uintptr{16, 0, 4, 8, 12} {
		t.Fatal("independent native C vector layout", got)
	}
}

func TestTraceRay9NativeOriginal(t *testing.T) {
	var rows []traceRay9Case
	npcVoiceOriginalDecode(t, traceRay9Original, "7219ace93c56863b06630657065e28c4eeb09a529acbd247d320c026f74b018b", "e43dd3d411a1515b1516a87f99b4ff72faf504cb71469e924611b089dc16bd1d", 16, &rows)
	for route := 0; route < 3; route++ {
		for _, placement := range []int{0, 64, 512} {
			raw, free := alloc.MakeHigh([]byte{}, 65536)
			t.Cleanup(free)
			if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(&raw[0])) <= math.MaxUint32 {
				t.Fatal("native vector owner must be above4GiB")
			}
			for _, c := range rows {
				t.Run(fmt.Sprintf("route%d/place%d/vector%d/word%x", route, placement, c.Input.Vector, c.Input.MapReturn), func(t *testing.T) {
					for i := range raw {
						raw[i] = byte(i*31 + 19)
					}
					var line unsafe.Pointer
					if c.Input.Vector == 0 {
						line = unsafe.Pointer(&raw[0x1000+placement])
					}
					before := bytes.Clone(raw)
					calls := 0
					restore := traceRay9FixtureService(func(got, position, grid unsafe.Pointer, flags int8) int32 {
						calls++
						if got != line || position != nil || grid != nil || flags != 9 || len(c.Events) != 1 {
							t.Fatal("original raw-pointer/null-output/flags contract changed")
						}
						args := [4]int32{int32(c.Input.Vector), -1, -1, int32(flags)}
						if args != c.Events[0].Args || c.Input.MapReturn != c.Events[0].Result {
							t.Fatal("original service observation changed")
						}
						return int32(c.Input.MapReturn)
					})
					defer restore()
					got := traceRay9ActualCall(route, line)
					if uint32(got) != c.Return || calls != 1 || !bytes.Equal(raw, before) {
						t.Fatalf("return/calls=%x/%d or64KiB input owner changed", uint32(got), calls)
					}
				})
			}
		}
	}
}

func TestTraceRay9NativeDefaultMap(t *testing.T) {
	previous := GetServer
	t.Cleanup(func() { GetServer = previous })
	for route := 0; route < 3; route++ {
		for _, placement := range []int{0, 64, 512} {
			for _, blocked := range []bool{false, true} {
				t.Run(fmt.Sprintf("route%d/place%d/blocked%t", route, placement, blocked), func(t *testing.T) {
					srv, _, _ := monsterLookAtFixture5125A0(t)
					srv.Map.Init()
					t.Cleanup(srv.Map.Free)
					if srv.Walls.InitHigh() == 0 {
						t.Fatal("cannot initialize actual native wall map")
					}
					t.Cleanup(srv.Walls.Free)
					GetServer = func() Server { return &arachnaphobiaCastLegacyServer52DC80{srv: srv} }
					raw, free := alloc.MakeHigh([]byte{}, 65536)
					t.Cleanup(free)
					for i := range raw {
						raw[i] = byte(i*31 + 19)
					}
					point := [4]float32{300, 300, 330, 300}
					want := int32(1)
					if blocked {
						point[0], point[1], want = -100, -200, 0
					}
					for i, value := range point {
						binary.LittleEndian.PutUint32(raw[0x1000+placement+4*i:], math.Float32bits(value))
					}
					before := bytes.Clone(raw)
					got := traceRay9ActualCall(route, unsafe.Pointer(&raw[0x1000+placement]))
					if got != want || !bytes.Equal(raw, before) {
						t.Fatalf("actual default map result%d want%d or vector owner changed", got, want)
					}
				})
			}
		}
	}
}
