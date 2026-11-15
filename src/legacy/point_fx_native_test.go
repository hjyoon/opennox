package legacy

import (
	"bytes"
	_ "embed"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"os"
	"os/exec"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/internal/netlist"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

//go:embed testdata/point_fx_original_522ff0_v1.json.gz
var pointFXOriginal522FF0 []byte

type pointFXOriginalCase struct {
	Input struct {
		Name       string     `json:"name"`
		Root       string     `json:"root"`
		Position   [2]uint32  `json:"position"`
		Op         uint32     `json:"op_word"`
		Subtype    uint32     `json:"subtype_word"`
		SendReturn uint32     `json:"send_return"`
		PointWrite *[2]uint32 `json:"point_write"`
	} `json:"input"`
	Events []struct {
		Kind   string    `json:"kind"`
		Owner  string    `json:"point_owner"`
		Packet string    `json:"packet_hex"`
		Size   int       `json:"size"`
		Before [2]uint32 `json:"position_before"`
		After  [2]uint32 `json:"position_after"`
		Return uint32    `json:"declared_send_return_u32"`
	} `json:"events"`
	RawEAX uint32  `json:"raw_forwarded_eax"`
	Return *uint32 `json:"current_int_compatibility_return_u32"`
}

func pointFXOriginalCases(t *testing.T) []pointFXOriginalCase {
	t.Helper()
	var rows []pointFXOriginalCase
	npcVoiceOriginalDecode(t, pointFXOriginal522FF0, "1788dbb7acec8af91b1140311e5a67500ee6155ab1edd0ca0c03424191832054", "e1e0c9ec249515b3dcfb18e4295252215e4c00838a436b4eb8e65ee33cb651bf", 2628, &rows)
	return rows
}

func pointFXRoot(t *testing.T, name string) int {
	t.Helper()
	for i, s := range []string{"point", "point-subtype", "spark", "generator-break"} {
		if name == s {
			return i
		}
	}
	t.Fatal("unknown original point FX root", name)
	return -1
}

func pointFXForeignOwner(t *testing.T, offset int) ([]byte, *types.Pointf) {
	t.Helper()
	raw, free := alloc.MakeHigh([]byte{}, 0x10000)
	t.Cleanup(free)
	point := (*types.Pointf)(unsafe.Pointer(&raw[offset]))
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(point)) <= math.MaxUint32 {
		t.Fatal("point owner must be above 4 GiB")
	}
	return raw, point
}

func pointFXWriteBits(raw []byte, offset int, bits [2]uint32) {
	binary.LittleEndian.PutUint32(raw[offset:offset+4], bits[0])
	binary.LittleEndian.PutUint32(raw[offset+4:offset+8], bits[1])
}

func TestPointFXNativeLayout(t *testing.T) {
	want := [6]uintptr{unsafe.Sizeof(types.Pointf{}), unsafe.Offsetof(types.Pointf{}.X), unsafe.Offsetof(types.Pointf{}.Y), unsafe.Sizeof(float32(0)), 1, 4}
	if got := pointFXLayoutC(); got != want {
		t.Fatalf("independent C/native point and scalar layout %v want %v", got, want)
	}
}

func TestPointFXNativeOriginalMatrix(t *testing.T) {
	rows := pointFXOriginalCases(t)
	old := pointFXRuntimeForCall522FF0
	t.Cleanup(func() { pointFXRuntimeForCall522FF0 = old })
	counts := [4]int{}
	for route := 0; route < 3; route++ {
		for _, offset := range []int{0x1000, 0x2044, 0xfef8} {
			raw, point := pointFXForeignOwner(t, offset)
			for _, c := range rows {
				t.Run(fmt.Sprintf("route%d/offset%x/%s", route, offset, c.Input.Name), func(t *testing.T) {
					root := pointFXRoot(t, c.Input.Root)
					if len(c.Events) != 1 || c.Events[0].Kind != "send" || c.Events[0].Owner != "input" || c.Events[0].Before != c.Input.Position || c.Events[0].Return != c.Input.SendReturn || c.RawEAX != c.Input.SendReturn {
						t.Fatal("unexpected original observation contract")
					}
					e := c.Events[0]
					packet, err := hex.DecodeString(e.Packet)
					if err != nil || len(packet) != e.Size {
						t.Fatal("bad original wire fixture", err)
					}
					for i := range raw {
						raw[i] = byte(i*31 + 19)
					}
					pointFXWriteBits(raw, offset, c.Input.Position)
					before := bytes.Clone(raw)
					want := bytes.Clone(raw)
					pointFXWriteBits(want, offset, e.After)
					calls := 0
					pointFXRuntimeForCall522FF0 = func() server.PointFXNativeRuntime522FF0 {
						return server.PointFXNativeRuntime522FF0{Send: func(owner *types.Pointf, data []byte) int32 {
							calls++
							if owner != point || !bytes.Equal(raw, before) || !bytes.Equal(data, packet) || len(data) != e.Size {
								t.Fatalf("send owner/state/packet: owner%p want%p packet%x want%x", owner, point, data, packet)
							}
							address := uintptr(unsafe.Pointer(unsafe.SliceData(data)))
							start := uintptr(unsafe.Pointer(unsafe.SliceData(raw)))
							if address >= start && address < start+uintptr(len(raw)) {
								t.Fatal("packet aliases foreign input owner")
							}
							if c.Input.PointWrite != nil {
								pointFXWriteBits(raw, offset, *c.Input.PointWrite)
							}
							if !bytes.Equal(data, packet) {
								t.Fatal("packet changed with send-time point mutation")
							}
							return int32(c.Input.SendReturn)
						}}
					}
					result := pointFXActualCall(route, root, byte(c.Input.Op), byte(c.Input.Subtype), point)
					if root == 3 {
						if c.Return != nil {
							t.Fatal("void entry must not acquire a defined return")
						}
					} else if c.Return == nil || uint32(result) != *c.Return {
						t.Fatalf("signed-dword result %#x want%v", uint32(result), c.Return)
					}
					if calls != 1 || !bytes.Equal(raw, want) {
						t.Fatalf("send count/whole foreign owner changed: calls%d", calls)
					}
					counts[root]++
					runtime.KeepAlive(point)
				})
			}
		}
	}
	if counts != [4]int{5913, 5913, 5913, 5913} {
		t.Fatal("missing original root/route/placement cases", counts)
	}
	t.Log("2628 original cases x3 public-C/typed-C/production-Go routes x3 foreign placements; full 64 KiB canaries, exact packet/point identity/send writes, signed int returns separate from void")
}

func TestPointFXNativeNestedPacketOwnership(t *testing.T) {
	old := pointFXRuntimeForCall522FF0
	t.Cleanup(func() { pointFXRuntimeForCall522FF0 = old })
	for route := 0; route < 3; route++ {
		raw, point := pointFXForeignOwner(t, 0x1000)
		pointFXWriteBits(raw, 0x1000, [2]uint32{math.Float32bits(1.5), math.Float32bits(-1.5)})
		depth, calls := 0, 0
		var outer []byte
		pointFXRuntimeForCall522FF0 = func() server.PointFXNativeRuntime522FF0 {
			return server.PointFXNativeRuntime522FF0{Send: func(owner *types.Pointf, packet []byte) int32 {
				calls++
				if owner != point {
					t.Fatal("nested native owner identity")
				}
				if depth == 0 {
					if !bytes.Equal(packet, []byte{0xf0, 1, 0, 0xff, 0xff}) {
						t.Fatal("outer packet", packet)
					}
					outer = packet
					for i := range outer {
						outer[i] = 0x5a
					}
					depth++
					if got := pointFXActualCall(route, 2, 0, 0x80, point); got != math.MinInt32 {
						t.Fatal("nested signed result", got)
					}
					depth--
					if !bytes.Equal(outer, []byte{0x5a, 0x5a, 0x5a, 0x5a, 0x5a}) {
						t.Fatal("nested call overwrote caller-owned packet")
					}
					return -1
				}
				if !bytes.Equal(packet, []byte{0x93, 1, 0, 0xff, 0xff, 0x80}) || unsafe.SliceData(packet) == unsafe.SliceData(outer) {
					t.Fatal("nested packet bytes/ownership", packet)
				}
				return math.MinInt32
			}}
		}
		if got := pointFXActualCall(route, 0, 0xf0, 0, point); got != -1 || calls != 2 {
			t.Fatal("outer result/count", got, calls)
		}
	}
}

func TestPointFXNativeDefaultNetwork(t *testing.T) {
	const child = "OPENNOX_POINT_FX_NETWORK_CHILD"
	if os.Getenv(child) != "1" {
		program, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(program, "-test.run=^TestPointFXNativeDefaultNetwork$", "-test.v", "-test.timeout=3m")
		cmd.Env = append(os.Environ(), child+"=1")
		out, err := cmd.CombinedOutput()
		t.Logf("actual default network composition:\n%s", out)
		if err != nil {
			t.Fatal(err)
		}
		return
	}
	s, receiver := generatorDamageCompositionServer(t)
	raw, point := pointFXForeignOwner(t, 0x1000)
	compositions := 0
	for _, c := range pointFXOriginalCases(t) {
		// The original matrix still checks invalid/out-of-viewport points via
		// its declared send boundary. Real network composition uses visible ones.
		x, y := math.Float32frombits(c.Input.Position[0]), math.Float32frombits(c.Input.Position[1])
		if math.IsNaN(float64(x)) || math.IsInf(float64(x), 0) || math.IsNaN(float64(y)) || math.IsInf(float64(y), 0) || math.Abs(float64(x)) > 65537 || math.Abs(float64(y)) > 65537 {
			continue
		}
		packet, err := hex.DecodeString(c.Events[0].Packet)
		if err != nil {
			t.Fatal(err)
		}
		for route := 0; route < 3; route++ {
			pointFXWriteBits(raw, 0x1000, c.Input.Position)
			receiver.PlayerUnit.PosVec = *point
			if got := pointFXActualCall(route, pointFXRoot(t, c.Input.Root), byte(c.Input.Op), byte(c.Input.Subtype), point); got != 0 {
				t.Fatalf("canonical network return%d", got)
			}
			// NetList must own the packet after both this local packet and the
			// original position are gone or changed.
			pointFXWriteBits(raw, 0x1000, [2]uint32{0, 0})
			if got := s.NetList.CopyPacketsA(ntype.PlayerInd(receiver.PlayerInd), netlist.Kind1); !bytes.Equal(got, packet) {
				t.Fatalf("%s/route%d actual NetList%x want%x", c.Input.Name, route, got, packet)
			}
			compositions++
		}
	}
	if compositions != 2412 {
		t.Fatal("missing real network compositions", compositions)
	}
	t.Logf("actual default compositions=%d, no runtime override; high native receiver/point, real viewport and NetList copy, original packets", compositions)
}
